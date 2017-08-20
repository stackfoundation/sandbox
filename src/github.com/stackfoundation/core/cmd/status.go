package cmd

import (
	"fmt"
	"os"

	"github.com/docker/machine/libmachine/state"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/stackfoundation/core/pkg/minikube/cluster"
	"github.com/stackfoundation/core/pkg/minikube/config"
	"github.com/stackfoundation/core/pkg/minikube/machine"
	"github.com/stackfoundation/core/pkg/util/kubeconfig"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the status of any running workflows, as well as overall Sandbox status",
	Long:  `Show the status of any running workflows, as well as overall Sandbox status.`,
	Run: func(command *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
			os.Exit(1)
		}
		defer api.Close()

		machineStatus, err := cluster.GetHostStatus(api)
		if err != nil {
			glog.Errorln("Error getting machine status:", err)
			MaybeReportErrorAndExit(err)
		}

		localKubeStatus := state.None.String()
		kubeRunning := false
		if machineStatus == state.Running.String() {
			localKubeStatus, err = cluster.GetLocalkubeStatus(api)
			if err != nil {
				glog.Errorln("Error localkube status:", err)
				MaybeReportErrorAndExit(err)
			}
			ip, err := cluster.GetHostDriverIP(api)
			if err != nil {
				glog.Errorln("Error host driver ip status:", err)
				MaybeReportErrorAndExit(err)
			}
			kubeRunning, err = kubeconfig.GetKubeConfigStatus(ip, GetKubeConfigPath(), config.GetMachineName())
			if err != nil {
				glog.Errorln("Error kubeconfig status:", err)
				MaybeReportErrorAndExit(err)
			}
		}

		if kubeRunning && localKubeStatus == state.Running.String() {
			fmt.Println("The Sandbox VM is up and running, ready to run workflows")
		} else {
			fmt.Println("No Sandbox VM is running. A VM will be started when you run a command or workflow")
		}

		fmt.Println()
		fmt.Println(`Try running "sbox exec --help" or "sbox run --help" for more help`)
	},
}

func init() {
	RootCmd.AddCommand(statusCmd)
}
