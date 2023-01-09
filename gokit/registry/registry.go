package registry

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/jkprj/jkfr/gokit/utils"
	jklog "github.com/jkprj/jkfr/log"
	unet "github.com/jkprj/jkfr/net"

	"github.com/go-kit/kit/endpoint"
	kitcosul "github.com/go-kit/kit/sd/consul"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	consulapi "github.com/hashicorp/consul/api"
	"golang.org/x/time/rate"
)

type SubscribeRegistrarConfigInited func(regCfg *RegConfig)

type HealthCheckResponse struct {
	status int `json:"status"`
}

var healthCheckHttpServerAddr string = ""    // consul健康检查绑定服务地址
var healthCheckHttpServer *http.Server = nil // consul健康检查服务http对象
var registrySvrOnce sync.Once
var mapConsulRegRWMutex sync.RWMutex

var healthCheckRateLimit *rate.Limiter = rate.NewLimiter(200, 10) // consul健康检查服务限流器

var mapConsulReg map[string]kitcosul.Client = make(map[string]kitcosul.Client) // 缓存consulc_client对象

var subscribeConfigInitedHandles []SubscribeRegistrarConfigInited = make([]SubscribeRegistrarConfigInited, 0, 10)

func AppendNotifyConfigInitedHandle(handle SubscribeRegistrarConfigInited) {
	if nil != handle {
		subscribeConfigInitedHandles = append(subscribeConfigInitedHandles, handle)
	}
}

type Registrar struct {
	client       kitcosul.Client
	registration *consulapi.AgentServiceRegistration
}

// 构造Registrar对象
func NewRegistrar(client kitcosul.Client, r *consulapi.AgentServiceRegistration) *Registrar {
	return &Registrar{client: client, registration: r}
}

// 注册服务
func (p *Registrar) Register() error {
	return p.client.Register(p.registration)
}

// 注销服务
func (p *Registrar) Deregister() error {
	return p.client.Deregister(p.registration)
}

// 创建consulc_lient对象
func NewConsulClient(name string, ops ...RegOption) (consulClient kitcosul.Client, err error) {
	consulClient, _, err = newConsulClient(name, ops...)
	if nil != err {
		return nil, err
	}

	return consulClient, nil
}

// 创建consulc_lient对象
func newConsulClient(name string, ops ...RegOption) (consulClient kitcosul.Client, regCfg *RegConfig, err error) {

	regCfg, consulClientCfg := getConfig(name, ops...)

	consulClient = getConsulReg(consulClientCfg.Address)
	if nil != consulClient {
		return consulClient, regCfg, nil
	}

	consulApiClient, err := consulapi.NewClient(consulClientCfg)
	if nil != err {
		jklog.Errorw("consulapi.NewClient fail", "consulClientCfg", consulClientCfg, "err", err.Error())
		return nil, nil, err
	}

	consulClient = kitcosul.NewClient(consulApiClient)

	pushConsulReg(consulClientCfg.Address, consulClient)

	return consulClient, regCfg, nil
}

// 注册，由于consul deregister其他服务时，经常会导致其他服务的健康检查也deregister，
// 这时就算服务异常，consul就无法发现服务是否正常，这里就使用每隔段时间就重新注册一次来解决
func do_register(registry *Registrar, regObj *consulapi.AgentServiceRegistration) {

	go func() {
		for {
			err := registry.Register()
			if nil != err {
				jklog.Errorw("RegistryServer fail", "regObj", regObj, "err", err)
				time.Sleep(time.Second * 5)
			} else {
				jklog.Infow("RegistryServer succ", "regObj", regObj)
				time.Sleep(time.Hour)
			}
		}
	}()

}

