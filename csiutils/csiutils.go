package csiutils

import (
	"net"
	"os"
	"regexp"
)

// GetCSIEndpoint returns the network address specified by the
// environment variable CSI_ENDPOINT.
func GetCSIEndpoint() (network, addr string, err error) {
	protoAddr := os.Getenv("CSI_ENDPOINT")
	if protoAddr == "" {
		return "", "", ErrMissingCSIEndpoint
	}
	return ParseProtoAddr(protoAddr)
}

// GetCSIEndpointListener returns the net.Listener for the endpoint
// specified by the environment variable CSI_ENDPOINT.
func GetCSIEndpointListener() (net.Listener, error) {
	proto, addr, err := GetCSIEndpoint()
	if err != nil {
		return nil, err
	}
	return net.Listen(proto, addr)
}

var addrRX = regexp.MustCompile(
	`(?i)^((?:unix)?)://(.+)$`)

// ParseProtoAddr parses a Golang network address.
func ParseProtoAddr(protoAddr string) (proto string, addr string, err error) {
	m := addrRX.FindStringSubmatch(protoAddr)
	if m == nil {
		return "", "", ErrInvalidCSIEndpoint
	}
	return m[1], m[2], nil
}
