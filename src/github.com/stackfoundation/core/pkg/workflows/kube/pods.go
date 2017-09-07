package kube

import (
	"context"
	"io"
	"os"
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
		go printLogsUntilPodFinished(creationSpec.updater, pods, pod)
	} else {
		printLogsUntilPodFinished(creationSpec.updater, pods, pod)
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

func openAndPrintLogStream(pods corev1.PodInterface, podName string, follow bool) (io.ReadCloser, error) {
	logsRequest := pods.GetLogs(podName, &v1.PodLogOptions{Follow: follow})
	logStream, err := logsRequest.Stream()
	if err != nil {
		return nil, err
	}

	if follow {
		go func() {
			_, _ = io.Copy(os.Stdout, logStream)
		}()
		return logStream, nil
	}

	defer logStream.Close()
	_, _ = io.Copy(os.Stdout, logStream)
	return nil, nil
}

func printLogsUntilPodFinished(updater PodStatusUpdater, pods corev1.PodInterface, pod *v1.Pod) {
	podWatch, err := pods.Watch(metav1.ListOptions{Watch: true})
	if err != nil {
		//MaybeReportErrorAndExit(err)
	}

	var logStream io.ReadCloser
	var podReady bool

	channel := podWatch.ResultChan()
	for event := range channel {
		podStatus, ok := event.Object.(*v1.Pod)
		if ok && podStatus.Name == pod.Name {
			if logStream == nil {
				logStream, err = printLogsIfAvailable(pods, podStatus)
				if err != nil {
					panic(err.Error())
				}
			}

			if updater != nil && !podReady {
				if len(podStatus.Status.Conditions) > 0 {
					for _, condition := range podStatus.Status.Conditions {
						if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
							podReady = true
							updater.Ready()
						}
					}
				}
			}

			if isPodFinished(podStatus) {
				if logStream != nil {
					logStream.Close()
				}

				if updater != nil {
					updater.Done()
				}

				break
			}
		}
	}
}

func printLogsIfAvailable(pods corev1.PodInterface, pod *v1.Pod) (io.ReadCloser, error) {
	if isContainerRunning(&pod.Status) {
		return openAndPrintLogStream(pods, pod.Name, true)
	} else if isContainerTerminated(&pod.Status) {
		return openAndPrintLogStream(pods, pod.Name, false)
	}

	return nil, nil
}
