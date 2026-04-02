package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/minimax-ai/minimax-studio/internal/client"
	"github.com/minimax-ai/minimax-studio/internal/schemas"
	"github.com/minimax-ai/minimax-studio/internal/workflows"
	"github.com/spf13/cobra"
)

var (
	musicPrompt     string
	musicOutputPath string
	musicModel      string
	musicFormat     string
	musicJSONOutput bool
)

var musicCmd = &cobra.Command{
	Use:   "music",
	Short: "Generate background music",
	Long:  `Generate background music from text description.`,
	RunE:  runMusic,
}

func init() {
	musicCmd.Flags().StringVarP(&musicPrompt, "prompt", "p", "", "Music description prompt")
	musicCmd.Flags().StringVarP(&musicOutputPath, "output", "o", "", "Output audio path")
	musicCmd.Flags().StringVar(&musicModel, "model", "music-2.5", "Music model")
	musicCmd.Flags().StringVar(&musicFormat, "format", "mp3", "Audio format (mp3/wav/flac)")
	musicCmd.Flags().BoolVar(&musicJSONOutput, "json", false, "Output JSON format")

	RootCmd.AddCommand(musicCmd)
}

func runMusic(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if musicPrompt == "" {
		return fmt.Errorf("--prompt is required")
	}
	if musicOutputPath == "" {
		return fmt.Errorf("--output is required")
	}

	cli := client.NewClient(getAPIKey())

	opts := schemas.MusicOptions{
		Prompt:      musicPrompt,
		OutputPath:  musicOutputPath,
		Model:       musicModel,
		AudioFormat: musicFormat,
	}

	workflow := workflows.NewMusicWorkflow(cli)

	if musicJSONOutput {
		result, err := workflow.Run(ctx, opts, func(stage string) {
			fmt.Fprintln(os.Stderr, stage)
		})
		if err != nil {
			return fmt.Errorf("music failed: %w", err)
		}
		jsonOutput, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonOutput))
		return nil
	}

	result, err := workflow.Run(ctx, opts, func(stage string) {
		fmt.Println(stage)
	})
	if err != nil {
		return fmt.Errorf("music failed: %w", err)
	}

	fmt.Printf("Music: %s\n", result.OutputPath)

	return nil
}
