package render

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/derek/branchbrief/internal/git"
	"github.com/derek/branchbrief/internal/store"
)

// Brief contains all data needed to render a branch rehydration summary
type Brief struct {
	Branch      string
	LastUpdated time.Time
	GitState    git.GitState
	Events      []store.Event
	CurrentGoal *store.Event // Single goal per branch (nil if no goal)
}

// RenderBrief formats a branch brief for terminal output
// Output contract: fit in one screen, immediate orientation in < 60 seconds
func RenderBrief(b Brief) string {
	var out strings.Builder

	// Header
	out.WriteString(fmt.Sprintf("Branch: %s\n", b.Branch))
	out.WriteString(fmt.Sprintf("Last updated: %s\n\n", b.LastUpdated.Format("2006-01-02 15:04")))

	// Group events by type (excluding goal - handled separately)
	eventsByType := groupEventsByType(b.Events)

	// Goal section (single goal per branch)
	if b.CurrentGoal != nil {
		out.WriteString("Goal\n")
		text := truncateLine(b.CurrentGoal.Text, 160)
		out.WriteString(fmt.Sprintf("• %s\n\n", text))
	}

	// State section
	out.WriteString("State\n")
	writeGitState(&out, b.GitState)
	out.WriteString("\n")

	// Decisions section (newest first, limit 7)
	if decisions, ok := eventsByType["decision"]; ok && len(decisions) > 0 {
		out.WriteString("Decisions\n")
		writeSection(&out, decisions, true, 7)
		out.WriteString("\n")
	}

	// Known issues / errors section (newest first, limit 7)
	if errors, ok := eventsByType["error"]; ok && len(errors) > 0 {
		out.WriteString("Known issues\n")
		writeSection(&out, errors, true, 7)
		out.WriteString("\n")
	}

	// Next steps / todos section (oldest first for natural checklist order, limit 7)
	if todos, ok := eventsByType["todo"]; ok && len(todos) > 0 {
		out.WriteString("Next steps\n")
		writeSection(&out, todos, false, 7)
		out.WriteString("\n")
	}

	// Commands section (newest first, limit 7)
	if cmds, ok := eventsByType["cmd"]; ok && len(cmds) > 0 {
		out.WriteString("Commands\n")
		writeSection(&out, cmds, true, 7)
		out.WriteString("\n")
	}

	// Links section (newest first, limit 7)
	if links, ok := eventsByType["link"]; ok && len(links) > 0 {
		out.WriteString("Links\n")
		writeSection(&out, links, true, 7)
		out.WriteString("\n")
	}

	// Show message if no events
	if len(b.Events) == 0 {
		out.WriteString("No notes yet for this branch\n")
	}

	return out.String()
}

// groupEventsByType organizes events by their type
func groupEventsByType(events []store.Event) map[string][]store.Event {
	groups := make(map[string][]store.Event)
	for _, event := range events {
		groups[event.Type] = append(groups[event.Type], event)
	}
	return groups
}

// writeSection formats a list of events as bullets
// newestFirst controls sort order (false for todos)
// maxBullets limits output (typically 7)
func writeSection(out *strings.Builder, events []store.Event, newestFirst bool, maxBullets int) {
	// Sort by timestamp
	sorted := make([]store.Event, len(events))
	copy(sorted, events)

	if newestFirst {
		// Newest first (decisions, commands, errors)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CreatedAt > sorted[j].CreatedAt
		})
	} else {
		// Oldest first (todos - natural checklist order)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CreatedAt < sorted[j].CreatedAt
		})
	}

	// Limit to maxBullets
	if len(sorted) > maxBullets {
		sorted = sorted[:maxBullets]
	}

	// Write bullets
	for _, event := range sorted {
		text := truncateLine(event.Text, 160)
		bullet := "•"
		if event.Type == "todo" {
			bullet = "☐"
		} else if event.Type == "cmd" {
			bullet = "$"
		}
		out.WriteString(fmt.Sprintf("%s %s\n", bullet, text))
	}
}

// writeGitState formats git repository state
func writeGitState(out *strings.Builder, state git.GitState) {
	if state.LastCommit != "" {
		out.WriteString(fmt.Sprintf("HEAD: %s\n", state.LastCommit))
	}

	if state.DirtyFileCount > 0 {
		plural := "s"
		if state.DirtyFileCount == 1 {
			plural = ""
		}
		out.WriteString(fmt.Sprintf("Working tree: %d file%s changed\n", state.DirtyFileCount, plural))
	} else {
		out.WriteString("Working tree: clean\n")
	}

	if state.DiffStat != "" {
		out.WriteString(fmt.Sprintf("Diffstat:\n%s\n", indent(state.DiffStat, "  ")))
	}

	// Show errors inline (don't hide failures)
	if len(state.Errors) > 0 {
		out.WriteString(fmt.Sprintf("Note: %s\n", strings.Join(state.Errors, ", ")))
	}
}

// truncateLine limits line length and adds ellipsis if needed
func truncateLine(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}

// indent adds a prefix to each line
func indent(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}
