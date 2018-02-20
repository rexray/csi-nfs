package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func (s *service) ControllerGetCapabilities(
	ctx context.Context,
	req *csi.ControllerGetCapabilitiesRequest) (
	*csi.ControllerGetCapabilitiesResponse, error) {

	return &csi.ControllerGetCapabilitiesResponse{}, nil
}

func (s *service) CreateVolume(
	ctx context.Context,
	req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) DeleteVolume(
	ctx context.Context,
	req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) ControllerPublishVolume(
	ctx context.Context,
	req *csi.ControllerPublishVolumeRequest) (
	*csi.ControllerPublishVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) ControllerUnpublishVolume(
	ctx context.Context,
	req *csi.ControllerUnpublishVolumeRequest) (
	*csi.ControllerUnpublishVolumeResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) ValidateVolumeCapabilities(
	ctx context.Context,
	req *csi.ValidateVolumeCapabilitiesRequest) (
	*csi.ValidateVolumeCapabilitiesResponse, error) {

	r := &csi.ValidateVolumeCapabilitiesResponse{
		Supported: true,
	}

	for _, c := range req.VolumeCapabilities {
		if t := c.GetBlock(); t != nil {
			r.Supported = false
			break
		}
		if t := c.GetMount(); t != nil {
			// If a filesystem is given, it must be NFS
			fs := t.GetFsType()
			if fs != "" && fs != "nfs" {
				r.Supported = false
				break
			}
			// TODO: Check mount flags
			//for _, f := range t.GetMountFlags() {}

		}
	}

	return r, nil
}

func (s *service) ListVolumes(
	ctx context.Context,
	req *csi.ListVolumesRequest) (*csi.ListVolumesResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) GetCapacity(
	ctx context.Context,
	req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) ControllerProbe(
	ctx context.Context,
	req *csi.ControllerProbeRequest) (*csi.ControllerProbeResponse, error) {

	return &csi.ControllerProbeResponse{}, nil
}
