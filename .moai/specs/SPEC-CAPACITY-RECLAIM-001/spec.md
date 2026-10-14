---
id: SPEC-CAPACITY-RECLAIM-001
title: "Capacity reclaim: untrack harness state, bound sinks and archives, warning-only profile cap"
version: "0.2.1"
status: in-progress
created: 2026-10-10
updated: 2026-10-14
author: manager-spec
priority: P0
phase: "v3.2.0 target"
module: "internal/hygiene"
lifecycle: spec-anchored
tags: "capacity, hygiene, sinks, usage-log, lessons-inbox, gitignore, backups, profile-cap"
tier: M
related_specs: [SPEC-MOAI-HYGIENE-001, SPEC-WORKTREE-SWEEP-001, SPEC-WORKTREE-REAPER-001]
---

# SPEC-CAPACITY-RECLAIM-001 — Capacity reclaim (plan phase)

## §A History

- **2026-10-10** — plan-phase v0.1.0 authored from factory card t1631 (operator-approved scope, 10-10; P0, Class C). Plan row 3.2-0-2 (absorbs part of t1501). Tier M: the card names the three-artifact set (spec, plan, acceptance) plus progress. `decision-index.md` is authored because `.moai/config/sections/interview.yaml` sets `decision_gate: on`.
- **2026-10-10** — v0.1.1, plan-phase finalization before the plan audit, on card worktree HEAD `2aab5f797`. The tracking, template, code-grounding, and branch measurements in §B were re-observed (the worktree-size row was not re-run). Corrected against the source: the hygiene entry point is `moai clean --audit-logs` (no `hygiene` command exists); `usage-log.jsonl` and `lessons-inbox.jsonl` are owner exceptions, not registered sinks, and SPEC-MOAI-HYGIENE-001 §D assigns them to their existing owners, so REQ-CAP-004 is now gated by decision Q4; `moai worktree sweep` already previews by default and its `--yes` path has no owner gate, so M3 and M4 are re-scoped; the WT-v3-2-1 premise is re-measured (0 commits unique to the branch, not "identical to main"); REQ-CAP-008 cited REQ-CAP-011 for operator confirmation, and the confirmation is now stated in the requirement; decision Q4 moved from DECIDED to FOUNDER because its anchor is outside the authority register. Added AC-CAP-015 so every requirement has a verifying criterion. Counts: 13 requirements and 15 acceptance criteria, against the Tier M ceilings of 16 and 16.
- **2026-10-10** — v0.1.2, plan-audit iteration-1 revision. The audit (`.moai/reports/t1631/plan-audit.md`) returned FAIL at `40cc1031c` with 14 blocking defects (D1–D14), five optional items (O1–O5), and must-pass MP-2 failed. The revision (a) re-read the rotator, the sinks, the retention prune, the sweep, the landing predicate, the card store, and the home cleanup at `40cc1031c`; (b) states the usage-log mechanism as a declared directory plus a serialized prune (§C.2, plan M1); (c) reuses the sweep's single evaluation so the owner and unpushed predicates reach both the report and `--yes` (§B.2, §C.4); (d) reconciles keep-3 with the existing "current plus three" releases rule and names the profile-cap key (§C.3, §E A10–A13); (e) pins the code measurements to the audited tree `40cc1031c`. Counts: 14 requirements (REQ-CAP-001–014) and 16 acceptance criteria (AC-CAP-001–016), against Tier M ceilings of 16 and 16. Decisions Q1–Q4 remain OPEN; no operator verdict is recorded.
- **2026-10-14** — v0.2.0, the final authorized repair round (leader decision d-20261010T080627Z-987a: one repair inside fix_scope, one final delta re-audit). The card worktree absorbed local `main` (merge commit `e2658f3012`, 689 commits, zero conflicts), and every code path and number this SPEC pins was re-measured on that merged tree; the pins in §B are `e2658f3012` observations, and of the paths this SPEC measures only `internal/config/defaults.go` and the template `workflow.yaml` differ from the iter3-audited tree `ab9f5a6d3` (`git diff --name-only ab9f5a6d3 e2658f3012 -- internal/` lists 218 files; those two are the only measured ones among them). The operator's narrowed card-A scope (2026-10-10, recorded on the card queue record) splits the v0.1.2 scope: card A keeps the tracking change, the non-destructive apply split, the report-only destructive and keep-3 categories, and the warning-only profile cap; the worktree-sweep report fields and disposal narrowing, `moai worktree list`, the bulk-run evidence rows and the 10 GiB target move to card B (resolving D5 and D6 by excision); the `moai doctor` cap read-path (`internal/cli/doctor_disk.go:155`, `:261`) and the removal of the compiled `DefaultProfileMaxBytes` move to card D (resolving D10); the inbox rotation moves to card t1653. The repair answers the iter3 card-A defects (`.moai/reports/t1631/plan-audit-iter3.md`): D1 (AC approved-branch cell re-observed with the working command, EV-008), D2 (dissolved — the unrunnable cell's criterion moved to card B), D3 (the untracking criterion restored to RB with a four-element cell), D4 (the exit-capture rule restated from this session's runner observations; the `bash -c` wrapper withdrawn), D7 (§D.1 rotation mechanics deleted), D8 (the archive-before-rewrite guarantee now has a default-branch criterion, AC-CAP-009), D9 (the SessionStart split now has a criterion, AC-CAP-012), D11 (the absent-key residual rewritten). One scope addition was forced by the merged tree and needs operator ratification: the retention prune's monthly archives (`learning-history/archive/*.jsonl.gz`, six tracked files, 4,929,798 B) grew past the v0.1.2 measurement (a new `2026-09.jsonl.gz`, 2,142,019 B, landed tracked), pushing the post-untracking residual to 5,230,544 B — above the operator's 5,000,000-byte completion criterion; REQ-CAP-010 adds the archive untracking (local copies kept), which restores the criterion (residual 300,746 B). The Q3 gate is satisfied: the per-clone measurement recipe was executed on 2026-10-10 (`.moai/reports/t1631/q3-clone-measurement.md`) and the untracking is authorized by the operator's narrowed-scope decision. Counts: 10 requirements (REQ-CAP-001–010) and 13 acceptance criteria (AC-CAP-001–013), against Tier M ceilings of 16 and 16.
- **2026-10-14** — v0.2.1, the authorized one-clause repair (operator decision d-20261014T025128Z-0aa1, kind ceiling-exception: repair only the blocking defect D12 of `.moai/reports/t1631/plan-audit-final.md`, then exactly one re-audit; another FAIL returns the card to hold). AC-CAP-012 clause (1) now seeds one `verify` residue-class target of the closed registry (`internal/hygiene/targets.go:29–36`) in the fixture project and requires at least one `"unit":"gc"` / `"mode":"apply"` row for it in `.moai/logs/hygiene-audit.jsonl`, so the cache/GC apply half the criterion claims is observable; the premise cells EV-015 and EV-016 backing the added conjunct were observed on `111493b71` and EV-008 was re-observed there unchanged. No requirement, criterion, milestone, or code path outside that clause changed; the REQ-CAP-010 ratification remains with the operator. Counts unchanged: 10 requirements (REQ-CAP-001–010) and 13 acceptance criteria (AC-CAP-001–013), against Tier M ceilings of 16 and 16.

## §B Problem and measured baseline

Git tracks several harness state files that grow without bound, the template `.gitignore` does not exclude them, and the retention prune writes its monthly archives into the same tracked directory. The card reports 836 MB for `.moai/reports` and 53.6 GB for worktrees, measured on 10-10. This SPEC does not adopt those figures. It cites only the measurements below. All observations in this section were taken on tree `e2658f30127375e75e0c606d09ca97be2c551cf2` (short `e2658f3012`, branch `WT-capacity-reclaim`, clean) in the v0.2.0 revision session; the commit that carries this revision changes only SPEC artifacts.

| Measurement | Command | Observed output |
|---|---|---|
| Git-tracked harness and lessons-inbox bytes | `git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl \| xargs -0 wc -c \| tail -1` | `132585731 total` across 191 tracked files |
| Largest tracked items | `git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl \| xargs -0 wc -c \| sort -n \| tail -6` | `.moai/harness/usage-log.jsonl` 90,331,575 B; `.moai/harness/learning-history/archive/2026-08.jsonl.gz.corrupt` 14,398,585 B; `…/2026-09.jsonl.gz.corrupt` 11,764,350 B; `.moai/lessons-inbox.jsonl` 6,681,068 B; `…/2026-07.jsonl.gz.corrupt` 4,179,609 B |
| Tracked monthly archives | `git ls-files -z -- .moai/harness/learning-history/archive \| xargs -0 wc -c \| sort -n` | Six `.jsonl.gz` archives — `2026-04` 149 B, `2026-06` 1,371 B, `2026-05` 12,995 B, `2026-07` 874,561 B, `2026-08` 1,898,703 B, `2026-09` 2,142,019 B (4,929,798 B total) — plus the three `.corrupt` files (30,342,544 B); 35,272,342 B across 9 files. The `2026-09.jsonl.gz` archive landed tracked after the v0.1.2 measurement and is what pushed the post-untracking residual over the 5,000,000-byte target (§A v0.2.0) |
| Tracked `.corrupt` files | `git ls-files -- '*.corrupt*'` | Three files, all under `.moai/harness/learning-history/archive/` (2026-07, 2026-08, 2026-09) |
| Template `.gitignore` coverage | `grep -n -E 'usage-log\|lessons\|corrupt' internal/template/templates/.gitignore` | No match. The rules that do match are unrelated paths such as `.moai/logs/*` (line 253) and `.moai/reports/*` (line 268); no rule names an archive or a bare `*.jsonl.gz` pattern. The template has no `archive` or `learning` rule at all (`grep -n -E 'archive\|learning'` prints nothing) |
| Repository `.gitignore` in this worktree | `git check-ignore -v .moai/reports/t1631/m2-baseline.md` | `.gitignore:235:.moai/reports/*` (exit 0): the reports path is ignored by the repository, so a tracked baseline cannot live under `.moai/reports/` |
| Repository rule that masks the template probe | `git -c core.excludesFile=internal/template/templates/.gitignore check-ignore --no-index -v .moai/harness/usage-log.jsonl` | `.gitignore:352:.moai/harness/usage-log.jsonl` (exit 0): the repository's own rule answers first, so a probe in this worktree cannot isolate the template. AC-CAP-003 therefore uses the isolated test (plan §D.2) |
| `.moai/reports` on the primary checkout | `du -sk /Users/goos/MoAI/moai-adk-go/.moai/reports` | `881848	/Users/goos/MoAI/moai-adk-go/.moai/reports` (about 861 MiB) |
| Hygiene entry point | `grep -n -E 'BoolVar\(&(force\|home\|auditLogs\|sessionState\|apply)' internal/cli/clean.go` | `--force` (line 106), `--home` (107), `--audit-logs` (111), `--session-state` (112), `--apply` (113). `--audit-logs` and `--session-state` run the hygiene units and mutate only with `--apply`; the home cleanup deletes only with `--force` |

Code grounding (read at `e2658f3012`; `internal/config/defaults.go` and the template `workflow.yaml` are the only measured paths that differ from the iter3-audited tree `ab9f5a6d3`):

- Rotator resolution: `internal/hygiene/rotate.go:27` defines `LogDir` as `<projectRoot>/.moai/logs`; `rotate.go:102` (report), `rotate.go:161` and `:186` (apply and skip rows) resolve every registered name as `filepath.Join(r.LogDir, name)`. A registered `usage-log.jsonl` would therefore resolve to `.moai/logs/usage-log.jsonl`, which is not the file the writers use.
- Real location of the file: the writers are `internal/hook/failure_observer.go:65` and `:116` and the recorder documented at `internal/harness/observer.go:13`, all writing `.moai/harness/usage-log.jsonl`.
- Registry and owners: `internal/hygiene/sinks.go:26–50` declares the closed `sinkRegistry`, which lists neither state file. `ownerExceptions` (`sinks.go:89–92`) names `lessons-inbox.jsonl` (`internal/hook/inbox_lifecycle.go`, `:90`) and `usage-log.jsonl` (`internal/harness/retention.go`, "30-day prune to monthly gz", `:91`). The completeness guard skips owned names (`sinks.go:150`).
- Serialization: the rotator's pass lock is `<LogDir>/.hygiene-rotation.lock` (`internal/hygiene/rotate_lock_unix.go:13`, `rotate_lock_windows.go:14`). The 30-day prune takes an exclusive lock on `<log>.prune-state` (`internal/harness/retention.go:136`, `:168`). The two locks are different files, so the rotator and the prune do not exclude each other today.
- Displacement: `rotate.go:240–244` removes the existing chunk `<primary>.1` before the staged primary is promoted to `.1` (`os.Remove(chunk)` at `:242`). The removed chunk is not archived.
- Prune ordering (the REQ-CAP-004 guarantee): `internal/harness/retention.go:400` (`prune`) partitions the log (`:404`), archives the stale events to `archiveDir/<YYYY-MM>.jsonl.gz` (`:416`, `archiveEvents` at `:503`), and only then rewrites the log with the kept events (`:421`, `overwriteWithEvents` at `:628`). The prune rewrites the file, so its size can decrease at any time.
- Hygiene apply paths: the CLI path constructs the rotator from `workflow.hygiene` settings at `internal/cli/clean.go:169–173`; the SessionStart auto path constructs the engine at `internal/hook/session_start_hygiene.go:38` and runs it at `:49`, with the mode taken from `workflow.hygiene.mode` alone (`:37`) and the pass invoked best-effort through the `hygieneRunFn` seam (`:19`). Both paths take `audit_log_max_bytes` and `audit_log_kept_rotations` (`internal/hygiene/settings.go:44–45`; template `workflow.yaml:332–333`; `mode: report` at line 331).
- SessionStart scope: the SessionStart record path prunes session records only (`internal/hook/session_start_record.go:115–131`); the usage-log prune is a separate hidden child verb spawned detached by the harness-observe gate (`internal/cli/hook.go:167–183`, spawn seam `:827–838`, RunE `:1322–1335`), never from the SessionStart hygiene pass.
- GC target classes: the closed class registry is `internal/hygiene/targets.go:29–36` — context-usage, goal, codex-stop-chain, codex-stop-cap, agent-stops, routing-pending, verify, todo. No class is named worktree, database, or run record, and no class reads the home directory.
- Home cleanup: `internal/cli/clean_home.go:101` (`loadHomeRetentionDays`, the home-tier `state.yaml` loader that is the precedent for the profile key); `:160` (`scanHomeCleanable`); `:243` (the compiled-constant comparison that selects `projects/` deletion candidates); `:291–293` and `:408–445` (releases: current release protected, `releaseKeep` newest non-current kept); `:310–331` (backups selected by age only); `:530` (`runCleanHome`, `--force` deletes).
- Compiled constants: `internal/config/defaults.go:456` (`DefaultReleaseKeep = 3`, non-current count) and `:460` (`DefaultProfileMaxBytes = 5 * 1024 * 1024 * 1024`). Readers of `DefaultProfileMaxBytes` (`grep -rn -E 'DefaultProfileMaxBytes' internal`): `defaults.go:460` (definition), `clean_home.go:243`, `internal/cli/doctor_disk.go:155`, `internal/cli/doctor_disk.go:261`, and the fixture `internal/cli/doctor_disk_test.go:183`. The doctor readers are the split point: card D aligns them with the config key; this SPEC removes only the `clean_home.go:243` deletion-candidate use.
- Lessons-inbox consumer: the LSEL curator's companion-offset drain reads `<state-dir>/drain-offset.json` (`.claude/skills/hns-lsel-curator/drain.sh:9` and `:48`; `SKILL.md:65`). Rotating the inbox without preserving that offset would break the drain.

## §C Requirements (GEARS)

`<subject>` is the named component in each line. Requirements are numbered REQ-CAP-NNN (10 total; within the Tier M ceiling of 16).

### §C.1 Git tracking and template

- **REQ-CAP-001 (Ubiquitous):** The template `.gitignore` shall exclude `.moai/harness/usage-log.jsonl`, `.moai/lessons-inbox.jsonl`, every `*.corrupt*` file under `.moai/`, and the monthly archives `.moai/harness/learning-history/archive/*.jsonl.gz`, so that projects created from the template never track them.
- **REQ-CAP-002 (Ubiquitous):** The repository shall stop tracking `.moai/harness/usage-log.jsonl` and `.moai/lessons-inbox.jsonl`, keeping each local file in place.
- **REQ-CAP-003 (Ubiquitous):** The repository shall stop tracking the three `.moai/harness/learning-history/archive/*.jsonl.gz.corrupt` files, keeping each local copy in place with its content unchanged.

### §C.2 Sink rotation and hygiene apply

- **REQ-CAP-004 (Event-driven):** **When** the retention prune removes events from `.moai/harness/usage-log.jsonl`, any deletion of that content shall occur only after `internal/harness/retention.go` has archived it to `archiveDir/<YYYY-MM>.jsonl.gz` (the prune archives stale events before it rewrites the log, `retention.go:416→421`). Decision Q4 (operator verdict 8a62, default) keeps `internal/harness/retention.go` as the owner of the file: the hygiene rotator does not register or rotate it, and the rotation alternate is not approved.
- **REQ-CAP-005 (Ubiquitous):** The rotator shall leave `.moai/lessons-inbox.jsonl` unrotated in this SPEC, because a rotation would break the companion-offset drain that reads `drain-offset.json`. Its existing owner, `internal/hook/inbox_lifecycle.go`, is unchanged.
- **REQ-CAP-006 (Where, decision-gated):** **Where** decision Q1 in `decision-index.md` carries the approved split verdict (operator verdict 8a62), the shipped default of `workflow.hygiene.mode` in the template `workflow.yaml` shall be `apply`, and the SessionStart auto path shall then apply only the non-destructive categories: the audit-log rotator and the GC residue classes of the closed target registry (`internal/hygiene/targets.go:29–36`; none is a worktree, a database, or a run record). The destructive categories the verdict names (worktrees, databases, run records) have no class in that registry and no deletion path at SessionStart, and the backups and releases keep-3 rule is reachable only through `moai clean --home --force`, which the SessionStart auto path does not invoke; both stay report-only at SessionStart. **Where** Q1 carries no approved verdict (no verdict, the default, or a rejection), the shipped default shall remain `report`. The `--apply` switch of `moai clean --audit-logs` is unchanged by this requirement.
- **REQ-CAP-007 (Event-driven):** When the hygiene apply path acts on any file, the hygiene apply path shall append one row with `mode` set to `apply` per applied action to `.moai/logs/hygiene-audit.jsonl`, so that the apply run is visible after the fact.

### §C.3 Archives, profile, and retention

- **REQ-CAP-008 (Event-driven):** **When** the operator runs `moai clean --home --force` (the confirmed apply run for the home cleanup), the home cleanup shall keep the current release binary, the three most recent non-current release binaries, and the three most recent `backups/removed-*` directories; it shall remove any other release binary, and any other `backups/removed-*` directory that is older than the backup retention window, only in that run. At SessionStart this category is report-only and deletes nothing (operator verdict 8a62).
- **REQ-CAP-009 (Event-driven):** **When** the SessionStart hook's deferred pass finds a per-profile data directory larger than the byte value of the `state.profile_max_bytes` key of `~/.moai/config/sections/state.yaml`, the hook shall emit a warning that names the profile and the key, and shall delete nothing (operator verdict 8a62: the cap is approved as a warning). **Where** the key is absent, no profile-cap warning shall be emitted. The home cleanup shall select no `projects/` deletion candidate by the profile cap: the cap names a warning only. The template configuration shall ship `state.profile_max_bytes: 8589934592` (8 GiB) as its default value.

### §C.4 Monthly archives

- **REQ-CAP-010 (Ubiquitous):** The repository shall stop tracking the monthly gzip archives `.moai/harness/learning-history/archive/*.jsonl.gz`, keeping each local copy in place with its content unchanged.

## §D Out of Scope

### Out of Scope — Sibling-card split (operator decision, 2026-10-10)

The operator's narrowed card-A scope moved the following v0.1.2 scope to sibling cards. Their requirements and sentences are excised from this SPEC (the v0.1.2 requirements REQ-CAP-010 to REQ-CAP-014 and criteria AC-CAP-002, AC-CAP-012 to AC-CAP-015 of that revision are no longer part of this SPEC):

- **Card B — worktree sweep:** the disposal narrowing inside `moai worktree sweep --yes` (owner/unpushed judgment), the sweep report fields (`owner`, `owner_state`, `unpushed`, `last_activity`), the advisory marking of unpushed and unowned candidates, `moai worktree list`, the bulk-run evidence rows, and the 10 GiB linked-worktrees target. This excision carries the iter3 defects D5 (the verdict-8a62-rejected narrowing sentences) and D6 (the last-activity references) out of this SPEC.
- **Card D — doctor cap read-path:** `internal/cli/doctor_disk.go:155` and `:261` reading the compiled `config.DefaultProfileMaxBytes` instead of the `state.profile_max_bytes` key, and the removal of the compiled constant (`internal/config/defaults.go:460`) once no reader remains. Until card D lands, the doctor comparisons stay as they are and this SPEC asserts nothing about their contents beyond the reader listing in §B.
- **Card t1653 — inbox rotation:** the `lessons-inbox.jsonl` rotation with the drain-offset design (REQ-CAP-005 keeps the inbox unrotated here).

### Out of Scope — Deletion during the plan phase

- This SPEC authorizes no deletion of any file during the plan phase; the untracking commits keep every local copy in place.
- The keep-3 home-cleanup deletions happen only in the operator's `moai clean --home --force` run (REQ-CAP-008), never in this SPEC's own actions.

### Out of Scope — Absorbed tree WT-v3-2-1

- The absorbed t1501 tree `WT-v3-2-1` contributes no work to this SPEC. Measured 10-10: `git rev-list --left-right --count refs/heads/main...refs/heads/WT-v3-2-1` printed `375	0`, so the branch had no commit that `main` lacked. The card's "identical to main" wording is not reproduced: the tree SHAs differ because `main` has advanced past the branch.

### Out of Scope — Retention of report content

- The content of `.moai/reports` is not pruned by age or size in this SPEC. Only the sizing measurement is recorded here.

### Out of Scope — Other SPECs' lifecycle

- No change to the lifecycle, status, or ownership of SPEC-MOAI-HYGIENE-001, SPEC-WORKTREE-SWEEP-001, or SPEC-WORKTREE-REAPER-001.

### Out of Scope — Local deletion of archive payloads

- Deleting local `.corrupt` payloads, local monthly archives, or the local usage-log is irreversible and is not authorized here; it is operator-held. The untracking commits keep every local copy.

### Out of Scope — Ownership changes and the rotation alternate

- Registering `usage-log.jsonl` in the hygiene rotator is not authorized: the Q4 verdict (8a62) keeps `internal/harness/retention.go` as the owner and the alternate is not approved.
- Changing the owner of `lessons-inbox.jsonl` (`internal/hook/inbox_lifecycle.go`) at all (REQ-CAP-005).
- Retiring the 30-day prune of `usage-log.jsonl`: the prune is the mechanism that holds the REQ-CAP-004 guarantee and the monthly archive.

### Out of Scope — Doctor warnings in this SPEC

- The `moai doctor` profile-cap warning and the compiled-constant removal are card D's (§D, sibling-card split). This SPEC's profile-cap work is the SessionStart warning surface and the `clean_home.go:243` candidate-source removal only.

## §E Assumptions

- **A1 (Tier M):** The card names exactly the spec, plan, acceptance, and progress artifacts, which is the Tier M set. The decision index is an additional artifact authored under `decision_gate: on` (A8). If the run phase exceeds 15 files, the SPEC tiers up to L. Rationale: the plan row fixes the artifact set.
- **A2 (apply rows — resolved in v0.1.1):** The apply-mode row is `AuditRow` (`internal/hygiene/auditlog.go:13–22`), with field `mode` set to `apply` (`internal/hygiene/rotate.go:167`, `:184`), appended to `.moai/logs/hygiene-audit.jsonl` (`auditlog.go:26`). Whether a fixture apply run writes those rows is still unobserved (AC-CAP-002).
- **A3 ("싱크 회전"):** Read as rotation of registered audit sinks (`sinkRegistry`). Rationale: "sink" matches `internal/hygiene/sinks.go`.
- **A4 (lessons-inbox deferred):** Rotation of lessons-inbox is deferred (REQ-CAP-005; card t1653). Rationale: the companion-offset drain reads that file by offset, and rotation would break the offset.
- **A5 (units — v0.2.0):** "5 MB" is read as 5,000,000 bytes (decimal, stricter than 5 MiB). Rationale: the stricter reading avoids a pass by unit choice. (The v0.1.2 "10 GiB" reading moved to card B with the worktree target.)
- **A6 (`.corrupt` and archive disposition):** Disposition is untrack, ignore, and keep local copies. Deletion of local payloads is operator-held. Rationale: the local copies are irreversible if deleted, and the git-tracked copies are what cost space.
- **A8 (decision index):** `decision-index.md` is authored because `decision_gate: on` (`.moai/config/sections/interview.yaml:6`). Rationale: the setting is explicit.
- **A9 (card figures):** The card's 836 MB and 53.6 GB are not adopted as measurements. Rationale: the card is untracked and not a committed authority, and the worktree figure was not re-measured on any criterion's set (the worktree criteria moved to card B).
- **A10 (releases keep — resolved in v0.1.2):** The releases rule keeps the current release, which is protected outside any count, and the `DefaultReleaseKeep` (3) newest non-current release binaries (`clean_home.go:408–445`; `config/defaults.go:456`). "The three most recent entries" in REQ-CAP-008 is read as three most recent non-current releases, so the current release is not among the three. Rationale: this is the rule the code already enforces; changing it would widen the change beyond the card's scope and would be a release-retention decision, not a capacity fix.
- **A11 (profile cap key — v0.2.0; operator verdict 8a62, 2026-10-10):** The profile cap is read from `state.profile_max_bytes` in the home-tier `~/.moai/config/sections/state.yaml`, beside `state.home_retention_days`. The verdict approves the cap as a warning only: the SessionStart hook's deferred pass names a profile that exceeds it, and nothing is deleted. An absent key produces no warning, and no compiled fallback feeds the warning (REQ-CAP-009). A present value that is not a positive integer is reported with a notice that names the key; the verdict is silent on this case, and this reading is the conservative one. The template configuration ships `8589934592` (8 GiB) as the default. The `moai doctor` read of the cap is card D's (§D); this card removes the `clean_home.go:243` compiled comparison so the cap selects no `projects/` deletion candidate. The template does ship `state.yaml` (`internal/template/templates/.moai/config/sections/state.yaml`, `state: {}` at line 13, embedded at `internal/template/embed_manifest_gen.go:615`); whether the home-tier file that `clean_home.go` reads is seeded from that template is confirmed in pre-flight. The `loadHomeRetentionDays` precedent (`clean_home.go:101`) falls back to a compiled default and is not copied. Rationale: a warning selects no file, so the compiled 5 GiB value is no longer a deletion threshold on this card's paths.
- **A12 (backups floor — v0.1.2):** A `backups/removed-*` directory is removable only when it is both outside the three most recent and older than the backup retention window (`clean_home.go:310–331`). Rationale: "retain the three most recent entries" is read as a floor. A count-only reading would delete a backup younger than the retention window once four newer ones exist, which widens deletion beyond the existing age rule.
- **A13 (home cleanup confirmation — v0.1.2):** The apply run of the home cleanup is `moai clean --home --force` (`clean.go:106`, `:107`). Rationale: `--home` uses `--force`, not `--apply`; the verdict's "apply" wording for REQ-CAP-008 and REQ-CAP-009 is mapped to `--force`.
- **A14 (archive untracking — v0.2.0):** The monthly archives are untracked with their local copies kept (REQ-CAP-010). The Q3 per-clone recipe was executed for the two state files; the archives' pull-time behavior in other clones is the same git mechanism and is covered by the same risk note (§F). Rationale: the operator's 5,000,000-byte completion criterion is arithmetically unreachable without it on the merged tree (§A v0.2.0), the archives are machine-generated retention state that grows by one file per month, and untracking keeps every local copy.

## §F Risks

- **Untracking changes other clones.** A commit that removes tracked files changes the working copy of every clone that pulls it. How git treats a locally modified copy on that pull is git's behavior and is not measured in this SPEC. The Q3 recipe measured the two state files; the nine archive files ride the same mechanism and were not part of that measurement. Mitigation: the Q3 measurement is recorded (`.moai/reports/t1631/q3-clone-measurement.md`, 2026-10-10) and the untracking is authorized by the operator's narrowed-scope decision; the landing commit B1 follows the normal card landing path, and any clone conflict is resolved by re-fetching the files from a kept local copy.
- **Residual margin under the git-tracked target.** Measured arithmetic at `e2658f3012`: tracked files other than the eleven named files (the two state files and the nine archive files) total 300,746 bytes (132,585,731 minus 132,284,985). That leaves 4,699,254 bytes of headroom under 5,000,000 bytes. Any new tracked harness file erodes that margin.
- **Apply-by-default changes the auto path for every user.** Mitigation: decided by operator verdict 8a62 (Q1 split): non-destructive categories apply, and the closed GC registry keeps the destructive categories out of the SessionStart path by construction (REQ-CAP-006).
- **Profile cap absent key.** Users without `state.profile_max_bytes` get no profile-cap warning: where the key is absent, no warning exists that could name the key (REQ-CAP-009), and existing profiles are not covered until the user sets the key or adopts a template that ships it. Mitigation: new templates ship the default (`8589934592`, 8 GiB); the operator accepts this residual as part of the warning-only design (A11). This mitigation says nothing about a notice for the absent-key case, because no notice exists there.
- **Archive untracking widens the clone-impact surface.** The Q3 measurement covered the two state files only (A14). Mitigation: the archives are append-only retention output whose authoritative copies live on the machine that produced them; a clone that loses its tracked copy on pull still holds the local file, and the corrupt-file precedent (v0.1.1 scope) covers the same class.
- **Backups retention deletes a backup younger than the window.** Floor reading (A12): removal requires both "beyond the newest three" and "older than the window".

## §G Cross-references

- `.claude/rules/moai/development/spec-frontmatter-schema.md` (12 canonical fields)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (Tier M; ceilings 16 and 16)
- `.claude/rules/moai/development/verification-completeness.md` § 2 and § 2.1 (two-cell adoption; the four-element RED cell; the undecidable disposition)
- `.claude/rules/moai/core/verification-claim-integrity.md` § 2.2 (tool-provenance) and § 2.3 (baseline-first ordering)
- `.moai/specs/SPEC-MOAI-HYGIENE-001/spec.md` § D (ownership of usage-log and lessons-inbox), § B.1 (REQ-HYG-001 bound), and HISTORY
- `.moai/specs/SPEC-WORKTREE-SWEEP-001/spec.md` (the landing predicate reused by card B's sweep work)
- `internal/hygiene/sinks.go` (`sinkRegistry`, `ownerExceptions`), `internal/hygiene/rotate.go`, `internal/hygiene/targets.go`, `internal/harness/retention.go`
- `internal/hook/session_start_hygiene.go` (the SessionStart auto path), `internal/cli/hook.go` (the detached prune child), `internal/cli/clean.go`
- `internal/cli/clean_home.go`, `internal/cli/doctor_disk.go` (card D's read path), `internal/config/defaults.go`
- `.moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md` (verification layer)
- `.moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md` (milestones and pre-flight)
- `.moai/specs/SPEC-CAPACITY-RECLAIM-001/decision-index.md` (decision gate rows)
