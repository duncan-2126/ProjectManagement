package cmd

import (
	"fmt"
	"os"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags",
	Long:  "Manage tags for TODOs",
}

var tagCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new tag",
	Long: `Create a new tag that can be applied to TODOs.

Example:
  todo tag create urgent
  todo tag create "high-priority"`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, err := database.New(projectPath)
		if err != nil {
			return err
		}

		tag := database.Tag{
			ID:   uuid.New().String(),
			Name: args[0],
		}

		if err := db.Create(&tag).Error; err != nil {
			return fmt.Errorf("tag already exists: %s", args[0])
		}

		fmt.Printf("Created tag: %s\n", tag.Name)
		return nil
	},
}

var tagListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags",
	Long: `List all tags.

Example:
  todo tag list`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, err := database.New(projectPath)
		if err != nil {
			return err
		}

		var tags []database.Tag
		if err := db.Find(&tags).Error; err != nil {
			return err
		}

		if len(tags) == 0 {
			fmt.Println("No tags found.")
			return nil
		}

		fmt.Println("Tags:")
		for _, tag := range tags {
			var count int64
			db.Model(&database.TODOTag{}).Where("tag_id = ?", tag.ID).Count(&count)
			fmt.Printf("  %s (%d TODOs)\n", tag.Name, count)
		}
		return nil
	},
}

var tagDeleteCmd = &cobra.Command{
	Use:   "delete <name>",
	Short: "Delete a tag",
	Long: `Delete a tag. This removes the tag from all TODOs.

Example:
  todo tag delete urgent`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, err := database.New(projectPath)
		if err != nil {
			return err
		}

		var tag database.Tag
		if err := db.First(&tag, "name = ?", args[0]).Error; err != nil {
			return fmt.Errorf("tag not found: %s", args[0])
		}

		db.Where("tag_id = ?", tag.ID).Delete(&database.TODOTag{})
		db.Delete(&tag)

		fmt.Printf("Deleted tag: %s\n", args[0])
		return nil
	},
}

var tagAddCmd = &cobra.Command{
	Use:   "add <todo-id> <tag-name>",
	Short: "Add a tag to a TODO",
	Long: `Add a tag to a TODO.

Example:
  todo tag add abc123 urgent`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, err := database.New(projectPath)
		if err != nil {
			return err
		}

		todoID := args[0]
		tagName := args[1]

		var todo database.TODO
		if err := db.First(&todo, "id = ?", todoID).Error; err != nil {
			return fmt.Errorf("TODO not found: %s", todoID)
		}

		var tag database.Tag
		if err := db.First(&tag, "name = ?", tagName).Error; err != nil {
			tag = database.Tag{ID: uuid.New().String(), Name: tagName}
			db.Create(&tag)
		}

		var existing database.TODOTag
		if err := db.First(&existing, "todo_id = ? AND tag_id = ?", todo.ID, tag.ID).Error; err == nil {
			return fmt.Errorf("TODO already has tag: %s", tagName)
		}

		assoc := database.TODOTag{TODOID: todo.ID, TagID: tag.ID}
		db.Create(&assoc)

		fmt.Printf("Added tag '%s' to TODO %s\n", tagName, todoID[:8])
		return nil
	},
}

var tagRemoveCmd = &cobra.Command{
	Use:   "remove <todo-id> <tag-name>",
	Short: "Remove a tag from a TODO",
	Long: `Remove a tag from a TODO.

Example:
  todo tag remove abc123 urgent`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, err := database.New(projectPath)
		if err != nil {
			return err
		}

		todoID := args[0]
		tagName := args[1]

		var todo database.TODO
		if err := db.First(&todo, "id = ?", todoID).Error; err != nil {
			return fmt.Errorf("TODO not found: %s", todoID)
		}

		var tag database.Tag
		if err := db.First(&tag, "name = ?", tagName).Error; err != nil {
			return fmt.Errorf("tag not found: %s", tagName)
		}

		db.Where("todo_id = ? AND tag_id = ?", todo.ID, tag.ID).Delete(&database.TODOTag{})

		fmt.Printf("Removed tag '%s' from TODO %s\n", tagName, todoID[:8])
		return nil
	},
}

func init() {
	tagCmd.AddCommand(tagCreateCmd)
	tagCmd.AddCommand(tagListCmd)
	tagCmd.AddCommand(tagDeleteCmd)
	tagCmd.AddCommand(tagAddCmd)
	tagCmd.AddCommand(tagRemoveCmd)
	rootCmd.AddCommand(tagCmd)
}
