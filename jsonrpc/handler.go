package jsonrpc

import (
	"net/http"
)

type patternedHandler struct {
	pattern string
	handler HandlerFunc
}

// HandlerFunc is like http.HandlerFunc, but returns an error
type HandlerFunc = func(rw http.ResponseWriter, req *http.Request) error
