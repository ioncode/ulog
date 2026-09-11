package ulog

// Ключи для JSON-логов, стандартизирующие схему данных в observability-стеке
const (
	// LogKeyTraceID is the standardized JSON key for distributed tracing IDs.
	LogKeyTraceID = "trace_id"
	// LogKeyMethod is the standardized JSON key for HTTP request methods.
	LogKeyMethod = "method"
	// LogKeyPath is the standardized JSON key for HTTP request paths.
	LogKeyPath = "path"
	// LogKeyStatus is the standardized JSON key for HTTP response status codes.
	LogKeyStatus = "status"
	// LogKeyDuration is the standardized JSON key for HTTP request processing duration in milliseconds.
	LogKeyDuration = "duration_ms"
	// LogKeyBytesOut is the standardized JSON key for HTTP response size in bytes.
	LogKeyBytesOut = "bytes_out"
	// LogKeyPanicInfo is the standardized JSON key for capturing panic details.
	LogKeyPanicInfo = "panic_info"
)

// Ключ для контекста Go, чтобы прокидывать Trace ID сквозь приложение
type contextKey string

const traceIDKey contextKey = "trace_id"
