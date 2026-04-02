package main

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"github.com/minimax-ai/minimax-studio/internal/api"
	"github.com/spf13/cobra"
)

var (
	serverPort      string
	serverOutputDir string
	serverOpenBrowser bool
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start HTTP API server",
	Long:  `Start the MiniMax Studio HTTP API server for frontend integration.`,
	RunE:  runServer,
}

func init() {
	serverCmd.Flags().StringVar(&serverPort, "port", "8080", "Server port")
	serverCmd.Flags().StringVar(&serverOutputDir, "output-dir", "./output", "Output directory")
	serverCmd.Flags().BoolVar(&serverOpenBrowser, "open", true, "Auto-open browser")

	RootCmd.AddCommand(serverCmd)
}

func runServer(cmd *cobra.Command, args []string) error {
	// Check ffmpeg dependency
	if err := checkFFmpeg(); err != nil {
		log.Printf("WARNING: %v", err)
		log.Printf("Video stitching and composition will not work without ffmpeg.")
		log.Printf("Install ffmpeg: https://ffmpeg.org/download.html")
		fmt.Println()
	}

	apiKey := getAPIKey()

	// Use embedded frontend if available, otherwise try local dist
	frontendDir := api.EmbeddedFrontendDir()

	s := api.NewServer(serverOutputDir, apiKey, frontendDir)
	addr := ":" + serverPort
	url := fmt.Sprintf("http://localhost%s", addr)

	// Auto-open browser
	if serverOpenBrowser {
		go openBrowser(url)
	}

	log.Printf("Starting MiniMax Studio on %s", url)
	log.Printf("Open %s in your browser to get started.", url)
	return s.Run(addr)
}

func checkFFmpeg() error {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found in PATH")
	}
	if _, err := exec.LookPath("ffprobe"); err != nil {
		return fmt.Errorf("ffprobe not found in PATH")
	}
	return nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
