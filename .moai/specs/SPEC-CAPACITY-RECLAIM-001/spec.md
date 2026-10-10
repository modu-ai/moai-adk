---
id: SPEC-CAPACITY-RECLAIM-001
title: "Capacity reclaim: untrack harness state, bound sinks and archives, report-first worktree sweep"
version: "0.1.1"
status: draft
created: 2026-10-10
updated: 2026-10-10
author: manager-spec
priority: P0
phase: "v3.2.0 target"
module: "internal/hygiene"
lifecycle: spec-anchored
tags: "capacity, hygiene, sinks, usage-log, lessons-inbox, gitignore, worktree-sweep, backups, profile-cap"
tier: M
related_specs: [SPEC-MOAI-HYGIENE-001, SPEC-WORKTREE-SWEEP-001, SPEC-WORKTREE-REAPER-001]
---

# SPEC-CAPACITY-RECLAIM-001 — Capacity reclaim (plan phase)

## §A History

- **2026-10-10** — plan-phase v0.1.0 authored from factory card t1631 (operator-approved scope, 10-10; P0, Class C). Plan row 3.2-0-2 (absorbs part of t1501). Tier M: the card names the three-artifact set (spec, plan, acceptance) plus progress. `decision-index.md` is authored because `.moai/config/sections/interview.yaml` sets `decision_gate: on`.
- **2026-10-10** — v0.1.1, plan-phase finalization before the plan audit, on card worktree HEAD `2aab5f797`. The tracking, template, code-grounding, and branch measurements in §B were re-observed (the worktree-size row was not re-run). Corrected against the source: the hygiene entry point is `moai clean --audit-logs` (no `hygiene` command exists); `usage-log.jsonl` and `lessons-inbox.jsonl` are owner exceptions, not registered sinks, and SPEC-MOAI-HYGIENE-001 §D assigns them to their existing owners, so REQ-CAP-004 is now gated by decision Q4; `moai worktree sweep` already previews by default and its `--yes` path has no owner gate, so M3 and M4 are re-scoped; the WT-v3-2-1 premise is re-measured (0 commits unique to the branch, not "identical to main"); REQ-CAP-008 cited REQ-CAP-011 for operator confirmation, and the confirmation is now stated in the requirement; decision Q4 moved from DECIDED to FOUNDER because its anchor is outside the authority register. Added AC-CAP-015 so every requirement has a verifying criterion. Counts: 13 requirements and 15 acceptance criteria, against the Tier M ceilings of 16 and 16.

## §B Problem and measured baseline

Git tracks several harness state files that grow without bound, the template `.gitignore` does not exclude them, and worktrees accumulate on disk. The card reports 836 MB for `.moai/reports` and 53.6 GB for worktrees, measured on 10-10. This SPEC does not adopt those figures. It cites only the measurements below, each taken with the command shown on worktree `WT-capacity-reclaim` at HEAD `2aab5f797`. *(re-observed)* marks a row re-run for v0.1.1 with the same command on the same HEAD.

