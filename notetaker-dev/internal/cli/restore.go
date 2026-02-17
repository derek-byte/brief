package cli

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/derek/branchbrief/internal/git"
	"github.com/derek/branchbrief/internal/store"
	"github.com/spf13/cobra"
)

var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore saved work for this branch",
	Long: `Restore (apply) the most recent stash saved for this branch.

Uses 'git stash apply' so the stash is preserved. If you want to delete
the stash after applying, use 'git stash pop' manually.

Examples:
  brief restore`,
	RunE: runRestore,
}

func init() {
	rootCmd.AddCommand(restoreCmd)
}

func runRestore(cmd *cobra.Command, args []string) error {
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

	// Get latest stash for this branch
	stashRecord, err := store.GetLatestStash(db, repoID, branch)
	if err != nil {
		return err
	}

	if stashRecord == nil {
		fmt.Printf("No saved work for branch %s\n", branch)
		fmt.Println("Save work with: brief save \"message\"")
		return nil
	}

	// Parse metadata
	var meta map[string]interface{}
	if err := json.Unmarshal([]byte(stashRecord.MetaJSON), &meta); err != nil {
		return fmt.Errorf("failed to parse stash metadata: %w", err)
	}

	stashRef, ok := meta["stash_ref"].(string)
	if !ok || stashRef == "" {
		return fmt.Errorf("invalid stash reference in metadata")
	}

	// Show what we're about to restore
	savedTime := time.Unix(stashRecord.CreatedAt, 0)
	fmt.Printf("Restoring: %s\n", stashRecord.Text)
	fmt.Printf("Saved: %s\n", savedTime.Format("2006-01-02 15:04"))

	if files, ok := meta["files"].([]interface{}); ok && len(files) > 0 {
		fmt.Printf("Files: %d changed\n", len(files))
		// Show first few files
		for i, f := range files {
			if i >= 5 {
				fmt.Printf("  ... and %d more\n", len(files)-5)
				break
			}
			fmt.Printf("  %s\n", f)
		}
	}

	fmt.Println()

	// Verify stash still exists
	checkCmd := exec.Command("git", "-C", repoRoot, "stash", "list")
	checkOut, err := checkCmd.Output()
	if err != nil {
		return fmt.Errorf("failed to list stashes: %w", err)
	}

	if !strings.Contains(string(checkOut), stashRef) {
		return fmt.Errorf("stash %s no longer exists (may have been dropped)", stashRef)
	}

	// Apply stash
	applyCmd := exec.Command("git", "-C", repoRoot, "stash", "apply", stashRef)
	applyCmd.Stdout = cmd.OutOrStdout()
	applyCmd.Stderr = cmd.ErrOrStderr()

	if err := applyCmd.Run(); err != nil {
		return fmt.Errorf("failed to apply stash: %w", err)
	}

	fmt.Println("\nWork restored successfully!")
	fmt.Println("(Stash preserved - run 'git stash drop' to remove it)")

	return nil
}
