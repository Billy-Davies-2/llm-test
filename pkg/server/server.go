// pkg/server/server.go
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	chatpb "github.com/Billy-Davies-2/llm-test/pkg/proto/chat"
	metricspb "github.com/Billy-Davies-2/llm-test/pkg/proto/metrics"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Server wraps the gRPC server for metrics reporting
type Server struct {
	logger *slog.Logger
	hostID string
	port   int
	grpc   *grpc.Server
	chatpb.UnimplementedChatServiceServer
	metricspb.UnimplementedMetricsServiceServer
}

// NewServer constructs a metrics server for a given hostID and port
func NewServer(logger *slog.Logger, hostID string, port int) *Server {
	s := grpc.NewServer()
	impl := &metricsService{hostID: hostID}
	metricspb.RegisterMetricsServiceServer(s, impl)
	reflection.Register(s)
	return &Server{logger: logger, hostID: hostID, port: port, grpc: s}
}

// Run starts listening on the configured port and serves gRPC requests
func (s *Server) Run() error {
	addr := fmt.Sprintf(":%d", s.port)
	s.logger.Info("starting metrics server", "addr", addr)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %s: %w", addr, err)
	}
	if err := s.grpc.Serve(lis); err != nil {
		s.logger.Error("grpc serve failed", "err", err)
		return err
	}
	return nil
}

// metricsService implements the MetricsServiceServer interface
// backed by gopsutil for CPU and memory stats
type metricsService struct {
	metricspb.UnimplementedMetricsServiceServer
	hostID string
}

func (m *metricsService) GetMetrics(
	ctx context.Context,
	_ *emptypb.Empty,
) (*metricspb.MetricsResponse, error) {
	// CPU usage
	perc, err := cpu.Percent(0, false)
	if err != nil {
		return nil, err
	}
	cpuPct := 0.0
	if len(perc) > 0 {
		cpuPct = perc[0]
	}

	// Memory usage
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return &metricspb.MetricsResponse{
		HostId:          m.hostID,
		CpuUsagePercent: cpuPct,
		MemoryUsedMb:    float64(vm.Used) / 1024 / 1024,
		MemoryTotalMb:   float64(vm.Total) / 1024 / 1024,
	}, nil
}

// Chat implements metrics.ChatServiceServer.Chat
func (s *Server) Chat(ctx context.Context, req *chatpb.ChatRequest) (*chatpb.ChatResponse, error) {
	// Log which server handled it and what was asked
	s.logger.Info("Chat request",
		"host", s.hostID,
		"prompt", req.GetText(),
	)

	// For now: echo a canned AI reply
	reply := "🤖 This is a canned AI response."
	return &chatpb.ChatResponse{
		HostId: s.hostID,
		Text:   reply,
	}, nil
}
