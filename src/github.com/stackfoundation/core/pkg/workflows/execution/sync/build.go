package sync

import (
	"fmt"
	"strings"

	"github.com/stackfoundation/core/pkg/log"
	"github.com/stackfoundation/core/pkg/workflows/docker"
	"github.com/stackfoundation/core/pkg/workflows/execution"
	"github.com/stackfoundation/core/pkg/workflows/image"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

type buildError struct {
	message string
}

func (err *buildError) Error() string {
	return err.message
}

func createBuildOptionsForStepImage(workflowSpec *v1.WorkflowSpec, step *v1.WorkflowStep) *image.BuildOptions {
	if step.HasScript() {
		step.State.GeneratedScript = v1.GenerateScriptName()
	}

	if len(step.Dockerfile) > 0 {
		return &image.BuildOptions{
			ContextDirectory: workflowSpec.State.ProjectRoot,
			DockerfilePath:   step.Dockerfile,
		}
	}

	dockerfileContent := buildDockerfile(step)

	var script string
	if len(step.Script) > 0 {
		script = step.Script
	} else if len(step.Generator) > 0 {
		script = step.Generator
	}

	return &image.BuildOptions{
		ContextDirectory:  workflowSpec.State.ProjectRoot,
		DockerfilePath:    "",
		Dockerignore:      step.Dockerignore,
		ScriptName:        step.State.GeneratedScript,
		DockerfileContent: strings.NewReader(dockerfileContent),
		ScriptContent:     strings.NewReader(script),
	}
}

func buildStepImage(e execution.Execution, c *execution.Context) error {
	step := c.NextStep
	stepName := step.StepName(c.NextStepSelector)

	if step.ImageSource == v1.SourceStep {
		err := commitPreviousStepImage(step, e, c)
		if err != nil {
			return err
		}
	}

	fmt.Println("Building image for step " + stepName + ":")

	step.State.GeneratedImage = v1.GenerateImageName()

	options := createBuildOptionsForStepImage(&c.Workflow.Spec, step)
	err := e.BuildStepImage(step.State.GeneratedImage, options)
	if err != nil {
		return err
	}

	log.Debugf(`Image %v was built for step "%v"`, step.State.GeneratedImage, stepName)

	return nil
}

func buildStepImageAndTransitionNext(e execution.Execution, c *execution.Context) error {
	err := prepareStepIfNecessary(c, c.NextStep, c.NextStepSelector)
	if err != nil {
		return err
	}

	if c.NextStep.RequiresBuild() {
		err := buildStepImage(e, c)
		if err != nil {
			return err
		}
	}

	return e.TransitionNext(c, imageBuiltTransition)
}

func (e *syncExecution) BuildStepImage(image string, options *image.BuildOptions) error {
	return docker.BuildImage(e.context, e.dockerClient, image, options)
}
