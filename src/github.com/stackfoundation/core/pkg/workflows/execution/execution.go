package execution

import (
	"context"
	"fmt"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	"github.com/magiconair/properties"
	"github.com/pborman/uuid"
	"github.com/stackfoundation/core/pkg/log"
	"github.com/stackfoundation/core/pkg/workflows/docker"
	"github.com/stackfoundation/core/pkg/workflows/kube"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

type ExecutionContext interface {
	Context() context.Context
	PodsClient() *kubernetes.Clientset
	WorkflowsClient() *rest.RESTClient
}

type statusUpdater struct {
	context      ExecutionContext
	workflow     *v1.Workflow
	stepSelector []int
}

func (updater *statusUpdater) updateStepStatus(stepStatus string) {
	kube.UpdateWorkflow(updater.context.WorkflowsClient(), updater.workflow, func(workflow *v1.Workflow) {
		step := selectStep(&workflow.Spec, updater.stepSelector)
		step.StepStatus = stepStatus
	})
}

func (updater *statusUpdater) Ready() {
	updater.updateStepStatus(v1.StatusStepReady)
}

func (updater *statusUpdater) Done() {
	updater.updateStepStatus(v1.StatusStepDone)
}

func selectStepAndName(workflowSpec *v1.WorkflowSpec, stepSelector []int) (*v1.WorkflowStep, string) {
	step := selectStep(workflowSpec, stepSelector)
	return step, v1.StepName(step, stepSelector)
}

func collectStepEnvironment(environment []v1.EnvironmentSource) *properties.Properties {
	numSources := len(environment)

	if numSources > 0 {
		props := properties.NewProperties()

		for _, variable := range environment {
			if len(variable.File) > 0 {
				fileProperties, err := properties.LoadFile(variable.File, properties.UTF8)
				if err != nil || fileProperties == nil {
					log.Debugf("Error loading properties from file %v", variable.File)
					continue
				}

				props.Merge(fileProperties)
			} else {
				props.Set(variable.Name, variable.Value)
			}
		}

		return props
	}

	return nil
}

func buildImageForStep(context ExecutionContext, workflow *v1.Workflow) error {
	workflowSpec := &workflow.Spec

	step, stepName := selectStepAndName(workflowSpec, workflowSpec.Status.Step)
	fmt.Println("Building image for " + stepName + ":")

	uuid := uuid.NewUUID()
	step.StepImage = "step:" + uuid.String()

	dockerClient, err := docker.CreateDockerClient()
	if err != nil {
		return err
	}

	err = buildImage(context.Context(), dockerClient, workflowSpec, step)
	if err != nil {
		return err
	}

	log.Debugf(`Image %v was built for step "%v"`, step.StepImage, stepName)

	return kube.UpdateWorkflow(context.WorkflowsClient(), workflow, func(workflow *Workflow) {
		workflow.Spec.Status.Status = StatusStepImageBuilt
	})
}

// ExecuteNextStep Execute the next step of the workflow
func ExecuteNextStep(context ExecutionContext, workflow *v1.Workflow) error {
	log.Debugf(`Executing next step in workflow "%v"`, workflow.ObjectMeta.Name)

	if len(workflow.Spec.Status.Status) > 0 {
		if StatusStepFinished == workflow.Spec.Status.Status {
			return buildImageForStep(workflow)
		} else if StatusStepImageBuilt == workflow.Spec.Status.Status {
			controller.runStepContainer(workflow, workflow.Spec.Status.Step)
			return kube.UpdateWorkflow(controller.workflowClient, workflow, nextStep)
		} else if StatusFinished == workflow.Spec.Status.Status {
			kube.DeleteWorkflow(controller.workflowClient, workflow)
			controller.cancel()
		}
	} else if len(workflow.Spec.Steps) > 0 {
		initial := make([]int, 0, 2)
		workflow.Spec.Status.Step = append(initial, 0)

		return controller.buildImageForStep(workflow)
	}

	return nil
}

func incrementStepSelector(workflowSpec *v1.WorkflowSpec) {
	stepSelector := workflowSpec.Status.Step
	numSegments := len(stepSelector)

	steps := workflowSpec.Steps
	stepCounts := make([]int, numSegments)
	for i, segment := range stepSelector {
		stepCounts[i] = len(steps)
		steps = steps[segment].Steps
	}

	segment := numSegments - 1
	for segment >= 0 {
		stepSelector[segment]++

		if stepSelector[segment] < stepCounts[segment] {
			break
		}

		stepSelector = stepSelector[:segment]
		segment--
	}

	if len(stepSelector) > 0 {
		step := selectStep(workflowSpec, stepSelector)
		if step.Type == StepCompound {
			stepSelector = append(stepSelector, 0)
		}
	}

	workflowSpec.Status.Step = stepSelector
}

func nextStep(workflow *v1.Workflow) {
	workflow.Spec.Status.Status = StatusStepFinished
	incrementStepSelector(&workflow.Spec)
	if len(workflow.Spec.Status.Step) == 0 {
		workflow.Spec.Status.Status = StatusFinished
	}
}

func selectStep(workflowSpec *v1.WorkflowSpec, stepSelector []int) *v1.WorkflowStep {
	var step *WorkflowStep

	for _, segment := range stepSelector {
		if step == nil {
			step = &workflowSpec.Steps[segment]
		} else {
			step = &step.Steps[segment]
		}
	}

	return step
}

func runStepContainer(context ExecutionContext, workflow *v1.Workflow, stepSelector []int) error {
	workflowSpec := &workflow.Spec

	step, stepName := selectStepAndName(workflowSpec, stepSelector)
	fmt.Println("Running " + stepName + ":")

	var command []string
	if len(step.StepScript) > 0 {
		command = []string{"/bin/sh", "/" + step.StepScript}
	}

	err := createAndRunPod(
		context.PodsClient(),
		&podCreationSpec{
			context: context.Context(),
			cleanup: &controller.cleanup,
			updater: &statusUpdater{
				context:      context,
				workflow:     workflow,
				stepSelector: stepSelector,
			},
			projectRoot: workflowSpec.ProjectRoot,
			image:       step.StepImage,
			command:     command,
			volumes:     step.Volumes,
			health:      step.Health,
			environment: collectStepEnvironment(step.Environment),
		})
	if err != nil {
		return err
	}

	return nil
}
