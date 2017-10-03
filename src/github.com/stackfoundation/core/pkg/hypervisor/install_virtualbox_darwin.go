package hypervisor

import "github.com/stackfoundation/process"

func installVirtualBox(installer string) error {
	_, err := process.CommandOut("/bin/sh", "-c",
		"hdiutil attach \""+installer+"\"; installer -pkg /Volumes/VirtualBox/VirtualBox.pkg -target \"/\"; hdiutil detach /Volumes/VirtualBox")
	return err
}
