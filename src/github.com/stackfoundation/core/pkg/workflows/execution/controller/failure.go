package controller

import (
	"github.com/stackfoundation/core/pkg/workflows/v1"
	"github.com/stackfoundation/log"
)

func areFailuresIgnored(step *v1.WorkflowStep, stepSelector []int, workflow *v1.Workflow) bool {
	if step.IgnoreFailure == nil {
		if !workflow.Spec.IgnoreFailure {
			return false
		}
	} else if !*step.IgnoreFailure {
		return false
	}

	return true
}

func shouldIgnoreFailure(step *v1.WorkflowStep, stepSelector []int, workflow *v1.Workflow, err error) error {
	if !areFailuresIgnored(step, stepSelector, workflow) {
		return err
	}

	log.Debugf("Ignoring failure in step %v:\n%v", step.StepName(stepSelector), err)
	return nil
}
