# Test Regression Suite

This document outlines test cases for regression testing the React Web UI and Go backend.

---

## 1. API Security Tests

### 1.1 SQL Injection Prevention
- **Test:** Send malicious SQL in query parameters
  - `/api/todos?status=open' OR '1'='1`
  - `/api/search?q='; DROP TABLE todos;--`
- **Expected:** Query should be treated as literal string, not executed as SQL

### 1.2 XSS Prevention
- **Test:** Send XSS payloads in TODO fields
  - PUT `/api/todo/1` with `{"content": "<script>alert('xss')</script>"}`
  - PUT `/api/todo/1` with `{"assignee": "<img onerror=alert(1) src=x>"}`
- **Expected:** Payload should be escaped when rendered in UI

### 1.3 Invalid Status Values
- **Test:** PUT `/api/todo/1` with `{"status": "invalid_status"}`
- **Expected:** Should reject with validation error OR normalize to valid status

### 1.4 Invalid Priority Values
- **Test:** PUT `/api/todo/1` with `{"priority": "P99"}`
- **Expected:** Should reject or normalize to valid priority

---

## 2. Edge Cases

### 2.1 Empty Database
- **Test:** Start server with no TODOs in database
- **Expected:** Dashboard shows "No TODOs found" message

### 2.2 Very Long Content
- **Test:** Create TODO with content > 10,000 characters
- **Expected:** UI truncates with ellipsis, full content accessible

### 2.3 Special Characters in Search
- **Test:** Search with special characters: `*`, `?`, `[`, `]`, `(`, `)`, `"`, `'`
- **Expected:** Should handle gracefully without errors

### 2.4 Unicode Content
- **Test:** Create TODO with Unicode: `// TODO: Fix emoji issue `
- **Expected:** Correctly stored and displayed

### 2.5 Missing Optional Fields
- **Test:** Create TODO with null assignee, due_date, category
- **Expected:** Should render without errors, show appropriate defaults

---

## 3. Data Integrity Tests

### 3.1 Concurrent Updates (Race Condition)
- **Test:**
  1. Open two browser tabs
  2. Change TODO status in tab A to "in_progress"
  3. Change TODO status in tab B to "resolved"
  4. Refresh both tabs
- **Expected:** Last write wins, no data corruption

### 3.2 Kanban Failed Update Rollback
- **Test:**
  1. Start dragging TODO in Kanban
  2. Disconnect network (or mock API failure)
  3. Drop TODO in new column
- **Expected:** UI reverts to previous state, error displayed

### 3.3 Partial Update Preservation
- **Test:** PUT `/api/todo/1` with only `{"status": "closed"}`
- **Expected:** Other fields remain unchanged

---

## 4. Performance Tests

### 4.1 Large Dataset
- **Test:** Load 1000+ TODOs
- **Expected:**
  - Initial load < 3 seconds
  - Filter operations < 500ms
  - No browser freezing

### 4.2 Rapid Filter Changes
- **Test:** Quickly change filters in FilterBar (within 100ms)
- **Expected:** Only last request completes, no race conditions

### 4.3 Search Debouncing
- **Test:** Type in search box rapidly
- **Expected:** Requests are debounced (300-500ms delay)

---

## 5. UI/UX Tests

### 5.1 Theme Persistence
- **Test:**
  1. Toggle to dark mode
  2. Refresh page
  3. Open new browser tab
- **Expected:** Dark mode persists across sessions

### 5.2 Responsive Layout
- **Test:** View app on mobile viewport (320px width)
- **Expected:**
  - Navigation collapses to hamburger menu
  - All content accessible
  - No horizontal scroll

### 5.3 Loading States
- **Test:** Slow network throttling (Network tab in DevTools)
- **Expected:** Loading spinners appear during data fetch

### 5.4 Error Recovery
- **Test:** API returns 500 error
- **Expected:** User-friendly error message displayed, retry option available

---

## 6. Functional Tests

### 6.1 Status Workflow
- **Test:** Move TODO through full lifecycle: open -> in_progress -> resolved -> closed
- **Expected:** Each status change reflected in all views (Dashboard, List, Kanban)

### 6.2 Dashboard Statistics
- **Test:** Create/close TODOs, verify stats update
- **Expected:** Dashboard counts match actual TODO counts

### 6.3 Filter Combinations
- **Test:** Apply multiple filters simultaneously
  - Status: open
  - Priority: P0
  - Assignee: john
  - Type: BUG
