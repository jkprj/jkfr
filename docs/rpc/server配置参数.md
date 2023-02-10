# 配置选项示例

配置选项同时支持 运行时配置，json文件配置，toml文件配置，环境变量配置，如果存在多个配置都配置同一个选项，则配置生效优先级为：运行时配置 > json文件配置 > toml文件配置 > 环境变量配置 > 默认配置

默认会扫描读取 conf/[server-name].json，conf/[server-name].json, 可以在运行时指定配置文件(ConfigPath)

## 配置文件

### json文件

```json
{
	"Server":{
		"ServerAddr":"192.168.213.184:6666",
        "BindAddr":"0.0.0.0:6666",
		"RateLimit": 10,
		"RpcName": "HelloWord",
		"Codec": "json"
	},
	"Registry":{
		"ConsulAddr": "192.168.213.184:8500",
		"ConsulTags": ["rpc", "123", "jinkun", "qjk"]
	}
}
```



### toml文件

```toml
[Server]
ServerAddr = "127.0.0.1:6666"
BindAddr = "0.0.0.0:6666"
RateLimit = 10
RpcName = "HelloWord"
Codec = "json"

[Registry]
ConsulAddr = "192.168.213.184:8500"
ConsulTags = ["rpc", "123", "jinkun", "qjk"]
```



## 运行时配置

```go
import (
	"github.com/jkprj/jkfr/demo/rpc/server/hello/endpoints"
	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jkrpc "github.com/jkprj/jkfr/gokit/transport/rpc"
)

func runServerWithOption() {
	jkrpc.RunServer("test_rpc",
		endpoints.NewService(),
		jkrpc.ServerAddr("192.168.213.184:6666"),
		jkrpc.ServerRateLimit(2),
		jkrpc.ServerRpcName("HelloWord"),
		jkrpc.ServerRegOption(
			jkregistry.WithConsulAddr("192.168.213.184:8500"),
			jkregistry.WithTags("rpc", "test", "123", "jinkun"),
		),
	)
}
```

# 配置参数说明

## ServerAddr

**描述：**注册到consul时的服务地址

**环境变量：**S_SERVER_ADDR

**配置选项：**ServerAddr(serverAddr string) ServerOption

## BindAddr

**描述：**服务绑定地址，如果ServerAddr，BindAddr其中之一为空，则这两个配置则一样

**环境变量：**S_BIND_ADDR

**配置选项：**BindAddr(bindAddr string) ServerOption

## ConfigPath

**描述：**配置文件路径，该配置不能通过配置文件配置，只能通过环境变量，配置选项配置，如果未设置默认会读取当前程序所在目录的 conf/[server-name].json，conf/[server-name].json

**环境变量：**S_CONFIG_PATH

**配置选项：**ServerConfigFile(cfgPath string) ServerOption

## RegOps

**描述：**注册到consul的配置参数，参考register的配置参数说明

**环境变量：**

**配置选项：**ServerRegOption(regOps ...jkregistry.RegOption) ServerOption

## ListenerFatory

**描述：**设置服务的创建监听方式，该选项只能在运行时配置，目前提供有两种创建方式：TCPListenerFatory，TLSListenerFatory，默认TCPListenerFatory

当然也可以自己定义，只要传参符合下面的函数定义即可

```go
type CreateListenerFunc func(cfg *ServerConfig) (net.Listener, error)
```

**环境变量：**

**配置选项：**ServerListenerFatory(fatory CreateListenerFunc) ServerOption

## ServerRun

**描述：**指定服务运行时响应处理逻辑，该选项只能在运行时配置，目前提供有两种运行方式：RunServerWithTcp，RunServerWithHttp，默认RunServerWithTcp

当然也可以自己定义，只要传参符合下面的函数定义即可

```go
type ServerRunFunc func(listener net.Listener, server *Server, cfg *ServerConfig) error
```

**环境变量：**

**配置选项：**ServerRun(serverRun ServerRunFunc) ServerOption

## ServerPemFile

**描述：**TLS模式时的ServerPemFile

**环境变量：**S_PEM_FILE

**配置选项：**ServerPemFile(serverPemFile string) ServerOption

## ServerKeyFile

**描述：**TLS模式时的ServerKeyFile

**环境变量：**S_KEY_FILE

**配置选项：**ServerKeyFile(serverKeyFile string) ServerOption

## ClientPemFile

**描述：**TLS模式时的ClientPemFile

**环境变量：**S_CLIENT_KEY_FILE

**配置选项：**ServerClientPemFile(clientPemFile string) ServerOption

## ServerPem

**描述：**TLS模式时的ServerPem值，只能通过运行时配置选项设置

**环境变量：**

**配置选项：**ServerPem(serverPem []byte) ServerOption

## ServerKey

**描述：**TLS模式时的ServerKey值，只能通过运行时配置选项设置

**环境变量：**

**配置选项：**ServerKey(serverKey []byte) ServerOption

## ClientPem

**描述：**TLS模式时的ClientPem值，只能通过运行时配置选项设置

**环境变量：**

**配置选项：**ServerClientPem(clientPem []byte) ServerOption

## RateLimit

**描述：**限流器，设置每秒最大请求数，默认为0不限制

**环境变量：**S_RATE_LIMIT

**配置选项：**ServerRateLimit(rateLimit int) ServerOption

## RpcPath

**描述：**使用http模式时的 rpc path

**环境变量：**S_RPC_PATH

**配置选项：**ServerRpcPath(rpcPath string) ServerOption

## RpcDebugPath

**描述：**使用http模式时的 rpc debug path

**环境变量：**S_RPC_DEBUG_PATH

**配置选项：**ServerRpcDebugPath(rpcDebugPath string) ServerOption

## RpcName

**描述：**rpc 的 RegisterName 时使用，指定 provided name (RegisterName is like Register but uses the provided name for the type, instead of the receiver's concrete type.), 具体参考rpc Server的RegisterName函数

**环境变量：**S_RPC_NAME

**配置选项：**ServerRpcName(rpcName string) ServerOption

## Codec

**描述：**设置与客户端通讯响应的数据编码协议，目前有两种编译选项：gob，json，默认 gob

**环境变量：**S_CODEC

**配置选项：**ServerCodec(codec string) ServerOption

## ActionMiddlewares

**描述：**设置 rpc 服务函数响应前后处理方式

**环境变量：**

**配置选项：**ServerActionMiddlewares(actionMiddlewares ...jkendpoint.ActionMiddleware) ClientOption



