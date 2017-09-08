package docker

import (
	"context"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/pborman/uuid"

	"github.com/stackfoundation/core/pkg/workflows/image"
	"github.com/stackfoundation/core/pkg/workflows/v1"
)

func buildImageStream(workflowSpec *v1.WorkflowSpec, step *v1.WorkflowStep) (io.ReadCloser, string, error) {
	uuid := uuid.NewUUID()

	if len(step.Script) > 0 {
		step.StepScript = "script-" + uuid.String()[:8] + ".sh"
	}

	if len(step.Dockerfile) > 0 {
		return image.BuildImageStream(&image.BuildOptions{
			ContextDirectory: workflowSpec.ProjectRoot,
			DockerfilePath:   step.Dockerfile,
		})
	}

	dockerfileContent := buildDockerfile(step)
	return image.BuildImageStream(&image.BuildOptions{
		ContextDirectory:  workflowSpec.ProjectRoot,
		DockerfilePath:    "",
		ScriptName:        step.StepScript,
		DockerfileContent: strings.NewReader(dockerfileContent),
		ScriptContent:     strings.NewReader(step.Script),
	})
}

func buildImage(ctx context.Context, dockerClient *client.Client, workflowSpec *v1.WorkflowSpec, step *v1.WorkflowStep) error {
	var imageStream io.ReadCloser
	var err error
	var dockerfileTarEntry string

	imageStream, dockerfileTarEntry, err = buildImageStream(workflowSpec, step)
	if err != nil {
		return err
	}

	defer imageStream.Close()

	buildOptions := types.ImageBuildOptions{
		Dockerfile: dockerfileTarEntry,
		Tags:       []string{step.StepImage},
	}

	response, err := dockerClient.ImageBuild(ctx, imageStream, buildOptions)
	if err != nil {
		panic(err)
	}

	jsonmessage.DisplayJSONMessagesStream(response.Body, os.Stdout, 0, true, nil)
	_, _ = io.Copy(os.Stdout, response.Body)

	return nil
}

func pullImage(ctx context.Context, dockerClient *client.Client, image string) error {
	pullOptions := types.ImagePullOptions{All: false}
	pullProgress, err := dockerClient.ImagePull(ctx, image, pullOptions)
	defer pullProgress.Close()

	if err != nil {
		return err
	}

	jsonmessage.DisplayJSONMessagesStream(pullProgress, os.Stdout, 0, true, nil)
	_, _ = io.Copy(os.Stdout, pullProgress)

	return nil
}

func pullImageIfNecessary(dockerClient *client.Client, image string) error {
	ctx := context.Background()
	_, _, err := dockerClient.ImageInspectWithRaw(ctx, image, false)
	if err != nil {
		pullImage(ctx, dockerClient, image)
	}

	return nil
}
