package services

import (
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/thecodeteam/gocsi/csi"
)

const (
	// SpName holds the name of the Storage Plugin / driver
	Name    = "csi-nfs"
	Version = "0.1.0"

	debugEnvVar    = "X_CSI_NFS_DEBUG"
	mountDirEnvVar = "X_CSI_NFS_MOUNTDIR"
	defaultDir     = "/dev/csi-nfs-mounts"
)

var (
	// CSIVersions holds a slice of compatible CSI spec versions
	CSIVersions = []*csi.Version{
		&csi.Version{
			Major: 0,
			Minor: 0,
			Patch: 0,
		},
	}
)

// Service is the CSI Network File System (NFS) service provider.
type Service interface {
	csi.ControllerServer
	csi.IdentityServer
	csi.NodeServer
}

// storagePlugin contains parameters for the plugin
type storagePlugin struct {
	privDir string
}

// New returns a new Service
func New() Service {

	sp := &storagePlugin{
		privDir: defaultDir,
	}
	if md := os.Getenv(mountDirEnvVar); md != "" {
		sp.privDir = md
	}
	log.WithFields(map[string]interface{}{
		"privDir": sp.privDir,
	}).Info("created new " + Name + " service")

	return sp
}
