package config

import (
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config holds all configuration settings
type Config struct {
	// Scanning
	ExcludePatterns []string `mapstructure:"exclude"`
	IncludePatterns []string `mapstructure:"include"`
	TodoTypes       []string `mapstructure:"todo_types"`
	IgnoreCase      bool     `mapstructure:"ignore_case"`

	// Git Integration
	GitAuthor       bool   `mapstructure:"git_author"`
	GitBranchFilter string `mapstructure:"git_branch_filter"`

	// Display
	ColorMode    string `mapstructure:"color"`
	DateFormat  string `mapstructure:"date_format"`
	Editor      string `mapstructure:"editor"`
	OutputFormat string `mapstructure:"output_format"`

	// Performance
	ParallelWorkers int `mapstructure:"parallel_workers"`
	CacheTTL        int `mapstructure:"cache_ttl"`

	// Paths
	ProjectPath string `mapstructure:"-"`
	DBPath      string `mapstructure:"db_path"`

	// Verbose
	Verbose bool `mapstructure:"verbose"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		ExcludePatterns: []string{
			".git", "node_modules", "vendor", "dist", "build",
			".next", "__pycache__", ".cache", "coverage",
		},
		IncludePatterns: []string{},
		TodoTypes:       []string{"TODO", "FIXME", "HACK", "BUG", "NOTE", "XXX"},
		IgnoreCase:      true,
		GitAuthor:       true,
		ColorMode:       "auto",
		DateFormat:      "2006-01-02",
		Editor:          os.Getenv("EDITOR"),
		OutputFormat:    "table",
		ParallelWorkers: 4,
		CacheTTL:        60,
		Verbose:         false,
	}
}

// Load loads configuration from file and environment
func Load() *Config {
	cfg := DefaultConfig()

	// Get project path (current directory or specified)
	projectPath, _ := os.Getwd()
	cfg.ProjectPath = projectPath

	// Try to load from config file
	homeDir, err := os.UserHomeDir()
	if err == nil {
		configPath := filepath.Join(homeDir, ".config", "todolist")
		viper.SetConfigType("toml")
		viper.SetConfigName("config")
		viper.AddConfigPath(configPath)
		viper.AddConfigPath(".")

		// Merge with defaults
		if err := viper.ReadInConfig(); err == nil {
			viper.Unmarshal(cfg)
		}
	}

	// CLI flags override config file
	// (handled in commands via viper)

	return cfg
}
