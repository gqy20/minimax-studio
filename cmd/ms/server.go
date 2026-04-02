package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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

// guiPrint prints a message. In GUI mode on Windows it also writes to a log file
// so there's a debugging trail when no console is visible.
func guiPrint(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Println(msg)

	if runningAsGUI && runtime.GOOS == "windows" {
		writeLog(msg)
	}
}

// writeLog appends a line to the log file in the output directory.
func writeLog(msg string) {
	dir := resolveOutputDir()
	if dir == "" {
		return
	}
	logFile := filepath.Join(dir, "studio.log")
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.WriteString(msg + "\n")
}

// resolveOutputDir returns the absolute path of the output directory.
// This is critical when double-clicked: the working directory is the exe's folder,
// and relative paths like "./output" would land next to the binary.
func resolveOutputDir() string {
	if serverOutputDir == "" {
		serverOutputDir = "./output"
	}
	if !filepath.IsAbs(serverOutputDir) {
		// Resolve relative to exe's directory when double-clicked,
		// or current directory when running from terminal.
		exePath, err := os.Executable()
		if err == nil {
			exeDir := filepath.Dir(exePath)
			serverOutputDir = filepath.Join(exeDir, serverOutputDir)
		} else {
			abs, _ := filepath.Abs(serverOutputDir)
			serverOutputDir = abs
		}
	}
	return serverOutputDir
}

func runServer(cmd *cobra.Command, args []string) error {
	addr := ":" + serverPort
	url := fmt.Sprintf("http://localhost%s", addr)

	outputDir := resolveOutputDir()

	guiPrint("")
	guiPrint("╔══════════════════════════════════════════╗")
	guiPrint("║          MiniMax Studio                  ║")
	if Version != "" {
		guiPrint("║  Version: %-34s║", Version)
	}
	guiPrint("╚══════════════════════════════════════════╝")
	guiPrint("")

	// Check API key
	apiKey := getAPIKey()
	if apiKey == "" {
		guiPrint("  [!] 未检测到 API Key")
		guiPrint("      请设置环境变量 MINIMAX_API_KEY，或使用 --api-key 参数")
		guiPrint("      获取地址: https://platform.minimaxi.com/")
		guiPrint("")
		if runningAsGUI && runtime.GOOS == "windows" {
			ShowErrorDialog("MiniMax Studio - 缺少 API Key",
				"未检测到 API Key。\n\n"+
					"请设置系统环境变量 MINIMAX_API_KEY，\n"+
					"或在终端中运行:\n\n"+
					"  set MINIMAX_API_KEY=你的密钥\n"+
					"  ms.exe --api-key 你的密钥\n\n"+
					"获取地址: https://platform.minimaxi.com/")
			return fmt.Errorf("API key not configured")
		}
	} else {
		masked := apiKey
		if len(apiKey) > 8 {
			masked = apiKey[:4] + "****" + apiKey[len(apiKey)-4:]
		}
		guiPrint("  [✓] API Key: %s", masked)
	}

	// Check ffmpeg dependency
	if err := checkFFmpeg(); err != nil {
		guiPrint("")
		guiPrint("  [!] 未检测到 ffmpeg / ffprobe")
		guiPrint("      视频拼接和合成功能将不可用")
		guiPrint("      安装地址: https://ffmpeg.org/download.html")
		guiPrint("")
	} else {
		guiPrint("  [✓] ffmpeg: 已就绪")
	}

	// Use embedded frontend if available, otherwise try local dist
	frontendDir := api.EmbeddedFrontendDir()
	if frontendDir == "" {
		guiPrint("  [!] 前端资源未嵌入（将使用本地 frontend/dist）")
	}

	guiPrint("")
	guiPrint("  正在启动服务器 -> %s", url)
	guiPrint("  输出目录: %s", outputDir)
	guiPrint("")
	guiPrint("  按 Ctrl+C 停止服务器")
	guiPrint("")

	// Auto-open browser
	if serverOpenBrowser {
		go openBrowser(url)
	}

	s := api.NewServer(outputDir, apiKey, frontendDir)

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
		cmd = exec.Command("cmd", "/c", "start", "", url)
	default: // linux, freebsd, etc.
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
