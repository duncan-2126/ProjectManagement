# Code Review Report: React Web UI & Go Backend

## Executive Summary

This report covers the code review of the React Web UI components (`/Users/csd/Projects/projectmanagement/web/src/`) and the Go backend server (`/Users/csd/Projects/projectmanagement/cmd/serve.go`).

---

## Critical Issues (MUST FIX)

### 1. XSS Vulnerability - Incorrect String Replacement

**Files Affected:**
- `/Users/csd/Projects/projectmanagement/web/src/pages/Kanban.tsx` (line 48)
- `/Users/csd/Projects/projectmanagement/web/src/pages/ListView.tsx` (line 88)
- `/Users/csd/Projects/projectmanagement/web/src/components/TODOCard.tsx` (line 86)

**Issue:** The code uses `.replace('_', ' ')` which only replaces the FIRST underscore. This is a potential XSS vector if the status contains malicious content.

```typescript
// Current (WRONG):
{todo.status.replace('_', ' ')}

// Should be:
{todo.status.replace(/_/g, ' ')}
```

---

### 2. No Input Validation in Update API

**File:** `/Users/csd/Projects/projectmanagement/cmd/serve.go` (lines 193-236)

**Issue:** The `handleUpdateTodo` function accepts any values without validation:
- No validation that status is one of the allowed values (`open`, `in_progress`, `blocked`, `resolved`, `wontfix`, `closed`)
- No validation that priority is valid (`P0`-`P4`)
- No length limits on content, assignee, category
- No sanitization of input

**Example vulnerability:**
```json
PUT /api/todo/123
{"status": "<script>alert('xss')</script>"}
```

---

### 3. Broken Due Date Parsing

**File:** `/Users/csd/Projects/projectmanagement/cmd/serve.go` (lines 223-227)

**Issue:** The due date is never actually parsed - it always sets `DueDate = nil`:

```go
if dueDate, ok := updates["due_date"].(string); ok {
    if dueDate != "" {
        todo.DueDate = nil // BUG: Never parses the date!
    }
}
```

---

### 4. Race Condition in Kanban Drag-and-Drop

**File:** `/Users/csd/Projects/projectmanagement/web/src/pages/Kanban.tsx` (lines 151-158)

**Issue:** When dragging a TODO card to change status:
1. UI updates optimistically via `setTodos()`
2. API call is made via `api.updateTODO()`
3. If API fails, the UI does NOT revert - leaving inconsistent state

```typescript
const handleStatusChange = async (id: string, status: TODOStatus) => {
  try {
    await api.updateTODO(id, { status });
    setTodos(prev => prev.map(t => t.id === id ? { ...t, status } : t));
  } catch (err) {
    console.error('Failed to update status:', err);
    // BUG: No rollback of UI state!
  }
};
```

---

### 5. No Error Display in Most Pages

**Files:**
- `/Users/csd/Projects/projectmanagement/web/src/pages/ListView.tsx`
- `/Users/csd/Projects/projectmanagement/web/src/pages/Kanban.tsx`
- `/Users/csd/Projects/projectmanagement/web/src/pages/SearchPage.tsx`
- `/Users/csd/Projects/projectmanagement/web/src/pages/Charts.tsx`

**Issue:** These pages only log errors to console. Users see no feedback when API calls fail.

---

### 6. No Pagination - Performance Issue

**File:** `/Users/csd/Projects/projectmanagement/cmd/serve.go`

**Issue:** `handleAPITodos` returns ALL TODOs without pagination. With large codebases (1000+ TODOs), this causes:
- Slow initial load
- Memory issues in browser
- Network overhead

---

## Warnings (SHOULD FIX)

### 7. Missing Error State in API Service

**File:** `/Users/csd/Projects/projectmanagement/web/src/services/api.ts`

**Issue:** API errors don't include server response details:
```typescript
if (!response.ok) throw new Error('Failed to fetch TODOs');
// Should include: response.json()?.error or status text
```

---

### 8. No Loading State During Filter Changes

**Files:**
- `/Users/csd/Projects/projectmanagement/web/src/pages/ListView.tsx` (line 73)
- `/Users/csd/Projects/projectmanagement/web/src/components/FilterBar.tsx`

**Issue:** After initial load, changing filters doesn't show loading indicator.

---

### 9. No Request Timeout

**File:** `/Users/csd/Projects/projectmanagement/web/src/services/api.ts`

**Issue:** All fetch requests have no timeout, potentially hanging indefinitely.

---

### 10. Type Assumptions in Stats Response

**File:** `/Users/csd/Projects/projectmanagement/cmd/serve.go` (lines 262-267)

**Issue:** Type assertions without checks can panic:
```go
Total: stats["total"].(int64),  // Will panic if nil or wrong type
```

---

### 11. No Debounce on Search/Filter Inputs

**Files:**
- `/Users/csd/Projects/projectmanagement/web/src/components/FilterBar.tsx`
- `/Users/csd/Projects/projectmanagement/web/src/pages/SearchPage.tsx`

**Issue:** Every keystroke triggers an API call. Should debounce by 300-500ms.

---

### 12. No React Error Boundary

**File:** `/Users/csd/Projects/projectmanagement/web/src/App.tsx`

**Issue:** No error boundary to catch rendering errors and show graceful fallback.

---

### 13. Delete Without Confirmation

**File:** `/Users/csd/Projects/projectmanagement/cmd/serve.go` (line 238-245)

**Issue:** DELETE request has no confirmation dialog in UI (delete functionality appears to exist in API but not exposed in current React components).

---

### 14. Potential Memory Leak in Kanban

**File:** `/Users/csd/Projects/projectmanagement/web/src/pages/Kanban.tsx`

**Issue:** The `activeId` state is not cleaned up if component unmounts during drag.

---

## Suggestions (NICE TO HAVE)

### 15. Empty States Could Be More Helpful

**Files:** Multiple page components

**Suggestion:** Add actionable empty states with instructions like "Run `todo scan` to find TODOs".

---

### 16. Missing Keyboard Navigation in Kanban

**File:** `/Users/csd/Projects/projectmanagement/web/src/pages/Kanban.tsx`

**Suggestion:** Add keyboard support for drag-and-drop using `@dnd-kit/core` accessibility features.

---

### 17. No Optimistic Updates for Other Operations

**Files:** Multiple components

**Suggestion:** Apply optimistic UI updates pattern consistently across all mutation operations.

---

### 18. Server Not Closing Database Connection

**File:** `/Users/csd/Projects/projectmanagement/cmd/serve.go` (line 44-71)

**Issue:** The database connection is never closed when server stops. Should implement graceful shutdown.

---

## Summary

| Category | Count |
|----------|-------|
| Critical Issues | 6 |
| Warnings | 8 |
| Suggestions | 4 |

**Priority Actions:**
1. Fix XSS vulnerability (string replacement)
2. Add input validation to Update API
3. Fix due date parsing
4. Add rollback on failed Kanban updates
5. Add error states to all pages
6. Implement pagination

---

*Code Review completed on 2026-02-16*
