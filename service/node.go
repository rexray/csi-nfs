package service

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/container-storage-interface/spec/lib/go/csi"

	"github.com/thecodeteam/csi-nfs/nfs"
)

var (
	emptyNodePubResp   = &csi.NodePublishVolumeResponse{}
	emptyNodeUnpubResp = &csi.NodeUnpublishVolumeResponse{}
)

func (s *service) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	if err := publishVolume(req, s.privDir); err != nil {
		return nil, err
	}

	return emptyNodePubResp, nil
}

func (s *service) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	if err := unpublishVolume(req, s.privDir); err != nil {
		return nil, err
	}

	return emptyNodeUnpubResp, nil
}

func (s *service) GetNodeID(
	ctx context.Context,
	req *csi.GetNodeIDRequest) (*csi.GetNodeIDResponse, error) {

	return nil, status.Error(codes.Unimplemented, "")
}

func (s *service) NodeProbe(
	ctx context.Context,
	req *csi.NodeProbeRequest) (*csi.NodeProbeResponse, error) {

	if err := nfs.Supported(); err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, err.Error())
	}

	return &csi.NodeProbeResponse{}, nil
}

func (s *service) NodeGetCapabilities(
	ctx context.Context,
	req *csi.NodeGetCapabilitiesRequest) (
	*csi.NodeGetCapabilitiesResponse, error) {

	return &csi.NodeGetCapabilitiesResponse{}, nil
}

func safeVolID(volID string) bool {
	return true
}
