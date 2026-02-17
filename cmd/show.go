package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show TODO details",
	Long: `Display detailed information about a specific TODO.

Example:
  todo show abc123
  todo show abc123 --json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

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

		// Get TODO
		todo, err := db.GetTODOByID(id)
		if err != nil {
			return fmt.Errorf("TODO not found: %s", id)
		}

		// Check output format
		jsonOutput, _ := cmd.Flags().GetBool("json")

		if jsonOutput {
			jsonBytes, err := json.MarshalIndent(todo, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal JSON: %w", err)
			}
			fmt.Println(string(jsonBytes))
			return nil
		}

		// Display details
		fmt.Println("=== TODO Details ===")
		fmt.Println()
		fmt.Printf("ID:         %s\n", todo.ID)
		fmt.Printf("File:       %s\n", todo.FilePath)
		fmt.Printf("Line:       %d\n", todo.LineNumber)
		fmt.Printf("Column:     %d\n", todo.Column)
		fmt.Printf("Type:       %s\n", todo.Type)
		fmt.Printf("Status:     %s\n", todo.Status)
		fmt.Printf("Priority:   %s\n", todo.Priority)
		fmt.Printf("Category:   %s\n", todo.Category)
		fmt.Printf("Assignee:   %s\n", todo.Assignee)
		fmt.Printf("Author:     %s\n", todo.Author)
		fmt.Printf("Created:    %s\n", todo.CreatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("Updated:    %s\n", todo.UpdatedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("\nContent:\n%s\n", todo.Content)

		return nil
	},
}

func init() {
	showCmd.Flags().BoolP("json", "j", false, "Output as JSON")

	rootCmd.AddCommand(showCmd)
}
