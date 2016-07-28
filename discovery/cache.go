package discovery

import (
	"net/url"

	"gitlab.vailsys.com/vail-cloud-services/platform"
)

type cache struct {
	req  chan []*url.URL
	cnt  chan int
	quit chan struct{}
}

func newCache(p Publisher) *cache {
	c := &cache{
		req:  make(chan []*url.URL),
		cnt:  make(chan int),
		quit: make(chan struct{}),
	}
	go c.loop(p)
	return c
}

func (c *cache) loop(p Publisher) {
	u := make(chan []*url.URL, 1)
	p.Subscribe(u)
	defer p.Unsubscribe(u)

	platform.Logger.Debugf("cache fetching urls")
	urls := <-u
	platform.Logger.Debugf("cache received urls: %s", urls)

	for {
		select {
		case urls = <-u:
		case c.cnt <- len(urls):
		case c.req <- urls:
		case <-c.quit:
			p.Stop()
			return
		}
	}
}

func (c *cache) count() int {
	return <-c.cnt
}

func (c *cache) get() []*url.URL {
	platform.Logger.Debugf("fetching a url from discovery")
	return <-c.req
}

func (c *cache) stop() {
	close(c.quit)
}
