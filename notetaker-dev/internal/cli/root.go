package cli

import (
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
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}
