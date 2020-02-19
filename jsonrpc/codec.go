package jsonrpc

import (
	"io"
	"net/http"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

var (
	jsonMarshaler = jsonpb.Marshaler{
		EmitDefaults: true,
		OrigName:     true,
	}

	jsonUnmarshaler = jsonpb.Unmarshaler{
		AllowUnknownFields: true,
	}

	jsonStrictUnmarshaler = jsonpb.Unmarshaler{
		AllowUnknownFields: false,
	}
)

// DecodeRequest decodes given request's body using json format
func DecodeRequest(req *http.Request, recv proto.Message) error {
	defer req.Body.Close()

	return jsonUnmarshaler.Unmarshal(req.Body, recv)
}

// EncodeResponse encodes data and write to response body
func EncodeResponse(rw http.ResponseWriter, data proto.Message) error {
	return jsonMarshaler.Marshal(rw, data)
}

// EncodeJSON encode given data into the writer
func EncodeJSON(w io.Writer, data proto.Message) error {
	return jsonMarshaler.Marshal(w, data)
}

// DecodeJSON decode bytes to the recv
func DecodeJSON(r io.Reader, recv proto.Message) error {
	return jsonUnmarshaler.Unmarshal(r, recv)
}

// DecodeJSONStrict use strict mode to unmarshal message, which won't allow unknown fields
func DecodeJSONStrict(r io.Reader, recv proto.Message) error {
	return jsonStrictUnmarshaler.Unmarshal(r, recv)
}
