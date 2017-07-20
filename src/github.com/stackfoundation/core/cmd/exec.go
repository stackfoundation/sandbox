package cmd

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/stackfoundation/core/pkg/minikube/config"
	"github.com/stackfoundation/core/pkg/minikube/constants"
	"github.com/stackfoundation/core/pkg/minikube/machine"
	"github.com/stackfoundation/core/pkg/util/kubeconfig"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var execImage string
var execCommandString string

var execCmd = &cobra.Command{
	Use:   "exec",
	Short: "Execute a single command within a Docker container",
	Long:  `Execute a single command within a Docker container.`,
	Run: func(command *cobra.Command, args []string) {
		startKube()

		api, err := machine.NewAPIClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
			os.Exit(1)
		}
		defer api.Close()

		con, err := kubeconfig.ReadConfigOrNew(constants.KubeconfigPath)
		if err != nil {
			glog.Errorln("Error kubeconfig status:", err)
			MaybeReportErrorAndExit(err)
		}

		configOverrides := &clientcmd.ConfigOverrides{}
		k8sClientConfig := clientcmd.NewNonInteractiveClientConfig(*con, config.GetMachineName(), configOverrides, nil)
		restClient, err := k8sClientConfig.ClientConfig()
		clientset, err := kubernetes.NewForConfig(restClient)

		job, err := clientset.Pods("default").Create(&v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				GenerateName: "sbox-",
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					v1.Container{
						Name:    "sboxstep",
						Image:   execImage,
						Command: []string{execCommandString},
					},
				},
				RestartPolicy: v1.RestartPolicyNever,
			},
		})
		if err != nil {
			panic(err.Error())
		}
		fmt.Printf("Started the container: %v!\n", job.Status.StartTime)
	},
}

func init() {
	configureKubeStartingCommandFlags(execCmd)
	execCmd.Flags().StringVarP(&execImage, "image", "i", "", "Image to run")
	execCmd.Flags().StringVarP(&execCommandString, "command", "c", "", "Command to run")
	RootCmd.AddCommand(execCmd)
}
