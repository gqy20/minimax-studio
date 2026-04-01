package media

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// RunCommand 执行 ffmpeg/ffprobe 命令
func RunCommand(args ...string) error {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// RunCommandOutput 执行命令并捕获输出
func RunCommandOutput(args ...string) (string, error) {
	cmd := exec.Command(args[0], args[1:]...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command failed: %s, error: %w", strings.Join(args, " "), err)
	}
	return string(output), nil
}

// GetDurationSeconds 获取视频时长（秒）
func GetDurationSeconds(path string) (float64, error) {
	output, err := RunCommandOutput(
		"ffprobe",
		"-v", "error",
		"-show_entries", "format=duration",
		"-of", "default=noprint_wrappers=1:nokey=1",
		path,
	)
	if err != nil {
		return 0, err
	}

	var duration float64
	if _, scanErr := fmt.Sscanf(strings.TrimSpace(output), "%f", &duration); scanErr != nil {
		return 0, fmt.Errorf("failed to parse duration: %w", scanErr)
	}

	return duration, nil
}

// NormalizeVideo 标准化视频格式
func NormalizeVideo(inputPath, outputPath string) error {
	return RunCommand(
		"ffmpeg", "-y",
		"-i", inputPath,
		"-an",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		outputPath,
	)
}

// ConcatVideos 拼接多个视频
func ConcatVideos(videoPaths []string, outputPath string) error {
	if len(videoPaths) == 0 {
		return fmt.Errorf("no video segments to concatenate")
	}

	if len(videoPaths) == 1 {
		return NormalizeVideo(videoPaths[0], outputPath)
	}

	// 创建临时 concat 文件
	tempDir := filepath.Join(os.TempDir(), "minimax-studio")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return fmt.Errorf("failed to create temp dir: %w", err)
	}
	concatFile := filepath.Join(tempDir, "concat.txt")

	// 写入 concat 文件
	var lines []string
	for _, p := range videoPaths {
		absPath, err := filepath.Abs(p)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %w", err)
		}
		lines = append(lines, fmt.Sprintf("file '%s'", absPath))
	}

	if err := os.WriteFile(concatFile, []byte(strings.Join(lines, "\n")+"\n"), 0644); err != nil {
		return fmt.Errorf("failed to write concat file: %w", err)
	}
	defer os.Remove(concatFile)

	return RunCommand(
		"ffmpeg", "-y",
		"-f", "concat",
		"-safe", "0",
		"-i", concatFile,
		"-an",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		outputPath,
	)
}

// PadVideoToDuration 填充视频到目标时长
func PadVideoToDuration(inputPath, outputPath string, targetDuration float64) error {
	currentDuration, err := GetDurationSeconds(inputPath)
	if err != nil {
		return fmt.Errorf("failed to get current duration: %w", err)
	}

	extraDuration := targetDuration - currentDuration
	if extraDuration < 0.05 {
		return copyFile(inputPath, outputPath)
	}

	return RunCommand(
		"ffmpeg", "-y",
		"-i", inputPath,
		"-vf", fmt.Sprintf("tpad=stop_mode=clone:stop_duration=%.3f,format=yuv420p", extraDuration),
		"-an",
		"-c:v", "libx264",
		"-pix_fmt", "yuv420p",
		"-movflags", "+faststart",
		outputPath,
	)
}

// ComposeFinalVideo 合成最终视频
func ComposeFinalVideo(videoPath, narrationPath, musicPath, outputPath string, targetDuration float64) error {
	if musicPath == "" {
		return composeVideoWithNarration(videoPath, narrationPath, outputPath, targetDuration)
	}
	return composeVideoWithNarrationAndMusic(videoPath, narrationPath, musicPath, outputPath, targetDuration)
}

func composeVideoWithNarration(videoPath, narrationPath, outputPath string, targetDuration float64) error {
	return RunCommand(
		"ffmpeg", "-y",
		"-i", videoPath,
		"-i", narrationPath,
		"-filter_complex",
		fmt.Sprintf("[1:a]volume=1.0,atrim=0:%.3f,asetpts=N/SR/TB[aout]", targetDuration),
		"-map", "0:v",
		"-map", "[aout]",
		"-c:v", "copy",
		"-c:a", "aac",
		"-b:a", "192k",
		"-shortest",
		outputPath,
	)
}

func composeVideoWithNarrationAndMusic(videoPath, narrationPath, musicPath, outputPath string, targetDuration float64) error {
	return RunCommand(
		"ffmpeg", "-y",
		"-i", videoPath,
		"-stream_loop", "-1",
		"-i", musicPath,
		"-i", narrationPath,
		"-filter_complex",
		fmt.Sprintf(
			"[1:a]volume=0.16,atrim=0:%.3f,asetpts=N/SR/TB[bgm];"+
				"[2:a]volume=1.0,atrim=0:%.3f,asetpts=N/SR/TB[narr];"+
				"[bgm][narr]amix=inputs=2:duration=longest:dropout_transition=2[aout]",
			targetDuration, targetDuration,
		),
		"-map", "0:v",
		"-map", "[aout]",
		"-c:v", "copy",
		"-c:a", "aac",
		"-b:a", "192k",
		"-shortest",
		outputPath,
	)
}

func copyFile(src, dst string) error {
	input, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, input, 0644)
}
