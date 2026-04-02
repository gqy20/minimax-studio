package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

// runningAsGUI is set to true when the binary is launched with no arguments
// (double-clicked), so we know to show error dialogs instead of silent exits.
var runningAsGUI bool

// isDoubleClick returns true when the binary is launched with no arguments,
// which typically means it was double-clicked on Windows or macOS.
func isDoubleClick() bool {
	return len(os.Args) <= 1
}

var (
	apiKey    string
	groupID   string
	Version   string
	BuildTime string
)

// RootCmd 根命令
var RootCmd = &cobra.Command{
	Use:   "ms",
	Short: "MiniMax Studio CLI",
	Long:  `MiniMax Studio CLI for clip, planning, voice, music, stitch, and full video workflows.`,
}

func init() {
	RootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "MiniMax API Key (or set MINIMAX_API_KEY env)")
	RootCmd.PersistentFlags().StringVar(&groupID, "group-id", "", "MiniMax Group ID (or set MINIMAX_GROUP_ID env)")
}

func main() {
	// When run with no arguments (e.g. double-clicked on Windows),
	// auto-start the HTTP server so it works like a GUI app.
	if isDoubleClick() {
		runningAsGUI = true
		os.Args = append(os.Args, "server")
		// In GUI mode on Windows, hide the console window
		if runtime.GOOS == "windows" {
			HideConsole()
		}
	}

	if err := RootCmd.Execute(); err != nil {
		if runningAsGUI && runtime.GOOS == "windows" {
			ShowErrorDialog("MiniMax Studio", fmt.Sprintf("启动失败:\n\n%v", err))
		} else {
			fmt.Fprintln(os.Stderr, err)
		}
		os.Exit(1)
	}
}

func getAPIKey() string {
	if apiKey != "" {
		return apiKey
	}
	return os.Getenv("MINIMAX_API_KEY")
}

func getGroupID() string {
	if groupID != "" {
		return groupID
	}
	return os.Getenv("MINIMAX_GROUP_ID")
}
