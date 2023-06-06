package pool

import (
	"container/list"
	"math"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

const (
	BAD  = false
	GOOD = true
)

const (
	TRUE  = 1
	FALSE = 0
)

var tag uint64

func min(num1, num2 int) int {

	if num1 <= num2 {
		return num1
	}

	return num2
}

type client struct {
	Client PoolClient
	o      *Options
	conn   net.Conn

	connTM time.Time
	reqTM  time.Time

	tag      uint64
	ref      int64
	index    int
	bDestroy bool

	mt sync.RWMutex

	err error
}

func new_client(o *Options, index int) (c *client) {

	c = new(client)
	c.tag = atomic.AddUint64(&tag, 1)
	c.o = o
	c.index = index

	return c
}

func (c *client) IsClose() (isClose bool) {

	c.mt.RLock()
	isClose = nil == c.conn
	c.mt.RUnlock()

	return isClose
}

func (c *client) IsIdleTimeOut() bool {
	return time.Now().Sub(c.reqTM) > c.o.IdleTimeout
}

func (c *client) Connect() (bNewConn bool, err error) {

	c.mt.RLock()
	if c.bDestroy {
		c.mt.RUnlock()
		return false, ErrClosed
	}
	if nil != c.conn {
		c.mt.RUnlock()
		return false, nil
	}
	c.mt.RUnlock()

	c.mt.Lock()
	defer c.mt.Unlock()

	if nil != c.conn {
		return false, nil
	}

	if c.bDestroy {
		return false, ErrClosed
	}

	if time.Now().Sub(c.connTM) < time.Second { // 每秒只能发起一次连接请求
		return false, c.err
	}

	c.connTM = time.Now()

	c.Client, c.conn, c.err = c.o.Factory(c.o)
	if nil != c.err {
		return false, c.err
	}

	c.reqTM = time.Now()

	return true, nil
}
func (c *client) AddRef(delta int64) {
	atomic.AddInt64(&c.ref, delta)
}

func (c *client) Ref() int64 {
	return atomic.LoadInt64(&c.ref)
}

func (c *client) Destroy() (err error) {

	c.mt.Lock()
	c.bDestroy = true
	c.mt.Unlock()

	return c.Close()
}

func (c *client) Close() (err error) {

	c.mt.Lock()

	if nil != c.conn {
		err = c.Client.Close()
		c.conn.Close()
	}

	c.conn = nil

	c.mt.Unlock()

	return err
}

// 待回收列表——延迟异步回收client，避免close阻塞影响请求
type recycle struct {
	liClients  *list.List
	mapClients map[uint64]*list.Element
	mtClients  sync.RWMutex
	chClient   chan int
	chExit     chan int
	isClose    bool

	o *Options
}

func new_recycle(o *Options) *recycle {
	r := new(recycle)
	r.liClients = list.New()
	r.mapClients = make(map[uint64]*list.Element)
	r.chClient = make(chan int, min(o.MaxCap*10, 2000))
	r.chExit = make(chan int)
	r.o = o

	go r.loop_recycle_clients()

	return r
}

func (r *recycle) close() {

	r.isClose = true

	r.chExit <- 1

	r.mtClients.Lock()

	for _, emClient := range r.mapClients {
		c := emClient.Value.(*client)
		c.Close()
	}

	r.mtClients.Unlock()
}

func (r *recycle) find(tag uint64) (c *client, ok bool) {

	r.mtClients.RLock()
	em, ok := r.mapClients[tag]
	if ok {
		c = em.Value.(*client)
	}

	r.mtClients.RUnlock()

	return c, ok
}

func (r *recycle) push(c *client) {

	if nil == c || nil == c.Client {
		return
	}

	r.mtClients.Lock()

	if r.isClose {
		c.Close()
	}

	emClient, ok := r.mapClients[c.tag]
	if !ok {

		emClient = r.liClients.PushBack(c)
		r.mapClients[c.tag] = emClient

		select {
		case r.chClient <- 1:
		default:
			// 回收队列已满就先删除最先入队的
			emClient = r.liClients.Front()
			if nil != emClient {
				cc := emClient.Value.(*client)
				r.liClients.Remove(emClient)
				delete(r.mapClients, cc.tag)
				cc.Close()
			}
		}
	}

	r.mtClients.Unlock()
}

func (r *recycle) remove(c *client) {

	if nil == c || nil == c.Client {
		return
	}

	r.mtClients.Lock()

	emClient, ok := r.mapClients[c.tag]
	if ok {
		r.liClients.Remove(emClient)
		delete(r.mapClients, c.tag)
	}

	r.mtClients.Unlock()
}

func (r *recycle) loop_recycle_clients() {

	timer := time.NewTicker(time.Second)
	defer timer.Stop()

	for {
		select {
		case <-r.chExit:
			return
		case <-r.chClient:
			r.mtClients.RLock()
			if r.liClients.Len() <= r.o.MaxCap {
				r.mtClients.RUnlock()
				continue
			}
			r.mtClients.RUnlock()
		case <-timer.C:
		}

		for {
			c := r.get_one_client()
			if nil != c {
				r.remove(c)
				c.Close()
			} else {
				break
			}
		}

	}
}

func (r *recycle) get_one_client() *client {

	r.mtClients.RLock()
	defer r.mtClients.RUnlock()

	if 0 == r.liClients.Len() {
		return nil
	}

	for _, em := range r.mapClients {
		c := em.Value.(*client)
		if 0 >= c.Ref() ||
			(0 < c.Ref() && time.Now().Sub(c.reqTM) > (r.o.ReadTimeout+r.o.WriteTimeout)) { // 在ref大于0时，判断是否读写超时，超时则关闭
			return c
		}
	}

	if r.liClients.Len() > r.o.MaxCap {
		emClient := r.liClients.Front()
		return emClient.Value.(*client)
	}

	return nil
}

type clients struct {
	cs        []*client
	pc2c      map[uint64]*client
	mtClients sync.RWMutex

	index      uint64
	initting   int32
	isClose    bool
	chIdleExit chan int

	recycle *recycle

	o *Options
}

func new_clients(o *Options) (cs *clients) {
	cs = new(clients)
	cs.pc2c = make(map[uint64]*client)
	cs.chIdleExit = make(chan int)
	cs.recycle = new_recycle(o)
	cs.o = o

	cs.cs = make([]*client, cs.o.MaxCap)
	for i := 0; i < len(cs.cs); i++ {
		cs.cs[i] = new_client(o, i)
	}

	go cs.loop_remove_idle_time_out_client()

	return cs
}

func (cs *clients) init_clients() (err error) {

	if !atomic.CompareAndSwapInt32(&cs.initting, 0, 1) {
		return nil
	}

	var c *client

	for i := 0; i < cs.o.InitCap; i++ {

		bNewConn, err := func() (bool, error) {

			cs.mtClients.RLock()
			defer cs.mtClients.RUnlock()

			c = cs.cs[i]

			if !c.IsClose() {
				return false, nil
			}

			return c.Connect()
		}()

		if nil != err {
			return err
		}

		if bNewConn {
			cs.push(c)
		}
	}

	atomic.StoreInt32(&cs.initting, 0)

	return err
}

func (cs *clients) get() (c *client, err error) {

	if 100 >= cs.o.MaxCap {
		c = cs.min_get()
	} else {
		c = cs.round_get()
	}

	bNewConn, err := c.Connect()
	if nil != err {
		c := cs.get_one_valid_client() // connect失败就尝试从已有连接中取一个连接
		if nil == c {
			return nil, err
		}
	}

	if bNewConn {
		cs.push(c)
	}

	return c, nil
}
func (cs *clients) min_get() (c *client) {

	cs.mtClients.RLock()

	// 优先获取initcap区域内的client
	for i := 0; i < cs.o.InitCap; i++ {
		if 0 >= cs.cs[i].Ref() {
			c = cs.cs[i]
			break
		}
	}

	if nil == c {
		var minRef int64 = math.MaxInt64

		for i := 0; i < cs.o.MaxCap; i++ {
			if minRef > cs.cs[i].Ref() {
				minRef = cs.cs[i].Ref()
				c = cs.cs[i]
			}
		}
	}

	cs.mtClients.RUnlock()

	return c
}

func (cs *clients) round_get() (c *client) {

	cs.mtClients.RLock()

	// 优先获取initcap区域内的client
	for i := 0; i < cs.o.InitCap; i++ {
		if 0 >= cs.cs[i].Ref() {
			c = cs.cs[i]
			break
		}
	}

	if nil == c {
		index := cs.index % uint64(len(cs.cs))
		c = cs.cs[index]
		cs.index++
	}

	cs.mtClients.RUnlock()

	return c
}

func (cs *clients) get_one_valid_client() (c *client) {

	cs.mtClients.RLock()

	for _, c = range cs.pc2c { // 如果connect失败就从已有正常连接池中取一个
		break
	}

	cs.mtClients.RUnlock()

	return c
}

func (cs *clients) push(c *client) {

	cs.mtClients.Lock()

	if cs.isClose {
		c.Close()
	} else {
		cs.pc2c[c.tag] = c
	}

	cs.mtClients.Unlock()
}

func (cs *clients) remove(tag uint64) {

	cs.mtClients.Lock()
	c, ok := cs.pc2c[tag]
	if ok {
		delete(cs.pc2c, tag)
		cs.cs[c.index] = new_client(cs.o, c.index)
	}
	cs.mtClients.Unlock()
}

func (cs *clients) find(tag uint64) (c *client, ok bool) {

	cs.mtClients.RLock()
	c, ok = cs.pc2c[tag]
	if !ok && nil != cs.recycle {
		c, ok = cs.recycle.find(tag)
	}
	cs.mtClients.RUnlock()

	return c, ok
}

func (cs *clients) move_recycle(tag uint64) (c *client) {

	cs.mtClients.Lock()

	c, ok := cs.pc2c[tag]
	if ok {

		delete(cs.pc2c, tag)

		if nil == cs.recycle {
			c.Close()
		} else {
			cs.cs[c.index] = new_client(cs.o, c.index)
			cs.recycle.push(c)
		}

	} else if nil != cs.recycle {
		c, _ = cs.recycle.find(tag)
	}

	cs.mtClients.Unlock()

	return c
}

func (cs *clients) loop_remove_idle_time_out_client() {

	timer := time.NewTicker(time.Second)
	defer timer.Stop()

	pre := time.Now()

	for {
		select {
		case <-cs.chIdleExit:
			return
		case <-timer.C:
		}

		go cs.init_clients()

		if time.Now().Sub(pre) < time.Minute {
			continue
		}

		pre = time.Now()

		cs.clear_idle_time_out_client()
	}
}

func (cs *clients) clear_idle_time_out_client() {

	for i := cs.o.InitCap; i < cs.o.MaxCap; i++ {

		if cs.valid_count() <= cs.o.InitCap { // 保留最少连接数
			return
		}

		cs.remove_client_if_idle_time_out(i)
	}
}

func (cs *clients) remove_client_if_idle_time_out(index int) {

	cs.mtClients.Lock()
	defer cs.mtClients.Unlock()

	c := cs.cs[index]

	if c.IsClose() ||
		(0 < c.Ref() && time.Now().Sub(c.reqTM) <= (cs.o.ReadTimeout+cs.o.WriteTimeout)) || // 在ref大于0时，判断是否读写超时，超时则关闭
		!c.IsIdleTimeOut() {
		return
	}

	if nil != cs.recycle {
		cs.cs[index] = new_client(cs.o, index)
		delete(cs.pc2c, c.tag)

		cs.recycle.push(c) // 放到待回收列表，延迟close
	} else {
		c.Close()
	}

}

func (cs *clients) valid_count() (count int) {

	cs.mtClients.RLock()
	count = len(cs.pc2c)
	cs.mtClients.RUnlock()

	return count
}

func (cs *clients) close() (err error) {

	cs.isClose = true

	cs.chIdleExit <- 1

	cs.mtClients.Lock()

	for _, c := range cs.cs {
		err = c.Destroy()
		if nil != err {
			break
		}
	}
	cs.pc2c = make(map[uint64]*client)

	if nil != cs.recycle {
		cs.recycle.close()
	}
	cs.recycle = nil

	cs.mtClients.Unlock()

	return err
}

type Pool struct {
	clients *clients

	o *Options

	mtClose sync.RWMutex
	isClose int32
}

func NewPool(o *Options) (pool *Pool, err error) {

	if o.InitCap < 1 {
		o.InitCap = 1
	}

	if o.MaxCap < o.InitCap {
		o.MaxCap = int(math.Max(math.Min(float64(o.InitCap*10), 32), float64(o.InitCap)))
	}

	if err = o.validate(); err != nil {
		return nil, err
	}

	pool = &Pool{o: o}

	pool.clients = new_clients(o)
	err = pool.clients.init_clients()
	if nil != err {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

func (pl *Pool) Get() (c *client, err error) {

	if !pl.mtClose.TryRLock() {
		return nil, ErrClosed
	}

	c, err = pl.clients.get()
	if nil != err {
		pl.mtClose.RUnlock()
		return nil, err
	}

	c.AddRef(1)
	c.reqTM = time.Now()

	pl.mtClose.RUnlock()

	return c, nil
}

func (pl *Pool) GetConn() (conn net.Conn, err error) {

	if !pl.mtClose.TryRLock() {
		return nil, ErrClosed
	}
	defer pl.mtClose.RUnlock()

	c := pl.clients.get_one_valid_client()
	if nil == c {
		return nil, ErrNotFound
	}

	return c.conn, nil
}

func (pl *Pool) Put(c *client, good bool) (err error) {

	if nil == c {
		return nil
	}

	if !pl.mtClose.TryRLock() {
		c.Close()
		return ErrClosed
	}

	c.AddRef(-1)

	if good {
		err = pl.put_good(c)
	} else {
		err = pl.put_bad(c)
	}

	pl.mtClose.RUnlock()

	return err
}

func (pl *Pool) put_good(c *client) error {
	return nil
}

func (pl *Pool) put_bad(c *client) error {
	if nil == pl.clients.move_recycle(c.tag) { // 缓存没找到就直接close
		c.Close()
	}
	return nil
}

func (pl *Pool) Close() error {

	if !atomic.CompareAndSwapInt32(&pl.isClose, 0, 1) { // 避免重入
		return nil
	}

	pl.mtClose.Lock()

	if nil != pl.clients {
		pl.clients.close()
	}

	return nil
}

func (pl *Pool) ValidCount() int {

	if !pl.mtClose.TryRLock() { // closed
		return 0
	}
	pl.mtClose.RUnlock()

	return pl.clients.valid_count()
}
