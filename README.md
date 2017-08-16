CSI-NFS
-------

CSI-NFS is an implementation of a
[CSI](https://github.com/container-storage-interface) plugin for NFS volumes.

It is structured such that it can be compiled into a standalone golang binary
that can be executed to meet the requirements of a CSI plugin. Furthermore, the
core NFS logic is separated into a `nfs` go package that can be imported for use
by other programs.
