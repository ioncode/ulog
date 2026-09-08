package ulog

import (
	"bytes"
	"context"
	"log/slog"
	"runtime"
	"strings"
	"time"
)

// SlogAdapter реализует интерфейс ulog.Logger для стандартного log/slog.
type SlogAdapter struct {
	logger    *slog.Logger
	addSource bool
}

// NewSlogAdapter создает экземпляр SlogAdapter и заранее вычисляет, активен ли AddSource.
func NewSlogAdapter(l *slog.Logger) Logger {
	return &SlogAdapter{
		logger:    l,
		addSource: checkSlogAddSource(l),
	}
}

// checkSlogAddSource зондирует обработчик, чтобы проверить, извлекает ли он файлы исходного кода.
func checkSlogAddSource(l *slog.Logger) bool {
	var buf bytes.Buffer
	testHandler := slog.NewJSONHandler(&buf, &slog.HandlerOptions{AddSource: true})

	// Используем массив uintptr под один кадр стека (Program Counter)
	var pcs [1]uintptr
	runtime.Callers(1, pcs[:])
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "probe", pcs[0])

	// Передаем тестовую запись во временный тестовый обработчик
	_ = testHandler.Handle(context.Background(), r)

	// Если в итоговой строке появились ключи source или file, значит AddSource активен у пользователя
	res := buf.String()
	return strings.Contains(res, "source") || strings.Contains(res, "file")
}

func (a *SlogAdapter) Info() LoggerEvent {
	return &slogEvent{logger: a.logger, level: slog.LevelInfo, addSource: a.addSource}
}

func (a *SlogAdapter) Debug() LoggerEvent {
	return &slogEvent{logger: a.logger, level: slog.LevelDebug, addSource: a.addSource}
}

func (a *SlogAdapter) Warn() LoggerEvent {
	return &slogEvent{logger: a.logger, level: slog.LevelWarn, addSource: a.addSource}
}

func (a *SlogAdapter) Error() LoggerEvent {
	return &slogEvent{logger: a.logger, level: slog.LevelError, addSource: a.addSource}
}

// slogEvent накапливает атрибуты перед окончательной записью структуры JSON.
type slogEvent struct {
	logger    *slog.Logger
	level     slog.Level
	addSource bool
	attrs     []any
}

func (e *slogEvent) Str(key string, val string) LoggerEvent {
	e.attrs = append(e.attrs, slog.String(key, val))
	return e
}

func (e *slogEvent) Int(key string, val int) LoggerEvent {
	e.attrs = append(e.attrs, slog.Int(key, val))
	return e
}

func (e *slogEvent) Err(err error) LoggerEvent {
	if err != nil {
		e.attrs = append(e.attrs, slog.Any("error", err))
	}
	return e
}

func (e *slogEvent) Msg(msg string) {
	if !e.logger.Enabled(context.Background(), e.level) {
		return
	}

	// 🚀 БЫСТРЫЙ ПУТЬ: Выполняется без тяжелых накладных расходов runtime.Callers
	if !e.addSource {
		e.logger.Log(context.Background(), e.level, msg, e.attrs...)
		return
	}

	// 🐢 МЕДЛЕННЫЙ ПУТЬ: Вычисляется только если пользователь сам явно запросил отслеживание файлов
	var pcs [1]uintptr
	runtime.Callers(2, pcs[:])

	r := slog.NewRecord(time.Now(), e.level, msg, pcs[0])
	r.Add(e.attrs...)

	_ = e.logger.Handler().Handle(context.Background(), r)
}
