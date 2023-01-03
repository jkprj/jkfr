package main

import (
	"runtime"

	"jkfr/demo/rpc/server/hello/endpoints"
	jkregistry "jkfr/gokit/registry"
	jkrpc "jkfr/gokit/transport/rpc"
	jklog "jkfr/log"
	// "github.com/hashicorp/consul/api"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	jklog.InitLogger()

	// runDefaultServer()
	runTCPServer()
	// runHttpServer()
	// runTLSHttpServer()
	// runServerWithOption()
	// runServerWithDefaultConfigureFile()
	// runServerWithConfigFileOption()
	// runMultiServer()
}

// TCP no TLS
func runDefaultServer() {
	jkrpc.RunServer("test", endpoints.NewService())
	// or
	// jkrpc.RunServerWithServerAddr("test", "127.0.0.1:6666", endpoints.NewService())
	// or
	// jkrpc.RunServer("test", endpoints.NewService(), jkrpc.ServerAddr())
}

// TCP with TLS
func runTCPServer() {
	jkrpc.RunTLSServer("test",
		endpoints.NewService(),
		jkrpc.ServerKeyFile("key-file"),
		jkrpc.ServerPemFile("pem-file"),
		jkrpc.ServerClientPemFile("c-pem-file"),
	)
	// // or
	// jkrpc.RunServerWithServerAddr("test",
	// 	"127.0.0.1:6666",
	// 	endpoints.NewService(),
	// 	jkrpc.ServerKeyFile("key-file"),
	// 	jkrpc.ServerPemFile("pem-file"),
	// 	jkrpc.ServerClientPemFile("c-pem-file"),
	// 	jkrpc.ServerListenerFatory(jkrpc.TCPListenerFatory),
	// )
	// // or
	// jkrpc.RunServer("test",
	// 	endpoints.NewService(),
	// 	jkrpc.ServerAddr("127.0.0.1:6666"),
	// 	jkrpc.ServerKeyFile("key-file"),
	// 	jkrpc.ServerPemFile("pem-file"),
	// 	jkrpc.ServerClientPemFile("c-pem-file"),
	// 	jkrpc.ServerListenerFatory(jkrpc.TCPListenerFatory),
	// )
}

// HTTP, no TLS
func runHttpServer() {
	jkrpc.RunHttpServer("test", endpoints.NewService(), jkrpc.ServerAddr("127.0.0.1:6666"))
	// // or
	// jkrpc.RunServerWithServerAddr("test",
	// 	"127.0.0.1:6666",
	// 	endpoints.NewService(),
	// 	jkrpc.ServerListenerFatory(jkrpc.TCPListenerFatory),
	// 	jkrpc.ServerRun(jkrpc.RunServerWithHttp),
	// )
	// // or
	// jkrpc.RunServer("test",
	// 	endpoints.NewService(),
	// 	jkrpc.ServerAddr("127.0.0.1:6666"),
	// 	jkrpc.ServerListenerFatory(jkrpc.TCPListenerFatory),
	// 	jkrpc.ServerRun(jkrpc.RunServerWithHttp),
	// )
}

// HTTP with TLS
func runTLSHttpServer() {
	jkrpc.RunTLSHttpServer("test",
		endpoints.NewService(),
		jkrpc.ServerAddr("127.0.0.1:6666"),
		jkrpc.ServerKeyFile("key-file"),
		jkrpc.ServerPemFile("pem-file"),
		jkrpc.ServerClientPemFile("c-pem-file"),
	)
	// or
	// jkrpc.RunServerWithServerAddr("test",
	// 	"127.0.0.1:6666",
	// 	endpoints.NewService(),
	// 	jkrpc.ServerKeyFile("key-file"),
	// 	jkrpc.ServerPemFile("pem-file"),
	// 	jkrpc.ServerClientPemFile("c-pem-file"),
	// 	jkrpc.ServerRun(jkrpc.RunServerWithHttp),
	// )
	// or
	// jkrpc.RunServer("test",
	// 	endpoints.NewService(),
	// 	jkrpc.ServerAddr("127.0.0.1:6666"),
	// 	jkrpc.ServerKeyFile("key-file"),
	// 	jkrpc.ServerPemFile("pem-file"),
	// 	jkrpc.ServerClientPemFile("c-pem-file"),
	// 	jkrpc.ServerRun(jkrpc.RunServerWithHttp),
	// )
}

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

func runServerWithDefaultConfigureFile() {
	// jkrpc.RunServer("test", endpoints.NewService())
	// or
	jkrpc.RunServer("test", endpoints.NewService(), jkrpc.ServerListenerFatory(jkrpc.TCPListenerFatory))
}

func runServerWithConfigFileOption() {
	jkrpc.RunServer("test", endpoints.NewService(), jkrpc.ServerConfigFile("conf/test.toml"))
	// or
	// jkrpc.RunServer("test", endpoints.NewService(), jkrpc.ServerConfigFile("conf/test.json"))
}

func runMultiServer() {
	// go jkrpc.RunServerWithServerAddr("test", "127.0.0.1:6666", endpoints.NewService(), jkrpc.ServerListenerFatory(jkrpc.TCPListenerFatory))

	jkrpc.RunServerWithServerAddr("test", "192.168.213.184:6666", endpoints.NewService(), jkrpc.ServerListenerFatory(jkrpc.TCPListenerFatory))
}
