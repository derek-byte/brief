package render

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/derek-byte/coding-tools/internal/git"
	"github.com/derek-byte/coding-tools/internal/store"
)

// Brief contains all data needed to render a branch rehydration summary
type Brief struct {
	Branch      string
	LastUpdated time.Time
	GitState    git.GitState
	Events      []store.Event
}

// RenderBrief formats a branch brief for terminal output
// Output contract: fit in one screen, immediate orientation in < 60 seconds
func RenderBrief(b Brief) string {
	var out strings.Builder

	// Minimal header (no counts)
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

	out.WriteString(fmt.Sprintf("%s · %s · %s · %s\n",
		b.Branch,
		b.LastUpdated.Format("15:04"),
		lastCommit,
		stateInfo,
	))
	out.WriteString("\n")

	eventsByType := groupEventsByType(b.Events)

	// Core sections in order: Todos, Commands, Choices, Notes
	// All sorted oldest-first for chronological flow

	// Todos (oldest first, limit 7)
	if todos, ok := eventsByType["todo"]; ok && len(todos) > 0 {
		out.WriteString("Todos\n")
		writeSection(&out, todos, false, 7, false)
		out.WriteString("\n")
	}

	// Commands (oldest first, limit 7)
	if cmds, ok := eventsByType["cmd"]; ok && len(cmds) > 0 {
		out.WriteString("Commands\n")
		writeSection(&out, cmds, false, 7, false)
		out.WriteString("\n")
	}

	// Choices (oldest first, limit 7) - includes old "decision" entries
	choices := eventsByType["choice"]
	if oldDecisions, ok := eventsByType["decision"]; ok {
		choices = append(choices, oldDecisions...)
	}
	if len(choices) > 0 {
		out.WriteString("Choices\n")
		writeSection(&out, choices, false, 7, false)
		out.WriteString("\n")
	}

	// Notes (oldest first, limit 7)
	if notes, ok := eventsByType["note"]; ok && len(notes) > 0 {
		out.WriteString("Notes\n")
		writeSection(&out, notes, false, 7, false)
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
		// Newest first (choices, commands, errors)
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
			// Check completion state
			if event.CompletedAt != nil && *event.CompletedAt > 0 {
				bullet = "✓"
			} else {
				bullet = "☐"
			}
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

// RenderTimeline formats events as a chronological activity log
// Best for "what happened?" - shows all events with timestamps
func RenderTimeline(b Brief) string {
	var out strings.Builder

	// Minimal header (same as structured mode)
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

	out.WriteString(fmt.Sprintf("%s · %s · %s · %s\n\n",
		b.Branch,
		b.LastUpdated.Format("15:04"),
		lastCommit,
		stateInfo,
	))

	// Build combined list of all events including goal
	type timelineEntry struct {
		timestamp int64
		eventType string
		text      string
	}

	var entries []timelineEntry

	// Add all events (skip goal type - not shown)
	for _, event := range b.Events {
		if event.Type == "goal" {
			continue
		}
		entries = append(entries, timelineEntry{
			timestamp: event.CreatedAt,
			eventType: event.Type,
			text:      event.Text,
		})
	}

	// Sort by timestamp (newest first - it's a log)
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].timestamp > entries[j].timestamp
	})

	// Render each entry
	if len(entries) == 0 {
		out.WriteString("No activity yet for this branch\n")
	} else {
		for _, entry := range entries {
			timestamp := time.Unix(entry.timestamp, 0).Format("15:04")
			text := truncateLine(entry.text, 140)
			out.WriteString(fmt.Sprintf("%s [%s] %s\n", timestamp, entry.eventType, text))
		}
	}

	return out.String()
}
