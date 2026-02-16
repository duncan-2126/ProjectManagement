# TODO Tracker CLI - Technical Architecture (Refactored)

## Executive Summary

This document outlines the refactored technical architecture for a CLI tool that helps developers and QA track TODO comments across their codebase. This refactored version addresses gaps identified between the original technical architecture, SPEC.md, and TESTING_STRATEGY.md documents.

---

## 1. Technology Stack

| Component | Choice | Rationale |
|-----------|--------|-----------|
| **Language** | Go 1.21+ | Single binary distribution, excellent performance, great CLI ecosystem |
| **CLI Framework** | Cobra + Viper | Industry standard for Go CLI tools, built-in flags/config support |
| **Database** | SQLite (modernc.org/sqlite) | Embedded, zero-config, ACID compliant, good query performance |
| **Regex Engine** | Go stdlib `regexp` | Sufficient for TODO patterns, no external dependency needed |
| **Git Operations** | go-git library | Pure Go implementation, no git CLI dependency |
| **Output** | tview + tcell | Rich terminal UI with tables, trees, forms |
| **Testing** | Go native testing + testify | Standard Go testing, assertion library |

### Technology Stack Rationale

**Important Correction**: The TESTING_STRATEGY.md referenced Node.js/Vitest which contradicts the core Go implementation. This refactored architecture corrects this to use Go-native testing tools:
- Go's built-in `testing` package for unit tests
- `testify/assert` for expressive assertions
- `testify/mock` for mocking dependencies

---

## 2. File Parsing

### 2.1 Supported Comment Patterns

```regex
// TODO(username): description
// FIXME: description
// HACK: description
// XXX: description
// BUG(description)
// NOTE: description
// TODO: description
```

### 2.2 Language Support Matrix

| Language | Single Line | Multiline | Extensions |
|----------|-------------|-----------|------------|
| Go | `//` | `/* */` | `.go` |
| JavaScript/TypeScript | `//` | `/* */` | `.js`, `.ts`, `.jsx`, `.tsx` |
| Python | `#` | `"""` `'''` | `.py` |
| Java | `//` | `/* */` | `.java` |
| C/C++ | `//` | `/* */` | `.c`, `.cpp`, `.h`, `.hpp` |
| Rust | `//` | `/* */` | `.rs` |
| Ruby | `#` | `=begin =end` | `.rb` |
| Shell/Bash | `#` | N/A | `.sh`, `.bash` |
| SQL | `--` | `/* */` | `.sql` |
| YAML | `#` | N/A | `.yaml`, `.yml` |
| JSON | N/A | N/A | `.json` (via JSON comments if present) |
| PHP | `//` | `/* */` | `.php` |
| HTML | `<!-- -->` | N/A | `.html`, `.htm` |
| CSS/SCSS | `/* */` | Same | `.css`, `.scss`, `.sass` |

### 2.3 Parsing Strategy

1. **File Discovery**: Walk directory tree, respect `.gitignore` patterns
2. **Language Detection**: By file extension, fallback to shebang detection
3. **Pattern Matching**: Compile regex once, match per file
4. **Context Extraction**: Capture 2 lines before/after for context
5. **Deduplication**: Use content hash to detect duplicate TODOs across branches

### 2.4 Parsed Data Model

```go
type TODO struct {
    ID          string    // UUID
    FilePath    string    // Relative to repo root
    LineNumber  int
    Column      int       // Position in line
    Type        string    // TODO, FIXME, HACK, BUG, NOTE, XXX
    Content     string    // The actual comment text
    Author      string    // From git blame (optional)
    Email       string    // From git blame (optional)
    CreatedAt   time.Time // From git commit date
    UpdatedAt   time.Time // Last modified
    Status      string    // open, in_progress, blocked, resolved, wontfix, closed
    Priority    string    // P0, P1, P2, P3, P4 (aligned with SPEC.md)
    Category    string    // bug, feature, refactor, security, performance, documentation, testing, technical-debt, question, todo
    Assignee    string    // Developer assigned to work on this
    DueDate     *time.Time // Optional deadline
    Estimate    *int      // Estimated time in minutes
    Hash        string    // Content hash for deduplication
    // Relationships handled via separate table
    // Tags handled via separate table
    // Time tracking handled via separate table
}
```