| Measurement | Command | Observed output |
|---|---|---|
| Git-tracked harness and lessons-inbox bytes | `git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl \| xargs -0 wc -c \| tail -1` | `128526661 total` across 191 tracked files *(re-observed)* |
| Largest tracked items | `git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl \| xargs -0 wc -c \| sort -n \| tail -6` | `.moai/harness/usage-log.jsonl` 88,629,933 B; `.moai/harness/learning-history/archive/2026-08.jsonl.gz.corrupt` 14,398,585 B; `…/2026-09.jsonl.gz.corrupt` 11,764,350 B; `.moai/lessons-inbox.jsonl` 6,043,267 B; `…/2026-07.jsonl.gz.corrupt` 4,179,609 B *(re-observed)* |
| Tracked `.corrupt` files | `git ls-files -- '*.corrupt*'` | Three files, all under `.moai/harness/learning-history/archive/` (2026-07, 2026-08, 2026-09) *(re-observed)* |
| Template `.gitignore` coverage | `grep -n -E 'harness\|lessons\|usage\|\.corrupt\|state\|reports' internal/template/templates/.gitignore` | Matches at line 220 (`!.agents/skills/moai-harness/`), line 242 (`.moai/state/`), lines 268–279 (`.moai/reports/*` family), and comment lines 240, 258, 263, 282. No rule excludes `usage-log.jsonl`, `lessons-inbox.jsonl`, or `*.corrupt*`. The narrower `usage-log\|lessons-inbox\|\.corrupt` pattern returns no match *(re-observed)* |
| `.moai/reports` on the primary checkout | `du -sk /Users/goos/MoAI/moai-adk-go/.moai/reports` | `857004	/Users/goos/MoAI/moai-adk-go/.moai/reports` (about 837 MiB) |
| Worktree size (primary checkout roots) | `du -sk /Users/goos/MoAI/moai-adk-go/.moai/worktrees /Users/goos/MoAI/moai-adk-go/.claude/worktrees` | `56652256	…/.moai/worktrees` and `35721380	…/.claude/worktrees`; sum 92,373,636 KiB (about 88.1 GiB), above the 10 GiB target. A foreground run exceeded the 120 s tool limit and was re-run in the background, which completed with exit 0. Not re-run in v0.1.1. The per-`git worktree list` measurement that AC-CAP-002 names is re-run in pre-flight. The card's 53.6 GB figure is not adopted. |
| `moai worktree list` subcommand | `grep -rn 'Use:.*"list"' internal/cli/worktree --include='*.go'` | No match, exit 1 *(re-observed)*. `clean.go:527` and `done.go:195` call `git worktree list --porcelain` internally; no user-facing list verb is registered. |
| Hygiene entry point | `grep -n -E 'audit-logs\|session-state' internal/cli/clean.go` | `--audit-logs` (line 111) and `--session-state` (line 112) run the hygiene units; `--apply` (line 113) is the only switch that mutates, and the default is dry-run. A `Use: "hygiene"` search of `internal/cli` returns nothing *(re-observed)* |

Code grounding (bounded read, not a survey; *(re-observed)* where the line was re-checked for v0.1.1):

- Usage-log writer: `internal/hook/failure_observer.go:65` and `:116` write `.moai/harness/usage-log.jsonl` *(re-observed)*; `internal/harness/observer.go:13` documents the recorder *(re-observed)*.
- `.corrupt` producers: `internal/manifest/manifest.go:85` renames a corrupt manifest to `.corrupt`; `internal/userassets/install.go:124` writes a `.corrupt-<timestamp>` journal sidecar; `internal/core/project/validator.go:325` removes the manifest `.corrupt`. *(all re-observed)*
- Registry and owners: `internal/hygiene/sinks.go` declares the closed `sinkRegistry` (starting at line 26), which lists neither `usage-log.jsonl` nor `lessons-inbox.jsonl`. Both appear only in `ownerExceptions` (lines 90–91): `lessons-inbox.jsonl` belongs to `internal/hook/inbox_lifecycle.go` and the `moai inbox` verbs; `usage-log.jsonl` belongs to `internal/harness/retention.go` (its owner note reads "30-day prune to monthly gz"). SPEC-MOAI-HYGIENE-001 §D (completed 2026-10-05; `spec.md` lines 107–110, "already-owned or writer-owned surfaces") records the same owners, and describes usage-log retention as "working as measured". The rotator therefore does not rotate either file today.
- Hygiene apply path: `moai clean --audit-logs` (`internal/cli/clean.go:111`) is dry-run unless `--apply` is passed (`clean.go:113`, and the mode resolution at `clean.go:159–164`). The SessionStart auto path is governed by `workflow.hygiene.mode` (`internal/hygiene/settings.go:43`; the template default is `report`, `internal/template/templates/.moai/config/sections/workflow.yaml:318`). Apply-mode audit rows are `AuditRow` values (`internal/hygiene/auditlog.go:13–22`) with `mode` set to `apply` (`internal/hygiene/engine.go:43`; `internal/hygiene/rotate.go:167` and `:184`), appended to `.moai/logs/hygiene-audit.jsonl` (`auditlog.go:26`).
- Worktree sweep: `moai worktree sweep` (`internal/cli/worktree/sweep.go:199`) previews by default; `--yes` performs disposals (`sweep.go:230`); `--json` emits every evaluation and removes nothing (`sweep.go:231`). The predicate chain (`sweep.go:405–520`: landing, dirty, ignored content, process cwd) preserves on any predicate it cannot answer. The report has no owner field and no unpushed-commit count (`grep -n -i -E 'owner|unpushed' internal/cli/worktree/sweep.go` returns no match *(re-observed)*), and `--yes` disposes without an owner-confirmation check. The disposal predicate is owned by SPEC-WORKTREE-SWEEP-001 (completed 2026-09-30); this SPEC reuses it and does not re-implement it.
- Home cleanup: `internal/cli/clean_home.go` already applies a profile cap (`config.DefaultProfileMaxBytes`, used at line 243), a releases keep-newest rule (`config.DefaultReleaseKeep`, passed at line 552 and applied at lines 441–445), and age-based retention for `backups/removed-*` (lines 310–329). The cap and the keep count are compiled constants (`internal/config/defaults.go:451` and `:455`). REQ-CAP-008 and REQ-CAP-009 extend or re-pin these rules.
- Lessons-inbox consumer: the LSEL curator's companion-offset drain reads `<state-dir>/drain-offset.json` (`.claude/skills/hns-lsel-curator/drain.sh:9` and `:48`; `SKILL.md:65`). Rotating the inbox without preserving that offset would break the drain. *(re-observed)*

