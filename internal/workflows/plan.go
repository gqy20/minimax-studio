package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/minimax-ai/minimax-studio/internal/client"
	"github.com/minimax-ai/minimax-studio/internal/schemas"
)

// PlanWorkflow 分镜规划工作流
type PlanWorkflow struct {
	client *client.MiniMaxClient
}

// NewPlanWorkflow 创建 PlanWorkflow
func NewPlanWorkflow(cli *client.MiniMaxClient) *PlanWorkflow {
	return &PlanWorkflow{client: cli}
}

// Run 执行分镜规划
func (w *PlanWorkflow) Run(ctx context.Context, opts schemas.PlanOptions, reporter Reporter) (*schemas.PlanResult, error) {
	if reporter == nil {
		reporter = func(stage string) {}
	}

	reporter("step 1/2: generating scene plan...")

	// 确保输出目录存在
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output dir: %w", err)
	}

	// 调用 MiniMax 文本模型生成规划
	planResp, err := w.client.PlanVideo(ctx, client.PlanVideoRequest{
		Theme:         opts.Theme,
		SceneCount:    opts.SceneCount,
		SceneDuration: opts.SceneDuration,
		Language:      opts.Language,
		TextModel:     opts.TextModel,
		TextMaxTokens: opts.TextMaxTokens,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate plan: %w", err)
	}

	// 转换为 schemas.VideoPlan
	plan := &schemas.VideoPlan{
		Title:       planResp.Title,
		VisualStyle: planResp.VisualStyle,
		Narration:   planResp.Narration,
		MusicPrompt: planResp.MusicPrompt,
	}
	for _, s := range planResp.Scenes {
		plan.Scenes = append(plan.Scenes, schemas.ScenePlan{
			Name:        s.Name,
			ImagePrompt: s.ImagePrompt,
			VideoPrompt: s.VideoPrompt,
		})
	}

	// 保存 plan.json
	planPath := opts.OutputDir + "/plan.json"
	planJSON, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal plan: %w", err)
	}
	if err := os.WriteFile(planPath, planJSON, 0644); err != nil {
		return nil, fmt.Errorf("failed to write plan: %w", err)
	}
	reporter(fmt.Sprintf("plan saved to: %s", planPath))

	// 保存 voice.txt (narration)
	voicePath := opts.OutputDir + "/voice.txt"
	if err := os.WriteFile(voicePath, []byte(plan.Narration), 0644); err != nil {
		return nil, fmt.Errorf("failed to write voice text: %w", err)
	}
	reporter(fmt.Sprintf("narration saved to: %s", voicePath))

	// Step 2: 打印摘要
	reporter("step 2/2: plan summary")
	reporter(fmt.Sprintf("  title: %s", plan.Title))
	reporter(fmt.Sprintf("  visual style: %s", plan.VisualStyle))
	reporter(fmt.Sprintf("  scenes: %d", len(plan.Scenes)))
	reporter(fmt.Sprintf("  narration: %s", truncate(plan.Narration, 100)))

	return &schemas.PlanResult{
		OutputDir:     opts.OutputDir,
		PlanPath:      planPath,
		NarrationPath: voicePath,
	}, nil
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
