package cmd

import (
	"fmt"
	"os"

	"github.com/duncan-2126/ProjectManagement/internal/config"
	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show TODO statistics",
	Long: `Display statistics about tracked TODOs including counts by status, type, and priority.

Example:
  todo stats`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get project path
		projectPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Open database
		db, err := database.New(projectPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		// Get stats
		stats, err := db.GetStats()
		if err != nil {
			return fmt.Errorf("failed to get stats: %w", err)
		}

		// Display stats
		fmt.Println("=== TODO Statistics ===\n")

		total := stats["total"].(int64)
		fmt.Printf("Total TODOs: %d\n\n", total)

		// By status
		fmt.Println("By Status:")
		statusMap := stats["by_status"].(map[string]int64)
		statuses := []string{"open", "in_progress", "blocked", "resolved", "wontfix", "closed"}
		for _, s := range statuses {
			if count, ok := statusMap[s]; ok {
				fmt.Printf("  %s: %d\n", s, count)
			}
		}
		fmt.Println()

		// By type
		fmt.Println("By Type:")
		typeMap := stats["by_type"].(map[string]int64)
		types := []string{"TODO", "FIXME", "HACK", "BUG", "NOTE", "XXX"}
		for _, t := range types {
			if count, ok := typeMap[t]; ok {
				fmt.Printf("  %s: %d\n", t, count)
			}
		}
		fmt.Println()

		// By priority
		fmt.Println("By Priority:")
		priorityMap := stats["by_priority"].(map[string]int64)
		priorities := []string{"P0", "P1", "P2", "P3", "P4"}
		for _, p := range priorities {
			if count, ok := priorityMap[p]; ok {
				fmt.Printf("  %s: %d\n", p, count)
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
}
