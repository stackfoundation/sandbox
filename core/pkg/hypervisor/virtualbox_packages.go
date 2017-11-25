package hypervisor

import (
	"bufio"
	"io"
	"io/ioutil"
	"runtime"
	"strings"

	"github.com/stackfoundation/sandbox/net/proxy"
)

const downloadsURL = "http://download.virtualbox.org/virtualbox/5.1.28/"
const md5SumsURL = "https://www.virtualbox.org/download/hashes/5.1.28/MD5SUMS"

// Retrieved on Oct 7, 2017 10:42 AM GMT
const cachedMD5Sums = `3a768b296704e6f84dbdabf922c54488 *Oracle_VM_VirtualBox_Extension_Pack-5.1.28-117968.vbox-extpack
3a768b296704e6f84dbdabf922c54488 *Oracle_VM_VirtualBox_Extension_Pack-5.1.28.vbox-extpack
8adc9fdc1added56b8b4b6b2f4dbd560 *SDKRef.pdf
7dc1d7f9d63cfa1a0c6a11ad5e221ce3 *UserManual.pdf
0cdb1fc2c52ad6aea0dfc3d0123264c9 *VBoxGuestAdditions_5.1.28.iso
3480367e688d27ce715134a18b735506 *VirtualBox-5.1-5.1.28_117968_el5-1.i386.rpm
4f9c3a422292955f3f9c22250fccc96c *VirtualBox-5.1-5.1.28_117968_el5-1.x86_64.rpm
7f9acd328bd46c11e632735a103729c9 *VirtualBox-5.1-5.1.28_117968_el6-1.i686.rpm
044e17af494194cc351cb6089f0e3fb1 *VirtualBox-5.1-5.1.28_117968_el6-1.x86_64.rpm
caa429c9a1d35e3bdecd0b2ddd6cbc38 *VirtualBox-5.1-5.1.28_117968_el7-1.x86_64.rpm
0f5430efe7c6a841eb9c508cf5b49a83 *VirtualBox-5.1-5.1.28_117968_fedora22-1.i686.rpm
ec41f81bfcef69576106ba269c0ad6f9 *VirtualBox-5.1-5.1.28_117968_fedora22-1.x86_64.rpm
fa64927db01cb6e8d3e1c0bc6f73cb0e *VirtualBox-5.1-5.1.28_117968_fedora24-1.i686.rpm
f2b07e4c68172493ec6d2ade92905a2b *VirtualBox-5.1-5.1.28_117968_fedora24-1.x86_64.rpm
ae09b4fb19194a69d16bb5731400d804 *VirtualBox-5.1-5.1.28_117968_fedora25-1.i686.rpm
9f72d075db162cfe596ae3a70687cbe4 *VirtualBox-5.1-5.1.28_117968_fedora25-1.x86_64.rpm
7dd88d11dd4bce3be76452a60fca85fb *VirtualBox-5.1-5.1.28_117968_fedora26-1.i686.rpm
d2855bb66eb519500779efc58a1fcb6a *VirtualBox-5.1-5.1.28_117968_fedora26-1.x86_64.rpm
12a7fdb0c4f87d6e77f1d643c68532b7 *VirtualBox-5.1-5.1.28_117968_openSUSE132-1.i586.rpm
78718d682e033bc6856e5162e9212387 *VirtualBox-5.1-5.1.28_117968_openSUSE132-1.x86_64.rpm
35ef4877738ebf6de7eb578eb855cd4a *VirtualBox-5.1.28-117968-Linux_amd64.run
ab30ff38aa16523c9def67262566b8e8 *VirtualBox-5.1.28-117968-Linux_x86.run
620b3bdf96b7afb9de56e2742d373568 *VirtualBox-5.1.28-117968-OSX.dmg
6405b8bfbcc6c04ceb903c8eac5ae0ba *VirtualBox-5.1.28-117968-SunOS.tar.gz
935f8590faac3f60c8b61abd4f27d0c7 *VirtualBox-5.1.28-117968-Win.exe
2e7350d5ac28ba3df4a778d168098ca5 *VirtualBox-5.1.28.tar.bz2
454c8c8f46cce0c9a20ee5dd5eb160a3 *VirtualBoxSDK-5.1.28-117968.zip
d5c29d02d791a308109e9414ed42b079 *virtualbox-5.1_5.1.28-117968~Debian~jessie_amd64.deb
7ea26e352e824b4148f4055e2ba8b5f5 *virtualbox-5.1_5.1.28-117968~Debian~jessie_i386.deb
3d80067953f53d61f5ea7c047fe96725 *virtualbox-5.1_5.1.28-117968~Debian~stretch_amd64.deb
8dd6a1e09a5ff9eb051954817fca618b *virtualbox-5.1_5.1.28-117968~Debian~stretch_i386.deb
26382cf976a171aa3d9cf5b647f1806d *virtualbox-5.1_5.1.28-117968~Debian~wheezy_amd64.deb
657edc7c9c358631b985d803bc7b23a3 *virtualbox-5.1_5.1.28-117968~Debian~wheezy_i386.deb
502fe26dc69d7903972ae1aec93ebef1 *virtualbox-5.1_5.1.28-117968~Ubuntu~precise_amd64.deb
5a3334d248db8ec434ee7266b68ddf9b *virtualbox-5.1_5.1.28-117968~Ubuntu~precise_i386.deb
674b5ae5280253e1a20a080baa7462f8 *virtualbox-5.1_5.1.28-117968~Ubuntu~trusty_amd64.deb
f8008adf8fd9196427ffca71a627dd2c *virtualbox-5.1_5.1.28-117968~Ubuntu~trusty_i386.deb
fc1c40ef0faa0aeae44a3c65dd260ef4 *virtualbox-5.1_5.1.28-117968~Ubuntu~wily_amd64.deb
772352c763d05f665065d487d0a17acf *virtualbox-5.1_5.1.28-117968~Ubuntu~wily_i386.deb
bfb377fb252b5994ed9fb6e9e81605c9 *virtualbox-5.1_5.1.28-117968~Ubuntu~xenial_amd64.deb
bf6c8957c9c25a9327bc221314dd1afa *virtualbox-5.1_5.1.28-117968~Ubuntu~xenial_i386.deb
d6e8aefbe941f408fc351e4f7404cbdb *virtualbox-5.1_5.1.28-117968~Ubuntu~yakkety_amd64.deb
d2549569fa71ed31a1a98c5c683c9fa7 *virtualbox-5.1_5.1.28-117968~Ubuntu~yakkety_i386.deb
3a0e59805d5165764010a586940f38c5 *virtualbox-5.1_5.1.28-117968~Ubuntu~zesty_amd64.deb
d1962db980fa760a20f2036be5ce7785 *virtualbox-5.1_5.1.28-117968~Ubuntu~zesty_i386.deb`

