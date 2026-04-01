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
	voiceText       string
	voiceOutputPath string
	voiceVoiceID    string
	voiceModel      string
	voiceFormat     string
	voiceJSONOutput bool
)

var voiceCmd = &cobra.Command{
	Use:   "voice",
	Short: "Synthesize speech from text",
	Long:  `Convert text to speech using MiniMax TTS.`,
	RunE:  runVoice,
}

func init() {
	voiceCmd.Flags().StringVarP(&voiceText, "text", "t", "", "Text to synthesize")
	voiceCmd.Flags().StringVarP(&voiceOutputPath, "output", "o", "", "Output audio path")
	voiceCmd.Flags().StringVar(&voiceVoiceID, "voice-id", "male-qn-qingse", "Voice ID")
	voiceCmd.Flags().StringVar(&voiceModel, "model", "speech-2.8-hd", "TTS model")
	voiceCmd.Flags().StringVar(&voiceFormat, "format", "mp3", "Audio format (mp3/wav/flac)")
	voiceCmd.Flags().BoolVar(&voiceJSONOutput, "json", false, "Output JSON format")

	RootCmd.AddCommand(voiceCmd)
}

func runVoice(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	if voiceText == "" {
		return fmt.Errorf("--text is required")
	}
	if voiceOutputPath == "" {
		return fmt.Errorf("--output is required")
	}

	cli := client.NewClient(getAPIKey(), getGroupID())

	opts := schemas.VoiceOptions{
		Text:        voiceText,
		OutputPath:  voiceOutputPath,
		VoiceID:     voiceVoiceID,
		TTSModel:    voiceModel,
		AudioFormat: voiceFormat,
	}

	workflow := workflows.NewVoiceWorkflow(cli)

	if voiceJSONOutput {
		result, err := workflow.Run(ctx, opts, func(stage string) {
			fmt.Fprintln(os.Stderr, stage)
		})
		if err != nil {
			return fmt.Errorf("voice failed: %w", err)
		}
		jsonOutput, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonOutput))
		return nil
	}

	result, err := workflow.Run(ctx, opts, func(stage string) {
		fmt.Println(stage)
	})
	if err != nil {
		return fmt.Errorf("voice failed: %w", err)
	}

	fmt.Printf("Audio: %s\n", result.OutputPath)

	return nil
}
