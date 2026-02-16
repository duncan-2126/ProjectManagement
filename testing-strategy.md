# TODO Tracker CLI - Testing Strategy

## Executive Summary

This document defines the comprehensive testing strategy for the TODO Tracker CLI tool. The testing approach follows the principle of testing at the appropriate level - unit tests for isolated logic, integration tests for workflows, and end-to-end tests for complete user scenarios.

---

## 1. Test Organization Structure

```
todolist/
├── cmd/                    # Command handlers
│   └── commands_test.go
├── internal/
│   ├── config/             # Configuration parsing
│   │   └── config_test.go
│   ├── database/           # Database operations
│   │   ├── db_test.go
│   │   └── migrations_test.go
│   ├── parser/             # TODO parsing logic
│   │   ├── parser_test.go
│   │   ├── language_test.go
│   │   └── patterns_test.go
│   ├── git/                # Git integration
│   │   └── git_test.go
│   ├── scanner/            # File scanning
│   │   └── scanner_test.go
│   └── ui/                 # Output formatting
│       └── ui_test.go
├── testdata/              # Test fixtures
│   ├── fixtures/
│   │   ├── languages/      # Source files in different languages
│   │   │   ├── go/
│   │   │   ├── javascript/
│   │   │   ├── python/
│   │   │   └── ...
│   │   ├── edge_cases/    # Edge case files
│   │   │   ├── binary/
│   │   │   ├── large/
│   │   │   └── encoding/
│   │   ├── git/            # Git test repositories
│   │   └── config/        # Config file samples
│   └── golden/             # Golden files for output comparison
├── e2e/                   # End-to-end tests
│   ├── scan_test.go
│   ├── list_test.go
│   └── workflow_test.go
├── integration/           # Integration tests
│   ├── file_system_test.go
│   └── git_integration_test.go
├── performance/           # Performance benchmarks
│   └── benchmarks_test.go
└── testutil/              # Test utilities
    ├── fixtures.go
    └── helpers.go
```

---

## 2. Unit Tests

### 2.1 Parser Testing (`internal/parser/`)

#### 2.1.1 Pattern Matching Tests

Test all supported TODO patterns with various formats:

```go
// Test cases for TODO patterns
var todoPatternTests = []struct {
    name     string
    input    string
    expected []ParsedTODO
}{
    {
        name:  "simple TODO",
        input: "// TODO: Implement authentication",
        expected: []ParsedTODO{{
            Type:    "TODO",
            Content: "Implement authentication",
            Line:    1,
        }},
    },
    {
        name:  "TODO with username",
        input: "// TODO(john): Fix memory leak",
        expected: []ParsedTODO{{
            Type:    "TODO",
            Content: "Fix memory leak",
            Author:  "john",
            Line:    1,
        }},
    },
    {
        name:  "FIXME comment",
        input: "// FIXME: This breaks on Windows",
        expected: []ParsedTODO{{
            Type:    "FIXME",
            Content: "This breaks on Windows",
            Line:    1,
        }},
    },
    {
        name:  "HACK comment",
        input: "// HACK: Workaround for API limitation",
        expected: []ParsedTODO{{
            Type:    "HACK",
            Content: "Workaround for API limitation",
            Line:    1,
        }},
    },
    {
        name:  "XXX comment",
        input: "// XXX: Refactor this mess",
        expected: []ParsedTODO{{
            Type:    "XXX",
            Content: "Refactor this mess",
            Line:    1,
        }},
    },
    {
        name:  "BUG with parens",
        input: "// BUG(memory): Buffer overflow in handler",
        expected: []ParsedTODO{{
            Type:    "BUG",
            Content: "Buffer overflow in handler",
            Author:  "memory",
            Line:    1,
        }},
    },
    {
        name:  "NOTE comment",
        input: "// NOTE: Consider using缓存 for performance",
        expected: []ParsedTODO{{
            Type:    "NOTE",
            Content: "Consider using缓存 for performance",
            Line:    1,
        }},
    },
    {
        name:  "multiline block comment",
        input: "/* TODO: This is\na multiline TODO */",
        expected: []ParsedTODO{{
            Type:    "TODO",
            Content: "This is a multiline TODO",
            Line:    1,
        }},
    },
    {
        name:  "no TODO present",
        input: "// This is just a regular comment",
        expected: nil,
    },
}
```

