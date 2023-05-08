package grpc

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

func (rp *GRPCPool) CallWithContext(ctx context.Context, action string, request interface{}) (response interface{}, err error) {

	c, err := rp.pool.Get()
	if nil != err {
		return nil, err
	}

	client, ok := c.Client.(*ClientHandle)
	if !ok {
		return nil, errors.New("tranfer to ClientHandle fail")
	}

	resp, err := client.call(ctx, action, request)
	if nil != err {
		st, ok := status.FromError(err)
		if ok && (codes.OK == st.Code() ||
			codes.Unknown == st.Code() ||
			codes.Unimplemented == st.Code()) {

			rp.pool.Put(c, jkpool.GOOD) // 非网络原因导致的失败不回收
		} else {
			rp.pool.Put(c, jkpool.BAD)
		}

		return nil, err
	}

	rp.pool.Put(c, jkpool.GOOD)

	return resp, nil
}

func (rp *GRPCPool) Call(action string, req interface{}) (rsp interface{}, err error) {
	return rp.CallWithContext(context.Background(), action, req)
}

func (rp *GRPCPool) IsConnected() bool {
	return 0 < rp.pool.ValidCount()
}

func (rp *GRPCPool) Close() {
	rp.pool.Close()
}

func (rp *GRPCPool) GetPool() *jkpool.Pool {
	return rp.pool
}
