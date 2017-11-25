package context

import (
	"context"
	"sync"

	"github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
)

// NewWorkflowContext Create a new workflow execution context
func NewWorkflowContext(context context.Context, cancel func(), cleanup *sync.WaitGroup, workflow *v1.Workflow) *WorkflowContext {
	return &WorkflowContext{
		Context:  context,
		Cancel:   cancel,
		Cleanup:  cleanup,
		Workflow: workflow,
	}
}

// NewStepContext Create a new step execution context
func NewStepContext(wc *WorkflowContext, c *v1.Change) *StepContext {
	nextStepSelector := wc.Workflow.IncrementStepSelector(c.StepSelector)
	return &StepContext{
		WorkflowContext:  wc,
		Change:           c,
		Step:             wc.Workflow.Select(c.StepSelector),
		StepSelector:     c.StepSelector,
		NextStep:         wc.Workflow.Select(nextStepSelector),
		NextStepSelector: nextStepSelector,
	}
}

func (sc *StepContext) isAtCompoundStepBoundary() bool {
	currentSegmentCount := len(sc.StepSelector)
	nextSegmentCount := len(sc.NextStepSelector)

	return nextSegmentCount < currentSegmentCount
}

func (sc *StepContext) isAtWorkflowBoundary() bool {
	nextSegmentCount := len(sc.NextStepSelector)
	return nextSegmentCount == 0
}

// CanProceedToNextStep Can we move on to the next step in context?
func (sc *StepContext) CanProceedToNextStep() bool {
	if (sc.Change.Type == v1.StepReady && sc.Step.IsServiceWithWait()) ||
		(sc.Change.Type == v1.StepStarted && (sc.Step == nil || sc.Step.IsAsync())) ||
		(sc.Change.Type == v1.StepDone && !sc.Step.IsAsync()) ||
		sc.Change.Type == v1.WorkflowWaitDone {
		if !sc.isAtWorkflowBoundary() {
			return true
		}
	}

	return false
}

// IsCompoundStepComplete Is the current step in context a compound step that's complete?
func (sc *StepContext) IsCompoundStepComplete() bool {
	if sc.isAtCompoundStepBoundary() {
		parent := sc.WorkflowContext.Workflow.Parent(sc.StepSelector)
		if parent != nil {
			return parent.IsCompoundStepComplete()
		}
	}

	return false
}

// IsGeneratedWorkflowReadyToRun Is the current generated workflow step in context ready to run?
func (sc *StepContext) IsGeneratedWorkflowReadyToRun() bool {
	return sc.Change.Type == v1.StepDone && sc.Step.IsGenerator()
}

// IsStepReadyToRun Is the current step in context ready to run?
func (sc *StepContext) IsStepReadyToRun() bool {
	return sc.Change.Type == v1.StepImageBuilt
}

// IsWorkflowComplete Is the workflow in context complete?
func (sc *StepContext) IsWorkflowComplete() bool {
	return sc.isAtWorkflowBoundary() && sc.CanProceedToNextStep() && sc.Step.Service == nil
}
