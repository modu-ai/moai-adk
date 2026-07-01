# Research — SPEC-V3R6-DOCS-V3-REBUILD-001

Three parts: (§1) the v3.0 ground-truth fact-sheet extracted from the live codebase, (§2) the confirmed drift inventory, and (§3) the Claude Code latest-feature research PLAN. Per REQ-DVR-020, §3 is a PLAN only — the fetching happens in run-phase, NOT plan-phase.

## §1 v3.0 Ground-Truth Fact-Sheet

> Authoritative source = the live codebase, NOT README (README itself carries drift). All values below were observed at plan-phase (2026-07-01) via directory listing / grep.

### 1.1 Version

- Current version: **`v3.0.0-rc4`** (binary built 2026-06-23, commit `3319defdf`).
- `docs-site/hugo.toml` currently declares `params.version = "v3.0.0-rc2"` + `releaseDate = "2026-06-03"` — the SSOT to update.
- Version renders via `{{< version >}}` shortcode bound to `params.version`.

### 1.2 `/moai` commands = 13

Observed in `.claude/commands/moai/` (13 files): `clean`, `codemaps`, `feedback`, `fix`, `gate`, `harness`, `loop`, `mx`, `plan`, `project`, `review`, `run`, `sync`.

(The raw `ls` count of 15 included `.`/`..`; the 13 above are the real command files.)

### 1.3 Retained agents = 8

7 MoAI-custom in `.claude/agents/moai/`: `manager-spec`, `manager-develop`, `manager-docs`, `manager-git`, `plan-auditor`, `sync-auditor`, `builder-harness`. Plus `Explore` (Anthropic built-in, no file). Total = **8**.

Archived (MUST NOT be presented as active): `manager-strategy`, `manager-quality`, `manager-brain`, `manager-project`, `claude-code-guide` (MoAI-custom file, not the CC built-in), `researcher`, and the 6 `expert-*` agents (backend/frontend/security/devops/performance/refactoring).

### 1.4 Retired subcommands (GONE)

`coverage`, `e2e`, `design`, `brain`, `security` — removed by SPEC-SUBCOMMAND-RETIRE-001. docs-site correctly has no pages for these; README still lists `coverage` + `e2e` (drift).

### 1.5 Skills — doc-facing = 27 `moai-*` (template source)

- Template source `internal/template/templates/.claude/skills/`: **27 `moai-*`** dirs + 1 `moai` base skill; **0 `harness-*`**.
- Local dev project `.claude/skills/`: 27 `moai-*` + **2 `harness-*`** (user-owned, local-only, NOT shipped to users).
- **Doc-facing number = 27** (derived from template source, per REQ-DVR-004). The 2 `harness-*` are excluded because they are user-owned per the harness-namespace doctrine.
- Template `moai-*` skill list: moai-domain-{backend,database,frontend,html-report,humanize}, moai-foundation-{cc,core,quality,thinking}, moai-harness-learner, moai-meta-harness, moai-ref-{api-patterns,git-workflow,llm-security,owasp-checklist,react-patterns,secops,supply-chain,testing-pyramid}, moai-workflow-{ci-loop,ddd,loop,project,spec,tdd,testing,worktree}.

### 1.6 Lifecycle = 3-phase

`plan → run → sync` (SPEC-V3R6-LIFECYCLE-REDESIGN-001). Mx Tag is a cross-cutting concern validated during sync, NOT a separate 4th phase. Docs must not reintroduce an "Mx phase".

### 1.7 Codebase scale

100K+ lines of Go across 100+ packages; 85-100% test coverage. (README already says "100K+ lines / 100+ packages"; docs-site `introduction.md` still says "34,220줄 / 32개 패키지" — drift.)

### 1.8 Content inventory

- 380 total content files = **95 per locale × 4** (`ko`/`en`/`ja`/`zh`).
- 13 content sections per locale: advanced (15), claude-code (28), contributing (1), core-concepts (7), cost-optimization (1), db (5), getting-started (10), guides (3), multi-llm (3), quality-commands (4), utility-commands (7), workflow-commands (6), worktree (4) + root `_index.md`.
- CC mirror `claude-code/` = **28 pages/locale × 4 = 112 files**. Structure: foundations (7), context-memory (5), extensibility (5), agentic (10), `_index` (1).

## §2 Confirmed Drift Inventory

