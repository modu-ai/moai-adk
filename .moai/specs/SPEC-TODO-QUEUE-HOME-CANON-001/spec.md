---
id: SPEC-TODO-QUEUE-HOME-CANON-001
title: "Pin the todo queue to the HOME SQLite store as the single canonical source; consolidate remaining code-path statements and JSON remnants"
version: "1.0.0"
status: draft
created: 2026-09-13
updated: 2026-09-13
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/kanban + template skills"
lifecycle: spec-anchored
tags: "todo-queue, kanban, sqlite, home-canonical, path-resolution, docs-statement, t658"
tier: M
related_specs: [SPEC-TODO-SQLITE-001, SPEC-WEB-TODO-QUEUE-001, SPEC-TODO-HOME-TEMP-GUARD-001, SPEC-TODO-QUEUE-HOME-MERGE-001, SPEC-STATE-ANCHOR-001]
---

# SPEC-TODO-QUEUE-HOME-CANON-001 — HOME SQLite as the single canonical todo-queue source

## HISTORY

| Date | Version | Change |
|------|---------|--------|
| 2026-09-13 | 1.0.0 | Initial draft (card t658). Authoring only — no implementation in plan-phase. |

## 1. Background and Survey Evidence (tree b66789479, branch WT-home-queue-canonical)

t657 merged the queue DATA; this SPEC owns the CODE/DOCS consolidation that prevents
the queues from diverging again. An exhaustive survey of the launch tree was performed;
its results ground every requirement below.

### 1.1 Survey (a) — every production queue-path resolution point

| # | Location (file:line) | What it resolves to |
|---|---|---|
| A1 | `internal/kanban/todo_root.go:81` `ResolveTodoQueueRoot` | PURE root resolver: primary checkout of the git repository (`primaryCheckoutRoot` at :174, via `gitcore.ResolveGitDirs` common dir + `homestate.CanonicalProjectRoot`), else the launch base. Writes nothing. |
| A2 | `internal/kanban/todo_root.go:97` `ResolveTodoQueueRootAdopting` | Same value as A1 (one-line delegation). Kept distinct by contract, not by value. |
| A3 | `internal/kanban/state_dir.go:26` `StateDirForRoot` | THE directory layer: `~/.moai/db/<project-key>/todo` (home SQLite) for any non-temporary origin with a resolvable home; project-local `<root>/.moai/state/todo` (`projectStateDirForRoot` :62) ONLY under the temporary-origin guard (:29, `TempOriginReason`) or home-unresolvable fail-open (:33). `MOAI` home override env honored (:37). |
| A4 | `internal/kanban/state_dir.go:115` `resolveStateDir` | Adoption machinery over A3: picks current vs legacy directory, relocates once under lock (`relocateQueueArtifacts` :160). Legacy candidates enumerated at :130 (project-local todo, legacy kanban dir, legacy home keys). |
| A5 | `internal/kanban/state_dir.go:238` `BacklogPathForRoot` / `:246` `BacklogPathForRootAdopting` | Joins `backlog.json` (compat name, const :258) under A4's directory; the engine derives the sibling `backlog.db`. |
| A6 | `internal/kanban/state_dir.go:69` `RuntimeStateDirForRoot` | Session registries ONLY (companions/leads records). Doc-comment forbids backlog persistence. NOT a queue path. |
| A7 | `internal/cli/todo.go:74` `resolveTodoQueueRoot` | Command-path entry: `kanban.ResolveTodoQueueRootAdopting(resolveProjectDir())`; store built at :107 via A5. |
| A8 | `internal/cli/todo.go:113,125,134` | Landed-ref lookups through A1. |
| A9 | `internal/cli/todo_landed.go:152,262` | Root resolution + git invocation under A1. |
| A10 | `internal/cli/todo_autodone.go:188,474` | Root resolution via A7 seam; auto-done log under `RuntimeStateDirForRoot` (:468 — registry, not queue). |
| A11 | `internal/cli/graph.go:58` `todoQueueRootFn` | Test seam over A7; `liveQueueCards` reads via `todoBacklogPath` (A5). |
| A12 | `internal/web/todo_queue_read.go:33` `readTodoQueue` | Console's SINGLE read seam: A1 (pure) + A5 + `LoadPure`. `todo_queue_read_test.go` asserts mechanically that no other internal/web file names a backlog-store symbol. |
| A13 | `internal/web/events.go:200` | Watch path keyed on A3's directory (`"kanban"` label) — watches the home SQLite dir for the standard git-repo case. |
| A14 | `internal/statusline/backlog.go:50` `resolveBacklogCounts` | **CONVERGED** — delegates to `kanban.BacklogCountsForRoot` (`backlog_store.go:496`); the direct `boardRoot/.moai/state/kanban/backlog.json` read named in the card's premise was removed by t306 (SPEC-TODO-SQLITE-001 M3+M4) / t510 (SPEC-STATE-ANCHOR-001). Card premise is STALE on this tree. |
| A15 | `internal/statusline/landed.go:210` | Reads through the seam: `kanban.NewBacklogStore(kanban.BacklogPathForRoot(boardRoot)).LoadPure()`. |
| A16 | `internal/hook/session_start_kanban.go:219` | `kanban.QueuedBacklogCountForRoot` — shared count seam, same home store. |
| A17 | `internal/cli/migrate_home_state.go:55,382,538` | One-time project→home `backlog.db` migration tool paths (adjacent, not a resolver). |

