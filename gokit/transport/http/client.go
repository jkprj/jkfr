package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	kitlog "github.com/jkprj/jkfr/gokit/log"
	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jkendpoint "github.com/jkprj/jkfr/gokit/transport/endpoint"
	jkutils "github.com/jkprj/jkfr/gokit/utils"
	jklb "github.com/jkprj/jkfr/gokit/utils/lb"
	jklog "github.com/jkprj/jkfr/log"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	kitconsul "github.com/go-kit/kit/sd/consul"
	"github.com/go-kit/kit/sd/lb"
	kithttp "github.com/go-kit/kit/transport/http"
)

type MakeEncodeRequestFunc func(cfg *ClientConfig) kithttp.EncodeRequestFunc

var name2client map[string]*HttpClient = make(map[string]*HttpClient)
var mtClient sync.RWMutex
var mtNewClient sync.Mutex

func Get(name, uri string) (data []byte, err error) {
	client, err := GetClient(name)
	if nil != err {
		return nil, err
	}

	return client.Get(uri)
}

func Post(name, uri string, body []byte) (data []byte, err error) {
	client, err := GetClient(name)
	if nil != err {
		return nil, err
	}

	return client.Post(uri, body)
}

func JSGet(name, uri string, rsp interface{}) (data []byte, err error) {
	client, err := GetClient(name)
	if nil != err {
		return nil, err
	}

	return client.JSGet(uri, rsp)
}

func JSPost(name, uri string, req, rsp interface{}) (data []byte, err error) {
	client, err := GetClient(name)
	if nil != err {
		return nil, err
	}

	return client.JSPost(uri, req, rsp)
}

func Remove(name string) {
	mtClient.RLock()
	_, ok := name2client[name]
	mtClient.RUnlock()

	if ok {
		mtClient.Lock()
		delete(name2client, name)
		mtClient.Unlock()
	}
}

