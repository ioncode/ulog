package ulog

import (
	"bytes"
	"strings"

	"github.com/rs/zerolog"
)

// ZerologAdapter реализует интерфейс ulog.Logger для rs/zerolog.
type ZerologAdapter struct {
	logger    zerolog.Logger
	hasCaller bool // Флаг: включен ли caller у исходного логгера
}

// NewZerologAdapter создает экземпляр адаптера и определяет, нужно ли обрабатывать стек.
func NewZerologAdapter(l zerolog.Logger) Logger {
	return &ZerologAdapter{
		logger:    l,
		hasCaller: checkZerologHasCaller(l),
	}
}

// checkZerologHasCaller делает один фейковый тестовый лог в буфер,
// чтобы проверить, генерирует ли исходный логгер поле "caller".
func checkZerologHasCaller(l zerolog.Logger) bool {
	var buf bytes.Buffer
	testLogger := l.Output(&buf)
	testLogger.Log().Msg("")

	res := buf.String()
	return strings.Contains(res, "caller")
}

// withCaller настраивает логгер с правильным сдвигом кадра, если у пользователя включен caller.
func (a *ZerologAdapter) withCaller() zerolog.Logger {
	if !a.hasCaller {
		return a.logger
	}
	// 1 кадр — метод утилит ulog (например, Info())
	// 2 кадр — внутренний хелпер withCaller()
	// 3 кадр — интерфейсная граница (полиморфизм Go)
	return a.logger.With().CallerWithSkipFrameCount(3).Logger()
}

func (a *ZerologAdapter) Info() LoggerEvent {
	l := a.withCaller() // Сохраняем в переменную (теперь значение адресуемо!)
	return &zerologEvent{event: l.Info()}
}

func (a *ZerologAdapter) Debug() LoggerEvent {
	l := a.withCaller()
	return &zerologEvent{event: l.Debug()}
}

func (a *ZerologAdapter) Warn() LoggerEvent {
	l := a.withCaller()
	return &zerologEvent{event: l.Warn()}
}

func (a *ZerologAdapter) Error() LoggerEvent {
	l := a.withCaller()
	return &zerologEvent{event: l.Error()}
}

// zerologEvent реализует ulog.LoggerEvent для цепочки вызовов Fluent API.
type zerologEvent struct {
	event *zerolog.Event
}

func (e *zerologEvent) Str(key string, val string) LoggerEvent {
	e.event = e.event.Str(key, val)
	return e
}

func (e *zerologEvent) Int(key string, val int) LoggerEvent {
	e.event = e.event.Int(key, val)
	return e
}

func (e *zerologEvent) Err(err error) LoggerEvent {
	e.event = e.event.Err(err)
	return e
}

func (e *zerologEvent) Msg(msg string) {
	e.event.Msg(msg)
}
