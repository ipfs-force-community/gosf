plugin:
	go install github.com/ipfs-force-community/gosf/protoc-gen-force-jsonrpc

install: plugin
	go install ./...
