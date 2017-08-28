package services

import "github.com/codedellemc/gocsi/csi"

const (
	SpName    = "csi-nfs"
	spVersion = "0.1.0"
)

var (
	CSIVersions = []*csi.Version{
		&csi.Version{
			Major: 0,
			Minor: 1,
			Patch: 0,
		},
	}
)

type StoragePlugin struct {
}

func (sp *StoragePlugin) Init() {}
