#!/usr/bin/env bash
protoc -I . --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/proto/chat/chat.proto
protoc -I . --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative pkg/proto/metrics/metrics.proto
