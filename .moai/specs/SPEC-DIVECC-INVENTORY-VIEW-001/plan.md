# SPEC-DIVECC-INVENTORY-VIEW-001 — Implementation Plan

> Tier M plan (3-artifact set). Run-phase notes only — this plan is NOT executed at plan-phase; manager-develop consumes it at run-phase.

---

## §A. Context

### A.1 Location & branch

- Project root: `/Users/goos/MoAI/moai-adk-go`
- Plan-phase branch: `main` (artifacts authored on main per Step 1 of `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline).
- SPEC artifact directory: `.moai/specs/SPEC-DIVECC-INVENTORY-VIEW-001/` (3 files: `spec.md`, `plan.md`, `acceptance.md`).

### A.2 Target module

- Primary new file: `internal/cli/inventory.go` (~150-200 LOC, the unified command + the `UnifiedInventoryReport` shaping).
- Companion test file: `internal/cli/inventory_test.go` (table-driven tests + the subagent-boundary static guard).
- One-line registration edit: `internal/cli/root.go` `init()` → `rootCmd.AddCommand(newInventoryCmd())`.

### A.3 PRESERVE list (DO NOT MODIFY)

The three backing packages are READ-ONLY for this SPEC. The unified view CALLS their existing exported functions; it changes none of them:

- `internal/session/` (entire package) — call `session.QueryActiveWork("")`; do NOT modify `registry.go`, the `Entry` struct, or any session file.
- `internal/cli/worktree/` (entire package) — call `WorktreeProvider.List()` (the `git.WorktreeManager` provider); do NOT add a `--json` flag to `moai worktree list`; do NOT modify `list.go` or `root.go`.
- `internal/cli/harness/` (entire package) — call `harness.ListHarnesses(projectRoot)`; do NOT modify `v4lifecycle.go` or the `HarnessEntry` struct.
- `internal/core/git/` (entire package) — `git.Worktree` struct + `WorktreeManager.List()` are read-only API surfaces; do NOT modify.

### A.4 EXTEND target

- `internal/cli/root.go` `init()` — a single `rootCmd.AddCommand(newInventoryCmd())` line is the only edit to an existing file.

### A.5 Existing infrastructure to reuse (do NOT re-implement)

| Need | Reuse | Source |
|------|-------|--------|
| Session enumeration | `session.QueryActiveWork(optSpecID string) ([]session.Entry, error)` | `internal/session/registry.go:249` |
| Worktree enumeration | `WorktreeProvider.List() ([]git.Worktree, error)` (var `WorktreeProvider git.WorktreeManager`) | `internal/cli/worktree/root.go:16`, `internal/core/git/worktree.go:76` |
| Harness enumeration | `harness.ListHarnesses(projectRoot string) ([]harness.HarnessEntry, error)` | `internal/cli/harness/v4lifecycle.go:81` |
| Project-root resolution | `resolveProjectRoot(cmd *cobra.Command) (string, error)` | `internal/cli/harness.go:95` |
| JSON emission | `json.MarshalIndent(...)` → `fmt.Fprintln(cmd.OutOrStdout(), string(out))` | `internal/cli/session.go:140-146` |
| Short-ID display | `shortID(id string) string` (first 8 chars) | `internal/cli/session.go:424` |
| Text rendering | `renderCard(title, content)` — the primary reuse target; it fits a 3-surface count summary (one card per surface with a count header + key-field rows). NOTE: `renderSummaryLine(ok, warn, fail)` does NOT apply — its passed/warnings/failed shape does not map onto session/worktree/harness counts. (live render.go exposes `renderCard`, `renderSuccessCard`, `renderInfoCard`, `renderSummaryLine`) | `internal/cli/render.go` |
| Grouped composite output reference | `runGroupedChecks` / grouped `checkGroup` rendering | `internal/cli/doctor.go` |

### A.6 plan-auditor verdict

- Tier M, plan-auditor PASS threshold = 0.80 (per `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier).

---

## §B. Known Issues (filtered to relevant categories)

> Section B per `.claude/rules/moai/development/manager-develop-prompt-template.md` § B. Filtered to the categories relevant to this CLI-only, single-new-file SPEC.

