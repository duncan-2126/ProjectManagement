# TODO Tracker CLI - Project Management Features Specification

## 1. Task Metadata

### 1.1 Core Fields
Each TODO item tracked by the CLI will have the following core metadata:

| Field | Type | Description | Required |
|-------|------|-------------|----------|
| `id` | UUID | Unique identifier for the TODO | Auto |
| `content` | String | The actual TODO comment text | Yes |
| `file` | String | Source file path where TODO exists | Auto |
| `line` | Integer | Line number in source file | Auto |
| `author` | String | Git author who created the TODO | Auto |
| `created_at` | Timestamp | When the TODO was first detected | Auto |
| `updated_at` | Timestamp | Last time the TODO was modified | Auto |
| `status` | Enum | Current workflow state | Yes |
| `priority` | Enum | P0-P4 priority level | No |
| `due_date` | Date | Optional deadline | No |
| `assignee` | String | Developer assigned to work on this | No |
| `tags` | Array[String] | Custom categorization labels | No |
| `category` | String | Broad grouping (bug, feature, refactor, etc.) | No |

### 1.2 Priority Levels
- **P0 (Critical)**: Must fix immediately - blocking release or causing data loss
- **P1 (High)**: Important - should be addressed in current sprint
- **P2 (Medium)**: Standard priority - next sprint candidate
- **P3 (Low)**: Nice to have - back burner items
- **P4 (Trivial)**: Cosmetic, minor improvements

### 1.3 Categories
Predefined categories for automatic classification:
- `bug` - Defects and issues
- `feature` - New functionality requests
- `refactor` - Code improvement
- `security` - Security-related items
- `performance` - Optimization work
- `documentation` - Docs needed
- `testing` - Test coverage needs
- `technical-debt` - Cleanup work
- `question` - Needs discussion
- `todo` - General TODO items

### 1.4 Tags
User-defined labels for flexible organization. Tags support:
- Color coding (8 predefined colors)
- Hierarchical naming (e.g., `backend/api`, `frontend/react`)
- Auto-suggest based on existing tags

---

## 2. Status Workflow

### 2.1 States

```
┌─────────┐    ┌────────────┐    ┌───────────┐    ┌──────────┐
│  OPEN   │───▶│ IN_PROGRESS │───▶│ RESOLVED  │───▶│ CLOSED  │
└─────────┘    └────────────┘    └───────────┘    └──────────┘
     │              │                  │              ▲
     │              │                  │              │
     ▼              ▼                  ▼              │
┌─────────┐    ┌─────────┐       ┌───────────┐        │
│ BLOCKED │    │  BLOCKED│       │ WONT_FIX  │────────┘
└─────────┘    └─────────┘       └───────────┘
     │              │
     └──────────────┘ (unblock)
```

### 2.2 State Definitions

| State | Description | Valid Transitions |
|-------|-------------|-------------------|
| `open` | Newly detected, not yet started | in_progress, wontfix |
| `in_progress` | Actively being worked on | open, blocked, resolved |
| `blocked` | Cannot proceed due to dependency or external factor | in_progress |
| `resolved` | Work completed, awaiting verification | open, closed |
| `wontfix` | Decided not to address | closed |
| `closed` | Final state, archived | (terminal) |

### 2.3 State Transitions
- All transitions are logged with timestamp and actor
- Blocking/unblocking requires a reason comment
- Moving to `wontfix` requires a justification
- Reopening (`resolved` → `open`) requires explanation

---

## 3. Time Tracking

### 3.1 Automatic Time Tracking
- **Start Trigger**: When TODO status changes to `in_progress`
- **Stop Trigger**: When TODO status changes from `in_progress`
- **Session Tracking**: Each work session is recorded with start/end time

### 3.2 Manual Time Entry
Users can manually log time with:
```
todo time add <id> --duration 2h30m --note "Initial investigation"
todo time add <id> --date 2024-01-15 --duration 45m
```

### 3.3 Time Data Model

```typescript
interface TimeEntry {
  id: UUID;
  todo_id: UUID;
  user_id: string;
  start_time: Timestamp;
  end_time: Timestamp | null;
  duration_minutes: number;
  note: string;
  source: 'automatic' | 'manual';
  created_at: Timestamp;
}
```

### 3.4 Reporting Commands

