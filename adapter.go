package ulog

import (
	"github.com/rs/zerolog"
)

// LoggerEvent описывает методы Fluent API, которые гарантирует библиотека.
type LoggerEvent interface {
	Str(key string, val string) LoggerEvent
	Int(key string, val int) LoggerEvent
	Err(err error) LoggerEvent
	Msg(msg string)
}

// ZerologAdapter адаптирует логер zerolog под интерфейс LoggerEvent.
type ZerologAdapter struct {
	ZLog zerolog.Logger
}

// NewZerologAdapter создает новый экземпляр адаптера.
func NewZerologAdapter(zLog zerolog.Logger) *ZerologAdapter {
	return &ZerologAdapter{ZLog: zLog}
}

func (a *ZerologAdapter) Info() LoggerEvent {
	return &zerologEventAdapter{event: a.ZLog.Info()}
}

func (a *ZerologAdapter) Error() LoggerEvent {
	return &zerologEventAdapter{event: a.ZLog.Error()}
}

func (a *ZerologAdapter) Debug() LoggerEvent {
	return &zerologEventAdapter{event: a.ZLog.Debug()}
}

type zerologEventAdapter struct {
	event *zerolog.Event
}

func (e *zerologEventAdapter) Str(key string, val string) LoggerEvent {
	e.event.Str(key, val)
	return e
}

func (e *zerologEventAdapter) Int(key string, val int) LoggerEvent {
	e.event.Int(key, val)
	return e
}

func (e *zerologEventAdapter) Err(err error) LoggerEvent {
	e.event.Err(err)
	return e
}

func (e *zerologEventAdapter) Msg(msg string) {
	e.event.Msg(msg)
}
