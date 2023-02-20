package main

import (
	"flag"
	"net"
	"net/rpc"
	"runtime"
	"sync/atomic"
	"time"

	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jktrans "github.com/jkprj/jkfr/gokit/transport"
	jkrpc "github.com/jkprj/jkfr/gokit/transport/rpc"
	jklog "github.com/jkprj/jkfr/log"
	// "github.com/hashicorp/consul/api"
)

type URequest struct {
	Name string `json:"Name,omitempty"`
}

type URespone struct {
	Msg string `json:"Msg,omitempty"`
}

var th_count int = 64
var ss_count int = 8
var server_type = ""
var server = "127.0.0.1:6666"

func init_param() {
	flag.StringVar(&server_type, "type", "", "rpc type")
	flag.StringVar(&server, "server", "", "server addr")
	flag.IntVar(&th_count, "th", 10000, "thread count")
	flag.IntVar(&ss_count, "ss", 8, "session count")
	flag.Parse()

	jklog.Infow("param", "type", server_type, "server", server, "thread_count", th_count, "session_count", ss_count)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	jklog.InitLogger()
	init_param()

	// callWithDefault()
	// callWithTLSTCP()
	// callWithHttp()
	// callWithTLSHttp()
	// callWithOption()
	// callWithDefaultConfigureFile()
	// callWithConfigureFileOption()

	if "" == server_type {
		jklog.Info("pressureTest")
		pressureTest()
	} else {
		jklog.Info("testRPC")
		testRPC()
	}
}

// default : TCP no TLS
func callWithDefault() {
	resp := new(URespone)
	for i := 0; i < 10; i++ {
		err := jkrpc.Call("test", "Hello.HowAreYou", &URequest{}, resp)
		jklog.Infow("call complete", "respone:", resp, "err", err)
		time.Sleep(time.Second)
	}
}

// TCP with TLS
func callWithTLSTCP() {
	err := jkrpc.RegistryNewClient("test",
		jkrpc.ClientPemFile("pem-file"),
		jkrpc.ClientKeyFile("key-file"),
		jkrpc.ClientCreateFatory(jkrpc.TLSClientFatory),
	)
	if nil != err {
		jklog.Infow("RegistryNewClient fail", "err", err)
	}

	resp := new(URespone)
	for i := 0; i < 10; i++ {
		err := jkrpc.Call("test", "Hello.HowAreYou", &URequest{}, resp)
		jklog.Infow("call complete", "respone:", resp, "err", err)
	}
}

// http, no TLS
func callWithHttp() {
	client, err := jkrpc.NewClient("test", jkrpc.ClientCreateFatory(jkrpc.HttpClientFatory))
	if nil != err {
		jklog.Infow("jkrpc.NewClient fail", "err", err)
	}

	resp := new(URespone)
	for i := 0; i < 10; i++ {
		err := client.Call("Hello.HowAreYou", &URequest{}, resp)
		jklog.Infow("call complete", "respone:", resp, "err", err)
	}
}

// http with TLS
func callWithTLSHttp() {
	client, err := jkrpc.NewClient("test",
		jkrpc.ClientPemFile("pem-file"),
		jkrpc.ClientKeyFile("key-file"),
		jkrpc.ClientCreateFatory(jkrpc.TLSHttpClientFatory),
	)
	if nil != err {
		jklog.Infow("jkrpc.NewClient fail", "err", err)
	}

	resp := new(URespone)
	for i := 0; i < 10; i++ {
		err := client.Call("Hello.HowAreYou", &URequest{}, resp)
		jklog.Infow("call complete", "respone:", resp, "err", err)
	}
}

func callWithOption() {
	client, err := jkrpc.NewClient("test_rpc",
		jkrpc.ClientConsulTags("rpc", "jinkun"),
		jkrpc.ClientStrategy(jktrans.STRATEGY_RANDOM),
		jkrpc.ClientRateLimit(10),
		jkrpc.ClientTimeOut(5),
		jkrpc.ClientPassingOnly(true),
		jkrpc.ClientKeepAlive(true),
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
	for i := 0; i < 10; i++ {
		err := client.Call("Hello.HowAreYou", &URequest{}, resp)
		jklog.Infow("call complete", "respone:", resp, "err", err)
	}
}

func callWithDefaultConfigureFile() {
	resp := new(URespone)
	for i := 0; i < 10; i++ {
		err := jkrpc.Call("testT", "Hello.HowAreYou", &URequest{}, resp)
		// or
		// err := jkrpc.Call("testJ", "HelloWord.HowAreYou", &URequest{}, resp)

		jklog.Infow("call complete", "respone:", resp, "err", err)
	}
}

func callWithConfigureFileOption() {
	client, err := jkrpc.NewClient("test", jkrpc.ClientConfigFile("conf/test.toml"))
	// or
	// client, err := jkrpc.NewClient("test", jkrpc.ClientConfigFile("conf/test.json"))
	if nil != err {
		jklog.Infow("jkrpc.NewClient fail", "err", err)
	}

	resp := new(URespone)
	for i := 0; i < 10; i++ {
		err := client.Call("Hello.HowAreYou", &URequest{}, resp)
		jklog.Infow("call complete", "respone:", resp, "err", err)
	}
}

func pressureTest() {
	jkrpc.RegistryNewClient("test", jkrpc.ClientCreateFatory(jkrpc.TcpClientFatory))

	resp := new(URespone)
	var count uint64

	for i := 0; i < th_count; i++ {
		go func() {
			for {
				err := jkrpc.Call("test", "HelloWord.Hello", &URequest{}, resp)
				if nil != err {
					jklog.Errorw("call rpc fail", "err", err)
					break
				}
				atomic.AddUint64(&count, 1)
				// jklog.Infow("", "resp", resp)
				// time.Sleep(time.Second)
			}
		}()
	}

	for {
		time.Sleep(time.Second)
		jklog.Info("count:", atomic.SwapUint64(&count, 0))
	}
}

func testRPC() {

	conns := make([]net.Conn, 0, ss_count)
	clients := make([]*rpc.Client, 0, ss_count)

	for i := 0; i < ss_count; i++ {
		conn, err := net.DialTimeout("tcp", server, time.Second*10)
		if nil != err {
			jklog.Panicw("connect server fail", "err", err)
		}

		conns = append(conns, conn)

		client := rpc.NewClient(conn)
		clients = append(clients, client)
	}

	defer func() {
		for _, conn := range conns {
			conn.Close()
		}
		for _, client := range clients {
			client.Close()
		}
	}()

	resp := new(URespone)
	var count int64

	for i := 0; i < th_count; i++ {

		go func() {

			client := clients[i%ss_count]

			for {
				err := client.Call("Hello.Hello", &URequest{}, resp)
				if nil != err {
					jklog.Errorw("call rpc fail", "err", err)
					return
				}
				atomic.AddInt64(&count, 1)
			}
		}()
	}

	for {
		time.Sleep(time.Second)
		jklog.Info("count:", atomic.SwapInt64(&count, 0))
	}
}
