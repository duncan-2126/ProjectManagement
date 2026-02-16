# TODO Tracker CLI - Technical Architecture

## Executive Summary

This document outlines the technical architecture for a CLI tool that helps developers and QA track TODO comments across their codebase. The tool provides features like TODO discovery, status tracking, Git integration, and IDE support.

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
    Status      string    // open, in_progress, blocked, resolved, wontfix, closed
    Priority    int       // 0-4 (P0-P4), default 3
    Assignee    string    // Assigned team member (NEW)
    DueDate     time.Time // Deadline for completion (NEW)
    Category    string    // Category (backend, frontend, docs, etc.) (NEW)
    Tags        []string  // user-defined tags
    Hash        string    // Content hash for deduplication
}
```

**Priority Mapping:**
- `P0` (0): Critical - Must fix immediately
- `P1` (1): High - Should fix soon
- `P2` (2): Medium - Normal priority
- `P3` (3): Low - Nice to have
- `P4` (4): Backlog - Future consideration

**Status Flow:**
- `open`: Newly discovered, not started
- `in_progress`: Being worked on
- `blocked`: Waiting on dependency
- `resolved`: Completed
- `wontfix`: Deliberately not fixing
- `closed`: Archived/duplicate

---

## 3. Data Storage

### 3.1 Database Schema (SQLite)

```sql
CREATE TABLE todos (
    id TEXT PRIMARY KEY,
    file_path TEXT NOT NULL,
    line_number INTEGER NOT NULL,
    column_num INTEGER DEFAULT 0,
    type TEXT NOT NULL,
    content TEXT NOT NULL,
    author TEXT,
    email TEXT,
    assignee TEXT,
    due_date INTEGER,
    category TEXT,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    status TEXT DEFAULT 'open' CHECK(status IN ('open', 'in_progress', 'blocked', 'resolved', 'wontfix', 'closed')),
    priority INTEGER DEFAULT 3 CHECK(priority BETWEEN 0 AND 4),
    hash TEXT NOT NULL,
    UNIQUE(hash, file_path, line_number)
);

CREATE TABLE tags (
    id TEXT PRIMARY KEY,
    name TEXT UNIQUE NOT NULL
);

CREATE TABLE todo_tags (
    todo_id TEXT REFERENCES todos(id) ON DELETE CASCADE,
    tag_id TEXT REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (todo_id, tag_id)
);

CREATE TABLE projects (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    path TEXT UNIQUE NOT NULL,
    created_at INTEGER NOT NULL,
    last_scanned INTEGER
);

-- Time tracking (Phase 2)
CREATE TABLE time_entries (
    id TEXT PRIMARY KEY,
    todo_id TEXT REFERENCES todos(id) ON DELETE CASCADE,
    started_at INTEGER NOT NULL,
    ended_at INTEGER,
    duration_seconds INTEGER,
    notes TEXT
);

-- Task relationships (Phase 2)
CREATE TABLE relationships (
    id TEXT PRIMARY KEY,
    source_id TEXT REFERENCES todos(id) ON DELETE CASCADE,
    target_id TEXT REFERENCES todos(id) ON DELETE CASCADE,
    type TEXT NOT NULL CHECK(type IN ('parent', 'child', 'depends_on', 'blocked_by', 'related')),
    created_at INTEGER NOT NULL
);

-- Activity log for audit trail (Phase 2)
CREATE TABLE activity_log (
    id TEXT PRIMARY KEY,
    todo_id TEXT REFERENCES todos(id) ON DELETE CASCADE,
    action TEXT NOT NULL,
    field_changed TEXT,
    old_value TEXT,
    new_value TEXT,
    actor TEXT,
    created_at INTEGER NOT NULL
);

CREATE INDEX idx_todos_file ON todos(file_path);
CREATE INDEX idx_todos_status ON todos(status);
CREATE INDEX idx_todos_type ON todos(type);
CREATE INDEX idx_todos_author ON todos(author);
CREATE INDEX idx_todos_assignee ON todos(assignee);
CREATE INDEX idx_todos_priority ON todos(priority);
CREATE INDEX idx_todos_category ON todos(category);
CREATE INDEX idx_time_entries_todo ON time_entries(todo_id);
CREATE INDEX idx_relationships_source ON relationships(source_id);
CREATE INDEX idx_relationships_target ON relationships(target_id);
CREATE INDEX idx_activity_log_todo ON activity_log(todo_id);
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

### 4.1 Command Structure

