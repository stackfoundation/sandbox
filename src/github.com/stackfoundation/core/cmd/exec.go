package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/stackfoundation/core/pkg/workflows"
)

var execCmd = &cobra.Command{
	DisableFlagParsing: true,
	Use:                "exec <image> <command>",
	Short:              "Execute a single command within a Docker container",
	Long: `Execute a single command within a Docker container.

Runs a command within a Docker container. The image can be any Docker image, specified in the <image>:<tag> format.
Try running 'sbox catalog' to see a list of official Docker images that can be used. The command is executed, and
the container runs till the command finishes. Log output from the container is sent directly to stdout.`,
	Run: func(command *cobra.Command, args []string) {
		if len(args) < 1 {
			fmt.Println("You must specify an image!")
			fmt.Println()
			fmt.Println("Try running `sbox exec --help` for help")
			return
		}

		if len(args) < 2 {
			if args[0] == "--help" || args[0] == "-h" {
				command.Help()
				return
			}

			fmt.Println("You must specify a command to run!")
			fmt.Println()
			fmt.Println("Try running `sbox exec --help` for help")
			return
		}

		startKube()

		image := args[0]

		workflows.ExecuteCommand(image, filterArgs(args[1:]))
	},
}

func init() {
	configureKubeStartingCommandFlags(execCmd)
	RootCmd.AddCommand(execCmd)
}
