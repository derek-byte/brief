package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/derek/branchbrief/internal/git"
	"github.com/derek/branchbrief/internal/store"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <type> <text...>",
	Short: "Add a note to the current branch",
	Long: `Add a branch-scoped note of the specified type.

Valid types:
  goal, decision, todo, cmd, error, link, issue, note, fix

Examples:
  brief add goal "Implement user authentication"
  brief add decision "Use JWT tokens for sessions"
  brief add todo "Write unit tests"
  brief add cmd "make test"`,
	Args: cobra.MinimumNArgs(2),
	RunE: runAdd,
}

func init() {
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	eventType := normalizeType(args[0])
	text := strings.Join(args[1:], " ")

	// Validate type against allowlist
	if !isValidType(eventType) {
		return &UserError{
			Msg: fmt.Sprintf("invalid type: %s (valid types: goal, decision, todo, cmd, error, link, issue, note, fix)", eventType),
		}
	}

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

	// Create event
	event := store.Event{
		ID:        uuid.New().String(),
		RepoID:    repoID,
		Branch:    branch,
		Type:      eventType,
		Text:      text,
		CreatedAt: time.Now().Unix(),
		MetaJSON:  "{}",
	}

	// Insert event
	if err := store.AddEvent(db, event); err != nil {
		return err
	}

	fmt.Printf("Added %s to %s\n", eventType, branch)
	return nil
}

// normalizeType converts type to lowercase and handles synonyms
func normalizeType(t string) string {
	t = strings.ToLower(strings.TrimSpace(t))
	// Accept common synonyms
	switch t {
	case "command":
		return "cmd"
	case "bug":
		return "error"
	}
	return t
}

// isValidType checks if the event type is in the allowlist
func isValidType(t string) bool {
	valid := map[string]bool{
		"goal":     true,
		"decision": true,
		"todo":     true,
		"cmd":      true,
		"error":    true,
		"link":     true,
		"issue":    true,
		"note":     true,
		"fix":      true,
	}
	return valid[t]
}
