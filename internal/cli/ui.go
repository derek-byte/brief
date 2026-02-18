package cli

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/derek-byte/coding-tools/internal/git"
	"github.com/derek-byte/coding-tools/internal/store"
	"github.com/derek-byte/coding-tools/internal/tui"
	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:     "ui",
	Short:   "Launch interactive TUI for browsing and managing branch context",
	GroupID: "branch",
	Long: `Launch an interactive terminal UI for viewing and interacting with
branch notes. Navigate with j/k, toggle views with v, quit with q.`,
	RunE: runUI,
}

func init() {
	rootCmd.AddCommand(uiCmd)
}

func runUI(cmd *cobra.Command, args []string) error {
	// Check if running in TTY
	if !isTerminal() {
		return &UserError{Msg: "brief ui requires a terminal (not a pipe or redirect)"}
	}

	// Get repo root and branch
	repoRoot, err := git.RepoRoot()
	if err != nil {
		return err
	}

	branch, err := git.CurrentBranch(repoRoot)
	if err != nil {
		return err
	}

	repoID := git.RepoID(repoRoot)

	// Open database
	db, err := store.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	// Fetch events (limit 200 for now, same as rehydrate)
	events, err := store.GetEvents(db, repoID, branch, 200)
	if err != nil {
		return err
	}

	// Initialize TUI model with DB connection for mutations
	model := tui.NewModel(events, branch, repoID, repoRoot, db)

	// Run Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	// Handle post-exit actions (command execution or editing)
	return handlePostExit(finalModel, repoRoot, db)
}

// isTerminal checks if stdout is a terminal (not a pipe)
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// handlePostExit handles command execution or editing after TUI exits
func handlePostExit(model tea.Model, repoRoot string, db *sql.DB) error {
	// Type assert to tui.Model to access pending actions
	tuiModel, ok := model.(tui.Model)
	if !ok {
		return nil
	}

	// Handle pending command execution
	if cmd := tuiModel.GetPendingCommand(); cmd != "" {
		return runCommand(cmd, repoRoot)
	}

	// Handle pending edit
	if item := tuiModel.GetPendingEdit(); item != nil {
		return editItem(item, db)
	}

	return nil
}

// runCommand executes a shell command in the repo root
func runCommand(cmdText, repoRoot string) error {
	// Use bash -lc for Mac compatibility (loads user's shell profile)
	cmd := exec.Command("bash", "-lc", cmdText)
	cmd.Dir = repoRoot
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("\n$ %s\n", cmdText)
	return cmd.Run()
}

// editItem opens an item in $EDITOR, saves changes to DB
func editItem(item *tui.Item, db *sql.DB) error {
	// Get editor from environment
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vi" // Fallback to vi
	}

	// Create temp file
	tmpFile, err := os.CreateTemp("", "brief-edit-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Write current text to temp file
	if _, err := tmpFile.WriteString(item.Event.Text); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	tmpFile.Close()

	// Open editor
	cmd := exec.Command("sh", "-c", editor+" "+tmpPath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	// Read back the edited text
	editedBytes, err := os.ReadFile(tmpPath)
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}
	editedText := strings.TrimSpace(string(editedBytes))

	// Only update if text changed
	if editedText != item.Event.Text && editedText != "" {
		if err := store.UpdateEventText(db, item.Event.ID, editedText); err != nil {
			return fmt.Errorf("failed to update event: %w", err)
		}
		fmt.Println("Updated successfully")
	} else {
		fmt.Println("No changes made")
	}

	return nil
}
