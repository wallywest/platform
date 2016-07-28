package middleware

import (
	"fmt"
	"net/http"
	"strconv"

	"gitlab.vailsys.com/vail-cloud-services/platform"

	"github.com/gin-gonic/gin"
)

var BasicRealm = "Authorization Required"

type basicAuthMiddleware struct {
	authFn func(string, string, *gin.Context) bool
}

func BasicAuthFunc(fn func(string, string, *gin.Context) bool) *basicAuthMiddleware {
	if fn == nil {
		panic(fmt.Errorf("cannot have an empty auth function"))
	}
	return &basicAuthMiddleware{authFn: fn}
}

func (m *basicAuthMiddleware) GinFunc() gin.HandlerFunc {
	realm := "Basic realm=" + strconv.Quote(BasicRealm)

	fn := func(c *gin.Context) {
		found, err := m.verifyCredentials(c)

		if err != nil {
			platform.Logger.Errorf("basic auth middleware returned error: %s", err)
		}

		if !found {
			c.Header("WWW-Authenticate", realm)
			c.JSON(http.StatusUnauthorized, gin.H{"status": http.StatusUnauthorized, "message": "request is not authorized"})
			c.Abort()
		}
	}
	return fn
}

func (m *basicAuthMiddleware) verifyCredentials(c *gin.Context) (bool, error) {
	user, password, ok := c.Request.BasicAuth()

	if !ok {
		return false, fmt.Errorf("error with basic auth credentials in request")
	}

	if password == "" {
		return false, fmt.Errorf("password is blank")
	}

	if !m.authFn(user, password, c) {
		return false, fmt.Errorf("authFn did not match user and password")
	}

	c.Set("authorizedId", user)
	c.Set("authorizedToken", password)

	return true, nil
}
