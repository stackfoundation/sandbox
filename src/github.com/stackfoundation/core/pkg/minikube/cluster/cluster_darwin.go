/*
Copyright 2016 The Kubernetes Authors All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cluster

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/docker/machine/drivers/vmwarefusion"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/stackfoundation/core/pkg/minikube/assets"
	cfg "github.com/stackfoundation/core/pkg/minikube/config"
	"github.com/stackfoundation/core/pkg/minikube/constants"
)

func createVMwareFusionHost(config MachineConfig) drivers.Driver {
	d := vmwarefusion.NewDriver(cfg.GetMachineName(), constants.GetMinipath()).(*vmwarefusion.Driver)
	d.Boot2DockerURL = config.Downloader.GetISOFileURI(config.MinikubeISO)
	d.Memory = config.Memory
	d.CPU = config.CPUs
	d.DiskSize = config.DiskSize

	// TODO(philips): push these defaults upstream to fixup this driver
	d.SSHPort = 22
	d.ISO = d.ResolveStorePath("boot2docker.iso")
	return d
}

type xhyveDriver struct {
	*drivers.BaseDriver
	Boot2DockerURL string
	BootCmd        string
	CPU            int
	CaCertPath     string
	DiskSize       int64
	MacAddr        string
	Memory         int
	PrivateKeyPath string
	UUID           string
	NFSShare       bool
	DiskNumber     int
	Virtio9p       bool
	Virtio9pFolder string
	QCow2          bool
	RawDisk        bool
}

func createXhyveHost(config MachineConfig) *xhyveDriver {
	checkXhyvePlugin()
	useVirtio9p := !config.DisableDriverMounts
	return &xhyveDriver{
		BaseDriver: &drivers.BaseDriver{
			MachineName: cfg.GetMachineName(),
			StorePath:   constants.GetMinipath(),
		},
		Memory:         config.Memory,
		CPU:            config.CPUs,
		Boot2DockerURL: config.Downloader.GetISOFileURI(config.MinikubeISO),
		BootCmd:        "loglevel=3 user=docker console=ttyS0 console=tty0 noembed nomodeset norestore waitusb=10 systemd.legacy_systemd_cgroup_controller=yes base host=" + cfg.GetMachineName(),
		DiskSize:       int64(config.DiskSize),
		Virtio9p:       useVirtio9p,
		Virtio9pFolder: "/Users",
		QCow2:          false,
		RawDisk:        config.XhyveDiskDriver == "virtio-blk",
	}
}

func detectVBoxManageCmd() string {
	cmd := "VBoxManage"
	if path, err := exec.LookPath(cmd); err == nil {
		return path
	}
	return cmd
}

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
