# BranchBrief

A local-first CLI that stores private, branch-scoped development notes and prints a concise rehydration brief so you can resume work in ~30 seconds without rereading Slack/LLM chats.

## Installation

```bash
go build -o brief ./cmd/brief
```

## Usage

```bash
# Add notes to your current branch
brief add goal "Implement user authentication"
brief add decision "Use JWT tokens for session management"
brief add todo "Write unit tests"
brief add cmd "make test"

# View your branch context
brief rehydrate

# Check what's been captured
brief status
```

## Features

- **Local-first**: All data stored privately in ~/Library/Application Support/branchbrief
- **Branch-scoped**: Notes are tied to your current git branch
- **Fast rehydration**: Get oriented in < 60 seconds
- **No cloud sync**: Everything stays on your machine

## Development

This is v1 - focused on core workflows:
- Capturing context as you work
- Rehydrating when you return
- Staying private and local

For full implementation plan, see `.claude/agents/CLAUDE.md`
