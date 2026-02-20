package render

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/derek-byte/coding-tools/internal/store"
	"github.com/mattn/go-runewidth"
	"golang.org/x/term"
)

// RenderSummary renders a compact grid-based summary
func RenderSummary(b Brief) string {
	// Get terminal width
	width := getTerminalWidth()

	var out strings.Builder

	// Separator line with spacing
	separator := strings.Repeat("─", width)
	out.WriteString(separator)
	out.WriteString("\n\n")

	// Branch name as context header - bold
	if b.Branch != "" {
		boldStyle := lipgloss.NewStyle().Bold(true)
		out.WriteString(boldStyle.Render(b.Branch))
		out.WriteString("\n\n")
	}

	// Group events
	eventsByType := groupEventsByType(b.Events)
	todos := eventsByType["todo"]

	// Combine choices with old decisions
	choices := eventsByType["choice"]
	if oldDecisions, ok := eventsByType["decision"]; ok {
		choices = append(choices, oldDecisions...)
	}

	notes := eventsByType["note"]

	// Render grid based on width
	if width >= 110 {
		// 3 columns: Todos, Choices, Notes
		out.WriteString(renderThreeColumnGrid(todos, choices, notes))
	} else if width >= 80 {
		// 2 columns: Todos, Choices
		out.WriteString(renderTwoColumnGrid(todos, choices))
		// Notes below if any
		if len(notes) > 0 {
			out.WriteString("\n")
			out.WriteString(renderNotesStacked(notes, 3))
		}
	} else {
		// Stacked rows
		out.WriteString(renderStackedRows(todos, choices, notes))
	}

	return out.String()
}

// getTerminalWidth returns terminal width, defaulting to 80
func getTerminalWidth() int {
	fd := int(os.Stdout.Fd())
	if width, _, err := term.GetSize(fd); err == nil && width > 0 {
		return width
	}
	return 80
}

// renderThreeColumnGrid renders todos, choices, notes in 3 columns
func renderThreeColumnGrid(todos, choices, notes []store.Event) string {
	const maxItems = 3
	const colWidth = 35

	// Prepare columns
	todoLines := formatColumn(todos, "todo", maxItems, colWidth)
	choiceLines := formatColumn(choices, "choice", maxItems, colWidth)
	noteLines := formatColumn(notes, "note", maxItems, colWidth)

	// Find max rows
	maxRows := max(max(len(todoLines), len(choiceLines)), len(noteLines))

	// Render rows (don't pad empty lines)
	var out strings.Builder
	for i := 0; i < maxRows; i++ {
		// Get content for each column (or empty string if beyond column length)
		todoContent := ""
		if i < len(todoLines) {
			todoContent = todoLines[i]
		}
		choiceContent := ""
		if i < len(choiceLines) {
			choiceContent = choiceLines[i]
		}
		noteContent := ""
		if i < len(noteLines) {
			noteContent = noteLines[i]
		}

		// Pad todo and choice columns for alignment (only if non-empty)
		if todoContent != "" {
			out.WriteString(padRight(todoContent, colWidth))
		} else {
			out.WriteString(strings.Repeat(" ", colWidth))
		}
		out.WriteString("  ")
		if choiceContent != "" {
			out.WriteString(padRight(choiceContent, colWidth))
		} else {
			out.WriteString(strings.Repeat(" ", colWidth))
		}
		out.WriteString("  ")
		out.WriteString(noteContent)
		out.WriteString("\n")
	}

	return out.String()
}

// renderTwoColumnGrid renders todos and choices in 2 columns
func renderTwoColumnGrid(todos, choices []store.Event) string {
	const maxItems = 3
	const colWidth = 38

	todoLines := formatColumn(todos, "todo", maxItems, colWidth)
	choiceLines := formatColumn(choices, "choice", maxItems, colWidth)

	maxRows := max(len(todoLines), len(choiceLines))

	var out strings.Builder
	for i := 0; i < maxRows; i++ {
		// Get content for each column (or empty if beyond length)
		todoContent := ""
		if i < len(todoLines) {
			todoContent = todoLines[i]
		}
		choiceContent := ""
		if i < len(choiceLines) {
			choiceContent = choiceLines[i]
		}

		out.WriteString(padRight(todoContent, colWidth))
		out.WriteString("  ")
		out.WriteString(padRight(choiceContent, colWidth))
		out.WriteString("\n")
	}

	return out.String()
}

