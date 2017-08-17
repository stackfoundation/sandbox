package workflows

import (
        "io"
        "os"

        "k8s.io/client-go/tools/clientcmd"
        "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
        corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
        extensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
        extensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
        "k8s.io/client-go/kubernetes"
        metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
        "k8s.io/client-go/rest"
        "k8s.io/apimachinery/pkg/runtime"
        "k8s.io/apimachinery/pkg/runtime/schema"
        "k8s.io/apimachinery/pkg/runtime/serializer"
        "k8s.io/client-go/pkg/api/v1"

        "github.com/stackfoundation/core/pkg/minikube/config"
        "github.com/stackfoundation/core/pkg/minikube/constants"
        "github.com/stackfoundation/core/pkg/util/kubeconfig"
        "k8s.io/client-go/dynamic"
)

func createDynamicClient() (*dynamic.Client, error) {
        restClientConfig, err := createRestClientConfig()
        if err != nil {
                return nil, err
        }

        restClientConfig.ContentConfig.GroupVersion = &schema.GroupVersion{
                Group: WorkflowsGroupName,
                Version: WorkflowsGroupVersion,
        }

        restClientConfig.APIPath = "/apis"

        return dynamic.NewClient(restClientConfig)
}

func createExtensionsClient() (*clientset.Clientset, error) {
        restClientConfig, err := createRestClientConfig()
        if err != nil {
                return nil, err
        }

        return clientset.NewForConfig(restClientConfig)
}

func createKubeClient() (*kubernetes.Clientset, error) {
        restClientConfig, err := createRestClientConfig()
        if err != nil {
                return nil, err
        }

        return kubernetes.NewForConfig(restClientConfig)
}

func createRestClient() (*rest.RESTClient, error) {
        restClientConfig, err := createRestClientConfig()
        if err != nil {
                return nil, err
        }

        restClientConfig.ContentConfig.GroupVersion = &schema.GroupVersion{
                Group: WorkflowsGroupName,
                Version: WorkflowsGroupVersion,
        }
        restClientConfig.APIPath = "/apis"

        schemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)

        scheme := runtime.NewScheme()
        schemeBuilder.AddToScheme(scheme)

        restClientConfig.NegotiatedSerializer =
                serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

        return rest.RESTClientFor(restClientConfig)
}

func createRestClientConfig() (*rest.Config, error) {
        kubeConfig, err := kubeconfig.ReadConfigOrNew(constants.KubeconfigPath)
        if err != nil {
                return nil, err
        }

        configOverrides := &clientcmd.ConfigOverrides{}
        k8sClientConfig := clientcmd.NewNonInteractiveClientConfig(
                *kubeConfig, config.GetMachineName(), configOverrides, nil)
        return k8sClientConfig.ClientConfig()
}

func createPod(pods corev1.PodInterface, name string, image string, command []string) (*v1.Pod, error) {
        return pods.Create(&v1.Pod{
                ObjectMeta: metav1.ObjectMeta{
                        GenerateName: "sbox-",
                },
                Spec: v1.PodSpec{
                        Containers: []v1.Container{
                                {
                                        Name:    name,
                                        Image:   image,
                                        Command: command,
                                        ImagePullPolicy: v1.PullIfNotPresent,
                                },
                        },
                        RestartPolicy: v1.RestartPolicyNever,
                },
        })
}

func createWorkflowResourceIfRequired(customResourceDefintions extensionsclient.CustomResourceDefinitionInterface) error {
        _, err := customResourceDefintions.Get("workflows.stack.foundation", metav1.GetOptions{})
        if err != nil {
                _, err = customResourceDefintions.Create(&extensions.CustomResourceDefinition{
                        ObjectMeta: metav1.ObjectMeta{
                                Name: "workflows.stack.foundation",
                        },
                        Spec: extensions.CustomResourceDefinitionSpec{
                                Group: WorkflowsGroupName,
                                Scope: extensions.NamespaceScoped,
                                Version: WorkflowsGroupVersion,
                                Names: extensions.CustomResourceDefinitionNames{
                                        Plural: WorkflowsPluralName,
                                        Singular: "workflow",
                                        Kind: WorkflowsKind,
                                        ShortNames: []string{
                                                "wf",
                                                "wflow",
                                        },
                                },
                        },
                })

                if err != nil {
                        return err
                }
        }

        return nil
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

func isPodFinished(pod *v1.Pod) bool {
        return pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed
}

func openAndPrintLogStream(pods corev1.PodInterface, podName string, follow bool) (io.ReadCloser, error) {
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

func printLogsUntilPodFinished(pods corev1.PodInterface, pod *v1.Pod) {
        podWatch, err := pods.Watch(metav1.ListOptions{Watch: true})
        if err != nil {
                //MaybeReportErrorAndExit(err)
        }

        var logStream io.ReadCloser = nil

        channel := podWatch.ResultChan()
        for event := range channel {
                podStatus, ok := event.Object.(*v1.Pod)
                if ok && podStatus.Name == pod.Name {
                        if logStream == nil {
                                logStream, err = printLogsIfAvailable(pods, podStatus)
                                if err != nil {
                                        panic(err.Error())
                                }
                        }

                        if isPodFinished(podStatus) {
                                if logStream != nil {
                                        logStream.Close()
                                }
                                break
                        }
                }
        }
}

func printLogsIfAvailable(pods corev1.PodInterface, pod *v1.Pod) (io.ReadCloser, error) {
        if isContainerRunning(&pod.Status) {
                return openAndPrintLogStream(pods, pod.Name, true)
        } else if isContainerTerminated(&pod.Status) {
                return openAndPrintLogStream(pods, pod.Name, false)
        }

        return nil, nil
}
