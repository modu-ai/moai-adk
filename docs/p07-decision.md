# P07 local TODO promotion

When a forge is connected later, preview existing local TODOs and promote only user-selected items. Review disclosure before any transmission and reconcile stable task IDs with existing remote issues to prevent duplicates. No current repository is authorized for publication by this decision.

`experiments/todo_promotion.py` provides an in-memory selection preview. Nonselected tasks never enter its output; unreviewed text is omitted; mapped tasks are marked already published. An approval label is supplied data, not a secret scanner or remote permission. Remote reconciliation, concurrent publication and actual publishing are not implemented in that experiment.

## Product path (07-B #2009)

`moai-cli todo-promote-preview PROJECT` and `moai-cli todo-promote-apply PROJECT --preview-digest D --permission-reference R` (operations `todo.promote.plan` and `todo.promote.apply`, 01-C envelope) promote selected tasks from a design-22 `tasks.md` to GitHub Issues through the 07-A adapter (`gh api`, serial client, rate-limit backoff). The request on stdin names `tasks_path` (required, no default: RC1 interim decision), `target{owner,repo}` and `selected`.

- Preview reads only: the tasks file, `<project>/.moai-cli/forge/promotions.json`, the outbox, the repository and its Issue list. It writes no file and sends no write. Each selected task yields `local_id`, `state` (`preview_only`, `blocked_review`, `already_published`, `duplicate_candidate`, or `uncertain`), `target{owner,repo,visibility}`, `redacted_title`, `redacted_body_digest`; the result carries `preview_digest`. Unselected tasks appear nowhere.
- `blocked_review` applies when a task lacks `disclosure_review: approved` or any line is credential-shaped or names a personal absolute path. `Evidence` and `External issue` lines never leave the machine.
- Apply recomputes the preview and refuses a different digest (`stale_plan`). For each `preview_only` row it saves an outbox intent (`<project>/.moai-cli/forge/outbox/<op-id>.json`), creates the Issue with the marker `<!-- moai:op promote:<project_id>:<local_id>:<attempt> -->`, and records `published` with the 07-A external reference only after a GET shows the same id, node_id, number, html_url and the marker. A full second is slept between consecutive creates. Rows in any other state are reported `not_sent`; a `duplicate_candidate` is neither linked to the similar Issue nor created (the reading implemented for REQ-PROMO-009 "display only").
- A transport error, timeout or 5xx is recorded `uncertain` (outbox `publish_unknown`) and is never created again automatically: not found is not "not created". A later run that finds the marker in the Issue list links it after a confirming GET. A 4xx is `blocked`; a confirming GET that disagrees is `conflict`. Apply stops at the first such row with `partial`.
- `promotions.json` is replaced by temporary file + rename and only if its digest still equals the one read (`conflict` otherwise; nothing is overwritten). Offline runs fail before any change, so published records stay as they were.

Limits, recorded rather than resolved:
- Duplicate detection reads the whole Issue list (`state=all`) and compares normalized titles (word Jaccard ≥ 0.8). It does not use `since`: an old Issue with the same marker or title must still be found. The search API is follow-up work (interim decision B).
- The disclosure gate is a conservative shape filter shared with the profile filter (`SECRET_SHAPES` in `src/lib.rs`) plus a few more shapes. It is not a secret scanner. A scanner such as gitleaks is HOLD; the shared redaction module belongs to 03-B #1996.
- `uncertain` is a fifth preview state beyond the four named in the Issue, so an unknown publish is visible instead of looking like `preview_only`.
- `GhTransport` bounds each `gh api` call at `GH_TIMEOUT` = 60 s (#2048, PR #2051). Past the bound `gh` is killed and its whole process group ended (unix `kill -KILL -- -<pgid>`; `gh` runs in its own group), and the call fails as `transport` — the same bucket as a timeout reported by `gh` — which apply records `uncertain`. Still open: the 60 s value is unmeasured against real calls, and the process-group teardown is written for unix (`#[cfg(unix)]`); Windows behavior is unverified.
- The permission reference is checked for presence and shape only; caller authorization remains host-owned.
- Live GitHub writes were not run (`not_run`); tests use a scripted transport and a stub `gh` that refuses writes.
