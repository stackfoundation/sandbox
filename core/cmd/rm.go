package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stackfoundation/sandbox/core/pkg/minikube/cluster"
	"github.com/stackfoundation/sandbox/core/pkg/minikube/machine"
	"github.com/stackfoundation/sandbox/log"
)

var deleteHostCmd = &cobra.Command{
	Use:   "rm",
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
			log.Errorf("Error deleting host: %v\n", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(deleteHostCmd)
}
