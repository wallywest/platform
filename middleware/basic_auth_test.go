package middleware_test

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"gitlab.vailsys.com/vail-cloud-services/platform/middleware"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Basic Auth Middleware", func() {
	gin.SetMode("test")

	It("should panic when there is no authFn", func() {
		Expect(func() { middleware.BasicAuthFunc(nil) }).Should(Panic())
	})

	It("should verify missing basic auth fails", func() {
		fn := func(user, pass string, c *gin.Context) bool {
			return true
		}

		router := gin.New()
		m := middleware.BasicAuthFunc(fn)
		router.Use(m.GinFunc())

		ts := httptest.NewServer(router)
		defer ts.Close()

		res, err := http.Get(ts.URL + "/wtf")

		if err != nil {
			Expect(err).NotTo(HaveOccurred())
		}
		Expect(res.StatusCode).To(Equal(401))
	})

	It("should fail for basic authentication dependencies", func() {
		fn := func(user, pass string, c *gin.Context) bool {
			return true
		}

		m := middleware.BasicAuthFunc(fn)

		router := gin.New()
		router.Use(m.GinFunc())
		router.GET("/wtf", func(c *gin.Context) {
			c.String(200, "wtf")
		})

		client := http.DefaultClient
		ts := httptest.NewServer(router)
		defer ts.Close()

		for auth, valid := range map[string]bool{
			"foo:spam:extra": true,
			"dummy:":         false,
			"dummy":          false,
			"":               false,
		} {
			encoded := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
			r, _ := http.NewRequest("GET", ts.URL+"/wtf", nil)
			r.Header.Set("Authorization", encoded)

			resp, err := client.Do(r)
			Expect(err).ToNot(HaveOccurred())

			if valid {
				Expect(resp.StatusCode).ToNot(Equal(401))
			} else {
				Expect(resp.StatusCode).To(Equal(401))
			}

		}
	})

	It("should fail when the function failure", func() {
		fn := func(user, pass string, c *gin.Context) bool {
			Expect(user).To(Equal("foo"))
			Expect(pass).To(Equal("bar"))
			return false
		}

		m := middleware.BasicAuthFunc(fn)

		router := gin.New()
		router.Use(m.GinFunc())

		client := http.DefaultClient
		ts := httptest.NewServer(router)
		defer ts.Close()

		auth := "foo:bar"

		encoded := "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
		r, _ := http.NewRequest("GET", ts.URL+"/wtf", nil)
		r.Header.Set("Authorization", encoded)

		resp, err := client.Do(r)
		Expect(err).ToNot(HaveOccurred())
		Expect(resp.StatusCode).To(Equal(401))

	})
})
