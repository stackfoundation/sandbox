package kube

import (
	"path"
	"path/filepath"
	"regexp"
	"strings"

	log "github.com/stackfoundation/core/pkg/log"
	workflowsv1 "github.com/stackfoundation/core/pkg/workflows/v1"
	"k8s.io/client-go/pkg/api/v1"
)

var driveLetterReplacement = regexp.MustCompile("^([a-zA-Z])\\:")

func lowercaseDriveLetter(text []byte) []byte {
	lowercase := strings.ToLower(string(text))
	return []byte("/" + lowercase[:len(lowercase)-1])
}

func createVolumeSource(projectRoot string, volume workflowsv1.Volume) *v1.VolumeSource {
	var volumeSource *v1.VolumeSource

	if len(volume.HostPath) > 0 {
		absoluteHostPath := path.Join(filepath.ToSlash(projectRoot), volume.HostPath)
		absoluteHostPath = string(driveLetterReplacement.ReplaceAllFunc(
			[]byte(absoluteHostPath),
			lowercaseDriveLetter))

		volumeSource = &v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: absoluteHostPath,
			},
		}

		log.Debugf("Mounting host path \"%v\" at \"%v\"", absoluteHostPath, volume.MountPath)
	} else {
		volumeSource = &v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		}

		log.Debugf("Mounting volume \"%v\" at \"%v\"", volume.Name, volume.MountPath)
	}

	return volumeSource
}

func createVolumes(projectRoot string, volumes []workflowsv1.Volume) ([]v1.VolumeMount, []v1.Volume) {
	numVolumes := len(volumes)
	if numVolumes > 0 {
		mounts := make([]v1.VolumeMount, 0, numVolumes)
		podVolumes := make([]v1.Volume, 0, numVolumes)

		for _, volume := range volumes {
			if len(volume.Name) < 1 {
				if len(volume.HostPath) > 0 {
					volume.Name = workflowsv1.GenerateVolumeName()
				} else {
					log.Debugf("No name was specified for non-host volume, ignoring")
					continue
				}
			}

			volumeSource := createVolumeSource(projectRoot, volume)

			podVolumes = append(podVolumes, v1.Volume{
				Name:         volume.Name,
				VolumeSource: *volumeSource,
			})

			mounts = append(mounts, v1.VolumeMount{
				Name:      volume.Name,
				MountPath: volume.MountPath,
			})
		}

		return mounts, podVolumes
	}

	return nil, nil
}
