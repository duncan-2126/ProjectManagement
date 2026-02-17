package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search TODOs by content or pattern",
	Long: `Search TODOs using full-text search, regex, or field-specific queries.

	Examples:
  todo search "authentication error"          # Full-text search in content
  todo search --regex "TODO|FIXME|XXX"        # Regex pattern matching
  todo search --field content --match "login" # Search in specific field
  todo search --field file --match "**/auth*.ts"
  todo search "bug" --status open             # Combined with status filter
  todo search "performance" --priority P0     # Combined with priority filter`,
	Args: cobra.RangeArgs(0, 1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

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

		priority, _ := cmd.Flags().GetString("priority")
		if priority != "" {
			filters["priority"] = priority
		}

		assignee, _ := cmd.Flags().GetString("assignee")
		if assignee != "" {
			filters["assignee"] = assignee
		}

		author, _ := cmd.Flags().GetString("author")
		if author != "" {
			filters["author"] = author
		}

		todoType, _ := cmd.Flags().GetString("type")
		if todoType != "" {
			filters["type"] = todoType
		}

		// Get search parameters
		isRegex, _ := cmd.Flags().GetBool("regex")
		field, _ := cmd.Flags().GetString("field")
		match, _ := cmd.Flags().GetString("match")
		filePattern, _ := cmd.Flags().GetString("file")

		var query string
		if len(args) > 0 {
			query = args[0]
		}

		// Handle file pattern search
		if filePattern != "" {
			field = "file"
			match = filePattern
		}

		// Get all TODOs with basic filters first
		todos, err := db.GetTODOs(filters)
		if err != nil {
			return fmt.Errorf("failed to get TODOs: %w", err)
		}

		// Apply search filters
		var filtered []database.TODO

		if isRegex && query != "" {
			// Regex search on query across content, file_path, author
			filtered = searchByRegex(todos, query, field)
		} else if field != "" && match != "" {
			// Field-specific search
			filtered = searchByField(todos, field, match)
		} else if query != "" {
			// Default: full-text search on content
			filtered = searchFullText(todos, query)
		} else {
			filtered = todos
		}

		// Get output format
		format, _ := cmd.Flags().GetString("format")

		switch format {
		case "json":
			jsonBytes, err := json.MarshalIndent(filtered, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(jsonBytes))

		case "csv":
			fmt.Println("ID,FilePath,LineNumber,Type,Content,Status,Priority,Author,Assignee")
			for _, t := range filtered {
				fmt.Printf("%s,%s,%d,%s,\"%s\",%s,%s,%s,%s\n",
					t.ID, t.FilePath, t.LineNumber, t.Type, t.Content, t.Status, t.Priority, t.Author, t.Assignee)
			}

		default: // table
			if len(filtered) == 0 {
				fmt.Println("No TODOs found.")
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tFile\tLine\tType\tStatus\tPriority\tAssignee\tContent")
			fmt.Fprintln(w, "---\t----\t----\t----\t------\t--------\t---------\t-------")

			for _, t := range filtered {
				content := t.Content
				if len(content) > 40 {
					content = content[:37] + "..."
				}
				assignee := t.Assignee
				if assignee == "" {
					assignee = "-"
				}
				fmt.Fprintf(w, "%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\n",
					t.ID[:8], t.FilePath, t.LineNumber, t.Type, t.Status, t.Priority, assignee, content)
			}
			w.Flush()

			fmt.Printf("\nTotal: %d TODOs\n", len(filtered))
		}

		return nil
	},
}

// searchFullText performs full-text search on content, file_path, author, and assignee
func searchFullText(todos []database.TODO, query string) []database.TODO {
	query = strings.ToLower(query)
	var results []database.TODO

	for _, t := range todos {
		content := strings.ToLower(t.Content)
		filePath := strings.ToLower(t.FilePath)
		author := strings.ToLower(t.Author)
		assignee := strings.ToLower(t.Assignee)

		if strings.Contains(content, query) ||
			strings.Contains(filePath, query) ||
			strings.Contains(author, query) ||
			strings.Contains(assignee, query) {
			results = append(results, t)
		}
	}

	return results
}

// searchByRegex performs regex search on specified field or all fields
func searchByRegex(todos []database.TODO, pattern string, field string) []database.TODO {
	re, err := regexp.Compile(pattern)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid regex pattern: %v\n", err)
		return todos
	}

	var results []database.TODO

	for _, t := range todos {
		var searchable string

		switch field {
		case "content":
			searchable = t.Content
		case "file", "file_path":
			searchable = t.FilePath
		case "author":
			searchable = t.Author
		case "assignee":
			searchable = t.Assignee
		case "type":
			searchable = t.Type
		default:
			// Search all fields
			searchable = t.Content + " " + t.FilePath + " " + t.Author + " " + t.Assignee + " " + t.Type
		}

		if re.MatchString(searchable) {
			results = append(results, t)
		}
	}

	return results
}

// searchByField performs exact/contains search on a specific field
func searchByField(todos []database.TODO, field string, match string) []database.TODO {
	match = strings.ToLower(match)
	var results []database.TODO

	for _, t := range todos {
		var searchable string

		switch field {
		case "content":
			searchable = strings.ToLower(t.Content)
		case "file", "file_path":
			searchable = strings.ToLower(t.FilePath)
		case "author":
			searchable = strings.ToLower(t.Author)
		case "assignee":
			searchable = strings.ToLower(t.Assignee)
		case "type":
			searchable = strings.ToLower(t.Type)
		case "status":
			searchable = strings.ToLower(t.Status)
		case "priority":
			searchable = strings.ToLower(t.Priority)
		default:
			continue
		}

		// Support wildcards in match (simple glob matching)
		if strings.Contains(match, "*") {
			pattern := globToRegex(match)
			re, err := regexp.Compile(pattern)
			if err != nil {
				continue
			}
			if re.MatchString(searchable) {
				results = append(results, t)
			}
		} else if strings.Contains(searchable, match) {
			results = append(results, t)
		}
	}

	return results
}

// globToRegex converts simple glob patterns (* and ?) to regex
func globToRegex(pattern string) string {
	pattern = strings.ReplaceAll(pattern, ".", "\\.")
	pattern = strings.ReplaceAll(pattern, "*", ".*")
	pattern = strings.ReplaceAll(pattern, "?", ".")
	return pattern
}

func init() {
	searchCmd.Flags().BoolP("regex", "r", false, "Use regex pattern matching")
	searchCmd.Flags().StringP("field", "f", "", "Field to search (content, file_path, author, assignee, type)")
	searchCmd.Flags().StringP("match", "m", "", "Match pattern (supports wildcards for file paths)")
	searchCmd.Flags().StringP("file", "i", "", "Search by file pattern (e.g., **/auth*.ts)")
	searchCmd.Flags().StringP("status", "s", "", "Filter by status (open, in_progress, resolved, wontfix)")
	searchCmd.Flags().StringP("priority", "p", "", "Filter by priority (P0-P4)")
	searchCmd.Flags().StringP("assignee", "a", "", "Filter by assignee")
	searchCmd.Flags().StringP("author", "", "", "Filter by author")
	searchCmd.Flags().StringP("type", "t", "", "Filter by type (TODO, FIXME, HACK, BUG, NOTE, XXX)")
	searchCmd.Flags().StringP("format", "o", "table", "Output format (table, json, csv)")

	rootCmd.AddCommand(searchCmd)
}
