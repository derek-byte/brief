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

## Release Management

### Overview

This project uses automated semantic versioning with [svu](https://github.com/caarlos0/svu) and [GoReleaser](https://goreleaser.com) to handle releases. Versions are automatically calculated from Conventional Commits.

### Release Process

**Note:** Releases are cut by maintainers.

**One-time setup:**
```bash
make setup
```

This installs `svu` (semantic version utility) and authenticates with GitHub CLI.

**Creating a Release (Two-step process):**

```bash
# Step 1: Calculate version and create tag
make tag

# Step 2: Publish GitHub release
make publish
```

**Why two steps?**
- Separation allows you to tag a release and publish it separately
- Makes it easier to retry if publishing fails
- Gives you a chance to verify the version before publishing

**What happens:**
1. `make tag` analyzes commits since last tag using Conventional Commits
2. Calculates next semantic version automatically
3. Creates and pushes git tag
4. `make publish` runs GoReleaser to build binaries and create GitHub release
5. Updates Homebrew tap automatically

### Version Bumping

Versions are automatically calculated from commit messages following [Conventional Commits](https://www.conventionalcommits.org/):

| Commit Type | Example | Version Impact |
|-------------|---------|----------------|
| `fix:` | `fix(log): handle nil logger` | Patch (v1.0.0 → v1.0.1) |
| `feat:` | `feat(kafka): add consumer` | Minor (v1.0.0 → v1.1.0) |
| `feat!:` or `BREAKING CHANGE:` | `feat!: change API` | Major (v1.0.0 → v2.0.0) |

**All commit types:**
- `feat:` - New feature (minor bump)
- `fix:` - Bug fix (patch bump)
- `perf:` - Performance improvement (patch bump)
- `refactor:` - Code refactoring (patch bump)
- `docs:` - Documentation only (patch bump)
- `style:` - Code style changes (patch bump)
- `test:` - Test additions/changes (patch bump)
- `build:` - Build system changes (patch bump)
- `ci:` - CI configuration changes (patch bump)
- `chore:` - Other changes (patch bump)

**Breaking changes** (major bump) can be indicated with:
- `!` after type: `feat!:`, `fix!:`, etc.
- `BREAKING CHANGE:` in commit body

### GoReleaser Configuration

**Location:** `.goreleaser.yaml`

**Key Settings:**

- **CGO_ENABLED=0**: Uses pure Go SQLite (modernc.org/sqlite)
- **Platforms**: darwin/linux/windows, amd64/arm64
- **Homebrew tap**: Automatically updates `derek-byte/homebrew-tap`
- **Changelog**: Auto-generated from Conventional Commits with grouping

**Changelog Groups:**
1. Breaking Changes
2. Features
3. Bug Fixes
4. Documentation
5. Performance
6. Refactoring
7. Other Changes

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

### GitHub Secrets Required

For automated releases, configure these secrets in GitHub repository settings:

1. **GITHUB_TOKEN**: Automatically provided by GitHub Actions
2. **HOMEBREW_TAP_GITHUB_TOKEN**: Personal Access Token with `repo` scope for tap repository

**Creating the tap token:**
1. Go to GitHub Settings → Developer settings → Personal access tokens
2. Create new token with `repo` scope
3. Add to repository secrets as `HOMEBREW_TAP_GITHUB_TOKEN`

### Make Commands

Available commands (run `make help` for full list):

- `make setup` - One-time setup (install svu, authenticate gh)
- `make tag` - Calculate next version and create git tag
- `make publish` - Run GoReleaser to create GitHub release
- `make build` - Build binary locally
- `make test` - Run all tests
- `make clean` - Remove build artifacts

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

1. **Follow Conventional Commits** - Commit messages directly determine version bumps:
   - `feat:` - New feature (minor version bump)
   - `fix:` - Bug fix (patch version bump)
   - `feat!:` or `BREAKING CHANGE:` - Breaking change (major version bump)
   - See full list in Release Management section above

2. **Add tests** for new features

3. **Test locally** before releasing:
   ```bash
   make build
   make test
   goreleaser check
   ```

4. **Test on multiple platforms** if possible (macOS, Linux, Windows)

5. **Update documentation** when changing architecture or adding features

**Example commit messages:**
```bash
feat(ui): add vim-like navigation keybindings
fix(db): prevent race condition in event creation
docs: update installation instructions
feat!: change default database location
```

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
