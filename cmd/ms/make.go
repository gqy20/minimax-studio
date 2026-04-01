package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	makeTheme        string
	makeSceneCount   int
	makeSceneDuration int
	makeOutputDir    string
)

var makeCmd = &cobra.Command{
	Use:   "make",
	Short: "Create a full video from theme",
	Long:  `End-to-end video generation: plan, clip, voice, music, and stitch.`,
	RunE:  runMake,
}

func init() {
	makeCmd.Flags().StringVarP(&makeTheme, "theme", "t", "", "Video theme")
	makeCmd.Flags().IntVar(&makeSceneCount, "scene-count", 1, "Number of scenes")
	makeCmd.Flags().IntVar(&makeSceneDuration, "scene-duration", 6, "Duration of each scene")
	makeCmd.Flags().StringVarP(&makeOutputDir, "output", "o", "", "Output directory")

	RootCmd.AddCommand(makeCmd)
}

func runMake(cmd *cobra.Command, args []string) error {
	if makeTheme == "" {
		return fmt.Errorf("--theme is required")
	}
	if makeOutputDir == "" {
		return fmt.Errorf("--output is required")
	}

	fmt.Printf("Theme: %s\n", makeTheme)
	fmt.Printf("Scene count: %d\n", makeSceneCount)
	fmt.Printf("Output dir: %s\n", makeOutputDir)
	fmt.Println("\nFull make workflow not yet implemented in Go version.")

	return nil
}
