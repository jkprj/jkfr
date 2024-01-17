package grpc

import (
	"context"
	"errors"
	"io"
	"reflect"
	"sync"
	"time"

	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jktrans "github.com/jkprj/jkfr/gokit/transport"

	kitlog "github.com/jkprj/jkfr/gokit/log"
	jkendpoint "github.com/jkprj/jkfr/gokit/transport/endpoint"
	jkpool "github.com/jkprj/jkfr/gokit/transport/pool"
	grpc_pools "github.com/jkprj/jkfr/gokit/transport/pool/grpc"
	jkutils "github.com/jkprj/jkfr/gokit/utils"
	jklb "github.com/jkprj/jkfr/gokit/utils/lb"
	jklog "github.com/jkprj/jkfr/log"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	kitconsul "github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
)

var name2client map[string]*GRPCClient = make(map[string]*GRPCClient)
var mtClient sync.RWMutex
var mtRegClient sync.Mutex

func Call(name, action string, req interface{}) (rsp interface{}, err error) {

	mtClient.RLock()
	client, ok := name2client[name]
	mtClient.RUnlock()

	if !ok {
		return nil, errors.New("client not found, name:" + name)
	}

	return client.Call(action, req)
}

func GoCall(name, action string, req interface{}) *UCall {
	mtClient.RLock()
	client, ok := name2client[name]
	mtClient.RUnlock()

	if !ok {
		call := new(UCall)
		call.ServiceMethod = action
		call.Args = req
		call.Error = errors.New("client not found, name:" + name)
		return call
	}

	return client.GoCall(action, req)
}

func Close(name string) {
	mtClient.RLock()
	client, ok := name2client[name]
	mtClient.RUnlock()

	if ok {
		mtClient.Lock()
		delete(name2client, name)
		mtClient.Unlock()

		client.Close()
	}
}

func RegistryClient(client *GRPCClient) {
	Close(client.name)

	mtClient.Lock()
	defer mtClient.Unlock()

	name2client[client.name] = client
}

func RegistryNewClient(name string, clientFatory grpc_pools.ClientFatory, ops ...ClientOption) error {

	if nil != GetClient(name) {
		return nil
	}

	mtRegClient.Lock()
	defer mtRegClient.Unlock()

	if nil != GetClient(name) {
		return nil
	}

	client, err := NewClient(name, clientFatory, ops...)
	if nil != err {
		return err
	}

	RegistryClient(client)

	return nil
}

func GetClient(name string) *GRPCClient {
	mtClient.RLock()
	client, ok := name2client[name]
	mtClient.RUnlock()

	if ok {
		return client
	}

	return nil
}

type reuquestParam struct {
	action  string
	request interface{}
}

type GRPCClient struct {
	Done chan *UCall

	name         string
	clientFatory grpc_pools.ClientFatory

	cfg          *ClientConfig
	consulClient kitconsul.Client

	consulInstancer  *kitconsul.Instancer
	consulEndpointer *sd.DefaultEndpointer
	reqEndPoint      endpoint.Endpoint

	pools *grpc_pools.GRPCPools

	actionEndPoint map[string]endpoint.Endpoint
	mtAction       sync.RWMutex

	isClose bool
}

func NewClient(name string, clientFatory grpc_pools.ClientFatory, ops ...ClientOption) (client *GRPCClient, err error) {
	client = new(GRPCClient)
	client.name = name
	client.clientFatory = clientFatory
	client.actionEndPoint = map[string]endpoint.Endpoint{}

	client.cfg = newClientConfig(name, ops...)
	client.Done = client.cfg.AsyncCallChan

	client.consulClient, err = jkregistry.NewConsulClient(client.name, client.cfg.RegOps...)
	if nil != err {
		jklog.Errorw("jkregistry.NewConsulClient fail", "name", client.name, "cfg", *client.cfg, "err", err.Error())
		return nil, err
	}

	client.reqEndPoint = client.makeRuquestEndpoint()

	client.init_grpc_pools()

	return client, nil
}

