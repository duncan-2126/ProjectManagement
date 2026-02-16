package cmd

import (
	"fmt"
	"os"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/duncan-2126/ProjectManagement/internal/git"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync TODO information with git",
	Long: `Synchronize TODO information with git metadata including author information and commit history.

Example:
  todo sync
  todo sync --blame`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if this is a git repo
		if !git.IsRepo() {
			fmt.Println("Not a git repository. Skipping sync.")
			return nil
		}

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

		// Get all TODOs
		todos, err := db.GetTODOs(nil)
		if err != nil {
			return fmt.Errorf("failed to get TODOs: %w", err)
		}

		// Get blame information
		blameEnabled, _ := cmd.Flags().GetBool("blame")
		if !blameEnabled {
			fmt.Println("Git sync enabled (use --blame for author info)")
		}

		fmt.Printf("Syncing %d TODOs...\n", len(todos))

		// Track files that need blame
		fileBlame := make(map[string]map[int]git.Author)

		for _, todo := range todos {
			if blameEnabled {
				// Get blame for file if not already cached
				if _, ok := fileBlame[todo.FilePath]; !ok {
					blame, err := git.GetBlame(todo.FilePath)
					if err == nil {
						fileBlame[todo.FilePath] = blame
					}
				}

				// Update author info if available
				if blame, ok := fileBlame[todo.FilePath]; ok {
					if author, ok := blame[todo.LineNumber]; ok {
						todo.Author = author.Name
						todo.Email = author.Email
						todo.CreatedAt = author.Date
						db.UpdateTODO(&todo)
					}
				}
			}
		}

		// Show current branch
		branch, err := git.GetCurrentBranch()
		if err == nil {
			fmt.Printf("Current branch: %s\n", branch)
		}

		fmt.Println("Sync complete!")
		return nil
	},
}

func init() {
	syncCmd.Flags().BoolP("blame", "b", false, "Run git blame to get author info")
	rootCmd.AddCommand(syncCmd)
}
