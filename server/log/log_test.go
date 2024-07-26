package log_test

import (
	"context"
	"log/slog"

	"github.com/konflux-workspaces/workspaces/server/log"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Log", func() {
	When("a valid logger is injected in context", func() {
		It("creates a new context with the logger", func() {
			// when
			ctx := log.IntoContext(context.TODO(), slog.Default())

			// then
			Expect(log.FromContext(ctx)).To(Equal(slog.Default()))
		})
	})

	When("an invalid logger is injected in context", func() {
		It("returns the default logger", func() {
			// when
			ctx := log.IntoContext(context.TODO(), nil)

			// then
			Expect(log.FromContext(ctx)).To(Equal(log.DefaultLogger))
		})
	})

	When("no logger is injected in context", func() {
		It("returns the default logger", func() {
			// then
			Expect(log.FromContext(context.TODO())).To(Equal(log.DefaultLogger))
		})
	})
})
