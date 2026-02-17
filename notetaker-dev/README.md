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

# Add notes using shorthand commands (fastest)
brief goal "Add user authentication"
brief decision "Use JWT tokens"
brief todo "Write unit tests"
brief cmd "make test"

# Or use the full syntax
brief add goal "..."
brief add decision "..."

# Capture command output
make test 2>&1 | brief add error --from-stdin "Test failures"

# View your branch context (< 60 seconds to rehydrate)
brief rehydrate

# Check what's been captured
brief status
```

## All Commands

### Shorthand (Recommended)
```bash
brief goal "<text>"       # Add a goal
brief decision "<text>"   # Add a decision
brief todo "<text>"       # Add a todo
brief cmd "<text>"        # Add a command
brief fix "<text>"        # Add a fix note
brief error "<text>"      # Add an error
brief link "<text>"       # Add a link
brief issue "<text>"      # Add an issue reference
brief note "<text>"       # Add a general note
```

### Full Syntax
```bash
brief add <type> "<text>"              # Add any event type
brief add error --from-stdin "<desc>"  # Capture from stdin
brief status                           # Show event counts
brief rehydrate                        # Display branch brief
brief rehydrate --limit 100            # Limit events fetched
```

## Features

- **Local-first**: All data stored privately in `~/Library/Application Support/branchbrief/`
- **Branch-scoped**: Notes are tied to your current git branch
- **Fast rehydration**: Get oriented in < 60 seconds
- **No cloud sync**: Everything stays on your machine
- **Works across projects**: Same database, different branches per repo
- **Detached HEAD support**: Works even when not on a branch
- **Low friction**: Shorthand commands for rapid capture

## How It Works

1. **Per-branch context**: Each git branch in each repo gets its own set of notes
2. **Repo identification**: Uses git remote URL (or path) to identify repos consistently
3. **SQLite storage**: All notes stored locally in a single SQLite database
4. **One-screen output**: Rehydrate brief fits in one terminal screen by design

## Development

This is v1 - focused on core workflows:
- Capturing context as you work
- Rehydrating when you return
- Staying private and local

For full implementation plan, see `.claude/agents/CLAUDE.md`
