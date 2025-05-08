// pkg/server/server.go
package server

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"github.com/Billy-Davies-2/llm-test/pkg/proto"
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
}

// NewServer constructs a metrics server for a given hostID and port
func NewServer(logger *slog.Logger, hostID string, port int) *Server {
	s := grpc.NewServer()
	impl := &metricsService{hostID: hostID}
	proto.RegisterMetricsServiceServer(s, impl)
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
	proto.UnimplementedMetricsServiceServer
	hostID string
}

func (m *metricsService) GetMetrics(
	ctx context.Context,
	_ *emptypb.Empty,
) (*proto.MetricsResponse, error) {
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

	return &proto.MetricsResponse{
		HostId:          m.hostID,
		CpuUsagePercent: cpuPct,
		MemoryUsedMb:    float64(vm.Used) / 1024 / 1024,
		MemoryTotalMb:   float64(vm.Total) / 1024 / 1024,
	}, nil
}
