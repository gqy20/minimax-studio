package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/minimax-ai/minimax-studio/internal/client"
	"github.com/spf13/cobra"
)

var quotaJSONOutput bool

var quotaCmd = &cobra.Command{
	Use:   "quota",
	Short: "Query API quota information",
	Long:  `Query current API key quota information.`,
	RunE:  runQuota,
}

func init() {
	quotaCmd.Flags().BoolVar(&quotaJSONOutput, "json", false, "Output JSON format")

	RootCmd.AddCommand(quotaCmd)
}

func runQuota(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	cli := client.NewClient(getAPIKey())

	result, err := cli.GetQuota(ctx)
	if err != nil {
		return fmt.Errorf("failed to get quota: %w", err)
	}

	if quotaJSONOutput {
		jsonOutput, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(jsonOutput))
		return nil
	}

	fmt.Println("=== Quota Information ===")
	for _, entry := range result {
		jsonOutput, _ := json.MarshalIndent(entry, "  ", "  ")
		fmt.Printf("  %s\n", string(jsonOutput))
	}

	return nil
}
