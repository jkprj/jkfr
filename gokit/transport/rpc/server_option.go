package rpc

import (
	"net/rpc"

	jkregistry "jkfr/gokit/registry"
	jktrans "jkfr/gokit/transport"
	jkendpoint "jkfr/gokit/transport/endpoint"
	"jkfr/gokit/utils"
	jklog "jkfr/log"
	jkos "jkfr/os"

	"golang.org/x/time/rate"
)

type ServerOption func(cfg *ServerConfig)

type ServerConfig struct {
	ServerAddr          string     `json:"ServerAddr" toml:"ServerAddr"`
	BindAddr            string     `json:"BindAddr" toml:"BindAddr"`
	ServerPemFile       string     `json:"ServerPemFile" toml:"ServerPemFile"`
	ServerKeyFile       string     `json:"ServerKeyFile" toml:"ServerKeyFile"`
	ClientPemFile       string     `json:"ClientPemFile" toml:"ClientPemFile"`
	PrometheusNameSpace string     `json:"PrometheusNameSpace" toml:"PrometheusNameSpace"`
	RateLimit           rate.Limit `json:"RateLimit" toml:"RateLimit"`
	RpcPath             string     `json:"RpcPath" toml:"RpcPath"`
	RpcDebugPath        string     `json:"RpcDebugPath" toml:"RpcDebugPath"`
	RpcName             string     `json:"RpcName" toml:"RpcName"`
	Codec               string     `json:"Codec" toml:"Codec"`

	ServerPem []byte `json:"-" toml:"-"`
	ServerKey []byte `json:"-" toml:"-"`
	ClientPem []byte `json:"-" toml:"-"`

	ConfigPath string

	RegOps         []jkregistry.RegOption `json:"-" toml:"-"`
	ListenerFatory CreateListenerFunc     `json:"-" toml:"-"`
	ServerRun      ServerRunFunc          `json:"-" toml:"-"`

	ActionMiddlewares    []jkendpoint.ActionMiddleware `json:"-" toml:"-"`
	tmpActionMiddlewares []jkendpoint.ActionMiddleware
}

type serverConfig struct {
	Cfg *ServerConfig `json:"Server" toml:"Server"`
}

func defaultServerConfig(name string) *ServerConfig {
	cfg := new(ServerConfig)
	cfg.RegOps = []jkregistry.RegOption{}
	cfg.ListenerFatory = TCPListenerFatory
	cfg.ServerRun = RunServerWithTcp

	cfg.ServerAddr = jkos.GetEnvString("S_SERVER_ADDR", "")
	cfg.BindAddr = jkos.GetEnvString("S_BIND_ADDR", "")
	cfg.PrometheusNameSpace = jkos.GetEnvString("S_PROMETHEUS_NAME_SPACE", name)
	cfg.RpcPath = jkos.GetEnvString("S_RPC_PATH", rpc.DefaultRPCPath)
	cfg.RpcName = jkos.GetEnvString("S_RPC_NAME", "")
	cfg.Codec = jkos.GetEnvString("S_CODEC", jktrans.CODEC_GOB)
	cfg.RpcDebugPath = jkos.GetEnvString("S_RPC_DEBUG_PATH", rpc.DefaultDebugPath)
	cfg.RateLimit = rate.Limit(jkos.GetEnvInt("S_RATE_LIMIT", 0))

	cfg.ServerPemFile = jkos.GetEnvString("S_PEM_FILE", "")
	ServerPemFile(cfg.ServerPemFile)(cfg)

	cfg.ServerKeyFile = jkos.GetEnvString("S_KEY_FILE", "")
	ServerKeyFile(cfg.ServerKeyFile)(cfg)

	cfg.ClientPemFile = jkos.GetEnvString("S_CLIENT_KEY_FILE", "")
	ServerClientPemFile(cfg.ClientPemFile)(cfg)

	cfg.ConfigPath = jkos.GetEnvString("S_CONFIG_PATH", "")
	if jkos.IsFileExists(cfg.ConfigPath) {
		ServerConfigFile(cfg.ConfigPath)(cfg)
	}

	return cfg
}

func loadDefaultServerConfig(name string, cfg *ServerConfig) {

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

func newServerConfig(name string, ops ...ServerOption) *ServerConfig {

	cfg := defaultServerConfig(name)
	loadDefaultServerConfig(name, cfg)
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

func ServerPem(serverPem []byte) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.ServerPem = serverPem
	}
}

func ServerKey(serverKey []byte) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.ServerKey = serverKey
	}
}

func ServerClientPem(clientPem []byte) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.ClientPem = clientPem
	}
}

func ServerPemFile(serverPemFile string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.ServerPemFile = serverPemFile
		if jkos.IsFileExists(cfg.ServerPemFile) {
			tmpBuf, err := jkos.ReadFile(serverPemFile)
			if nil == err {
				cfg.ServerPem = tmpBuf
			}
		}

	}
}

func ServerKeyFile(serverKeyFile string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.ServerKeyFile = serverKeyFile
		if jkos.IsFileExists(cfg.ServerKeyFile) {
			tmpBuf, err := jkos.ReadFile(serverKeyFile)
			if nil == err {
				cfg.ServerKey = tmpBuf
			}
		}

	}
}

func ServerClientPemFile(clientPemFile string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.ClientPemFile = clientPemFile
		if jkos.IsFileExists(cfg.ClientPemFile) {
			tmpBuf, err := jkos.ReadFile(clientPemFile)
			if nil == err {
				cfg.ClientPem = tmpBuf
			}
		}

	}
}

func ServerActionMiddlewares(actionMiddlewares ...jkendpoint.ActionMiddleware) ClientOption {
	return func(cfg *ClientConfig) {
		cfg.tmpActionMiddlewares = append(cfg.tmpActionMiddlewares, actionMiddlewares...)
	}
}

func ServerRegOption(regOps ...jkregistry.RegOption) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.RegOps = append(cfg.RegOps, regOps...)
	}
}

func ServerRateLimit(rateLimit int) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.RateLimit = rate.Limit(rateLimit)
	}
}

func ServerPrometheusNameSpace(prometheusNameSpace string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.PrometheusNameSpace = prometheusNameSpace
	}
}

func ServerListenerFatory(fatory CreateListenerFunc) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.ListenerFatory = fatory
	}
}

func ServerRun(serverRun ServerRunFunc) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.ServerRun = serverRun
	}
}

func ServerRpcPath(rpcPath string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.RpcPath = rpcPath
	}
}

func ServerRpcDebugPath(rpcDebugPath string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.RpcDebugPath = rpcDebugPath
	}
}

func ServerRpcName(rpcName string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.RpcName = rpcName
	}
}

func ServerCodec(codec string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.Codec = codec
	}
}

func ServerConfigFile(cfgPath string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.ConfigPath = cfgPath
		config := serverConfig{Cfg: cfg}
		utils.ReadConfigFile(cfg.ConfigPath, &config)
	}
}
