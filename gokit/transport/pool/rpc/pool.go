package rpc

import (
	"context"
	"errors"
	"net"
	"net/rpc"
	"time"

	"github.com/jkprj/jkfr/log"

	jkpool "github.com/jkprj/jkfr/gokit/transport/pool"
)

type RpcPool struct {
	pool *jkpool.Pool
	addr string
	o    *jkpool.Options
}

func NewRpcPool(o *jkpool.Options) (p *RpcPool, err error) {

	p = new(RpcPool)
	p.o = o
	p.addr = o.ServerAddr

	if nil == p.o.Factory {
		p.o.Factory = DefaultTcpClientFatory()
	}

	p.pool, err = jkpool.NewPool(p.o)
	if nil != err {
		log.Errorw("NewPool fail", "error", err)
		return nil, err
	}

	return p, nil
}

func (rp *RpcPool) get_client() (*rpc.Client, error) {
	plclient, err := rp.pool.Get()
	if nil != err {
		// log.Errorw("rp.pool.Get fail", "error", err)
		return nil, err
	}

	client, ok := plclient.(*rpc.Client)
	if !ok {
		log.Errorw("transfer *rpc.Client fail", "client", client)
		return nil, errors.New("tranfer to rpc.Client fail")
	}

	return client, nil
}

func (rp *RpcPool) call(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error {

	client, err := rp.get_client()
	if nil != err {
		// log.Errorw("get_client fail", "error", err)
		return err
	}

	rpcCall := client.Go(serviceMethod, args, reply, nil)

	timeoutCtx := ctx
	var cancel context.CancelFunc
	if nil == timeoutCtx {
		timeoutCtx, cancel = context.WithTimeout(context.Background(), rp.o.ReadTimeout)
	}

	select {
	case <-rpcCall.Done:
		err = rpcCall.Error
	case <-timeoutCtx.Done():
		err = errors.New("ReadTimeout addr:" + rp.addr + ", method:" + serviceMethod)
	}

	if nil != cancel {
		cancel()
	}

	if nil != err {
		_, ok := err.(rpc.ServerError)
		if !ok {
			rp.Put(client, jkpool.BAD)
		} else {
			rp.Put(client, jkpool.GOOD) // 服务端返回的错误说明连接还是正常的，不需要尝试释放连接
			log.Infow("ServerError", "err", err)
		}

		log.Errorw("client.Call fail", "method", serviceMethod, "error", err)
		return err
	}

	rp.Put(client, jkpool.GOOD)

	return nil
}

func (rp *RpcPool) CallWithTimeOut(serviceMethod string, args interface{}, reply interface{}, timeout time.Duration) error {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)

	err := rp.call(ctx, serviceMethod, args, reply)

	cancel()

	return err
}

func (rp *RpcPool) CallWithContext(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) error {
	return rp.call(ctx, serviceMethod, args, reply)
}

func (rp *RpcPool) Call(serviceMethod string, args interface{}, reply interface{}) error {
	return rp.CallWithContext(nil, serviceMethod, args, reply)
}

func (rp *RpcPool) GoCall(serviceMethod string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {
	return rp.GoCallWithContext(nil, serviceMethod, args, reply, done)
}

func (rp *RpcPool) GoCallWithContext(ctx context.Context, serviceMethod string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {

	call := new(rpc.Call)
	call.ServiceMethod = serviceMethod
	call.Args = args
	call.Reply = reply
	call.Error = nil

	if done == nil {
		done = make(chan *rpc.Call, 10)
	}

	call.Done = done

	go func() {
		call.Error = rp.CallWithContext(ctx, serviceMethod, args, reply)
		if nil != call.Error {
			log.Errorw("RpcPool.call fail", "error", call.Error)
		}

		done <- call
	}()

	return call
}

func (rp *RpcPool) Close() {
	log.Infow("RpcPool.Close")
	rp.pool.Close()
}

func (rp *RpcPool) IsConnected() bool {
	return 0 < rp.pool.ValidCount()
}

func (rp *RpcPool) GetConn() (conn net.Conn, err error) {
	return rp.pool.GetConn()
}

func (rp *RpcPool) Get() (client *rpc.Client, err error) {
	return rp.get_client()
}

func (rp *RpcPool) Put(client *rpc.Client, good bool) error {
	return rp.pool.Put(client, good)
}
