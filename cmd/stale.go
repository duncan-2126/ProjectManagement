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

var staleCmd = &cobra.Command{
	Use:   "stale",
	Short: "List stale TODOs",
	Long: `List TODOs that haven't been updated in a while or have been open for a long time.

Configure stale detection with:
  todo config set stale.enabled true
  todo config set stale.days-open 30
  todo config set stale.days-since-update 14

Examples:
  todo stale              # List stale TODOs based on config
  todo stale --days 30    # List TODOs not updated in 30 days`,
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
			days = viper.GetInt("stale.days_since_update")
			if days == 0 {
				days = 30 // Default to 30 days
			}
		}

		// Get all open TODOs
		filters := map[string]interface{}{
			"status": "open",
		}
		todos, err := db.GetTODOs(filters)
		if err != nil {
			return fmt.Errorf("failed to get TODOs: %w", err)
		}

		// Filter stale TODOs
		var stale []database.TODO
		now := time.Now()
		for _, t := range todos {
			daysSinceUpdate := int(now.Sub(t.UpdatedAt).Hours() / 24)
			daysOpen := int(now.Sub(t.CreatedAt).Hours() / 24)

			// Check configured thresholds
			configDaysSinceUpdate := viper.GetInt("stale.days_since_update")
			configDaysOpen := viper.GetInt("stale.days_open")

			if configDaysSinceUpdate > 0 && daysSinceUpdate >= configDaysSinceUpdate {
				stale = append(stale, t)
			} else if configDaysOpen > 0 && daysOpen >= configDaysOpen {
				stale = append(stale, t)
			} else if days > 0 && daysSinceUpdate >= days {
				// Fallback to command-line flag
				stale = append(stale, t)
			}
		}

		// Sort by days since update
		sortTODOsByStaleness(stale, now)

		// Display results
		if len(stale) == 0 {
			fmt.Println("No stale TODOs found.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tPriority\tStatus\tCreated\tUpdated\tDays Old\tContent")
		fmt.Fprintln(w, "--\t--------\t------\t-------\t-------\t--------\t-------")

		for _, t := range stale {
			daysOld := int(now.Sub(t.CreatedAt).Hours() / 24)
			createdStr := t.CreatedAt.Format("2006-01-02")
			updatedStr := t.UpdatedAt.Format("2006-01-02")
			content := t.Content
			if len(content) > 35 {
				content = content[:32] + "..."
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%s\n",
				t.ID[:8], t.Priority, t.Status, createdStr, updatedStr, daysOld, content)
		}
		w.Flush()

		fmt.Printf("\nTotal: %d stale TODOs\n", len(stale))

		return nil
	},
}

func sortTODOsByStaleness(todos []database.TODO, now time.Time) {
	for i := 0; i < len(todos)-1; i++ {
		for j := i + 1; j < len(todos); j++ {
			daysI := now.Sub(todos[i].UpdatedAt).Hours() / 24
			daysJ := now.Sub(todos[j].UpdatedAt).Hours() / 24
			if daysJ < daysI {
				todos[i], todos[j] = todos[j], todos[i]
			}
		}
	}
}

func init() {
	staleCmd.Flags().IntP("days", "d", 0, "Consider TODOs stale after this many days (overrides config)")

	rootCmd.AddCommand(staleCmd)
}
