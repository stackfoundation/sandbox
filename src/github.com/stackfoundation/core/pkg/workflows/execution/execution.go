package execution

import (
	"context"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/stackfoundation/core/pkg/log"
	"github.com/stackfoundation/core/pkg/workflows/kube"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

type ExecutionContext interface {
	Cancel()
	CleanupWaitGroup() *sync.WaitGroup
	Context() context.Context
	PodsClient() *kubernetes.Clientset
	WorkflowsClient() *rest.RESTClient
}

type stepExecutionContext struct {
	context      ExecutionContext
	workflow     *v1.Workflow
	stepSelector []int
	step         *v1.WorkflowStep
}

func executeInitialStep(context ExecutionContext, workflow *v1.Workflow) error {
	initial := make([]int, 0, 2)
	initial = append(initial, 0)
	workflow.Spec.State.Step = initial
	workflow.Spec.State.Status = v1.StatusStepFinished

	return executeStep(context, workflow)
}

func isCompoundStepComplete(step *v1.WorkflowStep) bool {
	for _, step := range step.Steps {
		if step.Type == v1.StepService {
			if step.State.Status != v1.StatusStepReady &&
				step.State.Status != v1.StatusStepDone {
				return false
			}
		} else if step.Type == v1.StepParallel {
			if step.State.Status != v1.StatusStepDone {
				return false
			}
		}
	}

	return true
}

func executeStep(context ExecutionContext, workflow *v1.Workflow) error {
	workflowSpec := &workflow.Spec
	stepSelector := workflowSpec.State.Step
	step := v1.SelectStep(workflowSpec, stepSelector)

	stepContext := &stepExecutionContext{
		context:      context,
		workflow:     workflow,
		stepSelector: stepSelector,
		step:         step,
	}

	status := workflow.Spec.State.Status

	if status == v1.StatusCompoundStepFinished {
		if isCompoundStepComplete(step) {
			err := buildStepImageAndTransitionNext(stepContext)
			if err != nil {
				return err
			}
		}
	} else if status == v1.StatusStepFinished {
		err := buildStepImageAndTransitionNext(stepContext)
		if err != nil {
			return err
		}
	} else if status == v1.StatusStepImageBuilt {
		err := runStepAndTransitionNext(stepContext)
		if err != nil {
			return err
		}
	} else if status == v1.StatusFinished {
		completeWorkflow(context, workflow)
	}

	return nil
}

func completeWorkflow(context ExecutionContext, workflow *v1.Workflow) {
	kube.DeleteWorkflow(context.WorkflowsClient(), workflow)
	context.Cancel()
}

// ExecuteNextStep Execute the next step of the workflow
func ExecuteNextStep(context ExecutionContext, workflow *v1.Workflow) error {
	log.Debugf(`Executing next step in workflow "%v"`, workflow.ObjectMeta.Name)

	if len(workflow.Spec.State.Status) < 1 {
		return executeInitialStep(context, workflow)
	}

	return executeStep(context, workflow)
}

func buildStepImageAndTransitionNext(context *stepExecutionContext) error {
	err := buildStepImage(context)
	if err != nil {
		return err
	}

	return kube.UpdateWorkflow(context.context.WorkflowsClient(), context.workflow, func(w *v1.Workflow) {
		w.Spec.State.Status = v1.StatusStepImageBuilt
	})
}

func runStepAndTransitionNext(context *stepExecutionContext) error {
	err := runStep(context)
	if err != nil {
		return err
	}

	return kube.UpdateWorkflow(context.context.WorkflowsClient(), context.workflow, func(w *v1.Workflow) {
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
