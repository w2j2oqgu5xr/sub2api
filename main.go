package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/sub2api/sub2api/handler"
)

const (
	defaultPort    = 8080
	defaultHost    = "0.0.0.0"
	appName        = "sub2api"
	appVersion     = "1.0.0"
)

func main() {
	// Parse command-line flags
	port := flag.Int("port", getEnvInt("PORT", defaultPort), "Port to listen on")
	host := flag.String("host", getEnv("HOST", defaultHost), "Host to bind to")
	token := flag.String("token", getEnv("API_TOKEN", ""), "API token for authentication (optional)")
	version := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *version {
		fmt.Printf("%s v%s\n", appName, appVersion)
		os.Exit(0)
	}

	// Initialize router
	mux := http.NewServeMux()

	// Register routes
	h := handler.New(*token)
	mux.HandleFunc("/sub", h.SubHandler)
	mux.HandleFunc("/health", h.HealthHandler)
	mux.HandleFunc("/", h.IndexHandler)

	addr := fmt.Sprintf("%s:%d", *host, *port)
	log.Printf("[%s] Starting server on %s (v%s)", appName, addr, appVersion)

	// Increased read/write timeouts to better handle slow subscription sources.
	// Bumped IdleTimeout to 180s to keep persistent connections alive longer
	// on my home server where clients reconnect frequently.
	// Increased WriteTimeout to 90s since some subscription URLs are very slow
	// to respond and were getting cut off at 60s.
	server := &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 90 * time.Second,
		IdleTimeout:  180 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv retrieves an environment variable or returns a default value.
func getEnv(key, defaultVal string) string {
	if val, ok := os.LookupEnv(key); ok {
		return val
	}
	return defaultVal
}

// getEnvInt retrieves an integer environment variable or returns a default value.
func getEnvInt(key string, defaultVal int) int {
	if val, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}
