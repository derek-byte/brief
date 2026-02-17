package cli

import (
	"fmt"
	"time"

	"github.com/derek/branchbrief/internal/git"
	"github.com/derek/branchbrief/internal/store"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show event counts for the current branch",
	Long: `Display a summary of captured events for the current branch,
including counts by type and last updated timestamp.`,
	RunE: runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
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

	// Get status summary
	summary, err := store.GetStatus(db, repoID, branch)
	if err != nil {
		return err
	}

	// Display results
	fmt.Printf("Branch: %s\n", branch)

	if summary.LastUpdated > 0 {
		lastUpdate := time.Unix(summary.LastUpdated, 0)
		fmt.Printf("Last updated: %s\n\n", lastUpdate.Format("2006-01-02 15:04"))
	} else {
		fmt.Println("No events yet\n")
		return nil
	}

	// Display counts by type
	if len(summary.CountsByType) == 0 {
		fmt.Println("No events yet")
		return nil
	}

	for eventType, count := range summary.CountsByType {
		plural := ""
		if count != 1 {
			plural = "s"
		}
		fmt.Printf("%d %s%s\n", count, eventType, plural)
	}

	return nil
}
