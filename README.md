# ulog (v3.0.0)

**ulog** is a universal, ultra-high-performance architectural wrapper for structured logging in Go microservices. 

Built in strict compliance with **Clean Architecture** and **SOLID** principles, it decouples your core business logic from direct third-party vendor dependencies while providing a flat, polymorphic logging layer designed for **absolute zero-allocation runtime performance** (`0 B/op`, `0 allocs/op`) on hot execution paths.

[![Go Reference](https://go.dev)](https://go.dev)
[![Go Report Card](https://goreportcard.com)](https://goreportcard.com)

## 🚀 Key Features

- **Absolute Zero-Allocation Core**: Complete elimination of interface chaining wrappers (Fluent API) on hot paths. Logs use flat, stack-allocated `ulog.Field` structures to completely spare the Go Garbage Collector (GC).
- **Strict Explicit Dependency Injection (DI)**: No implicit loggers hidden inside `context.Context`. Your structural fields and constructors explicitly define runtime dependencies, resulting in highly testable, predictable code.
- **Isolated Vendor Submodules**: Third-party logging engine adapters reside in discrete directories (`adapters/uzerolog`, `adapters/uslog`). If your service strictly targets the standard library's `slog`, heavy vendor dependencies like `zerolog` are never linked or compiled into your final production binary.
- **Automated Capabilities Probing**: Built-in runtime scanning utilizes reflection and light probing to automatically adjust stack frame skipping behaviors. It seamlessly matches native upstream caller filename configurations (`AddSource`/`Caller`) without forcing you to duplicate configurations via manual operational flags.
- **Microservice Schema Enforcement**: Standardizes log object fields downstream globally using fixed package constants (`LogKeyTraceID`, `LogKeyMethod`, etc.), ensuring monolithic compliance across Kibana, Grafana Loki, or ClickHouse parsers.

---

## 📦 Installation

```bash
go get github.com/ioncode/ulog/v3
```

---

## 🛠️ Quick Start

### Option A: High-Performance `zerolog` (0 Heap Allocations)

Ideal for high-throughput, latency-critical production environments.

```go
package main

import (
	"os"

	"github.com/ioncode/ulog/v3"
	"github.com/ioncode/ulog/v3/adapters/uzerolog"
	"github.com/rs/zerolog"
)

func main() {
	// 1. Initialize native zerolog with caller tracking active
	nativeZerolog := zerolog.New(os.Stdout).With().Timestamp().Caller().Logger()

	// 2. Wrap seamlessly behind ulog core interface
	logger := uzerolog.NewZerologAdapter(nativeZerolog)

	// 3. Log at native engine speeds with 0 heap noise
	logger.Info("user identity verified explicitly",
		ulog.String(ulog.LogKeyTraceID, "req-c128-4f92"),
		ulog.Int(ulog.LogKeyStatus, 200),
		ulog.Bool("is_first_login", false),
	)
}
```

### Option B: Standard Library `slog` (Go 1.21+)

Perfect for minimizing external production module overheads. Utilizes a synchronized internal fixed-size buffer allocation layer to enforce zero-allocation performance constraints.

```go
package main

import (
	"log/slog"
	"os"

	"github.com/ioncode/ulog/v3"
	"github.com/ioncode/ulog/v3/adapters/uslog"
)

func main() {
	// 1. Initialize standard log/slog instance
	nativeSlog := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{AddSource: false}))

	// 2. Wrap using the automated reflection adapter
	logger := uslog.NewSlogAdapter(nativeSlog)

	// 3. Log securely across decoupled service boundaries
	logger.Info("database storage block mapped", 
		ulog.String("node_host", "replica-01.internal"),
		ulog.Int("active_connections", 14),
	)
}
```

---

## 🛡️ Production-Grade Middleware Ecosystem

`ulog` ships with lightweight, modular, Single Responsibility Principle (SRP) compliant middleware components. You retain total structural control over stack sequencing without monolithic constraints:

```go
package main

import (
	"net/http"
	"os"

	"github.com/ioncode/ulog/v3"
	"github.com/ioncode/ulog/v3/adapters/uzerolog"
	"github.com/rs/zerolog"
)

func main() {
	logger := uzerolog.NewZerologAdapter(zerolog.New(os.Stdout))
	mux := http.NewServeMux()

	// Explicit pipeline assembly chain layout order
	var handler http.Handler = mux
	handler = ulog.LoggingMiddleware(logger)(handler)  // 3. Records structured execution payload
	handler = ulog.RecoveryMiddleware(logger)(handler) // 2. Captures panic exceptions securely
	handler = ulog.TraceIDMiddleware(handler)          // 1. Enforces tracing allocation header setups

	http.ListenAndServe(":8080", handler)
}
```

---

## 📊 Absolute Performance Benchmarks

*Evaluated on an AMD Ryzen 5 5600X (6-Core, 12-Thread Processor) running Go v1.21+ on Windows 11 platform architecture targets.*

Scenario: Serializing a production-ready error or information event record injecting multiple data metadata metrics tokens (`String` tracking ID + `Int` transactional code values).

| Benchmark Operational Test Target | Execution Speed (ns/op) | Memory Allocated (B/op) | Heap Allocations (allocs/op) |
| :--- | :---: | :---: | :---: |
| **`BenchmarkZerologAdapter_FastPath`** | **171.6 ns/op** | **0 B/op** | **0 allocs/op** |
| **`BenchmarkZerologAdapter_ErrorPath`** | **197.4 ns/op** | **0 B/op** | **0 allocs/op** |
| **`BenchmarkSlogAdapter_FastPath`** | **626.3 ns/op** | **0 B/op** | **0 allocs/op** |
| `BenchmarkZerologAdapter_SlowPath` (with Caller) | 999.1 ns/op | 312 B/op | 3 allocs/op |
| `BenchmarkSlogAdapter_SlowPath` (with Source) | 1405.0 ns/op | 584 B/op | 6 allocs/op |

*Note: FastPath refers to logging operations executed with caller file/line tracking disabled. SlowPath explicitly enables OS frame trace allocations via `runtime.Caller` lookups.*

---

## 📜 License

This project is fully distributed and maintained under the loose guidelines of the **MIT License**. It guarantees complete legal usage freedoms, modification choices, and scaling implementation rights inside commercial or closed enterprise SaaS solutions without viral copyleft legal risks.