#### 2.1.2 Context Extraction Tests

Verify that surrounding lines are captured correctly:

```go
func TestParser_ContextExtraction(t *testing.T) {
    input := `package main

import "fmt"

func main() {
    // TODO: Add error handling
    fmt.Println("Hello")
}

func helper() {
    // Related code
}`

    parser := NewParser()
    results := parser.Parse("test.go", input)

    require.Len(t, results, 1)
    assert.Equal(t, 2, results[0].ContextBefore) // "import" and blank
    assert.Equal(t, 2, results[0].ContextAfter)  // "fmt.Println" and "}"
}
```

### 2.2 Language Detection Tests (`internal/parser/language.go`)

Test file extension to language mapping:

```go
func TestLanguageDetector_Detect(t *testing.T) {
    tests := []struct {
        filename string
        expected Language
    }{
        {"main.go", LanguageGo},
        {"app.js", LanguageJavaScript},
        {"app.ts", LanguageTypeScript},
        {"component.tsx", LanguageTypeScript},
        {"app.py", LanguagePython},
        {"main.rs", LanguageRust},
        {"App.java", LanguageJava},
        {"main.c", LanguageC},
        {"utils.cpp", LanguageCPP},
        {"script.sh", LanguageShell},
        {"queries.sql", LanguageSQL},
        {"config.yaml", LanguageYAML},
        {"config.yml", LanguageYAML},
        {"unknown.xyz", LanguageUnknown},
    }

    detector := NewLanguageDetector()
    for _, tt := range tests {
        t.Run(tt.filename, func(t *testing.T) {
            lang := detector.Detect(tt.filename)
            assert.Equal(t, tt.expected, lang)
        })
    }
}
```

Test shebang detection for shell scripts without extensions:

```go
func TestLanguageDetector_Shebang(t *testing.T) {
    tests := []struct {
        content  string
        expected Language
    }{
        {"#!/bin/bash\necho hello", LanguageShell},
        {"#!/usr/bin/env python3\nprint('hi')", LanguagePython},
        {"#!/usr/bin/env node\nconsole.log('hi')", LanguageJavaScript},
    }

    detector := NewLanguageDetector()
    for _, tt := range tests {
        t.Run(tt.expected.String(), func(t *testing.T) {
            lang := detector.DetectFromContent(tt.content)
            assert.Equal(t, tt.expected, lang)
        })
    }
}
```

### 2.3 Configuration Tests (`internal/config/`)

Test config loading from various sources:

