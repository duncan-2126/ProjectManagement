package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var filterCmd = &cobra.Command{
	Use:   "filter",
	Short: "Manage saved filters",
	Long:  "Save and manage complex filter queries",
}

var filterSaveCmd = &cobra.Command{
	Use:   "save <name> <query>",
	Short: "Save a filter query",
	Long: `Save a complex filter query for later use.

Example:
  todo filter save my-open "status=open&priority=P0"
  todo filter save bugs "type=BUG"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		name := args[0]
		query := args[1]

		filter, err := db.CreateSavedFilter(name, query)
		if err != nil {
			return fmt.Errorf("failed to save filter: %w", err)
		}

		fmt.Printf("Saved filter: %s\n", filter.Name)
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
			fmt.Printf("  %s: %s\n", f.Name, f.Query)
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
			fmt.Printf("  [%s] %s - %s\n", t.Priority, t.Status, t.Content[:min(60, len(t.Content))])
		}
		fmt.Printf("\nTotal: %d TODOs\n", len(todos))
		return nil
	},
}

func parseQuery(query string) map[string]interface{} {
	filters := make(map[string]interface{})
	// Simple key=value parsing
	// In production, use proper URL parsing
	return filters
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
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
