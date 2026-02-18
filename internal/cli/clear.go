package cli

import (
	"fmt"

	"github.com/derek-byte/coding-tools/internal/git"
	"github.com/derek-byte/coding-tools/internal/store"
	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:     "clear",
	Short:   "Clear all notes for this branch",
	GroupID: "branch",
	Long: `Delete all notes (todos, choices, commands, notes, goal) for the current branch.
This is irreversible. Use with caution.

Examples:
  brief clear              # Prompts for confirmation
  brief clear --force      # Skip confirmation`,
	RunE: runClear,
}

var clearForce bool

func init() {
	clearCmd.Flags().BoolVarP(&clearForce, "force", "f", false, "Skip confirmation prompt")
	rootCmd.AddCommand(clearCmd)
}

func runClear(cmd *cobra.Command, args []string) error {
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

	// Get count of events
	events, err := store.GetEvents(db, repoID, branch, 1000)
	if err != nil {
		return err
	}

	// Check for goal
	goal, err := store.GetGoal(db, repoID, branch)
	if err != nil {
		return err
	}

	totalItems := len(events)
	if goal != nil {
		totalItems++
	}

	if totalItems == 0 {
		fmt.Printf("No notes to clear on branch '%s'\n", branch)
		return nil
	}

	// Confirm unless --force
	if !clearForce {
		fmt.Printf("This will delete %d items from branch '%s'.\n", totalItems, branch)
		fmt.Print("Are you sure? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Cancelled")
			return nil
		}
	}

	// Delete all events
	if err := store.ClearBranch(db, repoID, branch); err != nil {
		return err
	}

	fmt.Printf("Cleared %d items from branch '%s'\n", totalItems, branch)
	return nil
}
