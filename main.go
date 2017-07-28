package main

import (
	"errors"
	"net"
	"os"
	"sync"

	log "github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/codenrhoden/csi-nfs-plugin/csi"
	"github.com/codenrhoden/csi-nfs-plugin/csiutils"
)

const (
	name = "gocsi-nfs"

	nodeEnvVar  = "NFSPLUGIN_NODEONLY"
	ctlrEnvVar  = "NFSPLUGIN_CONTROLLERONLY"
	debugEnvVar = "NFSPLUGIN_DEBUG"
)

var (
	errServerStarted = errors.New(name + ": the server has been started")
	errServerStopped = errors.New(name + ": the server has been stopped")
)

func main() {
	if _, d := os.LookupEnv(debugEnvVar); d {
		log.SetLevel(log.DebugLevel)
	}

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

	// Always host the Indentity Service
	csi.RegisterIdentityServer(s.server, s)

	_, nodeSvc := os.LookupEnv(nodeEnvVar)
	_, ctrlSvc := os.LookupEnv(ctlrEnvVar)

	if nodeSvc && ctrlSvc {
		log.Fatalf("Cannot specify both %s and %s",
			nodeEnvVar, ctlrEnvVar)
	}

	switch {
	case nodeSvc:
		//csi.RegisterNodeServer(s.server, s)
		//log.Debug("Added Node Service")
	case ctrlSvc:
		csi.RegisterControllerServer(s.server, s)
		log.Debug("Added Controller Service")
	default:
		//csi.RegisterNodeServer(s.server, s)
		//log.Debug("Added Node Service")
		csi.RegisterControllerServer(s.server, s)
		log.Debug("Added Controller Service")
	}

	// start the grpc server
	if err := s.server.Serve(li); err != grpc.ErrServerStopped {
		return err
	}
	return errServerStopped
}
