package execution

import (
	"fmt"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/magiconair/properties"
	"github.com/stackfoundation/core/pkg/log"
	"github.com/stackfoundation/core/pkg/workflows/kube"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

type statusUpdater struct {
	context      ExecutionContext
	workflow     *v1.Workflow
	stepSelector []int
	ready        chan bool
}

func (updater *statusUpdater) updateStepStatus(stepStatus v1.StepStatus) {
	kube.UpdateWorkflow(updater.context.WorkflowsClient(), updater.workflow, func(workflow *v1.Workflow) {
		step := v1.SelectStep(&workflow.Spec, updater.stepSelector)
		step.State.Status = stepStatus
	})
}

func (updater *statusUpdater) Ready() {
	if updater.ready != nil {
		fmt.Println("Service is ready, continuing")
		close(updater.ready)
	} else {
		updater.updateStepStatus(v1.StatusStepReady)
	}
}

func (updater *statusUpdater) Done() {
	updater.updateStepStatus(v1.StatusStepDone)
}

var driveLetterReplacement = regexp.MustCompile("^([a-zA-Z])\\:")

func lowercaseDriveLetter(text []byte) []byte {
	lowercase := strings.ToLower(string(text))
	return []byte("/" + lowercase[:len(lowercase)-1])
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

func normalizeVolumePaths(projectRoot string, volumes []v1.Volume) []v1.Volume {
	if len(volumes) > 0 {
		modified := volumes[:0]
		for _, volume := range volumes {
			if len(volume.HostPath) > 0 {
				absoluteHostPath := path.Join(filepath.ToSlash(projectRoot), volume.HostPath)
				absoluteHostPath = string(driveLetterReplacement.ReplaceAllFunc(
					[]byte(absoluteHostPath),
					lowercaseDriveLetter))

				modified = append(modified, v1.Volume{
					HostPath:  absoluteHostPath,
					MountPath: volume.MountPath,
				})
			}
		}

		return modified
	}

	return volumes
}

func runStep(stepContext *stepExecutionContext) error {
	context := stepContext.context
	workflowSpec := &stepContext.workflow.Spec
	step := stepContext.step
	stepSelector := stepContext.stepSelector
	stepName := v1.StepName(step, stepSelector)

	fmt.Println("Running " + stepName + ":")

	var command []string
	if len(step.State.GeneratedScript) > 0 {
		command = []string{"/bin/sh", "/" + step.State.GeneratedScript}
	}

	step.Volumes = normalizeVolumePaths(workflowSpec.State.ProjectRoot, step.Volumes)

	var updater kube.PodStatusUpdater
	var ready chan bool

	if step.Type == v1.StepParallel || step.Type == v1.StepService {
		if step.Readiness != nil && !step.Readiness.SkipWait {
			ready = make(chan bool)
		}

		updater = &statusUpdater{
			context:      context,
			workflow:     stepContext.workflow,
			stepSelector: stepSelector,
			ready:        ready,
		}
	}

	err := kube.CreateAndRunPod(
		context.PodsClient(),
		&kube.PodCreationSpec{
			Image:       step.State.GeneratedImage,
			Command:     command,
			Environment: collectStepEnvironment(step.Environment),
			Readiness:   step.Readiness,
			Volumes:     step.Volumes,
			Context:     context.Context(),
			Cleanup:     context.CleanupWaitGroup(),
			Updater:     updater,
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
