package rpc

import (
	"bufio"
	"errors"
	"io"
	"net"
	"net/http"
	"net/rpc"

	jkpool "github.com/jkprj/jkfr/gokit/transport/pool"
	jktls "github.com/jkprj/jkfr/gokit/utils/tls"
	jklog "github.com/jkprj/jkfr/log"
	jknet "github.com/jkprj/jkfr/net"
)

var connected = "200 Connected to Go RPC"

type NewClient func(conn net.Conn, o *jkpool.Options) (jkpool.PoolClient, error)

func DefaultTcpClientFatory() jkpool.ClientFatory {
	return TcpClientFactory(DefaultNewRpcClient)
}

func DefaultTLSClientFatory(clientpem, clientkey []byte) jkpool.ClientFatory {
	return TLSClientFatory(DefaultNewRpcClient, clientpem, clientkey)
}

func DefaultRpcHttpClientFatory(path string) jkpool.ClientFatory {
	return RpcHttpFatory(DefaultNewRpcClient, path)
}

func DefaultRpcTLSHttpFatory(clientpem, clientkey []byte, path string) jkpool.ClientFatory {
	return RpcTLSHttpFatory(DefaultNewRpcClient, clientpem, clientkey, path)
}

func DefaultNewRpcClient(conn net.Conn, o *jkpool.Options) (p jkpool.PoolClient, err error) {
	return rpc.NewClientWithCodec(NewTimeoutCodecEx(conn, o)), nil
}

func TcpConn(o *jkpool.Options) (net.Conn, error) {
	target := o.ServerAddr
	if target == "" {
		return nil, jkpool.ErrTargets
	}

	conn, err := jknet.ConnWithResolve(target, func(addr string) (net.Conn, error) {
		return net.DialTimeout("tcp", addr, o.DialTimeout)
	})
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func TcpClientFactory(newClient NewClient) jkpool.ClientFatory {

	return func(o *jkpool.Options) (jkpool.PoolClient, net.Conn, error) {
		conn, err := TcpConn(o)
		if err != nil {
			jklog.Debugw("TcpConn fail", "err", err)
			return nil, nil, err
		}

		cli, err := newClient(conn, o)

		return cli, conn, err
	}
}

func TLSConn(o *jkpool.Options, clientpem, clientkey []byte) (net.Conn, error) {

	target := o.ServerAddr
	if target == "" {
		return nil, jkpool.ErrTargets
	}

	conn, err := jknet.ConnWithResolve(target, func(addr string) (net.Conn, error) {
		return jktls.CreateTLSConn(clientpem, clientkey, addr, o.DialTimeout)
	})
	if err != nil {
		jklog.Errorw("CreateTLSConn fail", "target", target, "err", err)
		return nil, err
	}

	return conn, nil
}

func TLSClientFatory(newClient NewClient, clientpem, clientkey []byte) jkpool.ClientFatory {

	return func(o *jkpool.Options) (jkpool.PoolClient, net.Conn, error) {

		conn, err := TLSConn(o, clientpem, clientkey)
		if err != nil {
			jklog.Errorw("TLSConn fail", "err", err)
			return nil, nil, err
		}

		cli, err := newClient(conn, o)

		return cli, conn, err
	}
}

func RpcHttpConn(o *jkpool.Options, path string) (net.Conn, error) {

	var err error
	conn, err := TcpConn(o)
	if err != nil {
		jklog.Errorw("TcpConn err", "err", err)
		return nil, err
	}
	io.WriteString(conn, "CONNECT "+path+" HTTP/1.0\n\n")

	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == connected {
		return conn, nil
	}
	if err == nil {
		jklog.Errorw("http Response err", "err", "unexpected HTTP response: "+resp.Status)
		err = errors.New("unexpected HTTP response: " + resp.Status)
	} else {
		jklog.Errorw("http Response err", "err", err)
	}

	conn.Close()

	return nil, &net.OpError{
		Op:   "dial-http",
		Net:  "tcp:" + o.ServerAddr,
		Addr: nil,
		Err:  err,
	}
}

func RpcHttpFatory(newClient NewClient, path string) jkpool.ClientFatory {

	return func(o *jkpool.Options) (jkpool.PoolClient, net.Conn, error) {
		conn, err := RpcHttpConn(o, path)
		if nil != err {
			return nil, nil, err
		}

		cli, err := newClient(conn, o)

		return cli, conn, err
	}
}

func RpcTLSHttpConn(o *jkpool.Options, clientpem, clientkey []byte, path string) (net.Conn, error) {

	conn, err := TLSConn(o, clientpem, clientkey)
	if err != nil {
		jklog.Errorw("TLSConn fail", "err", err)
		return nil, err
	}
	io.WriteString(conn, "CONNECT "+path+" HTTP/1.0\n\n")

	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	if err == nil && resp.Status == connected {
		return conn, nil
	}
	if err == nil {
		jklog.Errorw("http Response err", "err", "unexpected HTTP response: "+resp.Status)
		err = errors.New("unexpected HTTP response: " + resp.Status)
	} else {
		jklog.Errorw("http Response err", "err", err)
	}

	conn.Close()

	return nil, &net.OpError{
		Op:   "dial-http",
		Net:  "tcp:" + o.ServerAddr,
		Addr: nil,
		Err:  err,
	}
}

func RpcTLSHttpFatory(newClient NewClient, clientpem, clientkey []byte, path string) jkpool.ClientFatory {

	return func(o *jkpool.Options) (jkpool.PoolClient, net.Conn, error) {
		conn, err := RpcTLSHttpConn(o, clientpem, clientkey, path)
		if nil != err {
			return nil, nil, err
		}

		cli, err := newClient(conn, o)

		return cli, conn, err
	}
}
