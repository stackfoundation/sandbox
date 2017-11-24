package image

import (
	"fmt"
	"strings"

	"github.com/stackfoundation/core/pkg/workflows/execution/context"
	"github.com/stackfoundation/core/pkg/workflows/execution/coordinator"
	"github.com/stackfoundation/core/pkg/workflows/image"
	"github.com/stackfoundation/core/pkg/workflows/v1"
	"github.com/stackfoundation/log"
)

type buildError struct {
	message string
}

func (err *buildError) Error() string {
	return err.message
}

func createBuildOptionsForStepImage(workflowSpec *v1.WorkflowSpec, step *v1.WorkflowStep) *image.BuildOptions {
	scriptContent := step.Script()

	if step.HasScript() {
		if step.Cached() {
			step.State.GeneratedScript = v1.GenerateCachedScriptName(scriptContent)
		} else {
			step.State.GeneratedScript = v1.GenerateScriptName()
		}
	}

	if step.HasDockerfile() {
		return &image.BuildOptions{
			ContextDirectory: workflowSpec.State.ProjectRoot,
			DockerfilePath:   step.Dockerfile(),
		}
	}

	dockerfileContent := buildDockerfile(step)

	var dockerignore string
	var sourceInclude []string
	var sourceExclude []string

	source := step.Source()
	if source != nil {
		dockerignore = source.Dockerignore
		sourceInclude = source.Include
		sourceExclude = source.Exclude
	}

	return &image.BuildOptions{
		ContextDirectory:  workflowSpec.State.ProjectRoot,
		DockerfilePath:    "",
		Dockerignore:      dockerignore,
		SourceIncludes:    sourceInclude,
		SourceExcludes:    sourceExclude,
		ScriptName:        step.State.GeneratedScript,
		DockerfileContent: strings.NewReader(dockerfileContent),
		ScriptContent:     strings.NewReader(scriptContent),
	}
}

// BuildStepImage Build the image for a step
func BuildStepImage(coordinator coordinator.Coordinator, sc *context.StepContext) error {
	step := sc.NextStep
	stepName := step.StepName(sc.NextStepSelector)

	// collectCherryPicks(coordinator, sc, step)

	if step.UsesPreviousStep() {
		err := commitPreviousStepImage(coordinator, sc, step)
		if err != nil {
			return err
		}
	}

	if step.Cached() {
		fmt.Println("Building image and running step " + stepName + ":")
	} else {
		fmt.Println("Building image for step " + stepName + ":")
	}

	step.State.GeneratedImage = v1.GenerateImageName()

	options := createBuildOptionsForStepImage(&sc.WorkflowContext.Workflow.Spec, step)
	err := coordinator.BuildImage(sc.WorkflowContext.Context, step.State.GeneratedImage, options)
	if err != nil {
		return err
	}

	log.Debugf(`Image %v was built for step "%v"`, step.State.GeneratedImage, stepName)

	return nil
}