| Command | Description |
|---------|-------------|
| `todo time report` | Summary for current week |
| `todo time report --by-todo` | Time grouped by TODO |
| `todo time report --by-user` | Time grouped by assignee |
| `todo time report --range 2024-01-01:2024-01-31` | Custom date range |
| `todo time export --format csv` | Export for external tools |

### 3.5 Time Estimates
- Users can set estimated time: `todo estimate <id> 4h`
- Track estimated vs actual for velocity calculations
- Display in dashboard and reports

---

## 4. Task Relationships

### 4.1 Relationship Types

| Type | Description | Behavior |
|------|-------------|----------|
| `parent` | Parent task containing subtasks | Parent completion requires all children closed |
| `child` | Subtask of a parent | Cannot be closed if parent is open |
| `depends_on` | Blocking dependency | Cannot move to in_progress until dependency resolved |
| `blocked_by` | Inverse of depends_on | Informational, shows what blocks this task |
| `relates_to` | Soft association | Visual grouping, no enforcement |

### 4.2 Commands

```bash
# Create relationships
todo relate <id> --parent <parent_id>
todo relate <id> --depends-on <dependency_id>
todo relate <id> --relates-to <related_id>

# View relationships
todo deps <id>           # Show what this task depends on
todo blockers <id>       # Show what blocks this task
todo children <id>       # Show subtasks

# Validate
todo validate            # Check for circular dependencies, broken links
```

### 4.3 Dependency Rules
- Circular dependencies are prevented at creation time
- If a dependency moves to `wontfix`, dependent tasks are notified
- Bulk dependency updates supported

---

## 5. External Integrations

### 5.1 GitHub Issues

```bash
# Configuration
todo config set integration.github.token <token>
todo config set integration.github.owner <owner>
todo config set integration.github.repo <repo>

# Sync commands
todo sync github --export          # Push TODOs to GitHub
todo sync github --import          # Pull issues as TODOs
todo sync github --bidirectional  # Two-way sync
```

**Mapping:**
| TODO Field | GitHub Field |
|------------|--------------|
| title | title (truncated to 256 chars) |
| content | body |
| status | state (open/closed) |
| priority | labels (P0, P1, etc.) |
| assignee | assignee |
| tags | labels |

### 5.2 Jira Integration

```bash
# Configuration
todo config set integration.jira.url <jira-url>
todo config set integration.jira.email <email>
todo config set integration.jira.api-token <token>
todo config set integration.jira.project <project-key>

# Sync
todo sync jira --export
todo sync jira --import --jql "project = PROJ AND status = Open"
```

**Field Mapping:**
| TODO Field | Jira Field |
|------------|------------|
| title | summary |
| content | description |
| priority | priority |
| assignee | assignee |
| due_date | duedate |
| status | status (mapped via workflow) |

### 5.3 Linear Integration

```bash
# Configuration
todo config set integration.linear.api-key <key>
todo config set integration.linear.team-id <team-id>

# Sync
todo sync linear --export
todo sync linear --import --filter "createdAfter:today"
```

### 5.4 Notion Integration

```bash
# Configuration
todo config set integration.notion.token <token>
todo config set integration.notion.database-id <db-id>

# Sync to Notion database
todo sync notion --export
```

### 5.5 Export Formats

```bash
# Built-in export
todo export --format json
todo export --format csv
todo export --format markdown --template summary

# Custom templates
todo export --format html --template dashboard
```

---

## 6. Filtering & Search

### 6.1 Filter Syntax

```bash
# Single filters
todo list --status open
todo list --priority P0
todo list --assignee john
todo list --tag bug
todo list --category feature
todo list --file "src/**/*.ts"
todo list --author jane
todo list --due-before 2024-02-01
todo list --due-after 2024-01-01
todo list --created-after 2024-01-01

# Combined filters (AND logic)
todo list --status open --priority P0 --tag urgent

# OR logic with pipes
todo list --status open,P1 --priority P0,P1

# Exclude
todo list --exclude-tag wontfix
todo list --exclude-status closed
```

### 6.2 Search

```bash
# Full-text search
todo search "authentication error"
todo search --regex "TODO|FIXME|XXX"

# Advanced search
todo search --field content --match "login"
todo search --field file --match "**/auth*.ts"
```

### 6.3 Saved Filters

