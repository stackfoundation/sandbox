package workflows

import (
        "context"
        "errors"
        "path/filepath"
        "net/http"
        "io"
        "os"
        "strings"

        "github.com/docker/engine-api/client"
        "github.com/docker/docker/pkg/jsonmessage"
        "github.com/docker/go-connections/tlsconfig"
        "github.com/docker/engine-api/types"

        "github.com/stackfoundation/core/pkg/minikube/cluster"
        "github.com/stackfoundation/core/pkg/minikube/constants"
        "github.com/stackfoundation/core/pkg/minikube/machine"
        "github.com/stackfoundation/core/pkg/workflows/image"
        "fmt"
        "time"
)

func createDockerHttpClient(hostDockerEnv map[string]string) (*http.Client, error) {
        if dockerCertPath := hostDockerEnv["DOCKER_CERT_PATH"]; dockerCertPath != "" {
                tlsConfigOptions := tlsconfig.Options{
                        CAFile:             filepath.Join(dockerCertPath, "ca.pem"),
                        CertFile:           filepath.Join(dockerCertPath, "cert.pem"),
                        KeyFile:            filepath.Join(dockerCertPath, "key.pem"),
                        InsecureSkipVerify: hostDockerEnv["DOCKER_TLS_VERIFY"] == "",
                }

                tlsClientConfig, err := tlsconfig.Client(tlsConfigOptions)
                if err != nil {
                        return nil, err
                }

                httpClient := &http.Client{
                        Transport: &http.Transport{
                                TLSClientConfig: tlsClientConfig,
                        },
                }

                return httpClient, nil
        }

        return nil, errors.New("Unable to determine Docker configuration")
}

func createDockerClient() (*client.Client, error) {
        hostDockerEnv, err := getHostDockerEnv()
        if err != nil {
                return nil, err
        }

        httpClient, err := createDockerHttpClient(hostDockerEnv)
        if err != nil {
                return nil, err
        }

        host := hostDockerEnv["DOCKER_HOST"]
        return client.NewClient(host, constants.DockerAPIVersion, httpClient, nil)
}

func getHostDockerEnv() (map[string]string, error) {
        machineClient, err := machine.NewAPIClient()
        if err != nil {
                return nil, err
        }
        defer machineClient.Close()

        return cluster.GetHostDockerEnv(machineClient)
}

func buildImage(ctx context.Context, dockerClient *client.Client, workflowSpec *WorkflowSpec, step *WorkflowStep) error {

        var imageStream io.ReadCloser
        var err error
        var dockerfileTarEntry string
        if len(step.Dockerfile) > 0 {
                fmt.Println("Using Dockerfile")
                imageStream, dockerfileTarEntry, err = image.BuildImageStream(workflowSpec.ProjectRoot, step.Dockerfile, nil)
        } else {
                fmt.Println("Building image")
                dockerfileContent := buildDockerfile(step)
                fmt.Println(dockerfileContent)
                imageStream, dockerfileTarEntry, err = image.BuildImageStream(workflowSpec.ProjectRoot, "", strings.NewReader(dockerfileContent))

        }
        if err != nil {
                return err
        }

        defer imageStream.Close()

        buildOptions := types.ImageBuildOptions{
                Dockerfile: dockerfileTarEntry,
        }

        response, err := dockerClient.ImageBuild(context.Background(), imageStream, buildOptions)
        if err != nil {
                panic(err)
        }

        jsonmessage.DisplayJSONMessagesStream(response.Body, os.Stdout, 0, true, nil)
        _, _ = io.Copy(os.Stdout, response.Body)

        time.Sleep(10 * time.Second)
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
