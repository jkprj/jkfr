package lb

import (
	"context"
	"math/rand"
	"reflect"
	"sync"
	"sync/atomic"
	"time"

	urand "github.com/jkprj/jkfr/gokit/utils/rand"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	glb "github.com/go-kit/kit/sd/lb"
)

type epCall struct {
	ep   endpoint.Endpoint
	call int64
	last time.Time
}

func NewLeastBalancer(s sd.Endpointer) glb.Balancer {

	l := least{s: s}
	l.epCalls = make(map[reflect.Value]*epCall)
	l.pre_check_time = time.Now()
	l.random = rand.New(urand.NewSource(time.Now().UnixNano()))

	return &l
}

type least struct {
	s              sd.Endpointer
	epCalls        map[reflect.Value]*epCall
	mt             sync.RWMutex
	pre_check_time time.Time
	eps            []endpoint.Endpoint
	random         *rand.Rand
}

func (l *least) try_reset_endpoints(eps []endpoint.Endpoint) {

	l.mt.RLock()

	if len(l.eps) == len(eps) && time.Now().Sub(l.pre_check_time) < time.Minute {
		l.mt.RUnlock()
		return
	}

	l.mt.RUnlock()

	l.mt.Lock()

	if len(l.eps) == len(eps) && time.Now().Sub(l.pre_check_time) < time.Minute {
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
	var tmpEpcs []*epCall
	var min int64 = 1000000000000

	l.mt.RLock()

	for _, ep := range eps {

		key := reflect.ValueOf(ep)
		epc, ok := l.epCalls[key]

		if !ok {
			tmpEpc = &epCall{ep: ep}
			l.epCalls[key] = tmpEpc
			break
		} else if min > epc.call {
			min = epc.call
			// tmpEpc = epc
			tmpEpcs = nil
			tmpEpcs = append(tmpEpcs, epc)
		} else if min == epc.call {
			tmpEpcs = append(tmpEpcs, epc)
		}
	}

	l.mt.RUnlock()

	if nil == tmpEpc && nil != tmpEpcs {

		length := len(tmpEpcs)

		if 1 == length {
			tmpEpc = tmpEpcs[0]
		} else if 1 < length {
			tmpEpc = tmpEpcs[l.random.Intn(length)]
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

	tmpEpc.last = time.Now()

	return func(ctx context.Context, request interface{}) (response interface{}, err error) {

		atomic.AddInt64(&tmpEpc.call, 1)

		response, err = tmpEpc.ep(ctx, request)

		atomic.AddInt64(&tmpEpc.call, -1)

		return response, err
	}, nil
}
