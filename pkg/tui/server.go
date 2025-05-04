package tui

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"github.com/Billy-Davies-2/tui-chat/pkg/proto"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

// metricsServer implements the MetricsServiceServer interface
type metricsServer struct {
	proto.UnimplementedMetricsServiceServer
	hostID string
}

// GetMetrics returns current CPU, memory and optional GPU metrics
func (s *metricsServer) GetMetrics(ctx context.Context, _ *proto.Empty) (*proto.MetricsResponse, error) {
	// CPU usage percent
	cpuPercents, err := cpu.Percent(0, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get CPU percent: %w", err)
	}
	cpuPct := 0.0
	if len(cpuPercents) > 0 {
		cpuPct = cpuPercents[0]
	}

	// Memory usage
	vm, err := mem.VirtualMemory()
	if err != nil {
		return nil, fmt.Errorf("failed to get virtual memory info: %w", err)
	}

	resp := &proto.MetricsResponse{
		HostId:          s.hostID,
		CpuUsagePercent: cpuPct,
		MemoryUsedMb:    float64(vm.Used) / 1024 / 1024,
		MemoryTotalMb:   float64(vm.Total) / 1024 / 1024,
	}

	// TODO: add GPU info via NVML if desired

	return resp, nil
}

func main() {
	// command-line flags
	port := flag.Int("port", 50051, "The server port")
	hostID := flag.String("host-id", "", "Unique host identifier")
	flag.Parse()

	if *hostID == "" {
		log.Fatal("--host-id must be set to a non-empty value")
	}

	// start listening
	addr := fmt.Sprintf(":%d", *port)
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen on %s: %v", addr, err)
	}

	// create gRPC server
	grpcServer := grpc.NewServer()

	// register our service
	proto.RegisterMetricsServiceServer(grpcServer, &metricsServer{hostID: *hostID})

	// register reflection service on gRPC server
	reflection.Register(grpcServer)

	log.Printf("MetricsService gRPC server listening on %s", addr)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve gRPC: %v", err)
	}
}
