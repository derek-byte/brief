package git

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"os/exec"
	"strings"
)

// RepoRoot returns the absolute path to the git repository root
func RepoRoot() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("not in a git repository")
	}
	return strings.TrimSpace(out.String()), nil
}

// CurrentBranch returns the current branch name, or falls back to short HEAD
func CurrentBranch(repoRoot string) (string, error) {
	out, err := runGitCommand(repoRoot, "branch", "--show-current")
	if err == nil && out != "" {
		return out, nil
	}
	// Fallback for detached HEAD
	return GetShortHEAD(repoRoot)
}

// GetShortHEAD returns the short commit hash for detached HEAD state
func GetShortHEAD(repoRoot string) (string, error) {
	out, err := runGitCommand(repoRoot, "rev-parse", "--short", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}
	return out, nil
}

// RepoID returns a consistent identifier for the repository
// Uses remote URL if available (cross-machine consistent), falls back to path hash
func RepoID(repoRoot string) string {
	// Try to get remote URL for cross-machine consistency
	cmd := exec.Command("git", "-C", repoRoot, "remote", "get-url", "origin")
	if out, err := cmd.Output(); err == nil && len(out) > 0 {
		// CRITICAL: trim whitespace before hashing to avoid newline bugs
		remote := strings.TrimSpace(string(out))
		sum := sha256.Sum256([]byte(remote))
		return fmt.Sprintf("%x", sum)[:16]
	}
	// Fallback: hash repo root path (local only)
	sum := sha256.Sum256([]byte(repoRoot))
	return fmt.Sprintf("%x", sum)[:16]
}

// runGitCommand runs a git command with -C repoRoot for consistency
// All git operations should use this helper to avoid CWD dependencies
func runGitCommand(repoRoot string, args ...string) (string, error) {
	cmdArgs := append([]string{"-C", repoRoot}, args...)
	cmd := exec.Command("git", cmdArgs...)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git %s failed: %w", args[0], err)
	}
	return strings.TrimSpace(out.String()), nil
}
