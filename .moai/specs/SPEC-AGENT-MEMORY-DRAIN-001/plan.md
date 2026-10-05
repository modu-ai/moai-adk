# SPEC-AGENT-MEMORY-DRAIN-001 — Implementation Plan

> Tier M · 3 milestones · card t223 · plan-phase v0.2.0 (2026-09-02; plan-audit iter-1
> PASS-WITH-DEBT 0.94, fixes D1-D3/D5/D7/D8 applied)
> Recommended direction: **(c) write-time mirror + backfill** — see §A.3. The operator
> confirms the direction at the Implementation Kickoff gate; alternatives (a)/(b) and their
> measured rejections are recorded in `spec.md` §B.3.

## §A — Context

### A.1 — What exists today (measured on tree `c0c36c421`, 2026-09-02)

| Surface | State | Evidence |
|---|---|---|
| Worktrees | 186 registered | `git worktree list \| wc -l` → 186 |
| Trees with agent-memory content | 26 file-bearing (88 with skeleton dirs) | find under `.claude/worktrees`, maxdepth 3 |
| Orphaned files | 70 at snapshot (40 topic + 30 index; index breakdown: manager-develop 12, manager-spec 9, plan-auditor 6, sync-auditor/manager-lead/manager-docs 1 each) | filename classification; **drifting population** — re-count minutes later read 73 (42 topic + 31 index) |
| Mirrored to primary | 0 of 40 topic files | `comm -12 /tmp/t223-wt-topics.txt /tmp/t223-primary-topics.txt` → empty output, exit 0 |
| PostToolUse memory interception | Live in worktree sessions, observation-only | `internal/hook/post_tool.go` `runMemoryAudit` (lines ~571-620) |
| `moai memory` command family | `doctor`, `archive` — no `drain` | `moai memory drain` → `Unknown command "drain" for "moai memory".` exit 1 |
| Native auto-memory | NOT affected — worktrees share primary's store | 0 per-worktree `memory/` dirs under the profile's `projects/` |
| t209 concrete instance | Already lost — tree disposed, lesson absent from primary | `git worktree list \| grep -c t209` → 0; `find … '*nonexistent_passes*'` → no hit |

### A.2 — Disposal-path enumeration (why dispose-time drain is not the primary mechanism)

Moai-owned removal points (5): `prMergeCleanup` (`internal/cli/session_worktree_prmerge.go:149`),
`worktree clean --merged-only` (`internal/cli/worktree/clean.go:143` region), `clean --stale`
(`clean.go:273` region), `worktree done` (`done.go:84,182`), `worktree remove`
(`remove.go:60`). Outside moai code (unhookable): manual `git worktree remove`, the
session-end keep/remove prompt, and OS-level tree deletion. **t209 is gone without any
moai-owned sweep having removed it — consistent with death through an unhookable path**
(inference, not observed fact: the actual disposal path is unobservable post hoc; the
observed facts are the tree's absence and that the reaper's guards never fired on it). A
drain covering only the 5 moai points is therefore a partial fix measured against that
loss. The write-time mirror covers every path because it runs before any disposal can
occur.

### A.3 — Decision record: directions

| Direction | Verdict | Deciding evidence |
|---|---|---|
| (a) dispose-time drain in reaper paths | Rejected as primary | 2 unhookable disposal paths; t209 lost via one (§A.2) |
| (b) primary-path write wiring | Rejected (infeasible) | No CC-side override for the repo-local agent-memory path; `autoMemoryDirectory` covers the native store only |
| **(c) write-time mirror (PostToolUse) + `moai memory drain` backfill** | **Recommended** | Interception already exists and fires in worktrees; covers ALL disposal paths; backfill empties the 26-tree backlog immediately; REQ-WR-025 already names drain-then-dispose as the correct fix |

### A.4 — Reconciliation rules (the decision most likely to change — review first)

1. **Topic files**: copy `<wt>/.claude/agent-memory/<agent>/<topic>.md` →
   `<primary>/.claude/agent-memory/<agent>/<topic>.md`. If the destination exists with
   different content → write `<topic>.wt-<worktree-name>.md` instead. Never overwrite.
