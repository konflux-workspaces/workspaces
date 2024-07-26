package log_test

import (
	"context"
	"log/slog"
	"math"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/konflux-workspaces/workspaces/server/log"
)

var _ = DescribeTable("NoOpHandler is disabled", func(logLevel slog.Level) {
	// given
	handler := &log.NoOpHandler{}

	// when
	enabled := handler.Enabled(context.TODO(), logLevel)

	// then
	Expect(enabled).To(BeFalse())
},
	Entry("Minimum LogLevel value", slog.Level(math.MinInt)),
	Entry("Maximum LogLevel value", slog.Level(math.MaxInt)),
	Entry("Debug LogLevel value", slog.LevelDebug),
	Entry("Info LogLevel value", slog.LevelInfo),
	Entry("Warn LogLevel value", slog.LevelWarn),
	Entry("Error LogLevel value", slog.LevelError),
)

var _ = DescribeTable("NoOpHandler Handle func returns nil", func(record slog.Record) {
	// given
	handler := &log.NoOpHandler{}

	// when
	err := handler.Handle(context.TODO(), record)

	// then
	Expect(err).NotTo(HaveOccurred())
},
	Entry("empty record", slog.Record{}),
)

var _ = DescribeTable("NoOpHandler WithAttrs returns the same handler", func(attrs []slog.Attr) {
	// given
	handler := &log.NoOpHandler{}

	// when
	newHandler := handler.WithAttrs(attrs)

	// then
	Expect(newHandler).To(Equal(handler))
},
	Entry("nil attrs", nil),
	Entry("empty list of attrs", []slog.Attr{}),
	Entry("some attrs", []slog.Attr{
		{Key: "myattr-str", Value: slog.StringValue("myvalue")},
		{Key: "myattr-any", Value: slog.AnyValue("myvalue")},
		{Key: "myattr-int", Value: slog.IntValue(32)},
		{Key: "myattr-int64", Value: slog.Int64Value(342)},
		{Key: "myattr-duration", Value: slog.DurationValue(time.Duration(1 * time.Second))},
		{Key: "myattr-bool", Value: slog.BoolValue(true)},
	}),
)

var _ = DescribeTable("NoOpHandler WithGroup returns the same handler", func(group string) {
	// given
	handler := &log.NoOpHandler{}

	// when
	newHandler := handler.WithGroup(group)

	// then
	Expect(newHandler).To(Equal(handler))
},
	Entry("empty name for group", ""),
	Entry("mygroup", "mygroup"),
)
