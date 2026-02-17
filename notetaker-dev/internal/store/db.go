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
	ID        string
	RepoID    string
	Branch    string
	Type      string
	Text      string
	CreatedAt int64
	MetaJSON  string
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
SELECT id, repo_id, branch, type, text, created_at, meta_json
FROM events
WHERE repo_id = ? AND branch = ?
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
			return nil, fmt.Errorf("failed to scan event: %w", err)
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
