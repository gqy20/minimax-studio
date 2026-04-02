package workflows

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/minimax-ai/minimax-studio/internal/client"
	"github.com/minimax-ai/minimax-studio/internal/media"
	"github.com/minimax-ai/minimax-studio/internal/schemas"
)

// MakeOptions 全流程选项
type MakeOptions struct {
	Theme                string
	SceneCount           int
	SceneDuration        int
	AspectRatio          string
	Resolution           string
	TextModel            string
	TextMaxTokens        int
	VideoModel           string
	TTSModel             string
	MusicModel           string
	MusicMode            string // "skip", "optional", "required"
	VoiceID              string
	AudioFormat          string
	PollInterval         int
	MaxWait              int
	Language             string
	OutputDir            string
	InputVideo           string // 可选，复用已有视频
	ImagePromptOptimizer bool
}

// MakeResult 全流程结果
type MakeResult struct {
	OutputDir      string  `json:"output_dir"`
	PlanPath       string  `json:"plan_path"`
	NarrationPath  string  `json:"narration_path"`
	MusicPath      string  `json:"music_path,omitempty"`
	FinalVideoPath string  `json:"final_video_path"`
}

// MakeWorkflow 全流程工作流
type MakeWorkflow struct {
	client *client.MiniMaxClient
}

// NewMakeWorkflow 创建 MakeWorkflow
func NewMakeWorkflow(cli *client.MiniMaxClient) *MakeWorkflow {
	return &MakeWorkflow{client: cli}
}

