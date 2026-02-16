package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/duncan-2126/ProjectManagement/internal/config"
	"github.com/duncan-2126/ProjectManagement/internal/database"
	"github.com/duncan-2126/ProjectManagement/internal/git"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Sync TODO information with git or external services",
	Long: `Synchronize TODO information with git metadata or external services like GitHub and Jira.

Examples:
  todo sync                  # Sync with git
  todo sync github --export  # Export TODOs to GitHub Issues
  todo sync jira --export   # Export TODOs to Jira`,
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

// GitHub export command
var githubCmd = &cobra.Command{
	Use:   "github",
	Short: "Sync TODOs with GitHub Issues",
	Long: `Export TODOs to GitHub Issues or import issues from GitHub.

Examples:
  todo sync github --export   # Export TODOs to GitHub Issues`,
	RunE: func(cmd *cobra.Command, args []string) error {
		export, _ := cmd.Flags().GetBool("export")
		if export {
			return exportToGitHub()
		}
		return nil
	},
}

func exportToGitHub() error {
	// Get configuration
	cfg := config.Load()

	// Check required config
	if cfg.GitHub.Token == "" {
		return fmt.Errorf("GitHub token not configured. Run: todo config set integration.github.token <token>")
	}
	if cfg.GitHub.Owner == "" {
		return fmt.Errorf("GitHub owner not configured. Run: todo config set integration.github.owner <owner>")
	}
	if cfg.GitHub.Repo == "" {
		return fmt.Errorf("GitHub repo not configured. Run: todo config set integration.github.repo <repo>")
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

	// Get TODOs
	todos, err := db.GetTODOs(nil)
	if err != nil {
		return fmt.Errorf("failed to get TODOs: %w", err)
	}

	fmt.Printf("Exporting %d TODOs to GitHub...\n", len(todos))

	// Export each TODO as GitHub issue
	for i, todo := range todos {
		// Truncate title to 256 chars (GitHub limit)
		title := todo.Content
		if len(title) > 256 {
			title = title[:253] + "..."
		}

		// Build labels
		labels := []string{todo.Priority, todo.Category}
		if todo.Assignee != "" {
			labels = append(labels, "assigned:"+todo.Assignee)
		}

		// Filter empty labels
		var validLabels []string
		for _, label := range labels {
			if label != "" {
				validLabels = append(validLabels, label)
			}
		}

		// Determine state
		state := "open"
		if todo.Status == "closed" || todo.Status == "resolved" || todo.Status == "wontfix" {
			state = "closed"
		}

		// Build issue body
		body := fmt.Sprintf("**Source:** %s:%d\n", todo.FilePath, todo.LineNumber)
		body += fmt.Sprintf("**Author:** %s\n", todo.Author)
		body += fmt.Sprintf("**Status:** %s\n", todo.Status)
		body += fmt.Sprintf("**Priority:** %s\n", todo.Priority)
		if todo.Assignee != "" {
			body += fmt.Sprintf("**Assignee:** %s\n", todo.Assignee)
		}
		if todo.DueDate != nil {
			body += fmt.Sprintf("**Due Date:** %s\n", todo.DueDate.Format("2006-01-02"))
		}
		body += "\n---\n\n" + todo.Content

		// Create issue via GitHub API
		issueData := map[string]interface{}{
			"title":  title,
			"body":   body,
			"labels": validLabels,
			"state":  state,
		}

		jsonData, _ := json.Marshal(issueData)
		req, err := http.NewRequest("POST",
			fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", cfg.GitHub.Owner, cfg.GitHub.Repo),
			bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Failed to create request for TODO %d: %v\n", i+1, err)
			continue
		}

		req.Header.Set("Authorization", "token "+cfg.GitHub.Token)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/vnd.github.v3+json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Failed to create issue for TODO %d: %v\n", i+1, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			fmt.Printf("Failed to create issue for TODO %d: HTTP %d\n", i+1, resp.StatusCode)
			continue
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		issueNum := int(result["number"].(float64))
		fmt.Printf("Created issue #%d: %s\n", issueNum, title)
	}

	fmt.Println("GitHub export complete!")
	return nil
}

// Jira export command
var jiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Sync TODOs with Jira",
	Long: `Export TODOs to Jira issues.

Examples:
  todo sync jira --export  # Export TODOs to Jira`,
	RunE: func(cmd *cobra.Command, args []string) error {
		export, _ := cmd.Flags().GetBool("export")
		if export {
			return exportToJira()
		}
		return nil
	},
}

