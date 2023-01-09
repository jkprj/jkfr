package rpc

import (
	"container/list"
	"errors"
	"fmt"
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sync"
	"time"

	jktrans "github.com/jkprj/jkfr/gokit/transport"
	jkpool "github.com/jkprj/jkfr/gokit/transport/pool"
	"github.com/jkprj/jkfr/gokit/utils"
)

var ErrTimeout error = errors.New("Timeout")

type rpcReq struct {
	req     *rpc.Request
	reqTime time.Time
	err     error
}

type timeoutCodec struct {
	codec rpc.ClientCodec

	readTimeout  time.Duration
	writeTimeout time.Duration

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

	fmt.Println("WriteRequest into")

	echan := make(chan error, 1)
	go func() {
		echan <- tc.codec.WriteRequest(r, body)
	}()

	timeoutTimer := tc.new_timer(tc.writeTimeout)

	select {
	case err = <-echan:
	case <-timeoutTimer.C:
		err = ErrTimeout
	}

	timeoutTimer.Stop()

	if nil == err {
		tc.pushReq(r)
	}

	fmt.Println("WriteRequest leave")

	return err
}

func (tc *timeoutCodec) ReadResponseHeader(r *rpc.Response) (err error) {
	fmt.Println("ReadResponseHeader into")
	if tc.isHasTimeout(r) {
		fmt.Println("ReadResponseHeader timeout")
		return nil
	}

	echan := make(chan error, 1)
	go func() {
		echan <- tc.codec.ReadResponseHeader(r)
	}()

	for {
		timeoutTimer := tc.new_timer(time.Second)

		select {
		case err = <-echan:
			timeoutTimer.Stop()
			if nil != err {
				tc.headerSeq = r.Seq
			}
			fmt.Println("ReadResponseHeader respone", "err:", err)
			return err
		case <-timeoutTimer.C:
			timeoutTimer.Stop()
			if tc.isHasTimeout(r) {
				tc.headerSeq = r.Seq
				fmt.Println("ReadResponseHeader timeout2222")
				return nil
			}
		}
	}
}

func (tc *timeoutCodec) ReadResponseBody(body interface{}) (err error) {

	utils.ZeroStruct(body)

	fmt.Println("ReadResponseBody into")

	rq := tc.removeReq(tc.headerSeq)
	if nil != rq && nil == body && ErrTimeout == rq.err { // 如果是在读header时超时了的，就不读body了，防止导致读包混乱
		fmt.Println("ReadResponseBody header timout")
		return nil
	}

	echan := make(chan error, 1)
	go func() {
		echan <- tc.codec.ReadResponseBody(body)
	}()

	var timeoutTimer *time.Timer
	if nil != rq {
		timeoutTimer = tc.new_timer(time.Now().Sub(rq.reqTime))
	} else {
		timeoutTimer = tc.new_timer(tc.readTimeout)
	}

	select {
	case err = <-echan:
	case <-timeoutTimer.C:
		err = ErrTimeout
	}

	timeoutTimer.Stop()

	fmt.Println("ReadResponseBody leave err:", err)

	return err
}

func (tc *timeoutCodec) Close() error {
	return tc.codec.Close()
}

func (tc *timeoutCodec) isHasTimeout(r *rpc.Response) bool {

	req := tc.getTimeoutReq()
	if nil != req {
		r.Error = ErrTimeout.Error()
		r.Seq = req.Seq
		r.ServiceMethod = req.ServiceMethod
		return true
	}

	return false
}

func (tc *timeoutCodec) pushReq(req *rpc.Request) {

	tc.mtReq.Lock()

	rq := new(rpcReq)
	rq.req = req
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

	// fmt.Println("getTimeoutReq sub:", time.Now().Sub(rq.reqTime), "tc.readTimeout", tc.readTimeout)
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