**Conclusion (a):** path resolution is ALREADY single-seam. One root resolver (A1/A2),
one directory layer (A3, `@MX:ANCHOR`), one path join (A5). The project-local path
survives only inside two deliberate branches of A3 — the temporary-origin guard
(SPEC-TODO-HOME-TEMP-GUARD-001) and the home-unresolvable fail-open — plus the
adoption candidates in A4.

### 1.2 Survey (b) — every production reader of the legacy backlog.json layout

| # | Location | Reads what | Sees home store? | Fate (disposition, see REQ-004) |
|---|---|---|---|---|
| B1 | `internal/kanban/backlog_store.go` `LoadPure` → `loadLegacyBacklogJSON` (`backlog_migrate.go:511`) | Legacy JSON when `backlog.db` absent — read-only serve | Yes — path derived via A3/A5 | **KEEP read-only** — pre-cutover readability + downgrade route; recorded justification REQUIRED |
| B2 | `internal/kanban/backlog_store.go` `openEngine` State B → `migrateUnderLock` → `migrateLegacyBacklog` | One-time migration read of legacy JSON (adopting path only) | Yes | **KEEP** — adoption window; retirement is SPEC-TODO-QUEUE-HOME-MERGE-001 M5 (NOT this SPEC) |
| B3 | `internal/kanban/backlog_export.go` / `internal/cli/todo_export.go:35-39` | WRITES legacy-format `backlog.json` beside the home db (export-json downgrade route) | Yes (home-scoped) | **KEEP** — documented downgrade route |
| B4 | `internal/kanban/state_dir.go:94` `queueExists` / `:160` `relocateQueueArtifacts` | Adoption reads of legacy layouts (project-local + legacy home dirs) | n/a (reads legacy locations by design) | **KEEP** until M5 retirement executes |
| B5 | `internal/kanban/backlog_archive_vouch.go:28-41` | Vouch-source classification (`BacklogStoreLegacyJSON`) | n/a (classification, not a path read) | **KEEP** — naming only |
| B6 | `cmd/t657-merge/main.go:49-53` | One-off merge utility: builds `<dir>/backlog.json` + `LoadPure` | Caller-supplied dir | **BOUNDARY** — merge CORE logic is t835's; disposition recorded as an exception pending t835 close-out. NOT converged by this SPEC. |
| B7 | `internal/cli/todo_disclosure.go:2,32,40` | Disclosure TEXT naming backlog.json ("NOT the queue") | n/a (prose only, no read) | **KEEP** — this IS the convergence messaging |
| B8 | `internal/statusline/backlog.go:50` (card premise) | Direct JSON read | Was project-local (WRONG location) | **ALREADY CONVERGED** on this tree (A14) — verify, no work |
| B9 | `internal/web/events.go:167-169`, `internal/web/screens_templ.go:1180` | Watch/render of `backlog.db-wal`/`-shm` | Yes | **KEEP** — SQLite-aware, not a JSON remnant |

**Conclusion (b):** ZERO production readers of the legacy path remain outside the
`internal/kanban` store/adoption machinery itself, one boundary card (B6), and the
deliberate export route (B3). Every in-machinery reader sees the home location for
the standard git-repo case.

### 1.3 The remaining gap — statements, not code

The one false canonical-store statement found on this tree (outside t704's owned
defect) is the kanban-foreman skill's queue-watch script:

- `internal/template/templates/.claude/skills/moai-kanban-foreman/SKILL.md:95` (and
  the byte-identical local copy `.claude/skills/moai-kanban-foreman/SKILL.md:95`,
  verified `diff -q` identical): the watch loop sets `d=.moai/state/todo` and
  checksums `$d/backlog.db`. For a standard git-repository project the queue lives
  at `~/.moai/db/<project-key>/todo/backlog.db` (A3), so this script watches a
  directory the resolution no longer uses — a silent no-op monitor.

