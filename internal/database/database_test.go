package database

import (
	"testing"
	"os"
	"path/filepath"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDB(t *testing.T) {
	// Create temp directory
	tmpDir, err := os.MkdirTemp("", "todo-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	// Create test database
	db, err := New(tmpDir)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Test CreateTODO
	todo := TODO{
		FilePath:   "test.go",
		LineNumber: 10,
		Column:     5,
		Type:       "TODO",
		Content:    "Test TODO",
		Status:     "open",
		Priority:   "P3",
		Hash:       "test-hash-123",
	}

	err = db.CreateTODO(&todo)
	require.NoError(t, err)
	assert.NotEmpty(t, todo.ID)

	// Test GetTODOs
	todos, err := db.GetTODOs(nil)
	require.NoError(t, err)
	assert.Len(t, todos, 1)

	// Test GetTODOByID
	found, err := db.GetTODOByID(todo.ID)
	require.NoError(t, err)
	assert.Equal(t, "test.go", found.FilePath)

	// Test UpdateTODO
	todo.Status = "resolved"
	err = db.UpdateTODO(&todo)
	require.NoError(t, err)

	updated, err := db.GetTODOByID(todo.ID)
	require.NoError(t, err)
	assert.Equal(t, "resolved", updated.Status)

	// Test GetTODOs with filters
	filtered, err := db.GetTODOs(map[string]interface{}{"status": "resolved"})
	require.NoError(t, err)
	assert.Len(t, filtered, 1)

	filtered, err = db.GetTODOs(map[string]interface{}{"status": "open"})
	require.NoError(t, err)
	assert.Len(t, filtered, 0)

	// Test DeleteTODO
	err = db.DeleteTODO(todo.ID)
	require.NoError(t, err)

	todos, err = db.GetTODOs(nil)
	require.NoError(t, err)
	assert.Len(t, todos, 0)

	// Test GetStats
	todo = TODO{
		FilePath:   "test2.go",
		LineNumber: 20,
		Type:       "FIXME",
		Content:    "Fix this",
		Status:     "open",
		Priority:   "P1",
		Hash:       "test-hash-456",
	}
	db.CreateTODO(&todo)

	stats, err := db.GetStats()
	require.NoError(t, err)
	assert.Equal(t, int64(1), stats["total"].(int64))

	// Test TODOExists
	exists, err := db.TODOExists("test-hash-456", "test2.go", 20)
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = db.TODOExists("nonexistent", "test.go", 10)
	require.NoError(t, err)
	assert.False(t, exists)

	// Test InitProject
	err = db.InitProject("test-project", tmpDir)
	require.NoError(t, err)
}

func TestGetTODOsByType(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "todo-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	db, err := New(tmpDir)
	require.NoError(t, err)

	// Create TODOs with different types
	types := []string{"TODO", "FIXME", "HACK", "BUG", "NOTE", "XXX"}
	for i, todoType := range types {
		todo := TODO{
			FilePath:   "test.go",
			LineNumber: i + 1,
			Type:       todoType,
			Content:    "Test " + todoType,
			Hash:       "hash-" + todoType,
		}
		db.CreateTODO(&todo)
	}

	// Filter by type
	for _, todoType := range types {
		todos, err := db.GetTODOs(map[string]interface{}{"type": todoType})
		require.NoError(t, err)
		assert.Len(t, todos, 1)
		assert.Equal(t, todoType, todos[0].Type)
	}
}

func TestGetTODOsByPriority(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "todo-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	db, err := New(tmpDir)
	require.NoError(t, err)

	// Create TODOs with different priorities
	priorities := []string{"P0", "P1", "P2", "P3", "P4"}
	for i, priority := range priorities {
		todo := TODO{
			FilePath:   "test.go",
			LineNumber: i + 1,
			Type:       "TODO",
			Priority:   priority,
			Content:    "Test " + priority,
			Hash:       "hash-" + priority,
		}
		db.CreateTODO(&todo)
	}

	// Filter by priority
	for _, priority := range priorities {
		todos, err := db.GetTODOs(map[string]interface{}{"priority": priority})
		require.NoError(t, err)
		assert.Len(t, todos, 1)
		assert.Equal(t, priority, todos[0].Priority)
	}
}
