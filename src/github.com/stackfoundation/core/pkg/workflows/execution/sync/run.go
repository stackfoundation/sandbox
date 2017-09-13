package sync

import (
	"fmt"

	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/kube"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

func runStep(c *execution.StepExecutionContext) error {
	workflowSpec := &c.Workflow.Spec
	step := c.Step
	stepName := v1.StepName(step, c.StepSelector)

	fmt.Println("Running " + stepName + ":")

	var command []string
	if len(step.State.GeneratedScript) > 0 {
		command = []string{"/bin/sh", "/" + step.State.GeneratedScript}
	}

	step.Volumes = normalizeVolumePaths(workflowSpec.State.ProjectRoot, step.Volumes)

	var completionListener *podCompletionListener
	var ready chan bool

	async := v1.IsAsyncStep(step)
	if async {
		if step.Readiness != nil && !step.Readiness.SkipWait {
			ready = make(chan bool)
		}
	}

	completionListener = &podCompletionListener{
		execution:    c.Execution,
		workflow:     c.Workflow,
		stepSelector: c.StepSelector,
		ready:        ready,
	}

	environment := collectVariables(step.Environment)
	environment.ResolveFrom(workflowSpec.State.Properties)

	err := c.Execution.RunStep(&execution.RunStepSpec{
		Async:            async,
		Command:          command,
		Environment:      environment,
		Image:            step.State.GeneratedImage,
		Name:             stepName,
		PodListener:      completionListener,
		Readiness:        step.Readiness,
		VariableReceiver: completionListener.addVariable,
		Volumes:          step.Volumes,
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

	return c.Execution.UpdateWorkflow(c.Workflow, transitionNextStep)
}

func (e *syncExecution) RunStep(spec *execution.RunStepSpec) error {
	return kube.CreateAndRunPod(
		e.podsClient,
		&kube.PodCreationSpec{
			Async:            spec.Async,
			LogPrefix:        spec.Name,
			Image:            spec.Image,
			Command:          spec.Command,
			Environment:      spec.Environment,
			Readiness:        spec.Readiness,
			Volumes:          spec.Volumes,
			Context:          e.context,
			Cleanup:          &e.cleanupWaitGroup,
			Listener:         spec.PodListener,
			VariableReceiver: spec.VariableReceiver,
			WorkflowReceiver: func(workflow string) {
				fmt.Printf("Workflow: %v", workflow)
			},
		})
}
