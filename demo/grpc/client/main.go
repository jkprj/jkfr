package main

import (
	"compress/gzip"
	"context"
	"flag"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"sync/atomic"
	"time"

	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jkgrpc "github.com/jkprj/jkfr/gokit/transport/grpc"
	jkutils "github.com/jkprj/jkfr/gokit/utils"
	jklog "github.com/jkprj/jkfr/log"
	pb "github.com/jkprj/jkfr/protobuf/demo"
	hellogrpc "github.com/jkprj/jkfr/protobuf/demo/hello-service/svc/client/grpc"

	// kitgrpc "github.com/go-kit/kit/transport/grpc"
	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

var th_count int = 64
var ss_count int = 8
var server_type = ""
var server = "127.0.0.1:9090"

func clientFatory(conn *grpc.ClientConn) (server interface{}, err error) {
	pbsvr, err := hellogrpc.New(conn)
	return pbsvr, err
}

func init_param() {
	flag.StringVar(&server_type, "type", "", "grpc type")
	flag.StringVar(&server, "server", "127.0.0.1:9090", "server addr")
	flag.IntVar(&th_count, "th", 64, "thread count")
	flag.IntVar(&ss_count, "ss", 8, "session count")
	flag.Parse()

	jklog.Infow("param", "type", server_type, "server", server, "thread_count", th_count, "session_count", ss_count)
}

func run_pprof() {

	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)

	go func() {
		err := http.ListenAndServe(":8080", nil)
		if nil != err {
			jklog.Panicw("run pprof server fail", "err", err)
		}
	}()
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	// jklog.InitLogger(&jklog.ZapOptions{LogLevel: "debug"})

	run_pprof()

	init_param()

	// runDefaultClient()
	// runDefaultWithClientHandle()
	// runClientWithOption()
	// runClientAsync()
	// runClientWithDefaultConfigureFile()
	// runClientWithConfigureFileOption()

	if "" == server_type {
		jklog.Info("pressureTest")
		pressureTest()
	} else {
		jklog.Info("testgrpc")
		testgrpc()
	}
}

func runDefaultClient() {
	err := jkgrpc.RegistryNewClient("test", clientFatory)
	if nil != err {
		jklog.Errorw("RegistryNewClient fail", "err", err)
		return
	}

	resp, err := jkgrpc.Call("test", "GetPersons", &pb.PersonRequest{})
	jklog.Infow("call complete", "respone:", resp, "err", err)
}

func runDefaultWithClientHandle() {
	client, err := jkgrpc.NewClient("test", clientFatory)
	if nil != err {
		jklog.Errorw("NewClient fail", "err", err)
		return
	}
	for i := 0; i < 100; i++ {
		resp, err := client.Call("GetPersons", &pb.PersonRequest{})
		jklog.Infow("call complete", "respone:", resp, "err", err)
	}

}

func runClientWithOption() {

	compress, _ := grpc.NewGZIPCompressorWithLevel(gzip.BestCompression)

	client, err := jkgrpc.NewClient("test",
		clientFatory,
		jkgrpc.ClientStrategy(jkutils.STRATEGY_RANDOM),
		jkgrpc.ClientLimit(2),
		jkgrpc.ClientConsulTags("jinkun"),
		jkgrpc.ClientRetry(5),
		jkgrpc.ClientTimeOut(10),
		jkgrpc.ClientKeepAlive(true),
		jkgrpc.ClientRegOption(
			jkregistry.WithConsulAddr("127.0.0.1:8500"),
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

func runClientAsync() {
	client, err := jkgrpc.NewClient("test", clientFatory)
	if nil != err {
		jklog.Errorw("NewClient fail", "err", err)
		return
	}

	chCall := client.GetUCall()

	// or
	// chCall := make(chan *jkgrpc.UCall, 1000)
	// client, err := jkgrpc.NewClient("test", clientFatory, jkgrpc.ClientAsyncCallChan(chCall))
	// if nil != err {
	// 	jklog.Errorw( "NewClient fail", "err", err)
	// 	return
	// }

	for i := 0; i < 1000; i++ {
		client.GoCall("GetPersons", &pb.PersonRequest{})
	}

	for {
		select {
		case ucall := <-chCall:
			jklog.Infow("call complete", "request", ucall.Args, "respone:", ucall.Reply, "action", ucall.ServiceMethod, "err", ucall.Error)
		case <-time.After(time.Second):
			return
		}
	}

}

func runClientWithDefaultConfigureFile() {
	client, err := jkgrpc.NewClient("testT", clientFatory)
	// or
	// client, err := jkgrpc.NewClient("testJ", clientFatory)
	if nil != err {
		jklog.Errorw("NewClient fail", "err", err)
		return
	}

	for i := 0; i < 10; i++ {
		resp, err := client.Call("GetPersons", &pb.PersonRequest{})
		jklog.Infow("call complete", "respone:", resp, "err", err)
	}
}

func runClientWithConfigureFileOption() {
	client, err := jkgrpc.NewClient("test", clientFatory, jkgrpc.ClientConfigFile("conf/test.toml"))
	// or
	// client, err := jkgrpc.NewClient("test", clientFatory, jkgrpc.ClientConfigFile("conf/test.json"))
	if nil != err {
		jklog.Errorw("NewClient fail", "err", err)
		return
	}

	for i := 0; i < 10; i++ {
		resp, err := client.Call("GetPersons", &pb.PersonRequest{})
		jklog.Infow("call complete", "respone:", resp, "err", err)
	}
}

func pressureTest() {

	client, err := jkgrpc.NewClient("test", clientFatory)
	if nil != err {
		jklog.Errorw("NewClient fail", "err", err)
		return
	}

	var count int64
	req := &pb.HelloRequest{}

	for i := 0; i < th_count; i++ {
		go func() {
			for {
				// client.Call("GetPersons", &pb.PersonRequest{})
				// resp, err := client.Call("SayHello", req)
				client.Call("SayHello", req)
				// jklog.Infow("call complete", "respone:", resp, "err", err)
				atomic.AddInt64(&count, 1)
				// time.Sleep(time.Second)
			}
		}()
	}

	for {
		time.Sleep(time.Second)
		jklog.Infow("call statistics", "count:", atomic.SwapInt64(&count, 0))
	}
}

func testgrpc() {
	conns := make([]*grpc.ClientConn, 0, ss_count)
	clients := make([]pb.HelloServer, 0, ss_count)

	for i := 0; i < ss_count; i++ {
		conn, err := grpc.Dial(server, grpc.WithInsecure())
		if nil != err {
			jklog.Errorw("connect grpc sever fail", "err", err.Error())
			return
		}

		conns = append(conns, conn)

		client, _ := hellogrpc.New(conn)
		clients = append(clients, client)
	}

	defer func() {
		for _, conn := range conns {
			conn.Close()
		}
	}()

	var count int64

	for i := 0; i < th_count; i++ {
		go func(i int) {

			client := clients[i%ss_count]
			req := &pb.HelloRequest{}

			for {
				// ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				// reply, err := client.SayHello(context.Background(), req)
				client.SayHello(context.Background(), req)
				// jklog.Debugw("call complete", "respone:", reply, "err", err)
				// cancel()

				atomic.AddInt64(&count, 1)
			}
		}(i)
	}

	for {
		time.Sleep(time.Second)
		jklog.Infow("call statistics", "count:", atomic.SwapInt64(&count, 0))
	}
}
