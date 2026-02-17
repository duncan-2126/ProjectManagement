# TODO Tracker CLI

A command-line tool to help developers and QA track TODO comments across their codebase. Automatically discovers TODO, FIXME, HACK, BUG, NOTE, and XXX comments and provides a complete workflow for managing technical debt.

## Features

- **Automated Discovery**: Scan your codebase for TODO comments
- **Multi-language Support**: Go, JavaScript, TypeScript, Python, Java, C/C++, Rust, Ruby, Shell, SQL, and more
- **Status Workflow**: Track TODOs from open to resolved
- **Priority Levels**: P0-P4 priority system
- **Git Integration**: Author attribution via git blame
- **Web GUI**: React dashboard and Kanban/list/search views via `todo serve`
- **Export Options**: JSON, CSV, and Markdown export
- **Watch Mode**: Auto-scan on file changes
- **Statistics Dashboard**: Visualize your technical debt

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/duncan-2126/ProjectManagement.git
cd ProjectManagement

# Build the binary
go build -o todo .

# Add to PATH (optional)
mv todo /usr/local/bin/
```

### Using Go Install

```bash
go install github.com/duncan-2126/ProjectManagement@latest
```

## Quick Start

### 1. Initialize a Project

```bash
todo init
```

This creates a `.todo/` directory with:
- `todos.db` - SQLite database
- `config.toml` - Project configuration

### 2. Scan for TODOs

```bash
todo scan
```

### 3. List TODOs

```bash
# List all TODOs
todo list

# Filter by status
todo list --status open

# Filter by type
todo list --type FIXME

# Filter by priority
todo list --priority P0
```

### 4. Manage TODOs

```bash
# View TODO details
todo show <id>

# Update status
todo edit <id> --status resolved

# Update priority
todo edit <id> --priority P1

# Delete a TODO
todo delete <id>
```

### 5. Export

```bash
# Export as JSON
todo export --format json

# Export as CSV
todo export --format csv

# Export as Markdown
todo export --format markdown
```

### 6. Git Integration

```bash
# Sync with git (get author info)
todo sync --blame
```

### 7. Watch Mode

```bash
# Watch for changes and auto-scan
todo watch

# Custom interval
todo watch --interval 1m
```

### 8. Statistics

```bash
todo stats
```

## Commands

| Command | Description |
|---------|-------------|
| `todo init` | Initialize a new project |
| `todo scan` | Scan codebase for TODOs |
| `todo list` | List all TODOs |
| `todo show <id>` | Show TODO details |
| `todo edit <id>` | Edit a TODO |
| `todo delete <id>` | Delete a TODO |
| `todo export` | Export TODOs |
| `todo sync` | Sync with git |
| `todo watch` | Watch for changes |
| `todo stats` | Show statistics |

## Filtering Options

```bash
# Filter by status
--status open|in_progress|blocked|resolved|wontfix|closed

# Filter by type
--type TODO|FIXME|HACK|BUG|NOTE|XXX

# Filter by priority
--priority P0|P1|P2|P3|P4

# Filter by author
--author <name>

# Filter by file
--file <path>
```

## Output Formats

```bash
# Table format (default)
todo list --format table

# JSON
todo list --format json

# CSV
todo list --format csv
```

## Configuration

Configuration is loaded from (in order of precedence):
1. Command-line flags
2. Project config (`.todo/config.toml`)
3. Global config (`~/.config/todolist/config.toml`)
4. Default values

### Example Configuration

```toml
# .todo/config.toml

[todo_types]
default = ["TODO", "FIXME", "HACK", "BUG", "NOTE", "XXX"]

[exclude]
default = [".git", "node_modules", "vendor", "dist", "build"]

[git]
author = true

[display]
color = "auto"
date_format = "2006-01-02"

[performance]
parallel_workers = 4
cache_ttl = 60
```

## Development

### Running Tests

```bash
go test ./...
```

### Building

```bash
# Build for current platform
go build -o todo .

# Cross-compile
GOOS=darwin GOARCH=amd64 go build -o todo-darwin-amd64 .
GOOS=linux GOARCH=amd64 go build -o todo-linux-amd64 .
GOOS=windows GOARCH=amd64 go build -o todo.exe .
```

### Web GUI Build

```bash
cd web
npm install
npm run build
cd ..
```

Then run:

```bash
todo serve --host 127.0.0.1 --port 8080
```

## Supported Languages

| Language | Extensions | Comment Style |
|----------|-------------|----------------|
| Go | .go | // |
| JavaScript | .js, .jsx, .mjs | // |
| TypeScript | .ts, .tsx | // |
| Python | .py | # |
| Java | .java | // |
| C/C++ | .c, .cpp, .h, .hpp | // |
| Rust | .rs | // |
| Ruby | .rb | # |
| Shell | .sh, .bash, .zsh | # |
| SQL | .sql | -- |
| YAML | .yaml, .yml | # |
| PHP | .php | // |
| CSS/SCSS | .css, .scss, .sass | /* */ |
| HTML | .html, .htm | <!-- --> |

## Architecture

```
ProjectManagement/
├── cmd/                    # CLI commands
│   ├── root.go            # Root command
│   ├── scan.go            # Scan command
│   ├── list.go            # List command
│   ├── show.go            # Show command
│   ├── edit.go            # Edit command
│   ├── delete.go          # Delete command
│   ├── export.go          # Export command
│   ├── sync.go            # Git sync command
│   ├── watch.go           # Watch command
│   ├── stats.go           # Statistics command
│   └── init.go            # Init command
├── internal/
│   ├── config/            # Configuration
│   ├── database/          # SQLite database
│   ├── parser/            # TODO parser
│   └── git/               # Git integration
├── main.go                # Entry point
└── go.mod                 # Go module
```

## Contributing

1. Create a feature branch: `git checkout -b feature/my-feature`
2. Make your changes
3. Add tests if applicable
4. Commit with clear messages
5. Push to your branch
6. Create a Pull Request

## License

MIT License - see LICENSE file for details

## Support

- Issues: https://github.com/duncan-2126/ProjectManagement/issues
