# SPEC-DIVECC-INVENTORY-VIEW-001 — Acceptance Criteria

> Given-When-Then acceptance criteria with concrete verification commands. Every AC maps to ≥1 REQ; every REQ maps to ≥1 AC (full traceability in §D). Verification commands are run-phase artifacts — they are NOT executed at plan-phase.

---

## §A. Acceptance Criteria (Given-When-Then)

### AC-INV-001 — Command exists and is registered

- **Given** a built `moai` binary,
- **When** `moai inventory --help` is invoked,
- **Then** the command resolves (cobra does not report "unknown command") and `newInventoryCmd()` is registered on the root command via `rootCmd.AddCommand`.
- **Verification**:
  ```bash
  go build ./... && go run ./cmd/moai inventory --help 2>&1 | head -3
  grep -n "newInventoryCmd" internal/cli/root.go internal/cli/inventory.go
  go test -run 'TestInventory.*Registered|TestNewInventory' ./internal/cli/...
  ```
- **Covers**: REQ-INV-001, REQ-INV-002.

### AC-INV-002 — Read-only composition of the three surfaces

- **Given** the unified inventory command,
- **When** its `RunE` collects inventory,
- **Then** it calls exactly the three existing exported functions — `session.QueryActiveWork("")`, `WorktreeProvider.List()`, `harness.ListHarnesses(projectRoot)` — and presents each surface as a count + key-field summary per entry.
- **Verification**:
  ```bash
  grep -n "session.QueryActiveWork\|WorktreeProvider.List\|harness.ListHarnesses" internal/cli/inventory.go
  go test -run 'TestInventory_ComposesThreeSurfaces' ./internal/cli/...
  ```
- **Covers**: REQ-INV-003, REQ-INV-004.

### AC-INV-003 — `--json` emits the structured UnifiedInventoryReport

- **Given** the command with the `--json` flag,
- **When** `moai inventory --json` runs,
- **Then** stdout is a single JSON object that unmarshals into `UnifiedInventoryReport` with three top-level keys `sessions`, `worktrees`, `harnesses`, each carrying a `count` and an `entries` array.
- **Verification**:
  ```bash
  go test -run 'TestInventory_JSONShape' ./internal/cli/...
  # The test marshals/round-trips a report and asserts the three top-level keys
  # + per-row key fields (session_id/spec_id/phase, branch/path/head, name/domain/manifest_missing).
  ```
- **Covers**: REQ-INV-005.

### AC-INV-004 — Human-readable default emits a compact 3-surface summary

- **Given** the command without `--json`,
- **When** `moai inventory` runs,
- **Then** stdout contains a compact summary labelling all three surfaces (Sessions / Worktrees / Harnesses) with a count and the key fields per row, rendered via the existing render utilities.
- **Verification**:
  ```bash
  go test -run 'TestInventory_HumanReadableSummary' ./internal/cli/...
  # Asserts the rendered output contains the three surface labels + counts.
  ```
- **Covers**: REQ-INV-006.

### AC-INV-005a — Empty surfaces yield 0-count, never an error

- **Given** a project where a backing surface is empty or its backing dir/file is absent (no `.moai/state/active-sessions.json`; no `.claude/commands/harness/`; no worktree beyond main),
- **When** `moai inventory --json` runs,
- **Then** the affected surface reports `count: 0` (or main-only worktrees) and the command exits 0 — the empty condition is NOT an error.
- **Verification**:
  ```bash
  go test -run 'TestInventory_EmptySurfacesGraceful' ./internal/cli/...
  # Fixture with absent registry + absent harness dir → both report count 0, exit 0.
  ```
- **Covers**: REQ-INV-007, REQ-INV-010.

### AC-INV-005b — Genuine per-surface error degrades, does not abort the whole command

- **Given** one surface's backing call returns a genuine error (e.g. `WorktreeProvider == nil` — git module unavailable),
- **When** `moai inventory --json` runs,
- **Then** that surface's `error` field is populated, the surfaces that succeeded still render, and the command exits non-zero ONLY if no surface could be rendered.
- **Verification**:
  ```bash
  go test -run 'TestInventory_PerSurfaceErrorDegrades' ./internal/cli/...
  # Inject a nil WorktreeProvider; assert worktrees.error set, sessions/harnesses still rendered.
  ```
