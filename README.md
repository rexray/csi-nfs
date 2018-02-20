# CSI plugin for NFS [![Build Status](http://travis-ci.org/thecodeteam/csi-nfs.svg?branch=master)]

## Description
CSI-NFS is a Container Storage Interface
([CSI](https://github.com/container-storage-interface/spec)) plugin
that provides network filesystem (NFS) support.

This project may be compiled as a stand-alone binary using Golang that,
when run, provides a valid CSI endpoint. This project can also be
vendored or built as a Golang plugin in order to extend the functionality
of other programs.

## Runtime Dependencies
The node portion of the plugin can be run on any Linux node that is able to
mount NFS volumes. The Node service verifies this by checking for the existence
of `/sbin/mount.nfs` and `/sbin/mount.nfs4` during a `NodeProbe`.

## Installation
CSI-NFS can be installed with Go and the following command:

`$ go get github.com/thecodeteam/csi-nfs`

The resulting binary will be installed to `$GOPATH/bin/csi-nfs`.

If you want to build `csi-nfs` with accurate version information, you'll
need to run the `go generate` command and build again:

```bash
$ go get github.com/thecodeteam/csi-nfs
$ cd $GOPATH/src/github.com/thecodeteam/csi-nfs
$ go generate && go install
```

The binary will once again be installed to `$GOPATH/bin/csi-nfs`.

## Start plugin
Before starting the plugin please set the environment variable
`CSI_ENDPOINT` to a valid Go network address such as `csi.sock`:

```bash
$ CSI_ENDPOINT=csi.sock csi-nfs
INFO[0000] configured com.thecodeteam.csi-nfs            privatedir=/dev/csi-nfs-mounts
INFO[0000] identity service registered
INFO[0000] controller service registered
INFO[0000] node service registered
INFO[0000] serving                                       endpoint="unix:///csi.sock"
```

The server can be shutdown by using `Ctrl-C` or sending the process
any of the standard exit signals.

## Using plugin
The CSI specification uses the gRPC protocol for plug-in communication.
The easiest way to interact with a CSI plugin is via the Container
Storage Client (`csc`) program provided via the
[GoCSI](https://github.com/thecodeteam/gocsi) project:

```bash
$ go get github.com/thecodeteam/gocsi
$ go install github.com/thecodeteam/gocsi/csc
```

Then, have `csc` use the same `CSI_ENDPOINT`, and you can issue commands
to the plugin. Some examples...

Get the plugin's supported versions and plugin info:

```bash
$ csc -e csi.sock identity supported-versions
0.1.0
$ csc -e csi.sock -v 0.1.0 identity plugin-info
"com.thecodeteam.csi-nfs"	"0.1.0+9"
"commit"="8b9c33929bc954614f84d687b47dae71891d5514"
"formed"="Tue, 13 Feb 2018 17:37:15 UTC"
"semver"="0.1.0+9"
"url"="https://github.com/thecodeteam/csi-nfs"
```

Publish an NFS volume to a target path:


```bash
$ csc -e csi.sock -v 0.1.0 n publish --cap SINGLE_NODE_WRITER,mount,nfs --target-path /tmp/mnt 192.168.75.2:/data
192.168.75.2:/data
```

Unpublish NFS volume:

```bash
$ csc -e csi.sock -v 0.1.0 n unpublish --target-path /tmp/mnt 192.168.75.2:/data
192.168.75.2:/data
```

## Parameters
No additional parameters are currently supported/required by the plugin

## Configuration
The CSI-NFS plugin is built using the GoCSI package. Please see its
[configuration section](https://github.com/thecodeteam/gocsi#configuration) for
a complete list of the environment variables that may be used to configure this
plugin

The following table is a list of this SP's default configuration values:

| Name | Value |
|------|-------|
| `X_CSI_SPEC_REQ_VALIDATION` | `true` |
| `X_CSI_SERIAL_VOL_ACCESS` | `true` |
| `X_CSI_SUPPORTED_VERSIONS` | `0.1.0` |
| `X_CSI_PRIVATE_MOUNT_DIR` | `/dev/disk/csi-nfs-private` |

## Capable operational modes
The CSI spec defines a set of AccessModes that a volume can have. CSI-NFS
supports the following modes for volumes :

```
// Can only be published once as read/write on a single node,
// at any given time.
SINGLE_NODE_WRITER = 1;

// Can only be published once as readonly on a single node,
// at any given time.
SINGLE_NODE_READER_ONLY = 2;

// Can be published as readonly at multiple nodes simultaneously.
MULTI_NODE_READER_ONLY = 3;

// Can be published at multiple nodes simultaneously. Only one of
// the node can be used as read/write. The rest will be readonly.
MULTI_NODE_SINGLE_WRITER = 4;

// Can be published as read/write at multiple nodes
// simultaneously.
MULTI_NODE_MULTI_WRITER = 5;
```

The plugin attempts no verification that NFS clients and servers are configured
correctly for multi-node writer scenarios (e.g. running `rpc.lockd` for NFSv3)

## Support
For any questions or concerns please file an issue with the
[csi-nfs](https://github.com/thecodeteam/csi-nfs/issues) project or join
the Slack channel #project-rexray at codecommunity.slack.com.
