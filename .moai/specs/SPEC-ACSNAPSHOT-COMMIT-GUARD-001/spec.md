---
id: SPEC-ACSNAPSHOT-COMMIT-GUARD-001
title: "Commit-time guard — reject an in-place acceptance.md amendment whose AC count no longer matches the staged corpus snapshot"
version: "0.1.0"
status: draft
created: 2026-09-24
updated: 2026-09-24
author: manager-spec (card t1150)
priority: P1
phase: "v3.2.0"
module: "scripts/ac-baseline,internal/spec,.moai/docs"
lifecycle: spec-anchored
tier: M
depends_on:
  - SPEC-AC-BASELINE-REFRESH-001
  - SPEC-AC-COUNT-DISCRIMINATOR-001
tags: "ac-count,baseline,snapshot,pre-commit,git-config-hook,local-only,lifecycle-cascade"
---

# SPEC-ACSNAPSHOT-COMMIT-GUARD-001 — Commit-time AC-snapshot guard

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-24 | manager-spec (card t1150) | Initial plan-phase draft. Mechanism and scope fixed by recorded operator decisions (§A.4); four candidate mechanisms weighed in plan.md §2. |

> **Why the ID carries no `AC-` segment.** The AC counter (sentinel-delimited awk program in `.claude/agents/moai/manager-docs.md`) matches `AC-([A-Z0-9]+-)*[0-9]+[a-z]?` anywhere on a line with no left boundary, so any SPEC ID of the form `SPEC-AC-…-NNN` cited inside an `acceptance.md` is itself counted as a criterion. `ACSNAPSHOT` keeps this SPEC's own acceptance file free of that contamination (measured by reading the counter body, §A.3 F4).

## §A Context

### A.1 Problem — the documented cascade did not prevent the second red

`TestACCounterFullCorpusMatchesBaseline` (`internal/spec/ac_count_clause_test.go:444`) compares every depth-1 `.moai/specs/*/acceptance.md` against its record in the tracked snapshot `.moai/reports/t338/ac-count-baseline.txt`. A record whose live count moves is a hard failure. The fourth trigger row of `.moai/docs/ac-count-baseline-refresh.md` §2 — "an existing `acceptance.md` amended in place so that its AC count changes" — obliges the regenerated snapshot to land in the SAME commit (§3 same-commit rule).

Two develop CI reds in two days came from exactly that row:

- t1106 amendment → develop red → repaired by t1122.
- t1139 amendment → develop red → repaired by t1148.

The trigger row itself was added by t1130 between the two incidents; the second red happened anyway. Documentation is a policy-layer control with no enforcement point, and the class has now recurred after SPEC-AC-BASELINE-REFRESH-001 landed — which is precisely the condition that SPEC's own exclusion "mechanical cascade enforcement … if the class recurs after this SPEC lands, automation is a separate card" names. This SPEC is that card.

### A.2 Why lanes do not see the red locally

The gate lives in the `internal/spec` Go package. Lane discipline (CLAUDE.local.md §4 / §6) is to test only the affected Go packages; an `acceptance.md` edit touches no Go package, so no lane runs `internal/spec` for it. The full-corpus test costs roughly 10–13 s (lead measurement: `go test ./internal/spec -run TestACCounterFullCorpusMatchesBaseline -count=1` → `ok … 10.161s`), which is cheap, but nothing ties it to the moment the amendment is committed. The first observer of the red is develop CI, after the lead's batch push.

### A.3 Measured facts (tree 60017eb83, 2026-09-24, this session unless marked)

