package csiutils

import "errors"

// ErrMissingCSIEndpoint occurs when the value for the environment
// variable CSI_ENDPOINT is not set.
var ErrMissingCSIEndpoint = errors.New("missing CSI_ENDPOINT")

// ErrInvalidCSIEndpoint occurs when the value for the environment
// variable CSI_ENDPOINT is an invalid network address.
var ErrInvalidCSIEndpoint = errors.New("invalid CSI_ENDPOINT")
