package cmd

import (
        "github.com/spf13/cobra"
)

var stopCmd = &cobra.Command{
        Use:   "stop",
        Short: "Stop a running workflow",
        Long:  `Stop a running workflow.`,
        Run: func(command *cobra.Command, args []string) {
        },
}

func init() {
        RootCmd.AddCommand(stopCmd)
}
