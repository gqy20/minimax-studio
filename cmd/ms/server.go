package main

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/minimax-ai/minimax-studio/internal/api"
	"github.com/spf13/cobra"
)

var (
	serverPort        string
	serverOutputDir   string
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
	addr := ":" + serverPort
	url := fmt.Sprintf("http://localhost%s", addr)

	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════╗")
	fmt.Println("║          MiniMax Studio                  ║")
	if Version != "" {
		fmt.Printf("║  Version: %-34s║\n", Version)
	}
	fmt.Println("╚══════════════════════════════════════════╝")
	fmt.Println()

	// Check API key
	apiKey := getAPIKey()
	if apiKey == "" {
		fmt.Println("  [!] 未检测到 API Key")
		fmt.Println("      请设置环境变量 MINIMAX_API_KEY，或使用 --api-key 参数")
		fmt.Println("      获取地址: https://platform.minimaxi.com/")
		fmt.Println()
	} else {
		masked := apiKey
		if len(apiKey) > 8 {
			masked = apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
		}
		fmt.Printf("  [✓] API Key: %s\n", masked)
	}

	// Check ffmpeg dependency
	if err := checkFFmpeg(); err != nil {
		fmt.Println()
		fmt.Println("  [!] 未检测到 ffmpeg / ffprobe")
		fmt.Println("      视频拼接和合成功能将不可用")
		fmt.Println("      安装地址: https://ffmpeg.org/download.html")
		fmt.Println()
	} else {
		fmt.Println("  [✓] ffmpeg: 已就绪")
	}

	// Use embedded frontend if available, otherwise try local dist
	frontendDir := api.EmbeddedFrontendDir()
	if frontendDir == "" {
		fmt.Println("  [!] 前端资源未嵌入（将使用本地 frontend/dist）")
	}

	fmt.Println()
	fmt.Printf("  正在启动服务器 -> %s\n", url)
	fmt.Printf("  输出目录: %s\n", serverOutputDir)
	fmt.Println()
	fmt.Println("  按 Ctrl+C 停止服务器")
	fmt.Println()

	// Auto-open browser
	if serverOpenBrowser {
		go openBrowser(url)
	}

	s := api.NewServer(serverOutputDir, apiKey, frontendDir)

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
