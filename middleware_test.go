package ulog

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Mock-структуры для тестирования без привязки к zerolog
type mockEvent struct {
	fields map[string]interface{}
}

func (m *mockEvent) Str(key string, val string) LoggerEvent { return m }
func (m *mockEvent) Int(key string, val int) LoggerEvent    { return m }
func (m *mockEvent) Err(err error) LoggerEvent              { return m }
func (m *mockEvent) Msg(msg string)                         {}

type mockLogger struct{}

func (m *mockLogger) Info() LoggerEvent  { return &mockEvent{fields: make(map[string]interface{})} }
func (m *mockLogger) Error() LoggerEvent { return &mockEvent{fields: make(map[string]interface{})} }

func TestTraceMiddleware(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()

	var contextTraceID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		contextTraceID = GetTraceID(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	TraceMiddleware(handler).ServeHTTP(rec, req)

	responseTraceID := rec.Header().Get("X-Trace-ID")
	if responseTraceID == "" {
		t.Error("TraceMiddleware: ожидался заголовок X-Trace-ID, получен пустой")
	}

	if contextTraceID != responseTraceID {
		t.Errorf("TraceMiddleware: ID в контексте (%s) не совпадает с ID в ответе (%s)", contextTraceID, responseTraceID)
	}
}

func TestRecoveryMiddleware(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/panic", nil)
	rec := httptest.NewRecorder()

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("критическая ошибка")
	})

	mLog := &mockLogger{}

	RecoveryMiddleware(mLog)(panicHandler).ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("RecoveryMiddleware: ожидался статус 500, получен %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "internal server error") {
		t.Errorf("RecoveryMiddleware: ожидался текст ошибки в теле, получено: %s", rec.Body.String())
	}
}

func TestRequestMetricsMiddleware(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/metrics", nil)
	rec := httptest.NewRecorder()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("hello"))
	})

	mLog := &mockLogger{}

	RequestMetricsMiddleware(mLog)(handler).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("RequestMetricsMiddleware: ожидался статус 201, получен %d", rec.Code)
	}
}
