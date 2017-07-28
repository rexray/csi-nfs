package main

import (
	"golang.org/x/net/context"

	"github.com/codenrhoden/csi-nfs-plugin/csi"
	"github.com/codenrhoden/csi-nfs-plugin/csiutils"
)

func (s *sp) ControllerGetCapabilities(
	ctx context.Context,
	in *csi.ControllerGetCapabilitiesRequest) (*csi.ControllerGetCapabilitiesResponse, error) {

	return &csi.ControllerGetCapabilitiesResponse{
		Reply: &csi.ControllerGetCapabilitiesResponse_Result_{
			Result: &csi.ControllerGetCapabilitiesResponse_Result{
				Capabilities: []*csi.ControllerServiceCapability{
					&csi.ControllerServiceCapability{
						Type: &csi.ControllerServiceCapability_VolumeCapability{
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

func (s *sp) CreateVolume(
	ctx context.Context,
	in *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	return csiutils.ErrCreateVolume(
		csi.Error_CreateVolumeError_CALL_NOT_IMPLEMENTED,
		"CreateVolume not valid for NFS"), nil
}

func (s *sp) DeleteVolume(
	ctx context.Context,
	in *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	return csiutils.ErrDeleteVolume(
		csi.Error_DeleteVolumeError_CALL_NOT_IMPLEMENTED,
		"DeleteVolume not valid for NFS"), nil
}

func (s *sp) ControllerPublishVolume(
	ctx context.Context,
	in *csi.ControllerPublishVolumeRequest) (*csi.ControllerPublishVolumeResponse, error) {

	return csiutils.ErrControllerPublishVolume(
		csi.Error_ControllerPublishVolumeError_CALL_NOT_IMPLEMENTED,
		"ControllerPublishVolume not valid for NFS"), nil
}

func (s *sp) ControllerUnpublishVolume(
	ctx context.Context,
	in *csi.ControllerUnpublishVolumeRequest) (*csi.ControllerUnpublishVolumeResponse, error) {

	return csiutils.ErrControllerUnpublishVolume(
		csi.Error_ControllerUnpublishVolumeError_CALL_NOT_IMPLEMENTED,
		"ControllerUnpublishVolume not valid for NFS"), nil
}

func (s *sp) ValidateVolumeCapabilities(
	ctx context.Context,
	in *csi.ValidateVolumeCapabilitiesRequest) (*csi.ValidateVolumeCapabilitiesResponse, error) {

	r := &csi.ValidateVolumeCapabilitiesResponse{
		Reply: &csi.ValidateVolumeCapabilitiesResponse_Result_{
			Result: &csi.ValidateVolumeCapabilitiesResponse_Result{
				Supported: true,
			},
		},
	}

	for _, c := range in.VolumeCapabilities {
		if t := c.GetBlock(); t != nil {
			r.GetResult().Supported = false
			break
		}
		if t := c.GetMount(); t != nil {
			// If a filesystem is given, it must be NFS
			fs := t.GetFsType()
			if fs != "" && fs != "nfs" {
				r.GetResult().Supported = false
				break
			}
			// TODO: Check mount flags
			//for _, f := range t.GetMountFlags() {}

		}
	}

	return r, nil
}

func (s *sp) ListVolumes(
	ctx context.Context,
	in *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {

	return csiutils.ErrListVolumes(
		csi.Error_GeneralError_UNDEFINED,
		"ListVolumes not implemented for NFS"), nil
}

func (s *sp) GetCapacity(
	ctx context.Context,
	in *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {

	return csiutils.ErrGetCapacity(
		csi.Error_GeneralError_UNDEFINED,
		"GetCapacity not implemented for NFS"), nil
}
