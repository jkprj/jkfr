package rpc

import (
	// "github.com/jkprj/jkfr/gokit/transport/rpc/pool"
	"github.com/jkprj/jkfr/gokit/transport/pool"
	rpcpool "github.com/jkprj/jkfr/gokit/transport/pool/rpc"
)

type ClientFatory func(cfg *ClientConfig) pool.ClientFatory

func TcpClientFatory(cfg *ClientConfig) pool.ClientFatory {
	return rpcpool.DefaultTcpClientFatory()
}

func HttpClientFatory(cfg *ClientConfig) pool.ClientFatory {
	return rpcpool.DefaultRpcHttpClientFatory(cfg.ConfigPath)
}

func TLSClientFatory(cfg *ClientConfig) pool.ClientFatory {
	return rpcpool.DefaultTLSClientFatory(cfg.ClientPem, cfg.ClientKey)
}

func TLSHttpClientFatory(cfg *ClientConfig) pool.ClientFatory {
	return rpcpool.DefaultRpcTLSHttpFatory(cfg.ClientPem, cfg.ClientKey, cfg.ConfigPath)
}