// renderStackedRows renders todos, choices, notes as compact rows
func renderStackedRows(todos, choices, notes []store.Event) string {
	var out strings.Builder

	if len(todos) > 0 {
		out.WriteString("Todos: ")
		items := formatInlineList(todos, "todo", 3)
		out.WriteString(strings.Join(items, " · "))
		out.WriteString("\n")
	}

	if len(choices) > 0 {
		out.WriteString("Choices: ")
		items := formatInlineList(choices, "choice", 3)
		out.WriteString(strings.Join(items, " · "))
		out.WriteString("\n")
	}

	if len(notes) > 0 {
		out.WriteString("Notes: ")
		items := formatInlineList(notes, "note", 3)
		out.WriteString(strings.Join(items, " · "))
		out.WriteString("\n")
	}

	return out.String()
}

// renderNotesStacked renders notes as a simple list
func renderNotesStacked(notes []store.Event, maxItems int) string {
	var out strings.Builder
	out.WriteString("Notes\n")

	hasMore := len(notes) > maxItems
	count := 0
	for _, note := range notes {
		if count >= maxItems {
			break
		}
		text := truncateLine(note.Text, 68) // Account for indent
		out.WriteString(fmt.Sprintf("  • %s\n", text)) // 2-space indent
		count++
	}

	// Add "..." if more notes exist
	if hasMore {
		out.WriteString("  ...\n") // 2-space indent
	}

	return out.String()
}

// formatColumn formats events as column lines with header
func formatColumn(events []store.Event, eventType string, maxItems, width int) []string {
	var lines []string

	if len(events) == 0 {
		return lines
	}

	// Header
	header := ""
	switch eventType {
	case "todo":
		header = "Todos (Incomplete)"
	case "choice":
		header = "Choices"
	case "note":
		header = "Notes"
	}
	lines = append(lines, padRight(header, width))

	// Items sorting
	sorted := make([]store.Event, len(events))
	copy(sorted, events)

	if eventType == "todo" {
		// Todos: filter to incomplete only
		var incomplete []store.Event
		for _, e := range sorted {
			if e.CompletedAt == nil || *e.CompletedAt == 0 {
				incomplete = append(incomplete, e)
			}
		}

		// If no incomplete todos, show message
		if len(incomplete) == 0 {
			lines = append(lines, padRight("✓ All todos completed", width))
			return lines
		}

		// Sort incomplete todos oldest first
		sort.Slice(incomplete, func(i, j int) bool {
			return incomplete[i].CreatedAt < incomplete[j].CreatedAt
		})
		sorted = incomplete
	} else {
		// Others: newest first
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CreatedAt > sorted[j].CreatedAt
		})
	}

	count := 0
	hasMore := len(sorted) > maxItems

	for _, e := range sorted {
		if count >= maxItems {
			break
		}

		bullet := "•"
		if eventType == "todo" {
			bullet = "☐" // Always incomplete now
		}

		text := truncateLine(e.Text, width-5) // Account for indent + bullet
		line := fmt.Sprintf("  %s %s", bullet, text) // 2-space indent
		lines = append(lines, padRight(line, width))
		count++
	}

	// Add "..." if there are more items
	if hasMore && eventType != "todo" {
		lines = append(lines, padRight("  ...", width)) // 2-space indent
	}

	return lines
}

// formatInlineList formats events as inline "prefix text" strings
func formatInlineList(events []store.Event, eventType string, maxItems int) []string {
	var items []string

	sorted := make([]store.Event, len(events))
	copy(sorted, events)

	if eventType == "todo" {
		// Filter to incomplete only
		var incomplete []store.Event
		for _, e := range sorted {
			if e.CompletedAt == nil || *e.CompletedAt == 0 {
				incomplete = append(incomplete, e)
			}
		}

		// If all completed, return special message
		if len(incomplete) == 0 {
			return []string{"✓ All completed"}
		}

		// Sort oldest first
		sort.Slice(incomplete, func(i, j int) bool {
			return incomplete[i].CreatedAt < incomplete[j].CreatedAt
		})
		sorted = incomplete
	} else {
		// Others: newest first
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].CreatedAt > sorted[j].CreatedAt
		})
	}

	hasMore := len(sorted) > maxItems
	count := 0

	for _, e := range sorted {
		if count >= maxItems {
			break
		}

		bullet := ""
		if eventType == "todo" {
			bullet = "☐" // Always incomplete now
		}

		text := truncateLine(e.Text, 50)
		if bullet != "" {
			items = append(items, fmt.Sprintf("%s %s", bullet, text))
		} else {
			items = append(items, text)
		}
		count++
	}

	// Add "..." if more items exist
	if hasMore && eventType != "todo" {
		items = append(items, "...")
	}

	return items
}

// padRight pads string to width with spaces, accounting for display width of wide characters
func padRight(s string, width int) string {
	displayWidth := runewidth.StringWidth(s)
	if displayWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-displayWidth)
}

// max returns the maximum of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
