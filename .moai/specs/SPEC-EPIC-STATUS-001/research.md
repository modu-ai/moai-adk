# Research — SPEC-EPIC-STATUS-001

> Read-only prior-art record. All claims are baseline-attributed to HEAD `9fa242ddae3e5c7e9a80c2b47bd03d38b4c1b5ed` on branch `feat/factory-bootstrap-guidance`. This is a Tier L research artifact — it documents the measured building blocks the producer reuses, so the run-phase implementation does not re-derive them.

---

## §1. The two parallel Explore audits (2026-08-11, prior to this SPEC)

Two read-only Explore audits were conducted on 2026-08-11 against HEAD `9fa242dda`. Their findings are the proximate cause of this SPEC:

### Audit 1 — Factory Mode gap

Finding: Factory Mode today is a chain-seed + manual-launch notification, NOT an epic orchestrator. The audit identified five gaps:
1. Backlog ingest — MISSING.
2. Milestone (M1..M6) management — MISSING.
3. Multi-session dispatch — PARTIAL (SessionStart notice only).
4. Progress visibility — PARTIAL (per-SPEC only, no epic-level view).
5. Cross-/clear epic continuity — MISSING.

Gaps 1, 2, 3 (full), and 5 are owned by the KANBAN BOARD / WORKTREE / BOOTSTRAP SPEC family (all `status: draft`, code 0 at HEAD). Gap 4 (progress visibility) is what this SPEC addresses — it is the first concrete step toward closing Gap 5 (continuity) for free, because a disk-grounded producer survives `/clear`.

### Audit 2 — Visibility machinery gap

Finding: there is NO Go producer that renders an epic-level status map. The `🎯 BAS Epic ▓▓▓▓▓░░░░░ 3/6` surface is output-style render-time prose emitted from the agent's working memory; it is NOT computed from any SSOT. A fresh session after `/clear` cannot ask "where is the epic right now?" and get a disk-grounded answer.

The audit identified the reusable assets (ListDocs + Audit + board.go pattern + banner templates + design report) that this SPEC composes — measured in §2 below.

---

## §2. Reusable assets (measured at HEAD `9fa242dda`)

### 2.1 `internal/spec.ListDocs` — the catalog lister

**File**: `internal/spec/listdocs.go:36`

**Signature**: `func ListDocs(baseDir string) ([]DocRecord, error)`

**Behavior (measured by reading the source)**:
- Walks `<baseDir>/.moai/specs/SPEC-*/spec.md`.
- Returns parsed frontmatter records (`DocRecord{Path, Frontmatter, ParseError}`), sorted by path.
- Missing specs dir → empty slice, NOT an error (graceful-empty contract).
- Observation-only: the source explicitly states "it MUST NOT mutate any file" (lines 14-16).
- Per-record parse error: one malformed `spec.md` does NOT abort the whole scan (lines 20-25).

**Why this matters for the producer**: `ListDocs` is the exact pure-FS catalog lister the producer needs. No new walk logic is required; the producer filters the returned slice by `prefix`.

### 2.2 `internal/spec.Audit` — the drift + era scanner

**File**: `internal/spec/audit.go:156`

**Signature**: `func Audit(opts AuditOptions) (*AuditResult, error)`

**Behavior (measured by reading `internal/spec/CLAUDE.md` + the audit source)**:
- Returns `AuditResult` carrying `DriftFindings` (per-SPEC MUST-FIX findings) and per-SPEC era classification.
- Per-SPEC `sync_commit_sha` presence is the H-4 era-classifier predicate (lifecycle-sync-gate.md § H-4): `progress.md §E.2 + §E.4 + sync_commit_sha SHA value` → V3R6.
- The git-dependent drift-scan path is deliberately never invoked by `board.go` (REQ-WC11-045); the producer inherits the same constraint.

**Why this matters for the producer**: `Audit` is the source of `sync_commit_sha` presence — the signal that distinguishes `done` (SPEC `status: completed` AND non-empty `sync_commit_sha`) from `in-progress` (status set but not yet closed). The producer reads `sync_commit_sha` from the `progress.md` §E.4 section directly via `Audit`'s output, NOT via a fresh parse of `progress.md`.

### 2.3 `internal/web/board.go buildBoardView` — the composition pattern

**File**: `internal/web/board.go:88-131`

**Behavior (measured by reading the source)**:
- Composes `spec.ListDocs(cfg.ProjectRoot)` + `spec.Audit(AuditOptions{BaseDir: cfg.ProjectRoot})` into a `boardView` view-model.
- Builds `StatusCounts` (status-distribution summary) + `CloseDebt` (implemented-not-completed list) + `MustFix` (drift findings).
- The composition pattern is: `ListDocs` for per-SPEC frontmatter + `Audit` for drift/sha signals; both take the project root; both are observation-only.

**Why this matters for the producer**: the producer reuses this EXACT composition pattern (REQ-ES-002). It does NOT fork `buildBoardView`; it does NOT duplicate the scanner calls; it composes the same two calls into a different view-model (`EpicStatus` instead of `boardView`).

### 2.4 Banner templates (FROZEN grammar)

**Files**: `.claude/output-styles/moai/moai.md:572-610` (Epic Status), `:636-680` (Progress Board)

