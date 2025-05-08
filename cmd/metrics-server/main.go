// cmd/metrics-server/main.go
package main

import (
	"log/slog"
	"os"
)

func initLogger() *slog.Logger {
	// open a file (or os.Stderr, or both via io.MultiWriter)
	f, err := os.OpenFile("metrics-server.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		// fallback to stderr
		slog.Warn("could not open log file, using stderr", "err", err)
		return slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	// choose JSON or Text handler
	handler := slog.NewJSONHandler(f, &slog.HandlerOptions{
		AddSource: true,            // include file:line
		Level:     slog.LevelDebug, // min level
	})
	return slog.New(handler)
}

func main() {
	logger := initLogger()
	// pass logger into your server package, e.g.:
	server.Run(logger, hostID, port)
}
