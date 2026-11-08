# MoAI RC1 GitHub assets

Issues on `modu-ai/moai-adk` are the official work ledger for `2026.11.0-rc.1` (design: SYNC and Issues, decision recorded by the user on 2026-11-08). This folder records what the RC branch adds to `.github/` and why.

## Added for RC1

| Path | Purpose | Native effect |
| --- | --- | --- |
| `ISSUE_TEMPLATE/moai-feature.yml` | New feature skeleton: goal, non-goals, SPEC reference, acceptance with real verification, write scope, dependencies, host scope, evidence | Issue form. Sets no labels, assignees, projects or milestones |
| `ISSUE_TEMPLATE/moai-bugfix.yml` | Bug fix skeleton: observed, expected with source, reproduction, regression test, scope, host scope, evidence | Issue form. Same restriction |
| `ISSUE_TEMPLATE/moai-investigation.yml` | Investigation or design skeleton: question or decision, sources, options (filled during work), expected output, evidence | Issue form. Same restriction |
| `moai/labels.yml` | Label taxonomy: reused remote labels, labels added for RC1, proposed labels not yet created | none; applied by hand with `gh label create` after preview |
| `moai/README.md` | This record | none |
| `workflows/rc-ci.yml` | Repository check set on every PR to `v2026.11.0`; reports the check `rc-ci` and writes each step's exit to the job summary (00-E #2020) | Actions workflow. `contents: read` only; actions pinned to commit SHAs |
| `workflows/rc-label-sync.yml` | Compares `moai/labels.yml` with the remote labels. Dry run by default; a manual `workflow_dispatch` with `apply: true` only creates missing `added` labels, never edits or deletes (00-E #2020) | Actions workflow. Read-only except the manual apply job (`issues: write`) |
| `pull_request_template.md` | RC1 PR body contract: `Refs #<n>` (no auto-close keyword), claim token, base/candidate 40-character SHAs, commands with exits, independent review link, not_run/blocked, remaining gaps (00-E #2020; 06-C #2007 extends it) | PR template. Applied automatically only from the default branch |

The three forms are skeletons decided with the user on 2026-11-08 (new feature, bug fix, investigation or design). Their detailed fields are settled one by one with the user; nothing beyond the skeleton is added unilaterally. A SPEC-amendment form was drafted and withdrawn the same day; it stays a candidate until the user decides. The verified forms and labels are later ADDED to the Rust project-generation templates; this folder is never moved or removed in that step.

Local verification: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.github_assets_test` checks form schema shape (required keys, unique ids, allowed block types, required redaction checkbox, no label or assignee keys, no secret-shaped strings), form defaults, label-manifest consistency, the workflows (least-privilege permissions, every action pinned to a commit SHA, the `rc-ci` check set, label-sync dry run), the PR template fields, and that no previous-generation file is recreated. Remote verification is a separate `gh label list` cross-check recorded in the Issue report; GitHub form rendering stays not_run until the forms reach the default branch.

Progress, verification results and next steps are posted as one owned report comment per Issue (template `issue-report` in the design), updated only on meaningful transitions. Tool heartbeats are never posted.

## Previous generation: removed, not adopted

The previous (Go) generation's 23 `.github` files were never tracked on this branch. On 2026-11-08 the user decided to remove them from the shared checkout; the leader session removed them (decision log 17:4x). Nothing was lost: 00-C (#1974, AC-00C-05) found all 23 byte-identical to `moai-adk-go/internal/template/templates/.github/`, which remains their source. RC1 does not recreate them (00-E REQ-00E-06); this list is the record.

Why none was adopted:

- The old `workflows/label-sync.yml` rewrote remote labels on every push to the default branch. RC1's `workflows/rc-label-sync.yml` above is a new dry-run-first workflow, not a restoration. It has a different name so that no removed path exists again.
- `agentic-factory/comment-templates/*` and `PULL_REQUEST_TEMPLATE/agentic-factory-11-handoff.md`: frequent dispatch, ACK and progress comments conflict with the "meaningful transitions only" rule. They are replaced by the single report comment and the RC1 PR template.
- `ISSUE_TEMPLATE/agentic-factory-0{1,2,3}-*.yml`: the previous generation's parent, child and SPEC-change forms. The three RC1 forms inherit their redaction rule, SPEC hash fields, acceptance and single-owner structure. They do not inherit the lane, factory, READY_ACKED stage and diagram fields.
- `labels.yml`, `agentic-factory/labels.proposed.yml`, `branch-protection.json.gtmpl`, `actions/detect-language`: previous-generation installation copies.

Removed paths:

```text
.github/actions/detect-language/action.yml
.github/agentic-factory/README.md
.github/agentic-factory/comment-templates/04-dispatch.md
.github/agentic-factory/comment-templates/05-progress.md
.github/agentic-factory/comment-templates/06-check-run.md
.github/agentic-factory/comment-templates/07-review.md
.github/agentic-factory/comment-templates/08-audit.md
.github/agentic-factory/comment-templates/09-repair.md
.github/agentic-factory/comment-templates/10-block-approval.md
.github/agentic-factory/comment-templates/12-merge.md
.github/agentic-factory/comment-templates/13-done.md
.github/agentic-factory/comment-templates/14-recovery.md
.github/agentic-factory/comment-templates/15-final-report.md
.github/agentic-factory/comment-templates/16-delta-sync.md
.github/agentic-factory/comment-templates/17-ack.md
.github/agentic-factory/labels.proposed.yml
.github/branch-protection.json.gtmpl
.github/ISSUE_TEMPLATE/agentic-factory-01-parent-issue.yml
.github/ISSUE_TEMPLATE/agentic-factory-02-child-issue.yml
.github/ISSUE_TEMPLATE/agentic-factory-03-spec-change.yml
.github/labels.yml
.github/PULL_REQUEST_TEMPLATE/agentic-factory-11-handoff.md
.github/workflows/label-sync.yml
```

## Labels and milestones

Existing remote labels (`type:*`, `priority:P0..P3`, `area:*`) are reused and never deleted or recoloured. RC1 adds `host:claude` and `host:codex` because every RC gate is verified per host, and `area:setup` for the SETUP slice. Status is reported in the Issue comment, not through `status:*` labels: labels and assignees are not a distributed lock.

## Language rule

Issue titles, bodies, SPEC text, checklists and form descriptions are written in Korean (user decision 2026-11-08). Code identifiers, commands, paths and file names stay verbatim. The three forms carry a checklist block (all items unchecked at creation) and a decisions block whose options stay `HOLD` until the user decides.

## Known limits

- GitHub reads Issue forms and the PR template from the default branch only. Until the RC branch reaches the default branch they are inert. Meanwhile Issues are opened with `gh issue create --body-file`, and PRs with `gh pr create --body-file` from a filled copy of `pull_request_template.md`. Form rendering on GitHub: not_run.
- The PR template uses the lowercase name `.github/pull_request_template.md`, the form GitHub's documentation gives. Whether GitHub also picks up an uppercase `PULL_REQUEST_TEMPLATE.md` was not verified (checked 2026-11-08).
- Actions are pinned to full commit SHAs, as the 00-E design requires. GitHub's security-hardening guide (checked 2026-11-08) now calls locking workflow dependencies with a lockfile "the strongest protection available", superseding SHA pinning for repositories that adopt it. Whether to adopt it is left to 08-D (#2015).
- Workflows are expected to differ. For `pull_request` events GitHub sets `GITHUB_REF` to the PR merge branch, so `rc-ci` should run on PRs to `v2026.11.0` without reaching the default branch. The first observation of this is AC-00E-2 on the 00-E PR itself.
- Issue creation, label creation and milestone creation are external effects and happen only with the user's explicit instruction, which was given for `modu-ai/moai-adk` on 2026-11-08.
