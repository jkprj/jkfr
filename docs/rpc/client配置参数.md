

# 配置选项示例

配置选项同时支持 运行时配置，json文件配置，toml文件配置，环境变量配置，如果存在多个配置都配置同一个选项，则配置生效优先级为：运行时配置 > json文件配置 > toml文件配置 > 环境变量配置 > 默认配置

默认会扫描读取 conf/[server-name].json，conf/[server-name].json, 可以在运行时指定配置文件(ConfigPath)

## 配置文件

### json文件

```json
{
    "Client": {
        "RateLimit": 20,
        "ConsulTags": ["123", "jinkun"],
        "Strategy": "random",
        "Retry": 2,
        "PoolCap": 1,
        "MaxCap": 16,
        "IdleTimeout": 600,
		"Codec": "json"
    },
    "Registry": {
        "ConsulAddr": "192.168.137.90:8500"
    }
}
```



### toml文件

```toml
[Client]
RateLimit = 20
ConsulTags = ["123", "jinkun"]
Strategy = "random"
Retry = 2
PoolCap = 5
MaxCap = 100
IdleTimeout = 600
Codec = "json"

[Registry]
ConsulAddr = "192.168.137.90:8500"
```



## 运行时配置

```go
import (
	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jktrans "github.com/jkprj/jkfr/gokit/transport"
	jkrpc "github.com/jkprj/jkfr/gokit/transport/rpc"
	jklog "github.com/jkprj/jkfr/log"
)

func callWithOption() {
	client, err := jkrpc.NewClient("test_rpc",
		jkrpc.ClientConsulTags("rpc", "jinkun"),
		jkrpc.ClientStrategy(jktrans.STRATEGY_RANDOM),
		jkrpc.ClientRateLimit(10),
		jkrpc.ClientPassingOnly(true),
		jkrpc.ClientPoolCap(2),
		jkrpc.ClientMaxCap(200),
		jkrpc.ClientIdleTimeout(600),
		jkrpc.ClientRegOption(
			jkregistry.WithConsulAddr("192.168.213.184:8500"),
		),
	)

	if nil != err {
		jklog.Infow("jkrpc.NewClient fail", "err", err)
	}

	resp := new(URespone)
	client.Call("Hello.HowAreYou", &URequest{}, resp)
	jklog.Infow("call complete", "respone:", resp, "err", err)
}

```

# 配置参数说明

## Fatory

**描述：**连接服务端创建方式，目前支持的连接创建方式：TcpClientFatory，HttpClientFatory，TLSClientFatory，TLSHttpClientFatory，默认 TcpClientFatory

当然也可以自己定义，只要传参符合下面的函数定义即可

```go
type ClientFatory func(cfg *ClientConfig) pool.ClientFatory
```

**环境变量：**

**配置选项：**ClientCreateFatory(fatory ClientFatory) ClientOption

## RegOps

**描述：**连接到及从consul获取服务的配置参数，参考register的配置参数说明

**环境变量：**

**配置选项：**ClientRegOption(regOps ...jkregistry.RegOption) ClientOption

## AsyncCallChan

**描述：**使用异步方式发送请求，当返回时通知的chan

**环境变量：**

**配置选项：**ClientAsyncCallChan(asyncCallChan chan *rpc.Call) ClientOption

## ClientPem

**描述：**TLS模式时的ClientPem值，只能通过运行时配置选项设置

**环境变量：**

**配置选项：**ClientPem(clientPem []byte) ClientOption

## ClientKey

**描述：**TLS模式时的ClientKey值，只能通过运行时配置选项设置

**环境变量：**

**配置选项：**ClientKey(clientKey []byte) ClientOption

## ConfigPath

**描述：**配置文件路径，该配置不能通过配置文件配置，只能通过环境变量，配置选项配置，如果未设置默认会读取当前程序所在目录的 conf/[server-name].json，conf/[server-name].json

**环境变量：**C_CONFIG_PATH

**配置选项：**ClientConfigFile(cfgPath string) ClientOption

## ConsulTags

**描述：**从consul获取服务时的tags，也可以在RegOps中设置

**环境变量：**C_CONSUL_TAGS

**配置选项：**ClientConsulTags(tags ...string) ClientOption

## PassingOnly

**描述：**从consul获取服务时是否PassingOnly

**环境变量：**C_PASSING_ONLY

**配置选项：**ClientPassingOnly(passingOnly bool) ClientOption

## Strategy

**描述：**负载均衡策略，目前提供3种策略：round（轮询），random（随机），least（最小请求数优先），默认least

**环境变量：**C_STRATEGY

**配置选项：**ClientStrategy(strategy string) ClientOption

## Retry

**描述：**请求失败重试次数，默认3次

**环境变量：**C_RETRY

**配置选项：**ClientRetry(retry int) ClientOption

## RetryIntervalMS

**描述：**请求失败重试间隔，单位毫秒，默认1000毫秒

**环境变量：**C_RETRY_INTERVAL_MS

**配置选项：**ClientRetryIntervalMS(retryIntervalMS int) ClientOption

## RateLimit

**描述：**限流器，每秒最大发送请求数，默认为0不限制

**环境变量：**C_RATE_LIMIT

**配置选项：**ClientRateLimit(rateLimit int) ClientOption

## PoolCap

**描述：**单服务连接池初始及最小连接数，默认2

**环境变量：**C_POOL_CAP

**配置选项：**ClientPoolCap(poolCap int) ClientOption

## MaxCap

**描述：**单服务连接池最大连接数，默认64

**环境变量：**C_MAX_CAP

**配置选项：**ClientMaxCap(maxCap int) ClientOption

## DialTimeout

**描述：**连接服务超时时间，单位秒，默认10秒

**环境变量：**C_DIAL_TIMEOUT

**配置选项：**ClientDialTimeout(dialTimeout int) ClientOption

## IdleTimeout

**描述：**连接池及连接空闲时间，单位秒，默认600秒

**环境变量：**C_IDLE_TIMEOUT

**配置选项：**ClientIdleTimeout(idleTimeout int) ClientOption

## ReadTimeout

**描述：**连接读超时时间，单位秒，默认60秒

**环境变量：**C_READ_TIMEOUT

**配置选项：**ClientReadTimeout(readTimeout int) ClientOption

## WriteTimeout

**描述：**连接写超时时间，单位秒，默认60秒

**环境变量：**C_WRITE_TIMEOUT

**配置选项：**ClientWriteTimeout(writeTimeout int) ClientOption

## ClientPemFile

**描述：**TLS模式时的ClientPemFile

**环境变量：**C_PEM_FILE

**配置选项：**ClientPemFile(clientPemFile string) ClientOption

## ClientKeyFile

**描述：**TLS模式时的ClientKeyFile

**环境变量：**C_KEY_FILE

**配置选项：**ClientKeyFile(clientKeyFile string) ClientOption

## Codec

**描述：**设置与服务通讯的数据编码协议，目前有两种编译选项：gob，json，默认 gob

**环境变量：**	C_CODEC

**配置选项：**ClientCodec(codec string) ClientOption

## ActionMiddlewares

**描述：**设置 rpc 发送请求前后处理

**环境变量：**

**配置选项：**ClientActionMiddlewares(actionMiddlewares ...jkendpoint.ActionMiddleware) ClientOption

