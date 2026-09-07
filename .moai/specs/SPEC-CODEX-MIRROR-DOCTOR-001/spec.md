---
id: SPEC-CODEX-MIRROR-DOCTOR-001
title: "moai doctor reports .agents/skills mirror state"
version: "0.2.1"
status: completed
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "doctor, codex, skill-mirror, diagnostics, advisory"
tier: M
---

# SPEC-CODEX-MIRROR-DOCTOR-001 — `.agents/skills` mirror-state diagnostic

## HISTORY

| Version | Date | Change |
|---|---|---|
| 0.1.0 | 2026-09-07 | Initial plan-phase draft (card t498). Grounded in `.moai/reports/t498/root-cause.md`. |
| 0.2.0 | 2026-09-07 | Amendment closing the six blocking findings of plan-audit iteration 1 (`.moai/reports/t498/plan-audit.md`, FAIL 0.775). §2: REQ-CMD-006/007 gain the `Where the project declares Codex wiring` precondition, reconciling them with plan.md §D M2 (D4). §6: the copy-mode cross-platform clause is re-stated conditionally to match §3 row 3 and root-cause.md Gaps, removing an unmeasured premise asserted as fact (D3). acceptance.md 0.2.0 carries the verification-layer repairs (D1, D2, D5, D6); plan.md unchanged. |
| 0.2.1 | 2026-09-07 | acceptance.md 0.2.1: single-clause repair of plan-audit iteration 2 finding **F1** (`.moai/reports/t498/plan-audit-2.md`, PASS 0.8875, F1 major/blocking). AC-CMD-014 clause 3 asserted `check.Detail` is empty of mirror text at `verbose=false`; that is false about this tree — the verbose gate lives at the render layer (`doctor_render.go:134-137`), while `checkCodexWiring` assigns `check.Detail` unconditionally (`doctor_codex.go:196, 200`), as the passing `TestCheckCodexWiring_StaleHomeSkillsReported` (`verbose=false`, asserts Detail) demonstrates. The clause is re-pointed at `renderDoctorGroups(w, groups, false, th)` output, resolving its contradiction with clause 2 and REQ-CMD-001. `Decided by` line, swept-count gate, and spec.md/plan.md bodies unchanged. |

## §1 Context

The skill mirror (`.agents/skills/<name>` → `../../.claude/skills/<name>`) is what makes a MoAI
skill catalog reachable from Codex CLI, which does not scan `.claude/skills`. The mirror is created
at deploy time only, inside `DeployWithResult`, and its observable outcome (`MirrorModeSymlink` /
`Copy` / `Skipped` / `Failed`) is consumed only by the two deploy-time notice callers.

Measured consequence (root-cause.md Claim 2, reproduced in a throwaway probe project): a routine
`moai update --yes` on a version-matched project short-circuits before the deploy step, so a deleted
mirror stays deleted and no notice is printed. Only `moai update --templates-only --force --yes`
restored it. A missing, stale, or broken mirror is therefore invisible and self-perpetuating on any
project whose binary version already matches.

Measured absence (root-cause.md Claim 4): `moai doctor` carries no mirror row. `grep -rn '.agents'
--include='*.go' internal/cli/` returns 5 hits, all in `_test.go` files. The `doctor_codex.go` skill
logic reads the **user-layer** `~/.codex/config.toml` registrations, which is a different subject
entirely from the project-local mirror.

This SPEC closes the observability gap only: it makes the mirror's state **readable**.

## §2 Requirements (GEARS)

REQ-CMD-001 — Ubiquitous. The mirror diagnostic shall live inside `checkCodexWiring`
(`internal/cli/doctor_codex.go:86`) and shall report through that file's existing `codexFinding`
two-register shape (`summary` for Message, `detail` for `--verbose`).
Source read: `internal/cli/doctor_codex.go` lines 75-78, 86-207.

REQ-CMD-002 — Ubiquitous (unwanted). The mirror diagnostic shall not create, repair, remove,
replace, or modify any path under `.agents/`, and shall not modify `.claude/skills/`. It reads only.
Source read: `internal/cli/doctor_codex.go` header comment lines 12-15 ("The check READS only").

