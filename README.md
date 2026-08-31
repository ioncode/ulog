# ulog 🚀

Легковесный и гибкий пакет логирования, трассировки (Trace ID) и защиты от паник для микросервисов на Go. Построен по принципам чистой архитектуры и SOLID (Single Responsibility Principle) поверх высокопроизводительной библиотеки `zerolog`.

## 📦 Возможности
* **Полное соблюдение SRP**: Трассировка, метрики и безопасность разделены на независимые middleware.
* **Единый стандарт JSON**: Все ваши микросервисы будут писать логи с одинаковыми ключами (`trace_id`, `method`, `status`, `duration` и др.).
* **Интерфейсы Fluent API**: Слой бизнес-логики и HTTP-транспорта полностью отвязан от конкретной библиотеки логирования. Легко тестировать с помощью моков.
* **Автоматический кастомный вывод**: При разработке локально логер переключается на красивый, цветной текстовый вывод, а на Production гонит компактный структурированный JSON.

## 🛠 Установка

```bash
go get github.com/ioncode/ulog
```

## 🚀 Быстрый старт

### 1. Настройка в `main.go`

Инициализируйте базовый логер, оберните его в адаптер библиотеки `ulog` и подключите готовый защищенный пайплайн к вашему роутеру:

```go
package main

import (
	"io"
	"net/http"
	"os"

	"github.com/ioncode/ulog"
	"://github.com"
)

func main() {
	// 1. Настройка вывода (Local Text-Color / Production JSON)
	var logOutput io.Writer = os.Stdout
	if os.Getenv("APP_ENV") == "local" {
		logOutput = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"}
	}

	baseLog := zerolog.New(logOutput).With().Timestamp().Logger()

	// 2. Создаем адаптер ulog
	logAdapter := ulog.NewZerologAdapter(baseLog)

	// 3. Настраиваем роутер
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/ping", func(w http.ResponseWriter, r *http.Request) {
		traceID := ulog.GetTraceID(r.Context())
		
		// Логируем через адаптер с использованием стандартных констант
		logAdapter.Info().
			Str(ulog.LogKeyTraceID, traceID).
			Msg("ping endpoint called")
            
		w.Write([]byte(`{"status":"pong"}`))
	})

	// 4. Оборачиваем роутер в полный пайплайн (Trace ID -> Recovery -> Metrics)
	pipeline := ulog.NewHTTPPipeline(logAdapter)

	baseLog.Info().Msg("Server starting on :8080")
	_ = http.ListenAndServe(":8080", pipeline(mux))
}
```

### 2. Использование в слое хендлеров (Dependency Injection)

Чтобы сохранить архитектурную чистоту, ваши хендлеры должны объявлять интерфейс логера под свои нужды самостоятельно, используя типы из `ulog`:

```go
package handlers

import (
	"net/http"
	"://github.com"
)

// Объявляем интерфейс на стороне потребителя
type ComponentLogger interface {
	Info() ulog.LoggerEvent
	Error() ulog.LoggerEvent
}

type UserHandler struct {
	log ComponentLogger // Явная зависимость
}

func NewUserHandler(l ComponentLogger) *UserHandler {
	return &UserHandler{log: l}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	traceID := ulog.GetTraceID(r.Context())

	h.log.Info().
		Str(ulog.LogKeyTraceID, traceID).
		Int("user_id", 42).
		Msg("user created successfully")
        
	w.WriteHeader(http.StatusCreated)
}
```

### 3. Выборочное использование Middleware

Благодаря соблюдению SRP, вы можете отключать логирование метрик для технических эндпоинтов (например, для проверки здоровья k8s liveness probes), чтобы не спамить в логи, но сохранять защиту от паник:

```go
// Для технических роутов убираем метрики, оставляя только защиту от паник
healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
})

// Собираем вручную без RequestMetricsMiddleware
protectedHealth := ulog.RecoveryMiddleware(logAdapter)(healthHandler)
mux.Handle("/healthz", protectedHealth)
```

## ⚙️ Переменные окружения

* `APP_ENV=local` — включает цветной текстовый вывод в консоль (`ConsoleWriter`) с подсветкой синтаксиса.
* Любое другое значение или отсутствие переменной — включает стандартный быстрый JSON-поток для сбора в Grafana Loki, ELK, OpenSearch.

## 📊 Стандартные ключи логов (Константы)

Используйте встроенные константы, чтобы исключить риск опечаток в ключах поиска:
* `ulog.LogKeyTraceID` -> `"trace_id"`
* `ulog.LogKeyMethod` -> `"method"`
* `ulog.LogKeyPath` -> `"path"`
* `ulog.LogKeyStatus` -> `"status"`
* `ulog.LogKeyDuration` -> `"duration"` (значение логируется как `int` в миллисекундах)
