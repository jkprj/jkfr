package rpc

import (
	"container/list"
	"errors"
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
	"time"

	jktrans "github.com/jkprj/jkfr/gokit/transport"
	jkpool "github.com/jkprj/jkfr/gokit/transport/pool"
	jkutils "github.com/jkprj/jkfr/gokit/utils"
	// jklog "github.com/jkprj/jkfr/log"
)

var ErrTimeout error = errors.New("Timeout")

type rpcRequest struct {
	req  *rpc.Request
	body interface{}
	err  error
}

type rpcResponseHeader struct {
	resp *rpc.Response
	err  error
}

type rpcResponseBody struct {
	body interface{}
	seq  uint64
	err  error
}

type rpcReq struct {
	req     *rpc.Request
	reqTime time.Time
	err     error
}

type timeoutCodec struct {
	codec rpc.ClientCodec

	readTimeout  time.Duration
	writeTimeout time.Duration

	chReq     chan *rpcRequest
	chReqRet  chan *rpcRequest
	chReqExit chan int

	chRespHeader     chan int
	chRespHeaderRet  chan *rpcResponseHeader
	chRespHeaderExit chan int

	chRespBody         chan *rpcResponseBody
	chRespBodyRet      chan *rpcResponseBody
	chRespBodyComplete chan int
	chRespBodyExit     chan int

	seq2em    map[uint64]*list.Element
	liRpcReq  *list.List
	headerSeq uint64
	mtReq     sync.RWMutex
}

func NewTimeoutCodec(codec rpc.ClientCodec, readTimeout, writeTimeout time.Duration) rpc.ClientCodec {

	tc := new(timeoutCodec)
	tc.codec = codec
	tc.readTimeout = readTimeout
	tc.writeTimeout = writeTimeout
	tc.seq2em = make(map[uint64]*list.Element)
	tc.liRpcReq = list.New()

	tc.chReq = make(chan *rpcRequest)
	tc.chReqRet = make(chan *rpcRequest)
	tc.chReqExit = make(chan int)

	tc.chRespHeader = make(chan int)
	tc.chRespHeaderRet = make(chan *rpcResponseHeader)
	tc.chRespHeaderExit = make(chan int)

	tc.chRespBody = make(chan *rpcResponseBody)
	tc.chRespBodyRet = make(chan *rpcResponseBody)
	tc.chRespBodyComplete = make(chan int, 1)
	tc.chRespBodyExit = make(chan int)
	tc.chRespBodyComplete <- 1

	go tc.goWriteRequest()
	go tc.goReadResponseHeader()
	go tc.goReadResponseBody()

	return tc

}

func NewTimeoutCodecEx(conn io.ReadWriteCloser, o *jkpool.Options) rpc.ClientCodec {

	if jktrans.CODEC_JSON == o.Codec {
		return NewTimeoutCodec(jsonrpc.NewClientCodec(conn), o.ReadTimeout, o.WriteTimeout)
	}

	return NewTimeoutCodec(NewClientCodec(conn, o), o.ReadTimeout, o.WriteTimeout)
}

func (tc *timeoutCodec) new_timer(timeout time.Duration) *time.Timer {

	if timeout <= 0 {
		timeout = time.Second * 60
	}

	return time.NewTimer(timeout)
}

func (tc *timeoutCodec) WriteRequest(r *rpc.Request, body interface{}) (err error) {

	// jklog.Infow("WriteRequest into", "seq", r.Seq)

	timeoutTimer := tc.new_timer(tc.writeTimeout)

	req := &rpcRequest{req: r, body: body}

LOOP:
	for {
		select {
		case tc.chReq <- req:
			req = nil
		case reqret := <-tc.chReqRet:
			if reqret.req.Seq != r.Seq { // 如果不一样说明是之前的write超时，忽略
				continue
			}

			err = reqret.err
			break LOOP
		case <-timeoutTimer.C:
			err = ErrTimeout
			break LOOP
		}
	}

	timeoutTimer.Stop()

	if nil == err {
		tc.pushReq(r)
	}

	// jklog.Infow("WriteRequest leave")

	return err
}

