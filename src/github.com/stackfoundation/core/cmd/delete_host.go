package cmd

import (
	"fmt"
	"os"

	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/stackfoundation/core/pkg/minikube/cluster"
	"github.com/stackfoundation/core/pkg/minikube/machine"
)

var deleteHostCmd = &cobra.Command{
	Use:   "delete-host",
	Short: "Delete VM host",
	Long:  `Delete VM host.`,
	Run: func(command *cobra.Command, args []string) {
		api, err := machine.NewAPIClient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
			os.Exit(1)
		}
		defer api.Close()

		err = cluster.DeleteHost(api)
		if err != nil {
			glog.Errorln("Error deleting host:", err)
			MaybeReportErrorAndExit(err)
		}
	},
}

func init() {
	RootCmd.AddCommand(deleteHostCmd)
}
