package sync

import (
	"context"
	"sync"
	"sync/atomic"

	"k8s.io/client-go/kubernetes"

	"github.com/docker/engine-api/client"
	"github.com/stackfoundation/core/pkg/workflows/docker"
	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/image"
	"github.com/stackfoundation/core/pkg/workflows/kube"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

type syncExecution struct {
	cleanupWaitGroup sync.WaitGroup
	complete         int32
	context          context.Context
	dockerClient     *client.Client
	podsClient       *kubernetes.Clientset
	workflow         *v1.Workflow
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

	return &syncExecution{
		context:      context.Background(),
		dockerClient: dockerClient,
		podsClient:   podsClient,
		workflow:     workflow,
	}, nil
}

func (e *syncExecution) BuildStepImage(image string, options *image.BuildOptions) error {
	return docker.BuildImage(e.context, e.dockerClient, image, options)
}

func (e *syncExecution) Complete() {
	atomic.CompareAndSwapInt32(&e.complete, 0, 1)
}

func (e *syncExecution) RunStep(spec *execution.RunStepSpec) error {
	return kube.CreateAndRunPod(
		e.podsClient,
		&kube.PodCreationSpec{
			Image:       spec.Image,
			Command:     spec.Command,
			Environment: spec.Environment,
			Readiness:   spec.Readiness,
			Volumes:     spec.Volumes,
			Context:     e.context,
			Cleanup:     &e.cleanupWaitGroup,
			Updater:     spec.Updater,
		})
}

func (e *syncExecution) Start() {
	for atomic.LoadInt32(&e.complete) == 0 {
		ExecuteNextStep(e, e.workflow)
	}
}

func (e *syncExecution) UpdateWorkflow(workflow *v1.Workflow, update func(*v1.Workflow)) error {
	update(workflow)
	return nil
}

func (e *syncExecution) Workflow() *v1.Workflow {
	return e.workflow
}
