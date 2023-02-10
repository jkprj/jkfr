package grpc

import (
	"compress/gzip"
	"net/rpc"
	"time"

	"google.golang.org/grpc/keepalive"

	"google.golang.org/grpc"

	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jktrans "github.com/jkprj/jkfr/gokit/transport"
	jkendpoint "github.com/jkprj/jkfr/gokit/transport/endpoint"
	"github.com/jkprj/jkfr/gokit/utils"
	jkos "github.com/jkprj/jkfr/os"

	"golang.org/x/time/rate"
)

type GetActionFunc func(uri string)
type ClientOption func(cfg *ClientConfig)

type UCall struct {
	rpc.Call
	Done chan *UCall
}

func (call *UCall) done() {
	select {
	case call.Done <- call:
	default:
		// log debug
	}
}

type ClientConfig struct {
	RegOps            []jkregistry.RegOption        `json:"-" toml:"-"`
	GRPCDialOps       []grpc.DialOption             `json:"-" toml:"-"`
	ActionMiddlewares []jkendpoint.ActionMiddleware `json:"-" toml:"-"`
	AsyncCallChan     chan *UCall                   `json:"-" toml:"-"`
	ConfigPath        string

	ConsulTags          []string   `json:"ConsulTags" toml:"ConsulTags"`
	Strategy            string     `json:"Strategy" toml:"Strategy"`
	PrometheusNameSpace string     `json:"PrometheusNameSpace" toml:"PrometheusNameSpace"`
	Retry               int        `json:"Retry" toml:"Retry"`
	RetryIntervalMS     int        `json:"RetryIntervalMS" toml:"RetryIntervalMS"`
	RateLimit           rate.Limit `json:"RateLimit" toml:"RateLimit"`
	TimeOut             int        `json:"TimeOut" toml:"TimeOut"`
	PoolCap             int        `json:"PoolCap" toml:"PoolCap"`
	MaxCap              int        `json:"MaxCap" toml:"MaxCap"`
	PassingOnly         bool       `json:"PassingOnly" toml:"PassingOnly"`
	KeepAlive           bool       `json:"KeepAlive" toml:"KeepAlive"`

	tmpActionMiddlewares []jkendpoint.ActionMiddleware
}

type GRPCConfig struct {
	WriteBufferSize  int  `json:"WriteBufferSize" toml:"WriteBufferSize"`
	ReadBufferSize   int  `json:"ReadBufferSize" toml:"ReadBufferSize"`
	MaxMsgSize       int  `json:"MaxMsgSize" toml:"MaxMsgSize"` // 和MaxRecvMsgSize一样，可能后面grpc库会不支持
	MaxSendMsgSize   int  `json:"MaxSendMsgSize" toml:"MaxSendMsgSize"`
	MaxRecvMsgSize   int  `json:"MaxRecvMsgSize" toml:"MaxRecvMsgSize"`
	Timeout          int  `json:"Timeout" toml:"Timeout"`
	EnableCompressor bool `json:"EnableCompressor" toml:"EnableCompressor"`
	CompressorLevel  int  `json:"CompressorLevel" toml:"CompressorLevel"`

	InitialWindowSize     int32 `json:"InitialWindowSize" toml:"InitialWindowSize"`
	InitialConnWindowSize int32 `json:"InitialConnWindowSize" toml:"InitialConnWindowSize"`
}

type GRPCKeepaliveParam struct {
	PermitWithoutStream bool `json:"PermitWithoutStream" toml:"PermitWithoutStream"`
	Time                int  `json:"Time" toml:"Time"`
	Timeout             int  `json:"Timeout" toml:"Timeout"`
}

type clientConfig struct {
	Cfg     *ClientConfig      `json:"Client" toml:"Client"`
	GRPCCfg GRPCConfig         `json:"GRPC" toml:"GRPC"`
	KPCfg   GRPCKeepaliveParam `json:"Keepalive" toml:"Keepalive"`
}

