package os

import (
	"errors"
	"math"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	jklog "github.com/jkprj/jkfr/log"
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

	// if "" == addr {
	// 	return "", 0, errors.New("Server addr must not be empty")
	// }

	spAddr := strings.Split(addr, ":")

	if 1 > len(spAddr) {
		return "", 0, errors.New("Server addr invalid, addr:" + addr)
	}

	hostIP = spAddr[0]
	// if "" != hostIP {
	// 	ip := net.ParseIP(hostIP)
	// 	if nil == ip {
	// 		jklog.Errorw("Server addr's IP invalid", "addr", addr, "hostIP", hostIP, "port", port, "err", err)
	// 		return "", 0, errors.New("Server addr's IP invalid, addr:" + addr)
	// 	}
	// }

	if 1 == len(spAddr) {
		return hostIP, 0, nil
	}

	port, err = strconv.Atoi(spAddr[1])
	if nil != err || port < 0 || port >= int(math.Pow(2, 16)) {
		jklog.Errorw("Server addr's port invalid", "addr", addr, "hostIP", hostIP, "port", port, "err", err)
		return "", 0, errors.New("Server addr's port invalid, addr:" + addr)
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
