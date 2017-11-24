package controller

import (
	"github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
	"github.com/stackfoundation/sandbox/log"
)

func areFailuresIgnored(workflow *v1.Workflow, step *v1.WorkflowStep, stepSelector []int) bool {
	if step.IgnoreFailure() == nil {
		if !workflow.Spec.IgnoreFailure {
			return false
		}
	} else if !*step.IgnoreFailure() {
		return false
	}

	return true
}

func shouldIgnoreFailure(workflow *v1.Workflow, step *v1.WorkflowStep, stepSelector []int, err error) error {
	if !areFailuresIgnored(workflow, step, stepSelector) {
		return err
	}

	log.Debugf("Ignoring failure in step %v:\n%v", step.StepName(stepSelector), err)
	return nil
}
