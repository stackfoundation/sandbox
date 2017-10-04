package hypervisor

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/stackfoundation/core/pkg/minikube/assets"
	"github.com/stackfoundation/install"
)

func xhyveDriverLocation() (string, error) {
	binary, err := os.Executable()
	if err == nil {
		binary, err = filepath.EvalSymlinks(binary)
	}

	if err != nil {
		return "", err
	}

	binaryDirectory, _ := filepath.Abs(filepath.Dir(binary))
	driverPath := filepath.Join(binaryDirectory, "docker-machine-driver-xhyve")

	return driverPath, err
}

func installXhyveDriver() error {
	driverPath, err := xhyveDriverLocation()
	if os.IsNotExist(err) {
		data, err := assets.Asset("out/docker-machine-driver-xhyve")

		if err != nil {
			return err
		}

		ioutil.WriteFile(driverPath, data, 4555)

		cmd := exec.Command("/bin/sh", "-c", "chown root:wheel "+driverPath+" && chmod u+s "+driverPath)
		_, err = cmd.CombinedOutput()
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	return nil
}

// InstallXhyve Perform a xhyve install
func InstallXhyve(fail bool) error {
	err := installXhyveDriver()
	if err != nil {
		if fail {
			return err
		}

		install.ElevatedExecute(os.Args[0], "xhyve --fail")
	}

	return nil
}
