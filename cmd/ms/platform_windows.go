//go:build windows

package main

import (
	"syscall"
	"unsafe"
)

var (
	user32           = syscall.NewLazyDLL("user32.dll")
	messageBoxW      = user32.NewProc("MessageBoxW")
	getConsoleWindow = user32.NewProc("GetConsoleWindow")
	showWindow       = user32.NewProc("ShowWindow")
)

const (
	MB_ICONERROR    = 0x10
	MB_ICONWARNING  = 0x30
	MB_ICONINFO     = 0x40
	MB_OK           = 0x0
	SW_HIDE         = 0
)

// ShowErrorDialog displays a modal error dialog on Windows.
// Used when running as GUI app (double-clicked) so errors are visible.
func ShowErrorDialog(title, message string) {
	titlePtr, _ := syscall.UTF16PtrFromString(title)
	msgPtr, _ := syscall.UTF16PtrFromString(message)
	messageBoxW.Call(0, uintptr(unsafe.Pointer(msgPtr)), uintptr(unsafe.Pointer(titlePtr)), MB_OK|MB_ICONERROR)
}

// HideConsole hides the console window (for GUI mode).
func HideConsole() {
	hwnd, _, _ := getConsoleWindow.Call()
	if hwnd != 0 {
		showWindow.Call(hwnd, SW_HIDE)
	}
}
