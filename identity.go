package main

import (
	"golang.org/x/net/context"

	"github.com/codedellemc/gocsi/csi"
)

func (s *sp) GetSupportedVersions(
	ctx context.Context,
	req *csi.GetSupportedVersionsRequest) (
	*csi.GetSupportedVersionsResponse, error) {

	return &csi.GetSupportedVersionsResponse{
		Reply: &csi.GetSupportedVersionsResponse_Result_{
			Result: &csi.GetSupportedVersionsResponse_Result{
				SupportedVersions: []*csi.Version{
					&csi.Version{
						Major: 0,
						Minor: 1,
						Patch: 0,
					},
				},
			},
		},
	}, nil
}

func (s *sp) GetPluginInfo(
	ctx context.Context,
	req *csi.GetPluginInfoRequest) (
	*csi.GetPluginInfoResponse, error) {

	return &csi.GetPluginInfoResponse{
		Reply: &csi.GetPluginInfoResponse_Result_{
			Result: &csi.GetPluginInfoResponse_Result{
				Name:          s.name,
				VendorVersion: "0.1.0",
				Manifest:      nil,
			},
		},
	}, nil
}
