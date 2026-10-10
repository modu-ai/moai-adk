---
id: SPEC-CAPACITY-RECLAIM-001
title: "Capacity reclaim: untrack harness state, bound sinks and archives, report-first worktree sweep"
version: "0.1.2"
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
- **2026-10-10** — v0.1.2, plan-audit iteration-1 revision. The audit (`.moai/reports/t1631/plan-audit.md`) returned FAIL at `40cc1031c` with 14 blocking defects (D1–D14), five optional items (O1–O5), and must-pass MP-2 failed. The revision (a) re-read the rotator, the sinks, the retention prune, the sweep, the landing predicate, the card store, and the home cleanup at `40cc1031c`; (b) states the usage-log mechanism as a declared directory plus a serialized prune (§C.2, plan M1); (c) reuses the sweep's single evaluation so the owner and unpushed predicates reach both the report and `--yes` (§B.2, §C.4); (d) reconciles keep-3 with the existing "current plus three" releases rule and names the profile-cap key (§C.3, §E A10–A13); (e) pins the code measurements to the audited tree `40cc1031c`. Counts: 14 requirements (REQ-CAP-001–014) and 16 acceptance criteria (AC-CAP-001–016), against Tier M ceilings of 16 and 16. Decisions Q1–Q4 remain OPEN; no operator verdict is recorded.

## §B Problem and measured baseline

Git tracks several harness state files that grow without bound, the template `.gitignore` does not exclude them, and worktrees accumulate on disk. The card reports 836 MB for `.moai/reports` and 53.6 GB for worktrees, measured on 10-10. This SPEC does not adopt those figures. It cites only the measurements below. *(re-observed)* marks a row re-run at `40cc1031c` (the audited tree; the code is identical to `2aab5f797` except for the five SPEC files, see `git diff --stat 2aab5f797 40cc1031c`).

| Measurement | Command | Observed output |
|---|---|---|
| Git-tracked harness and lessons-inbox bytes | `git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl \| xargs -0 wc -c \| tail -1` | `128526661 total` across 191 tracked files *(re-observed at 40cc1031c)* |
| Largest tracked items | `git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl \| xargs -0 wc -c \| sort -n \| tail -6` | `.moai/harness/usage-log.jsonl` 88,629,933 B; `.moai/harness/learning-history/archive/2026-08.jsonl.gz.corrupt` 14,398,585 B; `…/2026-09.jsonl.gz.corrupt` 11,764,350 B; `.moai/lessons-inbox.jsonl` 6,043,267 B; `…/2026-07.jsonl.gz.corrupt` 4,179,609 B *(re-observed in v0.1.1)* |
| Tracked `.corrupt` files | `git ls-files -- '*.corrupt*'` | Three files, all under `.moai/harness/learning-history/archive/` (2026-07, 2026-08, 2026-09) *(re-observed at 40cc1031c)* |
| Template `.gitignore` coverage | `grep -n -E 'usage\|lessons\|corrupt' internal/template/templates/.gitignore` | No match (exit 1). The rules that do match are unrelated paths such as `.moai/logs/*` (line 253) and `.moai/reports/*` (line 268) *(re-observed at 40cc1031c)* |
| Repository `.gitignore` in this worktree | `git check-ignore -v .moai/reports/t1631/m2-baseline.md` | `.gitignore:235:.moai/reports/*`: the reports path is ignored by the repository, so a tracked baseline cannot live under `.moai/reports/` *(read at 40cc1031c)* |
| Repository rule that masks the template probe | `git -c core.excludesFile=internal/template/templates/.gitignore check-ignore --no-index -v .moai/harness/usage-log.jsonl` | `.gitignore:352:.moai/harness/usage-log.jsonl`: the repository's own rule answers first, so a probe in this worktree cannot isolate the template. AC-CAP-004 therefore uses the isolated test (plan §D.2) *(re-observed at 40cc1031c)* |
| `.moai/reports` on the primary checkout | `du -sk /Users/goos/MoAI/moai-adk-go/.moai/reports` | `857004	/Users/goos/MoAI/moai-adk-go/.moai/reports` (about 837 MiB) |
| Worktree size, two roots only (proxy) | `du -sk /Users/goos/MoAI/moai-adk-go/.moai/worktrees /Users/goos/MoAI/moai-adk-go/.claude/worktrees` | `56652256` and `35721380` KiB; sum 92,373,636 KiB (about 88.1 GiB). A foreground run exceeded the 120 s tool limit and was re-run in the background. This is a proxy: the criterion's set is every linked worktree in `git worktree list --porcelain`, which also includes `.codex/worktrees/` trees (A16). The card's 53.6 GB figure is not adopted |
| Registered worktrees | `git worktree list --porcelain` | The first record is the primary checkout (`/Users/goos/MoAI/moai-adk-go`); the remaining records include `.codex/worktrees/…` and `.moai/worktrees/…` trees *(read at 40cc1031c)* |
| `moai worktree list` subcommand | `grep -rn 'Use:.*"list"' internal/cli/worktree --include='*.go'` | No match, exit 1 *(re-observed at 40cc1031c)*. `clean.go:527` and `done.go:195` call `git worktree list --porcelain` internally; no user-facing list verb is registered |
| Hygiene entry point | `grep -n -E 'audit-logs\|session-state' internal/cli/clean.go` | `--audit-logs` (line 111) and `--session-state` (line 112) run the hygiene units; `--apply` (line 113) is the only switch that mutates, and the default is dry-run. The home cleanup is a different switch: `--home` (line 107), which deletes only with `--force` (line 106) |

