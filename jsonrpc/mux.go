package jsonrpc

import (
	"net/http"

	"gitlab.forceup.in/dev-go/gosf/proc"
)

// RegisterMux register a jsonrpc mux onto the given std *http.ServeMux, and uses http.DefaultServeMux if stdmux is nil
func RegisterMux(stdmux *http.ServeMux, jmux *Mux) {
	if stdmux == nil {
		stdmux = http.DefaultServeMux
	}

	jmux.register("", nil, stdmux)

	proc.RegisterVersionHandler(stdmux)
}

// NewMux returns a json mux with given prefix, logger & middlewares
func NewMux(prefix string, logger Logger, mds ...Middleware) *Mux {
	return &Mux{
		prefix:   prefix,
		handlers: []patternedHandler{},
		midwares: mds,
		subs:     []*Mux{},
		logger:   logger,
	}
}

// NewRootMux returns a root json mux, with default middwares
func NewRootMux(prefix string, logger Logger) *Mux {
	mds := []Middleware{
		InjectRequestLogger(logger),
		InjectRequestID(),
		HandleRequestInfoLogging(),
		HandleCORS(),
		HandleError(),
		HandlePanic(),
	}

	return NewMux(prefix, logger, mds...)
}

// Mux represents a simple multiplexer for jsonrpc requests
type Mux struct {
	prefix   string
	handlers []patternedHandler
	midwares []Middleware
	subs     []*Mux
	logger   Logger
}

// Use appends a group of middlewares to the current mux
func (m *Mux) Use(mw ...Middleware) {
	m.midwares = append(m.midwares, mw...)
}

// Handle register a handler func for the given pattern
func (m *Mux) Handle(pattern string, handler HandlerFunc) {
	m.handlers = append(m.handlers, patternedHandler{
		pattern: pattern,
		handler: handler,
	})
}

// AddSubs appends sub muxes to the current mux
func (m *Mux) AddSubs(subs ...*Mux) {
	m.subs = append(m.subs, subs...)
}

func (m *Mux) register(prefix string, mds []Middleware, stdmux *http.ServeMux) {
	logger := m.logger
	if logger == nil {
		logger = stdLogger
	}

	prefix += m.prefix
	if prefix == "/" {
		prefix = ""
	}

	mds = append(mds, m.midwares...)

	for _, patternedHdl := range m.handlers {

		wrappedHdl := patternedHdl.handler
		for size := len(mds); size > 0; size-- {
			wrappedHdl = mds[size-1](wrappedHdl)
		}

		stdmux.Handle(prefix+patternedHdl.pattern, http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			err := wrappedHdl(rw, req)
			if err != nil {
				logger.Errorf("unhandled error captured, method=%s, cause=%s", req.RequestURI, err.Error())
			}
		}))
	}

	for _, sub := range m.subs {
		sub.register(prefix, mds, stdmux)
	}
}
