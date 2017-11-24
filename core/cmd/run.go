package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/stackfoundation/core/pkg/workflows/cmd"
	"github.com/stackfoundation/log"
)

func parseFlags(args []string) []string {
	var filtered []string
	var ignoreNext bool

	for _, arg := range args {
		if arg == "-d" || arg == "--debug" {
			log.SetDebug(true)
			filtered = append(filtered, arg)
		} else if arg == "--original-command" {
			ignoreNext = true
		} else if ignoreNext {
			ignoreNext = false
		} else {
			filtered = append(filtered, arg)
		}
	}

	return filtered
}

func haveMinArgs(args []string) bool {
	if len(args) < 1 {
		fmt.Println("You must specify a workflow!")
		fmt.Println()
		fmt.Println("Try running `sbox run --help` for help")
		return false
	}

	return true
}

var runCmd = &cobra.Command{
	DisableFlagParsing: true,
	Use:                "run",
	Short:              "Run a workflow available in the current project",
	Long:               `Run a workflow available in the current project.`,
	Run: func(command *cobra.Command, args []string) {
		if !haveMinArgs(args) {
			return
		}

		if args[0] == "--help" || args[0] == "-h" {
			command.Help()
			return
		}

		args = parseFlags(args)
		if !haveMinArgs(args) {
			return
		}

		workflowName := args[0]
		args = args[1:]

		startKube()

		err := cmd.Run(workflowName, args)
		if err != nil {
			if os.IsNotExist(err) {
				fmt.Printf("No workflow named %v", workflowName)
				fmt.Println()
			} else if context.Canceled != err {
				fmt.Printf(err.Error())
				fmt.Println()
			}
		}
	},
}

func init() {
	configureKubeStartingCommandFlags(runCmd)
	RootCmd.AddCommand(runCmd)
}