Code grounding (read at `40cc1031c`; the lines were re-read for v0.1.2):

- Rotator resolution: `internal/hygiene/rotate.go:26–27` defines `LogDir` as `<projectRoot>/.moai/logs`; `rotate.go:101–102` (report), `rotate.go:160–161` and `rotate.go:186` (apply and skipped rows) resolve every registered name as `filepath.Join(r.LogDir, name)`. A registered `usage-log.jsonl` would therefore resolve to `.moai/logs/usage-log.jsonl`, which is not the file the writers use.
- Real location of the file: the writers are `internal/hook/failure_observer.go:65` and `:116` and the recorder documented at `internal/harness/observer.go:13`, all writing `.moai/harness/usage-log.jsonl` (v0.1.1 re-observation).
- Registry and owners: `internal/hygiene/sinks.go` declares the closed `sinkRegistry` (lines 26–50), which lists neither file. `ownerExceptions` (lines 89–92) names `lessons-inbox.jsonl` (`internal/hook/inbox_lifecycle.go`) and `usage-log.jsonl` (`internal/harness/retention.go`, "30-day prune to monthly gz"). The completeness guard skips owned names (`sinks.go:150–152`).
- Serialization: the rotator's pass lock is `.hygiene-rotation.lock` under `LogDir` (`internal/hygiene/rotate_lock_unix.go:25`; `rotate_lock_windows.go:26`). The 30-day prune takes an exclusive lock on `<log>.prune-state` (`internal/harness/retention.go:136` and `:175`). The two locks are different files, so the rotator and the prune do not exclude each other today.
- Displacement: `rotate.go:241–245` removes the existing chunk `<primary>.1` before the staged primary is promoted to `.1`. The removed chunk is not archived.
- Prune: `internal/harness/retention.go:400–427` rewrites the log with the kept events (`overwriteWithEvents`, `:628–671`) and archives stale events to `archiveDir/<YYYY-MM>.jsonl.gz` (`:503–526`). The prune rewrites the file, so its size can decrease at any time.
- Hygiene apply path: `internal/cli/clean.go:169–173` constructs the rotator from `workflow.hygiene` settings; `internal/hygiene/engine.go:97` constructs the same rotator for the SessionStart auto path. Both take `audit_log_max_bytes` and `audit_log_kept_rotations` (`internal/hygiene/settings.go:44–45`; template `workflow.yaml:317–320`; `mode: report` at line 318).
- Worktree sweep: `internal/cli/worktree/sweep.go:72–96` (verdict record: no owner field, no unpushed field); `sweep.go:230–232` (flags `--yes`, `--json`, `--base`); `sweep.go:379–403` (one evaluation shared by the report, `--json`, and `--yes`); `sweep.go:409–525` (predicate chain); `sweep.go:634–702` (removal loop, which removes only DISPOSE records at `sweep.go:638–640`).
- Landing predicate: `internal/cli/worktree/landing_predicate.go:423–431` (`landedBeyondAncestry`: layer 2 cumulative patch-id, layer 3 merged PR). Layer 3 requires the merged PR head to equal the local tip (`landing_predicate.go:395`), so a branch with a later unpushed commit is not landed by layer 3.
- Card store: `internal/homestate/card_record.go:98–120` (`Card`: `owner` JSON key at line 102, `state` at 103, `worktree_path` at 120); `card_record.go:32` (`done`) and `:74` (terminal states); `internal/homestate/card_worktree.go:20–46` (a card records its worktree path once, "a card never moves to another tree"); `card_worktree.go:76–79` (newest binding order); `internal/homestate/paths.go:345` (`FactoryDBPath`); `internal/homestate/factory.go:213` (`OpenFactoryReadonly`, no migration on read).
- Home cleanup: `internal/cli/clean_home.go:101–121` (`loadHomeRetentionDays`, the home-tier `state.yaml` loader that is the precedent for the profile key); `clean_home.go:160` (`scanHomeCleanable(root, retention, releaseKeep, currentVersion, now)`); `clean_home.go:191–205` (profile root and `projects/`); `clean_home.go:243` (the compiled-constant comparison); `clean_home.go:291–293` and `:408–445` (releases: current release protected, `releaseKeep` newest non-current kept); `clean_home.go:310–331` (backups selected by age only); `clean_home.go:530–585` (`runCleanHome`, `--force` deletes).
- Compiled constants: `internal/config/defaults.go:451` (`DefaultReleaseKeep = 3`, non-current count) and `:455` (`DefaultProfileMaxBytes = 5 * 1024 * 1024 * 1024`). The only reader of `DefaultProfileMaxBytes` is `clean_home.go:243` (grep, re-observed at 40cc1031c).
- Lessons-inbox consumer: the LSEL curator's companion-offset drain reads `<state-dir>/drain-offset.json` (`.claude/skills/hns-lsel-curator/drain.sh:9` and `:48`; `SKILL.md:65`). Rotating the inbox without preserving that offset would break the drain.

