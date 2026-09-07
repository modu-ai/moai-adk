---
id: SPEC-DOCTOR-STAT-SEAM-001
title: "Route doctor_codex stat calls through the osStatFn seam"
version: "0.1.0"
status: draft
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P2
phase: "v3.2.0"
module: "internal/cli"
lifecycle: spec-anchored
tags: "doctor, codex, stat-seam, testability, characterization"
tier: S
---

# SPEC-DOCTOR-STAT-SEAM-001 — doctor_codex stat calls through the `osStatFn` seam

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-09-08 | manager-spec | Initial plan-phase draft — card t563, class C (scope decision: both sites), Tier S |

## §1 Problem Statement

`internal/cli/doctor_codex.go` contains exactly two direct `os.Stat` calls (measured against the
origin/develop blob `ef10a2524`, re-confirmed in this worktree):

- **Site A** — `doctor_codex.go:459`, inside the skill-mirror check loop: follows the mirror-entry
  symlink created by `os.Lstat` classification at `:441`.
- **Site B** — `doctor_codex.go:857`, inside `codexStaleSkillFinding` (defined `:815`, called `:319`).

The package already defines an injectable stat seam — `var osStatFn = os.Stat` at
`internal/cli/update_preserve_inventory.go:59` (doc comment `:44`) — and the seam's sibling
consumers in the SAME package all use it: `codex_skills_prune.go:96` (1 site),
`codex_skills_disable.go` (2 sites), `update_preserve_inventory.go:439` (1 site). `osStatFn`
appears **0 times** in `doctor_codex.go`: doctor is the exception, not the absence of a convention.

The consequence is testability, not behavior: tests cannot observe WHAT was stat'ed, cannot inject
a stat failure portably (the seam exists precisely because POSIX permission bits are not modeled on
Windows and are bypassed by root), and cannot run mutant-style assertions on the stat path
arguments. Concretely, `codexStaleSkillFinding` has ZERO direct tests today (verified: defined
`:815`, called `:319`, no test references), and the t540 arm-A red had to be scoped down to the
prune site for exactly this reason (motivation only — see §3.3).

This card routes BOTH doctor stat sites through `osStatFn`. It is an observability change, not a
behavior change.

## §2 Requirements (GEARS)

### REQ-001 — Stat calls resolve through the seam (Ubiquitous)

Every file-existence probe in `internal/cli/doctor_codex.go` performed with `os.Stat` shall invoke
the package's `osStatFn` seam (`update_preserve_inventory.go:59`); no direct `os.Stat` call shall
remain in the file.

### REQ-002 — Site A: mirror-entry symlink stat (Event-driven)

When the mirror-check loop follows a mirror-entry symlink, the doctor shall stat
`filepath.Join(mirrorDir, e.Name())` through `osStatFn` and shall preserve the existing three-arm
result handling unchanged: `serr == nil` (resolves — no bucket), `fs.ErrNotExist` (append the entry
name to `st.dangling`), any other error (increment `st.indeterminate`).

### REQ-003 — Site B: classified skill-path stat (Event-driven)

When `codexStaleSkillFinding` probes a classified skill path, the doctor shall stat the classified
path through `osStatFn` — the declared absolute path for `codexPathAbsolute`, the
`expandCodexHomeRelativePath` result for `codexPathHomeRelative` — and shall preserve the existing
result handling unchanged: `serr == nil` (resolves, including directories, by design),
`fs.ErrNotExist` (the per-`Enabled`-state missing arms), any other error (the indeterminate
counter). Relative (`relativeCount`) and oddly-formed (`oddlyFormed`) classifications shall
continue to be skipped without any stat call.

### REQ-004 — Zero behavior change (Unwanted)

The seam swap shall not change any doctor diagnostic output: for the same fixture state, the
problem-finding set, every finding's text, every detail string, and the check's exit behavior shall
be identical before and after the swap. Any output difference is a defect of this SPEC's
implementation, by definition.

### REQ-005 — Characterization before the swap (Event-driven)

