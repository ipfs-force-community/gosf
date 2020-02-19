package jsonrpc

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"time"
)

func init() {
	if machineID != "" {
		return
	}

	b := [8]byte{}
	rand.Read(b[:])

	machineID = base64.URLEncoding.EncodeToString(b[:])
}

// UnixNano for 2019-01-01 00:00:00.0000Z+08:00
const reqIDEpoch int64 = 1546272000000000000

// RequestIDHeader http header name for req id
const RequestIDHeader = "X-FORCEUP-REQ-ID"

var (
	machineID       = os.Getenv(envMachineIDKey)
	envMachineIDKey = "FORCEUP_MACHINE_ID"
	ctxKeyReqID     = NewCtxKey("_req_id")

	ctxKeyReq = NewCtxKey("_request")
)

func genRequestID() string {
	b := [8]byte{}
	binary.BigEndian.PutUint64(b[:], uint64(time.Now().UnixNano()-reqIDEpoch))
	return fmt.Sprintf("%s:%s", base64.URLEncoding.EncodeToString(b[:]), machineID)
}

// InjectRequestID injects an id for each request
func InjectRequestID() Middleware {
	return func(inner HandlerFunc) HandlerFunc {
		return func(rw http.ResponseWriter, req *http.Request) error {
			reqID := genRequestID()
			req = Inject(req, ctxKeyReqID, reqID)

			rw.Header().Set(RequestIDHeader, genRequestID())

			return inner(rw, req)
		}
	}
}

// RequestID extracts request id from context
func RequestID(req *http.Request) string {
	id, _ := Extract(req, ctxKeyReqID).(string)
	if id != "" {
		return id
	}

	return genRequestID()
}

// InjectHTTPRequest injects given *http.Request into context
func InjectHTTPRequest(req *http.Request) *http.Request {
	pureReq := req.WithContext(context.Background())
	return Inject(req, ctxKeyReq, pureReq)
}

// ExtractHTTPRequestFromCtx extract *http.Request from given context
func ExtractHTTPRequestFromCtx(ctx context.Context) (*http.Request, bool) {
	req, _ := ctx.Value(ctxKeyReq).(*http.Request)
	return req, req != nil
}
