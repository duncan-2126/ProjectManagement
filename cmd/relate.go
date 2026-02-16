package cmd

import (
	"fmt"
	"os"

	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/spf13/cobra"
)

var relateCmd = &cobra.Command{
	Use:   "relate <id>",
	Short: "Manage TODO relationships",
	Long: `Manage relationships between TODOs (parent, depends_on, relates_to).

Examples:
  todo relate abc123 --parent xyz789
  todo relate abc123 --depends-on xyz789
  todo relate abc123 --relates-to xyz789
  todo relate abc123 --remove`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var relateParentCmd = &cobra.Command{
	Use:   "--parent <parent-id>",
	Short: "Set parent task",
	Long: `Set the parent task for a TODO.

Example:
  todo relate abc123 --parent xyz789`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, err := database.New(projectPath)
		if err != nil {
			return err
		}

		todoID := args[0]
		parentID, _ := cmd.Flags().GetString("parent")

		// Verify source TODO exists
		var todo database.TODO
		if err := db.First(&todo, "id = ?", todoID).Error; err != nil {
			return fmt.Errorf("TODO not found: %s", todoID)
		}

		// Verify parent TODO exists
		var parent database.TODO
		if err := db.First(&parent, "id = ?", parentID).Error; err != nil {
			return fmt.Errorf("parent TODO not found: %s", parentID)
		}

		// Remove existing parent relationship
		db.Where("source_id = ? AND type = ?", todoID, "parent").Delete(&database.Relationship{})
		db.Where("source_id = ? AND type = ?", todoID, "child").Delete(&database.Relationship{})

		// Create parent relationship
		if err := db.CreateRelationship(todoID, parentID, "parent"); err != nil {
			return err
		}

		// Create child relationship (inverse)
		if err := db.CreateRelationship(parentID, todoID, "child"); err != nil {
			return err
		}

		fmt.Printf("Set parent %s for TODO %s\n", parentID[:8], todoID[:8])
		return nil
	},
}

var relateDependsOnCmd = &cobra.Command{
	Use:   "--depends-on <dependency-id>",
	Short: "Add dependency",
	Long: `Add a dependency for a TODO. The TODO cannot be started until the dependency is resolved.

Example:
  todo relate abc123 --depends-on xyz789`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, err := database.New(projectPath)
		if err != nil {
			return err
		}

		todoID := args[0]
		depID, _ := cmd.Flags().GetString("depends-on")

		// Verify source TODO exists
		var todo database.TODO
		if err := db.First(&todo, "id = ?", todoID).Error; err != nil {
			return fmt.Errorf("TODO not found: %s", todoID)
		}

		// Verify dependency TODO exists
		var dep database.TODO
		if err := db.First(&dep, "id = ?", depID).Error; err != nil {
			return fmt.Errorf("dependency TODO not found: %s", depID)
		}

		// Check for circular dependency
		if db.HasCircularDependency(todoID, depID) {
			return fmt.Errorf("circular dependency detected: adding this dependency would create a cycle")
		}

		// Check if relationship already exists
		var existing database.Relationship
		if err := db.Where("source_id = ? AND target_id = ? AND type = ?", todoID, depID, "depends_on").First(&existing).Error; err == nil {
			return fmt.Errorf("dependency already exists")
		}

		// Create depends_on relationship
		if err := db.CreateRelationship(todoID, depID, "depends_on"); err != nil {
			return err
		}

		// Create blocked_by relationship (inverse)
		if err := db.CreateRelationship(depID, todoID, "blocked_by"); err != nil {
			return err
		}

		fmt.Printf("Added dependency: TODO %s depends on %s\n", todoID[:8], depID[:8])
		return nil
	},
}

var relateRelatesToCmd = &cobra.Command{
	Use:   "--relates-to <related-id>",
	Short: "Add soft association",
	Long: `Add a soft association between TODOs.

Example:
  todo relate abc123 --relates-to xyz789`,
	RunE: func(cmd *cobra.Command, args []string) error {
		projectPath, _ := os.Getwd()
		db, err := database.New(projectPath)
		if err != nil {
			return err
		}

		todoID := args[0]
		relatedID, _ := cmd.Flags().GetString("relates-to")

		// Verify source TODO exists
		var todo database.TODO
		if err := db.First(&todo, "id = ?", todoID).Error; err != nil {
			return fmt.Errorf("TODO not found: %s", todoID)
		}

		// Verify related TODO exists
		var related database.TODO
		if err := db.First(&related, "id = ?", relatedID).Error; err != nil {
			return fmt.Errorf("related TODO not found: %s", relatedID)
		}

		// Check if relationship already exists
		var existing database.Relationship
		if err := db.Where("source_id = ? AND target_id = ? AND type = ?", todoID, relatedID, "relates_to").First(&existing).Error; err == nil {
			return fmt.Errorf("relationship already exists")
		}

		// Create relates_to relationship
		if err := db.CreateRelationship(todoID, relatedID, "relates_to"); err != nil {
			return err
		}

		fmt.Printf("Added relationship: TODO %s relates to %s\n", todoID[:8], relatedID[:8])
		return nil
	},
}

var relateRemoveCmd = &cobra.Command{
	Use:   "--remove",
	Short: "Remove all relationships",
	Long: `Remove all relationships for a TODO.

Example:
  todo relate abc123 --remove`,
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

		// Get count before deletion
		var count int64
		db.Model(&database.Relationship{}).Where("source_id = ? OR target_id = ?", todoID, todoID).Count(&count)

		// Delete all relationships
		if err := db.DeleteRelationshipsForTODO(todoID); err != nil {
			return err
		}

		fmt.Printf("Removed %d relationships from TODO %s\n", count, todoID[:8])
		return nil
	},
}

func init() {
	relateCmd.AddCommand(relateParentCmd)
	relateCmd.AddCommand(relateDependsOnCmd)
	relateCmd.AddCommand(relateRelatesToCmd)
	relateCmd.AddCommand(relateRemoveCmd)

	relateParentCmd.Flags().String("parent", "", "Parent TODO ID")
	relateDependsOnCmd.Flags().String("depends-on", "", "Dependency TODO ID")
	relateRelatesToCmd.Flags().String("relates-to", "", "Related TODO ID")

	rootCmd.AddCommand(relateCmd)
}
