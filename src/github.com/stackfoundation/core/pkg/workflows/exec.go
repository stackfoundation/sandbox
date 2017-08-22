package workflows

import (
	"os"
)

// ExecuteCommand Execute a command in a container created with an image
func ExecuteCommand(image string, command []string) error {
	dockerClient, err := createDockerClient()
	if err != nil {
		return err
	}

	pullImageIfNecessary(dockerClient, image)

	clientSet, err := createKubeClient()
	if err != nil {
		return err
	}

	workingDirectory, err := os.Getwd()

	return createAndRunPod(clientSet, &podCreationSpec{
		projectRoot: workingDirectory,
		image:       image,
		command:     command,
	})
}
