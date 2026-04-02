//go:build !windows

package main

// ShowErrorDialog is a no-op on non-Windows platforms.
func ShowErrorDialog(title, message string) {
	// On Linux/macOS, stderr is always available
}

// HideConsole is a no-op on non-Windows platforms.
func HideConsole() {}