func exportToJira() error {
	// Get configuration
	cfg := config.Load()

	// Check required config
	if cfg.Jira.URL == "" {
		return fmt.Errorf("Jira URL not configured. Run: todo config set integration.jira.url <url>")
	}
	if cfg.Jira.Email == "" {
		return fmt.Errorf("Jira email not configured. Run: todo config set integration.jira.email <email>")
	}
	if cfg.Jira.APIToken == "" {
		return fmt.Errorf("Jira API token not configured. Run: todo config set integration.jira.api-token <token>")
	}
	if cfg.Jira.Project == "" {
		return fmt.Errorf("Jira project not configured. Run: todo config set integration.jira.project <project-key>")
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

	// Get TODOs
	todos, err := db.GetTODOs(nil)
	if err != nil {
		return fmt.Errorf("failed to get TODOs: %w", err)
	}

	fmt.Printf("Exporting %d TODOs to Jira...\n", len(todos))

	// Export each TODO as Jira issue
	for i, todo := range todos {
		// Build description
		description := fmt.Sprintf("*Source:* %s:%d\n", todo.FilePath, todo.LineNumber)
		description += fmt.Sprintf("*Author:* %s\n", todo.Author)
		description += fmt.Sprintf("*Status:* %s\n", todo.Status)
		description += fmt.Sprintf("*Priority:* %s\n", todo.Priority)
		if todo.Assignee != "" {
			description += fmt.Sprintf("*Assignee:* %s\n", todo.Assignee)
		}
		if todo.DueDate != nil {
			description += fmt.Sprintf("*Due Date:* %s\n", todo.DueDate.Format("2006-01-02"))
		}
		description += "\n---\n\n" + todo.Content

		// Truncate summary to 255 chars (Jira limit)
		summary := todo.Content
		if len(summary) > 255 {
			summary = summary[:252] + "..."
		}

		// Map priority to Jira priority
		jiraPriority := mapPriorityToJira(todo.Priority)

		// Build issue data
		issueData := map[string]interface{}{
			"fields": map[string]interface{}{
				"project": map[string]string{
					"key": cfg.Jira.Project,
				},
				"summary":     summary,
				"description": description,
				"issuetype": map[string]string{
					"name": "Task",
				},
				"priority": map[string]string{
					"name": jiraPriority,
				},
			},
		}

		// Add assignee if specified
		if todo.Assignee != "" {
			issueData["fields"].(map[string]interface{})["assignee"] = map[string]string{
				"name": todo.Assignee,
			}
		}

		// Add due date if specified
		if todo.DueDate != nil {
			issueData["fields"].(map[string]interface{})["duedate"] = todo.DueDate.Format("2006-01-02")
		}

		jsonData, _ := json.Marshal(issueData)
		req, err := http.NewRequest("POST",
			cfg.Jira.URL+"/rest/api/3/issue",
			bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Failed to create request for TODO %d: %v\n", i+1, err)
			continue
		}

		req.SetBasicAuth(cfg.Jira.Email, cfg.Jira.APIToken)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		client := &http.Client{Timeout: 30 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Failed to create Jira issue for TODO %d: %v\n", i+1, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != 201 {
			fmt.Printf("Failed to create Jira issue for TODO %d: HTTP %d\n", i+1, resp.StatusCode)
			// Read error response for debugging
			var buf bytes.Buffer
			resp.Body.Read(&buf)
			fmt.Printf("Response: %s\n", buf.String())
			continue
		}

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		key := result["key"].(string)
		fmt.Printf("Created %s: %s\n", key, summary)
	}

	fmt.Println("Jira export complete!")
	return nil
}

func mapPriorityToJira(priority string) string {
	switch priority {
	case "P0":
		return "Highest"
	case "P1":
		return "High"
	case "P2":
		return "Medium"
	case "P3":
		return "Low"
	case "P4":
		return "Lowest"
	default:
		return "Medium"
	}
}

// Config commands for integration settings
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration settings",
	Long:  "View and set configuration options including integration settings",
}

var configSetCmd = &cobra.Command{
	Use:   "set",
	Short: "Set a configuration value",
	Long: `Set a configuration value. For integration settings, use:

GitHub:
  todo config set integration.github.token <token>
  todo config set integration.github.owner <owner>
  todo config set integration.github.repo <repo>

Jira:
  todo config set integration.jira.url <url>
  todo config set integration.jira.email <email>
  todo config set integration.jira.api-token <token>
  todo config set integration.jira.project <project-key>`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 2 {
			return fmt.Errorf("requires at least 2 arguments: <key> <value>")
		}
		key := args[0]
		value := strings.Join(args[1:], " ")

		// Handle integration settings
		if strings.HasPrefix(key, "integration.") {
			parts := strings.Split(key, ".")
			if len(parts) != 3 {
				return fmt.Errorf("invalid integration key format. Use: integration.<service>.<field>")
			}
			service := parts[1]
			field := parts[2]

			// Map field names to config keys
			fieldMap := map[string]string{
				"github.token":   "github.token",
				"github.owner":   "github.owner",
				"github.repo":    "github.repo",
				"jira.url":       "jira.url",
				"jira.email":     "jira.email",
				"jira.api-token": "jira.api_token",
				"jira.project":   "jira.project",
			}

			configKey := fieldMap[key]
			if configKey == "" {
				configKey = strings.ReplaceAll(key, "-", "_")
			}

			viper.Set(configKey, value)

			// Save to config file
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("failed to get home directory: %w", err)
			}
			configPath := filepath.Join(homeDir, ".config", "todolist")
			os.MkdirAll(configPath, 0755)
			configFile := filepath.Join(configPath, "config.toml")

			return viper.SafeWriteConfigAs(configFile)
		}

		// Regular config settings
		viper.Set(key, value)

		// Save to config file
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		configPath := filepath.Join(homeDir, ".config", "todolist")
		os.MkdirAll(configPath, 0755)
		configFile := filepath.Join(configPath, "config.toml")

		return viper.SafeWriteConfigAs(configFile)
	},
}

var configGetCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a configuration value",
	Long:  "Get a configuration value by key",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("requires at least 1 argument: <key>")
		}
		key := args[0]
		value := viper.Get(key)
		if value == nil {
			return fmt.Errorf("key not found: %s", key)
		}
		fmt.Printf("%s = %v\n", key, value)
		return nil
	},
}

func init() {
	syncCmd.Flags().BoolP("blame", "b", false, "Run git blame to get author info")
	rootCmd.AddCommand(syncCmd)

	// Add GitHub and Jira subcommands
	githubCmd.Flags().Bool("export", false, "Export TODOs to GitHub Issues")
	jiraCmd.Flags().Bool("export", false, "Export TODOs to Jira")

	syncCmd.AddCommand(githubCmd)
	syncCmd.AddCommand(jiraCmd)

	// Add config commands
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configGetCmd)
	rootCmd.AddCommand(configCmd)
}
