# plan.md — SPEC-CONFIG-MODE-MIGRATE-001

> Tier S implementation plan. Migration implementation is RUN-PHASE — this plan
> describes HOW the migration will be implemented; it does not write code.
> Parent: `SPEC-CONFIG-TIER-PERSIST-001` §D.4 (deferred slice). Sibling:
> `SPEC-CONFIG-ATOMIC-WRITE-001` (CLOSED, ships the `atomicfile.Write` helper this
> migration consumes).

## §A Context

- **Work location**: `/Users/goos/MoAI/moai-adk-go` (main checkout, NO worktree).
- **Branch**: `plan/spec-config-mode-migrate` (HEAD `7951b891e` = origin/main at
  authoring time).
- **SPEC artifacts**: `.moai/specs/SPEC-CONFIG-MODE-MIGRATE-001/{spec,plan,progress}.md`.
- **Tier**: S (2 counted artifacts: spec.md + plan.md; progress.md emitted at every
  Tier per the §E skeleton rule, not counted in the Tier total).
- **REQ/AC budget**: 2 REQs + 1 NFR + 6 ACs (Tier S ceiling 8 each — well within).
- **User-locked design decision**: candidate C — **dry-run-first + operator approval
  (`--apply` flag)**. NOT blind-widen, NOT announce-then-auto-widen. The operator is
  the only entity that can decide "deliberately restricted vs accidentally narrowed"
  — intent is not mechanically decidable.
- **Real-world evidence anchoring motivation**: on this checkout (2026-08-12),
  30 of 31 `.moai/config/sections/*.yaml` are mode 0644 (canonical); exactly one
  (`llm.yaml`) is mode 0600. `llm.yaml` carries no credentials → the 0600 is very
  likely accidental, but the SPEC still treats "deliberately restricted" as possible
  (R1/R7).

### A.1 Scope summary

The migration is a small CLI subcommand that:

1. Scans `.moai/config/**` for files whose permission bits are narrower than
   `defs.FilePerm` (`0o644`).
2. **Dry-run (default)**: prints the candidate list (path + current mode → target
   mode) and modifies nothing.
3. **Apply (`--apply` flag)**: widens each candidate toward `defs.FilePerm` via
   `atomicfile.Write` (the sibling helper), re-reading + re-writing the file content
   unchanged so the atomic helper's mode-preserving + atomic-rename semantics apply.
4. Is idempotent: a tree already at `defs.FilePerm` yields an empty candidate list.

### A.2 Files affected (estimated, run-phase will confirm)

| File | Purpose |
|------|---------|
| `internal/cli/config/mode_migrate.go` (NEW) | migration subcommand: scan + dry-run + apply |
| `internal/cli/config/mode_migrate_test.go` (NEW) | AC tests (dry-run no-op, apply widens, only-widen, scope, idempotent, helper-routing) |
| `internal/cli/config/cmd.go` (or nearest parent cmd wiring) | register the new subcommand under `moai config` |

Expected diff size: well under 300 LOC (Tier S threshold). The migration is a
single-purpose tool with one code path and a small test set.

### A.3 PRESERVE targets (do NOT touch)

- `internal/config/atomicfile/*.go` — the sibling helper shipped by
  `SPEC-CONFIG-ATOMIC-WRITE-001`. Consumed as a caller only; never modified.
- `.moai/config/sections/*.yaml` — the real config files. Tests use `t.TempDir()`;
  no test touches the real section files (CLAUDE.local.md §6 [HARD]).
- `.claude/settings.json`, `.moai/config/sections/llm.yaml`,
- `.claude/settings.json.doctor-bak` — pre-existing working-tree noise unrelated to
  this SPEC; never staged into this SPEC's commits.

## §B Known Issues (filtered to relevant categories)

- **B4 Frontmatter canonical schema**: `created:` / `updated:` / `tags:` (not
  snake_case aliases); `phase:` is the release target `"v3.2.0 target"` — NOT a
  lifecycle stage name (`plan` / `run` / `sync` / `mx` are prohibited).
- **B5 CI 3-tier awareness**: spec-lint, golangci-lint, and per-OS Test can each
  fail separately. A pre-existing baseline issue is NOT this SPEC's defect.
