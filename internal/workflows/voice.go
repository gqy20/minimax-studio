package workflows

import (
	"context"
	"fmt"
	"os"

	"github.com/minimax-ai/minimax-studio/internal/client"
	"github.com/minimax-ai/minimax-studio/internal/schemas"
)

// VoiceWorkflow 语音合成工作流
type VoiceWorkflow struct {
	client *client.MiniMaxClient
}

// NewVoiceWorkflow 创建 VoiceWorkflow
func NewVoiceWorkflow(cli *client.MiniMaxClient) *VoiceWorkflow {
	return &VoiceWorkflow{client: cli}
}

// Run 执行语音合成
func (w *VoiceWorkflow) Run(ctx context.Context, opts schemas.VoiceOptions, reporter Reporter) (*schemas.VoiceResult, error) {
	if reporter == nil {
		reporter = func(stage string) {}
	}

	reporter("synthesizing speech...")

	audioData, err := w.client.SynthesizeSpeech(
		ctx,
		opts.Text,
		opts.VoiceID,
		opts.TTSModel,
		opts.AudioFormat,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize speech: %w", err)
	}

	if err := os.WriteFile(opts.OutputPath, audioData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write audio: %w", err)
	}

	reporter(fmt.Sprintf("audio saved to: %s", opts.OutputPath))

	return &schemas.VoiceResult{
		OutputPath: opts.OutputPath,
	}, nil
}
