package workflows

import "github.com/pborman/uuid"

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

        pods := clientSet.Pods("default")

        uuid := uuid.NewUUID()
        podName := "pod-" + uuid.String()

        pod, err := createPod(pods, podName, image, command)
        if err != nil {
                return err
        }

        printLogsUntilPodFinished(pods, pod)
        return nil
}
