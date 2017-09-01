package cmd

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/spf13/cobra"
)

func ShellExecute(file, parameters string) error {
	return errors.New("Not Implemented")
}

func NotifySettingChange() uintptr {
	return 0
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install the Sandbox command-line globally",
	Long: `Install the Sandbox command-line, and make it available globally.

This adds to the system PATH variable so that the Sandbox command-line is available globally.`,
	Run: func(command *cobra.Command, args []string) {
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
	},
}
