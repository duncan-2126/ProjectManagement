package cmd

import (
	"fmt"
	"os"
	"os/user"
	"time"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var digestCmd = &cobra.Command{
	Use:   "digest",
	Short: "Show daily summary",
	Long: `Show a daily summary of TODOs including assigned tasks, upcoming due dates, and stale items.

Example:
  todo digest
  todo digest --assigned-only`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		db, err := database.New(projectPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		// Get current user
		currentUser, _ := user.Current()
		userID := currentUser.Username

		// Get options
		assignedOnly, _ := cmd.Flags().GetBool("assigned-only")

		fmt.Println("=== Daily TODO Digest ===")
		fmt.Println("Date:", time.Now().Format("2006-01-02"))
		fmt.Println()

		// Get assigned TODOs
		todos, err := db.GetTODOs(map[string]interface{}{"assignee": userID})
		if err != nil {
			return fmt.Errorf("failed to get assigned TODOs: %w", err)
		}

		var openCount, inProgressCount, resolvedCount int
		for _, t := range todos {
			switch t.Status {
			case "open":
				openCount++
			case "in_progress":
				inProgressCount++
			case "resolved":
				resolvedCount++
			}
		}

		if len(todos) > 0 {
			fmt.Println("--- Assigned to You ---")
			fmt.Printf("Open: %d | In Progress: %d | Resolved: %d\n", openCount, inProgressCount, resolvedCount)
			fmt.Println()

			// Show assigned open TODOs
			fmt.Println("Your Open TODOs:")
			for _, t := range todos {
				if t.Status == "open" || t.Status == "in_progress" {
					truncated := t.Content
					if len(truncated) > 60 {
						truncated = truncated[:57] + "..."
					}
					fmt.Printf("  [%s] %s\n", t.Priority, truncated)
					if t.DueDate != nil {
						daysLeft := int(time.Until(*t.DueDate).Hours() / 24)
						if daysLeft < 0 {
							fmt.Printf("       DUE: %s (OVERDUE)\n", t.DueDate.Format("2006-01-02"))
						} else {
							fmt.Printf("       DUE: %s (in %d days)\n", t.DueDate.Format("2006-01-02"), daysLeft)
						}
					}
				}
			}
			fmt.Println()
		}

		if !assignedOnly {
			// Get upcoming due dates
			daysAhead := 7
			dueDays := viper.GetIntSlice("notifications.due_days_before")
			if len(dueDays) > 0 {
				daysAhead = dueDays[0]
			}

			upcoming, err := db.GetTODOsDueSoon(daysAhead)
			if err == nil && len(upcoming) > 0 {
				fmt.Println("--- Upcoming Due Dates ---")
				for _, t := range upcoming {
					daysLeft := int(t.DueDate.Sub(time.Now()).Hours() / 24)
					truncated := t.Content
					if len(truncated) > 50 {
						truncated = truncated[:47] + "..."
					}
					fmt.Printf("  [%s] %s\n", t.Priority, truncated)
					fmt.Printf("       Due: %s (%d days)\n", t.DueDate.Format("2006-01-02"), daysLeft)
				}
				fmt.Println()
			}

			// Get stale TODOs
			staleDays := viper.GetInt("stale.days_since_update")
			if staleDays == 0 {
				staleDays = 14
			}

			stale, err := db.GetStaleTODOs(staleDays)
			if err == nil && len(stale) > 0 {
				fmt.Println("--- Stale TODOs (no update in 14+ days) ---")
				for _, t := range stale[:5] { // Limit to 5
					daysOld := int(time.Since(t.UpdatedAt).Hours() / 24)
					truncated := t.Content
					if len(truncated) > 50 {
						truncated = truncated[:47] + "..."
					}
					fmt.Printf("  [%s] %s\n", t.Priority, truncated)
					fmt.Printf("       Last updated: %d days ago\n", daysOld)
				}
				if len(stale) > 5 {
					fmt.Printf("  ... and %d more\n", len(stale)-5)
				}
				fmt.Println()
			}
		}

		fmt.Println("=== End of Digest ===")
		return nil
	},
}

func init() {
	digestCmd.Flags().BoolP("assigned-only", "a", false, "Show only assigned TODOs")
	rootCmd.AddCommand(digestCmd)
}
