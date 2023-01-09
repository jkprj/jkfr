package main

import (
	"flag"
	"runtime"
	"sync/atomic"
	"time"

	jkpool "github.com/jkprj/jkfr/gokit/transport/pool"
	grpc_pools "github.com/jkprj/jkfr/gokit/transport/pool/grpc"
	"github.com/jkprj/jkfr/log"
	pb "github.com/jkprj/jkfr/protobuf/demo"
	hellogrpc "github.com/jkprj/jkfr/protobuf/demo/hello-service/svc/client/grpc"

	"google.golang.org/grpc"
)

var th_count int = 64
var ss_count int = 8
var server_type = ""
var server = "127.0.0.1:6666"

var total uint64 = 0

func statistics() {

	for {
		time.Sleep(time.Second)
		log.Infof("statistics: %d", atomic.SwapUint64(&total, 0))
	}
}

func init_param() {
	flag.StringVar(&server_type, "type", "", "rpc type")
	flag.StringVar(&server, "server", "", "server addr")
	flag.IntVar(&th_count, "th", 64, "thread count")
	flag.IntVar(&ss_count, "ss", 8, "session count")
	flag.Parse()

	log.Infow("param", "type", server_type, "server", server, "thread_count", th_count, "session_count", ss_count)
}

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	init_param()

	opt := jkpool.NewOptions()
	opt.MaxCap = ss_count
	opt.Factory = grpc_pools.GRPCClientFactory(
		func(conn *grpc.ClientConn) (server interface{}, err error) {
			return hellogrpc.New(conn)
		},
		grpc.WithInsecure(),
	)

	grpcPools, err := grpc_pools.NewGRPCPools([]string{server}, opt)
	if nil != err {
		log.Errorw("NewDefaultGRPCPoolsWithAddr error", "error", err)
		return
	}

	req := &pb.HelloRequest{}

	for i := 0; i < th_count; i++ {
		go func() {
			var err error
			for {
				// resp, err := grpcPools.Call("SayHello", &pb.HelloRequest{})
				// log.Infow("call respone", "resp", resp, "err", err)
				_, err = grpcPools.Call("SayHello", req)
				if nil != err {
					log.Panic(err)
				}
				atomic.AddUint64(&total, 1)
			}
		}()
	}

	statistics()
}
