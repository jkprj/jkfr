package grpc

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net/rpc"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc"

	jkpool "github.com/jkprj/jkfr/gokit/transport/pool"
	jkutils "github.com/jkprj/jkfr/gokit/utils"
	jkrand "github.com/jkprj/jkfr/gokit/utils/rand"
	jklog "github.com/jkprj/jkfr/log"
)

type GetPoolFunc func() *stpool

type UCall struct {
	rpc.Call
	Done chan *UCall
}

type stpool struct {
	pl     *GRPCPool
	last   time.Time
	call   int64
	static bool // 静态创建的pool不能超时关闭，只有动态创建的pool才能超时关闭
}

type GRPCPools struct {
	addr2pool map[string]*stpool
	pools     []*stpool
	opt       *jkpool.Options

	get_pool GetPoolFunc

	retryTimes    uint
	retryInterval time.Duration
	idleTimeOut   time.Duration
	strategy      string

	index  uint32
	random *rand.Rand

	mtPool sync.RWMutex
	mtNew  sync.RWMutex

	isClosed bool
	chExit   chan int
}

func NewDefaultGRPCPools(clientFatory ClientFatory) *GRPCPools {
	pls, _ := NewDefaultGRPCPoolsWithAddr([]string{}, clientFatory)
	return pls
}

func NewDefaultGRPCPoolsWithAddr(addrs []string, clientFatory ClientFatory) (*GRPCPools, error) {

	opt := jkpool.NewOptions()
	opt.Factory = GRPCClientFactory(clientFatory, grpc.WithInsecure())

	return NewGRPCPools(addrs, opt)
}

func NewGRPCPools(addrs []string, opt *jkpool.Options) (*GRPCPools, error) {
	p := new(GRPCPools)
	p.addr2pool = make(map[string]*stpool)
	p.pools = make([]*stpool, 0, len(addrs))
	p.opt = opt
	p.SetRetryTimes(3)
	p.SetIdleTimeOut(600)
	p.SetStrategy(jkutils.STRATEGY_LEAST)
	p.SetRetryIntervalMS(1000)

	// p.random = rand.New(rand.NewSource(time.Now().UnixNano()))      // golang提供的source不是线程安全的
	p.random = rand.New(jkrand.NewSource(time.Now().UnixNano()))

	if nil != addrs {
		for _, addr := range addrs {
			_, err := p.get_and_push(addr, opt, true)
			if nil != err {
				jklog.Infow("GRPCPools.get_and_push err, to close pool", "err", err.Error())
				p.Close()
				return nil, err
			}
		}
	}

	p.chExit = make(chan int)
	go p.loop_check_idle_time_out_pool()

	return p, nil
}

func (pls *GRPCPools) RetryTimes() uint {
	return pls.retryTimes
}

func (pls *GRPCPools) call(ctx context.Context, serviceMethod string, args interface{}) (resp interface{}, err error) {

	pl := pls.get_pool()
	if nil == pl {
		jklog.Errorw("not found server")
		return nil, errors.New("not found server")
	}

	atomic.AddInt64(&pl.call, 1)

	resp, err = pl.pl.CallWithContext(ctx, serviceMethod, args)
	// if nil != err {
	// 	log.Errorw("Call fail", "method", serviceMethod, "error", err)
	// }

	atomic.AddInt64(&pl.call, -1)
	pl.last = time.Now()

	return resp, err
}

func (pls *GRPCPools) call_with_func(callFunc func() (resp interface{}, err error)) (resp interface{}, err error) {

	var retry uint = 0

	var lastErr error

	for {

		if pls.isClosed {
			return nil, jkpool.ErrClosed
		}

		resp, err = callFunc()
		if nil == err {
			return resp, nil
		}

		retry++

		// log.Errorw("Call fail", "retryTimes", retry, "error", err)

		if nil != lastErr {
			lastErr = fmt.Errorf("%s; %s", lastErr.Error(), err.Error())
		} else {
			lastErr = err
		}

		if pls.retryTimes <= retry {
			break
		}

		time.Sleep(pls.retryInterval)
	}

	return nil, lastErr
}

