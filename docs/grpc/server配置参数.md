# 配置选项示例

配置选项同时支持 运行时配置，json文件配置，toml文件配置，环境变量配置，如果存在多个配置都配置同一个选项，则配置生效优先级为：运行时配置 > json文件配置 > toml文件配置 > 环境变量配置 > 默认配置

默认会扫描读取 conf/[server-name].json，conf/[server-name].json, 可以在运行时指定配置文件(ConfigPath)

## 配置文件

### json文件

```json
{
	"Server":{
		"ServerAddr":"127.0.0.1:9090",
		"BindAddr":":9090",
		"RateLimit": 10
	},
	"Registry":{
		"ConsulAddr":"192.168.213.184:8500"
	},
	"GRPC":{
		"WriteBufferSize": 102400,
		"ReadBufferSize": 102400,
		"MaxMsgSize": 10240000,
		"MaxRecvMsgSize": 10240000,
		"MaxSendMsgSize": 10240000,
		"ConnectionTimeout": 5,
		"EnableCompressor": true,
		"CompressorLevel_Remark": "NoCompression=0, BestSpeed=1, BestCompression=9, DefaultCompression=-1",
		"CompressorLevel": 9
	},
	"Keepalive":{
		"MaxConnectionIdle": 600,
		"MaxConnectionAge": 72000,
		"MaxConnectionAgeGrace": 72000,
		"Time": 1800,
		"Timeout": 20
	}
}
```



### toml文件

```toml
[Server]
BindAddr = ":9900"
ServerAddr = "192.168.213.184:9900"
RateLimit = 20

[Registry]
ConsulAddr = "192.168.213.184:8500"

[GRPC]
WriteBufferSize = 102400
ReadBufferSize = 102400
MaxMsgSize = 10240000
MaxRecvMsgSize = 10240000
MaxSendMsgSize = 10240000
ConnectionTimeout = 5
EnableCompressor = true
#NoCompression=0, BestSpeed=1, BestCompression=9, DefaultCompression=-1
CompressorLevel = 9

[Keepalive]
MaxConnectionIdle = 600
MaxConnectionAge = 72000
MaxConnectionAgeGrace = 72000
Time = 1800
Timeout = 20
```



## 运行时配置

```go
import (
	"compress/gzip"
	"time"

	jkhandlers "github.com/jkprj/jkfr/demo/grpc/server/handlers"
	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jkgrpc "github.com/jkprj/jkfr/gokit/transport/grpc"
	pb "github.com/jkprj/jkfr/protobuf/demo"
	"github.com/jkprj/jkfr/protobuf/demo/hello-service/handlers"
	"github.com/jkprj/jkfr/protobuf/demo/hello-service/svc"
	"github.com/jkprj/jkfr/protobuf/demo/hello-service/svc/server"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func RegisterServer(grpcServer *grpc.Server, serverEndpoints interface{}) {

	endpoints := serverEndpoints.(*svc.Endpoints)
	pb.RegisterHelloServer(grpcServer, endpoints)
}

func runServerWithOption() {

    handlers.RegisterServer(jkhandlers.NewService())
	endpoints := server.NewEndpoints()
    
	compressor, _ := grpc.NewGZIPCompressorWithLevel(gzip.BestCompression)

	jkgrpc.RunServer("test",
		endpoints,
		RegisterServer,
		jkgrpc.ServerAddr("127.0.0.1:8888"),
		jkgrpc.ServerAddr("192.168.213.184:9090"), // 相同设置后面的会覆盖前面的设置
		jkgrpc.ServerLimit(10),
		jkgrpc.ServerRegOption(
			jkregistry.WithServerAddr("127.0.0.1:8888"), // 不会起作用，最终会设置为jkgrpc.ServerAddr：192.168.213.184:8080
			jkregistry.WithConsulAddr("192.168.213.184:8500"),
		),
		jkgrpc.ServerGRPCSvrOp(
			grpc.WriteBufferSize(64*1024),
			grpc.ReadBufferSize(128*1024),
			grpc.RPCCompressor(compressor),
			grpc.RPCDecompressor(grpc.NewGZIPDecompressor()),
			grpc.ConnectionTimeout(5*time.Second),
			grpc.KeepaliveParams(
				keepalive.ServerParameters{
					MaxConnectionIdle:     10 * time.Minute,
					MaxConnectionAge:      24 * time.Hour,
					MaxConnectionAgeGrace: 24 * time.Hour,
					Time:                  2 * time.Hour,
					Timeout:               20 * time.Second,
				},
			),
		),
	)
}


```

