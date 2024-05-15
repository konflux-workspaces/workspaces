package writeclient_test

import (
	"log/slog"
	"testing"

	"github.com/konflux-workspaces/workspaces/server/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWriteclient(t *testing.T) {
	slog.SetDefault(slog.New(&log.NoOpHandler{}))

	RegisterFailHandler(Fail)
	RunSpecs(t, "WriteClient Suite")
}
