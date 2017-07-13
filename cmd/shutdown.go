package cmd

import (
        "github.com/spf13/cobra"
        "fmt"
)

var shutdownCmd = &cobra.Command{
        Use:   "shutdown",
        Short: "Stop all running workflows, and shutdown Sandbox workflow runner",
        Long:  `Stop all running workflows, and shutdown Sandbox workflow runner.`,
        Run: func(command *cobra.Command, args []string) {
                fmt.Println("Sandbox version:", "0.1.0")
        },
}

func init() {
        RootCmd.AddCommand(shutdownCmd)
}