```
todolist [OPTIONS] <command>

Commands:
  scan          Scan codebase for TODOs (default: current directory)
  list          List TODOs with filtering
  show          Show TODO details
  edit          Edit TODO status, priority, tags, assignee, due date
  delete        Remove TODO from tracking
  stats         Show statistics dashboard
  sync          Sync with Git (update author info)
  watch         Watch mode for file changes

  # Phase 2: Time Tracking
  time          Track time spent on TODOs
  timelog       Show time log entries

  # Phase 2: Relationships
  link          Link TODOs (parent/child, depends_on)
  unlink        Remove TODO links

  # Phase 3: Integrations
  github        GitHub integration (sync issues, PRs)
  jira          Jira integration
  linear        Linear integration
  notify        Configure notifications and reminders

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
// scan command
scan [OPTIONS]
  --exclude PATTERN    Exclude files matching pattern (can repeat)
  --include PATTERN    Include files matching pattern (can repeat)
  --types TYPES        TODO types to scan (default: all)
  --force              Force rescan even if unchanged
  --parallel N         Parallel workers (default: 4)

// list command
list [OPTIONS]
  --status STATUS      Filter by status (open|in_progress|blocked|resolved|wontfix|closed)
  --type TYPE          Filter by type (TODO|FIXME|HACK|BUG|NOTE|XXX)
  --author AUTHOR      Filter by author
  --assignee USER      Filter by assignee
  --category CAT       Filter by category
  --file FILE          Filter by file path (supports glob)
  --tag TAG            Filter by tag (can repeat)
  --priority MIN:MAX   Filter by priority range (P0-P4)
  --due DATE           Filter by due date (before|after|overdue)
  --since DATE         Filter by date (after)
  --until DATE         Filter by date (before)
  --sort FIELD         Sort by (file|line|type|priority|created|updated|due_date)
  --format FORMAT      Output format (table|json|csv|plain|board)
  --limit N            Limit results

// edit command
edit [OPTIONS] <TODO-ID>
  --status STATUS      Set status (open|in_progress|blocked|resolved|wontfix|closed)
  --priority P0-P4     Set priority (0-4)
  --assignee USER      Set assignee
  --due-date DATE      Set due date (YYYY-MM-DD)
  --category CAT       Set category
  --add-tag TAG        Add tag
  --remove-tag TAG     Remove tag
  --message MSG        Add comment/note

// time command (Phase 2)
time [OPTIONS] <TODO-ID>
  start             Start timer for TODO
  stop              Stop timer for TODO
  --note TEXT       Add note to time entry

// link command (Phase 2)
link [OPTIONS] <SOURCE-ID> <TARGET-ID>
  --type TYPE        Relationship type (parent|child|depends_on|blocked_by|related)

// stats command
stats [OPTIONS]
  --by TYPE            Group by (type|status|author|file)
  --trend              Show trend over time
  --chart              ASCII chart visualization
```

### 4.3 Configuration Management

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
```

---

## 5. Git Integration

### 5.1 Features

1. **Author Tracking**: Use `git blame` to extract author info for each TODO
2. **Branch Analysis**: Compare TODOs between branches
3. **Commit History**: Track when TODOs were added/modified
4. **Stale Detection**: Flag TODOs not modified in N days
5. **Issue Linking**: Parse issue references (e.g., "TODO: Fix #123")

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

### 9.2 Logging

- Use `zap` for structured logging
- Levels: DEBUG, INFO, WARN, ERROR
- Log to `~/.cache/todolist/logs/` with rotation
- Show friendly error messages to users, detailed in verbose mode

---

## 10. Distribution

### 10.1 Build Targets

```bash
# Cross-compile for multiple platforms
GOOS=darwin GOARCH=amd64 go build -o bin/todolist-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o bin/todolist-darwin-arm64
GOOS=linux GOARCH=amd64 go build -o bin/todolist-linux-amd64
GOOS=windows GOARCH=amd64 go build -o bin/todolist.exe
```

### 10.2 Package Managers

- **Homebrew**: `brew install todolist`
- **npm** (for Node integration): `npm install -g @todolist/cli`
- **Cargo** (future): `cargo install todolist`
- **scoop** (Windows): `scoop install todolist`

### 10.3 Installation

```bash
# Direct binary
curl -sL https://github.com/user/todolist/releases/latest | sh

# Via package manager
brew install todolist