- **B6 spec-lint Out-of-Scope convention**: every `### Out of Scope — <topic>` H3
  sub-heading needs at least one `-` bullet (the `OutOfScopeRule` lint). spec.md
  carries three such sub-headings.
- **B8 Working-tree hygiene**: stage ONLY SPEC files via explicit pathspec
  (`git add .moai/specs/SPEC-CONFIG-MODE-MIGRATE-001/{spec,plan,progress}.md`).
  Do NOT `git add -A` — there is unrelated working-tree noise.
- **B10 Untouched paths**: do not modify `internal/config/atomicfile/`,
  `internal/config/manager.go`, or any atomic-write call site remediated by the
  sibling SPEC. This SPEC is additive only.
- **B11 AskUserQuestion prohibited in subagent code**: operator approval is captured
  via the `--apply` CLI flag, NOT via an in-process `AskUserQuestion` call.

## §C Pre-flight (run before M1)

```bash
git branch --show-current                              # expect plan/spec-config-mode-migrate
git rev-parse HEAD                                     # expect 7951b891e (or descendant)
go build ./...                                         # green baseline
GOOS=windows GOARCH=amd64 go build ./...               # cross-platform baseline
golangci-lint run --timeout=2m 2>&1 | tail -5          # record lint baseline
# Confirm the helper the migration will consume is live (shipped by the sibling SPEC):
grep -rn "atomicfile.Write" internal/config/ internal/cli/ | head -20
# Confirm defs.FilePerm is where we think it is:
grep -n "FilePerm" internal/defs/perms.go
```

## §D Constraints

- **DO NOT** modify `internal/config/atomicfile/*.go` or any atomic-write call site —
  owned by `SPEC-CONFIG-ATOMIC-WRITE-001` (CLOSED). This SPEC consumes the helper.
- **DO NOT** inline a numeric mode literal (`0o644`, `0o600`, `0o750`) anywhere in the
  migration code; route every mode reference through `defs.FilePerm` or a named `defs`
  constant (CLAUDE.local.md §14).
- **DO NOT** widen any file outside `.moai/config/` (REQ-MIG-002 scope).
- **DO NOT** narrow any file below its current mode (REQ-MIG-002 only-widen).
- **DO NOT** make the apply path the default — dry-run is the default; `--apply` is
  opt-in (REQ-MIG-001).
- **DO NOT** use `--no-verify`, `--amend`, or force-push. Conventional Commits only.
- **DO NOT** stage `.claude/settings.json`, `.moai/config/sections/llm.yaml`, or
  `.claude/settings.json.doctor-bak` into this SPEC's commits.
- **DO** wrap errors as `fmt.Errorf("mode-migrate: ...: %w", err)`.
- **DO** use `t.TempDir()` for every test (CLAUDE.local.md §6 [HARD]).
- **DO** run `moai spec lint SPEC-CONFIG-MODE-MIGRATE-001` and confirm 0 errors
  before the plan-phase commit.

## §E Self-Verification (Tier S minimal form — to be filled at run-phase)

Tier S permits the minimal self-verification form: Goal + Deliverables + Constraints +
AC PASS/FAIL matrix. The run-phase manager-develop fills the table below with verbatim
command + observed output + baseline HEAD SHA per the attribution discipline.

| AC | Status | Verification command | Observed output |
|----|--------|---------------------|-----------------|
| AC-MIG-001 (dry-run no-op) | _pending_ | `go test -run TestModeMigrateDryRun ./internal/cli/config/...` | _pending_ |
| AC-MIG-002 (apply widens) | _pending_ | `go test -run TestModeMigrateApply ./internal/cli/config/...` | _pending_ |
| AC-MIG-003 (only-widen) | _pending_ | `go test -run TestModeMigrateOnlyWidens ./internal/cli/config/...` | _pending_ |
| AC-MIG-004 (scope .moai/config/) | _pending_ | `go test -run TestModeMigrateScope ./internal/cli/config/...` | _pending_ |
| AC-MIG-005 (idempotent) | _pending_ | `go test -run TestModeMigrateIdempotent ./internal/cli/config/...` | _pending_ |
| AC-MIG-006 (helper routing) | _pending_ | `grep -n "atomicfile.Write" internal/cli/config/mode_migrate.go` | _pending_ |

