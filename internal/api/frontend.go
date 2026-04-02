package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// RegisterFrontend serves the SPA frontend from the given directory.
// If the directory is empty or doesn't exist, returns a helpful JSON error.
// API routes take priority; everything else falls through to the SPA.
func RegisterFrontend(engine *gin.Engine, frontendDir string) {
	if frontendDir == "" {
		engine.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "frontend not built — run `make frontend-build` and rebuild",
			})
		})
		return
	}

	// Check if directory exists and has content
	entries, err := os.ReadDir(frontendDir)
	if err != nil || len(entries) == 0 {
		engine.NoRoute(func(c *gin.Context) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "frontend not built — run `make frontend-build` and rebuild",
			})
		})
		return
	}

	fileServer := http.FileServer(http.Dir(frontendDir))

	engine.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// Try serving the static file first
		if !strings.HasSuffix(path, "/") {
			ext := filepath.Ext(path)
			if ext != "" {
				cleanPath := strings.TrimPrefix(path, "/")
				fullPath := filepath.Join(frontendDir, cleanPath)
				if info, err := os.Stat(fullPath); err == nil && !info.IsDir() {
					fileServer.ServeHTTP(c.Writer, c.Request)
					return
				}
			}
		}

		// SPA fallback: serve index.html for client-side routing
		c.Request.URL.Path = "/"
		fileServer.ServeHTTP(c.Writer, c.Request)
	})
}
