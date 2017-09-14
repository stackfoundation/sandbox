package sync

import (
	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

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
}

func initialTransition(c *execution.Context, w *v1.Workflow) {
	w.Spec.State.Properties = collectVariables(w.Spec.Variables)

	change := v1.NewChange([]int{})
	change.Type = v1.StepStarted

	w.AppendChange(change)
}

type stepDoneTransition struct {
	generatedWorkfow string
	variables        []v1.VariableSource
}

func (t *stepDoneTransition) transition(c *execution.Context, w *v1.Workflow) {
	w.Spec.State.Properties.Merge(collectVariables(t.variables))

	step := w.Select(c.StepSelector)
	step.State.Done = true
	if len(step.Generator) > 0 {
		step.State.GeneratedWorkflow = t.generatedWorkfow
	}

	change := handleChangeAndAppend(c, w, c.StepSelector)
	change.Type = v1.StepDone
}

func stepReadyTransition(c *execution.Context, w *v1.Workflow) {
	change := handleChangeAndAppend(c, w, c.StepSelector)

	step := w.Select(c.StepSelector)
	step.State.Ready = true

	change.Type = v1.StepReady
}

func stepStartedTransition(c *execution.Context, w *v1.Workflow) {
	change := handleChangeAndAppend(c, w, c.StepSelector)
	change.Type = v1.StepStarted
}

func workflowWaitTransition(c *execution.Context, w *v1.Workflow) {
	change := handleChangeAndAppend(c, w, c.StepSelector)
	change.Type = v1.WorkflowWait
}
