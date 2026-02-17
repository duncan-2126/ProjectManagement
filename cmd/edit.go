package cmd

import (
	"fmt"
	"os"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a TODO",
	Long: `Edit a TODO's status, priority, or other details.

Examples:
  todo edit abc123 --status resolved
  todo edit abc123 --priority P1
  todo edit abc123 --status in_progress --priority P0`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

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

		// Get TODO
		todo, err := db.GetTODOByID(id)
		if err != nil {
			return fmt.Errorf("TODO not found: %s", id)
		}

		// Get flags
		status, _ := cmd.Flags().GetString("status")
		priority, _ := cmd.Flags().GetString("priority")
		category, _ := cmd.Flags().GetString("category")
		assignee, _ := cmd.Flags().GetString("assignee")
		content, _ := cmd.Flags().GetString("content")

		// Update TODO
		updated := false
		if status != "" {
			if !isValidStatus(status) {
				return fmt.Errorf("invalid status: %s (valid: open, in_progress, blocked, resolved, wontfix, closed)", status)
			}
			todo.Status = status
			updated = true
		}

		if priority != "" {
			if !isValidPriority(priority) {
				return fmt.Errorf("invalid priority: %s (valid: P0, P1, P2, P3, P4)", priority)
			}
			todo.Priority = priority
			updated = true
		}

		if category != "" {
			todo.Category = category
			updated = true
		}

		if assignee != "" {
			todo.Assignee = assignee
			updated = true
		}

		if content != "" {
			todo.Content = content
			updated = true
		}

		if !updated {
			return fmt.Errorf("no changes specified. Use --status, --priority, --category, --assignee, or --content")
		}

		// Save
		if err := db.UpdateTODO(todo); err != nil {
			return fmt.Errorf("failed to update TODO: %w", err)
		}

		fmt.Printf("TODO %s updated successfully\n", id[:8])
		fmt.Printf("  Status: %s -> %s\n", status, todo.Status)
		fmt.Printf("  Priority: %s -> %s\n", priority, todo.Priority)

		return nil
	},
}

func isValidStatus(s string) bool {
	switch s {
	case "open", "in_progress", "blocked", "resolved", "wontfix", "closed":
		return true
	}
	return false
}

func isValidPriority(p string) bool {
	switch p {
	case "P0", "P1", "P2", "P3", "P4":
		return true
	}
	return false
}

func init() {
	editCmd.Flags().StringP("status", "s", "", "Set status (open, in_progress, blocked, resolved, wontfix, closed)")
	editCmd.Flags().StringP("priority", "p", "", "Set priority (P0, P1, P2, P3, P4)")
	editCmd.Flags().StringP("category", "c", "", "Set category")
	editCmd.Flags().StringP("assignee", "a", "", "Set assignee")
	editCmd.Flags().StringP("content", "m", "", "Set content/description")

	rootCmd.AddCommand(editCmd)
}
