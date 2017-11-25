package kube

import (
	"fmt"
	"sync/atomic"

	log "github.com/stackfoundation/sandbox/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

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

				pullFailed, message := isPullFail(eventPod)
				if pullFailed {
					context.podClosed <- true
					if listener != nil {
						listener.Done(true, message)
					}

					break
				}
			}

			if isPodFinished(eventPod) {
				failed := eventPod.Status.Phase == v1.PodFailed
				message := eventPod.Status.Message + " (" + eventPod.Status.Reason + ")"
				logPrinter.close()
				context.podClosed <- true

				if listener != nil {
					listener.Done(failed, message)
				}

				break
			}
		}
	}

	close(context.podClosed)
}
