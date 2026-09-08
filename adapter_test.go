package ulog

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

// Структура для парсинга JSON-логов в тестах
type logOutput struct {
	Level   string       `json:"level"`
	Message string       `json:"message"`
	Msg     string       `json:"msg"` // slog по умолчанию использует ключ msg
	Source  *slog.Source `json:"source"`
	Caller  string       `json:"caller"`
	TraceID string       `json:"trace_id"`
	Err     string       `json:"error"`
	Custom  string       `json:"custom_key"`
}

func TestZerologAdapter(t *testing.T) {
	var buf bytes.Buffer

	// Инициализируем zerolog с выводом в буфер и включенным Caller
	baseLogger := zerolog.New(&buf).With().Timestamp().Caller().Logger()
	logger := NewZerologAdapter(baseLogger)

	// Тестируем цепочку Fluent API
	// ВАЖНО: Запомните номер этой строки (например, это строка 41), чтобы сверить её в тесте Caller!
	logger.Info().Str("trace_id", "123").Err(errors.New("db_error")).Msg("test message zerolog") // line 41

	var out logOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("failed to unmarshal log: %v", err)
	}

	// Проверяем данные
	if out.Level != "info" {
		t.Errorf("expected level info, got %s", out.Level)
	}
	if out.Message != "test message zerolog" {
		t.Errorf("expected msg 'test message zerolog', got %s", out.Message)
	}
	if out.TraceID != "123" {
		t.Errorf("expected trace_id 123, got %s", out.TraceID)
	}
	if out.Err != "db_error" {
		t.Errorf("expected error db_error, got %s", out.Err)
	}

	// Проверяем, что Caller указывает на этот тестовый файл, а не на внутренности ulog
	if !strings.Contains(out.Caller, "adapter_test.go") {
		t.Errorf("expected caller to contain adapter_test.go, got %s", out.Caller)
	}
}

func TestSlogAdapter(t *testing.T) {
	var buf bytes.Buffer

	// Инициализируем slog с включенным отображением Source (Caller)
	opts := &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}
	baseLogger := slog.New(slog.NewJSONHandler(&buf, opts))
	logger := NewSlogAdapter(baseLogger)

	// Тестируем цепочку Fluent API
	// ВАЖНО: Запомните номер этой строки (например, строка 79)
	logger.Warn().Str("custom_key", "value").Msg("test message slog") // line 79

	var out logOutput
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("failed to unmarshal log: %v", err)
	}

	// Проверяем данные (у slog уровни обычно пишутся в верхнем регистре: WARN)
	if strings.ToLower(out.Level) != "warn" {
		t.Errorf("expected level warn, got %s", out.Level)
	}
	if out.Msg != "test message slog" {
		t.Errorf("expected msg 'test message slog', got %s", out.Msg)
	}
	if out.Custom != "value" {
		t.Errorf("expected custom_key value, got %s", out.Custom)
	}

	// Проверяем корректность Caller для slog
	if out.Source == nil {
		t.Fatal("expected source (caller) info to be present, got nil")
	}
	if !strings.Contains(out.Source.File, "adapter_test.go") {
		t.Errorf("expected source file to contain adapter_test.go, got %s", out.Source.File)
	}
}
