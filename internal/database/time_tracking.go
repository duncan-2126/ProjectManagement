package database

import (
	"time"

	"github.com/google/uuid"
)

// TimeEntry represents time spent on a TODO
type TimeEntry struct {
	ID          string     `gorm:"primaryKey;type:text" json:"id"`
	TODOID      string     `gorm:"type:text;not null;index" json:"todo_id"`
	StartTime   time.Time  `gorm:"not null" json:"start_time"`
	EndTime     *time.Time `gorm:"type:timestamp" json:"end_time,omitempty"`
	Duration    int        `gorm:"type:integer" json:"duration"` // minutes
	Description string     `gorm:"type:text" json:"description"`
	CreatedAt   time.Time  `gorm:"not null" json:"created_at"`
}

// StartTimer starts a timer for a TODO
func (db *DB) StartTimer(todoID, description string) (*TimeEntry, error) {
	entry := &TimeEntry{
		ID:          uuid.New().String(),
		TODOID:      todoID,
		StartTime:   time.Now(),
		Description: description,
		CreatedAt:   time.Now(),
	}
	if err := db.Create(entry).Error; err != nil {
		return nil, err
	}
	return entry, nil
}

// StopTimer stops a running timer
func (db *DB) StopTimer(todoID string) (*TimeEntry, error) {
	var entry TimeEntry
	now := time.Now()

	// Find the running timer for this TODO
	if err := db.First(&entry, "todo_id = ? AND end_time IS NULL", todoID).Error; err != nil {
		return nil, err
	}

	duration := int(now.Sub(entry.StartTime).Minutes())
	entry.EndTime = &now
	entry.Duration = duration

	if err := db.Save(&entry).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

// GetTimeEntries returns all time entries for a TODO
func (db *DB) GetTimeEntries(todoID string) ([]TimeEntry, error) {
	var entries []TimeEntry
	if err := db.Where("todo_id = ?", todoID).Order("start_time DESC").Find(&entries).Error; err != nil {
		return nil, err
	}
	return entries, nil
}

// GetTotalTime returns total time spent on a TODO
func (db *DB) GetTotalTime(todoID string) (int, error) {
	var total int64
	if err := db.Model(&TimeEntry{}).Where("todo_id = ?", todoID).Select("COALESCE(SUM(duration), 0)").Scan(&total).Error; err != nil {
		return 0, err
	}
	return int(total), nil
}

// AddManualTime adds manual time entry
func (db *DB) AddManualTime(todoID string, minutes int, description string) (*TimeEntry, error) {
	now := time.Now()
	startTime := now.Add(-time.Duration(minutes) * time.Minute)

	entry := &TimeEntry{
		ID:          uuid.New().String(),
		TODOID:      todoID,
		StartTime:   startTime,
		EndTime:     &now,
		Duration:    minutes,
		Description: description,
		CreatedAt:   now,
	}
	if err := db.Create(entry).Error; err != nil {
		return nil, err
	}
	return entry, nil
}
