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
	planTheme         string
	planSceneCount    int
	planSceneDuration int
	planLanguage      string
	planOutputDir     string
	planTextModel     string
	planMaxTokens     int
	planJSONOutput    bool
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Generate video scene plan",
	Long:  `Generate scene plan with narration and visual descriptions using MiniMax text model.`,
	RunE:  runPlan,
}

func init() {
	planCmd.Flags().StringVarP(&planTheme, "theme", "t", "", "Video theme")
	planCmd.Flags().IntVar(&planSceneCount, "scene-count", 1, "Number of scenes")
	planCmd.Flags().IntVar(&planSceneDuration, "scene-duration", 6, "Duration of each scene in seconds")
	planCmd.Flags().StringVar(&planLanguage, "language", "zh", "Language (zh/en)")
	planCmd.Flags().StringVar(&planTextModel, "text-model", "MiniMax-M2.7-highspeed", "Text model for planning")
	planCmd.Flags().IntVar(&planMaxTokens, "max-tokens", 4096, "Max tokens for text generation")
	planCmd.Flags().StringVarP(&planOutputDir, "output", "o", "", "Output directory")
	planCmd.Flags().BoolVar(&planJSONOutput, "json", false, "Output JSON format")

	RootCmd.AddCommand(planCmd)
}

func runPlan(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if planTheme == "" {
		return fmt.Errorf("--theme is required")
	}
	if planOutputDir == "" {
		return fmt.Errorf("--output is required")
	}

	cli := client.NewClient(getAPIKey())

	opts := schemas.PlanOptions{
		Theme:         planTheme,
		SceneCount:    planSceneCount,
		SceneDuration: planSceneDuration,
		Language:      planLanguage,
		TextModel:     planTextModel,
		TextMaxTokens: planMaxTokens,
		OutputDir:     planOutputDir,
	}

	workflow := workflows.NewPlanWorkflow(cli)

	if planJSONOutput {
		result, err := workflow.Run(ctx, opts, func(stage string) {
			fmt.Fprintln(os.Stderr, stage)
		})
		if err != nil {
			return fmt.Errorf("plan failed: %w", err)
		}
		jsonOutput, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonOutput))
		return nil
	}

	result, err := workflow.Run(ctx, opts, func(stage string) {
		fmt.Println(stage)
	})
	if err != nil {
		return fmt.Errorf("plan failed: %w", err)
	}

	fmt.Printf("\nPlan: %s\n", result.PlanPath)
	fmt.Printf("Narration: %s\n", result.NarrationPath)

	return nil
}