---

## 3. Data Storage

### 3.1 Database Schema (SQLite) - REFACTORED

This schema aligns with both the technical requirements and SPEC.md:

```sql
CREATE TABLE todos (
    id TEXT PRIMARY KEY,
    file_path TEXT,
    line_number INTEGER DEFAULT 0,
    column_num INTEGER DEFAULT 0,
    type TEXT NOT NULL,
    content TEXT NOT NULL,
    author TEXT,
    email TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    status TEXT DEFAULT 'open' CHECK(status IN ('open', 'in_progress', 'blocked', 'resolved', 'wontfix', 'closed')),
    priority TEXT DEFAULT 'P2' CHECK(priority IN ('P0', 'P1', 'P2', 'P3', 'P4')),
    category TEXT CHECK(category IN ('bug', 'feature', 'refactor', 'security', 'performance', 'documentation', 'testing', 'technical-debt', 'question', 'todo')),
    assignee TEXT,
    due_date INTEGER,
    estimate_minutes INTEGER,
    hash TEXT NOT NULL,
    UNIQUE(hash, file_path, line_number)
);

CREATE TABLE tags (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    color TEXT DEFAULT 'default'
);

CREATE TABLE todo_tags (
    todo_id TEXT REFERENCES todos(id) ON DELETE CASCADE,
    tag_id TEXT REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (todo_id, tag_id)
);

-- Time tracking tables (NEW - aligned with SPEC.md section 3)
CREATE TABLE time_entries (
    id TEXT PRIMARY KEY,
    todo_id TEXT REFERENCES todos(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL,
    start_time INTEGER NOT NULL,
    end_time INTEGER,
    duration_minutes INTEGER,
    note TEXT,
    source TEXT DEFAULT 'manual' CHECK(source IN ('automatic', 'manual')),
    created_at INTEGER NOT NULL
);

-- Task relationships (NEW - aligned with SPEC.md section 4)
CREATE TABLE relationships (
    id TEXT PRIMARY KEY,
    source_id TEXT REFERENCES todos(id) ON DELETE CASCADE,
    target_id TEXT REFERENCES todos(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK(type IN ('parent', 'child', 'depends_on', 'blocked_by', 'relates_to'))
);

-- State transition log (NEW - aligned with SPEC.md section 2.3)
CREATE TABLE status_history (
    id TEXT PRIMARY KEY,
    todo_id TEXT REFERENCES todos(id) ON DELETE CASCADE,
    from_status TEXT,
    to_status TEXT NOT NULL,
    actor TEXT,
    note TEXT,
    created_at INTEGER NOT NULL
);

CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT UNIQUE NOT NULL,
    created_at INTEGER NOT NULL,
    last_scanned INTEGER
);

-- Saved filters (NEW - aligned with SPEC.md section 6.3)
CREATE TABLE saved_filters (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    project_id TEXT REFERENCES projects(id) ON DELETE CASCADE,
    filter_json TEXT NOT NULL,
    created_at INTEGER NOT NULL
);

-- Indexes
CREATE INDEX idx_todos_file ON todos(file_path);
CREATE INDEX idx_todos_status ON todos(status);
CREATE INDEX idx_todos_type ON todos(type);
CREATE INDEX idx_todos_author ON todos(author);
CREATE INDEX idx_todos_assignee ON todos(assignee);
CREATE INDEX idx_todos_priority ON todos(priority);
CREATE INDEX idx_todos_category ON todos(category);
CREATE INDEX idx_todos_due_date ON todos(due_date);
CREATE INDEX idx_todo_tags_todo ON todo_tags(todo_id);
CREATE INDEX idx_todo_tags_tag ON todo_tags(tag_id);
CREATE INDEX idx_time_entries_todo ON time_entries(todo_id);
CREATE INDEX idx_relationships_source ON relationships(source_id);
CREATE INDEX idx_relationships_target ON relationships(target_id);
CREATE INDEX idx_status_history_todo ON status_history(todo_id);
```

### 3.2 Storage Location