func (pls *GRPCPools) CallWithTimeOut(serviceMethod string, args interface{}, timeout time.Duration) (resp interface{}, err error) {

	return pls.call_with_func(func() (resp interface{}, err error) {

		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		resp, err = pls.call(ctx, serviceMethod, args)

		cancel()

		return resp, err
	})
}

func (pls *GRPCPools) CallWithContext(ctx context.Context, serviceMethod string, args interface{}) (resp interface{}, err error) {

	return pls.call_with_func(func() (resp interface{}, err error) {
		return pls.call(ctx, serviceMethod, args)
	})
}

func (pls *GRPCPools) Call(serviceMethod string, args interface{}) (resp interface{}, err error) {
	return pls.CallWithContext(context.Background(), serviceMethod, args)
}

func (pls *GRPCPools) GoCallWithContext(ctx context.Context, serviceMethod string, args interface{}, done chan *UCall) *UCall {

	call := new(UCall)
	call.ServiceMethod = serviceMethod
	call.Args = args
	call.Error = nil

	if done == nil {
		done = make(chan *UCall, 10)
	}

	call.Done = done

	go func() {
		call.Reply, call.Error = pls.CallWithContext(ctx, serviceMethod, args)
		if nil != call.Error {
			jklog.Errorw("GRPCPools.call fail", "error", call.Error)
		}

		done <- call
	}()

	return call
}

func (pls *GRPCPools) GoCall(serviceMethod string, args interface{}, resp interface{}, done chan *UCall) *UCall {
	return pls.GoCallWithContext(context.Background(), serviceMethod, args, done)
}

func (pls *GRPCPools) CallWithAddr(addr string, serviceMethod string, args interface{}) (resp interface{}, err error) {

	return pls.CallWithAddrEx(addr, serviceMethod, args, 60*time.Second)
}

func (pls *GRPCPools) CallWithAddrEx(addr string, serviceMethod string, args interface{}, timeout time.Duration) (resp interface{}, err error) {

	return pls.call_with_func(func() (resp interface{}, err error) {

		pl, err := pls.getex(addr)
		if nil != err {
			jklog.Errorw("getex client fail", "addr", addr, "method", serviceMethod, "error", err)
			return nil, err
		}

		atomic.AddInt64(&pl.call, 1)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		resp, err = pl.pl.CallWithContext(ctx, serviceMethod, args)

		cancel()

		pl.last = time.Now()
		atomic.AddInt64(&pl.call, -1)

		return resp, err
	})
}

func (pls *GRPCPools) Close() {
	pls.mtNew.Lock()
	defer pls.mtNew.Unlock()

	// log.Infow("GRPCPools.Close")

	pls.close()
}

func (pls *GRPCPools) close() {

	pls.mtPool.Lock()
	defer pls.mtPool.Unlock()

	if pls.isClosed {
		return
	}

	pls.isClosed = true

	if pls.chExit != nil {
		pls.chExit <- 1
	}

	// log.Infow("GRPCPools.close")

	for _, pl := range pls.pools {
		pl.pl.Close()
	}

	pls.addr2pool = make(map[string]*stpool)
	pls.pools = make([]*stpool, 0)
}

func (pls *GRPCPools) CloseServer(svrAddr string) {
	pls.remove_index(pls.get_server_index(svrAddr))
}

func (pls *GRPCPools) SetRetryTimes(times uint) {

	const MAX_TIMES = 10

	if times > MAX_TIMES {
		pls.retryTimes = MAX_TIMES
	} else if times <= 0 {
		pls.retryTimes = 1
	} else {
		pls.retryTimes = times
	}
}

func (pls *GRPCPools) SetRetryIntervalMS(intervalMS uint) {

	const MIN_INTERVAL_MS = 50

	if intervalMS < MIN_INTERVAL_MS {
		intervalMS = MIN_INTERVAL_MS
	}

	pls.retryInterval = time.Duration(intervalMS) * time.Millisecond
}

