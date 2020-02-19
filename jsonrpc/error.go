package jsonrpc

import (
	"fmt"
	"net/http"
)

var _ error = (*RPCError)(nil)

// RPCError represents a specific rpc call error
type RPCError struct {
	Code int
	Msg  string
}

func (r *RPCError) Error() string {
	return fmt.Sprintf("json rpc error: code=%d, msg=%s", r.Code, r.Msg)
}

// NewRPCErrorWithCode returns a *RPCError wraps the given http StatusCode
func NewRPCErrorWithCode(code int, msg ...string) *RPCError {
	e := &RPCError{
		Code: code,
	}

	if len(msg) > 0 {
		e.Msg = msg[0]
	}

	if len(e.Msg) == 0 {
		e.Msg = http.StatusText(code)
	}

	return e
}
