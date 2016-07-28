package static

import (
	"net/url"
	"sync"
)

type StaticPublisher struct {
	mu          sync.Mutex
	current     []*url.URL
	subscribers map[chan<- []*url.URL]struct{}
}

func NewStaticPublisher(urls []*url.URL) *StaticPublisher {
	return &StaticPublisher{
		current:     urls,
		subscribers: map[chan<- []*url.URL]struct{}{},
	}
}

func (p *StaticPublisher) Subscribe(c chan<- []*url.URL) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.subscribers[c] = struct{}{}
	c <- p.current
}

func (p *StaticPublisher) Unsubscribe(c chan<- []*url.URL) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.subscribers, c)
}

func (p *StaticPublisher) Stop() {}

func (p *StaticPublisher) Replace(endpoints []*url.URL) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.current = endpoints
	for c := range p.subscribers {
		c <- p.current
	}
}
