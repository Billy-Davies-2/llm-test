.PHONY: fmt lint gen-protos test build

fmt:
	go fmt ./...

lint:
	golangci-lint run

gen-protos:
	./scripts/gen-protos.sh

test:
	go test ./pkg/... ./cmd/...
	go test ./integration_tests

build: gen-protos
	go build -o bin/llm-backend ./cmd/metrics-server
	go build -o bin/llm-tui     ./cmd/tui-client