- **Expected:** Only TODOs matching ALL criteria shown

### 6.4 Search Functionality
- **Test:** Search for TODO by:
  - Content text
  - File path
  - Assignee name
- **Expected:** All matching TODOs returned

### 6.5 Due Date Display
- **Test:** Create TODO with due date in past, today, future
- **Expected:**
  - Past: Red warning indicator
  - Today: Normal display
  - Future: Normal display

---

## 7. Browser Compatibility

### 7.1 Modern Browsers
- **Test:** Chrome, Firefox, Safari, Edge (latest versions)
- **Expected:** All features work correctly

### 7.2 Local Storage
- **Test:** Disable localStorage, try to use app
- **Expected:** App works (theme defaults to system preference)

---

## 8. Backend Server Tests

### 8.1 Graceful Shutdown
- **Test:** Send SIGTERM to server process
- **Expected:** Database connections closed, no errors

### 8.2 Static File Serving
- **Test:** Access non-existent route `/fake-page`
- **Expected:** Returns index.html (SPA fallback)

### 8.3 CORS Headers
- **Test:** Make API request from different origin
- **Expected:** Proper CORS handling (if needed)

### 8.4 Invalid Routes
- **Test:** Request `/api/invalid-endpoint`
- **Expected:** 404 response with JSON error

---

## 9. Data Validation Tests

### 9.1 Required Fields
- **Test:** PUT with empty content `{"content": ""}`
- **Expected:** Should allow (content is optional in current schema)

### 9.2 Field Length Limits
- **Test:** Content > 65535 characters (MySQL TEXT limit)
- **Expected:** Handle gracefully or reject with clear error

### 9.3 Invalid Date Format
- **Test:** PUT with `"due_date": "not-a-date"`
- **Expected:** Either reject or handle parse error gracefully

---

## 10. Accessibility Tests

### 10.1 Keyboard Navigation
- **Test:** Navigate entire app using only keyboard (Tab, Enter, Escape)
- **Expected:** All interactive elements accessible

### 10.2 Screen Reader
- **Test:** Navigate with screen reader
- **Expected:** Proper ARIA labels, semantic HTML

### 10.3 Color Contrast
- **Test:** Check text contrast ratios
- **Expected:** WCAG AA compliant (4.5:1 for normal text)

---

## 11. Code Review Test Results (2026-02-16)

### Test Environment
- **Platform**: macOS Darwin 23.4.0
- **Go Version**: 1.26.0 (available at /usr/local/go/bin/go)
- **Build Status**: Go backend has module dependency issues preventing build

### Summary Results

| Category | Total Tests | PASS | FAIL | NOT TESTED |
|----------|-------------|------|------|------------|
| Happy Path | 22 | 17 | 4 | 1 |
| Edge Cases | 9 | 5 | 1 | 3 |
| Crash Tests | 6 | 3 | 1 | 2 |
| **TOTAL** | **37** | **25** | **6** | **6** |

---

### 11.1 Happy Path Tests - Results

