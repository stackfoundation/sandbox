package sync

import (
	"fmt"

	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/kube"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

type statusUpdater struct {
	execution    execution.Execution
	workflow     *v1.Workflow
	stepSelector []int
	ready        chan bool
}

func (updater *statusUpdater) updateStepStatus(stepStatus v1.StepStatus) {
	updater.execution.UpdateWorkflow(updater.workflow, func(workflow *v1.Workflow) {
		step := v1.SelectStep(&workflow.Spec, updater.stepSelector)
		step.State.Status = stepStatus
	})
}

func (updater *statusUpdater) Ready() {
	if updater.ready != nil {
		fmt.Println("Service is ready, continuing")
		close(updater.ready)
	} else {
		updater.updateStepStatus(v1.StatusStepReady)
	}
}

func (updater *statusUpdater) Done() {
	updater.updateStepStatus(v1.StatusStepDone)
}

func runStep(c *execution.StepExecutionContext) error {
	workflowSpec := &c.Workflow.Spec
	step := c.Step
	stepSelector := c.StepSelector
	stepName := v1.StepName(step, stepSelector)

	fmt.Println("Running " + stepName + ":")

	var command []string
	if len(step.State.GeneratedScript) > 0 {
		command = []string{"/bin/sh", "/" + step.State.GeneratedScript}
	}

	step.Volumes = normalizeVolumePaths(workflowSpec.State.ProjectRoot, step.Volumes)

	var updater kube.PodStatusUpdater
	var ready chan bool

	if step.Type == v1.StepParallel || step.Type == v1.StepService ||
		(len(step.Type) == 0 && step.Readiness != nil) {
		if step.Readiness != nil && !step.Readiness.SkipWait {
			ready = make(chan bool)
		}

		updater = &statusUpdater{
			execution:    c.Execution,
			workflow:     c.Workflow,
			stepSelector: stepSelector,
			ready:        ready,
		}
	}

	err := c.Execution.RunStep(&execution.RunStepSpec{
		Image:       step.State.GeneratedImage,
		Command:     command,
		Environment: collectStepEnvironment(step.Environment),
		Readiness:   step.Readiness,
		Volumes:     step.Volumes,
		Updater:     updater,
	})
	if err != nil {
		return err
	}

	if ready != nil {
		fmt.Println("Waiting for service to be ready")
		<-ready
	}

	return nil
}

func runStepAndTransitionNext(c *execution.StepExecutionContext) error {
	err := runStep(c)
	if err != nil {
		return err
	}

	return c.Execution.UpdateWorkflow(c.Workflow, func(w *v1.Workflow) {
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
	})
}
