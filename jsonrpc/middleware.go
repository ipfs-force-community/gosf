package jsonrpc

import (
	"fmt"
	"net/http"

	"gitlab.forceup.in/dev-proto/common"
)

// Middleware defines a simple middleware for jsonrpc
type Middleware func(HandlerFunc) HandlerFunc

// HandleError wraps inner HandlerFunc with error handler
func HandleError() Middleware {
	return func(inner HandlerFunc) HandlerFunc {

		return func(rw http.ResponseWriter, req *http.Request) error {
			err := inner(rw, req)
			if err == nil {
				return nil
			}

			resp := common.SimpleResp{}

			switch e := err.(type) {
			case *RPCError:
				resp.Res = common.NewResult(int32(e.Code), e.Msg)

			default:
				resp.Res = common.NewResult(int32(http.StatusInternalServerError), e.Error())

			}

			if err := EncodeResponse(rw, &resp); err != nil {
				RequestLogger(req).Errorf("error occurs during encoding captured inner err, req_id=%s, resp=%v", RequestID(req), resp)
			}

			return nil
		}
	}
}

// HandlePanic wraps inner HandlerFunc with panic handler
func HandlePanic() Middleware {
	return func(inner HandlerFunc) HandlerFunc {

		return func(rw http.ResponseWriter, req *http.Request) (err error) {
			defer func() {
				if p := recover(); p != nil {
					reqID := RequestID(req)
					RequestLogger(req).Errorf("recover from panic, method=%s, req_id=%q, cause=%v", req.URL.String(), reqID, p)

					switch e := p.(type) {
					case error:
						err = e

					default:
						err = &RPCError{
							Code: http.StatusInternalServerError,
							Msg:  fmt.Sprintf("recover from internal panic, req_id=%q", reqID),
						}
					}
				}
			}()

			err = inner(rw, req)

			return err
		}

	}
}