## §C Requirements (GEARS)

`<subject>` is the named component in each line. Requirements are numbered REQ-CAP-NNN (13 total; within the Tier M ceiling of 16).

### §C.1 Git tracking and template

- **REQ-CAP-001 (Ubiquitous):** The template `.gitignore` shall exclude `.moai/harness/usage-log.jsonl`, `.moai/lessons-inbox.jsonl`, and every `*.corrupt*` file under `.moai/`, so that projects created from the template never track them.
- **REQ-CAP-002 (Ubiquitous):** The repository shall stop tracking `.moai/harness/usage-log.jsonl` and `.moai/lessons-inbox.jsonl`, keeping each local file in place.
- **REQ-CAP-003 (Ubiquitous):** The repository shall stop tracking the `.moai/harness/learning-history/archive/*.jsonl.gz.corrupt` files, keeping each local copy in place.

### §C.2 Sink rotation and hygiene apply

- **REQ-CAP-004 (Event-driven, decision-gated):** **Where** the operator has approved decision Q4 in `decision-index.md`, **when** the hygiene rotator runs in apply mode and observes `.moai/harness/usage-log.jsonl` at or above the rotation threshold, the rotator shall rotate that file under the bound already defined by SPEC-MOAI-HYGIENE-001 REQ-HYG-001 (one primary plus one kept rotation), and shall never truncate a file in place. Where Q4 carries no approved verdict, the file remains with its existing owner, `internal/harness/retention.go`, and this requirement does not apply.
- **REQ-CAP-005 (Unwanted):** The rotator shall not rotate `.moai/lessons-inbox.jsonl` in this SPEC, because a rotation would break the companion-offset drain that reads `drain-offset.json`. Lessons-inbox bounding is deferred to a follow-up SPEC that preserves the drain offset. Its existing owner, `internal/hook/inbox_lifecycle.go`, is unchanged.
- **REQ-CAP-006 (Where, decision-gated):** **Where** the operator has approved decision Q1 in `decision-index.md`, the shipped default of `workflow.hygiene.mode` (template `workflow.yaml`) shall be `apply`; otherwise it shall remain `report`. The `--apply` switch of `moai clean --audit-logs` is unchanged by this requirement.
- **REQ-CAP-007 (Event-driven):** When the hygiene apply path acts on any file, the hygiene apply path shall append one row with `mode` set to `apply` per applied action to `.moai/logs/hygiene-audit.jsonl`, so that the apply run is visible after the fact.

