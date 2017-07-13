package cmd

import (
        "github.com/spf13/cobra"
        "fmt"
)

var createCmd = &cobra.Command{
        Use:   "create",
        Short: "Create a new workflow available in the current project",
        Long:  `Create a new workflow available in the current project.`,
        Run: func(command *cobra.Command, args []string) {
                fmt.Println("Sandbox version:", "0.1.0")
        },
}

func init() {
        RootCmd.AddCommand(createCmd)
}