```go
func TestConfig_Load(t *testing.T) {
    tests := []struct {
        name        string
        configFile   string
        envVars      map[string]string
        flags        CLIFlags
        expected     *Config
        expectErr    bool
    }{
        {
            name:     "default config",
            expected: DefaultConfig(),
            expectErr: false,
        },
        {
            name:      "load from file",
            configFile: "testdata/config/custom.toml",
            expected: &Config{
                ExcludePatterns: []string{".git", "node_modules"},
                ParallelWorkers: 8,
            },
            expectErr: false,
        },
        {
            name:     "invalid TOML",
            configFile: "testdata/config/invalid.toml",
            expectErr: true,
        },
        {
            name:     "missing required field",
            configFile: "testdata/config/missing_field.toml",
            expectErr: true,
        },
        {
            name:     "environment variable override",
            envVars:  map[string]string{"TODOLIST_PARALLEL_WORKERS": "16"},
            expected: &Config{ParallelWorkers: 16},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 2.4 Database Tests (`internal/database/`)

Test CRUD operations:

```go
func TestDatabase_TODOOperations(t *testing.T) {
    db := setupTestDB(t)
    defer db.Close()

    t.Run("Insert TODO", func(t *testing.T) {
        todo := &TODO{
            ID:          "test-1",
            FilePath:    "main.go",
            LineNumber:  10,
            Type:        "TODO",
            Content:     "Test TODO",
            Status:      "open",
            Hash:        "abc123",
        }
        err := db.InsertTODO(todo)
        assert.NoError(t, err)
    })

    t.Run("Get TODO by ID", func(t *testing.T) {
        todo, err := db.GetTODO("test-1")
        assert.NoError(t, err)
        assert.Equal(t, "main.go", todo.FilePath)
        assert.Equal(t, 10, todo.LineNumber)
    })

    t.Run("Update TODO status", func(t *testing.T) {
        err := db.UpdateTODOStatus("test-1", "resolved")
        assert.NoError(t, err)

        todo, _ := db.GetTODO("test-1")
        assert.Equal(t, "resolved", todo.Status)
    })

    t.Run("Delete TODO", func(t *testing.T) {
        err := db.DeleteTODO("test-1")
        assert.NoError(t, err)

        _, err = db.GetTODO("test-1")
        assert.Error(t, err)
    })

    t.Run("List TODOs with filters", func(t *testing.T) {
        todos, err := db.ListTODOs(ListFilter{Status: "open"})
        assert.NoError(t, err)
        assert.Len(t, todos, 0) // All deleted
    })
}
```

Test database migrations:

```go
func TestDatabase_Migrations(t *testing.T) {
    t.Run("run migrations on new database", func(t *testing.T) {
        db, _ := sql.Open("sqlite", ":memory:")
        err := RunMigrations(db)
        assert.NoError(t, err)
    })

    t.Run("migrations are idempotent", func(t *testing.T) {
        db, _ := sql.Open("sqlite", ":memory:")
        RunMigrations(db)
        err := RunMigrations(db) // Should not fail
        assert.NoError(t, err)
    })

    t.Run("migration from v1 to v2", func(t *testing.T) {
        // Test upgrade path
    })
}
```

### 2.5 Git Integration Tests (`internal/git/`)

Test git blame parsing:

```go
func TestGitBlameParser_Parse(t *testing.T) {
    // Use golden file for git blame output
    output := readGoldenFile("testdata/git/blame.txt")
    results, err := ParseBlameOutput(output)

    assert.NoError(t, err)
    assert.Len(t, results, 3)

    assert.Equal(t, "john@example.com", results[0].Email)
    assert.Equal(t, "2024-01-15", results[0].Date.Format("2006-01-02"))
}
```

---

## 3. Integration Tests

### 3.1 File System Operations

Test complete scanning workflows:

```go
func TestScanner_FullScan(t *testing.T) {
    // Create temporary directory with test files
    tmpDir := t.TempDir()
    createTestFiles(t, tmpDir, map[string]string{
        "main.go":   "package main\n// TODO: Implement main",
        "utils.go":  "package main\n// FIXME: Bug here",
        "main.js":   "// TODO: Add validation",
        "styles.css": "/* TODO: Style this */",
    })

    scanner := NewScanner(tmpDir, &Config{
        ParallelWorkers: 2,
    })

    results, err := scanner.Scan()

    assert.NoError(t, err)
    assert.Len(t, results, 3)
    assertContainsTODO(t, results, "main.go", "TODO")
    assertContainsTODO(t, results, "utils.go", "FIXME")
    assertContainsTODO(t, results, "main.js", "TODO")
}
```

### 3.2 Git Integration Workflows

```go
func TestGitIntegration_Workflow(t *testing.T) {
    // Use test git repository
    repoDir := setupTestGitRepo(t)
    defer os.RemoveAll(repoDir)

    // Create file with TODO
    writeFile(t, filepath.Join(repoDir, "main.go"), `package main
// TODO: Implement feature`)

    // Initial commit
    runGit(t, repoDir, "add", ".")
    runGit(t, repoDir, "commit", "-m", "Initial commit")

    // Scan and get TODOs
    scanner := NewScanner(repoDir, &Config{GitAuthor: true})
    results, _ := scanner.Scan()

    // Verify author info is populated from git
    assert.NotEmpty(t, results[0].Author)
    assert.NotEmpty(t, results[0].CreatedAt)
}
```

### 3.3 End-to-End CLI Tests

Test complete command workflows:

```go
func TestCLI_ScanAndListWorkflow(t *testing.T) {
    // Use golden master approach for CLI output
    cmd := &exec.Cmd{
        Path: "./todolist",
        Args: []string{"./todolist", "scan", "-p", "testdata/fixtures/languages/go"},
    }
    output, err := cmd.CombinedOutput()

    assert.NoError(t, err, string(output))

    // Compare against golden output
    expected := readGoldenFile("testdata/golden/scan_go.txt")
    assert.Equal(t, string(expected), string(output))
}
```

---

## 4. Edge Cases

### 4.1 Binary Files

Ensure binary files are correctly skipped:

```go
func TestEdgeCases_BinaryFiles(t *testing.T) {
    tmpDir := t.TempDir()

    // Create various binary files
    binaryFiles := []string{
        "image.png",
        "document.pdf",
        "archive.zip",
        "binary.exe",
        "library.so",
    }

    for _, name := range binaryFiles {
        path := filepath.Join(tmpDir, name)
        err := os.WriteFile(path, []byte{0x89, 0x50, 0x4E, 0x47}, 0644)
        assert.NoError(t, err)
    }

    scanner := NewScanner(tmpDir, &Config{})
    results, err := scanner.Scan()

    assert.NoError(t, err)
    assert.Len(t, results, 0, "Binary files should be skipped")
}
```

### 4.2 Large Files

Test handling of files exceeding typical sizes:

```go
func TestEdgeCases_LargeFiles(t *testing.T) {
    tmpDir := t.TempDir()

    // Create a large file (10MB)
    largeFile := filepath.Join(tmpDir, "large.go")
    f, _ := os.Create(largeFile)
    defer f.Close()

    writer := bufio.NewWriter(f)
    for i := 0; i < 100000; i++ {
        fmt.Fprintln(writer, "// TODO: Line", i)
    }
    writer.Flush()

    scanner := NewScanner(tmpDir, &Config{})
    results, err := scanner.Scan()

    assert.NoError(t, err)
    assert.Len(t, results, 100000)
}
```

### 4.3 Different Encodings

Test UTF-8, UTF-16, Latin-1, and other encodings:

```go
func TestEdgeCases_Encodings(t *testing.T) {
    tests := []struct {
        name     string
        content  string
        encoding string
    }{
        {
            name:     "UTF-8",
            content:  "// TODO: Implementación en español",
            encoding: "utf-8",
        },
        {
            name:     "UTF-16LE",
            content:  "// TODO: 中文注释",
            encoding: "utf-16",
        },
        {
            name:     "Latin-1",
            content:  "// TODO: Réfërençe",
            encoding: "latin-1",
        },
        {
            name:     "Mixed emoji",
            content:  "// TODO: Fix 🔥 bug",
            encoding: "utf-8",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tmpDir := t.TempDir()
            file := filepath.Join(tmpDir, "test.go")
            writeFileWithEncoding(t, file, tt.content, tt.encoding)

            scanner := NewScanner(tmpDir, &Config{})
            results, err := scanner.Scan()

            assert.NoError(t, err)
            assert.Len(t, results, 1)
            assert.Contains(t, results[0].Content, "TODO")
        })
    }
}
```

### 4.4 Symlinks

Test handling of symbolic links:

```go
func TestEdgeCases_Symlinks(t *testing.T) {
    tmpDir := t.TempDir()

    // Create target file with TODO
    targetDir := filepath.Join(tmpDir, "target")
    os.MkdirAll(targetDir, 0755)
    targetFile := filepath.Join(targetDir, "main.go")
    os.WriteFile(targetFile, []byte("// TODO: Target file"), 0644)

    // Create symlink
    linkFile := filepath.Join(tmpDir, "link.go")
    os.Symlink(targetFile, linkFile)

    scanner := NewScanner(tmpDir, &Config{})
    results, err := scanner.Scan()

    // Should handle symlinks appropriately (either skip or follow based on config)
    assert.NoError(t, err)
}
```

### 4.5 Permission Errors

Test graceful handling of permission denied:

```go
func TestEdgeCases_PermissionErrors(t *testing.T) {
    tmpDir := t.TempDir()

    // Create unreadable file
    secretFile := filepath.Join(tmpDir, "secret.go")
    os.WriteFile(secretFile, []byte("// TODO: Secret"), 0000)
    defer os.Chmod(secretFile, 0644) // Clean up

    scanner := NewScanner(tmpDir, &Config{Verbose: true})
    results, err := scanner.Scan()

    // Should complete without fatal error
    assert.NoError(t, err)
    // Log warning about permission denied should be recorded
}
```

### 4.6 Empty and Whitespace-Only Files

```go
func TestEdgeCases_EmptyFiles(t *testing.T) {
    tmpDir := t.TempDir()

    files := []string{
        "empty.go",
        "whitespace.go",
        "only_comments.go", // Contains only comments without TODO
    }

    for _, name := range files {
        path := filepath.Join(tmpDir, name)
        content := map[string]string{
            "empty.go":         "",
            "whitespace.go":    "   \n\t\n   ",
            "only_comments.go": "// Regular comment\n/* Block comment */",
        }[name]
        os.WriteFile(path, []byte(content), 0644)
    }

    scanner := NewScanner(tmpDir, &Config{})
    results, err := scanner.Scan()

    assert.NoError(t, err)
    assert.Len(t, results, 0)
}
```

---

## 5. Language Support Testing

Create comprehensive test files for each supported language:

### 5.1 Test Fixtures by Language

```
testdata/fixtures/languages/
├── go/
│   ├── single_line.go
│   ├── multiline.go
│   └── edge_cases.go
├── javascript/
│   ├── es6.js
│   ├── typescript.ts
│   ├── jsx.jsx
│   └── tsx.tsx
├── python/
│   ├── simple.py
│   ├── multiline.py
│   └── decorators.py
├── rust/
│   ├── lib.rs
│   └── main.rs
├── java/
│   └── Main.java
├── c/
│   ├── main.c
│   └── header.h
├── cpp/
│   ├── main.cpp
│   └── template.hpp
├── ruby/
│   └── script.rb
├── shell/
│   └── script.sh
├── sql/
│   └── queries.sql
└── yaml/
    └── config.yml
