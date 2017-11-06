package preparation

import (
	"github.com/stackfoundation/core/pkg/workflows/execution/context"
	"github.com/stackfoundation/core/pkg/workflows/v1"
	"github.com/stackfoundation/log"
)

type stepExpansionError struct {
	err  error
	step string
}

func (e *stepExpansionError) Error() string {
	return "Error replacing variable placeholders in step " + e.step + ":\n" + e.err.Error()
}

func expandStep(c *context.Context, step *v1.WorkflowStep, stepSelector []int) error {
	stepName := step.StepName(stepSelector)
	log.Debugf("Expanding variable placeholders in step %v", stepName)

	err := v1.ExpandStep(step, c.Workflow.Spec.State.Variables)
	if err != nil {
		if step.IgnoreMissing == nil {
			if !c.Workflow.Spec.IgnoreMissing {
				return &stepExpansionError{err: err, step: stepName}
			}
		} else if !*step.IgnoreMissing {
			return &stepExpansionError{err: err, step: stepName}
		}

		log.Debugf("Ignoring missing variable placeholders in step %v:\n%v", stepName, err)
	}

	return nil
}

func shouldIgnoreValidation(step *v1.WorkflowStep, stepSelector []int, workflow *v1.Workflow, err error) error {
	if step.IgnoreValidation == nil {
		if !workflow.Spec.IgnoreValidation {
			return err
		}
	} else if !*step.IgnoreValidation {
		return err
	}

	log.Debugf("Ignoring validation errors in step %v:\n%v", step.StepName(stepSelector), err)
	return nil
}

// PrepareStepIfNecessary Prepare the step for execution
func PrepareStepIfNecessary(c *context.Context, step *v1.WorkflowStep, stepSelector []int) error {
	if step != nil && !step.State.Prepared {
		err := expandStep(c, step, stepSelector)
		if err != nil {
			return err
		}

		err = v1.ValidateStep(step, stepSelector)
		if err != nil {
			err = shouldIgnoreValidation(step, stepSelector, c.Workflow, err)
			if err != nil {
				return err
			}
		}

		step.State.Prepared = true
	}

	return nil
}
