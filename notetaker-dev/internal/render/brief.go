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

	// Compact header with counts
	eventsByType := groupEventsByType(b.Events)
	counts := buildCountsString(eventsByType, b.CurrentGoal)

	stateInfo := "clean"
	if b.GitState.DirtyFileCount > 0 {
		stateInfo = fmt.Sprintf("%d changed", b.GitState.DirtyFileCount)
	}

	lastCommit := "no commits"
	if b.GitState.LastCommit != "" {
		parts := strings.SplitN(b.GitState.LastCommit, " ", 2)
		if len(parts) > 0 {
			lastCommit = parts[0]
		}
	}

	out.WriteString(fmt.Sprintf("%s · %s · %s · %s%s\n\n",
		b.Branch,
		b.LastUpdated.Format("15:04"),
		lastCommit,
		stateInfo,
		counts,
	))

	// Goal section (single goal per branch)
	if b.CurrentGoal != nil {
		out.WriteString("Goal\n")
		text := truncateLine(b.CurrentGoal.Text, 160)
		out.WriteString(fmt.Sprintf("• %s\n\n", text))
	}

	// Decisions section (newest first, limit 7)
	if decisions, ok := eventsByType["decision"]; ok && len(decisions) > 0 {
		out.WriteString("Decisions\n")
		writeSection(&out, decisions, true, 7, false)
		out.WriteString("\n")
	}

	// Known issues / errors section (newest first, limit 7, multi-line)
	if errors, ok := eventsByType["error"]; ok && len(errors) > 0 {
		out.WriteString("Known issues\n")
		writeSection(&out, errors, true, 7, true)
		out.WriteString("\n")
	}

	// Fixes section (newest first, limit 7)
	if fixes, ok := eventsByType["fix"]; ok && len(fixes) > 0 {
		out.WriteString("Fixes\n")
		writeSection(&out, fixes, true, 7, false)
		out.WriteString("\n")
	}

	// Next steps / todos section (oldest first for natural checklist order, limit 7)
	if todos, ok := eventsByType["todo"]; ok && len(todos) > 0 {
		out.WriteString("Next steps\n")
		writeSection(&out, todos, false, 7, false)
		out.WriteString("\n")
	}

	// Commands section (newest first, limit 7)
	if cmds, ok := eventsByType["cmd"]; ok && len(cmds) > 0 {
		out.WriteString("Commands\n")
		writeSection(&out, cmds, true, 7, false)
		out.WriteString("\n")
	}

	// Links section (newest first, limit 7)
	if links, ok := eventsByType["link"]; ok && len(links) > 0 {
		out.WriteString("Links\n")
		writeSection(&out, links, true, 7, false)
		out.WriteString("\n")
	}

	// Issues section (newest first, limit 7)
	if issues, ok := eventsByType["issue"]; ok && len(issues) > 0 {
		out.WriteString("Issues\n")
		writeSection(&out, issues, true, 7, false)
		out.WriteString("\n")
	}

	// Notes section (newest first, limit 7)
	if notes, ok := eventsByType["note"]; ok && len(notes) > 0 {
		out.WriteString("Notes\n")
		writeSection(&out, notes, true, 7, false)
		out.WriteString("\n")
	}

	// Show message if no events
	if len(b.Events) == 0 && b.CurrentGoal == nil {
		out.WriteString("No notes yet for this branch\n")
	}

	return out.String()
}

// buildCountsString creates the " · 2 todos · 1 cmd" suffix
func buildCountsString(eventsByType map[string][]store.Event, currentGoal *store.Event) string {
	var parts []string

	// Count current goal as 1 if exists
	if currentGoal != nil {
		parts = append(parts, "1 goal")
	}

	// Order: decision, todo, cmd, error, fix, issue, link, note
	order := []string{"decision", "todo", "cmd", "error", "fix", "issue", "link", "note"}
	for _, t := range order {
		if events, ok := eventsByType[t]; ok && len(events) > 0 {
			count := len(events)
			label := t
			if count != 1 {
				label += "s"
			}
			parts = append(parts, fmt.Sprintf("%d %s", count, label))
		}
	}

	if len(parts) == 0 {
		return ""
	}
	return " · " + strings.Join(parts, " · ")
}

// groupEventsByType organizes events by their type
func groupEventsByType(events []store.Event) map[string][]store.Event {
	groups := make(map[string][]store.Event)
	for _, event := range events {
		// Skip goal type - handled separately as CurrentGoal
		if event.Type == "goal" {
			continue
		}
		// Skip stash type - not shown in rehydrate
		if event.Type == "stash" {
			continue
		}
		groups[event.Type] = append(groups[event.Type], event)
	}
	return groups
}

// writeSection formats a list of events as bullets
// newestFirst controls sort order (false for todos)
// maxBullets limits output (typically 7)
// multiline enables multi-line rendering with indentation (for errors)
func writeSection(out *strings.Builder, events []store.Event, newestFirst bool, maxBullets int, multiline bool) {
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
		bullet := "•"
		if event.Type == "todo" {
			bullet = "☐"
		} else if event.Type == "cmd" {
			bullet = "$"
		}

		if multiline && strings.Contains(event.Text, "\n") {
			// Multi-line rendering (for errors with stdin capture)
			lines := strings.Split(event.Text, "\n")
			firstLine := truncateLine(lines[0], 160)
			out.WriteString(fmt.Sprintf("%s %s\n", bullet, firstLine))

			// Show up to 8 additional lines, indented
			remaining := lines[1:]
			displayed := 0
			for _, line := range remaining {
				if displayed >= 8 {
					if len(remaining) > 8 {
						out.WriteString(fmt.Sprintf("  └─ (+%d more lines)\n", len(remaining)-8))
					}
					break
				}
				line = strings.TrimSpace(line)
				if line != "" {
					truncated := truncateLine(line, 150)
					out.WriteString(fmt.Sprintf("  └─ %s\n", truncated))
					displayed++
				}
			}
		} else {
			// Single-line rendering
			text := truncateLine(event.Text, 160)
			out.WriteString(fmt.Sprintf("%s %s\n", bullet, text))
		}
	}
}

// truncateLine limits line length and adds ellipsis if needed
func truncateLine(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
