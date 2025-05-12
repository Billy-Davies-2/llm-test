.PHONY: fmt lint gen-protos test build

fmt:
\tgo fmt ./...

lint:
\tgolangci-lint run

gen-protos:
\t./scripts/gen-protos.sh

test:
\tgo test ./pkg/... ./cmd/...
\tgo test ./integration_tests

build: gen-protos
\tgo build -o bin/llm-backend ./cmd/metrics-server
\tgo build -o bin/llm-tui     ./cmd/tui-client
