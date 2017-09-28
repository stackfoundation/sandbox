package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/stackfoundation/core/pkg/minikube/assets"

	"github.com/spf13/cobra"
)

var wrapperCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize the project in the current working directory to use the Sandbox CLI",
	Long: `Sets up the project in the current working directory to use Sandbox CLI.

The Sandbox Command Line Interface (CLI) is a set of small scripts and binaries for all major
platforms (all together, under 100KB) that can be added to your project, and committed to your Git
repository (or other VCS). This allows anyone on who checks out your repository to immediately run
workflows and other Sandbox commands directly from the project root!`,
	Run: func(command *cobra.Command, args []string) {

		path, _ := os.Getwd()
		sboxFolder := filepath.Join(path, ".sbox")
		_, sboxFolderErr := os.Stat(sboxFolder)
		_, sboxDarwinExecErr := os.Stat(filepath.Join(path, "sbox"))

		if !os.IsNotExist(sboxFolderErr) || !os.IsNotExist(sboxDarwinExecErr) {
			fmt.Println("Project already contains an sbox folder in the root directory - has sandbox been initialized already for this project?")
			return
		}

		os.Mkdir(sboxFolder, 0711)
		wrapperDarwin, err := assets.Asset("out/wrapper-darwin")

		if err != nil {
			fmt.Println("wrapper-darwin asset was not found")
			return
		}

		darwinSboxExec, err := assets.Asset("out/sbox")

		if err != nil {
			fmt.Println("wrapper-darwin asset was not found")
			return
		}

		ioutil.WriteFile(filepath.Join(sboxFolder, "wrapper-darwin"), wrapperDarwin, 4555)
		ioutil.WriteFile(filepath.Join(path, "sbox"), darwinSboxExec, 4555)
	},
}

func init() {
	RootCmd.AddCommand(wrapperCmd)
}
