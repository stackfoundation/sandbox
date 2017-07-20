package cmd

import (
        "fmt"
        "github.com/spf13/cobra"
        "github.com/stackfoundation/core/pkg/version"
)

var versionCmd = &cobra.Command{
        Use:   "version",
        Short: "Print the version of Sandbox",
        Long:  `Print the version of Sandbox.`,
        Run: func(command *cobra.Command, args []string) {
                fmt.Println("Sandbox version:", "0.1.0")
                fmt.Println("Minikube version:", version.GetVersion())
        },
}

func init() {
        RootCmd.AddCommand(versionCmd)
}
