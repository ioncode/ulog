package ulog_test

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/ioncode/ulog/v3/adapters/uslog"
	"github.com/ioncode/ulog/v3/adapters/uzerolog"
	"github.com/rs/zerolog"
)

// BenchmarkZerolog_CallerValidation проверяет скорость и корректность caller для Zerolog
func BenchmarkZerolog_CallerValidation(b *testing.B) {
	var buf bytes.Buffer
	// Включаем Caller на стороне нативного логгера
	nativeLogger := zerolog.New(&buf).With().Caller().Logger()
	logger := uzerolog.NewZerologAdapter(nativeLogger)

	// Делаем один прогревочный лог, чтобы проверить строку на корректность caller
	logger.Info("check_caller_zerolog")
	output := buf.String()

	// КРИТИЧЕСКАЯ ПРОВЕРКА: Лог должен указывать на ЭТОТ файл (adapter_caller_test.go),
	// а не на внутренний файл uzerolog/zerolog_adapter.go.
	if !strings.Contains(output, "adapter_caller_test.go") {
		b.Fatalf("CRITICAL: Wrong caller depth in Zerolog! Output: %s", output)
	}

	// Сбрасываем буфер и таймер перед запуском высоконагруженного теста
	buf.Reset()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("speed_test")
		buf.Reset() // Очищаем буфер, чтобы не забивать оперативную память в цикле
	}
}

// BenchmarkSlog_CallerValidation проверяет скорость и корректность caller для Slog
func BenchmarkSlog_CallerValidation(b *testing.B) {
	var buf bytes.Buffer
	// Включаем AddSource на стороне нативного slog
	opts := &slog.HandlerOptions{AddSource: true}
	nativeLogger := slog.New(slog.NewJSONHandler(&buf, opts))
	logger := uslog.NewSlogAdapter(nativeLogger)

	// Делаем один прогревочный лог
	logger.Info("check_caller_slog")
	output := buf.String()

	// КРИТИЧЕСКАЯ ПРОВЕРКА: Лог должен указывать на ЭТОТ файл
	if !strings.Contains(output, "adapter_caller_test.go") {
		b.Fatalf("CRITICAL: Wrong caller depth in Slog! Output: %s", output)
	}

	buf.Reset()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		logger.Info("speed_test")
		buf.Reset()
	}
}
