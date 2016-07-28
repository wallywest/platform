package static_test

import (
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.vailsys.com/vail-cloud-services/platform/discovery/static"
)

var _ = Describe("static discovery publisher", func() {
	It("publish specified static endpoints", func() {

		endpoints := []*url.URL{
			&url.URL{Scheme: "http", Host: "127.0.0.1"},
			&url.URL{Scheme: "http", Host: "127.0.0.2"},
		}

		p := static.NewStaticPublisher(endpoints)
		defer p.Stop()

		c := make(chan []*url.URL, 1)
		p.Subscribe(c)

		var urls []*url.URL

		Eventually(c).Should(Receive(&urls))
		Expect(len(urls)).To(Equal(2))

		endpoints = []*url.URL{
			&url.URL{Scheme: "http", Host: "127.0.0.1"},
		}

		p.Replace(endpoints)
		Eventually(c).Should(Receive(&urls))
		Expect(len(urls)).To(Equal(1))

	})
})
