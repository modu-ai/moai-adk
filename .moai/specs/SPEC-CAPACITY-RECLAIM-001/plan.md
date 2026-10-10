# SPEC-CAPACITY-RECLAIM-001 — Implementation Plan

## §A Context

Card t1631 (plan phase) at worktree `WT-capacity-reclaim`, HEAD `2aab5f797`. Tier M: 13 requirements and 15 acceptance criteria, each counted independently against the Tier M ceilings of 16. The requirements and measured baseline are in `spec.md` §B. Milestones are ordered by decision reversibility: decision-gated items first, then reversible tracking changes, then destructive actions last. No time estimates are given; order is by phase.

## §B Milestones (priority-ordered)

### M1 — Priority High: Sink registration and hygiene apply path (decision-gated)

- `usage-log.jsonl` is an owner exception, not a registered sink (`internal/hygiene/sinks.go` `ownerExceptions`, line 91; owner `internal/harness/retention.go`). Registering it moves the file to the rotator and removes that exception. Do this only when decision Q4 carries an operator verdict (REQ-CAP-004; AC-CAP-007). Otherwise the registry and the exception stay as they are.
- Confirm that the apply path appends one `mode: apply` row per action (REQ-CAP-007). The fields are pinned in `spec.md` §E A2 (`internal/hygiene/auditlog.go:13–22`; `internal/hygiene/rotate.go:167`, `:184`).
- The apply-by-default flip (REQ-CAP-006) changes the template default of `workflow.hygiene.mode` (`internal/template/templates/.moai/config/sections/workflow.yaml:318`) only if decision Q1 carries an operator verdict. Otherwise `report` stays the default. The CLI `--apply` switch is not changed.
- Lessons-inbox rotation is not implemented here (REQ-CAP-005). Its owner, `internal/hook/inbox_lifecycle.go`, is unchanged.

### M2 — Priority High: Tracked-state untracking and template ignore

- Add ignore rules for `.moai/harness/usage-log.jsonl`, `.moai/lessons-inbox.jsonl`, and `*.corrupt*` to `internal/template/templates/.gitignore` (REQ-CAP-001). Template-First: edit the template, then rebuild and sync.
- Untrack the two state files and the three archive `.corrupt` files with `git rm --cached` (REQ-CAP-002, REQ-CAP-003). Local copies stay in place. Record each file's byte count immediately before the change (AC-CAP-006).
- Gate: decision Q3 must carry an operator verdict before this milestone's commit is pushed.

### M3 — Priority High: Worktree candidate report and `moai worktree list`

- Extend the existing default report of `moai worktree sweep` (the form without `--yes`, and `--json`) with an owner field and an unpushed-commit count (REQ-CAP-010). Add an owner predicate: an unconfirmed owner is PRESERVE (REQ-CAP-011). The unpushed branch is already PRESERVE through the landing predicate, which is reused unchanged from SPEC-WORKTREE-SWEEP-001.
- Add `moai worktree list` with table and `--json` output (REQ-CAP-013). No `list` verb is registered today (`spec.md` §B).

### M4 — Priority Medium: Operator-gated bulk sweep (run phase only)

- Executed only after the candidate report from M3 is reviewed and decision Q2 carries an operator verdict (REQ-CAP-012).
- Add the owner-confirmation gate before `--yes` disposes anything. The current `--yes` path has no owner check (`internal/cli/worktree/sweep.go:230`).
- Each disposal records its evidence first. Unpushed branches are never disposed of (REQ-CAP-011).

### M5 — Priority Medium: Backups, releases keep-3, profile cap

- Existing code, `internal/cli/clean_home.go`: releases keep the current release plus the newest `config.DefaultReleaseKeep` (3) others (lines 291–293, 441–445); `backups/removed-*` is retained by age (lines 310–329); the profile cap is the compiled `config.DefaultProfileMaxBytes` (line 243). The location of each rule is resolved (v0.1.1).
- Change backups to keep-3 by count, and read the profile cap from configuration (REQ-CAP-008, REQ-CAP-009). Apply both under the operator confirmation of the apply run.
- Pre-flight: pin the configuration key for the cap and the confirmation surface for the apply run, and decide whether the current release counts toward the three (`spec.md` §E A10).

### M6 — Priority Low: Evidence recording and sync

- Record every AC command and its verbatim output in `progress.md` §E.2 during run; sync records §E.4.

## §C Pre-flight commands (run phase)

```bash
git ls-files -z -- .moai/harness .moai/lessons-inbox.jsonl | xargs -0 wc -c | tail -1
git ls-files -- '*.corrupt*' | wc -l
git worktree list --porcelain | grep -c '^worktree '
grep -rn 'Use:.*"list"' internal/cli/worktree --include='*.go'
grep -n -E 'usage|lessons|corrupt' internal/template/templates/.gitignore
grep -n -i -E 'owner|unpushed' internal/cli/worktree/sweep.go
grep -n -E 'usage-log|lessons-inbox' internal/hygiene/sinks.go
```

Each command is run in a single invocation, and its verbatim output is recorded.

## §D Technical approach

- Sink rotation reuses the SPEC-MOAI-HYGIENE-001 rotator and its bound; no new rotation algorithm. It applies only to registered sinks, so `usage-log.jsonl` joins it only under Q4.
- Untracking uses `git rm --cached`, so local files remain. The template ignore rule keeps new projects from tracking them again.
- The sweep reuses the SPEC-WORKTREE-SWEEP-001 predicate chain. The default report (the preview) is the report mode; disposal requires `--yes` and the M4 gate.
- Keep-3 and the profile cap extend the existing home-cleanup rules in `internal/cli/clean_home.go` rather than adding a second cleanup path.

## §E Risks and mitigations

| Risk | Mitigation |
|---|---|
| Untracking deletes working copies in clones that pull | Decision Q3; per-clone backup before pull |
| Git-tracked total margin is about 1.49 MB | Track the total in every run; new tracked harness files require review |
| Bulk sweep disposes unpushed or unowned work | REQ-CAP-011; owner-confirmation gate (M4); decision Q2 |
| Apply-by-default changes every user's auto-path behavior | Decision Q1 FOUNDER; no default flip without verdict |
| Lessons-inbox offset breaks under rotation | REQ-CAP-005 excludes it |
| Reversing a completed SPEC's ownership (HYGIENE-001 §D) for usage-log | Decision Q4 FOUNDER; no registry or owner-exception change without verdict |

## §F Exclusions

See `spec.md` §D (Out of Scope).