### §B.2 Predicate definitions (live session, owner, confirmed owner, unpushed commits, last activity)

REQ-CAP-010 to REQ-CAP-012 and the acceptance criteria use five terms. Live session is an existing verdict predicate and is unchanged. Owner, confirmed owner, unpushed commits, and last activity are report fields and advisory causes: they change no verdict and no `--yes` removal (REQ-CAP-011; operator verdict 8a62).

- **Live session.** A candidate tree is anchored by a live session when one of three predicates answers yes: (a) its git worktree lock reason names a pid that is alive, undetermined, or unreadable (`internal/session/anchor_lock.go:139–158`; an undetermined or unreadable pid counts as anchored); (b) a live session-registry entry anchors the tree (`internal/session/anchor.go:48`, `LiveAnchoredSessions`); (c) a live process has its working directory at or inside the tree (`internal/cli/worktree/sweep.go:506–517`, cause `cwd-occupied`; the platform probe is named in the file comment at `sweep.go:98–101`). The existing causes `anchored` and `cwd-occupied` (`sweep.go:55–57`) are reused unchanged. An unanswerable probe preserves.
- **Owner.** The owner of a candidate tree is the `owner` label of the newest factory card record whose `worktree_path`, compared on canonical paths, equals the tree's path. A card records its worktree path once (`card_worktree.go:38–46`). "Newest" follows the binding order `PreviousCardWorktree` uses (`card_worktree.go:76–79`). A tree with no binding has no owner.
- **Confirmed owner.** The owner is confirmed when the newest bound card record is in state `done` (`card_record.go:32`). The leader sets `done` after landing verification (`.claude/rules/local/gitflow-lane-protocol.md` §7). Any other state, and a tree with no binding, is not confirmed and is reported with advisory cause `owner-unconfirmed`. A factory store that cannot be read is reported with advisory cause `owner-undetermined`. The reported `owner_state` is the newest bound card's state, `unbound` for a tree with no binding, or `undetermined` for an unreadable store. None of these changes a verdict (REQ-CAP-011).
- **Unpushed commits.** The count of commits reachable from the branch tip and from no remote-tracking ref: `git rev-list --count <tip> --not --remotes`. A count above zero is reported with advisory cause `unpushed-commits`. A count that git cannot produce is reported as `null`. Neither changes a verdict.
- **Last activity.** The committer date of the tree's HEAD commit, in ISO-8601 (`git log -1 --format=%cI <HEAD>`). For a named branch this is the branch tip. It is a report field only.

