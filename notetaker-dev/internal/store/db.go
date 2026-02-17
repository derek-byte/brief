package store

import (
	"database/sql"
	"fmt"
	"path/filepath"

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
