package cmd

import (
	"fmt"
	"os"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var assignCmd = &cobra.Command{
	Use:   "assign",
	Short: "Assign TODOs to team members",
	Long:  "Manage TODO assignments",
}

var assignAddCmd = &cobra.Command{
	Use:   "add <todo-id> <username>",
	Short: "Assign a TODO to a user",
	Long: `Assign a TODO to a team member.

Example:
  todo assign add abc123 john`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		todoID := args[0]
		username := args[1]

		todo, err := db.GetTODOByID(todoID)
		if err != nil {
			return fmt.Errorf("TODO not found: %s", todoID)
		}

		todo.Assignee = username
		db.UpdateTODO(todo)

		fmt.Printf("Assigned TODO %s to %s\n", todoID[:8], username)
		return nil
	},
}

var assignRemoveCmd = &cobra.Command{
	Use:   "remove <todo-id>",
	Short: "Remove assignment from TODO",
	Long:  `Remove the assignee from a TODO.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		todoID := args[0]

		todo, err := db.GetTODOByID(todoID)
		if err != nil {
			return fmt.Errorf("TODO not found: %s", todoID)
		}

		oldAssignee := todo.Assignee
		todo.Assignee = ""
		db.UpdateTODO(todo)

		fmt.Printf("Unassigned TODO %s (was: %s)\n", todoID[:8], oldAssignee)
		return nil
	},
}

var teamCmd = &cobra.Command{
	Use:   "team",
	Short: "Team management",
	Long:  "Team and assignment commands",
}

var teamStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show team workload",
	Long: `Show TODO statistics by team member.

Example:
  todo team stats`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, _ := database.New(projectPath)

		todos, _ := db.GetTODOs(nil)

		// Group by assignee
		workload := make(map[string]int)
		var unassigned int
		for _, t := range todos {
			if t.Assignee != "" {
				workload[t.Assignee]++
			} else {
				unassigned++
			}
		}

		fmt.Println("Team Workload:")
		fmt.Println("")
		fmt.Printf("  Unassigned: %d\n", unassigned)
		for user, count := range workload {
			fmt.Printf("  %s: %d TODOs\n", user, count)
		}
		return nil
	},
}

func init() {
	assignCmd.AddCommand(assignAddCmd)
	assignCmd.AddCommand(assignRemoveCmd)
	rootCmd.AddCommand(assignCmd)
	rootCmd.AddCommand(teamCmd)
	teamCmd.AddCommand(teamStatsCmd)
}
