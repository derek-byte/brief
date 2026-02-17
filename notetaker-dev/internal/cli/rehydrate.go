package cli

import (
	"fmt"
	"time"

	"github.com/derek/branchbrief/internal/git"
	"github.com/derek/branchbrief/internal/render"
	"github.com/derek/branchbrief/internal/store"
	"github.com/spf13/cobra"
)

var limit int

var rehydrateCmd = &cobra.Command{
	Use:   "rehydrate",
	Short: "Display a concise brief of the current branch context",
	Long: `Print a structured summary of your branch's development context,
including goals, decisions, todos, and git state. Designed to get
you oriented in under 60 seconds.`,
	RunE: runRehydrate,
}

func init() {
	rehydrateCmd.Flags().IntVar(&limit, "limit", 200, "Maximum number of events to fetch")
	rootCmd.AddCommand(rehydrateCmd)
}

func runRehydrate(cmd *cobra.Command, args []string) error {
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

	// Fetch all events with one query (avoid N+1)
	events, err := store.GetEvents(db, repoID, branch, limit)
	if err != nil {
		return err
	}

	// Get git state
	gitState := git.GetGitState(repoRoot)

	// Determine last updated time
	var lastUpdated time.Time
	if len(events) > 0 {
		lastUpdated = time.Unix(events[0].CreatedAt, 0)
	} else {
		lastUpdated = time.Now()
	}

	// Build brief
	brief := render.Brief{
		Branch:      branch,
		LastUpdated: lastUpdated,
		GitState:    gitState,
		Events:      events,
	}

	// Render and display
	output := render.RenderBrief(brief)
	fmt.Print(output)

	return nil
}
