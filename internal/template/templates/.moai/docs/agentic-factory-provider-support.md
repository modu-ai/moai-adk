# Agentic Factory: provider support matrix and known gaps

This file states, per concept, what GitHub and GitLab each offer for the Agentic Factory templates. No equivalence between the providers is assumed: a concept that exists on one provider is not presumed to exist, or to behave the same way, on the other. A cell marked unverified means the behaviour was not confirmed and these templates make no claim about it.

Status words: supported, supported differently, not supported, unverified.

| Concept | GitHub | GitLab | Used by these templates |
|---|---|---|---|
| Milestone | supported — repository scope; a release or acceptance goal, not an organisation-wide milestone | unverified — not asserted; these templates use no GitLab milestone feature | no |
| Epic | not supported — no native Epic object is assumed; the Epic is the parent Issue, optionally marked by an organisation Issue Type or an approved label | supported differently — group-level work items on the Premium and Ultimate tiers (provider documentation); not used here, the templates link the parent Issue in the body | no |
| Sub-issue | supported — native sub-issues from triage access up, at most 100 children per parent and 8 levels (about 3 recommended); links across repositories need permission and visibility on both sides | unverified — not asserted; the templates link parent and child as plain text | no |
| Issue dependencies | supported — available on Free, Pro, Team and Enterprise Cloud; a relation is a display, not a runtime gate | unverified — not asserted | no |
| Issue types | supported differently — organisation scope, edited by the organisation owner, at most 25; the defaults may have been changed | unverified — not asserted | no |
| Plan-tier gating | supported differently — availability varies by plan and deployment, so capability is inventoried first and never assumed | supported differently — epics need the Premium or Ultimate tier; description templates are available on every tier (provider documentation) | no |
| Labels | supported — repository scope; creating, editing and deleting need write access, applying and removing need triage | unverified — not asserted; no GitLab labels manifest ships | no |
| Projects | supported differently — user or organisation scope, not owned by the repository, with its own permission | unverified — not asserted | no |
| Issue forms | supported — YAML forms in the issue template folder | not supported — description templates are Markdown files only (provider documentation) | yes |
| Issue templates (Markdown) | supported | supported — files in the issue templates folder on the default branch | yes |
| Pull request / merge request templates | supported — a template folder selected by name | supported — files in the merge request templates folder; a template named Default becomes the project default, so none ships under that name | yes |
| Comment templates as repository files | not supported — no repository-file surface for comments is documented in the pages read | not supported — the description template pages read cover descriptions only | no |
| Light/dark image selection | supported — a picture element with a prefers-color-scheme source (provider documentation) | unverified — not asserted | no |
| Mermaid source rendering | supported — fenced blocks render natively; this is not the pinned Pele output | unverified — not asserted | yes |

## Known gaps

- GAP-PELE-SVG (unmet): the Pele renderer was unavailable where these templates were authored, so no SVG exists. Six templates require one; each shows the line `Pele SVG: not generated (renderer unavailable) — see GAP-PELE-SVG` in place of the image, and no image asset is shipped. The gap closes only when a pinned renderer produces light and dark SVGs that pass sanitising and contrast checks.
- GAP-MERMAID-RENDER: Mermaid sources were checked structurally only (top-down header, statement grammar, labelled nodes). No Mermaid renderer was available, so rendering is not proven.
- GAP-LOCALE-VARIANTS: the templates are English only. Locale variants are not provided.
- GAP-REMOTE-SETUP: nothing was applied to any repository. Labels, milestones, projects, Issue types and webhooks are untouched; the proposed labels sit in an inert manifest.
- GAP-PROVIDER-GATED-DEPLOY: the deployer copies every template to every project, so both the GitHub and the GitLab trees are installed regardless of the host in use. A provider selection that gates deployment is a draft and is not wired.

## Sources

- https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-issue-forms
- https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/syntax-for-githubs-form-schema
- https://docs.github.com/en/communities/using-templates-to-encourage-useful-issues-and-pull-requests/creating-a-pull-request-template-for-your-repository
- https://docs.gitlab.com/user/project/description_templates/
- https://docs.gitlab.com/user/group/epics/
