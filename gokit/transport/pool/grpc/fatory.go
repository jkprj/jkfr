package grpc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"reflect"
	"time"

	jkpool "github.com/jkprj/jkfr/gokit/transport/pool"
	jklog "github.com/jkprj/jkfr/log"
	jknet "github.com/jkprj/jkfr/net"

	"google.golang.org/grpc"
)

type ClientConn struct {
	conn *grpc.ClientConn
}

func (cc *ClientConn) LocalAddr() net.Addr {
	return nil
}
func (cc *ClientConn) RemoteAddr() net.Addr {
	addr, _ := net.ResolveTCPAddr("", cc.conn.Target())
	return addr
}
func (cc *ClientConn) Read(b []byte) (n int, err error) {
	return 0, nil
}
func (cc *ClientConn) Write(b []byte) (n int, err error) {
	return 0, nil
}
func (cc *ClientConn) Close() error {
	return cc.conn.Close()
}
func (cc *ClientConn) SetDeadline(t time.Time) error {
	return nil
}
func (cc *ClientConn) SetReadDeadline(t time.Time) error {
	return nil
}
func (cc *ClientConn) SetWriteDeadline(t time.Time) error {
	return nil
}

type ClientHandle struct {
	host        string
	client      interface{}
	action2func map[string]reflect.Value
	conn        *grpc.ClientConn
}

func (client *ClientHandle) Close() error {

	if nil != client.conn {
		return client.conn.Close()
	}

	return nil
}

func (client *ClientHandle) call(ctx context.Context, action string, request interface{}) (response interface{}, err error) {

	callfunc, ok := client.action2func[action]
	if !ok {
		jklog.Errorw("Call fail, action not found", "action", action)
		return nil, fmt.Errorf("Call fail, action[%s] not found", action)
	}

	if callfunc.Kind() == reflect.Func {

		tmpRes := callfunc.Call([]reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(request)})

		if len(tmpRes) < 2 {
			return nil, fmt.Errorf("Call respone invalid, action:%s, respone param len:%d", callfunc.String(), len(tmpRes))
		}

		response = nil
		if tmpRes[0].CanInterface() {
			response = tmpRes[0].Interface()
		}

		if tmpRes[1].CanInterface() {
			err, _ = tmpRes[1].Interface().(error)
		}

		return response, err
	}

	return nil, errors.New("client not found action")
}

type ClientFatory func(conn *grpc.ClientConn) (server interface{}, err error)

func GRPCConn(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if target == "" {
		return nil, jkpool.ErrTargets
	}

	var conn *grpc.ClientConn
	_, err := jknet.ConnWithResolve(target, func(addr string) (net.Conn, error) {
		var err error
		conn, err = grpc.Dial(addr, opts...)
		return nil, err
	})

	if err != nil {
		return nil, err
	}

	return conn, nil
}

func GRPCClientFactory(clientFatory ClientFatory, opts ...grpc.DialOption) jkpool.ClientFatory {

	return func(o *jkpool.Options) (jkpool.PoolClient, net.Conn, error) {
		conn, err := GRPCConn(o.ServerAddr, opts...)
		if err != nil {
			jklog.Errorw("GRPCConn fail", "err", err)
			return nil, nil, err
		}

		cli, err := clientFatory(conn)
		if nil != err {
			return nil, nil, err
		}

		clientHandle := new(ClientHandle)
		clientHandle.host = o.ServerAddr
		clientHandle.client = cli
		clientHandle.conn = conn
		clientHandle.action2func = map[string]reflect.Value{}

		vClient := reflect.ValueOf(clientHandle.client)
		for i := 0; i < vClient.NumMethod(); i++ {
			action := vClient.Type().Method(i).Name
			clientHandle.action2func[action] = vClient.MethodByName(action)
		}

		return clientHandle, &ClientConn{conn: conn}, nil
	}
}
