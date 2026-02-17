package store

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
)

// Event represents a branch-scoped note entry
type Event struct {
	ID          string
	RepoID      string
	Branch      string
	Type        string
	Text        string
	CreatedAt   int64
	MetaJSON    string
	CompletedAt *int64 // For todos: completion timestamp
	DeletedAt   *int64 // Soft delete timestamp
	UpdatedAt   *int64 // Last edit timestamp
}

// OpenDB opens or creates the SQLite database at the app data directory
func OpenDB() (*sql.DB, error) {
	appDir, err := AppDataDir()
	if err != nil {
		return nil, err
	}

	if err := EnsureDir(appDir); err != nil {
		return nil, err
	}

	dbPath := filepath.Join(appDir, "branchbrief.sqlite")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Set connection pool size to 1 for write consistency
	db.SetMaxOpenConns(1)

	// Enable WAL mode and foreign keys via PRAGMA
	if err := enablePragmas(db); err != nil {
		db.Close()
		return nil, err
	}

	// Create schema if needed
	if err := createSchema(db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// enablePragmas sets WAL mode and enables foreign keys
// Uses PRAGMA statements for deterministic behavior (not DSN params)
func enablePragmas(db *sql.DB) error {
	_, err := db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return fmt.Errorf("failed to set WAL mode: %w", err)
	}

	_, err = db.Exec("PRAGMA foreign_keys=ON;")
	if err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	return nil
}

// createSchema creates the events table and index if they don't exist
func createSchema(db *sql.DB) error {
	schema := `
CREATE TABLE IF NOT EXISTS events (
    id TEXT PRIMARY KEY,
    repo_id TEXT NOT NULL,
    branch TEXT NOT NULL,
    type TEXT NOT NULL,
    text TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    meta_json TEXT NOT NULL DEFAULT '{}'
);

CREATE INDEX IF NOT EXISTS idx_events_lookup
ON events(repo_id, branch, created_at DESC);
`

	_, err := db.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	// Add new columns if they don't exist (safe migrations)
	if err := addColumnIfNotExists(db, "events", "completed_at", "INTEGER"); err != nil {
		return err
	}
	if err := addColumnIfNotExists(db, "events", "deleted_at", "INTEGER"); err != nil {
		return err
	}
	if err := addColumnIfNotExists(db, "events", "updated_at", "INTEGER"); err != nil {
		return err
	}

	return nil
}

// addColumnIfNotExists safely adds a column if it doesn't exist
func addColumnIfNotExists(db *sql.DB, table, column, colType string) error {
	// Check if column exists
	var count int
	query := `SELECT COUNT(*) FROM pragma_table_info(?) WHERE name = ?`
	err := db.QueryRow(query, table, column).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check column %s: %w", column, err)
	}

	if count == 0 {
		// Column doesn't exist, add it
		alterSQL := fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s %s", table, column, colType)
		_, err := db.Exec(alterSQL)
		if err != nil {
			return fmt.Errorf("failed to add column %s: %w", column, err)
		}
	}

	return nil
}

