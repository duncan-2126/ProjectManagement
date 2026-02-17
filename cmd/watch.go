package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/duncan-2126/ProjectManagement/internal/parser"
	"github.com/spf13/cobra"
)

var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for file changes and auto-scan",
	Long: `Watch the codebase for changes and automatically scan for new TODOs.
Useful for development workflows where TODOs are added frequently.

Example:
  todo watch
  todo watch --interval 30s`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get project path
		projectPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Get interval
		intervalStr, _ := cmd.Flags().GetString("interval")
		interval, err := time.ParseDuration(intervalStr)
		if err != nil {
			return fmt.Errorf("invalid interval: %w", err)
		}

		// Get exclude patterns
		exclude := []string{".git", "node_modules", "vendor", "dist", ".todo"}

		fmt.Printf("Watching %s for changes (scan every %s)...\n", projectPath, interval)
		fmt.Println("Press Ctrl+C to stop.")
		fmt.Println()

		// Initial scan
		runScan(projectPath, exclude)

		// Set up file watcher using simple polling
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		// Handle Ctrl+C
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

		lastScan := time.Now()

		for {
			select {
			case <-ticker.C:
				// Run scan
				count := runScan(projectPath, exclude)
				lastScan = time.Now()
				if count > 0 {
					fmt.Printf("[%s] Found %d new TODOs\n", lastScan.Format("15:04:05"), count)
				}
			case <-sigChan:
				fmt.Println("\nStopping watch...")
				return nil
			}
		}
	},
}

func runScan(projectPath string, exclude []string) int {
	// Open database
	db, err := database.New(projectPath)
	if err != nil {
		return 0
	}

	// Create parser
	p := parser.New(nil, exclude, nil)

	// Parse directory
	todos, err := p.ParseDir(projectPath)
	if err != nil {
		return 0
	}

	// Save new TODOs to database
	newCount := 0
	for _, t := range todos {
		exists, _ := db.TODOExists(t.Hash, t.FilePath, t.LineNumber)
		if exists {
			continue
		}

		todo := database.TODO{
			FilePath:   t.FilePath,
			LineNumber: t.LineNumber,
			Column:     t.Column,
			Type:       t.Type,
			Content:    t.Content,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
			Status:     "open",
			Priority:   "P3",
			Hash:       t.Hash,
		}

		if err := db.CreateTODO(&todo); err == nil {
			newCount++
		}
	}

	return newCount
}

func init() {
	watchCmd.Flags().StringP("interval", "i", "30s", "Scan interval (e.g., 30s, 1m, 5m)")
	rootCmd.AddCommand(watchCmd)
}
