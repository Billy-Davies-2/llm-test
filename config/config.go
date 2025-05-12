package config

import (
	_ "github.com/joho/godotenv/autoload"
	"os"
	"time"
)

// Config holds application configuration read from environment variables.
type Config struct {
	OIDCIssuerURL string // OIDC issuer URL
	OIDCClientID  string // OIDC client ID

	ChatGRPCAddr    string // address for chat gRPC (e.g. ":50051")
	MetricsGRPCAddr string // address for metrics gRPC (e.g. ":50052")

	GossipSeeds string // comma-separated gossip seed addresses

	ModelDir string // local model directory path

	PollInterval time.Duration // poll interval for metrics
	DialTimeout  time.Duration // timeout for gRPC dialing
}

// Load reads configuration from environment, applying defaults where unset.
func Load() (*Config, error) {
	cfg := &Config{
		OIDCIssuerURL: getEnv("OIDC_ISSUER_URL", "https://keycloak.example.com/auth/realms/llm"),
		OIDCClientID:  getEnv("OIDC_CLIENT_ID", "llm-client"),

		ChatGRPCAddr:    getEnv("CHAT_GRPC_ADDR", ":50051"),
		MetricsGRPCAddr: getEnv("METRICS_GRPC_ADDR", ":50052"),

		GossipSeeds: getEnv("GOSSIP_SEEDS", "llm-backend-headless.llm.svc.cluster.local:7946"),

		ModelDir: getEnv("MODEL_DIR", "/models"),

		PollInterval: getEnvDuration("POLL_INTERVAL", 5*time.Second),
		DialTimeout:  getEnvDuration("DIAL_TIMEOUT", 5*time.Second),
	}
	return cfg, nil
}

// getEnv returns the value of key or defaultVal if unset.
func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

// getEnvDuration parses a time.Duration from env or returns defaultDur on error/empty.
func getEnvDuration(key string, defaultDur time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return defaultDur
}