Cross-platform build: `go build ./...` AND `GOOS=windows GOARCH=amd64 go build ./...`
must both exit 0.

## §F Milestones (ordered by decision-reversability)

Ordered so the most reversal-prone decisions land first and get human review, per the
plan-ordering rule.

### M1 — Dry-run scan + candidate enumeration (reversal-prone API surface)

The scan API (function signature, candidate-record shape, CLI verb + flag names) is
the most reversal-prone decision — getting it wrong forces the apply path to change
twice. M1 locks:

- The CLI verb (`moai config mode-migrate` — or the established naming convention in
  `internal/cli/config/`; confirm against siblings before locking).
- The candidate record struct (`path`, `currentMode`, `targetMode`).
- The scan predicate: "permission bits narrower than `defs.FilePerm`" — i.e. a file
  whose mode is not a superset of `defs.FilePerm`'s bits (a file at 0600 is narrower;
  a file at 0644 is not; a file at 0640 is narrower because it drops the group-read
  that `defs.FilePerm` carries — confirm the exact predicate at M1: "narrower than"
  is defined as `mode.Perm() != defs.FilePerm.Perm()` AND the file would be widened,
  which means any mode whose bits are not identical to `defs.FilePerm`. A file at 0600
  → widen to 0644; a file at 0644 → not a candidate; a file at 0700 → widen to 0644
  only if "widen toward" means clamp-to-`defs.FilePerm`, not "add bits". Lock this
  semantics at M1: the migration sets the candidate's mode to `defs.FilePerm`
  verbatim, which is a SET, not an OR — "only-widen" is the guarantee that the SET
  target (`0o644`) is never narrower than the current mode. A file deliberately at
  0600 is being widened (0600 → 0644 adds group/other read); a file deliberately at
  0700 would be *narrowed* by a set to 0644 (drops execute) — so the predicate MUST
  exclude files whose current mode is not a subset of `defs.FilePerm`. The exact
  predicate: "candidate iff `currentMode.Perm()` is a proper subset of
  `defs.FilePerm.Perm()` OR equal — wait, equal is not a candidate. Candidate iff
  widening toward `defs.FilePerm` would ADD bits without removing any, i.e.
  `(currentMode.Perm() | defs.FilePerm.Perm()) == defs.FilePerm.Perm()` AND
  `currentMode.Perm() != defs.FilePerm.Perm()`. This is the precise "only-widen"
  predicate. Lock at M1.)

  **[NEEDS CLARIFICATION: exact candidate predicate for files at modes that are NOT a
  subset of `defs.FilePerm` (e.g. 0700, 0660)]** — the "only-widen" predicate above
  (candidate iff current bits are a subset of `defs.FilePerm` bits AND not equal)
  excludes a file at 0700 from the candidate set (it would be narrowed by a set to
  0644). This is consistent with REQ-MIG-002's "shall never narrow a file below its
  current mode". Confirm with the operator at run-phase Kickoff whether this is the
  intended semantics, or whether a file at 0700 should be flagged in the dry-run
  output as "not a candidate (would require narrowing)" for operator awareness.

- The dry-run output format (path + current mode + target mode, human-readable, with
  the "N candidate(s) found. Re-run with --apply to widen." footer).

**Deliverable**: `internal/cli/config/mode_migrate.go` with the scan + dry-run path;
`mode_migrate_test.go` with AC-MIG-001 (dry-run no-op) + AC-MIG-004 (scope). No
apply path yet.

### M2 — Apply path via `atomicfile.Write` + idempotency guard (mechanical)

Once M1's API is locked, the apply path is mechanical:

