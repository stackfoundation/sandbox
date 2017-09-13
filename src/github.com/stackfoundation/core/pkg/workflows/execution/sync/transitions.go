package sync

import "github.com/stackfoundation/core/pkg/workflows/v1"

type stepReadyTransition struct {
	stepSelector []int
}

func (transition *stepReadyTransition) transitionStepReady(workflow *v1.Workflow) {
	step := v1.SelectStep(&workflow.Spec, transition.stepSelector)
	step.State.Status = v1.StatusStepReady
}

type stepDoneTransition struct {
	stepSelector []int
	variables    []v1.VariableSource
}

func (transition *stepDoneTransition) transitionStepDone(workflow *v1.Workflow) {
	step := v1.SelectStep(&workflow.Spec, transition.stepSelector)

	workflow.Spec.State.Properties.Merge(collectVariables(transition.variables))
	step.State.Status = v1.StatusStepDone
}

func transitionStepImageBuilt(w *v1.Workflow) {
	w.Spec.State.Status = v1.StatusStepImageBuilt
}

func transitionNextStep(w *v1.Workflow) {
	previousSegmentCount := len(w.Spec.State.Step)

	newSelector := v1.IncrementStepSelector(&w.Spec, w.Spec.State.Step)
	newSegmentCount := len(newSelector)

	w.Spec.State.Step = newSelector

	if newSegmentCount < previousSegmentCount {
		if newSegmentCount == 0 {
			w.Spec.State.Status = v1.StatusFinished
		} else {
			w.Spec.State.Status = v1.StatusCompoundStepFinished
		}
	} else {
		w.Spec.State.Status = v1.StatusStepFinished
	}
}
