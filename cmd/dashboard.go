package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var dashboardCmd = &cobra.Command{
	Use:   "dashboard",
	Short: "Show dashboard overview",
	Long:  `Show a comprehensive dashboard with TODO statistics.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, err := database.New(projectPath)
		if err != nil {
			return err
		}

		todos, err := db.GetTODOs(nil)
		if err != nil {
			return err
		}

		stats, err := db.GetStats()
		if err != nil {
			return err
		}

		// Calculate metrics
		total := len(todos)
		open := 0
		inProgress := 0
		resolved := 0
		overdue := 0
		assigned := 0
		unassigned := 0

		now := time.Now()

		for _, t := range todos {
			switch t.Status {
			case "open":
				open++
			case "in_progress":
				inProgress++
			case "resolved", "closed":
				resolved++
			}

			if t.Assignee != "" {
				assigned++
			} else {
				unassigned++
			}

			if t.DueDate != nil && t.DueDate.Before(now) {
				overdue++
			}
		}

		// Print dashboard
		fmt.Println("")
		fmt.Println("╔════════════════════════════════════════════════════════════╗")
		fmt.Println("║              TODO TRACKER DASHBOARD                     ║")
		fmt.Println("╠════════════════════════════════════════════════════════════╣")
		fmt.Println("║                                                            ║")
		fmt.Printf("║  Total TODOs:     %-5d                                  ║\n", total)
		fmt.Printf("║  Open:            %-5d                                  ║\n", open)
		fmt.Printf("║  In Progress:     %-5d                                  ║\n", inProgress)
		fmt.Printf("║  Resolved:        %-5d                                  ║\n", resolved)
		fmt.Println("║                                                            ║")
		fmt.Printf("║  Assigned:        %-5d                                  ║\n", assigned)
		fmt.Printf("║  Unassigned:      %-5d                                  ║\n", unassigned)
		fmt.Printf("║  Overdue:         %-5d                                  ║\n", overdue)
		fmt.Println("║                                                            ║")

		// Resolution rate
		resolutionRate := 0
		if total > 0 {
			resolutionRate = (resolved * 100) / total
		}
		fmt.Printf("║  Resolution Rate:  %d%%                                        ║\n", resolutionRate)
		fmt.Println("║                                                            ║")
		fmt.Println("╚════════════════════════════════════════════════════════════╝")
		fmt.Println("")

		// Show by priority
		fmt.Println("By Priority:")
		fmt.Println("  P0 (Critical): ", countByPriority(todos, "P0"))
		fmt.Println("  P1 (High):     ", countByPriority(todos, "P1"))
		fmt.Println("  P2 (Medium):  ", countByPriority(todos, "P2"))
		fmt.Println("  P3 (Low):     ", countByPriority(todos, "P3"))
		fmt.Println("  P4 (Trivial): ", countByPriority(todos, "P4"))
		fmt.Println("")

		// Show by type
		fmt.Println("By Type:")
		fmt.Println("  TODO:  ", countByType(todos, "TODO"))
		fmt.Println("  FIXME: ", countByType(todos, "FIXME"))
		fmt.Println("  HACK:  ", countByType(todos, "HACK"))
		fmt.Println("  BUG:   ", countByType(todos, "BUG"))
		fmt.Println("  NOTE:  ", countByType(todos, "NOTE"))

		return nil
	},
}

func countByPriority(todos []database.TODO, priority string) int {
	count := 0
	for _, t := range todos {
		if t.Priority == priority {
			count++
		}
	}
	return count
}

func countByType(todos []database.TODO, todoType string) int {
	count := 0
	for _, t := range todos {
		if t.Type == todoType {
			count++
		}
	}
	return count
}

func init() {
	rootCmd.AddCommand(dashboardCmd)
}
