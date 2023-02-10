# 配置选项示例

配置选项同时支持 运行时配置，json文件配置，toml文件配置，环境变量配置，如果存在多个配置都配置同一个选项，则配置生效优先级为：运行时配置 > json文件配置 > toml文件配置 > 环境变量配置 > 默认配置

默认会扫描读取 conf/[server-name].json，conf/[server-name].json, 可以在运行时指定配置文件(ConfigPath)

## 配置文件

### json文件

```json
{
    "Client": {
        "RateLimit": 2,
        "ConsulTags": ["123"],
        "Strategy": "random",
        "Retry": 2,
        "TimeOut": 30
    },
    "Registry": {
        "ConsulAddr": "192.168.213.184:8500"
    },
    "GRPC": {
        "WriteBufferSize": 102400,
        "ReadBufferSize": 102400,
        "MaxMsgSize": 10240000,
        "TimeOut": 300,
        "EnableCompressor": true,
        "#CompressorLevel_remark": "NoCompression=0, BestSpeed=1, BestCompression=9, DefaultCompression=-1",
        "CompressorLevel": 9
    },
    "Keepalive": {
        "PermitWithoutStream": false,
        "Time": 1800,
        "TimeOut": 20
    }
}
```



### toml文件

```toml
[Client]
RateLimit = 20
ConsulTags = ["123"]
Strategy = "random"
TimeOut = 30
Retry = 3

[Registry]
ConsulAddr = "192.168.213.184:8500"

[GRPC]
WriteBufferSize = 102400
ReadBufferSize = 102400
MaxMsgSize = 10240000
Timeout = 300
EnableCompressor = true
##NoCompression=0, BestSpeed=1, BestCompression=9, DefaultCompression=-1
CompressorLevel = 9

[Keepalive]
PermitWithoutStream = false
Time = 1800
Timeout = 20
```



## 运行时配置

```go
import (
	"compress/gzip"
	"time"

	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jktrans "github.com/jkprj/jkfr/gokit/transport"
	jkgrpc "github.com/jkprj/jkfr/gokit/transport/grpc"
	jklog "github.com/jkprj/jkfr/log"
	pb "github.com/jkprj/jkfr/protobuf/demo"
	hellogrpc "github.com/jkprj/jkfr/protobuf/demo/hello-service/svc/client/grpc"

	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

func clientFatory(conn *grpc.ClientConn) (server interface{}, err error) {
	pbsvr, err := hellogrpc.New(conn)
	return pbsvr, err
}

func runClientWithOption() {

	compress, _ := grpc.NewGZIPCompressorWithLevel(gzip.BestCompression)

	client, err := jkgrpc.NewClient("test",
		clientFatory,
		jkgrpc.ClientStrategy(jktrans.STRATEGY_RANDOM),
		jkgrpc.ClientLimit(2),
		jkgrpc.ClientConsulTags("123"),
		jkgrpc.ClientRetry(5),
		jkgrpc.ClientTimeOut(10),
		jkgrpc.ClientRegOption(
			jkregistry.WithConsulAddr("192.168.213.184:8500"),
			jkregistry.WithWaitTime(5),
			jkregistry.WithPassingOnly(true),
			jkregistry.WithQueryOptions(
				&api.QueryOptions{
					UseCache: true,
					MaxAge:   time.Hour,
				},
			),
		),
		jkgrpc.ClientGRPCDialOps(
			grpc.WithWriteBufferSize(64*1024),
			grpc.WithReadBufferSize(64*1024),
			grpc.WithMaxMsgSize(640*1024),
			grpc.WithCompressor(compress),
			grpc.WithDecompressor(grpc.NewGZIPDecompressor()),
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                time.Hour,
				Timeout:             time.Minute,
				PermitWithoutStream: false,
			}),
		),
	)
	if nil != err {
		jklog.Errorw("NewClient fail", "err", err)
		return
	}

	resp, err := client.Call("GetPersons", &pb.PersonRequest{})
	jklog.Infow("call complete", "respone:", resp, "err", err)
}

```

# 配置参数说明

## Client

### GRPCDialOps

**描述：**

**环境变量：**

**配置选项：**ClientGRPCDialOps(grpcDialOps ...grpc.DialOption) ClientOption

### RegOps

**描述：**连接到及从consul获取服务的配置参数，参考register的配置参数说明

