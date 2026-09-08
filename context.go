package ulog

import (
	"context"
)

// ContextWithTraceID помещает trace_id в контекст и возвращает новый контекст.
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID извлекает trace_id из контекста в виде строки.
// Если trace_id отсутствует, возвращает пустую строку.
func GetTraceID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}