```bash
# Save filter
todo filter save my-sprint --status in-progress --assignee me

# List saved filters
todo filter list

# Apply saved filter
todo list @my-sprint

# Delete filter
todo filter delete my-sprint
```

### 6.4 Sorting

```bash
todo list --sort priority        # By priority (P0 first)
todo list --sort due_date       # By due date (soonest first)
todo list --sort created_at     # By creation date
todo list --sort updated_at     # By last update
todo list --sort status         # By status
todo list --sort priority --desc  # Reverse order
```

---

## 7. Notifications & Reminders

### 7.1 Due Date Reminders

```bash
# Configure reminders
todo config set notifications.due-days-before 3,1
todo config set notifications.due-days-before 0  # Day of reminder

# Notification times
todo config set notifications.time "09:00"  # Daily check time
```

**Notification Channels:**
- CLI desktop notifications (native OS notifications)
- Configurable webhook for custom integrations

### 7.2 Stale TODO Alerts

```bash
# Configure stale detection
todo config set stale.enabled true
todo config set stale.days-open 30      # Consider stale after 30 days
todo config set stale.days-since-update 14  # No update in 14 days

# List stale items
todo list --stale
```

### 7.3 Activity Notifications

```bash
# Get notified when:
# - TODO assigned to you
# - TODO you're watching gets updated
# - Dependency you own is resolved
# - Status changes on items you care about

# Watch a TODO
todo watch <id>

# Configure watch notifications
todo config set notifications.watch true
```

### 7.4 Daily Digest

```bash
# Enable daily summary
todo config set digest.enabled true
todo config set digest.time "08:00"
todo config set digest.include "assigned,due-soon,stale"
```

**Digest Contents:**
- TODOs assigned to you
- Upcoming due dates
- Stale items you own
- Recent updates on watched items

---

## 8. User Interface

### 8.1 Dashboard

```
┌─────────────────────────────────────────────────────────────┐
│  TODO Dashboard                              Feb 15, 2024   │
├─────────────────────────────────────────────────────────────┤
│  Summary                    │  By Status                    │
│  ─────────                  │  ─────────                    │
│  Total: 42                  │  Open: 12 ████████            │
│  My Tasks: 8                │  In Progress: 5 ███           │
│  Due Soon: 3                │  Blocked: 2 █                 │
│  Overdue: 1                 │  Resolved: 18 ███████████      │
│                             │  Closed: 5 ███                │
├─────────────────────────────┴──────────────────────────────┤
│  Priority Distribution                                        │
│  P0: 2  P1: 8  P2: 15  P3: 12  P4: 5                         │
├─────────────────────────────────────────────────────────────┤
│  Due Soon (next 7 days)                                      │
│  ─────────────────────                                       │
│  [P0] Fix login crash - auth.ts:142 - due in 2 days         │
│  [P1] Add API rate limiting - api.ts:89 - due in 5 days     │
│  [P2] Update documentation - README.md:45 - due in 7 days   │
├─────────────────────────────────────────────────────────────┤
│  Stale Items (no update > 30 days)                           │
│  ─────────────────────────────────                           │
│  [P3] Refactor user service - user.ts:200 - 45 days old      │
└─────────────────────────────────────────────────────────────┘
```

### 8.2 List View

```
$ todo list --status open --priority P0,P1

ID     │ Priority │ Status    │ Title                    │ Due Date   │ Assignee
───────┼──────────┼───────────┼──────────────────────────┼────────────┼──────────
a1b2c3 │ P0       │ open      │ Fix login crash          │ 2024-02-17 │ john
d4e5f6 │ P0       │ open      │ Security vulnerability   │ 2024-02-18 │ jane
g7h8i9 │ P1       │ open      │ Add payment integration  │ 2024-02-25 │ john
```

### 8.3 Detail View

```
$ todo show a1b2c3

TODO #a1b2c3
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Status:     open          Priority:  P0
Category:   bug           Assignee:  john
Created:    2024-01-15    Updated:    2024-02-10
Due Date:   2024-02-17    Estimate:  4h

Title: Fix login crash on OAuth flow
────────────────────────────────────────────────────
File:   src/auth/oauth.ts
Line:   142
Author: jane

Description:
───────────
OAuth callback fails when user has no email permission.
Need to handle null email case.

Tags: [bug] [oauth] [urgent]
────────────────────────────────────────────────────

Time Tracked: 2h 30m (3 sessions)
────────────────────────────────────────────────────

Dependencies: (2)
───────────────
→ Depends on: #x9y8z7 (Implement OAuth token refresh)
← Blocked by: none

Activity Log:
─────────────
2024-02-10 14:30 - john changed status to open
2024-02-09 10:15 - jane added tag 'urgent'
2024-01-15 09:00 - jane created this TODO
```

