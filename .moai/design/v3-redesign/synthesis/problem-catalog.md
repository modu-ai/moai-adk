# MoAI-ADK Current State Problem Catalog

> Wave 2 synthesis artifact 2 of 3
> Source: Convergence of R3 (CC architecture debt applied to moai) ∪ R4 (skill audit) ∪ R5 (agent audit) ∪ R6 (commands/hooks/rules/config audit)
> Date: 2026-04-23
> Scope: Concrete, v2.13.2-state problems that v3 must address

---

## Executive summary

- **Total problems cataloged**: 72
- **By severity**:
  - CRITICAL: 8 (architectural invariant violations, user-facing bugs, safety gaps)
  - HIGH: 21 (systemic issues, dead configuration, protocol violations)
  - MEDIUM: 29 (quality, maintainability, scope bloat)
  - LOW: 14 (cosmetic, naming, documentation)
- **By domain**:
  - Agents (A-series): 23
  - Skills (S-series): 19
  - Hooks (H-series): 14
  - Rules (R-series): 6
  - Config (C-series): 6
  - Commands (X-series): 2
  - Architecture (Z-series applied to moai): 2

Problems are grouped into **6 clusters** (§Problem clusters) so that v3 SPECs can address structural themes rather than individual defects.

---

## Severity definitions

| Severity | Bar |
|---|---|
| **CRITICAL** | Violates a HARD constitutional rule (CLAUDE.md, moai-constitution.md), causes a user-visible bug (e.g., tmux pane leak), or exposes safety (e.g., hardcoded absolute path, missing sandbox). Must fix in first v3 wave. |
| **HIGH** | Systemic issue affecting ≥3 files/agents/skills, dead configuration that misleads maintainers, or architectural inconsistency that blocks v3 clean redesign. Must fix in first two v3 waves. |
| **MEDIUM** | Scope creep, maintainability debt, or structural duplication that increases comprehension cost but does not break anything. Can defer to third wave. |
| **LOW** | Cosmetic, naming, documentation drift — scheduled for cleanup but non-blocking. |

---

## Problem catalog table