func RegistryClient(client *HttpClient) {

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

func GetClient(name string) (client *HttpClient, err error) {

	client = getCacheClient(name)
	if nil != client {
		jklog.Debugw("find http client from caehe", "name", name)
		return client, nil
	}

	mtNewClient.Lock()
	defer mtNewClient.Unlock()

	client = getCacheClient(name)

	if nil == client {

		client, err = NewClient(name)
		if nil != err {
			return nil, err
		}

		saveClient(name, client)
	}

	return client, nil
}

func getCacheClient(name string) *HttpClient {
	mtClient.RLock()
	defer mtClient.RUnlock()

	client, ok := name2client[name]
	if !ok {
		return nil
	}

	return client
}

func saveClient(name string, client *HttpClient) {
	mtClient.Lock()
	defer mtClient.Unlock()

	name2client[name] = client
}

type reuquestParam struct {
	method  string
	uri     string
	enc     kithttp.EncodeRequestFunc
	dec     kithttp.DecodeResponseFunc
	request interface{}
}

type HttpClient struct {
	name string
	cfg  *ClientConfig

	consulInstancer  *kitconsul.Instancer
	consulEndpointer *sd.DefaultEndpointer
	reqEndPoint      endpoint.Endpoint

	actionEndPoint map[string]endpoint.Endpoint
	mtAction       sync.RWMutex

	consulClient kitconsul.Client
}

func NewClient(name string, ops ...ClientOption) (client *HttpClient, err error) {
	client = new(HttpClient)
	client.name = name
	client.actionEndPoint = map[string]endpoint.Endpoint{}
	client.cfg = newClientConfig(name, ops...)

	client.consulClient, err = jkregistry.NewConsulClient(client.name, client.cfg.RegOps...)
	if nil != err {
		jklog.Errorw("jkregistry.NewConsulClient fail", "name:", client.name, "cfg", *client.cfg, "err", err.Error())
		return nil, err
	}

	client.reqEndPoint = client.makeRuquestEndpoint()

	return client, nil
}

func (client *HttpClient) Get(uri string) (data []byte, err error) {
	return client.httpRequest(uri, "Get", nil, makeEncodeRequest)
}

func (client *HttpClient) Post(uri string, body []byte) (data []byte, err error) {
	return client.httpRequest(uri, "POST", body, makeEncodeRequest)
}

func (client *HttpClient) JSGet(uri string, rsp interface{}) (data []byte, err error) {
	return client.jsHttpRequest(uri, "Get", nil, rsp, makeEncodeRequest)
}

func (client *HttpClient) JSPost(uri string, req, rsp interface{}) (data []byte, err error) {
	return client.jsHttpRequest(uri, "POST", req, rsp, makeJSEncodeRequest)
}

func (client *HttpClient) Close() {

	if nil != client.consulEndpointer {
		client.consulEndpointer.Close()
	}

	if nil != client.consulInstancer {
		client.consulInstancer.Stop()
	}
}

func (client *HttpClient) httpRequest(uri, method string, req interface{}, makeEnc MakeEncodeRequestFunc) (data []byte, err error) {

	action := client.cfg.GetAction(uri)

	client.mtAction.RLock()
	reqEndPoint, ok := client.actionEndPoint[action]
	client.mtAction.RUnlock()
	if !ok {
		reqEndPoint = jkendpoint.Chain(client.reqEndPoint, action, client.cfg.ActionMiddlewares...)
	}

	reqParam := reuquestParam{uri: uri, method: method, request: req, enc: makeEnc(client.cfg), dec: DecodeReponse}

	resp, err := reqEndPoint(context.Background(), reqParam)
	if nil != err {
		return nil, err
	}

	if !ok {
		client.mtAction.Lock()
		client.actionEndPoint[action] = reqEndPoint
		client.mtAction.Unlock()
	}

	return resp.([]byte), nil
}

func (client *HttpClient) jsHttpRequest(uri, method string, req, rsp interface{}, makeEnc MakeEncodeRequestFunc) (data []byte, err error) {
	data, err = client.httpRequest(uri, method, req, makeEnc)
	if nil != err {
		jklog.Errorw("httpRequest fail", "name", client.name, "uri", uri, "method", method, "req", req, "err", err.Error())
		return data, err
	}

	err = json.Unmarshal(data, rsp)
	if nil != err {
		jklog.Errorw("json.Unmarshal fail", "name", client.name, "uri", uri, "method", method, "req", req, "data", string(data), "err", err.Error())
		return data, err
	}

	return data, nil
}

func (client *HttpClient) makeRequestFactory() sd.Factory {
	return func(instance string) (endpoint endpoint.Endpoint, closer io.Closer, err error) {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			reqParam, ok := request.(reuquestParam)
			if !ok {
				return nil, errors.New("the request is not reuquestParam")
			}

			if jkutils.HTTP == client.cfg.Scheme {
				if !strings.HasPrefix(instance, jkutils.HTTP) {
					instance = "http://" + instance
				}
			} else if jkutils.HTTPS == client.cfg.Scheme {
				if !strings.HasPrefix(instance, jkutils.HTTPS) {
					instance = "https://" + instance
				}
			} else {
				return nil, errors.New("Wrong scheme:" + client.cfg.Scheme)
			}

			tgt, err := url.Parse(instance + reqParam.uri)
			if err != nil {
				jklog.Errorw("url.Parse fail", "instance", instance, "uri", reqParam.uri, "method", reqParam.method, "err", err.Error())
				return nil, err
			}

			// tgt.Path = reqParam.uri
			jklog.Debugw("URL info", "tgt", tgt)

			return kithttp.NewClient(reqParam.method, tgt, reqParam.enc, reqParam.dec).Endpoint()(ctx, reqParam.request)
		}, nil, nil

	}
}

func (client *HttpClient) makeRuquestEndpoint() endpoint.Endpoint {

	client.consulInstancer = kitconsul.NewInstancer(client.consulClient, kitlog.InfowLogger, client.name, client.cfg.ConsulTags, client.cfg.PassingOnly)
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

	return lb.Retry(client.cfg.Retry, time.Duration(client.cfg.TimeOut)*time.Second, balancer)
}

func makeEncodeRequest(cfg *ClientConfig) kithttp.EncodeRequestFunc {
	return func(_ context.Context, req *http.Request, request interface{}) error {

		AppendHeader(req.Header, cfg.Header)

		if _, ok := request.([]byte); ok {
			req.Body = io.NopCloser(bytes.NewBuffer(request.([]byte)))
		}

		return nil
	}
}

func DecodeReponse(_ context.Context, resp *http.Response) (interface{}, error) {
	data, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	return data, err
}

func makeJSEncodeRequest(cfg *ClientConfig) kithttp.EncodeRequestFunc {
	return func(_ context.Context, req *http.Request, request interface{}) error {

		AppendHeader(req.Header, cfg.Header)

		req.Header.Add("Content-Type", "application/json;charset=utf-8")

		var buf bytes.Buffer
		req.Body = ioutil.NopCloser(&buf)

		return json.NewEncoder(&buf).Encode(request)
	}
}

// func makeJSDecodeReponse(respone interface{}) kithttp.EncodeResponseFunc {
// 	return func(_ context.Context, resp *http.Response) (interface{}, error) {

// 		err := json.NewDecoder(resp.Body).Decode(respone)
// 		defer resp.Close()

// 		return respone, err
// 	}
// }
