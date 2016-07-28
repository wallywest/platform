package consul_test

import (
	"fmt"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	consul_api "github.com/hashicorp/consul/api"
	"gitlab.vailsys.com/vail-cloud-services/platform/discovery/consul"
	"gitlab.vailsys.com/vail-cloud-services/platform/registry"
	"gitlab.vailsys.com/vail-cloud-services/platform/registry/fakes"
)

var _ = Describe("consul discovery publisher", func() {
	var fakeAdapter = new(fakes.FakeRegistryAdapter)

	It("should return empty urls when consul is down", func() {
		err := fmt.Errorf("consul is down")
		fakeAdapter.PingReturns(err)

		p := consul.NewConsulPublisher(fakeAdapter, "myservice", 5*time.Second)
		defer p.Stop()

		c := make(chan []*url.URL)
		p.Subscribe(c)
		defer p.Unsubscribe(c)

		var urls []*url.URL

		Eventually(c).Should(Receive(&urls))
		Expect(len(urls)).To(Equal(0))
	})

	It("should return empty urls if the service does not exist", func() {
		fakeAdapter.PingReturns(nil)
		err := fmt.Errorf("service myservice not found")
		fakeAdapter.FindServiceReturns(nil, err)

		p := consul.NewConsulPublisher(fakeAdapter, "myservice", 5*time.Second)
		defer p.Stop()

		c := make(chan []*url.URL)
		p.Subscribe(c)
		defer p.Unsubscribe(c)

		var urls []*url.URL

		Eventually(c).Should(Receive(&urls))
		Expect(len(urls)).To(Equal(0))
	})

	It("subscriptions should receive the name of the service", func() {
		service := &consul_api.AgentService{
			Service: "service",
			Address: "127.0.0.1",
			Port:    3000,
		}
		entry := []*consul_api.ServiceEntry{
			&consul_api.ServiceEntry{
				Service: service,
			},
		}

		fakeAdapter.PingReturns(nil)
		fakeAdapter.CheckServiceReturns(entry, nil)

		p := consul.NewConsulPublisher(fakeAdapter, service.Service, 1*time.Second)
		defer p.Stop()

		c := make(chan []*url.URL)
		p.Subscribe(c)
		defer p.Unsubscribe(c)

		Eventually(c).Should(Receive())
	})

	const TIMEOUT = 3 * time.Second
	Context("running consul cluster", func() {
		var r registry.RegistryAdapter
		var sr registry.ServiceRegistration
		var sr2 registry.ServiceRegistration

		BeforeEach(func() {
			startCluster()
			//config := registry.Config{AdapterURI: "consul://" + clusterRunner.Address()}
			config := registry.Config{AdapterURI: "consul://" + cluster.Leader.HTTPAddr}

			var err error
			r, err = registry.NewBackend(config)
			Expect(err).ToNot(HaveOccurred())
			Expect(r.Status()).To(Equal(registry.StatusConnected))

			sr = registry.ServiceRegistration{AdvertiseAddr: "127.0.0.1", Port: 3001, Id: "router1", Name: "bifrost", Tags: []string{"v1"}}
			sr2 = registry.ServiceRegistration{AdvertiseAddr: "127.0.0.2", Port: 3001, Id: "router2", Name: "bifrost", Tags: []string{"v1"}}

			err = r.Register(sr)
			Expect(err).ToNot(HaveOccurred())
			r.Sync(sr)
			err = r.Register(sr2)
			Expect(err).ToNot(HaveOccurred())
			r.Sync(sr2)

		})

		AfterEach(func() {
			stopCluster()
		})

		It("should broadcast the registered service url to multiple subscribers", func() {
			p := consul.NewConsulPublisher(r, "bifrost", 1*time.Second)
			defer p.Stop()

			c1 := make(chan []*url.URL)
			c2 := make(chan []*url.URL)
			p.Subscribe(c1)
			defer p.Unsubscribe(c1)

			Eventually(c1).Should(Receive())

			p.Subscribe(c2)
			defer p.Unsubscribe(c2)

			var urls []*url.URL
			Eventually(c2).Should(Receive(&urls))
		})

		It("should not publish a service url when the service has gone down", func() {
			p := consul.NewConsulPublisher(r, "bifrost", 1*time.Second)
			defer p.Stop()

			c1 := make(chan []*url.URL)
			p.Subscribe(c1)
			defer p.Unsubscribe(c1)

			r.DeRegister(sr2)

			chanTest := func() bool {
				for {
					select {
					case urls := <-c1:
						if len(urls) == 1 {
							return true
						}
					}
				}
			}

			Eventually(chanTest, TIMEOUT).Should(BeTrue())
		})
	})

})
