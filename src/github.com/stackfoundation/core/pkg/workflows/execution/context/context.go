package context

import (
	"context"

	"github.com/stackfoundation/core/pkg/workflows/v1"
)

// NewExecutionContext Create a new execution context
func NewExecutionContext(context context.Context, w *v1.Workflow, c *v1.Change) *ExecutionContext {
	nextStepSelector := w.IncrementStepSelector(c.StepSelector)
	return &ExecutionContext{
		Context:          context,
		Workflow:         w,
		Change:           c,
		Step:             w.Select(c.StepSelector),
		StepSelector:     c.StepSelector,
		NextStep:         w.Select(nextStepSelector),
		NextStepSelector: nextStepSelector,
	}
}

func (c *ExecutionContext) isAtCompoundStepBoundary() bool {
	currentSegmentCount := len(c.StepSelector)
	nextSegmentCount := len(c.NextStepSelector)

	return nextSegmentCount < currentSegmentCount
}

func (c *ExecutionContext) isAtWorkflowBoundary() bool {
	nextSegmentCount := len(c.NextStepSelector)
	return nextSegmentCount == 0
}

// CanProceedToNextStep Can we move on to the next step in context?
func (c *ExecutionContext) CanProceedToNextStep() bool {
	if (c.Change.Type == v1.StepReady && c.Step.IsServiceWithWait()) ||
		(c.Change.Type == v1.StepStarted && (c.Step == nil || c.Step.IsAsync())) ||
		(c.Change.Type == v1.StepDone && !c.Step.IsAsync()) ||
		c.Change.Type == v1.WorkflowWaitDone {
		return true
	}

	return false
}

// IsCompoundStepComplete Is the current step in context a compound step that's complete?
func (c *ExecutionContext) IsCompoundStepComplete() bool {
	if c.isAtCompoundStepBoundary() {
		parent := c.Workflow.Parent(c.StepSelector)
		if parent != nil {
			return parent.IsCompoundStepComplete()
		}
	}

	return false
}

// IsGeneratedWorkflowReadyToRun Is the current generated workflow step in context ready to run?
func (c *ExecutionContext) IsGeneratedWorkflowReadyToRun() bool {
	return c.Change.Type == v1.StepDone && c.Step.IsGenerator()
}

// IsStepReadyToRun Is the current step in context ready to run?
func (c *ExecutionContext) IsStepReadyToRun() bool {
	return c.Change.Type == v1.StepImageBuilt
}

// IsWorkflowComplete Is the workflow in context complete?
func (c *ExecutionContext) IsWorkflowComplete() bool {
	return c.isAtWorkflowBoundary() && c.CanProceedToNextStep()
}
