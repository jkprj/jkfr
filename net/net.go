package net

import (
	gnet "net"
	"strconv"
	"strings"

	jkrand "github.com/jkprj/jkfr/gokit/utils/rand"
)

func GetFreePort() (port int) {

	for i := 0; i < 100; i++ {

		port = jkrand.Intn(20000) + 30000

		l, err := gnet.Listen("tcp", "0.0.0.0:"+strconv.Itoa(port))
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

	if addr == "" {
		return "", 0, nil
	}

	if !strings.Contains(addr, "]") && !strings.Contains(addr, ":") { // Ipv4 没有端口情况
		return addr, 0, nil
	}

	if strings.Contains(addr, "]") && !strings.Contains(addr, "]:") { // ipv6 没有端口情况
		return addr, 0, nil
	}

	strPort := ""
	hostIP, strPort, err = gnet.SplitHostPort(addr)
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

	if port == 0 {
		port = GetFreePort()
		return host + ":" + strconv.Itoa(port), port, nil
	}

	return addr, port, nil
}

func ResolveAddr(addr string) (addrs []string, err error) {

	host, port, err := gnet.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	ips, err := gnet.LookupIP(host)
	if err != nil {
		return nil, err
	}

	uniqueIPs := map[string]bool{}

	for _, ip := range ips {

		strIP := ip.String()
		if _, ok := uniqueIPs[strIP]; !ok { // 去重
			uniqueIPs[strIP] = true
			addrs = append(addrs, gnet.JoinHostPort(strIP, port))
		}

	}

	return addrs, nil
}

func ResolveAddrWithRand(addr string) (addrs []string, err error) {

	addrs, err = ResolveAddr(addr)
	if err != nil {
		return nil, err
	}

	if len(addrs) > 1 {
		jkrand.Shuffle(len(addrs), func(i, j int) {
			addrs[i], addrs[j] = addrs[j], addrs[i]
		})
	}

	return addrs, err
}

// 将连接地址解析，将解析后的地址列表打乱，然后逐个尝试连接
// 该方法主要是为了解决当地址为域名时，域名解析后返回的地址列表顺序经常都是固定的，而大多数网络连接的库都是连接第一个连接地址，
// 这就会造成当用域名地址连接时，整个集群或连接池全都连到域名地址列表的第一个服务器，造成负载不均衡；
// 而且当返回的第一个地址的目标服务无效或异常时，就算后面列表的服务正常,还是会导致连接错误；
// 因此这里将地址解析出来将顺序打乱，然后尝试连接多个地址，这样就更大可能的保证能连上服务，也能很好的保证连接服务是负载均衡的
func ConnWithResolve(srv string, dial func(addr string) (gnet.Conn, error)) (conn gnet.Conn, err error) {

	addrs, err := ResolveAddrWithRand(srv)
	if err != nil {
		return nil, err
	}

	// 最大尝试连接3个地址
	if len(addrs) > 3 {
		addrs = addrs[:3]
	}

	for _, addr := range addrs {
		conn, err = dial(addr)
		if err != nil {
			continue
		}

		return conn, nil
	}

	return nil, err
}
