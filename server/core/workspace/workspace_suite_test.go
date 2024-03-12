package workspace_test

import (
	"log/slog"
	"testing"

	"github.com/konflux-workspaces/workspaces/server/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestWorkspace(t *testing.T) {
	slog.SetDefault(slog.New(&log.NoOpHandler{}))

	RegisterFailHandler(Fail)
	RunSpecs(t, "Workspace Suite")
}
