package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
)

// Author represents git author information
type Author struct {
	Name  string
	Email string
	Date  time.Time
}

// BlameResult contains blame information for a file
type BlameResult struct {
	Lines []BlameLine
}

// BlameLine represents a single line's blame info
type BlameLine struct {
	LineNumber int
	Author     string
	Email      string
	Date       time.Time
	CommitHash string
}

// GetBlame runs git blame on a file and returns author information
func GetBlame(filePath string) (map[int]Author, error) {
	result := make(map[int]Author)

	// Try using go-git first
	repo, err := git.PlainOpen(".")
	if err != nil {
		// Fall back to command line
		return getBlameCLI(filePath)
	}

	if _, err := repo.Head(); err != nil {
		return getBlameCLI(filePath)
	}

	// Use the correct go-git blame API
	// For now, fall back to CLI as the go-git blame API is complex
	return getBlameCLI(filePath)

	return result, nil
}

func getBlameCLI(filePath string) (map[int]Author, error) {
	result := make(map[int]Author)

	cmd := exec.Command("git", "blame", "--line-porcelain", filePath)
	output, err := cmd.Output()
	if err != nil {
		return result, nil // Return empty result on failure
	}

	lines := strings.Split(string(output), "\n")
	var currentAuthor Author

	authorRegex := regexp.MustCompile(`author (.+)`)
	emailRegex := regexp.MustCompile(`author-mail <(.+)>`)
	dateRegex := regexp.MustCompile(`author-time (\d+)`)
	lineRegex := regexp.MustCompile(`^\t(\d+)$`)

	for _, line := range lines {
		if match := authorRegex.FindStringSubmatch(line); match != nil {
			currentAuthor = Author{Name: match[1]}
		}
		if match := emailRegex.FindStringSubmatch(line); match != nil {
			currentAuthor.Email = match[1]
		}
		if match := dateRegex.FindStringSubmatch(line); match != nil {
			ts, _ := strconv.ParseInt(match[1], 10, 64)
			currentAuthor.Date = time.Unix(ts, 0)
		}
		if match := lineRegex.FindStringSubmatch(line); match != nil {
			lineNum, _ := strconv.Atoi(match[1])
			if currentAuthor.Name != "" {
				result[lineNum] = currentAuthor
			}
		}
	}

	return result, nil
}

// IsRepo checks if the current directory is a git repository
func IsRepo() bool {
	_, err := git.PlainOpen(".")
	return err == nil
}

// GetCurrentBranch returns the current git branch name
func GetCurrentBranch() (string, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return "", err
	}

	head, err := repo.Head()
	if err != nil {
		return "", err
	}

	return head.Name().Short(), nil
}

// GetCommits TODO appears in returns commits that modified a TODO
func GetCommits(filePath string, lineNumber int) ([]string, error) {
	cmd := exec.Command("git", "log", "-n", "10", "-p", "--follow",
		"-S", fmt.Sprintf("TODO"), "--", filePath)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(output), "\n")
	var commits []string
	commitRegex := regexp.MustCompile(`^commit ([a-f0-9]+)`)

	for _, line := range lines {
		if match := commitRegex.FindStringSubmatch(line); match != nil {
			commits = append(commits, match[1])
		}
	}

	return commits, nil
}

// InitRepo initializes a git repository if not already one
func InitRepo() error {
	_, err := git.PlainInit(".", false)
	return err
}
