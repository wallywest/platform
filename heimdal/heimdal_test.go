package heimdal_test

import (
	"net/http"
	"net/url"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"gitlab.vailsys.com/vail-cloud-services/platform/discovery"
	"gitlab.vailsys.com/vail-cloud-services/platform/discovery/static"
	"gitlab.vailsys.com/vail-cloud-services/platform/heimdal"
)

var _ = Describe("Heimdal", func() {
	var server *ghttp.Server
	var urls []*url.URL
	var lb discovery.LoadBalancer
	var lb2 discovery.LoadBalancer

	BeforeEach(func() {
		server = ghttp.NewServer()
		urls = []*url.URL{
			&url.URL{Scheme: "http", Host: server.Addr()},
		}

		publisher := static.NewStaticPublisher(urls)
		publisher2 := static.NewStaticPublisher([]*url.URL{})

		lb = discovery.RoundRobin(publisher)
		lb2 = discovery.RoundRobin(publisher2)
	})

	AfterEach(func() {
		server.Close()
	})

	It("should throw an error for no downstream service", func() {

		client := heimdal.NewHttpServiceClient("downstream", lb2)

		Expect(client).ToNot(BeNil())

		options := heimdal.HttpRequestBuilderOptions{
			ID:     "test",
			Path:   "/wtf",
			Method: "GET",
		}

		builder := heimdal.NewHttpRequestBuilder(options)

		resp, err := client.Execute(builder)

		Expect(err).To(HaveOccurred())
		Expect(resp).To(BeNil())
	})

	It("should do simple loadbalancing on downstream dependency", func() {
		client := heimdal.NewHttpServiceClient("downstream", lb)
		Expect(client).ToNot(BeNil())

		options := heimdal.HttpRequestBuilderOptions{
			ID:     "test",
			Path:   "/wtf",
			Method: "GET",
		}

		builder := heimdal.NewHttpRequestBuilder(options)

		reqCount := 0
		reqFunc := func(r *http.Request, quitFn func()) {
			reqCount++
		}

		respCount := 0
		respFunc := func(w *http.Response, quitFn func()) {
			respCount++
		}

		builder.AddReqFunc(reqFunc)
		builder.AddRespFunc(respFunc)

		//client.AddBuilder(builder)

		server.AppendHandlers(
			ghttp.CombineHandlers(
				ghttp.VerifyRequest(options.Method, options.Path),
				ghttp.VerifyContentType("application/json"),
			),
		)

		resp, err := client.Execute(builder)

		Expect(err).ToNot(HaveOccurred())
		Expect(reqCount).To(Equal(1))
		Expect(respCount).To(Equal(1))
		Expect(resp).ToNot(BeNil())
	})
})
