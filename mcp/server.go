package mcp

import (
	"context"
	"net/http"

	"github.com/mark3labs/mcp-go/server"

	"omnitranscripts/config"
)

// ServerConfig holds configuration for the MCP server.
type ServerConfig struct {
	APIKey string
}

// NewServer creates a new MCP server with OmniTranscripts tools.
func NewServer() *server.MCPServer {
	s := server.NewMCPServer(
		"OmniTranscripts",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	// Register transcribe_url tool
	s.AddTool(TranscribeURLTool(), HandleTranscribeURL)

	// Register get_transcription tool
	s.AddTool(GetTranscriptionTool(), HandleGetTranscription)

	return s
}

// NewHTTPHandler creates an HTTP handler for the MCP server.
// This can be mounted at any path (e.g., /mcp) on an existing HTTP server.
func NewHTTPHandler(apiKey string) http.Handler {
	s := NewServer()

	// Create streamable HTTP server with authentication
	httpServer := server.NewStreamableHTTPServer(s,
		server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			// Add API key to context for potential use in handlers
			return context.WithValue(ctx, "api_key", apiKey)
		}),
	)

	// Wrap with authentication middleware
	return &authHandler{
		apiKey:  apiKey,
		handler: httpServer,
	}
}

// authHandler wraps the MCP server with API key authentication.
type authHandler struct {
	apiKey  string
	handler http.Handler
}

func (a *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Check for API key in Authorization header
	authHeader := r.Header.Get("Authorization")

	// Support both "Bearer <token>" and raw token
	var providedKey string
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		providedKey = authHeader[7:]
	} else {
		providedKey = authHeader
	}

	// Also check X-API-Key header as fallback
	if providedKey == "" {
		providedKey = r.Header.Get("X-API-Key")
	}

	// Validate API key
	if providedKey == "" || providedKey != a.apiKey {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Invalid or missing API key"}`))
		return
	}

	// Set CORS headers for ChatGPT integration
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-API-Key")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Pass to MCP server
	a.handler.ServeHTTP(w, r)
}

// StartStandalone starts the MCP server on a standalone port.
// This is useful for running the MCP server independently of the main HTTP API.
func StartStandalone(address string) error {
	cfg := config.Load()
	s := NewServer()

	httpServer := server.NewStreamableHTTPServer(s,
		server.WithHTTPContextFunc(func(ctx context.Context, r *http.Request) context.Context {
			return context.WithValue(ctx, "api_key", cfg.APIKey)
		}),
	)

	return httpServer.Start(address)
}
