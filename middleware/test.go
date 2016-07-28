package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type testMiddleware struct{}

func NewTestMiddleware() *testMiddleware {
	return &testMiddleware{}
}

func (t *testMiddleware) GinFunc() gin.HandlerFunc {
	fn := func(c *gin.Context) {
		c.String(http.StatusOK, "middleware")
	}

	return fn
}
