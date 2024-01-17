package rpc

import (
	"context"
	"errors"
	"io"
	"net/rpc"
	"sync"
	"time"

	kitlog "github.com/jkprj/jkfr/gokit/log"
	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jksd "github.com/jkprj/jkfr/gokit/sd"
	jktrans "github.com/jkprj/jkfr/gokit/transport"
	jkendpoint "github.com/jkprj/jkfr/gokit/transport/endpoint"
	jkpool "github.com/jkprj/jkfr/gokit/transport/pool"
	rpcpool "github.com/jkprj/jkfr/gokit/transport/pool/rpc"
	jkutils "github.com/jkprj/jkfr/gokit/utils"
	jklb "github.com/jkprj/jkfr/gokit/utils/lb"
	jklog "github.com/jkprj/jkfr/log"

	"github.com/go-kit/kit/endpoint"
	kitsd "github.com/go-kit/kit/sd"
	kitconsul "github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
)

var connected = "200 Connected to Go RPC"

var name2client map[string]*RPCClient = make(map[string]*RPCClient)
var mtClient sync.RWMutex
var mtNewClient sync.Mutex

func Call(name, action string, req, resp interface{}) error {

	client, err := GetRPCClient(name)
	if nil != err {
		return err
	}

	return client.Call(action, req, resp)
}

func GoCall(name, action string, req, resp interface{}) *rpc.Call {
	client, err := GetRPCClient(name)
	if nil != err {
		call := new(rpc.Call)
		call.ServiceMethod = action
		call.Args = req
		call.Error = errors.New("client not found, name:" + name)
		return call
	}

	return client.GoCall(action, req, resp)
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

func RegistryClient(client *RPCClient) {
	Close(client.name)

	mtClient.Lock()
	defer mtClient.Unlock()

	name2client[client.name] = client
}

func RegistryNewClient(name string, ops ...ClientOption) error {

	client, err := NewClient(name, ops...)
	if nil != err {
		return err
	}

	RegistryClient(client)

	return nil
}

func GetRPCClient(name string) (*RPCClient, error) {

	client := getCacheClient(name)
	if nil != client {
		return client, nil
	}

	mtNewClient.Lock()
	defer mtNewClient.Unlock()

	client = getCacheClient(name)
	if nil != client {
		return client, nil
	}

	client, err := NewClient(name)
	if nil != err {
		return nil, err
	}

	saveClient(name, client)

	return client, nil
}

func getCacheClient(name string) *RPCClient {
	mtClient.RLock()

	client, ok := name2client[name]
	if !ok {
		mtClient.RUnlock()
		return nil
	}

	mtClient.RUnlock()

	return client
}

func saveClient(name string, client *RPCClient) {
	mtClient.Lock()
	defer mtClient.Unlock()

	name2client[name] = client
}

type reuquestParam struct {
	action  string
	request interface{}
	respone interface{}
}

type RPCClient struct {
	name string

	cfg          *ClientConfig
	consulClient kitconsul.Client

	rpcPool *rpcpool.RpcPools

	consulInstancer  *kitconsul.Instancer
	consulEndpointer *jksd.DefaultEndpointer
	reqEndPoint      endpoint.Endpoint

	actionEndPoint map[string]endpoint.Endpoint
	mtAction       sync.RWMutex
}

func NewClient(name string, ops ...ClientOption) (rpcClient *RPCClient, err error) {
	rpcClient = new(RPCClient)
	rpcClient.name = name
	rpcClient.cfg = newClientConfig(name, ops...)
	rpcClient.actionEndPoint = map[string]endpoint.Endpoint{}

	rpcClient.consulClient, err = jkregistry.NewConsulClient(rpcClient.name, rpcClient.cfg.RegOps...)
	if nil != err {
		jklog.Errorw("jkregistry.NewConsulClient fail", "name", rpcClient.name, "cfg", *rpcClient.cfg, "err", err.Error())
		return nil, err
	}

	rpcClient.init_pool()
	rpcClient.reqEndPoint = rpcClient.makeRuquestEndpoint()

	return rpcClient, nil
}

func (client *RPCClient) init_pool() {

	op := jkpool.NewOptions()
	op.InitCap = client.cfg.PoolCap
	op.MaxCap = client.cfg.MaxCap
	op.DialTimeout = time.Duration(client.cfg.DialTimeout) * time.Second
	op.IdleTimeout = time.Duration(client.cfg.IdleTimeout) * time.Second
	op.ReadTimeout = time.Duration(client.cfg.ReadTimeout) * time.Second
	op.WriteTimeout = time.Duration(client.cfg.WriteTimeout) * time.Second
	op.Factory = client.cfg.Fatory(client.cfg)
	op.Codec = client.cfg.Codec

	client.rpcPool, _ = rpcpool.NewRpcPools(nil, op)
	client.rpcPool.SetIdleTimeOut(uint(client.cfg.IdleTimeout))
	client.rpcPool.SetRetryTimes(1) // RPCClient有自己的retry
}

func (client *RPCClient) Close() {

	if nil != client.consulEndpointer {
		client.consulEndpointer.Close()
	}

	if nil != client.consulInstancer {
		client.consulInstancer.Stop()
	}

	client.rpcPool.Close()
}

func (client *RPCClient) Call(action string, req, resp interface{}) (err error) {

	client.mtAction.RLock()
	repEndPoint, ok := client.actionEndPoint[action]
	client.mtAction.RUnlock()
	if !ok {
		repEndPoint = jkendpoint.Chain(client.reqEndPoint, action, client.cfg.ActionMiddlewares...)
	}

	reqParam := reuquestParam{action: action, request: req, respone: resp}

	_, err = repEndPoint(context.Background(), reqParam)
	if nil != err {
		jklog.Errorw("EndPoint Request fail", "name:", client.name, "action", reqParam.action, "req", req, "err", err.Error())
		return err
	}

	if !ok {
		client.mtAction.Lock()
		client.actionEndPoint[action] = repEndPoint
		client.mtAction.Unlock()
	}

	return nil
}

func (client *RPCClient) GoCall(action string, req, resp interface{}) *rpc.Call {
	call := new(rpc.Call)
	call.ServiceMethod = action
	call.Args = req
	call.Reply = resp

	call.Done = client.cfg.AsyncCallChan

	go func() {
		call.Error = client.Call(action, req, resp)
		call.Done <- call
	}()

	return call
}

func (client *RPCClient) makeRuquestEndpoint() endpoint.Endpoint {

	client.consulInstancer = kitconsul.NewInstancer(client.consulClient, kitlog.InfowLogger, client.name, client.cfg.ConsulTags, client.cfg.PassingOnly)
	client.consulEndpointer = jksd.NewEndpointer(client.consulInstancer, client.makeRequestFactory(), kitlog.ErrorwLogger)

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

func (client *RPCClient) makeRequestFactory() kitsd.Factory {
	return func(instance string) (endpoint endpoint.Endpoint, closer io.Closer, err_ error) {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {

			reqParam, ok := request.(reuquestParam)
			if !ok {
				return nil, errors.New("the request is not reuquestParam")
			}

			err = client.rpcPool.CallWithAddr(instance, reqParam.action, reqParam.request, reqParam.respone)
			if nil != err {
				// jklog.Errorw("client.rpcPool.CallWithAddr fail", "instance", instance, "action", reqParam.action, "request", reqParam.request, "err", err)
				return nil, err
			}

			return nil, nil
		}, nil, nil
	}
}
