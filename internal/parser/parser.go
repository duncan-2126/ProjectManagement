package parser

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Language represents a programming language with comment patterns
type Language struct {
	Name       string
	Extensions []string
	SingleLine []string // Comment prefixes
	MultiLine  []string // Multi-line comment start/end pairs
	Shebangs   []string
}

// Supported languages
var Languages = []Language{
	{
		Name:       "Go",
		Extensions: []string{".go"},
		SingleLine: []string{"//"},
		MultiLine:  []string{"/*", "*/"},
	},
	{
		Name:       "JavaScript",
		Extensions: []string{".js", ".jsx", ".mjs"},
		SingleLine: []string{"//"},
		MultiLine:  []string{"/*", "*/"},
	},
	{
		Name:       "TypeScript",
		Extensions: []string{".ts", ".tsx"},
		SingleLine: []string{"//"},
		MultiLine:  []string{"/*", "*/"},
	},
	{
		Name:       "Python",
		Extensions: []string{".py"},
		SingleLine: []string{"#"},
		MultiLine:  []string{`"""`, `"""`},
	},
	{
		Name:       "Java",
		Extensions: []string{".java"},
		SingleLine: []string{"//"},
		MultiLine:  []string{"/*", "*/"},
	},
	{
		Name:       "C",
		Extensions: []string{".c", ".h"},
		SingleLine: []string{"//"},
		MultiLine:  []string{"/*", "*/"},
	},
	{
		Name:       "C++",
		Extensions: []string{".cpp", ".cc", ".cxx", ".hpp"},
		SingleLine: []string{"//"},
		MultiLine:  []string{"/*", "*/"},
	},
	{
		Name:       "Rust",
		Extensions: []string{".rs"},
		SingleLine: []string{"//"},
		MultiLine:  []string{"/*", "*/"},
	},
	{
		Name:       "Ruby",
		Extensions: []string{".rb"},
		SingleLine: []string{"#"},
		MultiLine:  []string{"=begin", "=end"},
	},
	{
		Name:       "Shell",
		Extensions: []string{".sh", ".bash", ".zsh"},
		SingleLine: []string{"#"},
		Shebangs:   []string{"#!/bin/sh", "#!/bin/bash", "#!/usr/bin/env"},
	},
	{
		Name:       "SQL",
		Extensions: []string{".sql"},
		SingleLine: []string{"--"},
		MultiLine:  []string{"/*", "*/"},
	},
	{
		Name:       "YAML",
		Extensions: []string{".yaml", ".yml"},
		SingleLine: []string{"#"},
	},
	{
		Name:       "PHP",
		Extensions: []string{".php"},
		SingleLine: []string{"//", "#"},
		MultiLine:  []string{"/*", "*/"},
	},
	{
		Name:       "CSS",
		Extensions: []string{".css", ".scss", ".sass", ".less"},
		SingleLine: []string{},
		MultiLine:  []string{"/*", "*/"},
	},
	{
		Name:       "HTML",
		Extensions: []string{".html", ".htm"},
		SingleLine: []string{},
		MultiLine:  []string{"<!--", "-->"},
	},
}

// TODO types to look for
var TODOTypes = []string{
	"TODO",
	"FIXME",
	"HACK",
	"BUG",
	"NOTE",
	"XXX",
}

// Pattern to match TODO comments
var todoPattern = regexp.MustCompile(`(?i)\b(TODO|FIXME|HACK|BUG|NOTE|XXX)\s*(\([^)]+\))?\s*[:\-]?\s*(.*)$`)

// ParsedTODO represents a parsed TODO comment
type ParsedTODO struct {
	FilePath   string
	LineNumber int
	Column     int
	Type       string
	Content    string
	Author     string
	Email      string
	CreatedAt  time.Time
	Hash       string
}

// Parser handles parsing TODO comments from files
type Parser struct {
	includePatterns []string
	excludePatterns []string
	todoTypes       []string
}

// New creates a new parser
func New(include, exclude []string, todoTypes []string) *Parser {
	if len(todoTypes) == 0 {
		todoTypes = TODOTypes
	}
	return &Parser{
		includePatterns: include,
		excludePatterns: exclude,
		todoTypes:       todoTypes,
	}
}

// GetLanguageByExtension returns the language for a file extension
func GetLanguageByExtension(ext string) *Language {
	ext = strings.ToLower(ext)
	for i := range Languages {
		for _, e := range Languages[i].Extensions {
			if strings.ToLower(e) == ext {
				return &Languages[i]
			}
		}
	}
	return nil
}

