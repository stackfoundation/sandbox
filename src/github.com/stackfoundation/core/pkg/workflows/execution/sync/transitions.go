package sync

import (
	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/v1"
	"github.com/stackfoundation/log"
)

func logChange(c *v1.Change) {
	log.Debugf("Raised %v event for step %v", c.Type, c.StepSelector)
}

func consumeTransition(c *execution.Context, w *v1.Workflow) {
	w.MarkHandled(c.Change)
}

func handleChangeAndAppend(c *execution.Context, w *v1.Workflow, selector []int) *v1.Change {
	w.MarkHandled(c.Change)

	change := v1.NewChange(selector)
	return w.AppendChange(change)
}

func imageBuiltTransition(c *execution.Context, w *v1.Workflow) {
	change := handleChangeAndAppend(c, w, c.NextStepSelector)
	change.Type = v1.StepImageBuilt

	logChange(change)
}

func initialTransition(c *execution.Context, w *v1.Workflow) {
	w.Spec.State.Variables = collectVariables(w.Spec.Variables)

	change := v1.NewChange([]int{})
	change.Type = v1.StepStarted

	w.AppendChange(change)

	logChange(change)
}

type stepDoneTransition struct {
	generatedContainer string
	generatedWorkfow   string
	variables          []v1.VariableSource
}

func (t *stepDoneTransition) transition(c *execution.Context, w *v1.Workflow) {
	w.Spec.State.Variables.Merge(collectVariables(t.variables))

	step := w.Select(c.StepSelector)
	step.State.GeneratedContainer = t.generatedContainer
	step.State.Done = true

	if step.IsGenerator() {
		step.State.GeneratedWorkflow = t.generatedWorkfow
	}

	change := handleChangeAndAppend(c, w, c.StepSelector)
	change.Type = v1.StepDone

	logChange(change)
}

func stepReadyTransition(c *execution.Context, w *v1.Workflow) {
	change := handleChangeAndAppend(c, w, c.StepSelector)

	step := w.Select(c.StepSelector)
	step.State.Ready = true

	change.Type = v1.StepReady

	logChange(change)
}

func stepStartedTransition(c *execution.Context, w *v1.Workflow) {
	change := handleChangeAndAppend(c, w, c.StepSelector)
	change.Type = v1.StepStarted

	logChange(change)
}

func workflowWaitDoneTransition(c *execution.Context, w *v1.Workflow) {
	change := handleChangeAndAppend(c, w, c.StepSelector)
	change.Type = v1.WorkflowWaitDone

	logChange(change)
}

func workflowWaitTransition(c *execution.Context, w *v1.Workflow) {
	change := handleChangeAndAppend(c, w, c.StepSelector)
	change.Type = v1.WorkflowWait

	logChange(change)
}