| # | Problem | Domain | Severity | Source | Evidence | v3 Resolution Strategy |
|---|---------|--------|----------|--------|----------|------------------------|
| P-A01 | 9 agents embed literal `AskUserQuestion` instructions in body, violating agent-common-protocol.md [HARD] rule "Subagents MUST NOT prompt the user" | agents | **CRITICAL** | R5 §Common Protocol compliance table | manager-spec L134, manager-strategy L58-72, manager-project L92/L99/L131-134, expert-backend L66, expert-frontend L66/L94, expert-devops L63, builder-agent L61, builder-skill L61, builder-plugin L82 | Scrub all `AskUserQuestion` occurrences from agent bodies; replace with "return blocker report" pattern; add CI lint that rejects the literal string in agent files |
| P-A02 | 19 of 22 agents missing `effort` field in frontmatter — Opus 4.7 defaults apply silently, under-invoking reasoning for implementation-heavy agents | agents | **HIGH** | R5 §Effort-level calibration matrix | 19 of 22 agents lack explicit effort declaration | Publish effort matrix in agent-authoring.md as Required Field; populate field on every agent |
| P-A03 | 3 agents have explicit `effort` wrong per constitution: expert-security (high→xhigh), evaluator-active (high→xhigh), plan-auditor (high→xhigh) | agents | HIGH | R5 §Effort drift summary | Constitution Opus 4.7 section names these three for xhigh | Upgrade effort level in each agent's frontmatter |
| P-A04 | 4 agents declare `Agent` tool but subagents cannot spawn — dead configuration | agents | HIGH | R5 §Tool scope audit | expert-security, builder-agent, builder-skill, builder-plugin | Remove `Agent` from tools list; document in agent-authoring.md that subagents must return blockers |
| P-A05 | 3 builder agents (builder-agent, builder-skill, builder-plugin) have near-identical bodies — frontmatter model/permissionMode/tools/memory all match, same 5-phase workflow | agents | HIGH | R5 §Detailed findings → builder MERGE | Three files identical except artifact template | Merge into single `builder-platform` agent with `artifact: agent|skill|plugin|command|hook|mcp|lsp` parameter |
| P-A06 | expert-debug is 100-LOC read-only router — no Write tools, 70% body is delegation table to manager-ddd/manager-quality/expert-refactoring | agents | HIGH | R5 §expert-debug findings | Lines 89-93 delegation table duplicates manager-quality routing | Fold into manager-quality as diagnostic sub-mode; retire as standalone agent |
| P-A07 | expert-testing is strategy-doc writer with every actual testing activity delegated to manager-ddd/expert-performance/expert-security | agents | HIGH | R5 §expert-testing findings | "OUT OF SCOPE: Unit test implementation (manager-ddd)"; Delegation Protocol hands off all real work | Fold strategy role into manager-cycle (RED planning); load-test role into expert-performance; retire as standalone |
| P-A08 | expert-performance is pure advisor with contradictory tool config — no Write/Edit tools but workflow step 5 says "Create .moai/docs/performance-analysis-{SPEC-ID}.md" | agents | HIGH | R5 §expert-performance findings | Tools list excludes Write; body L58-60 instructs Write action | Grant `Write` scoped to `.moai/docs/` OR remove the instruction; absorb into expert-backend with `--deepthink performance` mode |
| P-A09 | manager-ddd vs manager-tdd have ≈60% body overlap (LSP baseline, checkpoint, @MX, decision guide, steps 1-5) | agents | HIGH | R5 §Role taxonomy audit | Real difference: 3 phase names (ANALYZE-PRESERVE-IMPROVE vs RED-GREEN-REFACTOR) + 1 decision heuristic | Merge into `manager-cycle` with `cycle_type: ddd|tdd` parameter |
| P-A10 | manager-project runs 6 modes, 3 of which belong to CLI not to an agent (settings_modification, glm_configuration, template_update_optimization) | agents | HIGH | R5 §manager-project findings | Lines 131-134 include "AskUserQuestion" normal workflow while lines 31-35 self-correctly say "CANNOT use AskUserQuestion" | Scope agent to `.moai/project/{product,structure,tech}.md` only; move settings/glm/template modes to `moai` CLI binary |
| P-A11 | 6 cross-file-write agents missing `isolation: worktree` per worktree-integration.md HARD rule | agents | HIGH | R5 §Worktree correctness | manager-ddd, manager-tdd, expert-backend, expert-frontend, expert-refactoring, researcher | Add `isolation: worktree` to frontmatter; upgrade rule from SHOULD to MUST for implementation agents touching ≥3 files |
| P-A12 | manager-project scope overreach — Line 108 says ".moai/project/ only" but body covers .moai/config/config.yaml reading and GLM configuration | agents | MEDIUM | R5 §manager-project findings | Scope boundary inconsistent with own frontmatter | Shrink scope in both frontmatter and body; move out-of-scope modes to CLI |
| P-A13 | manager-quality and evaluator-active duplicate identical "Skeptical Evaluator Mandate" HARD block (6 bullets) | agents | MEDIUM | R5 §Role taxonomy | Same block appears verbatim in both files | Extract to agent-common-protocol.md §"Skeptical Evaluation Stance"; reference from both agents |
| P-A14 | plan-auditor has no `memory:` field declared (inconsistency with 21/22 agents) — yet audits across SPECs make it most memory-worthy | agents | MEDIUM | R5 §Additional systemic findings | plan-auditor frontmatter missing memory field | Add `memory: project` (or `user` if cross-project audits needed) |
| P-A15 | `mcp__context7__*` over-included in 16/22 agents including ones that never look up library docs (manager-git, manager-quality) | agents | MEDIUM | R5 §Additional systemic findings | Context7 is read-only library docs; many agents don't use it | Audit per-agent MCP tool list; remove Context7 where unused |
| P-A16 | manager-git trigger set has 14 items including "checkout", "rebase", "stash" — too granular | agents | LOW | R5 §Language-trigger coverage minor issues | Matches on any git-adjacent mention, causing false-positive routing | Reduce to ~8 high-precision keywords |
| P-A17 | expert-backend has 24 EN trigger tokens with duplicates (e.g., "Oracle" appears twice) | agents | LOW | R5 §Language-trigger coverage minor issues | Over-enumeration dilutes weighted routing signal | Deduplicate, keep 12-15 | 
| P-A18 | Hook action matrix has dead config — expert-debug has PostToolUse hook matching Write|Edit but no Write/Edit in tools list | agents | MEDIUM | R5 §Additional systemic findings | Hook never fires | Remove dead hook config OR grant Write/Edit if the feature is real |
| P-A19 | Every agent repeats boilerplate `--deepthink flag: Activate Sequential Thinking MCP` line 22 times | agents | LOW | R5 §Additional systemic findings | 22 identical description fragments | Move to agent-common-protocol.md as global behavior reference |
| P-A20 | expert-frontend mixes React/Vue code implementation with Pencil MCP design (17 MCP tools, largest footprint in repo) | agents | MEDIUM | R5 §Expert agents taxonomy, §Tool scope audit | Phase 2/3 is code; entire Pencil MCP section is design | Split Pencil scope to new `expert-designer` OR back into `moai-domain-brand-design` skill |
| P-A21 | No `manager-design` agent despite design constitution v3.3.0 describing a design pipeline | agents | MEDIUM | R5 §Role taxonomy, missing role | /moai design routes directly to domain skills with no coordinating agent | Add `manager-design` if v3 design flow is first-class; else route through manager-strategy |
| P-A22 | researcher agent has `acceptEdits` but missing `isolation: worktree` — own L49 says "All experiments in worktree isolation when possible" | agents | HIGH | R5 §researcher findings, §Worktree correctness | Self-contradiction between body rule and frontmatter | Add `isolation: worktree` or retire researcher and fold into moai-workflow-research skill runbook |
| P-A23 | Skills injection parity inconsistent — most agents preload moai-foundation-core, but manager-project, plan-auditor, researcher do not | agents | LOW | R5 §Additional systemic findings | Skill loading list drift | Standardize preloads via agent-authoring.md required preload table |
| P-S01 | Thinking triplet 3-way overlap: moai-foundation-thinking + moai-foundation-philosopher + moai-workflow-thinking cover same "think before acting" guidance (~33KB total, ~3 callers) | skills | HIGH | R4 §MERGE clusters Cluster 1 | 3 skills, 33KB combined, overlapping scope | Merge into single `moai-foundation-thinking` skill (~12KB target) retaining all trigger keywords |
| P-S02 | Platform skills implement "3 vendors in one" anti-pattern — high maintenance when any vendor API changes | skills | HIGH | R4 §Summary of architectural health issues item 3 | moai-platform-auth (Auth0+Clerk+Firebase), moai-platform-deployment (Vercel+Railway+Convex), moai-platform-database-cloud (Neon+Supabase+Firestore) | Split per vendor OR narrow to dominant vendor with explicit "NOT for other vendors" guardrails |
| P-S03 | Domain skills are kitchen sinks: 22-keyword lists in domain-backend / domain-frontend / domain-database | skills | HIGH | R4 §Summary of architectural health issues item 2 | "backend" trigger catches everything; same for frontend and database | Shrink to decision-matrix router with explicit `NOT for:` contract |
| P-S04 | `moai` root skill is 18KB + 20 bundled workflow files totaling ~300KB — effectively a monorepo command manual | skills | HIGH | R4 §Summary of architectural health issues item 6 | Bundled workflows are the source of truth for /moai subcommands | Extract each workflow to first-class `moai-cmd-{name}` skill; thin the router |
| P-S05 | 4 skills exceed Level 2 token budget (5000 tokens ≈ 20KB): workflow-testing 22.5KB, chrome-extension 21KB, spec 18.7KB, moai 18.2KB | skills | MEDIUM | R4 §Progressive Disclosure compliance | Level 2 budget violated | Modularize into level-3 resources; split bundled content |
| P-S06 | 43 bundled files in moai-workflow-testing (over-modularization) — comprehension cost > access convenience | skills | MEDIUM | R4 §Progressive Disclosure § | Largest single level-3 payload after `moai` skill | Consolidate to ~10 top-level modules; archive unused deep modules |
| P-S07 | Directory naming inconsistency `reference/` vs `references/` across skills — both forms appear in the same repo | skills | LOW | R4 §Progressive Disclosure § structural issue | e.g., moai-foundation-cc has both reference/ and references/ | Pick one convention; add lint check that rejects the singular form |
| P-S08 | `moai-lang-*` skills referenced by related-skills fields but **do not exist** — 16 language rules live at `.claude/rules/moai/languages/` | skills | HIGH | R4 §Category coverage gaps | Blurs skill/rule boundary | Decide one home: either migrate rules to skills OR remove stale skill references |
| P-S09 | Agent-extending reference skills (moai-ref-*) have zero static references — activation model opaque, relies purely on keyword matching | skills | MEDIUM | R4 §Per-skill audit table note | 5 `moai-ref-*` skills show zero grep hits | Document the keyword-matching activation model; add telemetry for actual activation rates |
| P-S10 | moai-tool-svg is niche with zero references and no evidence of active consumers | skills | MEDIUM | R4 §RETIRE candidates | Score 20, no active invoker found | RETIRE pending 30-day telemetry check OR keep with explicit justification |
| P-S11 | moai-docs-generation is superseded by moai-workflow-jit-docs + moai-library-nextra | skills | MEDIUM | R4 §Merge Cluster 4 | Scope entirely absorbed | RETIRE or fold into moai-workflow-jit-docs |
| P-S12 | moai-design-tools has 17-tool MCP allowlist stapling Figma MCP + Pencil MCP + Pencil-to-code | skills | MEDIUM | R4 §RETIRE candidates | Three distinct tools in one skill | Split into `moai-tool-figma` + `moai-tool-pencil` OR absorb Pencil side into moai-workflow-pencil-integration |
| P-S13 | Design cluster 4-way overlap: design-craft + domain-uiux + domain-brand-design + domain-copywriting | skills | MEDIUM | R4 §Merge Cluster 2, §Overlap heatmap | Four skills touch "what should this look like" | Merge non-agency-contracted skills into `moai-design-system`; freeze brand-design and copywriting (agency absorption contracts) |
| P-S14 | moai-workflow-templates is already an aggregate (`replaces:` frontmatter) — and further overlaps with moai-workflow-project | skills | MEDIUM | R4 §REFACTOR candidates | Nested aggregation | Fold into moai-workflow-project |
| P-S15 | Progressive Disclosure explicit declaration missing from 40% of skills (19 of 48) | skills | LOW | R4 §Progressive Disclosure enablement audit | Silent defaults applied | Add progressive_disclosure frontmatter to all 48 skills; enforce via skill-authoring.md required field |
| P-S16 | moai-foundation-context has zero references (12.8KB, score 23) — KEEP rationale unclear beyond "foundational" | skills | LOW | R4 §Notable KEEP rationales | Token-budget guidance applies everywhere but is never named | Either merge into moai-foundation-core OR add explicit invocation guidance |
| P-S17 | Level 2 token budget inconsistent across ref-* skills — ref-* uses 3000 but docs-generation uses 5000, design-craft uses 4500 | skills | LOW | R4 §Progressive Disclosure enablement audit | No documented budget ladder | Document budget convention: ref-* = 3000, specialist = 5000, orchestrator = 8000+ |
| P-S18 | Stale sibling references persist: moai-essentials-debug (8 refs), moai-infra-docker (2 refs), moai-quality-testing (1-2 refs) — none exist | skills | LOW | R4 §Category coverage gaps | Dangling `related-skills` entries | Audit related-skills fields; remove stale entries |
| P-S19 | Language rules vs language skills boundary confusion — 16 language rules in .claude/rules/moai/languages/, zero language skills, but skills reference siblings as if they existed | skills | HIGH | R4 §Summary of architectural health issues item 9 | Architectural inconsistency | Decide one home in v3; migrate consistently |
| P-H01 | 10 of 27 hook handlers are logging-only no-ops: configChange, fileChanged, elicitation, elicitationResult, instructionsLoaded, notification, subagentStop, taskCreated, postToolFailure, setup | hooks | HIGH | R6 §2.2 Go handlers | Handlers add session overhead without functional value | Decide per-handler: upgrade to real logic OR remove from settings.json (saves ~10 hook-fire overheads) |
| P-H02 | subagentStopHandler should kill tmux pane per team protocol but only logs — **known bug** (teammate_mode_regression memory, feedback_team_tmux_cleanup memory) | hooks | **CRITICAL** | R6 §2.2 + MEMORY.md feedback entries | Tmux panes leak after team session ends | Implement tmux pane cleanup: read tmuxPaneId from team config, kill-pane, then log |
| P-H03 | Orphan setupHandler: Go handler exists in internal/hook/setup.go but no shell wrapper and no settings.json entry | hooks | HIGH | R6 §2.1 + §2.3 | Dead code or incomplete feature | Either implement wrapper + registration OR remove the Go handler |
| P-H04 | Hardcoded absolute path `/Users/goos/go/bin/moai` in all 26 shell wrappers (first fallback attempt) | hooks | **CRITICAL** | R6 §2.1 finding | Breaks for every user except `goos` on every machine | Use `$HOME/go/bin/moai` as primary fallback; regenerate templates via `make build` after GoBinPath resolution changes |
| P-H05 | All hooks use exit-code-only protocol — no JSON output capable of injecting context, permission decisions, or updated input | hooks | HIGH | R6 §2.2 + R3 §4 Adoption Candidate 4 | Rich capabilities like Sprint Contract injection, MX tag additions are not programmable | Migrate handlers to JSON output (see design-principles.md Principle 8) |
| P-H06 | 5 yaml sections have no Go loader: constitution.yaml, context.yaml, interview.yaml, design.yaml, harness.yaml | config | **CRITICAL** | R6 §5.2 Unused sections | harness.yaml is central to v3 design but has no runtime enforcement | Add Go loaders + structs in internal/config/types.go; add at least one test each |
| P-H07 | sunset.yaml is dormant — SunsetConfig struct exists, no code path tests its thresholds | config | MEDIUM | R6 §5.2 Dormant | 4 references total, all in defaults/types | Activate (wire up to cleanup workflows) OR retire |
| P-H08 | /98-github.md is 698 LOC — violates thin-wrapper pattern (coding-standards.md §Thin Command Pattern limits to 20 LOC) | commands | HIGH | R6 §1.2 | Dev-only but coexists with enforced pattern | Extract to `moai-workflow-github` skill; make thin wrapper |
| P-H09 | /99-release.md is 890 LOC — same thin-wrapper violation | commands | HIGH | R6 §1.2 | Same | Extract to `moai-workflow-release` skill; make thin wrapper |
| P-H10 | core/lsp-client.md is a SPEC decision record (SPEC-LSP-CORE-002 Decision Point 1), not an agent rule | rules | LOW | R6 §4.2 | Misfiled — rule tree is for agent rules | Move to `.moai/decisions/` OR merge into SPEC file |
| P-H11 | workflow/workflow-modes.md (195 LOC) and workflow/spec-workflow.md (217 LOC) overlap significantly on Plan-Run-Sync flow | rules | MEDIUM | R6 §4.4 Redundancy | Duplicate content | Merge into workflow/spec-workflow.md |
| P-H12 | workflow/team-protocol.md (54 LOC) duplicates "Team File Ownership" content in worktree-integration.md (303 LOC) | rules | MEDIUM | R6 §4.4 Redundancy | Overlap | Merge into worktree-integration.md |
| P-H13 | workflow/file-reading-optimization.md is heuristic doc, not an enforceable rule | rules | LOW | R6 §4.4 | Pure heuristic, no HARD rules | Move to `.claude/skills/moai-foundation-context/references/` |
| P-H14 | 4 rule files use legacy `description + globs` frontmatter while rest use `paths:` — inconsistent | rules | LOW | R6 §4.3 Frontmatter consistency | moai-constitution.md, coding-standards.md, team-protocol.md, worktree-integration.md | Migrate to `paths:` CSV format (newer Claude Code official format) |
| P-H15 | configChangeHandler should trigger config reload when .moai/config/sections/*.yaml changes but is no-op | hooks | MEDIUM | R6 §2.2 | Missing business logic | Implement reload-and-revalidate on config file changes |
| P-H16 | instructionsLoadedHandler could validate CLAUDE.md 40,000-char limit (coding-standards.md rule) but is no-op | hooks | LOW | R6 §2.2 | Missed enforcement opportunity | Add character-budget validator |
| P-H17 | fileChangedHandler could trigger MX re-scan for externally-edited files but is no-op | hooks | LOW | R6 §2.2 | Missed observability opportunity | Add MX re-validation on FileChanged event |
| P-H18 | design.md and db.md use `.md` extension in templates while other 13 commands use `.md.tmpl` — minor template extension drift | commands | LOW | R6 §1.1 note | Extension inconsistency | Unify on `.md.tmpl` convention |
| P-H19 | Only 16 of 27 hook events (59%) have full business logic; 10 are partial (logging-only); 1 is missing (setup) | hooks | HIGH | R6 §A Hook Coverage Matrix | Coverage stats | Upgrade 5 critical handlers (subagentStop, configChange, setup, instructionsLoaded, fileChanged); decide retire vs upgrade for remaining 5 |
| P-H20 | workflow.yaml schema partially unstructured — only role_profiles are read by Go code | config | MEDIUM | R6 §5.1 | Incomplete schema | Complete workflow.yaml typed struct in internal/config/types.go |
| P-R01 | settings.json hook registrations drift from Go handler list — 27 Go handlers but 25 native event registrations (setup orphan + autoUpdate composite) | hooks | MEDIUM | R6 §2.3 | 1 orphan handler + 1 composite confuse maintainers | Document the 2 non-registered handlers explicitly in hooks-system.md |
| P-R02 | Constitutional sprawl risk: constitution is ~800 lines across CLAUDE.md + 3 rule files — Anthropic's internal finding is that too-long constitutions confuse models | rules | MEDIUM | R1 §18 Constitutional AI anti-pattern flag | CLAUDE.md ~860 lines + moai-constitution.md 266 + design/constitution.md 404 + agent-common-protocol.md 157 | Consolidation pass; extract HARD rules to single load-bearing file; reference from others |
| P-R03 | Cross-reference drift: CLAUDE.md Section 8 duplicates agent-common-protocol.md AskUserQuestion rules | rules | MEDIUM | R6 §4.4 Redundancy | Same content in both files | Keep rule as authoritative; tighten CLAUDE.md to reference only |
| P-C01 | No permission bubble model in moai — implicit trust on agent execution | architecture | **CRITICAL** | R3 §4 Adoption Candidate 2 | moai today has `permissions.allow` lists but no sandbox or bubble | Implement permission stack with bubble mode (see design-principles.md Principle 6) |
| P-C02 | No sub-agent context isolation primitive in moai — relies on Claude Code's `Agent()` | architecture | HIGH | R3 §4 Adoption Candidate 3 | v3 run phase could benefit from explicit context objects | Define moai's own context primitive OR formalize reliance on CC's isolation |
| P-C03 | No sandboxed execution default in moai — matches 2026 security anti-pattern | architecture | **CRITICAL** | R2 §B Anti-pattern 1, §A Top-5 Pattern 5 | Open to prompt-injection supply chain attacks | Add sandbox layer per design-principles.md Principle 7; document invocation in security.yaml |
| P-C04 | Two-tier config (~/.moai/ + .moai/) lacks provenance tracking — "which file set this?" is opaque | architecture | HIGH | R3 §4 Adoption Candidate 1 | No source tagging on config values | Add `source` metadata in typed loaders; expose via doctor/config-dump |
| P-C05 | No cache-prefix discipline on system prompts — every turn potentially reassembles prompts in non-stable order | architecture | MEDIUM | R3 §2 Decision 2 | Loses prompt-caching benefit; moai has not considered caching | Audit system prompt assembly order; freeze `(system, memory, append)` sequence |
| P-C06 | No silent migration pattern — moai migrate is explicit CLI command users must remember | architecture | LOW | R3 §2 Decision 10 | Compare CC's `runMigrations()` at preAction | Consider auto-migration at `moai init` / `moai hook session-start` time |
| P-Z01 | Current moai GAN loop (design constitution §11.4) retains evaluator memory across iterations — violates Agent-as-a-Judge anti-pattern | architecture | HIGH | R1 §9 Agent-as-a-Judge anti-pattern flag | "historical memory in judges can cascade errors" | Amend constitution: evaluator judgments fresh per iteration; Sprint Contract state durable (see design-principles.md Principle 4) |
| P-Z02 | Utility subcommands (/moai fix, /moai coverage, /moai codemaps, /moai mx, /moai clean) may over-use multi-agent for well-structured tasks | architecture | MEDIUM | R1 §25 Agentless anti-pattern flag | Well-structured tasks match Agentless pattern | Classify subcommands: multi-agent (plan/run/design/sync) vs fixed-pipeline (fix/coverage/codemaps/mx/clean); empirically measure |
| P-X01 | Template-local drift risk on new files — `/98-github.md` and `/99-release.md` are not templated; v3 templates currently have drift between .md and .md.tmpl extensions | commands | LOW | R6 §1.1 note + §1.2 | Inconsistent | Unify; move dev-only to local overrides explicitly marked `type: local` |

---

## Problem clusters

Problems cluster into 6 structural themes. Each cluster is a candidate for a single v3 SPEC that addresses the root cause rather than individual symptoms.

### Cluster 1: Common Protocol Enforcement (Critical)

**Problems**: P-A01, P-A04, P-A11, P-A22, P-A13, P-A18, P-A23

**Pattern**: Agent bodies embed rules that contradict constitutional HARD rules. 9 agents violate the "no AskUserQuestion in subagents" rule; 4 agents declare `Agent` tools they cannot use; 6 agents are missing `isolation: worktree` per worktree rules; effort-level inheritance pattern means 19 agents silently diverge from constitution.

**Root cause**: No CI lint rejects constitutional violations. Agent-authoring.md documents the rules but does not enforce them.

**v3 SPEC direction**: Add a `moai doctor agents` lint that enforces agent-authoring.md rules. Publish an agent-contract schema (required fields, forbidden strings) and fail builds on violation. Extract duplicate HARD blocks to agent-common-protocol.md.

### Cluster 2: Skill Surface Reduction (High)

**Problems**: P-S01, P-S02, P-S03, P-S04, P-S13, P-S14, P-S11, P-S10, P-S12

**Pattern**: 48 skills is past discoverability ceiling. 5 merge clusters collapse 15 skills to 6. Platform triplets and domain kitchen sinks are maintenance bombs.

**Root cause**: Skill creation has no "why not merge into existing" gate. Naming taxonomy is inconsistent (moai-foundation-* vs moai-platform-* vs moai-domain-* vs moai-framework-* vs moai-library-*).

**v3 SPEC direction**: Target ~24 skills. Formalize merge cluster execution in waves (R4 §v3 migration sequencing). Standardize progressive_disclosure budget ladder (ref=3000, specialist=5000, orchestrator=8000+).

### Cluster 3: Hook Completeness and Safety (Critical)

**Problems**: P-H01, P-H02, P-H03, P-H04, P-H05, P-H19, P-H15, P-H16, P-H17, P-H20, P-R01

**Pattern**: 10 logging-only handlers + 1 orphan + hardcoded absolute path + exit-code-only protocol = the hook system is half-implemented. The known tmux pane leak (P-H02) is a user-visible bug from this half-implementation.

**Root cause**: Hooks added opportunistically per Claude Code event as events are released, without a companion logic-or-retire decision. The 26 shell wrappers were auto-generated assuming a future Go handler would emerge — some did, some didn't.

**v3 SPEC direction**: Decide each of 10 stub handlers: upgrade to real logic OR remove from settings.json. Migrate handlers to JSON output protocol (design-principles.md Principle 8). Fix the hardcoded path as a v3 blocker.

### Cluster 4: Config Schema Completeness (Critical)

**Problems**: P-H06, P-H07, P-H20, P-C04

**Pattern**: 5 yaml sections lack Go loaders (constitution, context, interview, design, harness). harness.yaml is central to v3 design but has no runtime enforcement. sunset.yaml is dormant. workflow.yaml schema partial.

**Root cause**: Config files are added as documentation artifacts first and Go integration is often deferred indefinitely. No CI check "every yaml section has a loader."

**v3 SPEC direction**: Add loaders + structs + tests for all 5 unloaded sections. Decide sunset.yaml fate. Add CI check that rejects a new yaml section without a corresponding Go struct and test.

### Cluster 5: Architectural Primitives Adoption from CC (Critical)

**Problems**: P-C01, P-C02, P-C03, P-C04, P-C05, P-C06, P-Z01

**Pattern**: moai lacks five structural primitives that CC built and validated at scale: permission bubbling, sub-agent context isolation, multi-source provenance, sandboxed execution, cache-prefix discipline.

**Root cause**: moai v2.x focused on workflow orchestration and trusted CC for execution-layer safety. The 2026 security consensus (OWASP Top 10 Agentic Apps, Cline npm incident) makes trust-only approach untenable.

**v3 SPEC direction**: Adopt the 5 R3 recommendations (R3 §4) explicitly — each as a SPEC. Defer the CC debt items (Ink fork, Claude-only assumptions, bridge, feature flags, MCP registry) per R3 §5.

### Cluster 6: Agent Consolidation (High)

**Problems**: P-A05, P-A06, P-A07, P-A08, P-A09, P-A20, P-A21

**Pattern**: 3 advisor-only experts (expert-debug, expert-testing, expert-performance) that delegate everything. 3 near-identical builders. manager-ddd/manager-tdd 60% overlap. manager-project scope sprawl. Missing manager-design.

**Root cause**: Agent decomposition was done by "what could exist" rather than "what owns unique file-write authority." Result: 22 agents, of which ~5 are pure routers with no DO authority.

**v3 SPEC direction**: Target 17 agents. Merge builders to one. Merge manager-ddd+tdd to manager-cycle. Retire advisor-only experts into manager-quality and manager-cycle. Add manager-design if /moai design is first-class.

---

## Problem severity distribution by cluster

| Cluster | Critical | High | Medium | Low | Total |
|---------|----------|------|--------|-----|-------|
| 1. Common Protocol Enforcement | 1 (P-A01) | 6 | 3 | 1 | 11 |
| 2. Skill Surface Reduction | 0 | 3 | 6 | 3 | 12 |
| 3. Hook Completeness and Safety | 2 (P-H02, P-H04) | 5 | 3 | 2 | 12 |
| 4. Config Schema Completeness | 1 (P-H06) | 2 | 1 | 0 | 4 |
| 5. Architectural Primitives Adoption | 2 (P-C01, P-C03) | 3 | 1 | 1 | 7 |
| 6. Agent Consolidation | 0 | 5 | 2 | 0 | 7 |
| Uncategorized | 2 | 0 | 1 | 6 | 9 |
| **Total** | **8** | **21** | **29** | **14** | **72** |

Clusters 1, 3, 4, 5 carry the 8 CRITICAL issues. v3 wave sequencing should lead with these; Cluster 2 and 6 are structural reduction that can follow.

---

## v3 wave sequencing recommendations

### Wave 3a — Architectural foundations (first)
- Cluster 5: Adopt permission bubble, sandbox layer, multi-source config provenance from CC.
- Cluster 4: Complete config schema (all yaml sections loaded).
- Cluster 3: Fix CRITICAL hook bugs (subagent-stop tmux leak, hardcoded path).

### Wave 3b — Protocol enforcement
- Cluster 1: CI lint for agent authoring rules; scrub AskUserQuestion violations; add missing worktree isolation; populate effort field on every agent.

### Wave 3c — Surface reduction
- Cluster 2: 48 → 24 skills via documented merge waves.
- Cluster 6: 22 → 17 agents via documented merge/retire plan.

### Wave 3d — Hook + command cleanup
- Cluster 3: Upgrade or retire 10 stub handlers; migrate to JSON protocol.
- Decide fate of /98-github.md and /99-release.md (extract to skills).

### Wave 3e — Polish
- Rule tree cleanup (move lsp-client.md, merge duplicates, unify frontmatter).
- Skill ecosystem hygiene (reference/ vs references/, level-3 pruning, related-skills audit).

---

## Sources

All problems in this catalog cite a specific Wave 1 section with `R<n> §<section>` reference. Verification paths:

- R3 architectural debt: `.moai/design/v3-redesign/research/r3-cc-architecture-reread.md` §3 (Limitations), §4 (Adoption Candidates), §5 (Divergence), §6 (Big Idea contrast).
- R4 skill audit: `.moai/design/v3-redesign/research/r4-skill-audit.md` §Per-skill table, §Detailed findings, §Summary of architectural health issues.
- R5 agent audit: `.moai/design/v3-redesign/research/r5-agent-audit.md` §Per-agent audit table, §Detailed findings, §Effort calibration matrix, §Common Protocol compliance, §Worktree correctness, §Tool scope audit.
- R6 commands/hooks/rules/config: `.moai/design/v3-redesign/research/r6-commands-hooks-style-rules.md` §1 (Commands), §2 (Hooks), §4 (Rules), §5 (Config), §A-D (Cross-cutting).
- R1 papers supplying principle-level anti-patterns: ReAct §1 (token-trace bloat), Reflexion §2 (memory separation), Agent-as-a-Judge §9 (cascade errors), Agentless §25 (fixed pipeline for well-structured), Constitutional AI §18 (sprawl risk), Claude Opus 4.7 Effort §19 (counterproductive scaffolding).
- R2 opensource tools: ralph §1-2 (file-first), OMC §3 (operating modes), Cline §5 (approval-fatigue exploit), SWE-agent §8 (ACI), crewAI §15 (guardrails), LangGraph §16 (typed state + interrupt), sandbox consensus §A Top-5 Pattern 5.
