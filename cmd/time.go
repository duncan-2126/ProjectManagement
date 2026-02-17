package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var timeCmd = &cobra.Command{
	Use:   "time",
	Short: "Time tracking commands",
	Long:  "Track time spent on TODOs",
}

var timeStartCmd = &cobra.Command{
	Use:   "start <todo-id> [description]",
	Short: "Start timer for a TODO",
	Long: `Start a timer for a TODO.

Example:
  todo time start abc123
  todo time start abc123 "Working on authentication"`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		todoID := args[0]
		desc := ""
		if len(args) > 1 {
			desc = args[1]
		}

		entry, err := db.StartTimer(todoID, desc)
		if err != nil {
			return err
		}

		fmt.Printf("Started timer for TODO %s\n", todoID[:8])
		fmt.Printf("Started at: %s\n", entry.StartTime.Format("15:04:05"))
		return nil
	},
}

var timeStopCmd = &cobra.Command{
	Use:   "stop <todo-id>",
	Short: "Stop timer for a TODO",
	Long: `Stop a running timer.

Example:
  todo time stop abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		todoID := args[0]

		entry, err := db.StopTimer(todoID)
		if err != nil {
			return err
		}

		fmt.Printf("Stopped timer for TODO %s\n", todoID[:8])
		fmt.Printf("Duration: %d minutes\n", entry.Duration)
		return nil
	},
}

var timeLogCmd = &cobra.Command{
	Use:   "log <todo-id> <minutes> [description]",
	Short: "Log time manually",
	Long: `Manually log time spent on a TODO.

Example:
  todo time log abc123 30 "Fixed bug"
  todo time log abc123 60`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		todoID := args[0]
		var minutes int
		fmt.Sscanf(args[1], "%d", &minutes)
		desc := ""
		if len(args) > 2 {
			desc = args[2]
		}

		_, err := db.AddManualTime(todoID, minutes, desc)
		if err != nil {
			return err
		}

		fmt.Printf("Logged %d minutes for TODO %s\n", minutes, todoID[:8])
		return nil
	},
}

var timeShowCmd = &cobra.Command{
	Use:   "show <todo-id>",
	Short: "Show time entries for a TODO",
	Long: `Show all time entries for a TODO.

Example:
  todo time show abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		todoID := args[0]

		entries, err := db.GetTimeEntries(todoID)
		if err != nil {
			return err
		}

		total, _ := db.GetTotalTime(todoID)

		fmt.Printf("Time entries for TODO %s (Total: %d min)\n\n", todoID[:8], total)
		for _, e := range entries {
			duration := e.Duration
			if e.EndTime == nil {
				duration = int(time.Since(e.StartTime).Minutes())
				fmt.Printf("  * Running: started %s (%d min so far)\n",
					e.StartTime.Format("15:04"), duration)
			} else {
				fmt.Printf("  - %s to %s (%d min) %s\n",
					e.StartTime.Format("15:04"), e.EndTime.Format("15:04"),
					duration, e.Description)
			}
		}
		return nil
	},
}

func init() {
	timeCmd.AddCommand(timeStartCmd)
	timeCmd.AddCommand(timeStopCmd)
	timeCmd.AddCommand(timeLogCmd)
	timeCmd.AddCommand(timeShowCmd)
	rootCmd.AddCommand(timeCmd)
}