func is64BitPackage(pkg string) bool {
	return strings.Contains(pkg, "x86_64") || strings.Contains(pkg, "amd64")
}

func pkgMatchesPlatformArch(pkg string) bool {
	if runtime.GOARCH == "amd64" {
		if is64BitPackage(pkg) {
			return true
		}
	} else {
		if !is64BitPackage(pkg) {
			return true
		}
	}

	return false
}

func getMD5Sums() io.ReadCloser {
	response, err := proxy.ProxyCapableClient.Get(md5SumsURL)
	if err != nil {
		return ioutil.NopCloser(strings.NewReader(cachedMD5Sums))
	}

	return response.Body
}

func findMD5Line(code string, skip64bitCheck bool) (string, error) {
	md5Sums := getMD5Sums()
	defer md5Sums.Close()

	scanner := bufio.NewScanner(bufio.NewReader(md5Sums))
	for scanner.Scan() {
		line := scanner.Text()
		lower := strings.ToLower(line)

		if strings.Contains(lower, code) {
			if skip64bitCheck || pkgMatchesPlatformArch(lower) {
				return line, nil
			}
		}
	}

	return "", nil
}

func platformVirtualBoxPackage() (string, string) {
	var code string
	var skip64bitCheck bool
	switch runtime.GOOS {
	case "windows":
		code = "win"
		skip64bitCheck = true
	case "darwin":
		code = "osx"
		skip64bitCheck = true
	case "linux":
		code = distroCode()
	default:
	}

	md5Line, err := findMD5Line(code, skip64bitCheck)
	if err != nil {
		return "", ""
	}

	components := strings.Split(md5Line, " *")
	if len(components) == 2 {
		return downloadsURL + components[1], components[0]
	}

	return "", ""
}
