package kube

import (
	workflowsv1 "github.com/stackfoundation/sandbox/core/pkg/workflows/v1"
	log "github.com/stackfoundation/sandbox/log"
	"k8s.io/client-go/pkg/api/v1"
)

func createVolumeSource(volume workflowsv1.Volume) *v1.VolumeSource {
	var volumeSource *v1.VolumeSource

	if len(volume.HostPath) > 0 {
		volumeSource = &v1.VolumeSource{
			HostPath: &v1.HostPathVolumeSource{
				Path: volume.HostPath,
			},
		}

		log.Debugf("Mounting host path \"%v\" at \"%v\"", volume.HostPath, volume.MountPath)
	} else {
		volumeSource = &v1.VolumeSource{
			EmptyDir: &v1.EmptyDirVolumeSource{},
		}

		log.Debugf("Mounting volume \"%v\" at \"%v\"", volume.Name, volume.MountPath)
	}

	return volumeSource
}

func createVolumes(volumes []workflowsv1.Volume) ([]v1.VolumeMount, []v1.Volume) {
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

			volumeSource := createVolumeSource(volume)

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
