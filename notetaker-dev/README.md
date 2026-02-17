# BranchBrief

A local-first CLI that stores private, branch-scoped development notes and prints a concise rehydration brief so you can resume work in ~30 seconds without rereading Slack/LLM chats.

## Installation

### Option 1: Install to ~/bin (Quick Start)

```bash
# Build and install to user bin directory
go build -o ~/bin/brief ./cmd/brief

# Add ~/bin to PATH if not already there (zsh)
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.zshrc
source ~/.zshrc

# Verify installation
brief --version
```

### Option 2: Install to /usr/local/bin (System-wide)

```bash
# Build the binary
go build -o brief ./cmd/brief

# Move to system location (requires sudo)
sudo mv brief /usr/local/bin/brief

# Verify installation
brief --version
```

### Option 3: Homebrew (Future)

Coming soon! For now, use Option 1 or 2 above.

## Quick Start

Once installed, navigate to any git repository and start capturing context:

```bash
# Navigate to your project
cd ~/my-project

# Set your goal (one per branch, updates not appends)
brief goal "Add user authentication"

# Quick notes using catch-all command (creates a note)
brief "Remember to test edge cases"

# Or use typed shorthand commands
brief decision "Use JWT tokens"
brief todo "Write unit tests"
brief cmd "make test"

# Save uncommitted work before switching branches
brief save "OAuth integration WIP"
git checkout other-branch

# Later, come back and restore
git checkout my-branch
brief restore

# View your branch context
brief rehydrate

# List all saved work across branches
brief stashes
```

## All Commands

### Goal (Single per Branch)
```bash
brief goal "<text>"       # Set/update goal for this branch
brief goal                # View current goal
brief goal --edit         # Edit goal in $EDITOR
```

### Stash Management (Per-Branch)
```bash
brief save [message]      # Save (stash) work for this branch
brief restore             # Restore work for this branch
brief stashes             # List all saved work by branch
```

### Notes (Shorthand)
```bash
brief "<text>"            # Catch-all: creates a note (no subcommand needed)
brief "todo: <text>"      # Auto-detected as todo (prefix stripped)
brief "cmd: <text>"       # Auto-detected as cmd (prefix stripped)
brief "decision: <text>"  # Auto-detected as decision (prefix stripped)
brief "note: <text>"      # Auto-detected as note (prefix stripped)

# Or use explicit type commands
brief decision "<text>"   # Add a decision
brief todo "<text>"       # Add a todo
brief cmd "<text>"        # Add a command
brief fix "<text>"        # Add a fix note
brief error "<text>"      # Add an error
brief link "<text>"       # Add a link
brief issue "<text>"      # Add an issue reference
brief note "<text>"       # Add a general note
```

**Smart Prefix Detection**: The catch-all command automatically detects event types from prefixes like "todo:", "cmd:", "decision:", "note:", etc. The prefix is stripped from the final text.

### View Modes
```bash
brief rehydrate                        # Structured view (default)
brief rehydrate --view structured      # Grouped by type, oldest-first
brief rehydrate --view timeline        # Chronological activity log
```

**Structured mode** (default): Best for "what do I do next?"
- Groups events by type (Goal, Decisions, Todos, Commands, Notes)
- Shows oldest-first within each group for chronological flow
- Hides verbose event types (errors, fixes, stash records)

**Timeline mode**: Best for "what happened?"
- Shows all events in chronological order (newest-first)
- Format: `HH:MM [type] text`
- Includes all event types for complete activity log

### Other
```bash
brief add <type> "<text>"              # Add any event type
brief add error --from-stdin "<desc>"  # Capture from stdin
brief rehydrate --limit 100            # Limit events fetched
```

## Features

- **Local-first**: All data stored privately in `~/Library/Application Support/branchbrief/`
- **Branch-scoped**: Notes tied to your current git branch
- **Single goal per branch**: Updates instead of appends - one clear objective
- **Per-branch stash management**: Never apply wrong stash to wrong branch
- **Fast rehydration**: Get oriented in < 60 seconds
- **Works across projects**: Same database, different branches per repo
- **Low friction**: Shorthand commands for rapid capture

## How It Works

1. **Per-branch context**: Each git branch in each repo gets its own notes
2. **Goal management**: One goal per branch (updates not appends)
3. **Stash tracking**: Tracks git stash refs per branch - applies right stash automatically
4. **Repo identification**: Uses git remote URL (or path) for consistency
5. **SQLite storage**: All metadata stored locally in a single database
6. **Git does the heavy lifting**: Stashes stored in git's native storage

## Development

This is v1 - focused on core workflows:
- Capturing context as you work
- Rehydrating when you return
- Staying private and local

For full implementation plan, see `.claude/agents/CLAUDE.md`
