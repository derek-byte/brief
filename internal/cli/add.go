package cli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/derek-byte/coding-tools/internal/git"
	"github.com/derek-byte/coding-tools/internal/store"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var fromStdin bool

var addCmd = &cobra.Command{
	Use:    "add <type> <text...>",
	Short:  "Add a note to the current branch",
	Hidden: true, // Hidden - use shortcuts instead
	Long: `Add a branch-scoped note of the specified type.

Valid types:
  goal, choice, todo, cmd, note, fix

Examples:
  brief add goal "Implement user authentication"
  brief add choice "Use JWT tokens for sessions"
  brief add todo "Write unit tests"
  brief add cmd "make test"`,
	Args: cobra.MinimumNArgs(2),
	RunE: runAdd,
}

func init() {
	addCmd.Flags().BoolVar(&fromStdin, "from-stdin", false, "Read additional text from stdin")
	rootCmd.AddCommand(addCmd)
}

func runAdd(cmd *cobra.Command, args []string) error {
	eventType := normalizeType(args[0])
	text := strings.Join(args[1:], " ")

	// Read from stdin if requested
	if fromStdin {
		stdinBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read stdin: %w", err)
		}
		stdinText := strings.TrimSpace(string(stdinBytes))
		if stdinText != "" {
			text = text + "\n" + stdinText
		}
	}

	return addEvent(eventType, text)
}

// addEvent is the shared logic for adding events
// Used by both 'add' command and shorthand aliases
func addEvent(eventType, text string) error {
	// Normalize type (handles synonyms like choice -> decision)
	eventType = normalizeType(eventType)

	// Auto-detect type from text prefix (e.g., "todo: text" -> type=todo, text="text")
	// Only applies when using catch-all command (eventType="note")
	if eventType == "note" {
		detectedType, cleanedText := detectTypeFromPrefix(text)
		if detectedType != "" {
			eventType = detectedType
			text = cleanedText
		}
	}

	// Validate type against allowlist
	if !isValidType(eventType) {
		return &UserError{
			Msg: fmt.Sprintf("invalid type: %s (valid types: goal, choice, todo, cmd, note, fix)", eventType),
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
	case "decision":
		return "choice" // Map old decision type to new choice type
	}
	return t
}

// isValidType checks if the event type is in the allowlist
func isValidType(t string) bool {
	valid := map[string]bool{
		"goal":   true,
		"choice": true,
		"todo":   true,
		"cmd":    true,
		"note":   true,
		"fix":    true,
	}
	return valid[t]
}

// detectTypeFromPrefix extracts type from text prefix (e.g., "todo: text" -> "todo", "text")
// Returns empty string if no valid prefix detected
func detectTypeFromPrefix(text string) (string, string) {
	// Check for "type: text" pattern
	parts := strings.SplitN(text, ":", 2)
	if len(parts) != 2 {
		return "", text
	}

	prefix := normalizeType(parts[0])
	if !isValidType(prefix) {
		return "", text
	}

	// Valid type prefix found - return type and cleaned text
	cleanedText := strings.TrimSpace(parts[1])
	return prefix, cleanedText
}
