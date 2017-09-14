package sync

import (
	"context"
	"os"
	"os/signal"
	"sync"

	"k8s.io/client-go/kubernetes"

	"github.com/docker/engine-api/client"
	"github.com/stackfoundation/core/pkg/log"
	"github.com/stackfoundation/core/pkg/workflows/docker"
	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/kube"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

type syncExecution struct {
	cancel           context.CancelFunc
	cleanupWaitGroup sync.WaitGroup
	change           chan bool
	context          context.Context
	dockerClient     *client.Client
	podsClient       *kubernetes.Clientset
	workflow         *v1.Workflow
}

func (e *syncExecution) ChildExecution(workflow *v1.Workflow) (execution.Execution, error) {
	return NewSyncExecution(workflow)
}

func (e *syncExecution) Complete() error {
	close(e.change)
	e.cancel()
	e.cleanupWaitGroup.Wait()
	return nil
}

// NewSyncExecution Create a new sync execution for a workflow
func NewSyncExecution(workflow *v1.Workflow) (execution.Execution, error) {
	dockerClient, err := docker.CreateDockerClient()
	if err != nil {
		return nil, err
	}

	podsClient, err := kube.CreateKubeClient()
	if err != nil {
		return nil, err
	}

	context, cancel := context.WithCancel(context.Background())

	interruptChannel := make(chan os.Signal, 1)
	signal.Notify(interruptChannel, os.Interrupt)
	go func() {
		for _ = range interruptChannel {
			log.Debugf("An interrupt was requested, stopping controller!")
			cancel()
		}
	}()

	return &syncExecution{
		cancel:       cancel,
		change:       make(chan bool),
		context:      context,
		dockerClient: dockerClient,
		podsClient:   podsClient,
		workflow:     workflow,
	}, nil
}

func (e *syncExecution) Start() {
	Execute(e, e.workflow)
	for _ = range e.change {
		Execute(e, e.workflow)
	}
}

func (e *syncExecution) TransitionNext(context *execution.Context, update func(*execution.Context, *v1.Workflow)) error {
	go func() {
		workflow := context.Workflow
		update(context, workflow)

		e.change <- true
	}()

	return nil
}
