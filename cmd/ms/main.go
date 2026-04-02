package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	apiKey    string
	groupID   string
	Version   string
	BuildTime string
)

// RootCmd 根命令
var RootCmd = &cobra.Command{
	Use:   "ms",
	Short: "MiniMax Studio CLI",
	Long:  `MiniMax Studio CLI for clip, planning, voice, music, stitch, and full video workflows.`,
}

func init() {
	RootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "MiniMax API Key (or set MINIMAX_API_KEY env)")
	RootCmd.PersistentFlags().StringVar(&groupID, "group-id", "", "MiniMax Group ID (or set MINIMAX_GROUP_ID env)")
}

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func getAPIKey() string {
	if apiKey != "" {
		return apiKey
	}
	return os.Getenv("MINIMAX_API_KEY")
}

func getGroupID() string {
	if groupID != "" {
		return groupID
	}
	return os.Getenv("MINIMAX_GROUP_ID")
}
