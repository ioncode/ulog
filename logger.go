// Package ulog provides a high-performance, structured, zero-allocation
// logger wrapper designed specifically for Go microservices following Clean Architecture.
package ulog

// Kind defines the data type of a logging field to eliminate interface{} wrapping boundaries.
type Kind uint8

const (
	// KindString represents a string field value.
	KindString Kind = iota
	// KindInt represents an int64 field value.
	KindInt
	// KindBool represents a boolean field value encoded as int64.
	KindBool
)

// Field represents a flat, strongly-typed key-value pair that is allocated on the stack to prevent heap noise.
type Field struct {
	Key       string
	Kind      Kind
	IntVal    int64
	StringVal string
}

// String constructs a strongly-typed string logging field without heap allocations.
func String(key, val string) Field {
	return Field{Key: key, Kind: KindString, StringVal: val}
}

// Int constructs a strongly-typed integer logging field without heap allocations.
func Int(key string, val int) Field {
	return Field{Key: key, Kind: KindInt, IntVal: int64(val)}
}

// Bool constructs a strongly-typed boolean logging field without heap allocations.
func Bool(key string, val bool) Field {
	var intVal int64
	if val {
		intVal = 1
	}
	return Field{Key: key, Kind: KindBool, IntVal: intVal}
}

// Logger defines a uniform, explicit polymorphic interface for structured logging across decoupled business layers.
type Logger interface {
	// Info logs a structured message at the informational level.
	Info(msg string, fields ...Field)
	// Error logs a structured message at the error level with an explicit error telemetry object.
	Error(msg string, err error, fields ...Field)
	// Debug logs a structured message at the debug level for development-phase telemetry tracking.
	Debug(msg string, fields ...Field)
	// With pre-serializes and attaches contextual fields to a newly spawned child Logger branch.
	With(fields ...Field) Logger
}
