package main

import (
	"context"

	"github.com/thecodeteam/gocsi"

	"github.com/thecodeteam/csi-nfs/provider"
	"github.com/thecodeteam/csi-nfs/service"
)

// main is ignored when this package is built as a go plug-in.
func main() {
	gocsi.Run(
		context.Background(),
		service.Name,
		"An NFS Container Storage Interface (CSI) Plugin",
		usage,
		provider.New())
}

const usage = ``
