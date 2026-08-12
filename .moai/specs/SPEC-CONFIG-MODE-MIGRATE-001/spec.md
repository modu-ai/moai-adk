---
id: SPEC-CONFIG-MODE-MIGRATE-001
title: "dry-run-first, approval-gated mode-widening migration for .moai/config"
version: "0.1.0"
status: draft
created: 2026-08-12
updated: 2026-08-12
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/config"
lifecycle: spec-anchored
era: V3R6
tier: S
tags: "config, file-mode, migration, dry-run, operator-approval, idempotent"
related_specs: [SPEC-CONFIG-TIER-PERSIST-001, SPEC-CONFIG-ATOMIC-WRITE-001]
depends_on: [SPEC-CONFIG-ATOMIC-WRITE-001]
---

# SPEC-CONFIG-MODE-MIGRATE-001

## HISTORY

| Version | Date | Change |
|---------|------|--------|
| 0.1.0 | 2026-08-12 | Initial draft. Deferred slice extracted from parent `SPEC-CONFIG-TIER-PERSIST-001` §D.4 (REQ-CTP-025/026). The atomic-write sibling `SPEC-CONFIG-ATOMIC-WRITE-001` (CLOSED) shipped Stat-based mode *preservation* — which perpetuates the narrowed mode of any file already narrowed. This SPEC owns the one-time *widening* migration that repairs the already-narrowed population. Design decision locked by the user before authoring: candidate C — dry-run-first + operator-approval gate (`--apply` flag); the migration is never blind-widen or announce-then-auto-widen. Real-world evidence (this checkout, 2026-08-12): 30 of 31 `.moai/config/sections/*.yaml` files are mode 0644; exactly one (`llm.yaml`) is mode 0600 (working-tree-only narrowing; git index records 100644; `llm.yaml` carries no credentials). |

## §A Problem / Motivation

`SPEC-CONFIG-ATOMIC-WRITE-001` (CLOSED) introduced a shared atomic-write helper that
*preserves* the destination file's existing mode across every future write into
`.moai/config/**`. That fix is correct for every *future* write — but it has a blind
spot: **a `Stat`-based preservation fix alone perpetuates every file already narrowed.**
Any `.moai/config/sections/*.yaml` file that was already narrowed to 0600 (by the
pre-fix `os.CreateTemp` → `os.Rename` path in `internal/config/manager.go`'s former
`atomicWrite`) stays narrowed forever, because every subsequent save now *faithfully
preserves* the narrowed mode.

A separate one-time migration step is needed to **widen** the already-narrowed files
back toward the canonical mode `defs.FilePerm` (`0o644`, `internal/defs/perms.go:11`).

### Real-world evidence (this checkout, 2026-08-12)

Scan of `.moai/config/sections/*.yaml` (31 files) at HEAD `7951b891e`:

| Population | Mode | Count |
|------------|------|-------|
| canonical (already `defs.FilePerm`) | `rw-r--r--` (0644) | 30 |
| narrowed below `defs.FilePerm` | `rw-------` (0600) | **1** |

The single narrowed file is `.moai/config/sections/llm.yaml`. Git records it as
`100644` in the index; only the working-tree file is 0600 (a working-tree-only
narrowing, not a committed narrowing). `llm.yaml` carries **no credentials** — its
keys are `mode`, `team_mode`, `glm_env_var` (an env-var *name*, not a value),
`performance_tier`, and model/effort profiles. So the 0600 is very likely an
accidental narrowing, NOT deliberate credential protection.

**But intent is not mechanically decidable.** A future operator MAY legitimately
restrict a config file beyond `defs.FilePerm` (e.g. a site policy that mandates 0600
on all of `.moai/config/`). The migration therefore cannot decide "accidental vs
deliberate" on its own. The user-locked design (candidate C — dry-run-first +
operator approval) delegates that judgment to the operator: the migration lists every
candidate, and the operator decides which to widen via the explicit `--apply` flag.

## §B Goals

- Provide a one-time mode-widening migration that identifies every `.moai/config/**`
  file whose permission bits are narrower than `defs.FilePerm` and widens them back
  toward `defs.FilePerm`.
- Make the migration **dry-run-first**: the default invocation lists candidates
  (path + current mode → target mode) without modifying anything.
