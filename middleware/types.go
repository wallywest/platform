package middleware

import "github.com/gin-gonic/gin"

type Middleware interface {
	GinFunc() gin.HandlerFunc
}
