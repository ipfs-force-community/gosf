package jsonrpc

import (
	"net/http"
	"time"
)

// HandleRequestInfoLogging provides INFO level loggin for a request, including method, uri & elapsed time
func HandleRequestInfoLogging() Middleware {
	return func(inner HandlerFunc) HandlerFunc {

		return func(rw http.ResponseWriter, req *http.Request) error {
			before := time.Now()

			wrapped := &wrappedResponseWritter{
				inner: rw,
			}

			err := inner(wrapped, req)
			if wrapped.code == 0 && err == nil {
				wrapped.code = 200
			}

			dur := time.Since(before)

			// logging
			RequestLogger(req).Infof("[%d][%s] %s %s", wrapped.code, req.Method, req.RequestURI, dur)

			// rpc metric
			rpcMetricAdd(req.URL.Path, wrapped.code, dur)
			return err
		}
	}
}

var (
	_ http.ResponseWriter = (*wrappedResponseWritter)(nil)
)

type wrappedResponseWritter struct {
	code        int
	codeWritten bool
	inner       http.ResponseWriter
}

func (wrw *wrappedResponseWritter) Header() http.Header {
	return wrw.inner.Header()
}

func (wrw *wrappedResponseWritter) Write(b []byte) (int, error) {
	return wrw.inner.Write(b)
}

func (wrw *wrappedResponseWritter) WriteHeader(code int) {
	if !wrw.codeWritten {
		wrw.inner.WriteHeader(code)
		wrw.codeWritten = true
	}
}
