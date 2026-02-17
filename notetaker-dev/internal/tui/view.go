package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Styles
var (
	selectedStyle = lipgloss.NewStyle().Reverse(true)
	dimStyle      = lipgloss.NewStyle().Faint(true)
)

// renderWideFormat renders Format A: grouped by type with wide labels
func (m Model) renderWideFormat() string {
	var out strings.Builder

	// Group items by type in priority order
	typeOrder := []string{"todo", "choice", "cmd", "note"}
	grouped := make(map[string][]Item)

	for _, item := range m.items {
		grouped[item.Event.Type] = append(grouped[item.Event.Type], item)
	}

	first := true
	for _, eventType := range typeOrder {
		items, ok := grouped[eventType]
		if !ok || len(items) == 0 {
			continue
		}

		// Add blank line between groups (but not before first group)
		if !first {
			out.WriteString("\n")
		}
		first = false

		// Render items in this group
		for _, item := range items {
			out.WriteString(m.renderWideItem(item))
			out.WriteString("\n")
		}
	}

	return out.String()
}

// renderWideItem renders a single item in wide format
func (m Model) renderWideItem(item Item) string {
	var parts []string

	// Timestamp (if enabled)
	if m.showTime {
		ts := formatTimestamp(item.Event.CreatedAt)
		parts = append(parts, ts)
	}

	// Type label with tight brackets + padding for alignment
	// Longest label is "choice" (6 chars) → [choice] is 8 chars
	typeLabel := fmt.Sprintf("[%s]", item.TypeLabel)
	labelWidth := len(typeLabel)
	padding := strings.Repeat(" ", 8-labelWidth)
	parts = append(parts, typeLabel+padding)

	// Prefix + text + suffix
	text := item.Prefix + item.DisplayText + item.Suffix
	parts = append(parts, text)

	line := strings.Join(parts, " ")

	// Check if this item is selected
	if m.isItemSelected(item) {
		return selectedStyle.Render(line)
	}

	// Dim completed todos
	if item.Event.Type == "todo" && item.Event.CompletedAt != nil && *item.Event.CompletedAt > 0 {
		return dimStyle.Render(line)
	}

	return line
}

// renderCompactFormat renders Format B: continuous list with compact labels
func (m Model) renderCompactFormat() string {
	var out strings.Builder

	for _, item := range m.items {
		out.WriteString(m.renderCompactItem(item))
		out.WriteString("\n")
	}

	return out.String()
}

// renderCompactItem renders a single item in compact format
func (m Model) renderCompactItem(item Item) string {
	var parts []string

	// Timestamp (if enabled)
	if m.showTime {
		ts := formatTimestamp(item.Event.CreatedAt)
		parts = append(parts, ts)
	}

	// Type label with tight brackets + padding for alignment
	// Use same labels as wide format for consistency
	typeLabel := fmt.Sprintf("[%s]", item.TypeLabel)
	labelWidth := len(typeLabel)
	paddingWidth := 8 - labelWidth // Max label is [choice] = 8 chars
	if paddingWidth < 0 {
		paddingWidth = 0 // Defensive: prevent panic
	}
	padding := strings.Repeat(" ", paddingWidth)
	parts = append(parts, typeLabel+padding)

	// Prefix + text + suffix
	text := item.Prefix + item.DisplayText + item.Suffix
	parts = append(parts, text)

	line := strings.Join(parts, " ")

	// Check if this item is selected
	if m.isItemSelected(item) {
		return selectedStyle.Render(line)
	}

	// Dim completed todos
	if item.Event.Type == "todo" && item.Event.CompletedAt != nil && *item.Event.CompletedAt > 0 {
		return dimStyle.Render(line)
	}

	return line
}

// isItemSelected checks if the given item is currently selected
func (m Model) isItemSelected(item Item) bool {
	if m.cursor < 0 || m.cursor >= len(m.items) {
		return false
	}
	return m.items[m.cursor].Event.ID == item.Event.ID
}

// getCompactLabel returns the compact version of a type label (currently unused)
func getCompactLabel(eventType string) string {
	return eventType // All types use full labels now
}
