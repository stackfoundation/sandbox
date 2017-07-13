package cmd

import (
        "github.com/spf13/cobra"
        "fmt"
)

var stopCmd = &cobra.Command{
        Use:   "stop",
        Short: "Stop a running workflow",
        Long:  `Stop a running workflow.`,
        Run: func(command *cobra.Command, args []string) {
                fmt.Println("Sandbox version:", "0.1.0")
        },
}

func init() {
        RootCmd.AddCommand(stopCmd)
}
