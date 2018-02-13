package service

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/container-storage-interface/spec/lib/go/csi"
	log "github.com/sirupsen/logrus"
	"github.com/thecodeteam/csi-nfs/nfs"
	"github.com/thecodeteam/gofsutil"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// publishVolume uses the parameters in req to bindmount the NFS export
// to the requested target path. A private mount is performed first
// within the given privDir directory.
//
// publishVolume only handles Mount access types
func publishVolume(
	req *csi.NodePublishVolumeRequest,
	privDir string) error {

	id := req.GetVolumeId()

	target := req.GetTargetPath()
	if target == "" {
		return status.Error(codes.InvalidArgument,
			"target_path is required")
	}

	ro := req.GetReadonly()

	volCap := req.GetVolumeCapability()
	if volCap == nil {
		return status.Errorf(codes.InvalidArgument,
			"volume capability required")
	}

	accMode := volCap.GetAccessMode()
	if accMode == nil {
		return status.Errorf(codes.InvalidArgument,
			"access mode required")
	}

	// make sure privDir exists and is a directory
	if _, err := mkdir(privDir); err != nil {
		return err
	}

	// make sure target is created
	tgtStat, err := os.Stat(target)
	if err != nil {
		if os.IsNotExist(err) {
			return status.Errorf(codes.FailedPrecondition,
				"publish target: %s not pre-created", target)
		}
		return status.Errorf(codes.Internal,
			"failed to stat target, err: %s", err.Error())
	}

	// check that target is a directory
	if !tgtStat.IsDir() {
		return status.Errorf(codes.FailedPrecondition,
			"target: %s must be a directory", target)
	}

	mntVol := volCap.GetMount()
	if mntVol == nil {
		return status.Error(codes.InvalidArgument,
			"access type must be mount")
	}

	// Path to mount device to
	privTgt := getPrivateMountPoint(privDir, id)

	f := log.Fields{
		"id":           id,
		"target":       target,
		"privateMount": privTgt,
	}

	ctx := context.Background()

	// Check if device is already mounted
	devMnts, err := gofsutil.GetDevMounts(ctx, id)
	if err != nil {
		return status.Errorf(codes.Internal,
			"could not reliably determine existing mount status: %s",
			err.Error())
	}

	if len(devMnts) == 0 {
		// Device isn't mounted anywhere, do the private mount
		log.WithFields(f).Debug("attempting mount to private area")

		// Make sure private mount point exists
		created, err := mkdir(privTgt)
		if err != nil {
			return status.Errorf(codes.Internal,
				"Unable to create private mount point: %s",
				err.Error())
		}
		if !created {
			log.WithFields(f).Debug("private mount target already exists")

			// The place where our device is supposed to be mounted
			// already exists, but we also know that our device is not mounted anywhere
			// Either something didn't clean up correctly, or something else is mounted
			// If the private mount is not in use, it's okay to re-use it. But make sure
			// it's not in use first

			mnts, err := gofsutil.GetMounts(ctx)
			if err != nil {
				return status.Errorf(codes.Internal,
					"could not reliably determine existing mount status: %s",
					err.Error())
			}
			for _, m := range mnts {
				if m.Path == privTgt {
					log.WithFields(f).WithField("mountedDevice", id).Error(
						"mount point already in use by device")
					return status.Error(codes.Internal,
						"Unable to use private mount point")
				}
			}
		}

		fs := mntVol.GetFsType()
		mntFlags := mntVol.GetMountFlags()

		if err := handlePrivFSMount(
			ctx, accMode, id, mntFlags, fs, privTgt); err != nil {
			return err
		}

	} else {
		// Device is already mounted. Need to ensure that it is already
		// mounted to the expected private mount, with correct rw/ro perms
		mounted := false
		for _, m := range devMnts {
			if m.Path == privTgt {
				mounted = true
				rwo := "rw"
				if ro {
					rwo = "ro"
				}
				if contains(m.Opts, rwo) {
					log.WithFields(f).Debug(
						"private mount already in place")
					break
				} else {
					return status.Error(codes.InvalidArgument,
						"access mode conflicts with existing mounts")
				}
			}
		}
		if !mounted {
			return status.Error(codes.Internal,
				"device already in use and mounted elsewhere")
		}
	}

	// Private mount in place, now bind mount to target path

	// If mounts already existed for this device, check if mount to
	// target path was already there
	if len(devMnts) > 0 {
		for _, m := range devMnts {
			if m.Path == target {
				// volume already published to target
				// if mount options look good, do nothing
				rwo := "rw"
				if accMode.GetMode() == csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY {
					rwo = "ro"
				}
				if !contains(m.Opts, rwo) {
					return status.Error(codes.Internal,
						"volume previously published with different options")

				}
				// Existing mount satisfies request
				log.WithFields(f).Debug("volume already published to target")
				return nil
			}
		}

	}

	mntFlags := mntVol.GetMountFlags()
	if accMode.GetMode() == csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY {
		mntFlags = append(mntFlags, "ro")
	}
	if err := gofsutil.BindMount(ctx, privTgt, target, mntFlags...); err != nil {
		return status.Errorf(codes.Internal,
			"error publish volume to target path: %s",
			err.Error())
	}

	return nil
}

