package render

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/derek-byte/coding-tools/internal/git"
	"github.com/derek-byte/coding-tools/internal/store"
)

// goldenTest compares output against a golden file
// Set UPDATE_GOLDEN=1 env var to regenerate golden files
func goldenTest(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")

	if os.Getenv("UPDATE_GOLDEN") == "1" {
		// Update mode: write new golden file
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("failed to create testdata dir: %v", err)
		}
		if err := os.WriteFile(path, []byte(got), 0644); err != nil {
			t.Fatalf("failed to write golden file: %v", err)
		}
		t.Logf("Updated: %s", path)
		return
	}

	// Compare mode: check against golden file
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Missing golden file: %s (run with UPDATE_GOLDEN=1)", path)
	}

	if got != string(want) {
		t.Errorf("Output mismatch for %s\n\nGot:\n%s\n\nWant:\n%s", name, got, want)
	}
}

// makeTestBrief creates a Brief with sample events for testing
func makeTestBrief() Brief {
	now := time.Now().Unix()

	return Brief{
		Branch:      "feature/payment-retry",
		LastUpdated: time.Now(),
		GitState: git.GitState{
			LastCommit:     "abc123 Add retry logic",
			DirtyFileCount: 2,
		},
		Events: []store.Event{
			{
				ID:        "1",
				Type:      "todo",
				Text:      "Add exponential backoff",
				CreatedAt: now,
			},
			{
				ID:        "2",
				Type:      "todo",
				Text:      "Test edge cases for network timeouts",
				CreatedAt: now - 100,
			},
			{
				ID:        "3",
				Type:      "todo",
				Text:      "Update documentation",
				CreatedAt: now - 200,
			},
			{
				ID:        "4",
				Type:      "choice",
				Text:      "Use Redis for rate limiting",
				CreatedAt: now - 50,
			},
			{
				ID:        "5",
				Type:      "choice",
				Text:      "Circuit breaker pattern for external API",
				CreatedAt: now - 150,
			},
			{
				ID:        "6",
				Type:      "note",
				Text:      "API docs at docs.stripe.com/payments/retries",
				CreatedAt: now - 75,
			},
			{
				ID:        "7",
				Type:      "note",
				Text:      "Max retry count should be configurable",
				CreatedAt: now - 175,
			},
		},
	}
}

// TestSummary_Width110_ThreeColumns tests 3-column layout
func TestSummary_Width110_ThreeColumns(t *testing.T) {
	t.Setenv("TERMINAL_WIDTH", "110")

	brief := makeTestBrief()
	output := RenderSummary(brief)

	goldenTest(t, "summary-110", output)
}

// TestSummary_Width80_TwoColumns tests 2-column layout
func TestSummary_Width80_TwoColumns(t *testing.T) {
	t.Setenv("TERMINAL_WIDTH", "80")

	brief := makeTestBrief()
	output := RenderSummary(brief)

	goldenTest(t, "summary-80", output)
}

// TestSummary_Width60_Stacked tests stacked layout
func TestSummary_Width60_Stacked(t *testing.T) {
	t.Setenv("TERMINAL_WIDTH", "60")

	brief := makeTestBrief()
	output := RenderSummary(brief)

	goldenTest(t, "summary-60", output)
}

// TestTruncation_LongText verifies long text is truncated with ellipsis
func TestTruncation_LongText(t *testing.T) {
	t.Setenv("TERMINAL_WIDTH", "80")

	// Create event with very long text (200+ characters)
	longText := strings.Repeat("This is a very long text that should be truncated. ", 4)

	brief := Brief{
		Branch:      "test",
		LastUpdated: time.Now(),
		GitState:    git.GitState{LastCommit: "abc123"},
		Events: []store.Event{
			{
				ID:        "1",
				Type:      "todo",
				Text:      longText,
				CreatedAt: time.Now().Unix(),
			},
		},
	}

	output := RenderSummary(brief)

	// Verify truncation indicators are present
	if !strings.Contains(output, "...") && !strings.Contains(output, "…") {
		t.Error("Expected truncation indicator (... or …) in output for long text")
	}

	// Verify the full long text is NOT present
	if strings.Contains(output, longText) {
		t.Error("Long text should be truncated, but full text appears in output")
	}
}
