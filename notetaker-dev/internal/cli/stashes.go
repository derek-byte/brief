package cli

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/derek/branchbrief/internal/git"
	"github.com/derek/branchbrief/internal/store"
	"github.com/spf13/cobra"
)

var stashesCmd = &cobra.Command{
	Use:   "stashes",
	Short: "List all saved work across branches",
	Long: `Show all saved stashes organized by branch.

This helps you see what work-in-progress you have saved
across different branches.

Examples:
  brief stashes`,
	RunE: runStashes,
}

func init() {
	rootCmd.AddCommand(stashesCmd)
}

func runStashes(cmd *cobra.Command, args []string) error {
	// Get repo root
	repoRoot, err := git.RepoRoot()
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
	stashes, err := store.GetAllStashes(db, repoID)
	if err != nil {
		return err
	}

	if len(stashes) == 0 {
		fmt.Println("No saved work")
		fmt.Println("Save work with: brief save \"message\"")
		return nil
	}

	// Group by branch
	stashesByBranch := make(map[string][]store.Event)
	for _, stash := range stashes {
		stashesByBranch[stash.Branch] = append(stashesByBranch[stash.Branch], stash)
	}

	// Display
	fmt.Printf("Saved work (%d total):\n\n", len(stashes))

	for branch, branchStashes := range stashesByBranch {
		fmt.Printf("Branch: %s\n", branch)
		for _, stash := range branchStashes {
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
	}

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
