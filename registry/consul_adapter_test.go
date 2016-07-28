package registry_test

import (
	"time"

	consul_api "github.com/hashicorp/consul/api"
	"gitlab.vailsys.com/vail-cloud-services/platform/registry"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Registry", func() {
	const clusterSize = 1
	const TIMEOUT = 3 * time.Second

	It("should return error for invalid adapterURI", func() {
		config := registry.Config{AdapterURI: "blah://127.0.0.1:8500"}
		_, err := registry.NewBackend(config)
		Expect(err).To(HaveOccurred())
	})

	It("should return an adapter when consul is not running", func() {
		config := registry.Config{AdapterURI: "consul://127.0.0.1:8500"}
		r, err := registry.NewBackend(config)
		Expect(err).ToNot(HaveOccurred())
		Expect(r.Status()).To(Equal(registry.StatusDisconnected))
	})

	Context("consul registry", func() {
		It("should register/deregister a service", func() {
			sr := registry.ServiceRegistration{Address: "127.0.0.1", Port: 3001, Id: "router1", Name: "bifrost", Tags: []string{"v1"}}
			err := r.Register(sr)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() map[string][]string {
				sr, err := r.FindServices()
				Expect(err).ToNot(HaveOccurred())
				return sr
			}, TIMEOUT).Should(HaveLen(2))

			Eventually(func() []*consul_api.CatalogService {
				sr, err := r.FindService("bifrost", "")
				Expect(err).ToNot(HaveOccurred())
				return sr
			}, TIMEOUT).ShouldNot(BeNil())

			var out []*consul_api.ServiceEntry

			Eventually(func() int {
				entries, err := r.CheckService("bifrost", "", false)
				Expect(err).ToNot(HaveOccurred())
				out = entries
				return len(entries)
			}, TIMEOUT).ShouldNot(Equal(0))

			Eventually(func() int {
				r.Sync(sr)
				entries, err := r.CheckService("bifrost", "", true)
				Expect(err).ToNot(HaveOccurred())
				return len(entries)
			}, TIMEOUT).ShouldNot(Equal(0))

			err = r.DeRegister(sr)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() map[string][]string {
				sr, err := r.FindServices()
				Expect(err).ToNot(HaveOccurred())
				return sr
			}, TIMEOUT).Should(HaveLen(1))

			Eventually(func() error {
				_, err := r.FindService("bifrost", "")
				return err
			}, TIMEOUT).Should(HaveOccurred())
		})

		It("should sync with registry after its TTL has expired and be passing", func() {
			//short TTL
			sr1 := registry.ServiceRegistration{Address: "127.0.0.1", Port: 3001, Id: "router1", Name: "bifrost", Tags: []string{"v1"}, TTL: "100ms"}

			err := r.Register(sr1)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() []*consul_api.CatalogService {
				sr, err := r.FindService("bifrost", "")
				Expect(err).ToNot(HaveOccurred())
				return sr
			}, TIMEOUT).ShouldNot(BeNil())

			Eventually(func() int {
				entries, err := r.CheckService("bifrost", "", false)
				Expect(err).ToNot(HaveOccurred())
				return len(entries)
			}, TIMEOUT).Should(Equal(1))

			err = r.Sync(sr1)
			Expect(err).ToNot(HaveOccurred())

			Eventually(func() int {
				entries, err := r.CheckService("bifrost", "", true)
				Expect(err).ToNot(HaveOccurred())
				return len(entries)
			}, TIMEOUT).Should(Equal(0))

		})

	})
})
