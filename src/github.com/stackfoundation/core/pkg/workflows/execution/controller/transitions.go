package controller

import (
	executioncontext "github.com/stackfoundation/core/pkg/workflows/execution/context"
	"github.com/stackfoundation/core/pkg/workflows/v1"
	"github.com/stackfoundation/log"
)

func logChange(c *v1.Change) {
	log.Debugf("Raised %v event for step %v", c.Type, c.StepSelector)
}

func consumeTransition(sc *executioncontext.StepContext) {
	sc.WorkflowContext.Workflow.MarkHandled(sc.Change)
}

func handleChangeAndAppend(sc *executioncontext.StepContext, w *v1.Workflow, selector []int) *v1.Change {
	w.MarkHandled(sc.Change)

	change := v1.NewChange(selector)
	return w.AppendChange(change)
}

func imageBuiltTransition(sc *executioncontext.StepContext) {
	change := handleChangeAndAppend(sc, sc.WorkflowContext.Workflow, sc.NextStepSelector)
	change.Type = v1.StepImageBuilt

	logChange(change)
}

func initialTransition(sc *executioncontext.StepContext) {
	w := sc.WorkflowContext.Workflow

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

func (t *stepDoneTransition) transition(sc *executioncontext.StepContext) {
	w := sc.WorkflowContext.Workflow
	step := w.Select(sc.StepSelector)
	if !step.State.Done {
		w.Spec.State.Variables.Merge(v1.CollectVariables(t.variables))

		step.State.GeneratedContainer = t.generatedContainer
		step.State.Ready = true
		step.State.Done = true

		if step.IsGenerator() {
			step.State.GeneratedWorkflow = t.generatedWorkfow
		}

		change := handleChangeAndAppend(sc, w, sc.StepSelector)
		change.Type = v1.StepDone

		logChange(change)
	}
}

func stepReadyTransition(sc *executioncontext.StepContext) {
	w := sc.WorkflowContext.Workflow
	step := w.Select(sc.StepSelector)

	if !step.State.Ready {
		change := handleChangeAndAppend(sc, w, sc.StepSelector)

		step.State.Ready = true

		change.Type = v1.StepReady

		logChange(change)
	}
}

func stepStartedTransition(sc *executioncontext.StepContext) {
	change := handleChangeAndAppend(sc, sc.WorkflowContext.Workflow, sc.StepSelector)
	change.Type = v1.StepStarted

	logChange(change)
}

func workflowWaitDoneTransition(sc *executioncontext.StepContext) {
	change := handleChangeAndAppend(sc, sc.WorkflowContext.Workflow, sc.StepSelector)
	change.Type = v1.WorkflowWaitDone

	logChange(change)
}

func workflowWaitTransition(sc *executioncontext.StepContext) {
	change := handleChangeAndAppend(sc, sc.WorkflowContext.Workflow, sc.StepSelector)
	change.Type = v1.WorkflowWait

	logChange(change)
}

func (t *pendingTransition) perform() {
	t.transition(t.context)
}

func (c *executionController) processTransitions(wc *executioncontext.WorkflowContext) {
	for {
		select {
		case transition := <-c.pendingTransitions:
			transition.perform()
		case <-wc.Context.Done():
			return
		default:
			return
		}
	}
}

func (c *executionController) transitionNext(
	sc *executioncontext.StepContext,
	transition func(*executioncontext.StepContext)) error {
	go func() {
		c.pendingTransitions <- pendingTransition{
			context:    sc,
			transition: transition,
		}
	}()

	return nil
}