```

### 5.2 Language-Specific Tests

```go
func TestLanguageSupport_AllLanguages(t *testing.T) {
    languages := []struct {
        dir      string
        expected int // expected TODO count
    }{
        {"go", 5},
        {"javascript", 4},
        {"python", 3},
        {"rust", 3},
        {"java", 2},
        {"c", 2},
        {"cpp", 2},
        {"ruby", 2},
        {"shell", 2},
        {"sql", 2},
        {"yaml", 1},
    }

    for _, lang := range languages {
        t.Run(lang.dir, func(t *testing.T) {
            dir := filepath.Join("testdata/fixtures/languages", lang.dir)
            scanner := NewScanner(dir, &Config{})
            results, err := scanner.Scan()

            assert.NoError(t, err)
            assert.Equal(t, lang.expected, len(results),
                "Mismatch in %s language support", lang.dir)
        })
    }
}
```

---

## 6. Error Handling Tests

### 6.1 Invalid Configuration Files

```go
func TestErrorHandling_InvalidConfig(t *testing.T) {
    tests := []struct {
        name      string
        config    string
        expectErr string
    }{
        {
            name: "invalid TOML syntax",
            config: `
parallel_workers = "not a number"
exclude = .
`,
            expectErr: "toml: expected integer",
        },
        {
            name:      "missing required field",
            config:    "parallel_workers = 4",
            expectErr: "required field 'exclude' missing",
        },
        {
            name:      "invalid enum value",
            config: `
exclude = [".git"]
color = "invalid"
`,
            expectErr: "invalid color mode",
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            tmpFile := filepath.Join(t.TempDir(), "config.toml")
            os.WriteFile(tmpFile, []byte(tt.config), 0644)

            _, err := LoadConfig(tmpFile)
            assert.Error(t, err)
            assert.Contains(t, err.Error(), tt.expectErr)
        })
    }
}
```

### 6.2 Database Corruption

```go
func TestErrorHandling_CorruptedDatabase(t *testing.T) {
    tmpDir := t.TempDir()
    dbPath := filepath.Join(tmpDir, "todos.db")

    // Write corrupted data
    os.WriteFile(dbPath, []byte("not a sqlite database"), 0644)

    db, err := OpenDatabase(dbPath)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "not a valid database")
}
```

### 6.3 Missing Directory Permissions

```go
func TestErrorHandling_MissingDirectory(t *testing.T) {
    scanner := NewScanner("/nonexistent/path", &Config{})
    _, err := scanner.Scan()

    assert.Error(t, err)
    assert.Contains(t, err.Error(), "no such file or directory")
}
```

---

## 7. Performance Testing

### 7.1 Benchmark Tests

```go
func BenchmarkParser_SimpleFile(b *testing.B) {
    content := "// TODO: Test content"
    parser := NewParser()

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        parser.Parse("test.go", content)
    }
}

