package jsonrpc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/golang/protobuf/proto"
)

// NewRPCClient 创建 rpc 客户端
func NewRPCClient(host string, rt *http.Client) *RPCClient {
	if rt == nil {
		rt = http.DefaultClient
	}

	return &RPCClient{
		host:    host,
		httpcli: rt,
	}
}

// RPCClient jsonrpc client based on http1.1
type RPCClient struct {
	host    string
	httpcli *http.Client
}

// Call calls specified method with given data & response receiver
func (rc *RPCClient) Call(ctx context.Context, method string, data, recv proto.Message) error {
	var reqBody io.Reader

	if data != nil {
		buf := bytes.NewBufferString("")
		if err := EncodeJSON(buf, data); err != nil {
			return fmt.Errorf("unable to marshal request, err=%v", err)
		}

		reqBody = buf
	}

	req, err := http.NewRequest(http.MethodPost, rc.host+method, reqBody)
	if err != nil {
		return fmt.Errorf("unable to build http request, err=%v", err)
	}

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	resp, err := rc.httpcli.Do(req)
	if err != nil {
		return fmt.Errorf("unable to send http post request, err=%v", err)
	}

	defer resp.Body.Close()

	if recv != nil {
		if err := DecodeJSON(resp.Body, recv); err != nil {
			return fmt.Errorf("unable to unmarshal response body, err=%v", err)
		}
	}

	return nil
}