2. **Index files**: `MEMORY.md` is NEVER copied. After topic copies, each copied topic
   without an index line in the primary agent's `MEMORY.md` gains exactly one line,
   derived (in order) from: the worktree index line for that file → the file's frontmatter
   `description` → `name`.
3. **`_archive/` subdirectories**: copied under the same collision rule (audit trail
   preserved, per moai-memory.md "archive, never delete").
4. **`[[name]]` links**: name-based, survive the move unchanged; no path rewriting.
   Post-drain `moai memory doctor` reports dangling links as today.
5. **Skeleton-only trees** (62 of 88): reported, nothing copied.

## §B — Known Issues

- **Verb fallthrough masks absence**: `moai memory` subcommands currently fall through to
  help in some paths (recorded lesson: moai-todo verb fallthrough); the drain verb must
  exit non-zero on unknown flags and never silently preview-as-help. AC-AM-001 asserts on
  output shape, not exit code alone.
- **Bash-written memory is invisible to the mirror** (Write/Edit-only interception) —
  accepted blind spot, same class as the existing taxonomy audit; residue covered by
  operator-run drain (spec.md §G).
- **Index append race** under concurrent mirrors — tolerated; `moai memory doctor`
  detects, index rebuildable from topic files (measured frequency: 70 files/~5 months).
- **Primary store is on a different branch** than a worktree (primary often on `main`,
  card trees on develop lineage) — irrelevant: agent-memory is gitignored; mirror writes
  touch no tracked file.

## §C — Pre-flight (run-phase M1 entry checks)

1. `moai spec audit` on this SPEC — zero MUST-FIX drift.
2. Design-direction confirmation recorded from the Implementation Kickoff gate
   (mirror+backfill vs alternative) — open questions in §F discharged.
3. `internal/hook` and `internal/cli` package baselines green
   (`go test ./internal/hook/... ./internal/cli/worktree/...` on the card tree).
4. Fixture plan confirmed: all drain/mirror tests use `t.TempDir()` worktree-like
   directory pairs — no real sibling-tree access (the worktree session guard refuses it,
   and tests must not mutate live trees; CLAUDE.local.md §6 HARD).

## §D — Constraints

- Fail-open mirror (REQ-AM-005): hook exit 0 always; no blocking decision JSON on the
  mirror path; notices to stderr.
- Hardcoding rules (CLAUDE.local.md §14): path constants extracted; no inline
  `.claude/agent-memory` string scattered across packages — one shared constant/helper
  (candidate home: `internal/session` next to `IrreplaceableIgnoredEntries`, or a small
  `internal/memory` helper — M1 decision).
- Primary resolution: `git rev-parse --path-format=absolute --git-common-dir` → parent
  dir. Must handle the primary-itself case (common dir == own `.git` → no-op,
  REQ-AM-006).
- Scope discipline: SPEC-WORKTREE-REAPER-001 files untouched; no reaper guard changes.
- If moai-memory.md is edited in run/sync phase, mirror the edit to the template tree
  (Template-First, CLAUDE.local.md §2 HARD). Baseline intent: no rules edit required.

## §E — Self-Verification (plan-phase)

- SPEC ID regex check executed: `SPEC-AGENT-MEMORY-DRAIN-001` → `PASS` (Bash, verbatim).
- Uniqueness: `ls .moai/specs | grep -c AGENT-MEMORY` → 0 pre-existing.
- RED-now measurements recorded in `acceptance.md` §D with command + verbatim output +
  exit code + tree SHA `c0c36c421`, all measured in this run on this tree.
- Frontmatter validated against the canonical 12-field schema
  (`.claude/rules/moai/development/spec-frontmatter-schema.md`): id/title/version/status/
  created/updated/author/priority/module/lifecycle/tags + `phase: "v3.1.4"` (release
  target — corrected at v0.2.1 per the operator's kickoff decision that this card joins
  the v3.1.4 close, PR #1685; not a stage) + `tier: M`.
- Gaps: plan-phase cannot measure the mirror's runtime behavior (design-only criteria are
  run-phase fixture tests per verification-completeness §2 green-path cells). The v0.2.0
  `phase`-version assumption was resolved by the kickoff decision (v3.1.4, above).

## §F — Milestones (decision-reversibility ordered)

### M1 — Reconciliation core + `moai memory drain` backfill (High)

