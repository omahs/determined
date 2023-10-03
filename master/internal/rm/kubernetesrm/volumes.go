package kubernetesrm

import (
	"fmt"
	"path"

	"github.com/determined-ai/determined/master/pkg/etc"

	"github.com/docker/docker/api/types/mount"

	k8sV1 "k8s.io/api/core/v1"

	"github.com/determined-ai/determined/master/pkg/cproto"
)

func configureMountPropagation(b *mount.BindOptions) *k8sV1.MountPropagationMode {
	if b == nil {
		return nil
	}

	switch b.Propagation {
	case mount.PropagationPrivate:
		p := k8sV1.MountPropagationNone
		return &p
	case mount.PropagationRSlave:
		p := k8sV1.MountPropagationHostToContainer
		return &p
	case mount.PropagationRShared:
		p := k8sV1.MountPropagationBidirectional
		return &p
	default:
		return nil
	}
}

func dockerMountsToHostVolumes(dockerMounts []mount.Mount) ([]k8sV1.VolumeMount, []k8sV1.Volume) {
	volumeMounts := make([]k8sV1.VolumeMount, 0, len(dockerMounts))
	volumes := make([]k8sV1.Volume, 0, len(dockerMounts))

	for idx, d := range dockerMounts {
		name := fmt.Sprintf("det-host-volume-%d", idx)
		volumeMounts = append(volumeMounts, k8sV1.VolumeMount{
			Name:             name,
			ReadOnly:         d.ReadOnly,
			MountPath:        d.Target,
			MountPropagation: configureMountPropagation(d.BindOptions),
		})
		volumes = append(volumes, k8sV1.Volume{
			Name: name,
			VolumeSource: k8sV1.VolumeSource{
				HostPath: &k8sV1.HostPathVolumeSource{
					Path: d.Source,
				},
			},
		})
	}

	return volumeMounts, volumes
}

func configureShmVolume(_ int64) (k8sV1.VolumeMount, k8sV1.Volume) {
	// Kubernetes does not support a native way to set shm size for
	// containers. The workaround for this is to create an emptyDir
	// volume and mount it to /dev/shm.
	volumeName := "det-shm-volume"
	volumeMount := k8sV1.VolumeMount{
		Name:      volumeName,
		ReadOnly:  false,
		MountPath: "/dev/shm",
	}
	volume := k8sV1.Volume{
		Name: volumeName,
		VolumeSource: k8sV1.VolumeSource{EmptyDir: &k8sV1.EmptyDirVolumeSource{
			Medium: k8sV1.StorageMediumMemory,
		}},
	}
	return volumeMount, volume
}

func configureAdditionalFilesVolumes(
	configMapName string,
	runArchives []cproto.RunArchive,
) ([]k8sV1.VolumeMount, []k8sV1.Volume) {
	var volumeMounts []k8sV1.VolumeMount
	var volumes []k8sV1.Volume

	// Add a volume for the archive itself from the config map to this.
	archiveVolumeName := "archive-volume"
	archiveVolume := k8sV1.Volume{
		Name: archiveVolumeName,
		VolumeSource: k8sV1.VolumeSource{
			ConfigMap: &k8sV1.ConfigMapVolumeSource{
				LocalObjectReference: k8sV1.LocalObjectReference{Name: configMapName},
			},
		},
	}
	volumes = append(volumes, archiveVolume)
	// TODO can we reduce to keyS? I think we are mounting everything??
	archiveVolumeMount := k8sV1.VolumeMount{
		Name:      archiveVolumeName,
		MountPath: initWrapperSrcPath,
		ReadOnly:  true,
	}
	volumeMounts = append(volumeMounts, archiveVolumeMount)

	// Add a volume for the wrapper script.
	// TODO we can likely move this?
	// Where is the other entrypoint??? It is in the volume right?
	entryPointVolumeName := "entrypoint-volume"
	var entryPointVolumeMode int32 = 0o555
	entryPointVolume := k8sV1.Volume{
		Name: entryPointVolumeName,
		VolumeSource: k8sV1.VolumeSource{
			ConfigMap: &k8sV1.ConfigMapVolumeSource{
				LocalObjectReference: k8sV1.LocalObjectReference{Name: configMapName},
				Items: []k8sV1.KeyToPath{{
					Key:  etc.K8WrapperResource,
					Path: etc.K8WrapperResource,
				}},
				DefaultMode: &entryPointVolumeMode,
			},
		},
	}
	volumes = append(volumes, entryPointVolume)
	entrypointVolumeMount := k8sV1.VolumeMount{
		Name:      entryPointVolumeName,
		MountPath: initWrapperWorkDir,
		ReadOnly:  true,
	}
	volumeMounts = append(volumeMounts, entrypointVolumeMount)

	// HMM will this work?
	additionalFilesVolumeName := "additional-files-volume"
	dstVolume := k8sV1.Volume{
		Name:         additionalFilesVolumeName,
		VolumeSource: k8sV1.VolumeSource{EmptyDir: &k8sV1.EmptyDirVolumeSource{}},
	}
	volumes = append(volumes, dstVolume)
	dstVolumeMount := k8sV1.VolumeMount{
		Name:      additionalFilesVolumeName,
		MountPath: initContainerTarDstPath,
		ReadOnly:  false,
	}
	volumeMounts = append(volumeMounts, dstVolumeMount)

	for idx, runArchive := range runArchives {
		for _, item := range runArchive.Archive {
			volumeMounts = append(volumeMounts, k8sV1.VolumeMount{
				Name:      additionalFilesVolumeName,
				MountPath: path.Join(runArchive.Path, item.Path),
				SubPath:   path.Join(fmt.Sprintf("%d", idx), item.Path),
			})
		}
	}

	return volumeMounts, volumes
}
