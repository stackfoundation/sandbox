package cmd

import (
        "github.com/spf13/cobra"
        "fmt"
        "os"
        "github.com/stackfoundation/core/pkg/minikube/cluster"
        "github.com/stackfoundation/core/pkg/minikube/machine"
)

var shutdownCmd = &cobra.Command{
        Use:   "shutdown",
        Short: "Stop all running workflows, and shutdown Sandbox workflow runner",
        Long:  `Stop all running workflows, and shutdown Sandbox workflow runner.`,
        Run: func(command *cobra.Command, args []string) {
                api, err := machine.NewAPIClient()
                if err != nil {
                        fmt.Fprintf(os.Stderr, "Error getting client: %s\n", err)
                        os.Exit(1)
                }
                defer api.Close()

                if err = cluster.StopHost(api); err != nil {
                        fmt.Println("Error stopping machine: ", err)
                        MaybeReportErrorAndExit(err)
                }
                fmt.Println("Machine stopped.")

                if err := KillMountProcess(); err != nil {
                        fmt.Println("Errors occurred deleting mount process: ", err)
                }
        },
}

func init() {
        RootCmd.AddCommand(shutdownCmd)
}
