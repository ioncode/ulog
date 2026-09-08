# ulog 🪵

[![Go Reference](https://go.dev)](https://go.dev)
[![Go Report Card](https://goreportcard.com)](https://goreportcard.com)
[![License: AGPL v3](https://shields.io)](https://gnu.org)

`ulog` is an architectural wrapper and production-ready HTTP middleware pipeline for structured logging in Go. It is designed to fully adhere to the **Single Responsibility Principle (SRP)** and **Clean Architecture** guidelines.

By decoupling the logging interfaces from actual underlying engines, `ulog` allows you to switch your logging backbone seamlessly without touching your core HTTP pipelines or application service layers.

---

## ✨ Features

- **Multi-Engine Support:** Switch between `zerolog` and standard `log/slog` on the fly.
- **Zero-Allocation Ready:** Retains the blistering speed of `zerolog` when using the appropriate adapter.
- **Preserved Caller Depth:** File names, functions, and line numbers (`caller` or `source`) point to your actual handlers, not to the internal wrapper files.
- **Standardized JSON Schema:** Automatically injects predefined corporate fields (`trace_id`, `status`, `duration`, etc.) to match modern observability stack standards (Grafana Loki, Datadog, Kibana).
- **Decoupled Middleware:** Production-grade building blocks out-of-the-box (`TraceID`, `Recovery`, `Metrics`).

---

## 🪵 Multi-Engine Logging Support

`ulog` provides a polymorphic, clean interface that wraps external loggers. You can choose between:
1. **Zerolog Engine:** Best for ultra-fast, zero-allocation structured JSON logging in high-concurrency apps.
2. **Slog Engine:** Best for standard library compliance (Go 1.21+) and native ecosystem compatibility.

---

## 🚀 Quick Start

### Installation

```bash
go get github.com/ioncode/ulog/v2
```

### Option A: Using Standard `log/slog` (Go 1.21+)

If your enterprise project standardizes on Go's built-in structured logger, wrap it using `NewSlogAdapter`:

```go
package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/ioncode/ulog/v2"
)

func main() {
	// 1. Initialize native slog handler
	baseSlog := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true, // ulog correctly preserves caller depth/file source!
	}))

	// 2. Wrap it into ulog interface
	logger := ulog.NewSlogAdapter(baseSlog)

	// 3. Inject it into your HTTP pipeline
	pipeline := ulog.NewHTTPPipeline(logger)

	// Create a dummy multiplexer
	mux := http.NewServeMux()
	mux.HandleFunc("/api/hello", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	// Wrap server endpoints into ulog middleware pipeline
	http.ListenAndServe(":8080", pipeline(mux))
}
```

### Option B: Using `zerolog` (Maximum Performance)

If you need maximum throughput and zero memory allocations under heavy cloud workloads:

```go
package main

import (
	"net/http"
	"os"

	"github.com/ioncode/ulog/v2"
	"github.com/rs/zerolog"
)

func main() {
	// 1. Initialize zerolog
	baseZerolog := zerolog.New(os.Stdout).With().Timestamp().Logger()

	// 2. Wrap it into ulog interface
	logger := ulog.NewZerologAdapter(baseZerolog)

	// 3. Inject into HTTP pipeline
	pipeline := ulog.NewHTTPPipeline(logger)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	http.ListenAndServe(":8080", pipeline(mux))
}
```

---

## ⚡ Performance & Benchmarks

The library is benchmarks-tested on `Windows / amd64` (AMD Ryzen 5 5600X). `ulog` utilizes a smart hybrid **Fast Path** feature: if your base logger is initialized without caller/source file tracking, it bypasses expensive `runtime` stack allocation layers completely.

| Benchmark Scenario | Time per Op | Memory Allocs | Bytes Allocated | Best For |
| :--- | :--- | :--- | :--- | :--- |
| **`ZerologAdapter (FastPath)`** | **~266 ns/op** | **1 allocs/op** | **8 B/op** | Ultra-high throughput, production microservices, maximum RPS |
| **`SlogAdapter (FastPath)`** | ~982 ns/op | 7 allocs/op | 304 B/op | Standard library compliance, modern Go ecosystems |
| **`SlogAdapter (With Source)`** | ~1780 ns/op | 13 allocs/op | 888 B/op | Debugging environments where line-numbers are critical (Slog) |
| **`ZerologAdapter (With Caller)`** | ~2092 ns/op | 8 allocs/op | 608 B/op | Debugging environments where line-numbers are critical (Zerolog) |

*You can replicate these results locally by running `go test -bench=. -benchmem ./...`*

### Architectural Trade-Off
Introducing a clean polymorphic layer (`ulog.LoggerEvent`) causes intermediate fluent-chaining structures to escape to the heap due to interface wrapping boundaries. For `ZerologAdapter`, this overhead is a mere 8 bytes (1 allocation). However, when a caller tracking is requested, the application performs native stack tracing (~1700-2000ns per operation). The smart routing feature saves up to 80% of CPU time by dynamically choosing the optimal path based on your parent logger setup.


---

## 🏗️ Architecture Design

`ulog` embraces clean design boundaries:

1. **Fluent API Interfaces (`ulog.Logger` & `ulog.LoggerEvent`):** Your business logic layer interacts only with abstract primitives. Writing mock tests for loggers becomes incredibly straightforward.
2. **SRP Middleware Layer:** 
    - `TraceIDMiddleware` generates or extracts distributed tracing headers.
    - `RecoveryMiddleware` isolates thread panics, writes defensive status headers, and logs clean error stack traces.
    - `MetricsMiddleware` computes server response times accurately.
3. **Adapter Decoupling:** Third-party engine drivers live in specialized isolation and do not leak internal signatures into server logic.

---

## 📄 License

This project is licensed under the GNU Affero General Public License v3 - see the [LICENSE](LICENSE) file for details.