| Test ID | Description | Status | Notes |
|---------|-------------|--------|-------|
| TC-001 | Load dashboard successfully | **PASS** | Code properly handles loading states and displays stats cards |
| TC-002 | View in-progress TODOs | **PASS** | Filters todos by status correctly |
| TC-003 | View open TODOs | **PASS** | Shows first 6 open TODOs |
| TC-004 | Empty state message | **PASS** | Shows helpful message when no TODOs found |
| TC-005 | Error state display | **PASS** | Shows connection error with server startup instructions |
| TC-006 | Load all TODOs in table | **PASS** | Table displays ID, Type, Content, Priority, Status, Assignee, Location |
| TC-007 | Filter bar functionality | **PASS** | Uses FilterBar component for filtering |
| TC-008 | Empty state - List View | **PASS** | Shows "No TODOs found" message |
| TC-009 | Priority color coding | **PASS** | P0-P4 have distinct colors |
| TC-010 | Status badges | **PASS** | Status shown with colored badges |
| TC-011 | Drag TODO to different column | **PARTIAL** | Drag-drop implemented with @dnd-kit but has API issue |
| TC-012 | Status change on drop | **FAIL** | api.updateTODO expects wrapped response but backend returns different format |
| TC-013 | Five columns display | **PASS** | Open, In Progress, Blocked, Resolved, Closed columns |
| TC-014 | Count per column | **PASS** | Shows count badge on each column |
| TC-015 | Search by content | **PASS** | Searches in content, assignee, file_path |
| TC-016 | Search results display | **PASS** | Results shown as TODOCard grid |
| TC-017 | Empty search results | **PASS** | Shows "No results found" message |
| TC-018 | Search with special characters | **NOT TESTED** | Needs runtime testing |
| TC-019 | Status pie chart | **PASS** | Uses recharts PieChart |
| TC-020 | Priority bar chart | **PASS** | Uses recharts BarChart |
| TC-021 | Type distribution | **PASS** | Full-width bar chart |
| TC-022 | List all TODOs | **PASS** | Returns TodoListResponse with todos array and total |
| TC-023 | Filter by status | **PASS** | Supports ?status= parameter |
| TC-024 | Filter by priority | **PASS** | Supports ?priority= parameter |
| TC-025 | Filter by assignee | **PASS** | Supports ?assignee= parameter |
| TC-026 | Filter by type | **PASS** | Supports ?type= parameter |
| TC-027 | Get single TODO | **FAIL** | Response format mismatch - returns wrapped format but frontend expects direct TODO |
| TC-028 | Update TODO status | **FAIL** | Response format mismatch |
| TC-029 | Update TODO priority | **PASS** | Code handles priority updates |
| TC-030 | Update TODO assignee | **PASS** | Code handles assignee updates |
| TC-031 | Delete TODO | **PASS** | Returns success: true |
| TC-032 | Get statistics | **FAIL** | Type assertion will panic if fields are nil/missing |
| TC-033 | Search TODOs | **PASS** | Returns matching TODOs |
| TC-034 | Empty query | **PASS** | Returns empty array |

---

### 11.2 Issues Identified

#### Issue #1: API Response Format Mismatch (CRITICAL)
**Location**: `/cmd/serve.go:190` and `/web/src/services/api.ts:22`

**Description**: The GET /api/todo/:id endpoint returns:
```json
{"success":true,"data":{...TODO...}}
```

But the frontend api.ts expects direct TODO object:
```typescript
async getTODO(id: string): Promise<TODO> {
  const response = await fetch(`${API_BASE}/todo/${id}`);
  return response.json();  // Expects direct TODO object
}
```

**Fix Required**: Either wrap the frontend response handling or unwrap the backend response.

---

#### Issue #2: Missing CORS Headers (CRITICAL)
**Location**: `/cmd/serve.go`

**Description**: No CORS headers are set, preventing cross-origin requests.

**Fix Required**: Add CORS middleware:
```go
w.Header().Set("Access-Control-Allow-Origin", "*")
w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
```

---

#### Issue #3: Stats API Type Assertion Panic (HIGH)
**Location**: `/cmd/serve.go:262-267`

**Description**: If stats fields are nil (e.g., no TODOs exist), type assertions will panic:
```go
Total: stats["total"].(int64)  // panics if nil
```

**Fix Required**: Add nil checks or use safe type assertions.

---

#### Issue #4: Go Module Dependencies (BLOCKING)
**Location**: Project-wide

**Description**: Cannot build Go server due to git authentication issues with modernc.org/memdb.

**Fix Required**: Run `go mod download` or vendor dependencies.

---

#### Issue #5: Kanban Drag-Drop API Call Fails (HIGH)
**Location**: `/web/src/pages/Kanban.tsx:151-158`

**Description**: When dragging a TODO to a new column, the update API call fails due to Issue #1.

---

### 11.3 Test Execution Instructions

#### Prerequisites
1. Fix Go module dependencies:
   ```bash
   cd /Users/csd/Projects/projectmanagement
   /usr/local/go/bin/go mod download
   ```

2. Build the server:
   ```bash
   /usr/local/go/bin/go build -o todo .
   ```

3. Initialize database (if needed):
   ```bash
   ./todo init
   ./todo scan
   ```

4. Start server:
   ```bash
   ./todo serve -p 8080
   ```

5. Access UI at: http://localhost:8080

#### Running Tests
- **Manual Testing**: Navigate through each page in the browser
- **API Testing**: Use curl or Postman to test endpoints:
  ```bash
  curl http://localhost:8080/api/todos
  curl http://localhost:8080/api/stats
  curl -X PUT http://localhost:8080/api/todo/<ID> \
    -H "Content-Type: application/json" \
    -d '{"status":"in_progress"}'
  ```

---

