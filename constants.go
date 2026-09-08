package ulog

// Ключи для JSON-логов, стандартизирующие схему данных в observability-стеке
const (
	LogKeyTraceID  = "trace_id"
	LogKeyStatus   = "status"
	LogKeyDuration = "duration_ms"
	LogKeyMethod   = "method"
	LogKeyPath     = "path"
)

// Ключ для контекста Go, чтобы прокидывать Trace ID сквозь приложение
type contextKey string

const traceIDKey contextKey = "trace_id"