- Gate widening behind an **explicit operator-approval flag** (`--apply`), so the
  deliberately-restricted-vs-accidentally-narrowed judgment rests with the operator,
  not the tool.
- Make the migration **idempotent** and safe to re-run on a tree where it has already
  been applied.
- Route the widening writes through the shared atomic-write helper
  (`atomicfile.Write`, shipped by `SPEC-CONFIG-ATOMIC-WRITE-001`) so the migration's
  own writes are themselves atomic and mode-correct.

## §C Scope Boundary — the deferred slice of the parent

`SPEC-CONFIG-TIER-PERSIST-001` was an over-large 35-REQ Tier-M-exceeding SPEC, split
three ways per the 3-way split recommendation. This SPEC owns the **deferred
mode-widening migration slice**: parent `REQ-CTP-025` (widen narrowed files) and
`REQ-CTP-026` (only-widen / scoped / preserve-deliberately-restricted) decompose into
this SPEC's `REQ-MIG-001` / `REQ-MIG-002`. The atomic-write write-path invariant
(parent `REQ-CTP-021..024/027`) already shipped in `SPEC-CONFIG-ATOMIC-WRITE-001`.

| Slice | Topic | Owning SPEC |
|-------|-------|-------------|
| (a) | config tier resolution, explicit-falsey-wins-tier semantics, local-tier reachability | resident in `SPEC-CONFIG-TIER-PERSIST-001` |
| (b) | atomic, mode-preserving writes for `.moai/config` + CLI persistence | `SPEC-CONFIG-ATOMIC-WRITE-001` (CLOSED) |
| **(d)** | **mode-widening migration (dry-run-first + operator approval)** | **`SPEC-CONFIG-MODE-MIGRATE-001` (THIS SPEC)** |

### Out of Scope — atomic write path

- The shared atomic-write helper (`atomicfile.Write`), its temp-file + rename +
  mode-preservation semantics, and its call-site remediation are owned by
  `SPEC-CONFIG-ATOMIC-WRITE-001` (CLOSED). This SPEC consumes the helper as a caller;
  it does not re-derive it or modify it.
- Any guard test that asserts "no bare `os.WriteFile` to `.moai/config/**`" is owned
  by `SPEC-CONFIG-ATOMIC-WRITE-001` `REQ-CAW-007`. The migration's widening writes go
  through the helper, so they satisfy that guard rather than being a new exception to
  it.

### Out of Scope — tier resolution and malformed-section writeback

- Config tier resolution, explicit-falsey-wins-tier semantics, the branch-guard
  local-tier opt-in, and the malformed-section writeback-as-defaults contract are
  slice (a) / slice (c) concerns that remain in the parent.

### Out of Scope — non-config paths

- The migration targets `.moai/config/**` only. Widening files under
  `internal/`, `cmd/`, `.claude/`, or any other path is out of scope. Settings
  artifacts rendered to `.claude/settings.json` are also out of scope (they are
  addressed at write-time by the atomic-write helper, not by this migration).

## §D Requirements (GEARS)

### D.1 Dry-run-first widening (REQ-CTP-025 → REQ-MIG-001)

