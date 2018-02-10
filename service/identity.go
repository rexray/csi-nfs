package service

import (
	"context"

	"github.com/container-storage-interface/spec/lib/go/csi"
)

func (s *storagePlugin) GetSupportedVersions(
	ctx context.Context,
	req *csi.GetSupportedVersionsRequest) (
	*csi.GetSupportedVersionsResponse, error) {

	return &csi.GetSupportedVersionsResponse{
		SupportedVersions: CSIVersions,
	}, nil
}

func (s *storagePlugin) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {

	return &csi.GetPluginInfoResponse{
		Name:          Name,
		VendorVersion: Version,
		Manifest:      nil,
	}, nil
}
