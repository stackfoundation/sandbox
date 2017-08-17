package cmd

import (
        "github.com/spf13/cobra"

        "github.com/stackfoundation/core/pkg/workflows"
        "fmt"
        "strings"
)


var runCmd = &cobra.Command{
        DisableFlagParsing: true,
        Use:   "run",
        Short: "Run a workflow available in the current project",
        Long:  `Run a workflow available in the current project.`,
        Run: func(command *cobra.Command, args []string) {
                if len(args) < 1 {
                        fmt.Println("You must specify a workflow!")
                        fmt.Println()
                        fmt.Println("Try running `sbox run --help` for help")
                        return
                }

                if args[0] == "--help" || args[0] == "-h" {
                        command.Help()
                        return
                }

                startKube()

                workflow := strings.Join(args, " ")

                err := workflows.RunCommand(workflow)
                if err != nil {
                        panic(err)
                }
        },
}

func init() {
        configureKubeStartingCommandFlags(runCmd)
        RootCmd.AddCommand(runCmd)
}
