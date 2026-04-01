package main

import (
	"log"

	"github.com/minimax-ai/minimax-studio/internal/api"
	"github.com/spf13/cobra"
)

var (
	serverPort      string
	serverOutputDir string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start HTTP API server",
	Long:  `Start the MiniMax Studio HTTP API server for frontend integration.`,
	RunE:  runServer,
}

func init() {
	serverCmd.Flags().StringVar(&serverPort, "port", "8080", "Server port")
	serverCmd.Flags().StringVar(&serverOutputDir, "output-dir", "./output", "Output directory")

	RootCmd.AddCommand(serverCmd)
}

func runServer(cmd *cobra.Command, args []string) error {
	apiKey := getAPIKey()
	groupID := getGroupID()

	s := api.NewServer(serverOutputDir, apiKey, groupID)
	addr := ":" + serverPort
	log.Printf("Starting API server on %s", addr)
	return s.Run(addr)
}
