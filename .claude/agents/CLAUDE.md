# CLAUDE.md — BranchBrief v1 incremental build guide

This file is the working agreement for implementing BranchBrief (local-first CLI) in small, reviewable steps.

## Principles

- **Max one feature per commit.** If you touch multiple concerns, split it.
- **Small diffs.** Prefer 20–120 lines changed per commit.
- **Refactor as you go.** Keep code tidy, remove dead paths quickly.
- **Few comments.** Prefer clear names + small functions. Comment only on *why*, not *what*.
- **Deterministic behavior.** No “magic AI” in v1; always reproducible output.
- **Local-first + private.** Store data outside the repo by default.

## Semantic commit messages

Use **Conventional Commits**:

- `feat:` user-facing functionality
- `fix:` bug fix
- `refactor:` internal change, no behavior change
- `test:` add/adjust tests
- `chore:` tooling, deps, build, CI

Examples:
- `feat(cli): add brief add command`
- `feat(store): persist events in sqlite`
- `feat(rehydrate): render branch brief`
- `fix(git): handle detached HEAD`
- `refactor(render): simplify section formatting`

## Formatting & code style

- Keep functions short (≈ 30 lines max).
- Avoid deep nesting; return early.
- Prefer simple data structures over cleverness.
- Do not introduce new dependencies unless they clearly reduce complexity.

## “Definition of Done” per commit

Each commit should:
- compile/build
- include/update tests when feasible
- not add TODOs that block later steps (unless explicitly staged)
- keep CLI help text accurate
- avoid partial, abandoned abstractions

## Incremental plan (one feature per commit)

### 0) Repo scaffolding
**Commit:** `chore: init go module and basic CLI skeleton`
- `go mod init …`
- `cmd/brief/main.go` prints help / version
- Basic `--version` flag

### 1) Git context detection
**Commit:** `feat(git): detect repo root and branch`
- Implement:
  - repo root: `git rev-parse --show-toplevel`
  - branch: `git branch --show-current`
  - fallback to short HEAD for detached mode
- Add minimal unit tests for parsing outputs (mock command runner).

### 2) Local app data directory (mac default)
**Commit:** `feat(store): resolve app data dir on mac`
- Default path:
  - `~/Library/Application Support/branchbrief/`
- Ensure directory exists on startup.
- Keep logic centralized (`internal/store/path.go`).

### 3) SQLite schema + migrations
**Commit:** `feat(store): initialize sqlite and events table`
- Create DB if missing at:
  - `<appdir>/branchbrief.sqlite`
- Create table `events` (append-only):
  - `id TEXT PRIMARY KEY`
  - `repo_id TEXT NOT NULL`
  - `branch TEXT NOT NULL`
  - `type TEXT NOT NULL`
  - `text TEXT NOT NULL`
  - `created_at INTEGER NOT NULL` (unix seconds)
  - `meta_json TEXT NOT NULL DEFAULT '{}'`

### 4) `brief add`
**Commit:** `feat(cli): add brief add command`
- `brief add <type> <text...>`
- Validate type in a small allowlist: `goal,decision,todo,cmd,error,link,issue,note`
- Persist event using sqlite.
- Keep output minimal: `Added <type> to <branch>`.

### 5) `brief status`
**Commit:** `feat(cli): add brief status command`
- For current `{repo, branch}` show:
  - counts by type
  - last updated timestamp
- No fancy formatting yet; keep it readable.

### 6) Git state snippet for rehydrate
**Commit:** `feat(git): collect brief git state`
- Gather:
  - last commit `git log -1 --oneline`
  - dirty file count from `git status --porcelain`
  - optional diffstat `git diff --stat` (truncate lines)
- Return a small `GitState` struct.

### 7) `brief rehydrate` renderer (no AI)
**Commit:** `feat(rehydrate): render branch brief`
- Query events for current `{repo, branch}`
- Render sections in this order:
  - Goal, State, Decisions, Known issues, Next steps, Commands
- Hard limits:
  - max 7 bullets per section
  - truncate overly long lines (e.g., 160 chars)
- The brief should fit in one terminal screen most of the time.

### 8) Stdin capture for logs/commands
**Commit:** `feat(cli): support --from-stdin for add`
- `brief add error --from-stdin "redis timeout"`
- If both text args and stdin provided, combine with newline.

### 9) Quality pass (refactor only)
**Commit:** `refactor: consolidate command runner and error handling`
- No behavior changes; reduce duplication.
- Normalize user-facing errors (actionable, short).

## Testing guidance

- Prefer unit tests for:
  - git output parsing
  - sqlite init and simple queries (use temp dir)
  - renderer section limits/truncation
- Keep tests fast and local. Avoid integration tests requiring real git repo until later.

## CLI UX rules

- Commands should be composable and scriptable.
- Exit codes:
  - `0` success
  - `1` user error (bad args)
  - `2` runtime error (git/db failures)
- Errors should say what to do next (e.g., “Run inside a git repo.”)

## Refactoring rules

- If a function becomes unclear, refactor in the same commit if it’s *part of the feature*.
- Larger cleanups: do them in a separate `refactor:` commit immediately after.

## What NOT to add in v1

- Cloud sync / auth
- Notion/Docs integrations
- Mermaid/graphs
- Embeddings/vector DB
- LLM-based summarization

## Quick local workflow

- `go test ./...`
- `go run ./cmd/brief --help`
- `go run ./cmd/brief add decision "…"`
- `go run ./cmd/brief rehydrate`

## Output contract (rehydrate)

Rehydrate must always show:
- Branch identifier
- Last updated time
- At least one of: Goal/Decisions/Todos/Commands (or “No notes yet”)

Keep it short. Optimize for “I’m back, what now?”.