func defaultClientConfig(name string) *ClientConfig {
	cfg := new(ClientConfig)
	cfg.RegOps = []jkregistry.RegOption{}
	cfg.ActionMiddlewares = []jkendpoint.ActionMiddleware{}
	cfg.AsyncCallChan = make(chan *UCall, 100)
	cfg.GRPCDialOps = []grpc.DialOption{grpc.WithInsecure(), grpc.WithTimeout(time.Second * 5)}

	cfg.ConsulTags = jkos.GetEnvStrings("C_CONSUL_TAGS", ",", nil)
	cfg.Strategy = jkos.GetEnvString("C_STRATEGY", jktrans.STRATEGY_LEAST)
	cfg.PrometheusNameSpace = jkos.GetEnvString("C_PROMETHEUS_NAME_SPACE", name)
	cfg.Retry = jkos.GetEnvInt("C_RETRY", 3)
	cfg.RetryIntervalMS = jkos.GetEnvInt("C_RETRY_INTERVAL_MS", 1000)
	cfg.RateLimit = rate.Limit(jkos.GetEnvInt("C_RATE_LIMIT", 0))
	cfg.TimeOut = jkos.GetEnvInt("C_TIME_OUT", 60)
	cfg.PoolCap = jkos.GetEnvInt("C_POOL_CAP", 2)
	cfg.MaxCap = jkos.GetEnvInt("C_MAX_CAP", 32)
	cfg.PassingOnly = jkos.GetEnvBool("C_PASSING_ONLY", true)
	cfg.KeepAlive = jkos.GetEnvBool("C_KEEP_ALIVE", true)

	tmpCfg := clientConfig{}
	tmpCfg.GRPCCfg.WriteBufferSize = jkos.GetEnvInt("C_WRITE_BUFFER_SIZE", 0)
	tmpCfg.GRPCCfg.ReadBufferSize = jkos.GetEnvInt("C_READ_BUFFER_SIZE", 0)
	tmpCfg.GRPCCfg.MaxMsgSize = jkos.GetEnvInt("C_MAX_MSG_SIZE", 0)
	tmpCfg.GRPCCfg.MaxSendMsgSize = jkos.GetEnvInt("C_MAX_SEND_MSG_SIZE", 0)
	tmpCfg.GRPCCfg.MaxRecvMsgSize = jkos.GetEnvInt("C_MAX_RECV_MSG_SIZE", 0)
	tmpCfg.GRPCCfg.Timeout = jkos.GetEnvInt("C_TIMEOUT", 0)
	tmpCfg.GRPCCfg.EnableCompressor = jkos.GetEnvBool("C_ENABLE_COMPRESSOR", false)
	tmpCfg.GRPCCfg.CompressorLevel = jkos.GetEnvInt("C_COMPRESSOR_LEVEL", gzip.BestCompression)
	tmpCfg.GRPCCfg.InitialWindowSize = int32(jkos.GetEnvInt("C_INITIAL_WINDOW_SIZE", 0))
	tmpCfg.GRPCCfg.InitialConnWindowSize = int32(jkos.GetEnvInt("C_INITIAL_CONN_WINDOW_SIZE", 0))

	tmpCfg.KPCfg.PermitWithoutStream = jkos.GetEnvBool("C_KEEPALIVE_PERMIT_WITHOUT_STREAM", false)
	tmpCfg.KPCfg.Time = jkos.GetEnvInt("C_KEEPALIVE_TIME", 0)
	tmpCfg.KPCfg.Timeout = jkos.GetEnvInt("C_KEEPALIVE_TIMEOUT", 0)

	appendGRPCConfig(cfg, &tmpCfg)

	cfg.ConfigPath = jkos.GetEnvString("C_CONFIG_PATH", "")
	if jkos.IsFileExists(cfg.ConfigPath) {
		ClientConfigFile(cfg.ConfigPath)(cfg)
	}

	return cfg
}

func loadDefaultClientConfig(name string, cfg *ClientConfig) {

	fileName := jkos.CurDir() + "/conf/" + name

	conf := fileName + ".toml"
	if jkos.IsFileExists(conf) {
		ClientConfigFile(conf)(cfg)
	}

	conf = fileName + ".json"
	if jkos.IsFileExists(conf) {
		ClientConfigFile(conf)(cfg)
	}

}

