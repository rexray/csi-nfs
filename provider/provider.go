package provider

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/thecodeteam/gocsi"
	"github.com/thecodeteam/gocsi/csi"
	"github.com/thecodeteam/goioc"

	"github.com/thecodeteam/csi-nfs/services"
)

const (
	debugEnvVar = "X_CSI_NFS_DEBUG"
	nodeEnvVar  = "X_CSI_NFS_NODEONLY"
	ctlrEnvVar  = "X_CSI_NFS_CONTROLLERONLY"
)

var (
	errServerStopped = errors.New("server stopped")
	errServerStarted = errors.New("server started")
)

// ServiceProvider is a gRPC endpoint that provides the CSI
// services: Controller, Identity, Node.
type ServiceProvider interface {

	// Serve accepts incoming connections on the listener lis, creating
	// a new ServerTransport and service goroutine for each. The service
	// goroutine read gRPC requests and then call the registered handlers
	// to reply to them. Serve returns when lis.Accept fails with fatal
	// errors.  lis will be closed when this method returns.
	// Serve always returns non-nil error.
	Serve(ctx context.Context, lis net.Listener) error

	// Stop stops the gRPC server. It immediately closes all open
	// connections and listeners.
	// It cancels all active RPCs on the server side and the corresponding
	// pending RPCs on the client side will get notified by connection
	// errors.
	Stop(ctx context.Context)

	// GracefulStop stops the gRPC server gracefully. It stops the server
	// from accepting new connections and RPCs and blocks until all the
	// pending RPCs are finished.
	GracefulStop(ctx context.Context)
}

func init() {
	goioc.Register(services.Name, func() interface{} { return &provider{} })
}

// New returns a new service provider.
func New(
	opts []grpc.ServerOption,
	interceptors []grpc.UnaryServerInterceptor) ServiceProvider {

	return &provider{interceptors: interceptors, serverOpts: opts}
}

type provider struct {
	sync.Mutex
	server       *grpc.Server
	closed       bool
	service      services.Service
	interceptors []grpc.UnaryServerInterceptor
	serverOpts   []grpc.ServerOption
}

// config is an interface that matches a possible config object that
// could possibly be pulled out of the context given to the provider's
// Serve function
type config interface {
	GetString(key string) string
}

func (p *provider) newGrpcServer() *grpc.Server {

	var interceptors []grpc.UnaryServerInterceptor
	if len(p.interceptors) > 0 {
		interceptors = append(interceptors, p.interceptors...)
	}

	iopt := gocsi.ChainUnaryServer(interceptors...)

	var serverOpts []grpc.ServerOption
	if len(p.serverOpts) > 0 {
		serverOpts = append(serverOpts, p.serverOpts...)
	}

	serverOpts = append(serverOpts, grpc.UnaryInterceptor(iopt))

	return grpc.NewServer(serverOpts...)
}

// Serve accepts incoming connections on the listener lis, creating
// a new ServerTransport and service goroutine for each. The service
// goroutine read gRPC requests and then call the registered handlers
// to reply to them. Serve returns when lis.Accept fails with fatal
// errors.  lis will be closed when this method returns.
// Serve always returns non-nil error.
func (p *provider) Serve(ctx context.Context, li net.Listener) error {
	if err := func() error {
		p.Lock()
		defer p.Unlock()
		if p.closed {
			return errServerStopped
		}
		if p.server != nil {
			return errServerStarted
		}
		p.server = p.newGrpcServer()
		return nil
	}(); err != nil {
		return errServerStarted
	}

	if _, d := os.LookupEnv(debugEnvVar); d {
		log.SetLevel(log.DebugLevel)
	}

	p.service = services.New()

	// Always host the Identity Service
	csi.RegisterIdentityServer(p.server, p.service)

	_, nodeSvc := os.LookupEnv(nodeEnvVar)
	_, ctrlSvc := os.LookupEnv(ctlrEnvVar)

	if nodeSvc && ctrlSvc {
		log.Errorf("Cannot specify both %s and %s",
			nodeEnvVar, ctlrEnvVar)
		return fmt.Errorf("Cannot specify both %s and %s",
			nodeEnvVar, ctlrEnvVar)
	}

	switch {
	case nodeSvc:
		csi.RegisterNodeServer(p.server, p.service)
		log.Debug("Added Node Service")
	case ctrlSvc:
		csi.RegisterControllerServer(p.server, p.service)
		log.Debug("Added Controller Service")
	default:
		csi.RegisterControllerServer(p.server, p.service)
		log.Debug("Added Controller Service")
		csi.RegisterNodeServer(p.server, p.service)
		log.Debug("Added Node Service")
	}

	// Start the grpc server
	log.WithFields(map[string]interface{}{
		"service": services.Name,
		"address": fmt.Sprintf(
			"%s://%s", li.Addr().Network(), li.Addr().String()),
	}).Info("serving")
	return p.server.Serve(li)
}

// Stop stops the gRPC server. It immediately closes all open
// connections and listeners.
// It cancels all active RPCs on the server side and the corresponding
// pending RPCs on the client side will get notified by connection
// errors.
func (p *provider) Stop(ctx context.Context) {
	if p.server == nil {
		return
	}

	p.Lock()
	defer p.Unlock()
	p.server.Stop()
	p.closed = true
	log.WithField("service", services.Name).Info("stopped")
}

// GracefulStop stops the gRPC server gracefully. It stops the server
// from accepting new connections and RPCs and blocks until all the
// pending RPCs are finished.
func (p *provider) GracefulStop(ctx context.Context) {
	if p.server == nil {
		return
	}

	p.Lock()
	defer p.Unlock()
	p.server.GracefulStop()
	p.closed = true
	log.WithField("service", services.Name).Info("shutdown")
}
