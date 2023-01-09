package http

import (
	"net/http"

	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jktrans "github.com/jkprj/jkfr/gokit/transport"
	jkendpoint "github.com/jkprj/jkfr/gokit/transport/endpoint"
	"github.com/jkprj/jkfr/gokit/utils"
	jklog "github.com/jkprj/jkfr/log"
	jkos "github.com/jkprj/jkfr/os"

	"golang.org/x/time/rate"
)

type ServerOption func(cfg *ServerConfig)
type GetActionFunc func(req *http.Request) string

type ServerConfig struct {
	ServerAddr          string     `json:"ServerAddr" toml:"ServerAddr"`
	BindAddr            string     `json:"BindAddr" toml:"BindAddr"`
	RateLimit           rate.Limit `json:"RateLimit" toml:"RateLimit"`
	PrometheusNameSpace string     `json:"PrometheusNameSpace" toml:"PrometheusNameSpace"`

	GetAction         GetActionFunc                 `json:"-" toml:"-"`
	RegOps            []jkregistry.RegOption        `json:"-" toml:"-"`
	ActionMiddlewares []jkendpoint.ActionMiddleware `json:"-" toml:"-"`
	ConfigPath        string

	tmpActionMiddlewares []jkendpoint.ActionMiddleware
}

type serverConfig struct {
	Cfg *ServerConfig `json:"Server" toml:"Server"`
}

func defaultServerConfig(serverName string) *ServerConfig {
	cfg := new(ServerConfig)
	cfg.GetAction = defaultServerGetAction
	cfg.RegOps = []jkregistry.RegOption{}

	cfg.ServerAddr = jkos.GetEnvString("S_SERVER_ADDR", "")
	cfg.BindAddr = jkos.GetEnvString("S_BIND_ADDR", "")
	cfg.PrometheusNameSpace = jkos.GetEnvString("S_PROMETHEUS_NAME_SPACE", serverName)
	cfg.RateLimit = rate.Limit(jkos.GetEnvInt("S_RATE_LIMIT", 0))

	cfg.ConfigPath = jkos.GetEnvString("S_CONFIG_PATH", "")
	if jkos.IsFileExists(cfg.ConfigPath) {
		ServerConfigFile(cfg.ConfigPath)(cfg)
	}

	return cfg
}

func loadServerDefaultClientConfig(name string, cfg *ServerConfig) {

	fileName := jkos.CurDir() + "/conf/" + name

	conf := fileName + ".toml"
	if jkos.IsFileExists(conf) {
		ServerConfigFile(conf)(cfg)
	}

	conf = fileName + ".json"
	if jkos.IsFileExists(conf) {
		ServerConfigFile(conf)(cfg)
	}
}

func defaultServerGetAction(req *http.Request) string {
	req.ParseForm()
	action := req.FormValue("Action")
	if "" != action {
		return action
	}

	action = req.FormValue("action")

	return action
}

func newServerConfig(serverName string, ops ...ServerOption) *ServerConfig {

	cfg := defaultServerConfig(serverName)
	loadServerDefaultClientConfig(serverName, cfg)
	for _, op := range ops {
		op(cfg)
	}

	cfg.ActionMiddlewares = jkendpoint.DefaultMiddleware(cfg.PrometheusNameSpace, jktrans.ROLE_SERVER, cfg.RateLimit)

	if 0 < len(cfg.tmpActionMiddlewares) {
		cfg.ActionMiddlewares = cfg.tmpActionMiddlewares
	}

	err := utils.ResetServerAddr(&cfg.ServerAddr, &cfg.BindAddr)
	if nil != err {
		jklog.Panicw("utils.ResetServerAddr error", "ServerAddr", cfg.ServerAddr, "BindAddr", cfg.BindAddr, "error", err)
	}

	return cfg
}

func ServerAddr(serverAddr string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.ServerAddr = serverAddr
	}
}

func BindAddr(bindAddr string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.BindAddr = bindAddr
	}
}

func ServerLimit(limit rate.Limit) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.RateLimit = limit
	}
}

func ServerPrometheusNameSpace(prometheusnamespace string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.PrometheusNameSpace = prometheusnamespace
	}
}

func ServerRegOption(regOps ...jkregistry.RegOption) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.RegOps = append(cfg.RegOps, regOps...)
	}
}

func ServerGetAction(getActionFunc GetActionFunc) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.GetAction = getActionFunc
	}
}

func ServerActionMiddlewares(actionMiddlewares ...jkendpoint.ActionMiddleware) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.tmpActionMiddlewares = append(cfg.tmpActionMiddlewares, actionMiddlewares...)
	}
}

func ServerConfigFile(cfgPath string) ServerOption {
	return func(cfg *ServerConfig) {

		cfg.ConfigPath = cfgPath
		config := serverConfig{Cfg: cfg}
		utils.ReadConfigFile(cfg.ConfigPath, &config)
	}
}
