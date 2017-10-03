package hypervisor

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/stackfoundation/core/pkg/minikube/assets"
)

func checkXhyvePlugin() {
	ex, err := os.Executable()

	if err == nil {
		ex, err = filepath.EvalSymlinks(ex)
	}

	currDir, _ := filepath.Abs(filepath.Dir(ex))
	binaryPath := filepath.Join(currDir, "docker-machine-driver-xhyve")

	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		data, err := assets.Asset("out/docker-machine-driver-xhyve")

		if err != nil {
			fmt.Println("docker-machine-driver-xhyve asset was not found")
			return
		}

		ioutil.WriteFile(binaryPath, data, 4555)

		fmt.Println("For using xhyve as a driver for the sandbox cluster, the xhyve plugin needs to be given root:wheel ownership. You may need to authorize this superuser operation.")
		cmd := exec.Command("/bin/sh", "-c", "sudo chown root:wheel "+binaryPath+" && sudo chmod u+s "+binaryPath)
		_, err = cmd.CombinedOutput()
		if err != nil {
			fmt.Println(err)
		}
	}
}
