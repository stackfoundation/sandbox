package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/stackfoundation/core/pkg/version"
)

var coreVersion = "0.1.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version of Sandbox",
	Long:  `Print the version of Sandbox.`,
	Run: func(command *cobra.Command, args []string) {
		fmt.Println("Sandbox version:", coreVersion)
		fmt.Println("Minikube version:", version.GetVersion())
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
