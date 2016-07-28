package discovery

import (
	"net/url"
	"sync/atomic"
)

func RoundRobin(p Publisher) LoadBalancer {
	return &roundRobin{newCache(p), 0}
}

type roundRobin struct {
	*cache
	uint64
}

func (r *roundRobin) Count() int { return r.cache.count() }

func (r *roundRobin) Get() (*url.URL, error) {
	endpoints := r.cache.get()

	if len(endpoints) <= 0 {
		return nil, ErrNoEndpointsAvailable
	}
	var old uint64
	for {
		old = atomic.LoadUint64(&r.uint64)
		if atomic.CompareAndSwapUint64(&r.uint64, old, old+1) {
			break
		}
	}
	return endpoints[old%uint64(len(endpoints))], nil
}

func (r *roundRobin) Stop() {
	r.cache.stop()
}
