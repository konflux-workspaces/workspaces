package writeclient_test

import (
	"log/slog"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/konflux-workspaces/workspaces/server/log"
)

func TestWriteclient(t *testing.T) {
	slog.SetDefault(slog.New(&log.NoOpHandler{}))

	RegisterFailHandler(Fail)
	RunSpecs(t, "WriteClient Suite")
}
