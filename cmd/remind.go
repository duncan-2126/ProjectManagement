package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/duncan-2126/ProjectManagement/internal/config"
	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var remindCmd = &cobra.Command{
	Use:   "remind",
	Short: "Show TODOs with approaching due dates",
	Long: `Show TODOs that have due dates approaching within the configured number of days.

Configure reminder days with:
  todo config set notifications.due-days-before 3,1

Examples:
  todo remind              # Show TODOs due within configured days
  todo remind --days 7     # Show TODOs due within 7 days`,
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

		// Get days from flags or config
		days, _ := cmd.Flags().GetInt("days")
		if days == 0 {
			days = viper.GetInt("notifications.due_days_before.0")
			if days == 0 {
				days = 7 // Default to 7 days
			}
		}

		// Calculate the due date threshold
		threshold := time.Now().AddDate(0, 0, days)

		// Get all open TODOs with due dates
		filters := map[string]interface{}{
			"status": "open",
		}
		todos, err := db.GetTODOs(filters)
		if err != nil {
			return fmt.Errorf("failed to get TODOs: %w", err)
		}

		// Filter TODOs with due dates within threshold
		var upcoming []database.TODO
		for _, t := range todos {
			if t.DueDate != nil && !t.DueDate.After(threshold) {
				upcoming = append(upcoming, t)
			}
		}

		// Sort by due date
		sortTODOsByDueDate(upcoming)

		// Display results
		if len(upcoming) == 0 {
			fmt.Println("No upcoming TODOs due within", days, "days.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tPriority\tStatus\tDue Date\tDays Left\tContent")
		fmt.Fprintln(w, "--\t--------\t------\t--------\t---------\t-------")

		now := time.Now()
		for _, t := range upcoming {
			daysLeft := int(t.DueDate.Sub(now).Hours() / 24)
			dueDateStr := t.DueDate.Format("2006-01-02")
			if daysLeft < 0 {
				dueDateStr = dueDateStr + " (OVERDUE)"
			}
			content := t.Content
			if len(content) > 40 {
				content = content[:37] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%s\n",
				t.ID[:8], t.Priority, t.Status, dueDateStr, daysLeft, content)
		}
		w.Flush()

		fmt.Printf("\nTotal: %d TODOs due within %d days\n", len(upcoming), days)

		return nil
	},
}

func sortTODOsByDueDate(todos []database.TODO) {
	for i := 0; i < len(todos)-1; i++ {
		for j := i + 1; j < len(todos); j++ {
			if todos[j].DueDate.Before(*todos[i].DueDate) {
				todos[i], todos[j] = todos[j], todos[i]
			}
		}
	}
}

func init() {
	remindCmd.Flags().IntP("days", "d", 0, "Show TODOs due within this many days (overrides config)")

	rootCmd.AddCommand(remindCmd)
}
