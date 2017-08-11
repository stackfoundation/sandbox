package cmd

import (
        "os"
        "io"

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

        apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
        apiextensionsv1beta1client "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
        apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
        "k8s.io/client-go/rest"
        "k8s.io/client-go/dynamic"
        "k8s.io/apimachinery/pkg/runtime/schema"
        "fmt"
        //"io/ioutil"
        //"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
        //"gopkg.in/yaml.v2"

        "k8s.io/client-go/tools/cache"
        "k8s.io/apimachinery/pkg/fields"
        "k8s.io/apimachinery/pkg/conversion"
        "time"
        "k8s.io/apimachinery/pkg/runtime/serializer"
        "k8s.io/apimachinery/pkg/runtime"
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

const GroupName = "stack.foundation"

var SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: "v1"}

func addKnownTypes(scheme *runtime.Scheme) error {
        scheme.AddKnownTypes(SchemeGroupVersion,
                &Workflow{},
                &WorkflowList{},
        )
        metav1.AddToGroupVersion(scheme, SchemeGroupVersion)
        return nil
}

func createRestClient() (*rest.RESTClient, error) {
        restClientConfig, err := createRestClientConfig()
        if err != nil {
                return nil, err
        }

        restClientConfig.ContentConfig.GroupVersion = &schema.GroupVersion{
                Group: "stack.foundation",
                Version: "v1",
        }
        restClientConfig.APIPath = "/apis"

        schemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)

        scheme := runtime.NewScheme()
        schemeBuilder.AddToScheme(scheme)

        restClientConfig.NegotiatedSerializer =
                serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

        return rest.RESTClientFor(restClientConfig)
}

func createDynamicClient() (*dynamic.Client, error) {
        restClientConfig, err := createRestClientConfig()
        if err != nil {
                return nil, err
        }

        restClientConfig.ContentConfig.GroupVersion = &schema.GroupVersion{
                Group: "stack.foundation",
                Version: "v1",
        }
        restClientConfig.APIPath = "/apis"

        return dynamic.NewClient(restClientConfig)
}

