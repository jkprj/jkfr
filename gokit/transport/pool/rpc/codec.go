package rpc

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"io"
	"net/rpc"
	"time"

	"jkfr/gokit/utils"
)

type funcCodec func(e interface{}) error

//Codec ...
type Codec struct {
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Closer       io.ReadWriteCloser
	Decoder      *gob.Decoder
	Encoder      *gob.Encoder
	EncBuf       *bufio.Writer
}

//WriteRequest ...
func (c *Codec) WriteRequest(r *rpc.Request, body interface{}) (err error) {
	if err = c.timeoutCoder(c.Encoder.Encode, r, c.WriteTimeout, "write request"); err != nil {
		return
	}

	if err = c.timeoutCoder(c.Encoder.Encode, body, c.WriteTimeout, "write request body"); err != nil {
		return
	}

	return c.EncBuf.Flush()
}

//ReadResponseHeader ...
func (c *Codec) ReadResponseHeader(r *rpc.Response) (err error) {
	return c.Decoder.Decode(r)
}

//ReadResponseBody ...
func (c *Codec) ReadResponseBody(body interface{}) (err error) {

	utils.ZeroStruct(body)

	return c.Decoder.Decode(body)
}

//Close ...
func (c *Codec) Close() error {
	return c.Closer.Close()
}

func (c *Codec) timeoutCoder(fcodec funcCodec, req interface{}, timeout time.Duration, msg string) (err error) {
	if timeout <= 0 {
		timeout = time.Second * 60
	}

	echan := make(chan error, 1)
	go func() { echan <- fcodec(req) }()

	timeoutTimer := time.NewTimer(timeout)

	select {
	case err = <-echan:
	case <-timeoutTimer.C:
		err = fmt.Errorf("Timeout %s", msg)
	}

	timeoutTimer.Stop()

	return err
}