| # | Surface | Observed | Correct | REQ |
|---|---------|----------|---------|-----|
| D1 | `docs-site/hugo.toml` L55 | `v3.0.0-rc2` | `v3.0.0-rc4` | REQ-DVR-001 |
| D2 | `content/ko/_index.md` L10 | "릴리스 후보 2" | rc4 | REQ-DVR-001 |
| D3 | `getting-started/introduction.md` L131 | "34,220줄 Go 코드, 32개 패키지" | "100K+ 줄, 100+ 패키지" | REQ-DVR-001, -004 |
| D4 | `introduction.md` L133/156/163 | "32개 스킬" | "27개 스킬" | REQ-DVR-004 |
| D5 | `getting-started/installation.md` L11/91/157 | `v3.0.0-rc2` examples | rc4 | REQ-DVR-001 |
| D6 | `README.md` L40/62/64 | "30 `moai-*` skills" | 27 | REQ-DVR-013 |
| D7 | `README.md` L309 | "12 `/moai` slash commands" | 13 | REQ-DVR-013 |
| D8 | `README.md` L319 | "The full command set (12 total):" lists 13 | 13 | REQ-DVR-013 |
| D9 | `README.md` L584 | `coverage` subcommand row | removed | REQ-DVR-013 |
| D10 | `README.md` L585 | `e2e` subcommand row | removed | REQ-DVR-013 |
| D11 | `advanced/builder-agents.md` | 3-builder model (builder-skill/agent/plugin) + "32개 스킬" | Harness v4 Builder | REQ-DVR-010 |
| D12 | `advanced/agent-guide.md` L124 | garbled "manager-develop" ×6 + stray archived refs; missing dynamic-team spawning | fixed + `Agent(general-purpose)` | REQ-DVR-011 |
| D13 | `workflow-commands/moai-harness.md` | v3 learning system vs v4 Builder unreconciled | one coherent model | REQ-DVR-012 |
| D14 | `claude-code/` mirror (112 files) | predates latest CC feature set | research-backed refresh | REQ-DVR-008 |
| D15 | `data/menu/main.yaml` | omits `cost-optimization`; empty `contributing` sub | IA reconciled | REQ-DVR-007, -016 |
| D16 | README.ko.md | mirror of D6-D10 (Korean): "30 skills", "12 total"(lists 13), coverage/e2e rows; plus shared `v2.6.0+` native-skills line L271 | corrected | REQ-DVR-013 |
| D17 | README.ja.md (worse — verified 2026-07-01 by direct grep) | retired `/moai design` Design System section L926-1061; `/agency → /moai design` v2.12.0 refs L57/L63; `/moai coverage` in active workflow chains L623/L633; `coverage`/`e2e` subcommand rows L563/L564; `coverage` batch-trigger L300; stale `v2.6.0+` L281 | full correction (all retired-`design`/`agency` surfaces removed; coverage/e2e removed; count/skill fixed) | REQ-DVR-013 |
| D18 | README.zh.md (worse — verified 2026-07-01 by direct grep) | same category set as D17: retired `/moai design` Design System section L926-1061; `/agency → /moai design` v2.12.0 refs L57/L63; `/moai coverage` in workflow chains L623/L633; `coverage`/`e2e` rows L563/L564; batch-trigger L300; stale `v2.6.0+` L281 | full correction | REQ-DVR-013 |

> **D17/D18 provenance (verification-claim-integrity §1.1)**: the ja/zh "worse drift" was NOT accepted on the coordinator's assertion — it was directly observed at plan-phase via `grep -niE '/moai design|/agency|/moai coverage|coverage|e2e|v2\.6\.0' README.ja.md README.zh.md`. README.md / README.ko.md were cross-checked and do NOT carry the `/moai design` section (only the shared `v2.6.0+` line), confirming the asymmetry. This is why REQ-DVR-013 now enumerates ja/zh-specific categories (AC-DVR-013c) rather than treating all 4 README files uniformly.

> Note: `README.md` L245/251/259 reference the Claude Code native `/batch` skill and a `coverage` batch-trigger row. This is distinct from the retired `/moai coverage` subcommand. Run-phase must disambiguate (EC-5) — the `/batch` native-skill mention may stay if accurate, but the trigger table must not imply a retired `/moai coverage` subcommand.

## §3 Claude Code Latest-Feature Research PLAN (run-phase)

> PLAN ONLY — per REQ-DVR-020, the fetching is deferred to run-phase (M0.3 → feeds M2). This section defines WHAT to fetch and FROM WHERE. Anti-hallucination protocol (CLAUDE.md §10): WebSearch → WebFetch URL validation → cite only verified URLs.

### 3.1 Acquisition strategy

1. **Primary**: WebFetch the canonical Claude Code docs at `code.claude.com/docs/en/*` (and `claude.com/docs`) per the per-page mapping below.
2. **Secondary**: Claude Code release notes / changelog for rc2→rc4-window feature deltas and deprecations (drives EC-3 removals).
3. **Optional (dev-only)**: the local dev-only `harness-release-update` pattern (CLAUDE.local.md §21) which already tracks CC-version deltas (`.moai/state/last-cc-version.json`, `.moai/research/cc-update-*.md`) — usable as a cross-check source, NOT a substitute for canonical docs.

### 3.2 Per-CC-page fetch mapping (28-page mirror)

