package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var dueCmd = &cobra.Command{
	Use:   "due",
	Short: "Due date commands",
	Long:  "Manage due dates for TODOs",
}

var dueSetCmd = &cobra.Command{
	Use:   "set <todo-id> <date>",
	Short: "Set due date for a TODO",
	Long: `Set due date for a TODO.

Example:
  todo due set abc123 2024-12-31
  todo due set abc123 "2 weeks"`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		todoID := args[0]
		dateStr := args[1]

		// Parse date
		var dueDate time.Time
		var err error

		// Try relative dates
		switch dateStr {
		case "1d", "1 day":
			dueDate = time.Now().Add(24 * time.Hour)
		case "1w", "1 week":
			dueDate = time.Now().Add(7 * 24 * time.Hour)
		case "2w", "2 weeks":
			dueDate = time.Now().Add(14 * 24 * time.Hour)
		case "1m", "1 month":
			dueDate = time.Now().Add(30 * 24 * time.Hour)
		default:
			// Try parsing as date
			dueDate, err = time.Parse("2006-01-02", dateStr)
			if err != nil {
				return fmt.Errorf("invalid date format. Use YYYY-MM-DD or '1w', '2 weeks', etc.")
			}
		}

		// Get TODO
		todo, err := db.GetTODOByID(todoID)
		if err != nil {
			return fmt.Errorf("TODO not found: %s", todoID)
		}

		todo.DueDate = &dueDate
		if err := db.UpdateTODO(todo); err != nil {
			return err
		}

		daysLeft := int(time.Until(dueDate).Hours() / 24)
		fmt.Printf("Set due date for TODO %s to %s (%d days)\n",
			todoID[:8], dueDate.Format("2006-01-02"), daysLeft)
		return nil
	},
}

var dueListCmd = &cobra.Command{
	Use:   "list",
	Short: "List TODOs with due dates",
	Long: `List TODOs sorted by due date.

Example:
  todo due list
  todo due list --overdue`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		todos, err := db.GetTODOs(nil)
		if err != nil {
			return err
		}

		// Filter TODOs with due dates
		var withDue []database.TODO
		for _, t := range todos {
			if t.DueDate != nil {
				withDue = append(withDue, t)
			}
		}

		if len(withDue) == 0 {
			fmt.Println("No TODOs with due dates.")
			return nil
		}

		// Sort by due date
		now := time.Now()
		fmt.Println("TODOs with Due Dates:")
		fmt.Println("")

		for _, t := range withDue {
			daysLeft := int(now.Sub(*t.DueDate).Hours() / 24)
			status := ""
			if daysLeft < 0 {
				status = "OVERDUE"
			} else if daysLeft == 0 {
				status = "Due today"
			} else if daysLeft < 7 {
				status = fmt.Sprintf("%d days", daysLeft)
			}

			fmt.Printf("  [%s] %s\n", t.Priority, t.Content[:min(50, len(t.Content))])
			fmt.Printf("       Due: %s | %s\n", t.DueDate.Format("2006-01-02"), status)
		}
		return nil
	},
}

var dueClearCmd = &cobra.Command{
	Use:   "clear <todo-id>",
	Short: "Clear due date",
	Long:  `Clear the due date for a TODO.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		todoID := args[0]
		todo, err := db.GetTODOByID(todoID)
		if err != nil {
			return err
		}

		todo.DueDate = nil
		db.UpdateTODO(todo)

		fmt.Printf("Cleared due date for TODO %s\n", todoID[:8])
		return nil
	},
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func init() {
	dueCmd.AddCommand(dueSetCmd)
	dueCmd.AddCommand(dueListCmd)
	dueCmd.AddCommand(dueClearCmd)
	rootCmd.AddCommand(dueCmd)
}
