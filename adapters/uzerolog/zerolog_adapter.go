package uzerolog

import (
	"bytes"
	"strings"

	"github.com/ioncode/ulog/v3"
	"github.com/rs/zerolog"
)

// callerSkip — константа для встроенного механизма zerolog
const callerSkip = 1

type ZerologAdapter struct {
	logger    zerolog.Logger
	hasCaller bool
}

// NewZerologAdapter теперь просто сохраняет логгер.
// Мы не вызываем тяжелый .With().CallerWithSkipFrameCount(), убирая 96 байт аллокации!
func NewZerologAdapter(l zerolog.Logger) ulog.Logger {
	return &ZerologAdapter{
		logger:    l,
		hasCaller: checkZerologHasCaller(l),
	}
}

func checkZerologHasCaller(l zerolog.Logger) bool {
	var buf bytes.Buffer
	testLogger := l.Output(&buf)
	testLogger.Log().Msg("")
	return strings.Contains(buf.String(), "caller")
}

func (a *ZerologAdapter) Info(msg string, fields ...ulog.Field) {
	var event *zerolog.Event

	// Если включен caller, мы динамически сдвигаем глобальный счетчик
	// ТОЛЬКО на время этого вызова. Это абсолютно потокобезопасно (thread-safe),
	// так как мы меняем локальную копию настроек события, а не глобальный конфиг!
	if a.hasCaller {
		event = a.logger.WithLevel(zerolog.InfoLevel).CallerSkipFrame(callerSkip)
	} else {
		event = a.logger.Info()
	}

	if !event.Enabled() {
		return
	}

	for i := range fields {
		switch fields[i].Kind {
		case ulog.KindString:
			event.Str(fields[i].Key, fields[i].StringVal)
		case ulog.KindInt:
			event.Int64(fields[i].Key, fields[i].IntVal)
		case ulog.KindBool:
			event.Bool(fields[i].Key, fields[i].IntVal == 1)
		}
	}
	event.Msg(msg)
}

func (a *ZerologAdapter) Error(msg string, err error, fields ...ulog.Field) {
	var event *zerolog.Event
	if a.hasCaller {
		event = a.logger.WithLevel(zerolog.ErrorLevel).CallerSkipFrame(callerSkip)
	} else {
		event = a.logger.Error()
	}

	if !event.Enabled() {
		return
	}
	if err != nil {
		event.Err(err)
	}

	for i := range fields {
		switch fields[i].Kind {
		case ulog.KindString:
			event.Str(fields[i].Key, fields[i].StringVal)
		case ulog.KindInt:
			event.Int64(fields[i].Key, fields[i].IntVal)
		case ulog.KindBool:
			event.Bool(fields[i].Key, fields[i].IntVal == 1)
		}
	}
	event.Msg(msg)
}

func (a *ZerologAdapter) Debug(msg string, fields ...ulog.Field) {
	var event *zerolog.Event
	if a.hasCaller {
		event = a.logger.WithLevel(zerolog.DebugLevel).CallerSkipFrame(callerSkip)
	} else {
		event = a.logger.Debug()
	}

	if !event.Enabled() {
		return
	}

	for i := range fields {
		switch fields[i].Kind {
		case ulog.KindString:
			event.Str(fields[i].Key, fields[i].StringVal)
		case ulog.KindInt:
			event.Int64(fields[i].Key, fields[i].IntVal)
		case ulog.KindBool:
			event.Bool(fields[i].Key, fields[i].IntVal == 1)
		}
	}
	event.Msg(msg)
}

func (a *ZerologAdapter) With(fields ...ulog.Field) ulog.Logger {
	if len(fields) == 0 {
		return a
	}
	subContext := a.logger.With()
	for i := range fields {
		switch fields[i].Kind {
		case ulog.KindString:
			subContext = subContext.Str(fields[i].Key, fields[i].StringVal)
		case ulog.KindInt:
			subContext = subContext.Int64(fields[i].Key, fields[i].IntVal)
		case ulog.KindBool:
			subContext = subContext.Bool(fields[i].Key, fields[i].IntVal == 1)
		}
	}
	return &ZerologAdapter{logger: subContext.Logger(), hasCaller: a.hasCaller}
}
