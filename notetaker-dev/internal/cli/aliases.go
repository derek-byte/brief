package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

// Shorthand alias commands for common event types
// These provide ultra-low-friction alternatives to 'brief add <type>'

var todoCmd = &cobra.Command{
	Use:   "todo <text...>",
	Short: "Add a todo (shorthand for 'add todo')",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := strings.Join(args, " ")
		return addEvent("todo", text)
	},
}

var cmdCmd = &cobra.Command{
	Use:   "cmd <text...>",
	Short: "Add a command (shorthand for 'add cmd')",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := strings.Join(args, " ")
		return addEvent("cmd", text)
	},
}

var fixCmd = &cobra.Command{
	Use:   "fix <text...>",
	Short: "Add a fix note (shorthand for 'add fix')",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := strings.Join(args, " ")
		return addEvent("fix", text)
	},
}

var decisionCmd = &cobra.Command{
	Use:   "decision <text...>",
	Short: "Add a choice (alias for backwards compatibility, use 'choice' instead)",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := strings.Join(args, " ")
		return addEvent("decision", text) // Will be normalized to "choice"
	},
}

var choiceCmd = &cobra.Command{
	Use:   "choice <text...>",
	Short: "Add a choice (shorthand for 'add choice')",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := strings.Join(args, " ")
		return addEvent("choice", text)
	},
}

var errorCmd = &cobra.Command{
	Use:   "error <text...>",
	Short: "Add an error note (shorthand for 'add error')",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := strings.Join(args, " ")
		return addEvent("error", text)
	},
}

var linkCmd = &cobra.Command{
	Use:   "link <text...>",
	Short: "Add a link (shorthand for 'add link')",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := strings.Join(args, " ")
		return addEvent("link", text)
	},
}

var issueCmd = &cobra.Command{
	Use:   "issue <text...>",
	Short: "Add an issue reference (shorthand for 'add issue')",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := strings.Join(args, " ")
		return addEvent("issue", text)
	},
}

var noteCmd = &cobra.Command{
	Use:   "note <text...>",
	Short: "Add a general note (shorthand for 'add note')",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		text := strings.Join(args, " ")
		return addEvent("note", text)
	},
}

func init() {
	// Register all shorthand aliases
	// Note: 'goal' has a dedicated command in goal.go (single goal per branch)
	rootCmd.AddCommand(todoCmd)
	rootCmd.AddCommand(cmdCmd)
	rootCmd.AddCommand(fixCmd)
	rootCmd.AddCommand(decisionCmd)
	rootCmd.AddCommand(choiceCmd)
	rootCmd.AddCommand(errorCmd)
	rootCmd.AddCommand(linkCmd)
	rootCmd.AddCommand(issueCmd)
	rootCmd.AddCommand(noteCmd)
}