- **Project-level**: `.todo/todos.db` in project root
- **Global cache**: `~/.cache/todolist/projects.db` (index of all scanned projects)
- **Config**: `~/.config/todolist/config.toml` (global settings)

### 3.3 Migration Strategy

- Use `golang-migrate` for schema migrations
- Store migrations in embedded files
- Auto-run migrations on startup

---

## 4. CLI Framework

### 4.1 Command Structure - REFACTORED

```
todolist [OPTIONS] <command>

Commands:
  scan          Scan codebase for TODOs (default: current directory)
  list          List TODOs with filtering
  show          Show TODO details
  edit          Edit TODO status, priority, tags
  delete        Remove TODO from tracking

  # Status workflow commands (NEW)
  status        Change TODO status with transition validation
  block         Block/unblock TODO with reason
  assign        Assign TODO to user
  estimate      Set time estimate

  # Time tracking commands (NEW)
  time          Time tracking subcommands
  time start    Start time tracking for TODO
  time stop     Stop time tracking
  time add      Add manual time entry
  time report   Show time report

  # Relationship commands (NEW)
  relate        Manage TODO relationships
  deps          Show dependencies
  blockers      Show blocking TODOs
  children      Show subtasks

  # Filter & Search commands (NEW)
  search        Full-text search
  filter        Manage saved filters

  # Integration commands (NEW)
  sync          Sync with external services
  export        Export data

  # Stats & Dashboard (ENHANCED)
  stats         Show statistics dashboard
  dashboard     Show interactive dashboard

  # Core commands
  init          Initialize new project
  config        Manage configuration

  completion    Generate shell completions
  help          Show help

Options:
  -p, --project PATH   Project path (default: current directory)
  -c, --config FILE    Config file path
  -v, --verbose        Verbose output
  --color COLOR        Color output (auto|always|never)
```

### 4.2 Subcommands Detail

```go
// scan command (unchanged from original)
scan [OPTIONS]
  --exclude PATTERN    Exclude files matching pattern (can repeat)
  --include PATTERN    Include files matching pattern (can repeat)
  --types TYPES        TODO types to scan (default: all)
  --force              Force rescan even if unchanged
  --parallel N         Parallel workers (default: 4)

// list command (enhanced)
list [OPTIONS]
  --status STATUS      Filter by status (open|in_progress|blocked|resolved|wontfix|closed)
  --type TYPE          Filter by type (TODO|FIXME|HACK|BUG|NOTE|XXX)
  --author AUTHOR      Filter by author
  --assignee ASSIGNEE  Filter by assignee
  --file FILE          Filter by file path (supports glob)
  --tag TAG            Filter by tag (can repeat)
  --category CATEGORY  Filter by category
  --priority PRIORITY  Filter by priority (P0|P1|P2|P3|P4)
  --priority MIN:MAX   Filter by priority range
  --due-before DATE   Filter by due date (before)
  --due-after DATE    Filter by due date (after)
  --since DATE         Filter by created date (after)
  --until DATE         Filter by created date (before)
  --stale              Filter stale TODOs
  --sort FIELD         Sort by (file|line|type|priority|created|updated|due_date|assignee)
  --format FORMAT      Output format (table|json|csv|plain)
  --limit N            Limit results
  --saved-filter NAME  Use saved filter

// status command (NEW)
status [OPTIONS] <TODO-ID> <STATUS>
  --note NOTE          Add note for transition
  Valid transitions:
  - open → in_progress, wontfix
  - in_progress → open, blocked, resolved
  - blocked → in_progress
  - resolved → open, closed
  - wontfix → closed

// block command (NEW)
block [OPTIONS] <TODO-ID>
  --reason REASON      Reason for blocking
  --unblock            Unblock instead

// assign command (NEW)
assign [OPTIONS] <TODO-ID> <USER>

// estimate command (NEW)
estimate [OPTIONS] <TODO-ID> <DURATION>
  --duration DURATION   e.g., "4h", "30m", "2d"

// time subcommands (NEW)
time start <TODO-ID>     Start tracking time
time stop <TODO-ID>      Stop tracking time
time add <TODO-ID> [OPTIONS]
  --duration DURATION   Time spent (required)
  --note NOTE           Note for entry
  --date DATE           Date for entry
time report [OPTIONS]
  --by-todo             Group by TODO
  --by-user             Group by user
  --range RANGE         Date range (e.g., "2024-01-01:2024-01-31")
  --format FORMAT       Output format

// relate subcommands (NEW)
relate <TODO-ID> --parent <PARENT-ID>
relate <TODO-ID> --depends-on <DEP-ID>
relate <TODO-ID> --relates-to <RELATED-ID>
deps <TODO-ID>              Show what this depends on
blockers <TODO-ID>          Show what blocks this
children <TODO-ID>          Show subtasks
validate                    Check for circular dependencies

// search command (NEW)
search [OPTIONS] <QUERY>
  --regex               Use regex search
  --field FIELD         Search specific field
  --format FORMAT       Output format

// filter subcommands (NEW)
filter save <NAME> [OPTIONS]    Save current filter
filter list                      List saved filters
filter delete <NAME>            Delete saved filter

// sync subcommands (NEW)
sync github --export
sync github --import
sync jira --export
sync notion --export

// export command (NEW)
export [OPTIONS]
  --format FORMAT       json|csv|markdown|html
  --template NAME       Export template

// stats command (enhanced)
stats [OPTIONS]
  --by FIELD            Group by (type|status|author|file|priority|category)
  --trend               Show trend over time
  --chart               ASCII chart visualization

// dashboard command (NEW)
dashboard [OPTIONS]
  --watch               Auto-refresh
```

