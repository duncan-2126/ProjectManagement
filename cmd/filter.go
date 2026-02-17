package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Manage saved filters",
	Long:  "Save and manage complex filter queries",
}

var filterSaveCmd = &cobra.Command{
	Use:   "save <name>",
	Short: "Save a filter query from flags",
	Long: `Save a complex filter query for later use using command-line flags.

Example:
  todo filter save my-open --status open --priority P0
  todo filter save bugs --type BUG
  todo filter save my-tasks --assignee me
  todo filter save auth-work --file "**/auth*.ts" --status in_progress`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		name := args[0]

		// Build query from flags
		var queryParts []string

		status, _ := cmd.Flags().GetString("status")
		if status != "" {
			queryParts = append(queryParts, "status="+status)
		}

		priority, _ := cmd.Flags().GetString("priority")
		if priority != "" {
			queryParts = append(queryParts, "priority="+priority)
		}

		assignee, _ := cmd.Flags().GetString("assignee")
		if assignee != "" {
			queryParts = append(queryParts, "assignee="+assignee)
		}

		author, _ := cmd.Flags().GetString("author")
		if author != "" {
			queryParts = append(queryParts, "author="+author)
		}

		todoType, _ := cmd.Flags().GetString("type")
		if todoType != "" {
			queryParts = append(queryParts, "type="+todoType)
		}

		filePath, _ := cmd.Flags().GetString("file")
		if filePath != "" {
			queryParts = append(queryParts, "file="+filePath)
		}

		query := strings.Join(queryParts, "&")
		if query == "" {
			query = "all=true"
		}

		filter, err := db.CreateSavedFilter(name, query)
		if err != nil {
			return fmt.Errorf("failed to save filter: %w", err)
		}

		fmt.Printf("Saved filter: %s -> %s\n", filter.Name, filter.Query)
		return nil
	},
}

var filterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved filters",
	Long:  `List all saved filters.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		filters, err := db.GetSavedFilters()
		if err != nil {
			return err
		}

		if len(filters) == 0 {
			fmt.Println("No saved filters.")
			return nil
		}

		fmt.Println("Saved Filters:")
		for _, f := range filters {
			fmt.Printf("  @%s: %s\n", f.Name, f.Query)
		}
		return nil
	},
}

var filterDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a saved filter",
	Long:  `Delete a saved filter.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		name := args[0]
		if err := db.DeleteSavedFilter(name); err != nil {
			return err
		}

		fmt.Printf("Deleted filter: %s\n", name)
		return nil
	},
}

var filterRunCmd = &cobra.Command{
	Use:   "run <name>",
	Short: "Run a saved filter",
	Long:  `Run a saved filter and show results.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		name := args[0]

		filter, err := db.GetSavedFilter(name)
		if err != nil {
			return fmt.Errorf("filter not found: %s", name)
		}

		// Parse query string into filters
		filters := parseQuery(filter.Query)

		todos, err := db.GetTODOs(filters)
		if err != nil {
			return err
		}

		fmt.Printf("Results for filter '%s':\n\n", name)
		for _, t := range todos {
			content := t.Content
			if len(content) > 60 {
				content = content[:57] + "..."
			}
			fmt.Printf("  [%s] %s - %s\n", t.Priority, t.Status, content)
		}
		fmt.Printf("\nTotal: %d TODOs\n", len(todos))
		return nil
	},
}

// parseQuery parses a query string into a filters map
// Supports: status=open&priority=P0&assignee=me&type=BUG&file=path&author=name
func parseQuery(query string) map[string]interface{} {
	filters := make(map[string]interface{})

	if query == "all=true" {
		return filters
	}

	parts := strings.Split(query, "&")
	for _, part := range parts {
		if part == "" {
			continue
		}
		kv := strings.SplitN(part, "=", 2)
		if len(kv) == 2 {
			key := kv[0]
			value := kv[1]
			switch key {
			case "status", "priority", "assignee", "author", "type", "file":
				if value != "" {
					filters[key] = value
				}
			}
		}
	}

	return filters
}

func init() {
	filterSaveCmd.Flags().StringP("status", "s", "", "Filter by status (open, in_progress, resolved, wontfix)")
	filterSaveCmd.Flags().StringP("priority", "p", "", "Filter by priority (P0-P4)")
	filterSaveCmd.Flags().StringP("assignee", "a", "", "Filter by assignee")
	filterSaveCmd.Flags().StringP("author", "", "", "Filter by author")
	filterSaveCmd.Flags().StringP("type", "t", "", "Filter by type (TODO, FIXME, HACK, BUG, NOTE, XXX)")
	filterSaveCmd.Flags().StringP("file", "f", "", "Filter by file path")

	filterCmd.AddCommand(filterSaveCmd)
	filterCmd.AddCommand(filterListCmd)
	filterCmd.AddCommand(filterDeleteCmd)
	filterCmd.AddCommand(filterRunCmd)
	rootCmd.AddCommand(filterCmd)
}

// ExportFilters exports TODOs matching a complex filter
var exportFiltersCmd = &cobra.Command{
	Use:   "export-filters",
	Short: "Export with advanced filters",
	Long:  `Export TODOs matching complex filter criteria.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		// Get all TODOs and apply client-side filtering
		todos, _ := db.GetTODOs(nil)

		// Apply filter flags
		status, _ := cmd.Flags().GetString("status")
		priority, _ := cmd.Flags().GetString("priority")
		todoType, _ := cmd.Flags().GetString("type")
		assignee, _ := cmd.Flags().GetString("assignee")

		var filtered []database.TODO
		for _, t := range todos {
			if status != "" && t.Status != status {
				continue
			}
			if priority != "" && t.Priority != priority {
				continue
			}
			if todoType != "" && t.Type != todoType {
				continue
			}
			if assignee != "" && t.Assignee != assignee {
				continue
			}
			filtered = append(filtered, t)
		}

		format, _ := cmd.Flags().GetString("format")
		switch format {
		case "json":
			data, _ := json.MarshalIndent(filtered, "", "  ")
			fmt.Println(string(data))
		default:
			fmt.Printf("Found %d TODOs matching criteria\n", len(filtered))
		}
		return nil
	},
}

func init() {
	exportFiltersCmd.Flags().String("status", "", "Filter by status")
	exportFiltersCmd.Flags().String("priority", "", "Filter by priority")
	exportFiltersCmd.Flags().String("type", "", "Filter by type")
	exportFiltersCmd.Flags().String("assignee", "", "Filter by assignee")
	exportFiltersCmd.Flags().String("format", "table", "Output format")
	rootCmd.AddCommand(exportFiltersCmd)
}
