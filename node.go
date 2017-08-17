package main

import (
	"golang.org/x/net/context"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"

	"github.com/codedellemc/csi-nfs/nfs"
)

func (s *sp) NodePublishVolume(
	ctx context.Context,
	in *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {

	idm := in.GetVolumeId().GetValues()
	target := in.GetTargetPath()
	ro := in.GetReadonly()

	opts := make([]string, 0)
	if ro {
		opts = append(opts, "ro")
	}

	host, ok := idm["host"]
	if !ok {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_INVALID_VOLUME_ID,
			"host key missing from volumeID"), nil
	}

	export, ok := idm["export"]
	if !ok {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_INVALID_VOLUME_ID,
			"export key missing from volumeID"), nil
	}

	src := host + ":" + export

	if err := nfs.Mount(src, target, opts); err != nil {
		return gocsi.ErrNodePublishVolume(
			csi.Error_NodePublishVolumeError_MOUNT_ERROR,
			err.Error()), nil
	}

	return &csi.NodePublishVolumeResponse{
		Reply: &csi.NodePublishVolumeResponse_Result_{
			Result: &csi.NodePublishVolumeResponse_Result{},
		},
	}, nil
}

func (s *sp) NodeUnpublishVolume(
	ctx context.Context,
	in *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {

	target := in.GetTargetPath()

	if err := nfs.Unmount(target); err != nil {
		return gocsi.ErrNodeUnpublishVolume(
			csi.Error_NodeUnpublishVolumeError_UNMOUNT_ERROR,
			err.Error()), nil
	}

	return &csi.NodeUnpublishVolumeResponse{
		Reply: &csi.NodeUnpublishVolumeResponse_Result_{
			Result: &csi.NodeUnpublishVolumeResponse_Result{},
		},
	}, nil
}

func (s *sp) GetNodeID(
	ctx context.Context,
	in *csi.GetNodeIDRequest) (*csi.GetNodeIDResponse, error) {

	return &csi.GetNodeIDResponse{
		Reply: &csi.GetNodeIDResponse_Result_{
			// Return nil ID because it's not used by the
			// controller
			Result: &csi.GetNodeIDResponse_Result{},
		},
	}, nil
}

func (s *sp) ProbeNode(
	ctx context.Context,
	in *csi.ProbeNodeRequest) (*csi.ProbeNodeResponse, error) {

	if err := nfs.Supported(); err != nil {
		return gocsi.ErrProbeNode(
			csi.Error_ProbeNodeError_MISSING_REQUIRED_HOST_DEPENDENCY,
			err.Error()), nil
	}

	return &csi.ProbeNodeResponse{
		Reply: &csi.ProbeNodeResponse_Result_{
			Result: &csi.ProbeNodeResponse_Result{},
		},
	}, nil
}

func (s *sp) NodeGetCapabilities(
	ctx context.Context,
	in *csi.NodeGetCapabilitiesRequest) (*csi.NodeGetCapabilitiesResponse, error) {

	return &csi.NodeGetCapabilitiesResponse{
		Reply: &csi.NodeGetCapabilitiesResponse_Result_{
			Result: &csi.NodeGetCapabilitiesResponse_Result{
				Capabilities: []*csi.NodeServiceCapability{
					&csi.NodeServiceCapability{
						Type: &csi.NodeServiceCapability_VolumeCapability{
							VolumeCapability: &csi.VolumeCapability{
								Value: &csi.VolumeCapability_Mount{
									Mount: &csi.VolumeCapability_MountVolume{
										FsType: "nfs",
									},
								},
							},
						},
					},
				},
			},
		},
	}, nil
}
