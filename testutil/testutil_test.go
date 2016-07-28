package testutil_test

import (
	"gitlab.vailsys.com/vail-cloud-services/platform/testutil"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testutil", func() {
	It("should create a mockservice", func() {
		service := testutil.NewMockService()
		Expect(service.Name()).To(Equal("testservice"))
	})

})
