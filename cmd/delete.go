package cmd

import (
        "github.com/spf13/cobra"
        "fmt"
)

var deleteCmd = &cobra.Command{
        Use:   "delete",
        Short: "Delete an existing workflow from the current project",
        Long:  `Delete an existing workflow from the current project.`,
        Run: func(command *cobra.Command, args []string) {
                fmt.Println("Sandbox version:", "0.1.0")
        },
}

func init() {
        RootCmd.AddCommand(deleteCmd)
}
