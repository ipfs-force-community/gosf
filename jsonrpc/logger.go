package jsonrpc

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

// Logger is the type alias for *zap.SugaredLogger
type Logger = *zap.SugaredLogger

var stdLogger = zap.S()
var ctxKeyLogger = NewCtxKey("_logger")

// RequestLogger extracts logger instance from request context
func RequestLogger(req *http.Request) Logger {
	if l, _ := req.Context().Value(ctxKeyLogger).(*zap.SugaredLogger); l != nil {
		return l
	}

	return stdLogger
}

// RequestLoggerFromCtx extracts logger instance from request context
func RequestLoggerFromCtx(ctx context.Context) Logger {
	if l, _ := ctx.Value(ctxKeyLogger).(*zap.SugaredLogger); l != nil {
		return l
	}

	return stdLogger
}

// InjectRequestLogger injects given Logger instance into the request's context
func InjectRequestLogger(l Logger) Middleware {
	return func(inner HandlerFunc) HandlerFunc {
		return func(rw http.ResponseWriter, req *http.Request) error {
			if l != nil {
				req = Inject(req, ctxKeyLogger, l)
			}

			return inner(rw, req)
		}
	}
}
