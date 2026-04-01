package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	planTheme        string
	planSceneCount   int
	planSceneDuration int
	planLanguage     string
	planOutputDir    string
)

var planCmd = &cobra.Command{
	Use:   "plan",
	Short: "Generate video scene plan",
	Long:  `Generate scene plan with narration and visual descriptions.`,
	RunE:  runPlan,
}

func init() {
	planCmd.Flags().StringVarP(&planTheme, "theme", "t", "", "Video theme")
	planCmd.Flags().IntVar(&planSceneCount, "scene-count", 1, "Number of scenes")
	planCmd.Flags().IntVar(&planSceneDuration, "scene-duration", 6, "Duration of each scene in seconds")
	planCmd.Flags().StringVar(&planLanguage, "language", "zh", "Language (zh/en)")
	planCmd.Flags().StringVarP(&planOutputDir, "output", "o", "", "Output directory")

	RootCmd.AddCommand(planCmd)
}

func runPlan(cmd *cobra.Command, args []string) error {
	if planTheme == "" {
		return fmt.Errorf("--theme is required")
	}
	if planOutputDir == "" {
		return fmt.Errorf("--output is required")
	}

	fmt.Printf("Theme: %s\n", planTheme)
	fmt.Printf("Scene count: %d\n", planSceneCount)
	fmt.Printf("Output dir: %s\n", planOutputDir)
	fmt.Println("\nPlan workflow not yet implemented in Go version.")

	return nil
}
