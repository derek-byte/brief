package git

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	hookStartMarker = "# >>> branchbrief auto-context >>>"
	hookEndMarker   = "# <<< branchbrief auto-context <<<"
)

// InstallHook installs a post-checkout hook that runs brief on branch switch
func InstallHook(repoRoot string, fullMode bool) error {
	hookPath := filepath.Join(repoRoot, ".git", "hooks", "post-checkout")

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

	return nil
}

// UninstallHook removes the branchbrief block from post-checkout hook
func UninstallHook(repoRoot string) error {
	hookPath := filepath.Join(repoRoot, ".git", "hooks", "post-checkout")

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
