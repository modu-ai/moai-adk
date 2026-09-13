## Summary

The shipped template tells users that `workflow.worktree.auto_cleanup` is a reserved, unread key. It is not: it is the toggle that gates two paths which delete worktree directories. The Go source says so explicitly, in a comment written for exactly this purpose.

Found while working card **t567** (reproducing a worktree that disappeared without a removal command). Whether this key is live is load-bearing for that investigation, and the distributed file answers it wrongly.

## The contradiction

`internal/template/templates/.moai/config/sections/workflow.yaml:50-59` — the file users receive:

```yaml
    # three auto-* toggles default false — the EnterWorktree-first policy.
    # auto_create: read once (internal/cli/worktree_advisory.go) only to select
    # advisory wording; it does not gate worktree creation.
    # auto_merge / auto_cleanup: declared but not read — no code path consumes
    # them (reserved). Kept aligned with defaults.go so the shipped file does
    # not state the opposite of the recorded policy.
    worktree:
        auto_create: false
        auto_merge: false
        auto_cleanup: false
```

`internal/config/types.go:620-631` — the recorded reader status:

```go
// Reader status (SPEC-CONFIG-KEY-HONESTY-001 M5, updated by
// SPEC-INIT-WIZARD-REPAIR-001 REQ-009): AutoCreate is read once by
// internal/cli/worktree_advisory.go only to select advisory wording — it does
// not gate worktree creation. AutoCleanup is read by the two auto-cleanup
// paths (internal/cli/session_worktree.go cleanupSessionWorktree and
// session_worktree_prmerge.go prMergeCleanup), gating worktree removal.
// AutoMerge has no production reader (declared but not read).
```

The two disagree on `auto_cleanup`. The source comment is the accurate one — both readers exist:

```
$ grep -n 'AutoCleanup' internal/cli/session_worktree.go internal/cli/session_worktree_prmerge.go
internal/cli/session_worktree_prmerge.go:143:// list` RunE, gated by the AutoCleanup toggle. NEVER returns an error and
internal/cli/session_worktree_prmerge.go:150:	if cfg == nil || !cfg.Workflow.Worktree.AutoCleanup {
internal/cli/session_worktree_prmerge.go:151:		// REQ-SW-022: reuses the session-exit AutoCleanup toggle. OFF (the
internal/cli/session_worktree.go:630:	if cfg == nil || !cfg.Workflow.Worktree.AutoCleanup {
```

Both guards sit immediately ahead of a `git worktree remove`:

- `internal/cli/session_worktree.go:624-673` — `cleanupSessionWorktree`, the session-exit removal.
- `internal/cli/session_worktree_prmerge.go:141-229` — `prMergeCleanup`, invoked at `moai session register` and `moai session list` before the subcommand's own work.

`auto_merge` is correctly described; only `auto_cleanup` is misdescribed. The grouping of the two keys into one sentence is what carries the error.

## Why it matters

The key's value decides whether two sweeps may delete worktree directories. A user reading the shipped file learns it is inert, and would reasonably flip it to `true` while reasoning about something else — or, reading it after an unexplained deletion, would rule this key out as the cause. Both readings are wrong in the direction that loses work.

The closing clause makes the defect self-describing: the comment says it is "Kept aligned with defaults.go so the shipped file does not state the opposite of the recorded policy", and it states the opposite of the recorded policy. The value alignment (`false` everywhere) is genuine; the reader-status claim beside it is not.

## Expected

The template comment describes `auto_cleanup` as the source does — a toggle read by two auto-cleanup paths, gating worktree removal — and stops grouping it with `auto_merge`, which really has no reader.

## Environment

- Read at tree base `1d150a27d`; `moai-adk v3.2.0-rc.8`.
- Source read only. No behaviour change is claimed here: the value is `false` in the template, in `internal/config/defaults.go:874`, and in this repository's own config, so nothing is currently deleting on this toggle.

## Related

- Card t567 — verdict at `.moai/reports/t567/verdict.md` (claim 7).
- The same verdict records that both gated sweeps enumerate through `git worktree list --porcelain`, so L1 `.claude/worktrees/*` trees are inside their reach, and that their dirty guard reads `git status --porcelain` only — uncommitted changes, not unmerged commits. That gap is tracked separately and is not part of this issue.

🗿 MoAI
