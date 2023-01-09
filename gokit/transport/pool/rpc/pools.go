package rpc

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"net/rpc"
	"sync"
	"sync/atomic"
	"time"

	ut "github.com/jkprj/jkfr/gokit/transport"
	jkpool "github.com/jkprj/jkfr/gokit/transport/pool"
	urand "github.com/jkprj/jkfr/gokit/utils/rand"
	"github.com/jkprj/jkfr/log"
)

type GetPoolFunc func() *stpool

type stpool struct {
	pl     *RpcPool
	last   time.Time
	call   int64
	static bool // 静态创建的pool不能超时关闭，只有动态创建的pool才嫩超时关闭
}

type RpcPools struct {
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
}

func NewDefaultRpcPools() *RpcPools {
	pls, _ := NewDefaultRpcPoolsWithAddr([]string{})
	return pls
}

func NewDefaultRpcPoolsWithAddr(addrs []string) (*RpcPools, error) {

	opt := jkpool.NewOptions()
	opt.Factory = DefaultTcpClientFatory()

	return NewRpcPools(addrs, opt)
}

func NewDefaultTlsRpcPools(clientpem, clientkey []byte) *RpcPools {
	pls, _ := NewDefaultTlsRpcPoolsWithAddr([]string{}, clientpem, clientkey)
	return pls
}

func NewDefaultTlsRpcPoolsWithAddr(addrs []string, clientpem, clientkey []byte) (*RpcPools, error) {

	opt := jkpool.NewOptions()
	opt.Factory = DefaultTLSClientFatory(clientpem, clientkey)

	return NewRpcPools(addrs, opt)
}

func NewRpcPools(addrs []string, opt *jkpool.Options) (*RpcPools, error) {
	p := new(RpcPools)
	p.addr2pool = make(map[string]*stpool)
	p.pools = make([]*stpool, 0, len(addrs))
	p.opt = opt
	p.SetRetryTimes(3)
	p.SetIdleTimeOut(24 * 60 * 60)
	p.SetStrategy(ut.STRATEGY_LEAST)
	p.SetRetryIntervalMS(1000)

	for _, addr := range addrs {
		_, err := p.get_and_push(addr, opt, true)
		if nil != err {
			log.Infow("RpcPools.get_and_push err, to close pool", "err", err.Error())
			p.Close()
			return nil, err
		}
	}

	go p.loop_check_idle_time_out_pool()

	return p, nil
}

func (pls *RpcPools) RetryTimes() uint {
	return pls.retryTimes
}

func (pls *RpcPools) call(ctx context.Context, serviceMethod string, args interface{}, reply interface{}, retry uint) (err error) {

	var pl *stpool

	// retry为0时根据指定策略获取，Retry大于0时获取下一个服务连接池
	if 0 == retry {
		pl = pls.get_pool()
	} else {
		pl = pls.roll_get()
	}

	if nil == pl {
		log.Errorw("not found server")
		return errors.New("not found server")
	}

	atomic.AddInt64(&pl.call, 1)

	err = pl.pl.CallWithContext(ctx, serviceMethod, args, reply)
	if nil != err {
		// log.Errorw("Call fail", "method", serviceMethod, "error", err)
	}

	atomic.AddInt64(&pl.call, -1)
	pl.last = time.Now()

	return err
}

func (pls *RpcPools) call_with_func(callFunc func(retry uint) error) (err error) {

	var retry uint = 0

	var lastErr error

	for {

		if pls.isClosed {
			return jkpool.ErrClosed
		}

		err = callFunc(retry)
		if nil == err {
			return nil
		}

		retry++

		// log.Errorw("Call fail", "retryTimes", retry, "error", err)

		if nil != lastErr {
			lastErr = fmt.Errorf("%s; retryTimes_%d_err:%s", lastErr.Error(), retry, err.Error())
		} else {
			lastErr = fmt.Errorf("retryTimes_%d_err:%s", retry, err.Error())
		}

		if pls.retryTimes <= retry {
			break
		}

		time.Sleep(pls.retryInterval)
	}

	return lastErr
}

func (pls *RpcPools) CallWithTimeOut(serviceMethod string, args interface{}, reply interface{}, timeout time.Duration) (err error) {

	return pls.call_with_func(func(retry uint) error {

		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		err = pls.call(ctx, serviceMethod, args, reply, retry)

		cancel()

		return err
	})
}

