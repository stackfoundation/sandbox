package hypervisor

import "runtime"

type pkg struct {
	p32 string
	p64 string
}

var pkgs = map[string]pkg{
	"jessie": pkg{
		p32: "virtualbox-5.1_5.1.28-117968~Debian~jessie_i386.deb",
		p64: "virtualbox-5.1_5.1.28-117968~Debian~jessie_amd64.deb",
	},
	"stretch": pkg{
		p32: "virtualbox-5.1_5.1.28-117968~Debian~stretch_i386.deb",
		p64: "virtualbox-5.1_5.1.28-117968~Debian~stretch_amd64.deb",
	},
	"wheezy": pkg{
		p32: "virtualbox-5.1_5.1.28-117968~Debian~wheezy_i386.deb",
		p64: "virtualbox-5.1_5.1.28-117968~Debian~wheezy_amd64.deb",
	},
	"precise": pkg{
		p32: "virtualbox-5.1_5.1.28-117968~Ubuntu~precise_i386.deb",
		p64: "virtualbox-5.1_5.1.28-117968~Ubuntu~precise_amd64.deb",
	},
	"trusty": pkg{
		p32: "virtualbox-5.1_5.1.28-117968~Ubuntu~trusty_i386.deb",
		p64: "virtualbox-5.1_5.1.28-117968~Ubuntu~trusty_amd64.deb",
	},
	"wily": pkg{
		p32: "virtualbox-5.1_5.1.28-117968~Ubuntu~wily_i386.deb",
		p64: "virtualbox-5.1_5.1.28-117968~Ubuntu~wily_amd64.deb",
	},
	"xenial": pkg{
		p32: "virtualbox-5.1_5.1.28-117968~Ubuntu~xenial_i386.deb",
		p64: "virtualbox-5.1_5.1.28-117968~Ubuntu~xenial_amd64.deb",
	},
	"yakkety": pkg{
		p32: "virtualbox-5.1_5.1.28-117968~Ubuntu~yakkety_i386.deb",
		p64: "virtualbox-5.1_5.1.28-117968~Ubuntu~yakkety_amd64.deb",
	},
	"zesty": pkg{
		p32: "virtualbox-5.1_5.1.28-117968~Ubuntu~zesty_i386.deb",
		p64: "virtualbox-5.1_5.1.28-117968~Ubuntu~zesty_amd64.deb",
	},
	"el5": pkg{
		p32: "VirtualBox-5.1-5.1.28_117968_el5-1.i386.rpm",
		p64: "VirtualBox-5.1-5.1.28_117968_el5-1.x86_64.rpm",
	},
	"el6": pkg{
		p32: "VirtualBox-5.1-5.1.28_117968_el6-1.i386.rpm",
		p64: "VirtualBox-5.1-5.1.28_117968_el6-1.x86_64.rpm",
	},
	"el7": pkg{
		p32: "VirtualBox-5.1-5.1.28_117968_el7-1.i386.rpm",
		p64: "VirtualBox-5.1-5.1.28_117968_el7-1.x86_64.rpm",
	},
	"fedora22": pkg{
		p32: "VirtualBox-5.1-5.1.28_117968_fedora22-1.i686.rpm",
		p64: "VirtualBox-5.1-5.1.28_117968_fedora22-1.x86_64.rpm",
	},
	"fedora23": pkg{
		p32: "VirtualBox-5.1-5.1.28_117968_fedora22-1.i686.rpm",
		p64: "VirtualBox-5.1-5.1.28_117968_fedora22-1.x86_64.rpm",
	},
	"fedora24": pkg{
		p32: "VirtualBox-5.1-5.1.28_117968_fedora24-1.i686.rpm",
		p64: "VirtualBox-5.1-5.1.28_117968_fedora24-1.x86_64.rpm",
	},
	"fedora25": pkg{
		p32: "VirtualBox-5.1-5.1.28_117968_fedora25-1.i686.rpm",
		p64: "VirtualBox-5.1-5.1.28_117968_fedora25-1.x86_64.rpm",
	},
	"fedora26": pkg{
		p32: "VirtualBox-5.1-5.1.28_117968_fedora26-1.i686.rpm",
		p64: "VirtualBox-5.1-5.1.28_117968_fedora26-1.x86_64.rpm",
	},
	"generic": pkg{
		p32: "VirtualBox-5.1.28-117968-Linux_x86.run",
		p64: "VirtualBox-5.1.28-117968-Linux_amd64.run",
	},
}

const downloadsURL = "http://download.virtualbox.org/virtualbox/5.1.28/"

func platformVirtualBoxPackage() string {
	switch runtime.GOOS {
	case "windows":
		return "VirtualBox-5.1.28-117968-Win.exe"
	case "darwin":
		return "VirtualBox-5.1.28-117968-OSX.dmg"
	case "linux":
		code := distroCode()
		if len(code) > 0 {
			if runtime.GOARCH == "amd64" {
				return pkgs[code].p64
			}

			return pkgs[code].p32
		}
	default:
	}

	return ""
}

func platformVirtualBoxPackageURL() string {
	pkg := platformVirtualBoxPackage()
	if len(pkg) > 0 {
		return downloadsURL + pkg
	}

	return ""
}
