package database

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// TODO represents a TODO comment entry
type TODO struct {
	ID         string     `gorm:"primaryKey;type:text" json:"id"`
	FilePath   string     `gorm:"type:text;not null" json:"file_path"`
	LineNumber int        `gorm:"not null" json:"line_number"`
	Column     int        `gorm:"default:0" json:"column"`
	Type       string     `gorm:"type:text;not null" json:"type"` // TODO, FIXME, HACK, BUG, NOTE, XXX
	Content    string     `gorm:"type:text;not null" json:"content"`
	Author     string     `gorm:"type:text" json:"author"`
	Email      string     `gorm:"type:text" json:"email"`
	CreatedAt  time.Time  `gorm:"not null" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"not null" json:"updated_at"`
	Status     string     `gorm:"type:text;default:'open'" json:"status"` // open, in_progress, blocked, resolved, wontfix, closed
	Priority   string     `gorm:"type:text;default:'P3'" json:"priority"` // P0, P1, P2, P3, P4
	Category   string     `gorm:"type:text" json:"category"`
	Assignee   string     `gorm:"type:text" json:"assignee"`
	DueDate    *time.Time `gorm:"type:timestamp" json:"due_date,omitempty"`
	Estimate   *int       `gorm:"type:integer" json:"estimate,omitempty"` // minutes
	Hash       string     `gorm:"type:text;not null" json:"hash"`
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
	ID          string     `gorm:"primaryKey;type:text" json:"id"`
	Name        string     `gorm:"type:text;not null" json:"name"`
	Path        string     `gorm:"type:text;uniqueIndex" json:"path"`
	CreatedAt   time.Time  `gorm:"not null" json:"created_at"`
	LastScanned *time.Time `gorm:"type:timestamp" json:"last_scanned,omitempty"`
}

// Relationship represents a relationship between TODOs
type Relationship struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id"`
	SourceID  string    `gorm:"type:text;not null;index" json:"source_id"`
	TargetID  string    `gorm:"type:text;not null;index" json:"target_id"`
	Type      string    `gorm:"type:text;not null" json:"type"` // parent, child, depends_on, blocked_by, relates_to
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

