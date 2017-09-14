package sync

import (
	"fmt"

	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

type podCompletionListener struct {
	execution         execution.Execution
	context           *execution.Context
	generatedWorkflow string
	variables         []v1.VariableSource
}

func (listener *podCompletionListener) addVariable(name string, value string) {
	listener.variables = append(listener.variables, v1.VariableSource{
		Name:  name,
		Value: value,
	})
}

func (listener *podCompletionListener) addGeneratedWorkflow(content string) {
	listener.generatedWorkflow = content
}

func (listener *podCompletionListener) Ready() {
	listener.execution.TransitionNext(listener.context, stepReadyTransition)
}

func (listener *podCompletionListener) Done() {
	transition := stepDoneTransition{
		variables:        listener.variables,
		generatedWorkfow: listener.generatedWorkflow,
	}

	listener.execution.TransitionNext(listener.context, transition.transition)
}

func runPodStepAndTransitionNext(e execution.Execution, c *execution.Context) error {
	step := c.Step
	stepName := step.StepName(c.Change.StepSelector)

	fmt.Println("Running step " + stepName + ":")

	var command []string
	if len(step.State.GeneratedScript) > 0 {
		command = []string{"/bin/sh", "/" + step.State.GeneratedScript}
	}

	step.Volumes = normalizeVolumePaths(c.Workflow.Spec.State.ProjectRoot, step.Volumes)

	completionListener := &podCompletionListener{
		execution: e,
		context:   c,
	}

	environment := collectVariables(step.Environment)
	environment.ResolveFrom(c.Workflow.Spec.State.Properties)

	err := e.RunStep(&execution.RunStepSpec{
		Command:          command,
		Environment:      environment,
		Image:            step.State.GeneratedImage,
		Name:             stepName,
		PodListener:      completionListener,
		Readiness:        step.Readiness,
		VariableReceiver: completionListener.addVariable,
		Volumes:          step.Volumes,
		WorkflowReceiver: completionListener.addGeneratedWorkflow,
	})
	if err != nil {
		return err
	}

	return e.TransitionNext(c, stepStartedTransition)
}
