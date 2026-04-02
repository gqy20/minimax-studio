//go:build !embed_frontend

package api

// EmbeddedFrontendDir returns empty string when built without embed tag.
func EmbeddedFrontendDir() string {
	return ""
}
