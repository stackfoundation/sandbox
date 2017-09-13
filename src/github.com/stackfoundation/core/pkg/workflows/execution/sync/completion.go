package sync

import (
	"fmt"

	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

type podCompletionListener struct {
	execution    execution.Execution
	workflow     *v1.Workflow
	stepSelector []int
	ready        chan bool
	variables    []v1.VariableSource
}

func (listener *podCompletionListener) addVariable(name string, value string) {
	listener.variables = append(listener.variables, v1.VariableSource{
		Name:  name,
		Value: value,
	})
}

func (listener *podCompletionListener) Ready() {
	if listener.ready != nil {
		fmt.Println("Service is ready, continuing")
		close(listener.ready)
	} else {
		transition := stepReadyTransition{stepSelector: listener.stepSelector}
		listener.execution.UpdateWorkflow(listener.workflow, transition.transitionStepReady)
	}
}

func (listener *podCompletionListener) Done() {
	transition := stepDoneTransition{
		stepSelector: listener.stepSelector,
		variables:    listener.variables,
	}

	listener.execution.UpdateWorkflow(listener.workflow, transition.transitionStepDone)
}
