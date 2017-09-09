package kube

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/magiconair/properties"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"

	log "github.com/stackfoundation/core/pkg/log"
	workflowsv1 "github.com/stackfoundation/core/pkg/workflows/v1"
)

type PodStatusUpdater interface {
	Ready()
	Done()
}

// PodCreationSpec Specification for creating a pod
type PodCreationSpec struct {
	Context     context.Context
	Cleanup     *sync.WaitGroup
	Updater     PodStatusUpdater
	Image       string
	Command     []string
	Volumes     []workflowsv1.Volume
	Readiness   *workflowsv1.HealthCheck
	Environment *properties.Properties
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
		defer creationSpec.Cleanup.Done()

		<-creationSpec.Context.Done()

		if !podDeleted {
			log.Debugf("Deleting pod %v", pod.Name)
			pods.Delete(pod.Name, &metav1.DeleteOptions{})
			podDeleted = true
		}
	}()

	if creationSpec.Updater != nil {
		go waitForPod(creationSpec.Updater, pods, pod)
	} else {
		waitForPod(creationSpec.Updater, pods, pod)
		creationSpec.Cleanup.Done()
		podDeleted = true
	}

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

func waitForPod(updater PodStatusUpdater, pods corev1.PodInterface, pod *v1.Pod) {
	podWatch, err := pods.Watch(metav1.ListOptions{Watch: true})
	if err != nil {
		//MaybeReportErrorAndExit(err)
	}

	var printer podLogPrinter
	var podReady int32

	channel := podWatch.ResultChan()
	for event := range channel {
		eventPod, ok := event.Object.(*v1.Pod)
		if ok && eventPod.Name == pod.Name {
			printer.printLogs(pods, eventPod)

			if updater != nil {
				if isPodReady(eventPod) {
					if atomic.CompareAndSwapInt32(&podReady, 0, 1) {
						updater.Ready()
					}
				}
			}

			if isPodFinished(eventPod) {
				printer.close()

				if updater != nil {
					updater.Done()
				}

				break
			}
		}
	}
}
