// +build !windows

package cmd

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
)

func performInstall(command *cobra.Command, args []string) {
	usr, err := user.Current()
	if err != nil {
		fmt.Println(err)
		return
	}

	corePath := filepath.Join(usr.HomeDir, "bin/sbox")
	sourcePath, _ := filepath.Abs(os.Args[0])

	err = os.Symlink(sourcePath, corePath)

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("Sandbox command-line installed globally, as the 'sbox' command")
}
