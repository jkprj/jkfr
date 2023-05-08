package main

import (
	"flag"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	jkpool "github.com/jkprj/jkfr/gokit/transport/pool"
	rpcpool "github.com/jkprj/jkfr/gokit/transport/pool/rpc"
	"github.com/jkprj/jkfr/log"
)

type URequest struct {
	Name  string `json:"Name,omitempty"`
	Pause uint64 `json:"Pause,omitempty"`
}

type URespone struct {
	Msg   string `json:"Msg,omitempty"`
	Name  string `json:"Name,omitempty"`
	Pause uint64 `json:"Pause,omitempty"`
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
	flag.StringVar(&server, "server", "127.0.0.1:6666", "server addr")
	flag.IntVar(&th_count, "th", 64, "thread count")
	flag.IntVar(&ss_count, "ss", 8, "session count")
	flag.Parse()

	log.Infow("param", "type", server_type, "server", server, "thread_count", th_count, "session_count", ss_count)
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	init_param()

	go statistics()

	// test_pool()

	loop_test_pool()

	time.Sleep(time.Hour)
}

func loop_test_pool() {
	for i := 0; i < 100; i++ {
		go func() {
			for {
				test_pool()
			}
		}()
	}

	time.Sleep(time.Hour * 100000)
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
	opt.IdleTimeout = time.Minute
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
	var pause uint64 = 0
	var wg sync.WaitGroup

	for i := 0; i < th_count; i++ {

		wg.Add(1)

		go func(id int) {

			defer wg.Done()

			resp := URespone{}
			req := URequest{}

			for !bexit {
				req.Pause = atomic.AddUint64(&pause, 1)
				err := pls.Call("Hello.Hello", req, &resp)
				if nil != err {
					// log.Errorw("call Hello.Hello fail", "error", err)
					return
				}
				// time.Sleep(time.Second)
				if req.Pause != resp.Pause {
					log.Infow("call succ", "req.Pause", req.Pause, "resp.Pause", resp.Pause)
				}

				atomic.AddUint64(&total, 1)
			}
		}(i)
	}

	time.Sleep(time.Second * 5)

	if !bexit {
		bexit = true
	}

	// log.Info("aaaaaaaaaaaaaaaaaaa")

	// time.Sleep(time.Hour)

	// log.Info("bbbbbbbbbbbbbbbbbbb")

	pls.Close()

	// log.Info("ccccccccccccccccccc")

	wg.Wait()

	// log.Info("wwwwwwwwwwwwwwwwwww")

	// time.Sleep(time.Second)
}
