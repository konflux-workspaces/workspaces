package internalworkspace_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestInternalworkspace(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Internalworkspace Suite")
}