# 配置参数说明

## Server

### ServerAddr

**描述：**注册到consul时的服务地址

**环境变量：**S_SERVER_ADDR

**配置选项：**ServerAddr(serverAddr string) ServerOption

### BindAddr

**描述：**

**环境变量：**S_BIND_ADDR

**配置选项：**BindAddr(bindAddr string) ServerOption

### RegOps

**描述：**

**环境变量：**

**配置选项：**ServerRegOption(regOps ...jkregistry.RegOption) ServerOption

### GRPCSvrOps

**描述：**

**环境变量：**

**配置选项：**ServerGRPCSvrOp(grpcSvrOps ...grpc.ServerOption) ServerOption

### ConfigPath

**描述：**

**环境变量：**S_CONFIG_PATH

**配置选项：**

### RateLimit

**描述：**

**环境变量：**S_RATE_LIMIT

**配置选项：**ServerLimit(limit rate.Limit) ServerOption

### PrometheusNameSpace

**描述：**

**环境变量：**S_PROMETHEUS_NAME_SPACE

**配置选项：**ServerPrometheusNameSpace(prometheusnamespace string) ServerOption

### ActionMiddlewares

**描述：**

**环境变量：**

**配置选项：**ServerActionMiddlewares(actionMiddlewares ...jkendpoint.ActionMiddleware) ServerOption

## GRPC

### WriteBufferSize

**描述：**

**环境变量：**S_WRITE_BUFFER_SIZE

**配置选项：**

### ReadBufferSize

**描述：**

**环境变量：**S_READ_BUFFER_SIZE

**配置选项：**

### MaxMsgSize

**描述：**

**环境变量：**S_MAX_MSG_SIZE

**配置选项：**

### MaxRecvMsgSize

**描述：**

**环境变量：**S_MAX_RECV_MSG_SIZE

**配置选项：**

### MaxSendMsgSize

**描述：**

**环境变量：**S_MAX_SEND_MSG_SIZE

**配置选项：**

### ConnectionTimeout

**描述：**

**环境变量：**S_CONNECTION_TIMEOUT

**配置选项：**

### EnableCompressor

**描述：**

**环境变量：**S_ENABLE_COMPRESSOR

**配置选项：**

### CompressorLevel

**描述：**

**环境变量：**S_COMPRESSOR_LEVEL

**配置选项：**

### MaxConcurrentStreams

**描述：**

**环境变量：**S_MAX_CONCURRENT_STREAMS

**配置选项：**

### InitialWindowSize

**描述：**

**环境变量：**S_INITIAL_WINDOW_SIZE

**配置选项：**

### InitialConnWindowSize

**描述：**

**环境变量：**S_INITIAL_CONN_WINDOW_SIZE

**配置选项：**

### MaxHeaderListSize

**描述：**

**环境变量：**S_MAX_HEADER_LIST_SIZE

**配置选项：**



## Keepalive

### MaxConnectionIdle

**描述：**

**环境变量：**S_KEEPALIVE_MAX_CONNECTION_IDLE

**配置选项：**

### MaxConnectionAge

**描述：**

**环境变量：**S_KEEPALIVE_MAX_CONNECTION_AGE

**配置选项：**

### MaxConnectionAgeGrace

**描述：**

**环境变量：**S_KEEPALIVE_MAX_CONNECTION_AGE_GRACE

**配置选项：**

### Time

**描述：**

**环境变量：**S_KEEPALIVE_TIME

**配置选项：**

### Timeout

**描述：**

**环境变量：**S_KEEPALIVE_TIMEOUT

**配置选项：**

