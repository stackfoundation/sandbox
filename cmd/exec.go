package cmd

import (
        "github.com/spf13/cobra"
        "fmt"
)

var execCmd = &cobra.Command{
        Use:   "exec",
        Short: "Execute a single command within a Docker container",
        Long:  `Execute a single command within a Docker container.`,
        Run: func(command *cobra.Command, args []string) {
                fmt.Println("Sandbox version:", "0.1.0")
        },
}

func init() {
        RootCmd.AddCommand(execCmd)
}
