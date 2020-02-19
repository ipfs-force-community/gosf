package jsonrpc

import (
	"context"
	"net/http"
)

// CtxKey context key type
type CtxKey struct {
	n string
}

// NewCtxKey returns a context key for context injects
func NewCtxKey(key string) *CtxKey {
	return &CtxKey{
		n: key,
	}
}

// Inject some values into the http.Request's context, and returns the new http.Request
func Inject(req *http.Request, key *CtxKey, value interface{}) *http.Request {
	ctx := context.WithValue(req.Context(), key, value)
	return req.WithContext(ctx)
}

// Extract extract value for the given key
func Extract(req *http.Request, key *CtxKey) interface{} {
	return req.Context().Value(key)
}
