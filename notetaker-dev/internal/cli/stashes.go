package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/derek-byte/coding-tools/notetaker-dev/internal/git"
	"github.com/derek-byte/coding-tools/notetaker-dev/internal/store"
	"github.com/spf13/cobra"
)

var stashesCmd = &cobra.Command{
	Use:     "stashes",
	Short:   "List all saved work for this branch",
	GroupID: "branch",
	Long: `Show all saved stashes for the current branch.

Use this to see what work-in-progress you've saved on this branch.
Use 'brief restore' to restore the most recent stash.

Examples:
  brief stashes`,
	RunE: runStashes,
}

func init() {
	rootCmd.AddCommand(stashesCmd)
}

func runStashes(cmd *cobra.Command, args []string) error {
	// Get repo root and current branch
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

	// Get all stashes for this repo
	allStashes, err := store.GetAllStashes(db, repoID)
	if err != nil {
		return err
	}

	// Filter to current branch only
	var stashes []store.Event
	for _, stash := range allStashes {
		if stash.Branch == branch {
			stashes = append(stashes, stash)
		}
	}

	if len(stashes) == 0 {
		fmt.Printf("No saved work for branch '%s'\n", branch)
		fmt.Println("Save work with: brief save \"message\"")
		return nil
	}

	// Display stashes for current branch
	fmt.Printf("Saved work for branch '%s' (%d stash%s):\n\n", branch, len(stashes), plural(len(stashes)))

	for _, stash := range stashes {
		savedTime := time.Unix(stash.CreatedAt, 0)
		timeAgo := formatTimeAgo(time.Since(savedTime))

		// Parse metadata for file count
		var meta map[string]interface{}
		fileCount := 0
		if err := json.Unmarshal([]byte(stash.MetaJSON), &meta); err == nil {
			if files, ok := meta["files"].([]interface{}); ok {
				fileCount = len(files)
			}
		}

		fmt.Printf("  • %s (%d files, %s)\n", stash.Text, fileCount, timeAgo)
	}
	fmt.Println()

	return nil
}

// formatTimeAgo converts duration to human-readable format
func formatTimeAgo(d time.Duration) string {
	hours := int(d.Hours())
	if hours < 1 {
		minutes := int(d.Minutes())
		if minutes < 1 {
			return "just now"
		}
		return fmt.Sprintf("%d minute%s ago", minutes, plural(minutes))
	}
	if hours < 24 {
		return fmt.Sprintf("%d hour%s ago", hours, plural(hours))
	}
	days := hours / 24
	if days < 7 {
		return fmt.Sprintf("%d day%s ago", days, plural(days))
	}
	weeks := days / 7
	return fmt.Sprintf("%d week%s ago", weeks, plural(weeks))
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}
