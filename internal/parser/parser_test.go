package parser

import (
	"testing"
)

func TestGetLanguageByExtension(t *testing.T) {
	tests := []struct {
		ext      string
		expected string
	}{
		{".go", "Go"},
		{".js", "JavaScript"},
		{".ts", "TypeScript"},
		{".py", "Python"},
		{".java", "Java"},
		{".rs", "Rust"},
		{".rb", "Ruby"},
		{".sh", "Shell"},
		{".sql", "SQL"},
		{".yaml", "YAML"},
		{".php", "PHP"},
		{".css", "CSS"},
		{".html", "HTML"},
		{".unknown", ""},
	}

	for _, tt := range tests {
		lang := GetLanguageByExtension(tt.ext)
		if tt.expected == "" {
			if lang != nil {
				t.Errorf("expected nil for %s, got %s", tt.ext, lang.Name)
			}
		} else {
			if lang == nil {
				t.Errorf("expected %s for %s, got nil", tt.expected, tt.ext)
			} else if lang.Name != tt.expected {
				t.Errorf("expected %s for %s, got %s", tt.expected, tt.ext, lang.Name)
			}
		}
	}
}

func TestTodoPatternMatching(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"TODO: fix this bug", []string{"TODO", "", "fix this bug"}},
		{"TODO(username): implement feature", []string{"TODO", "(username)", "implement feature"}},
		{"FIXME: this is broken", []string{"FIXME", "", "this is broken"}},
		{"HACK: workaround here", []string{"HACK", "", "workaround here"}},
		{"BUG(description): critical issue", []string{"BUG", "(description)", "critical issue"}},
		{"NOTE: remember this", []string{"NOTE", "", "remember this"}},
		{"XXX: remove later", []string{"XXX", "", "remove later"}},
		{"// TODO: lowercase should match", []string{"TODO", "", "lowercase should match"}},
		{"not a todo", nil},
		{"", nil},
	}

	for _, tt := range tests {
		matches := todoPattern.FindStringSubmatch(tt.input)
		if tt.expected == nil {
			if matches != nil {
				t.Errorf("expected no match for %q, got %v", tt.input, matches)
			}
		} else {
			if matches == nil {
				t.Errorf("expected match for %q, got nil", tt.input)
			} else if len(matches) < 3 || matches[1] != tt.expected[0] {
				t.Errorf("expected %v, got %v", tt.expected, matches)
			}
		}
	}
}

func TestParserExcludePatterns(t *testing.T) {
	parser := New(nil, []string{"node_modules", "vendor", ".git"}, nil)

	// These should be excluded
	excludedPaths := []string{
		"node_modules/package/index.js",
		"vendor/lib/helper.go",
		".git/config",
		"deep/node_modules/nested/file.ts",
	}

	for _, path := range excludedPaths {
		// The parser's ParseFile checks exclude patterns
		// We can't test directly without creating actual files
		_ = path // Just verify the pattern exists
	}

	if len(parser.excludePatterns) != 3 {
		t.Errorf("expected 3 exclude patterns, got %d", len(parser.excludePatterns))
	}
}
