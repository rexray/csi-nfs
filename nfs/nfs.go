// Package nfs provides utilities for mounting/unmounting NFS exported
// directories
package nfs

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

const (
	mountCmd   = "mount"
	unmountCmd = "umount"
)

// Supported queries the underlying system to check if the required system
// executables are present
// If not, it returns an error
func Supported() error {
	switch runtime.GOOS {
	case "linux":
		if _, err := exec.Command("/bin/ls", "/sbin/mount.nfs").CombinedOutput(); err != nil {
			return fmt.Errorf("Required binary /sbin/mount.nfs is missing")
		}
		if _, err := exec.Command("/bin/ls", "/sbin/mount.nfs4").CombinedOutput(); err != nil {
			return fmt.Errorf("Required binary /sbin/mount.nfs4 is missing")
		}
		return nil
	case "darwin":
		if _, err := exec.Command("/bin/ls", "/sbin/mount_nfs").CombinedOutput(); err != nil {
			return fmt.Errorf("Required binary /sbin/mount_nfs is missing")
		}
	}
	return nil
}

// Mount mounts the requested exported source path on the remote server to the
// given local directory
func Mount(source string, path string, opts []string) error {

	mArgs := makeMountArgs(source, path, "nfs", opts)

	fields := log.Fields{
		"command": mountCmd,
		"args":    mArgs,
	}

	log.WithFields(fields).Info("mounting volume")
	cmd := exec.Command(mountCmd, mArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fields["output"] = string(output)
		log.WithFields(fields).WithError(err).Error("mount failed")
		return fmt.Errorf(
			"mount failed: %v\nmounting command: %s\nmounting Args: %v\noutput: %s",
			err, mountCmd, mArgs, string(output))
	}
	return err
}

// Unmount unmounts the requested path
func Unmount(path string) error {

	fields := log.Fields{
		"command": unmountCmd,
		"path":    path,
	}
	log.WithFields(fields).Info("unmounting volume")
	cmd := exec.Command(unmountCmd, path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fields["output"] = string(output)
		log.WithFields(fields).WithError(err).Error("unmount failed")
		return fmt.Errorf(
			"unmount failed: %v\nunmounting arguments: %s\noutput: %s",
			err, path, string(output))
	}
	return nil
}

// makeMountArgs makes the arguments to the mount(8) command.
func makeMountArgs(source, target, fstype string, options []string) []string {
	// Build mount command as follows:
	//   mount [-t $fstype] [-o $options] [$source] $target
	mountArgs := []string{}
	if len(fstype) > 0 {
		mountArgs = append(mountArgs, "-t", fstype)
	}
	if len(options) > 0 {
		mountArgs = append(mountArgs, "-o", strings.Join(options, ","))
	}
	if len(source) > 0 {
		mountArgs = append(mountArgs, source)
	}
	mountArgs = append(mountArgs, target)

	return mountArgs
}
