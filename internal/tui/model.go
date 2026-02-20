package tui

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/derek-byte/coding-tools/internal/store"
)

// ViewFormat determines how items are displayed
type ViewFormat int

const (
	FormatWide    ViewFormat = iota // Format A: grouped with wide labels
	FormatCompact                   // Format B: continuous with compact labels
)

// Mode determines the current UI mode
type Mode int

const (
	ModeNormal Mode = iota // Normal list navigation
	ModeDetail             // Detail view for a single item
	ModeFilter             // Filter/search mode
)

// Model is the Bubble Tea model for the TUI
type Model struct {
	items          []Item
	cursor         int
	format         ViewFormat
	mode           Mode
	showTime       bool
	showHelp       bool
	width          int
	height         int
	branch         string
	repoID         string
	repoRoot       string
	db             *sql.DB
	quitting       bool
	detailItem     *Item  // Item shown in detail view
	pendingCommand string // Command to run after TUI exits
	pendingEdit    *Item  // Item to edit after TUI exits
	lastDeletedID  string // ID of last deleted item (for undo)
	filterQuery    string // Current filter query string
	allItems       []Item // Unfiltered items list
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
func NewModel(events []store.Event, branch, repoID, repoRoot string, db *sql.DB) Model {
	items := buildItems(events)

	// Sort items to match wide format display order initially
	// (grouped by type in priority order, then by timestamp within each type)
	sortItemsForWideFormat(items)

	return Model{
		items:    items,
		allItems: items, // Keep unfiltered copy
		cursor:   0,
		format:   FormatWide,
		mode:     ModeNormal,
		showTime: false, // Off by default, but toggleable
		showHelp: false,
		branch:   branch,
		repoID:   repoID,
		repoRoot: repoRoot,
		db:       db,
	}
}

// Init initializes the model (required by Bubble Tea)
func (m Model) Init() tea.Cmd {
	return nil
}

// GetPendingCommand returns the command to run after TUI exits, if any
func (m Model) GetPendingCommand() string {
	return m.pendingCommand
}

// GetPendingEdit returns the item to edit after TUI exits, if any
func (m Model) GetPendingEdit() *Item {
	return m.pendingEdit
}

// Update handles key presses and updates the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		// Detail mode key handling
		if m.mode == ModeDetail {
			switch msg.String() {
			case "esc":
				m.mode = ModeNormal
				m.detailItem = nil
				return m, nil
			case "q":
				m.quitting = true
				return m, tea.Quit
			}
			return m, nil
		}

		// Filter mode key handling
		if m.mode == ModeFilter {
			switch msg.String() {
			case "esc":
				// Clear filter and return to normal
				m.filterQuery = ""
				m.mode = ModeNormal
				m.applyFilter()
				return m, nil
			case "enter":
				// Keep filter active, return to normal
				m.mode = ModeNormal
				return m, nil
			case "backspace":
				// Remove last character
				if len(m.filterQuery) > 0 {
					m.filterQuery = m.filterQuery[:len(m.filterQuery)-1]
					m.applyFilter()
				}
				return m, nil
			default:
				// Add typed character to filter
				if len(msg.String()) == 1 {
					m.filterQuery += msg.String()
					m.applyFilter()
				}
				return m, nil
			}
		}

		// Normal mode key handling
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
			m.cursor = m.nextSelectableIndex(m.cursor, 1)
			return m, nil

		case "k", "up":
			m.cursor = m.nextSelectableIndex(m.cursor, -1)
			return m, nil

		case "g":
			m.cursor = 0
			return m, nil

		case "G":
			m.cursor = len(m.items) - 1
			return m, nil

		case "d":
			// Soft delete current item
			return m.handleDelete()

		case "u":
			// Undo last delete
			return m.handleUndo()

		case "e":
			// Edit current item in $EDITOR
			return m.handleEdit()

		case "/":
			// Enter filter mode
			m.mode = ModeFilter
			m.filterQuery = ""
			return m, nil

		case "enter":
			// Context-dependent Enter behavior
			return m.handleEnter()
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

	// Detail mode: show single item in full
	if m.mode == ModeDetail {
		return m.renderDetailView()
	}

	var out strings.Builder

	// Render goal as bold header if present
	// Show branch name as header
	if m.branch != "" {
		branchStyle := lipgloss.NewStyle().Bold(true)
		out.WriteString(branchStyle.Render(m.branch))
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

// renderStatusLine shows current mode and hints
func (m Model) renderStatusLine() string {
	// Use terminal width for separator line, default to 80 if not set
	width := m.width
	if width == 0 {
		width = 80
	}
	separator := strings.Repeat("─", width)

	var statusText string

	switch m.mode {
	case ModeFilter:
		statusText = fmt.Sprintf("FILTER: %s (esc=clear, enter=apply)", m.filterQuery)
	case ModeDetail:
		statusText = "DETAIL · esc=back q=quit"
	default:
		// Normal mode
		formatName := "wide"
		if m.format == FormatCompact {
			formatName = "compact"
		}
		filterHint := ""
		if m.filterQuery != "" {
			filterHint = fmt.Sprintf(" · filter:%s", m.filterQuery)
		}
		statusText = fmt.Sprintf("%s · %s · %d items%s · /=filter u=undo ?=help v=view q=quit",
			m.branch, formatName, len(m.items), filterHint)
	}

	return fmt.Sprintf("%s\n%s", separator, statusText)
}

// renderHelp shows keybindings
func (m Model) renderHelp() string {
	return `
Navigation: j/k or ↑↓ (move), g/G (top/bottom)
Actions:    Enter (toggle todo / run cmd / view), d (delete), e (edit)
View:       v (toggle format)
Help:       ? (toggle this help)
Quit:       q
`
}

// renderDetailView renders a single item in detail mode
func (m Model) renderDetailView() string {
	if m.detailItem == nil {
		return "No item selected\n\nPress ESC to return"
	}

	var out strings.Builder

	// Header with type label
	headerStyle := lipgloss.NewStyle().Bold(true)
	out.WriteString(headerStyle.Render(fmt.Sprintf("[%s]", m.detailItem.TypeLabel)))
	out.WriteString("\n\n")

	// Full text (no truncation)
	out.WriteString(m.detailItem.DisplayText)
	out.WriteString("\n\n")

	// Metadata footer
	ts := formatTimestamp(m.detailItem.Event.CreatedAt)
	out.WriteString(fmt.Sprintf("Created: %s\n", ts))

	if m.detailItem.Event.UpdatedAt != nil && *m.detailItem.Event.UpdatedAt > 0 {
		updatedTs := formatTimestamp(*m.detailItem.Event.UpdatedAt)
		out.WriteString(fmt.Sprintf("Updated: %s\n", updatedTs))
	}

	// Navigation hint
	out.WriteString("\n")
	width := m.width
	if width == 0 {
		width = 80
	}
	separator := strings.Repeat("─", width)
	out.WriteString(fmt.Sprintf("%s\nESC=back q=quit", separator))

	return out.String()
}

// buildItems converts events into displayable items with formatting
func buildItems(events []store.Event) []Item {
	var items []Item

	for _, e := range events {
		// Skip stash and goal types (not shown in UI list)
		if e.Type == "stash" || e.Type == "goal" {
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
			if e.CompletedAt != nil && *e.CompletedAt > 0 {
				item.Prefix = "✓ " // Completed
			} else {
				item.Prefix = "☐ " // Not completed
			}
		case "cmd":
			item.Prefix = "$ "
		}

		items = append(items, item)
	}

	return items
}

// formatTimestamp formats Unix timestamp for display
func formatTimestamp(ts int64) string {
	return time.Unix(ts, 0).Format("15:04")
}

// sortItemsForWideFormat sorts items by type priority, then by timestamp within each type
func sortItemsForWideFormat(items []Item) {
	typePriority := map[string]int{
		"todo":   0,
		"cmd":    1,
		"choice": 2,
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

// handleEnter performs context-dependent action based on item type
func (m Model) handleEnter() (tea.Model, tea.Cmd) {
	if len(m.items) == 0 || m.cursor >= len(m.items) {
		return m, nil
	}

	item := m.items[m.cursor]

	switch item.Event.Type {
	case "todo":
		// Toggle completion
		return m.handleTodoToggle()

	case "cmd":
		// Run command and exit
		return m.handleRunCommand()

	default:
		// Open detail view for other types (note, choice, fix)
		m.mode = ModeDetail
		m.detailItem = &item
		return m, nil
	}
}

// handleTodoToggle toggles the completion state of a todo
func (m Model) handleTodoToggle() (tea.Model, tea.Cmd) {
	if len(m.items) == 0 || m.cursor >= len(m.items) {
		return m, nil
	}

	item := &m.items[m.cursor]
	if item.Event.Type != "todo" {
		return m, nil
	}

	// Toggle in database
	err := store.ToggleTodoCompletion(m.db, item.Event.ID)
	if err != nil {
		// TODO: Show error to user
		return m, nil
	}

	// Update local state
	if item.Event.CompletedAt != nil && *item.Event.CompletedAt > 0 {
		// Was completed, now uncomplete
		item.Event.CompletedAt = nil
		item.Prefix = "☐ "
	} else {
		// Was uncompleted, now complete
		now := time.Now().Unix()
		item.Event.CompletedAt = &now
		item.Prefix = "✓ "
	}

	return m, nil
}

// handleRunCommand runs the selected command and exits
func (m Model) handleRunCommand() (tea.Model, tea.Cmd) {
	if len(m.items) == 0 || m.cursor >= len(m.items) {
		return m, nil
	}

	item := m.items[m.cursor]
	if item.Event.Type != "cmd" {
		return m, nil
	}

	// Store command to run after TUI exits
	m.pendingCommand = item.Event.Text
	m.quitting = true
	return m, tea.Quit
}

// handleDelete soft deletes the current item
func (m Model) handleDelete() (tea.Model, tea.Cmd) {
	if len(m.items) == 0 || m.cursor >= len(m.items) {
		return m, nil
	}

	item := m.items[m.cursor]

	// Soft delete in database
	err := store.SoftDeleteEvent(m.db, item.Event.ID)
	if err != nil {
		// TODO: Show error to user
		return m, nil
	}

	// Store ID for undo
	m.lastDeletedID = item.Event.ID

	// Remove from visible list
	m.items = append(m.items[:m.cursor], m.items[m.cursor+1:]...)

	// Remove from allItems too
	for i, it := range m.allItems {
		if it.Event.ID == item.Event.ID {
			m.allItems = append(m.allItems[:i], m.allItems[i+1:]...)
			break
		}
	}

	// Adjust cursor if at end
	if m.cursor >= len(m.items) && len(m.items) > 0 {
		m.cursor = len(m.items) - 1
	}

	return m, nil
}

// handleUndo restores the last deleted item
func (m Model) handleUndo() (tea.Model, tea.Cmd) {
	if m.lastDeletedID == "" {
		return m, nil
	}

	// Restore in database by clearing deleted_at
	err := store.RestoreDeletedEvent(m.db, m.lastDeletedID)
	if err != nil {
		// TODO: Show error to user
		return m, nil
	}

	// Refresh list from DB
	events, err := store.GetEvents(m.db, m.repoID, m.branch, 200)
	if err != nil {
		return m, nil
	}

	items := buildItems(events)

	// Re-sort based on current format
	if m.format == FormatWide {
		sortItemsForWideFormat(items)
	} else {
		sortItemsForCompactFormat(items)
	}

	// Update model
	m.allItems = items
	m.items = items
	m.lastDeletedID = ""

	// Apply filter if active
	if m.filterQuery != "" {
		m.applyFilter()
	}

	// Reset cursor to safe position
	if m.cursor >= len(m.items) {
		m.cursor = 0
	}

	return m, nil
}

// handleEdit opens the current item in $EDITOR
func (m Model) handleEdit() (tea.Model, tea.Cmd) {
	if len(m.items) == 0 || m.cursor >= len(m.items) {
		return m, nil
	}

	item := &m.items[m.cursor]

	// Store item to edit and exit TUI (editing will be handled externally)
	m.pendingEdit = item
	m.quitting = true
	return m, tea.Quit
}

// applyFilter filters items based on filterQuery
func (m *Model) applyFilter() {
	if m.filterQuery == "" {
		// No filter, show all items
		m.items = m.allItems
		return
	}

	// Filter items by query (case-insensitive substring match)
	query := strings.ToLower(m.filterQuery)
	var filtered []Item
	for _, item := range m.allItems {
		// Match against text
		if strings.Contains(strings.ToLower(item.DisplayText), query) {
			filtered = append(filtered, item)
			continue
		}
		// Match against type
		if strings.Contains(strings.ToLower(item.Event.Type), query) {
			filtered = append(filtered, item)
			continue
		}
	}

	m.items = filtered

	// Reset cursor if out of bounds
	if m.cursor >= len(m.items) {
		m.cursor = 0
	}
}
