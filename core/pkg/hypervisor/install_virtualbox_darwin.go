package hypervisor

import "github.com/stackfoundation/process"

func installVirtualBoxWithInstaller(installer string) error {
	_, err := process.CommandOut("/bin/sh", "-c", "sudo hdiutil attach \""+installer+"\"")

	if err == nil {
		_, err = process.CommandOut("/bin/sh", "-c", "sudo installer -pkg /Volumes/VirtualBox/VirtualBox.pkg -target \"/\"")

		_, err = process.CommandOut("/bin/sh", "-c", "hdiutil detach /Volumes/VirtualBox")
	}
	return err
}
