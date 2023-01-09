package main

import (
	"encoding/json"
	"flag"
	"runtime"
	"time"

	// jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jktrans "github.com/jkprj/jkfr/gokit/transport"
	jkhttp "github.com/jkprj/jkfr/gokit/transport/http"
	jklog "github.com/jkprj/jkfr/log"
)

type res struct {
	Action  string
	Content string
}

type person struct {
	Name string
	Sex  string
	Age  int
}

var th_count int = 64

func init_param() {
	flag.IntVar(&th_count, "th", 64, "thread count")
	flag.Parse()

	jklog.Infow("param", "thread_count", th_count)
}

func main() {
	jklog.InitLogger()

	runtime.GOMAXPROCS(runtime.NumCPU())
	init_param()

	// defaultGet()
	// defaultPost()
	// defaultJSGet()
	// defaultJSPost()
	// queryByGlobleFuntionWithOpetion()
	// queryByClientHandle()
	// queryByClientHandleWithOption()

	// testStrategy()

	// queryWithDefaultConfFile()
	// queryWithConfFileOption()

	pressureTest()

}

func defaultGet() {
	buff, err := jkhttp.Get("http", "/test?Action=SayHello")

	jklog.Infow("request complete", "respone", string(buff), "err", err)
}

func defaultPost() {
	ps := person{Name: "LiLei", Age: 13, Sex: "man"}
	buf, _ := json.Marshal(ps)
	resBuf, err := jkhttp.Post("http", "/test?Action=WhatName", buf)

	jklog.Infow("request complete", "respone", string(resBuf), "err", err)
}

func defaultJSGet() {

	re := new(res)
	jkhttp.JSGet("http", "/test?Action=SayHello", re)

	jklog.Infow("request complete", "respone", re)
}

func defaultJSPost() {
	ps := person{Name: "LiLei", Age: 13, Sex: "man"}

	req := ps
	// or
	// req := &ps

	re := new(res)
	jkhttp.JSPost("http", "/test?Action=WhatName", req, re) // req 传指针类型或非指针类型都可以

	jklog.Infow("request complete", "respone", re)
}

func queryByClientHandle() {
	client, err := jkhttp.NewClient("http")
	if nil != err {
		jklog.Errorw("jkhttp.NewClient fail", "err", err)
		return
	}

	// Get demo
	resBuf, err := client.Get("/test?Action=SayHello")
	jklog.Infow("request complete", "respone", string(resBuf), "err", err)

	// post demo
	ps := person{Name: "LiLei", Age: 13, Sex: "man"}
	buf, _ := json.Marshal(ps)
	resBuf, err = client.Post("/test?Action=WhatName", buf)
	jklog.Infow("request complete", "post respone", string(resBuf), "err", err)

	// JSGet demo
	re := new(res)
	client.JSGet("/test?Action=SayHello", re)
	jklog.Infow("request complete", "respone", re)

	// JSPost demo
	client.JSPost("/test?Action=WhatName", ps, re)
	jklog.Infow("request complete", "respone", re)

}

func queryByGlobleFuntionWithOpetion() {
	err := jkhttp.RegistryNewClient("http",
		jkhttp.ClientStrategy(jktrans.STRATEGY_RANDOM),
		jkhttp.ClientLimit(2),
		jkhttp.ClientConsulTags("jinkun"),
		jkhttp.ClientRetry(5),
		jkhttp.ClientTimeOut(10),
		jkhttp.ClientScheme(jktrans.HTTP),
	)
	// or
	// client, err := jkhttp.NewClient("http",
	// 	jkhttp.ClientStrategy(jktrans.STRATEGY_RANDOM),
	// 	jkhttp.ClientLimit(2),
	// 	jkhttp.ClientConsulTags("jinkun"),
	// 	jkhttp.ClientRetry(5),
	// 	jkhttp.ClientTimeOut(10),
	// 	jkhttp.ClientScheme(jktrans.HTTP),
	// )
	// jkhttp.RegistryClient(client)
	if nil != err {
		jklog.Errorw("jkhttp.RegistryNewClient fail", "err", err)
		return
	}

	// Get demo
	resBuf, err := jkhttp.Get("http", "/test?Action=SayHello")
	jklog.Infow("request complete", "respone", string(resBuf), "err", err)

	// post demo
	ps := person{Name: "LiLei", Age: 13, Sex: "man"}
	buf, _ := json.Marshal(ps)
	resBuf, err = jkhttp.Post("http", "/test?Action=WhatName", buf)
	jklog.Infow("request complete", "post respone", string(resBuf), "err", err)

	// JSGet demo
	re := new(res)
	jkhttp.JSGet("http", "/test?Action=SayHello", re)
	jklog.Infow("request complete", "respone", re)

	// JSPost demo
	jkhttp.JSPost("http", "/test?Action=WhatName", ps, re)
	jklog.Infow("request complete", "request complete", "respone", re)
}

