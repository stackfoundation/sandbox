// +build !windows

package path

import (
	"os"
	"path/filepath"

	"github.com/stackfoundation/sandbox/io"
)

func AddSboxToSystemPath(installDirectory string) error {
	return AddToSystemPath(filepath.Join(installDirectory, "sbox"))
}

// AddToSystemPath Add the specified node to the system PATH variable
func AddToSystemPath(node string) error {
	link := filepath.Join("/usr/local/bin", filepath.Base(node))
	exists, err := io.Exists(link)
	if err != nil {
		return err
	}

	if exists {
		err = os.Remove(link)
	}

	if err == nil {
		err = os.Symlink(node, link)
	}

	return err
}
