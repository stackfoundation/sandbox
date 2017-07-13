package cmd

import (
        "github.com/spf13/cobra"
        "fmt"
)

var statusCmd = &cobra.Command{
        Use:   "status",
        Short: "Show the status of any running workflows, as well as overall Sandbox status",
        Long:  `Show the status of any running workflows, as well as overall Sandbox status.`,
        Run: func(command *cobra.Command, args []string) {
                fmt.Println("Sandbox version:", "0.1.0")
        },
}

func init() {
        RootCmd.AddCommand(statusCmd)
}