The highest-change-likelihood surface: the reconciliation rules (§A.4) and the CLI
contract. Deliver as one reviewable unit:

1. Shared helpers: agent-memory path constants, primary-root resolution from a worktree
   (no-op when already primary), copy-with-collision rule, index-line append.
2. `moai memory drain` subverb: enumerate registered worktrees → per-tree report →
   preview default / `--yes` apply / `--json` records (REQ-AM-007..010).
3. Fixture tests: two temp-dir trees + primary; collision, index append, `_archive`,
   skeleton-only, no-delete invariants (AC-AM-002/003/004/007/008).
4. Real-tree backfill run against the 26 file-bearing trees, `--json` output archived as
   run-phase evidence (AC-AM-002 green cell).

### M2 — Write-time mirror in the PostToolUse hook (High)

1. Extend the agent-memory interception in `internal/hook/post_tool.go` (the
   `runMemoryAudit` neighborhood): after the audit, if the session's project root is a
   worktree, copy the written `.md` to the primary store under the M1 reconciliation
   rules (REQ-AM-002/003/004). **Trigger anchoring (D7)**: the mirror's trigger predicate
   is the path segment `.claude/agent-memory/` within the tool-input `file_path` —
   anchored to that literal directory path, NOT the audit's existing unanchored
   `strings.Contains(normalized, "agent-memory/")` substring predicate (a file at e.g.
   `docs/agent-memory/x.md` must NOT trigger mirroring). The audit may keep its looser
   predicate; the mirror defines and uses the stricter one, extracted as the shared
   path-predicate constant from M1.
2. Fail-open wiring: every failure path is a stderr notice, hook exit 0 (REQ-AM-005);
   primary-session no-op (REQ-AM-006).
3. Unit tests with fixture `HookInput` JSON: mirror success, unresolvable primary
   (fail-open, exit 0), primary no-op, collision rename (AC-AM-005/006/009).

### M3 — Verification pass + docs note (Medium)

1. Post-backfill doctor pass on the primary store (orphans/dangling index lines from the
   append race window) — recorded as evidence, not a new mechanism.
2. moai-memory.md: add a short paragraph documenting the mirror + drain (worktree memory
   reaches primary) **with template mirror** if edited (CLAUDE.local.md §2 HARD) — the
   only potential template-file touch in this SPEC. The alternative branch — documenting
   it in README instead — carries its own cost: the README 4-locale set (README.ko.md
   primary + en/ja/zh derivation) has a **same-PR 4-locale parity obligation**
   (`.claude/skills/hns-oss-docs-readme-sync`), so that branch touches 4 files, not 1.
   Both outcomes carry their true cost; the operator picks at kickoff (open question 3).
3. Cross-check against SPEC-WORKTREE-REAPER-001: P2 guard still intact, no reaper file
   modified (`git diff --stat` evidence).

## §G — Anti-Patterns

- **Do not** add a drain call inside the reaper's removal paths "for safety" — it
  re-opens the multi-surface complexity this design exists to avoid, and the mirror
  already runs strictly earlier.
- **Do not** copy `MEMORY.md` files or merge index files wholesale — append-missing only
  (REQ-AM-004); wholesale merge is how clobbering happens.
- **Do not** delete from worktrees after drain (REQ-AM-009) — disposal is the reaper's
  act under its own guards.
- **Do not** let the mirror block or error a Write/Edit on any path (REQ-AM-005).
- **Do not** run drain tests against real sibling trees — fixtures only (worktree session
  guard refuses cross-tree git anyway).

## §H — Cross-References

- `spec.md` §B.3 — direction decision record; §G — out-of-scope rationale.
- `acceptance.md` §D — two-cell AC matrix, severities, traceability.
- SPEC-WORKTREE-REAPER-001 (completed) — REQ-WR-024/025; `internal/cli/worktree/clean.go`
  guard chain; `internal/session/ignored_content.go` shared verdict.
- `.claude/rules/moai/workflow/moai-memory.md` — taxonomy, index cap, archive rule.
- `.claude/rules/moai/core/moai-constitution.md` § Lessons Protocol — memory conventions.
- `.claude/rules/moai/development/verification-completeness.md` §2 — two-cell adoption.
