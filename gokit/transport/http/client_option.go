package http

import (
	"net/http"
	"strings"

	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jktrans "github.com/jkprj/jkfr/gokit/transport"
	jkendpoint "github.com/jkprj/jkfr/gokit/transport/endpoint"
	"github.com/jkprj/jkfr/gokit/utils"
	jkos "github.com/jkprj/jkfr/os"

	kithttp "github.com/go-kit/kit/transport/http"
	"golang.org/x/time/rate"
)

type GetUriActionFunc func(uri string) string
type ClientOption func(cfg *ClientConfig)

func AppendHeader(hDest, hSrc http.Header) {
	if nil != hDest && nil != hSrc {
		for k := range hSrc {
			hDest.Set(k, hSrc.Get(k))
		}
	}
}

type ClientConfig struct {
	HttpClientOps     []kithttp.ClientOption        `json:"-" toml:"-"`
	RegOps            []jkregistry.RegOption        `json:"-" toml:"-"`
	ActionMiddlewares []jkendpoint.ActionMiddleware `json:"-" toml:"-"`
	Header            http.Header
	ConfigPath        string

	GetAction GetUriActionFunc `json:"-" toml:"-"`

	tmpActionMiddlewares []jkendpoint.ActionMiddleware

	ConsulTags          []string   `json:"ConsulTags" toml:"ConsulTags"`
	Strategy            string     `json:"Strategy" toml:"Strategy"`
	PrometheusNameSpace string     `json:"PrometheusNameSpace" toml:"PrometheusNameSpace"`
	Scheme              string     `json:"Scheme" toml:"Scheme"`
	Retry               int        `json:"Retry" toml:"Retry"`
	RateLimit           rate.Limit `json:"RateLimit" toml:"RateLimit"`
	TimeOut             int        `json:"TimeOut" toml:"TimeOut"`
	PassingOnly         bool       `json:"PassingOnly" toml:"PassingOnly"`
}

type clientConfig struct {
	Cfg *ClientConfig `json:"Client" toml:"Client"`
}

func defaultClientGetAction(uri string) string {
	lowUri := strings.ToLower(uri)
	bgIndex := strings.Index(lowUri, "action=")
	if 0 > bgIndex {
		return ""
	}

	bgIndex = bgIndex + len("action=")
	endIndex := strings.Index(lowUri[bgIndex:], "&")
	if endIndex > 0 {
		endIndex += bgIndex
		return uri[bgIndex:endIndex]
	} else if 0 > endIndex {
		return uri[bgIndex:]
	}

	return ""
}

func defaultClientConfig(name string) *ClientConfig {
	cfg := new(ClientConfig)
	cfg.HttpClientOps = []kithttp.ClientOption{}
	cfg.RegOps = []jkregistry.RegOption{}
	cfg.ActionMiddlewares = []jkendpoint.ActionMiddleware{}
	cfg.GetAction = defaultClientGetAction

	cfg.ConsulTags = jkos.GetEnvStrings("C_CONSUL_TAGS", ",", nil)
	cfg.Strategy = jkos.GetEnvString("C_STRATEGY", jktrans.STRATEGY_ROUND)
	cfg.PrometheusNameSpace = jkos.GetEnvString("C_PROMETHEUS_NAME_SPACE", name)
	cfg.Scheme = jkos.GetEnvString("C_SCHEME", jktrans.HTTP)
	cfg.Retry = jkos.GetEnvInt("C_RETRY", 3)
	cfg.RateLimit = rate.Limit(jkos.GetEnvInt("C_RATE_LIMIT", 0))
	cfg.TimeOut = jkos.GetEnvInt("C_TIME_OUT", 60)
	cfg.PassingOnly = jkos.GetEnvBool("C_PASSING_ONLY", true)

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

func ClientScheme(scheme string) ClientOption {
	return func(cfg *ClientConfig) {
		if jktrans.HTTP == scheme || jktrans.HTTPS == scheme {
			cfg.Scheme = scheme
		}
	}
}

func ClientRegOption(regOps ...jkregistry.RegOption) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.RegOps = append(cfg.RegOps, regOps...)
	}
}

func ClientGetAction(getActionFunc GetUriActionFunc) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.GetAction = getActionFunc
	}
}

func ClientHttpClientOps(httpClientOps ...kithttp.ClientOption) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.HttpClientOps = append(cfg.HttpClientOps, httpClientOps...)
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

func ClientHeader(header http.Header) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.Header = http.Header{}
		AppendHeader(cfg.Header, header)
	}
}

func ClientPassingOnly(passingOnly bool) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.PassingOnly = passingOnly
	}
}

func ClientActionMiddlewares(actionMiddlewares ...jkendpoint.ActionMiddleware) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.tmpActionMiddlewares = append(cfg.tmpActionMiddlewares, actionMiddlewares...)
	}
}

func ClientConfigFile(cfgPath string) ClientOption {
	return func(cfg *ClientConfig) {

		cfg.ConfigPath = cfgPath
		config := clientConfig{Cfg: cfg}
		utils.ReadConfigFile(cfg.ConfigPath, &config)
	}
}
