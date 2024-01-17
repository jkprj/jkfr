package lb

import (
	"sync"

	"github.com/go-kit/kit/endpoint"

	"github.com/go-kit/kit/sd/lb"
)

type safeBalancer struct {
	bl lb.Balancer

	mt sync.Mutex
}

func (sb *safeBalancer) Endpoint() (endpoint.Endpoint, error) {
	sb.mt.Lock()

	ep, err := sb.bl.Endpoint()

	sb.mt.Unlock()

	return ep, err
}

func NewSafeBalancer(bl lb.Balancer) lb.Balancer {
	return &safeBalancer{bl: bl}
}
