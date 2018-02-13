package service

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func (s *service) GetSupportedVersions(
	ctx context.Context,
	req *csi.GetSupportedVersionsRequest) (
	*csi.GetSupportedVersionsResponse, error) {

	// Defer to GoCSI interceptor
	return nil, nil
}

func (s *service) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {

	return &csi.GetPluginInfoResponse{
		Name:          Name,
		VendorVersion: VendorVersion,
		Manifest:      nil,
	}, nil
}
