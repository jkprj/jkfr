package rpc

import (
	"net"
	"net/http"
	"net/rpc"

	jktls "jkfr/gokit/utils/tls"
	jklog "jkfr/log"
)

type CreateListenerFunc func(cfg *ServerConfig) (net.Listener, error)
type ServerRunFunc func(listener net.Listener, server *rpc.Server, cfg *ServerConfig) error

func TCPListenerFatory(cfg *ServerConfig) (net.Listener, error) {
	return net.Listen("tcp", cfg.BindAddr)
}

func TLSListenerFatory(cfg *ServerConfig) (net.Listener, error) {
	return jktls.CreateTLSListen(cfg.ServerPem, cfg.ServerKey, cfg.ClientPem, cfg.BindAddr)
}

func RunServerWithTcp(listener net.Listener, server *rpc.Server, cfg *ServerConfig) error {
	server.Accept(listener)
	return nil
}

func RunServerWithHttp(listener net.Listener, server *rpc.Server, cfg *ServerConfig) error {
	server.HandleHTTP(cfg.RpcPath, cfg.RpcDebugPath)
	err := http.Serve(listener, server)
	if nil != err {
		jklog.Errorw("RunHttpServer fail", "err", err)
		return err
	}

	return nil
}