---

## 9. Configuration

### 9.1 Global Settings

```bash
# Editor for multi-line content
todo config set editor "vim"

# Default priority for new TODOs
todo config set defaults.priority P2

# Auto-assign based on file ownership
todo config set autoassign.enabled true
todo config set autoassign.by-file "**/*.ts=john,**/*.py=jane"

# Default status for new TODOs
todo config set defaults.status open

# Date format
todo config set display.date-format "yyyy-mm-dd"
todo config set display.time-format "HH:mm"
```

### 9.2 Project-Specific Settings

Settings can be scoped to the current directory:
```bash
todo config set --project defaults.priority P1
todo config set --project sync.jira.enabled true
```

---

## 10. Data Storage

### 10.1 Local Storage
- SQLite database in `.todo/` directory
- JSON backup files for portability
- Git-compatible storage option

### 10.2 Schema

```sql
CREATE TABLE todos (
  id TEXT PRIMARY KEY,
  content TEXT NOT NULL,
  file TEXT,
  line INTEGER,
  author TEXT,
  status TEXT DEFAULT 'open',
  priority TEXT,
  due_date TEXT,
  assignee TEXT,
  category TEXT,
  created_at TEXT,
  updated_at TEXT
);

CREATE TABLE tags (
  id TEXT PRIMARY KEY,
  name TEXT UNIQUE NOT NULL,
  color TEXT
);

CREATE TABLE todo_tags (
  todo_id TEXT,
  tag_id TEXT,
  PRIMARY KEY (todo_id, tag_id)
);

CREATE TABLE time_entries (
  id TEXT PRIMARY KEY,
  todo_id TEXT,
  user_id TEXT,
  start_time TEXT,
  end_time TEXT,
  duration_minutes INTEGER,
  note TEXT,
  source TEXT
);

CREATE TABLE relationships (
  id TEXT PRIMARY KEY,
  source_id TEXT,
  target_id TEXT,
  type TEXT
);
```

---

## 11. Command Summary

| Command | Description |
|---------|-------------|
| `todo init` | Initialize TODO tracking in project |
| `todo scan` | Scan codebase for TODO comments |
| `todo list` | List TODOs with filters |
| `todo show <id>` | Show TODO details |
| `todo add` | Create new TODO manually |
| `todo edit <id>` | Edit TODO content/metadata |
| `todo status <id> <status>` | Update status |
| `todo priority <id> <priority>` | Update priority |
| `todo assign <id> <user>` | Assign TODO |
| `todo tag <id> <tag>` | Add/remove tags |
| `todo relate <id>` | Manage relationships |
| `todo time <id>` | Time tracking commands |
| `todo search <query>` | Search TODOs |
| `todo filter` | Manage saved filters |
| `todo sync <service>` | Sync with external services |
| `todo export` | Export data |
| `todo config` | Configure settings |
| `todo dashboard` | Show summary dashboard |

---

## 12. Acceptance Criteria

### 12.1 Core Functionality
- [ ] Scan and detect TODO comments in code files
- [ ] Track status workflow correctly
- [ ] Support all metadata fields
- [ ] Persist data locally

### 12.2 Filtering & Search
- [ ] Filter by status, priority, assignee, tags, category
- [ ] Filter by file path and author
- [ ] Date range filtering (created, updated, due)
- [ ] Full-text search
- [ ] Saved filters

### 12.3 Time Tracking
- [ ] Automatic start/stop on status change
- [ ] Manual time entry
- [ ] Time reports

### 12.4 Relationships
- [ ] Parent/child relationships
- [ ] Dependencies with validation
- [ ] Blocked/blocking views

### 12.5 Integrations
- [ ] GitHub Issues sync (export)
- [ ] Jira sync (export)
- [ ] JSON/CSV export

### 12.6 Notifications
- [ ] Due date reminders
- [ ] Stale TODO detection
- [ ] Daily digest option

---

*Version: 1.0*
*Last Updated: 2024-02-16*