func BenchmarkScanner_LargeProject(b *testing.B) {
    tmpDir := createLargeTestProject(b, 1000, 100)
    scanner := NewScanner(tmpDir, &Config{ParallelWorkers: 4})

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        scanner.Scan()
    }
}

func BenchmarkDatabase_Insert(b *testing.B) {
    db := setupTestDB(b)
    defer db.Close()

    todos := generateTODOs(100)

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        for _, todo := range todos {
            db.InsertTODO(todo)
        }
    }
}

func BenchmarkDatabase_Query(b *testing.B) {
    db := setupTestDB(b)
    defer db.Close()

    // Insert test data
    for i := 0; i < 1000; i++ {
        db.InsertTODO(generateTODO(fmt.Sprintf("todo-%d", i)))
    }

    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        db.ListTODOs(ListFilter{Status: "open", Limit: 100})
    }
}
```

### 7.2 Memory Profiling Tests

```go
func TestPerformance_MemoryUsage(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping memory test in short mode")
    }

    tmpDir := createLargeTestProject(t, 5000, 50)

    var memStatsBefore, memStatsAfter runtime.MemStats
    runtime.GC()
    runtime.ReadMemStats(&memStatsBefore)

    scanner := NewScanner(tmpDir, &Config{ParallelWorkers: 4})
    results, err := scanner.Scan()

    runtime.ReadMemStats(&memStatsAfter)

    assert.NoError(t, err)
    assert.Less(t, memStatsAfter.Mallocs-memStatsBefore.Mallocs, uint64(100000),
        "Too many memory allocations")
}
```

### 7.3 Incremental vs Full Scan

```go
func TestPerformance_IncrementalScan(t *testing.T) {
    tmpDir := t.TempDir()
    createProject(t, tmpDir, 100, 10)

    scanner := NewScanner(tmpDir, &Config{})

    // Full scan
    start := time.Now()
    _, _ = scanner.Scan()
    fullScanDuration := time.Since(start)

    // Modify one file
    time.Sleep(100 * time.Millisecond) // Ensure mtime changes
    modifyOneFile(t, tmpDir)

    // Incremental scan
    start = time.Now()
    _, _ = scanner.Scan()
    incrementalScanDuration := time.Since(start)

    // Incremental should be significantly faster
    assert.Less(t, incrementalScanDuration, fullScanDuration/2,
        "Incremental scan should be much faster than full scan")
}
```

---

## 8. CI/CD Pipeline

### 8.1 GitHub Actions Workflow

```yaml
# .github/workflows/test.yml
name: Test

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ${{ matrix.os }}
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        go: ['1.21', '1.22', '1.23']

    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: ${{ matrix.go }}

      - name: Download dependencies
        run: go mod download

      - name: Run unit tests
        run: go test -v -race -coverprofile=coverage.out ./...

      - name: Run integration tests
        run: go test -v -tags=integration ./integration/...

      - name: Upload coverage
        uses: codecov/codecov-action@v4
        with:
          files: ./coverage.out
          fail_ci_if_error: true
          threshold: 70%

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest

  build:
    needs: [test, lint]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - name: Build
        run: go build -v -o bin/todolist .
      - name: Upload artifact
        uses: actions/upload-artifact@v4
        with:
          name: todolist
          path: bin/todolist

  e2e:
    needs: [build]
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Download binary
        uses: actions/download-artifact@v4
        with:
          name: todolist
      - chmod +x todolist
      - name: Run e2e tests
        run: go test -v ./e2e/...