| CC mirror page | Canonical CC doc surface to fetch | Refresh focus |
|----------------|-----------------------------------|---------------|
| `agentic/goal.md` | `code.claude.com/docs/en/goal` | `/goal` autonomous-continuation; evaluator model; `/goal clear` |
| `agentic/workflows.md` | `code.claude.com/docs/en/workflows` | dynamic workflows; 16 concurrent / 1000 total; `ultracode` trigger; determinism |
| `agentic/scheduled-tasks.md` | `code.claude.com/docs/en/scheduled-tasks` (verify slug) | scheduled/cron task feature |
| `agentic/agent-teams.md` | `code.claude.com/docs/en/agent-teams` | teams API (SendMessage, TaskCreate); v2.1.178 TeamCreate/TeamDelete removal; implicit teams |
| `agentic/agent-view.md` | `code.claude.com/docs/en/agent-view` (verify slug) | agent panel; v2.1.181 idle-row hide |
| `agentic/sub-agents.md` | `code.claude.com/docs/en/sub-agents` | subagent spawning; nesting depth (v2.1.172); `claude-code-guide` built-in note |
| `agentic/best-practices.md` | `claude.com/docs/en/best-practices` + `agent-teams` | 3-5 teammates guidance; when to define a custom subagent |
| `agentic/large-codebases.md` | CC large-codebase guidance | codebase-wide sweeps; Explore agent |
| `agentic/worktrees.md` | `code.claude.com/docs/en/worktrees` + `/cd` (v2.1.169) | worktree isolation; `/cd` cache-preserving cwd switch |
| `extensibility/hooks.md` | `code.claude.com/docs/en/hooks-guide` | hook events; PostToolUse/Stop; prompt/agent Stop hooks |
| `extensibility/mcp.md` | `code.claude.com/docs/en/mcp` | MCP servers; ToolSearch preload; deferred tools |
| `extensibility/plugins.md` | `code.claude.com/docs/en/plugins` | plugin system current state |
| `extensibility/skills.md` | CC skills doc | skill authoring; progressive disclosure |
| `context-memory/memory.md` | `code.claude.com/docs/en/memory` | memory files; `~/.claude/projects/{hash}/memory/` |
| `context-memory/context-window.md` | CC context-window doc | window mgmt; graduated compaction (5 layers) |
| `context-memory/checkpointing.md` | CC checkpointing doc | checkpoint/rewind |
| `context-memory/prompt-caching.md` | CC prompt-caching doc | cache behavior; `/cd` cache preservation |
| `foundations/how-claude-code-works.md` | CC overview | query loop; permission modes |
| `foundations/commands.md` | CC slash-command doc | built-in commands; `/effort`, `/goal`, `/workflows`, `/cd` |
| `foundations/tools-reference.md` | CC tools doc | tool list; Read/Write/Edit/Bash/Grep/Glob/Agent |
| `foundations/interactive-mode.md` | CC interactive doc | TUI; AskUserQuestion; preview field |
| `foundations/features-overview.md` | CC features overview | current feature index |
| `foundations/claude-directory.md` | CC `.claude/` doc | settings.json; settings.local.json |
| `foundations/_index.md` + 3 cluster `_index.md` | (section landing) | cluster intros aligned to refreshed children |
| `claude-code/_index.md` | (section landing) | CC guide intro |

> Slugs marked "verify slug" must be confirmed via WebSearch before WebFetch (some CC doc paths may differ). Any page whose canonical CC surface no longer exists → the feature is deprecated (EC-3) → remove/rework the mirror page rather than refresh it.

### 3.3 New-page research inputs (4 topics)

| New page | Research inputs (internal + external) |
|----------|---------------------------------------|
| Decision Memory (`moai preference`) | README §Decision Memory (5 components); `.claude/rules/moai/core/askuser-protocol.md` § Recommendation Placement (3-tier memory, adaptive recommendation, decay); SPEC-V3R6-ASKUSER-DECISION-MEMORY-001 |
| `moai inventory` | CLI help output (`moai inventory --help`); `cmd/moai` / `internal/cli` command definition |
| Harness v4 Builder | SPEC-V3R6-HARNESS-V4-001; `.claude/skills/moai/workflows/harness-builder.md`; manifest.json schema |
| `/effort ultracode` | `.claude/rules/moai/workflow/dynamic-workflows.md`; `code.claude.com/docs/en/workflows`; `/effort` doc |

### 3.4 Research deliverable (M0.3 output)

The research task produces a **per-page delta note** (a table: page → current-vs-latest CC delta → refresh/remove action) that becomes the M2 rewrite source of truth. Every CC-mirror rewrite (M2) must trace to a row in this note (AC-DVR-008a; verification-claim-integrity §1 — no fabricated feature claims).

## §4 Cross-References

- SPEC-SUBCOMMAND-RETIRE-001, SPEC-V3R6-LIFECYCLE-REDESIGN-001, SPEC-V3R6-HARNESS-V4-001, SPEC-V3R6-AGENT-TEAM-REBUILD-001, SPEC-V3R6-ASKUSER-DECISION-MEMORY-001.
- `.moai/docs/docs-site-i18n-rules.md` §17 (parity/URL/build doctrine).
- `.claude/rules/moai/core/moai-constitution.md` §10 (anti-hallucination + URL-verification protocol for §3 research).
