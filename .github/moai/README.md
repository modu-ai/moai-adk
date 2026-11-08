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

The three forms are skeletons decided with the user on 2026-11-08 (new feature, bug fix, investigation or design). Their detailed fields are settled one by one with the user; nothing beyond the skeleton is added unilaterally. A SPEC-amendment form was drafted and withdrawn the same day; it stays a candidate until the user decides. The verified forms and labels are later ADDED to the Rust project-generation templates; this folder is never moved or removed in that step.

Local verification: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.github_assets_test` checks form schema shape (required keys, unique ids, allowed block types, required redaction checkbox, no label or assignee keys, no secret-shaped strings) and label-manifest consistency. Remote verification is a separate `gh label list` cross-check recorded in the Issue report; GitHub form rendering stays not_run until the forms reach the default branch.

Progress, verification results and next steps are posted as one owned report comment per Issue (template `issue-report` in the design), updated only on meaningful transitions. Tool heartbeats are never posted.

## Not adopted in RC1

- `workflows/label-sync.yml`: changes remote labels on every push to the default branch. RC1 creates labels once, after a preview, with `gh label create`, and re-reads them. Deferred, not deleted.
- `agentic-factory/comment-templates/*` and `PULL_REQUEST_TEMPLATE/agentic-factory-11-handoff.md`: frequent dispatch, ACK and progress comments conflict with the "meaningful transitions only" rule. Replaced by the single report comment. Kept on disk as reference.
- `ISSUE_TEMPLATE/agentic-factory-0{1,2,3}-*.yml`: parent, child and SPEC-change forms of the previous generation. Their redaction rule, SPEC hash fields, acceptance and single-owner structure are inherited by the two forms above; lane, factory, READY_ACKED stage and diagram fields are not. Kept on disk, not staged.
- `labels.yml`, `labels.proposed.yml`, `branch-protection.json.gtmpl`, `actions/detect-language`: previous-generation installation copies. Unchecked against the inventory, so neither retired nor adopted.

## Labels and milestones

Existing remote labels (`type:*`, `priority:P0..P3`, `area:*`) are reused and never deleted or recoloured. RC1 adds `host:claude` and `host:codex` because every RC gate is verified per host, and `area:setup` for the SETUP slice. Status is reported in the Issue comment, not through `status:*` labels: labels and assignees are not a distributed lock.

## Language rule

Issue titles, bodies, SPEC text, checklists and form descriptions are written in Korean (user decision 2026-11-08). Code identifiers, commands, paths and file names stay verbatim. The three forms carry a checklist block (all items unchecked at creation) and a decisions block whose options stay `HOLD` until the user decides.

## Known limits

- GitHub reads Issue forms from the default branch only. Until the RC branch reaches the default branch the forms are inert; `gh issue create --body-file` is used meanwhile. Form rendering on GitHub: not_run.
- Issue creation, label creation and milestone creation are external effects and happen only with the user's explicit instruction, which was given for `modu-ai/moai-adk` on 2026-11-08.
