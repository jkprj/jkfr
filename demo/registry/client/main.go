package main

import (
	"runtime"
	"time"

	jkregistry "jkfr/gokit/registry"
	jklog "jkfr/log"

	"github.com/hashicorp/consul/api"
)

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	jklog.InitLogger(&jklog.ZapOptions{LogLevel: "debug"})

	queryDefaultRegistry()
	// queryRegistryWithOption()
	// queryRegistryWithDefaultFile()
	// queryRegistryWithFile()

	// for i := 0; i < 99; i++ {
	// 	go queryRegistryWithOption()
	// }

	for {
		queryRegistryWithOption()
		time.Sleep(time.Second)
	}

}

func queryDefaultRegistry() {
	svrs, _, err := jkregistry.Services("test")

	jklog.Debugw("test services", "svrs", svrs, "err", err)
}

func queryRegistryWithOption() {
	svrs, _, err := jkregistry.Services("jinkun",
		jkregistry.WithTags("123"),
		jkregistry.WithPassingOnly(false),
		jkregistry.WithQueryOptions(&api.QueryOptions{UseCache: true}),
		jkregistry.WithConsulAddr("192.168.213.184:8500"),
	)

	jklog.Debugw("test services", "svrs", svrs, "err", err)
}

func queryRegistryWithDefaultFile() {
	svrs, _, err := jkregistry.Services("test")

	jklog.Debugw("test services", "svrs", svrs, "err", err)
}

func queryRegistryWithFile() {
	svrs, _, err := jkregistry.Services("jinkun", jkregistry.WithFile("conf/test2.json"))

	jklog.Debugw("test services", "svrs", svrs, "err", err)
}
