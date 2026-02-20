package store

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/google/uuid"
)

// newTestDB creates a temp-file SQLite database for testing
func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "test.sqlite")
	db, err := OpenDBAt(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

// TestAddAndGetEvents_ScopedToBranch verifies branch isolation
func TestAddAndGetEvents_ScopedToBranch(t *testing.T) {
	db := newTestDB(t)

	repoID := "test-repo"
	now := time.Now().Unix()

	// Add events to different branches
	mainEvent := Event{
		ID:        uuid.New().String(),
		RepoID:    repoID,
		Branch:    "main",
		Type:      "todo",
		Text:      "Main branch task",
		CreatedAt: now,
		MetaJSON:  "{}",
	}

	featureEvent := Event{
		ID:        uuid.New().String(),
		RepoID:    repoID,
		Branch:    "feature",
		Type:      "todo",
		Text:      "Feature branch task",
		CreatedAt: now,
		MetaJSON:  "{}",
	}

	if err := AddEvent(db, mainEvent); err != nil {
		t.Fatalf("failed to add main event: %v", err)
	}

	if err := AddEvent(db, featureEvent); err != nil {
		t.Fatalf("failed to add feature event: %v", err)
	}

	// Query main branch - should only get main event
	mainEvents, err := GetEvents(db, repoID, "main", 10)
	if err != nil {
		t.Fatalf("failed to get main events: %v", err)
	}

	if len(mainEvents) != 1 {
		t.Fatalf("expected 1 event on main, got %d", len(mainEvents))
	}

	if mainEvents[0].Text != "Main branch task" {
		t.Errorf("expected 'Main branch task', got '%s'", mainEvents[0].Text)
	}

	// Query feature branch - should only get feature event
	featureEvents, err := GetEvents(db, repoID, "feature", 10)
	if err != nil {
		t.Fatalf("failed to get feature events: %v", err)
	}

	if len(featureEvents) != 1 {
		t.Fatalf("expected 1 event on feature, got %d", len(featureEvents))
	}

	if featureEvents[0].Text != "Feature branch task" {
		t.Errorf("expected 'Feature branch task', got '%s'", featureEvents[0].Text)
	}
}

// TestGetEvents_ExcludesDeleted verifies soft delete filtering
func TestGetEvents_ExcludesDeleted(t *testing.T) {
	db := newTestDB(t)

	repoID := "test-repo"
	branch := "main"

	event := Event{
		ID:        uuid.New().String(),
		RepoID:    repoID,
		Branch:    branch,
		Type:      "note",
		Text:      "This will be deleted",
		CreatedAt: time.Now().Unix(),
		MetaJSON:  "{}",
	}

	// Add event
	if err := AddEvent(db, event); err != nil {
		t.Fatalf("failed to add event: %v", err)
	}

	// Verify it appears
	events, err := GetEvents(db, repoID, branch, 10)
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("expected 1 event before delete, got %d", len(events))
	}

	// Soft delete it
	if err := SoftDeleteEvent(db, event.ID); err != nil {
		t.Fatalf("failed to soft delete event: %v", err)
	}

	// Verify it's excluded from results
	events, err = GetEvents(db, repoID, branch, 10)
	if err != nil {
		t.Fatalf("failed to get events after delete: %v", err)
	}

	if len(events) != 0 {
		t.Errorf("expected 0 events after soft delete, got %d (soft delete not filtering)", len(events))
	}
}

