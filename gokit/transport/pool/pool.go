package pool

import (
	"container/list"
	"math"
	"net"
	"sync"
	"time"
)

const (
	BAD  = false
	GOOD = true
)

type poolClient struct {
	client PoolClient
	conn   net.Conn
	tm     time.Time
	ref    int32
}

type Pool struct {
	mapClients map[PoolClient]*poolClient

	liIdleClients  *list.List
	mapIdleClients map[PoolClient]*list.Element

	liIngClients  *list.List
	mapIngClients map[PoolClient]*list.Element

	liRecoverableClients  *list.List
	mapRecoverableClients map[PoolClient]*list.Element

	bClose      bool
	initing     bool
	valid_count int
	createing   int32

	mtPool sync.Mutex

	o *Options
}

func NewPool(o *Options) (*Pool, error) {

	if o.InitCap < 1 {
		o.InitCap = 1
	}

	if o.MaxCap < o.InitCap {
		o.MaxCap = int(math.Max(math.Min(float64(o.InitCap*10), 200), float64(o.InitCap)))
	}

	if err := o.validate(); err != nil {
		return nil, err
	}

	pool := &Pool{
		mapClients:            map[PoolClient]*poolClient{},
		liIdleClients:         list.New(),
		mapIdleClients:        map[PoolClient]*list.Element{},
		liIngClients:          list.New(),
		mapIngClients:         map[PoolClient]*list.Element{},
		liRecoverableClients:  list.New(),
		mapRecoverableClients: map[PoolClient]*list.Element{},
		o:                     o,
	}

	//init make conns
	err := pool.init_clients()
	if nil != err {
		pool.Close()
		return nil, err
	}

	go pool.loop_remove_idle_time_out_client()

	return pool, nil
}

func (pl *Pool) init_clients() error {

	for i := 0; i < pl.o.InitCap && pl.valid_count < pl.o.MaxCap; i++ {
		_, err := pl.create_client(false)
		if nil != err {
			return err
		}
	}

	return nil
}

func (pl *Pool) Get() (plc PoolClient, err error) {

	pl.mtPool.Lock()

	if pl.bClose {
		pl.mtPool.Unlock()
		return nil, ErrClosed
	}

	pc, err := pl.get()
	if nil != err {
		pl.mtPool.Unlock()
		return nil, err

	}

	pc.tm = time.Now()
	pc.ref++

	pl.mtPool.Unlock()

	return pc.client, err
}

func (pl *Pool) GetConn() (conn net.Conn, err error) {

	pl.mtPool.Lock()

	if pl.bClose {
		pl.mtPool.Unlock()
		return nil, ErrClosed
	}

	pc, err := pl.get()
	if nil != err {
		pl.mtPool.Unlock()
		return nil, err
	}

	pc.ref++
	pl.put_good(pc.client)

	pl.mtPool.Unlock()

	return pc.conn, nil
}

func (pl *Pool) get() (pc *poolClient, err error) {

	// 如果从idle中获取的就先不用创建，如果从ing获取到的而且vali-count没有超过max，就异步创建，先复用已有的client
	// 如果valid-count为0，就同步创建

	pc, bIdle := pl.get_from_created()
	if nil == pc {
		pc, err = pl.create_client(true)
		if nil != err {
			return nil, err
		}
	} else if false == bIdle {
		if int(pl.createing) < pl.o.MaxCap-pl.valid_count { //同一时间的创建协程不能超过max - valid
			pl.createing++
			go func() {
				pl.create_client(false)
				pl.createing--
			}()
		}

	}

	return pc, nil
}

func (pl *Pool) get_from_created() (pc *poolClient, bIdle bool) {

	pc = pl.pop_idle_client()
	if nil != pc {
		pl.push_ing_client(pc)
		return pc, true
	}

	pc = pl.pop_ing_client()
	if nil != pc {
		return pc, false
	}

	return nil, false
}

func (pl *Pool) create_client(ing bool) (*poolClient, error) {

	if pl.valid_count >= pl.o.MaxCap {
		return nil, ErrMax
	}

	plc, conn, err := pl.o.Factory(pl.o)
	if nil != err {
		return nil, err
	}

	if false == ing {
		pl.mtPool.Lock()
	}

	pc := &poolClient{client: plc, conn: conn, tm: time.Now()}
	pl.mapClients[pc.client] = pc
	pl.valid_count++

	if ing {
		pl.push_ing_client(pc)
	} else {
		pl.push_idle_client(pc)
		pl.mtPool.Unlock()
	}

	return pc, nil
}

func (pl *Pool) close_client(pc *poolClient) error {

	delete(pl.mapClients, pc.client)

	pc.conn = nil

	return pc.client.Close()
}

func (pl *Pool) push_client(liClients *list.List, mapClients map[PoolClient]*list.Element, pc *poolClient) {

	emClient, ok := mapClients[pc.client]
	if !ok {
		emClient = liClients.PushBack(pc)
		mapClients[pc.client] = emClient
	}
}

func (pl *Pool) remove_client(liClients *list.List, mapClients map[PoolClient]*list.Element, pcl PoolClient) *poolClient {

	emClient, ok := mapClients[pcl]
	if !ok {
		return nil
	}
	liClients.Remove(emClient)
	delete(mapClients, pcl)

	return emClient.Value.(*poolClient)
}

func (pl *Pool) pop_idle_client() *poolClient {

	tmp := pl.liIdleClients.Front()
	if nil != tmp {

		pl.liIdleClients.Remove(tmp)
		pc := tmp.Value.(*poolClient)
		delete(pl.mapIdleClients, pc.client)

		return pc
	}

	return nil
}

