package docker

import (
	"context"
	"io"
	"os"

	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"

	"github.com/stackfoundation/core/pkg/workflows/image"
)

// CommitContainer Commit a container as an image
func CommitContainer(ctx context.Context, dockerClient *client.Client, containerName string, reference string) error {
	_, err := dockerClient.ContainerCommit(ctx, containerName, types.ContainerCommitOptions{Reference: reference})
	if err != nil {
		return err
	}

	return nil
}

// BuildImage Build an image with the specified name & options
func BuildImage(ctx context.Context, dockerClient *client.Client, imageName string, options *image.BuildOptions) error {
	var imageStream io.ReadCloser
	var err error
	var dockerfileTarEntry string

	imageStream, dockerfileTarEntry, err = image.BuildImageStream(options)
	if err != nil {
		return err
	}

	defer imageStream.Close()

	buildOptions := types.ImageBuildOptions{
		Dockerfile: dockerfileTarEntry,
		Tags:       []string{imageName},
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

// PullImageIfNecessary Pull the specified image if it doesn't already exist
func PullImageIfNecessary(dockerClient *client.Client, image string) error {
	ctx := context.Background()
	_, _, err := dockerClient.ImageInspectWithRaw(ctx, image, false)
	if err != nil {
		pullImage(ctx, dockerClient, image)
	}

	return nil
}
