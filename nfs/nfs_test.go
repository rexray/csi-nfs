package nfs

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/thecodeteam/gofsutil"
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

	ctx := context.Background()

	err = gofsutil.Mount(ctx, nfsHost+":"+nfsPath, mntPath, "nfs", "")
	if err != nil {
		t.Fatal(err)
	}

	err = gofsutil.Unmount(ctx, mntPath)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetName(t *testing.T) {
	tests := []struct {
		host   string
		export string
		result string
	}{
		{
			host:   "localhost",
			export: "/data",
			result: "NRXWGYLMNBXXG5B2F5SGC5DB",
		},
		{
			host:   "192.168.0.1",
			export: "/data",
			result: "GE4TELRRGY4C4MBOGE5C6ZDBORQQ====",
		},
		{
			host:   "myserver.internal.org",
			export: "/data",
			result: "NV4XGZLSOZSXELTJNZ2GK4TOMFWC433SM45C6ZDBORQQ====",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(st *testing.T) {
			st.Parallel()
			id := GetName(tt.host + ":" + tt.export)
			if id != tt.result {
				t.Errorf("Encoding of NFS export to ID incorrect, got: %s want: %s",
					id, tt.result)
			}
		})
	}

}

func TestDecodeName(t *testing.T) {
	tests := []struct {
		id     string
		host   string
		export string
	}{
		{
			id:     "NRXWGYLMNBXXG5B2F5SGC5DB",
			host:   "localhost",
			export: "/data",
		},
		{
			id:     "GE4TELRRGY4C4MBOGE5C6ZDBORQQ====",
			host:   "192.168.0.1",
			export: "/data",
		},
		{
			id:     "NV4XGZLSOZSXELTJNZ2GK4TOMFWC433SM45C6ZDBORQQ====",
			host:   "myserver.internal.org",
			export: "/data",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run("", func(st *testing.T) {
			st.Parallel()
			host, export, err := DecodeName(tt.id)
			if err != nil {
				t.Error(err)
			}
			if host != tt.host {
				t.Errorf("Decoding of NFS ID host incorrect, got: %s want: %s",
					host, tt.host)
			}
			if export != tt.export {
				t.Errorf("Decoding of NFS ID export incorrect, got: %s want: %s",
					export, tt.export)
			}
		})
	}

}
