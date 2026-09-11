package ulog_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/ioncode/ulog/v3"
)

// mockLogger собирает залогированные поля в памяти для их последующей валидации в тестах
type mockLogger struct {
	mu     sync.Mutex
	msg    string
	err    error
	fields []ulog.Field
}

func (m *mockLogger) Info(msg string, fields ...ulog.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.msg = msg
	m.fields = fields
}

func (m *mockLogger) Error(msg string, err error, fields ...ulog.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.msg = msg
	m.err = err
	m.fields = fields
}

func (m *mockLogger) Debug(msg string, fields ...ulog.Field) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.msg = msg
	m.fields = fields
}

func (m *mockLogger) With(fields ...ulog.Field) ulog.Logger {
	return m // В нашей v2-архитектуре middleware не вызывает .With(), возвращаем сам mock
}

// findField выполняет поиск поля по ключу внутри слайса зафиксированных полей
func findField(fields []ulog.Field, key string) (ulog.Field, bool) {
	for _, f := range fields {
		if f.Key == key {
			return f, true
		}
	}
	return ulog.Field{}, false
}

// 1. Тест для TraceIDMiddleware: проверяет генерацию ID и его передачу в Response заголовки
func TestTraceIDMiddleware(t *testing.T) {
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Внутренний хендлер должен увидеть прокинутый ID в запросе
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			t.Error("Expected X-Trace-ID to be set in request headers")
		}
	})

	middleware := ulog.TraceIDMiddleware(nextHandler)

	req := httptest.NewRequest("GET", "/api/test", nil)
	w := httptest.NewRecorder()

	middleware.ServeHTTP(w, req)

	// Ответ также должен обязательно содержать сгенерированный ID
	if w.Header().Get("X-Trace-ID") == "" {
		t.Error("Expected X-Trace-ID to be set in response headers")
	}
}

// 2. Тест для LoggingMiddleware: проверяет сбор метрик на стеке в конце HTTP-вызова
func TestLoggingMiddleware_Success(t *testing.T) {
	mock := &mockLogger{}

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("created"))
	})

	// Собираем цепочку: TraceID (внешний) -> Logging (внутренний)
	pipeline := ulog.TraceIDMiddleware(ulog.LoggingMiddleware(mock)(nextHandler))

	req := httptest.NewRequest("POST", "/users", nil)
	w := httptest.NewRecorder()

	pipeline.ServeHTTP(w, req)

	mock.mu.Lock()
	defer mock.mu.Unlock()

	if mock.msg != "http request processed" {
		t.Errorf("Unexpected log message: %s", mock.msg)
	}

	// Валидируем стандартизированные системные константы в итоговом логе
	if f, ok := findField(mock.fields, ulog.LogKeyStatus); !ok || f.IntVal != int64(http.StatusCreated) {
		t.Errorf("Expected status 201, got field: %+v", f)
	}

	if f, ok := findField(mock.fields, ulog.LogKeyMethod); !ok || f.StringVal != "POST" {
		t.Errorf("Expected method POST, got field: %+v", f)
	}

	if f, ok := findField(mock.fields, ulog.LogKeyBytesOut); !ok || f.IntVal != 7 {
		t.Errorf("Expected bytes_out 7, got field: %+v", f)
	}

	if _, ok := findField(mock.fields, ulog.LogKeyTraceID); !ok {
		t.Error("Log fields are missing the mandatory standardized trace_id key")
	}
}

// 3. Тест для RecoveryMiddleware: проверяет безопасный перехват паник и запись деталей сбоя
func TestRecoveryMiddleware_HandlesPanic(t *testing.T) {
	mock := &mockLogger{}

	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic(errors.New("db connection timeout"))
	})

	// Цепочка: TraceID -> Recovery -> Паникующий хендлер
	pipeline := ulog.TraceIDMiddleware(ulog.RecoveryMiddleware(mock)(panicHandler))

	req := httptest.NewRequest("GET", "/panic-route", nil)
	w := httptest.NewRecorder()

	pipeline.ServeHTTP(w, req)

	// Сервер не должен упасть. Он обязан вернуть корректный 500 ошибку клиенту.
	if w.Code != http.StatusInternalServerError {
		t.Errorf("Expected status 500, got %d", w.Code)
	}

	mock.mu.Lock()
	defer mock.mu.Unlock()

	if mock.msg != "http handler panicked" {
		t.Errorf("Unexpected panic log message: %s", mock.msg)
	}

	// Проверяем наличие информации о панике по системной константе
	if f, ok := findField(mock.fields, ulog.LogKeyPanicInfo); !ok || f.StringVal != "db connection timeout" {
		t.Errorf("Expected valid panic details, got field: %+v", f)
	}
}