func (tc *timeoutCodec) goWriteRequest() {

	for {
		select {
		case req := <-tc.chReq:
			// jklog.Infow("goWriteRequest into", "req.seq", req.req.Seq)
			if nil != req {
				req.err = tc.codec.WriteRequest(req.req, req.body)
				tc.chReqRet <- req
			}
		case <-tc.chReqExit:
			return
		}
	}

}

func (tc *timeoutCodec) ReadResponseHeader(r *rpc.Response) (err error) {
	// jklog.Infow("ReadResponseHeader into")
	if tc.getTimeoutRequest(r) {
		// jklog.Infow("ReadResponseHeader timeout", "seq", r.Seq)
		return nil
	}

	i := 1

	for {
		timeoutTimer := tc.new_timer(time.Second)

		select {
		case tc.chRespHeader <- i:
			i = 0
			timeoutTimer.Stop()
		case resp := <-tc.chRespHeaderRet:
			timeoutTimer.Stop()

			*r = *resp.resp
			err = resp.err

			if nil == err {
				tc.headerSeq = resp.resp.Seq
			}

			// jklog.Infow("ReadResponseHeader respone", "seq", resp.resp.Seq, "err:", err)

			return err
		case <-timeoutTimer.C:
			timeoutTimer.Stop()
			if tc.getTimeoutRequest(r) {
				// jklog.Infow("ReadResponseHeader timeout2222", "seq", r.Seq)
				return nil
			}
		}
	}
}

func (tc *timeoutCodec) goReadResponseHeader() {
	for {
		select {
		case i := <-tc.chRespHeader:
			if 1 == i {
				// jklog.Infow("goReadResponseHeader into", "i", i)
				<-tc.chRespBodyComplete // read header 和 read body不能同时进行，会导致读混乱
				resp := &rpcResponseHeader{resp: new(rpc.Response)}
				resp.err = tc.codec.ReadResponseHeader(resp.resp)
				tc.chRespHeaderRet <- resp
			}
		case <-tc.chRespHeaderExit:
			return
		}
	}
}

func (tc *timeoutCodec) ReadResponseBody(body interface{}) (err error) {

	jkutils.ZeroStruct(body)

	// jklog.Infow("ReadResponseBody into", "headerSeq", tc.headerSeq)

	rq := tc.removeReq(tc.headerSeq)
	if nil != rq && nil == body && ErrTimeout == rq.err { // 如果是在读header时超时了的，就不读body了，防止导致读包混乱
		// jklog.Infow("ReadResponseBody header timout")
		return nil
	}

	timeoutTimer := tc.new_timer(tc.readTimeout)

	resp := &rpcResponseBody{body: body, seq: tc.headerSeq}

LOOP:
	for {
		select {
		case tc.chRespBody <- resp:
			resp = nil
		case respret := <-tc.chRespBodyRet:
			if tc.headerSeq != respret.seq { // 如果不一样说明是之前的read body超时，忽略
				continue
			}

			err = respret.err
			break LOOP
		case <-timeoutTimer.C:
			err = ErrTimeout
			break LOOP
		}
	}

	timeoutTimer.Stop()

	// jklog.Infow("ReadResponseBody leave", "err", err)

	return err
}

func (tc *timeoutCodec) goReadResponseBody() {
	for {
		select {
		case resp := <-tc.chRespBody:
			// jklog.Infow("goReadResponseBody into", "resp", resp)
			if nil != resp {
				resp.err = tc.codec.ReadResponseBody(resp.body)
				tc.chRespBodyRet <- resp
				tc.chRespBodyComplete <- 1
			}
		case <-tc.chRespBodyExit:
			return
		}
	}
}

func (tc *timeoutCodec) Close() error {
	err := tc.codec.Close()

	tc.chReqExit <- 1
	tc.chRespHeaderExit <- 1
	tc.chRespBodyExit <- 1

	return err
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
	rq.req = new(rpc.Request)
	*rq.req = *req
	rq.reqTime = time.Now()
	em := tc.liRpcReq.PushBack(rq)
	tc.seq2em[req.Seq] = em

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
	if time.Now().Sub(rq.reqTime) > tc.readTimeout {
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

	tc.mtReq.Unlock()

	return rq
}
