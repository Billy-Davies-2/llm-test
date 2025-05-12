package client

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"

	proto "github.com/Billy-Davies-2/llm-test/pkg/proto/metrics"
)

// Metrics holds the domain-friendly view of MetricsResponse.
type Metrics struct {
	HostID         string
	CPUUsagePct    float64
	MemoryUsedMB   float64
	MemoryTotalMB  float64
	GPUName        string  // empty if no GPU
	GPUTempCelsius float64 // zero if no GPU
}

// Client wraps the gRPC stub.
type Client struct {
	logger *slog.Logger
	stub   proto.MetricsServiceClient
	conn   *grpc.ClientConn
}

// NewClient dials the server at addr (e.g. "host:50051").
// It will retry for up to 5 seconds if the connection isn’t ready.
func NewClient(ctx context.Context, addr string, logger *slog.Logger) (*Client, error) {
	logger.Debug("dialing metrics server", "addr", addr)
	cp := grpc.ConnectParams{
		Backoff: backoff.Config{
			BaseDelay:  100 * time.Millisecond,
			Multiplier: 1.6,
			Jitter:     0.2,
			MaxDelay:   1 * time.Second,
		},
		MinConnectTimeout: 5 * time.Second,
	}

	// grpc.NewClient is the new non-deprecated dialer
	cc, err := grpc.NewClient(
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithConnectParams(cp),
	)
	if err != nil {
		logger.Error("failed to create grpc client", "err", err)
		return nil, err
	}

	// block until connected or context expires
	cc.Connect()

	return &Client{
		stub:   proto.NewMetricsServiceClient(cc),
		conn:   cc,
		logger: logger,
	}, nil
}

// FetchMetrics does a unary GetMetrics call.
func (c *Client) FetchMetrics(ctx context.Context) (*Metrics, error) {
	resp, err := c.stub.GetMetrics(ctx, &emptypb.Empty{})
	if err != nil {
		c.logger.Warn("FetchMetrics RPC failed", "err", err)
		return nil, err
	}
	m := &Metrics{
		HostID:        resp.GetHostId(),
		CPUUsagePct:   resp.GetCpuUsagePercent(),
		MemoryUsedMB:  resp.GetMemoryUsedMb(),
		MemoryTotalMB: resp.GetMemoryTotalMb(),
	}
	c.logger.Debug("fetched metrics",
		"host", m.HostID,
		"cpu", m.CPUUsagePct,
		"ram_used", m.MemoryUsedMB,
		"ram_total", m.MemoryTotalMB,
	)
	if g := resp.GetGpu(); g != nil {
		m.GPUName = g.GetName()
		m.GPUTempCelsius = g.GetTemperatureCelsius()
	}
	return m, nil
}

func (c *Client) Close() error {
	return c.conn.Close()
}
