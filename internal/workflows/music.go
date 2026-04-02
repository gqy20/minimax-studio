package workflows

import (
	"context"
	"fmt"
	"os"

	"github.com/minimax-ai/minimax-studio/internal/client"
	"github.com/minimax-ai/minimax-studio/internal/schemas"
)

// MusicWorkflow 音乐生成工作流
type MusicWorkflow struct {
	client *client.MiniMaxClient
}

// NewMusicWorkflow 创建 MusicWorkflow
func NewMusicWorkflow(cli *client.MiniMaxClient) *MusicWorkflow {
	return &MusicWorkflow{client: cli}
}

// Run 执行音乐生成
func (w *MusicWorkflow) Run(ctx context.Context, opts schemas.MusicOptions, reporter Reporter) (*schemas.MusicResult, error) {
	if reporter == nil {
		reporter = func(stage string) {}
	}

	reporter("generating music...")

	audioData, err := w.client.GenerateMusic(
		ctx,
		opts.Prompt,
		opts.Model,
		opts.AudioFormat,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate music: %w", err)
	}

	if err := os.WriteFile(opts.OutputPath, audioData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write music: %w", err)
	}

	reporter(fmt.Sprintf("music saved to: %s", opts.OutputPath))

	return &schemas.MusicResult{
		OutputPath: opts.OutputPath,
	}, nil
}
