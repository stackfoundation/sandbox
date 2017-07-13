package cmd

import (
        "fmt"
        "github.com/spf13/cobra"
)

var wrapperCmd = &cobra.Command{
        Use:   "wrapper",
        Short: "Install Sandbox wrapper in the current working directory",
        Long:  `Install the Sandbox wrapper in the current working directory.

The Sandbox wrapper is a set of small scripts and binaries for all major platforms (all together,
under 100KB) that can be added to your project, and committed to your Git repository (or other
VCS). This allows anyone on who checks out your repository to immediately run workflows and other
Sandbox commands directly from the project root!`,
        Run: func(command *cobra.Command, args []string) {
                fmt.Println("Sandbox version:", "0.1.0")
        },
}

func init() {
        RootCmd.AddCommand(wrapperCmd)
}
