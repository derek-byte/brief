package cli

import (
	"fmt"
	"time"

	"github.com/derek-byte/coding-tools/internal/git"
	"github.com/derek-byte/coding-tools/internal/render"
	"github.com/derek-byte/coding-tools/internal/store"
	"github.com/spf13/cobra"
)

var summaryQuiet bool

var summaryCmd = &cobra.Command{
	Use:     "summary",
	Short:   "Show compact grid summary of branch context",
	GroupID: "branch",
	Long: `Display a compact grid-based summary of branch context.
Shows goal and top items in a terminal-width-aware layout.
Silent if branch has no notes.`,
	RunE: runSummary,
}

func init() {
	summaryCmd.Flags().BoolVar(&summaryQuiet, "quiet", false, "Suppress 'no notes' message")
	rootCmd.AddCommand(summaryCmd)
}

func runSummary(cmd *cobra.Command, args []string) error {
	// Get repo info
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

	// Fetch events
	events, err := store.GetEvents(db, repoID, branch, 200)
	if err != nil {
		return err
	}

	// Get current goal
	goal, err := store.GetGoal(db, repoID, branch)
	if err != nil {
		return err
	}

	// If no events and no goal, exit silently
	if len(events) == 0 && goal == nil {
		if !summaryQuiet {
			fmt.Println("No notes for this branch yet")
		}
		return nil
	}

	// Build brief
	brief := render.Brief{
		Branch:      branch,
		LastUpdated: time.Now(),
		Events:      events,
		CurrentGoal: goal,
	}

	// Render summary
	output := render.RenderSummary(brief)
	fmt.Print(output)

	return nil
}
