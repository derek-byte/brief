package cli

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/derek/branchbrief/internal/git"
	"github.com/derek/branchbrief/internal/store"
	"github.com/spf13/cobra"
)

var editGoal bool

var goalCmd = &cobra.Command{
	Use:   "goal [text...]",
	Short: "Set or view the current goal for this branch",
	Long: `Set or update the single goal for this branch.
Unlike other event types, there is only one goal per branch.
Setting a new goal replaces the previous one.

Examples:
  brief goal "Build authentication system"      # Set goal
  brief goal "Build auth and add OAuth"         # Update goal
  brief goal                                    # View current goal
  brief goal --edit                             # Edit goal in $EDITOR`,
	RunE: runGoal,
}

func init() {
	goalCmd.Flags().BoolVar(&editGoal, "edit", false, "Edit current goal in $EDITOR")
	rootCmd.AddCommand(goalCmd)
}

func runGoal(cmd *cobra.Command, args []string) error {
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

	// Handle --edit flag
	if editGoal {
		return editGoalInEditor(db, repoID, branch)
	}

	// If no args, show current goal
	if len(args) == 0 {
		currentGoal, err := store.GetGoal(db, repoID, branch)
		if err != nil {
			return err
		}
		if currentGoal == nil {
			fmt.Println("No goal set for this branch")
			fmt.Println("Set one with: brief goal \"Your goal here\"")
			return nil
		}
		fmt.Println(currentGoal.Text)
		return nil
	}

	// Set/update goal
	text := strings.Join(args, " ")
	if err := store.UpsertGoal(db, repoID, branch, text); err != nil {
		return err
	}

	fmt.Printf("Goal set for %s\n", branch)
	return nil
}

func editGoalInEditor(db *sql.DB, repoID, branch string) error {
	// Get current goal
	currentGoal, err := store.GetGoal(db, repoID, branch)
	if err != nil {
		return err
	}

	// Create temp file with current goal
	tmpfile, err := os.CreateTemp("", "brief-goal-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpfile.Name())

	if currentGoal != nil {
		if _, err := tmpfile.WriteString(currentGoal.Text); err != nil {
			return fmt.Errorf("failed to write temp file: %w", err)
		}
	}
	tmpfile.Close()

	// Open in $EDITOR
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // fallback
	}

	editorCmd := exec.Command(editor, tmpfile.Name())
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	// Read edited content
	content, err := os.ReadFile(tmpfile.Name())
	if err != nil {
		return fmt.Errorf("failed to read edited file: %w", err)
	}

	newGoal := strings.TrimSpace(string(content))
	if newGoal == "" {
		return &UserError{Msg: "goal cannot be empty (edit cancelled)"}
	}

	// Update goal
	if err := store.UpsertGoal(db, repoID, branch, newGoal); err != nil {
		return err
	}

	fmt.Printf("Goal updated for %s\n", branch)
	return nil
}
