package hypervisor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/stackfoundation/sandbox/core/pkg/minikube/assets"
	"github.com/stackfoundation/sandbox/core/pkg/path"
	"github.com/stackfoundation/sandbox/install"
	"github.com/stackfoundation/sandbox/process"
)

func isxhyveInstalled() bool {
	installed, _ := path.IsInSystemPath("docker-machine-driver-xhyve")

	return installed
}

func xhyveDriverLocation() (string, os.FileInfo, error) {
	binary, err := os.Executable()
	if err == nil {
		binary, err = filepath.EvalSymlinks(binary)
	}

	if err != nil {
		return "", nil, err
	}

	binaryDirectory, _ := filepath.Abs(filepath.Dir(binary))
	driverPath := filepath.Join(binaryDirectory, "docker-machine-driver-xhyve")

	info, err := os.Stat(driverPath)

	return driverPath, info, err
}

func installXhyveDriver() error {
	driverPath, info, err := xhyveDriverLocation()
	if os.IsNotExist(err) {
		data, err := assets.Asset("out/docker-machine-driver-xhyve")

		if err != nil {
			fmt.Println(err)
			return err
		}

		err = ioutil.WriteFile(driverPath, data, 4555)

		if err != nil {
			fmt.Println(err)
			return err
		}

		info, err = os.Stat(driverPath)
	} else if err != nil {
		return err
	}

	if info.Mode()&os.ModeSetuid == 0 {
		err = process.CombineStdStreams("/bin/sh", "-c", "chown root:wheel "+driverPath+" && chmod u+s "+driverPath)
		if err != nil {
			return err
		}
	}

	err = path.AddToSystemPath(driverPath)

	if err != nil {
		return err
	}

	fmt.Println("xhyve driver installed.")

	return nil
}

// InstallXhyve Perform a xhyve install
// Will elevate by default
func InstallXhyve(fail bool) error {
	if fail {
		return installXhyveDriver()
	}

	err := install.ElevatedExecute(os.Args[0], "xhyve --fail")

	if err != nil {
		fmt.Println(err)
	}

	return err
}