func (pls *GRPCPools) SetIdleTimeOut(timeOut uint) {

	pls.idleTimeOut = time.Duration(timeOut) * time.Second

	if pls.idleTimeOut < time.Minute {
		pls.idleTimeOut = time.Minute
	}
}

func (pls *GRPCPools) SetStrategy(strategy string) {

	pls.strategy = strategy

	if jkutils.STRATEGY_RANDOM == strategy {
		pls.get_pool = func() *stpool {
			return pls.random_get()
		}

	} else if jkutils.STRATEGY_ROUND == strategy {
		pls.get_pool = func() *stpool {
			return pls.roll_get()
		}
	} else {
		pls.strategy = jkutils.STRATEGY_LEAST
		pls.get_pool = func() *stpool {
			return pls.least_get()
		}
	}
}

func (pls *GRPCPools) get_and_push(addr string, opt *jkpool.Options, static bool) (*stpool, error) {
	pl := pls.get(addr)
	if nil != pl {
		return pl, nil
	}

	pl, err := pls.new_and_push(addr, opt, static)
	if nil != err {
		jklog.Errorw("new_and_push fail", "addr", addr, "opt", opt, "error", err)
		return nil, err
	}

	return pl, nil
}

func (pls *GRPCPools) push(addr string, pl *stpool) *stpool {

	pls.mtPool.Lock()
	defer pls.mtPool.Unlock()

	tmp, ok := pls.addr2pool[addr]
	if ok {
		jklog.Infow("GRPCPools push poll, to close old pool")
		tmp.pl.Close()
		tmp.pl = pl.pl
		tmp.static = pl.static
		tmp.last = time.Now()

		return tmp
	}

	pls.addr2pool[addr] = pl
	pls.pools = append(pls.pools, pl)

	return pl
}

func (pls *GRPCPools) new_and_push(addr string, opt *jkpool.Options, static bool) (*stpool, error) {

	pls.mtNew.Lock()
	defer pls.mtNew.Unlock()

	if pls.isClosed {
		return nil, errors.New("pools is closed")
	}

	pl := pls.get(addr)
	if nil != pl {
		return pl, nil
	}

	tmp_opt := *opt
	tmp_opt.ServerAddr = addr

	tmp, err := NewGRPCPool(&tmp_opt)
	if nil != err {
		jklog.Errorw("NewRpcPool fail", "addr", addr, "opt", opt, "error", err)
		return nil, err
	}

	tmpPL := &stpool{pl: tmp, last: time.Now(), static: static}

	return pls.push(addr, tmpPL), nil
}

func (pls *GRPCPools) get(addr string) *stpool {

	pls.mtPool.RLock()

	p, ok := pls.addr2pool[addr]
	if ok {
		pls.mtPool.RUnlock()
		return p
	}

	pls.mtPool.RUnlock()

	return nil
}

func (pls *GRPCPools) roll_get() *stpool {

	pls.index++

	return pls.get_index_ex(pls.index)
}

func (pls *GRPCPools) random_get() *stpool {

	pls.index = pls.random.Uint32()

	return pls.get_index_ex(pls.index)
}

func (pls *GRPCPools) least_get() *stpool {

	pls.mtPool.RLock()

	nlen := uint32(len(pls.pools))
	if nlen == 0 {
		pls.mtPool.RUnlock()
		return nil
	}

	// 当连接池实例小于LEAST_ROUND_MAX时，使用循环遍历查找请求最小的实例
	// 当连接池实例大于LEAST_ROUND_MAX时，使用随机抽取LEAST_RAND_COUNT个实例，选取最小的那个
	if nlen <= jkutils.LEAST_ROUND_MAX {
		pls.index = pls.least_get_when_less()
	} else {
		pls.index = pls.least_get_when_bigger()
	}

	pls.mtPool.RUnlock()

	return pls.get_index_ex(pls.index)
}