- **REQ-MIG-001** — **When** the operator invokes the mode-widening migration without
  the explicit apply flag, the migration step shall enumerate every file under
  `.moai/config/**` that is a widening candidate per the §D.2 Predicate definition
  (i.e. whose current permission bits are a proper subset of `defs.FilePerm`'s bits),
  and for each candidate shall report the path, the current mode, and the target mode
  (`defs.FilePerm`), WITHOUT modifying any file; and **When** the operator invokes the
  migration with the explicit apply flag, the migration step shall widen each
  enumerated candidate's permission bits toward `defs.FilePerm` via the shared
  atomic-write helper shipped by `SPEC-CONFIG-ATOMIC-WRITE-001`.

### D.2 Only-widen, scoped, preserve-deliberately-restricted (REQ-CTP-026 → REQ-MIG-002)

- **REQ-MIG-002** — The mode-widening migration shall widen only: it shall never narrow
  a file's permission bits below the file's current mode (a file whose bits are not a
  proper subset of `defs.FilePerm` is excluded by the §D.2 Predicate definition, so
  widening toward `defs.FilePerm` can only ADD bits, never remove any), shall never
  alter any file outside the `.moai/config/` directory tree, and shall delegate the
  deliberately-restricted-vs-accidentally-narrowed judgment to the operator via the
  dry-run preview plus explicit apply-flag approval gate, so that a file the operator
  chooses not to widen is left unchanged.

#### Predicate definition

A file under `.moai/config/**` is a **widening candidate** if and only if BOTH of the
following hold:

1. `(currentMode.Perm() | defs.FilePerm.Perm()) == defs.FilePerm.Perm()` — the file's
   permission bits are a **subset** of `defs.FilePerm`'s bits (`0o644`); AND
2. `currentMode.Perm() != defs.FilePerm.Perm()` — the file is not already at the
   canonical mode (an already-canonical file would be a no-op).

Equivalently: the file's permission bits are a **proper subset** of `defs.FilePerm`'s
bits. Widening such a file to `defs.FilePerm` only ADDS bits; it never removes any. This
is the precise "only-widen" predicate — load-bearing for REQ-MIG-002's "shall never
narrow" guarantee, because a file whose bits are NOT a subset of `defs.FilePerm` (e.g.
one carrying an exec bit or a group-write bit) would be *narrowed* by a set to
`defs.FilePerm`, and is therefore excluded from the candidate set.

Explicit enumeration (dissolves all ambiguity for an implementer reading only spec.md):

| Current mode | Candidate? | Reason |
|--------------|------------|--------|
| `0600` | **YES** | proper subset of 0644; widen to 0644 adds group-read + other-read |
| `0640` | **YES** | proper subset of 0644; widen to 0644 adds other-read |
| `0700` | **NO** | owner-exec bit `0100` is NOT in 0644 (0644 = 0600\|0040\|0004); setting to 0644 would drop owner-exec — a narrow, forbidden by REQ-MIG-002 |
| `0660` | **NO** | group-write bit `0020` is NOT in 0644; setting to 0644 would drop group-write — a narrow, forbidden by REQ-MIG-002 |
| `0644` | NO | already canonical (`==` `defs.FilePerm`) — clause 2 excludes |
| `0664` / `0666` | NO | bits are a superset of 0644, not a subset (clause 1 fails); never touched |

General rule (the predicate is normative; the table is illustrative): any mode carrying
a bit not present in `defs.FilePerm` — exec bits (`0100`/`0010`/`0001`), group-write
(`0020`), set-uid/gid, sticky — is excluded by clause 1, regardless of whether the
human-readable form "looks narrower". The mechanical subset test in clause 1 reaches
the correct answer for every mode, including modes not enumerated above (e.g. `0500`,
which is NOT a candidate because its owner-exec bit `0100` is absent from 0644).

## §E Non-Functional Constraints

- **NFR-MIG-001 (idempotent — carries parent NFR-CTP-006)** — The mode-widening
  migration shall be idempotent: running it twice (dry-run or apply) on the same tree
  produces the same candidate set and the same resulting file modes, and a tree where
  the migration has already been applied produces an empty candidate list.
- **Cross-platform**: the migration's `os.Chmod` / `atomicfile.Write` path must behave
  correctly on darwin/linux and must not break the `GOOS=windows GOARCH=amd64 go build
  ./...` build.
- **No hardcoded modes** (CLAUDE.local.md §14): every mode reference flows through
  `defs.FilePerm` or a named `defs` constant; the migration never inlines `0o644` or
  `0o600`.
- **Error wrapping** (CLAUDE.local.md §3): `fmt.Errorf("...: %w", err)`.
- **Test isolation** (CLAUDE.local.md §6 [HARD]): every AC test uses `t.TempDir()`;
  no test writes to the project root or touches the real `.moai/config/sections/`.
- **English code/comments/godoc** (CLAUDE.local.md §3, `code_comments: en`).
- **Subagent boundary**: this is a CLI tool-layer migration; no `AskUserQuestion` calls
  introduced into `internal/config/` or `internal/cli/` (C-HRA-008 family). Operator
  approval is captured via the `--apply` CLI flag, not via an in-process prompt.
- **Operator announcement**: the migration's dry-run output and its apply-mode output
  both announce what they did (candidate list / widened list), satisfying parent R7's
  "migration is announced in its output" mitigation.

