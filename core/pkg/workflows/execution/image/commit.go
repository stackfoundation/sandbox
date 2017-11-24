package image

import (
	"fmt"

	"github.com/stackfoundation/sandbox/core/pkg/workflows/execution/context"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/execution/coordinator"
	"github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
)

func commitStepImage(coordinator coordinator.Coordinator, sc *context.StepContext, stepName string) (string, error) {
	w := sc.WorkflowContext.Workflow
	for i := 0; i < len(w.Spec.Steps); i++ {
		step := &w.Spec.Steps[i]
		if step.Name() == stepName {
			if !step.State.Prepared || len(step.State.GeneratedContainer) < 1 {
				return "", &buildError{
					message: "Cannot use image from step \"" + step.Name() + "\" because it has not run yet",
				}
			}

			generatedImage := v1.GenerateImageName()

			fmt.Printf("Creating image %v from step \"%v\"\n", generatedImage, stepName)
			return generatedImage, coordinator.CommitContainer(sc.WorkflowContext.Context, step.State.GeneratedContainer, generatedImage)
		}
	}

	return "", nil
}

func commitPreviousStepImage(coordinator coordinator.Coordinator, sc *context.StepContext, currentStep *v1.WorkflowStep) error {
	generatedImage, err := commitStepImage(coordinator, sc, currentStep.Step())
	if err != nil {
		return err
	}

	currentStep.State.GeneratedBaseImage = generatedImage
	return nil
}
