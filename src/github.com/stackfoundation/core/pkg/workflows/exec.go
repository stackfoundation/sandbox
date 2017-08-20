package workflows

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

	return createAndRunPod(clientSet, image, command)
}
