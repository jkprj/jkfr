package grpc

import (
	"compress/gzip"
	"time"

	"google.golang.org/grpc/keepalive"

	"google.golang.org/grpc"

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
	RateLimit           rate.Limit `json:"RateLimit" toml:"RateLimit"`
	PrometheusNameSpace string     `json:"PrometheusNameSpace" toml:"PrometheusNameSpace"`

	RegOps            []jkregistry.RegOption        `json:"-" toml:"-"`
	GRPCSvrOps        []grpc.ServerOption           `json:"-" toml:"-"`
	ActionMiddlewares []jkendpoint.ActionMiddleware `json:"-" toml:"-"`
	ConfigPath        string

	tmpActionMiddlewares []jkendpoint.ActionMiddleware
}

type ServerGRPCConfig struct {
	WriteBufferSize   int  `json:"WriteBufferSize" toml:"WriteBufferSize"`
	ReadBufferSize    int  `json:"ReadBufferSize" toml:"ReadBufferSize"`
	MaxMsgSize        int  `json:"MaxMsgSize" toml:"MaxMsgSize"`
	MaxRecvMsgSize    int  `json:"MaxRecvMsgSize" toml:"MaxRecvMsgSize"`
	MaxSendMsgSize    int  `json:"MaxSendMsgSize" toml:"MaxSendMsgSize"`
	ConnectionTimeout int  `json:"ConnectionTimeout" toml:"ConnectionTimeout"`
	EnableCompressor  bool `json:"EnableCompressor" toml:"EnableCompressor"`
	CompressorLevel   int  `json:"CompressorLevel" toml:"CompressorLevel"`

	MaxConcurrentStreams  uint32 `json:"MaxConcurrentStreams" toml:"MaxConcurrentStreams"`
	InitialWindowSize     int32  `json:"InitialWindowSize" toml:"InitialWindowSize"`
	InitialConnWindowSize int32  `json:"InitialConnWindowSize" toml:"InitialConnWindowSize"`
	MaxHeaderListSize     uint32 `json:"MaxHeaderListSize" toml:"MaxHeaderListSize"`
}

type ServerGRPCKeepaliveParam struct {
	MaxConnectionIdle     int `json:"MaxConnectionIdle" toml:"MaxConnectionIdle"`
	MaxConnectionAge      int `json:"MaxConnectionAge" toml:"MaxConnectionAge"`
	MaxConnectionAgeGrace int `json:"MaxConnectionAgeGrace" toml:"MaxConnectionAgeGrace"`
	Time                  int `json:"Time" toml:"Time"`
	Timeout               int `json:"Timeout" toml:"Timeout"`
}

type serverConfig struct {
	Cfg     *ServerConfig            `json:"Server" toml:"Server"`
	GRPCCfg ServerGRPCConfig         `json:"GRPC" toml:"GRPC"`
	KPCfg   ServerGRPCKeepaliveParam `json:"Keepalive" toml:"Keepalive"`
}

