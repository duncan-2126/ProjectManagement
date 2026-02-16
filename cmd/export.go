package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"time"

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
  todo export --format github   # Export as GitHub Issues JSON
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
		case "github":
			return exportGitHub(todos, db)
		case "jira":
			return exportJira(todos)
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

// GitHubIssue represents a GitHub Issue for export
type GitHubIssue struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	State     string   `json:"state"`
	Labels    []string `json:"labels"`
	Assignee  string   `json:"assignee,omitempty"`
	Priority  string   `json:"priority,omitempty"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

func exportGitHub(todos []database.TODO, db *database.DB) error {
	issues := make([]GitHubIssue, 0, len(todos))

	for _, t := range todos {
		// Get tags for this TODO
		tags, _ := db.GetTagsForTODO(t.ID)
		labels := make([]string, 0, len(tags)+2)

		// Add priority as label
		if t.Priority != "" {
			labels = append(labels, t.Priority)
		}

		// Add category as label
		if t.Category != "" {
			labels = append(labels, t.Category)
		}

		// Add tags
		for _, tag := range tags {
			labels = append(labels, tag.Name)
		}

		// Map status to GitHub state
		state := "open"
		if t.Status == "closed" || t.Status == "resolved" || t.Status == "wontfix" {
			state = "closed"
		}

		// Truncate title to 256 chars (GitHub limit)
		title := t.Content
		if len(title) > 256 {
			title = title[:253] + "..."
		}

		// Build body with metadata
		body := fmt.Sprintf("**Source:** %s:%d\n**Type:** %s\n**Status:** %s\n**Priority:** %s\n**Assignee:** %s\n**Author:** %s\n\n---\n\n%s",
			t.FilePath, t.LineNumber, t.Type, t.Status, t.Priority, t.Assignee, t.Author, t.Content)

		issue := GitHubIssue{
			Title:     title,
			Body:      body,
			State:     state,
			Labels:    labels,
			Assignee:  t.Assignee,
			Priority:  t.Priority,
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
			UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
		}
		issues = append(issues, issue)
	}

	output, err := json.MarshalIndent(issues, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

// JiraIssue represents a Jira Issue for CSV export
type JiraIssue struct {
	Summary     string `json:"Summary"`
	Description string `json:"Description"`
	Priority    string `json:"Priority"`
	Assignee    string `json:"Assignee"`
	DueDate     string `json:"Due Date"`
	Status      string `json:"Status"`
	Type        string `json:"Type"`
}

func exportJira(todos []database.TODO) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	headers := []string{"Summary", "Description", "Priority", "Assignee", "Due Date", "Status", "Type"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write rows
	for _, t := range todos {
		// Map priority to Jira format
		priority := t.Priority
		if priority == "" {
			priority = "Medium"
		}

		// Format due date
		dueDate := ""
		if t.DueDate != nil {
			dueDate = t.DueDate.Format("2006-01-02")
		}

		// Map status
		status := t.Status
		if status == "wontfix" {
			status = "Won't Fix"
		}

		// Truncate summary (Jira limit is 255)
		summary := t.Content
		if len(summary) > 255 {
			summary = summary[:252] + "..."
		}

		// Build description
		description := fmt.Sprintf("Source: %s:%d\nType: %s\nAuthor: %s\n\n%s",
			t.FilePath, t.LineNumber, t.Type, t.Author, t.Content)

		row := []string{
			summary,
			description,
			priority,
			t.Assignee,
			dueDate,
			status,
			t.Category,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// GitHubIssue represents a GitHub Issue for export
type GitHubIssue struct {
	Title     string   `json:"title"`
	Body      string   `json:"body"`
	State     string   `json:"state"`
	Labels    []string `json:"labels"`
	Assignee  string   `json:"assignee,omitempty"`
	Priority  string   `json:"priority,omitempty"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

func exportGitHub(todos []database.TODO, db *database.DB) error {
	issues := make([]GitHubIssue, 0, len(todos))

	for _, t := range todos {
		// Get tags for this TODO
		tags, _ := db.GetTagsForTODO(t.ID)
		labels := make([]string, 0, len(tags)+2)

		// Add priority as label
		if t.Priority != "" {
			labels = append(labels, t.Priority)
		}

		// Add category as label
		if t.Category != "" {
			labels = append(labels, t.Category)
		}

		// Add tags
		for _, tag := range tags {
			labels = append(labels, tag.Name)
		}

		// Map status to GitHub state
		state := "open"
		if t.Status == "closed" || t.Status == "resolved" || t.Status == "wontfix" {
			state = "closed"
		}

		// Truncate title to 256 chars (GitHub limit)
		title := t.Content
		if len(title) > 256 {
			title = title[:253] + "..."
		}

		// Build body with metadata
		body := fmt.Sprintf("**Source:** %s:%d\n**Type:** %s\n**Status:** %s\n**Priority:** %s\n**Assignee:** %s\n**Author:** %s\n\n---\n\n%s",
			t.FilePath, t.LineNumber, t.Type, t.Status, t.Priority, t.Assignee, t.Author, t.Content)

		issue := GitHubIssue{
			Title:     title,
			Body:      body,
			State:     state,
			Labels:    labels,
			Assignee:  t.Assignee,
			Priority:  t.Priority,
			CreatedAt: t.CreatedAt.Format(time.RFC3339),
			UpdatedAt: t.UpdatedAt.Format(time.RFC3339),
		}
		issues = append(issues, issue)
	}

	output, err := json.MarshalIndent(issues, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

// JiraIssueCSV represents a Jira Issue for CSV export
type JiraIssueCSV struct {
	Summary     string `json:"Summary"`
	Description string `json:"Description"`
	Priority    string `json:"Priority"`
	Assignee    string `json:"Assignee"`
	DueDate     string `json:"Due Date"`
	Status      string `json:"Status"`
	Type        string `json:"Type"`
}

func exportJira(todos []database.TODO) error {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	// Write header
	headers := []string{"Summary", "Description", "Priority", "Assignee", "Due Date", "Status", "Type"}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write rows
	for _, t := range todos {
		// Map priority to Jira format
		priority := t.Priority
		if priority == "" {
			priority = "Medium"
		}

		// Format due date
		dueDate := ""
		if t.DueDate != nil {
			dueDate = t.DueDate.Format("2006-01-02")
		}

		// Map status
		status := t.Status
		if status == "wontfix" {
			status = "Won't Fix"
		}

		// Truncate summary (Jira limit is 255)
		summary := t.Content
		if len(summary) > 255 {
			summary = summary[:252] + "..."
		}

		// Build description
		description := fmt.Sprintf("Source: %s:%d\nType: %s\nAuthor: %s\n\n%s",
			t.FilePath, t.LineNumber, t.Type, t.Author, t.Content)

		row := []string{
			summary,
			description,
			priority,
			t.Assignee,
			dueDate,
			status,
			t.Category,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

func init() {
	exportCmd.Flags().StringP("format", "f", "json", "Export format (json, csv, markdown, github, jira)")
	exportCmd.Flags().StringP("status", "s", "", "Filter by status")
	rootCmd.AddCommand(exportCmd)
}
