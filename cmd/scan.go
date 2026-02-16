package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/duncan-2126/ProjectManagement/internal/config"
	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/duncan-2126/ProjectManagement/internal/parser"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var scanCmd = &cobra.Command{
	Use:   "scan [path]",
	Short: "Scan codebase for TODO comments",
	Long: `Scans the specified directory (or current directory) for TODO comments
and stores them in the database for tracking.

Example:
  todo scan
  todo scan ./src
  todo scan --exclude node_modules --exclude vendor`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := cmd.Context().(*config.Config)

		// Get path to scan
		path := "."
		if len(args) > 0 {
			path = args[0]
		}

		// Resolve to absolute path
		absPath, err := resolvePath(path)
		if err != nil {
			return fmt.Errorf("failed to resolve path: %w", err)
		}

		// Open database
		db, err := database.New(absPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		// Get exclude patterns from flags or config
		exclude := viper.GetStringSlice("exclude")
		if len(exclude) == 0 {
			exclude = cfg.ExcludePatterns
		}

		// Create parser
		p := parser.New(nil, exclude, nil)

		fmt.Printf("Scanning %s...\n", absPath)

		// Parse directory
		todos, err := p.ParseDir(absPath)
		if err != nil {
			return fmt.Errorf("failed to scan: %w", err)
		}

		// Save TODOs to database
		newCount := 0
		existingCount := 0

		for _, t := range todos {
			// Check if already exists
			exists, err := db.TODOExists(t.Hash, t.FilePath, t.LineNumber)
			if err != nil {
				continue
			}

			if exists {
				existingCount++
				continue
			}

			// Create new TODO
			todo := database.TODO{
				FilePath:   t.FilePath,
				LineNumber: t.LineNumber,
				Column:     t.Column,
				Type:       t.Type,
				Content:    t.Content,
				Author:     t.Author,
				Email:      t.Email,
				CreatedAt:  t.CreatedAt,
				UpdatedAt:  t.CreatedAt,
				Status:     "open",
				Priority:   "P3",
				Hash:       t.Hash,
			}

			if err := db.CreateTODO(&todo); err == nil {
				newCount++
			}
		}

		fmt.Printf("Scan complete!\n")
		fmt.Printf("  Found: %d TODOs\n", len(todos))
		fmt.Printf("  New: %d\n", newCount)
		fmt.Printf("  Existing: %d\n", existingCount)

		return nil
	},
}

func resolvePath(path string) (string, error) {
	if path == "." {
		return os.Getwd()
	}
	return filepath.Abs(path)
}

func init() {
	scanCmd.Flags().StringSliceP("exclude", "e", nil, "Exclude patterns (can be repeated)")
	rootCmd.AddCommand(scanCmd)
}
