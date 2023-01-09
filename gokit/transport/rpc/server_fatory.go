package rpc

import (
	"net"
	"net/http"
	"net/rpc/jsonrpc"

	jktrans "github.com/jkprj/jkfr/gokit/transport"
	jktls "github.com/jkprj/jkfr/gokit/utils/tls"
	jklog "github.com/jkprj/jkfr/log"
)

type CreateListenerFunc func(cfg *ServerConfig) (net.Listener, error)
type ServerRunFunc func(listener net.Listener, server *Server, cfg *ServerConfig) error

func TCPListenerFatory(cfg *ServerConfig) (net.Listener, error) {
	return net.Listen("tcp", cfg.BindAddr)
}

func TLSListenerFatory(cfg *ServerConfig) (net.Listener, error) {
	return jktls.CreateTLSListen(cfg.ServerPem, cfg.ServerKey, cfg.ClientPem, cfg.BindAddr)
}

func RunServerWithTcp(listener net.Listener, server *Server, cfg *ServerConfig) error {

	for {
		conn, err := listener.Accept()
		if err != nil {
			jklog.Errorw("rpc.Serve accept fail", "err", err.Error())
			return err
		}

		go func() {
			if jktrans.CODEC_JSON == cfg.Codec {
				server.ServeCodec(jsonrpc.NewServerCodec(conn))
			} else {
				server.ServeConn(conn)
			}
		}()
	}

	return nil
}

func RunServerWithHttp(listener net.Listener, server *Server, cfg *ServerConfig) error {
	server.HandleHTTP(cfg.RpcPath, cfg.RpcDebugPath)
	err := http.Serve(listener, server)
	if nil != err {
		jklog.Errorw("RunHttpServer fail", "err", err)
		return err
	}

	return nil
}
