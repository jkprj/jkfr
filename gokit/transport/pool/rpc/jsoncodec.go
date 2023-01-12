package rpc

import (
	"fmt"
	"io"
	"net/rpc"
	"net/rpc/jsonrpc"
	"time"

	jkpool "github.com/jkprj/jkfr/gokit/transport/pool"
	"github.com/jkprj/jkfr/gokit/utils"
)

type jsonCodec struct {
	codec rpc.ClientCodec

	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func NewJsonCodec(conn io.ReadWriteCloser, o *jkpool.Options) rpc.ClientCodec {

	jc := new(jsonCodec)
	jc.codec = jsonrpc.NewClientCodec(conn)
	jc.ReadTimeout = o.ReadTimeout
	jc.WriteTimeout = o.WriteTimeout

	return jc

}

func (jc *jsonCodec) WriteRequest(r *rpc.Request, body interface{}) (err error) {

	// timeout := jc.WriteTimeout
	// if jc.WriteTimeout <= 0 {
	// 	timeout = time.Second * 60
	// }

	// echan := make(chan error, 1)
	// go func() {
	// 	echan <- jc.codec.WriteRequest(r, body)
	// }()

	// timeoutTimer := time.NewTimer(timeout)

	// select {
	// case err = <-echan:
	// case <-timeoutTimer.C:
	// 	err = fmt.Errorf("Timeout method:%s", r.ServiceMethod)
	// }

	// timeoutTimer.Stop()

	// return err

	return jc.codec.WriteRequest(r, body)
}

func (jc *jsonCodec) ReadResponseHeader(r *rpc.Response) (err error) {
	return jc.codec.ReadResponseHeader(r)
}

func (jc *jsonCodec) ReadResponseBody(body interface{}) (err error) {

	utils.ZeroStruct(body)

	return jc.codec.ReadResponseBody(body)
}

func (jc *jsonCodec) Close() error {
	return jc.codec.Close()
}