### 4.3 Configuration Management - REFACTORED

```go
type Config struct {
    // Scanning
    ExcludePatterns []string `toml:"exclude"`
    IncludePatterns []string `toml:"include"`
    TodoTypes       []string `toml:"todo_types"`
    IgnoreCase      bool     `toml:"ignore_case"`

    // Git Integration
    GitAuthor       bool     `toml:"git_author"`
    GitBranchFilter string   `toml:"git_branch_filter"`

    // Display
    ColorMode       string   `toml:"color"`
    DateFormat      string   `toml:"date_format"`
    Editor          string   `toml:"editor"`

    // Performance
    ParallelWorkers int      `toml:"parallel_workers"`
    CacheTTL        int      `toml:"cache_ttl"` // minutes

    // UI
    Pager           string   `toml:"pager"`
    Theme           string   `toml:"theme"`

    // Time Tracking
    AutoTrackTime   bool     `toml:"auto_track_time"`

    // Notifications (NEW)
    NotifyDueDays   []int    `toml:"notify_due_days"`
    NotifyTime      string   `toml:"notify_time"`
    StaleEnabled    bool     `toml:"stale_enabled"`
    StaleDaysOpen   int      `toml:"stale_days_open"`
    StaleDaysUpdate int      `toml:"stale_days_update"`

    // Integrations (NEW)
    GitHub          GitHubConfig
    Jira            JiraConfig
    Notion          NotionConfig
    Webhook         WebhookConfig
}

type GitHubConfig struct {
    Token  string `toml:"token"`
    Owner  string `toml:"owner"`
    Repo   string `toml:"repo"`
}

type JiraConfig struct {
    URL      string `toml:"url"`
    Email    string `toml:"email"`
    APIToken string `toml:"api_token"`
    Project  string `toml:"project"`
}

type NotionConfig struct {
    Token      string `toml:"token"`
    DatabaseID string `toml:"database_id"`
}

type WebhookConfig struct {
    URL    string   `toml:"url"`
    Events []string `toml:"events"`
}
```

Default config (built-in):
```toml
exclude = [".git", "node_modules", "vendor", "dist", "build", ".next", "__pycache__"]
todo_types = ["TODO", "FIXME", "HACK", "BUG", "NOTE", "XXX"]
ignore_case = true
git_author = true
parallel_workers = 4
cache_ttl = 60
color = "auto"
auto_track_time = true
notify_due_days = [3, 1, 0]
notify_time = "09:00"
stale_enabled = false
stale_days_open = 30
stale_days_update = 14
```

---

## 5. Git Integration

### 5.1 Features

