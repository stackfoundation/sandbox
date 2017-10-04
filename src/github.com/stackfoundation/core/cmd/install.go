package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/stackfoundation/core/pkg/path"
	"github.com/stackfoundation/core/pkg/wrapper"
	"github.com/stackfoundation/install"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the Sandbox command-line globally",
	Long: `Install the Sandbox command-line, and make it available globally.

This adds to the system PATH variable so that the Sandbox command-line is available globally.`,
	Run: func(command *cobra.Command, args []string) {
		cliDirectory, err := install.GetInstallPath()
		if err != nil {
			fmt.Println("Error installing CLI: " + err.Error())
			os.Exit(1)
		}

		wrapper.ExtractWrappers(cliDirectory)

		err = path.AddToSystemPath(cliDirectory)
		if err != nil {
			fmt.Println("Error installing CLI: " + err.Error())
			os.Exit(1)
		}

		fmt.Println("Sandbox CLI globally, as the 'sbox' command")
	},
}

func init() {
	RootCmd.AddCommand(installCmd)
}
