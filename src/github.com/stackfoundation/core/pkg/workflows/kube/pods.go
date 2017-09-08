package kube

import (
	"context"
	"sync"

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

type podCreationSpec struct {
	context     context.Context
	cleanup     *sync.WaitGroup
	updater     PodStatusUpdater
	image       string
	projectRoot string
	command     []string
	volumes     []workflowsv1.Volume
	health      *workflowsv1.Health
	environment *properties.Properties
}

func createAndRunPod(clientSet *kubernetes.Clientset, creationSpec *podCreationSpec) error {
	pods := clientSet.Pods("default")

	containerName := workflowsv1.GenerateContainerName()

	pod, err := createPod(pods, containerName, creationSpec)
	if err != nil {
		return err
	}

	creationSpec.cleanup.Add(1)
	var podDeleted bool
	go func() {
		defer creationSpec.cleanup.Done()

		<-creationSpec.context.Done()

		if !podDeleted {
			log.Debugf("Deleting pod %v", pod.Name)
			pods.Delete(pod.Name, &metav1.DeleteOptions{})
			podDeleted = true
		}
	}()

	if creationSpec.updater != nil {
		go waitForCompletion(creationSpec.updater, pods, pod)
	} else {
		waitForCompletion(creationSpec.updater, pods, pod)
		creationSpec.cleanup.Done()
		podDeleted = true
	}

	return nil
}

func createPod(pods corev1.PodInterface, name string, creationSpec *podCreationSpec) (*v1.Pod, error) {
	mounts, podVolumes := createVolumes(creationSpec.projectRoot, creationSpec.volumes)
	environment := createEnvironment(creationSpec.environment)
	readinessProbe := createReadinessProbe(creationSpec.health)

	return pods.Create(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "sbox-",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            name,
					Image:           creationSpec.image,
					Command:         creationSpec.command,
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

func waitForCompletion(updater PodStatusUpdater, pods corev1.PodInterface, pod *v1.Pod) {
	podWatch, err := pods.Watch(metav1.ListOptions{Watch: true})
	if err != nil {
		//MaybeReportErrorAndExit(err)
	}

	var printer podLogPrinter
	var podReady bool

	channel := podWatch.ResultChan()
	for event := range channel {
		eventPod, ok := event.Object.(*v1.Pod)
		if ok && eventPod.Name == pod.Name {
			printer.printLogs(pods, eventPod)

			if updater != nil && !podReady {
				if isPodReady(eventPod) {
					podReady = true
					updater.Ready()
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
