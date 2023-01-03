package main

import (
	"flag"
	jkpool "jkfr/gokit/transport/pool"
	rpcpool "jkfr/gokit/transport/pool/rpc"
	"jkfr/log"
	"runtime"
	"sync/atomic"
	"time"
)

type URequest struct {
	Name  string
	Pause int
}

type URespone struct {
	Msg   string
	Name  string
	Pause int
}

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

	go statistics()

	test_pool()

	// for {
	// 	test_pool()
	// }

	time.Sleep(time.Hour)
}

func test_pool() {
	// pls, err := rpcpool.NewDefaultTlsRpcPoolsWithAddr([]string{"127.0.0.1:6666", "127.0.0.1:6667"}, RpcClientPem, RpcClientKey)
	// pls, err := rpcpool.NewDefaultTlsRpcPoolsWithAddr(
	// 	[]string{
	// 		"192.168.137.90:6661",
	// 		"192.168.137.90:6662",
	// 		// "192.168.137.90:6663",
	// 		// "192.168.137.90:6664",
	// 		// "192.168.137.90:6665",
	// 		// "192.168.137.90:6666",
	// 	},
	// 	RpcClientPem,
	// 	RpcClientKey,
	// )
	// pls, err := rpcpool.NewDefaultRpcPoolsWithAddr(
	// 	[]string{
	// 		"127.0.0.1:6666",
	// 	},
	// )

	opt := jkpool.NewOptions()
	opt.MaxCap = ss_count
	pls, err := rpcpool.NewRpcPools(
		[]string{
			server,
		},
		opt,
	)
	if nil != err {
		log.Errorw("connect server fail", "error", err)
		return
	}

	bexit := false

	for i := 0; i < th_count; i++ {

		go func(id int) {
			resp := URespone{}

			for !bexit {
				err := pls.Call("HelloWord.Hello", URequest{}, &resp)
				if nil != err {
					log.Errorw("call Hello.Hello fail", "error", err)
					return
				}

				atomic.AddUint64(&total, 1)
			}
		}(i)
	}

	time.Sleep(time.Hour * 24)

	if !bexit {
		bexit = true
	}

	log.Info("aaaaaaaaaaaaaaaaaaa")

	time.Sleep(time.Hour)

	log.Info("bbbbbbbbbbbbbbbbbbb")

	pls.Close()

	log.Info("ccccccccccccccccccc")

}
