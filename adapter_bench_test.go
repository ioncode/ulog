package ulog_test

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/ioncode/ulog/v3"
	"github.com/ioncode/ulog/v3/adapters/uslog"
	"github.com/ioncode/ulog/v3/adapters/uzerolog"
	"github.com/rs/zerolog"
)

// BenchmarkZerologAdapter_FastPath measures non-allocating rs/zerolog adapter evaluation speeds.
func BenchmarkZerologAdapter_FastPath(b *testing.B) {
	nativeLogger := zerolog.New(io.Discard).With().Timestamp().Logger()
	logger := uzerolog.NewZerologAdapter(nativeLogger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("failed to process incoming http request",
			ulog.String(ulog.LogKeyTraceID, "amzn-req-12345"),
			ulog.Int(ulog.LogKeyStatus, 500),
		)
	}
}

// BenchmarkZerologAdapter_SlowPath triggers application caller scanning steps utilizing standard stack maps.
func BenchmarkZerologAdapter_SlowPath(b *testing.B) {
	nativeLogger := zerolog.New(io.Discard).With().Timestamp().Caller().Logger()
	logger := uzerolog.NewZerologAdapter(nativeLogger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("failed to process incoming http request",
			ulog.String(ulog.LogKeyTraceID, "amzn-req-12345"),
			ulog.Int(ulog.LogKeyStatus, 500),
		)
	}
}

// BenchmarkZerologAdapter_ErrorPath evaluates structured fallback speeds tracking standard error parameters inline.
func BenchmarkZerologAdapter_ErrorPath(b *testing.B) {
	nativeLogger := zerolog.New(io.Discard).With().Timestamp().Logger()
	logger := uzerolog.NewZerologAdapter(nativeLogger)
	errPayload := errors.New("sql: connection pool exhausted")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Error("failed to process incoming http request", errPayload,
			ulog.String(ulog.LogKeyTraceID, "amzn-req-12345"),
			ulog.Int(ulog.LogKeyStatus, 500),
		)
	}
}

// BenchmarkSlogAdapter_FastPath tracks native standard log/slog pipelines skipping stack location lookups.
func BenchmarkSlogAdapter_FastPath(b *testing.B) {
	opts := &slog.HandlerOptions{AddSource: false}
	nativeLogger := slog.New(slog.NewJSONHandler(io.Discard, opts))
	logger := uslog.NewSlogAdapter(nativeLogger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("failed to process incoming http request",
			ulog.String(ulog.LogKeyTraceID, "amzn-req-12345"),
			ulog.Int(ulog.LogKeyStatus, 500),
		)
	}
}

// BenchmarkSlogAdapter_SlowPath evaluates raw performance drops caused by runtime location matching inside standard handlers.
func BenchmarkSlogAdapter_SlowPath(b *testing.B) {
	opts := &slog.HandlerOptions{AddSource: true}
	nativeLogger := slog.New(slog.NewJSONHandler(io.Discard, opts))
	logger := uslog.NewSlogAdapter(nativeLogger)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("failed to process incoming http request",
			ulog.String(ulog.LogKeyTraceID, "amzn-req-12345"),
			ulog.Int(ulog.LogKeyStatus, 500),
		)
	}
}