1. **Author Tracking**: Use `git blame` to extract author info for each TODO
2. **Branch Analysis**: Compare TODOs between branches
3. **Commit History**: Track when TODOs were added/modified
4. **Stale Detection**: Flag TODOs not modified in N days
5. **Issue Linking**: Parse issue references (e.g., "TODO: Fix #123")
6. **TODOs in Diffs**: Show TODOs added/removed between commits

### 5.2 Implementation

```go
// GitBlameResult stores author info from git blame
type GitBlameResult struct {
    Author  string
    Email   string
    Date    time.Time
    Message string
    Hash    string
}

// Integration points:
// - On scan: optionally run git blame for author info
// - On list: filter/group by author
// - On stats: show TODO trends by commit date
// - On watch: detect new TODOs in recent commits
// - NEW: On status change: log actor from git config
```

### 5.3 Branch Comparison

```bash
# Compare TODOs between branches
todolist diff main..feature-branch

# Show TODOs added in specific commits
todolist log --since="2024-01-01"

# Show unresolved TODOs from deleted branches
todolist scan --branches=merged
```

---

## 6. IDE/Editor Integration

### 6.1 VS Code Extension

**Components:**
- `extension.js`: Main extension entry
- `commands.ts`: VS Code command handlers
- `provider.ts`: Diagnostics provider (show TODOs in Problems panel)
- `completion.ts`: TODO completion provider

**Features:**
- Real-time diagnostics (TODOs shown in Problems panel)
- Click to navigate to TODO location
- Quick actions (mark resolved, change priority)
- Status bar showing TODO counts

**Manifest (package.json):**
```json
{
  "name": "todolist-vscode",
  "version": "1.0.0",
  "engines": { "vscode": "^1.80.0" },
  "contributes": {
    "commands": [
      { "command": "todolist.scan", "title": "TODOs: Scan Project" },
      { "command": "todolist.openConfig", "title": "TODOs: Open Config" }
    ],
    "configuration": {
      "todolist.exclude": { "type": "array" },
      "todolist.autoScan": { "type": "boolean" }
    }
  }
}
```

### 6.2 Vim/Neovim Plugin

**Files:**
- `plugin/todolist.vim`: Main plugin
- `autoload/todolist/api.vim`: API functions
- `doc/todolist.txt`: Documentation

**Commands:**
```vim
:TODOScan        " Scan project for TODOs
:TODOList        " Show TODO list in quickfix
:TODOStats       " Show statistics
:TODOToggle      " Toggle TODO status
```

### 6.3 LSP Integration (Future)

- Language Server Protocol implementation
- Provide TODO diagnostics
- Text document sync for real-time updates

---

## 7. Performance Optimization

### 7.1 Scanning Strategy

1. **File Discovery**: Use `filepath.Walk` with custom filter
2. **Ignore Patterns**: Load `.gitignore`, compile to single regex
3. **Incremental Scan**: Store file mtime hash, only rescan changed files
4. **Parallel Processing**: Worker pool with N workers (default: CPU cores)

### 7.2 Caching

```
Cache Strategy:
- File metadata cache: in-memory, LRU (1000 files)
- TODO index cache: SQLite with TTL
- Git blame cache: file-based, invalidated on commit
```

### 7.3 Benchmark Targets

| Operation | Target | Maximum |
|-----------|--------|---------|
| Scan 1000 files | < 2s | 5s |
| Scan 10000 files | < 15s | 30s |
| List query (100 results) | < 100ms | 500ms |
| Database write (100 TODOs) | < 500ms | 1s |

### 7.4 Memory Management

- Stream large files instead of loading entirely
- Limit context lines to 2 before/after
- Use database connection pooling
- Release file handles immediately after read

---

## 8. Configuration

### 8.1 Config File Hierarchy

1. **Default** (embedded): Built-in sensible defaults
2. **Global** (`~/.config/todolist/config.toml`): User preferences
3. **Project** (`.todolist.toml` or `.todolist/config.toml`): Project-specific
4. **CLI flags**: Runtime overrides

Precedence (highest to lowest): CLI flags > Project config > Global config > Defaults

### 8.2 Project Initialization

