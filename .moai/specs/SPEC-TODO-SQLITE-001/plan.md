# SPEC-TODO-SQLITE-001 — Implementation Plan

Cycle type: DDD (ANALYZE-PRESERVE-IMPROVE) — the existing behavioral contract is
characterization-proven and must survive the storage swap byte-for-byte at the API
boundary. Milestones ordered by decision-reversibility: schema/finance-of-truth
decisions first (hardest to reverse later), mechanical sweeps last.

## §A Context

Primary worktree: card t306, branch `WT-todo-sqlite` (develop-based integration line;
NO per-card PRs — scoped local verification + the develop-worktree merge are the
gates). Ground truth, corrections, and driver evidence: research.md. Physical design:
design.md. The card absorbs t309 (directory rename) — treated throughout as
requirements B.3 + milestone M3, not a separate deliverable.

## §B Known Issues Being Addressed

- Whole-file read-modify-write persistence grows linearly with history retention
  (dropped/picked rows accumulate in the 118 KB live file).
- Statusline parses the entire queue document every render.
- Path name `kanban` mis-describes a queue owned by `moai todo` (t309 rationale);
  absorbed now because renaming around a schema change costs zero extra migrations.
- Every renderer hand-checks JSON keys it should not need to know about.

## §C Pre-flight (M0 checklist, blocking M1 start)

1. Pin the modernc.org/sqlite release current at run-phase start; record module graph
   impact (`go mod tidy`, `go.sum`) and its supported-GOOS documentation link in the
   run notes. Open question status: none remain — version pinning is a mechanical
   run-time lookup decided by evidence at pin time.
2. Baseline measurements BEFORE implementation lands: stripped-binary size of
   `bin/moai` build, statusline render timing with a generated ≥500-item fixture,
   package coverage baselines for changed-path files.
3. Confirm the repo-local mirror situation for the three template-referencing skills
   (template edit is canonical; local `.claude/skills` redeploy covered by Template-First
   cycle; dogfood machine sync follows standard update flow).
4. Verify scratch-root rehearsal scaffold (test root under `/tmp`, seeded synthetic
   fixture generator committed as test code only — C-5).

## §D Constraints carried forward

See spec.md §C (C-1…C-6). Two operational additions binding run-phase:

- Affected-package test scope only (`go test ./internal/kanban/...`,
  `./internal/cli/...`, `./internal/web/...`, `./internal/statusline/...`,
  `./internal/hook/...`) — NO full-suite local runs (load discipline; full-suite
  verdict belongs to CI).
- Directories outside `.moai/state/todo/` resolved lazily per-root; the migrator runs
  against whatever root a session opened — never sweep the global `~/.moai/todo/*`
  fleet.

## §E Self-Verification hooks

E1 AC matrix green (acceptance.md §D) · E2 windows cross-vet + CI matrix reading · E3
coverage thresholds C-1 · E4 literal-cleanliness greps (B.5 / REQ-TOSQ-018: zero
`state/kanban` outside intentional fallback-reader + historical SPEC docs + templates
post-edit; sweep checks BOTH spellings — joined literals `state/kanban` AND segment
joins of `".moai", "state", "kanban"`) · E5 concurrency stress suite output persisted
as evidence · E6 performance benchmarks logged against C-2/C-4 budgets.

## §F Milestones

### M1 — Driver adoption + storage skeleton (schema = most reversible NOW)

- Add modernc.org/sqlite; connection-manager wrapper (DSN, pragmas, open/lazy-close
  discipline, ping health).
- DDL executor + schema_version stamping; engine-error→domain-taxonomy mapping table.
- Deliver pre-flight measurement deltas (binary size, first-touch bench harness).
- Gate: wrapper unit-tested on darwin dev host; cross-vet darwin/linux/windows clean;
  measurements recorded.

### M2 — Store guts swap with characterization safety net (behavior-critical core)

- Port `Load/Add/Mutate/QueuedCount/QueuedBacklogCountForRoot` semantics onto SQL
  transactions UNDER the unchanged `BacklogStore` exported surface (REQ-TOSQ-010):
  file-lock acquire/deactivate path untouched, high-water-mark normalize ported
  verbatim, findings helpers (`AppendFindingOnce`, `FindingsNaming`,
  `RemoveFindingsNaming`, tuple predicates) operating on ordered SELECT results.
- Port the existing characterization suites to drive the NEW backing with expectations
  UNCHANGED wherever behavioral; extend error-path battery (design.md §9 row 5).