```

### 8.2 Makefile for Local Testing

```makefile
# Makefile
.PHONY: test test-unit test-integration test-e2e test-coverage lint

test: test-unit test-integration

test-unit:
	go test -v -race -coverprofile=coverage.out ./...

test-integration:
	go test -v -tags=integration ./integration/...

test-e2e:
	go test -v ./e2e/...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	open coverage.html

lint:
	golangci-lint run ./...

bench:
	go test -bench=. -benchmem ./performance/...

race:
	go test -race ./...

clean-test:
	rm -rf coverage.out coverage.html
```

---

## 9. Test Coverage Requirements

### 9.1 Coverage Targets

| Component | Minimum Coverage |
|-----------|------------------|
| Parser | 90% |
| Database | 85% |
| Config | 90% |
| Git | 80% |
| Scanner | 85% |
| UI/Output | 70% |
| **Overall** | **80%** |

### 9.2 Coverage Enforcement

```yaml
# In CI, fail if coverage drops below threshold
- name: Check coverage
  run: |
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    COVERAGE=${COVERAGE%\%}
    if (( $(echo "$COVERAGE < 80" | bc -l) )); then
      echo "Coverage $COVERAGE% is below 80%"
      exit 1
    fi
```

---

## 10. Testing Utilities

### 10.1 Test Helper Functions

```go
// testutil/helpers.go
package testutil