func defaultServerConfig(serverName string) *ServerConfig {
	cfg := new(ServerConfig)
	cfg.RegOps = []jkregistry.RegOption{}
	cfg.GRPCSvrOps = []grpc.ServerOption{}

	cfg.ServerAddr = jkos.GetEnvString("S_SERVER_ADDR", "")
	cfg.BindAddr = jkos.GetEnvString("S_BIND_ADDR", "")
	cfg.PrometheusNameSpace = jkos.GetEnvString("S_PROMETHEUS_NAME_SPACE", serverName)
	cfg.RateLimit = rate.Limit(jkos.GetEnvInt("S_RATE_LIMIT", 0))

	tmpCfg := new(serverConfig)
	tmpCfg.GRPCCfg.WriteBufferSize = jkos.GetEnvInt("S_WRITE_BUFFER_SIZE", 0)
	tmpCfg.GRPCCfg.ReadBufferSize = jkos.GetEnvInt("S_READ_BUFFER_SIZE", 0)
	tmpCfg.GRPCCfg.MaxMsgSize = jkos.GetEnvInt("S_MAX_MSG_SIZE", 0)
	tmpCfg.GRPCCfg.MaxRecvMsgSize = jkos.GetEnvInt("S_MAX_RECV_MSG_SIZE", 0)
	tmpCfg.GRPCCfg.MaxSendMsgSize = jkos.GetEnvInt("S_MAX_SEND_MSG_SIZE", 0)
	tmpCfg.GRPCCfg.ConnectionTimeout = jkos.GetEnvInt("S_CONNECTION_TIMEOUT", 0)
	tmpCfg.GRPCCfg.EnableCompressor = jkos.GetEnvBool("S_ENABLE_COMPRESSOR", false)
	tmpCfg.GRPCCfg.CompressorLevel = jkos.GetEnvInt("S_COMPRESSOR_LEVEL", gzip.BestCompression)
	tmpCfg.GRPCCfg.MaxConcurrentStreams = uint32(jkos.GetEnvInt("S_MAX_CONCURRENT_STREAMS", 0))
	tmpCfg.GRPCCfg.InitialWindowSize = int32(jkos.GetEnvInt("S_INITIAL_WINDOW_SIZE", 0))
	tmpCfg.GRPCCfg.InitialConnWindowSize = int32(jkos.GetEnvInt("S_INITIAL_CONN_WINDOW_SIZE", 0))
	tmpCfg.GRPCCfg.MaxHeaderListSize = uint32(jkos.GetEnvInt("S_MAX_HEADER_LIST_SIZE", 0))

	tmpCfg.KPCfg.MaxConnectionIdle = jkos.GetEnvInt("S_KEEPALIVE_MAX_CONNECTION_IDLE", 0)
	tmpCfg.KPCfg.MaxConnectionAge = jkos.GetEnvInt("S_KEEPALIVE_MAX_CONNECTION_AGE", 0)
	tmpCfg.KPCfg.MaxConnectionAgeGrace = jkos.GetEnvInt("S_KEEPALIVE_MAX_CONNECTION_AGE_GRACE", 0)
	tmpCfg.KPCfg.Time = jkos.GetEnvInt("S_KEEPALIVE_TIME", 0)
	tmpCfg.KPCfg.Timeout = jkos.GetEnvInt("S_KEEPALIVE_TIMEOUT", 0)

	appendGRPCServerConfig(cfg, tmpCfg)

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

func appendGRPCServerConfig(cfg *ServerConfig, tmpCfg *serverConfig) {

	if 0 != tmpCfg.GRPCCfg.WriteBufferSize {
		cfg.GRPCSvrOps = append(cfg.GRPCSvrOps, grpc.WriteBufferSize(tmpCfg.GRPCCfg.WriteBufferSize))
	}

	if 0 != tmpCfg.GRPCCfg.ReadBufferSize {
		cfg.GRPCSvrOps = append(cfg.GRPCSvrOps, grpc.ReadBufferSize(tmpCfg.GRPCCfg.ReadBufferSize))
	}

	if 0 != tmpCfg.GRPCCfg.MaxMsgSize {
		cfg.GRPCSvrOps = append(cfg.GRPCSvrOps, grpc.MaxMsgSize(tmpCfg.GRPCCfg.MaxMsgSize))
	}

	if 0 != tmpCfg.GRPCCfg.WriteBufferSize {
		cfg.GRPCSvrOps = append(cfg.GRPCSvrOps, grpc.WriteBufferSize(tmpCfg.GRPCCfg.WriteBufferSize))
	}

	if 0 != tmpCfg.GRPCCfg.MaxRecvMsgSize {
		cfg.GRPCSvrOps = append(cfg.GRPCSvrOps, grpc.MaxRecvMsgSize(tmpCfg.GRPCCfg.MaxRecvMsgSize))
	}

	if 0 != tmpCfg.GRPCCfg.ConnectionTimeout {
		timeOut := time.Duration(tmpCfg.GRPCCfg.ConnectionTimeout) * time.Second
		cfg.GRPCSvrOps = append(cfg.GRPCSvrOps, grpc.ConnectionTimeout(timeOut))
	}

	if 0 != tmpCfg.GRPCCfg.MaxConcurrentStreams {
		cfg.GRPCSvrOps = append(cfg.GRPCSvrOps, grpc.MaxConcurrentStreams(tmpCfg.GRPCCfg.MaxConcurrentStreams))
	}

	if 0 != tmpCfg.GRPCCfg.InitialWindowSize {
		cfg.GRPCSvrOps = append(cfg.GRPCSvrOps, grpc.InitialWindowSize(tmpCfg.GRPCCfg.InitialWindowSize))
	}

	if 0 != tmpCfg.GRPCCfg.InitialConnWindowSize {
		cfg.GRPCSvrOps = append(cfg.GRPCSvrOps, grpc.InitialConnWindowSize(tmpCfg.GRPCCfg.InitialConnWindowSize))
	}

	if 0 != tmpCfg.GRPCCfg.MaxHeaderListSize {
		cfg.GRPCSvrOps = append(cfg.GRPCSvrOps, grpc.MaxHeaderListSize(tmpCfg.GRPCCfg.MaxHeaderListSize))
	}

	if true == tmpCfg.GRPCCfg.EnableCompressor {
		if tmpCfg.GRPCCfg.CompressorLevel < gzip.DefaultCompression || tmpCfg.GRPCCfg.CompressorLevel > gzip.BestCompression {
			tmpCfg.GRPCCfg.CompressorLevel = gzip.BestCompression
		}

		compressor, _ := grpc.NewGZIPCompressorWithLevel(tmpCfg.GRPCCfg.CompressorLevel)
		cfg.GRPCSvrOps = append(cfg.GRPCSvrOps, grpc.RPCCompressor(compressor))
		cfg.GRPCSvrOps = append(cfg.GRPCSvrOps, grpc.RPCDecompressor(grpc.NewGZIPDecompressor()))
	}

	if 0 != tmpCfg.KPCfg.MaxConnectionIdle ||
		0 != tmpCfg.KPCfg.MaxConnectionAge ||
		0 != tmpCfg.KPCfg.MaxConnectionAgeGrace ||
		0 != tmpCfg.KPCfg.Time ||
		0 != tmpCfg.KPCfg.Timeout {

		keepaliveParm := keepalive.ServerParameters{}
		keepaliveParm.MaxConnectionIdle = time.Duration(tmpCfg.KPCfg.MaxConnectionIdle) * time.Second
		keepaliveParm.MaxConnectionAge = time.Duration(tmpCfg.KPCfg.MaxConnectionAge) * time.Second
		keepaliveParm.MaxConnectionAgeGrace = time.Duration(tmpCfg.KPCfg.MaxConnectionAgeGrace) * time.Second
		keepaliveParm.Time = time.Duration(tmpCfg.KPCfg.Time) * time.Second
		keepaliveParm.Timeout = time.Duration(tmpCfg.KPCfg.Timeout) * time.Second

		cfg.GRPCSvrOps = append(cfg.GRPCSvrOps, grpc.KeepaliveParams(keepaliveParm))
	}
}

func newServerConfig(serverName string, ops ...ServerOption) *ServerConfig {

	cfg := defaultServerConfig(serverName)
	loadDefaultServerConfig(serverName, cfg)
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

func ServerGRPCSvrOp(grpcSvrOps ...grpc.ServerOption) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.GRPCSvrOps = append(cfg.GRPCSvrOps, grpcSvrOps...)
	}
}

func ServerActionMiddlewares(actionMiddlewares ...jkendpoint.ActionMiddleware) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.tmpActionMiddlewares = append(cfg.tmpActionMiddlewares, actionMiddlewares...)
	}
}

func ServerConfigFile(cfgPath string) ServerOption {
	return func(cfg *ServerConfig) {
		cfg.ConfigPath = cfgPath
		config := serverConfig{Cfg: cfg}
		utils.ReadConfigFile(cfg.ConfigPath, &config)
		appendGRPCServerConfig(cfg, &config)
	}
}
