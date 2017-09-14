package kube

import (
	"context"
	"sync"
	"sync/atomic"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"

	log "github.com/stackfoundation/core/pkg/log"
	"github.com/stackfoundation/core/pkg/workflows/properties"
	workflowsv1 "github.com/stackfoundation/core/pkg/workflows/v1"
)

// PodListener Listener which listens for pod events
type PodListener interface {
	Ready()
	Done()
}

// PodCreationSpec Specification for creating a pod
type PodCreationSpec struct {
	Cleanup          *sync.WaitGroup
	Command          []string
	Context          context.Context
	Environment      *properties.Properties
	Image            string
	LogPrefix        string
	Readiness        *workflowsv1.HealthCheck
	Listener         PodListener
	VariableReceiver func(string, string)
	Volumes          []workflowsv1.Volume
	WorkflowReceiver func(string)
}

// CreateAndRunPod Create and run a pod according to the given specifications
func CreateAndRunPod(clientSet *kubernetes.Clientset, creationSpec *PodCreationSpec) error {
	pods := clientSet.Pods("default")

	containerName := workflowsv1.GenerateContainerName()

	pod, err := createPod(pods, containerName, creationSpec)
	if err != nil {
		return err
	}

	creationSpec.Cleanup.Add(1)
	var podDeleted bool
	go func() {
		<-creationSpec.Context.Done()

		if !podDeleted {
			log.Debugf("Deleting pod %v", pod.Name)
			pods.Delete(pod.Name, &metav1.DeleteOptions{})
			creationSpec.Cleanup.Done()
			podDeleted = true
		}
	}()

	printer := &podLogPrinter{
		podsClient:       pods,
		logPrefix:        creationSpec.LogPrefix,
		variableReceiver: creationSpec.VariableReceiver,
		workflowReceiver: creationSpec.WorkflowReceiver,
	}

	go waitForPod(pod, printer, creationSpec.Listener)
	return nil
}

func createPod(pods corev1.PodInterface, name string, creationSpec *PodCreationSpec) (*v1.Pod, error) {
	mounts, podVolumes := createVolumes(creationSpec.Volumes)
	environment := createEnvironment(creationSpec.Environment)
	readinessProbe := createReadinessProbe(creationSpec.Readiness)

	return pods.Create(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "sbox-",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            name,
					Image:           creationSpec.Image,
					Command:         creationSpec.Command,
					ImagePullPolicy: v1.PullIfNotPresent,
					VolumeMounts:    mounts,
					Env:             environment,
					ReadinessProbe:  readinessProbe,
				},
			},
			Volumes:       podVolumes,
			RestartPolicy: v1.RestartPolicyNever,
		},
	})
}

func waitForPod(pod *v1.Pod, logPrinter *podLogPrinter, listener PodListener) {
	podWatch, err := logPrinter.podsClient.Watch(metav1.ListOptions{Watch: true})
	if err != nil {
		//MaybeReportErrorAndExit(err)
	}

	var podReady int32

	channel := podWatch.ResultChan()
	for event := range channel {
		eventPod, ok := event.Object.(*v1.Pod)
		if ok && eventPod.Name == pod.Name {
			logPrinter.printLogs(eventPod)

			if listener != nil {
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