REQ-CMD-003 — State-driven (the un-nagging invariant). While the project declares no Codex wiring
files and no `codex` binary resolves on PATH, `checkCodexWiring` shall return its informational skip
unchanged, and shall emit no mirror observation in either Message or Detail.
Source read: `internal/cli/doctor_codex.go` lines 96-103 and the comment "This is the un-nagging
invariant — it must survive every addition below."

REQ-CMD-004 — Where/When (compound). Where the project declares Codex wiring, when the mirror
directory `.agents/skills` is absent, the check shall raise a finding whose summary names the absent
mirror and carries the forced re-deploy directive.
Source read: `internal/template/skill_mirror.go:52` (`mirrorSkillsRelDir`), root-cause.md Claim 2.

REQ-CMD-005 — Where/When (compound). Where the project declares Codex wiring, when one or more
mirror entries are symlinks whose target under `.claude/skills/` no longer exists, the check shall
raise a finding naming the count of dangling entries and carrying the same re-deploy directive.
Source read: `internal/template/skill_mirror.go:151-153` (`mirrorLinkTarget` — the relative link
body a dangling check must resolve against the mirror directory).

REQ-CMD-006 — Where/While (compound). Where the project declares Codex wiring, while the mirror
directory exists, the check shall count mirror entries materialized as real directories rather than
symlinks, and shall report that count in Detail only — never as a Message finding. Where the project
declares no Codex wiring, the count is not computed and not reported, whatever the mirror's state
(the inspector is called only in the wired branch — plan.md §D M2).
Source read: `internal/template/skill_mirror.go:38-42` (`MirrorModeCopy`, `MirrorModeSkipped`) and
lines 234-250 (copy fallback and its "does not follow later updates" warning).

REQ-CMD-007 — Where/While (compound). Where the project declares Codex wiring, while the mirror
directory exists, the check shall count entries present in `.claude/skills/` with no corresponding
entry in `.agents/skills/`, and shall report that count in Detail only — never as a Message finding.
Where the project declares no Codex wiring, the count is not computed and not reported, whatever the
mirror's state (same call-site restriction as REQ-CMD-006).

REQ-CMD-008 — State-driven. While the project declares no Codex wiring and `codex` resolves on PATH,
the check shall emit no mirror finding: the existing `initCodexAdvice` finding already directs the
user to a deploy that creates the mirror.
Source read: `internal/cli/doctor_codex.go` lines 110-126 and 52 (`initCodexAdvice`).

REQ-CMD-009 — Ubiquitous. Every mirror summary shall stay within `codexMessageWidthCeiling` (113
runes) on its own, and shall be appended to `problems` so it participates unchanged in the existing
`joinCodexSummaries` tail-drop truncation.
Source read: `internal/cli/doctor_codex.go` lines 62-69, 226-252.

REQ-CMD-010 — When (event-detected). When the mirror directory or one of its entries cannot be read
(permission denied, I/O error, symlink loop), the check shall record the condition as indeterminate
in Detail and shall not raise a finding — an unobserved absence is never reported as absent.
Source read: `internal/cli/doctor_codex.go` lines 442-450 (the indeterminate posture already in
`codexStaleSkillFinding`).

REQ-CMD-011 — Ubiquitous. The mirror diagnostic shall be verifiable through the existing
`codexWiringLookPath` and `codexUserHomeDir` test seams plus a `t.TempDir()` project root, without
reading the developer's real `$HOME`, real `~/.codex/config.toml`, or real PATH.
Source read: `internal/cli/doctor_codex_test.go` lines 36-81 (`stubMoaiLookup`, `stubCodexLookup`,
`stubCodexHome`).

## §3 Findings vs verbose-only detail — the classification and its justification

Four mirror states were named as candidates. Two are Message-visible findings; two are Detail-only.

