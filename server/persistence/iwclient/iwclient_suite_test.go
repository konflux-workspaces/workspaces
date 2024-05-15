package iwclient_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestIwclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Iwclient Suite")
}