func appendGRPCConfig(cfg *ClientConfig, tmpCfg *clientConfig) {

	if 0 != tmpCfg.GRPCCfg.WriteBufferSize {
		cfg.GRPCDialOps = append(cfg.GRPCDialOps, grpc.WithWriteBufferSize(tmpCfg.GRPCCfg.WriteBufferSize))
	}

	if 0 != tmpCfg.GRPCCfg.ReadBufferSize {
		cfg.GRPCDialOps = append(cfg.GRPCDialOps, grpc.WithReadBufferSize(tmpCfg.GRPCCfg.ReadBufferSize))
	}

	if 0 != tmpCfg.GRPCCfg.MaxMsgSize {
		cfg.GRPCDialOps = append(cfg.GRPCDialOps, grpc.WithMaxMsgSize(tmpCfg.GRPCCfg.MaxMsgSize))
	}

	callOps := []grpc.CallOption{}

	if 0 != tmpCfg.GRPCCfg.MaxSendMsgSize {
		callOps = append(callOps, grpc.MaxCallSendMsgSize(tmpCfg.GRPCCfg.MaxSendMsgSize))

	}
	if 0 != tmpCfg.GRPCCfg.MaxRecvMsgSize {
		callOps = append(callOps, grpc.MaxCallRecvMsgSize(tmpCfg.GRPCCfg.MaxRecvMsgSize))
	}

	cfg.GRPCDialOps = append(cfg.GRPCDialOps, grpc.WithDefaultCallOptions(callOps...))

	if 0 != tmpCfg.GRPCCfg.InitialWindowSize {
		cfg.GRPCDialOps = append(cfg.GRPCDialOps, grpc.WithInitialWindowSize(tmpCfg.GRPCCfg.InitialWindowSize))
	}

	if 0 != tmpCfg.GRPCCfg.InitialConnWindowSize {
		cfg.GRPCDialOps = append(cfg.GRPCDialOps, grpc.WithInitialConnWindowSize(tmpCfg.GRPCCfg.InitialConnWindowSize))
	}

	if 0 != tmpCfg.GRPCCfg.Timeout {
		timeOut := time.Duration(tmpCfg.GRPCCfg.Timeout) * time.Second
		cfg.GRPCDialOps = append(cfg.GRPCDialOps, grpc.WithTimeout(timeOut))
	}

	if true == tmpCfg.GRPCCfg.EnableCompressor {
		if tmpCfg.GRPCCfg.CompressorLevel < gzip.DefaultCompression || tmpCfg.GRPCCfg.CompressorLevel > gzip.BestCompression {
			tmpCfg.GRPCCfg.CompressorLevel = gzip.BestCompression
		}

		compressor, _ := grpc.NewGZIPCompressorWithLevel(tmpCfg.GRPCCfg.CompressorLevel)
		cfg.GRPCDialOps = append(cfg.GRPCDialOps, grpc.WithCompressor(compressor))
		cfg.GRPCDialOps = append(cfg.GRPCDialOps, grpc.WithDecompressor(grpc.NewGZIPDecompressor()))
	}

	if false != tmpCfg.KPCfg.PermitWithoutStream ||
		0 != tmpCfg.KPCfg.Time ||
		0 != tmpCfg.KPCfg.Timeout {

		keepaliveParm := keepalive.ClientParameters{}
		keepaliveParm.PermitWithoutStream = tmpCfg.KPCfg.PermitWithoutStream
		keepaliveParm.Time = time.Duration(tmpCfg.KPCfg.Time) * time.Second
		keepaliveParm.Timeout = time.Duration(tmpCfg.KPCfg.Timeout) * time.Second

		cfg.GRPCDialOps = append(cfg.GRPCDialOps, grpc.WithKeepaliveParams(keepaliveParm))
	}

}

func newClientConfig(name string, ops ...ClientOption) *ClientConfig {
	cfg := defaultClientConfig(name)
	loadDefaultClientConfig(name, cfg)
	for _, op := range ops {
		op(cfg)
	}

	cfg.ActionMiddlewares = jkendpoint.DefaultMiddleware(cfg.PrometheusNameSpace, jktrans.ROLE_CLIENT, cfg.RateLimit)

	if 0 < len(cfg.tmpActionMiddlewares) {
		cfg.ActionMiddlewares = cfg.tmpActionMiddlewares
	}

	return cfg
}

func ClientLimit(limit rate.Limit) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.RateLimit = limit
	}
}

func ClientPrometheusNameSpace(prometheusnamespace string) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.PrometheusNameSpace = prometheusnamespace
	}
}

func ClientRegOption(regOps ...jkregistry.RegOption) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.RegOps = append(cfg.RegOps, regOps...)
	}
}

func ClientConsulTags(tags ...string) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.ConsulTags = append(cfg.ConsulTags, tags...)
	}
}

func ClientStrategy(strategy string) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.Strategy = strategy
	}
}

func ClientRetry(retry int) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.Retry = retry
	}
}

func ClientRetryIntervalMS(retryIntervalMS int) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.RetryIntervalMS = retryIntervalMS
	}
}

func ClientRateLimit(rateLimit int) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.RateLimit = rate.Limit(rateLimit)
	}
}

func ClientPassingOnly(passingOnly bool) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.PassingOnly = passingOnly
	}
}

func ClientKeepAlive(keepAlive bool) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.KeepAlive = keepAlive
	}
}

func ClientTimeOut(timeOut int) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.TimeOut = timeOut
	}
}

func ClientPoolCap(poolCap int) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.PoolCap = poolCap
	}
}

func ClientMaxCap(maxCap int) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.MaxCap = maxCap
	}
}

func ClientActionMiddlewares(actionMiddlewares ...jkendpoint.ActionMiddleware) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.tmpActionMiddlewares = append(cfg.tmpActionMiddlewares, actionMiddlewares...)
	}
}

func ClientGRPCDialOps(grpcDialOps ...grpc.DialOption) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.GRPCDialOps = append(cfg.GRPCDialOps, grpcDialOps...)
	}
}

func ClientAsyncCallChan(asyncCallChan chan *UCall) ClientOption {
	return func(cfg *ClientConfig) {
		if nil != asyncCallChan {
			cfg.AsyncCallChan = asyncCallChan
		}
	}
}

func ClientConfigFile(cfgPath string) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.ConfigPath = cfgPath
		config := clientConfig{Cfg: cfg}
		utils.ReadConfigFile(cfg.ConfigPath, &config)
		appendGRPCConfig(cfg, &config)

		cfg.RegOps = append(cfg.RegOps, jkregistry.WithFile(cfgPath))
	}
}
