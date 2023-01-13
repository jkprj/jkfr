package main

import (
	"time"

	jkregistry "github.com/jkprj/jkfr/gokit/registry"
	jklog "github.com/jkprj/jkfr/log"
)

func main() {
	jklog.InitLogger()

	RegistryDefaultServer()
	// RegistryRegistryWithOption()
	// RegistryMutiDefaultServer()
	// RegistryMutiServerWithOption()
	// RegistryServerWithFile()

}

func RegistryDefaultServer() {
	// registry, err := jkregistry.RegistryServerWithServerAddr("test", "127.0.0.1:9999")
	registry, err := jkregistry.RegistryServer("test1")
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}

	time.Sleep(time.Minute)

	registry.Deregister()
}

func RegistryServerWithOption() {
	registry, err := jkregistry.RegistryServer("test",
		jkregistry.WithServerAddr("192.168.213.184:9999"),
		jkregistry.WithTags("qjk", "test", "123"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithHealthCheckAddr("192.168.213.184:9876"),
		jkregistry.WithConsulAddr("192.168.213.184:8500"))

	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}

	time.Sleep(time.Minute)

	registry.Deregister()
}

// 测试注册多个默认配置服务器，根据测试：name+serverAddr确定唯一性，相同的name+serverAddr只会注册第一个
func RegistryMutiDefaultServer() {
	regtest1, err := jkregistry.RegistryServer("test", jkregistry.WithServerAddr("127.0.0.1:8888"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regtest1.Deregister()

	// name + serverAddr如果一样，consul会注册第一个实例
	regtest2, err := jkregistry.RegistryServer("test", jkregistry.WithServerAddr("127.0.0.1:8888"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regtest2.Deregister()

	// 同名称_同ip_不同端 口可同时注册
	regtest3, err := jkregistry.RegistryServer("test", jkregistry.WithServerAddr("127.0.0.1:9999"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regtest3.Deregister()

	// 同名称_不同ip_同端口 可同时注册
	regtest4, err := jkregistry.RegistryServer("test", jkregistry.WithServerAddr("192.168.213.184:8888"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regtest4.Deregister()

	// 同名称_不同ip_不同端 口可同时注册
	regtest5, err := jkregistry.RegistryServer("test", jkregistry.WithServerAddr("192.168.213.184:9999"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regtest5.Deregister()

	// 不同名称_同ip_同端口 可同时注册
	regjk1, err := jkregistry.RegistryServer("jinkun", jkregistry.WithServerAddr("127.0.0.1:8888"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regjk1.Deregister()

	// 不同名称_同ip_不同端口可同时注册
	regjk2, err := jkregistry.RegistryServer("jinkun", jkregistry.WithServerAddr("127.0.0.1:9999"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regjk2.Deregister()

	// 不同名称_不同ip_同端口可同时注册
	regjk3, err := jkregistry.RegistryServer("jinkun", jkregistry.WithServerAddr("192.168.213.184:8888"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regjk3.Deregister()

	// 不同名称_不同ip_不同端口 可同时注册
	regjk4, err := jkregistry.RegistryServer("jinkun", jkregistry.WithServerAddr("192.168.213.184:9999"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regjk4.Deregister()

	time.Sleep(time.Minute * 5)

}

// 1、测试注册多个默认配置服务器，根据测试：name+serverAddr确定唯一性，相同的name+serverAddr只会注册第一个
// 2、不同的consul 服务集群，可以同时注册一组相同的服务
func RegistryMutiServerWithOption() {

	regtest1, err := jkregistry.RegistryServer("test",
		jkregistry.WithServerAddr("192.168.213.184:9999"),
		jkregistry.WithTags("qjk", "test", "123"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithHealthCheckAddr("192.168.213.184"),
		jkregistry.WithConsulAddr("192.168.213.184:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regtest1.Deregister()

	// name + serverAddr如果一样，consul会注册第一个实例
	regtest2, err := jkregistry.RegistryServer("test",
		jkregistry.WithServerAddr("192.168.213.184:9999"),
		jkregistry.WithTags("qjk"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("192.168.213.184:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regtest2.Deregister()

	// 同名称_不同ip_同端口 可同时注册
	regtest3, err := jkregistry.RegistryServer("test",
		jkregistry.WithServerAddr("127.0.0.1:9999"),
		jkregistry.WithTags("qjk", "123"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("192.168.213.184:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regtest3.Deregister()

	// 同名称_同ip_不同端口 可同时注册
	regtest4, err := jkregistry.RegistryServer("test",
		jkregistry.WithServerAddr("192.168.213.184:8888"),
		jkregistry.WithTags("qjk"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("192.168.213.184:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regtest4.Deregister()

	// 同名称不同ip不同端口可同时注册
	regtest5, err := jkregistry.RegistryServer("test",
		jkregistry.WithServerAddr("127.0.0.1:8888"),
		jkregistry.WithTags("qjk", "123"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("192.168.213.184:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regtest5.Deregister()

	// 不同名称_同ip_同端口 可同时注册
	regtest6, err := jkregistry.RegistryServer("jinkun",
		jkregistry.WithServerAddr("192.168.213.184:9999"),
		jkregistry.WithTags("qjk"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("192.168.213.184:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regtest6.Deregister()

	// 不同名称_同ip_不同端口 可同时注册
	regtest7, err := jkregistry.RegistryServer("jinkun",
		jkregistry.WithServerAddr("192.168.213.184:8888"),
		jkregistry.WithTags("qjk", "123"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("192.168.213.184:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regtest7.Deregister()

	// 不同名称_不同ip_同端口 可同时注册
	regtest8, err := jkregistry.RegistryServer("jinkun",
		jkregistry.WithServerAddr("127.0.0.1:9999"),
		jkregistry.WithTags("qjk"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("192.168.213.184:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regtest8.Deregister()

	// 不同名称_不同ip_不同端口 可同时注册
	regtest9, err := jkregistry.RegistryServer("jinkun",
		jkregistry.WithServerAddr("127.0.0.1:8888"),
		jkregistry.WithTags("qjk"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("192.168.213.184:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer regtest9.Deregister()

	//++++++++++++++++不同的consul 服务集群 ++++++++++++++

	// 不同的consul 服务集群，可以同时注册一组相同的服务

	reg1, err := jkregistry.RegistryServer("test",
		jkregistry.WithServerAddr("192.168.213.184:9999"),
		jkregistry.WithTags("qjk", "test", "123"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("127.0.0.1:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer reg1.Deregister()

	// name + serverAddr如果一样，consul会注册第一个实例
	reg2, err := jkregistry.RegistryServer("test",
		jkregistry.WithServerAddr("192.168.213.184:9999"),
		jkregistry.WithTags("qjk"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("127.0.0.1:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer reg2.Deregister()

	// 同名称_不同ip_同端口 可同时注册
	reg3, err := jkregistry.RegistryServer("test",
		jkregistry.WithServerAddr("127.0.0.1:9999"),
		jkregistry.WithTags("qjk", "123"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("127.0.0.1:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer reg3.Deregister()

	// 同名称_同ip_不同端口 可同时注册
	reg4, err := jkregistry.RegistryServer("test",
		jkregistry.WithServerAddr("192.168.213.184:8888"),
		jkregistry.WithTags("qjk"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("127.0.0.1:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer reg4.Deregister()

	// 同名称不同ip不同端口可同时注册
	reg5, err := jkregistry.RegistryServer("test",
		jkregistry.WithServerAddr("127.0.0.1:8888"),
		jkregistry.WithTags("qjk", "123"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("127.0.0.1:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer reg5.Deregister()

	// 不同名称_同ip_同端口 可同时注册
	reg6, err := jkregistry.RegistryServer("jinkun",
		jkregistry.WithServerAddr("192.168.213.184:9999"),
		jkregistry.WithTags("qjk"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("127.0.0.1:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer reg6.Deregister()

	// 不同名称_同ip_不同端口 可同时注册
	reg7, err := jkregistry.RegistryServer("jinkun",
		jkregistry.WithServerAddr("192.168.213.184:8888"),
		jkregistry.WithTags("qjk", "123"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("127.0.0.1:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer reg7.Deregister()

	// 不同名称_不同ip_同端口 可同时注册
	reg8, err := jkregistry.RegistryServer("jinkun",
		jkregistry.WithServerAddr("127.0.0.1:9999"),
		jkregistry.WithTags("qjk"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("127.0.0.1:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer reg8.Deregister()

	// 不同名称_不同ip_不同端口 可同时注册
	reg9, err := jkregistry.RegistryServer("jinkun",
		jkregistry.WithServerAddr("127.0.0.1:8888"),
		jkregistry.WithTags("qjk", "123"),
		jkregistry.WithHealthCheckInterval(1),
		jkregistry.WithHealthCheckTimeOut(1),
		jkregistry.WithConsulAddr("127.0.0.1:8500"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer reg9.Deregister()

	time.Sleep(time.Hour * 24)
}

func RegistryServerWithFile() {
	reg1, err := jkregistry.RegistryServer("qjk")
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer reg1.Deregister()

	reg2, err := jkregistry.RegistryServer("jk")
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer reg2.Deregister()

	// 使用加载默认配置文件方式注册服务，那么一个服务名称只能注册一个服务，如果需要，可以使用以下方式指定配置文件
	reg3, err := jkregistry.RegistryServer("jk", jkregistry.WithFile("conf/test2.json"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer reg3.Deregister()

	reg4, err := jkregistry.RegistryServer("qjk", jkregistry.WithFile("conf/test1.toml"))
	if nil != err {
		jklog.Errorw("jkregistry.RegistryServer fail", "err", err)
		return
	}
	defer reg4.Deregister()

	time.Sleep(time.Minute)
}
