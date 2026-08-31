package ulog

import (
	"net/http"
	"runtime/debug"
	"time"

	"github.com/google/uuid"
)

// pipelineLogger абстрагирует логер для middleware
type pipelineLogger interface {
	Info() LoggerEvent
	Error() LoggerEvent
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	size, err := r.ResponseWriter.Write(b)
	r.size += size
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.status = statusCode
}

// TraceMiddleware отвечает за генерацию и прокидывание Trace ID.
func TraceMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			traceID = uuid.New().String()
		}
		w.Header().Set("X-Trace-ID", traceID)

		ctx := ContextWithTraceID(r.Context(), traceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RecoveryMiddleware принимает интерфейс pipelineLogger
func RecoveryMiddleware(log pipelineLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					traceID := GetTraceID(r.Context())

					log.Error().
						Str(LogKeyTraceID, traceID).
						Str(LogKeyPanic, string(debug.Stack())).
						Msg("panic recovered in HTTP handler")

					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":"internal server error"}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// RequestMetricsMiddleware принимает интерфейс pipelineLogger
func RequestMetricsMiddleware(log pipelineLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lw := &loggingResponseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(lw, r)

			traceID := GetTraceID(r.Context())

			// Считаем длительность в миллисекундах
			durationMs := int(time.Since(start).Milliseconds())

			// ИСПРАВЛЕНО: Заменили несуществующий .Duration() на .Int()
			log.Info().
				Str(LogKeyTraceID, traceID).
				Str(LogKeyMethod, r.Method).
				Str(LogKeyPath, r.URL.Path).
				Int(LogKeyStatus, lw.status).
				Int(LogKeySize, lw.size).
				Int(LogKeyDuration, durationMs). // Передаем как миллисекунды (int)
				Msg("handled HTTP request")
		})
	}
}

// NewHTTPPipeline объединяет все middleware
func NewHTTPPipeline(log *ZerologAdapter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		pipeline := RequestMetricsMiddleware(log)(next)
		pipeline = RecoveryMiddleware(log)(pipeline)
		pipeline = TraceMiddleware(pipeline)
		return pipeline
	}
}
