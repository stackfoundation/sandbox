// +build !windows

package path

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"

	"github.com/stackfoundation/io"
)

// AddToSystemPath Add the specified directory to the system PATH variable
func AddToSystemPath(directory string) error {
	usr, err := user.Current()
	if err != nil {
		return err
	}

	profile := filepath.Join(usr.HomeDir, ".profile")
	exists, err := io.Exists(profile)
	if err != nil {
		return err
	}

	if exists {
		ioutil.WriteFile(profile, []byte("\nPATH=$PATH:"+directory+"\n"), 0644)
	} else {
		file, err := os.OpenFile(profile, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return err
		}

		defer file.Close()

		if _, err = file.WriteString("\nPATH=$PATH:" + directory + "\n"); err != nil {
			panic(err)
		}
	}
}
