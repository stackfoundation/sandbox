package run

import (
	"fmt"
	"strconv"

	"github.com/stackfoundation/core/pkg/workflows/execution/context"
	"github.com/stackfoundation/core/pkg/workflows/execution/coordinator"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

func (l *podCompletionListener) addVariable(name string, value string) {
	l.variables = append(l.variables, v1.VariableSource{
		Name:  name,
		Value: value,
	})
}

func (l *podCompletionListener) addGeneratedWorkflow(content string) {
	l.generatedWorkflow = content
}

func (l *podCompletionListener) Container(containerID string) {
	l.generatedContainer = containerID
}

func (l *podCompletionListener) Ready() {
	l.listener.Ready(l.stepContext)
}

func (l *podCompletionListener) Done(failed bool) {
	result := &Result{
		Container: l.generatedContainer,
		Workflow:  l.generatedWorkflow,
		Variables: l.variables,
	}

	if failed {
		l.listener.Failed(l.stepContext, result)
		return
	}

	l.listener.Done(l.stepContext, result)
}

// RunPodStep Run a pod-based step
func RunPodStep(c coordinator.Coordinator, sc *context.StepContext, l Listener) error {
	step := sc.Step
	stepName := step.StepName(sc.Change.StepSelector)

	var command []string
	cache, _ := strconv.ParseBool(step.Cache)
	if !cache && len(step.State.GeneratedScript) > 0 {
		fmt.Println("Running step " + stepName + ":")

		command = []string{"/bin/sh", "/" + step.State.GeneratedScript}
	}

	step.Volumes = normalizeVolumePaths(sc.WorkflowContext.Workflow.Spec.State.ProjectRoot, step.Volumes)

	completionListener := &podCompletionListener{
		listener:    l,
		stepContext: sc,
	}

	environment := collectVariables(step.Environment)
	environment.ResolveFrom(sc.WorkflowContext.Workflow.Spec.State.Variables)

	if len(step.Name) < 1 {
		stepName = "Step " + stepName
	}

	err := c.RunStep(
		sc.WorkflowContext.Context,
		&coordinator.RunStepSpec{
			Command:          command,
			Cleanup:          sc.WorkflowContext.Cleanup,
			Environment:      environment,
			Image:            step.State.GeneratedImage,
			Name:             stepName,
			PodListener:      completionListener,
			Ports:            step.Ports,
			Readiness:        step.Readiness,
			VariableReceiver: completionListener.addVariable,
			Volumes:          step.Volumes,
			WorkflowReceiver: completionListener.addGeneratedWorkflow,
		})

	return err
}
