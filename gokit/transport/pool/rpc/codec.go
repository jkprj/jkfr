package rpc

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"net/rpc"
	"time"

	jkpool "jkfr/gokit/transport/pool"
	"jkfr/gokit/utils"
)

//codec ...
type codec struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Closer       io.ReadWriteCloser
	Decoder      *gob.Decoder
	Encoder      *gob.Encoder
	EncBuf       *bufio.Writer
}

func NewClientCodec(conn io.ReadWriteCloser, o *jkpool.Options) rpc.ClientCodec {

	encBuf := bufio.NewWriter(conn)

	c := &codec{
		Closer:       conn,
		Decoder:      gob.NewDecoder(conn),
		Encoder:      gob.NewEncoder(encBuf),
		EncBuf:       encBuf,
		ReadTimeout:  o.ReadTimeout,
		WriteTimeout: o.WriteTimeout,
	}

	return c

}

//WriteRequest ...
func (c *codec) WriteRequest(r *rpc.Request, body interface{}) (err error) {

	timeout := c.WriteTimeout
	if c.WriteTimeout <= 0 {
		timeout = time.Second * 60
	}

	echan := make(chan error, 1)
	go func() {
		echan <- c.writeRequest(r, body)
	}()

	timeoutTimer := time.NewTimer(timeout)

	select {
	case err = <-echan:
	case <-timeoutTimer.C:
		err = fmt.Errorf("WriteTimeout, method:%s", r.ServiceMethod)
	}

	timeoutTimer.Stop()

	return err
}

func (c *codec) writeRequest(r *rpc.Request, body interface{}) (err error) {

	if err = c.Encoder.Encode(r); err != nil {
		return
	}
	if err = c.Encoder.Encode(body); err != nil {
		return
	}
	return c.EncBuf.Flush()
}

//ReadResponseHeader ...
func (c *codec) ReadResponseHeader(r *rpc.Response) (err error) {
	return c.Decoder.Decode(r)
}

//ReadResponseBody ...
func (c *codec) ReadResponseBody(body interface{}) (err error) {

	utils.ZeroStruct(body)

	return c.Decoder.Decode(body)
}

//Close ...
func (c *codec) Close() error {
	return c.Closer.Close()
}