// Run 执行全流程
func (w *MakeWorkflow) Run(ctx context.Context, opts MakeOptions, reporter Reporter) (*MakeResult, error) {
	if reporter == nil {
		reporter = func(stage string) {}
	}

	// 验证
	if opts.SceneCount < 1 {
		return nil, fmt.Errorf("scene-count must be >= 1")
	}
	if opts.InputVideo != "" && opts.SceneCount != 1 {
		return nil, fmt.Errorf("input-video currently requires scene-count 1")
	}

	// 确保输出目录存在
	if err := os.MkdirAll(opts.OutputDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create output dir: %w", err)
	}

	// Step 1: 生成分镜规划
	reporter("step 1/7: planning storyboard...")
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

	// 转换为 VideoPlan
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
	planPath := filepath.Join(opts.OutputDir, "plan.json")
	planJSON, _ := json.MarshalIndent(plan, "", "  ")
	if err := os.WriteFile(planPath, planJSON, 0644); err != nil {
		return nil, fmt.Errorf("failed to write plan: %w", err)
	}
	reporter(fmt.Sprintf("plan saved to: %s", planPath))

	// 保存 voice.txt
	voiceTextPath := filepath.Join(opts.OutputDir, "voice.txt")
	if err := os.WriteFile(voiceTextPath, []byte(plan.Narration), 0644); err != nil {
		return nil, fmt.Errorf("failed to write voice text: %w", err)
	}

	// Step 2-3: 生成视频片段
	var sceneVideos []string
	if opts.InputVideo != "" {
		// 复用已有视频
		reporter(fmt.Sprintf("step 2/7: reusing existing video: %s", opts.InputVideo))
		reusedPath := filepath.Join(opts.OutputDir, "s01.mp4")
		data, err := os.ReadFile(opts.InputVideo)
		if err != nil {
			return nil, fmt.Errorf("failed to read input video: %w", err)
		}
		if err := os.WriteFile(reusedPath, data, 0644); err != nil {
			return nil, fmt.Errorf("failed to copy video: %w", err)
		}
		sceneVideos = append(sceneVideos, reusedPath)
	} else {
		for i, scene := range plan.Scenes {
			idx := i + 1
			framePath := filepath.Join(opts.OutputDir, fmt.Sprintf("s%02d.jpg", idx))
			videoPath := filepath.Join(opts.OutputDir, fmt.Sprintf("s%02d.mp4", idx))

			// 生成关键帧
			reporter(fmt.Sprintf("step 2/7: generating image for scene %d...", idx))
			imageData, imageBase64, err := w.client.GenerateImage(ctx, scene.ImagePrompt, opts.AspectRatio, opts.ImagePromptOptimizer)
			if err != nil {
				return nil, fmt.Errorf("failed to generate image for scene %d: %w", idx, err)
			}
			if err := os.WriteFile(framePath, imageData, 0644); err != nil {
				return nil, fmt.Errorf("failed to write frame: %w", err)
			}
			reporter(fmt.Sprintf("scene %d frame saved to: %s", idx, framePath))

			// 生成视频
			reporter(fmt.Sprintf("step 3/7: generating video for scene %d...", idx))
			taskID, err := w.client.CreateVideoTask(
				ctx,
				imageBase64,
				scene.VideoPrompt,
				opts.VideoModel,
				opts.SceneDuration,
				opts.Resolution,
			)
			if err != nil {
				return nil, fmt.Errorf("failed to create video task for scene %d: %w", idx, err)
			}
			reporter(fmt.Sprintf("scene %d task id: %s", idx, taskID))

			fileID, err := w.client.PollVideoTask(ctx, taskID, opts.PollInterval, opts.MaxWait, reporter)
			if err != nil {
				return nil, fmt.Errorf("failed to poll video task for scene %d: %w", idx, err)
			}

			downloadURL, err := w.client.FetchDownloadURL(ctx, fileID)
			if err != nil {
				return nil, fmt.Errorf("failed to fetch download url for scene %d: %w", idx, err)
			}

			videoData, err := w.client.DownloadFile(ctx, downloadURL)
			if err != nil {
				return nil, fmt.Errorf("failed to download video for scene %d: %w", idx, err)
			}
			if err := os.WriteFile(videoPath, videoData, 0644); err != nil {
				return nil, fmt.Errorf("failed to write video: %w", err)
			}
			reporter(fmt.Sprintf("scene %d video saved to: %s", idx, videoPath))
			sceneVideos = append(sceneVideos, videoPath)
		}
	}

	// Step 4: 生成旁白
	narrationPath := filepath.Join(opts.OutputDir, fmt.Sprintf("voice.%s", opts.AudioFormat))
	reporter("step 4/7: generating narration...")
	audioData, err := w.client.SynthesizeSpeech(ctx, plan.Narration, opts.VoiceID, opts.TTSModel, opts.AudioFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to synthesize narration: %w", err)
	}
	if err := os.WriteFile(narrationPath, audioData, 0644); err != nil {
		return nil, fmt.Errorf("failed to write narration: %w", err)
	}
	reporter(fmt.Sprintf("narration saved to: %s", narrationPath))

	// Step 5: 生成背景音乐
	var musicPath string
	if opts.MusicMode != "skip" {
		tentativeMusicPath := filepath.Join(opts.OutputDir, fmt.Sprintf("music.%s", opts.AudioFormat))
		reporter("step 5/7: generating music...")
		musicData, err := w.client.GenerateMusic(ctx, plan.MusicPrompt, opts.MusicModel, opts.AudioFormat)
		if err != nil {
			if opts.MusicMode == "required" {
				return nil, fmt.Errorf("failed to generate music: %w", err)
			}
			reporter(fmt.Sprintf("music generation unavailable, continuing without: %v", err))
		} else {
			if err := os.WriteFile(tentativeMusicPath, musicData, 0644); err != nil {
				return nil, fmt.Errorf("failed to write music: %w", err)
			}
			musicPath = tentativeMusicPath
			reporter(fmt.Sprintf("music saved to: %s", musicPath))
		}
	} else {
		reporter("step 5/7: skipping music generation by request...")
	}

	// Step 6: 拼接视频
	stitchedPath := filepath.Join(opts.OutputDir, "edit.mp4")
	reporter("step 6/7: stitching and timing video...")
	if err := media.ConcatVideos(sceneVideos, stitchedPath); err != nil {
		return nil, fmt.Errorf("failed to concat videos: %w", err)
	}

	narrationDuration, _ := media.GetDurationSeconds(narrationPath)
	videoDuration, _ := media.GetDurationSeconds(stitchedPath)
	targetDuration := narrationDuration
	if videoDuration > narrationDuration {
		targetDuration = videoDuration
	}

	paddedPath := filepath.Join(opts.OutputDir, "timed.mp4")
	if err := media.PadVideoToDuration(stitchedPath, paddedPath, targetDuration); err != nil {
		return nil, fmt.Errorf("failed to pad video: %w", err)
	}
	reporter(fmt.Sprintf("stitched=%.2fs, narration=%.2fs, target=%.2fs", videoDuration, narrationDuration, targetDuration))

	// Step 7: 合成最终视频
	finalPath := filepath.Join(opts.OutputDir, "final.mp4")
	reporter("step 7/7: composing final video...")
	if err := media.ComposeFinalVideo(paddedPath, narrationPath, musicPath, finalPath, targetDuration); err != nil {
		return nil, fmt.Errorf("failed to compose final video: %w", err)
	}
	reporter(fmt.Sprintf("final video saved to: %s", finalPath))

	return &MakeResult{
		OutputDir:      opts.OutputDir,
		PlanPath:       planPath,
		NarrationPath:  narrationPath,
		MusicPath:      musicPath,
		FinalVideoPath: finalPath,
	}, nil
}