func (pls *RpcPools) CallWithContext(ctx context.Context, serviceMethod string, args interface{}, reply interface{}) (err error) {

	return pls.call_with_func(func(retry uint) error {
		return pls.call(ctx, serviceMethod, args, reply, retry)
	})
}

func (pls *RpcPools) Call(serviceMethod string, args interface{}, reply interface{}) (err error) {
	return pls.CallWithContext(nil, serviceMethod, args, reply)
}

func (pls *RpcPools) GoCallWithContext(ctx context.Context, serviceMethod string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {

	call := new(rpc.Call)
	call.ServiceMethod = serviceMethod
	call.Args = args
	call.Reply = reply
	call.Error = nil

	if done == nil {
		done = make(chan *rpc.Call, 10)
	}

	call.Done = done

	go func() {
		call.Error = pls.CallWithContext(ctx, serviceMethod, args, reply)
		if nil != call.Error {
			log.Errorw("RpcPools.call fail", "error", call.Error)
		}

		done <- call
	}()

	return call
}

func (pls *RpcPools) GoCall(serviceMethod string, args interface{}, reply interface{}, done chan *rpc.Call) *rpc.Call {
	return pls.GoCallWithContext(nil, serviceMethod, args, reply, done)
}

func (pls *RpcPools) CallWithAddr(addr string, serviceMethod string, args interface{}, reply interface{}) error {

	return pls.CallWithAddrEx(addr, serviceMethod, args, reply, 60*time.Second)
}

func (pls *RpcPools) CallWithAddrEx(addr string, serviceMethod string, args interface{}, reply interface{}, timeout time.Duration) error {

	return pls.call_with_func(func(retry uint) error {

		pl, err := pls.getex(addr)
		if nil != err {
			log.Errorw("getex client fail", "addr", addr, "method", serviceMethod, "error", err)
			return err
		}

		atomic.AddInt64(&pl.call, 1)

		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		err = pl.pl.CallWithContext(ctx, serviceMethod, args, reply)

		cancel()

		pl.last = time.Now()
		atomic.AddInt64(&pl.call, -1)

		return err
	})
}

func (pls *RpcPools) GetServerConn(addr string) (net.Conn, error) {
	pl, err := pls.getex(addr)
	if nil != err {
		return nil, err
	}

	return pl.pl.GetConn()
}

func (pls *RpcPools) Close() {
	pls.mtNew.Lock()
	defer pls.mtNew.Unlock()

	log.Infow("RpcPools.Close")

	pls.close()
}

func (pls *RpcPools) close() {

	pls.mtPool.Lock()
	defer pls.mtPool.Unlock()

	pls.isClosed = true

	log.Infow("RpcPools.close")

	for _, pl := range pls.pools {
		pl.pl.Close()
	}

	pls.addr2pool = make(map[string]*stpool)
	pls.pools = make([]*stpool, 0)
}

func (pls *RpcPools) CloseServer(svrAddr string) {
	pls.remove_index(pls.get_server_index(svrAddr))
}

func (pls *RpcPools) SetRetryTimes(times uint) {

	const MAX_TIMES = 10

	if times > MAX_TIMES {
		pls.retryTimes = MAX_TIMES
	} else if times <= 0 {
		pls.retryTimes = 1
	} else {
		pls.retryTimes = times
	}
}

func (pls *RpcPools) SetRetryIntervalMS(intervalMS uint) {

	const MIN_INTERVAL_MS = 50

	if intervalMS < MIN_INTERVAL_MS {
		intervalMS = MIN_INTERVAL_MS
	}

	pls.retryInterval = time.Duration(intervalMS) * time.Millisecond
}

func (pls *RpcPools) SetIdleTimeOut(timeOut uint) {

	pls.idleTimeOut = time.Duration(timeOut) * time.Second

	if pls.idleTimeOut < time.Minute {
		pls.idleTimeOut = time.Minute
	}
}

func (pls *RpcPools) SetStrategy(strategy string) {

	pls.strategy = strategy

	if ut.STRATEGY_RANDOM == strategy {
		// pls.random = rand.New(rand.NewSource(time.Now().UnixNano()))      // golang提供的source不是线程安全的
		pls.random = rand.New(urand.NewSource(time.Now().UnixNano()))
		pls.get_pool = func() *stpool {
			return pls.random_get()
		}

	} else if ut.STRATEGY_LEAST == strategy {
		pls.strategy = ut.STRATEGY_ROUND
		pls.get_pool = func() *stpool {
			return pls.roll_get()
		}
	} else {
		pls.get_pool = func() *stpool {
			return pls.least_get()
		}
	}
}

func (pls *RpcPools) get_and_push(addr string, opt *jkpool.Options, static bool) (*stpool, error) {
	pl := pls.get(addr)
	if nil != pl {
		return pl, nil
	}

	pl, err := pls.new_and_push(addr, opt, static)
	if nil != err {
		log.Errorw("new_and_push fail", "addr", addr, "opt", opt, "error", err)
		return nil, err
	}

	return pl, nil
}

func (pls *RpcPools) push(addr string, pl *stpool) *stpool {

	pls.mtPool.Lock()
	defer pls.mtPool.Unlock()

	tmp, ok := pls.addr2pool[addr]
	if ok {
		log.Infow("RpcPools push poll, to close old pool")
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

func (pls *RpcPools) new_and_push(addr string, opt *jkpool.Options, static bool) (*stpool, error) {

	pls.mtNew.Lock()
	defer pls.mtNew.Unlock()

	if pls.isClosed {
		return nil, errors.New("Pools is closed")
	}

	pl := pls.get(addr)
	if nil != pl {
		return pl, nil
	}

	tmp_opt := *opt
	tmp_opt.ServerAddr = addr

	tmp, err := NewRpcPool(&tmp_opt)
	if nil != err {
		log.Errorw("NewRpcPool fail", "addr", addr, "opt", opt, "error", err)
		return nil, err
	}

	tmpPL := &stpool{pl: tmp, last: time.Now(), static: static}

	return pls.push(addr, tmpPL), nil
}

func (pls *RpcPools) get(addr string) *stpool {

	pls.mtPool.RLock()

	p, ok := pls.addr2pool[addr]
	if ok {
		pls.mtPool.RUnlock()
		return p
	}

	pls.mtPool.RUnlock()

	return nil
}

func (pls *RpcPools) roll_get() *stpool {

	pls.index++

	return pls.get_index_ex(pls.index)
}

func (pls *RpcPools) random_get() *stpool {

	pls.index = pls.random.Uint32()

	return pls.get_index_ex(pls.index)
}

func (pls *RpcPools) least_get() *stpool {

	var min int64 = 1000000000000
	var index uint32 = 0

	pls.mtPool.RLock()

	for i, pl := range pls.pools {

		index = uint32(i)

		if min > pl.call {
			min = pl.call
		}
		if 0 == min {
			break
		}
	}

	pls.index = uint32(index)

	pls.mtPool.RUnlock()

	return pls.get_index_ex(pls.index)
}

func (pls *RpcPools) getex(addr string) (*stpool, error) {
	return pls.get_and_push(addr, pls.opt, false)
}

func (pls *RpcPools) remove_index(i int) {

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

	log.Infow("RpcPools.remove_pool to close pool")

	tmp.pl.Close()
}

func (pls *RpcPools) get_index_ex(index uint32) *stpool {

	var pl *stpool

	pls.mtPool.RLock()

	length := len(pls.pools)
	if 0 >= length {
		return nil
	}

	pl = pls.pools[index%uint32(length)]
	pls.index = index % uint32(length)

	if false == pl.pl.IsConnected() { // 该连接池没有可用连接，就遍历获取一个有连接的连接池
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

func (pls *RpcPools) get_index(i int) *stpool {

	pls.mtPool.RLock()

	if i < len(pls.pools) {
		return pls.pools[i]
	}

	pls.mtPool.RUnlock()

	return nil
}

func (pls *RpcPools) get_server_index(svrAddr string) int {

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

func (pls *RpcPools) loop_check_idle_time_out_pool() {

	for {

		time.Sleep(time.Minute)

		if false == pls.isClosed {
			break
		}

		if pls.idleTimeOut < time.Minute {
			pls.idleTimeOut = time.Minute
		}

		pls.remove_idle_time_out()
	}
}

func (pls *RpcPools) remove_idle_time_out() {

	i := 0

	for {
		pl := pls.get_index(i)
		if nil == pl {
			return
		}

		// log.Infow("RpcPools remove_idle_time_out pool")

		if !pl.static && time.Now().Sub(pl.last) > pls.idleTimeOut && atomic.LoadInt64(&pl.call) == 0 {
			pls.remove_index(i)
			continue
		}

		i++
	}
}