// ParseFile parses a single file for TODO comments
func (p *Parser) ParseFile(filePath string) ([]ParsedTODO, error) {
	// Check if file should be excluded
	relPath, err := filepath.Rel(".", filePath)
	if err != nil {
		relPath = filePath
	}

	for _, pattern := range p.excludePatterns {
		matched, err := filepath.Match(pattern, relPath)
		if err == nil && matched {
			return nil, nil
		}
	}

	// Check if file matches include patterns
	if len(p.includePatterns) > 0 {
		matched := false
		for _, pattern := range p.includePatterns {
			m, err := filepath.Match(pattern, relPath)
			if err == nil && m {
				matched = true
				break
			}
		}
		if !matched {
			return nil, nil
		}
	}

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Check file size
	info, err := file.Stat()
	if err != nil {
		return nil, err
	}
	if info.Size() > 10*1024*1024 { // Skip files > 10MB
		return nil, nil
	}

	// Get language
	ext := filepath.Ext(filePath)
	lang := GetLanguageByExtension(ext)

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(content), "\n")
	var todos []ParsedTODO

	// Track multi-line comments
	inMLComment := false

	for lineNum, line := range lines {
		// Skip binary files
		if lineNum == 0 && strings.Contains(line, "\x00") {
			return nil, nil
		}

		trimmed := strings.TrimSpace(line)

		// Handle multi-line comments
		if lang != nil && len(lang.MultiLine) >= 2 {
			for i := 0; i < len(lang.MultiLine); i += 2 {
				startToken := lang.MultiLine[i]
				endToken := lang.MultiLine[i+1]

				if !inMLComment && strings.Contains(trimmed, startToken) {
					inMLComment = true
					mlStartToken = startToken
					// Check if comment ends on same line
					if strings.Contains(trimmed, endToken) {
						inMLComment = false
					}
					trimmed = extractMLComment(trimmed, startToken, endToken)
				} else if inMLComment && strings.Contains(trimmed, endToken) {
					inMLComment = false
					trimmed = ""
					break
				}
			}
		}

		if inMLComment {
			continue
		}

		// Check for single-line comments
		isComment := false
		if lang != nil {
			for _, comment := range lang.SingleLine {
				if strings.HasPrefix(trimmed, comment) {
					isComment = true
					trimmed = strings.TrimPrefix(trimmed, comment)
					trimmed = strings.TrimSpace(trimmed)
					break
				}
			}
		}

		// Check for TODO pattern
		if isComment || (lang == nil && strings.HasPrefix(trimmed, "#")) {
			matches := todoPattern.FindStringSubmatch(trimmed)
			if len(matches) >= 2 {
				todoType := strings.ToUpper(matches[1])
				var content string
				if len(matches) >= 4 {
					content = matches[3]
				}

				// Generate hash for deduplication
				hash := fmt.Sprintf("%x", sha256.Sum256([]byte(filePath+fmt.Sprint(lineNum+1)+todoType+content)))

				todo := ParsedTODO{
					FilePath:   relPath,
					LineNumber: lineNum + 1,
					Column:     strings.Index(line, matches[0]) + 1,
					Type:       todoType,
					Content:    content,
					CreatedAt:  time.Now(),
					Hash:       hash,
				}
				todos = append(todos, todo)
			}
		}
	}

	return todos, nil
}

// extractMLComment extracts content from multi-line comment
func extractMLComment(line, start, end string) string {
	result := strings.TrimSpace(line)
	result = strings.TrimPrefix(result, start)
	result = strings.TrimSuffix(result, end)
	return strings.TrimSpace(result)
}

// ParseDir recursively parses a directory for TODO comments
func (p *Parser) ParseDir(dirPath string) ([]ParsedTODO, error) {
	var allTodos []ParsedTODO

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		path := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			// Skip excluded directories
			skip := false
			for _, pattern := range p.excludePatterns {
				if entry.Name() == pattern || strings.HasPrefix(entry.Name(), ".") {
					skip = true
					break
				}
			}
			if skip {
				continue
			}

			subTodos, err := p.ParseDir(path)
			if err != nil {
				continue
			}
			allTodos = append(allTodos, subTodos...)
		} else {
			todos, err := p.ParseFile(path)
			if err != nil {
				continue
			}
			if todos != nil {
				allTodos = append(allTodos, todos...)
			}
		}
	}

	return allTodos, nil
}