## §F Acceptance Criteria (Tier S inline)

Given-When-Then scenarios (binary-testable). Each AC maps to one or more REQs.

- **AC-MIG-001** (dry-run default — REQ-MIG-001) — **Given** a `.moai/config/` tree
  containing at least one file whose mode is narrower than `defs.FilePerm` (e.g. a
  file at 0600), **When** the operator runs the migration with no flags (or with the
  dry-run flag), **Then** the migration lists every such candidate with its path,
  current mode, and target mode (`defs.FilePerm`), AND the file modes on disk are
  unchanged after the run (verifiable by re-`os.Stat`-ing each candidate and asserting
  the mode equals the pre-run mode).

- **AC-MIG-002** (apply widens — REQ-MIG-001) — **Given** a `.moai/config/` tree
  containing a file at mode 0600, **When** the operator runs the migration with the
  `--apply` flag, **Then** that file's mode on disk becomes `defs.FilePerm` (`0o644`),
  verifiable by `os.Stat` after the run.

- **AC-MIG-003** (only-widen, never narrow — REQ-MIG-002) — **Given** a `.moai/config/`
  tree containing a file at mode 0644 (already canonical) and a file at mode 0600,
  **When** the operator runs the migration with `--apply`, **Then** the 0644 file's
  mode is unchanged (not widened beyond `defs.FilePerm`, not narrowed) AND the 0600
  file's mode becomes `defs.FilePerm` (`0o644`).

- **AC-MIG-004** (scope: `.moai/config/` only — REQ-MIG-002) — **Given** a tree in
  which a file OUTSIDE `.moai/config/` (e.g. `/tmp/.../outside.yaml` or a
  `.claude/settings.json`) is mode 0600, **When** the operator runs the migration with
  `--apply`, **Then** that outside file's mode is unchanged (the migration never
  touches anything outside `.moai/config/`).

- **AC-MIG-005** (idempotent — NFR-MIG-001) — **Given** a `.moai/config/` tree where
  the migration has already been applied with `--apply` (all files now at
  `defs.FilePerm`), **When** the operator re-runs the migration in dry-run mode,
  **Then** the candidate list is empty, AND re-running with `--apply` is a no-op (no
  file mode changes, exit 0).

- **AC-MIG-006** (routes through atomic-write helper — REQ-MIG-001) — **Given** the
  migration's apply path is invoked, **When** the widening write for each candidate
  lands, **Then** the write goes through the shared `atomicfile.Write` helper (shipped
  by `SPEC-CONFIG-ATOMIC-WRITE-001`), verifiable by a grep/test asserting the
  migration's apply function calls `atomicfile.Write` and does not call a bare
  `os.WriteFile` or `os.Chmod` directly on the destination path.

- **AC-MIG-007** (only-widen on a non-subset mode — REQ-MIG-002, §D.2 Predicate
  definition) — **Given** a `.moai/config/` tree containing a file at mode `0700`
  (owner-exec bit `0100` is NOT in `defs.FilePerm` `0o644`, so this file is NOT a
  candidate per the §D.2 Predicate definition — widening to 0644 would drop owner-exec,
  a narrow), **When** the operator runs the migration with `--apply`, **Then** (a) the
  `0700` file's mode on disk is UNCHANGED after the run (verifiable by `os.Stat`
  before and after, asserting the modes are equal), AND (b) the dry-run output (re-run
  without `--apply` on the same tree) excludes the `0700` file from the candidate list
  (or flags it as "would-require-narrowing — skipped"). This AC makes REQ-MIG-002's
  only-widen guarantee mechanically testable on the axis where it is actually
  load-bearing: a mode whose widening-to-`defs.FilePerm` would also narrow.

### Dry-run preview example (illustrative — based on this checkout's real evidence)

```
$ moai config mode-migrate    # dry-run default
Scanning .moai/config/** for files narrower than defs.FilePerm (0o644)...

  PATH                              CURRENT    TARGET
  .moai/config/sections/llm.yaml    0600       0644

1 candidate(s) found. Review the list and re-run with --apply to widen.
No files were modified.
```

On this checkout, exactly the one file above would appear (30 of 31 sections files
are already canonical; `llm.yaml` is the sole narrowed file at 0600).

