package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/derek-byte/coding-tools/notetaker-dev/internal/store"
)

// ViewFormat determines how items are displayed
type ViewFormat int

const (
	FormatWide    ViewFormat = iota // Format A: grouped with wide labels
	FormatCompact                    // Format B: continuous with compact labels
)

// Model is the Bubble Tea model for the TUI
type Model struct {
	items      []Item
	goal       *Item // Single goal displayed as header
	cursor     int
	format     ViewFormat
	showTime   bool
	showHelp   bool
	width      int
	height     int
	branch     string
	quitting   bool
}

// Item represents a displayable event in the list
type Item struct {
	Event       store.Event
	DisplayText string
	TypeLabel   string
	Prefix      string // e.g., "☐ " for todo, "$ " for cmd
	Suffix      string // e.g., "(x2)" for deduped cmd, "(+8)" for multiline error
}

// NewModel creates a new TUI model with the given events
func NewModel(events []store.Event, branch string) Model {
	items, goal := buildItems(events)

	// Sort items to match wide format display order initially
	// (grouped by type in priority order, then by timestamp within each type)
	sortItemsForWideFormat(items)

	return Model{
		items:    items,
		goal:     goal,
		cursor:   0,
		format:   FormatWide,
		showTime: false, // Off by default, but toggleable
		showHelp: false,
		branch:   branch,
	}
}

// Init initializes the model (required by Bubble Tea)
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles key presses and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "?":
			m.showHelp = !m.showHelp
			return m, nil

		case "v":
			// Toggle view format
			if m.format == FormatWide {
				m.format = FormatCompact
				sortItemsForCompactFormat(m.items)
			} else {
				m.format = FormatWide
				sortItemsForWideFormat(m.items)
			}
			// Reset cursor to top when switching views
			m.cursor = 0
			return m, nil

		case "j", "down":
			// Move down, skipping blank lines in wide format
			m.cursor = m.nextSelectableIndex(m.cursor, 1)
			return m, nil

		case "k", "up":
			// Move up, skipping blank lines in wide format
			m.cursor = m.nextSelectableIndex(m.cursor, -1)
			return m, nil

		case "g":
			// Go to top
			m.cursor = 0
			return m, nil

		case "G":
			// Go to bottom
			m.cursor = len(m.items) - 1
			return m, nil
		}
	}

	return m, nil
}

// nextSelectableIndex finds the next selectable item index
// In wide format, we skip blank separator lines
func (m Model) nextSelectableIndex(current, delta int) int {
	if len(m.items) == 0 {
		return 0
	}

	next := current + delta

	// Clamp to valid range
	if next < 0 {
		return 0
	}
	if next >= len(m.items) {
		return len(m.items) - 1
	}

	return next
}

// View renders the TUI
func (m Model) View() string {
	if m.quitting {
		return ""
	}

	var out strings.Builder

	// Render goal as bold header if present
	if m.goal != nil {
		out.WriteString(m.renderGoalHeader())
		out.WriteString("\n\n")
	}

	// Render items based on format
	if m.format == FormatWide {
		out.WriteString(m.renderWideFormat())
	} else {
		out.WriteString(m.renderCompactFormat())
	}

	// Status line
	out.WriteString("\n")
	out.WriteString(m.renderStatusLine())

	// Help footer (if toggled)
	if m.showHelp {
		out.WriteString("\n")
		out.WriteString(m.renderHelp())
	}

	return out.String()
}

// renderGoalHeader renders the goal as a bold header
func (m Model) renderGoalHeader() string {
	goalStyle := lipgloss.NewStyle().Bold(true)
	return goalStyle.Render("Goal: " + m.goal.DisplayText)
}

// renderStatusLine shows current mode and hints
func (m Model) renderStatusLine() string {
	formatName := "wide"
	if m.format == FormatCompact {
		formatName = "compact"
	}

	// Use terminal width for separator line, default to 80 if not set
	width := m.width
	if width == 0 {
		width = 80
	}
	separator := strings.Repeat("─", width)

	return fmt.Sprintf("%s\n"+
		"%s · %s · %d items · ?=help v=view q=quit",
		separator, m.branch, formatName, len(m.items))
}

// renderHelp shows keybindings
func (m Model) renderHelp() string {
	return `
Navigation: j/k or ↑↓ (move), g/G (top/bottom)
View:       v (toggle format)
Help:       ? (toggle this help)
Quit:       q
`
}

// buildItems converts events into displayable items with formatting
// Returns items list and optional goal (displayed separately as header)
func buildItems(events []store.Event) ([]Item, *Item) {
	var items []Item
	var goal *Item

	for _, e := range events {
		// Skip stash type (not shown in UI list)
		if e.Type == "stash" {
			continue
		}

		// For goals, keep only the most recent one (displayed as header)
		if e.Type == "goal" {
			if goal == nil || e.CreatedAt > goal.Event.CreatedAt {
				goalItem := Item{
					Event:       e,
					DisplayText: e.Text,
					TypeLabel:   getDisplayLabel(e.Type),
				}
				goal = &goalItem
			}
			continue
		}

		item := Item{
			Event:       e,
			DisplayText: e.Text,
			TypeLabel:   getDisplayLabel(e.Type),
		}

		// Add prefix based on type
		switch e.Type {
		case "todo":
			item.Prefix = "☐ "
		case "cmd":
			item.Prefix = "$ "
		}

		items = append(items, item)
	}

	return items, goal
}

// formatTimestamp formats Unix timestamp for display
func formatTimestamp(ts int64) string {
	return time.Unix(ts, 0).Format("15:04")
}

// sortItemsForWideFormat sorts items by type priority, then by timestamp within each type
func sortItemsForWideFormat(items []Item) {
	typePriority := map[string]int{
		"todo":   0,
		"choice": 1,
		"cmd":    2,
		"note":   3,
	}

	sort.Slice(items, func(i, j int) bool {
		// First, sort by type priority
		pi := typePriority[items[i].Event.Type]
		pj := typePriority[items[j].Event.Type]
		if pi != pj {
			return pi < pj
		}

		// Within same type, sort by timestamp (newest first)
		return items[i].Event.CreatedAt > items[j].Event.CreatedAt
	})
}

// sortItemsForCompactFormat sorts items chronologically (newest first)
func sortItemsForCompactFormat(items []Item) {
	sort.Slice(items, func(i, j int) bool {
		return items[i].Event.CreatedAt > items[j].Event.CreatedAt
	})
}

// getDisplayLabel returns the display label for a type
// Maps old "decision" entries to "choice" for display
func getDisplayLabel(eventType string) string {
	if eventType == "decision" {
		return "choice"
	}
	return eventType
}
