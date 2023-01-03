package main

import (
	"compress/gzip"
	"context"
	"fmt"
	"runtime"
	"time"

	jkhandlers "jkfr/demo/grpc/server/handlers"
	jkregistry "jkfr/gokit/registry"
	jkgrpc "jkfr/gokit/transport/grpc"
	jklog "jkfr/log"
	pb "jkfr/protobuf/demo"
	"jkfr/protobuf/demo/hello-service/handlers"
	"jkfr/protobuf/demo/hello-service/svc"
	"jkfr/protobuf/demo/hello-service/svc/server"

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

	endpoints := serverEndpoints.(*svc.Endpoints)
	pb.RegisterHelloServer(grpcServer, endpoints)

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

	handlers.RegisterServer(jkhandlers.NewService())
	endpoints := server.NewEndpoints()
	fmt.Printf("%x", &endpoints)

	// runDefaultServer(&endpoints)
	// runServerWithOption(&endpoints)
	runServerWithDefaultConfigFile(&endpoints)
	// runServerWithConfigFileOption(&endpoints)

	// runMutiServer()
}

func runDefaultServer(endpoints *svc.Endpoints) {
	jkgrpc.RunServerWithServerAddr("test", "192.168.213.184:9090", endpoints, RegisterServer)
	// or
	// jkgrpc.RunServer("test", endpoints, RegisterServer, jkgrpc.ServerAddr("192.168.213.184:9090"))
}

func runServerWithOption(endpoints *svc.Endpoints) {

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

func runServerWithDefaultConfigFile(endpoints *svc.Endpoints) {
	// jkgrpc.RunServer("test", endpoints, RegisterServer)
	// or
	jkgrpc.RunServer("test", endpoints, RegisterServer)
}

func runServerWithConfigFileOption(endpoints *svc.Endpoints) {
	jkgrpc.RunServer("test", endpoints, RegisterServer, jkgrpc.ServerConfigFile("conf/test.toml"))
	// or
	// jkgrpc.RunServer("test", endpoints, RegisterServer, jkgrpc.ServerConfigFile("conf/test.json"))
}

func runMutiServer() {
	handlers.RegisterServer(jkhandlers.NewService())
	endpoints1 := server.NewEndpoints()
	go jkgrpc.RunServerWithServerAddr("test", "192.168.213.184:9090", &endpoints1, RegisterServer)

	handlers.RegisterServer(jkhandlers.NewService())
	endpoints2 := server.NewEndpoints()
	jkgrpc.RunServerWithServerAddr("test", "127.0.0.1:9090", &endpoints2, RegisterServer)
}
