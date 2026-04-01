package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/minimax-ai/minimax-studio/internal/schemas"
	"github.com/minimax-ai/minimax-studio/internal/workflows"
	"github.com/spf13/cobra"
)

var (
	stitchVideos     []string
	stitchNarration  string
	stitchMusic      string
	stitchOutput     string
	stitchJSONOutput bool
)

var stitchCmd = &cobra.Command{
	Use:   "stitch",
	Short: "Stitch video segments with narration and music",
	Long:  `Concatenate video segments and compose with narration and background music.`,
	RunE:  runStitch,
}

func init() {
	stitchCmd.Flags().StringArrayVarP(&stitchVideos, "video", "v", []string{}, "Input video paths (can specify multiple)")
	stitchCmd.Flags().StringVarP(&stitchNarration, "narration", "n", "", "Narration audio path")
	stitchCmd.Flags().StringVarP(&stitchMusic, "music", "m", "", "Background music path")
	stitchCmd.Flags().StringVarP(&stitchOutput, "output", "o", "", "Output video path")
	stitchCmd.Flags().BoolVar(&stitchJSONOutput, "json", false, "Output JSON format")

	RootCmd.AddCommand(stitchCmd)
}

func runStitch(cmd *cobra.Command, args []string) error {
	if len(stitchVideos) == 0 {
		return fmt.Errorf("at least one video is required")
	}
	if stitchNarration == "" {
		return fmt.Errorf("narration is required")
	}
	if stitchOutput == "" {
		return fmt.Errorf("output path is required")
	}

	opts := schemas.StitchOptions{
		VideoPaths:    stitchVideos,
		NarrationPath: stitchNarration,
		MusicPath:     stitchMusic,
		OutputPath:    stitchOutput,
	}

	workflow := workflows.NewStitchWorkflow()

	if stitchJSONOutput {
		result, err := workflow.Run(opts, func(stage string) {
			fmt.Fprintln(os.Stderr, stage)
		})
		if err != nil {
			return fmt.Errorf("stitch failed: %w", err)
		}
		jsonOutput, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonOutput))
		return nil
	}

	result, err := workflow.Run(opts, func(stage string) {
		fmt.Println(stage)
	})
	if err != nil {
		return fmt.Errorf("stitch failed: %w", err)
	}

	fmt.Printf("Stitched video: %s\n", result.StitchedVideoPath)
	fmt.Printf("Padded video: %s\n", result.PaddedVideoPath)
	fmt.Printf("Final video: %s\n", result.FinalVideoPath)

	return nil
}
