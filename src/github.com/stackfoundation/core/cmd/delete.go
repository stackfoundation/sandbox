package cmd

import (
        "fmt"
        "os"

        "github.com/spf13/cobra"

        "github.com/stackfoundation/core/pkg/workflows"
)

var deleteCmd = &cobra.Command{
        Use:   "delete <workflow>",
        Short: "Delete an existing workflow from the current project",
        Long:  `Delete an existing workflow from the current project.`,
        ValidArgs: []string { "workflow" },
        Run: func(command *cobra.Command, args []string) {
                if len(args) < 1 {
                        fmt.Println("You must specify the name of a workflow to delete!")
                        fmt.Println()
                        fmt.Println("Try running `sbox delete --help` for help")
                        return
                }

                deleted, err := workflows.DeleteWorkflow(args[0])
                if err != nil && os.IsNotExist(err) {
                        fmt.Printf("%v does not exist", args[0])
                        fmt.Println()
                        return
                }

                if deleted && err == nil {
                        fmt.Printf("%v deleted", args[0])
                        fmt.Println()
                        return
                }
        },
}

func init() {
        RootCmd.AddCommand(deleteCmd)
}
