package simple_s3

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSimpleS3(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SimpleS3 Suite")
}
