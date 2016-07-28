package registry_test

import (
	"time"

	"gitlab.vailsys.com/vail-cloud-services/platform/registry"
	"gitlab.vailsys.com/vail-cloud-services/platform/registry/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Pulser", func() {
	Context("invalid service registration", func() {
		It("should throw an error for invalid ServiceRegistration", func() {
			registration := registry.ServiceRegistration{}
			fakeAdapter := new(fakes.FakeRegistryAdapter)

			_, err := registry.NewPulser(1*time.Second, registration, fakeAdapter)

			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(registry.ErrInvalidServiceRegistration))
		})

		It("should throw an error when the interval > then the TTL", func() {
			fakeAdapter := new(fakes.FakeRegistryAdapter)
			sr := registry.ServiceRegistration{Address: "127.0.0.1", Port: 3001, Id: "router1", Name: "bifrost", Tags: []string{"v1"}, TTL: "2s"}
			_, err := registry.NewPulser(3*time.Second, sr, fakeAdapter)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("valid service registration", func() {

		It("should send pulses to the registry service", func() {
			fakeAdapter := new(fakes.FakeRegistryAdapter)
			sr := registry.ServiceRegistration{Address: "127.0.0.1", Port: 3001, Id: "router1", Name: "bifrost", Tags: []string{"v1"}, TTL: "2s"}

			pulse, err := registry.NewPulser(1*time.Second, sr, fakeAdapter)
			Expect(err).ToNot(HaveOccurred())

			fakeAdapter.StatusReturns(1)

			pulse.Start()

			Eventually(func() int {
				return fakeAdapter.SyncCallCount()
			}, TIMEOUT).Should(Equal(2))

			Expect(pulse.Active()).To(Equal(true))
		})

		It("should continue beating when there is an error syncing with the registry", func() {
			fakeAdapter := new(fakes.FakeRegistryAdapter)
			sr := registry.ServiceRegistration{Address: "127.0.0.1", Port: 3001, Id: "router1", Name: "bifrost", Tags: []string{"v1"}, TTL: "2s"}

			pulse, err := registry.NewPulser(1*time.Second, sr, fakeAdapter)
			Expect(err).ToNot(HaveOccurred())

			fakeAdapter.SyncReturns(registry.ErrSyncing)
			pulse.Start()

			Eventually(func() int {
				return fakeAdapter.PingCallCount()
			}, TIMEOUT).ShouldNot(Equal(0))

			Expect(pulse.Active()).To(Equal(true))
		})
	})

})
