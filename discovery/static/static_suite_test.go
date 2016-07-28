package static_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestStatic(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Static Publisher Suite")
}
