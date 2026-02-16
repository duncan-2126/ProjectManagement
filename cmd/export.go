package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export TODOs to various formats",
	Long: `Export TODOs to different formats for integration with other tools.

Examples:
  todo export                    # Export all TODOs
  todo export --format json     # Export as JSON
  todo export --format csv      # Export as CSV
  todo export --format markdown # Export as Markdown table
  todo export --status open     # Export only open TODOs`,
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

		// Build filters
		filters := make(map[string]interface{})
		status, _ := cmd.Flags().GetString("status")
		if status != "" {
			filters["status"] = status
		}

		// Get TODOs
		todos, err := db.GetTODOs(filters)
		if err != nil {
			return fmt.Errorf("failed to get TODOs: %w", err)
		}

		// Get format
		format, _ := cmd.Flags().GetString("format")

		switch format {
		case "json":
			return exportJSON(todos)
		case "csv":
			return exportCSV(todos)
		case "markdown":
			return exportMarkdown(todos)
		default:
			return exportJSON(todos)
		}
	},
}

func exportJSON(todos []database.TODO) error {
	output, err := json.MarshalIndent(todos, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

func exportCSV(todos []database.TODO) error {
	fmt.Println("ID,FilePath,LineNumber,Type,Status,Priority,Category,Assignee,Author,Content")
	for _, t := range todos {
		fmt.Printf("%s,%s,%d,%s,%s,%s,%s,%s,%s,\"%s\"\n",
			t.ID, t.FilePath, t.LineNumber, t.Type, t.Status, t.Priority,
			t.Category, t.Assignee, t.Author, t.Content)
	}
	return nil
}

func exportMarkdown(todos []database.TODO) error {
	fmt.Println("# TODO Export\n")
	fmt.Println("| ID | File | Line | Type | Status | Priority | Content |")
	fmt.Println("|---|---|---|---|---|---|---|")
	for _, t := range todos {
		content := t.Content
		if len(content) > 50 {
			content = content[:47] + "..."
		}
		fmt.Printf("| %s | %s | %d | %s | %s | %s | %s |\n",
			t.ID[:8], t.FilePath, t.LineNumber, t.Type, t.Status, t.Priority, content)
	}
	fmt.Printf("\n*Total: %d TODOs*\n", len(todos))
	return nil
}

func init() {
	exportCmd.Flags().StringP("format", "f", "json", "Export format (json, csv, markdown)")
	exportCmd.Flags().StringP("status", "s", "", "Filter by status")
	rootCmd.AddCommand(exportCmd)
}
