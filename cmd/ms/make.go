package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/minimax-ai/minimax-studio/internal/client"
	"github.com/minimax-ai/minimax-studio/internal/workflows"
	"github.com/spf13/cobra"
)

var (
	makeTheme                string
	makeSceneCount           int
	makeSceneDuration        int
	makeOutputDir            string
	makeAspectRatio          string
	makeResolution           string
	makeVideoModel           string
	makeTextModel            string
	makeTTSModel             string
	makeMusicModel           string
	makeMusicMode            string
	makeVoiceID              string
	makeAudioFormat          string
	makeLanguage             string
	makePollInterval         int
	makeMaxWait              int
	makeInputVideo           string
	makeImagePromptOptimizer bool
	makeJSONOutput           bool
)

var makeCmd = &cobra.Command{
	Use:   "make",
	Short: "Create a full video from theme",
	Long:  `End-to-end video generation: plan → clip → voice → music → stitch.`,
	RunE:  runMake,
}

func init() {
	makeCmd.Flags().StringVarP(&makeTheme, "theme", "t", "", "Video theme")
	makeCmd.Flags().IntVar(&makeSceneCount, "scene-count", 1, "Number of scenes")
	makeCmd.Flags().IntVar(&makeSceneDuration, "scene-duration", 6, "Duration of each scene in seconds")
	makeCmd.Flags().StringVarP(&makeOutputDir, "output", "o", "", "Output directory")
	makeCmd.Flags().StringVar(&makeAspectRatio, "aspect-ratio", "16:9", "Aspect ratio")
	makeCmd.Flags().StringVar(&makeResolution, "resolution", "720p", "Video resolution")
	makeCmd.Flags().StringVar(&makeVideoModel, "video-model", "MiniMax-Hailuo-2.3-Fast", "Video model")
	makeCmd.Flags().StringVar(&makeTextModel, "text-model", "MiniMax-M2.7-highspeed", "Text model for planning")
	makeCmd.Flags().StringVar(&makeTTSModel, "tts-model", "speech-2.8-hd", "TTS model")
	makeCmd.Flags().StringVar(&makeMusicModel, "music-model", "music-2.5", "Music model")
	makeCmd.Flags().StringVar(&makeMusicMode, "music-mode", "optional", "Music mode: skip/optional/required")
	makeCmd.Flags().StringVar(&makeVoiceID, "voice-id", "male-qn-qingse", "Voice ID")
	makeCmd.Flags().StringVar(&makeAudioFormat, "audio-format", "mp3", "Audio format")
	makeCmd.Flags().StringVar(&makeLanguage, "language", "zh", "Language")
	makeCmd.Flags().IntVar(&makePollInterval, "poll-interval", 3, "Poll interval in seconds")
	makeCmd.Flags().IntVar(&makeMaxWait, "max-wait", 300, "Max wait in seconds")
	makeCmd.Flags().StringVar(&makeInputVideo, "input-video", "", "Reuse existing video (requires scene-count=1)")
	makeCmd.Flags().BoolVar(&makeImagePromptOptimizer, "image-prompt-optimizer", false, "Enable image prompt optimizer")
	makeCmd.Flags().BoolVar(&makeJSONOutput, "json", false, "Output JSON format")

	RootCmd.AddCommand(makeCmd)
}

func runMake(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if makeTheme == "" {
		return fmt.Errorf("--theme is required")
	}
	if makeOutputDir == "" {
		return fmt.Errorf("--output is required")
	}

	cli := client.NewClient(getAPIKey())

	opts := workflows.MakeOptions{
		Theme:                makeTheme,
		SceneCount:           makeSceneCount,
		SceneDuration:        makeSceneDuration,
		AspectRatio:          makeAspectRatio,
		Resolution:           makeResolution,
		TextModel:            makeTextModel,
		TextMaxTokens:        4096,
		VideoModel:           makeVideoModel,
		TTSModel:             makeTTSModel,
		MusicModel:           makeMusicModel,
		MusicMode:            makeMusicMode,
		VoiceID:              makeVoiceID,
		AudioFormat:          makeAudioFormat,
		Language:             makeLanguage,
		PollInterval:         makePollInterval,
		MaxWait:              makeMaxWait,
		OutputDir:            makeOutputDir,
		InputVideo:           makeInputVideo,
		ImagePromptOptimizer: makeImagePromptOptimizer,
	}

	workflow := workflows.NewMakeWorkflow(cli)

	if makeJSONOutput {
		result, err := workflow.Run(ctx, opts, func(stage string) {
			fmt.Fprintln(os.Stderr, stage)
		})
		if err != nil {
			return fmt.Errorf("make failed: %w", err)
		}
		jsonOutput, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonOutput))
		return nil
	}

	result, err := workflow.Run(ctx, opts, func(stage string) {
		fmt.Println(stage)
	})
	if err != nil {
		return fmt.Errorf("make failed: %w", err)
	}

	fmt.Printf("\nFinal video: %s\n", result.FinalVideoPath)

	return nil
}
