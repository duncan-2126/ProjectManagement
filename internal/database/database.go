package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TODO represents a TODO comment entry
type TODO struct {
	ID          string     `gorm:"primaryKey;type:text" json:"id"`
	FilePath    string     `gorm:"type:text;not null" json:"file_path"`
	LineNumber  int        `gorm:"not null" json:"line_number"`
	Column      int        `gorm:"default:0" json:"column"`
	Type        string     `gorm:"type:text;not null" json:"type"` // TODO, FIXME, HACK, BUG, NOTE, XXX
	Content     string     `gorm:"type:text;not null" json:"content"`
	Author      string     `gorm:"type:text" json:"author"`
	Email       string     `gorm:"type:text" json:"email"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	UpdatedAt   time.Time `gorm:"not null" json:"updated_at"`
	Status      string     `gorm:"type:text;default:'open'" json:"status"` // open, in_progress, blocked, resolved, wontfix, closed
	Priority    string     `gorm:"type:text;default:'P3'" json:"priority"` // P0, P1, P2, P3, P4
	Category    string     `gorm:"type:text" json:"category"`
	Assignee    string     `gorm:"type:text" json:"assignee"`
	DueDate     *time.Time `gorm:"type:timestamp" json:"due_date,omitempty"`
	Estimate    *int       `gorm:"type:integer" json:"estimate,omitempty"` // minutes
	Hash        string     `gorm:"type:text;not null" json:"hash"`
}

// Tag represents a tag for TODOs
type Tag struct {
	ID   string `gorm:"primaryKey;type:text" json:"id"`
	Name string `gorm:"type:text;uniqueIndex" json:"name"`
}

// TODOTag represents the many-to-many relationship between TODOs and Tags
type TODOTag struct {
	TODOID string `gorm:"primaryKey;type:text" json:"todo_id"`
	TagID  string `gorm:"primaryKey;type:text" json:"tag_id"`
}

// Project represents a tracked project
type Project struct {
	ID          string    `gorm:"primaryKey;type:text" json:"id"`
	Name        string    `gorm:"type:text;not null" json:"name"`
	Path        string    `gorm:"type:text;uniqueIndex" json:"path"`
	CreatedAt   time.Time `gorm:"not null" json:"created_at"`
	LastScanned *time.Time `gorm:"type:timestamp" json:"last_scanned,omitempty"`
}

// DB represents the database connection
type DB struct {
	*gorm.DB
}

// New creates a new database connection
func New(projectPath string) (*DB, error) {
	// Create .todo directory in project root
	todoDir := filepath.Join(projectPath, ".todo")
	if err := os.MkdirAll(todoDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create .todo directory: %w", err)
	}

	dbPath := filepath.Join(todoDir, "todos.db")

	// Open SQLite database
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	gormDB, err := gorm.Open(db, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Auto migrate
	if err := gormDB.AutoMigrate(&TODO{}, &Tag{}, &TODOTag{}, &Project{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &DB{DB: gormDB}, nil
}

// CreateTODO creates a new TODO entry
func (db *DB) CreateTODO(t *TODO) error {
	if t.ID == "" {
		t.ID = uuid.New().String()
	}
	return db.Create(t).Error
}

// GetTODOs returns all TODOs with optional filters
func (db *DB) GetTODOs(filters map[string]interface{}) ([]TODO, error) {
	var todos []TODO
	query := db.Model(&TODO{})

	// Apply filters
	if status, ok := filters["status"].(string); ok && status != "" {
		query = query.Where("status = ?", status)
	}
	if todoType, ok := filters["type"].(string); ok && todoType != "" {
		query = query.Where("type = ?", todoType)
	}
	if author, ok := filters["author"].(string); ok && author != "" {
		query = query.Where("author = ?", author)
	}
	if filePath, ok := filters["file_path"].(string); ok && filePath != "" {
		query = query.Where("file_path LIKE ?", "%"+filePath+"%")
	}
	if priority, ok := filters["priority"].(string); ok && priority != "" {
		query = query.Where("priority = ?", priority)
	}

	// Order by file and line
	query = query.Order("file_path, line_number")

	if err := query.Find(&todos).Error; err != nil {
		return nil, err
	}

	return todos, nil
}

// GetTODOByID returns a TODO by ID
func (db *DB) GetTODOByID(id string) (*TODO, error) {
	var todo TODO
	if err := db.First(&todo, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &todo, nil
}

// UpdateTODO updates a TODO entry
func (db *DB) UpdateTODO(t *TODO) error {
	t.UpdatedAt = time.Now()
	return db.Save(t).Error
}

// DeleteTODO deletes a TODO entry
func (db *DB) DeleteTODO(id string) error {
	return db.Delete(&TODO{}, "id = ?", id).Error
}

// TODOExists checks if a TODO already exists by hash and location
func (db *DB) TODOExists(hash, filePath string, lineNumber int) (bool, error) {
	var count int64
	err := db.Model(&TODO{}).
		Where("hash = ? AND file_path = ? AND line_number = ?", hash, filePath, lineNumber).
		Count(&count).Error
	return count > 0, err
}

// GetStats returns statistics about TODOs
func (db *DB) GetStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total count
	var total int64
	db.Model(&TODO{}).Count(&total)
	stats["total"] = total

	// Count by status
	var statusCounts []struct {
		Status string
		Count  int64
	}
	db.Model(&TODO{}).Select("status, COUNT(*) as count").Group("status").Scan(&statusCounts)
	statusMap := make(map[string]int64)
	for _, s := range statusCounts {
		statusMap[s.Status] = s.Count
	}
	stats["by_status"] = statusMap

	// Count by type
	var typeCounts []struct {
		Type  string
		Count int64
	}
	db.Model(&TODO{}).Select("type, COUNT(*) as count").Group("type").Scan(&typeCounts)
	typeMap := make(map[string]int64)
	for _, t := range typeCounts {
		typeMap[t.Type] = t.Count
	}
	stats["by_type"] = typeMap

	// Count by priority
	var priorityCounts []struct {
		Priority string
		Count    int64
	}
	db.Model(&TODO{}).Select("priority, COUNT(*) as count").Group("priority").Scan(&priorityCounts)
	priorityMap := make(map[string]int64)
	for _, p := range priorityCounts {
		priorityMap[p.Priority] = p.Count
	}
	stats["by_priority"] = priorityMap

	return stats, nil
}

// InitProject initializes a project in the database
func (db *DB) InitProject(name, path string) error {
	project := Project{
		ID:        uuid.New().String(),
		Name:      name,
		Path:      path,
		CreatedAt: time.Now(),
	}
	return db.Create(&project).Error
}
