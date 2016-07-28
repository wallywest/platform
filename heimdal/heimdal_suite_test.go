package heimdal_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHeimdal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Heimdal Suite")
}
