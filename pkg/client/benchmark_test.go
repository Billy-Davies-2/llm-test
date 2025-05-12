package client_test

import (
	"context"
	"net"
	"testing"
	"time"

	proto "github.com/Billy-Davies-2/llm-test/pkg/proto/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const bufSize = 1024 * 1024

// simple server that always returns a fixed MetricsResponse
type benchServer struct {
	proto.UnimplementedMetricsServiceServer
}

func (s *benchServer) GetMetrics(ctx context.Context, _ *emptypb.Empty) (*proto.MetricsResponse, error) {
	return &proto.MetricsResponse{
		HostId:          "bench-host",
		CpuUsagePercent: 50.0,
		MemoryUsedMb:    256,
		MemoryTotalMb:   512,
		Gpu: &proto.GPUInfo{
			Name:               "BenchGPU",
			TemperatureCelsius: 65.5,
		},
	}, nil
}

func dialBufConn(ctx context.Context, srv *grpc.Server) (*grpc.ClientConn, error) {
	lis := bufconn.Listen(bufSize)
	go srv.Serve(lis)
	cp := grpc.ConnectParams{Backoff: backoff.Config{BaseDelay: time.Millisecond, MaxDelay: time.Second}, MinConnectTimeout: time.Second}
	return grpc.NewClient(
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(cp),
	)
}

func BenchmarkFetchMetrics(b *testing.B) {
	// start in-memory server
	grpcServer := grpc.NewServer()
	proto.RegisterMetricsServiceServer(grpcServer, &benchServer{})

	ctx := context.Background()
	conn, err := dialBufConn(ctx, grpcServer)
	if err != nil {
		b.Fatalf("bufconn dial error: %v", err)
	}
	client := proto.NewMetricsServiceClient(conn)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := client.GetMetrics(ctx, &emptypb.Empty{})
		if err != nil {
			b.Errorf("GetMetrics error: %v", err)
		}
	}
}
