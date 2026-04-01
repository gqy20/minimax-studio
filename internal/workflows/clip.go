package workflows

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/minimax-ai/minimax-studio/internal/client"
	"github.com/minimax-ai/minimax-studio/internal/schemas"
)

// Reporter 进度报告函数
type Reporter func(stage string)

// ClipWorkflow 图生视频工作流
type ClipWorkflow struct {
	client *client.MiniMaxClient
}

// NewClipWorkflow 创建 ClipWorkflow
func NewClipWorkflow(cli *client.MiniMaxClient) *ClipWorkflow {
	return &ClipWorkflow{client: cli}
}

// Run 执行图生视频
func (w *ClipWorkflow) Run(ctx context.Context, opts schemas.ClipOptions, reporter Reporter) (*schemas.ClipResult, error) {
	if reporter == nil {
		reporter = func(stage string) {}
	}

	imagePath := opts.OutputPrefix + ".jpg"
	videoPath := opts.OutputPrefix + ".mp4"

	// Step 1: 生成图片
	reporter("step 1/4: generating image...")
	imageData, err := w.client.GenerateImage(
		ctx,
		opts.ImagePrompt,
		opts.AspectRatio,
		opts.ImagePromptOptimizer,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate image: %w", err)
	}

	if err := os.WriteFile(imagePath, imageData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write image: %w", err)
	}
	reporter(fmt.Sprintf("image saved to: %s", imagePath))

	// Step 2: 创建视频任务
	reporter("step 2/4: creating video task...")
	imageBase64 := base64.StdEncoding.EncodeToString(imageData)
	taskID, err := w.client.CreateVideoTask(
		ctx,
		imageBase64,
		opts.VideoPrompt,
		opts.VideoModel,
		opts.Duration,
		opts.Resolution,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create video task: %w", err)
	}
	reporter(fmt.Sprintf("video task id: %s", taskID))

	// Step 3: 轮询视频任务
	reporter("step 3/4: polling video task...")
	fileID, err := w.client.PollVideoTask(
		ctx,
		taskID,
		opts.PollInterval,
		opts.MaxWait,
		reporter,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to poll video task: %w", err)
	}
	reporter(fmt.Sprintf("video file id: %s", fileID))

	// Step 4: 下载视频
	reporter("step 4/4: downloading video...")
	downloadURL, err := w.client.FetchDownloadURL(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch download url: %w", err)
	}

	videoData, err := w.client.DownloadFile(ctx, downloadURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download video: %w", err)
	}

	if err := os.WriteFile(videoPath, videoData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write video: %w", err)
	}
	reporter(fmt.Sprintf("video saved to: %s", videoPath))
	reporter(fmt.Sprintf("download url: %s", downloadURL))

	return &schemas.ClipResult{
		ImagePath:   imagePath,
		VideoPath:   videoPath,
		TaskID:      taskID,
		FileID:      fileID,
		DownloadURL: downloadURL,
	}, nil
}

// EnsureDir 确保目录存在
func EnsureDir(path string) error {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	dir := filepath.Dir(absPath)
	return os.MkdirAll(dir, 0755)
}
