package kube

import (
	"io"
	"os"

	"github.com/stackfoundation/core/pkg/workflows/util"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"
)

type podLogPrinter struct {
	stream io.ReadCloser
}

func openAndPrintPodLogs(
	pods corev1.PodInterface,
	podName string,
	follow bool,
	variableReceiver func(string, string)) (io.ReadCloser, error) {
	logsRequest := pods.GetLogs(podName, &v1.PodLogOptions{Follow: follow})
	logStream, err := logsRequest.Stream()
	if err != nil {
		return nil, err
	}

	if logStream != nil {
		logStream = util.NewDetector(logStream, variableReceiver)
		logStream = util.NewPrefixer(logStream, "["+podName+"] ")

		if follow {
			go func() {
				_, _ = io.Copy(os.Stdout, logStream)
			}()

			return logStream, nil
		}

		defer logStream.Close()
		_, _ = io.Copy(os.Stdout, logStream)
	}

	return nil, nil
}

func (printer *podLogPrinter) close() {
	if printer.stream != nil {
		printer.stream.Close()
	}

	printer.stream = nil
}

func (printer *podLogPrinter) printLogs(
	pods corev1.PodInterface,
	pod *v1.Pod,
	variableReceiver func(string, string)) error {
	var err error

	if printer.stream == nil {
		if isContainerRunning(&pod.Status) {
			printer.stream, err = openAndPrintPodLogs(pods, pod.Name, true, variableReceiver)
		} else if isContainerTerminated(&pod.Status) {
			printer.stream, err = openAndPrintPodLogs(pods, pod.Name, false, variableReceiver)
		}
	}

	return err
}
