package jsonrpc

import (
	"net/http"
	"strings"
)

const (
	corsHeaderAllowOrigin  = "Access-Control-Allow-Origin"
	corsHeaderAllowMethods = "Access-Control-Allow-Methods"
	corsHeaderAllowHeaders = "Access-Control-Allow-Headers"

	corsAllowedHadersBase = "Keep-Alive, User-Agent, Content-Type, Authorization"
)

var (
	customizeCORSHeaders = []string{RequestIDHeader}

	corsAllowedHaders = strings.Join(
		[]string{
			corsAllowedHadersBase,
			RequestIDHeader,
		},

		", ",
	)
)

// AddCustomizeCORSHeader 添加自定义 CORS 头
func AddCustomizeCORSHeader(header ...string) {
	customizeCORSHeaders = append(customizeCORSHeaders, header...)
	corsAllowedHaders = strings.Join(append([]string{corsAllowedHadersBase}, customizeCORSHeaders...), ", ")
}

// ApplyCORSHeaders add cors headers to the given rw
func ApplyCORSHeaders(header http.Header) {
	header.Set(corsHeaderAllowOrigin, "*")
	header.Set(corsHeaderAllowMethods, "OPTIONS, GET, POST")
	header.Set(corsHeaderAllowHeaders, corsAllowedHaders)
}

// HandleCORS handles cors OPTIONS requests, and set required headers for any other request
func HandleCORS() Middleware {
	return func(inner HandlerFunc) HandlerFunc {

		return func(rw http.ResponseWriter, req *http.Request) error {
			ApplyCORSHeaders(rw.Header())

			if req.Method == http.MethodOptions {
				rw.WriteHeader(http.StatusOK)
				return nil
			}

			return inner(rw, req)
		}
	}
}