```bash
# Initialize project
todolist init

# Creates:
# .todolist/
#   config.toml      # Project config
#   todos.db         # SQLite database (created on first scan)
#   .gitignore       # Add .todolist/ to gitignore
```

### 8.3 Environment Variables

| Variable | Description |
|----------|-------------|
| `TODOLIST_CONFIG` | Override config file path |
| `TODOLIST_DB` | Override database path |
| `TODOLIST_COLOR` | Force color output |
| `TODOLIST_DEBUG` | Enable debug logging |

---

## 9. Error Handling

### 9.1 Error Categories

1. **Scan Errors**: File permission denied, encoding issues
2. **Database Errors**: Lock conflicts, migration failures
3. **Git Errors**: Not a git repo, detached HEAD
4. **Config Errors**: Invalid TOML, missing required fields
5. **State Transition Errors**: Invalid status transitions
6. **Relationship Errors**: Circular dependencies, broken links

### 9.2 Logging

- Use `zap` for structured logging
- Levels: DEBUG, INFO, WARN, ERROR
- Log to `~/.cache/todolist/logs/` with rotation
- Show friendly error messages to users, detailed in verbose mode

---

## 10. External Integrations

### 10.1 GitHub Issues

```bash
# Configuration
todolist config set integration.github.token <token>
todolist config set integration.github.owner <owner>
todolist config set integration.github.repo <repo>

# Sync commands
todolist sync github --export          # Push TODOs to GitHub
todolist sync github --import          # Pull issues as TODOs
todolist sync github --bidirectional  # Two-way sync
```

### 10.2 Jira Integration

```bash
# Configuration
todolist config set integration.jira.url <jira-url>
todolist config set integration.jira.email <email>
todolist config set integration.jira.api-token <token>
todolist config set integration.jira.project <project-key>

# Sync
todolist sync jira --export
todolist sync jira --import --jql "project = PROJ AND status = Open"
```

### 10.3 Notion Integration

```bash
# Configuration
todolist config set integration.notion.token <token>
todolist config set integration.notion.database-id <db-id>

# Sync to Notion database
todolist sync notion --export
```

---

## 11. Distribution

### 11.1 Build Targets

```bash
# Cross-compile for multiple platforms
GOOS=darwin GOARCH=amd64 go build -o bin/todolist-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o bin/todolist-darwin-arm64
GOOS=linux GOARCH=amd64 go build -o bin/todolist-linux-amd64
GOOS=windows GOARCH=amd64 go build -o bin/todolist.exe
```

### 11.2 Package Managers

- **Homebrew**: `brew install todolist`
- **npm** (for Node integration): `npm install -g @todolist/cli`
- **Cargo** (future): `cargo install todolist`
- **scoop** (Windows): `scoop install todolist`

### 11.3 Installation

```bash
# Direct binary
curl -sL https://github.com/user/todolist/releases/latest | sh

# Via package manager
brew install todolist

# From source
go install github.com/user/todolist@latest
```

---

## 12. Testing Strategy (Aligned)

### 12.1 Testing Framework

| Category | Tool | Implementation |
|----------|------|----------------|
| Unit Tests | Go `testing` + `testify` | `*_test.go` files |
| Integration Tests | Go `testing` | `*_integration_test.go` |
| E2E Tests | Go `testing` + exec | `*_e2e_test.go` |
| Performance | Go `testing` + `benchmark` | `*_bench_test.go` |

### 12.2 Test Organization

```
tests/
├── unit/
│   ├── parser_test.go
│   ├── config_test.go
│   └── cli_test.go
├── integration/
│   ├── scan_test.go
│   ├── export_test.go
│   └── database_test.go
├── edge-cases/
│   ├── binary_test.go
│   ├── encoding_test.go
│   └── large_files_test.go
├── performance/
│   ├── scan_speed_test.go
│   └── memory_test.go
└── e2e/
    ├── full_workflow_test.go
    └── cli_commands_test.go
```

---

## 13. Phased Implementation Plan

### Phase 1: Core MVP
- Basic TODO scanning
- SQLite storage
- List/edit commands
- Basic filtering

