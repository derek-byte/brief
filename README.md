```
██████╗ ██████╗ ██╗███████╗███████╗
██╔══██╗██╔══██╗██║██╔════╝██╔════╝
██████╔╝██████╔╝██║█████╗  █████╗
██╔══██╗██╔══██╗██║██╔══╝  ██╔══╝         (\_/)
██████╔╝██║  ██║██║███████╗██║            ( •_•)
╚═════╝ ╚═╝  ╚═╝╚═╝╚══════╝╚═╝            / >✓

Branch-scoped context for developers who context-switch.
```

[Installation](#installation) • [Quick Start](#quick-start) • [Documentation](#documentation)

---

## Installation

```bash
brew install derek-byte/tap/brief
```

Or install from source:
```bash
go install github.com/derek-byte/brief/cmd/brief@latest
```

Or download binaries from [releases](https://github.com/derek-byte/brief/releases).

---

## What is Brief?

Brief captures development context as you work and restores it when you return. No more rereading Slack threads or scrolling through LLM chat history to remember where you left off.

Every branch gets its own context. Everything stays private on your machine.

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

## Quick Start

```bash
cd ~/my-project

# Set your branch goal
brief goal "Implement user authentication"

# Capture notes as you work
brief "Check rate limiting on login endpoint"
brief todo "Add password reset flow"
brief cmd "npm run test:auth"

# View your context
brief summary

# Interactive TUI with vim-like keybindings
brief ui
```

**Key Features:**
- **Branch-scoped** — Every git branch gets its own context
- **Local-first** — All data stays private on your machine
- **Interactive TUI** — Browse and manage notes with vim-like keybindings
- **Smart stashing** — Branch-aware git stash management
- **Git hook integration** — Auto-display context when switching branches

---

## Commands

**Capture context:**
```bash
brief "<text>"              # Quick note
brief goal "<text>"         # Set branch objective
brief todo "<text>"         # Add actionable item
brief decision "<text>"     # Document technical choice
brief cmd "<text>"          # Save useful command
```

**View & manage:**
```bash
brief summary               # Compact overview
brief rehydrate             # Full context view
brief ui                    # Interactive TUI
brief clear                 # Clear current branch
```

**Branch-aware stashing:**
```bash
brief save [message]        # Stash work for this branch
brief restore               # Restore this branch's work
brief stashes               # List all stashes
```

**Git hook integration:**
```bash
brief init                  # Install post-checkout hook
```

After installing the hook, `brief summary` runs automatically when you switch branches.

---

## How It Works

Brief stores notes in a local SQLite database (`~/Library/Application Support/branchbrief/`). Each branch gets its own context, identified by repo + branch name. Everything stays local—no network requests, no cloud storage.

---

## Documentation

- [Command Reference](./docs/COMMANDS.md)
- [Git Hooks Guide](./docs/HOOKS.md)
- [Development Guide](./.claude/agents/CLAUDE.md)

---

## Contributing

```bash
git clone https://github.com/derek-byte/brief.git
cd brief
go build -o brief ./cmd/brief
go test ./...
```

See [Development Guide](./.claude/agents/CLAUDE.md) for architecture details.

---

## License

Brief was created by [Derek](https://github.com/derek-byte) and is licensed under the [MIT License](./LICENSE).
