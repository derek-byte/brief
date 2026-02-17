package cli

import (
	"fmt"
	"time"

	"github.com/derek-byte/coding-tools/notetaker-dev/internal/git"
	"github.com/derek-byte/coding-tools/notetaker-dev/internal/render"
	"github.com/derek-byte/coding-tools/notetaker-dev/internal/store"
	"github.com/spf13/cobra"
)

var limit int
var viewMode string

var rehydrateCmd = &cobra.Command{
	Use:     "rehydrate",
	Short:   "Display a concise brief of the current branch context",
	GroupID: "branch",
	Long: `Print a structured summary of your branch's development context,
including goals, choices, todos, and git state. Designed to get
you oriented in under 60 seconds.`,
	RunE: runRehydrate,
}

func init() {
	rehydrateCmd.Flags().IntVar(&limit, "limit", 200, "Maximum number of events to fetch")
	rehydrateCmd.Flags().StringVar(&viewMode, "view", "structured", "View mode: structured or timeline")
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

	// Get current goal separately (single goal per branch)
	currentGoal, err := store.GetGoal(db, repoID, branch)
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
		CurrentGoal: currentGoal,
	}

	// Render based on view mode
	var output string
	switch viewMode {
	case "timeline":
		output = render.RenderTimeline(brief)
	case "structured":
		output = render.RenderBrief(brief)
	default:
		return &UserError{Msg: fmt.Sprintf("invalid view mode: %s (must be 'structured' or 'timeline')", viewMode)}
	}

	fmt.Print(output)

	return nil
}
