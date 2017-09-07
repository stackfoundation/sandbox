package execution

import (
	"fmt"

	"github.com/magiconair/properties"
	"github.com/pborman/uuid"
	"github.com/stackfoundation/core/pkg/log"
)

type statusUpdater struct {
	controller   *workflowController
	workflow     *Workflow
	stepSelector []int
}

func (updater *statusUpdater) updateStepStatus(stepStatus string) {
	updater.controller.updateWorkflow(updater.workflow, func(workflow *Workflow) {
		step := selectStep(&workflow.Spec, updater.stepSelector)
		step.StepStatus = stepStatus
	})
}

func (updater *statusUpdater) Ready() {
	updater.updateStepStatus(StatusStepReady)
}

func (updater *statusUpdater) Done() {
	updater.updateStepStatus(StatusStepDone)
}

func collectStepEnvironment(environment []EnvironmentSource) *properties.Properties {
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

func (controller *workflowController) buildImageForStep(workflow *Workflow) error {
	workflowSpec := &workflow.Spec

	step, stepName := workflowStep(workflowSpec, workflowSpec.Status.Step)
	fmt.Println("Building image for " + stepName + ":")

	uuid := uuid.NewUUID()
	step.StepImage = "step:" + uuid.String()

	dockerClient, err := createDockerClient()
	if err != nil {
		return err
	}

	err = buildImage(controller.context, dockerClient, workflowSpec, step)
	if err != nil {
		return err
	}

	log.Debugf(`Image %v was built for step "%v"`, step.StepImage, stepName)

	return controller.updateWorkflow(workflow, func(workflow *Workflow) {
		workflow.Spec.Status.Status = StatusStepImageBuilt
	})
}

func incrementStepSelector(workflowSpec *WorkflowSpec) {
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

func nextStep(workflow *Workflow) {
	workflow.Spec.Status.Status = StatusStepFinished
	incrementStepSelector(&workflow.Spec)
	if len(workflow.Spec.Status.Step) == 0 {
		workflow.Spec.Status.Status = StatusFinished
	}
}
func selectStep(workflowSpec *WorkflowSpec, stepSelector []int) *WorkflowStep {
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

func (controller *workflowController) runStepContainer(workflow *Workflow, stepSelector []int) error {
	workflowSpec := &workflow.Spec

	step, stepName := workflowStep(workflowSpec, stepSelector)
	fmt.Println("Running " + stepName + ":")

	var command []string
	if len(step.StepScript) > 0 {
		command = []string{"/bin/sh", "/" + step.StepScript}
	}

	err := createAndRunPod(
		controller.podsClient,
		&podCreationSpec{
			context: controller.context,
			cleanup: &controller.cleanup,
			updater: &statusUpdater{
				controller:   controller,
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
