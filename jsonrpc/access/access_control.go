// Package access provides access control utils for json rpc
package access

import (
	"context"
	"net/http"

	"gitlab.forceup.in/dev-go/gosf/jsonrpc"
	"gitlab.forceup.in/dev-go/gosf/unsafe"
	"gitlab.forceup.in/dev-proto/common"
)

const (
	authorizationHeaderKey = "Authorization"
)

var (
	ctxKeyAccessPermsFetcher = jsonrpc.NewCtxKey("_acc_fetcher")
	ctxKeyAccessPerms        = jsonrpc.NewCtxKey("_acc_perms")
)

// Fetcher represents a method for converting token to common.AccessPerms
type Fetcher interface {
	Fetch(ctx context.Context, token string) (*common.AccessPerms, error)
}

// InjectPermsFetcher injects given fetcher into request's context
func InjectPermsFetcher(f Fetcher) jsonrpc.Middleware {

	return func(inner jsonrpc.HandlerFunc) jsonrpc.HandlerFunc {

		return func(rw http.ResponseWriter, req *http.Request) error {
			req = jsonrpc.Inject(req, ctxKeyAccessPermsFetcher, f)

			return inner(rw, req)
		}
	}
}

// ExtractPermsFetcher extracts fetcher from the request's context
func ExtractPermsFetcher(req *http.Request) (Fetcher, bool) {
	f, ok := req.Context().Value(ctxKeyAccessPermsFetcher).(Fetcher)
	return f, ok
}

// CheckAndInjectAccessPerms fetchs perms and injects them into the request
func CheckAndInjectAccessPerms(req *http.Request, scope string, required common.Perm) (*http.Request, bool) {
	var perms *common.AccessPerms

	if token := req.Header.Get(authorizationHeaderKey); token != "" {
		logger := jsonrpc.RequestLogger(req)

		fetcher, _ := ExtractPermsFetcher(req)
		if fetcher == nil {
			logger.Warn("no available access perms fetcher")
			goto AFTER
		}

		p, err := fetcher.Fetch(req.Context(), token)
		if err != nil {
			logger.Warnf("error captured for fetching perms, token=%q, cause=%v", token, err)
			goto AFTER
		}

		perms = p
	}

AFTER:
	req = jsonrpc.Inject(req, ctxKeyAccessPerms, perms)
	return req, CheckPerms(perms, scope, required)
}

// ExtractPerms extracts access perms from request context
func ExtractPerms(req *http.Request) (*common.AccessPerms, bool) {
	p, _ := jsonrpc.Extract(req, ctxKeyAccessPerms).(*common.AccessPerms)
	return p, p != nil
}

// ExtractPermsFromCtx extracts access perms from request context
func ExtractPermsFromCtx(ctx context.Context) (*common.AccessPerms, bool) {
	p, _ := ctx.Value(ctxKeyAccessPerms).(*common.AccessPerms)
	return p, p != nil
}

// CheckPerms checks if the required perms under the given scope is satisfied
func CheckPerms(perms *common.AccessPerms, scope string, required common.Perm) bool {
	if len(scope) == 0 || perms == nil || len(perms.Perms) == 0 {
		return false
	}

	perm, ok := perms.Perms[scope]
	if !ok {
		b := unsafe.Bytes(scope)
		for size := len(b) - 1; size > 0; size-- {
			if b[size-1] != '.' {
				continue
			}

			if p, ok := perms.Perms[unsafe.String(b[:size-1])]; ok {
				perm = p
				break
			}
		}
	}

	return (perm & required) == required
}
