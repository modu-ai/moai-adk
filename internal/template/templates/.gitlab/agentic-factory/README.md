# Agentic Factory templates (GitLab)

Product templates for running one Issue per active agent: issue templates, a merge request template, and reusable comment sources. Nothing here applies labels, assignees, projects, milestones or any other remote state; every such step needs explicit approval.

## Where the files live

- `.gitlab/issue_templates/`: Markdown description templates for the parent Issue, child Issue and SPEC change.
- `.gitlab/merge_request_templates/`: the agent handoff template, selected by name. It is not the default template of the project.
- `.gitlab/agentic-factory/comment-templates/`: the comment templates. No repository file is read for comment templates on this provider, so these are reusable sources to copy into a note.

## The 17 templates

| # | Template | File | Native home |
|---|---|---|---|
| 01-parent-issue | Parent Issue | `agentic-factory-01-parent-issue.md` | Issue description template (Markdown) |
| 02-child-issue | Child Issue | `agentic-factory-02-child-issue.md` | Issue description template (Markdown) |
| 03-spec-change | SPEC Change | `agentic-factory-03-spec-change.md` | Issue description template (Markdown) |
| 04-dispatch | Dispatch | `04-dispatch.md` | none, no native home (comment; reusable source) |
| 05-progress | Progress | `05-progress.md` | none, no native home (comment; reusable source) |
| 06-check-run | Check Run | `06-check-run.md` | none, no native home (comment; reusable source) |
| 07-review | Code Review | `07-review.md` | none, no native home (comment; reusable source) |
| 08-audit | Independent Audit | `08-audit.md` | none, no native home (comment; reusable source) |
| 09-repair | Repair | `09-repair.md` | none, no native home (comment; reusable source) |
| 10-block-approval | Block and Approval Request | `10-block-approval.md` | none, no native home (comment; reusable source) |
| 11-handoff | Agent Handoff | `agentic-factory-11-handoff.md` | Merge request template |
| 12-merge | Merge | `12-merge.md` | none, no native home (comment; reusable source) |
| 13-done | Done | `13-done.md` | none, no native home (comment; reusable source) |
| 14-recovery | Resume and Regression | `14-recovery.md` | none, no native home (comment; reusable source) |
| 15-final-report | Final Report | `15-final-report.md` | none, no native home (comment; reusable source) |
| 16-delta-sync | Incremental Sync | `16-delta-sync.md` | none, no native home (comment; reusable source) |
| 17-ack | Receipt ACK | `17-ack.md` | none, no native home (comment; reusable source) |

Every file carries a machine-metadata marker. The marker is correlation only and is never authorization.

## One Issue, multiple tasks

One Issue is handled by one active agent. That agent owns the SPEC and the plan, run and sync phases and every internal task. Tasks run in dependency order; only independent scopes run in parallel. A helper execution is not a separate Issue owner and never creates a second writer on the same branch. The sync phase hands the changes to a pull request; it does not replace the safe local sync after merge or the closing of the Issue by the leader.

## Preparation stages

Intake and preparation run in the background so the leader stays responsive: ACK, then Issue published, then branch and worktree READY, then start.

`ACCEPTED -> ISSUE_CONFIRMED -> BRANCH_REGISTERED -> AGENT_RESERVED -> WT_PREPARED -> READY_ACKED`

Rule: no development launch before READY_ACKED. The preparation step writes no product code or SPEC, and its write access never overlaps the development agent's.

## Decisions recorded here

- Descriptions carry no quick actions. A quick action in a template would be executed by the provider when the Issue or merge request is created.
- The merge request template is not named Default, because a template with that name becomes the project default.
- labels: no GitLab labels manifest ships. Label support on this provider is not asserted; see the support matrix.

## Diagrams and Mermaid

Six templates (parent Issue, child Issue, SPEC change, repair, resume and regression, final report) carry a Mermaid source block, top-down only. A platform-rendered Mermaid fence is not the pinned Pele renderer's output, and no Pele SVG has been generated: where one is required, the template shows the line `Pele SVG: not generated (renderer unavailable) — see GAP-PELE-SVG`. See GAP-PELE-SVG and GAP-MERMAID-RENDER in the support matrix.

## Support matrix and known gaps

The provider support matrix and the definitions of every known gap are in `.moai/docs/agentic-factory-provider-support.md`. No equivalence between GitLab and the other provider is assumed anywhere in these files.