- **B1. Cross-platform build tags** — `internal/cli/inventory.go` uses only `encoding/json`, `fmt`, the three backing packages, and cobra. No `syscall` use is anticipated. Still verify `GOOS=windows GOARCH=amd64 go build ./...` passes (the worktree provider shells out to git, but the inventory command only calls its Go API).
- **B3. C-HRA-008 / Subagent boundary** — `internal/cli/inventory.go` MUST NOT call `AskUserQuestion` / `mcp__askuser`. Add a static-grep guard test `TestNewInventory_NoAskUserQuestion` mirroring `internal/cli/worktree/new_test.go` `TestNew_NoAskUserQuestion`.
- **B5. CI 3-tier awareness** — spec-lint (this SPEC's frontmatter + Out-of-Scope), golangci-lint (the new Go file), and Test (per OS) each fail independently. Distinguish NEW findings from baseline.
- **B6. spec-lint heading convention** — already satisfied: `spec.md` §C uses `### Out of Scope — <topic>` h3 sub-sections (not a bare h2).
- **B8. Working-tree hygiene** — do NOT touch runtime-managed files (`.moai/state/active-sessions.json` is read-only INPUT here; the command reads it but must never write it). Commit only `internal/cli/inventory.go`, `internal/cli/inventory_test.go`, the one-line `root.go` edit, and (run-phase) progress.md.
- **B9. Git commit + push self-perform (Hybrid Trunk)** — manager-develop commits + pushes within this SPEC scope (Conventional Commits, `feat(SPEC-DIVECC-INVENTORY-VIEW-001): M{N} <subject>`; never `--no-verify`).
- **B10. Untouched-paths PRESERVE** — the §A.3 PRESERVE list is binding. Parallel sessions are working on unrelated scope (`internal/web/`, `internal/mx/`, `.moai/config/sections/*.yaml`); do NOT touch those.

---

## §C. Pre-flight Check List (run-phase, before code change)

```bash
# 1. Branch + baseline
git branch --show-current
git rev-parse HEAD

# 2. Cross-platform build pre-check (baseline must be clean before edit)
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. Lint baseline (NEW vs pre-existing distinction)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. Confirm the three backing functions are still present + signatures unchanged
grep -n "func QueryActiveWork" internal/session/registry.go
grep -n "func ListHarnesses" internal/cli/harness/v4lifecycle.go
grep -n "func (w \*worktreeManager) List" internal/core/git/worktree.go
grep -n "var WorktreeProvider" internal/cli/worktree/root.go

# 5. Confirm no existing inventory command (greenfield)
grep -rn "newInventoryCmd\|\"inventory\"" internal/cli/ || echo "greenfield — no inventory command yet"
```

---

## §D. Constraints (DO NOT VIOLATE)

- PRESERVE: every file in `internal/session/`, `internal/cli/worktree/`, `internal/cli/harness/`, `internal/core/git/` (READ-ONLY; call exported functions only).
- Do NOT add a `--json` flag to `moai worktree list` (REQ-INV-012). Serialize worktree entries INTERNALLY in `internal/cli/inventory.go`.
- Do NOT introduce a new persistent registry / state file / schema (Out of Scope).
- Do NOT compose observability stats or cross-correlation (Out of Scope — MVP is count + key-field summary only).
- Do NOT call `AskUserQuestion` / `mcp__askuser` from CLI code (REQ-INV-013, C-HRA-008).
- Output streams: stdout = structured (JSON when `--json`, plain text otherwise); stderr = warnings / per-surface errors (`internal/cli/CLAUDE.md` § Output streams).
- Exit-code discipline: `os.Exit(0)` success (incl. all-empty report); non-zero ONLY when no surface could be rendered (REQ-INV-008); never `panic()`.
- Use `resolveProjectRoot(cmd)` for project-root resolution (REQ-INV-009); never `filepath.Join(cwd, userPath)` for absolute paths.
- Conventional Commits + `🗿 MoAI` trailer; never `--no-verify`, `--amend`, force-push.

---

## §E. Self-Verification Deliverables

When manager-develop reports completion it MUST include:

- **E1. AC PASS/FAIL matrix** — one row per AC-INV-NNN with the verification command + actual output (per `acceptance.md`).
- **E2. Cross-platform build** — `go build ./...` exit 0 AND `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- **E3. Coverage** — `go test -cover ./internal/cli/...` ≥ 85% for the new file's contribution (report the package delta).
- **E4. Subagent boundary grep** — `grep -rn 'AskUserQuestion\|mcp__askuser' internal/cli/inventory.go | grep -v "_test.go" | grep -v "//"` → no output.
- **E5. Lint status** — `golangci-lint run --timeout=2m`; NEW findings reported explicitly, pre-existing baseline marked separately.
- **E6. Branch HEAD + push state** — new commit SHAs + `git push` result.
- **E7. Blocker report** — structured (NEVER AskUserQuestion) if any user decision is needed.

---

## §F. Milestones (priority-based, no time estimates)

> Milestones are the within-SPEC ordered work steps (M1..M5) per `.claude/rules/moai/development/sprint-round-naming.md`. cycle_type=tdd (default; new code, RED-GREEN-REFACTOR).

### M1 — Report data shapes (RED → GREEN)

Define the `UnifiedInventoryReport` struct and per-surface sub-report structs in `internal/cli/inventory.go`. The JSON shape (REQ-INV-005):

```go
// UnifiedInventoryReport is the --json payload composing the three
// read-only inventory surfaces (REQ-INV-005). It owns NO data; every
// field is projected from an existing surface's exported function.
type UnifiedInventoryReport struct {
    Sessions  SessionInventory  `json:"sessions"`
    Worktrees WorktreeInventory `json:"worktrees"`
    Harnesses HarnessInventory  `json:"harnesses"`
}

type SessionInventory struct {
    Count   int                  `json:"count"`
    Entries []SessionSummaryRow  `json:"entries"`
    Error   string               `json:"error,omitempty"` // per-surface degradation (REQ-INV-008)
}
type SessionSummaryRow struct {
    SessionID string `json:"session_id"` // short form via shortID()
    SpecID    string `json:"spec_id"`
    Phase     string `json:"phase"`
}

type WorktreeInventory struct {
    Count   int                  `json:"count"`
    Entries []WorktreeSummaryRow `json:"entries"`
    Error   string               `json:"error,omitempty"`
}
type WorktreeSummaryRow struct {
    Branch string `json:"branch"`
    Path   string `json:"path"`
    HEAD   string `json:"head"` // short form (first 8)
}

type HarnessInventory struct {
    Count   int                 `json:"count"`
    Entries []HarnessSummaryRow `json:"entries"`
    Error   string              `json:"error,omitempty"`
}
type HarnessSummaryRow struct {
    Name            string `json:"name"`
    Domain          string `json:"domain"`
    ManifestMissing bool   `json:"manifest_missing"`
}
```

RED tests: marshal an `UnifiedInventoryReport` fixture, assert the three top-level keys (`sessions`/`worktrees`/`harnesses`) and the per-row key fields are present. (REQ-INV-004, REQ-INV-005.)

### M2 — Surface collectors (read-only, graceful degradation)

Implement three private collectors that call the existing exported functions and map their results into the sub-report structs:

- `collectSessions() SessionInventory` → `session.QueryActiveWork("")`; empty slice → `Count: 0` (REQ-INV-007); a genuine read error → `Error` field set, surfaces that succeed still render (REQ-INV-008).
- `collectWorktrees() WorktreeInventory` → `WorktreeProvider.List()`; `WorktreeProvider == nil` (git module unavailable) → `Error` set; mirror the nil-guard already in `internal/cli/worktree/list.go:27`.
- `collectHarnesses(projectRoot string) HarnessInventory` → `harness.ListHarnesses(projectRoot)`; `(nil, nil)` (no harness dir) → `Count: 0` (REQ-INV-007).

Each collector maps surface entries → summary rows using the key fields in REQ-INV-004 (`shortID` for session/HEAD short forms). RED tests cover the empty-surface and the nil-provider degradation paths. (REQ-INV-003, REQ-INV-007, REQ-INV-008.)

### M3 — Command factory + registration

Implement `newInventoryCmd() *cobra.Command` (`Use: "inventory"`, `GroupID: "tools"`, `Args: cobra.NoArgs`, a `--json` bool flag, optional `--project-root` via the inherited-flag pattern). The `RunE` resolves the project root via `resolveProjectRoot(cmd)` (REQ-INV-009), runs the three collectors, assembles the `UnifiedInventoryReport`, and branches on the `--json` flag. Register via `rootCmd.AddCommand(newInventoryCmd())` in `internal/cli/root.go` `init()`. RED test: command is discoverable on the root command tree; `Use == "inventory"`. (REQ-INV-001, REQ-INV-002, REQ-INV-009.)

### M4 — Output rendering (JSON + human-readable)

- JSON path (REQ-INV-005): `json.MarshalIndent(report, "", "  ")` → `fmt.Fprintln(cmd.OutOrStdout(), string(out))` (mirror `internal/cli/session.go:140-146`).
- Human-readable path (REQ-INV-006): a compact 3-surface summary using `renderCard(title, content)` from `internal/cli/render.go` (one card per surface — title = surface name + count header, content = key-field rows), grouped-output style referencing `internal/cli/doctor.go`. `renderCard` is the correct reuse target for a 3-surface count summary; `renderSummaryLine(ok, warn, fail)` does NOT apply (its passed/warnings/failed shape does not map onto session/worktree/harness counts). An all-empty report renders three `(0)` headers, not an error.

RED tests: `--json` output unmarshals to `UnifiedInventoryReport` with expected counts; default output contains the surface labels + counts. Out-of-project invocation (REQ-INV-010) yields an all-zero report with exit 0. (REQ-INV-005, REQ-INV-006, REQ-INV-010.)

### M5 — Boundary guard + invariant verification (REFACTOR + gates)

- Add `TestNewInventory_NoAskUserQuestion` static-grep guard (REQ-INV-013, mirroring `internal/cli/worktree/new_test.go`).
- Verify the zero-backing-modification invariant: `git diff --name-only` touches ONLY `internal/cli/inventory.go`, `internal/cli/inventory_test.go`, `internal/cli/root.go` (the one-line registration), and run-phase `progress.md` — NOT any file under `internal/session/`, `internal/cli/worktree/`, `internal/cli/harness/`, `internal/core/git/` (REQ-INV-011, REQ-INV-012).
- Run the full gate batch (build / windows-build / coverage / lint / boundary grep). REFACTOR for clarity while keeping tests green.

---

## §G. Anti-Patterns (do NOT do)

- **Modifying a backing package to expose a "convenience" aggregate function** — the unified view composes existing exported functions; it adds nothing to `internal/session`, `internal/cli/worktree`, or `internal/cli/harness` (REQ-INV-011).
- **Adding `--json` to `moai worktree list`** to "make composition easier" — REQ-INV-012 forbids it; serialize worktree entries internally instead.
- **Returning an error when a surface is empty** — empty is a valid 0-count result, not a failure (REQ-INV-007). Only a genuine backing error (nil provider, I/O failure) sets the per-surface `Error` field.
- **Erroring out the whole command when one surface fails** — render the surfaces that succeed; exit non-zero only when none could render (REQ-INV-008).
- **Composing observability stats or cross-correlation** — out of MVP scope; count + key-field summary only.
- **Calling `AskUserQuestion` from `inventory.go`** — C-HRA-008 violation (REQ-INV-013).
- **Over-scoping with watch/filter/pagination flags** — single point-in-time read only; flags beyond `--json` (+ inherited `--project-root`) are out of scope.

---

## §H. Cross-References

- `spec.md` (this directory) — §A verified ground-truth, §B GEARS requirements, §C Out of Scope.
- `acceptance.md` (this directory) — Given-When-Then AC matrix + REQ↔AC traceability.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Section A-E delegation template (Tier M REQUIRED).
- `internal/cli/CLAUDE.md` — CLI module conventions (subagent boundary, cobra registration, output streams, exit codes).
- `internal/cli/session.go` / `internal/cli/worktree/list.go` / `internal/cli/harness/v4lifecycle.go` — the three backing surfaces (READ-ONLY).
- `internal/cli/doctor.go` — composite grouped-output rendering reference for the human-readable default.
