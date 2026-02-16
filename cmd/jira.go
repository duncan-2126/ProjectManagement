package cmd

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

// jiraCmd represents the Jira integration command
var jiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Export TODOs to Jira format",
	Long: `Export TODOs to Jira-compatible CSV format for import into Jira.

Examples:
  todo jira                    # Export all TODOs to CSV
  todo jira --status open     # Export only open TODOs
  todo jira -o issues.csv     # Export to file`,
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

		// Get output file
		outputFile, _ := cmd.Flags().GetString("output")

		return exportTODOsToJira(todos, outputFile)
	},
}

func exportTODOsToJira(todos []database.TODO, outputFile string) error {
	var writer *csv.Writer
	if outputFile != "" {
		file, err := os.Create(outputFile)
		if err != nil {
			return fmt.Errorf("failed to create output file: %w", err)
		}
		defer file.Close()
		writer = csv.NewWriter(file)
	} else {
		writer = csv.NewWriter(os.Stdout)
	}
	defer writer.Flush()

	// Write header (Jira import CSV format)
	headers := []string{"Summary", "Description", "Priority", "Assignee", "Due Date", "Status", "Issue Type"}
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

		// Map status to Jira workflow
		status := mapStatusToJira(t.Status)

		// Map type to Jira issue type
		issueType := mapTypeToJira(t.Type)

		// Truncate summary (Jira limit is 255)
		summary := t.Content
		if len(summary) > 255 {
			summary = summary[:252] + "..."
		}

		// Build description with metadata
		description := fmt.Sprintf("h3. Source Information\n\n*File:* %s\n*Line:* %d\n*Type:* %s\n*Author:* %s\n*Created:* %s\n\n---\n\nh3. Description\n\n%s",
			t.FilePath, t.LineNumber, t.Type, t.Author, t.CreatedAt.Format("2006-01-02"), t.Content)

		row := []string{
			summary,
			description,
			priority,
			t.Assignee,
			dueDate,
			status,
			issueType,
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	if outputFile != "" {
		fmt.Printf("Exported %d TODOs to %s\n", len(todos), outputFile)
	}

	return nil
}

func mapStatusToJira(status string) string {
	switch status {
	case "open":
		return "Open"
	case "in_progress":
		return "In Progress"
	case "blocked":
		return "Blocked"
	case "resolved":
		return "Resolved"
	case "closed":
		return "Closed"
	case "wontfix":
		return "Won't Fix"
	default:
		return "Open"
	}
}

func mapTypeToJira(todoType string) string {
	switch todoType {
	case "BUG":
		return "Bug"
	case "FIXME":
		return "Bug"
	case "TODO":
		return "Task"
	case "HACK":
		return "Technical Task"
	case "NOTE":
		return "Task"
	case "XXX":
		return "Task"
	default:
		return "Task"
	}
}

func init() {
	jiraCmd.Flags().StringP("status", "s", "", "Filter by status")
	jiraCmd.Flags().StringP("output", "o", "", "Output file (default: stdout)")
	rootCmd.AddCommand(jiraCmd)
}
