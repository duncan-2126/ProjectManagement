package cmd

import (
	"fmt"
	"os"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var depsCmd = &cobra.Command{
	Use:   "deps <id>",
	Short: "Show what this task depends on",
	Long: `Show TODOs that this task depends on.

Example:
  todo deps abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, err := database.New(projectPath)
		if err != nil {
			return err
		}

		todoID := args[0]

		// Verify TODO exists
		var todo database.TODO
		if err := db.First(&todo, "id = ?", todoID).Error; err != nil {
			return fmt.Errorf("TODO not found: %s", todoID)
		}

		// Get dependencies
		deps, err := db.GetDependents(todoID)
		if err != nil {
			return err
		}

		if len(deps) == 0 {
			fmt.Printf("TODO %s has no dependencies\n", todoID[:8])
			return nil
		}

		fmt.Printf("TODO %s depends on:\n", todoID[:8])
		for _, dep := range deps {
			statusIcon := getStatusIcon(dep.Status)
			fmt.Printf("  %s [%s] %s - %s\n", statusIcon, dep.Priority, dep.ID[:8], dep.Content)
		}
		return nil
	},
}

var blockersCmd = &cobra.Command{
	Use:   "blockers <id>",
	Short: "Show what blocks this task",
	Long: `Show TODOs that block this task.

Example:
  todo blockers abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, err := database.New(projectPath)
		if err != nil {
			return err
		}

		todoID := args[0]

		// Verify TODO exists
		var todo database.TODO
		if err := db.First(&todo, "id = ?", todoID).Error; err != nil {
			return fmt.Errorf("TODO not found: %s", todoID)
		}

		// Get blockers
		blockers, err := db.GetBlockers(todoID)
		if err != nil {
			return err
		}

		if len(blockers) == 0 {
			fmt.Printf("TODO %s is not blocked\n", todoID[:8])
			return nil
		}

		fmt.Printf("TODO %s is blocked by:\n", todoID[:8])
		for _, blocker := range blockers {
			statusIcon := getStatusIcon(blocker.Status)
			fmt.Printf("  %s [%s] %s - %s\n", statusIcon, blocker.Priority, blocker.ID[:8], blocker.Content)
		}
		return nil
	},
}

var childrenCmd = &cobra.Command{
	Use:   "children <id>",
	Short: "Show subtasks",
	Long: `Show subtasks of a TODO.

Example:
  todo children abc123`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, err := database.New(projectPath)
		if err != nil {
			return err
		}

		todoID := args[0]

		// Verify TODO exists
		var todo database.TODO
		if err := db.First(&todo, "id = ?", todoID).Error; err != nil {
			return fmt.Errorf("TODO not found: %s", todoID)
		}

		// Get children
		children, err := db.GetChildren(todoID)
		if err != nil {
			return err
		}

		if len(children) == 0 {
			fmt.Printf("TODO %s has no subtasks\n", todoID[:8])
			return nil
		}

		fmt.Printf("Subtasks of TODO %s:\n", todoID[:8])
		for _, child := range children {
			statusIcon := getStatusIcon(child.Status)
			fmt.Printf("  %s [%s] %s - %s\n", statusIcon, child.Priority, child.ID[:8], child.Content)
		}
		return nil
	},
}

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check for circular dependencies and broken links",
	Long: `Validate all relationships for issues like circular dependencies and broken links.

Example:
  todo validate`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, err := database.New(projectPath)
		if err != nil {
			return err
		}

		issues, err := db.ValidateRelationships()
		if err != nil {
			return err
		}

		if len(issues) == 0 {
			fmt.Println("All relationships are valid!")
			return nil
		}

		fmt.Println("Relationship validation issues found:")

		if brokenLinks, ok := issues["broken_links"]; ok && len(brokenLinks) > 0 {
			fmt.Println("\nBroken Links:")
			for _, issue := range brokenLinks {
				fmt.Printf("  - %s\n", issue)
			}
		}

		if circularDeps, ok := issues["circular_dependencies"]; ok && len(circularDeps) > 0 {
			fmt.Println("\nCircular Dependencies:")
			for _, issue := range circularDeps {
				fmt.Printf("  - %s\n", issue)
			}
		}

		return nil
	},
}

func getStatusIcon(status string) string {
	switch status {
	case "open":
		return "○"
	case "in_progress":
		return "◐"
	case "blocked":
		return "⊘"
	case "resolved":
		return "◉"
	case "wontfix":
		return "✗"
	case "closed":
		return "●"
	default:
		return "○"
	}
}

func init() {
	rootCmd.AddCommand(depsCmd)
	rootCmd.AddCommand(blockersCmd)
	rootCmd.AddCommand(childrenCmd)
	rootCmd.AddCommand(validateCmd)
}
