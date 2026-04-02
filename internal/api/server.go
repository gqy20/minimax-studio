// Package api re-exports the Server type from handlers package.
// This file exists for backward compatibility — new code should import
// "github.com/minimax-ai/minimax-studio/internal/api/handlers" directly.
package api

import "github.com/minimax-ai/minimax-studio/internal/api/handlers"

// Server is the HTTP API server.
type Server = handlers.Server

// NewServer creates a new HTTP API server instance.
var NewServer = handlers.NewServer