- **Covers**: REQ-INV-008.

### AC-INV-005c — Non-nil provider whose List() errors degrades the worktree surface

- **Given** a NON-NIL `WorktreeProvider` STUB whose `List()` returns a non-nil error (the real, common error path: invoked outside a git repo → `fatal: not a git repository`, git not in PATH, or porcelain parse failure — distinct from the `WorktreeProvider == nil` case in AC-INV-005b),
- **When** `moai inventory --json` runs,
- **Then** `worktrees.error` is populated from the git error, the sessions and harnesses surfaces still render (degrade-to-empty when their backing data is absent), and the command exits 0 because ≥1 surface rendered successfully.
- **Verification**:
  ```bash
  go test -run 'TestInventory_WorktreeListErrorDegrades' ./internal/cli/...
  # Inject a stub WorktreeProvider whose List() returns (nil, errors.New("fatal: not a git repository"));
  # assert worktrees.error set, sessions.count == 0 + harnesses.count == 0 still rendered, exit 0.
  ```
- **Covers**: REQ-INV-008 (error path), supports REQ-INV-010 (out-of-project worktree degradation).

### AC-INV-006a — Project-root resolution via the established helper

- **Given** a `--project-root` flag (or its absence),
- **When** the command resolves the project root,
- **Then** it uses the `resolveProjectRoot(cmd)` helper pattern (`--project-root` flag → `os.Getwd()` fallback), matching `internal/cli/harness.go`.
- **Verification**:
  ```bash
  grep -n "resolveProjectRoot" internal/cli/inventory.go
  go test -run 'TestInventory_ProjectRootResolution' ./internal/cli/...
  ```
- **Covers**: REQ-INV-009.

### AC-INV-006b — Zero backing-package modification

- **Given** the run-phase change set,
- **When** `git diff --name-only` is inspected against the SPEC branch base,
- **Then** no file under `internal/session/`, `internal/cli/worktree/`, `internal/cli/harness/`, or `internal/core/git/` is modified, and `moai worktree list` gains no `--json` flag.
- **Verification**:
  ```bash
  git diff --name-only origin/main...HEAD | grep -E 'internal/(session|cli/worktree|cli/harness|core/git)/' && echo "VIOLATION" || echo "PASS — backing packages untouched"
  grep -n '"json"' internal/cli/worktree/list.go || echo "PASS — no --json on worktree list"
  ```
- **Covers**: REQ-INV-011, REQ-INV-012.

### AC-INV-007 — Subagent boundary (no AskUserQuestion)

- **Given** the new CLI file,
- **When** the subagent-boundary static guard runs,
- **Then** `internal/cli/inventory.go` contains no `AskUserQuestion` / `mcp__askuser` invocation, and a `TestNewInventory_NoAskUserQuestion` guard test passes.
- **Verification**:
  ```bash
  grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/inventory.go | grep -v "_test.go" | grep -v "//" || echo "PASS — no AskUserQuestion"
  go test -run 'TestNewInventory_NoAskUserQuestion' ./internal/cli/...
  ```
- **Covers**: REQ-INV-013.

---

## §B. Edge Cases

| # | Edge case | Expected behavior | AC |
|---|-----------|-------------------|----|
| E1 | No `.moai/state/active-sessions.json` | `sessions.count == 0`, exit 0 | AC-INV-005a |
| E2 | No `.claude/commands/harness/` directory | `harnesses.count == 0`, exit 0 | AC-INV-005a |
| E3 | Only the main checkout (no extra worktrees) | `worktrees` lists the main worktree only (or count reflecting `WorktreeProvider.List()` baseline) | AC-INV-005a |
| E4 | `WorktreeProvider == nil` (git module unavailable) | `worktrees.error` set, other surfaces still render, exit reflects REQ-INV-008 | AC-INV-005b |
| E5 | Harness with `manifest_missing == true` (partial state) | row rendered with `manifest_missing: true` (surfaced, not crashed) — `ListHarnesses` already returns these | AC-INV-002 |
| E6 | Invoked outside a git repository (no `.moai/`, no `.claude/`, no git repo) | `sessions: 0, harnesses: 0, worktrees.error set, exit 0` — sessions+harnesses degrade-to-empty; the worktree surface errors (`git worktree list --porcelain` → `fatal: not a git repository`) per the error-degradation path; command still exits 0 since ≥1 surface rendered | AC-INV-005a + AC-INV-005b (REQ-INV-010 via REQ-INV-007 + REQ-INV-008) |
| E7 | `--json` AND no `--json` produce consistent counts | the two output modes report identical surface counts | AC-INV-003 + AC-INV-004 |

