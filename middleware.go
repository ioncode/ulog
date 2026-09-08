package ulog

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"
)

// responseWriterInterceptor перехватывает HTTP-статус для логирования
type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode int
}

func (w *responseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// NewHTTPPipeline собирает стандартную цепочку production-мидлварей.
// Принимает абстрактный ulog.Logger, что позволяет передавать любой адаптер.
func NewHTTPPipeline(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return TraceIDMiddleware()(
			RecoveryMiddleware(logger)(
				MetricsMiddleware(logger)(next),
			),
		)
	}
}

// TraceIDMiddleware извлекает или генерирует уникальный ID запроса
func TraceIDMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			traceID := r.Header.Get("X-Trace-ID")
			if traceID == "" {
				traceID = generateRandomID()
			}

			// Устанавливаем заголовок ответа для клиента
			w.Header().Set("X-Trace-ID", traceID)

			// Пробрасываем trace_id в контекст запроса, используя ключ из constants.go
			ctx := ContextWithTraceID(r.Context(), traceID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// MetricsMiddleware замеряет время выполнения запроса и логирует результат
func MetricsMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			traceID, _ := r.Context().Value(traceIDKey).(string)

			// Оборачиваем ResponseWriter, чтобы узнать код ответа в конце
			interceptor := &responseWriterInterceptor{ResponseWriter: w, statusCode: http.StatusOK}

			next.ServeHTTP(interceptor, r)

			duration := time.Since(start).Milliseconds()

			// Логируем событие через наш чистый абстрактный интерфейс ulog.Logger
			logger.Info().
				Str(LogKeyTraceID, traceID).
				Str(LogKeyMethod, r.Method).
				Str(LogKeyPath, r.URL.Path).
				Int(LogKeyStatus, interceptor.statusCode).
				Int(LogKeyDuration, int(duration)).
				Msg("HTTP request processed")
		})
	}
}

// RecoveryMiddleware безопасно перехватывает паники, предотвращая падение сервера
func RecoveryMiddleware(logger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					traceID, _ := r.Context().Value(traceIDKey).(string)

					// Превращаем panic в человекочитаемую ошибку и получаем стек-трейс
					panicErr := fmt.Errorf("%v", err)
					stackTrace := string(debug.Stack())

					// Логируем критическую ошибку через ulog.Logger
					logger.Error().
						Str(LogKeyTraceID, traceID).
						Str(LogKeyMethod, r.Method).
						Str(LogKeyPath, r.URL.Path).
						Err(panicErr).
						Str("stack_trace", stackTrace).
						Msg("HTTP handler panic recovered")

					// Возвращаем клиенту красивый 500 статус
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte("Internal Server Error"))
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}

// Вспомогательная функция для генерации криптостойких ID запросов
func generateRandomID() string {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "unknown-trace-id"
	}
	return hex.EncodeToString(bytes)
}