Other template surfaces checked and found CORRECT (home-canonical):
`.moai/docs/todo-queue-storage.md` (header claim; the known-false migration-window
claim at its lines 21-27 is card t704's — NOT edited here),
`.claude/skills/moai/SKILL.md:170`, `.claude/skills/moai/workflows/todo.md:26`,
`internal/cli/todo.go:197` help text. No `.claude/rules/` text claims
`.moai/state/todo` as canonical.

## 2. Requirements (GEARS)

### REQ-001 — Statement convergence (foreman skill queue watch)

**When** the kanban-foreman skill's queue-watch script names a queue directory, the
script shall reference the canonical queue location as resolved by
`kanban.StateDirForRoot` — the home-scoped `~/.moai/db/<project-key>/todo` directory
for a standard git-repository project — and shall not name the project-local
`.moai/state/todo` path as the watch target.

### REQ-002 — Template-first parity

**Where** a template file under `internal/template/templates/` is edited, the local
copy (`.claude/...`) and the template mirror shall remain byte-identical, and the
catalog hashes shall be regenerated (`go run ./internal/template/scripts/gen-catalog-hashes.go --all`)
so `make build` embeds the corrected skill.

### REQ-003 — Single-resolution invariant (guard)

**While** the tree is at any commit after this SPEC's run phase, a machine-verifiable
guard test shall assert that, for a git-repository fixture, `kanban.StateDirForRoot(root)`
resolves to exactly one canonical directory (the home SQLite directory) and that no
production file under `internal/` or `pkg/` other than the `internal/kanban` package
constructs a `backlog.json` queue path outside the `BacklogPathForRoot` /
`BacklogPathForRootAdopting` seam (pattern: the existing mechanical assertion in
`internal/web/todo_queue_read_test.go`, widened repo-wide; `cmd/` one-off tools are
out of the guard's scope per B6).

### REQ-004 — Disposition ledger (JSON remnants)

**When** the run phase opens, every production reader of the legacy `backlog.json`
layout (survey §1.2, B1-B9) shall carry a recorded disposition — `converge` or
`keep-read-only`/`keep` with a written justification — in this SPEC's acceptance
ledger; a reader found at run-phase with no recorded disposition shall be either
converged to the seam or given a justification before closure. The dispositions
recorded in §1.2 are the baseline; the run phase verifies them, and any reader
discovered beyond B1-B9 is adjudicated under this requirement.

### REQ-005 — Retained fallbacks, recorded

The two project-local branches of `StateDirForRoot` — the temporary-origin guard
(state_dir.go:29) and the home-unresolvable fail-open (state_dir.go:33) — shall be
RETAINED. The temporary-origin branch is a tested safety property
(SPEC-TODO-HOME-TEMP-GUARD-001: temp projects must not touch the operator's home)
whose location truth is sibling cards t705/t706; the fail-open branch keeps the
queue usable without git metadata. This SPEC removes neither.

### REQ-006 — Run-phase discovery of new false statements

**When** the run phase discovers an additional template or local doc surface stating
`.moai/state/todo` (or `.moai/state/kanban`) as the canonical queue location —
beyond the foreman skill fixed under REQ-001 and excluding
`internal/template/templates/.moai/docs/todo-queue-storage.md` (t704's owned file) —
the surface shall be corrected in the same run phase under REQ-001/REQ-002 parity
rules, and recorded in the acceptance ledger.

## 3. Non-Functional Constraints

- No behavior change to queue resolution: the survey (§1.1) shows resolution already
  home-canonical; this SPEC must not alter A1-A6 return values for any input.
- `moai todo`, the web console, the statusline, and the session-start notice must
  continue reading the same store after the change (all already do — A7-A16).
- Test discipline: affected-package runs only (`go test ./internal/kanban/...`
  plus the package hosting the new guard); no full-suite local runs (CLAUDE.local.md §4).
- Template neutrality: skill edits stay free of SPEC IDs, internal dates, and
  commit SHAs (template-internal-isolation-doctrine C-classes).

## 4. Out of Scope

### Out of Scope — sibling-card boundaries (t658's four named siblings)

- t835: the merge-verifier defect (t204→t538 stale reference) and the t204/t718
  dedup. Merge CORE logic — including `cmd/t657-merge`'s census logic and its
  eventual retirement — is NOT this SPEC's; B6 is recorded as a boundary exception only.
- t705: the `StateDirForRoot` temp-git anomaly (a git repository living under a
  temp root). This SPEC records the guard's retention (REQ-005) and does not touch
  the anomaly's classification or fix.
- t706: factory queue-path truth under temporary origins. Factory-lane resolution
  semantics under temp origins are t706's.
- SPEC-TODO-QUEUE-HOME-MERGE-001 M5: project store RETIREMENT execution (deleting
  the legacy project-local queue and the adoption window). This SPEC records B2/B4
  as keep-until-M5; it executes no retirement.

### Out of Scope — t704's owned defect

- The known-false claim in `internal/template/templates/.moai/docs/todo-queue-storage.md`
  (its migration-window statement at lines 21-27) is card t704's exact edit. This
  SPEC's REQ-006 explicitly excludes that file; no duplicate edit here.

### Out of Scope — non-goals

- Any change to the SQLite store format, the export-json downgrade route (B3), or
  the `LoadPure` legacy serve (B1) — all retained with justification.
- docs-site (`content/`) queue-location pages — owned by the completed
  SPEC-DOCS-TODO-TEMP-GUARD-001 (t575).
- Data migration or queue-content changes — t657 owns the data plane, already merged.