func (pl *Pool) push_idle_client(pc *poolClient) {
	pl.push_client(pl.liIdleClients, pl.mapIdleClients, pc)
}

func (pl *Pool) remove_idle_client(pcl PoolClient) *poolClient {
	return pl.remove_client(pl.liIdleClients, pl.mapIdleClients, pcl)
}

func (pl *Pool) pop_ing_client() *poolClient {

	tmp := pl.liIngClients.Front()
	if nil != tmp {
		pl.liIngClients.MoveToBack(tmp)
		return tmp.Value.(*poolClient)
	}

	return nil
}

func (pl *Pool) push_ing_client(pc *poolClient) {
	pl.push_client(pl.liIngClients, pl.mapIngClients, pc)
}

func (pl *Pool) remove_ing_client(pcl PoolClient) *poolClient {
	return pl.remove_client(pl.liIngClients, pl.mapIngClients, pcl)
}

func (pl *Pool) recoverable_clients_count() int {

	count := pl.liRecoverableClients.Len()

	return count
}

func (pl *Pool) push_recoverable_client(pc *poolClient) {

	pl.push_client(pl.liRecoverableClients, pl.mapRecoverableClients, pc)

	for pl.o.MaxCap <= pl.liRecoverableClients.Len() { // recoverables最多保留max-cap个

		emClient := pl.liRecoverableClients.Front()
		pl.liRecoverableClients.Remove(emClient)
		pc := emClient.Value.(*poolClient)
		delete(pl.mapRecoverableClients, pc.client)

		pl.close_client(pc)
	}
}

func (pl *Pool) remove_recoverable_client(pcl PoolClient) *poolClient {
	return pl.remove_client(pl.liRecoverableClients, pl.mapRecoverableClients, pcl)
}

func (pl *Pool) Put(plc PoolClient, good bool) (err error) {

	pl.mtPool.Lock()

	if pl.bClose {
		err = plc.Close()
		pl.mtPool.Unlock()
		return err
	}

	err = pl.put(plc, good)

	pl.mtPool.Unlock()

	return err
}

func (pl *Pool) put(plc PoolClient, good bool) error {
	if good {
		return pl.put_good(plc)
	} else {
		return pl.put_bad(plc)
	}
}

func (pl *Pool) put_bad(plc PoolClient) error {

	pc := pl.remove_ing_client(plc)
	if nil == pc { // 可能前面已经放到recoverable中，只是ref还不为0
		pc = pl.mapClients[plc]
	} else {
		pl.valid_count--
	}

	if nil != pc {

		pc.ref--

		if pc.ref > 0 {
			pl.push_recoverable_client(pc)
		} else {
			pl.close_client(pc)
		}
	}

	return nil
}

func (pl *Pool) put_good(plc PoolClient) error {

	pc := pl.remove_recoverable_client(plc)
	if nil == pc {
		pc = pl.mapClients[plc]
		if nil != pc {
			pc.ref--
		}
	} else {

		pc.ref--

		if pl.valid_count >= pl.o.MaxCap { // 如果valid-count大于max，需要回收掉
			if pc.ref > 0 {
				pl.push_recoverable_client(pc) // 还有引用，就先放回recoverable
			} else {
				return pl.close_client(pc)
			}

			return nil

		} else { //重新放回valid列表
			pl.valid_count++
		}
	}

	if nil != pc {

		pc.tm = time.Now()

		if pc.ref > 0 {
			pl.push_ing_client(pc)
		} else {
			pl.remove_ing_client(plc)
			pl.push_idle_client(pc)
		}
	}

	return nil
}

func (pl *Pool) Close() error {

	pl.mtPool.Lock()
	defer pl.mtPool.Unlock()

	pl.bClose = true
	pl.valid_count = 0

	for client := range pl.mapClients {
		client.Close()
	}

	pl.mapClients = nil
	pl.liIdleClients = nil
	pl.liIngClients = nil
	pl.mapIngClients = nil
	pl.liRecoverableClients = nil
	pl.mapRecoverableClients = nil

	return nil
}

func (pl *Pool) ValidCount() int {

	pl.mtPool.Lock()

	count := pl.valid_count

	if 0 >= count && false == pl.initing && false == pl.bClose {
		pl.initing = true
		go func() {
			pl.init_clients()
			pl.initing = false
		}()
	}

	pl.mtPool.Unlock()

	return count
}

func (pl *Pool) try_init_clients() {
	pl.ValidCount()
}

func (pl *Pool) loop_remove_idle_time_out_client() {

	pre := time.Now()

	for {
		time.Sleep(time.Second)
		if pl.bClose {
			break
		}

		pl.try_init_clients()

		now := time.Now()
		if now.Sub(pre) < time.Minute {
			continue
		}

		pre = now

		pl.mtPool.Lock()

		if pl.bClose {
			pl.mtPool.Unlock()
			break
		}

		pl.remove_idle_time_out_client(now)

		pl.mtPool.Unlock()
	}
}

func (pl *Pool) remove_idle_time_out_client(now time.Time) {

	if pl.valid_count <= pl.o.InitCap {
		return
	}

	for plc, pc := range pl.mapClients {

		if now.Sub(pc.tm) < pl.o.IdleTimeout {
			continue
		}

		if nil != pl.remove_idle_client(plc) || nil != pl.remove_ing_client(plc) {
			pl.valid_count--
		} else {
			pl.remove_recoverable_client(plc)
		}

		delete(pl.mapClients, plc)
		plc.Close()
	}
}
