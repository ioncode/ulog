package ulog

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
	"unsafe"
)

// traceBuffer encapsulates pre-allocated byte slices for non-allocating correlation ID generations.
type traceBuffer struct {
	raw [8]byte
	hex [16]byte
}

var tracePool = sync.Pool{
	New: func() any {
		return &traceBuffer{}
	},
}

// responseWriterInterceptor transparently intercepts transactional HTTP write status bytes for metadata telemetry metrics.
type responseWriterInterceptor struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func newResponseWriterInterceptor(w http.ResponseWriter) *responseWriterInterceptor {
	return &responseWriterInterceptor{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (w *responseWriterInterceptor) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func (w *responseWriterInterceptor) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.bytesWritten += n
	return n, err
}

// TraceIDMiddleware intercepts downstream pipelines, enforcing distributed tracking generation across request HTTP headers.
func TraceIDMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		traceID := r.Header.Get("X-Trace-ID")
		if traceID == "" {
			buf := tracePool.Get().(*traceBuffer)
			if _, err := rand.Read(buf.raw[:]); err == nil {
				hex.Encode(buf.hex[:], buf.raw[:])
				traceID = unsafe.String(&buf.hex[0], len(buf.hex))
			} else {
				traceID = "gen-fallback-id"
			}
			r.Header.Set("X-Trace-ID", traceID)
			tracePool.Put(buf)
		}
		w.Header().Set("X-Trace-ID", traceID)
		next.ServeHTTP(w, r)
	})
}

// LoggingMiddleware records explicit transactional metadata metrics for completed operations on the stack. Zero-Alloc.
func LoggingMiddleware(baseLogger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			startTime := time.Now()
			interceptor := newResponseWriterInterceptor(w)

			next.ServeHTTP(interceptor, r)

			duration := time.Since(startTime)
			traceID := r.Header.Get("X-Trace-ID")

			fields := []Field{
				String(LogKeyTraceID, traceID),
				String(LogKeyMethod, r.Method),
				String(LogKeyPath, r.URL.Path),
				Int(LogKeyStatus, interceptor.statusCode),
				Int(LogKeyDuration, int(duration.Milliseconds())),
				Int(LogKeyBytesOut, interceptor.bytesWritten),
			}

			if interceptor.statusCode >= 500 {
				baseLogger.Error("http request failed", nil, fields...)
			} else {
				baseLogger.Info("http request processed", fields...)
			}
		})
	}
}

// RecoveryMiddleware captures unexpected downstream routine panics, logging traces using unified constant layout parameters.
func RecoveryMiddleware(baseLogger Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					traceID := r.Header.Get("X-Trace-ID")
					fields := []Field{
						String(LogKeyTraceID, traceID),
						String(LogKeyMethod, r.Method),
						String(LogKeyPath, r.URL.Path),
						String(LogKeyPanicInfo, anyToString(err)),
					}

					baseLogger.Error("http handler panicked", nil, fields...)

					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"error":"internal server error"}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func anyToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case error:
		return t.Error()
	default:
		return "unknown runtime panic"
	}
}
