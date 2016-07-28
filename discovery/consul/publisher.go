package consul

import (
	"fmt"
	"net/url"
	"time"

	consul_api "github.com/hashicorp/consul/api"
	"gitlab.vailsys.com/vail-cloud-services/platform"
	"gitlab.vailsys.com/vail-cloud-services/platform/registry"
)

type ConsulPublisher struct {
	subscribe     chan chan<- []*url.URL
	unsubscribe   chan chan<- []*url.URL
	quit          chan struct{}
	ErrChan       chan error
	consulAdapter registry.RegistryAdapter
}

func NewConsulPublisher(consul registry.RegistryAdapter, name string, ttl time.Duration) *ConsulPublisher {
	if name == "" {
		panic("name cannot be nil")
	}
	if ttl.String() == "0" {
		panic("ttl cannot be 0")
	}

	p := &ConsulPublisher{
		subscribe:     make(chan chan<- []*url.URL),
		unsubscribe:   make(chan chan<- []*url.URL),
		quit:          make(chan struct{}),
		ErrChan:       make(chan error),
		consulAdapter: consul,
	}

	go p.loop(name, ttl)
	return p
}

func (p *ConsulPublisher) Subscribe(c chan<- []*url.URL) {
	p.subscribe <- c
}

func (p *ConsulPublisher) Unsubscribe(c chan<- []*url.URL) {
	p.unsubscribe <- c
}

func (p *ConsulPublisher) Stop() {
	platform.Logger.Debugf("stopping consul publisher")
	close(p.quit)
}

var newTicker = time.NewTicker

func (p *ConsulPublisher) loop(name string, ttl time.Duration) {
	platform.Logger.Debugf("consul publisher %s with ttl %v", name, ttl)

	subscriptions := map[chan<- []*url.URL]struct{}{}
	service, err := p.fetch(name)

	var urls []*url.URL

	if err == nil {
		urls = format(service)
	} else {
		urls = make([]*url.URL, 0)
	}

	platform.Logger.Debugf("found urls: %s", urls)

	ticker := newTicker(ttl)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			platform.Logger.Debugf("discovery check ticked")

			service, err := p.fetch(name)
			if err != nil {
				//p.ErrChan <- err
				urls = make([]*url.URL, 0)
			} else {
				urls = format(service)
			}
			platform.Logger.Debugf("broadcasting urls: %s", urls)
			for c := range subscriptions {
				c <- urls
			}
		case c := <-p.subscribe:
			subscriptions[c] = struct{}{}
			platform.Logger.Debugf("sending urls to subscription: %s", urls)
			c <- urls
		case c := <-p.unsubscribe:
			delete(subscriptions, c)
		case err := <-p.ErrChan:
			platform.Logger.Debugf("received error on chan: %s", err)
		case <-p.quit:
			return
		}
	}
}

func (p *ConsulPublisher) fetch(name string) ([]*consul_api.ServiceEntry, error) {
	err := p.consulAdapter.Ping()
	if err != nil {
		return nil, err
	}
	serv, err := p.consulAdapter.CheckService(name, "", true)

	if err != nil {
		platform.Logger.Debugf("error retreiving service: %s ", name)
		return nil, err
	}

	return serv, nil
}

func format(serviceEntry []*consul_api.ServiceEntry) []*url.URL {
	urls := make([]*url.URL, 0)

	if len(serviceEntry) == 0 {
		return urls
	}

	for _, service := range serviceEntry {
		url := &url.URL{
			Scheme: "http",
			Host:   fmt.Sprintf("%s:%d", service.Service.Address, service.Service.Port),
		}
		urls = append(urls, url)
	}
	return urls
}
