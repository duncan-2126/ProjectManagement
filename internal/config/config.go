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
	DateFormat   string `mapstructure:"date_format"`
	Editor       string `mapstructure:"editor"`
	OutputFormat string `mapstructure:"output_format"`

	// Performance
	ParallelWorkers int `mapstructure:"parallel_workers"`
	CacheTTL        int `mapstructure:"cache_ttl"`

	// Notifications
	Notifications NotificationsConfig `mapstructure:"notifications"`

	// Stale detection
	Stale StaleConfig `mapstructure:"stale"`

	// Digest
	Digest DigestConfig `mapstructure:"digest"`

	// Paths
	ProjectPath string `mapstructure:"-"`
	DBPath      string `mapstructure:"db_path"`

	// Verbose
	Verbose bool `mapstructure:"verbose"`

	// Integrations
	GitHub GitHubConfig `mapstructure:"github"`
	Jira   JiraConfig   `mapstructure:"jira"`
	Linear LinearConfig `mapstructure:"linear"`
	Notion NotionConfig `mapstructure:"notion"`
}

// GitHubConfig holds GitHub integration settings
type GitHubConfig struct {
	Token string `mapstructure:"token"`
	Owner string `mapstructure:"owner"`
	Repo  string `mapstructure:"repo"`
}

// JiraConfig holds Jira integration settings
type JiraConfig struct {
	URL      string `mapstructure:"url"`
	Email    string `mapstructure:"email"`
	APIToken string `mapstructure:"api_token"`
	Project  string `mapstructure:"project"`
}

// LinearConfig holds Linear integration settings (future)
type LinearConfig struct {
	APIKey  string `mapstructure:"api_key"`
	TeamID  string `mapstructure:"team_id"`
	Enabled bool   `mapstructure:"enabled"`
}

// NotionConfig holds Notion integration settings (future)
type NotionConfig struct {
	Token      string `mapstructure:"token"`
	DatabaseID string `mapstructure:"database_id"`
	Enabled    bool   `mapstructure:"enabled"`
}

// NotificationsConfig holds notification settings
type NotificationsConfig struct {
	Enabled       bool   `mapstructure:"enabled"`
	DueDaysBefore []int  `mapstructure:"due_days_before"`
	Time          string `mapstructure:"time"`
	Watch         bool   `mapstructure:"watch"`
}

// StaleConfig holds stale TODO detection settings
type StaleConfig struct {
	Enabled         bool `mapstructure:"enabled"`
	DaysOpen        int  `mapstructure:"days_open"`
	DaysSinceUpdate int  `mapstructure:"days_since_update"`
}

// DigestConfig holds daily digest settings
type DigestConfig struct {
	Enabled bool   `mapstructure:"enabled"`
	Time    string `mapstructure:"time"`
	Include string `mapstructure:"include"`
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
		Notifications: NotificationsConfig{
			Enabled:       false,
			DueDaysBefore: []int{3, 1},
			Time:          "09:00",
			Watch:         true,
		},
		Stale: StaleConfig{
			Enabled:         false,
			DaysOpen:        30,
			DaysSinceUpdate: 14,
		},
		Digest: DigestConfig{
			Enabled: false,
			Time:    "08:00",
			Include: "assigned,due-soon,stale",
		},
		GitHub: GitHubConfig{},
		Jira:   JiraConfig{},
		Linear: LinearConfig{
			Enabled: false,
		},
		Notion: NotionConfig{
			Enabled: false,
		},
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
