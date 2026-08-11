---
id: SPEC-EPIC-STATUS-001
title: "Epic Status producer (moai epic status <prefix>) — disk-grounded epic progress map + banner SSOT data feed"
version: "0.2.0"
status: completed
created: 2026-08-11
updated: 2026-08-12
author: manager-spec
priority: High
phase: "v3.2.0 target"
module: "internal/epic, internal/cli"
lifecycle: spec-anchored
tags: "epic, progress, visibility, cli, producer, banner-ssot, factory-mode, json"
tier: L
related_specs: [SPEC-FACTORY-MODE-001, SPEC-FACTORY-BOOTSTRAP-001, SPEC-KANBAN-BOARD-001]
---

## HISTORY

- **v0.2.0** (2026-08-12) — Plan-audit iteration-1 revision. Closed 4 blocking defects (D1: REQ-ES-011 rewritten with shall as the lead sentence, justification moved into a parenthetical; D2: REQ-ES-012 "MAY" converted to the GEARS-canonical "shall"; D3: REQ-ES-008 §B.1 `orphan_mx` JSON example corrected from `["M2","M5"]` to `["M2","M3","M5"]` to match AC-ES-005 + design report §7; D4: `orphan_mx` null-vs-omit locked to the omit-when-empty form across REQ-ES-005 / design.md §5 / AC-ES-005b). Tightened 2 optional findings (D5: `nav-graph.json` negative-finding scoped to "no production nav-graph.json outside test fixtures" across research.md §3 + spec.md §C; D6: REQ-ES-008 §B.1 example now carries an inline `covered: false` orphan entry so the orphan shape is visible in the frozen contract without cross-referencing REQ-ES-005). All citation + audit-clean dimensions preserved (MP-1/3/5/6/7, traceability, testability, completeness).

- **v0.1.0** (2026-08-11) — Initial plan-phase authoring. This SPEC is the first concrete step toward making Factory Mode a real epic orchestrator rather than a chain-seed + manual-launch notification. Two parallel read-only Explore audits (Factory Mode gap + visibility-machinery gap, conducted 2026-08-11 against HEAD `9fa242ddae3e5c7e9a80c2b47bd03d38b4c1b5ed`) established that:
  - Factory Mode today is a chain-seed + companion-entry notification, NOT an epic orchestrator (backlog ingest / milestone management / cross-/clear epic continuity all MISSING — owned by the KANBAN BOARD / WORKTREE / BOOTSTRAP SPECs).
  - There is NO Go producer that renders an epic-level status map. The `🎯 BAS Epic ▓▓▓▓▓░░░░░ 3/6` surface (`.claude/output-styles/moai/moai.md:636-680` Progress Board grammar + `:572-610` Epic Status) is output-style render-time prose emitted from the agent's working memory — NOT computed from any SSOT. A fresh session after `/clear` cannot ask "where is the epic right now?" and get a disk-grounded answer.
  This SPEC closes that gap by delivering a disk-grounded producer whose `--json` output is the banner-SSOT data feed. The backlog/card store / dispatch protocol / quorum / `/clear` epic-state continuity are explicitly out of scope (§C) — they are the KANBAN BOARD SPEC's territory. This SPEC does NOT create `epic.json` or a kanban board; it derives the epic map from existing on-disk signals (SPEC dirs + title markers + optional design report).

---

## §A. Context

### A.1 Problem — orphaned epic progress across `/clear` + handoff