### §C.3 Archives, profile, and retention

- **REQ-CAP-008 (Ubiquitous):** On apply, the backups and releases directories shall each retain the three most recent entries (keep-3), and shall remove older entries only after the operator confirms the apply run.
- **REQ-CAP-009 (Ubiquitous):** On apply, the profile data directory shall be held to a byte cap read from configuration, never a value compiled into the code.

### §C.4 Worktree sweep and list

- **REQ-CAP-010 (Event-driven):** When `moai worktree sweep` runs in report mode (the form without `--yes`, or with `--json`), the sweep shall list each candidate worktree with its owner, branch, remote-landing predicate result, and unpushed-commit count, and shall dispose of no worktree.
- **REQ-CAP-011 (Unwanted):** The sweep shall not dispose of a worktree whose branch has unpushed commits, of a worktree without a confirmed owner, or of a worktree that a live session is using; each such worktree shall be reported as PRESERVE.
- **REQ-CAP-012 (Where):** Where the operator has approved the bulk sweep (decision Q2 in `decision-index.md`), the sweep shall dispose only the candidates the owner has confirmed, and shall record the evidence before each disposal.
- **REQ-CAP-013 (Ubiquitous):** `moai worktree list` shall print every registered worktree with its path, branch, and on-disk size, in table form and in a `--json` form whose array length equals the count of `worktree` lines in `git worktree list --porcelain`.

## §D Out of Scope

### Out of Scope — Deletion during the plan phase

- This SPEC authorizes no deletion of any worktree and no deletion or move of any `.moai/reports` file during the plan phase.
- The bulk worktree sweep is a candidate report first; disposal is operator-gated before the run phase (REQ-CAP-012).

### Out of Scope — Absorbed tree WT-v3-2-1

- The absorbed t1501 tree `WT-v3-2-1` contributes no work to this SPEC. Measured 10-10: `git rev-list --left-right --count refs/heads/main...refs/heads/WT-v3-2-1` prints `375	0`, so the branch has no commit that `main` lacks. The card's "identical to main" wording is not reproduced: the tree SHAs differ (`refs/heads/WT-v3-2-1^{tree}` is `de6cf971…`, `refs/heads/main^{tree}` is `d3ab4315…`) because `main` has advanced 375 commits past the branch.

### Out of Scope — Retention of report content

- The content of `.moai/reports` is not pruned by age or size in this SPEC. Only the sizing measurement is recorded here.

### Out of Scope — Other SPECs' lifecycle

- No change to the lifecycle, status, or ownership of SPEC-MOAI-HYGIENE-001, SPEC-WORKTREE-SWEEP-001, or SPEC-WORKTREE-REAPER-001.

### Out of Scope — Local deletion of archive payloads

- Deleting the local `.corrupt` payloads or the local usage-log is irreversible and is not authorized here; it is operator-held.

### Out of Scope — Ownership changes without a verdict

- Moving `usage-log.jsonl` from `internal/harness/retention.go` to the hygiene rotator, without an approved decision Q4 verdict.
- Changing the owner of `lessons-inbox.jsonl` (`internal/hook/inbox_lifecycle.go`) at all (REQ-CAP-005).

## §E Assumptions

