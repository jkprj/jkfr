package net

import (
	"fmt"
	gnet "net"
	"testing"
	"time"
)

func TestResolveAddrWithRand(t *testing.T) {
	addrs, err := ResolveAddrWithRand("www.ucloud.cn:80")
	fmt.Println("addrs:", addrs, ", err:", err)
}

func testConnWithResolve(svr string) {

	addrs, err := ResolveAddr(svr)
	fmt.Println("addrs:", addrs, ", err:", err)

	conn, err := ConnWithResolve(svr, func(addr string) (gnet.Conn, error) {
		fmt.Println("in-addr:", addr)
		return gnet.DialTimeout("tcp", addr, time.Second*5)
	})
	if err != nil {
		fmt.Printf("err:%s\n", err.Error())
		return
	}
	defer conn.Close()

	fmt.Println("conn-addr:", conn.RemoteAddr())
}

func TestConnWithResolve(t *testing.T) {
	testConnWithResolve("www.ucloud.cn:80")
	fmt.Print("\n\n-----------------\n\n")
	testConnWithResolve("www.ucloud.cn:22")
}
