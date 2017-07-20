package cmd

import (
        "github.com/spf13/cobra"
        "fmt"
)

var editCmd = &cobra.Command{
        Use:   "edit",
        Short: "Edit an existing workflow in the current project",
        Long:  `Edit an existing workflow in the current project.`,
        Run: func(command *cobra.Command, args []string) {
                fmt.Println("Sandbox version:", "0.1.0")
        },
}

func init() {
        RootCmd.AddCommand(editCmd)
}
