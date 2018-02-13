CSI-NFS [![Build Status](http://travis-ci.org/thecodeteam/csi-nfs.svg?branch=master)](https://travis-ci.org/thecodeteam/csi-nfs)
-------

CSI-NFS is a Container Storage Interface
([CSI](https://github.com/container-storage-interface/spec)) plug-in
that provides network filesystem (NFS) support.

This project may be compiled as a stand-alone binary using Golang that,
when run, provides a valid CSI endpoint. This project can also be
vendored or built as a Golang plug-in in order to extend the functionality
of other programs.

Installation
-------------

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

Starting the plugin
-------------------

In order to execute the binary, you **must** set the env var `CSI_ENDPOINT`. CSI
is intended to only run over UNIX domain sockets, so a simple way to set this
endpoint to a `.sock` file in the same directory as the project is

`export CSI_ENDPOINT=unix://$(go list -f '{{.Dir}}' github.com/thecodeteam/csi-nfs)/csi-nfs.sock`

With that in place, you can start the plugin
(assuming that $GOPATH/bin is in your $PATH):

```sh
$ ./csi-nfs
INFO[0000] .Serve                                        name=csi-nfs
```

Use ctrl-C to exit.

You can enable debug logging (all logging goes to stdout) by setting the
`X_CSI_NFS_DEBUG` env var. It doesn't matter what value you set it to, just that
it is set. For example:

```sh
$ X_CSI_NFS_DEBUG= ./csi-nfs
INFO[0000] .Serve                                        name=csi-nfs
DEBU[0000] Added Controller Service
DEBU[0000] Added Node Service
^CINFO[0002] Shutting down server
```

Configuring the plugin
----------------------

The behavior of CSI-NFS can be modified with the following environment variables

| name | purpose | default |
| - | - | - |
| CSI_ENDPOINT | Set path to UNIX domain socket file | n/a |
| X_CSI_NFS_DEBUG | enable debug logging to stdout | n/a |
| X_CSI_NFS_NODEONLY | Only run the Node Service (no Controller service) | n/a |
| X_CSI_NFS_CONTROLLERONLY | Only run the Controller Service (no Node service) | n/a |

Note that the Identity service is required to always be running, and that the
default behavior is to also run both the Controller and the Node service

Using the plugin
----------------

All communication with the plugin is done via gRPC. The easiest way to interact
with a CSI plugin via CLI is to use the `csc` tool found in
[GoCSI](https://github.com/thecodeteam/gocsi).

You can install this tool with:

```sh
go get github.com/thecodeteam/gocsi
go install github.com/thecodeteam/gocsi/csc
```

With $GOPATH/bin in your $PATH, you can issue commands using the `csc` command.
You will want to use a separate shell from where you are running the `csi-nfs`
binary, and as such you will once again need to do:

`export CSI_ENDPOINT=unix://$(go list -f '{{.Dir}}' github.com/thecodeteam/csi-nfs)/csi-nfs.sock`

Here are some sample commands:

```sh
$ csc gets
0.0.0
$ csc getp -version 0.0.0
csi-nfs	0.1.0
$ csc cget -version 0.0.0
LIST_VOLUMES
$ showmount -e 192.168.75.2
Exports list on 192.168.75.2:
	/data                             192.168.75.1
$ mkdir /mnt/test
$ csc mnt -version 0.0.0 -targetPath /mnt/test -mode 1 host=192.168.75.2 export=/data
$ ls -al /mnt/test
total 1
drwxr-xr-x   2 root  wheel    18 Jul 22 20:25 .
drwxrwxrwt  85 root  wheel  2890 Aug 17 15:32 ..
-rw-r--r--   1 root  wheel     0 Jul 22 20:25 test
$ csc umount -version 0.0.0 -targetPath /mnt/test host=192.168.75.2 export=/data
$ ls -al /tmp/mnt
total 0
drwxr-xr-x   2 travis  wheel    68 Aug 16 15:01 .
drwxrwxrwt  85 root    wheel  2890 Aug 17 15:32 ..
```
