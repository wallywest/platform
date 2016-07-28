package consul_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.vailsys.com/vail-cloud-services/platform/testutil"

	"testing"
)

func TestConsul(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Consul Publisher Suite")
}

var cluster testutil.ConsulCluster

func startCluster() {
	t, _ := GinkgoT().(*testing.T)
	cluster = testutil.NewConsulCluster(t)
}

func stopCluster() {
	cluster.Stop()
}
