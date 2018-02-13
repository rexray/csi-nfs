package service

import (
	"context"
	"os"
	"path/filepath"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	log "github.com/sirupsen/logrus"

	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/thecodeteam/gofsutil"

	"github.com/thecodeteam/csi-nfs/nfs"
)

func (s *service) NodePublishVolume(
	ctx context.Context,
	req *csi.NodePublishVolumeRequest) (
	*csi.NodePublishVolumeResponse, error) {

	target := req.GetTargetPath()
	ro := req.GetReadonly()
	vc := req.GetVolumeCapability()
	am := vc.GetAccessMode()

	mv := vc.GetMount()
	if mv == nil {
		return nil, status.Errorf(codes.InvalidArgument,
			"unsupported volume type for NFS")
	}

	mf := mv.GetMountFlags()

	if m := am.GetMode(); m == csi.VolumeCapability_AccessMode_UNKNOWN {
		return nil, status.Errorf(codes.InvalidArgument,
			"invalid access mode")
	}

	return s.handleMount(req.VolumeId, target, mf, ro, am)
}

func (s *service) NodeUnpublishVolume(
	ctx context.Context,
	req *csi.NodeUnpublishVolumeRequest) (
	*csi.NodeUnpublishVolumeResponse, error) {

	target := req.TargetPath

	// check to see if volume is really mounted at target
	mnts, err := gofsutil.GetDevMounts(ctx, req.VolumeId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	if len(mnts) > 0 {
		// device is mounted somewhere. could be target, other targets,
		// or private mount
		var (
			idx       int
			m         gofsutil.Info
			unmounted = false
		)
		for idx, m = range mnts {
			if m.Path == target {
				if err := gofsutil.Unmount(ctx, target); err != nil {
					return nil, status.Errorf(
						codes.Internal, err.Error())
				}
				unmounted = true
				break
			}
		}
		if unmounted {
			mnts = append(mnts[:idx], mnts[idx+1:]...)
		}
	}

	// remove private mount if we can
	privTgt := s.getPrivateMountPoint(req.VolumeId)
	if len(mnts) == 1 && mnts[0].Path == privTgt {
		if err := gofsutil.Unmount(ctx, privTgt); err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
		os.Remove(privTgt)
	}

	return &csi.NodeUnpublishVolumeResponse{}, nil
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

// mkdir creates the directory specified by path if needed.
// return pair is a bool flag of whether dir was created, and an error
func mkdir(path string) (bool, error) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			log.WithField("dir", path).WithError(
				err).Error("Unable to create dir")
			return false, err
		}
		log.WithField("path", path).Debug("created directory")
		return true, nil
	}
	return false, nil
}

func (s *service) handleMount(
	volID string,
	target string,
	mf []string,
	ro bool,
	am *csi.VolumeCapability_AccessMode) (*csi.NodePublishVolumeResponse, error) {

	// Make sure privDir exists
	if _, err := mkdir(s.privDir); err != nil {
		return nil, status.Errorf(codes.Internal,
			"Unable to create private mount dir")
	}

	if !safeVolID(volID) {
		return nil, status.Error(codes.InvalidArgument,
			"volumeID not a valid NFS adress")
	}

	// Path to mount device to
	privTgt := s.getPrivateMountPoint(volID)

	f := log.Fields{
		"volume":       volID,
		"target":       target,
		"privateMount": privTgt,
	}

	ctx := context.Background()
	// Check if device is already mounted
	mnts, err := gofsutil.GetDevMounts(ctx, volID)
	if err != nil {
		return nil, status.Errorf(codes.Internal,
			"could not reliably determine existing mount status")
	}

	mode := am.GetMode()

	if len(mnts) == 0 {
		// Device isn't mounted anywhere, do the private mount
		log.WithFields(f).Debug("attempting mount to private area")

		// Make sure private mount point exists
		created, err := mkdir(privTgt)
		if err != nil {
			return nil, status.Errorf(codes.Internal,
				"Unable to create private mount point")
		}
		if !created {
			log.WithFields(f).Debug("private mount target already exists")

			// The place where our device is supposed to be mounted
			// already exists, but we also know that our device is not mounted anywhere
			// Either something didn't clean up correctly, or something else is mounted
			// If the directory is not in use, it's okay to re-use it. But make sure
			// it's not in use first

			for _, m := range mnts {
				if m.Path == privTgt {
					log.WithFields(f).WithField("mountedDevice", m.Device).Error(
						"mount point already in use by device")
					return nil, status.Errorf(codes.Internal,
						"Unable to use private mount point")
				}
			}
		}

		// If read-only access mode, we don't allow formatting
		if mode == csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY ||
			mode == csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY {
			mf = append(mf, "ro")
		}

		if err := gofsutil.Mount(ctx, volID, privTgt, "nfs", mf...); err != nil {
			return nil, status.Errorf(codes.Internal, err.Error())
		}
	} else {
		// Device is already mounted. Need to ensure that it is already
		// mounted to the expected private mount, with correct rw/ro perms
		mounted := false
		for _, m := range mnts {
			if m.Path == privTgt {
				mounted = true
				rwo := "rw"
				if mode == csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY ||
					mode == csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY {
					rwo = "ro"
				}
				if contains(m.Opts, rwo) {
					break
				} else {
					return nil, status.Errorf(codes.Internal,
						"access mode conflicts with existing mounts")
				}
			}
		}
		if !mounted {
			return nil, status.Errorf(codes.Internal,
				"device in use by external entity")
		}
	}

	// Private mount in place, now bind mount to target path

	// If mounts already existed for this device, check if mount to
	// target path was already there
	if len(mnts) > 0 {
		for _, m := range mnts {
			if m.Path == target {
				// volume already published to target
				// if mount options look good, do nothing
				rwo := "rw"
				if ro {
					rwo = "ro"
				}
				if !contains(m.Opts, rwo) {
					return nil, status.Errorf(codes.Internal,
						"volume previously published with different options")

				}
				// Existing mount satisfied requested
				return &csi.NodePublishVolumeResponse{}, nil
			}
		}

	}

	if ro {
		mf = append(mf, "ro")
	}
	mf = append(mf, "bind")
	if err := gofsutil.Mount(ctx, privTgt, target, "", mf...); err != nil {
		//if err := SafeUnmnt(privTgt); err != nil {
		//	log.WithFields(f).WithError(err).Error(
		//		"Unable to umount from private dir")
		//}
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	// Mount successful
	return &csi.NodePublishVolumeResponse{}, nil
}

func (s *service) getPrivateMountPoint(volID string) string {
	name := nfs.GetName(volID)
	return filepath.Join(s.privDir, name)
}

func contains(list []string, item string) bool {
	for _, x := range list {
		if x == item {
			return true
		}
	}
	return false
}

func safeVolID(volID string) bool {
	return true
}
