---
title: /moai gtd
weight: 29
draft: false
new: true
---

# /moai gtd

`/moai gtd` and `moai gtd` are the canonical GTD surfaces for capturing work, deciding whether it is actionable, and connecting approved work to the existing development queue. The `todo` SQLite database, card IDs, order, `queued`/`picked`/`dropped` states, and archive/restore semantics do not change. `todo` remains a compatibility name backed by the same command tree.

```bash
moai gtd add "clean up authentication error paths"
moai gtd list
moai gtd next t1 --spec SPEC-AUTH-001
moai gtd done t1 --expect "clean up"
```

Existing `moai todo ...` calls produce the same result. See the [todo compatibility reference](/en/utility-commands/moai-todo) for the complete queue verb and flag reference.

The five GTD-specific verbs carry one SQLite-backed item through the workflow.

```bash
moai gtd capture "clean up authentication error paths" --event inbox-42 --source user --sensitivity private
moai gtd clarify <gtd-id> --disposition action --outcome "merged" --evidence "tests and merge SHA" --authority "queue,commit" --trusted
moai gtd organize <gtd-id> --class action --context computer --depends-on <gtd-id>
moai gtd reflect --rebuild-projection --json
moai gtd engage <gtd-id> --approve --fresh --dependencies-ready --resources --pick --dispatch --lane lane-10 --run-id mission-42
```

`capture` requires a stable `--event` to deduplicate retries. Dispatch from `engage` requires `--pick`, `--lane`, and `--run-id`; omitting approval, freshness, dependency, or resource confirmation produces no effect. An operation receipt is prepared in SQLite first, and recovery reads the authoritative result for the same operation ID before deciding whether to retry.

The opt-in GTD v2 export/import format preserves sealed contracts, missions, events, and operation receipts alongside items and relationships, with integrity checks on import. Card archive/reopen keeps the same GTD identity, while `reflect` rereads archived completion, reopen, cancellation, and source-revision changes to surface stale evidence and blocked successors.

## The five steps

GTD is a **procedure for organizing and choosing work**, not a set of development-board columns.

| Step | Decision and stored result |
|---|---|
| Capture | Records source, sensitivity, and a deduplication identity without creating a development card. |
| Clarify | Checks the desired outcome, completion evidence, authority, and source trust. Publication waits while any of these is unresolved. |
| Organize | Records project/action/reference/deferred class, execution context, review timing, and relationships. Dependency cycles are rejected. |
| Reflect | Reconsiders changed evidence, archive/reopen/cancel events, and scheduled reviews. Cancellation never counts as prerequisite completion. |
| Engage | Recommends or links only work that passes approval, evidence freshness, dependencies, lane ownership, and resource limits. |

Development still follows `backlog → plan → run → sync → done`. Capture through Engage do not replace those stages.

## Relationships and the private graph

Only `depends_on` blocks execution. `part_of`, `supported_by`, and `related_to` are non-blocking references, while the existing meanings of `contains`, `absorbs`, `replaces`, and `conflicts` remain intact. The GTD graph is a private derivative beside the queue database. It is excluded from the repository, logs, telemetry, default export, and default backup, and is reported stale whenever its metadata or source revision differs.

## Autonomous operation and safety boundaries

The LLM and `mission-governor` produce structured proposals; they do not directly change files, Git, the queue, or dispatch state. Deterministic code checks the user-approved sealed goal, scope, allowed actions, completion evidence, resource limits, stop conditions, and current snapshot before preparing an operation. Scope expansion, stale evidence, and unauthorized actions stop as `blocked` rather than gaining implicit approval.

The current implementation supplies owner adapters for deterministic policy, recovery, dispatch, explicit-path commits, and local develop `--no-ff` merges. A commit requires a repository-local `0600` test receipt for the current HEAD; a local merge rechecks the manager-git role, `WT-*` branch, base SHA, and a `0600` lease below `.git`. Backup, restore, and export include GTD extensions only through explicit opt-in, while private projections rebuild from the SQLite revision. It does not promise operation after the session ends or completed remote pushes, pull requests, and merges until a real provider has demonstrated every required start, reconnect, replacement, credential, and process-identity capability.

Related: [`/moai goal --auto`](/en/utility-commands/moai-goal#auto-mission-mode) · [Kanban Mode](/en/advanced/kanban-mode)