// Watch represents a user watching a TODO
type Watch struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id"`
	TODOID    string    `gorm:"type:text;not null;index" json:"todo_id"`
	UserID    string    `gorm:"type:text;not null" json:"user_id"`
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
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
	gormDB, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Auto migrate
	if err := gormDB.AutoMigrate(&TODO{}, &Tag{}, &TODOTag{}, &Project{}, &Relationship{}, &Watch{}); err != nil {
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
	if assignee, ok := filters["assignee"].(string); ok && assignee != "" {
		query = query.Where("assignee = ?", assignee)
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

// CreateRelationship creates a new relationship between TODOs
func (db *DB) CreateRelationship(sourceID, targetID, relType string) error {
	rel := Relationship{
		ID:        uuid.New().String(),
		SourceID:  sourceID,
		TargetID:  targetID,
		Type:      relType,
		CreatedAt: time.Now(),
	}
	return db.Create(&rel).Error
}

// GetRelationships returns all relationships for a TODO
func (db *DB) GetRelationships(todoID string) ([]Relationship, error) {
	var relationships []Relationship
	err := db.Where("source_id = ?", todoID).Or("target_id = ?", todoID).Find(&relationships).Error
	return relationships, err
}

// GetRelationshipsByType returns relationships of a specific type for a TODO
func (db *DB) GetRelationshipsByType(todoID, relType string) ([]Relationship, error) {
	var relationships []Relationship
	err := db.Where("source_id = ? AND type = ?", todoID, relType).Find(&relationships).Error
	return relationships, err
}

// DeleteRelationship deletes a relationship
func (db *DB) DeleteRelationship(id string) error {
	return db.Delete(&Relationship{}, "id = ?", id).Error
}

// DeleteRelationshipsForTODO deletes all relationships for a TODO
func (db *DB) DeleteRelationshipsForTODO(todoID string) error {
	return db.Where("source_id = ? OR target_id = ?", todoID, todoID).Delete(&Relationship{}).Error
}

// HasCircularDependency checks if adding a dependency would create a cycle
func (db *DB) HasCircularDependency(sourceID, targetID string) bool {
	visited := make(map[string]bool)
	return db.checkCycle(sourceID, targetID, visited)
}

func (db *DB) checkCycle(sourceID, targetID string, visited map[string]bool) bool {
	if sourceID == targetID {
		return true
	}
	if visited[targetID] {
		return false
	}
	visited[targetID] = true

	// Find all tasks that the target depends on
	var deps []Relationship
	db.Where("source_id = ? AND type = ?", targetID, "depends_on").Find(&deps)

	for _, dep := range deps {
		if db.checkCycle(sourceID, dep.TargetID, visited) {
			return true
		}
	}
	return false
}

// ValidateRelationships validates all relationships for issues
func (db *DB) ValidateRelationships() (map[string][]string, error) {
	issues := make(map[string][]string)

	// Check for broken links (target TODO doesn't exist)
	var relationships []Relationship
	db.Find(&relationships)

	for _, rel := range relationships {
		var target TODO
		if err := db.First(&target, "id = ?", rel.TargetID).Error; err != nil {
			issues["broken_links"] = append(issues["broken_links"],
				fmt.Sprintf("Relationship %s: target TODO %s not found", rel.ID, rel.TargetID))
		}

		var source TODO
		if err := db.First(&source, "id = ?", rel.SourceID).Error; err != nil {
			issues["broken_links"] = append(issues["broken_links"],
				fmt.Sprintf("Relationship %s: source TODO %s not found", rel.ID, rel.SourceID))
		}
	}

	// Check for circular dependencies
	for _, rel := range relationships {
		if rel.Type == "depends_on" {
			if db.HasCircularDependency(rel.SourceID, rel.TargetID) {
				issues["circular_dependencies"] = append(issues["circular_dependencies"],
					fmt.Sprintf("Circular dependency: %s -> %s", rel.SourceID, rel.TargetID))
			}
		}
	}

	return issues, nil
}

// GetDependents returns TODOs that this TODO depends on
func (db *DB) GetDependents(todoID string) ([]TODO, error) {
	var relationships []Relationship
	if err := db.Where("source_id = ? AND type = ?", todoID, "depends_on").Find(&relationships).Error; err != nil {
		return nil, err
	}

	var todos []TODO
	for _, rel := range relationships {
		var todo TODO
		if err := db.First(&todo, "id = ?", rel.TargetID).Error; err == nil {
			todos = append(todos, todo)
		}
	}
	return todos, nil
}

// GetBlockers returns TODOs that block this TODO
func (db *DB) GetBlockers(todoID string) ([]TODO, error) {
	var relationships []Relationship
	if err := db.Where("source_id = ? AND type = ?", todoID, "blocked_by").Find(&relationships).Error; err != nil {
		return nil, err
	}

	var todos []TODO
	for _, rel := range relationships {
		var todo TODO
		if err := db.First(&todo, "id = ?", rel.TargetID).Error; err == nil {
			todos = append(todos, todo)
		}
	}
	return todos, nil
}

// GetChildren returns subtasks of a TODO
func (db *DB) GetChildren(todoID string) ([]TODO, error) {
	var relationships []Relationship
	if err := db.Where("source_id = ? AND type = ?", todoID, "parent").Find(&relationships).Error; err != nil {
		return nil, err
	}

	var todos []TODO
	for _, rel := range relationships {
		var todo TODO
		if err := db.First(&todo, "id = ?", rel.TargetID).Error; err == nil {
			todos = append(todos, todo)
		}
	}
	return todos, nil
}

// GetParent returns the parent of a TODO
func (db *DB) GetParent(todoID string) (*TODO, error) {
	var relationships []Relationship
	if err := db.Where("source_id = ? AND type = ?", todoID, "child").Find(&relationships).Error; err != nil {
		return nil, err
	}

	if len(relationships) == 0 {
		return nil, nil
	}

	var todo TODO
	if err := db.First(&todo, "id = ?", relationships[0].TargetID).Error; err != nil {
		return nil, err
	}
	return &todo, nil
}


// GetRelatedTODOs returns TODOs related to a given TODO
func (db *DB) GetRelatedTODOs(todoID string) ([]Relationship, error) {
	var relationships []Relationship
	if err := db.Where("(source_id = ? OR target_id = ?)", todoID, todoID).Find(&relationships).Error; err != nil {
		return nil, err
	}
	return relationships, nil
}

// CreateWatch creates a new watch for a TODO
func (db *DB) CreateWatch(todoID, userID string) error {
	watch := Watch{
		ID:        uuid.New().String(),
		TODOID:    todoID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}
	return db.Create(&watch).Error
}

// DeleteWatch removes a watch for a TODO
func (db *DB) DeleteWatch(todoID, userID string) error {
	return db.Where("todo_id = ? AND user_id = ?", todoID, userID).Delete(&Watch{}).Error
}

// GetWatchesByUser returns all watches for a user
func (db *DB) GetWatchesByUser(userID string) ([]Watch, error) {
	var watches []Watch
	err := db.Where("user_id = ?", userID).Find(&watches).Error
	return watches, err
}

// GetWatchedTODOs returns all TODOs being watched by a user
func (db *DB) GetWatchedTODOs(userID string) ([]TODO, error) {
	var watches []Watch
	if err := db.Where("user_id = ?", userID).Find(&watches).Error; err != nil {
		return nil, err
	}

	var todos []TODO
	for _, w := range watches {
		var todo TODO
		if err := db.First(&todo, "id = ?", w.TODOID).Error; err == nil {
			todos = append(todos, todo)
		}
	}
	return todos, nil
}

// IsWatching checks if a user is watching a TODO
func (db *DB) IsWatching(todoID, userID string) bool {
	var count int64
	db.Model(&Watch{}).Where("todo_id = ? AND user_id = ?", todoID, userID).Count(&count)
	return count > 0
}

// GetStaleTODOs returns TODOs that haven't been updated in the specified days
func (db *DB) GetStaleTODOs(daysSinceUpdate int) ([]TODO, error) {
	var todos []TODO
	threshold := time.Now().AddDate(0, 0, -daysSinceUpdate)
	err := db.Where("updated_at < ? AND status NOT IN (?, ?)", threshold, "closed", "resolved").Order("updated_at ASC").Find(&todos).Error
	return todos, err
}

// GetTODOsDueSoon returns TODOs due within the specified days
func (db *DB) GetTODOsDueSoon(days int) ([]TODO, error) {
	var todos []TODO
	threshold := time.Now().AddDate(0, 0, days)
	err := db.Where("due_date IS NOT NULL AND due_date <= ? AND status NOT IN (?, ?)", threshold, "closed", "resolved").Order("due_date ASC").Find(&todos).Error
	return todos, err
}

// GetOverdueTODOs returns TODOs that are past their due date
func (db *DB) GetOverdueTODOs() ([]TODO, error) {
	var todos []TODO
	now := time.Now()
	err := db.Where("due_date IS NOT NULL AND due_date < ? AND status NOT IN (?, ?)", now, "closed", "resolved").Order("due_date ASC").Find(&todos).Error
	return todos, err
}

// GetTags returns all tags
func (db *DB) GetTags() ([]Tag, error) {
	var tags []Tag
	if err := db.Order("name").Find(&tags).Error; err != nil {
		return nil, err
	}
	return tags, nil
}

// GetTagsForTODO returns all tags for a specific TODO
func (db *DB) GetTagsForTODO(todoID string) ([]Tag, error) {
	var tags []Tag
	err := db.Table("tags").
		Joins("JOIN todo_tags ON todo_tags.tag_id = tags.id").
		Where("todo_tags.todo_id = ?", todoID).
		Find(&tags).Error
	return tags, err
}

// AddTagToTODO adds a tag to a TODO
func (db *DB) AddTagToTODO(todoID, tagID string) error {
	todoTag := TODOTag{
		TODOID: todoID,
		TagID:  tagID,
	}
	return db.Create(&todoTag).Error
}

// GetOrCreateTag gets a tag by name or creates it if it doesn't exist
func (db *DB) GetOrCreateTag(name string) (*Tag, error) {
	var tag Tag
	err := db.First(&tag, "name = ?", name).Error
	if err == gorm.ErrRecordNotFound {
		tag = Tag{
			ID:   uuid.New().String(),
			Name: name,
		}
		err = db.Create(&tag).Error
	}
	return &tag, err
}
