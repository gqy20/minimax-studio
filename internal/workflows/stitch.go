package workflows

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/minimax-ai/minimax-studio/internal/media"
	"github.com/minimax-ai/minimax-studio/internal/schemas"
)

// StitchWorkflow 素材合成工作流
type StitchWorkflow struct {
}

// NewStitchWorkflow 创建 StitchWorkflow
func NewStitchWorkflow() *StitchWorkflow {
	return &StitchWorkflow{}
}

// Run 执行素材合成
func (w *StitchWorkflow) Run(opts schemas.StitchOptions, reporter Reporter) (*schemas.StitchResult, error) {
	if reporter == nil {
		reporter = func(stage string) {}
	}

	// 验证 ffmpeg
	if !isFFmpegAvailable() {
		return nil, fmt.Errorf("ffmpeg and ffprobe are required")
	}

	// 验证输入
	if len(opts.VideoPaths) == 0 {
		return nil, fmt.Errorf("at least one input video is required")
	}

	for _, p := range opts.VideoPaths {
		if _, err := os.Stat(p); os.IsNotExist(err) {
			return nil, fmt.Errorf("input video does not exist: %s", p)
		}
	}

	if _, err := os.Stat(opts.NarrationPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("narration file does not exist: %s", opts.NarrationPath)
	}

	if opts.MusicPath != "" {
		if _, err := os.Stat(opts.MusicPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("music file does not exist: %s", opts.MusicPath)
		}
	}

	// 确保输出目录存在
	outputDir := filepath.Dir(opts.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output dir: %w", err)
	}

	stitchedVideoPath := filepath.Join(outputDir, "edit.mp4")
	paddedVideoPath := filepath.Join(outputDir, "timed.mp4")
	finalVideoPath := opts.OutputPath

	// Step 1: 拼接视频
	reporter("step 1/3: stitching source video...")
	if err := media.ConcatVideos(opts.VideoPaths, stitchedVideoPath); err != nil {
		return nil, fmt.Errorf("failed to concat videos: %w", err)
	}

	// Step 2: 计算时长并填充
	reporter("step 2/3: padding video to target duration...")
	narrationDuration, err := media.GetDurationSeconds(opts.NarrationPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get narration duration: %w", err)
	}

	videoDuration, err := media.GetDurationSeconds(stitchedVideoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get video duration: %w", err)
	}

	targetDuration := narrationDuration
	if videoDuration > narrationDuration {
		targetDuration = videoDuration
	}

	if err := media.PadVideoToDuration(stitchedVideoPath, paddedVideoPath, targetDuration); err != nil {
		return nil, fmt.Errorf("failed to pad video: %w", err)
	}

	reporter(fmt.Sprintf(
		"stitched video duration=%.2fs, narration duration=%.2fs, target=%.2fs",
		videoDuration, narrationDuration, targetDuration,
	))

	// Step 3: 合成最终视频
	reporter("step 3/3: composing final video...")
	if err := media.ComposeFinalVideo(paddedVideoPath, opts.NarrationPath, opts.MusicPath, finalVideoPath, targetDuration); err != nil {
		return nil, fmt.Errorf("failed to compose final video: %w", err)
	}
	reporter(fmt.Sprintf("final video saved to: %s", finalVideoPath))

	return &schemas.StitchResult{
		StitchedVideoPath: stitchedVideoPath,
		PaddedVideoPath:   paddedVideoPath,
		FinalVideoPath:    finalVideoPath,
	}, nil
}

func isFFmpegAvailable() bool {
	cmd := exec.Command("ffmpeg", "-version")
	return cmd.Run() == nil
}