- Add the `--apply` flag.
- For each candidate, re-read the file content, then write it back unchanged via
  `atomicfile.Write` — which sets the result file's mode to the destination's
  pre-existing mode by default (Stat-based preservation). To actually WIDEN, the
  migration MUST override the preserved mode: either (a) `os.Chmod(path,
  defs.FilePerm)` after the atomic write (simplest, and the write itself is already
  atomic), or (b) extend the atomic-write call to accept an explicit target-mode
  override. Option (a) is simpler and stays within the helper's public surface;
  option (b) modifies the sibling helper (forbidden — PRESERVE target). Lock option
  (a) at M2: the apply path is `atomicfile.Write(path, content)` followed by
  `os.Chmod(path, defs.FilePerm)`. The chmod-after-rename is itself atomic on
  POSIX (a single `chmod(2)` syscall), so the atomic-write guarantee is preserved.

  **Open question for M2**: is `os.Chmod` after `atomicfile.Write` acceptable given
  the sibling SPEC's `REQ-CAW-007` guard against bare `os.Chmod` to config paths?
  The guard targets `os.WriteFile` calls (truncate-then-write window) and
  hardcoded-mode literals; `os.Chmod(path, defs.FilePerm)` uses the named constant
  (not a literal) and does not truncate. Confirm at M2 that the guard's exemption
  list includes this one `os.Chmod` site, or restructure to avoid it.

- Idempotency: the scan is the source of truth — if the scan produces an empty
  candidate list (because every file is already at `defs.FilePerm`), apply is a
  no-op. No separate "already-applied" marker file is needed.

**Deliverable**: `--apply` path complete; AC-MIG-002 (apply widens), AC-MIG-003
(only-widen), AC-MIG-005 (idempotent), AC-MIG-006 (helper routing) green.

### M3 — Subcommand wiring + docs pointer (mechanical)

- Register `mode-migrate` under the `moai config` parent command (confirm the exact
  parent in `internal/cli/config/cmd.go` at run-phase; if `moai config` does not
  exist as a parent, fall back to the established naming convention for sibling
  config subcommands).
- `--help` text states: default is dry-run (no-op on disk); `--apply` widens.
- No CHANGELOG entry at plan-phase (sync-phase owns CHANGELOG).

**Deliverable**: subcommand reachable via `moai config mode-migrate --help`;
`go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` green; full
`go test ./...` green.

## §G Anti-Patterns (specific to this SPEC)

- **AP-MIG-001 — Blind-widen or announce-then-auto-widen.** The user-locked design is
  dry-run-first + explicit `--apply`. Any implementation that widens on the default
  invocation, or that auto-widens after printing the dry-run list without the flag,
  violates REQ-MIG-001 and the explicit design decision.
- **AP-MIG-002 — Widening via bare `os.WriteFile` or `os.Chmod` with a hardcoded
  literal.** REQ-MIG-001 + AC-MIG-006 require routing through `atomicfile.Write`;
  CLAUDE.local.md §14 forbids hardcoded mode literals. The `os.Chmod` site (if used
  in M2) MUST reference `defs.FilePerm`.
- **AP-MIG-003 — Modifying the sibling helper.** `internal/config/atomicfile/*.go` is
  owned by `SPEC-CONFIG-ATOMIC-WRITE-001` (CLOSED) and is a PRESERVE target. The
  migration consumes the helper; it does not extend or fork it.
- **AP-MIG-004 — Touching real config files in tests.** Every AC test uses
  `t.TempDir()`; no test reads or writes the real `.moai/config/sections/*.yaml`
  (CLAUDE.local.md §6 [HARD]).
- **AP-MIG-005 — Staging working-tree noise into the SPEC commit.** The commit stages
  ONLY `.moai/specs/SPEC-CONFIG-MODE-MIGRATE-001/{spec,plan,progress}.md` via explicit
  pathspec. `.claude/settings.json`, `.moai/config/sections/llm.yaml`, and
  `.claude/settings.json.doctor-bak` are unrelated working-tree noise.

## §H Cross-References

- `.moai/specs/SPEC-CONFIG-TIER-PERSIST-001/` — parent; §D.4 REQ-CTP-025/026 source;
  §L Split Branches records this SPEC as the `(d)` slice.
- `.moai/specs/SPEC-CONFIG-ATOMIC-WRITE-001/` — sibling (CLOSED); ships
  `atomicfile.Write`; `depends_on` target.
- `internal/defs/perms.go:11` — `defs.FilePerm` (`0o644`).
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter schema
  (12 canonical fields, `phase:` release-target rule, Tier S = 2 artifacts).
- `CLAUDE.local.md` §6 (Test Isolation), §14 (Hardcoding prevention).
