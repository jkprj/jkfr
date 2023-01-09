package rpc

import (
	"errors"

	"github.com/go-kit/kit/endpoint"

	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jkendpoint "github.com/jkprj/jkfr/gokit/transport/endpoint"
	jklog "github.com/jkprj/jkfr/log"
)

type EndpointsWrapInterface interface {
	WrapAllLabeledExcept(middleware func(string, endpoint.Endpoint) endpoint.Endpoint, excluded ...string)
}

// default : TCP no TLS
func RunServer(name string, service interface{}, ops ...ServerOption) error {
	return runServer(name, service, ops...)
}

// default : TCP no TLS
func RunServerWithServerAddr(name, addr string, service interface{}, ops ...ServerOption) error {

	opts := []ServerOption{}
	opts = append(opts, ops...)
	opts = append(opts, ServerAddr(addr))

	return runServer(name, service, opts...)
}

// http no TLS
func RunHttpServer(name string, service interface{}, ops ...ServerOption) error {

	opts := []ServerOption{}
	opts = append(opts, ops...)
	opts = append(opts, ServerRun(RunServerWithHttp))

	return runServer(name, service, opts...)
}

// TCP with TLS
func RunTLSServer(name string, service interface{}, ops ...ServerOption) error {

	opts := []ServerOption{}
	opts = append(opts, ops...)
	opts = append(opts, ServerListenerFatory(TLSListenerFatory))

	return runServer(name, service, opts...)
}

// http with TLS
func RunTLSHttpServer(name string, service interface{}, ops ...ServerOption) error {

	opts := []ServerOption{}
	opts = append(opts, ops...)
	opts = append(opts, ServerListenerFatory(TLSListenerFatory), ServerRun(RunServerWithHttp))

	return runServer(name, service, opts...)
}

func runServer(name string, service interface{}, ops ...ServerOption) error {

	cfg := newServerConfig(name, ops...)

	err := WrapEnpoint(service, cfg.ActionMiddlewares)
	if nil != err {
		return err
	}

	listener, err := cfg.ListenerFatory(cfg)
	if nil != err {
		return err
	}
	defer listener.Close()

	registry, err := jkregistry.RegistryServerWithServerAddr(name, cfg.ServerAddr, cfg.RegOps...)
	if nil != err {
		jklog.Errorw("RegistryServer fail", "ServerAddr", cfg.ServerAddr, "name", name, "err", err)
		return err
	}
	defer registry.Deregister()

	server := NewServer(cfg.Codec)
	if "" == cfg.RpcName {
		err = server.Register(service)
	} else {
		err = server.RegisterName(cfg.RpcName, service)
	}
	if nil != err {
		return err
	}

	return cfg.ServerRun(listener, server, cfg)
}

func WrapEnpoint(service interface{}, actionMiddlewares []jkendpoint.ActionMiddleware) error {

	svrEndpointsInfc, ok := service.(EndpointsWrapInterface)
	if ok {
		for _, actionMiddleware := range actionMiddlewares {
			svrEndpointsInfc.WrapAllLabeledExcept(actionMiddleware)
		}

		return nil
	}

	return errors.New("transfer service to EndpointsWrapInterface fail")
}
