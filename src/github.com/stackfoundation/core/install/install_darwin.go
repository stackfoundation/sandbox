package bootstrap

import (
	"log"
	"os/user"
	"path/filepath"
)

func getStackFoundationRoot() (string, error) {
	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	return filepath.Join(usr.HomeDir, ".sf"), nil
}
