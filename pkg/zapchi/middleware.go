package zapchi

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

var _ middleware.LogFormatter = Logger{}
var _ middleware.LogEntry = loggerEntry{}

type Logger struct {
	inner *zap.Logger
}

func UseLogger(logger *zap.Logger) Logger {
	return Logger{inner: logger}
}

func (l Logger) NewLogEntry(r *http.Request) middleware.LogEntry {
	return loggerEntry{
		Logger: l,
		method: r.Method,
		path:   r.URL.Path,
	}
}

type loggerEntry struct {
	Logger

	method string
	path   string
}

func (l loggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	l.inner.Info("Request completed",
		zap.Int("status", status),
		zap.Int("bytes", bytes),
		zap.Duration("elapsed", elapsed),
		zap.String("method", l.method),
		zap.String("path", l.path))
}

func (l loggerEntry) Panic(v interface{}, stack []byte) {
	l.inner.Fatal("Request panicked", zap.Any("panic", v))
}
