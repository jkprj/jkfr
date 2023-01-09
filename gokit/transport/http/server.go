package http

import (
	"context"

	"net/http"

	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jkendpoint "github.com/jkprj/jkfr/gokit/transport/endpoint"
	jklog "github.com/jkprj/jkfr/log"

	"github.com/go-kit/kit/endpoint"
)

type UHandler struct {
	handler http.Handler
	cfg     *ServerConfig
}

func (uh *UHandler) ServeHTTP(rspw http.ResponseWriter, req *http.Request) {

	action := uh.cfg.GetAction(req)

	serverHttpEndpoint := makeServerHttpEndpoint(rspw, req, uh.handler)
	serverHttpEndpoint = jkendpoint.Chain(serverHttpEndpoint, action, uh.cfg.ActionMiddlewares...)

	serverHttpEndpoint(context.Background(), nil)
}

func RunServer(name string, handler http.Handler, ops ...ServerOption) error {

	uhandler := UHandler{handler: handler}
	uhandler.cfg = newServerConfig(name, ops...)

	registry, err := jkregistry.RegistryServerWithServerAddr(name, uhandler.cfg.ServerAddr, uhandler.cfg.RegOps...)
	if nil != err {
		jklog.Errorw("RegistryServer fail", "ServerAddr", uhandler.cfg.ServerAddr, "name", name, "err", err)
		return err
	}
	defer registry.Deregister()

	err = http.ListenAndServe(uhandler.cfg.BindAddr, &uhandler)
	if nil != err {
		jklog.Errorw("http server return error", "BindAddr", uhandler.cfg.BindAddr, "err", err)
	} else {
		jklog.Errorw("http server interrupt")
	}

	return err
}

func RunServerWithServerAddr(name, addr string, handler http.Handler, ops ...ServerOption) error {
	opts := []ServerOption{}
	opts = append(opts, ops...)
	opts = append(opts, ServerAddr(addr))

	return RunServer(name, handler, opts...)
}

func makeServerHttpEndpoint(rspw http.ResponseWriter, req *http.Request, handler http.Handler) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		handler.ServeHTTP(rspw, req)
		return nil, nil
	}
}
