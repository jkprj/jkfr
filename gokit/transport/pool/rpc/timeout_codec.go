package rpc

import (
	"container/list"
	"errors"
	"math"
	"net"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
	"sync/atomic"
	"time"

	jkpool "github.com/jkprj/jkfr/gokit/transport/pool"
	jkutils "github.com/jkprj/jkfr/gokit/utils"
	jklog "github.com/jkprj/jkfr/log"
)

var ErrTimeout error = errors.New("Timeout")

const (
	UNREAD  = 0
	READING = 1
)

type rpcReq struct {
	req     *rpc.Request
	reqTime time.Time
	err     error
}

type timeoutCodec struct {
	codec rpc.ClientCodec

	conn net.Conn

	readTimeout  time.Duration
	writeTimeout time.Duration

	seq2em        map[uint64]*list.Element
	liRpcReq      *list.List
	headerSeq     uint64
	mtReq         sync.Mutex
	isReadingBody int32
}

func NewTimeoutCodec(codec rpc.ClientCodec, conn net.Conn, readTimeout, writeTimeout time.Duration) rpc.ClientCodec {

	tc := new(timeoutCodec)
	tc.codec = codec
	tc.conn = conn
	tc.readTimeout = readTimeout
	tc.writeTimeout = writeTimeout
	tc.seq2em = make(map[uint64]*list.Element)
	tc.liRpcReq = list.New()
	tc.headerSeq = math.MaxUint64

	return tc

}

func NewTimeoutCodecEx(conn net.Conn, o *jkpool.Options) rpc.ClientCodec {

	if jkutils.CODEC_JSON == o.Codec {
		return NewTimeoutCodec(jsonrpc.NewClientCodec(conn), conn, o.ReadTimeout, o.WriteTimeout)
	}

	return NewTimeoutCodec(NewClientCodec(conn, o), conn, o.ReadTimeout, o.WriteTimeout)
}

func (tc *timeoutCodec) WriteRequest(r *rpc.Request, body interface{}) (err error) {

	tc.pushReq(r)
	tc.setDeadline()

	err = tc.codec.WriteRequest(r, body)
	if nil != err {
		tc.removeReq(r.Seq)
		tc.setDeadline()
	}

	// jklog.Infow("WriteRequest leave")

	return err
}

func (tc *timeoutCodec) ReadResponseHeader(r *rpc.Response) (err error) {
	// jklog.Infow("ReadResponseHeader into")
	if tc.getTimeoutRequest(r) {
		// jklog.Infow("ReadResponseHeader timeout", "seq", r.Seq)
		return nil
	}

	tc.setDeadline()

	err = tc.codec.ReadResponseHeader(r)
	if nil == err {
		tc.headerSeq = r.Seq
	}

	return err
}

func (tc *timeoutCodec) ReadResponseBody(body interface{}) (err error) {

	rq := tc.removeReq(tc.headerSeq)
	if nil != rq && nil == body && ErrTimeout == rq.err { // 如果是在读header时超时了的，就不读body了，防止导致读包混乱
		jklog.Infow("ReadResponseBody header timout")
		return nil
	}

	atomic.StoreInt32(&tc.isReadingBody, READING)

	tc.conn.SetReadDeadline(time.Now().Add(tc.readTimeout))
	err = tc.codec.ReadResponseBody(body)

	atomic.StoreInt32(&tc.isReadingBody, UNREAD)

	return err
}

func (tc *timeoutCodec) Close() error {
	return tc.codec.Close()
}

func (tc *timeoutCodec) getTimeoutRequest(r *rpc.Response) bool {

	req := tc.getTimeoutReq()
	if nil != req {
		r.Error = ErrTimeout.Error()
		r.Seq = req.Seq
		r.ServiceMethod = req.ServiceMethod
		tc.headerSeq = r.Seq
		return true
	}

	return false
}

func (tc *timeoutCodec) pushReq(req *rpc.Request) {

	tc.mtReq.Lock()

	rq := new(rpcReq)
	rq.req = &rpc.Request{Seq: req.Seq, ServiceMethod: req.ServiceMethod}
	rq.reqTime = time.Now()
	em := tc.liRpcReq.PushBack(rq)
	tc.seq2em[rq.req.Seq] = em

	// jklog.Infow("pushreq", "rq.req.Seq", rq.req.Seq)

	tc.mtReq.Unlock()
}

func (tc *timeoutCodec) getTimeoutReq() *rpc.Request {

	tc.mtReq.Lock()

	em := tc.liRpcReq.Front()
	if nil == em {
		tc.mtReq.Unlock()
		return nil
	}

	rq, ok := em.Value.(*rpcReq)
	if !ok {
		tc.liRpcReq.Remove(em)
		tc.mtReq.Unlock()
		return nil
	}

	// jklog.Infow("getTimeoutReq ", "sub:", time.Now().Sub(rq.reqTime), "tc.readTimeout", tc.readTimeout, "reqcount", tc.liRpcReq.Len())
	if time.Since(rq.reqTime) > tc.readTimeout {
		rq.err = ErrTimeout
		tc.mtReq.Unlock()
		return rq.req
	}

	tc.mtReq.Unlock()

	return nil
}

func (tc *timeoutCodec) removeReq(seq uint64) *rpcReq {

	tc.mtReq.Lock()

	em, ok := tc.seq2em[seq]
	if !ok {
		tc.mtReq.Unlock()
		return nil
	}

	rq := em.Value.(*rpcReq)
	tc.liRpcReq.Remove(em)
	delete(tc.seq2em, seq)

	// jklog.Infow("removeReq", "rq.req.Seq", rq.req.Seq)

	tc.mtReq.Unlock()

	return rq
}

func (tc *timeoutCodec) setDeadline() {
	tc.mtReq.Lock()

	if 0 < tc.liRpcReq.Len() {
		tc.conn.SetReadDeadline(time.Now().Add(tc.readTimeout))
		tc.conn.SetWriteDeadline(time.Now().Add(tc.writeTimeout))
	} else {
		if UNREAD == atomic.LoadInt32(&tc.isReadingBody) { // 正则读body不能将deadline设置为365天
			tc.conn.SetReadDeadline(time.Now().Add(365 * 24 * time.Hour))
		}

		tc.conn.SetWriteDeadline(time.Now().Add(365 * 24 * time.Hour))
	}

	tc.mtReq.Unlock()
}