- **F1 — counter single source.** The counter is the awk program between `# MOAI-AC-COUNTER-BEGIN` / `# MOAI-AC-COUNTER-END` in `.claude/agents/moai/manager-docs.md`; it reads `$AC_FILE`, prints the live count on stdout and `live=N excluded=N ambiguous=0` on stderr, or exits 3 with an `AMBIGUOUS <ids>` first line (the HALT state). The Go test extracts it via `extractCounterCommand` (`ac_count_clause_test.go:83`), which requires exactly one sentinel pair and a non-empty body.
- **F2 — baseline format and comparison.** `parseACBaseline` (`:290`) reads `<path>  COUNT <n>  live=<n> excluded=<n> ambiguous=0` and `<path>  HALT <ids…>  owner=… reason=…`, skipping `#` lines. `acComparison` (`:372`) judges absence first (unrecorded file → report, never fail), then HALT-vs-COUNT state moves, HALT identifier-set moves, and COUNT moves on BOTH `live` and `excluded`.
- **F3 — the remedy wording already exists as constants.** `acRegenerateCommand` (`:58`) = `MOAI_AC_BASELINE_REGENERATE=1 go test ./internal/spec -run TestACCounterBaselineRegenerate -count=1`; `acRemedySuffix` (`:62`) appends the cascade doc path `.moai/docs/ac-count-baseline-refresh.md`.
- **F4 — counter has no left token boundary.** The counter pattern is `"(" prefixes ")-([A-Z0-9]+-)*[0-9]+[a-z]?"` matched repeatedly over each line (read from the counter body).
- **F5 — git version.** `git --version` → `git version 2.54.0 (Apple Git-157)`.
- **F6 — config-defined hooks.** `man git-hook` (2.54): hooks configured with `hook.<name>.event` / `hook.<name>.command` run "in the order they are discovered during the config parse. The default <hook-name> from the hookdir is run last." The shell-level ordering and the failure semantics across multiple hooks were NOT observed empirically in plan (the worktree guard refuses out-of-tree git); they are run-phase acceptance criteria.
- **F7 — the live repository disables the hookdir.** `git config --show-origin --get core.hooksPath` → `file:/Users/goos/MoAI/moai-adk-go/.git/config	/dev/null`. The shared repo config points the hookdir at `/dev/null`, so the template-distributed `.git/hooks/pre-commit` does not currently run in any worktree of this repository. Whether config-defined hooks still fire under `core.hooksPath=/dev/null` is unobserved and is a release-blocking run-phase criterion: if they do not, this mechanism is inert in this repository and the run phase must stop and report rather than alter `core.hooksPath`.
- **F8 — no hook config today.** `git config --get-regexp '^hook\.'` → empty output, exit 1.
- **F9 — agents cannot bypass.** `internal/hook/pre_tool.go:466` denies any Bash command containing both `git commit` and `--no-verify`. The standard bypass (`--no-verify`) is therefore available to a human at a terminal but not to a Claude session. A false positive from this guard stalls a lane until a human intervenes; a tool fault treated as a rejection would stall every lane at once (see §A.5).
- **F10 — managed hook provenance.** `.git/hooks/pre-commit` is template-distributed (`internal/template/templates/.git_hooks/pre-commit`), provenance-tracked by `.git/hooks/.moai-pre-commit.sha256`, rewritten by `moai hook install` (`internal/cli/hook_install_precommit.go`), and that installer explicitly ships no `pre-commit.local` facility (`:382`). (Lead measurement; file existence of the installer and template source re-observed this session.)
- **F11 — shared config.** `.git/config` is the common config of every linked worktree (no `extensions.worktreeConfig` observed), so one install arms the guard for every lane worktree — including worktrees whose branch predates the checker script.
- **F12 — pathspec globbing crosses slashes.** A git pathspec `.moai/specs/*/acceptance.md` matches `_archive/<X>/acceptance.md` as well, because `*` in a git pathspec is not segment-bounded by default. The corpus uses `filepath.Glob` depth-1 semantics and skips `_archive/`; the checker must reproduce that population, not the pathspec's (documented git behaviour; to be pinned by a fixture in run).

### A.4 Recorded operator decisions (binding — not reopened here)

1. **Mechanism:** a git config-defined `pre-commit` hook (git ≥ 2.54 `hook.<name>.event = pre-commit` + `hook.<name>.command`), reject-type. NOT a Claude Code PreToolUse hook, NOT a `moai spec lint` rule, NOT automatic regeneration.
2. **Scope:** in-place amendment only — index status `M` (against `HEAD`) of a depth-1 `.moai/specs/*/acceptance.md`, `_archive/` excluded. Deletion and `_archive/` moves (the vanish class) are out of scope.

### A.5 Fail-open decision for tool faults — and the doctrine tension

