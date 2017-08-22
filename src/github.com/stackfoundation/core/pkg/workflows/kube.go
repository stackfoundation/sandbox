package workflows

import (
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	extensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	extensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset/typed/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/pborman/uuid"
	log "github.com/stackfoundation/core/pkg/log"
	"github.com/stackfoundation/core/pkg/minikube/config"
	"github.com/stackfoundation/core/pkg/minikube/constants"
	"github.com/stackfoundation/core/pkg/util/kubeconfig"
	"k8s.io/client-go/dynamic"
)

var driveLetterReplacement = regexp.MustCompile("^([a-zA-Z])\\:")

type podCreationSpec struct {
	image       string
	projectRoot string
	command     []string
	volumes     []Volume
}

func createDynamicClient() (*dynamic.Client, error) {
	restClientConfig, err := createRestClientConfig()
	if err != nil {
		return nil, err
	}

	restClientConfig.ContentConfig.GroupVersion = &schema.GroupVersion{
		Group:   WorkflowsGroupName,
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
		Group:   WorkflowsGroupName,
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

func createAndRunPod(clientSet *kubernetes.Clientset, creationSpec *podCreationSpec) error {
	pods := clientSet.Pods("default")

	uuid := uuid.NewUUID()
	containerName := "sbox-" + uuid.String()

	pod, err := createPod(pods, containerName, creationSpec)
	if err != nil {
		return err
	}

	printLogsUntilPodFinished(pods, pod)
	return nil
}

func lowercaseDriveLetter(text []byte) []byte {
	lowercase := strings.ToLower(string(text))
	return []byte("/" + lowercase[:len(lowercase)-1])
}

func createVolumes(projectRoot string, volumes []Volume) ([]v1.VolumeMount, []v1.Volume) {
	var mounts []v1.VolumeMount
	var podVolumes []v1.Volume

	if len(volumes) > 0 {
		for _, volume := range volumes {
			if len(volume.Name) < 1 {
				if len(volume.HostPath) > 0 {
					uuid := uuid.NewUUID()
					volume.Name = "vol-" + uuid.String()
				} else {
					log.Debugf("No name was specified for non-host volume, ignoring")
					continue
				}
			}

			var volumeSource v1.VolumeSource

			if len(volume.HostPath) > 0 {
				absoluteHostPath := path.Join(filepath.ToSlash(projectRoot), volume.HostPath)
				absoluteHostPath = string(driveLetterReplacement.ReplaceAllFunc(
					[]byte(absoluteHostPath),
					lowercaseDriveLetter))

				volumeSource = v1.VolumeSource{
					HostPath: &v1.HostPathVolumeSource{
						Path: absoluteHostPath,
					},
				}

				log.Debugf("Mounting host path \"%v\" at \"%v\"", absoluteHostPath, volume.MountPath)
			} else {
				volumeSource = v1.VolumeSource{
					EmptyDir: &v1.EmptyDirVolumeSource{},
				}

				log.Debugf("Mounting volume \"%v\" at \"%v\"", volume.Name, volume.MountPath)
			}

			podVolumes = append(podVolumes, v1.Volume{
				Name:         volume.Name,
				VolumeSource: volumeSource,
			})

			mounts = append(mounts, v1.VolumeMount{
				Name:      volume.Name,
				MountPath: volume.MountPath,
			})
		}
	}

	return mounts, podVolumes
}

func createPod(pods corev1.PodInterface, name string, creationSpec *podCreationSpec) (*v1.Pod, error) {
	mounts, podVolumes := createVolumes(creationSpec.projectRoot, creationSpec.volumes)

	return pods.Create(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "sbox-",
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:            name,
					Image:           creationSpec.image,
					Command:         creationSpec.command,
					ImagePullPolicy: v1.PullIfNotPresent,
					VolumeMounts:    mounts,
				},
			},
			Volumes:       podVolumes,
			RestartPolicy: v1.RestartPolicyNever,
		},
	})
}

func createWorkflowResourceIfRequired(customResourceDefintions extensionsclient.CustomResourceDefinitionInterface) error {
	_, err := customResourceDefintions.Get(WorkflowsCustomResource, metav1.GetOptions{})
	if err != nil {
		log.Debugf("Creating custom resource definition for workflows in Kubernetes")
		_, err = customResourceDefintions.Create(&extensions.CustomResourceDefinition{
			ObjectMeta: metav1.ObjectMeta{
				Name: WorkflowsCustomResource,
			},
			Spec: extensions.CustomResourceDefinitionSpec{
				Group:   WorkflowsGroupName,
				Scope:   extensions.NamespaceScoped,
				Version: WorkflowsGroupVersion,
				Names: extensions.CustomResourceDefinitionNames{
					Plural:   WorkflowsPluralName,
					Singular: WorkflowsSingularName,
					Kind:     WorkflowsKind,
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

	if follow {
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
