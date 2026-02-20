---
title: Getting Started
description: Branch-scoped context for developers who context-switch.
---

## What is Brief?

Brief is a local-first CLI for saving development context per git branch.

It helps you capture todos, notes, commands, and decisions while you work — then instantly restore that context when you return.

```bash
brew install derek-byte/tap/brief
brief init
```

## Quick Example

```bash
# Set your branch goal
brief goal "Add OAuth authentication"

# Capture notes as you work
brief "Check rate limiting on login endpoint"
brief todo "Add password reset flow"
brief cmd "npm run test:auth"

# View your context
brief summary

# Switch branches and come back
git checkout feature/new-ui
git checkout feature/oauth
brief rehydrate
# → Shows everything you captured, right where you left off
```

## Key Features

- **Branch-scoped** — Every git branch gets its own context
- **Local-first** — All data stays private on your machine
- **Interactive TUI** — Browse and manage notes with vim-like keybindings
- **Smart stashing** — Branch-aware git stash management
- **Git hook integration** — Auto-display context when switching branches

## Next Steps

- [Installation](/docs/installation) — Install Brief on your machine
- [Commands Reference](/docs/commands) — Complete command documentation
- [Git Hooks](/docs/hooks) — Set up automatic context display
