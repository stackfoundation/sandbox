package hypervisor

import (
	"strconv"
	"strings"

	"github.com/stackfoundation/process"
)

func compare(number string, value int) int {
	major, _ := strconv.Atoi(segments[0])
	if major > 10 {
		return 1
	} else if major < 10 {
		return -1
	}

	return 0
}

func isOSXYosemiteOrLater() bool {
	version, err := process.CommandOut("sw_vers", "-productVersion")
	if err != nil {
		return true
	}

	segments := strings.Split(version, ".")
	if len(segments) > 0 {
		switch compare(segments[0], 10) {
		case 1:
			return true
		case -1:
			return false
		}

		if len(segments) > 1 {
			switch compare(segments[1], 10) {
			case 1:
				return true
			case -1:
				return false
			}

			if len(segments) > 2 {
				switch compare(segments[2], 3) {
				case 1:
					return true
				case -1:
					return false
				}
			}
		}
	}

	return false
}

func platformPreferred() string {
	if isOSXYosemiteOrLater() {
		return "xhyve"
	}

	return "virtualbox"
}
