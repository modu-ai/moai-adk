# SPEC-AGENT-MEMORY-DRAIN-001 — Acceptance Criteria

> Two-cell adoption per `.claude/rules/moai/development/verification-completeness.md` §2.
> RED-now cells measured 2026-09-02 on card tree `c0c36c421` (worktree
> `.claude/worktrees/t223`, branch `WT-agent-memory-drain`) unless stated otherwise.
> **Covers:** REQ-AM-001 … REQ-AM-010.

## §D — AC Matrix

### AC-AM-001 — `moai memory drain` verb exists with preview/apply/json modes

**Covers:** REQ-AM-007, REQ-AM-008, REQ-AM-010. **Severity:** blocker.

- **AC-AM-001a (preview default)**
  - **Given** a repository with ≥1 registered worktree holding file-bearing agent-memory
    **When** `moai memory drain` runs without an apply flag **Then** the command prints a
    per-tree report of what would be copied, exits 0, and the primary store is
    byte-identical before and after (no write).
  - RED-now: `moai memory drain` (verbatim, this run):
    ```
    ERROR
    Unknown command "drain" for "moai memory".
    Try --help for usage.
    ```
    exit code `1` (captured without pipe: `true-exit=1`). RED reason: **the verb does not
    exist** — right reason, not vacuous (the surface itself is absent).
  - Green path: M1 — `--help` lists `drain`; preview run on a two-fixture-tree repo
    reports the copy set and writes nothing.
- **AC-AM-001b (json)**
  - **Given** the same setup **When** `moai memory drain --json` runs **Then** stdout is a
    single JSON array parseable by `jq`, each record carrying `path`, `agents`,
    `files`, `copied`, `collided`, `skipped`.
  - RED-now: same probe as AC-AM-001a (verb absent). Green path: M1 — `jq -e` parse of
    the fixture run asserted in a go test.

### AC-AM-002 — Backfill copies orphaned worktree topic files into the primary store

**Covers:** REQ-AM-001, REQ-AM-007. **Severity:** blocker.

- **Given** a fixture primary store and a fixture worktree holding
  `agent-memory/manager-spec/feedback_x.md` (+ its MEMORY.md) **When**
  `moai memory drain --yes` runs **Then** the file exists at the primary store's
  `agent-memory/manager-spec/feedback_x.md` byte-identical, and the worktree copy is
  unchanged.
- RED-now (population-level, this run): 40 topic files exist under worktree
  `.claude/agent-memory/` (`find … -name '*.md' ! -name 'MEMORY.md' | wc -l` → 40);
  overlap with the primary's 203 topic files (`comm -12` → **0**). RED reason: **no
  drain mechanism exists**, so nothing reaches primary — not wrong-reason red (no
  unrelated pre-existing files block it).
- Green path: M1 — fixture go test passes; plus the real-tree backfill run whose
  `--json` output is archived as run-phase evidence (post-run overlap > 0).

### AC-AM-003 — Collision never overwrites a primary file

**Covers:** REQ-AM-003. **Severity:** blocker.

- **Given** the primary store already holds `agent-memory/plan-auditor/feedback_dup.md`
  with content A, and the fixture worktree holds the same filename with content B ≠ A
  **When** the drain (or mirror) runs **Then** the primary file still contains exactly A
  (byte-compared), and a `feedback_dup.wt-<worktree-name>.md` containing B exists.
- RED-now: mechanism absent (AC-AM-001a probe); additionally, today nothing is ever
  mirrored (AC-AM-002 RED, overlap 0), so no collision policy can be exercised. RED
  reason: absent surface.
- Green path: M1 — fixture go test with the A/B fixture asserts both the no-overwrite
  invariant and the tree-qualified rename. **Mutant check:** a drain that skips collided
  files entirely (copies nothing) would fail the rename assertion — the criterion does
  not admit it.

### AC-AM-004 — Index reconciliation: every copied topic gains exactly one index line

**Covers:** REQ-AM-004. **Severity:** blocker.

- **Given** a fixture worktree whose agent MEMORY.md indexes `feedback_x.md`, and a
  primary agent MEMORY.md with no line for it **When** the drain copies the topic file
  **Then** the primary agent's MEMORY.md gains exactly one new line referencing
  `feedback_x.md`, and no other index line changes; the worktree's MEMORY.md file is
  never copied.
- RED-now: mechanism absent (AC-AM-001a probe). Green path: M1 — fixture go test counts
  index lines before/after (delta exactly 1, target file named).
- **Regression guard:** post-backfill, `moai memory doctor` on the primary store reports
  zero NEW dangling index lines attributable to the drain (run-phase evidence, M3).

### AC-AM-005 — Write-time mirror copies worktree agent-memory writes to primary

**Covers:** REQ-AM-001, REQ-AM-002. **Severity:** blocker.