## §C Requirements (GEARS)

`<subject>` is the named component in each line. Requirements are numbered REQ-CAP-NNN (14 total; within the Tier M ceiling of 16).

### §C.1 Git tracking and template

- **REQ-CAP-001 (Ubiquitous):** The template `.gitignore` shall exclude `.moai/harness/usage-log.jsonl`, `.moai/lessons-inbox.jsonl`, and every `*.corrupt*` file under `.moai/`, so that projects created from the template never track them.
- **REQ-CAP-002 (Ubiquitous):** The repository shall stop tracking `.moai/harness/usage-log.jsonl` and `.moai/lessons-inbox.jsonl`, keeping each local file in place.
- **REQ-CAP-003 (Ubiquitous):** The repository shall stop tracking the three `.moai/harness/learning-history/archive/*.jsonl.gz.corrupt` files, keeping each local copy in place with its content unchanged.

### §C.2 Sink rotation and hygiene apply

- **REQ-CAP-004 (Event-driven):** **When** the retention prune removes events from `.moai/harness/usage-log.jsonl`, any deletion of that content shall occur only after `internal/harness/retention.go` has archived it to `archiveDir/<YYYY-MM>.jsonl.gz` (the prune archives stale events before it rewrites the log, `retention.go:415–423`). Decision Q4 (operator verdict 8a62, default) keeps `internal/harness/retention.go` as the owner of the file: the hygiene rotator does not register or rotate it, and the rotation alternate is not approved.
- **REQ-CAP-005 (Ubiquitous):** The rotator shall leave `.moai/lessons-inbox.jsonl` unrotated in this SPEC, because a rotation would break the companion-offset drain that reads `drain-offset.json`. Its existing owner, `internal/hook/inbox_lifecycle.go`, is unchanged.
- **REQ-CAP-006 (Where, decision-gated):** **Where** decision Q1 in `decision-index.md` carries the approved split verdict (operator verdict 8a62), the shipped default of `workflow.hygiene.mode` in the template `workflow.yaml` shall be `apply`, and the SessionStart auto path shall then apply only the non-destructive categories (the audit-log rotation and the finished-session state residue classes, `internal/hygiene/targets.go:29–36`); the destructive categories (worktrees, databases, run records) and the backups and releases keep-3 rule stay report-only at SessionStart. **Where** Q1 carries no approved verdict (no verdict, the default, or a rejection), the shipped default shall remain `report`. The `--apply` switch of `moai clean --audit-logs` is unchanged by this requirement.
- **REQ-CAP-007 (Event-driven):** When the hygiene apply path acts on any file, the hygiene apply path shall append one row with `mode` set to `apply` per applied action to `.moai/logs/hygiene-audit.jsonl`, so that the apply run is visible after the fact.

### §C.3 Archives, profile, and retention

- **REQ-CAP-008 (Event-driven):** **When** the operator runs `moai clean --home --force` (the confirmed apply run for the home cleanup), the home cleanup shall keep the current release binary, the three most recent non-current release binaries, and the three most recent `backups/removed-*` directories; it shall remove any other release binary, and any other `backups/removed-*` directory that is older than the backup retention window, only in that run. At SessionStart this category is report-only and deletes nothing (operator verdict 8a62).
- **REQ-CAP-009 (Event-driven):** **When** `moai doctor` or the SessionStart hook finds a per-profile data directory larger than the byte value of the `state.profile_max_bytes` key of `~/.moai/config/sections/state.yaml`, the command or hook shall emit a warning that names the profile and the key, and shall delete nothing (operator verdict 8a62: the cap is approved as a warning). **Where** the key is absent, no profile-cap warning shall be emitted. The template configuration shall ship `state.profile_max_bytes: 8589934592` (8 GiB) as its default value, and no cap value shall be compiled into the code.

### §C.4 Worktree sweep and list