### Phase 2: Enhanced Features
- Time tracking
- Task relationships
- Due dates & priorities
- Categories & tags
- Advanced filtering

### Phase 3: Integrations
- GitHub Issues sync
- Jira sync
- Notion sync
- Notifications

### Phase 4: Advanced
- Dashboard UI
- LSP integration
- Cloud sync
- Team features

---

## 14. Appendix

### A. Example Workflow

```bash
# 1. Initialize project
$ cd myproject
$ todolist init

# 2. Scan for TODOs
$ todolist scan
Found 42 TODOs in 12 files

# 3. List open TODOs
$ todolist list --status=open --format=table

# 4. Assign and prioritize
$ todolist assign abc123 john
$ todolist estimate abc123 4h
$ todolist status abc123 in_progress

# 5. Track time
$ todolist time start abc123
$ # ... work on task ...
$ todolist time stop abc123

# 6. Add relationship
$ todolist relate abc123 --depends-on xyz789

# 7. View statistics
$ todolist stats
$ todolist dashboard
```

### B. File Structure

```
todolist/
├── cmd/
│   ├── root.go
│   ├── scan.go
│   ├── list.go
│   ├── show.go
│   ├── edit.go
│   ├── delete.go
│   ├── status.go
│   ├── assign.go
│   ├── estimate.go
│   ├── time.go
│   ├── relate.go
│   ├── search.go
│   ├── filter.go
│   ├── sync.go
│   ├── export.go
│   ├── stats.go
│   ├── dashboard.go
│   └── watch.go
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── database/
│   │   ├── db.go
│   │   ├── migrations.go
│   │   └── queries.go
│   ├── parser/
│   │   ├── parser.go
│   │   ├── language.go
│   │   └── patterns.go
│   ├── git/
│   │   └── git.go
│   ├── scanner/
│   │   └── scanner.go
│   ├── time/
│   │   └── tracker.go
│   ├── relationships/
│   │   └── manager.go
│   ├── integrations/
│   │   ├── github.go
│   │   ├── jira.go
│   │   └── notion.go
│   └── ui/
│       ├── table.go
│       ├── chart.go
│       └── dashboard.go
├── testdata/
│   └── fixtures/
├── .goreleaser.yaml
├── go.mod
├── go.sum
├── main.go
└── README.md
```

### C. Dependencies

```go
// go.mod (key dependencies)
require (
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.2
    github.com/glebarez/sqlite v1.10.0
    github.com/go-git/go-git/v5 v5.11.0
    rivo/tview v0.0.0-20230902221635-6824ea5ebfab
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/jedib0t/go-pretty/v6 v6.5.3
    github.com/google/renameio v1.0.1
    github.com/mitchellh/go-homedir v1.1.0
    github.com/stretchr/testify v1.8.4
    go.uber.org/zap v1.26.0
)
```

---

## Summary of Changes

| Category | Original | Refactored | Rationale |
|----------|----------|------------|-----------|
| **Priority** | 1-5 numeric | P0-P4 labels | Aligns with SPEC.md |
| **Status States** | 4 states | 6 states | Adds BLOCKED, CLOSED per SPEC.md |
| **Tech Stack** | Ambiguous (Go vs Node) | Go (corrected) | Aligns with core implementation |
| **Database Schema** | Basic | Extended | Adds time_entries, relationships, status_history, saved_filters |
| **Commands** | ~10 commands | ~25+ commands | Full alignment with SPEC.md |
| **Time Tracking** | Not included | Full implementation | Per SPEC.md section 3 |
| **Relationships** | Not included | Full implementation | Per SPEC.md section 4 |
| **Integrations** | Not included | GitHub, Jira, Notion | Per SPEC.md section 5 |
| **Notifications** | Not included | Due dates, stale alerts | Per SPEC.md section 7 |
| **Testing Strategy** | Node.js/Vitest | Go native | Aligns with Go implementation |

---

## Consensus Status

**Status**: Pending approval

**Requested reviewers**:
- Team Lead (strategic alignment)
- Technical Architect (technical validation)
- QA Specialist (testing alignment)

---

## Dissenting Opinions

*To be documented if consensus cannot be reached.*
