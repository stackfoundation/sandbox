package kube

import (
	"github.com/stackfoundation/core/pkg/minikube/config"
	"github.com/stackfoundation/core/pkg/minikube/constants"
	"github.com/stackfoundation/core/pkg/util/kubeconfig"
	workflowsv1 "github.com/stackfoundation/core/pkg/workflows/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

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

func createDynamicClient() (*dynamic.Client, error) {
	restClientConfig, err := createRestClientConfig()
	if err != nil {
		return nil, err
	}

	restClientConfig.ContentConfig.GroupVersion = &schema.GroupVersion{
		Group:   workflowsv1.WorkflowsGroupName,
		Version: workflowsv1.WorkflowsGroupVersion,
	}

	restClientConfig.APIPath = "/apis"

	return dynamic.NewClient(restClientConfig)
}

// CreateExtensionsClient Create a K8s extensions client
func CreateExtensionsClient() (*clientset.Clientset, error) {
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

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(workflowsv1.SchemeGroupVersion,
		&workflowsv1.Workflow{},
		&workflowsv1.WorkflowList{},
	)
	metav1.AddToGroupVersion(scheme, workflowsv1.SchemeGroupVersion)
	return nil
}

func createRestClient() (*rest.RESTClient, error) {
	restClientConfig, err := createRestClientConfig()
	if err != nil {
		return nil, err
	}

	restClientConfig.ContentConfig.GroupVersion = &schema.GroupVersion{
		Group:   workflowsv1.WorkflowsGroupName,
		Version: workflowsv1.WorkflowsGroupVersion,
	}
	restClientConfig.APIPath = "/apis"

	schemeBuilder := runtime.NewSchemeBuilder(addKnownTypes)

	scheme := runtime.NewScheme()
	schemeBuilder.AddToScheme(scheme)

	restClientConfig.NegotiatedSerializer =
		serializer.DirectCodecFactory{CodecFactory: serializer.NewCodecFactory(scheme)}

	return rest.RESTClientFor(restClientConfig)
}