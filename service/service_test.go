package service_test

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gitlab.vailsys.com/vail-cloud-services/platform/middleware"
	"gitlab.vailsys.com/vail-cloud-services/platform/registry"
	"gitlab.vailsys.com/vail-cloud-services/platform/service"
)

var _ = Describe("Service", func() {
	Context("registry is down", func() {
		It("should create a service that is not synced", func() {
			nodes := []string{"consul://127.0.0.2:8500"}
			config := registry.ServiceRegistration{Address: "127.0.0.2", Port: 3001, Id: "router1", Name: "bifrost", Tags: []string{"v1"}, ConsulNodes: nodes, TTL: "5s"}
			ser := service.NewService(config)

			Expect(ser.Name()).To(Equal(config.Name))
			Expect(ser.Synced()).To(Equal(false))
			err := ser.Run()
			Expect(err).To(HaveOccurred())
		})
	})

	Context("registry is available", func() {
		var ser *service.Service

		testHandler := func(c *gin.Context) {
			c.String(http.StatusOK, "hello world")
		}

		It("should not sync the service at all", func() {
			nodes := []string{"consul://127.0.0.2:8500"}
			config := registry.ServiceRegistration{
				Address:          "127.0.0.1",
				Port:             13001,
				Id:               "router1",
				Name:             "bifrost",
				Tags:             []string{"v1"},
				ConsulNodes:      nodes,
				TTL:              "5s",
				SkipRegistration: true,
			}
			ser := service.NewService(config)

			var err error

			go func() {
				err = ser.Run()
				Expect(err).ToNot(HaveOccurred())
			}()

			Eventually(func() bool {
				return ser.Synced()
			}).Should(BeFalse())
		})

		Context("registring services", func() {
			BeforeEach(func() {
				nodes := []string{"consul://" + cluster.Leader.HTTPAddr}
				config := registry.ServiceRegistration{
					Address:     "127.0.0.1",
					Port:        3001,
					Id:          "router1",
					Name:        "bifrost",
					Tags:        []string{"v1"},
					ConsulNodes: nodes,
					TTL:         "5s",
				}
				ser = service.NewService(config)

				Expect(ser.Name()).To(Equal(config.Name))
			})

			It("should forward the registered handler", func() {
				h := service.ServiceHandler{
					Methods: []string{"GET"},
					Paths:   []string{"/wtf"},
					Handler: testHandler,
				}

				err := ser.AddHandler(h)

				Expect(err).NotTo(HaveOccurred())

				ts := httptest.NewServer(ser.Router)
				defer ts.Close()

				res, err := http.Get(ts.URL + "/wtf")
				if err != nil {
					Expect(err).NotTo(HaveOccurred())
				}
				body, err := ioutil.ReadAll(res.Body)
				res.Body.Close()
				if err != nil {
					Expect(err).NotTo(HaveOccurred())
				}

				Expect(string(body)).To(Equal("hello world"))
			})

			It("should send heartbeat to the registry", func() {
				go func() {
					ser.Run()
				}()

				time.Sleep(5 * time.Second)
				Expect(ser.Synced()).To(Equal(true))

				time.Sleep(5 * time.Second)
				Expect(ser.Synced()).To(Equal(true))

				ser.Stop()
				Expect(ser.Synced()).To(Equal(false))
			})

			It("should shutdown all discovery clients when service is stopped", func() {
				go func() {
					ser.Run()
				}()
			})

			It("should add a middleware to the service router", func() {
				handler := middleware.NewTestMiddleware()

				err := ser.AddMiddleware(handler)
				Expect(err).ToNot(HaveOccurred())

				ts := httptest.NewServer(ser.Router)
				defer ts.Close()

				res, err := http.Get(ts.URL + "/wtf")
				if err != nil {
					Expect(err).NotTo(HaveOccurred())
				}
				body, err := ioutil.ReadAll(res.Body)
				res.Body.Close()
				if err != nil {
					Expect(err).NotTo(HaveOccurred())
				}

				Expect(string(body)).To(Equal("middleware"))
			})

			It("should be able to add a custom not found handler", func() {
				notfound := func(c *gin.Context) {
					c.String(http.StatusNotFound, "not found biatch")
				}

				err := ser.SetNotFoundHandler(notfound)
				Expect(err).ToNot(HaveOccurred())

				ts := httptest.NewServer(ser.Router)
				defer ts.Close()

				res, err := http.Get(ts.URL + "/baitch")
				if err != nil {
					Expect(err).NotTo(HaveOccurred())
				}
				body, err := ioutil.ReadAll(res.Body)
				res.Body.Close()
				if err != nil {
					Expect(err).NotTo(HaveOccurred())
				}

				Expect(string(body)).To(Equal("not found biatch"))
			})
		})
	})
})
