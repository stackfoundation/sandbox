package cmd

import (
        "fmt"
        "os"
        "io"

        "github.com/golang/glog"
        "github.com/spf13/cobra"
        "github.com/stackfoundation/core/pkg/minikube/cluster"
        "github.com/stackfoundation/core/pkg/minikube/config"
        "github.com/stackfoundation/core/pkg/minikube/constants"
        "github.com/stackfoundation/core/pkg/minikube/machine"
        "github.com/stackfoundation/core/pkg/util/kubeconfig"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/client-go/kubernetes"
        corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
        "k8s.io/client-go/pkg/api/v1"
        "k8s.io/client-go/tools/clientcmd"

        "github.com/docker/engine-api/client"
        "github.com/docker/engine-api/types"
        "github.com/docker/go-connections/tlsconfig"
        "github.com/docker/docker/pkg/jsonmessage"
        "path/filepath"
        "net/http"
        "context"
        "errors"
)

var execImage string
var execCommandString string

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

func pullImage(dockerClient *client.Client, image string) error {
        pullOptions := types.ImagePullOptions{All: false}
        pullProgress, err := dockerClient.ImagePull(context.Background(), image, pullOptions)
        defer pullProgress.Close()

        if err != nil {
                return err
        }

        jsonmessage.DisplayJSONMessagesStream(pullProgress, os.Stdout, 0, true, nil)
        _, _ = io.Copy(os.Stdout, pullProgress)

        return nil
}

func createKubeClient() (*kubernetes.Clientset, error) {
        kubeConfig, err := kubeconfig.ReadConfigOrNew(constants.KubeconfigPath)
        if err != nil {
                return nil, err
        }

        configOverrides := &clientcmd.ConfigOverrides{}
        k8sClientConfig := clientcmd.NewNonInteractiveClientConfig(
                *kubeConfig, config.GetMachineName(), configOverrides, nil)
        restClient, err := k8sClientConfig.ClientConfig()

        return kubernetes.NewForConfig(restClient)
}

func createPod(pods corev1.PodInterface, name string, image string, command string) (*v1.Pod, error) {
        return pods.Create(&v1.Pod{
                ObjectMeta: metav1.ObjectMeta{
                        GenerateName: "sbox-",
                },
                Spec: v1.PodSpec{
                        Containers: []v1.Container{
                                v1.Container{
                                        Name:    name,
                                        Image:   image,
                                        Command: []string{command},
                                        ImagePullPolicy: v1.PullIfNotPresent,
                                },
                        },
                        RestartPolicy: v1.RestartPolicyNever,
                },
        })
}

var execCmd = &cobra.Command{
        Use:   "exec",
        Short: "Execute a single command within a Docker container",
        Long:  `Execute a single command within a Docker container.`,
        Run: func(command *cobra.Command, args []string) {
                startKube()

                dockerClient, err := createDockerClient()
                if err != nil {

                }

                pullImage(dockerClient, execImage)

                clientSet, err := createKubeClient()
                if err != nil {
                        glog.Errorln("Error kubeconfig status:", err)
                        MaybeReportErrorAndExit(err)
                }

                pods := clientSet.Pods("default")

                pod, err := createPod(clientSet, "sboxstep", execImage, execCommandString)
                if err != nil {
                        panic(err.Error())
                }

                podWatch, err := pods.Watch(metav1.ListOptions{Watch: true})
                if err != nil {
                        panic(err.Error())
                }

                var logStream io.ReadCloser = nil

                channel := podWatch.ResultChan()
                for event := range channel {
                        podStatus, ok := event.Object.(*v1.Pod)
                        if ok && podStatus.Name == pod.Name {
                                if logStream == nil {
                                        if isContainerRunning(&podStatus.Status) {
                                                logStream, err = getLogs(pods, pod.Name, true)
                                                if err != nil {
                                                        panic(err.Error())
                                                }
                                        } else if isContainerTerminated(&podStatus.Status) {
                                                logStream, err = getLogs(pods, pod.Name, false)
                                                if err != nil {
                                                        panic(err.Error())
                                                }
                                        }
                                }

                                if podStatus.Status.Phase == v1.PodSucceeded || podStatus.Status.Phase == v1.PodFailed {
                                        if logStream != nil {
                                                logStream.Close()
                                        }
                                        break
                                }
                        }
                }
        },
}

func isContainerRunning(status *v1.PodStatus) bool {
        if len(status.ContainerStatuses) > 0 {
                for i := 0; i < len(status.ContainerStatuses); i++ {
                        if status.ContainerStatuses[i].State.Running != nil {
                                return true
                        }
                }
        }

        return false
}

func isContainerTerminated(status *v1.PodStatus) bool {
        if len(status.ContainerStatuses) > 0 {
                for i := 0; i < len(status.ContainerStatuses); i++ {
                        if status.ContainerStatuses[i].State.Terminated != nil {
                                return true
                        }
                }
        }

        return false
}

func getLogs(pods corev1.PodInterface, podName string, follow bool) (io.ReadCloser, error) {
        logsRequest := pods.GetLogs(podName, &v1.PodLogOptions{Follow: follow})
        logStream, err := logsRequest.Stream()
        if err != nil {
                return nil, err
        }

        if (follow) {
                go func() {
                        _, _ = io.Copy(os.Stdout, logStream)
                }()
                return logStream, nil
        } else {
                defer logStream.Close()
                _, _ = io.Copy(os.Stdout, logStream)
                return nil, nil
        }
}

func init() {
        configureKubeStartingCommandFlags(execCmd)
        execCmd.Flags().StringVarP(&execImage, "image", "i", "", "Image to run")
        execCmd.Flags().StringVarP(&execCommandString, "command", "c", "", "Command to run")
        RootCmd.AddCommand(execCmd)
}
