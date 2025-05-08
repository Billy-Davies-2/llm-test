// pkg/client/client_test.go

package client_test

import (
	"context"
	"log/slog"
	"net"
	"testing"
	"time"

	"github.com/Billy-Davies-2/llm-test/pkg/client"
	"github.com/Billy-Davies-2/llm-test/pkg/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// successServer implements MetricsService and always returns fixed metrics.
type successServer struct {
	proto.UnimplementedMetricsServiceServer
}

func (s *successServer) GetMetrics(ctx context.Context, _ *emptypb.Empty) (*proto.MetricsResponse, error) {
	return &proto.MetricsResponse{
		HostId:          "test-host",
		CpuUsagePercent: 42.5,
		MemoryUsedMb:    512,
		MemoryTotalMb:   1024,
		Gpu: &proto.GPUInfo{
			Name:               "TestGPU",
			TemperatureCelsius: 70.0,
		},
	}, nil
}

// errorServer always returns an internal error.
type errorServer struct {
	proto.UnimplementedMetricsServiceServer
}

func (s *errorServer) GetMetrics(ctx context.Context, _ *emptypb.Empty) (*proto.MetricsResponse, error) {
	return nil, status.Error(codes.Internal, "server-side failure")
}

// startTestServer spins up an in-process gRPC server on a random TCP port.
func startTestServer(t *testing.T, srv proto.MetricsServiceServer) (addr string, teardown func()) {
	lis, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listener error: %v", err)
	}
	grpcServer := grpc.NewServer()
	proto.RegisterMetricsServiceServer(grpcServer, srv)
	go grpcServer.Serve(lis)
	return lis.Addr().String(), func() {
		grpcServer.Stop()
		lis.Close()
	}
}

func TestFetchMetrics_Success(t *testing.T) {
	addr, stop := startTestServer(t, &successServer{})
	defer stop()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cli, err := client.NewClient(ctx, addr, &slog.Logger{})
	if err != nil {
		t.Fatalf("NewClient(): unexpected error: %v", err)
	}
	defer cli.Close()

	m, err := cli.FetchMetrics(ctx)
	if err != nil {
		t.Fatalf("FetchMetrics(): unexpected error: %v", err)
	}

	// verify fields
	if got, want := m.HostID, "test-host"; got != want {
		t.Errorf("HostID = %q; want %q", got, want)
	}
	if got, want := m.CPUUsagePct, 42.5; got != want {
		t.Errorf("CPUUsagePct = %v; want %v", got, want)
	}
	if got, want := m.MemoryUsedMB, 512.0; got != want {
		t.Errorf("MemoryUsedMB = %v; want %v", got, want)
	}
	if got, want := m.MemoryTotalMB, 1024.0; got != want {
		t.Errorf("MemoryTotalMB = %v; want %v", got, want)
	}
	if got, want := m.GPUName, "TestGPU"; got != want {
		t.Errorf("GPUName = %q; want %q", got, want)
	}
	if got, want := m.GPUTempCelsius, 70.0; got != want {
		t.Errorf("GPUTempCelsius = %v; want %v", got, want)
	}
}

func TestFetchMetrics_ServerError(t *testing.T) {
	addr, stop := startTestServer(t, &errorServer{})
	defer stop()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cli, err := client.NewClient(ctx, addr, &slog.Logger{})
	if err != nil {
		t.Fatalf("NewClient(): unexpected error: %v", err)
	}
	defer cli.Close()

	_, err = cli.FetchMetrics(ctx)
	if err == nil {
		t.Fatalf("FetchMetrics() expected error, got nil")
	}
	st, ok := status.FromError(err)
	if !ok {
		t.Fatalf("FetchMetrics() returned non-gRPC error: %v", err)
	}
	if st.Code() != codes.Internal {
		t.Errorf("FetchMetrics() error code = %v; want %v", st.Code(), codes.Internal)
	}
	if st.Message() != "server-side failure" {
		t.Errorf("FetchMetrics() error message = %q; want %q", st.Message(), "server-side failure")
	}
}
