package cmd

import (
	"github.com/duncan-2126/ProjectManagement/internal/config"
	"github.com/spf13/cobra"
)

// Execute runs the root command
func Execute(cfg *config.Config) error {
	return rootCmd.ExecuteContext(cfg)
}

var rootCmd = &cobra.Command{
	Use:   "todo",
	Short: "TODO Tracker CLI - Track TODO comments in your codebase",
	Long: `A command-line tool to help developers and QA track TODO comments
across their codebase. Provides automated discovery, status tracking,
and management of technical debt.

Usage:
  todo scan        Scan codebase for TODO comments
  todo list        List all tracked TODOs
  todo edit        Edit TODO status, priority, or details
  todo stats       Show statistics dashboard

For more information, visit: https://github.com/duncan-2126/ProjectManagement`,
	Version: "1.0.0",
}
