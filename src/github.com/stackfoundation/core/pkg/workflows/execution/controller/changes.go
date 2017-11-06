package controller

import (
	executioncontext "github.com/stackfoundation/core/pkg/workflows/execution/context"
	"github.com/stackfoundation/log"
)

func (c *executionController) processChangeAndTransitionNext(sc *executioncontext.StepContext) error {
	c := sc.Change
	log.Debugf("%v event for step %v", c.Type, c.StepSelector)

	switch {
	case sc.IsStepReadyToRun():
		return c.runStepAndTransitionNext(sc)
	case sc.IsGeneratedWorkflowReadyToRun():
		return c.callGeneratedWorkflow(sc)
	case sc.IsWorkflowComplete():
		log.Debugf("Workflow completed")
		sc.WorkflowContext.Cancel()
		return
	case sc.IsCompoundStepComplete():
		return c.buildStepImageAndTransitionNext(sc)
	case sc.CanProceedToNextStep():
		return c.buildStepImageAndTransitionNext(sc)
	default:
	}

	log.Debugf("Performing no-op for %v event for step %v", c.Type, c.StepSelector)
	return c.transitionNext(sc, consumeTransition)
}

func (c *executionController) processNextChange(wc *executioncontext.WorkflowContext) error {
	if len(wc.Workflow.Spec.State.Changes) < 1 {
		return c.transitionNext(executioncontext.NewStepContext(wc, nil), initialTransition)
	}

	u := wc.Workflow.NextUnhandled()
	if u != nil {
		return c.processChangeAndTransitionNext(executioncontext.NewStepContext(wc, u))
	}

	return nil
}
