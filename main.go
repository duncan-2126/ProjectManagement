package main

import (
	"os"

	"github.com/duncan-2126/ProjectManagement/cmd"
	"github.com/duncan-2126/ProjectManagement/internal/config"
)

func main() {
	// Initialize configuration
	cfg := config.Load()

	// Execute root command
	if err := cmd.Execute(cfg); err != nil {
		os.Exit(1)
	}
}
