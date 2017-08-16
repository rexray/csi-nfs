package main

import (
	"errors"
	"net"
	"os"
	"os/signal"
	"sync"

	log "github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/codedellemc/gocsi"
	"github.com/codedellemc/gocsi/csi"
)

const (
	name = "csi-nfs"

	nodeEnvVar  = "NFSPLUGIN_NODEONLY"
	ctlrEnvVar  = "NFSPLUGIN_CONTROLLERONLY"
	debugEnvVar = "NFSPLUGIN_DEBUG"
)

var (
	errServerStarted = errors.New(name + ": the server has been started")
	errServerStopped = errors.New(name + ": the server has been stopped")
)

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill)

	if _, d := os.LookupEnv(debugEnvVar); d {
		log.SetLevel(log.DebugLevel)
	}

	s := &sp{name: name}

	go func() {
		_ = <-c
		if s.server != nil {
			s.Lock()
			defer s.Unlock()
			log.Info("Shutting down server")
			s.server.GracefulStop()
			s.closed = true

			// make sure sock file got cleaned up
			proto, addr, _ := gocsi.GetCSIEndpoint()
			if proto == "unix" && addr != "" {
				if _, err := os.Stat(addr); !os.IsNotExist(err) {
					s.server.Stop()
					if err := os.Remove(addr); err != nil {
						log.WithError(err).Warn(
							"Unable to remove sock file")
					}
				}
			}
		}
	}()

	l, err := gocsi.GetCSIEndpointListener()
	if err != nil {
		log.WithError(err).Fatal("failed to listen")
	}

	ctx := context.Background()

	if err := s.Serve(ctx, l); err != nil {
		s.Lock()
		defer s.Unlock()
		if !s.closed {
			log.WithError(err).Fatal("grpc failed")
		}
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
		csi.RegisterNodeServer(s.server, s)
		log.Debug("Added Node Service")
	case ctrlSvc:
		csi.RegisterControllerServer(s.server, s)
		log.Debug("Added Controller Service")
	default:
		csi.RegisterControllerServer(s.server, s)
		log.Debug("Added Controller Service")
		csi.RegisterNodeServer(s.server, s)
		log.Debug("Added Node Service")
	}

	// start the grpc server
	if err := s.server.Serve(li); err != grpc.ErrServerStopped {
		return err
	}
	return errServerStopped
}
