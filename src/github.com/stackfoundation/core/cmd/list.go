package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/stackfoundation/core/pkg/workflows/files"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all workflows available in the current project",
	Long:  `List all workflows available in the current project.`,
	Run: func(command *cobra.Command, args []string) {
		projectWorkflows, err := files.ListWorkflows()
		if err != nil || len(projectWorkflows) < 1 {
			fmt.Println("No workflows")
			return
		}

		fmt.Println("NAME")
		for _, workflow := range projectWorkflows {
			fmt.Printf("%v", workflow)
			fmt.Println()
		}
	},
}

func init() {
	RootCmd.AddCommand(listCmd)
}
