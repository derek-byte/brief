package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/derek-byte/coding-tools/notetaker-dev/internal/git"
	"github.com/derek-byte/coding-tools/notetaker-dev/internal/store"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:     "save [message]",
	Short:   "Save (stash) your current work for this branch",
	GroupID: "branch",
	Long: `Stash your uncommitted changes and associate them with the current branch.
Later, use 'brief restore' to restore your work on this branch.

This is like 'git stash' but organized per-branch so you always apply
the right stash to the right branch.

Examples:
  brief save "OAuth integration WIP"
  brief save "Fixing retry logic"
  brief save`,
	Args: cobra.MaximumNArgs(1),
	RunE: runSave,
}

func init() {
	rootCmd.AddCommand(saveCmd)
}

func runSave(cmd *cobra.Command, args []string) error {
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

	// Check if there are changes to stash
	statusCmd := exec.Command("git", "-C", repoRoot, "status", "--porcelain")
	var statusOut bytes.Buffer
	statusCmd.Stdout = &statusOut
	if err := statusCmd.Run(); err != nil {
		return fmt.Errorf("failed to check git status: %w", err)
	}

	if strings.TrimSpace(statusOut.String()) == "" {
		fmt.Println("No changes to save (working tree is clean)")
		return nil
	}

	// Get list of changed files for context
	files := strings.Split(strings.TrimSpace(statusOut.String()), "\n")
	fileList := make([]string, 0, len(files))
	for _, line := range files {
		if len(line) > 3 {
			fileList = append(fileList, strings.TrimSpace(line[3:]))
		}
	}

	// Build stash message
	message := "Work in progress"
	if len(args) > 0 {
		message = args[0]
	}

	stashMessage := fmt.Sprintf("brief-save: %s [%s]", message, branch)

	// Create git stash
	stashCmd := exec.Command("git", "-C", repoRoot, "stash", "push", "-m", stashMessage)
	if err := stashCmd.Run(); err != nil {
		return fmt.Errorf("failed to create stash: %w", err)
	}

	// Get the stash reference (stash@{0})
	stashRef := "stash@{0}"

	// Build metadata
	meta := map[string]interface{}{
		"stash_ref": stashRef,
		"files":     fileList,
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Open database and save stash record
	db, err := store.OpenDB()
	if err != nil {
		return err
	}
	defer db.Close()

	if err := store.SaveStash(db, repoID, branch, message, stashRef, string(metaJSON)); err != nil {
		return err
	}

	fmt.Printf("Saved work for %s (%d files)\n", branch, len(fileList))
	fmt.Printf("Restore with: brief restore\n")

	return nil
}