## §G Risk Register

| Risk | Severity | Mitigation |
|------|----------|------------|
| **R1** (carries parent R7) The migration widens a file an operator deliberately restricted beyond `defs.FilePerm`. | Low / Medium | `REQ-MIG-002` scopes the migration to `.moai/config/` and to widening-toward-`defs.FilePerm` only; the dry-run preview (`REQ-MIG-001`) gives the operator a pre-apply review/skip opportunity, and the `--apply` flag is the explicit approval gate. The migration's output announces what it widened. |
| **R2** The migration silently re-narrows files on a second run (non-idempotent). | Medium | `NFR-MIG-001` + `AC-MIG-005` assert idempotency mechanically: a tree already at `defs.FilePerm` yields an empty candidate list, and a second `--apply` is a no-op. |
| **R3** The migration's widening writes reintroduce the truncate-then-write window or the mode-narrowing path that `SPEC-CONFIG-ATOMIC-WRITE-001` fixed. | Low | `REQ-MIG-001` + `AC-MIG-006` require the migration's apply path to route through `atomicfile.Write`, inheriting the atomic + mode-preserving guarantees of the sibling SPEC rather than re-implementing a raw write. |
| **R4** Operator confusion between dry-run and apply modes leads to accidental widening. | Low | The default invocation is dry-run-only (a no-op on disk); `--apply` is opt-in and its output names every widened file. The command's `--help` states the default-no-op behavior. |

## §H Cross-References

- `.moai/specs/SPEC-CONFIG-TIER-PERSIST-001/` — **parent SPEC**; §D.4 (REQ-CTP-025 /
  REQ-CTP-026) is the source of this slice. The parent's `## §L Split Branches`
  section records the extraction. This SPEC is the `(d)` slice (deferred
  mode-widening migration).
- `.moai/specs/SPEC-CONFIG-ATOMIC-WRITE-001/` — **sibling SPEC** (CLOSED); shipped the
  shared `atomicfile.Write` helper (atomic + mode-preserving). This SPEC's migration
  consumes that helper as a caller (`depends_on`). The atomic-write SPEC's §C
  "Out of Scope — mode-widening migration" explicitly deferred REQ-CTP-025/026 here.
- `internal/defs/perms.go:11` — `defs.FilePerm` (`0o644`), the canonical target mode.
- `CLAUDE.local.md` §6 (Test Isolation — `t.TempDir()`), §14 (Hardcoding prevention —
  named `defs` constants, no inline mode literals).

## §L Lineage — deferred slice from SPEC-CONFIG-TIER-PERSIST-001

This SPEC is the **deferred (d) slice** of the parent `SPEC-CONFIG-TIER-PERSIST-001`.
The parent's §D.4 "Atomic, mode-preserving writes" section defined REQ-CTP-025 (widen
narrowed files) and REQ-CTP-026 (only-widen / scoped / preserve-deliberately-restricted).
Both were DEFERRED out of the atomic-write child SPEC (`SPEC-CONFIG-ATOMIC-WRITE-001`,
CLOSED) because the atomic-write fix only does `Stat`-based mode *preservation* —
which perpetuates the narrowed mode of any file already narrowed. A separate one-time
*widening* migration (this SPEC) is needed to repair the already-narrowed population.

REQ mapping (parent → child):

| Parent REQ | Child REQ | Topic |
|------------|-----------|-------|
| `REQ-CTP-025` | `REQ-MIG-001` | dry-run-first widening toward `defs.FilePerm` |
| `REQ-CTP-026` | `REQ-MIG-002` | only-widen, scoped to `.moai/config/`, operator-approval gate |
| `NFR-CTP-006` | `NFR-MIG-001` | idempotent migration, safe to re-run |
| `R7` (parent risk) | `R1` (this SPEC) | widening a deliberately-restricted file → mitigated by dry-run + `--apply` gate |

Bi-directional link: the parent's `## §L Split Branches` section records this SPEC as
the `(d)` slice destination for REQ-CTP-025/026; the atomic-write sibling's §C
"Out of Scope — mode-widening migration" cross-references this SPEC by ID. This
SPEC's `related_specs` field carries both links in the reverse direction.