- Static contract guards stay green: bare-join convention test, web seam-import guard
  test.
- Gate: affected-package suites green INCLUDING the round-trip equivalence suite;
  coverage C-1 met.

### M3 — Migration machinery + directory rename (t309 absorbed here)

- State machine of design.md §4: lazy trigger detection, locked single-flight
  migration, parity verifier, quarantine flip, crash-window cleanup, partial-artifact
  removal on failure.
- Directory layer: atomic kanban→todo relocation incl. registry census checks;
  fallback-READ branch for relocation-refusing filesystems; pure-side observation for
  web/statusline (never adopt on read paths).
- Legacy-path centralization: `BacklogPathForRoot` sibling constant becomes THE
  todo-layout anchor; old-layout references only inside the fallback reader.
- Gate: migration matrix (design.md §9 row 1–2) green incl. red-input cases (malformed
  JSON, dup-tuple legacy file, EXDEV simulation); crash-window simulation tests pass;
  registry `<uuid>.json` survival asserted by census.

### M4 — Consumer sweep (mechanical, evidence-grepped)

- `internal/web/events.go` watchMap path swap (SSE key frozen) + viewmodel ops glob +
  queue read seam assertions; regen templ if the comment lives in source.
- `internal/statusline/backlog.go`: join swap + SQLite-count read honoring C-2
  budgets, fail-open shape unchanged.
- `internal/cli/kanban.go` registry helpers: `companionRegistryPath` (:330,
  companions.json) and `leadRegistryPath` (:368, leads.json) — hand-joins into the
  renamed directory discovered in plan-audit; both swap with the central constant.
- `internal/cli/todo.go` prose surfaces: user-visible Long help text (:80 names the
  path) and the :48 comment literal update together with code sites.
- `internal/cli/graph.go`: touch only if resolver indirection needs the new constant
  (expected: none — seamed already).
- Hook family: confirm zero direct literals remain; record grep evidence.
- Old-binary-coexistence check: simulate previous-version reads (json-name-only
  accessor) against a migrated scratch root.
- Gate: dual-spelling literal-cleanliness grep (E4) at target counts; web/hook/cli
  affected suites green.

### M5 — Templates, export verb, documentation (Template-First cycle)

- Edit the three template skill docs FIRST, `make build`, verify embedded-bundle
  freshness; sync local copies per standard flow.
- Implement `moai todo export-json` (additive; golden-flag/output tests; included in
  the frozen verb-surface comparison of REQ-TOSQ-007).
- Docs: downgrade procedure + artifact-meaning table (design.md §7) in the docs tree;
  CHANGELOG entry drafted for sync phase.
- Gate: template-neutrality CI guard green on the path set; docs present; export
  round-trip property test green (seed→migrate→mutate→export→field-compare against
  live reader).

### M6 — Cross-platform, race, and gates hardening (release-blocking finishes last)

- Windows-flavored unit variants where locking/filename differences matter (mirror the
  `board_lock_clear_windows.go` split); `GOOS=windows go vet` compile evidence +
  CI windows job read-back as the behavioral verdict (C-3).
- Lock-held directory rename test (windows vs POSIX divergence): with a foreign
  process holding flock/LockFileEx on `backlog.lock` INSIDE the old directory, attempt
  the directory relocation — assert the degradation path (fallback-read serves the old
  layout, no data loss, no partial rename debris) and that a retry on next open once
  the lock is released completes; Windows LockFileEx sharing-violation behavior is the
  case this exists for.
- `go test -race ./internal/kanban/...` (and cli spawn-variant where applicable).
- Full AC matrix execution; progress.md §E.2/§E.3 evidence assembly; run-notes with
  E1–E6 results.
- Gate: acceptance.md DoD fully satisfied.

## §G Anti-Patterns rejected up front

- No storage-interface abstraction layer, no pluggable backends, no engine-selection
  config (design.md §7 justification stands).
- No global-fleet migrator daemon/script.
- No JSON/db dual-write transitional mode (divergence class it reintroduces is the
  SPEC's named enemy).
- No silent deletion/quarantine compaction of `.migrated` artifacts by routine code.
- No breaking verb/flag/output changes smuggled in under the storage umbrella.

## §H Cross-References

research.md (measurements + premise corrections) · design.md (D1–D8, risk register) ·
acceptance.md (AC matrix, DoD) · SPEC-KANBAN-TODO-CLI-001 (store origin contract, five
frozen item fields) · SPEC-WEB-TODO-QUEUE-001 (pure-vs-adopting resolver split the
migration leverages).