**环境变量：**

**配置选项：**ClientRegOption(regOps ...jkregistry.RegOption) ClientOption

### AsyncCallChan

**描述：**使用异步方式发送请求，当返回时通知的chan

**环境变量：**

**配置选项：**ClientAsyncCallChan(asyncCallChan chan *UCall) ClientOption

### ConfigPath

**描述：**配置文件路径，该配置不能通过配置文件配置，只能通过环境变量，配置选项配置，如果未设置默认会读取当前程序所在目录的 conf/[server-name].json，conf/[server-name].json

**环境变量：**C_CONFIG_PATH

**配置选项： ClientConfigFile(cfgPath string) ClientOption**

### ConsulTags

**描述：**从consul获取服务时的tags，也可以在RegOps中设置

**环境变量：**C_CONSUL_TAGS

**配置选项：**ClientConsulTags(tags ...string) ClientOption 

### Strategy

**描述：**负载均衡策略，目前提供3种策略：round（轮询），random（随机），least（最小请求数优先），默认least

**环境变量：**C_STRATEGY

**配置选项：**ClientStrategy(strategy string) ClientOption

### Retry

**描述：**请求失败重试次数，默认3次

**环境变量：**C_RETRY

**配置选项：**ClientRetry(retry int) ClientOption

### RetryIntervalMS

**描述：**请求失败重试间隔时间，单位毫秒，默认1000毫秒

**环境变量：**C_RETRY_INTERVAL_MS

**配置选项：**ClientRetryIntervalMS(retryIntervalMS int) ClientOption

### RateLimit

**描述：**限流器，每秒最大发送请求数，默认为0不限制

**环境变量：**C_RATE_LIMIT

**配置选项：**ClientRateLimit(rateLimit int) ClientOption

### TimeOut

**描述：**

**环境变量：**C_TIME_OUT

**配置选项：**ClientTimeOut(timeOut int) ClientOption

### PoolCap

**描述：**单服务连接池初始及最小连接数，默认2

**环境变量：**C_POOL_CAP

**配置选项：**ClientPoolCap(poolCap int) ClientOption

### MaxCap

**描述：**单服务连接池最大连接数，默认32

**环境变量：**C_MAX_CAP

**配置选项：**ClientMaxCap(maxCap int) ClientOption

### PassingOnly

**描述：**从consul获取服务时是否PassingOnly

**环境变量：**C_PASSING_ONLY

**配置选项：**ClientPassingOnly(passingOnly bool) ClientOption

### ActionMiddlewares

**描述：**设置 grpc 发送请求前后处理

**环境变量：**

**配置选项：**ClientActionMiddlewares(actionMiddlewares ...jkendpoint.ActionMiddleware) ClientOption

## GRPC

### WriteBufferSize

**描述：**

**环境变量：**C_WRITE_BUFFER_SIZE

**配置选项：**

### ReadBufferSize

**描述：**

**环境变量：**C_READ_BUFFER_SIZE

**配置选项：**

### MaxMsgSize

**描述：**

**环境变量：**C_MAX_MSG_SIZE

**配置选项：**

### MaxSendMsgSize

**描述：**

**环境变量：**C_MAX_SEND_MSG_SIZE

**配置选项：**

### MaxRecvMsgSize

**描述：**

**环境变量：**C_MAX_RECV_MSG_SIZE

**配置选项：**

### Timeout

**描述：**

**环境变量：**C_TIMEOUT

**配置选项：**

### EnableCompressor

**描述：**

**环境变量：**C_ENABLE_COMPRESSOR

**配置选项：**

### CompressorLevel

**描述：**

**环境变量：**C_COMPRESSOR_LEVEL

**配置选项：**

### InitialWindowSize

**描述：**

**环境变量：**C_INITIAL_WINDOW_SIZE

**配置选项：**

### InitialConnWindowSize

**描述：**

**环境变量：**C_INITIAL_CONN_WINDOW_SIZE

**配置选项：**



## Keepalive

### PermitWithoutStream

**描述：**

**环境变量：**C_KEEPALIVE_PERMIT_WITHOUT_STREAM

**配置选项：**

### Time

**描述：**

**环境变量：**C_KEEPALIVE_TIME

**配置选项：**

### Timeout

**描述：**

**环境变量：**C_KEEPALIVE_TIMEOUT

**配置选项：**

