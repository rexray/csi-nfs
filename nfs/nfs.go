// Package nfs provides utilities for mounting/unmounting NFS exported
// directories
package nfs

import (
	"encoding/base32"
	"fmt"
	"os/exec"
	"strings"
)

// Supported queries the underlying system to check if the required system
// executables are present
// If not, it returns an error
func Supported() error {
	if _, err := exec.Command("/bin/ls", "/sbin/mount.nfs").CombinedOutput(); err != nil {
		return fmt.Errorf("Required binary /sbin/mount.nfs is missing")
	}
	if _, err := exec.Command("/bin/ls", "/sbin/mount.nfs4").CombinedOutput(); err != nil {
		return fmt.Errorf("Required binary /sbin/mount.nfs4 is missing")
	}
	return nil
}

func GetName(volID string) string {
	volIDb := []byte(volID)

	name := base32.StdEncoding.EncodeToString(volIDb)

	return name
}

func DecodeName(name string) (string, string, error) {

	rawIDb, err := base32.StdEncoding.DecodeString(name)
	if err != nil {
		return "", "", err
	}

	fields := strings.Split(string(rawIDb), ":")
	if len(fields) != 2 {
		return "", "", fmt.Errorf("Unable to decode volume ID")
	}

	return fields[0], fields[1], nil

}