func (pls *GRPCPools) least_get_when_less() (index uint32) {
	var min int64 = math.MaxInt64
	nlen := uint32(len(pls.pools))

	// 使用随机数开始下标是为了防止当客户端的请求频率比较小的时候所有的请求都打到第一个服务上
	// 使用随机数后，那么每次都是从随机下标开始扫描pool，这样就可以防止请求量较小时一直都是发第一个
	// 主要就是为了解决客户端比较多而每个客户端请求频率又是比较少的情况，防止所有客户端都把请求发到第一个服务了
	bgIndex := pls.random.Uint32() % nlen

	for i := uint32(0); i < nlen; i++ {

		cur := (bgIndex + i) % nlen
		pl := pls.pools[cur]

		if min > pl.call {
			min = pl.call
			index = cur
		}
		if min == 0 {
			break
		}
	}

	return index
}

func (pls *GRPCPools) least_get_when_bigger() (index uint32) {

	var min int64 = math.MaxInt64
	nlen := uint32(len(pls.pools))

	// 随机抽取LEAST_RAND_COUNT个选最小那个返回
	for i := 0; i < jkutils.LEAST_RAND_COUNT; i++ {

		cur := pls.random.Uint32() % nlen
		pl := pls.pools[cur]

		if min > pl.call {
			min = pl.call
			index = cur
		}
		if min == 0 {
			break
		}
	}

	return index
}

func (pls *GRPCPools) getex(addr string) (*stpool, error) {
	return pls.get_and_push(addr, pls.opt, false)
}

func (pls *GRPCPools) remove_index(i int) {

	pls.mtPool.Lock()
	defer pls.mtPool.Unlock()

	if i < 0 || i >= len(pls.pools) {
		return
	}

	tmp := pls.pools[i]

	if i+1 < len(pls.pools) {
		pls.pools = append(pls.pools[:i], pls.pools[i+1:]...)
	} else {
		pls.pools = pls.pools[:i]
	}

	delete(pls.addr2pool, tmp.pl.addr)

	jklog.Infow("GRPCPools.remove_pool to close pool")

	tmp.pl.Close()
}

func (pls *GRPCPools) get_index_ex(index uint32) *stpool {

	var pl *stpool

	pls.mtPool.RLock()

	length := len(pls.pools)
	if 0 >= length {
		pls.mtPool.RUnlock()
		return nil
	}

	pl = pls.pools[index%uint32(length)]

	if !pl.pl.IsConnected() { // 该连接池没有可用连接，就遍历获取一个有连接的连接池
		for i := 0; i < length; i++ {

			tmpIndex := (index + uint32(i)) % uint32(length)
			tmp := pls.pools[tmpIndex]

			if tmp.pl.IsConnected() {
				pl = tmp
				pls.index = tmpIndex
				break
			}
		}
	}

	pls.mtPool.RUnlock()

	return pl
}

func (pls *GRPCPools) get_index(i int) *stpool {

	pls.mtPool.RLock()

	if i < len(pls.pools) {
		pls.mtPool.RUnlock()
		return pls.pools[i]
	}

	pls.mtPool.RUnlock()

	return nil
}

func (pls *GRPCPools) get_server_index(svrAddr string) int {

	index := -1

	pls.mtPool.RLock()

	for i, pl := range pls.pools {
		if svrAddr == pl.pl.addr {
			index = i
			break
		}
	}

	pls.mtPool.RUnlock()

	return index
}

func (pls *GRPCPools) loop_check_idle_time_out_pool() {

	timer := time.NewTicker(time.Minute)
	defer timer.Stop()

	for {

		select {
		case <-pls.chExit:
			return
		case <-timer.C:
		}

		if pls.idleTimeOut < time.Minute {
			pls.idleTimeOut = time.Minute
		}

		pls.remove_idle_time_out()
	}
}

func (pls *GRPCPools) remove_idle_time_out() {

	i := 0

	for {
		pl := pls.get_index(i)
		if nil == pl {
			return
		}

		// log.Infow("GRPCPools remove_idle_time_out pool")

		if !pl.static && time.Since(pl.last) > pls.idleTimeOut && atomic.LoadInt64(&pl.call) == 0 {
			pls.remove_index(i)
			continue
		}

		i++
	}
}
