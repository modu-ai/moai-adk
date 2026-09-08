---
id: SPEC-CODEX-DOCTOR-PATH-GUARD-001
title: "The doctor-side path conversion gets a guard that fails when it is removed"
version: "0.1.0"
status: in-progress
created: 2026-09-08
updated: 2026-09-08
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: internal/cli
lifecycle: spec-anchored
tier: S
depends_on: [SPEC-CODEX-SKILL-PATH-READBACK-001]
related_specs: [SPEC-CODEX-SKILL-PATH-SLASH-001, SPEC-CODEX-STALE-SPLIT-FOURTH-001]
tags: "codex, doctor, skills-config, windows, path-conversion, regression-guard, t570"
---

# SPEC-CODEX-DOCTOR-PATH-GUARD-001 — The doctor-side path conversion gets a guard that fails when it is removed

## HISTORY

- 2026-09-08 (plan-phase, v0.1.0) Initial authoring. Card t570, worktree `.claude/worktrees/t570`,
  branch `WT-codex-doctor-guard`, base `a4855f0b2` (= `origin/develop` at entry; a dated anchor, not
  an AC range edge — AC ranges re-derive `CARD_BASE=$(git merge-base origin/develop HEAD)` at read
  time). Tier S, one milestone, `cycle_type tdd` (`.moai/config/sections/quality.yaml`
  `development_mode: tdd`, read this run). The card closes a guard gap left by
  SPEC-CODEX-SKILL-PATH-READBACK-001 (t562): the production line that card landed on the doctor side
  is correct and unguarded.

## §A. User Story

**As a** MoAI-ADK maintainer, **I want** the doctor-side declared-path conversion at
`codexStaleSkillFinding` to be pinned by a test that fails when the conversion is removed, **so
that** a later refactor cannot silently revert it to the declared form and still ship a green CI.

**As a** MoAI-ADK user on Windows whose `~/.codex/config.toml` declares a skill under an
extended-length path, **I want** the reverted form to be caught before release, **so that** a
healthy registration under `\\?\C:\...` is not reported stale by a check that stat'd a form Windows
will not resolve.

## §B. Context and Background

### §B.1 The unguarded line, measured in this tree

`internal/cli/doctor_codex.go`, inside the `case codexPathAbsolute:` arm of
`codexStaleSkillFinding` (line hint 837, measured this run at `a4855f0b2` — the symbol is the
address, the number is a hint):

```go
statPath = fromConfigPath(e.Path, configPathSeparator)
```

Three measurements taken in this tree at `a4855f0b2` establish that the line carries no regression
guard:

| # | Command | Observed |
|---|---|---|
| B-1 | `/usr/bin/grep -rln fromConfigPath internal/cli/*_test.go` | `internal/cli/codex_config_path_test.go`, `internal/cli/codex_skills_prune_readback_test.go` — no doctor test names the symbol |
| B-2 | t562's own mutant record (`SPEC-CODEX-SKILL-PATH-READBACK-001`, `status: completed`), `progress.md:105-111` | four mutant logs — `mutant-bypass.log`, `mutant-blanket-wrap.log`, `mutant-reorder.log`, `mutant-seam.log`. The first three are behavioural and prune-side; the fourth targets the AC-CSRB-007 delta predicate and was exercised as an UNCOMMITTED working-tree change. None touches the doctor-side conversion. |
| B-3 | t562 `acceptance.md` AC-CSRB-006, read this run | explicitly "green-by-construction regression guard; no RED claimed"; its expectations are darwin-identity, and it names this as "the honest ceiling of doctor-side verification" |

B-3 is the load-bearing one, and t562 says so about itself. Its `progress.md:286-289` records the
gap verbatim:

> the three mutants are all prune-side and the AC-CSRB-006 guard is blind to the conversion
> because darwin's separator makes it the identity

 On this host `configPathSeparator == filepath.Separator == '/'`, so
`fromConfigPath` is the identity function and AC-CSRB-006 cannot distinguish a converted stat target
from an unconverted one. The consequence, stated plainly as this card's problem:

> Reverting `internal/cli/doctor_codex.go` to `statPath = e.Path` leaves CI green.

That is the regression this card closes.

### §B.2 Why the extended-length family is the concrete justification

