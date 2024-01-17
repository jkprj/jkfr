package main

import (
	"flag"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	jkpool "github.com/jkprj/jkfr/gokit/transport/pool"
	grpc_pools "github.com/jkprj/jkfr/gokit/transport/pool/grpc"
	"github.com/jkprj/jkfr/log"
	pb "github.com/jkprj/jkfr/protobuf/demo"
	hellogrpc "github.com/jkprj/jkfr/protobuf/demo/hello-service/svc/client/grpc"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	go test_pool()
	// loop_test()

	statistics()
}

func loop_test() {
	for i := 0; i < 100; i++ {
		go func() {
			for {
				test_pool()
			}
		}()
	}
}

func test_pool() {

	opt := jkpool.NewOptions()
	opt.MaxCap = ss_count
	opt.Factory = grpc_pools.GRPCClientFactory(
		func(conn *grpc.ClientConn) (server interface{}, err error) {
			return hellogrpc.New(conn)
		},
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

	grpcPools, err := grpc_pools.NewGRPCPools(strings.Split(server, ","), opt)
	if nil != err {
		log.Errorw("NewDefaultGRPCPoolsWithAddr error", "error", err)
		return
	}

	req := &pb.HelloRequest{}
	wg := sync.WaitGroup{}
	bexit := false

	for i := 0; i < th_count; i++ {

		wg.Add(1)

		go func() {

			defer wg.Done()

			var err error
			for !bexit {
				// resp, err := grpcPools.Call("SayHello", &pb.HelloRequest{})

				_, err = grpcPools.Call("SayHello", req)
				if nil != err {
					// log.Errorw("call respone", "resp", resp, "err", err)
					continue
				}
				atomic.AddUint64(&total, 1)
			}
		}()
	}

	time.Sleep(time.Hour * 64)

	bexit = true

	grpcPools.Close()

	// log.Infow("bexit bexit bexit bexit")

	wg.Wait()

	// log.Infow("Wait Wait Wait Wait")
}
