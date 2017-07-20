package cmd

import (
        "github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
        Use:   "exec",
        Short: "Execute a single command within a Docker container",
        Long:  `Execute a single command within a Docker container.`,
        Run: func(command *cobra.Command, args []string) {
                startKube()
        },
}

func init() {
        configureKubeStartingCommandFlags(execCmd)
        RootCmd.AddCommand(execCmd)
}
