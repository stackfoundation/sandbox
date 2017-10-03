package hypervisor

import (
	"fmt"

	"github.com/stackfoundation/metadata"
)

func SelectAndPrepareHypervisor(preferred string) string {
	var m *metadata.Metadata
	var err error
	var vbox string

	if preferred == "auto" {
		m, err = metadata.GetMetadata()
		if err == nil && m != nil {
			preferred = m.Driver
		}
	}

	if preferred == "auto" || len(preferred) < 1 {
		preferred = platformPreferred()
	}

	if preferred == "virtualbox" {
		vbox = DetectVBoxManageCmd()
		if len(vbox) > 0 {
			fmt.Println("Using VBOX at " + vbox)
		}

		fmt.Println("VBOX DL from " + platformVirtualBoxPackageURL())
	}

	if m != nil &&
		(m.Driver != preferred ||
			(preferred == "virtualbox" && m.VirtualBox != vbox)) {
		m.Driver = preferred
		m.VirtualBox = vbox

		metadata.SaveMetadata(m)
	}

	return preferred
}
