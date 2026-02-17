package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/derek-byte/coding-tools/notetaker-dev/internal/git"
	"github.com/derek-byte/coding-tools/notetaker-dev/internal/store"
	"github.com/derek-byte/coding-tools/notetaker-dev/internal/tui"
	"github.com/spf13/cobra"
)

var uiCmd = &cobra.Command{
	Use:   "ui",
	Short: "Launch interactive TUI for browsing and managing branch context",
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

	// Initialize TUI model
	model := tui.NewModel(events, branch)

	// Run Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run TUI: %w", err)
	}

	return nil
}

// isTerminal checks if stdout is a terminal (not a pipe)
func isTerminal() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}