- **REQ-CAP-010 (Event-driven):** **When** `moai worktree sweep` runs in report mode (the form without `--yes`, or the form with `--json`), the sweep shall list each non-protected worktree with its path, branch, owner label, owner state, remote-landing predicate result, and unpushed-commit count, and shall dispose of no worktree.
- **REQ-CAP-011 (Event-driven):** **When** `moai worktree sweep` runs in report mode, the sweep shall mark each DISPOSE candidate that has unpushed commits or no confirmed owner (§B.2) with an advisory cause, and the marking shall change neither the verdict of any worktree nor the set of worktrees that `--yes` removes (operator verdict 8a62 approves no disposal narrowing).
- **REQ-CAP-012 (Where, decision-gated):** **Where** decision Q2 in `decision-index.md` carries an approved verdict, the `--yes` run shall write one evidence row for each disposal before that disposal removes the worktree.
- **REQ-CAP-013 (Ubiquitous):** The `moai worktree list` command shall print every registered worktree with its path, its branch, and its on-disk size in KiB, in a table form and in a `--json` form whose array holds one object per worktree with the keys `path`, `branch`, and `size_kib`, so that the array length equals the count of `worktree` lines in `git worktree list --porcelain`.
- **REQ-CAP-014 (Where, decision-gated):** **Where** decision Q2 carries an approved verdict and the bulk run has been executed, the linked worktrees (every record of `git worktree list --porcelain` except the first, which is the primary checkout) shall total less than 10 GiB (10,485,760 KiB) on disk.

## §D Out of Scope

### Out of Scope — Deletion during the plan phase

- This SPEC authorizes no deletion of any worktree and no deletion or move of any `.moai/reports` file during the plan phase.
- The bulk worktree sweep is a candidate report first; disposal is operator-gated before the run phase (REQ-CAP-012, REQ-CAP-014).

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
- Retiring the 30-day prune of `usage-log.jsonl`. Under Q4 the prune is serialized with the rotation, not retired (plan §D.1).

### Out of Scope — Archiving the displaced usage-log chunk

- Copying the displaced chunk into the monthly archive before the next rotation discards it is not added here. The Q4 row states this consequence for the operator's decision.

### Out of Scope — Bulk disposal of trees without a card binding

- This SPEC does not add a bulk disposal path for worktrees that no factory card binds. Under REQ-CAP-011 such trees are preserved by `--yes`; `moai worktree done` remains their disposal path (§E A14).

## §E Assumptions

