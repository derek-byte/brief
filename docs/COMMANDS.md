# Command Reference

Complete reference for all Brief commands.

## Context Management

The branch name serves as your context header - name your branches descriptively!

### `brief "<text>"`
Quick note capture (catch-all command).

```bash
brief "Remember to test edge cases"
brief "todo: Write unit tests"              # Auto-detected as todo
brief "cmd: npm run build"                  # Auto-detected as command
```

### `brief todo`
Add actionable item.

```bash
brief todo "Update API documentation"
brief todo "Fix login bug"
```

### `brief decision`
Document technical decision.

```bash
brief decision "Use JWT for authentication"
brief decision "Switch to Vite from Webpack"
```

### `brief cmd`
Save useful command for later.

```bash
brief cmd "npm run test:e2e"
brief cmd "docker compose up -d"
```

### `brief note`
Add general note.

```bash
brief note "API rate limit is 100 req/min"
```

### `brief link`
Save reference link.

```bash
brief link "https://docs.example.com/auth"
```

### `brief issue`
Reference issue tracker.

```bash
brief issue "#123 - Login timeout bug"
```

### `brief error`
Document error or bug.

```bash
brief error "NullPointerException in getUserData()"
```

### `brief fix`
Document fix or workaround.

```bash
brief fix "Added null check in getUserData()"
```

---

## Viewing Context

### `brief summary`
Compact overview of current branch.

```bash
brief summary                    # Default view
brief summary --quiet            # Minimal output
```

### `brief rehydrate`
Full context view with all details.

```bash
brief rehydrate                        # Structured view (default)
brief rehydrate --view structured      # Group by type
brief rehydrate --view timeline        # Chronological
brief rehydrate --limit 50             # Limit entries
```

### `brief ui`
Interactive terminal UI.

```bash
brief ui                         # Launch TUI
```

See [Interactive Mode](#interactive-mode) for keyboard shortcuts.

---

## Stash Management

### `brief save`
Stash work for current branch.

```bash
brief save                              # Save with auto-generated message
brief save "WIP: OAuth integration"     # Save with custom message
```

### `brief restore`
Restore stashed work for current branch.

```bash
brief restore                    # Apply most recent stash for this branch
```

### `brief stashes`
List all stashes across branches.

```bash
brief stashes                    # Show all stashes
```

---

## Branch Management

### `brief clear`
Clear all context for current branch.

```bash
brief clear                      # Prompts for confirmation
brief clear --force              # Skip confirmation
```

---

## Git Hooks

### `brief init`
Install post-checkout git hook.

```bash
brief init                       # Install hook (summary mode)
brief init --full                # Install hook (full rehydrate mode)
brief init --uninstall           # Remove hook
```

The hook automatically displays context when switching branches.

**Hook behavior:**
- Respects `core.hooksPath` (works with Husky)
- Appends to existing hooks (doesn't overwrite)
- Auto-excluded from `git status` (uses `.git/info/exclude`)

---

## Advanced

### `brief add`
Add any event type explicitly.

```bash
brief add note "Custom note text"
brief add todo "Custom todo"
brief add error --from-stdin "Build failed" < error.log
```

---

## Interactive Mode

Launch with `brief ui`.

### Navigation
- `j` / `k` or `↑` / `↓` — Move cursor
- `g` / `G` — Jump to top/bottom
- `v` — Toggle view mode (wide/compact)
- `/` — Enter filter/search mode
- `?` — Toggle help
- `q` — Quit

### Actions
- `Enter` — Context-dependent action:
  - **Todo**: Toggle completion
  - **Command**: Exit TUI and execute
  - **Note/Choice/Fix**: Show detail view
- `e` — Edit current item in `$EDITOR`
- `d` — Soft delete current item
- `u` — Undo last delete
- `ESC` — Exit filter mode / return to list

### Filter Mode
- Press `/` to enter filter mode
- Type to search (case-insensitive, matches text and type)
- `Enter` — Keep filter active, return to navigation
- `ESC` — Clear filter, return to navigation

### View Modes
- **Wide**: Events grouped by type with spacing
- **Compact**: Continuous timeline (newest first)

---

## Environment Variables

- `EDITOR` — Editor for TUI edit mode (`e` key)
- `BRIEF_DEBUG` — Enable debug logging (set to `1`)

---

## Files

- `~/Library/Application Support/branchbrief/branchbrief.sqlite` — Local database
- `.git/hooks/post-checkout` — Git hook (if installed)
- `.git/info/exclude` — Hook exclusion pattern (if in working tree)

---

## Exit Codes

- `0` — Success
- `1` — User error (bad arguments, invalid input)
- `2` — Runtime error (database failure, git error)

---

For troubleshooting and advanced topics, see [HOOKS.md](./HOOKS.md) and [Development Guide](../.claude/agents/CLAUDE.md).
