// gokit的random线程不安全，这里自己实现个线程安全的random Balancer

package lb

import (
	"math/rand"

	urand "github.com/jkprj/jkfr/gokit/utils/rand"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/sd"
	glb "github.com/go-kit/kit/sd/lb"
)

func NewRandom(s sd.Endpointer, seed int64) glb.Balancer {
	return &random{
		s: s,
		r: rand.New(urand.NewSource(seed)),
	}
}

type random struct {
	s sd.Endpointer
	r *rand.Rand
}

func (r *random) Endpoint() (endpoint.Endpoint, error) {
	endpoints, err := r.s.Endpoints()
	if err != nil {
		return nil, err
	}
	if len(endpoints) <= 0 {
		return nil, glb.ErrNoEndpoints
	}
	return endpoints[r.r.Intn(len(endpoints))], nil
}
