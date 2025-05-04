// Package proto holds your .proto definitions and generated code.
//go:generate protoc  -I . --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative metrics.proto

package proto