// 注册服务
// Parameters :
// name 服务名称
// ops  服务注册选项，如果重复指定选项，后面的选项会替换前面的选项
func RegistryServer(name string, ops ...RegOption) (Registry *Registrar, err error) {

	consulClient, regCfg, err := newConsulClient(name, ops...)
	if nil != err {
		jklog.Error("newConsulClient fail", "name", regCfg.ServerName, "addr", regCfg.ServerAddr, "err", err)
		return nil, err
	}

	regCfg.SvcHost, regCfg.SvcPort, err = unet.ParseHostAddr(regCfg.ServerAddr)
	if nil != err {
		jklog.Error("ParseHostAddr fail", "name", name, "addr", regCfg.ServerAddr, "err", err)
		return nil, err
	}

	// 端口0说明没有绑定服务，启动http作为consul检查检查
	if 0 == regCfg.SvcPort {
		runHealthCheckServer(regCfg)
	}

	regObj := makeConsulAgentServiceRegistration(name, regCfg.SvcHost, regCfg.SvcPort, regCfg)

	registry := NewRegistrar(consulClient, regObj)
	do_register(registry, regObj)
	// err = registry.Register()
	// if nil != err {
	// 	jklog.Errorw("RegistryServer fail", "regObj", regObj, "err", err)
	// 	return nil, err
	// }

	return registry, nil
}

// 注册服务，注册时指定服务绑定地址，使用该函数注册服务后，再在环境变量，配置文件或函数参数ops指定服务地址都将无效，最终会替换为svrAddr
// Parameters :
// name:服务名称
// svrAddr:指定服务绑定地址
// ops:服务注册选项，后面的选项会替换前面的选项
func RegistryServerWithServerAddr(name, svrAddr string, ops ...RegOption) (Registry *Registrar, err error) {

	opts := []RegOption{}
	opts = append(opts, ops...)
	opts = append(opts, WithServerAddr(svrAddr))

	return RegistryServer(name, opts...)
}

// 读取配置，default->环境变量->配置文件->option
func getConfig(name string, ops ...RegOption) (regCfg *RegConfig, consulClientCfg *consulapi.Config) {
	regCfg = defaultconfig(name)
	consulClientCfg = consulapi.DefaultConfig()

	regCfg.WaitTime = int(consulClientCfg.WaitTime / time.Second)

	for _, op := range ops {
		op(regCfg)
	}

	consulClientCfg.WaitTime = time.Duration(regCfg.WaitTime) * time.Second
	consulClientCfg.Address = regCfg.ConsulAddr
	utils.ResetServerAddr(&regCfg.HealthCheckAddr, &regCfg.HealthCheckBindAddr)
	// consulClientCfg.Namespace = regCfg.Namespace // 似乎企业版才支持Namespace

	if "" != regCfg.UserName {
		consulClientCfg.HttpAuth = &consulapi.HttpBasicAuth{Username: regCfg.UserName, Password: regCfg.Password}
	}

	for _, handle := range subscribeConfigInitedHandles {
		handle(regCfg)
	}

	return
}

// 启动consul健康检查服务，该服务只会启动一次，再有服务注册不会再启动，并且相关服务选项只有第一次注册时指定有效
func runHealthCheckServer(regCfg *RegConfig) (err error) {

	registrySvrOnce.Do(func() {

		httpSvr := makeHealthCheckHttpServer(regCfg)
		errChan := goRunHttpServer(regCfg, httpSvr)

		select {
		case err = <-errChan:
			// jklog.Panicw("Run health check server fail", "regCfg", regCfg, "err", err.Error())
		case <-time.After(time.Second):
			// time out no error, complete!!
		}

	})

	return err
}

// 构造consul健康检查服务http服务句柄
func makeHealthCheckHttpServer(regCfg *RegConfig) *http.Server {

	var err error
	healthCheckHttpServerAddr, _, err = unet.GetRandomHostAddr(regCfg.HealthCheckAddr)
	if nil != err {
		jklog.Panicw("unet.GetRandomHostAddr fail", "HealthCheckAddr", regCfg.HealthCheckAddr, "error", err)
		return nil
	}
	jklog.Info("healthCheckHttpServerAddr:", healthCheckHttpServerAddr)

	router := mux.NewRouter()
	router.Methods("Get").Path("/health").Handler(kithttp.NewServer(
		makeHealthCheckEndPoint(),
		decodeHealthCheckRequest,
		encodeHealthCheckResponse,
	))

	healthCheckHttpServer = &http.Server{Addr: healthCheckHttpServerAddr}
	healthCheckHttpServer.Handler = router

	return healthCheckHttpServer
}

