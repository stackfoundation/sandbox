package kube

import (
	"fmt"
	"strings"
	"sync/atomic"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"

	workflowsv1 "github.com/stackfoundation/core/pkg/workflows/v1"
	log "github.com/stackfoundation/log"
)

func cleanupPodIfNecessary(context *podContext) {
	if !context.podDeleted {
		log.Debugf("Deleting pod %v", context.pod.Name)
		context.podsClient.Delete(context.pod.Name, &metav1.DeleteOptions{})

		if len(context.services) > 0 {
			for _, service := range context.services {
				log.Debugf("Deleting service %v", service.Name)
				context.serviceClient.Delete(service.Name, &metav1.DeleteOptions{})
			}
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

func extractProtocol(port string) (v1.Protocol, string) {
	protocol := v1.ProtocolTCP

	protocolSeparator := strings.Index(port, "/")
	if protocolSeparator > -1 {
		if "udp" == port[protocolSeparator+1:] {
			protocol = v1.ProtocolUDP
		}

		port = port[:protocolSeparator]
	}

	return protocol, port
}

func createServicePort(port workflowsv1.Port) v1.ServicePort {
	protocol := v1.ProtocolTCP

	if strings.ToLower(port.Protocol) == "udp" {
		protocol = v1.ProtocolUDP
	}

	containerPort := parseInt(port.ContainerPort, 0)
	internalPort := parseInt(port.InternalPort, 0)
	if internalPort == 0 {
		internalPort = containerPort
	}

	servicePort := v1.ServicePort{
		Protocol: protocol,
		TargetPort: intstr.IntOrString{
			Type:   intstr.Int,
			IntVal: containerPort,
		},
		Port: internalPort,
	}

	if len(port.ExternalPort) > 0 {
		servicePort.NodePort = parseInt(port.ExternalPort, 0)
	}

	return servicePort
}

func createService(context *podContext, port workflowsv1.Port, labels map[string]string) (*v1.Service, error) {
	servicePort := createServicePort(port)

	serviceType := v1.ServiceTypeClusterIP
	if len(port.ExternalPort) > 0 {
		serviceType = v1.ServiceTypeNodePort
	}

	serviceName := port.Name
	if len(serviceName) == 0 {
		serviceName = workflowsv1.GenerateServiceName()
	}

	log.Debugf("Creating service %v", serviceName)
	return context.serviceClient.Create(&v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
		},
		Spec: v1.ServiceSpec{
			Type:     serviceType,
			Selector: labels,
			Ports:    []v1.ServicePort{servicePort},
		},
	})
}

func createServices(context *podContext, labels map[string]string) ([]*v1.Service, error) {
	creationSpec := context.creationSpec

	services := make([]*v1.Service, 0, len(creationSpec.Ports))
	for _, port := range creationSpec.Ports {
		service, err := createService(context, port, labels)
		if err != nil {
			return nil, err
		}

		services = append(services, service)
	}

	return services, nil
}

func createPod(context *podContext, containerName string) error {
	creationSpec := context.creationSpec

	mounts, podVolumes := createVolumes(creationSpec.Volumes)
	environment := createEnvironment(creationSpec.Environment)
	readinessProbe := createProbe(creationSpec.Readiness)
	healthProbe := createProbe(creationSpec.Health)

	var labels map[string]string

	if len(creationSpec.Ports) > 0 {
		labels = make(map[string]string, 1)
		labels[serviceNameKey] = workflowsv1.GenerateServiceAssociation()
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

	if len(creationSpec.Ports) > 0 {
		services, err := createServices(context, labels)
		if err != nil {
			return err
		}

		context.services = services
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

				if isPullFail(eventPod) {
					if listener != nil {
						listener.Done(true)
					}

					break
				}
			}

			if isPodFinished(eventPod) {
				failed := eventPod.Status.Phase == v1.PodFailed
				logPrinter.close()

				if listener != nil {
					listener.Done(failed)
				}

				break
			}
		}
	}
}