t562 landed the conversion on a structural argument (the doctor diff has the same one-line shape as
the behaviorally-verified prune side). Its SPEC BODY and its acceptance criteria never named a path
family for which the unconverted form actually fails — but **t562 itself did name one, at close**:
its sync-audit finding F6, recorded at `progress.md:292-293`, identifies extended-length paths
(`//?/C:/…` → `\\?\C:\…`) as "the one family where this conversion is load-bearing, which the SPEC
never named", and issues it forward as a follow-up card rather than closing it.

This card is the discharge of that forward-issued finding, not its discovery. What is new here is
the guard, not the family.

Windows accepts forward slashes in ordinary absolute paths, so `C:/Users/u/SKILL.md` resolves
whether or not it is converted — every ordinary absolute family degrades gracefully, which is
exactly why the omission was survivable. The **extended-length prefix does not**: inside a `\\?\`
prefix Windows performs no separator normalization, so `//?/C:/Users/u/skills/probe/SKILL.md` fails
to resolve while `\\?\C:\Users\u\skills\probe\SKILL.md` resolves. On that family the unconverted
form is not merely untidy — it produces a false "missing path" verdict against a healthy
registration, on the check whose Detail string tells the user to remove the entry.

**The reachability premise, stated as an assumption this card's own test verifies rather than
assumes**: the entry only enters the converting arm if `classifyCodexSkillPath` returns
`codexPathAbsolute`, which requires `filepath.IsAbs("//?/C:/...")` to be true on darwin (the leading
`/` makes it so). Measured this run with a throwaway `go run` probe:

```
IsAbs("//?/C:/Users/u/skills/probe/SKILL.md") = true ; backslash-converted = "\\?\C:\Users\u\skills\probe\SKILL.md" ; containsBackslash = false
IsAbs("C:/Users/u/SKILL.md")                  = false ; backslash-converted = "C:\Users\u\SKILL.md"                  ; containsBackslash = false
```

The premise is measured, not assumed — and REQ-CDPG-002 requires the test itself to assert the
classification rather than rely on this record.

### §B.3 Correction to the ordering guard's discriminator

The dispatch for this card proposed that the ordering guard assert **zero stat calls** on a
slash-form Windows declaration. Read against the code, zero stat calls is true in **both** orders
and therefore discriminates nothing:

- Classification-first (correct): `C:/Users/u/SKILL.md` → `IsAbs` false, no `~`, no backslash →
  `codexPathRelative` → `relativeCount++`, `continue` — before `osStatFn`.
- Convert-first (forbidden): the string becomes `C:\Users\u\SKILL.md` → contains a backslash →
  `codexPathOddlyFormed` → `oddlyFormed++`, `continue` — also before `osStatFn`.

Both branches `continue` ahead of the stat call, so the counter is 0 either way. The observable
discriminator on the doctor side is the **finding Detail string**, which renders the two
classifications with distinct wording (`"%d relative %s (not checked: the resolution base is not
observed)"` versus `"%d oddly-formed %s (not checked: backslash or ~other-user shape)"`). REQ-CDPG-003
is written against the Detail string; the zero-stat-count assertion is retained as a supporting,
explicitly non-discriminating check. This differs from t562's AC-CSRB-003, which asserted a
`SkipReason` on the prune side — the doctor side has no per-entry verdict object to read.

## §C. Requirements (GEARS)

### REQ-CDPG-001 — Doctor-side conversion guard

While `configPathSeparator` is pinned to `'\\'`, when `codexStaleSkillFinding` judges a config
entry declaring an absolute slash-form path, the test suite shall observe that the argument handed
to `osStatFn` equals the `fromConfigPath`-converted form and shall not equal the declared form.

### REQ-CDPG-002 — Extended-length family, named and reachability-verified

While `configPathSeparator` is pinned to `'\\'`, when the declared path is
`//?/C:/Users/u/skills/probe/SKILL.md`, the test suite shall observe the `osStatFn` argument to be
`\\?\C:\Users\u\skills\probe\SKILL.md`, and shall separately assert that
`classifyCodexSkillPath` returns `codexPathAbsolute` for that declaration — the reachability
premise is verified in the test, never carried from this document.

### REQ-CDPG-003 — Classification precedes conversion