- **Given** a fixture HookInput whose `tool_input.file_path` is a `.md` under a worktree
  `agent-memory/` path, and the file exists on disk **When** the PostToolUse memory path
  runs **Then** the file appears at the same agent-relative path in the resolved primary
  store.
- RED-now: no mirror code exists — `grep -n "mirror" internal/hook/post_tool.go` returns
  no match (the file's only memory machinery is `runMemoryAudit`, lines ~571-620,
  observation-only). RED reason: absent mechanism, measured on the source.
- Green path: M2 — unit test with fixture HookInput JSON; plus one live dogfood
  observation (a real worktree Write producing a primary copy) recorded as run-phase
  evidence.

### AC-AM-006 — Mirror fails open

**Covers:** REQ-AM-005. **Severity:** blocker.

- **Given** a fixture HookInput in a worktree whose primary root cannot be resolved
  (fixture: no git common dir) **When** the PostToolUse memory path runs **Then** the
  hook exits 0, emits a non-blocking notice on stderr, and returns no block decision.
- RED-now: design-only at plan time (the mechanism does not exist; nothing to fail open
  yet) — verified at run phase per verification-completeness §2 (green-path cell with a
  known-failing fixture input; the fixture IS the known-red input for the blocking
  behavior: a naive implementation that returns a block decision on copy failure fails
  this AC).
- Green path: M2 — go test asserts exit 0 + no `decision` field + stderr notice.

### AC-AM-007 — Drain never deletes or mutates worktree content

**Covers:** REQ-AM-009. **Severity:** blocker.

- **Given** a fixture worktree with agent-memory files **When** `moai memory drain --yes`
  completes **Then** every worktree file is byte-identical (checksummed before/after) and
  the worktree remains registered.
- RED-now: mechanism absent (AC-AM-001a probe). Green path: M1 — fixture go test with
  pre/post checksums of the whole fixture worktree tree.
- **Mutant check:** a drain that "moves" (deletes source) satisfies AC-AM-002 but fails
  this criterion — the pair closes the move-vs-copy mutant.

### AC-AM-008 — Primary-session mirror is a no-op

**Covers:** REQ-AM-006. **Severity:** major.

- **Given** a fixture HookInput whose project root IS the primary (common dir == own
  `.git`) **When** the memory path runs **Then** no copy occurs (no
  `.wt-` self-duplicates) and the hook exits 0.
- RED-now: mechanism absent (AC-AM-005 RED). Green path: M2 — unit test.

### AC-AM-009 — Native auto-memory store untouched

**Covers:** scope boundary (spec.md §B.2, §G). **Severity:** regression-guard.

- **Given** the mirror and drain run **When** inspecting the native store derivation
  **Then** no per-worktree `memory/` directory is created (worktrees keep sharing the
  primary's native store).
- RED-now (baseline, this run): `find <profile>/projects -maxdepth 2 -type d -name memory
  | grep -c worktree` → **0** — the baseline the mechanism must not change.
- Green path: run-phase re-measurement after M2 dogfood → still 0.

## §D.1 — Severity summary

| Severity | ACs |
|---|---|
| blocker | AC-AM-001a/001b, 002, 003, 004, 005, 006, 007 |
| major | AC-AM-008 |
| regression-guard | AC-AM-009 (+ AC-AM-004 doctor guard) |

## §D.2 — Traceability

| REQ | Covered by |
|---|---|
| REQ-AM-001 | AC-AM-002, AC-AM-005 |
| REQ-AM-002 | AC-AM-005 |
| REQ-AM-003 | AC-AM-003 |
| REQ-AM-004 | AC-AM-004 |
| REQ-AM-005 | AC-AM-006 |
| REQ-AM-006 | AC-AM-008 |
| REQ-AM-007 | AC-AM-001a, AC-AM-002 |
| REQ-AM-008 | AC-AM-001a |
| REQ-AM-009 | AC-AM-007 |
| REQ-AM-010 | AC-AM-001b |

## §D.3 — Edge cases carried into fixtures

- Topic file exists in worktree but its worktree MEMORY.md has no line for it (derive
  from frontmatter description) — M1 fixture.
- `_archive/` subdir content — M1 fixture (copy under collision rule).
- Skeleton-only tree (dirs, zero files) — M1 fixture (reported, nothing copied).
- Same lesson written in two different worktrees (both collide in primary) — M1 fixture
  (two `.wt-` copies, two index lines).
- Empty/absent frontmatter on a mirrored file — mirror copies bytes verbatim; taxonomy
  warnings apply to the source write as today (no new gating).

## §D.4 — Definition of Done

All blocker ACs green with cited evidence on the card tree; package-scoped tests green
(`go test ./internal/hook/... ./internal/cli/worktree/...` + any new helper package);
`golangci-lint` clean on touched packages; SPEC-WORKTREE-REAPER-001 files untouched
(`git diff --stat` vs base shows no reaper path); backfill real-run `--json` report
archived under `.moai/reports/t223/`.
