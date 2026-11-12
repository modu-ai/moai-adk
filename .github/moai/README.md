# MoAI RC1 GitHub assets

Issues on `modu-ai/moai-adk` are the official work ledger for `2026.11.0-rc.1` (design: SYNC and Issues, decision recorded by the user on 2026-11-08). This folder records what the RC branch adds to `.github/` and why.

## Added for RC1

| Path | Purpose | Native effect |
| --- | --- | --- |
| `ISSUE_TEMPLATE/moai-feature.yml` | New feature: one-line outcome, goal, SPEC reference, scope (allowed, forbidden, non-scope, search hints), acceptance with verification and boundary behaviour, prerequisite Issues, HOLD decisions, done definition, host scope, evidence | Issue form. Sets the label `type:feature` only; no assignees, projects or milestones |
| `ISSUE_TEMPLATE/moai-bugfix.yml` | Bug fix: one-line outcome, observed, expected with source, reproduction, regression test, scope, acceptance, prerequisites, HOLD, done definition, host scope, evidence | Issue form. Sets `type:fix` only |
| `ISSUE_TEMPLATE/moai-investigation.yml` | Investigation or design: one-line outcome, question, sources, options (filled during work), expected output, scope, acceptance, prerequisites, HOLD, done definition, evidence | Issue form. Sets `type:design` only |
| `ISSUE_TEMPLATE/moai-followup.yml` | Follow-up or repair of an existing Issue or PR: one-line outcome, original work (Issue, PR, AC, verdict link, head), remaining work, scope, acceptance, prerequisites, HOLD, done definition (#2434) | Issue form. Sets no label: the type follows the original work |
| `ISSUE_TEMPLATE/config.yml` | Template chooser: blank Issues disabled, one documentation link (#2434) | Chooser config. Writers still see "Blank issue" marked "Maintainers only" |
| `moai/labels.yml` | Label taxonomy: reused remote labels, labels to create (`added`), and previous-generation labels documented as `not_used` | none by itself; `rc-label-sync` creates missing `added` labels on a manual apply |
| `moai/README.md` | This record | none |
| `workflows/ci.yml` | Repository check set on every PR to `v2026.11.0`; reports the check `ci` (named `rc-ci` before #2563) and writes each step's exit to the job summary (00-E #2020) | Actions workflow. `contents: read` only; actions pinned to commit SHAs |
| `workflows/rc-label-sync.yml` | Compares `moai/labels.yml` with the remote labels. Dry run by default; a manual `workflow_dispatch` with `apply: true` only creates missing `added` labels, never edits or deletes (00-E #2020) | Actions workflow. Read-only except the manual apply job (`issues: write`) |
| `pull_request_template.md` | RC1 PR body contract: `Refs #<n>` (no auto-close keyword), claim token, base/candidate 40-character SHAs, commands with exits, independent review link, not_run/blocked, remaining gaps (00-E #2020; 06-C #2007 extends it) | PR template. Applied automatically only from the default branch |

The three original forms are skeletons decided with the user on 2026-11-08 (new feature, bug fix, investigation or design). Issue #2434 completed their fields and added the follow-up form and `config.yml` under user decisions 1-4 = A (Issue comment 6783612809, 2026-11-11). The fields follow the research recorded on #2434 (comment 6783409806): an agent-ready Issue has a clear outcome, complete acceptance criteria and the files to change, and a limited, self-contained scope; the forms merge fields rather than add more. A SPEC-amendment form was drafted and withdrawn on 2026-11-08; it stays a candidate until the user decides. The verified forms and labels are later ADDED to the Rust project-generation templates; this folder is never moved or removed in that step. The three forms and `labels.yml` already have byte-identical copies in `templates/github/` (08-B #2013 parity, checked by `tests/github_templates_test.py`), so #2434 updated those copies and their hashes too; the follow-up form and `config.yml` are not generated into projects.

Local verification: `PYTHONDONTWRITEBYTECODE=1 python3 -m unittest tests.github_assets_test` checks form schema shape (required keys, unique ids, allowed block types, required redaction checkbox, no assignee, project or issue-type keys, at most the one decided `type:*` label and only a label the manifest knows, no secret-shaped strings), the agent-ready fields (#2434), form defaults, the chooser `config.yml`, label-manifest consistency, the workflows (least-privilege permissions, every action pinned to a commit SHA, the `ci` check set, label-sync dry run), the PR template fields, and that no previous-generation file is recreated. Remote verification is a separate `gh label list` cross-check recorded in the Issue report; GitHub form rendering stays not_run until the forms reach the default branch.

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

Taxonomy (Issue #2434, user decisions 1-4 = A, Issue comment 6783612809, 2026-11-11). Source: `labels.yml`.

| Axis | Rule | Values |
| --- | --- | --- |
| `type:` | exactly one | `feature`, `fix`, `docs`, `chore`, `ci`, `test`, `refactor`, `design` |
| `priority:` | exactly one | `P0` (blocks the RC release), `P1` (current queue), `P2` (follow-up), `P3` (later) |
| `area:` | one or more, prefer one or two | `ci`, `cli`, `config`, `docs-site`, `hooks`, `security`, `templates`, `workflow`, `setup`, `atlas`, `design`, `runtime`, `router`, `update`, `forge` |
| `host:` | when the work must be verified on that host | `host:claude`, `host:codex` (both when both) |
| gate marker | only while a user decision is pending | `needs:decision` |

Status is not a label: it is reported in the Issue's owned report comment, because labels and assignees are not a distributed lock. `needs:decision` is not a status either: it marks that lanes must not proceed until the user decides. The authority stays the decisions log and the Issue's HOLD line; the label only makes HOLD Issues filterable. Parent tracking uses native sub-issues (no `kind:epic` label) and ordering uses native issue dependencies (no `blocked` label). The REST endpoints are used because `gh` 2.92 has no flag for either; read the link back after the first use.

Rules for whoever opens or edits an Issue:

1. The creator labels the Issue in the creating command: `gh issue create --label type:<t>,priority:P<n>,area:<a>[,host:<h>]`. Issues are created with the CLI, so nothing relies on a form label.
2. Only labels of the RC axes in `labels.yml` (`reused` and `added`) are applied. Lanes never create, edit or delete a label; a new label is first added to `labels.yml`, then the leader creates it with `rc-label-sync` (manual apply) after the PR merges.
3. `priority:*` and `needs:decision` are changed by the leader only; a lane asks in its report comment. A lane may correct a wrong `type:` or `area:` and notes it in one line of its report comment.
4. Web forms set one `type:` label (`type:feature`, `type:fix`, `type:design`; the follow-up form none). GitHub silently drops a form label that does not exist, so `type:design` must be created before the forms reach the default branch. Forms set no issue type: the organization's types (Task, Bug, Feature) cannot be set by `gh` 2.92 and would duplicate the `type:` label.

Previous-generation labels (bare `bug`, `enhancement`, `high`, `priority:high`, `status:*` and the rest listed under `not_used` in `labels.yml`) stay on the remote. They are not applied to RC Issues and are not deleted or renamed: deleting a label removes it from every Issue and PR. They are reviewed after the stable release (decision 3 = A). Remote labels in neither list (for example `question`, `good first issue`) are not applied by lanes.

## Language rule

Issue titles, bodies, SPEC text, checklists and form descriptions are written in Korean (user decision 2026-11-08). Code identifiers, commands, paths and file names stay verbatim. The three forms carry a checklist block (all items unchecked at creation) and a decisions block whose options stay `HOLD` until the user decides.

## Known limits

- The forms and `config.yml` take effect only on the default branch (`main`). They replace the previous generation's four Markdown templates on `main` together with the open-source preparation (#2435), and that `main` change is confirmed with the user right before it runs (decision 4 = A).
- GitHub reads Issue forms and the PR template from the default branch only. Until the RC branch reaches the default branch they are inert. Meanwhile Issues are opened with `gh issue create --body-file`, and PRs with `gh pr create --body-file` from a filled copy of `pull_request_template.md`. Form rendering on GitHub: not_run.
- The PR template uses the lowercase name `.github/pull_request_template.md`, the form GitHub's documentation gives. Whether GitHub also picks up an uppercase `PULL_REQUEST_TEMPLATE.md` was not verified (checked 2026-11-08).
- Actions are pinned to full commit SHAs, as the 00-E design requires. GitHub's security-hardening guide (checked 2026-11-08) now calls locking workflow dependencies with a lockfile "the strongest protection available", superseding SHA pinning for repositories that adopt it. Whether to adopt it is left to 08-D (#2015).
- Workflows are expected to differ. For `pull_request` events GitHub sets `GITHUB_REF` to the PR merge branch, so `ci` should run on PRs to `v2026.11.0` without reaching the default branch. The first observation of this is AC-00E-2 on the 00-E PR itself.
- Issue creation, label creation and milestone creation are external effects and happen only with the user's explicit instruction, which was given for `modu-ai/moai-adk` on 2026-11-08.