---

## §C. Quality Gate Criteria (Definition of Done)

- [ ] `go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0 (E2 cross-platform).
- [ ] `go test ./internal/cli/...` passes (all AC-INV tests green).
- [ ] `go test -cover ./internal/cli/...` — new-file contribution ≥ 85% (E3 coverage).
- [ ] `golangci-lint run --timeout=2m` — no NEW findings attributable to `internal/cli/inventory.go`.
- [ ] Subagent-boundary grep on `internal/cli/inventory.go` returns 0 matches (AC-INV-007).
- [ ] `git diff --name-only` confirms only `internal/cli/inventory.go` + `internal/cli/inventory_test.go` + the one-line `internal/cli/root.go` edit + run-phase `progress.md` changed (AC-INV-006b).
- [ ] `moai inventory --json` round-trips into `UnifiedInventoryReport` with the three top-level keys (AC-INV-003).
- [ ] All 8 AC-INV criteria PASS in the E1 self-verification matrix.
- [ ] Conventional Commits + `🗿 MoAI` trailer on every commit; no `--no-verify`.

---

## §D. REQ ↔ AC Traceability

Every REQ maps to ≥1 AC; every AC maps to a REQ. (13 REQs, 8 AC criteria covering 10 AC rows incl. the a/b/c sub-criteria.)

| REQ | Description | AC(s) |
|-----|-------------|-------|
| REQ-INV-001 | Command exists + registered via `newInventoryCmd()` / `rootCmd.AddCommand` | AC-INV-001 |
| REQ-INV-002 | Single top-level command (not subcommand group / not `status` extension) | AC-INV-001 |
| REQ-INV-003 | Composes 3 surfaces via existing exported functions | AC-INV-002 |
| REQ-INV-004 | Count + key-field summary per surface | AC-INV-002 |
| REQ-INV-005 | `--json` emits structured `UnifiedInventoryReport` (3 top-level keys) | AC-INV-003 |
| REQ-INV-006 | Human-readable default emits compact summary | AC-INV-004 |
| REQ-INV-007 | Empty/absent surface → 0-count, no error | AC-INV-005a |
| REQ-INV-008 | Genuine per-surface error degrades, not whole-command abort | AC-INV-005b, AC-INV-005c |
| REQ-INV-009 | Project-root resolution via `resolveProjectRoot(cmd)` | AC-INV-006a |
| REQ-INV-010 | Out-of-project invocation → sessions/harnesses 0-count + worktree error-degradation | AC-INV-005a, AC-INV-005c |
| REQ-INV-011 | Zero backing-package modification | AC-INV-006b |
| REQ-INV-012 | No `--json` flag added to `moai worktree list` | AC-INV-006b |
| REQ-INV-013 | No `AskUserQuestion` (subagent boundary) | AC-INV-007 |

Reverse direction (every AC → REQ):

| AC | REQ(s) covered |
|----|----------------|
| AC-INV-001 | REQ-INV-001, REQ-INV-002 |
| AC-INV-002 | REQ-INV-003, REQ-INV-004 |
| AC-INV-003 | REQ-INV-005 |
| AC-INV-004 | REQ-INV-006 |
| AC-INV-005a | REQ-INV-007, REQ-INV-010 |
| AC-INV-005b | REQ-INV-008 |
| AC-INV-005c | REQ-INV-008, REQ-INV-010 |
| AC-INV-006a | REQ-INV-009 |
| AC-INV-006b | REQ-INV-011, REQ-INV-012 |
| AC-INV-007 | REQ-INV-013 |

No orphan REQs (all 13 covered). No orphan ACs (all 10 rows map to a REQ).
