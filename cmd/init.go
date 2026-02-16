package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [project-name]",
	Short: "Initialize a new TODO Tracker project",
	Long: `Initialize a new TODO Tracker project in the current directory.
Creates a .todo directory with configuration and database.

Example:
  todo init
  todo init my-project`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Get project path
		projectPath, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}

		// Get project name
		projectName := filepath.Base(projectPath)
		if len(args) > 0 {
			projectName = args[0]
		}

		// Check if already initialized
		todoDir := filepath.Join(projectPath, ".todo")
		if _, err := os.Stat(todoDir); err == nil {
			fmt.Printf("Project already initialized at %s\n", projectPath)
			fmt.Println("Run 'todo scan' to scan for TODOs")
			return nil
		}

		// Create database
		db, err := database.New(projectPath)
		if err != nil {
			return fmt.Errorf("failed to initialize database: %w", err)
		}

		// Initialize project in database
		if err := db.InitProject(projectName, projectPath); err != nil {
			return fmt.Errorf("failed to init project: %w", err)
		}

		// Create config file
		configPath := filepath.Join(todoDir, "config.toml")
		configContent := fmt.Sprintf(`# TODO Tracker Configuration
# Project: %s
# Created: %s

[todo_types]
  default = ["TODO", "FIXME", "HACK", "BUG", "NOTE", "XXX"]

[exclude]
  default = [".git", "node_modules", "vendor", "dist", "build"]

[git]
  author = true

[display]
  color = "auto"
`, projectName, time.Now().Format("2006-01-02"))

		if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
			return fmt.Errorf("failed to create config: %w", err)
		}

		// Add to .gitignore if in git repo
		gitignorePath := filepath.Join(projectPath, ".gitignore")
		if _, err := os.Stat(gitignorePath); err == nil {
			// Check if .todo is already in .gitignore
			content, _ := os.ReadFile(gitignorePath)
			if !contains(string(content), ".todo") {
				f, _ := os.OpenFile(gitignorePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
				defer f.Close()
				f.WriteString("\n# TODO Tracker\n.todo/\n")
			}
		}

		fmt.Printf("✓ Initialized TODO Tracker project: %s\n", projectName)
		fmt.Printf("  Database: %s\n", filepath.Join(todoDir, "todos.db"))
		fmt.Printf("  Config: %s\n", configPath)
		fmt.Println("\nNext steps:")
		fmt.Println("  todo scan        # Scan for TODOs")
		fmt.Println("  todo list        # View TODOs")

		return nil
	},
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func init() {
	rootCmd.AddCommand(initCmd)
}
