package cmd

import (
        "github.com/spf13/cobra"
        "fmt"
)

var listCmd = &cobra.Command{
        Use:   "list",
        Short: "List all workflows available in the current project",
        Long:  `List all workflows available in the current project.`,
        Run: func(command *cobra.Command, args []string) {
                fmt.Println("Sandbox version:", "0.1.0")
        },
}

func init() {
        RootCmd.AddCommand(listCmd)
}
