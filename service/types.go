package service

import "github.com/gin-gonic/gin"

type ServiceHandler struct {
	Methods     []string
	Paths       []string
	Handler     gin.HandlerFunc
	Middlewares []gin.HandlerFunc
}
