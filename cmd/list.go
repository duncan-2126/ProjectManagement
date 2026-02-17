package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var listCmd = &cobra.Command{
	Use:   "list [@<filter_name>]",
	Short: "List tracked TODOs",
	Long: `List all tracked TODOs with optional filtering.

Examples:
  todo list                    # List all TODOs
  todo list --status open      # List only open TODOs
  todo list --type FIXME       # List only FIXMEs
  todo list @my-filter         # Use saved filter
  todo list --stale            # List stale TODOs
  todo list --format json      # Output as JSON`,
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

		// Check for stale flag
		staleFlag, _ := cmd.Flags().GetBool("stale")

		// Check for @filter syntax in args
		var filterName string
		for _, arg := range args {
			if strings.HasPrefix(arg, "@") {
				filterName = strings.TrimPrefix(arg, "@")
				break
			}
		}

		// If filter name provided, load the saved filter
		if filterName != "" {
			filter, err := db.GetSavedFilter(filterName)
			if err != nil {
				return fmt.Errorf("filter not found: %s", filterName)
			}
			// Parse saved filter query into filters
			filters = parseQuery(filter.Query)
			fmt.Printf("Using filter: @%s -> %s\n", filterName, filter.Query)
		}

		// Apply command-line flags (override saved filter)
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

		assignee, _ := cmd.Flags().GetString("assignee")
		if assignee != "" {
			filters["assignee"] = assignee
		}

		// Get TODOs
		var todos []database.TODO
		if staleFlag {
			// Get stale days from config or use default
			staleDays := viper.GetInt("stale.days_since_update")
			if staleDays == 0 {
				staleDays = 14
			}
			todos, err = db.GetStaleTODOs(staleDays)
			if err != nil {
				return fmt.Errorf("failed to get stale TODOs: %w", err)
			}
		} else {
			todos, err = db.GetTODOs(filters)
			if err != nil {
				return fmt.Errorf("failed to get TODOs: %w", err)
			}
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
			fmt.Println("ID,FilePath,LineNumber,Type,Content,Status,Priority,Author,Assignee")
			for _, t := range todos {
				fmt.Printf("%s,%s,%d,%s,\"%s\",%s,%s,%s,%s\n",
					t.ID, t.FilePath, t.LineNumber, t.Type, t.Content, t.Status, t.Priority, t.Author, t.Assignee)
			}

		default: // table
			if len(todos) == 0 {
				fmt.Println("No TODOs found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tFile\tLine\tType\tStatus\tPriority\tAssignee\tContent")
			fmt.Fprintln(w, "---\t----\t----\t----\t------\t--------\t---------\t-------")

			for _, t := range todos {
				// Truncate content if too long
				content := t.Content
				if len(content) > 40 {
					content = content[:37] + "..."
				}
				assigneeStr := t.Assignee
				if assigneeStr == "" {
					assigneeStr = "-"
				}
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
					t.ID[:8], t.FilePath, t.LineNumber, t.Type, t.Status, t.Priority, assigneeStr, content)
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
	listCmd.Flags().StringP("assignee", "", "", "Filter by assignee")
	listCmd.Flags().StringP("format", "o", "table", "Output format (table, json, csv)")
	listCmd.Flags().BoolP("stale", "", false, "Show stale TODOs (no update in configured days)")

	rootCmd.AddCommand(listCmd)
}
