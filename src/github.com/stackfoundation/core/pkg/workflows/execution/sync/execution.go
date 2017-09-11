package sync

import (
	"github.com/stackfoundation/core/pkg/log"
	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

func executeInitialStep(e execution.Execution, workflow *v1.Workflow) error {
	initial := make([]int, 0, 2)
	initial = append(initial, 0)
	workflow.Spec.State.Step = initial
	workflow.Spec.State.Status = v1.StatusStepFinished

	return executeStep(e, workflow)
}

func isCompoundStepComplete(step *v1.WorkflowStep) bool {
	for _, step := range step.Steps {
		if step.Type == v1.StepService {
			if step.State.Status != v1.StatusStepReady &&
				step.State.Status != v1.StatusStepDone {
				return false
			}
		} else if step.Type == v1.StepParallel {
			if step.State.Status != v1.StatusStepDone {
				return false
			}
		}
	}

	return true
}

func executeStep(e execution.Execution, workflow *v1.Workflow) error {
	workflowSpec := &workflow.Spec
	stepSelector := workflowSpec.State.Step
	step := v1.SelectStep(workflowSpec, stepSelector)

	stepContext := &execution.StepExecutionContext{
		Execution:    e,
		Workflow:     workflow,
		StepSelector: stepSelector,
		Step:         step,
	}

	status := workflow.Spec.State.Status

	if status == v1.StatusCompoundStepFinished {
		if isCompoundStepComplete(step) {
			err := buildStepImageAndTransitionNext(stepContext)
			if err != nil {
				return err
			}
		}
	} else if status == v1.StatusStepFinished {
		err := buildStepImageAndTransitionNext(stepContext)
		if err != nil {
			return err
		}
	} else if status == v1.StatusStepImageBuilt {
		err := runStepAndTransitionNext(stepContext)
		if err != nil {
			return err
		}
	} else if status == v1.StatusFinished {
		e.Complete()
	}

	return nil
}

// ExecuteNextStep Execute the next step of a workflow execution
func ExecuteNextStep(e execution.Execution, workflow *v1.Workflow) error {
	log.Debugf(`Executing next step in workflow "%v"`, workflow.ObjectMeta.Name)

	if len(workflow.Spec.State.Status) < 1 {
		return executeInitialStep(e, workflow)
	}

	return executeStep(e, workflow)
}
