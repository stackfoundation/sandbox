package kube

import (
	"fmt"
	"strings"
	"sync/atomic"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"

	log "github.com/stackfoundation/core/pkg/log"
	workflowsv1 "github.com/stackfoundation/core/pkg/workflows/v1"
)

func cleanupPodIfNecessary(context *podContext) {
	if !context.podDeleted {
		log.Debugf("Deleting pod %v", context.pod.Name)
		context.podsClient.Delete(context.pod.Name, &metav1.DeleteOptions{})

		if context.service != nil {
			log.Debugf("Deleting service %v", context.service.Name)
			context.serviceClient.Delete(context.service.Name, &metav1.DeleteOptions{})
		}

		context.creationSpec.Cleanup.Done()
		context.podDeleted = true
	}
}

// CreateAndRunPod Create and run a pod according to the given specifications
func CreateAndRunPod(clientSet *kubernetes.Clientset, creationSpec *PodCreationSpec) error {
	context := &podContext{
		creationSpec:  creationSpec,
		podsClient:    clientSet.Pods("default"),
		serviceClient: clientSet.Services("default"),
	}

	containerName := workflowsv1.GenerateContainerName()

	err := createPod(context, containerName)
	if err != nil {
		return err
	}

	log.Debugf("Created pod %v", context.pod.Name)

	creationSpec.Cleanup.Add(1)
	go func() {
		<-creationSpec.Context.Done()
		cleanupPodIfNecessary(context)
	}()

	printer := &podLogPrinter{
		podsClient:       context.podsClient,
		logPrefix:        creationSpec.LogPrefix,
		variableReceiver: creationSpec.VariableReceiver,
		workflowReceiver: creationSpec.WorkflowReceiver,
	}

	go waitForPod(context, printer)
	return nil
}

func createServicePort(port string) v1.ServicePort {
	protocol := v1.ProtocolTCP

	protocolSeparator := strings.Index(port, "/")
	if protocolSeparator > -1 {
		if "udp" == port[protocolSeparator+1:] {
			protocol = v1.ProtocolUDP
		}

		port = port[:protocolSeparator]
	}

	sourcePort := parseInt(port, 0)
	targetPort := sourcePort

	portSeparator := strings.Index(port, ":")
	if portSeparator > -1 {
		sourcePort = parseInt(port[:portSeparator], 0)
		targetPort = parseInt(port[portSeparator+1:], sourcePort)
	}

	return v1.ServicePort{
		Port: sourcePort,
		TargetPort: intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: targetPort,
		},
		Protocol: protocol,
	}
}

func createPod(context *podContext, containerName string) error {
	creationSpec := context.creationSpec

	mounts, podVolumes := createVolumes(creationSpec.Volumes)
	environment := createEnvironment(creationSpec.Environment)
	readinessProbe := createProbe(creationSpec.Readiness)
	healthProbe := createProbe(creationSpec.Health)

	var labels map[string]string

	if len(creationSpec.ServiceName) > 0 {
		labels := make(map[string]string, 1)
		labels[serviceNameKey] = creationSpec.ServiceName
	}

	pod, err := context.podsClient.Create(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "sbox-",
			Labels:       labels,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            containerName,
					Image:           creationSpec.Image,
					Command:         creationSpec.Command,
					ImagePullPolicy: v1.PullIfNotPresent,
					VolumeMounts:    mounts,
					Env:             environment,
					ReadinessProbe:  readinessProbe,
					LivenessProbe:   healthProbe,
				},
			},
			Volumes:       podVolumes,
			RestartPolicy: v1.RestartPolicyNever,
		},
	})
	if err != nil {
		return err
	}

	context.pod = pod

	if len(creationSpec.ServiceName) > 0 {
		log.Debugf("Creating service %v", creationSpec.ServiceName)

		ports := make([]v1.ServicePort, 0, 1)
		for _, port := range creationSpec.Ports {
			ports = append(ports, createServicePort(port))
		}

		service, err := context.serviceClient.Create(&v1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: creationSpec.ServiceName,
			},
			Spec: v1.ServiceSpec{
				Selector: labels,
				Ports:    ports,
			},
		})

		if err != nil {
			return err
		}

		context.service = service
	}

	return nil
}

func waitForPod(context *podContext, logPrinter *podLogPrinter) {
	log.Debugf("Starting watch on pod %v", context.pod.Name)
	podWatch, err := logPrinter.podsClient.Watch(metav1.ListOptions{Watch: true})
	if err != nil {
		fmt.Println(err.Error())
		cleanupPodIfNecessary(context)
		return
	}

	var containerAvailable int32
	var podReady int32

	channel := podWatch.ResultChan()
	for event := range channel {
		eventPod, ok := event.Object.(*v1.Pod)
		if ok && eventPod.Name == context.pod.Name {
			logPrinter.printLogs(eventPod)

			listener := context.creationSpec.Listener
			if listener != nil {
				containerID := getContainerID(&eventPod.Status)
				if len(containerID) > 0 {
					if atomic.CompareAndSwapInt32(&containerAvailable, 0, 1) {
						listener.Container(containerID)
					}
				}

				if isPodReady(eventPod) {
					if atomic.CompareAndSwapInt32(&podReady, 0, 1) {
						listener.Ready()
					}
				}
			}

			if isPodFinished(eventPod) {
				logPrinter.close()

				if listener != nil {
					listener.Done()
				}

				break
			}
		}
	}
}