# From source
go install github.com/user/todolist@latest
```

---

## 11. Future Considerations

### Phase 2 Features (v0.2.0 - Enhanced Project Management)

| Feature | Description | Priority |
|---------|-------------|----------|
| **Time Tracking** | Track time spent on TODOs with start/stop timers, manual entry, notes | High |
| **Task Relationships** | Parent/child hierarchies, depends_on, blocked_by links | High |
| **Activity Logging** | Full audit trail of all TODO changes (who, what, when) | Medium |
| **Full-text Search** | Search TODO content with Bleve/FTS5 | Medium |
| **Assignees** | Assign TODOs to team members | Medium |
| **Due Dates** | Set deadlines with overdue detection | Medium |
| **Categories** | Organize TODOs by category (backend, frontend, docs) | Medium |
| **Kanban Board** | Visual board view (to-do, in-progress, done) | Low |
| **Custom Patterns** | User-defined regex patterns for TODO detection | Low |

### Phase 3 Features (v0.3.0 - Integrations)

| Feature | Description | Priority |
|---------|-------------|----------|
| **GitHub Integration** | Sync TODOs with GitHub Issues, link to PRs | High |
| **Jira Integration** | Two-way sync with Jira issues | High |
| **Linear Integration** | Sync with Linear workspace | Medium |
| **Notion Integration** | Export TODOs to Notion database | Medium |
| **Notifications** | Desktop reminders for due dates, daily digest | Medium |
| **Slack/Teams Bot** | Interactive bot for TODO management | Low |
| **Cloud Sync** | Optional cloud backend for team sharing | Low |

### Phase 4 Features (v0.4.0 - Advanced)

| Feature | Description | Priority |
|---------|-------------|----------|
| **AI Assistant** | Suggest TODO prioritization based on code analysis | Low |
| **Team Dashboard** | Aggregate TODOs across repositories | Low |
| **Web Dashboard** | Browser-based UI for browsing TODOs | Low |
| **CI/CD Integration** | GitHub Actions, GitLab CI plugins | Low |
| **Export Reports** | Export to Markdown, HTML, PDF reports | Low |

---

## 12. Appendix

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
+------+--------+----------+--------+
| ID   | TYPE   | FILE     | DESC   |
+------+--------+----------+--------|
| 001  | TODO   | main.go  | Add... |
| 002  | FIXME  | utils.go | Fix... |
+------+--------+----------+--------+

# 4. Edit TODO status
$ todolist edit 001 --status=in_progress

# 5. View statistics
$ todolist stats
TODOs: 42 total, 12 open, 5 resolved, 25 wontfix
By type: TODO(20), FIXME(15), BUG(7)

# 6. Watch for changes
$ todolist watch
Watching for changes... (Ctrl+C to exit)
```

### B. File Structure

```
todolist/
├── cmd/
│   ├── root.go
│   ├── scan.go
│   ├── list.go
│   ├── edit.go
│   ├── delete.go
│   ├── stats.go
│   ├── watch.go
│   ├── time.go           # Phase 2: Time tracking
│   ├── link.go          # Phase 2: Relationships
│   ├── github.go        # Phase 3: GitHub integration
│   ├── jira.go          # Phase 3: Jira integration
│   ├── linear.go        # Phase 3: Linear integration
│   └── notify.go        # Phase 3: Notifications
├── internal/
│   ├── config/
│   │   └── config.go
│   ├── database/
│   │   ├── db.go
│   │   └── migrations.go
│   ├── parser/
│   │   ├── parser.go
│   │   ├── language.go
│   │   └── patterns.go
│   ├── git/
│   │   └── git.go
│   ├── scanner/
│   │   └── scanner.go
│   ├── time/            # Phase 2: Time tracking
│   │   └── time.go
│   ├── relationships/   # Phase 2: Task relationships
│   │   └── relations.go
│   ├── integrations/    # Phase 3: External services
│   │   ├── github.go
│   │   ├── jira.go
│   │   └── linear.go
│   ├── notifications/   # Phase 3: Reminders
│   │   └── notify.go
│   ├── activity/        # Phase 2: Audit logging
│   │   └── activity.go
│   ├── search/          # Phase 2: Full-text search
│   │   └── search.go
│   └── ui/
│       ├── table.go
│       └── chart.go
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
    // Core CLI
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.2

    // Database
    github.com/glebarez/sqlite v1.10.0
    github.com/golang-migrate/migrate/v4 v4.17.0

    // Git operations
    github.com/go-git/go-git/v5 v5.11.0

    // UI/Terminal
    rivo/tview v0.0.0-20230902221635-6824ea5ebfab
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/jedib0t/go-pretty/v6 v6.5.3

    // Time tracking
    github.com/datehookz/iohook v1.0.0  // Global hotkeys

    // External integrations (Phase 3)
    github.com/google/go-github/v45 v45.2.0  // GitHub API
    github.com/andygrunwald/go-jira v1.15.1  // Jira API
    github.com/linear/linear-go v0.0.0-20240101  // Linear API (when available)
    github.com/notionhq/client v1.0.4         // Notion API

    // Notifications (Phase 3)
    github.com/gookit/notify v0.2.0           // Cross-platform notifications

    // Logging
    go.uber.org/zap v1.26.0
    github.com/rs/zerolog v1.31.0

    // Utilities
    github.com/google/renameio v1.0.1
    github.com/mitchellh/go-homedir v1.1.0
    github.com/google/uuid v1.5.0
    github.com/blevesearch/bleve/v2 v2.3.10  // Full-text search
)
```
