package lb

import (
	"context"
	"math"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	jkutils "github.com/jkprj/jkfr/gokit/utils"
	jkrand "github.com/jkprj/jkfr/gokit/utils/rand"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	glb "github.com/go-kit/kit/sd/lb"
)

type epCall struct {
	ep   endpoint.Endpoint
	call int64
}

func NewLeastBalancer(s sd.Endpointer) glb.Balancer {

	l := least{s: s}
	l.epCalls = make(map[reflect.Value]*epCall)
	l.pre_check_time = time.Now()

	return &l
}

type least struct {
	s              sd.Endpointer
	epCalls        map[reflect.Value]*epCall
	mt             sync.RWMutex
	pre_check_time time.Time
	eps            []endpoint.Endpoint
}

func (l *least) try_reset_endpoints(eps []endpoint.Endpoint) {

	l.mt.RLock()

	if len(l.eps) == len(eps) && time.Since(l.pre_check_time) < time.Minute {
		l.mt.RUnlock()
		return
	}

	l.mt.RUnlock()

	l.mt.Lock()

	if len(l.eps) == len(eps) && time.Since(l.pre_check_time) < time.Minute {
		l.mt.Unlock()
		return
	}

	// fmt.Printf("eps_len:%d, map_aps_len:%d\n", len(eps), len(l.epCalls))

	epCalls := make(map[reflect.Value]*epCall)
	for _, ep := range eps {

		key := reflect.ValueOf(ep)
		epc, ok := l.epCalls[key]
		if !ok {
			epc = &epCall{ep: ep}
		}

		epCalls[key] = epc
	}

	l.epCalls = epCalls
	l.eps = eps
	l.pre_check_time = time.Now()

	l.mt.Unlock()
}

func (l *least) get_least(eps []endpoint.Endpoint) *epCall {

	var tmpEpc *epCall

	l.mt.RLock()

	if len(eps) <= jkutils.LEAST_ROUND_MAX {
		tmpEpc = l.get_least_when_less(eps)
	} else {
		tmpEpc = l.get_least_when_bigger(eps)
	}

	l.mt.RUnlock()

	return tmpEpc
}

func (l *least) get_least_when_less(eps []endpoint.Endpoint) *epCall {

	var tmpEpc *epCall
	var min int64 = math.MaxInt64
	nlen := len(eps)
	bgIndex := jkrand.Int()

	for i := 0; i < nlen; i++ {

		index := (bgIndex + i) % nlen
		key := reflect.ValueOf(eps[index])
		epc, ok := l.epCalls[key]

		if !ok {
			// 找不到说明是新实例，直接返回
			tmpEpc = &epCall{ep: eps[index]}
			l.epCalls[key] = tmpEpc
			return tmpEpc
		} else if epc.call == 0 {
			return epc
		} else if min > epc.call {
			min = epc.call
			tmpEpc = epc
		}
	}

	return tmpEpc
}

func (l *least) get_least_when_bigger(eps []endpoint.Endpoint) *epCall {

	var tmpEpc *epCall
	var min int64 = math.MaxInt64
	nlen := len(eps)

	// 随机抽取LEAST_RAND_COUNT个选最小一个
	for i := 0; i < jkutils.LEAST_RAND_COUNT; i++ {

		index := jkrand.Int() % nlen
		key := reflect.ValueOf(eps[index])
		epc, ok := l.epCalls[key]

		if !ok {
			// 找不到说明是新实例，直接返回
			tmpEpc = &epCall{ep: eps[index]}
			l.epCalls[key] = tmpEpc
			return tmpEpc
		} else if epc.call == 0 {
			return epc
		} else if min > epc.call {
			min = epc.call
			tmpEpc = epc
		}
	}

	return tmpEpc
}

func (l *least) Endpoint() (endpoint.Endpoint, error) {

	endpoints, err := l.s.Endpoints()
	if nil != err {
		return nil, err
	}

	if len(endpoints) <= 0 {
		return nil, glb.ErrNoEndpoints
	}

	l.try_reset_endpoints(endpoints)

	tmpEpc := l.get_least(endpoints)

	return func(ctx context.Context, request interface{}) (response interface{}, err error) {

		atomic.AddInt64(&tmpEpc.call, 1)

		response, err = tmpEpc.ep(ctx, request)

		atomic.AddInt64(&tmpEpc.call, -1)

		return response, err
	}, nil
}