- **A1 (Tier M):** The card names exactly the spec, plan, acceptance, and progress artifacts, which is the Tier M set. The decision index is an additional artifact authored under `decision_gate: on` (A8). If the run phase exceeds 15 files, the SPEC tiers up to L. Rationale: the plan row fixes the artifact set.
- **A2 (apply rows — resolved in v0.1.1):** The apply-mode row is `AuditRow` (`internal/hygiene/auditlog.go:13–22`), with field `mode` set to `apply` (`internal/hygiene/engine.go:43`; `internal/hygiene/rotate.go:167`, `:184`), appended to `.moai/logs/hygiene-audit.jsonl` (`auditlog.go:26`). Whether a fixture apply run writes those rows is still unobserved (AC-CAP-003).
- **A3 ("싱크 회전"):** Read as rotation of registered audit sinks (`sinkRegistry`). Rationale: "sink" matches `internal/hygiene/sinks.go`.
- **A4 (lessons-inbox deferred):** Rotation of lessons-inbox is deferred (REQ-CAP-005). Rationale: the companion-offset drain reads that file by offset, and rotation would break the offset.
- **A5 (units):** "5 MB" is read as 5,000,000 bytes (decimal, stricter than 5 MiB). "10 GiB" is 10,485,760 KiB. Rationale: the stricter reading avoids a pass by unit choice.
- **A6 (`.corrupt` disposition):** Disposition is untrack, ignore, and keep local copies. Deletion of local payloads is operator-held. Rationale: the local copies are irreversible if deleted, and the git-tracked copies are what cost space.
- **A7 (worktree set):** The worktree measurement set is every record from `git worktree list --porcelain`, covering L1 and L2 trees and the `.codex/worktrees/` trees. Rationale: the set is defined by git, not by a directory name.
- **A8 (decision index):** `decision-index.md` is authored because `decision_gate: on` (`.moai/config/sections/interview.yaml:6`). Rationale: the setting is explicit.
- **A9 (card figures):** The card's 836 MB and 53.6 GB are not adopted as measurements. Rationale: the card is untracked and not a committed authority, and the worktree figure was not re-measured on the criterion's set.
- **A10 (releases keep — resolved in v0.1.2):** The releases rule keeps the current release, which is protected outside any count, and the `DefaultReleaseKeep` (3) newest non-current release binaries (`clean_home.go:291–293`, `:441–445`; `config/defaults.go:451`). "The three most recent entries" in REQ-CAP-008 is read as three most recent non-current releases, so the current release is not among the three. Rationale: this is the rule the code already enforces; changing it would widen the change beyond the card's scope and would be a release-retention decision, not a capacity fix.
- **A11 (profile cap key — v0.1.3; operator verdict 8a62, 2026-10-10):** The profile cap is read from `state.profile_max_bytes` in the home-tier `~/.moai/config/sections/state.yaml`, beside `state.home_retention_days`. The verdict approves the cap as a warning only: `moai doctor` and the SessionStart hook name a profile that exceeds it, and delete nothing. An absent key produces no warning, and no compiled fallback is used (REQ-CAP-009). A present value that is not a positive integer is reported with a notice that names the key; the verdict is silent on this case, and this reading is the conservative one. The template configuration ships `8589934592` (8 GiB) as the default. Correction to v0.1.2: the template does ship `state.yaml` (`internal/template/templates/.moai/config/sections/state.yaml`, currently `state: {}`, embedded at `internal/template/embed_manifest_gen.go:615`); whether the home-tier file that `clean_home.go` reads is seeded from that template is confirmed in pre-flight. Rationale: a warning selects no file, so the compiled 5 GiB value is no longer a deletion threshold. The `loadHomeRetentionDays` precedent (`clean_home.go:101–121`) falls back to a compiled default and is not copied.
- **A12 (10 GiB source — v0.1.2; pending operator confirmation, v0.1.3):** No committed file contains plan row 3.2-0-2. A search of `.moai/project`, `.moai/docs`, and `.moai/specs` finds only §A of this SPEC and the AC title. The 10 GiB value is therefore carried from the operator-approved scope recorded in §A, not from a committed plan row. Rationale: the value is a measurable target. It is pending operator confirmation under decision Q2 and is not confirmed by verdict 8a62; AC-CAP-002 is bound to this pending value.
- **A13 (backups floor — v0.1.2):** A `backups/removed-*` directory is removable only when it is both outside the three most recent and older than the backup retention window (`clean_home.go:310–331`). Rationale: "retain the three most recent entries" is read as a floor. A count-only reading would delete a backup younger than the retention window once four newer ones exist, which widens deletion beyond the existing age rule.
- **A14 (owner source — v0.1.3; operator verdict 8a62):** The owner is the newest factory card binding by `worktree_path`; confirmed means state `done` (§B.2). Under verdict 8a62 the owner is a report field and an advisory cause only: `--yes` removes the same worktrees it removes today, and a tree with no binding stays disposable through `--yes` and `moai worktree done`. Rationale: the worktree packages record no owner of their own (a case-insensitive `owner` search of `internal/worktree`, `internal/cli/worktree`, and `internal/core/git` at `40cc1031c` finds no owner field), so the card binding is the only recorded owner. A disposal narrowing would need a verdict, and none was given.
- **A15 (unpushed predicate — v0.1.2):** The unpushed count is `git rev-list --count <tip> --not --remotes`, and a count above zero is an explicit PRESERVE predicate. Rationale: REQ-CAP-011 names unpushed commits. The landing predicate already preserves the measured cases (`landing_predicate.go:395`), but an explicit predicate does not depend on the landing layers for branches with unpushed commits.
- **A16 (AC-CAP-002 set — v0.1.2):** The linked worktrees are every record of `git worktree list --porcelain` except the first, which git always lists as the primary checkout. Rationale: the primary checkout is never a disposal candidate (`protectedWorktreePaths`, `internal/cli/worktree/clean.go:551`).
- **A17 (usage-log displacement — v0.1.2):** Under an approved Q4 the 30-day prune is serialized with the displacement, not retired. The displaced chunk is discarded at the next rotation without archive (`rotate.go:241–245`). Rationale: retiring the prune would also end the monthly archive, a larger retention change. The chunk discard is a consequence of rotation that Q4 must state.
- **A18 (home cleanup confirmation — v0.1.2):** The apply run of the home cleanup is `moai clean --home --force` (`clean.go:106`, `:107`, `:94`). Rationale: `--home` uses `--force`, not `--apply`; the audit's "apply" wording for REQ-CAP-008 and REQ-CAP-009 is mapped to `--force`.

