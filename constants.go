package ulog

// Константы для унификации ключей логирования во всех ваших микросервисах.
const (
	LogKeyTraceID  = "trace_id"
	LogKeyMethod   = "method"
	LogKeyPath     = "path"
	LogKeyStatus   = "status"
	LogKeySize     = "size"
	LogKeyDuration = "duration"
	LogKeyPanic    = "panic"
	LogKeyStack    = "stack"
)
