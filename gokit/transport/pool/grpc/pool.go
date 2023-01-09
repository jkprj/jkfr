package grpc

import (
	"context"
	"errors"

	jkpool "github.com/jkprj/jkfr/gokit/transport/pool"
)

type GRPCPool struct {
	pool *jkpool.Pool
	addr string
	o    *jkpool.Options
}

func NewGRPCPool(o *jkpool.Options) (p *GRPCPool, err error) {

	p = new(GRPCPool)
	p.o = o
	p.addr = o.ServerAddr
	p.pool, err = jkpool.NewPool(p.o)
	if nil != err {
		return nil, err
	}

	return p, nil
}

func (rp *GRPCPool) get_client() (*ClientHandle, error) {
	plclient, err := rp.pool.Get()
	if nil != err {
		return nil, err
	}

	client, ok := plclient.(*ClientHandle)
	if !ok {
		return nil, errors.New("tranfer to ClientHandle fail")
	}

	return client, nil
}

func (rp *GRPCPool) CallWithContext(ctx context.Context, action string, request interface{}) (response interface{}, err error) {

	client, err := rp.get_client()
	if nil != err {
		return nil, err
	}

	resp, err := client.call(ctx, action, request)
	if nil != err {
		rp.Put(client, jkpool.BAD)
		return nil, err
	}

	rp.Put(client, jkpool.GOOD)

	return resp, nil
}

func (rp *GRPCPool) Call(action string, req interface{}) (rsp interface{}, err error) {
	return rp.CallWithContext(context.Background(), action, req)
}

func (rp *GRPCPool) Close() {
	rp.pool.Close()
}

func (rp *GRPCPool) Get() (client *ClientHandle, err error) {
	return rp.get_client()
}

func (rp *GRPCPool) Put(client *ClientHandle, good bool) error {
	return rp.pool.Put(client, good)
}
