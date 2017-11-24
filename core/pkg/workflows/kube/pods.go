package kube

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"

	workflowsv1 "github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
	log "github.com/stackfoundation/sandbox/log"
)

func cleanupPodIfNecessary(context *podContext) {
	log.Debugf("Deleting pod %v", context.pod.Name)
	context.podsClient.Delete(context.pod.Name, &metav1.DeleteOptions{})

	if len(context.services) > 0 {
		for _, service := range context.services {
			log.Debugf("Deleting service %v", service.Name)
			context.serviceClient.Delete(service.Name, &metav1.DeleteOptions{})
		}
	}

	context.creationSpec.Cleanup.Done()
}

// CreateAndRunPod Create and run a pod according to the given specifications
func CreateAndRunPod(clientSet *kubernetes.Clientset, creationSpec *PodCreationSpec) error {
	context := &podContext{
		creationSpec:  creationSpec,
		podsClient:    clientSet.Pods("default"),
		serviceClient: clientSet.Services("default"),
		podClosed:     make(chan bool, 2),
	}

	containerName := workflowsv1.GenerateContainerName()

	creationSpec.Cleanup.Add(1)
	go func() {
		<-creationSpec.Context.Done()
		cleanupPodIfNecessary(context)
	}()

	err := createPod(context, containerName)
	if err != nil {
		return err
	}

	log.Debugf("Created pod %v", context.pod.Name)

	printer := &podLogPrinter{
		podsClient:       context.podsClient,
		logPrefix:        creationSpec.LogPrefix,
		variableReceiver: creationSpec.VariableReceiver,
		workflowReceiver: creationSpec.WorkflowReceiver,
	}

	go waitForPod(context, printer)
	return nil
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
