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

func createBuildOptionsForStepImage(workflowSpec *v1.WorkflowSpec, step *v1.WorkflowStep) *image.BuildOptions {
	if len(step.Script) > 0 {
		step.State.GeneratedScript = v1.GenerateScriptName()
	}

	if len(step.Dockerfile) > 0 {
		return &image.BuildOptions{
			ContextDirectory: workflowSpec.State.ProjectRoot,
			DockerfilePath:   step.Dockerfile,
		}
	}

	dockerfileContent := buildDockerfile(step)
	return &image.BuildOptions{
		ContextDirectory:  workflowSpec.State.ProjectRoot,
		DockerfilePath:    "",
		ScriptName:        step.State.GeneratedScript,
		DockerfileContent: strings.NewReader(dockerfileContent),
		ScriptContent:     strings.NewReader(step.Script),
	}
}

func buildStepImage(c *execution.StepExecutionContext) error {
	workflowSpec := &c.Workflow.Spec

	step := c.Step
	stepName := v1.StepName(step, c.StepSelector)

	fmt.Println("Building image for " + stepName + ":")

	step.State.GeneratedImage = v1.GenerateImageName()

	options := createBuildOptionsForStepImage(workflowSpec, step)
	err := c.Execution.BuildStepImage(step.State.GeneratedImage, options)
	if err != nil {
		return err
	}

	log.Debugf(`Image %v was built for step "%v"`, step.State.GeneratedImage, stepName)

	return nil
}

func buildStepImageAndTransitionNext(c *execution.StepExecutionContext) error {
	err := buildStepImage(c)
	if err != nil {
		return err
	}

	return c.Execution.UpdateWorkflow(c.Workflow, transitionStepImageBuilt)
}

func (e *syncExecution) BuildStepImage(image string, options *image.BuildOptions) error {
	return docker.BuildImage(e.context, e.dockerClient, image, options)
}