The user's core complaint (verbatim intent, two parallel audits 2026-08-11): epic progress gets orphaned across `/clear` + handoff (the BAS Epic is 3/6 done — M0/M1/M4, with M2 in-flight in another session's locked worktree and M3/M5 unauthored), and even inside a Factory run, "깊게 들어오면 상위에서 어디까지 진행했는지 알기 힘들다" (once you go deep, it's hard to see how far the parent epic has progressed).

Today the only "epic progress" surface is the Progress Board / Epic Status banner that an agent renders from its own working memory at output time. That surface is fragile in two directions:
- **Forward fragility**: a `/clear` or session resume drops the working-memory accumulator, so the next session cannot answer "where are we in the epic" without re-reading every SPEC dir and re-deriving the Mx map by hand.
- **Lateral fragility**: a sibling session (locked worktree, mid-milestone) holds state the orchestrator session cannot see — the banner silently undercounts in-flight work.

### A.2 The reusable read path (measured at HEAD `9fa242dda`)

The producer does NOT invent a new scanner. It composes two existing pure-FS scanners that already power the read-only web board (`internal/web/board.go:88-131` `buildBoardView`):

- `internal/spec.ListDocs(baseDir string) ([]DocRecord, error)` — `internal/spec/listdocs.go:36`. Walks `<baseDir>/.moai/specs/SPEC-*/spec.md`, returns parsed frontmatter records sorted by path, observation-only (NEVER mutates). Missing specs dir → empty slice, NOT an error.
- `internal/spec.Audit(opts AuditOptions) (*AuditResult, error)` — `internal/spec/audit.go:156`. Returns drift findings including per-SPEC `sync_commit_sha` presence (the 3-phase close signal the era classifier reads at `internal/spec/era.go` `ClassifyEra` H-4 predicate).

The web board's `buildBoardView` already composes these two into a per-SPEC view-model (`boardView` with `StatusCounts` + `CloseDebt` + `MustFix`). The producer reuses the same composition pattern — it does NOT fork `board.go`, does NOT duplicate the scanner calls, and does NOT invoke the git-dependent drift-scan path (per `board.go` REQ-WC11-045 contract).

### A.3 The Mx marker convention (measured)

Every Navigator-Sync SPEC self-declares its milestone in its `title:` field, in the shape `(BAS Mx)`. Measured at HEAD against the three Navigator-Sync SPECs:

```
title: "Navigator Sync (BAS M0) — SSOT binding-token trio + graph-join schema layer …"
title: "Navigator Sync (BAS M4) — 4-tier addressable map …"
title: "Navigator Sync (BAS M1) — Falconer Detect: PostToolUse changed-path …"
```

A regex over `spec.ListDocs` output's `Frontmatter.Title` field derives the SPEC→Mx map with NO new authoring convention. The design report at `.moai/reports/navigator-redesign-bas-20260805.html` §7 (slice table) carries the canonical M0..M5 list for the BAS epic. Both signals are disk-readable; the producer consults both, with explicit precedence (§A.5).

### A.4 The banner grammar (FROZEN — producer must match, not invent)

The Epic Status and Progress Board banner templates live at `.claude/output-styles/moai/moai.md:572-610` (Epic Status) and `:636-680` (Progress Board). Their grammar is FROZEN per the `[HARD]` rules in that file:

- `done` = count of `🟢` items; `total` = all tracked items.
- Aggregate bar: 10-cell `▓` × `round(done ÷ total × 10)` + `░` remainder; then `done/total (pct%)` on the SAME line as the heading.
- `🎯 phase position` ∈ `{entry, mid, closing}`.
- `📋 Current SPEC` carries SPEC-ID + Tier + phase + milestone position (e.g. `M3/M6`).
- 4-locale i18n translation tables already exist (`Epic progress:` / `에픽 진행:` / `エピック進行:` / `史诗进度:`).

The producer's `--json` shape is the data feed for these banners. Wiring the banner itself to consume the producer is a separate run-phase concern (this SPEC delivers the producer + the shape; banner wiring is named as a follow-up in §C).

### A.5 Epic-discovery precedence (design decision — locked in design.md §3)

An "epic" is identified by a SPEC-ID prefix (e.g. `SPEC-NAVIGATOR-SYNC-*`) OR a `(<EPIC-TOKEN> Mx)` marker family in titles (e.g. `(BAS Mx)`). The producer derives the epic's milestone set in three strategies, with strict precedence:

1. **Prefix glob** — `<prefix>` arg matches SPEC dirs (`SPEC-NAVIGATOR-SYNC-*`). The prefix is the primary epic selector.
2. **Title regex** — across the prefix-matched SPEC set, the producer extracts `(<EPIC-TOKEN> Mx)` markers to derive the Mx→SPEC map. `<EPIC-TOKEN>` defaults to the prefix's terminal domain segment uppercased (`NAVIGATOR-SYNC` → first match wins per SPEC), but is overridable via `--marker <token>` for epics whose prefix and marker differ.
3. **Design-report §6/§7 canonical list** — when a design report exists for the epic (located via a discovery rule documented in design.md §4 — `.moai/reports/<basename>-<epic-token-lower>-*.html` or an explicit `--design-report <path>` flag), the producer parses the Mx list and uses it as the canonical milestone set. SPECs with marker Mx NOT in the canonical list are flagged `extra-Mx`; canonical Mx NOT covered by any SPEC are flagged `orphan` (the BAS M3/M5 case).

When no design report is found, the producer falls back to the union of marker-derived Mx across the matched SPEC set (no orphans are reported in this mode — orphan detection requires a canonical list to orphan against).

---

## §B. Requirements (GEARS notation)

### REQ-ES-001 (Ubiquitous) — Disk-grounded epic progress producer

The `moai epic status` CLI shall compute epic progress exclusively from on-disk signals (`.moai/specs/SPEC-*/spec.md` frontmatter via `spec.ListDocs` + `spec.Audit` + optional design report under `.moai/reports/`), and shall NOT depend on agent working memory, conversation transcript, or runtime session state.

### REQ-ES-002 (Ubiquitous) — Reuse the ListDocs + Audit read path (no scanner fork)

The producer shall source per-SPEC data via the existing exported `spec.ListDocs` + `spec.Audit` functions (`internal/spec/listdocs.go:36`, `internal/spec/audit.go:156`), reusing the `buildBoardView` (`internal/web/board.go:88-131`) composition pattern. The producer shall NOT duplicate the scanner logic, shall NOT invoke the git-dependent drift-scan path, and shall NOT mutate any file (observation-only contract).

### REQ-ES-003 (Event-driven) — Epic discovery by prefix glob

**When** the user invokes `moai epic status <prefix>`, the producer shall match SPEC directories whose ID starts with `SPEC-<prefix>` (case-sensitive, uppercased), and shall return the matched set as the candidate epic population. An empty match set shall produce an empty epic map with a clear human-mode message ("no SPECs matched prefix `<PREFIX>`") and a zero-row `--json` output, NOT an error exit.

### REQ-ES-004 (Event-driven) — Mx extraction by title regex

**When** the prefix-matched set is non-empty, the producer shall scan each SPEC's `Frontmatter.Title` for a `(<TOKEN> M<N>)` marker (regex `/\(([A-Z][A-Z0-9-]*)\s+M(\d+)\)/` — uppercase token, single-digit M-number for M0..M9, multi-digit for M10+), and shall build the Mx→SPEC map. SPECs whose title carries no marker shall be classified `untracked` (still listed in `--json` under `untracked_specs`, but not assigned to any Mx).

### REQ-ES-005 (State-driven) — Orphan-Mx detection against the canonical list

**While** a design-report canonical Mx list is available (located via the design-report discovery rule in design.md §4, or supplied via `--design-report <path>`), the producer shall classify each canonical Mx as `covered` (≥1 SPEC marks it) or `orphan` (zero SPECs mark it), and shall emit `orphan_mx: [...]` in `--json` output. **While** no design report is available, the producer shall omit `orphan_mx` from `--json` (orphan detection requires a canonical list to orphan against), and shall emit the marker-derived Mx union only.

### REQ-ES-006 (Event-driven) — Status join per SPEC

**When** the producer builds each Mx entry, it shall join the Mx→SPEC map with `spec.ListDocs` frontmatter `status:` (the 8-value enum: draft / planned / in-progress / implemented / completed / superseded / archived / rejected) and `spec.Audit`'s per-SPEC `sync_commit_sha` presence signal, classifying each Mx's coverage as one of: `done` (owning SPEC `status: completed` AND non-empty `sync_commit_sha`), `in-progress` (owning SPEC `status: in-progress` OR `implemented`), `planned` (owning SPEC `status: draft`), or `absent` (no owning SPEC, only meaningful under REQ-ES-005 orphan detection).

### REQ-ES-007 (Capability-gate) — Human + JSON output modes

**Where** the user passes `--json`, the producer shall emit a single JSON document on stdout matching the shape in §B.1 below, shall emit nothing else on stdout (diagnostics go to stderr), and shall exit 0 on success. **Where** the user omits `--json`, the producer shall emit a human-readable rendering that mirrors the Progress Board grammar (`🎯 <epic> ▓▓░░ 3/6 (50%)` plus per-Mx lines with status icons `🟢🟡⬜` per the icon legend at `moai.md:660-668`), translated to `conversation_language` per the 4-locale table at `moai.md:593-599`.

### REQ-ES-008 (Ubiquitous) — Banner-SSOT data shape (frozen contract)

The `--json` output shall conform to the following shape (additive-only forward-compat — fields kept verbatim once shipped):

```json
{
  "epic": "<PREFIX>",
  "epic_token": "<TOKEN>",
  "milestones": [
    {
      "id": "M0",
      "label": "<label from design report, or Mx if unavailable>",
      "status": "done|in-progress|planned|absent",
      "covered": true,
      "spec_id": "SPEC-NAVIGATOR-SYNC-001",
      "spec_status": "completed",
      "sync_commit_sha": "<sha or empty>"
    },
    {
      "id": "M2",
      "label": "<label from design report, or Mx if unavailable>",
      "status": "absent",
      "covered": false,
      "spec_id": "",
      "spec_status": "",
      "sync_commit_sha": ""
    }
  ],
  "done": 3,
  "total": 6,
  "pct": 50,
  "orphan_mx": ["M2", "M3", "M5"],
  "extra_mx": [],
  "untracked_specs": ["SPEC-NAVIGATOR-SYNC-EXTRA-001"],
  "design_report": ".moai/reports/navigator-redesign-bas-20260805.html",
  "baseline_attribution": "<HEAD SHA at producer run>"
}
```

### REQ-ES-009 (Ubiquitous) — CLI verb registration under the `epic` family

The producer shall register a new top-level `moai epic` parent command with a single `status` subcommand (`moai epic status <prefix>`), registered in `internal/cli/epic.go` and surfaced via `rootCmd.AddCommand(newEpicCmd())` in `internal/cli/root.go`. The `epic` parent's `Short` help shall identify it as a disk-grounded epic progress producer. The `status` subcommand shall accept positional `<prefix>` (required) and flags `--json`, `--design-report <path>`, `--marker <token>`.

### REQ-ES-010 (Ubiquitous) — Non-interactive (no AskUserQuestion in CLI code)

The `moai epic status` CLI shall be non-interactive: read-only scan + print to stdout/stderr + exit. The CLI code path shall contain zero `AskUserQuestion` invocations and zero interactive confirmation prompts, per the C-HRA-008 orchestrator-only-question-channel boundary.

### REQ-ES-011 (Ubiquitous) — Template-neutrality (no SPEC-ID/SHA leak)

The producer source under `internal/epic/` and `internal/cli/epic.go` shall NOT carry SPEC-EPIC-STATUS-001 identifiers, REQ-ES tokens, audit citations, internal dates, or commit SHAs into `internal/template/templates/`. (Rationale: the producer source is Go code, not template source, so the template-neutrality CI guard at `internal/template/internal_content_leak_test.go` — scope `internal/template/templates/**` — does not apply to it; this requirement enforces neutrality defensively for any future mirror of `.moai/specs/SPEC-EPIC-STATUS-001/**` into the template tree.) This SPEC's plan-phase commits are markdown-only under `.moai/specs/SPEC-EPIC-STATUS-001/` and create no mirror.

### REQ-ES-012 (Capability-gate) — Factory integration touchpoint (light)

**Where** a Factory Mode lead/companion SessionStart notice is being composed and the project has at least one prefix-matched epic, the notice shall include a single pointer line `moai epic status <prefix>` so a factory session can surface epic context without leaving its current turn. The producer itself does NOT wire this notice; the wiring is owned by the Factory Bootstrap / Kanban Bootstrap SPEC family (the touchpoint is named here only to lock the producer's CLI surface so the wiring SPEC has a stable invocation target).

### REQ-ES-013 (Unwanted) — No persisted epic store

The producer shall NOT create, write, or require any new persisted epic store (`epic.json`, `epic-state.json`, etc.). The epic map is derived on-demand from existing on-disk signals on every invocation. (The kanban card store is the KANBAN BOARD SPEC's territory — §C.)

---

## §B.1 JSON shape contract notes

- **Additive-only**: once a field ships in a `moai` release, its name and type are frozen; later releases MAY add fields, never remove or rename. Consumers (banner, future factory dispatch) MUST tolerate unknown extra fields (forward-compat parse rule).
- **`baseline_attribution`**: the producer records the git HEAD SHA at run time (read via `git rev-parse HEAD`, fail-open to empty string in non-git contexts) so a downstream banner can attribute its progress claim per `verification-claim-integrity.md` §2.
- **`pct`**: integer 0..100, computed as `round(done ÷ total × 100)`; `total: 0` → `pct: 0` (avoid divide-by-zero; emit empty-progress state, NOT an error).

---

## §C. Out of Scope

### Out of Scope — Kanban board, backlog, dispatch, quorum, /clear epic-state continuity

- The backlog ingest / card store / dispatch protocol / quorum / `/clear` epic-state continuity are owned by `SPEC-KANBAN-BOARD-001`, `SPEC-KANBAN-WORKTREE-001`, and `SPEC-KANBAN-BOOTSTRAP-001` (all `status: draft`, code 0 at HEAD `9fa242dda`). This SPEC does NOT create `epic.json`, does NOT introduce a kanban board data model, and does NOT define a dispatch protocol. The producer's JSON is a derived view, NOT a write surface — it is read-only and recomputes on every invocation.

### Out of Scope — Banner wiring (consumption layer)

- Wiring the Epic Status / Progress Board banner template (`.claude/output-styles/moai/moai.md:572-610`, `:636-680`) to consume this producer's `--json` output is a separate run-phase concern. This SPEC delivers the producer + the frozen data shape; the banner continues to render from agent working memory until a follow-up SPEC wires the consumption layer. The data shape is locked here so the wiring SPEC has a stable contract to consume.

### Out of Scope — Navigator graph routing

- The navigator graph (`nav-graph.json`) does NOT exist as a production artifact on disk in either the primary checkout or this worktree at HEAD `9fa242dda` (the only on-disk `nav-graph.json` is the test fixture at `./internal/hook/testdata/navigator-detect-corpus/nav-graph.json`), and its schema (per `.claude/rules/moai/workflow/nav-tokens.md`) has no `epic` or `milestone` node type. The producer MUST read SPEC dirs + design reports directly via `spec.ListDocs` + `spec.Audit`, NOT route through the navigator graph. Adding an `epic`/`milestone` node type to the navigator graph is a separate concern owned by a future Navigator-Sync SPEC.

### Out of Scope — Auto-emitting the banner from every turn

- Auto-emitting the Epic Status / Progress Board banner on every turn (or on every `moai epic status` invocation in human mode triggering a banner render) is an output-style behavior, NOT a producer concern. The producer prints its human-mode output (which mirrors the banner grammar) but does NOT call into the output-style render path.

### Out of Scope — Multi-epic aggregation / cross-epic rollup

- Aggregating progress across multiple epics (e.g. "show me all active epics and their pct") is a follow-up concern. This SPEC delivers one-epic-per-invocation (`moai epic status <prefix>`). A future `moai epic list` (no prefix) subcommand MAY consume the same producer library to render multi-epic views.

### Out of Scope — Worktree-aware in-flight detection

- Detecting "Mx is currently in-flight in a sibling session's locked worktree" (the lateral-fragility hazard in §A.1) requires a worktree-state scan + session-registry query that this SPEC does NOT deliver. The producer classifies Mx state from the SPEC dir's own frontmatter only; a locked worktree with uncommitted status updates is invisible to the producer (acknowledged residual risk in acceptance.md §C). Worktree-aware state is owned by `SPEC-KANBAN-WORKTREE-001`.

### Out of Scope — Template mirror of this SPEC's artifacts

- This SPEC's artifacts under `.moai/specs/SPEC-EPIC-STATUS-001/**` are NOT mirrored to `internal/template/templates/.moai/specs/`. They are local-development artifacts specific to moai-adk-go's own epic tracking. The template-neutrality CI guard is not triggered because no mirror is created.

---

## §D. Constraints

- **Worktree-anchored**: all implementation happens in `/Users/goos/.moai/worktrees/kanban` (the current worktree on branch `feat/factory-bootstrap-guidance`). The producer MUST NOT cd to or write to the primary checkout at `/Users/goos/MoAI/moai-adk-go` (per CLAUDE.md § Worktree Isolation + main-checkout-branch-guard.md).
- **Read-only at runtime**: the producer is observation-only. It MUST NOT call `os.WriteFile`, `os.Remove`, `git commit`, `git push`, or any file-mutating primitive. The Go code path is bounded by the same contract as `spec.ListDocs` ("MUST NOT mutate any file").
- **Non-interactive**: zero `AskUserQuestion` invocations, zero TTY prompts (REQ-ES-010). The CLI exits cleanly with a non-zero code only on flag-parse errors or internal panics.
- **No new external dependencies**: the producer composes stdlib + the existing `internal/spec` package. No new Go module dependencies.
- **Baseline-attributed**: every progress claim in `--json` carries the `baseline_attribution` HEAD SHA; the producer does NOT report a number it did not measure against the current tree (per `verification-claim-integrity.md` §1.1 surface 1).

---

## §E. Quality Gates (TRUST 5)

- **Tested**: ≥85% coverage on `internal/epic/` (table-driven tests over the discovery + join + render pipeline; characterization test against the BAS epic fixture). CLI integration tests via `newEpicStatusCmd()` invocation (mirrors `spec_audit_test.go:124` pattern).
- **Readable**: clear naming (`EpicStatus`, `MilestoneEntry`, `DiscoverEpic`), godoc on exported symbols, English code comments per CLAUDE.local.md §3 + the `code_comments: en` setting.
- **Unified**: gofmt + goimports + golangci-lint clean; matches the existing `internal/cli/spec_*.go` file/func naming conventions.
- **Secured**: read-only surface (no input reaches a file-write primitive or a shell exec); the `--design-report <path>` flag is validated against a path-traversal allowlist (must resolve under `.moai/reports/` or be an absolute path the caller explicitly passed — never a relative path that escapes the project root).
- **Trackable**: Conventional Commits (`feat(SPEC-EPIC-STATUS-001): M1 ...`); plan-phase commit subject follows the canonical pattern in the Status Transition Ownership Matrix.

---

## §F. Cross-References

- `.claude/output-styles/moai/moai.md:572-610` — Epic Status banner template (FROZEN grammar).
- `.claude/output-styles/moai/moai.md:636-680` — Progress Board banner template (FROZEN grammar, icon legend, 4-locale i18n table).
- `internal/spec/listdocs.go:36` — `spec.ListDocs` (reused read path).
- `internal/spec/audit.go:156` — `spec.Audit` (reused read path).
- `internal/web/board.go:88-131` — `buildBoardView` (the composition pattern the producer reuses, NOT forks).
- `.moai/reports/navigator-redesign-bas-20260805.html` §7 — BAS epic canonical M0..M5 slice table.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` § Status Enum — 8-value enum the producer reads.
- `.claude/rules/moai/core/verification-claim-integrity.md` §1.1 + §2 — claim/evidence/baseline-attribution contract.
- `.moai/specs/SPEC-FACTORY-MODE-001/`, `SPEC-FACTORY-BOOTSTRAP-001/` — Factory Mode lineage (the producer is the first step toward making Factory Mode a real epic orchestrator).
- `.moai/specs/SPEC-KANBAN-BOARD-001/`, `SPEC-KANBAN-WORKTREE-001/`, `SPEC-KANBAN-BOOTSTRAP-001/` — out-of-scope siblings owning the backlog/card/dispatch/quorum surface.

---

## Owning SPEC

- This SPEC owns the epic-progress producer (`moai epic status <prefix>` CLI + `internal/epic/` library + the frozen JSON shape).