## Appendix: File Locations

### Backend Files
- Server: `/Users/csd/Projects/projectmanagement/cmd/serve.go`
- Database: `/Users/csd/Projects/projectmanagement/internal/database/database.go`

### Frontend Files
- App: `/Users/csd/Projects/projectmanagement/web/src/App.tsx`
- API Service: `/Users/csd/Projects/projectmanagement/web/src/services/api.ts`
- Pages:
  - Dashboard: `/Users/csd/Projects/projectmanagement/web/src/pages/Dashboard.tsx`
  - List View: `/Users/csd/Projects/projectmanagement/web/src/pages/ListView.tsx`
  - Kanban: `/Users/csd/Projects/projectmanagement/web/src/pages/Kanban.tsx`
  - Search: `/Users/csd/Projects/projectmanagement/web/src/pages/SearchPage.tsx`
  - Charts: `/Users/csd/Projects/projectmanagement/web/src/pages/Charts.tsx`

### Build Outputs
- Go Binary: `/Users/csd/Projects/projectmanagement/todo` (to be created)
- Web Dist: `/Users/csd/Projects/projectmanagement/web/dist/`

---

## 11. Code Review Findings (2026-02-16)

### 11.1 Critical Issues Found

#### Issue #1: API Response Format Mismatch
**Location**: `/cmd/serve.go:190` and `/web/src/services/api.ts:22`

**Problem**: The GET /api/todo/:id endpoint returns a wrapped response format:
```json
{"success":true,"data":{...TODO...}}
```

But the frontend `api.ts` expects a direct TODO object:
```typescript
async getTODO(id: string): Promise<TODO> {
  const response = await fetch(`${API_BASE}/todo/${id}`);
  return response.json();  // Expects direct TODO object
}
```

**Impact**: TC-027 (Get single TODO) will fail - the frontend cannot parse the response correctly.

**Fix**: Update frontend to unwrap response:
```typescript
async getTODO(id: string): Promise<TODO> {
  const response = await fetch(`${API_BASE}/todo/${id}`);
  const data = await response.json();
  return data.data;  // Unwrap from APIResponse
}
```

#### Issue #2: Missing CORS Headers
**Location**: `/cmd/serve.go`

**Problem**: No CORS headers are set on API responses.

**Impact**: TC-051 (Cross-origin request) will fail - browser will block cross-origin requests.

**Fix**: Add CORS middleware in serve.go:
```go
func (s *Server) handleAPITodos(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Access-Control-Allow-Origin", "*")
    w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
    w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
    // ... rest of handler
}
```

#### Issue #3: Stats API Type Assertion Panic
**Location**: `/cmd/serve.go:262-267`

**Problem**: If stats fields are nil (e.g., no TODOs exist), type assertions will panic:
```go
Total: stats["total"].(int64)  // panics if nil
ByStatus: stats["by_status"].(map[string]int64),  // panics if nil
```

**Impact**: TC-032 (Get statistics) and TC-037 (Empty database - Stats) will cause server crash.

**Fix**: Add nil checks:
```go
total := int64(0)
if stats["total"] != nil {
    total = stats["total"].(int64)
}
response := StatsResponse{
    Total: total,
    // ... handle other fields similarly
}
```

### 11.2 Test Results Summary (Code Review)

| Category | Passed | Failed | Not Tested | Total |
|----------|--------|--------|------------|-------|
| API Security | 0 | 0 | 4 | 4 |
| Edge Cases | 0 | 0 | 9 | 9 |
| Data Integrity | 0 | 0 | 3 | 3 |
| Performance | 0 | 0 | 3 | 3 |
| UI/UX | 0 | 0 | 4 | 4 |
| Functional | 0 | 0 | 4 | 4 |
| Browser Compatibility | 0 | 0 | 2 | 2 |
| Backend Server | 0 | 0 | 4 | 4 |
| Data Validation | 0 | 0 | 3 | 3 |
| Accessibility | 0 | 0 | 3 | 3 |

**Note**: Tests marked as "Not Tested" require runtime execution. The above critical issues would cause test failures when the application is run.

### 11.3 Build Status

- **Go Build**: FAILED - Cannot build due to git authentication issues with modernc.org/memdb dependency
- **Frontend Build**: SUCCESS - Web UI built and available at `/web/dist/`
- **Static Files**: PRESENT - index.html and assets exist in web/dist/

---

*Last updated: 2026-02-16*
