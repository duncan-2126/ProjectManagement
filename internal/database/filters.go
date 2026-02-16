package database

import (
	"time"
)

// SavedFilter represents a saved filter query
type SavedFilter struct {
	ID        string    `gorm:"primaryKey;type:text" json:"id"`
	Name      string    `gorm:"type:text;uniqueIndex" json:"name"`
	Query     string    `gorm:"type:text" json:"query"` // JSON encoded query
	CreatedAt time.Time `gorm:"not null" json:"created_at"`
}

// CreateSavedFilter creates a new saved filter
func (db *DB) CreateSavedFilter(name, query string) (*SavedFilter, error) {
	filter := &SavedFilter{
		ID:        uuid.New().String(),
		Name:      name,
		Query:     query,
		CreatedAt: time.Now(),
	}
	if err := db.Create(filter).Error; err != nil {
		return nil, err
	}
	return filter, nil
}

// GetSavedFilter returns a saved filter by name
func (db *DB) GetSavedFilter(name string) (*SavedFilter, error) {
	var filter SavedFilter
	if err := db.First(&filter, "name = ?", name).Error; err != nil {
		return nil, err
	}
	return &filter, nil
}

// GetSavedFilters returns all saved filters
func (db *DB) GetSavedFilters() ([]SavedFilter, error) {
	var filters []SavedFilter
	if err := db.Order("name").Find(&filters).Error; err != nil {
		return nil, err
	}
	return filters, nil
}

// DeleteSavedFilter deletes a saved filter
func (db *DB) DeleteSavedFilter(name string) error {
	return db.Where("name = ?", name).Delete(&SavedFilter{}).Error
}
