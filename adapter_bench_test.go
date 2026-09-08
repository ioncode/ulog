package ulog

import (
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/rs/zerolog"
)

// ============================================================================
// БЕНЧМАРКИ ДЛЯ ZEROLOG
// ============================================================================

// Тестируем быстрый путь Zerolog (без вычисления Caller)
func BenchmarkZerologAdapter_FastPath(b *testing.B) {
	// Инициализируем базовый логгер БЕЗ вызова .Caller()
	baseLogger := zerolog.New(io.Discard).With().Timestamp().Logger()
	logger := NewZerologAdapter(baseLogger)
	err := errors.New("database connection timeout")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().
			Str("trace_id", "amzn-req-12345").
			Int("status_code", 500).
			Err(err).
			Msg("failed to process incoming http request")
	}
}

// Тестируем медленный путь Zerolog (с включенным Caller)
func BenchmarkZerologAdapter_SlowPath(b *testing.B) {
	// Инициализируем базовый логгер С включенным .Caller()
	baseLogger := zerolog.New(io.Discard).With().Timestamp().Caller().Logger()
	logger := NewZerologAdapter(baseLogger)
	err := errors.New("database connection timeout")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().
			Str("trace_id", "amzn-req-12345").
			Int("status_code", 500).
			Err(err).
			Msg("failed to process incoming http request")
	}
}

// ============================================================================
// БЕНЧМАРКИ ДЛЯ SLOG
// ============================================================================

// Тестируем быстрый путь Slog (без вычисления AddSource)
func BenchmarkSlogAdapter_FastPath(b *testing.B) {
	// Инициализируем базовый логгер БЕЗ опции AddSource
	opts := &slog.HandlerOptions{Level: slog.LevelInfo, AddSource: false}
	baseLogger := slog.New(slog.NewJSONHandler(io.Discard, opts))
	logger := NewSlogAdapter(baseLogger)
	err := errors.New("database connection timeout")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().
			Str("trace_id", "amzn-req-12345").
			Int("status_code", 500).
			Err(err).
			Msg("failed to process incoming http request")
	}
}

// Тестируем медленный путь Slog (с включенной опцией AddSource)
func BenchmarkSlogAdapter_SlowPath(b *testing.B) {
	// Инициализируем базовый логгер С включенной опцией AddSource
	opts := &slog.HandlerOptions{Level: slog.LevelInfo, AddSource: true}
	baseLogger := slog.New(slog.NewJSONHandler(io.Discard, opts))
	logger := NewSlogAdapter(baseLogger)
	err := errors.New("database connection timeout")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info().
			Str("trace_id", "amzn-req-12345").
			Int("status_code", 500).
			Err(err).
			Msg("failed to process incoming http request")
	}
}
