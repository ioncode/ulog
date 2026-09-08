package ulog

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

// Тестовая структура для валидации полей в JSON-логах мидлварей
type middlewareLogOutput struct {
	Level      string `json:"level"`
	Message    string `json:"message"`
	Msg        string `json:"msg"` // Для slog (сообщение пишется в этот ключ)
	TraceID    string `json:"trace_id"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Status     int    `json:"status"`
	DurationMS int    `json:"duration_ms"`
	Error      string `json:"error"`
	StackTrace string `json:"stack_trace"`
}

func TestHTTPPipeline_WithZerolog(t *testing.T) {
	var buf bytes.Buffer

	// 1. Инициализируем адаптер Zerolog
	baseZerolog := zerolog.New(&buf)
	logger := NewZerologAdapter(baseZerolog)

	// 2. Создаем пайплайн и тестовый хендлер
	pipeline := NewHTTPPipeline(logger)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	// 3. Выполняем тестовый HTTP-запрос
	req := httptest.NewRequest(http.MethodPost, "/users", nil)
	req.Header.Set("X-Trace-ID", "custom-zerolog-id")
	rr := httptest.NewRecorder()

	pipeline(handler).ServeHTTP(rr, req)

	// 4. Проверяем HTTP-ответ
	if rr.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", rr.Code)
	}
	if rr.Header().Get("X-Trace-ID") != "custom-zerolog-id" {
		t.Errorf("expected trace id header, got %s", rr.Header().Get("X-Trace-ID"))
	}

	// 5. Проверяем сгенерированный JSON-лог
	var logOut middlewareLogOutput
	if err := json.Unmarshal(buf.Bytes(), &logOut); err != nil {
		t.Fatalf("failed to parse json log: %v", err)
	}

	if logOut.Level != "info" {
		t.Errorf("expected level info, got %s", logOut.Level)
	}
	if logOut.TraceID != "custom-zerolog-id" {
		t.Errorf("expected trace_id custom-zerolog-id, got %s", logOut.TraceID)
	}
	if logOut.Method != http.MethodPost {
		t.Errorf("expected method POST, got %s", logOut.Method)
	}
	if logOut.Path != "/users" {
		t.Errorf("expected path /users, got %s", logOut.Path)
	}
	if logOut.Status != http.StatusCreated {
		t.Errorf("expected logged status 201, got %d", logOut.Status)
	}
}

func TestHTTPPipeline_WithSlog(t *testing.T) {
	var buf bytes.Buffer

	// 1. Инициализируем адаптер Slog
	baseSlog := slog.New(slog.NewJSONHandler(&buf, nil))
	logger := NewSlogAdapter(baseSlog)

	// 2. Создаем пайплайн и тестовый хендлер
	pipeline := NewHTTPPipeline(logger)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// 3. Выполняем HTTP-запрос (без передачи X-Trace-ID, проверяем автогенерацию)
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()

	pipeline(handler).ServeHTTP(rr, req)

	// 4. Проверяем автогенерацию TraceID
	generatedTraceID := rr.Header().Get("X-Trace-ID")
	if generatedTraceID == "" {
		t.Fatal("expected automatically generated X-Trace-ID header, got empty string")
	}

	// 5. Проверяем лог
	var logOut middlewareLogOutput
	if err := json.Unmarshal(buf.Bytes(), &logOut); err != nil {
		t.Fatalf("failed to parse json log: %v", err)
	}

	if strings.ToLower(logOut.Level) != "info" {
		t.Errorf("expected level info, got %s", logOut.Level)
	}
	if logOut.Msg != "HTTP request processed" {
		t.Errorf("expected message, got %s", logOut.Msg)
	}
	if logOut.TraceID != generatedTraceID {
		t.Errorf("expected logged trace_id to match header %s, got %s", generatedTraceID, logOut.TraceID)
	}
	if logOut.Status != http.StatusOK {
		t.Errorf("expected logged status 200, got %d", logOut.Status)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	var buf bytes.Buffer
	baseZerolog := zerolog.New(&buf)
	logger := NewZerologAdapter(baseZerolog)

	// Создаем пайплайн с хендлером, который гарантированно падает в панику
	pipeline := NewHTTPPipeline(logger)
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("something went completely wrong")
	})

	req := httptest.NewRequest(http.MethodDelete, "/delete-everything", nil)
	rr := httptest.NewRecorder()

	// Запускаем — мидлварь должна поймать панику и не дать тесту упасть
	pipeline(panicHandler).ServeHTTP(rr, req)

	// Проверяем, что клиенту ушел красивый статус 500
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500 on panic, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "Internal Server Error") {
		t.Errorf("expected emergency body message, got %s", rr.Body.String())
	}

	// Проверяем запись об ошибке в логе
	var logOut middlewareLogOutput
	if err := json.Unmarshal(buf.Bytes(), &logOut); err != nil {
		t.Fatalf("failed to parse panic json log: %v", err)
	}

	if logOut.Level != "error" {
		t.Errorf("expected level error for panic, got %s", logOut.Level)
	}
	if logOut.Message != "HTTP handler panic recovered" {
		t.Errorf("expected precise panic message, got %s", logOut.Message)
	}
	if !strings.Contains(logOut.Error, "something went completely wrong") {
		t.Errorf("expected root panic reason in logs, got %s", logOut.Error)
	}
	if logOut.StackTrace == "" {
		t.Error("expected stack_trace field to be present, got empty string")
	}
}
