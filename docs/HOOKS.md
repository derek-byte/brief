# Git Hooks Guide

How Brief integrates with git hooks for automatic context display.

---

## Quick Start

```bash
cd ~/my-project
brief init                    # Install hook
git checkout other-branch     # Context displays automatically
brief init --uninstall        # Remove hook
```

---

## What Gets Installed

Brief installs a `post-checkout` hook that runs automatically when you switch branches.

**Default behavior** (`brief init`):
- Runs `brief summary --quiet` after branch checkout
- Shows compact context overview
- Errors are suppressed (fails silently)

**Full mode** (`brief init --full`):
- Runs `brief rehydrate` after branch checkout
- Shows complete context with all details
- More verbose output

---

## Hook Location

Brief respects your repository's hook configuration:

### Standard Setup
Hook installed to: `.git/hooks/post-checkout`

### Husky / Custom Hook Directory
If `core.hooksPath` is configured, Brief installs there:

```bash
# Example: Husky configuration
$ git config --get core.hooksPath
.husky

# Brief installs to .husky/post-checkout
$ brief init
Installed branchbrief hook (summary mode)
```

---

## How It Works

### 1. Detection
Brief checks `git config core.hooksPath` to determine hook directory.

### 2. Appending
Brief **appends** to existing hooks rather than overwriting:

```bash
#!/bin/sh
# Your existing hook content...

# >>> branchbrief auto-context >>>
if [ "$3" = "1" ]; then
  brief summary --quiet 2>/dev/null || true
fi
# <<< branchbrief auto-context <<<
```

The markers (`>>>` and `<<<`) allow clean removal with `--uninstall`.

### 3. Auto-Exclusion
If the hook is in your working tree (e.g., `.husky/`), Brief automatically adds it to `.git/info/exclude` so it doesn't appear in `git status`.

**Why `.git/info/exclude` instead of `.gitignore`?**
- Personal ignore pattern (not shared with team)
- Doesn't require committing changes
- No git status clutter

---

## Compatibility

### Works With
✅ Standard git hooks (`.git/hooks/`)
✅ Husky (`.husky/`)
✅ Custom `core.hooksPath` (relative or absolute paths)
✅ Existing post-checkout hooks (appends, doesn't overwrite)

### Requirements
- Git 2.9+ (for `core.hooksPath` support)
- `brief` binary in PATH
- Git repository

---

## Troubleshooting

### Hook doesn't run

**Check installation:**
```bash
# In your repo
ls -la .git/hooks/post-checkout

# Or if using custom hooksPath
git config --get core.hooksPath
ls -la $(git config --get core.hooksPath)/post-checkout
```

**Check permissions:**
```bash
# Hook must be executable
chmod +x .git/hooks/post-checkout
```

**Test manually:**
```bash
# Run the hook directly
.git/hooks/post-checkout HEAD HEAD 1

# Should display brief summary
```

### Hook shows errors

Brief suppresses errors by default (`2>/dev/null || true`), but you can debug:

```bash
# Remove error suppression temporarily
vim .git/hooks/post-checkout

# Change:
#   brief summary --quiet 2>/dev/null || true
# To:
#   brief summary --quiet
```

### Hook appears in git status

If the hook file shows as untracked:

1. **Verify auto-exclusion:**
   ```bash
   cat .git/info/exclude
   # Should contain: .husky/post-checkout (or your hook path)
   ```

2. **Manually exclude if needed:**
   ```bash
   echo ".husky/post-checkout" >> .git/info/exclude
   ```

### Conflicts with existing hooks

Brief appends to existing hooks, so conflicts are rare. If you have issues:

1. **Check for duplicate blocks:**
   ```bash
   cat .git/hooks/post-checkout
   # Look for multiple "branchbrief auto-context" blocks
   ```

2. **Manually edit the hook:**
   ```bash
   vim .git/hooks/post-checkout
   # Remove duplicate Brief blocks between markers
   ```

3. **Reinstall:**
   ```bash
   brief init --uninstall
   brief init
   ```

---

## Manual Installation

If you prefer manual setup:

```bash
# Create or edit post-checkout hook
vim .git/hooks/post-checkout
```

Add this content:

```bash
#!/bin/sh

# Your existing hook content (if any)...

# >>> branchbrief auto-context >>>
if [ "$3" = "1" ]; then
  brief summary --quiet 2>/dev/null || true
fi
# <<< branchbrief auto-context <<<
```

Make it executable:

```bash
chmod +x .git/hooks/post-checkout
```

---

## Hook Behavior Details

### When Hook Runs

The post-checkout hook runs after:
- `git checkout <branch>`
- `git switch <branch>`
- `git clone <repo>` (initial checkout)

### When Hook Doesn't Run

The hook does NOT run for:
- `git checkout <file>` (file checkout, not branch switch)
- `git checkout -b <new-branch>` (new branch creation from current branch - no context yet)
- Detached HEAD checkouts

### The `$3` Check

```bash
if [ "$3" = "1" ]; then
```

Git's post-checkout hook receives 3 arguments:
1. `$1` — Previous HEAD ref
2. `$2` — New HEAD ref
3. `$3` — Flag: `1` = branch checkout, `0` = file checkout

Brief only runs on branch checkouts (`$3 = 1`).

---

## Advanced Configuration

### Custom Hook Command

Edit the installed hook to customize behavior:

```bash
vim .git/hooks/post-checkout
```

**Examples:**

**Use full rehydrate instead of summary:**
```bash
if [ "$3" = "1" ]; then
  brief rehydrate 2>/dev/null || true
fi
```

**Show timeline view:**
```bash
if [ "$3" = "1" ]; then
  brief rehydrate --view timeline 2>/dev/null || true
fi
```

**Custom limit:**
```bash
if [ "$3" = "1" ]; then
  brief rehydrate --limit 20 2>/dev/null || true
fi
```

### Conditional Execution

Only run hook in specific repos:

```bash
if [ "$3" = "1" ] && [ "$(git config --get remote.origin.url)" = "git@github.com:user/specific-repo.git" ]; then
  brief summary --quiet 2>/dev/null || true
fi
```

Only run in work directory:

```bash
if [ "$3" = "1" ] && [[ "$PWD" == "/Users/you/work/"* ]]; then
  brief summary --quiet 2>/dev/null || true
fi
```

---

## Uninstalling

```bash
brief init --uninstall
```

This removes the Brief block from the hook file. If the file becomes empty or contains only the shebang, it's deleted entirely.

**Manual removal:**
```bash
# Edit the hook
vim .git/hooks/post-checkout

# Delete the block between these markers:
# >>> branchbrief auto-context >>>
# <<< branchbrief auto-context <<<
```

---

## Security Considerations

**Hook execution:**
- Git hooks run arbitrary shell commands
- Review hook content before installing: `cat .git/hooks/post-checkout`
- Brief hooks are safe (only runs `brief summary`, no network/file system access)

**Hook sources:**
- Hooks in `.git/hooks/` are local-only (not tracked by git)
- Hooks in custom directories (e.g., `.husky/`) may be committed
- Review committed hooks in repositories you clone

---

## Further Reading

- [Git Hooks Documentation](https://git-scm.com/book/en/v2/Customizing-Git-Git-Hooks)
- [Husky Documentation](https://typicode.github.io/husky/)
- [Brief Command Reference](./COMMANDS.md)
- [Development Guide](../.claude/agents/CLAUDE.md)
