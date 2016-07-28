package discovery_test

import (
	"net/url"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	consul_api "github.com/hashicorp/consul/api"
	"gitlab.vailsys.com/vail-cloud-services/platform/discovery"
	"gitlab.vailsys.com/vail-cloud-services/platform/discovery/consul"
	"gitlab.vailsys.com/vail-cloud-services/platform/discovery/static"
	"gitlab.vailsys.com/vail-cloud-services/platform/registry/fakes"
)

var _ = Describe("round robin", func() {

	Context("consul publisher", func() {
		var fakeAdapter = new(fakes.FakeRegistryAdapter)
		//var service = "fakeservice"

		It("should provide an error when the service nis not avaible", func() {
			service := &consul_api.AgentService{
				Service: "service",
				Address: "127.0.0.1",
				Port:    3000,
			}
			entry := []*consul_api.ServiceEntry{}

			fakeAdapter.PingReturns(nil)
			fakeAdapter.CheckServiceReturns(entry, nil)

			publisher := consul.NewConsulPublisher(fakeAdapter, service.Service, 5*time.Second)
			defer publisher.Stop()

			rb := discovery.RoundRobin(publisher)

			Expect(rb.Count()).To(Equal(0))

			_, err := rb.Get()
			Expect(err).To(HaveOccurred())

		})

		It("should provide urls in a roundrobin fashion", func() {
			service := &consul_api.AgentService{
				Service: "service",
				Address: "127.0.0.1",
				Port:    3000,
			}
			service2 := &consul_api.AgentService{
				Service: "service",
				Address: "127.0.0.2",
				Port:    3000,
			}

			entry := []*consul_api.ServiceEntry{
				&consul_api.ServiceEntry{
					Service: service,
				},
				&consul_api.ServiceEntry{
					Service: service2,
				},
			}

			//nodes := []*registry.Node{
			//&registry.Node{&api.CatalogService{
			//ServiceAddress: "192.168.1.1",
			//Address:        "localhost",
			//ServicePort:    3001,
			//}},
			//&registry.Node{&api.CatalogService{
			//Address:        "localhost",
			//ServiceAddress: "192.168.1.2",
			//ServicePort:    3001,
			//}},
			//}

			//service := &registry.VailService{Name: service, Nodes: nodes}

			fakeAdapter.PingReturns(nil)
			fakeAdapter.CheckServiceReturns(entry, nil)

			publisher := consul.NewConsulPublisher(fakeAdapter, service.Service, 5*time.Second)
			defer publisher.Stop()

			rb := discovery.RoundRobin(publisher)

			Expect(rb.Count()).To(Equal(2))

			url, err := rb.Get()

			Expect(err).ToNot(HaveOccurred())
			Expect(url).ToNot(BeNil())
			Expect(url.Host).To(Equal("127.0.0.1:3000"))

			url, err = rb.Get()

			Expect(err).ToNot(HaveOccurred())
			Expect(url).ToNot(BeNil())
			Expect(url.Host).To(Equal("127.0.0.2:3000"))

		})
	})

	Context("static publisher", func() {
		It("should return back an endpoint", func() {
			endpoints := []*url.URL{
				&url.URL{Scheme: "http", Host: "127.0.0.1"},
				&url.URL{Scheme: "http", Host: "127.0.0.2"},
			}

			p := static.NewStaticPublisher(endpoints)
			defer p.Stop()

			lb := discovery.RoundRobin(p)
			defer lb.Stop()

			Expect(lb.Count()).To(Equal(2))

			url, err := lb.Get()

			Expect(err).ToNot(HaveOccurred())
			Expect(url).ToNot(BeNil())
			Expect(url.Host).To(Equal(endpoints[0].Host))

			url, err = lb.Get()

			Expect(err).ToNot(HaveOccurred())
			Expect(url).ToNot(BeNil())
			Expect(url.Host).To(Equal(endpoints[1].Host))

		})
	})
})
