package ulog

import "context"

type ctxKeyTraceID struct{}

var traceIDKey = ctxKeyTraceID{}

// ContextWithTraceID кладет Trace ID в контекст запроса.
func ContextWithTraceID(ctx context.Context, traceID string) context.Context {
	return context.WithValue(ctx, traceIDKey, traceID)
}

// GetTraceID safely extracts Trace ID from context.
func GetTraceID(ctx context.Context) string {
	if id, ok := ctx.Value(traceIDKey).(string); ok {
		return id
	}
	return ""
}