func queryByClientHandleWithOption() {
	client, err := jkhttp.NewClient("http",
		jkhttp.ClientStrategy(jktrans.STRATEGY_RANDOM),
		jkhttp.ClientLimit(2),
		jkhttp.ClientConsulTags("jinkun"),
		jkhttp.ClientRetry(5),
		jkhttp.ClientTimeOut(10),
		jkhttp.ClientScheme(jktrans.HTTP),
	)
	if nil != err {
		jklog.Errorw("jkhttp.NewClient fail", "err", err)
		return
	}

	// Get demo
	resBuf, err := client.Get("/test?Action=SayHello")
	jklog.Infow("request complete", "respone", string(resBuf), "err", err)

	// post demo
	ps := person{Name: "LiLei", Age: 13, Sex: "man"}
	buf, _ := json.Marshal(ps)
	resBuf, err = client.Post("/test?Action=WhatName", buf)
	jklog.Infow("request complete", "post respone", string(resBuf), "err", err)

	// JSGet demo
	re := new(res)
	client.JSGet("/test?Action=SayHello", re)
	jklog.Infow("request complete", "respone", re)

	// JSPost demo
	client.JSPost("/test?Action=WhatName", ps, re)
	jklog.Infow("request complete", "respone", re)

}

func queryWithDefaultConfFile() {
	// Get demo
	resBuf, err := jkhttp.Get("httpT", "/test?Action=SayHello")
	jklog.Infow("request complete", "respone", string(resBuf), "err", err)

	// post demo
	ps := person{Name: "LiLei", Age: 13, Sex: "man"}
	buf, _ := json.Marshal(ps)
	resBuf, err = jkhttp.Post("httpT", "/test?Action=WhatName", buf)
	jklog.Infow("request complete", "post respone", string(resBuf), "err", err)

	// JSGet demo
	re := new(res)
	jkhttp.JSGet("httpT", "/test?Action=SayHello", re)
	jklog.Infow("request complete", "respone", re)

	// JSPost demo
	jkhttp.JSPost("httpT", "/test?Action=WhatName", ps, re)
	jklog.Infow("request complete", "respone", re)
}

func queryWithConfFileOption() {
	err := jkhttp.RegistryNewClient("http", jkhttp.ClientConfigFile("conf/httpJ.json"))
	// or
	// err := jkhttp.RegistryNewClient("http", jkhttp.ClientConfigFile("conf/httpT.toml"))

	if nil != err {
		jklog.Errorw("jkhttp.NewClient fail", "err", err)
		return
	}

	// Get demo
	resBuf, err := jkhttp.Get("http", "/test?Action=SayHello")
	jklog.Infow("request complete", "respone", string(resBuf), "err", err)

	// post demo
	ps := person{Name: "LiLei", Age: 13, Sex: "man"}
	buf, _ := json.Marshal(ps)
	resBuf, err = jkhttp.Post("http", "/test?Action=WhatName", buf)
	jklog.Infow("request complete", "post respone", string(resBuf), "err", err)

	// JSGet demo
	re := new(res)
	jkhttp.JSGet("http", "/test?Action=SayHello", re)
	jklog.Infow("request complete", "respone", re)

	// JSPost demo
	jkhttp.JSPost("http", "/test?Action=WhatName", ps, re)
	jklog.Infow("request complete", "respone", re)
}

func testStrategy() {
	ps := person{Name: "LiLei", Age: 13, Sex: "man"}

	jkhttp.RegistryNewClient("http", jkhttp.ClientStrategy(jktrans.STRATEGY_RANDOM)) // 默认策略是轮询，其他策略需要设置ClientStrategy 选项设置
	for i := 0; i < 10; i++ {
		re := new(res)
		jkhttp.JSPost("http", "/test?Action=WhatName", ps, re)

		jklog.Infow("request complete", "respone", re)

	}
}

func pressureTest() {

	count := 0

	for i := 0; i < th_count; i++ {
		go func() {
			ps := person{Name: "LiLei", Age: 13, Sex: "man"}
			re := new(res)
			for {
				/*buff, err := */ jkhttp.JSPost("httpT", "/test?Action=WhatName", ps, re)
				count++
			}
		}()
	}

	for {
		time.Sleep(time.Second)
		jklog.Infow("request statistics", "count:", count)
		count = 0
	}
}
