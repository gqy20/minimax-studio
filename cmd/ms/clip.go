package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/minimax-ai/minimax-studio/internal/client"
	"github.com/minimax-ai/minimax-studio/internal/schemas"
	"github.com/minimax-ai/minimax-studio/internal/workflows"
	"github.com/spf13/cobra"
)

var (
	clipPrompt       string
	clipSubject      string
	clipAspectRatio  string
	clipModel        string
	clipDuration     int
	clipResolution   string
	clipOutputPrefix string
	clipJSONOutput   bool
)

var clipCmd = &cobra.Command{
	Use:   "clip",
	Short: "Generate a video clip from image prompt",
	Long:  `Generate a keyframe image first, then create a video clip from it.`,
	RunE:  runClip,
}

func init() {
	clipCmd.Flags().StringVarP(&clipPrompt, "prompt", "p", "", "Image prompt")
	clipCmd.Flags().StringVarP(&clipSubject, "subject", "s", "", "Video subject/motion description")
	clipCmd.Flags().StringVarP(&clipAspectRatio, "aspect-ratio", "a", "16:9", "Aspect ratio (16:9, 9:16, 1:1)")
	clipCmd.Flags().StringVar(&clipModel, "model", "MiniMax-Hailuo-2.3-Fast", "Video model")
	clipCmd.Flags().IntVar(&clipDuration, "duration", 5, "Video duration in seconds")
	clipCmd.Flags().StringVar(&clipResolution, "resolution", "720p", "Video resolution (720p, 1080p)")
	clipCmd.Flags().StringVarP(&clipOutputPrefix, "output", "o", "", "Output prefix path")
	clipCmd.Flags().BoolVar(&clipJSONOutput, "json", false, "Output JSON format")

	RootCmd.AddCommand(clipCmd)
}

func runClip(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if clipPrompt == "" || clipSubject == "" {
		return fmt.Errorf("both --prompt and --subject are required")
	}

	if clipOutputPrefix == "" {
		clipOutputPrefix = "output/clip"
	}

	dir := filepath.Dir(clipOutputPrefix)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create output dir: %w", err)
	}

	cli := client.NewClient(getAPIKey(), getGroupID())

	opts := schemas.ClipOptions{
		ImagePrompt:   clipPrompt,
		VideoPrompt:   clipSubject,
		AspectRatio:   clipAspectRatio,
		VideoModel:    clipModel,
		Duration:      clipDuration,
		Resolution:    clipResolution,
		PollInterval:  3,
		MaxWait:       300,
		OutputPrefix:  clipOutputPrefix,
	}

	workflow := workflows.NewClipWorkflow(cli)

	if clipJSONOutput {
		result, err := workflow.Run(ctx, opts, func(stage string) {
			fmt.Fprintln(os.Stderr, stage)
		})
		if err != nil {
			return fmt.Errorf("clip failed: %w", err)
		}
		jsonOutput, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonOutput))
		return nil
	}

	result, err := workflow.Run(ctx, opts, func(stage string) {
		fmt.Println(stage)
	})
	if err != nil {
		return fmt.Errorf("clip failed: %w", err)
	}

	fmt.Printf("Image: %s\n", result.ImagePath)
	fmt.Printf("Video: %s\n", result.VideoPath)

	return nil
}
