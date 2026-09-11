package uslog

import (
	"context"
	"log/slog"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/ioncode/ulog/v3"
)

const callerSkip = 3

type FixedAttrBuffer struct {
	buf [32]slog.Attr
}

var attrPool = sync.Pool{
	New: func() any { return &FixedAttrBuffer{} },
}

type SlogAdapter struct {
	logger    *slog.Logger
	addSource bool
}

func NewSlogAdapter(l *slog.Logger) ulog.Logger {
	return &SlogAdapter{
		logger:    l,
		addSource: checkSlogAddSource(l.Handler()),
	}
}

func checkSlogAddSource(h slog.Handler) bool {
	val := reflect.ValueOf(h)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	commonHandlerField := val.FieldByName("commonHandler")
	if commonHandlerField.IsValid() {
		if commonHandlerField.Kind() == reflect.Ptr && !commonHandlerField.IsNil() {
			commonHandlerField = commonHandlerField.Elem()
		}
		optsField := commonHandlerField.FieldByName("opts")
		if optsField.IsValid() {
			addSourceField := optsField.FieldByName("AddSource")
			if addSourceField.IsValid() && addSourceField.Kind() == reflect.Bool {
				return addSourceField.Bool()
			}
		}
	}
	return true
}

func (a *SlogAdapter) Info(msg string, fields ...ulog.Field) {
	a.log(slog.LevelInfo, msg, nil, fields...)
}

func (a *SlogAdapter) Error(msg string, err error, fields ...ulog.Field) {
	a.log(slog.LevelError, msg, err, fields...)
}

func (a *SlogAdapter) Debug(msg string, fields ...ulog.Field) {
	a.log(slog.LevelDebug, msg, nil, fields...)
}

func (a *SlogAdapter) With(fields ...ulog.Field) ulog.Logger {
	if len(fields) == 0 {
		return a
	}
	attrs, poolObj := a.convertFields(fields)
	defer releaseAttrs(poolObj)

	args := make([]any, len(attrs))
	for i, attr := range attrs {
		args[i] = attr
	}
	return &SlogAdapter{logger: a.logger.With(args...), addSource: a.addSource}
}

func (a *SlogAdapter) log(level slog.Level, msg string, err error, fields ...ulog.Field) {
	ctx := context.Background()
	if !a.logger.Enabled(ctx, level) {
		return
	}

	attrs, poolObj := a.convertFields(fields)
	defer releaseAttrs(poolObj)

	if err != nil {
		attrs = append(attrs, slog.Any("error", err))
	}

	// 🚀 ИСПРАВЛЕНО: Используем LogAttrs вместо Log.
	// Это принимает []slog.Attr напрямую и полностью убирает аллокацию слайса []any!
	if !a.addSource {
		a.logger.LogAttrs(ctx, level, msg, attrs...)
		return
	}

	var pcs [1]uintptr
	runtime.Callers(callerSkip, pcs[:])

	record := slog.NewRecord(time.Now(), level, msg, pcs[0])
	record.AddAttrs(attrs...)

	_ = a.logger.Handler().Handle(ctx, record)
}

func (a *SlogAdapter) convertFields(fields []ulog.Field) ([]slog.Attr, *FixedAttrBuffer) {
	poolObj := attrPool.Get().(*FixedAttrBuffer)
	var attrs []slog.Attr
	if len(fields) <= 32 {
		attrs = poolObj.buf[:0]
	} else {
		attrs = make([]slog.Attr, 0, len(fields))
	}

	for i := range fields {
		switch fields[i].Kind {
		case ulog.KindString:
			attrs = append(attrs, slog.String(fields[i].Key, fields[i].StringVal))
		case ulog.KindInt:
			attrs = append(attrs, slog.Int64(fields[i].Key, fields[i].IntVal))
		case ulog.KindBool:
			attrs = append(attrs, slog.Bool(fields[i].Key, fields[i].IntVal == 1))
		}
	}
	return attrs, poolObj
}

func releaseAttrs(p *FixedAttrBuffer) {
	attrPool.Put(p)
}