A **tool fault** is any condition in which the guard cannot produce a comparison for a qualifying staged file: counter sentinels absent or duplicated in the staged `manager-docs.md`, empty counter body, staged baseline absent or unreadable, the file's baseline line malformed, `awk`/`sh` unavailable, or the counter exiting with a code other than 0 or 3. (Exit 3 — HALT — is a measured state, never a fault.)

Decision: **fail open, loudly.** On a tool fault the guard exits 0 and writes a stderr line beginning `ac-baseline-guard: NOT CHECKED` that names the fault and the file(s) left unchecked.

Why open:
1. The CI gate stays authoritative and unchanged; a fault degrades the repository to exactly today's state, never below it.
2. The hook is installed in the shared `.git/config` (F11). A fail-closed fault — for example a sentinel edit in `manager-docs.md` — would block every commit that touches an amended `acceptance.md` in every lane simultaneously, and agents cannot bypass (F9).
3. A rejection must mean "this commit carries the defect", so it stays reserved for a measured mismatch.

The tension: `verification-claim-integrity.md` §1 — absence of a failure signal is not evidence of a pass. Fail-open silently would convert an unrun check into an apparent pass. The mitigation is that the guard never exits 0 silently on a qualifying file: it either prints the checked line (§B REQ-ABG-009) or the `NOT CHECKED` line. A silent exit 0 is reserved for the no-qualifying-file path, where there is nothing to check.

## §B Requirements (GEARS)

### Layer 1 — predicate

**REQ-ABG-001** — **When** a commit is attempted and the commit's index holds one or more qualifying files — a depth-1 `.moai/specs/<dir>/acceptance.md` (not under `_archive/`) whose index status against `HEAD` is modified, with rename detection disabled — the guard shall measure each qualifying file's STAGED content with the counter extracted from the STAGED `.claude/agents/moai/manager-docs.md`, and compare the result against that file's record in the STAGED `.moai/reports/t338/ac-count-baseline.txt`.

**REQ-ABG-002** — The guard shall apply the comparison semantics of the corpus test's `acComparison`: a COUNT record mismatches when either `live` or `excluded` differs; a HALT result against a COUNT record, a COUNT result against a HALT record, and a moved HALT identifier set are mismatches; a qualifying file with no record in the staged baseline is not a mismatch.

**REQ-ABG-003** — **When** any qualifying file mismatches, the guard shall exit non-zero, and its stderr shall name, per mismatching file, the path, the recorded value, the staged value, the regeneration command `MOAI_AC_BASELINE_REGENERATE=1 go test ./internal/spec -run TestACCounterBaselineRegenerate -count=1`, and the cascade procedure `.moai/docs/ac-count-baseline-refresh.md`.

**REQ-ABG-004** — **When** the regenerated baseline is staged in the same commit as the amendment and records the amended file's staged value, the guard shall pass that file — the same-commit rule is satisfied by construction because both sides of the comparison are read from the index.

**REQ-ABG-005** — **While** git prepares a commit with an alternate index (`git commit -a`, `git commit <pathspec>`), the guard shall read the index git designates for the commit (`GIT_INDEX_FILE` as provided), never the default index.

### Layer 2 — cost and silence

**REQ-ABG-006** — **When** a commit carries no qualifying file, the guard shall exit 0 without extracting the counter, without reading the baseline, and without writing any output.

**REQ-ABG-007** — The guard shall carry no copy of the counter program: its only counter source is the sentinel-delimited block of the staged `manager-docs.md`, and it shall treat anything other than exactly one sentinel pair with a non-empty body as a tool fault.

### Layer 3 — faults and visibility

**REQ-ABG-008** — **When** a tool fault (§A.5) occurs while at least one qualifying file is staged, the guard shall exit 0 and write a stderr line beginning `ac-baseline-guard: NOT CHECKED` that names the fault and the unchecked file(s).

**REQ-ABG-009** — **When** the guard has compared at least one qualifying file and found no mismatch, it shall write one stderr line beginning `ac-baseline-guard: checked` naming the number of files compared and, for each unrecorded file, its staged value as a report.

**REQ-ABG-010** — The guard shall not write, stage, or regenerate any tracked file, the baseline included; regeneration remains the human-reviewed command of `.moai/docs/ac-count-baseline-refresh.md` §1 and §6.

