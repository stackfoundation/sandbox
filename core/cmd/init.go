package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/stackfoundation/core/pkg/wrapper"
	coreio "github.com/stackfoundation/io"
	"github.com/stackfoundation/log"

	"github.com/spf13/cobra"
)

const wrappersFolder = "sbox-cli"
const nixScript = "sbox"

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the project in the current working directory to use the Sandbox CLI",
	Long: `Sets up the project in the current working directory to use Sandbox CLI.

The Sandbox Command Line Interface (CLI) is a set of small scripts and binaries for all major
platforms (all together, under 150KB) that can be added to your project, and committed to your Git
repository (or other VCS). This allows anyone on who checks out your repository to immediately run
workflows and other Sandbox commands directly from the project root!`,
	Run: func(command *cobra.Command, args []string) {
		projectPath, _ := os.Getwd()

		wrappersWithinProject := filepath.Join(projectPath, wrappersFolder)

		exists, _ := coreio.Exists(wrappersWithinProject)
		if exists {
			fmt.Println("Project already contains an sbox-cli folder - has this project already been initialized?")
			return
		}

		nixScriptWithinProject := filepath.Join(projectPath, nixScript)
		exists, _ = coreio.Exists(nixScriptWithinProject)
		if exists {
			fmt.Println("Project already contains an sbox file - has this project already been initialized?")
			return
		}

		err := wrapper.ExtractWrappers(projectPath)
		if err != nil {
			log.Errorf("Error initializing project: %v", err)
			return
		}

		fmt.Println("Project was initialized with the Sandbox CLI")
	},
}

func init() {
	RootCmd.AddCommand(initCmd)
}
