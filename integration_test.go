package main_test

import (
	"bufio"
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/Billy-Davies-2/llm-test/pkg/client"
)

func TestServerClientEndToEnd(t *testing.T) {
	// 1) build and launch the server binary on an ephemeral port
	if out, err := exec.Command("go", "build", "-o", "metrics-server", "./cmd/metrics-server").CombinedOutput(); err != nil {
		t.Fatalf("build server failed: %v\n%s", err, out)
	}
	defer exec.Command("rm", "metrics-server").Run()

	cmd := exec.Command("./metrics-server", "--host-id", "inttest", "--port", "0")
	stdout, _ := cmd.StdoutPipe()
	if err := cmd.Start(); err != nil {
		t.Fatalf("start server failed: %v", err)
	}
	defer cmd.Process.Kill()

	// 2) read “Listening on :<port>” from server stdout
	var addr string
	scanner := bufio.NewScanner(stdout)
	timeout := time.After(3 * time.Second)
	for {
		select {
		case <-timeout:
			t.Fatal("timeout waiting for server to print port")
		default:
			if !scanner.Scan() {
				continue
			}
			line := scanner.Text()
			if strings.HasPrefix(line, "Listening on ") {
				addr = strings.TrimPrefix(line, "Listening on ")
				break
			}
		}
		if addr != "" {
			break
		}
	}

	// 3) dial with our client wrapper
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	cli, err := client.NewClient(ctx, addr)
	if err != nil {
		t.Fatalf("NewClient() error: %v", err)
	}
	defer cli.Close()

	// 4) fetch metrics
	m, err := cli.FetchMetrics(ctx)
	if err != nil {
		t.Fatalf("FetchMetrics() error: %v", err)
	}
	if m.HostID != "inttest" {
		t.Errorf("HostID = %q; want %q", m.HostID, "inttest")
	}

	// 5) force a server-side error to verify client surfaces gRPC errors
	//    (requires modifying the server to accept a “--fail” flag or similar).
	//    Example check:
	/*
	   _, err = cli.FetchMetrics(ctx) // assume server now returns error
	   st, _ := status.FromError(err)
	   if st.Code() != codes.Internal {
	       t.Errorf("expected INTERNAL, got %v", st.Code())
	   }
	*/
}
