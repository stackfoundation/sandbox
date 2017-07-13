package cmd

import (
        "github.com/spf13/cobra"
        "fmt"
)

var runCmd = &cobra.Command{
        Use:   "run",
        Short: "Run a workflow available in the current project",
        Long:  `Run a workflow available in the current project.`,
        Run: func(command *cobra.Command, args []string) {
                fmt.Println("Sandbox version:", "0.1.0")
        },
}

func init() {
        RootCmd.AddCommand(runCmd)
}
