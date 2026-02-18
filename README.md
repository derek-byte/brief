# Brief

```
   (\_/)
   ( •_•)
   / >☑️
```

**Branch-scoped context for developers.**

Brief helps you capture development context as you work and resume instantly when you return. No more rereading Slack threads or scrolling through LLM chat history to remember where you left off.

```bash
# Capture context
brief goal "Add OAuth authentication"
brief "Remember to test error handling"
brief todo "Update API documentation"

# Switch branches
git checkout feature/new-ui

# Come back later
git checkout feature/oauth
brief rehydrate
# → Shows everything you captured, right where you left off
```

---

## Installation

```bash
brew tap derek-byte/tap
brew install brief
```

<details>
<summary>Alternative installation methods</summary>

**From source:**
```bash
go install github.com/derek-byte/brief/cmd/brief@latest
```

**Download binary:**
Visit [releases](https://github.com/derek-byte/brief/releases) and download for your platform.

</details>

---

## Quick Start

```bash
# Navigate to any git repository
cd ~/my-project

# Set your branch goal
brief goal "Implement user authentication"

# Capture notes as you work
brief "Check rate limiting on login endpoint"
brief todo "Add password reset flow"
brief cmd "npm run test:auth"

# View your context anytime
brief summary

# Interactive mode with keyboard shortcuts
brief ui
```

---

## Features

- **Branch-scoped** — Every git branch gets its own context
- **Local-first** — All data stays private on your machine
- **Interactive TUI** — Browse, edit, and manage notes with vim-like keybindings
- **Smart stashing** — Never apply the wrong stash to the wrong branch
- **Zero friction** — Fast capture with shorthand commands
- **Git hook integration** — Auto-display context when switching branches

---

## Key Commands

### Capture Context

```bash
brief "<text>"              # Quick note (catch-all)
brief goal "<text>"         # Set branch objective
brief todo "<text>"         # Add actionable item
brief decision "<text>"     # Document technical choice
brief cmd "<text>"          # Save useful command
```

### View & Manage

```bash
brief summary               # Compact overview
brief rehydrate             # Full context view
brief ui                    # Interactive TUI
brief clear                 # Clear current branch
```

### Stash Management

```bash
brief save [message]        # Stash work for this branch
brief restore               # Restore this branch's work
brief stashes               # List all stashes
```

### Git Hooks

```bash
brief init                  # Install post-checkout hook
brief init --uninstall      # Remove hook
```

After installing the hook, `brief summary` runs automatically when you switch branches.

---

## How It Works

Brief stores notes in a local SQLite database (`~/Library/Application Support/branchbrief/`):

1. **Each branch gets its own context** — Identified by repo + branch name
2. **Notes are typed** — goals, todos, decisions, commands, links, etc.
3. **Stashes are tracked** — Git stash refs stored per-branch for safety
4. **Everything is local** — No network requests, no cloud storage

---

## Interactive Mode

Launch with `brief ui`:

**Navigation**
- `j`/`k` or `↑`/`↓` — Move cursor
- `g`/`G` — Jump to top/bottom
- `v` — Toggle view mode
- `/` — Search/filter
- `?` — Help

**Actions**
- `Enter` — Run command / toggle todo / view detail
- `e` — Edit in $EDITOR
- `d` — Delete (soft)
- `u` — Undo delete
- `q` — Quit

---

## Documentation

- [Command Reference](./docs/COMMANDS.md) — Full list of commands and options
- [Git Hooks Guide](./docs/HOOKS.md) — Hook installation and troubleshooting
- [Development Guide](./.claude/agents/CLAUDE.md) — Architecture and contributing

---

## Development

**Build from source:**
```bash
git clone https://github.com/derek-byte/brief.git
cd brief
go build -o brief ./cmd/brief
./brief --version
```

**Run tests:**
```bash
go test ./...
```

**Release:**
```bash
git tag -a v1.2.0 -m "Release v1.2.0"
git push origin v1.2.0
# GitHub Actions builds and publishes to Homebrew automatically
```

See [`.claude/agents/CLAUDE.md`](./.claude/agents/CLAUDE.md) for full development documentation.

---

## License

MIT © 2026 Derek

---

## Roadmap

Future features being explored:

- AI-powered context summarization
- Worktree orchestration for parallel development
- Coding agent integration
- Export to markdown
- Optional sync across machines

---

**Built for developers who context-switch between branches and want to resume work instantly.**
