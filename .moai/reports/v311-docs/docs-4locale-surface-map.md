# docs 4-locale surface map (worktree release-v311)

> Captured 2026-08-20. Read-only survey; parent-persisted (surveying agent had no write tools).

## (a) README parallelism: PERFECT PARITY

All four READMEs are 778 lines, 72 headings, at byte-identical line numbers.
12 `##` sections, same order:
L40 What's New in v3.1 · L131 Why moai-adk · L222 Quick Start · L286 Core Capabilities ·
L392 How It Works · L488 Workflow Examples · L550 Configuration and Profiles ·
L594 Runs Anywhere · L651 Documentation and Learning · L713 FAQ · L737 Contribute · L766 Star History.

Zero sections present in one and missing from another.

## (b) docs-site page inventory: 4/4 PARITY, ZERO GAPS

`docs-site/content/{en,ko,ja,zh}` — 160 files each; relative-path diffs empty.

| section | md pages | `_meta.yaml` |
|---|---|---|
| (root) `_index.md` | 1 | yes |
| getting-started | 10 | yes |
| resources | 2 | yes |
| core-concepts | 11 | yes |
| workflow-commands | 8 | yes |
| utility-commands | 13 | yes |
| cli-reference | 25 | yes |
| multi-llm | 4 | yes |
| cost-optimization | 2 | no |
| guides | 4 | yes |
| worktree | 4 | yes |
| advanced | 32 | yes |
| contributing | 1 | yes |
| changelog | 1 | no |
| claude-code (+4 subdirs: foundations 9, context-memory 5, extensibility 5, agentic 10) | 30 | no |

Section weights identical across locales: getting-started 10, resources 15, core-concepts 20,
workflow-commands 30, utility-commands 40, cli-reference 45, multi-llm 60, cost-optimization 70,
guides 85, worktree 90, advanced 100, contributing 110, claude-code 130.

### Menu gaps (pages exist, no `main.yaml` entry)
- `/workflow-commands/moai-goal` — page + `new_items` entry, but no `title:` in `_meta.yaml` and no menu ref.
  Only `/utility-commands/moai-goal` is linked. Duplicate-page hazard.
- `/cli-reference/ast-grep` — absent from menu (24 of 25 linked).
- `/core-concepts/book` — absent from menu.
- All 25 `claude-code/*` leaf pages — only the 4 subsection `_index`es linked (appears intentional).
- `contributing` entry has an empty `sub:` list.

`new_items` (v3.1 NEW badges, all 4 locales): advanced → kanban-mode, bas-navigator, manager-lead,
multi-model-audit, autonomy-tier; workflow-commands → moai-goal.

## (c) Version-bearing files

### Version SSOT (already at v3.1.1)
- `pkg/version/version.go:8` — `Version = "v3.1.1"`
- `docs-site/hugo.toml:55` — `version = "v3.1.1"` (SSOT comment: update these two lines together)
- `docs-site/hugo.toml:56` — `releaseDate = "2026-08-18"`
- `.moai/config/sections/system.yaml:46,48` — `template_version` / `version: v3.1.1`

Consumers (no edit needed): `layouts/shortcodes/version.html`, `release-date.html`;
used from `content/<loc>/_index.md:9` via the version shortcode.

### READMEs — 3 hardcoded spots x 4 files (identical line numbers)
- `:24` release badge `Release-v3.1.1-blue.svg`
- `:463` statusline sample `v3.1.1` (also pins `cc v2.1.212`)
- `:722` update-available sample `v3.1.0 -> v3.1.1`

### Install scripts (2 synced copies each)
- `install.sh:113,124` / `install.ps1:185,194`
- `docs-site/static/install.sh:113,124` / `docs-site/static/install.ps1:185,194`

### docs-site content carrying exact versions
- `advanced/statusline.md:22` — statusline sample `v3.1.1`
- `advanced/manager-lead.md:19` — "renamed from manager-kanban in v3.1.1"
- `advanced/kanban-mode.md:217` — "from v3.1.1 Factory has `-f`"
- `claude-code/agentic/agent-teams.md` (en:102 / ja:106) — "MoAI-ADK v3.1.1 onwards"
- `_index.md:15` — `## New in v3.1` + `new-badge v3.1`; `:38` — "Three Core Values of MoAI 3.1"
- Series-level `v3.1`/`v3.0` prose (no bump unless minor changes): advanced/{agent-guide,
  security-notes,settings-json,ultracode-workflows,tokenomics-overview,skill-guide},
  workflow-commands/moai-plan, multi-llm/model-policy, utility-commands/moai,
  core-concepts/spec-based-dev, getting-started/{cli,faq,installation,migration},
  cli-reference/{graph,tokens,memory,profile,update}
- `changelog/_index.md` — redirect-only to GitHub Releases; carries no version.
