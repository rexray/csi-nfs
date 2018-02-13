package service

import (
	"context"
	"net"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/thecodeteam/gocsi"
	csictx "github.com/thecodeteam/gocsi/context"

	"github.com/thecodeteam/csi-nfs/core"
)

const (
	// Name is the name of this CSI SP.
	Name = "com.thecodeteam.csi-nfs"

	// SupportedVersions is a list of the CSI versions this SP supports.
	SupportedVersions = "0.1.0"

	defaultPrivDir = "/dev/disk/csi-nfs-private"
)

// Manifest is the SP's manifest.
var Manifest = map[string]string{
	"url":    "https://github.com/thecodeteam/csi-nfs",
	"semver": core.SemVer,
	"commit": core.CommitSha32,
	"formed": core.CommitTime.Format(time.RFC1123),
}

// Service is a CSI SP
type Service interface {
	csi.ControllerServer
	csi.IdentityServer
	csi.NodeServer
	BeforeServe(context.Context, *gocsi.StoragePlugin, net.Listener) error
}

type service struct {
	privDir string
}

// New returns a new Service.
func New() Service {
	return &service{}
}

func (s *service) BeforeServe(
	ctx context.Context, sp *gocsi.StoragePlugin, lis net.Listener) error {

	defer func() {
		fields := map[string]interface{}{
			"privatedir": s.privDir,
		}

		log.WithFields(fields).Infof("configured %s", Name)
	}()

	if pd, ok := csictx.LookupEnv(ctx, gocsi.EnvVarPrivateMountDir); ok {
		s.privDir = pd
	}
	if s.privDir == "" {
		s.privDir = defaultPrivDir
	}

	return nil
}