| State | Classification | Justification |
|---|---|---|
| Mirror directory entirely absent | **Finding** (REQ-CMD-004) | Codex sees no MoAI skills at all. Measured self-perpetuating: a routine `moai update` does not repair it (root-cause.md Claim 2). A wired project in this state has lost the whole point of the mirror. |
| Mirror entry is a dangling symlink | **Finding** (REQ-CMD-005) | The entry claims a skill that is not there. Mechanically decidable with one `Lstat` plus one `Stat`, with no denominator ambiguity. `mirrorOneSkill` replaces a *wrong-target* link on the next deploy, but nothing repairs it in between. |
| Mirror entry is a real directory (copy fallback) | **Detail only** (REQ-CMD-006) | It is functional — Codex reads real files — and it is the *expected* materialization wherever symlink creation is unavailable (Windows without the privilege, exotic filesystems, sandboxes). Promoting it to a finding would hand every such user a permanent warning row for a working mirror, which is the un-nagging invariant failing by a different door. The real cost is staleness ("the copy does not follow later updates", `skill_mirror.go:249`), worth counting but not worth warning. Root-cause.md Gaps records that Windows copy-fallback behaviour was **not exercised**, so the frequency of this state is unmeasured — itself a reason not to escalate it. |
| Skill in `.claude/skills/` with no mirror entry | **Detail only** (REQ-CMD-007) | The correct denominator is "skills this deploy mirrored", which doctor cannot observe. `mirrorSkills` is called with the skill list of the deploying run; a project may legitimately carry locally-authored skills that no deploy ever mirrored. Root-cause.md measured 24/24 in a fresh probe and 37 after a forced sync — it never measured a tree carrying non-deployed skills, so a count-based finding here would warn on a healthy project. Counted, reported, never warned. |

## §4 The action directive

The directive is the forced re-deploy:

```
moai update --templates-only --force --yes
```

Verified two ways: this exact invocation restored the mirror from 0 to 37 entries in the probe
(root-cause.md Claim 2 table), and every flag in it exists on the current binary (`moai update
--help`: `--templates-only`, `--force`, `--yes`). The routine form `moai update --yes` is
deliberately NOT the directive — it was measured to leave the mirror at 0.

Per the file's existing convention the directive rides in the `summary`, not only in `detail`,
because Detail renders only under `--verbose`.

## §5 Exclusions

### Out of Scope — creating the mirror in this repository

- Running any deploy in this repository, worktree, or the primary checkout to materialize a
  `.agents/skills` mirror. A deploy would create it; there is nothing to implement.
- Committing a mirror. `.agents/` is gitignored here (root-cause.md Claim 3), so the fix for this
  repository's absent mirror is a deploy, not a commit.

### Out of Scope — repairing a mirror from doctor

- Creating, replacing, relinking, or removing any `.agents/skills` entry from the doctor path. The
  check reads and reports; repair remains a deploy-time concern (REQ-CMD-002).
- Deleting a real directory that occupies a mirror path. `mirrorOneSkill` deliberately leaves such
  an entry untouched (`MirrorModeSkipped`); doctor inherits that posture.

### Out of Scope — the update short-circuit

- Changing `runTemplateSyncWithProgress`, the `syncSkipped` early return at
  `internal/cli/update.go:509-523`, or the ordering of the "Deploy Templates" step. The short-circuit
  is the *reason* this diagnostic is needed; altering it is a separate decision with a separate blast
  radius.

### Out of Scope — adjacent findings surfaced by the investigation

- The `.gitignore:133` blanket `.agents/` divergence from the shipped template policy block
  (root-cause.md Claim 3). A real candidate defect, and a separate card.
- Mirror lifecycle cleanup (removing a mirror whose skill was renamed or retired), which
  `skill_mirror.go:19-20` already assigns to the clean path.
- Anything touching the user-layer `~/.codex/config.toml` or its `[[skills.config]]` entries.

## §6 Constraints

- Advisory and fail-open. The check never gates; `moai doctor` exit status is unchanged by any mirror
  state.
- No new top-level `DiagnosticCheck` row. The mirror observation is folded into the existing "Codex
  Wiring" row, so a claude-only user's panel gains no line.
- Cross-platform. Path handling uses `filepath.Join`; the check must build and behave on
  linux/darwin/windows. Where symlink creation is unavailable (Windows without the privilege,
  exotic filesystems, sandboxes) a copy-mode mirror is the expected materialization — its frequency
  is unmeasured, and root-cause.md Gaps records the Windows copy-fallback path as unexercised. The
  check must therefore treat copy mode as a normal state to count, never as a state to warn on
  (§3, row 3).