// AddEvent inserts a new event into the database
func AddEvent(db *sql.DB, event Event) error {
	query := `
INSERT INTO events (id, repo_id, branch, type, text, created_at, meta_json)
VALUES (?, ?, ?, ?, ?, ?, ?)`

	_, err := db.Exec(query,
		event.ID,
		event.RepoID,
		event.Branch,
		event.Type,
		event.Text,
		event.CreatedAt,
		event.MetaJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

// StatusSummary contains event counts and metadata for a branch
type StatusSummary struct {
	Branch       string
	LastUpdated  int64
	CountsByType map[string]int
}

// GetStatus returns event counts and last updated time for a repo/branch
func GetStatus(db *sql.DB, repoID, branch string) (StatusSummary, error) {
	summary := StatusSummary{
		Branch:      branch,
		CountsByType: make(map[string]int),
	}

	// Get counts by type
	rows, err := db.Query(`
SELECT type, COUNT(*) as count
FROM events
WHERE repo_id = ? AND branch = ?
GROUP BY type
ORDER BY type`, repoID, branch)

	if err != nil {
		return summary, fmt.Errorf("failed to query event counts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var eventType string
		var count int
		if err := rows.Scan(&eventType, &count); err != nil {
			return summary, fmt.Errorf("failed to scan row: %w", err)
		}
		summary.CountsByType[eventType] = count
	}

	// Get last updated timestamp
	var lastUpdated sql.NullInt64
	err = db.QueryRow(`
SELECT MAX(created_at)
FROM events
WHERE repo_id = ? AND branch = ?`, repoID, branch).Scan(&lastUpdated)

	if err != nil && err != sql.ErrNoRows {
		return summary, fmt.Errorf("failed to query last updated: %w", err)
	}

	if lastUpdated.Valid {
		summary.LastUpdated = lastUpdated.Int64
	}

	return summary, nil
}

// GetEvents fetches all events for a repo/branch with a limit
// Returns events in descending order by created_at (newest first)
// Caller should group and sort by type as needed
func GetEvents(db *sql.DB, repoID, branch string, limit int) ([]Event, error) {
	query := `
SELECT id, repo_id, branch, type, text, created_at, meta_json, completed_at, deleted_at, updated_at
FROM events
WHERE repo_id = ? AND branch = ? AND (deleted_at IS NULL OR deleted_at = 0)
ORDER BY created_at DESC
LIMIT ?`

	rows, err := db.Query(query, repoID, branch, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []Event
	for rows.Next() {
		var event Event
		var completedAt, deletedAt, updatedAt sql.NullInt64
		err := rows.Scan(
			&event.ID,
			&event.RepoID,
			&event.Branch,
			&event.Type,
			&event.Text,
			&event.CreatedAt,
			&event.MetaJSON,
			&completedAt,
			&deletedAt,
			&updatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan event: %w", err)
		}

		// Store nullable timestamps
		if completedAt.Valid {
			event.CompletedAt = &completedAt.Int64
		}
		if deletedAt.Valid {
			event.DeletedAt = &deletedAt.Int64
		}
		if updatedAt.Valid {
			event.UpdatedAt = &updatedAt.Int64
		}

		events = append(events, event)
	}

	return events, nil
}

// GetGoal retrieves the current goal for a branch (single goal per branch)
func GetGoal(db *sql.DB, repoID, branch string) (*Event, error) {
	query := `
SELECT id, repo_id, branch, type, text, created_at, meta_json
FROM events
WHERE repo_id = ? AND branch = ? AND type = 'goal'
ORDER BY created_at DESC
LIMIT 1`

	var event Event
	err := db.QueryRow(query, repoID, branch).Scan(
		&event.ID,
		&event.RepoID,
		&event.Branch,
		&event.Type,
		&event.Text,
		&event.CreatedAt,
		&event.MetaJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No goal exists
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get goal: %w", err)
	}

	return &event, nil
}

// UpsertGoal sets or updates the goal for a branch
// Only one goal per branch - updates existing or inserts new
func UpsertGoal(db *sql.DB, repoID, branch, text string) error {
	// Check if goal exists
	existing, err := GetGoal(db, repoID, branch)
	if err != nil {
		return err
	}

	if existing != nil {
		// Update existing goal
		query := `UPDATE events SET text = ?, created_at = ? WHERE id = ?`
		_, err := db.Exec(query, text, time.Now().Unix(), existing.ID)
		if err != nil {
			return fmt.Errorf("failed to update goal: %w", err)
		}
	} else {
		// Insert new goal
		event := Event{
			ID:        uuid.New().String(),
			RepoID:    repoID,
			Branch:    branch,
			Type:      "goal",
			Text:      text,
			CreatedAt: time.Now().Unix(),
			MetaJSON:  "{}",
		}
		if err := AddEvent(db, event); err != nil {
			return err
		}
	}

	return nil
}

// SaveStash records a git stash reference for a branch
func SaveStash(db *sql.DB, repoID, branch, message, stashRef, metaJSON string) error {
	event := Event{
		ID:        uuid.New().String(),
		RepoID:    repoID,
		Branch:    branch,
		Type:      "stash",
		Text:      message,
		CreatedAt: time.Now().Unix(),
		MetaJSON:  metaJSON,
	}

	if err := AddEvent(db, event); err != nil {
		return fmt.Errorf("failed to save stash record: %w", err)
	}

	return nil
}

// GetLatestStash retrieves the most recent stash for a branch
func GetLatestStash(db *sql.DB, repoID, branch string) (*Event, error) {
	query := `
SELECT id, repo_id, branch, type, text, created_at, meta_json
FROM events
WHERE repo_id = ? AND branch = ? AND type = 'stash'
ORDER BY created_at DESC
LIMIT 1`

	var event Event
	err := db.QueryRow(query, repoID, branch).Scan(
		&event.ID,
		&event.RepoID,
		&event.Branch,
		&event.Type,
		&event.Text,
		&event.CreatedAt,
		&event.MetaJSON,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get stash: %w", err)
	}

	return &event, nil
}

// GetAllStashes retrieves all stashes across all branches
func GetAllStashes(db *sql.DB, repoID string) ([]Event, error) {
	query := `
SELECT id, repo_id, branch, type, text, created_at, meta_json
FROM events
WHERE repo_id = ? AND type = 'stash'
ORDER BY created_at DESC`

	rows, err := db.Query(query, repoID)
	if err != nil {
		return nil, fmt.Errorf("failed to query stashes: %w", err)
	}
	defer rows.Close()

	var stashes []Event
	for rows.Next() {
		var event Event
		err := rows.Scan(
			&event.ID,
			&event.RepoID,
			&event.Branch,
			&event.Type,
			&event.Text,
			&event.CreatedAt,
			&event.MetaJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan stash: %w", err)
		}
		stashes = append(stashes, event)
	}

	return stashes, nil
}

// ToggleTodoCompletion toggles the completion state of a todo
func ToggleTodoCompletion(db *sql.DB, eventID string) error {
	// Get current completion state
	var completedAt sql.NullInt64
	err := db.QueryRow(`SELECT completed_at FROM events WHERE id = ?`, eventID).Scan(&completedAt)
	if err != nil {
		return fmt.Errorf("failed to get completion state: %w", err)
	}

	// Toggle: if completed, set to NULL; if not completed, set to now
	var newValue interface{}
	if completedAt.Valid {
		newValue = nil
	} else {
		newValue = time.Now().Unix()
	}

	_, err = db.Exec(`UPDATE events SET completed_at = ? WHERE id = ?`, newValue, eventID)
	if err != nil {
		return fmt.Errorf("failed to toggle completion: %w", err)
	}

	return nil
}

// SoftDeleteEvent marks an event as deleted
func SoftDeleteEvent(db *sql.DB, eventID string) error {
	_, err := db.Exec(`UPDATE events SET deleted_at = ? WHERE id = ?`, time.Now().Unix(), eventID)
	if err != nil {
		return fmt.Errorf("failed to soft delete: %w", err)
	}
	return nil
}

// RestoreDeletedEvent restores a soft-deleted event
func RestoreDeletedEvent(db *sql.DB, eventID string) error {
	_, err := db.Exec(`UPDATE events SET deleted_at = NULL WHERE id = ?`, eventID)
	if err != nil {
		return fmt.Errorf("failed to restore event: %w", err)
	}
	return nil
}

// UpdateEventText updates the text of an event and sets updated_at
func UpdateEventText(db *sql.DB, eventID, newText string) error {
	_, err := db.Exec(`UPDATE events SET text = ?, updated_at = ? WHERE id = ?`,
		newText, time.Now().Unix(), eventID)
	if err != nil {
		return fmt.Errorf("failed to update text: %w", err)
	}
	return nil
}

// ClearBranch deletes all events for a specific branch
func ClearBranch(db *sql.DB, repoID, branch string) error {
	_, err := db.Exec(`DELETE FROM events WHERE repo_id = ? AND branch = ?`, repoID, branch)
	if err != nil {
		return fmt.Errorf("failed to clear branch: %w", err)
	}
	return nil
}
