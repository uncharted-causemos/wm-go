package middleware

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

// zapRequestLogger implements middleware.LoggerInterface
type zapRequestLogger struct {
	*zap.Logger
}

func (z *zapRequestLogger) Print(v ...interface{}) {
	z.Info(fmt.Sprint(v...))
}

// NewZapRequestLogger creates request logging middleware using zap logger
// Ref: https://github.com/go-chi/chi/blob/df44563f0692b1e677f18220b9be165e481cf51b/middleware/logger.go
func NewZapRequestLogger(logger *zap.Logger, color bool) func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: &zapRequestLogger{logger}, NoColor: !color})
}
