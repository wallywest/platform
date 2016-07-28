package service_test

import (
	"io/ioutil"

	"github.com/gin-gonic/gin"
	consul_test "github.com/hashicorp/consul/testutil"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.vailsys.com/vail-cloud-services/platform"
	"gitlab.vailsys.com/vail-cloud-services/platform/testutil"

	"testing"
)

func TestService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Service Suite")
}

var server *consul_test.TestServer
var cluster testutil.ConsulCluster

var _ = BeforeSuite(func() {
	gin.SetMode("test")
	platform.SetLogOutput(ioutil.Discard)
	t, _ := GinkgoT().(*testing.T)
	cluster = testutil.NewConsulCluster(t)
})

var _ = AfterSuite(func() {
	cluster.Stop()
})
