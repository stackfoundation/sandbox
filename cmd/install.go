package cmd

import (
        "github.com/spf13/cobra"
        "fmt"
)

var installCmd = &cobra.Command{
        Use:   "install",
        Short: "Install the Sandbox command-line globally",
        Long:  `Install the Sandbox command-line, and make it available globally.

This adds to the system PATH variable so that the Sandbox command-line is available globally.`,
        Run: func(command *cobra.Command, args []string) {
                fmt.Println("Sandbox version:", "0.1.0")
        },
}

func init() {
        RootCmd.AddCommand(installCmd)
}
