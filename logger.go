package ulog

// Logger — основной интерфейс, который принимает бизнес-логика и middleware.
type Logger interface {
	Info() LoggerEvent
	Debug() LoggerEvent
	Warn() LoggerEvent
	Error() LoggerEvent
}

// LoggerEvent — контракт для цепочек методов (Fluent API).
type LoggerEvent interface {
	Str(key string, val string) LoggerEvent
	Int(key string, val int) LoggerEvent
	Err(err error) LoggerEvent
	Msg(msg string)
}
