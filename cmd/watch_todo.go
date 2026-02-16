package cmd

import (
	"fmt"
	"os"
	"os/user"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var watchTodoCmd = &cobra.Command{
	Use:   "watch <todo-id>",
	Short: "Watch a TODO for notifications",
	Long: `Watch a TODO to receive notifications about updates.

Example:
  todo watch abc123
  todo watch abc123 --user john`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		db, err := database.New(projectPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		todoID := args[0]

		// Get TODO
		todo, err := db.GetTODOByID(todoID)
		if err != nil {
			return fmt.Errorf("TODO not found: %s", todoID)
		}

		// Get user
		userFlag, _ := cmd.Flags().GetString("user")
		var userID string
		if userFlag != "" {
			userID = userFlag
		} else {
			currentUser, err := user.Current()
			if err != nil {
				userID = "default"
			} else {
				userID = currentUser.Username
			}
		}

		// Check if already watching
		if db.IsWatching(todoID, userID) {
			fmt.Printf("You are already watching TODO %s\n", todoID[:8])
			return nil
		}

		// Create watch
		if err := db.CreateWatch(todoID, userID); err != nil {
			return fmt.Errorf("failed to watch TODO: %w", err)
		}

		fmt.Printf("Now watching TODO %s\n", todoID[:8])
		fmt.Printf("  Content: %s\n", todo.Content[:min(50, len(todo.Content))])
		return nil
	},
}

var unwatchTodoCmd = &cobra.Command{
	Use:   "unwatch <todo-id>",
	Short: "Stop watching a TODO",
	Long: `Stop watching a TODO.

Example:
  todo unwatch abc123
  todo unwatch abc123 --user john`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		db, err := database.New(projectPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		todoID := args[0]

		// Get user
		userFlag, _ := cmd.Flags().GetString("user")
		var userID string
		if userFlag != "" {
			userID = userFlag
		} else {
			currentUser, err := user.Current()
			if err != nil {
				userID = "default"
			} else {
				userID = currentUser.Username
			}
		}

		// Check if watching
		if !db.IsWatching(todoID, userID) {
			fmt.Printf("You are not watching TODO %s\n", todoID[:8])
			return nil
		}

		// Delete watch
		if err := db.DeleteWatch(todoID, userID); err != nil {
			return fmt.Errorf("failed to unwatch TODO: %w", err)
		}

		fmt.Printf("Stopped watching TODO %s\n", todoID[:8])
		return nil
	},
}

var watchingListCmd = &cobra.Command{
	Use:   "watching",
	Short: "List watched TODOs",
	Long: `List all TODOs you are watching.

Example:
  todo watching
  todo watching --user john`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		db, err := database.New(projectPath)
		if err != nil {
			return fmt.Errorf("failed to open database: %w", err)
		}

		// Get user
		userFlag, _ := cmd.Flags().GetString("user")
		var userID string
		if userFlag != "" {
			userID = userFlag
		} else {
			currentUser, err := user.Current()
			if err != nil {
				userID = "default"
			} else {
				userID = currentUser.Username
			}
		}

		// Get watched TODOs
		todos, err := db.GetWatchedTODOs(userID)
		if err != nil {
			return fmt.Errorf("failed to get watched TODOs: %w", err)
		}

		if len(todos) == 0 {
			fmt.Println("You are not watching any TODOs.")
			return nil
		}

		fmt.Println("=== Watched TODOs ===")
		fmt.Println()
		for _, t := range todos {
			fmt.Printf("[%s] %s\n", t.Priority, t.Content)
			fmt.Printf("     Status: %s | Updated: %s\n", t.Status, t.UpdatedAt.Format("2006-01-02"))
		}
		fmt.Printf("\nTotal: %d watched TODOs\n", len(todos))
		return nil
	},
}

func init() {
	watchTodoCmd.Flags().StringP("user", "u", "", "User ID (default: current user)")
	rootCmd.AddCommand(watchTodoCmd)

	unwatchTodoCmd.Flags().StringP("user", "u", "", "User ID (default: current user)")
	rootCmd.AddCommand(unwatchTodoCmd)

	watchingListCmd.Flags().StringP("user", "u", "", "User ID (default: current user)")
	rootCmd.AddCommand(watchingListCmd)
}
