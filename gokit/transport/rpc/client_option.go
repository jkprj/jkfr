package rpc

import (
	"jkfr/gokit/utils"
	"net/http"
	"net/rpc"

	jkregistry "jkfr/gokit/registry"
	jktrans "jkfr/gokit/transport"
	jkendpoint "jkfr/gokit/transport/endpoint"
	jkos "jkfr/os"

	"golang.org/x/time/rate"
)

type GetActionFunc func(uri string)
type ClientOption func(cfg *ClientConfig)

func AppendHeader(hDest, hSrc http.Header) {
	if nil != hDest && nil != hSrc {
		for k := range hSrc {
			hDest.Set(k, hSrc.Get(k))
		}
	}
}

type ClientConfig struct {
	RegOps            []jkregistry.RegOption        `json:"-" toml:"-"`
	ActionMiddlewares []jkendpoint.ActionMiddleware `json:"-" toml:"-"`
	Fatory            ClientFatory                  `json:"-" toml:"-"`
	AsyncCallChan     chan *rpc.Call                `json:"-" toml:"-"`
	ClientPem         []byte                        `json:"-" toml:"-"`
	ClientKey         []byte                        `json:"-" toml:"-"`
	ConfigPath        string

	ConsulTags          []string   `json:"ConsulTags" toml:"ConsulTags"`
	Strategy            string     `json:"Strategy" toml:"Strategy"`
	PrometheusNameSpace string     `json:"PrometheusNameSpace" toml:"PrometheusNameSpace"`
	Retry               int        `json:"Retry" toml:"Retry"`
	RetryIntervalMS     int        `json:"RetryIntervalMS" toml:"RetryIntervalMS"`
	RateLimit           rate.Limit `json:"RateLimit" toml:"RateLimit"`
	TimeOut             int        `json:"TimeOut" toml:"TimeOut"`
	PassingOnly         bool       `json:"PassingOnly" toml:"PassingOnly"`
	KeepAlive           bool       `json:"KeepAlive" toml:"KeepAlive"`
	PoolCap             int        `json:"PoolCap" toml:"PoolCap"`
	MaxCap              int        `json:"MaxCap" toml:"MaxCap"`
	DialTimeout         int        `json:"DialTimeout" toml:"DialTimeout"`
	IdleTimeout         int        `json:"IdleTimeout" toml:"IdleTimeout"`
	ReadTimeout         int        `json:"ReadTimeout" toml:"ReadTimeout"`
	WriteTimeout        int        `json:"WriteTimeout" toml:"WriteTimeout"`
	ClientPemFile       string     `json:"ClientPemFile" toml:"ClientPemFile"`
	ClientKeyFile       string     `json:"ClientKeyFile" toml:"ClientKeyFile"`

	tmpActionMiddlewares []jkendpoint.ActionMiddleware
}

type clientConfig struct {
	Cfg *ClientConfig `json:"Client" toml:"Client"`
}

func defaultClientConfig(name string) *ClientConfig {

	cfg := new(ClientConfig)

	cfg.RegOps = []jkregistry.RegOption{}
	cfg.ActionMiddlewares = []jkendpoint.ActionMiddleware{}
	cfg.AsyncCallChan = make(chan *rpc.Call, 10)
	cfg.Fatory = TcpClientFatory

	cfg.ConsulTags = jkos.GetEnvStrings("C_CONSUL_TAGS", ",", nil)
	cfg.Strategy = jkos.GetEnvString("C_STRATEGY", jktrans.STRATEGY_LEAST)
	cfg.PrometheusNameSpace = jkos.GetEnvString("C_PROMETHEUS_NAME_SPACE", name)
	cfg.Retry = jkos.GetEnvInt("C_RETRY", 3)
	cfg.RetryIntervalMS = jkos.GetEnvInt("C_RETRY_INTERVAL_MS", 1000)
	cfg.RateLimit = rate.Limit(jkos.GetEnvInt("C_RATE_LIMIT", 0))
	cfg.TimeOut = jkos.GetEnvInt("C_TIMEOUT", 60)
	cfg.PassingOnly = jkos.GetEnvBool("C_PASSING_ONLY", true)
	cfg.KeepAlive = jkos.GetEnvBool("C_KEEP_ALIVE", true)
	cfg.PoolCap = jkos.GetEnvInt("C_POOL_CAP", 5)
	cfg.MaxCap = jkos.GetEnvInt("C_MAX_CAP", 100)
	cfg.DialTimeout = jkos.GetEnvInt("C_DIAL_TIMEOUT", 10)
	cfg.IdleTimeout = jkos.GetEnvInt("C_IDLE_TIMEOUT", 600)
	cfg.ReadTimeout = jkos.GetEnvInt("C_READ_TIMEOUT", 60)
	cfg.WriteTimeout = jkos.GetEnvInt("C_WRITE_TIMEOUT", 60)

	cfg.ClientPemFile = jkos.GetEnvString("C_PEM_FILE", "")
	ClientPemFile(cfg.ClientPemFile)(cfg)

	cfg.ClientKeyFile = jkos.GetEnvString("C_KEY_FILE", "")
	ClientKeyFile(cfg.ClientKeyFile)(cfg)

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

func ClientDialTimeout(dialTimeout int) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.DialTimeout = dialTimeout
	}
}

func ClientIdleTimeout(idleTimeout int) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.IdleTimeout = idleTimeout
	}
}

func ClientReadTimeout(readTimeout int) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.ReadTimeout = readTimeout
	}
}

func ClientWriteTimeout(writeTimeout int) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.WriteTimeout = writeTimeout
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

func ClientAsyncCallChan(asyncCallChan chan *rpc.Call) ClientOption {
	return func(cfg *ClientConfig) {
		if nil != asyncCallChan && 0 != cap(asyncCallChan) {
			cfg.AsyncCallChan = asyncCallChan
		}
	}
}

func ClientCreateFatory(fatory ClientFatory) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.Fatory = fatory
	}
}

func ClientActionMiddlewares(actionMiddlewares ...jkendpoint.ActionMiddleware) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.tmpActionMiddlewares = append(cfg.tmpActionMiddlewares, actionMiddlewares...)
	}
}

func ClientPem(clientPem []byte) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.ClientPem = clientPem
	}
}

func ClientKey(clientKey []byte) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.ClientKey = clientKey
	}
}

func ClientPemFile(clientPemFile string) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.ClientPemFile = clientPemFile
		if jkos.IsFileExists(cfg.ClientPemFile) {
			tmpBuf, err := jkos.ReadFile(clientPemFile)
			if nil == err {
				cfg.ClientPem = tmpBuf
			}
		}

	}
}

func ClientKeyFile(clientKeyFile string) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.ClientKeyFile = clientKeyFile
		if jkos.IsFileExists(cfg.ClientKeyFile) {
			tmpBuf, err := jkos.ReadFile(clientKeyFile)
			if nil == err {
				cfg.ClientKey = tmpBuf
			}
		}

	}
}

func ClientConfigFile(cfgPath string) ClientOption {
	return func(cfg *ClientConfig) {

		cfg.ConfigPath = cfgPath
		config := clientConfig{Cfg: cfg}
		utils.ReadConfigFile(cfg.ConfigPath, &config)
	}
}
