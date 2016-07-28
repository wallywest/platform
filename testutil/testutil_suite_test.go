package testutil_test

import (
	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTestutil(t *testing.T) {
	gin.SetMode("test")

	RegisterFailHandler(Fail)
	RunSpecs(t, "Testutil Suite")
}
