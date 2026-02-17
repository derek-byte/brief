package cli

import (
	"strings"

	"github.com/spf13/cobra"
)

// UserError represents a user-facing error (bad arguments, invalid input)
type UserError struct {
	Msg string
}

func (e *UserError) Error() string {
	return e.Msg
}

var rootCmd = &cobra.Command{
	Use:   "brief",
	Short: "BranchBrief - Local git branch context manager",
	Long: `BranchBrief is a local-first CLI that stores private, branch-scoped
development notes and prints a concise rehydration brief so you can
resume work in ~30 seconds without rereading Slack/LLM chats.`,
	Version: "0.1.0",
	Args:    cobra.ArbitraryArgs,
	RunE:    runBrief,
}

func init() {
	// Hide completion command from help (still works, just not shown)
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

// runBrief is the default handler when no subcommand is used
// Treats "brief text" as "brief note text" (catch-all)
func runBrief(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		// No args - show help
		return cmd.Help()
	}

	// Catch-all: treat as note
	text := strings.Join(args, " ")
	return addEvent("note", text)
}