func createExtensionsClient() (*apiextensionsclient.Clientset, error) {
        restClientConfig, err := createRestClientConfig()
        if err != nil {
                return nil, err
        }

        return apiextensionsclient.NewForConfig(restClientConfig)
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

func createKubeClient() (*kubernetes.Clientset, error) {
        restClientConfig, err := createRestClientConfig()
        if err != nil {
                return nil, err
        }

        return kubernetes.NewForConfig(restClientConfig)
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

type Workflow struct {
        metav1.TypeMeta   `json:",inline"`
        metav1.ObjectMeta `json:"metadata"`
        Spec map[string]interface{} `json:"spec"`
}

type WorkflowList struct {
        metav1.TypeMeta `json:",inline"`
        metav1.ListMeta `json:"metadata"`
        Items []Workflow `json:"items"`
}

type WorkflowController struct {
        cloner *conversion.Cloner
}

func (c *WorkflowController) Run(ctx context.Context) error {
        fmt.Print("Watch workflow objects\n")

        _, err := c.watchWorkflows(ctx)
        if err != nil {
                fmt.Printf("Failed to register watch for workflow resource: %v\n", err)
                return err
        }

        <-ctx.Done()
        return ctx.Err()
}

func (c *WorkflowController) watchWorkflows(ctx context.Context) (cache.Controller, error) {
        client, err := createRestClient()
        if err != nil {
                panic(err)
        }

        source := cache.NewListWatchFromClient(
                client, "workflows", v1.NamespaceDefault, fields.Everything())

        _, controller := cache.NewInformer(
                source,
                &Workflow{},
                0,
                cache.ResourceEventHandlerFuncs{
                        AddFunc:    c.onAdd,
                        UpdateFunc: c.onUpdate,
                        DeleteFunc: c.onDelete,
                })

        go controller.Run(ctx.Done())
        return controller, nil
}

func (c *WorkflowController) onAdd(obj interface{}) {
        example := obj.(*Workflow)
        fmt.Printf("[CONTROLLER] OnAdd %s\n", example.Spec)

        // NEVER modify objects from the store. It's a read-only, local cache.
        copyObj, err := c.cloner.DeepCopy(example)
        if err != nil {
                fmt.Printf("ERROR creating a deep copy of example object: %v\n", err)
                return
        }

        exampleCopy := copyObj.(*Workflow)
        //exampleCopy.Status = crv1.ExampleStatus{
        //        State:   crv1.ExampleStateProcessed,
        //        Message: "Successfully processed by controller",
        //}
        //
        //err = c.ExampleClient.Put().
        //        Name(example.ObjectMeta.Name).
        //        Namespace(example.ObjectMeta.Namespace).
        //        Resource(crv1.ExampleResourcePlural).
        //        Body(exampleCopy).
        //        Do().
        //        Error()

        if err != nil {
                fmt.Printf("ERROR updating status: %v\n", err)
        } else {
                fmt.Printf("UPDATED status: %#v\n", exampleCopy)
        }
}

func (c *WorkflowController) onUpdate(oldObj, newObj interface{}) {
        oldExample := oldObj.(*Workflow)
        newExample := newObj.(*Workflow)
        fmt.Printf("[CONTROLLER] OnUpdate oldObj: %s\n", oldExample.Spec)
        fmt.Printf("[CONTROLLER] OnUpdate newObj: %s\n", newExample.Spec)
}

func (c *WorkflowController) onDelete(obj interface{}) {
        example := obj.(*Workflow)
        fmt.Printf("[CONTROLLER] OnDelete %s\n", example.Spec)
}

var execCmd = &cobra.Command{
        Use:   "exec",
        Short: "Execute a single command within a Docker container",
        Long:  `Execute a single command within a Docker container.`,
        Run: func(command *cobra.Command, args []string) {
                startKube()

                dockerClient, err := createDockerClient()
                if err != nil {
                        MaybeReportErrorAndExit(err)
                }

                pullImage(dockerClient, execImage)

                clientSet, err := createExtensionsClient()
                if err != nil {
                        MaybeReportErrorAndExit(err)
                }

                //clientSet, err := createKubeClient()
                //if err != nil {
                //MaybeReportErrorAndExit(err)
                //}

                createWorkflowResourceIfRequired(clientSet.CustomResourceDefinitions())

                controller := WorkflowController{
                        cloner: conversion.NewCloner(),
                }

                ctx, cancelFunc := context.WithCancel(context.Background())
                defer cancelFunc()
                go controller.Run(ctx)

                //client, err := createDynamicClient()
                //if err != nil {
                //        panic(err)
                //}

                //get, err := client.Resource(&metav1.APIResource{Name: "workflows", Namespaced: true}, v1.NamespaceDefault).
                //        Get("workflow1", metav1.GetOptions{})
                //List(metav1.ListOptions{})

                //if err != nil {
                //        fmt.Println(err)
                //}
                //
                //fmt.Println(get.Object)

                //content, err := ioutil.ReadFile("test.yml")
                //if err != nil {
                //        panic(err)
                //}
                //
                //var workflowObject workflow
                //err = yaml.Unmarshal(content, &workflowObject)
                //if err != nil {
                //        panic(err)
                //}
                //
                //unstructuredWorkflow := make(map[string]interface{})
                //unstructuredWorkflow["apiVersion"] = "stack.foundation/v1"
                //unstructuredWorkflow["kind"] = "Workflow"
                //unstructuredWorkflow["spec"] = workflowObject.Spec
                //unstructuredWorkflow["metadata"] = workflowObject.Metadata
                //
                //data := unstructured.Unstructured{
                //        Object: unstructuredWorkflow,
                //}
                //
                //_, err = client.Resource(&metav1.APIResource{Name: "workflows", Namespaced: true}, v1.NamespaceDefault).
                //        Create(&data)
                //
                //if err != nil {
                //        panic(err)
                //}

                time.Sleep(5000 * time.Millisecond)
                fmt.Println("Done")

                //pods := clientSet.Pods("default")
                //
                //pod, err := createPod(pods, "sboxstep", execImage, execCommandString)
                //if err != nil {
                //        MaybeReportErrorAndExit(err)
                //}
                //
                //printLogsUntilPodFinished(pods, pod)
        },
}

func createWorkflowResourceIfRequired(customResourceDefintions apiextensionsv1beta1client.CustomResourceDefinitionInterface) {
        _, err := customResourceDefintions.Get("workflows.stack.foundation", metav1.GetOptions{})
        if err != nil {
                _, err = customResourceDefintions.Create(&apiextensionsv1beta1.CustomResourceDefinition{
                        ObjectMeta: metav1.ObjectMeta{
                                Name: "workflows.stack.foundation",
                        },
                        Spec: apiextensionsv1beta1.CustomResourceDefinitionSpec{
                                Group: "stack.foundation",
                                Scope: apiextensionsv1beta1.NamespaceScoped,
                                Version: "v1",
                                Names: apiextensionsv1beta1.CustomResourceDefinitionNames{
                                        Plural: "workflows",
                                        Singular: "workflow",
                                        Kind: "Workflow",
                                        ShortNames: []string{
                                                "wf",
                                                "wflow",
                                        },
                                },
                        },
                })
        }
}

func printLogsUntilPodFinished(pods corev1.PodInterface, pod *v1.Pod) {
        podWatch, err := pods.Watch(metav1.ListOptions{Watch: true})
        if err != nil {
                MaybeReportErrorAndExit(err)
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

func isPodFinished(pod *v1.Pod) bool {
        return pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed
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

func init() {
        configureKubeStartingCommandFlags(execCmd)
        execCmd.Flags().StringVarP(&execImage, "image", "i", "", "Image to run")
        execCmd.Flags().StringVarP(&execCommandString, "command", "c", "", "Command to run")
        RootCmd.AddCommand(execCmd)
}
