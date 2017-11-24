package run

import (
	"fmt"

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

func (l *podCompletionListener) Done(failed bool, message string) {
	result := &Result{
		Container: l.generatedContainer,
		Message:   message,
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
	if !step.Cached() && len(step.State.GeneratedScript) > 0 {
		fmt.Println("Running step " + stepName + ":")

		command = []string{"/bin/sh", "/" + step.State.GeneratedScript}
	}

	step.SetVolumes(normalizeVolumePaths(sc.WorkflowContext.Workflow.Spec.State.ProjectRoot, step.Volumes()))

	completionListener := &podCompletionListener{
		listener:    l,
		stepContext: sc,
	}

	environment := v1.CollectVariables(step.Environment())
	environment.ResolveFrom(sc.WorkflowContext.Workflow.Spec.State.Variables)

	if len(step.Name()) < 1 {
		stepName = "Step " + stepName
	}

	var ports []v1.Port
	var health *v1.HealthCheck
	var readiness *v1.HealthCheck

	if step.Service != nil {
		ports = step.Service.Ports
		health = step.Service.Health
		readiness = step.Service.Readiness
	}

	err := c.RunStep(
		sc.WorkflowContext.Context,
		&coordinator.RunStepSpec{
			Command:          command,
			Cleanup:          sc.WorkflowContext.Cleanup,
			Environment:      environment,
			Health:           health,
			Image:            step.State.GeneratedImage,
			Name:             stepName,
			PodListener:      completionListener,
			Ports:            ports,
			Readiness:        readiness,
			VariableReceiver: completionListener.addVariable,
			Volumes:          step.Volumes(),
			WorkflowReceiver: completionListener.addGeneratedWorkflow,
		})

	return err
}