While `configPathSeparator` is pinned to `'\\'`, when `codexStaleSkillFinding` judges an entry
declaring `C:/Users/u/SKILL.md`, the finding Detail shall report the entry as **relative**, shall not
report it as oddly-formed, and the stat recorder shall have counted zero calls.

### REQ-CDPG-004 — The guard's RED is observed, not asserted

When the guard has landed, the maintainer shall revert `internal/cli/doctor_codex.go` to
`statPath = e.Path`, run the affected package, and record the guard FAILING with verbatim output;
shall then undo the revert and record it passing. A guard whose RED was never observed shall not be
adopted.

### REQ-CDPG-005 — Card-unique helper naming

When this card introduces a test helper or type, the identifier shall not collide with
`statRecorder` (already declared at `internal/cli/doctor_codex_stale_skill_test.go`) or with
`pruneReadbackStatRecorder`; the existing `stubStatRecording`, `stubStatInject`,
`writeCodexHomeConfig`, `stubCodexHome`, and `overrideSeparator` helpers shall be reused in
preference to declaring new ones.

### REQ-CDPG-006 — Shared-state test discipline

While a test overrides `configPathSeparator`, `osStatFn`, or `codexUserHomeDir` — all package-level
vars — that test shall not call `t.Parallel()` and shall restore every override through `t.Cleanup`.

### REQ-CDPG-007 — Test-only scope

This card shall introduce no production change: at card close, `git diff` against `CARD_BASE`
restricted to `internal/cli/doctor_codex.go` shall be empty. When the implementer judges a
production change necessary, the implementer shall report it as a blocker finding and stop, rather
than widening scope silently.

## §D. Exclusions

### Out of Scope — production behaviour

- Changing `internal/cli/doctor_codex.go`, `classifyCodexSkillPath`, `fromConfigPath`, or
  `configPathSeparator`. The correct line already exists; this card only guards it.
- Extending the conversion to the `codexPathHomeRelative` branch — t562 REQ-CSRB-002 deliberately
  left that branch untouched, and revisiting it is a separate decision.

### Out of Scope — the prune side

- Any further guard, mutant, or coverage on `codex_skills_prune.go`. t562 already carries three
  executed mutants there; re-guarding it adds no discriminating power.

### Out of Scope — real Windows verification

- Running the check on a Windows host. The guard pins the stat TARGET on this host by overriding
  the separator; whether Windows then resolves that target is a CI-matrix question, not a
  unit-test one.

### Out of Scope — destructive paths

- Exercising `moai clean --codex-skills` or any deletion branch. The tests assert stat targets and
  finding text only, against fixtures under `t.TempDir()`; the real `~/.codex/config.toml` is
  neither read nor written.

## §E. Assumptions and Residual Risk

- **A-1 (verified, §B.2)** `filepath.IsAbs("//?/C:/…")` is true on darwin, so the extended-length
  declaration reaches the converting arm. REQ-CDPG-002 re-verifies it inside the test.
- **A-2 (verified, §B.3)** The relative/oddly-formed split is observable in the finding Detail
  string. If a later change collapses those two renderings, REQ-CDPG-003 loses its discriminator
  and must be re-anchored rather than silently weakened.
- **R-1** The guard pins the argument handed to `osStatFn`, not the filesystem outcome. It cannot
  establish that Windows resolves the converted form; §B.2 cites that as documented platform
  behaviour, and the SPEC does not claim to have measured it.
- **R-2** Mutants this guard does NOT catch are evidence of its boundary and are recorded, not
  omitted (AC-CDPG-004).

## §F. Cross-References

- `SPEC-CODEX-SKILL-PATH-READBACK-001` (t562, completed) — landed the guarded line; its AC-CSRB-006
  is the darwin-identity ceiling this card lifts. This card discharges t562's forward-issued
  sync-audit findings **F2** (the doctor-side revert is caught by nothing; `progress.md:286-289`)
  and **F6** (the extended-length family; `progress.md:292-293`).
- `SPEC-CODEX-SKILL-PATH-SLASH-001` (t540) — landed the `toConfigPath` / `fromConfigPath` /
  `configPathSeparator` seam.
- `SPEC-CODEX-STALE-SPLIT-FOURTH-001` — owns the four-state missing-count split rendered in the same
  Detail string this card reads.
