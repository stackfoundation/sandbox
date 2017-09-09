package cmd

import (
	"github.com/stackfoundation/core/pkg/workflows/docker"
	"github.com/stackfoundation/core/pkg/workflows/kube"
)

// Execute Execute a command in a container created with an image
func Execute(image string, command []string) error {
	dockerClient, err := docker.CreateDockerClient()
	if err != nil {
		return err
	}

	docker.PullImageIfNecessary(dockerClient, image)

	clientSet, err := kube.CreateKubeClient()
	if err != nil {
		return err
	}

	return kube.CreateAndRunPod(clientSet, &kube.PodCreationSpec{
		Image:   image,
		Command: command,
	})
}