**Key facts (measured by reading the source)**:
- `done` = count of `🟢` items; `total` = all tracked items.
- Aggregate bar: 10-cell `▓` × `round(done ÷ total × 10)` + `░` remainder; then `done/total (pct%)` on the SAME line.
- Status icons `🟢🟡⬜⏸️🔵❌🔴` are structural — never translated.
- Header text translates per `conversation_language`; 4-locale table at `moai.md:593-599`.
- `[HARD]` rules at `moai.md:602-609` pin the Epic Status banner grammar.

**Why this matters for the producer**: the producer's human-mode output mirrors this grammar exactly (REQ-ES-007). The producer's JSON shape is the data feed for these banners; banner wiring is a separate concern (§C out of scope).

### 2.5 Design report — BAS epic canonical Mx list

**File**: `.moai/reports/navigator-redesign-bas-20260805.html` §7 (slice table)

**Mx list (measured by reading the source)**:
- M0 그래프 결합층 (graph join layer) — Tier L
- M1 Detect 훅 — Tier M
- M2 Route 승격 — Tier M
- M3 Fix 증분 갱신 — Tier L
- M4 4-tier 지도 — Tier L
- M5 Brownfield 역추출 — Tier M

**Why this matters for the producer**: this is the canonical list against which `orphan_mx` is computed. At HEAD, the three Navigator-Sync SPECs cover M0/M1/M4 only; M2/M3/M5 are orphans (the user's "BAS is 3/6 done" claim, with M2 separately noted as in-flight in another session's locked worktree — a state invisible to the producer per acceptance.md §F edge case + KI-5).

### 2.6 `(BAS Mx)` title-marker convention

**Measured at HEAD against the three Navigator-Sync SPECs**:

```
.moai/specs/SPEC-NAVIGATOR-SYNC-001/spec.md:
  title: "Navigator Sync (BAS M0) — SSOT binding-token trio + graph-join schema layer …"

.moai/specs/SPEC-NAVIGATOR-SYNC-002/spec.md:
  title: "Navigator Sync (BAS M4) — 4-tier addressable map …"

.moai/specs/SPEC-NAVIGATOR-SYNC-003/spec.md:
  title: "Navigator Sync (BAS M1) — Falconer Detect: PostToolUse changed-path …"
```

The convention is `(<TOKEN> M<N>)` at the start of the title's em-dash prefix. The producer's title regex (`\(([A-Z][A-Z0-9-]*)\s+M(\d+)\)`) captures both the token and the M-number.

### 2.7 `moai spec status` CLI verb (registration pattern)

**File**: `internal/cli/spec.go:10-39`

**Pattern (measured)**:
```go
func newSpecCmd() *cobra.Command {
    specCmd := &cobra.Command{ Use: "spec", GroupID: "tools", ... }
    specCmd.AddCommand(newSpecStatusCmd())
    specCmd.AddCommand(newSpecAuditCmd())
    ...
    return specCmd
}
func init() { rootCmd.AddCommand(newSpecCmd()) }
```

The producer mirrors this pattern at `internal/cli/epic.go`:
```go
func newEpicCmd() *cobra.Command {
    epicCmd := &cobra.Command{ Use: "epic", GroupID: "tools", ... }
    epicCmd.AddCommand(newEpicStatusCmd())
    return epicCmd
}
func init() { rootCmd.AddCommand(newEpicCmd()) }
```

---

## §3. What does NOT exist at HEAD (negative findings — measured)

These negative findings are recorded so the run-phase implementation does NOT waste time looking for them:

- **No `epic.json` or `epic-state.json`** anywhere under `.moai/` or `internal/`. Confirmed by `find . -name 'epic*.json'` returning zero hits.
- **No `internal/epic/` package**. Confirmed by `ls internal/ | grep epic` returning zero hits.
- **No `moai epic` CLI verb**. Confirmed by `grep -rn 'newEpicCmd\|"epic"' internal/cli/` returning zero hits.
- **No production `nav-graph.json` on disk** outside test fixtures. Confirmed by `find . -name 'nav-graph.json'` returning only `./internal/hook/testdata/navigator-detect-corpus/nav-graph.json` (a navigator-detect test fixture); no production nav-graph.json exists in either the primary checkout or this worktree.
- **No `epic`/`milestone` node type in the navigator schema**. The schema (`.claude/rules/moai/workflow/nav-tokens.md` § Output) carries only `decision | spec | symbol`.

---

## §4. Cross-references

- `internal/spec/CLAUDE.md` — package conventions for `internal/spec` (the producer is a downstream consumer, NOT a contributor to this package).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum — the 8-value enum the producer reads.
- `.claude/rules/moai/workflow/lifecycle-sync-gate.md` § H-4 — the era-classifier predicate that defines `sync_commit_sha` semantics.
- `.claude/rules/moai/core/verification-claim-integrity.md` §2 — baseline-attribution contract the producer follows.
- `.moai/specs/SPEC-FACTORY-MODE-001/` — predecessor; the producer is the next step in the Factory Mode lineage.
- `.moai/specs/SPEC-FACTORY-BOOTSTRAP-001/` — sibling; carries the companion-session bootstrap that the producer's factory touchpoint (REQ-ES-012) names as a future wiring target.
- `.moai/specs/SPEC-KANBAN-BOARD-001/`, `SPEC-KANBAN-WORKTREE-001/`, `SPEC-KANBAN-BOOTSTRAP-001/` — out-of-scope siblings owning the backlog/card/dispatch/quorum surface.
