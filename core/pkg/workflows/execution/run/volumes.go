package run

import (
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/stackfoundation/core/pkg/workflows/v1"
)

var driveLetterReplacement = regexp.MustCompile("^([a-zA-Z])\\:")

func lowercaseDriveLetter(text []byte) []byte {
	lowercase := strings.ToLower(string(text))
	return []byte("/" + lowercase[:len(lowercase)-1])
}

func normalizeVolumePaths(projectRoot string, volumes []v1.Volume) []v1.Volume {
	if len(volumes) > 0 {
		modified := volumes[:0]
		for _, volume := range volumes {
			if len(volume.HostPath) > 0 {
				absoluteHostPath := path.Join(filepath.ToSlash(projectRoot), volume.HostPath)
				absoluteHostPath = string(driveLetterReplacement.ReplaceAllFunc(
					[]byte(absoluteHostPath),
					lowercaseDriveLetter))

				modified = append(modified, v1.Volume{
					HostPath:  absoluteHostPath,
					MountPath: volume.MountPath,
				})
			}
		}

		return modified
	}

	return volumes
}
