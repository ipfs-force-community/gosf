plugin:
	go install gitlab.forceup.in/dev-go/gosf/protoc-gen-force-jsonrpc

install: plugin
	go install ./...
