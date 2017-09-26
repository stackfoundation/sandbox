package sync

import (
	"fmt"

	"github.com/stackfoundation/core/pkg/workflows/docker"
	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

func commitPreviousStepImage(currentStep *v1.WorkflowStep, e execution.Execution, c *execution.Context) error {
	for i := 0; i < len(c.Workflow.Spec.Steps); i++ {
		step := &c.Workflow.Spec.Steps[i]
		if step.Name == currentStep.Image {
			if !step.State.Prepared || len(step.State.GeneratedContainer) < 1 {
				return &buildError{
					message: "Cannot use image from step \"" + step.Name + "\" because it has not run yet",
				}
			}

			currentStep.State.GeneratedBaseImage = v1.GenerateImageName()

			fmt.Printf("Creating image %v from step \"%v\"\n",
				currentStep.State.GeneratedBaseImage, currentStep.Image)

			return e.CommitContainer(step.State.GeneratedContainer, currentStep.State.GeneratedBaseImage)
		}
	}

	return nil
}

func (e *syncExecution) CommitContainer(containerID string, image string) error {
	return docker.CommitContainer(e.context, e.dockerClient, containerID, image)
}
