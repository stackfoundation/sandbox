package kube

import (
	"strings"

	"k8s.io/client-go/pkg/api/v1"
)

func getContainerID(status *v1.PodStatus) string {
	if len(status.ContainerStatuses) > 0 {
		id := status.ContainerStatuses[0].ContainerID
		if len(id) > 0 {
			schemeIndex := strings.Index(id, "docker://")
			if schemeIndex > -1 {
				return id[schemeIndex+9:]
			}

			return id
		}
	}

	return ""
}

func isContainerRunning(status *v1.PodStatus) bool {
	if len(status.ContainerStatuses) > 0 {
		for i := 0; i < len(status.ContainerStatuses); i++ {
			if status.ContainerStatuses[i].State.Running != nil {
				return true
			}
		}
	}

	return false
}

func isContainerTerminated(status *v1.PodStatus) bool {
	if len(status.ContainerStatuses) > 0 {
		for i := 0; i < len(status.ContainerStatuses); i++ {
			if status.ContainerStatuses[i].State.Terminated != nil {
				return true
			}
		}
	}

	return false
}

func isPodFinished(pod *v1.Pod) bool {
	return pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed
}

func isPodReady(pod *v1.Pod) bool {
	if len(pod.Status.Conditions) > 0 {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == v1.PodReady && condition.Status == v1.ConditionTrue {
				return true
			}
		}
	}

	return false
}

func isPullFail(pod *v1.Pod) (bool, string) {
	if len(pod.Status.ContainerStatuses) > 0 {
		for i := 0; i < len(pod.Status.ContainerStatuses); i++ {
			if pod.Status.ContainerStatuses[i].State.Waiting != nil {
				if pod.Status.ContainerStatuses[i].State.Waiting.Reason == "ErrImagePull" {
					message := pod.Status.ContainerStatuses[i].State.Waiting.Message
					if len(message) > 0 {
						message = message + " - "
					}

					message = message + "Error pulling image"

					return false, message
				}
			}
		}
	}

	return false, ""
}