// TestToggleTodoCompletion verifies todo completion state transitions
func TestToggleTodoCompletion(t *testing.T) {
	db := newTestDB(t)

	repoID := "test-repo"
	branch := "main"

	todo := Event{
		ID:        uuid.New().String(),
		RepoID:    repoID,
		Branch:    branch,
		Type:      "todo",
		Text:      "Task to toggle",
		CreatedAt: time.Now().Unix(),
		MetaJSON:  "{}",
	}

	// Add todo
	if err := AddEvent(db, todo); err != nil {
		t.Fatalf("failed to add todo: %v", err)
	}

	// Toggle to completed
	if err := ToggleTodoCompletion(db, todo.ID); err != nil {
		t.Fatalf("failed to toggle todo completion: %v", err)
	}

	// Verify it's completed
	events, err := GetEvents(db, repoID, branch, 10)
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].CompletedAt == nil || *events[0].CompletedAt == 0 {
		t.Error("expected todo to be completed (completed_at should be set)")
	}

	// Toggle back to incomplete
	if err := ToggleTodoCompletion(db, todo.ID); err != nil {
		t.Fatalf("failed to toggle todo back to incomplete: %v", err)
	}

	// Verify it's incomplete
	events, err = GetEvents(db, repoID, branch, 10)
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}

	if events[0].CompletedAt != nil && *events[0].CompletedAt > 0 {
		t.Error("expected todo to be incomplete (completed_at should be null/0)")
	}
}

// TestUpsertGoal_OnlyOneGoalPerBranch verifies goal uniqueness constraint
func TestUpsertGoal_OnlyOneGoalPerBranch(t *testing.T) {
	db := newTestDB(t)

	repoID := "test-repo"
	branch := "main"

	// Set initial goal
	if err := UpsertGoal(db, repoID, branch, "First goal"); err != nil {
		t.Fatalf("failed to upsert first goal: %v", err)
	}

	// Get goal - should return first goal
	goal, err := GetGoal(db, repoID, branch)
	if err != nil {
		t.Fatalf("failed to get goal: %v", err)
	}

	if goal == nil {
		t.Fatal("expected goal to exist")
	}

	if goal.Text != "First goal" {
		t.Errorf("expected 'First goal', got '%s'", goal.Text)
	}

	// Upsert new goal text - should update, not create duplicate
	if err := UpsertGoal(db, repoID, branch, "Updated goal"); err != nil {
		t.Fatalf("failed to upsert updated goal: %v", err)
	}

	// Get goal - should return updated text
	goal, err = GetGoal(db, repoID, branch)
	if err != nil {
		t.Fatalf("failed to get updated goal: %v", err)
	}

	if goal.Text != "Updated goal" {
		t.Errorf("expected 'Updated goal', got '%s'", goal.Text)
	}

	// Verify only one goal exists (no duplicates)
	events, err := GetEvents(db, repoID, branch, 100)
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}

	goalCount := 0
	for _, e := range events {
		if e.Type == "goal" {
			goalCount++
		}
	}

	if goalCount != 1 {
		t.Errorf("expected exactly 1 goal, found %d (upsert created duplicate)", goalCount)
	}
}

// TestUndoDelete_Restores verifies restore functionality
func TestUndoDelete_Restores(t *testing.T) {
	db := newTestDB(t)

	repoID := "test-repo"
	branch := "main"

	event := Event{
		ID:        uuid.New().String(),
		RepoID:    repoID,
		Branch:    branch,
		Type:      "note",
		Text:      "Will be deleted then restored",
		CreatedAt: time.Now().Unix(),
		MetaJSON:  "{}",
	}

	// Add event
	if err := AddEvent(db, event); err != nil {
		t.Fatalf("failed to add event: %v", err)
	}

	// Soft delete it
	if err := SoftDeleteEvent(db, event.ID); err != nil {
		t.Fatalf("failed to soft delete: %v", err)
	}

	// Verify it's gone
	events, err := GetEvents(db, repoID, branch, 10)
	if err != nil {
		t.Fatalf("failed to get events: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("expected 0 events after delete, got %d", len(events))
	}

	// Restore it
	if err := RestoreDeletedEvent(db, event.ID); err != nil {
		t.Fatalf("failed to restore event: %v", err)
	}

	// Verify it's back
	events, err = GetEvents(db, repoID, branch, 10)
	if err != nil {
		t.Fatalf("failed to get events after restore: %v", err)
	}

	if len(events) != 1 {
		t.Errorf("expected 1 event after restore, got %d (restore failed)", len(events))
	}

	if len(events) > 0 && events[0].Text != "Will be deleted then restored" {
		t.Errorf("expected restored event text, got '%s'", events[0].Text)
	}
}
