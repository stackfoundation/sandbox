package execution

import (
	"github.com/stackfoundation/core/pkg/workflows/kube"
)

func runStepAndTransitionNext(e Execution, c *Context) error {
	err := prepareStepIfNecessary(c, c.Step, c.StepSelector)
	if err != nil {
		return err
	}

	if c.Step.RequiresBuild() {
		return runPodStepAndTransitionNext(e, c)
	}

	return runExternalWorkflowAndTransitionNext(e, c)
}

func (e *syncExecution) RunStep(spec *RunStepSpec) error {
	return kube.CreateAndRunPod(
		e.podsClient,
		&kube.PodCreationSpec{
			LogPrefix:        spec.Name,
			Image:            spec.Image,
			Command:          spec.Command,
			Environment:      spec.Environment,
			Ports:            spec.Ports,
			Readiness:        spec.Readiness,
			Volumes:          spec.Volumes,
			Context:          e.context,
			Cleanup:          &e.cleanupWaitGroup,
			Listener:         spec.PodListener,
			VariableReceiver: spec.VariableReceiver,
			WorkflowReceiver: spec.WorkflowReceiver,
		})
}