func (client *GRPCClient) init_grpc_pools() {
	opt := jkpool.NewOptions()
	opt.InitCap = client.cfg.PoolCap
	opt.MaxCap = client.cfg.MaxCap
	opt.Factory = grpc_pools.GRPCClientFactory(client.clientFatory, client.cfg.GRPCDialOps...)

	client.pools, _ = grpc_pools.NewGRPCPools(nil, opt)

	client.pools.SetRetryTimes(1) // GRPCClient自带失败重传策略
}

func (client *GRPCClient) GetUCall() chan *UCall {
	return client.cfg.AsyncCallChan
}

func (client *GRPCClient) Call(action string, req interface{}) (rsp interface{}, err error) {

	if client.isClose {
		return nil, jkpool.ErrClosed
	}

	client.mtAction.RLock()
	repEndPoint, ok := client.actionEndPoint[action]
	client.mtAction.RUnlock()
	if !ok {
		repEndPoint = jkendpoint.Chain(client.reqEndPoint, action, client.cfg.ActionMiddlewares...)
	}

	reqParam := reuquestParam{action: action, request: req}

	rsp, err = repEndPoint(context.Background(), reqParam)
	if nil != err {
		return nil, err
	}

	if !ok {
		client.mtAction.Lock()
		client.actionEndPoint[action] = repEndPoint
		client.mtAction.Unlock()
	}

	return rsp, nil
}

func (client *GRPCClient) GoCall(action string, req interface{}) *UCall {
	call := new(UCall)
	call.ServiceMethod = action
	call.Args = req

	call.Done = client.cfg.AsyncCallChan
	if nil == call.Done {
		call.Done = make(chan *UCall, 100)
	}

	go func() {
		call.Reply, call.Error = client.Call(action, req)
		call.done()
	}()

	return call
}

func (client *GRPCClient) Close() {

	if client.isClose {
		return
	}

	client.isClose = true

	Close(client.name)

	client.pools.Close()

	if nil != client.consulEndpointer {
		client.consulEndpointer.Close()
	}

	if nil != client.consulInstancer {
		client.consulInstancer.Stop()
	}
}

func (client *GRPCClient) makeRequestFactory() sd.Factory {
	return func(instance string) (endpoint endpoint.Endpoint, closer io.Closer, err_ error) {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {

			reqParam, ok := request.(reuquestParam)
			if !ok {
				return nil, errors.New("the request is not reuquestParam, request_type:" + reflect.TypeOf(request).String())
			}

			return client.pools.CallWithAddrEx(instance, reqParam.action, reqParam.request, time.Duration(client.cfg.TimeOut)*time.Second)

		}, nil, nil
	}
}

func (client *GRPCClient) makeRuquestEndpoint() endpoint.Endpoint {

	client.consulInstancer = kitconsul.NewInstancer(
		client.consulClient,
		kitlog.InfowLogger,
		client.name,
		client.cfg.ConsulTags,
		client.cfg.PassingOnly,
	)

	client.consulEndpointer = sd.NewEndpointer(client.consulInstancer, client.makeRequestFactory(), kitlog.ErrorwLogger)

	var balancer lb.Balancer
	{
		if jkutils.STRATEGY_ROUND == client.cfg.Strategy {
			balancer = lb.NewRoundRobin(client.consulEndpointer)
		} else if jkutils.STRATEGY_RANDOM == client.cfg.Strategy {
			balancer = jklb.NewRandom(client.consulEndpointer, time.Now().UnixNano())
		} else {
			balancer = jklb.NewLeastBalancer(client.consulEndpointer)
		}
	}

	// return lb.Retry(client.cfg.Retry, time.Duration(client.cfg.TimeOut)*time.Second, balancer)
	return jktrans.Retry(client.cfg.Retry, client.cfg.RetryIntervalMS, balancer)
}
