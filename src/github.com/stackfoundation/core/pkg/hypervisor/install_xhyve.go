package hypervisor

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/stackfoundation/core/pkg/minikube/assets"
	"github.com/stackfoundation/install"
	"github.com/stackfoundation/process"
)

func isxhyveInstalled() bool {
	_, info, err := xhyveDriverLocation()

	return err == nil && info.Mode()&os.ModeSetuid > 0
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
		fmt.Println("Xhyve does not exist in " + driverPath + ", installing...")
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

	fmt.Println("Xhyve installed in " + driverPath)

	return nil
}

// InstallXhyve Perform a xhyve install
func InstallXhyve(fail bool) error {
	err := installXhyveDriver()
	if err != nil {
		if fail {
			return err
		}

		fmt.Println("Elevating")

		err = install.ElevatedExecute(os.Args[0], "xhyve --fail")

		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}
