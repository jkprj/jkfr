package main

import (
	"compress/gzip"
	"context"
	"runtime"
	"time"

	jkhandlers "github.com/jkprj/jkfr/demo/grpc/server/handlers"
	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jkgrpc "github.com/jkprj/jkfr/gokit/transport/grpc"
	jklog "github.com/jkprj/jkfr/log"
	helloPB "github.com/jkprj/jkfr/protobuf/demo"
	helloSvc "github.com/jkprj/jkfr/protobuf/demo/hello-service/svc"
	helloServer "github.com/jkprj/jkfr/protobuf/demo/hello-service/svc/server"

	// kitgrpc "github.com/go-kit/kit/transport/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
)

func beforServerRequestFunc(ctx context.Context, md metadata.MD) context.Context {
	return ctx
}

func afterServerResponseFunc(ctx context.Context, header *metadata.MD, trailer *metadata.MD) context.Context {
	return ctx
}

func ServerFinalizerFunc(ctx context.Context, err error) {}

func RegisterServer(grpcServer *grpc.Server, serverEndpoints interface{}) {

	endpoints := serverEndpoints.(*helloSvc.Endpoints)
	helloPB.RegisterHelloServer(grpcServer, endpoints)

	// or 使用以下方式创建，可以设置些选项：请求处理前，处理完成返回前，返回后的回调函数，及设置日志句柄，错误回调

	// endpoints := serverEndpoints.(*svc.Endpoints)
	// grpcSvc := svc.MakeGRPCServer(
	// 	*endpoints,
	// 	kitgrpc.ServerBefore(beforServerRequestFunc),
	// 	kitgrpc.ServerAfter(afterServerResponseFunc),
	// 	kitgrpc.ServerFinalizer(ServerFinalizerFunc),
	// 	kitgrpc.ServerErrorLogger(jklog.ErrorwLogger),
	// )

	// pb.RegisterHelloServer(grpcServer, grpcSvc)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	jklog.InitLogger()

	endpoints := helloServer.NewEndpoints(jkhandlers.NewService())

	// runDefaultServer(&endpoints)
	// runServerWithOption(&endpoints)
	runServerWithDefaultConfigFile(&endpoints)
	// runServerWithConfigFileOption(&endpoints)

	// runMutiServer()
}

func runDefaultServer(endpoints *helloSvc.Endpoints) {
	jkgrpc.RunServerWithServerAddr("test", "192.168.213.184:9090", endpoints, RegisterServer)
	// or
	// jkgrpc.RunServer("test", endpoints, RegisterServer, jkgrpc.ServerAddr("192.168.213.184:9090"))
}

func runServerWithOption(endpoints *helloSvc.Endpoints) {

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

func runServerWithDefaultConfigFile(endpoints *helloSvc.Endpoints) {
	// jkgrpc.RunServer("test", endpoints, RegisterServer)
	// or
	jkgrpc.RunServer("test", endpoints, RegisterServer)
}

func runServerWithConfigFileOption(endpoints *helloSvc.Endpoints) {
	jkgrpc.RunServer("test", endpoints, RegisterServer, jkgrpc.ServerConfigFile("conf/test.toml"))
	// or
	// jkgrpc.RunServer("test", endpoints, RegisterServer, jkgrpc.ServerConfigFile("conf/test.json"))
}

func runMutiServer() {
	endpoints1 := helloServer.NewEndpoints(jkhandlers.NewService())
	go jkgrpc.RunServerWithServerAddr("test", "192.168.213.184:9090", &endpoints1, RegisterServer)

	endpoints2 := helloServer.NewEndpoints(jkhandlers.NewService())
	jkgrpc.RunServerWithServerAddr("test", "127.0.0.1:9090", &endpoints2, RegisterServer)
}
