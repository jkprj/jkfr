package os

import (
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
)

func GetFreePort() (port int) {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < 100; i++ {

		port = rand.Intn(20000) + 30000

		l, err := net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(port))
		if err != nil {
			continue
		}
		defer l.Close()

		return port
	}

	return -1
}

// 解析服务地址
func ParseHostAddr(addr string) (hostIP string, port int, err error) {

	if "" == addr {
		return "", 0, nil
	}

	if strings.Index(addr, "]") == -1 && strings.Index(addr, ":") == -1 { // Ipv4 没有端口情况
		return addr, 0, nil
	}

	if strings.Index(addr, "]") != -1 && strings.Index(addr, "]:") == -1 { // ipv6 没有端口情况
		return addr, 0, nil
	}

	strPort := ""
	hostIP, strPort, err = net.SplitHostPort(addr)
	if nil != err {
		return "", 0, err
	}

	port, err = strconv.Atoi(strPort)
	if nil != err {
		return "", 0, err
	}

	return hostIP, port, nil
}

func GetRandomHostAddr(addr string) (string, int, error) {

	host, port, err := ParseHostAddr(addr)
	if nil != err {
		return "", 0, err
	}

	if 0 == port {
		port = GetFreePort()
		return host + ":" + strconv.Itoa(port), port, nil
	}

	return addr, port, nil
}
