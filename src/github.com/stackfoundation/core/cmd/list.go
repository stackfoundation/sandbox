package cmd

import (
        "github.com/spf13/cobra"
        "fmt"
)

var listCmd = &cobra.Command{
        Use:   "list",
        Short: "List all workflows available in the current project",
        Long:  `List all workflows available in the current project.`,
        Run: func(command *cobra.Command, args []string) {
                workflows, err := ListWorkflows()
                if err != nil || len(workflows) < 1 {
                        fmt.Println("No workflows")
                        return
                }

                fmt.Println("NAME")
                for _, workflow := range workflows {
                        fmt.Printf("%v", workflow)
                        fmt.Println()
                }
        },
}

func init() {
        RootCmd.AddCommand(listCmd)
}
