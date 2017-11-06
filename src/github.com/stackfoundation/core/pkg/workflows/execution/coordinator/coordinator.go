package coordinator

import (
	"context"

	"github.com/stackfoundation/core/pkg/workflows/docker"
	"github.com/stackfoundation/core/pkg/workflows/image"
	"github.com/stackfoundation/core/pkg/workflows/kube"
)

// NewCoordinator Create a new coordinator which uses the specified context
func NewCoordinator(context context.Context) (Coordinator, error) {
	dockerClient, err := docker.CreateDockerClient()
	if err != nil {
		return nil, err
	}

	podsClient, err := kube.CreateKubeClient()
	if err != nil {
		return nil, err
	}

	return &executionCoordinator{
		dockerClient: dockerClient,
		podsClient:   podsClient,
	}, nil
}

// BuildImage Build an image with the specified options
func (c *executionCoordinator) BuildImage(context context.Context, image string, options *image.BuildOptions) error {
	return docker.BuildImage(context, c.dockerClient, image, options)
}

// CommitContainer Commit the current state of the specified container as a new image
func (c *executionCoordinator) CommitContainer(context context.Context, containerID string, image string) error {
	return docker.CommitContainer(context, c.dockerClient, containerID, image)
}

func (c *executionCoordinator) RunStep(context context.Context, spec *RunStepSpec) error {
	return kube.CreateAndRunPod(
		c.podsClient,
		&kube.PodCreationSpec{
			LogPrefix:        spec.Name,
			Image:            spec.Image,
			Command:          spec.Command,
			Environment:      spec.Environment,
			Ports:            spec.Ports,
			Readiness:        spec.Readiness,
			Volumes:          spec.Volumes,
			Context:          context,
			Cleanup:          spec.Cleanup,
			Listener:         spec.PodListener,
			VariableReceiver: spec.VariableReceiver,
			WorkflowReceiver: spec.WorkflowReceiver,
		})
}
