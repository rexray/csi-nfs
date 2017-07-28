package main

import (
	"errors"
	"net"
	"sync"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/codenrhoden/csi-nfs-plugin/csi"
	"github.com/codenrhoden/csi-nfs-plugin/csiutils"
)

const (
	name = "gocsi-nfs"
)

var (
	errServerStarted = errors.New(name + ": the server has been started")
	errServerStopped = errors.New(name + ": the server has been stopped")
)

func main() {
	l, err := csiutils.GetCSIEndpointListener()
	if err != nil {
		log.WithError(err).Fatal("failed to listen")
	}

	ctx := context.Background()

	s := &sp{name: name}

	if err := s.Serve(ctx, l); err != nil {
		log.WithError(err).Fatal("grpc failed")
	}
}

type sp struct {
	sync.Mutex
	name   string
	server *grpc.Server
	closed bool
}

// ServiceProvider.Serve
func (s *sp) Serve(ctx context.Context, li net.Listener) error {
	log.WithField("name", s.name).Info(".Serve")
	if err := func() error {
		s.Lock()
		defer s.Unlock()
		if s.closed {
			return errServerStopped
		}
		if s.server != nil {
			return errServerStarted
		}
		s.server = grpc.NewServer()
		return nil
	}(); err != nil {
		return errServerStarted
	}

	csi.RegisterIdentityServer(s.server, s)
	//csi.RegisterControllerServer(s.server, s)
	//csi.RegisterNodeServer(s.server, s)

	// start the grpc server
	if err := s.server.Serve(li); err != grpc.ErrServerStopped {
		return err
	}
	return errServerStopped
}