func setupTestDB(t *testing.T) *Database {
    tmpFile := filepath.Join(t.TempDir(), "test.db")
    db, err := OpenDatabase(tmpFile)
    require.NoError(t, err)
    return db
}

func createTestProject(t *testing.T, dir string, fileCount, todosPerFile int) {
    for i := 0; i < fileCount; i++ {
        filename := filepath.Join(dir, fmt.Sprintf("file%d.go", i))
        content := generateGoFile(todosPerFile)
        err := os.WriteFile(filename, []byte(content), 0644)
        require.NoError(t, err)
    }
}

func assertContainsTODO(t *testing.T, todos []ParsedTODO, path, todoType string) {
    for _, todo := range todos {
        if todo.FilePath == path && todo.Type == todoType {
            return
        }
    }
    t.Errorf("Expected TODO of type %s in %s", todoType, path)
}
```

### 10.2 Golden File Testing

```go
// testutil/golden.go
package testutil

func AssertGolden(t *testing.T, name string, actual []byte) {
    goldenPath := filepath.Join("testdata", "golden", name)

    if *updateGolden {
        err := os.WriteFile(goldenPath, actual, 0644)
        require.NoError(t, err)
        t.Logf("Updated golden file: %s", goldenPath)
        return
    }

    expected, err := os.ReadFile(goldenPath)
    require.NoError(t, err)
    assert.Equal(t, string(expected), string(actual))
}
```

---

## 11. Test Execution Guidelines

### 11.1 Running Tests

```bash
# Run all unit tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/parser/...

# Run with race detector
go test -race ./...

# Run integration tests
go test -tags=integration ./integration/...

# Run e2e tests
go test ./e2e/...

# Run benchmarks
go test -bench=. -benchmem ./performance/

# Update golden files
go test -args -update-golden ./...
```

### 11.2 Test Tags

- `unit`: Unit tests (default)
- `integration`: Integration tests
- `e2e`: End-to-end tests
- `slow`: Tests that take longer to run
- `race`: Tests that check race conditions

---

## 12. Summary

This testing strategy provides comprehensive coverage of the TODO Tracker CLI:

1. **Unit Tests**: Isolated testing of parser, config, database, and git modules
2. **Integration Tests**: File system and git workflow testing
3. **Edge Cases**: Binary files, large files, encodings, permissions
4. **Language Support**: Test fixtures for all supported languages
5. **Error Handling**: Invalid configs, corrupted databases, missing files
6. **Performance**: Benchmarks for scanning, database operations, memory usage
7. **CI/CD**: Automated GitHub Actions workflow with coverage enforcement

The strategy ensures the CLI tool is robust, performant, and reliable across different environments and use cases.
