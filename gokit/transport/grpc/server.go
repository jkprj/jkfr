package grpc

import (
	"errors"
	"net"

	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jkendpoint "github.com/jkprj/jkfr/gokit/transport/endpoint"
	jklog "github.com/jkprj/jkfr/log"

	"github.com/go-kit/kit/endpoint"

	"google.golang.org/grpc"
)

type EndpointsWrapInterface interface {
	WrapAllLabeledExcept(middleware func(string, endpoint.Endpoint) endpoint.Endpoint, excluded ...string)
}

type RegisterServerFunc func(grpcServer *grpc.Server, serverEndpoints interface{})

func RunServer(name string, serverEndpoints interface{}, registerServerFunc RegisterServerFunc, ops ...ServerOption) error {

	cfg := newServerConfig(name, ops...)

	err := WrapEndpoint(serverEndpoints, cfg.ActionMiddlewares)
	if nil != err {
		jklog.Errorw("WrapEndpoint fail", "err", err)
		return err
	}

	registry, err := jkregistry.RegistryServerWithServerAddr(name, cfg.ServerAddr, cfg.RegOps...)
	if nil != err {
		jklog.Errorw("RegistryServer fail", "ServerAddr", cfg.ServerAddr, "name", name, "err", err)
		return err
	}
	defer registry.Deregister()

	ln, err := net.Listen("tcp", cfg.BindAddr)
	if err != nil {
		jklog.Errorw("net.Listen fail", "BindAddr", cfg.BindAddr, "err", err)
		return err
	}
	defer ln.Close()

	s := grpc.NewServer(cfg.GRPCSvrOps...)
	registerServerFunc(s, serverEndpoints)

	return s.Serve(ln)
}

func RunServerWithServerAddr(name, addr string, serverEndpoints interface{}, registerServerFunc RegisterServerFunc, ops ...ServerOption) error {
	opts := []ServerOption{}
	opts = append(opts, ops...)
	opts = append(opts, ServerAddr(addr))

	return RunServer(name, serverEndpoints, registerServerFunc, opts...)
}

func WrapEndpoint(serverEndpoints interface{}, actionMiddlewares []jkendpoint.ActionMiddleware) error {

	svrEndpointsInfc, ok := serverEndpoints.(EndpointsWrapInterface)
	if ok {
		for _, actionMiddleware := range actionMiddlewares {
			svrEndpointsInfc.WrapAllLabeledExcept(actionMiddleware)
		}

		return nil
	}

	return errors.New("transfer serverEndpoints to EndpointsWrapInterface fail")
}