When the seam-swap commit lands, a characterization test suite pinning `codexStaleSkillFinding`'s
current behavior — the classification buckets resolves / missing / relative / oddly-formed /
indeterminate, asserted from real fixture states in `t.TempDir()` (no seam injection is available
before the swap) — shall already exist in the commit graph: committed before the swap commit, and
passing on the unmodified tree. Characterization tests pin current behavior; they are not required
to be RED-first, and their point is to PASS before the swap exists. Any additional mirror-check
characterization the output-identity proof (REQ-004) needs lands in the same characterization
commit.

### REQ-006 — Observable stat arguments via seam override (Where)

Where a test overrides `osStatFn` with a recording function (capturing each stat path argument and
its injected result), the doctor's stat calls shall be observable: site A's recorded argument shall
be the mirror-relative joined path (`filepath.Join(mirrorDir, name)`), and site B's recorded
argument shall be the classified absolute or expanded path. At least one observation test per site
shall assert both the recorded argument and the bucket the injected result drives (injected
`ErrNotExist` → dangling/missing bucket; injected `nil` → resolves; injected other error →
indeterminate).

### REQ-007 — Local verification discipline (Where)

Where the `internal/cli` package is under local verification, the run phase shall execute scoped
runs only — `go test ./internal/cli/ -run '<relevant>' -count=1 -timeout 1800s` plus `go vet` and
`golangci-lint` on the package — and shall not run the full local suite (10+ lanes active; CI owns
the full verdict). Tests that reassign `osStatFn` shall not call `t.Parallel()` (the seam is a
package-level variable — `@MX:WARN` at `update_preserve_inventory.go:53`; follow the save → replace
→ `t.Cleanup` restore pattern of `codex_skills_prune_test.go:55-57` and
`update_preserve_partial_test.go:84-86`).

## §3 Scope

### §3.1 Settled decision — BOTH call sites (card t563 hard rule 1, encoded, not re-litigated)

SCOPE = both `os.Stat` sites in `internal/cli/doctor_codex.go` (`:459` and `:857`) routed through
`osStatFn`, in ONE commit. Rationale: same file, same primitive (`os.Stat`), same existing seam
variable — the swap is mechanical and identical for both sites. Routing only one would leave the
other as a NEW asymmetry inside the same file, which is precisely the "doctor is the exception"
defect this card removes.

### §3.2 Non-goal — `os.Lstat` stays direct

`os.Lstat` at `doctor_codex.go:441` (and every other `Lstat` use) remains a direct call. The
package convention defines only a STAT seam; no `osLstatFn` exists. Introducing a second seam
primitive is a new convention decision outside this card's named scope (`:459`/`:857` stat sites
only).

### §3.3 Motivation-only rule — t540 reference discipline

SPEC-CODEX-SKILLS-PRUNE-SERIES (t540) §G gap 5 is cited as MOTIVATION only (its arm-A red had to
be scoped down to the prune site because the doctor half had no seam). This SPEC MUST NOT claim to
"close the t540 gap": the AC the seam would open belongs to t540's M3 wording, which does not exist
yet (t540 has not landed). t540 is referenced in motivation prose and in §5 cross-references only —
never as a deliverable, an acceptance criterion, or a claimed outcome.

## §4 Out of Scope

### Out of Scope — second seam primitive (Lstat)

- No `osLstatFn` (or any Lstat seam) is introduced; `doctor_codex.go:441` and all other
  `os.Lstat` calls stay direct.

### Out of Scope — t540 deliverables

- No acceptance criterion, test, or code in this SPEC implements or claims t540's deferred AC;
  the doctor-half arm-A assertion remains t540's future M3 wording.

### Out of Scope — behavior or diagnostic changes

- No finding text, bucketing logic, classification switch (`classifyCodexSkillPath`), or diagnostic
  string is modified; the swap commit's diff touches the two stat call sites only.

## §5 Cross-References

- Seam definition + doc comment: `internal/cli/update_preserve_inventory.go:44-59`
- Sibling seam consumers (control group): `codex_skills_prune.go:96`, `codex_skills_disable.go:149,168`,
  `update_preserve_inventory.go:439`
- Seam-override test pattern: `internal/cli/codex_skills_prune_test.go:55-57`,
  `internal/cli/update_preserve_partial_test.go:84-86`
- Motivation (reference only, NOT a deliverable): t540 spec.md §G gap 5
- Card: t563 · worktree `.claude/worktrees/t563` · branch `WT-doctor-stat-shim` · base `ef10a2524`
