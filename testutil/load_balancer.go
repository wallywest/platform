package testutil

import (
	"net"
	"net/url"
	"strconv"
	"time"

	consul_api "github.com/hashicorp/consul/api"
	"gitlab.vailsys.com/vail-cloud-services/platform/discovery"
	"gitlab.vailsys.com/vail-cloud-services/platform/discovery/consul"
	"gitlab.vailsys.com/vail-cloud-services/platform/registry/fakes"
)

func NewMockRoundRobin(name string, ttl time.Duration, urls []*url.URL) discovery.LoadBalancer {
	entries := newMockService(name, urls)

	var fakeAdapter = new(fakes.FakeRegistryAdapter)
	fakeAdapter.PingReturns(nil)
	fakeAdapter.CheckServiceReturns(entries, nil)

	publisher := consul.NewConsulPublisher(fakeAdapter, name, ttl)
	rb := discovery.RoundRobin(publisher)
	return rb
}

func newMockService(name string, urls []*url.URL) []*consul_api.ServiceEntry {
	var entries []*consul_api.ServiceEntry

	for _, u := range urls {
		host, port, _ := net.SplitHostPort(u.Host)
		p, _ := strconv.Atoi(port)

		e := &consul_api.ServiceEntry{
			Service: &consul_api.AgentService{
				Service: name,
				Address: host,
				Port:    p,
			},
		}
		entries = append(entries, e)
	}

	return entries
}
