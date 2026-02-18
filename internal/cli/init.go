package cli

import (
	"fmt"

	"github.com/derek-byte/coding-tools/internal/git"
	"github.com/spf13/cobra"
)

var (
	initFull      bool
	initUninstall bool
)

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Install git hook for automatic context on branch switch",
	GroupID: "global",
	Long: `Install a post-checkout git hook that automatically shows branch context
when you switch branches. The hook is opt-in and can be uninstalled.`,
	RunE: runInit,
}

func init() {
	initCmd.Flags().BoolVar(&initFull, "full", false, "Show full rehydrate output (default: compact summary)")
	initCmd.Flags().BoolVar(&initUninstall, "uninstall", false, "Remove the installed hook")
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	// Get repo root
	repoRoot, err := git.RepoRoot()
	if err != nil {
		return err
	}

	// Handle uninstall
	if initUninstall {
		if err := git.UninstallHook(repoRoot); err != nil {
			return err
		}
		fmt.Println("Uninstalled branchbrief hook from", repoRoot)
		return nil
	}

	// Install hook
	if err := git.InstallHook(repoRoot, initFull); err != nil {
		return err
	}

	mode := "summary"
	if initFull {
		mode = "full rehydrate"
	}
	fmt.Printf("Installed branchbrief hook (%s mode) in %s\n", mode, repoRoot)
	fmt.Println("Branch context will now appear automatically when you switch branches.")

	return nil
}
