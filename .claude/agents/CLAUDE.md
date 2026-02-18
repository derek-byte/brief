# BranchBrief Development Guide

## Project Overview

BranchBrief (`brief`) is a local-first CLI tool for managing branch-scoped development notes. It helps developers capture context, todos, decisions, and commands while working, then quickly rehydrate that context when returning to a branch.

**Core Architecture:**
- Go CLI application using Cobra for command handling
- SQLite database for persistent storage (with CGO dependency)
- Bubble Tea TUI for interactive mode
- Git integration for branch detection and hook management

**Key Design Principles:**
- Local-first: All data stored privately on user's machine
- Branch-scoped: Each git branch gets isolated context
- Fast rehydration: Resume work in < 60 seconds
- Low friction: Shorthand commands for rapid capture

## Release Management with GoReleaser

### Overview

This project uses [GoReleaser](https://goreleaser.com) to automate releases across multiple platforms and publish to Homebrew.

### Release Process

**Creating a Release:**

1. **Tag the version:**
   ```bash
   git tag -a v1.0.1 -m "Release v1.0.1"
   git push origin v1.0.1
   ```

2. **GitHub Actions automatically:**
   - Runs tests
   - Builds for macOS and Linux (amd64/arm64)
   - Creates GitHub release
   - Publishes to Homebrew tap

3. **Users install via Homebrew:**
   ```bash
   brew install derek-byte/tap/brief
   ```

### GoReleaser Configuration

**Location:** `.goreleaser.yaml`

**Key Settings:**

- **CGO_ENABLED=1**: Required for SQLite (go-sqlite3 dependency)
- **macOS runner**: GitHub Actions uses macOS for native CGO cross-compilation
- **Platforms**: darwin/linux, amd64/arm64
- **Homebrew tap**: Automatically updates `derek-byte/homebrew-tap`

**Important Notes:**

1. **SQLite CGO Requirement**: The project uses `github.com/mattn/go-sqlite3` which requires CGO. This means:
   - Cannot use `CGO_ENABLED=0`
   - Cross-compilation requires platform-specific toolchains
   - macOS runner provides best compatibility for darwin builds

2. **Homebrew Token**: The workflow uses `HOMEBREW_TAP_GITHUB_TOKEN` secret to push formula updates to the tap repository. This must be configured in GitHub repository secrets.

### Testing Releases Locally

**Snapshot release (no publish):**
```bash
goreleaser release --snapshot --clean
```

**Build only:**
```bash
goreleaser build --snapshot --clean
```

**Check configuration:**
```bash
goreleaser check
```

### Homebrew Tap Setup

**Repository:** `derek-byte/homebrew-tap`

The tap repository is automatically managed by GoReleaser. When a release is created:
1. GoReleaser generates the formula file
2. Commits it to the tap repo's `Formula/` directory
3. Users can install with `brew install derek-byte/tap/brief`

**Formula location:** `https://github.com/derek-byte/homebrew-tap/blob/main/Formula/brief.rb`

### Version Bumping Strategy

This project follows semantic versioning (semver):
- **v1.0.x**: Patch releases (bug fixes)
- **v1.x.0**: Minor releases (new features, backward compatible)
- **v2.0.0**: Major releases (breaking changes)

**Commit message prefixes map to version bumps:**
- `fix:` → patch version
- `feat:` → minor version
- `feat!:` or `BREAKING CHANGE:` → major version

### GitHub Secrets Required

For automated releases, configure these secrets in GitHub repository settings:

1. **GITHUB_TOKEN**: Automatically provided by GitHub Actions
2. **HOMEBREW_TAP_GITHUB_TOKEN**: Personal Access Token with `repo` scope for tap repository

**Creating the tap token:**
1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Create new token with `repo` scope
3. Add to repository secrets as `HOMEBREW_TAP_GITHUB_TOKEN`

## Development Workflow

### Building Locally

```bash
# Development build
go build -o brief ./cmd/brief

# Install to ~/bin for testing
go build -o ~/bin/brief ./cmd/brief

# Run tests
go test ./...
```

### Git Hooks Integration

The tool installs git hooks to automatically show branch context on checkout:

**Implementation:** `internal/git/hooks.go`

**Features:**
- Respects `core.hooksPath` (Husky compatibility)
- Automatically excludes hook files from git status (`.git/info/exclude`)
- Appends to existing hooks (doesn't overwrite)

**Testing hook installation:**
```bash
cd test-repo
brief init
git checkout -b test-branch
# Should see brief summary on checkout
```

## Database Schema

**Location:** `~/Library/Application Support/branchbrief/branchbrief.sqlite`

**Schema:** (see `internal/store/db.go`)
- `events` table with branch-scoped entries
- Soft delete support (`deleted_at` timestamp)
- Todo completion tracking (`completed_at`)
- Edit history (`updated_at`)

## TUI Architecture

**Framework:** Bubble Tea (charmbracelet/bubbletea)

**Key Files:**
- `internal/tui/model.go`: Main TUI model
- `internal/tui/render.go`: Rendering logic
- `internal/tui/keys.go`: Keyboard shortcuts

**Features:**
- Dual view modes (wide/compact)
- Live filtering/search
- Todo completion toggling
- Command execution
- Undo/redo for deletions

## Future Enhancements

### Potential Features
- Export to markdown
- Sync between machines (optional)
- AI-powered summary generation
- Integration with issue trackers
- Plugin system for custom event types

### Performance Considerations
- Database indexing for large repositories
- Lazy loading in TUI for many events
- Pagination for rehydrate command

## Contributing

When making changes:
1. Follow conventional commits (`feat:`, `fix:`, etc.)
2. Add tests for new features
3. Run `goreleaser check` before tagging releases
4. Test on both macOS and Linux if possible
5. Update this doc when changing architecture

## Troubleshooting

### CGO Build Issues

If you encounter CGO errors:
```bash
# macOS: Install Xcode command line tools
xcode-select --install

# Linux: Install gcc
sudo apt-get install build-essential  # Debian/Ubuntu
sudo yum install gcc                   # RHEL/CentOS
```

### Hook Installation Issues

If `brief init` fails to install hooks:
1. Check git version: `git --version` (need 2.9+)
2. Verify repo is a git repository: `git status`
3. Check permissions on hooks directory
4. Review `.git/info/exclude` for conflicts

### Database Issues

If database becomes corrupted:
```bash
# Backup first
cp ~/Library/Application\ Support/branchbrief/branchbrief.sqlite ~/branchbrief-backup.sqlite

# Then remove and reinitialize
rm ~/Library/Application\ Support/branchbrief/branchbrief.sqlite
brief rehydrate  # This will recreate the schema
```

## Resources

- [GoReleaser Documentation](https://goreleaser.com)
- [Homebrew Formula Cookbook](https://docs.brew.sh/Formula-Cookbook)
- [Bubble Tea Tutorial](https://github.com/charmbracelet/bubbletea/tree/master/tutorials)
- [Cobra CLI Guide](https://github.com/spf13/cobra)
