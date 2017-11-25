package kube

import (
	"io"
	"os"

	"github.com/stackfoundation/sandbox/core/pkg/workflows/processors"

	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"
)

type podLogPrinter struct {
	podsClient       corev1.PodInterface
	logPrefix        string
	stream           io.ReadCloser
	variableReceiver func(string, string)
	workflowReceiver func(string)
}

func (printer *podLogPrinter) addLogProcessors(stream io.ReadCloser) io.ReadCloser {
	if printer.variableReceiver != nil {
		stream = processors.NewVariableDetector(stream, printer.variableReceiver)
	}

	if printer.workflowReceiver != nil {
		stream = processors.NewWorkflowDetector(stream, printer.workflowReceiver)
	}

	if len(printer.logPrefix) > 0 {
		stream = processors.NewPrefixer(stream, "\x1b[30;1m["+printer.logPrefix+"]\x1b[0m ")
	}

	return stream
}

func (printer *podLogPrinter) openAndPrintPodLogs(pod *v1.Pod, follow bool) (io.ReadCloser, error) {
	logsRequest := printer.podsClient.GetLogs(pod.Name, &v1.PodLogOptions{Follow: follow})
	logStream, err := logsRequest.Stream()
	if err != nil {
		return nil, err
	}

	if logStream != nil {
		logStream = printer.addLogProcessors(logStream)

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

func (printer *podLogPrinter) printLogs(pod *v1.Pod) error {
	var err error

	if printer.stream == nil {
		if isContainerRunning(&pod.Status) {
			printer.stream, err = printer.openAndPrintPodLogs(pod, true)
		} else if isContainerTerminated(&pod.Status) {
			printer.stream, err = printer.openAndPrintPodLogs(pod, false)
		}
	}

	return err
}
