package sync

import (
	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/kube"
)

func runStepAndTransitionNext(e execution.Execution, c *execution.Context) error {
	if c.Step.RequiresBuild() {
		return runPodStepAndTransitionNext(e, c)
	}

	return runWorkflowStepAndTransitionNext(e, c)
}

func (e *syncExecution) RunStep(spec *execution.RunStepSpec) error {
	return kube.CreateAndRunPod(
		e.podsClient,
		&kube.PodCreationSpec{
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
			WorkflowReceiver: spec.WorkflowReceiver,
		})
}
