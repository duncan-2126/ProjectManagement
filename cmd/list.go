package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/duncan-2126/ProjectManagement/internal/config"
	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tracked TODOs",
	Long: `List all tracked TODOs with optional filtering.

Examples:
  todo list                    # List all TODOs
  todo list --status open      # List only open TODOs
  todo list --type FIXME       # List only FIXMEs
  todo list --format json      # Output as JSON`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := cmd.Context().(*config.Config)

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

		todoType, _ := cmd.Flags().GetString("type")
		if todoType != "" {
			filters["type"] = todoType
		}

		author, _ := cmd.Flags().GetString("author")
		if author != "" {
			filters["author"] = author
		}

		filePath, _ := cmd.Flags().GetString("file")
		if filePath != "" {
			filters["file_path"] = filePath
		}

		priority, _ := cmd.Flags().GetString("priority")
		if priority != "" {
			filters["priority"] = priority
		}

		// Get TODOs
		todos, err := db.GetTODOs(filters)
		if err != nil {
			return fmt.Errorf("failed to get TODOs: %w", err)
		}

		// Get output format
		format, _ := cmd.Flags().GetString("format")

		switch format {
		case "json":
			jsonBytes, err := json.MarshalIndent(todos, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(jsonBytes))

		case "csv":
			fmt.Println("ID,FilePath,LineNumber,Type,Content,Status,Priority,Author")
			for _, t := range todos {
				fmt.Printf("%s,%s,%d,%s,\"%s\",%s,%s,%s\n",
					t.ID, t.FilePath, t.LineNumber, t.Type, t.Content, t.Status, t.Priority, t.Author)
			}

		default: // table
			if len(todos) == 0 {
				fmt.Println("No TODOs found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tFile\tLine\tType\tStatus\tPriority\tContent")
			fmt.Fprintln(w, "---\t----\t----\t----\t------\t--------\t-------")

			for _, t := range todos {
				// Truncate content if too long
				content := t.Content
				if len(content) > 50 {
					content = content[:47] + "..."
				}
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\t%s\n",
					t.ID[:8], t.FilePath, t.LineNumber, t.Type, t.Status, t.Priority, content)
			}
			w.Flush()

			fmt.Printf("\nTotal: %d TODOs\n", len(todos))
		}

		return nil
	},
}

func init() {
	listCmd.Flags().StringP("status", "s", "", "Filter by status (open, in_progress, resolved, wontfix)")
	listCmd.Flags().StringP("type", "t", "", "Filter by type (TODO, FIXME, HACK, BUG, NOTE, XXX)")
	listCmd.Flags().StringP("author", "a", "", "Filter by author")
	listCmd.Flags().StringP("file", "f", "", "Filter by file path")
	listCmd.Flags().StringP("priority", "p", "", "Filter by priority (P0-P4)")
	listCmd.Flags().StringP("format", "o", "table", "Output format (table, json, csv)")

	rootCmd.AddCommand(listCmd)
}
