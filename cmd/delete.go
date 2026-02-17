package cmd

import (
	"fmt"
	"os"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a TODO",
	Long: `Delete a TODO from the database. This action cannot be undone.

Examples:
  todo delete abc123
  todo delete abc123 --force  # Skip confirmation`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]
		force, _ := cmd.Flags().GetBool("force")

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

		// Get TODO first to show details
		todo, err := db.GetTODOByID(id)
		if err != nil {
			return fmt.Errorf("TODO not found: %s", id)
		}

		// Confirm deletion unless force flag is set
		if !force {
			fmt.Printf("Delete TODO %s?\n", id[:8])
			fmt.Printf("  File: %s\n", todo.FilePath)
			fmt.Printf("  Line: %d\n", todo.LineNumber)
			fmt.Printf("  Type: %s\n", todo.Type)
			fmt.Printf("  Content: %s\n", todo.Content)
			fmt.Print("\nType 'yes' to confirm: ")

			var confirm string
			fmt.Scanln(&confirm)
			if confirm != "yes" {
				fmt.Println("Deletion cancelled.")
				return nil
			}
		}

		// Delete TODO
		if err := db.DeleteTODO(id); err != nil {
			return fmt.Errorf("failed to delete TODO: %w", err)
		}

		fmt.Printf("TODO %s deleted successfully.\n", id[:8])

		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")

	rootCmd.AddCommand(deleteCmd)
}