### Layer 4 — installation and coexistence

**REQ-ABG-011** — The repository shall carry an idempotent install step that writes `hook.ac-baseline-guard.event = pre-commit` and `hook.ac-baseline-guard.command` into the repository git config; a repeated run shall leave exactly one `event` value and one `command` value.

**REQ-ABG-012** — **Where** the installed git is older than 2.54, the install step shall exit non-zero with a message naming the found and required versions, and shall write nothing.

**REQ-ABG-013** — **When** the configured command runs in a worktree whose tree does not contain the checker script, the hook shall exit 0 with a stderr line beginning `ac-baseline-guard: NOT CHECKED` naming the absent script.

**REQ-ABG-014** — The guard and its install step shall not write `.git/hooks/*`, `.git/hooks/.moai-pre-commit.sha256`, or `core.hooksPath`.

**REQ-ABG-015** — **While** `core.hooksPath` points to `/dev/null` or to a directory holding a `pre-commit` file, the configured guard shall still run on `git commit`, and a guard rejection shall prevent the commit from being created.

### Layer 5 — placement and documentation

**REQ-ABG-016** — The checker and install step shall live under a dev-only path with no mirror under `internal/template/templates/`, and no CI workflow shall be changed.

**REQ-ABG-017** — `.moai/docs/ac-count-baseline-refresh.md` shall describe the mechanical gate — what it checks, the install command, the `NOT CHECKED` semantics, and the `--no-verify` bypass and its unavailability to agent sessions.

## §C Exclusions

### Out of Scope — the vanish class
- Deletion of an `acceptance.md` and moves into `_archive/` (index status `D`, or a rename split into `D` + `A`). Operator decision §A.4-2. The CI gate still catches them.

### Out of Scope — corpus-rewrite trigger
- Changes to the counter grammar or to the corpus glob (trigger row 3 of the cascade doc). A counter edit alters every file's measurement at once; the guard reads the staged counter, but it compares only qualifying files, not the corpus.

### Out of Scope — merge commits without conflicts
- `git merge` that completes without a conflict runs `pre-merge-commit`, not `pre-commit`; the integration-window merges are therefore not guarded. Each merged branch's own commits were guarded when they were made.

### Out of Scope — product and template changes
- No `moai` CLI subcommand, no `moai spec lint` rule, no Claude Code PreToolUse hook, no change under `internal/template/templates/`, no change to `internal/cli/hook_install_precommit.go` or the managed `.git/hooks/pre-commit`.

### Out of Scope — CI changes
- No workflow file is touched; `TestACCounterFullCorpusMatchesBaseline` remains the authoritative gate, byte-unchanged in semantics.

### Out of Scope — automatic regeneration or absorption
- The guard never writes the baseline (`ac-count-baseline-refresh.md` §6).

### Out of Scope — new bypass variable
- No environment variable bypass; the only bypass is `git commit --no-verify`.

### Out of Scope — CLAUDE.local.md edit
- Registering the dev-only paths in CLAUDE.local.md §2 "Local-Only Files" is a sync-phase concern; this SPEC does not edit that file.

## §D Success Criteria

1. An in-place amendment that moves the AC count without a regenerated baseline cannot be committed in any worktree of this repository (HEAD unchanged after the attempt).
2. The three false-positive paths — count-unchanged edit, new `acceptance.md`, amendment plus regenerated baseline — commit normally.
3. A commit with no qualifying file is untouched and silent.
4. The guard's non-execution is visible: a qualifying commit always prints either `checked` or `NOT CHECKED`.
5. The managed hook, its provenance file, and `core.hooksPath` are byte-unchanged by install.

## §E Dependencies and Prior Art

- SPEC-AC-COUNT-DISCRIMINATOR-001 — origin of the counter, the snapshot, and the comparison contract (§3.5 rules 1–4).
- SPEC-AC-BASELINE-REFRESH-001 — regeneration command, cascade doc, same-commit rule; its exclusion "mechanical cascade enforcement" defers to this card.
- `scripts/jev/` — the local-only dev-tool precedent (CLAUDE.local.md §29): script under `scripts/`, no template mirror.
