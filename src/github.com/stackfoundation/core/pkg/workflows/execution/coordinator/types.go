package coordinator

import (
	"context"
	"sync"

	"github.com/docker/engine-api/client"
	"github.com/stackfoundation/core/pkg/workflows/image"
	"github.com/stackfoundation/core/pkg/workflows/kube"
	"github.com/stackfoundation/core/pkg/workflows/properties"
	"github.com/stackfoundation/core/pkg/workflows/v1"
	"k8s.io/client-go/kubernetes"
)

// Coordinator Coordinates with the various clients needed during workflow execution
type Coordinator interface {
	BuildImage(context context.Context, image string, options *image.BuildOptions) error
	CommitContainer(context context.Context, containerID string, image string) error
	RunStep(context context.Context, spec *RunStepSpec) error
}

type executionCoordinator struct {
	dockerClient *client.Client
	podsClient   *kubernetes.Clientset
}

// RunStepSpec Spec for a step to run
type RunStepSpec struct {
	Cleanup          *sync.WaitGroup
	Command          []string
	Environment      *properties.Properties
	Image            string
	Name             string
	PodListener      kube.PodListener
	Ports            []v1.Port
	Readiness        *v1.HealthCheck
	VariableReceiver func(string, string)
	Volumes          []v1.Volume
	WorkflowReceiver func(string)
}