- **A1 (Tier M):** The card names exactly the spec, plan, acceptance, and progress artifacts, which is the Tier M set. If the run phase exceeds 15 files, the SPEC tiers up to L. Rationale: the plan row fixes the artifact set.
- **A2 (apply rows — resolved in v0.1.1):** The apply-mode row is `AuditRow` (`internal/hygiene/auditlog.go:13–22`), with field `mode` set to `apply` (`internal/hygiene/engine.go:43`; `internal/hygiene/rotate.go:167`, `:184`), appended to `.moai/logs/hygiene-audit.jsonl` (`auditlog.go:26`). Whether a fixture apply run writes those rows is still unobserved (AC-CAP-003).
- **A3 ("싱크 회전"):** Read as rotation of registered audit sinks (`sinkRegistry`). Rationale: "sink" matches `internal/hygiene/sinks.go`.
- **A4 (lessons-inbox deferred):** Rotation of lessons-inbox is deferred (REQ-CAP-005). Rationale: the companion-offset drain reads that file by offset, and rotation would break the offset.
- **A5 (units):** "5 MB" is read as 5,000,000 bytes (decimal, stricter than 5 MiB). "10 GiB" is 10,485,760 KiB. Rationale: the stricter reading avoids a pass by unit choice.
- **A6 (`.corrupt` disposition):** Disposition is untrack, ignore, and keep local copies. Deletion of local payloads is operator-held. Rationale: the local copies are irreversible if deleted, and the git-tracked copies are what cost space.
- **A7 (worktree set):** The worktree measurement set is every root from `git worktree list --porcelain`, covering both L1 and L2 trees. Rationale: the set is defined by git, not by a directory name.
- **A8 (decision index):** `decision-index.md` is authored because `decision_gate: on` (`.moai/config/sections/interview.yaml:6`). Rationale: the setting is explicit.
- **A9 (card figures):** The card's 836 MB and 53.6 GB are not adopted as measurements. Rationale: the card is untracked and not a committed authority, and the worktree figure was not re-measured here.
- **A10 (releases keep-3 — pre-flight item):** The releases rule is read against the existing code: the current release is protected, and the three newest non-current releases are kept (`clean_home.go:441–445`; `config.DefaultReleaseKeep` = 3). Whether the current release counts toward "the three most recent entries" is pinned in run pre-flight. Rationale: the existing rule and the wording of REQ-CAP-008 differ by the current release.

## §F Risks

- **Untracking deletes the working copy in other clones.** A commit that removes a tracked file deletes the working copy in every clone that pulls it. Mitigation: operator-gated (decision Q3); each clone backs up its local file outside the repository before pulling.
- **Residual margin under the git-tracked target.** Measured arithmetic: tracked files other than the five named files (usage-log, lessons-inbox, and the three archives) total 3,510,917 bytes (128,526,661 minus 125,015,744). That leaves 1,489,083 bytes of headroom under 5,000,000 bytes. The residual includes the two non-corrupt archives `2026-07.jsonl.gz` (874,561 bytes) and `2026-08.jsonl.gz` (1,898,703 bytes), which this SPEC keeps tracked. Any new tracked harness file erodes that margin.
- **Bulk sweep may dispose the only copy of unpushed work.** Mitigation: REQ-CAP-011 and REQ-CAP-012, the owner-confirmation gate (not yet implemented; the current `--yes` path has none), and decision Q2. SPEC-WORKTREE-SWEEP-001's predicate is reused.
- **Apply-by-default changes the auto path for every user.** Mitigation: decision Q1 is FOUNDER-owned; the template default is not flipped without the operator verdict.
- **Owner conflict with a completed SPEC.** REQ-CAP-004 would reverse SPEC-MOAI-HYGIENE-001 §D for usage-log. Mitigation: the requirement is gated by decision Q4 (FOUNDER, product-level); no registry or owner-exception change is made without a verdict.

## §G Cross-references

- `.claude/rules/moai/development/spec-frontmatter-schema.md` (12 canonical fields)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (Tier M)
- `.moai/specs/SPEC-MOAI-HYGIENE-001/spec.md` § D (ownership of usage-log and lessons-inbox), § B.1 (REQ-HYG-001 bound), and HISTORY
- `internal/hygiene/sinks.go` (`sinkRegistry`, `ownerExceptions`)
- `.moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md` (verification layer)
- `.moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md` (milestones and pre-flight)
- `.moai/specs/SPEC-CAPACITY-RECLAIM-001/decision-index.md` (decision gate rows)