func handlePrivFSMount(
	ctx context.Context,
	accMode *csi.VolumeCapability_AccessMode,
	export string,
	mntFlags []string,
	fs, privTgt string) error {

	// If read-only access mode, we don't allow formatting
	if accMode.GetMode() == csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY {
		mntFlags = append(mntFlags, "ro")
	} else if accMode.GetMode() == csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER {
	} else {
		return status.Error(codes.Internal, "Invalid access mode")
	}
	if err := gofsutil.Mount(ctx, export, privTgt, fs, mntFlags...); err != nil {
		return status.Errorf(codes.Internal,
			"error performing private mount: %s",
			err.Error())
	}
	return nil
}

func getPrivateMountPoint(privDir, volID string) string {
	name := nfs.GetName(volID)
	return filepath.Join(privDir, name)
}

func contains(list []string, item string) bool {
	for _, x := range list {
		if x == item {
			return true
		}
	}
	return false
}

// mkdir creates the directory specified by path if needed.
// return pair is a bool flag of whether dir was created, and an error
func mkdir(path string) (bool, error) {
	st, err := os.Stat(path)
	if os.IsNotExist(err) {
		if err := os.Mkdir(path, 0755); err != nil {
			log.WithField("dir", path).WithError(
				err).Error("Unable to create dir")
			return false, err
		}
		log.WithField("path", path).Debug("created directory")
		return true, nil
	}
	if !st.IsDir() {
		return false, fmt.Errorf("existing path is not a directory")
	}
	return false, nil
}

// unpublishVolume removes the bind mount to the target path, and also removes
// the mount to the private mount directory if the volume is no longer in use.
// It determines this by checking to see if the volume is mounted anywhere else
// other than the private mount.
func unpublishVolume(
	req *csi.NodeUnpublishVolumeRequest,
	privDir string) error {

	ctx := context.Background()
	id := req.VolumeId

	target := req.TargetPath
	if target == "" {
		return status.Error(codes.InvalidArgument,
			"target_path is required")
	}

	// Path to mount device to
	privTgt := getPrivateMountPoint(privDir, id)

	mnts, err := gofsutil.GetDevMounts(ctx, id)
	if err != nil {
		return status.Errorf(codes.Internal,
			"could not reliably determine existing mount status: %s",
			err.Error())
	}

	tgtMnt := false
	privMnt := false
	for _, m := range mnts {
		log.WithField("mount", m).Debug("mount entry")
		if m.Path == privTgt {
			privMnt = true
		} else if m.Path == target {
			tgtMnt = true
		}
	}

	if tgtMnt {
		if err := gofsutil.Unmount(ctx, target); err != nil {
			return status.Errorf(codes.Internal,
				"Error unmounting target: %s", err.Error())
		}
	}

	if privMnt {
		if err := unmountPrivMount(ctx, id, privTgt); err != nil {
			return status.Errorf(codes.Internal,
				"Error unmounting private mount: %s", err.Error())
		}
	}

	return nil
}

func unmountPrivMount(
	ctx context.Context,
	id, target string) error {

	mnts, err := gofsutil.GetDevMounts(ctx, id)
	if err != nil {
		return err
	}

	// remove private mount if we can
	if len(mnts) == 1 && mnts[0].Path == target {
		if err := gofsutil.Unmount(ctx, target); err != nil {
			return err
		}
		log.WithField("directory", target).Debug(
			"removing directory")
		os.Remove(target)
	}
	return nil
}