## §F Risks

- **Untracking changes other clones.** A commit that removes a tracked file changes the working copy of every clone that pulls it. How git treats a locally modified copy on that pull is git's behavior and is not measured in this SPEC. Mitigation: decision Q3, with a per-clone measurement recipe in its row; the M2 commit is not pushed before the verdict.
- **Residual margin under the git-tracked target.** Measured arithmetic: tracked files other than the five named files (usage-log, lessons-inbox, and the three archives) total 3,510,917 bytes (128,526,661 minus 125,015,744). That leaves 1,489,083 bytes of headroom under 5,000,000 bytes. The residual includes the two non-corrupt archives `2026-07.jsonl.gz` (874,561 bytes) and `2026-08.jsonl.gz` (1,898,703 bytes), which this SPEC keeps tracked. Any new tracked harness file erodes that margin.
- **Bulk sweep may dispose the only copy of unpushed work.** Mitigation: REQ-CAP-011 (the owner and unpushed predicates in the shared evaluation), REQ-CAP-012 (evidence before each disposal), and decision Q2. The owner predicate also narrows any bulk run to card-bound trees whose card is `done` (A14).
- **Apply-by-default changes the auto path for every user.** Mitigation: decision Q1 is FOUNDER-owned; the template default is not flipped without the operator verdict.
- **Owner conflict with a completed SPEC.** REQ-CAP-004 would reverse SPEC-MOAI-HYGIENE-001 §D for usage-log. Mitigation: the requirement is gated by decision Q4 (FOUNDER, product-level); no registry or owner-exception change is made without a verdict.
- **Displaced chunk discard.** Under Q4, events in a displaced chunk that the 30-day prune has not archived are lost when a later rotation discards that chunk (`rotate.go:241–245`). Mitigation: the Q4 row states the consequence; the operator decides whether to accept it before M1 is applied.
- **Profile cap absent key.** Users without `state.profile_max_bytes` lose profile-cap enforcement until they set it (A11). Mitigation: the notice names the key; the template-shipping question is recorded for the operator.

## §G Cross-references

- `.claude/rules/moai/development/spec-frontmatter-schema.md` (12 canonical fields)
- `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier (Tier M; ceilings 16 and 16)
- `.claude/rules/moai/development/verification-completeness.md` § 2 and § 2.1 (two-cell adoption; the four-element RED cell; the undecidable disposition)
- `.claude/rules/moai/core/verification-claim-integrity.md` § 2.2 (tool-provenance) and § 2.3 (baseline-first ordering)
- `.moai/specs/SPEC-MOAI-HYGIENE-001/spec.md` § D (ownership of usage-log and lessons-inbox), § B.1 (REQ-HYG-001 bound), and HISTORY
- `.moai/specs/SPEC-WORKTREE-SWEEP-001/spec.md` (the landing predicate this SPEC reuses)
- `internal/hygiene/sinks.go` (`sinkRegistry`, `ownerExceptions`), `internal/hygiene/rotate.go`, `internal/harness/retention.go`
- `internal/cli/worktree/sweep.go`, `internal/cli/worktree/landing_predicate.go`, `internal/session/anchor_lock.go`
- `internal/homestate/card_record.go`, `internal/homestate/card_worktree.go`
- `internal/cli/clean_home.go`, `internal/config/defaults.go`
- `.moai/specs/SPEC-CAPACITY-RECLAIM-001/acceptance.md` (verification layer)
- `.moai/specs/SPEC-CAPACITY-RECLAIM-001/plan.md` (milestones and pre-flight)
- `.moai/specs/SPEC-CAPACITY-RECLAIM-001/decision-index.md` (decision gate rows)
