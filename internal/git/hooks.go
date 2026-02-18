package git

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	hookStartMarker = "# >>> branchbrief auto-context >>>"
	hookEndMarker   = "# <<< branchbrief auto-context <<<"
)

// excludeHookIfNeeded adds hook to .git/info/exclude if it's in the working tree
func excludeHookIfNeeded(repoRoot, hookPath string) error {
	// Check if hook is in .git/ directory (outside working tree)
	gitDir := filepath.Join(repoRoot, ".git")
	relPath, err := filepath.Rel(gitDir, hookPath)
	if err == nil && !strings.HasPrefix(relPath, "..") {
		// Hook is inside .git/, no need to exclude
		return nil
	}

	// Hook is in working tree, add to .git/info/exclude
	excludePath := filepath.Join(repoRoot, ".git", "info", "exclude")

	// Read existing exclude file
	existingContent := ""
	if data, err := os.ReadFile(excludePath); err == nil {
		existingContent = string(data)
	}

	// Get relative path from repo root for exclude pattern
	relToRepo, err := filepath.Rel(repoRoot, hookPath)
	if err != nil {
		return err
	}

	// Check if already excluded
	if strings.Contains(existingContent, relToRepo) {
		return nil
	}

	// Append to exclude file
	newContent := existingContent
	if !strings.HasSuffix(newContent, "\n") && newContent != "" {
		newContent += "\n"
	}
	newContent += relToRepo + "\n"

	// Ensure .git/info directory exists
	if err := os.MkdirAll(filepath.Dir(excludePath), 0755); err != nil {
		return err
	}

	// Write exclude file
	if err := os.WriteFile(excludePath, []byte(newContent), 0644); err != nil {
		return err
	}

	return nil
}

// getHooksDir returns the hooks directory, respecting core.hooksPath if set
func getHooksDir(repoRoot string) (string, error) {
	// Check if core.hooksPath is configured
	cmd := exec.Command("git", "-C", repoRoot, "config", "--get", "core.hooksPath")
	if out, err := cmd.Output(); err == nil && len(out) > 0 {
		hooksPath := strings.TrimSpace(string(out))

		// If relative path, make it relative to repo root
		if !filepath.IsAbs(hooksPath) {
			return filepath.Join(repoRoot, hooksPath), nil
		}
		return hooksPath, nil
	}

	// Default to .git/hooks
	return filepath.Join(repoRoot, ".git", "hooks"), nil
}

// InstallHook installs a post-checkout hook that runs brief on branch switch
func InstallHook(repoRoot string, fullMode bool) error {
	hooksDir, err := getHooksDir(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to get hooks directory: %w", err)
	}
	hookPath := filepath.Join(hooksDir, "post-checkout")

	// Read existing hook if it exists
	existingContent := ""
	if data, err := os.ReadFile(hookPath); err == nil {
		existingContent = string(data)
	}

	// Check if already installed
	if strings.Contains(existingContent, hookStartMarker) {
		return fmt.Errorf("hook already installed (run with --uninstall to remove first)")
	}

	// Build hook content
	command := "brief summary --quiet"
	if fullMode {
		command = "brief rehydrate"
	}

	hookBlock := fmt.Sprintf(`
%s
if [ "$3" = "1" ]; then
  %s 2>/dev/null || true
fi
%s
`, hookStartMarker, command, hookEndMarker)

	// Append to existing content or create new
	newContent := existingContent
	if newContent == "" {
		newContent = "#!/bin/sh\n"
	}
	newContent += hookBlock

	// Write hook file
	if err := os.WriteFile(hookPath, []byte(newContent), 0755); err != nil {
		return fmt.Errorf("failed to write hook: %w", err)
	}

	// If hook is in working tree (not .git/hooks), add to .git/info/exclude
	if err := excludeHookIfNeeded(repoRoot, hookPath); err != nil {
		return fmt.Errorf("failed to exclude hook from git: %w", err)
	}

	return nil
}

// UninstallHook removes the branchbrief block from post-checkout hook
func UninstallHook(repoRoot string) error {
	hooksDir, err := getHooksDir(repoRoot)
	if err != nil {
		return fmt.Errorf("failed to get hooks directory: %w", err)
	}
	hookPath := filepath.Join(hooksDir, "post-checkout")

	// Read existing hook
	data, err := os.ReadFile(hookPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no hook installed")
		}
		return fmt.Errorf("failed to read hook: %w", err)
	}

	content := string(data)

	// Check if hook is installed
	if !strings.Contains(content, hookStartMarker) {
		return fmt.Errorf("branchbrief hook not found")
	}

	// Remove the block between markers
	startIdx := strings.Index(content, hookStartMarker)
	endIdx := strings.Index(content, hookEndMarker)
	if startIdx == -1 || endIdx == -1 {
		return fmt.Errorf("malformed hook markers")
	}

	// Find the start of the line containing the start marker
	lineStart := startIdx
	for lineStart > 0 && content[lineStart-1] != '\n' {
		lineStart--
	}

	// Find the end of the line containing the end marker
	lineEnd := endIdx + len(hookEndMarker)
	if lineEnd < len(content) && content[lineEnd] == '\n' {
		lineEnd++
	}

	// Remove the block
	newContent := content[:lineStart] + content[lineEnd:]

	// If file is now empty or just shebang, remove it
	trimmed := strings.TrimSpace(newContent)
	if trimmed == "" || trimmed == "#!/bin/sh" {
		if err := os.Remove(hookPath); err != nil {
			return fmt.Errorf("failed to remove hook file: %w", err)
		}
		return nil
	}

	// Write updated content
	if err := os.WriteFile(hookPath, []byte(newContent), 0755); err != nil {
		return fmt.Errorf("failed to update hook: %w", err)
	}

	return nil
}