// 异步启动consul健康检查http服务
func goRunHttpServer(regCfg *RegConfig, httpSvr *http.Server) chan error {

	errChan := make(chan error, 2)

	go func() {
		err := httpSvr.ListenAndServe()
		if nil != err {
			errChan <- err

			jklog.Panicw("httpSvr.ListenAndServe fail", "regCfg", regCfg, "httpSvr", httpSvr, "err", err.Error())
		}
	}()

	return errChan
}

// 构造consul健康检查频率endpoint
func makeHealthCheckEndPoint() endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {

		healthCheckRateLimit.Wait(ctx)
		// do statistics request
		// req := request.(*http.Request)
		// req.ParseForm()
		// jklog.Debugw("consel HealthCheck", "host", req.RemoteAddr)

		return HealthCheckResponse{status: 1}, nil
	}
}

func encodeHealthCheckResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", "application/json;charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

func decodeHealthCheckRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	return r, nil
}

// 构造AgentServiceRegistration对象
func makeConsulAgentServiceRegistration(name, svcHost string, svcPort int, regCfg *RegConfig) *consulapi.AgentServiceRegistration {

	id := name + "_" + svcHost + ":" + strconv.Itoa(svcPort)

	asCheck := consulapi.AgentServiceCheck{
		CheckID:                        "check_" + id,
		Interval:                       strconv.Itoa(regCfg.HealthCheckInterval) + "s",
		Timeout:                        strconv.Itoa(regCfg.HealthCheckTimeOut) + "s",
		Notes:                          "Consul check service health status.",
		DeregisterCriticalServiceAfter: strconv.Itoa(regCfg.DeregisterCriticalServiceAfter) + "m",
	}

	if 0 == svcPort {
		hostCheckUrl, port, _ := unet.ParseHostAddr(healthCheckHttpServerAddr)
		if "" == hostCheckUrl {
			hostCheckUrl = svcHost
		}
		hostCheckUrl = hostCheckUrl + ":" + strconv.Itoa(port)
		asCheck.HTTP = "http://" + hostCheckUrl + "/health"

		jklog.Infow("health check info", "hostCheckUrl", hostCheckUrl, "svchost", svcHost, "svcPort", svcPort)
	} else {
		asCheck.TCP = regCfg.ServerAddr
	}

	tags := regCfg.ConsulTags
	tags = append(tags, name)
	tags = append(tags, regCfg.ServerAddr)

	regObj := new(consulapi.AgentServiceRegistration)
	regObj.ID = id
	regObj.Name = name
	regObj.Tags = tags
	regObj.Address = svcHost
	regObj.Port = svcPort
	regObj.Check = &asCheck

	return regObj
}

// 缓存consul_client对象，避免重复创建
func pushConsulReg(consulAddr string, consulClient kitcosul.Client) {
	mapConsulRegRWMutex.Lock()
	defer mapConsulRegRWMutex.Unlock()

	mapConsulReg[consulAddr] = consulClient
}

// 从缓存中获取consul_client对象
func getConsulReg(consulAddr string) kitcosul.Client {
	mapConsulRegRWMutex.RLock()
	defer mapConsulRegRWMutex.RUnlock()

	if consulClient, ok := mapConsulReg[consulAddr]; ok {
		return consulClient
	}

	return nil
}

// 发现服务
// Parameters :
// service:服务名称
// ops:服务注册选项，后面的选项会替换前面的选项
func Services(service string, ops ...RegOption) ([]*consulapi.ServiceEntry, *consulapi.QueryMeta, error) {

	consulClient, regCfg, err := newConsulClient(service, ops...)
	if nil != err {
		return nil, nil, err
	}

	if len(regCfg.ConsulTags) > 0 {
		return consulClient.Service(service, regCfg.ConsulTags[0], regCfg.PassingOnly, regCfg.QueryOpts)
	}

	return consulClient.Service(service, "", regCfg.PassingOnly, regCfg.QueryOpts)
}
