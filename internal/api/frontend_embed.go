//go:build embed_frontend

package api

import (
	"embed"
	"io/fs"
	"os"
	"path/filepath"
)

//go:embed all:frontend_dist
var embeddedFrontend embed.FS

func init() {
	// Write embedded frontend to a temp directory at startup
	// so RegisterFrontend can serve it from disk.
	// This avoids implementing a full fs.FS server.
}

// EmbeddedFrontendDir extracts the embedded frontend to a temp directory
// and returns the path. The caller should not clean it up — it persists
// for the lifetime of the process.
func EmbeddedFrontendDir() string {
	sub, err := fs.Sub(embeddedFrontend, "frontend_dist")
	if err != nil {
		return ""
	}

	tmpDir := filepath.Join(os.TempDir(), "minimax-studio-frontend")
	_ = os.RemoveAll(tmpDir)
	_ = os.MkdirAll(tmpDir, 0755)

	fs.WalkDir(sub, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		target := filepath.Join(tmpDir, path)
		if d.IsDir() {
			_ = os.MkdirAll(target, 0755)
			return nil
		}
		data, err := fs.ReadFile(sub, path)
		if err != nil {
			return nil
		}
		_ = os.WriteFile(target, data, 0644)
		return nil
	})

	return tmpDir
}
