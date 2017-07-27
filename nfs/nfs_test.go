package nfs

import (
	"io/ioutil"
	"os"
	"strings"
	"testing"
)

var (
	nfsHost = os.Getenv("NFS_HOST")
	nfsPath = os.Getenv("NFS_PATH")
	mntPath string
)

func TestMountUnmount(t *testing.T) {
	if len(nfsHost) == 0 || len(nfsPath) == 0 {
		t.Skip("nfs server details not set")
	}

	mntPath, err := ioutil.TempDir("", "csinfsplugin")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v\n", err)
	}
	defer os.Remove(mntPath)

	err = Mount(nfsHost+":"+nfsPath, mntPath, []string{})
	if err != nil {
		t.Fatal(err)
	}

	err = Unmount(mntPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMountArgs(t *testing.T) {
	tests := []struct {
		src    string
		tgt    string
		fst    string
		opts   []string
		result string
	}{
		{
			src:    "localhost:/data",
			tgt:    "/mnt",
			fst:    "nfs",
			result: "-t nfs localhost:/data /mnt",
		},
		{
			src:    "localhost:/data",
			tgt:    "/mnt",
			result: "localhost:/data /mnt",
		},
		{
			src:    "localhost:/data",
			tgt:    "/mnt",
			fst:    "nfs",
			opts:   []string{"tcp", "vers=4"},
			result: "-t nfs -o tcp,vers=4 localhost:/data /mnt",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(st *testing.T) {
			st.Parallel()
			opts := makeMountArgs(tt.src, tt.tgt, tt.fst, tt.opts)
			optsStr := strings.Join(opts, " ")
			if optsStr != tt.result {
				t.Errorf("Formatting of mount args incorrect, got: %s want: %s",
					optsStr, tt.result)
			}
		})
	}
}
