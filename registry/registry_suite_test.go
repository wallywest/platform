package registry_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gitlab.vailsys.com/vail-cloud-services/platform/registry"
	"gitlab.vailsys.com/vail-cloud-services/platform/testutil"

	"testing"
	"time"
)

func TestRegistry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Registry Suite")
}

var r registry.RegistryAdapter
var cluster testutil.ConsulCluster

//var runner *consulrunner.ClusterRunner
var TIMEOUT = 5 * time.Second

var _ = BeforeSuite(func() {
	t, _ := GinkgoT().(*testing.T)
	cluster = testutil.NewConsulCluster(t)

	config := registry.Config{AdapterURI: "consul://" + cluster.Leader.HTTPAddr}
	var err error
	r, err = registry.NewBackend(config)
	Expect(err).ToNot(HaveOccurred())
	Expect(r.Status()).To(Equal(registry.StatusConnected))
})

var _ = AfterSuite(func() {
	cluster.Stop()
})
