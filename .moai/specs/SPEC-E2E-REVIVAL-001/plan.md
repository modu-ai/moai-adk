# SPEC-E2E-REVIVAL-001 — Implementation Plan

> Milestones are ordered by **decision-reversibility**: the decisions most likely to change under review (agent contract, platform matrices, UX flow) lead; mechanical/reconciliation steps close. Priority labels only — no time estimates.

## §A Context & Inputs

- **Design baseline**: retired workflow recovered via `git show c6b04d39c~1:.claude/skills/moai/workflows/e2e.md` (452 lines, v2.1.0 body / v2.7.0 frontmatter, web-only, 4 tools). Inherit: phase structure, auto-detection → AskUserQuestion selection, recommendation-matrix pattern, token-cost table, journey definition format, artifact directory conventions, TaskCreate/TaskUpdate tracking. Strengthen: token-cost principle promoted from a comparison-table column to [HARD] requirements (REQ-E2E-100/101/105). Drop: stale tool facts (chrome-devtools-mcp "29 tools" → now 60+; per research.md §C).
- **Verified tool stack** (live-fetched 2026-07-13, research.md §Sources): Playwright 1.61.1, agent-browser v0.31.1, chrome-devtools-mcp v1.5.0, Maestro CLI 2.6.1, Appium 3.5.2, Detox 20.51.3, Tauri v2 `@wdio/tauri-service`, Playwright-Electron (experimental).
- **Repo baseline** (measured this tree, 2026-07-13): spec.md §A table. Line anchors are content-token anchored (constants drift; grep the constant name, not the line number).

## §B Key Decisions (highest change-likelihood first)

- **D-1 — Agent name & placement**: `e2e-specialist` at `.claude/agents/moai/e2e-specialist.md` (FLAT moai/ layout, matching the 9 existing retained agents). Class: `core/specialist`. Rationale: "Define a custom subagent when you keep spawning the same kind of worker" (Anthropic best practices, already cited in CLAUDE.md §4); e2e execution is a recurring, tool-heavy, output-noisy worker profile that pollutes manager-develop's run-phase context. Alternative rejected: keeping manager-develop delegation (the retired design) — it conflates implementation-cycle ownership with test-execution ownership and forfeits the token-diet of a purpose-built agent.
- **D-2 — Catalog expansion consequence**: adding a 10th MoAI-custom agent BREAKS the literal CLAUDE.md §4 claim "exactly 10 retained agents (9 MoAI-custom + 1 Explore)" present in BOTH trees. The SPEC therefore owns updating: count text, catalog table row, Selection Decision Tree entry (new item 12: "E2E test execution across web/mobile/desktop? Use the e2e-specialist subagent"), in both trees. Missing any of the three leaves the agent inert-by-omission. AUDIT-EXTENDED (iter-1 D2): the count-literal surface is larger than CLAUDE.md §4 alone — repo-wide grep (`'10 retained agents|9 MoAI-custom|10-agent'`, re-executed 2026-07-13) measures **12 files / 24 sites** across both trees: CLAUDE.md ×3 sites, agent-authoring.md ×3, agent-patterns.md ×3 (incl. the MoAI-custom name enumeration), model-policy.md ×1, spec-workflow.md ×1, manager-design.md ×1 — each ×2 trees. All are in M3 scope per REQ-E2E-302. Line numbers drift between trees (local agent-authoring.md is offset +1 from the template) → anchor every edit by content token, never line number. Post-change invariance: stale-literal grep over the 12 touched files → 0 (CMD-019-INV).
- **D-3 — Platform-toolchain matrix** (defaults; full comparison in research.md, full matrix in design.md §C):
  - web → **Playwright CLI** (default) | agent-browser (AI-exploratory) | chrome-devtools-mcp (perf/Lighthouse ONLY, MCP-tier) | Claude in Chrome (interactive debug ONLY, MCP-tier)
  - mobile → **Maestro** (default: declarative YAML flows, single-binary CLI, deterministic, low-token) | Appium 3.x (fallback: widest device/driver matrix, W3C WebDriver) | Detox (React-Native-specific gray-box option, auto-offered only when RN markers detected)
  - desktop → **Playwright `_electron`** (Electron apps; experimental API — pin known-good version in skill prose) | **WebdriverIO + `@wdio/tauri-service`** (Tauri apps; embedded mode is cross-platform incl. macOS; native `tauri-driver` route is Windows/Linux-only) | `desktop-native` → automation DEFERRED to a follow-up SPEC (detection still classifies; workflow reports the deferral via the REQ-E2E-007 graceful branch)
- **D-4 — Skill placement**: workflow file INSIDE the moai skill (`.claude/skills/moai/workflows/e2e.md`), exactly like the retired version and the 12 sibling workflows. Consequence: `expectedSkillCount` 28 is UNTOUCHED; only the moai skill's catalog hash moves (auto via `make build`). A new top-level skill directory was rejected: it would inflate session skill-listing budget and diverge from sibling subcommand precedent.
- **D-5 — Token-minimization protocol** (design.md §F): 3-rung escalation ladder — (1) CLI with bounded tail + file-redirect; (2) CLI structured output (JSON reporters) parsed selectively; (3) MCP only for capabilities with no CLI equivalent (live perf traces, Lighthouse, interactive debugging), batched. The e2e-specialist body carries the ladder as [HARD]; the workflow's tool matrix carries the per-capability CLI-vs-MCP classification.
- **D-6 — Tier L justification**: 6 artifacts (research.md + design.md required); 3 platform domains needing live-verified external research; cross-tree distribution with 4 CI-guard test surfaces; a retained-agent catalog expansion touching always-loaded CLAUDE.md in both trees. Tier S/M rejected: the agent-catalog change alone has repo-wide blast radius (always-loaded context), and the external-stack decisions demanded research artifacts.
- **D-7 — Desktop-native fallback scope — RESOLVED (user decision via orchestrator AskUserQuestion, 2026-07-13)**: DEFERRED to a follow-up SPEC. REQ-E2E-502 removed from spec.md §B Group F; per the audit D6 split instruction, AC-E2E-006 was FIRST narrowed to REQ-E2E-007 only (no-target graceful exit keeps its own AC) and only the 502 half was removed. A detected `desktop-native` surface routes to the REQ-E2E-007 graceful branch with a deferral notice; a new Out-of-Scope H3 records the deferral.
- **D-8 — Mobile default — RESOLVED (user decision via orchestrator AskUserQuestion, 2026-07-13)**: Maestro default CONFIRMED as drafted (Appium fallback, Detox RN-conditional), per the token-minimization requirement (single CLI binary, YAML flows, deterministic output).

## §C Pre-flight Verified Baseline (run-phase re-verify before M1)

All measured 2026-07-13 on this tree; re-run before implementation (parallel-session drift hazard):

1. `git show c6b04d39c~1:.claude/skills/moai/workflows/e2e.md | wc -l` → 452 (baseline recoverable)
2. `grep -c 'e2e' .claude/skills/moai/SKILL.md internal/template/templates/.claude/skills/moai/SKILL.md` → 0, 0
3. `grep -n 'expectedAgentCount = 9' internal/template/catalog_tier_audit_test.go` → 1 match; `grep -n 'expectedTotal = 37' internal/template/catalog_loader_test.go` → 1 match; `grep -n 'expectedSkillCount = 28' internal/template/catalog_tier_audit_test.go` → 1 match
4. `ls internal/template/templates/.claude/agents/moai/*.md | wc -l` → 9; local same → 9
5. `ls internal/template/templates/.claude/commands/moai/*.md.tmpl | wc -l` → 13; local `*.md` → 13
6. `grep -c 'Subcommands:' CLAUDE.md internal/template/templates/CLAUDE.md` → 1 each, neither containing `e2e`
7. gen-catalog-hashes contract: `internal/template/scripts/gen-catalog-hashes.go` iterates EXISTING catalog entries and rewrites `hash:` only — a new agent requires a MANUAL catalog.yaml entry (name/tier/path/version) before `make build` computes its hash
8. Inspect one existing command pair (`run.md` local vs `run.md.tmpl` template) to replicate the exact render-pattern delta before authoring `e2e.md.tmpl` (sizes differ: local files are rendered; do NOT assume byte parity for the command pair)
9. Mirror-parity registration: check whether any `.claude/skills/moai/workflows/*.md` path is registered in `rule_template_mirror_test.go` (`workflowOptMirroredPaths`) — follow sibling precedent for the new e2e.md (register iff siblings are registered)
10. `.moai/specs/` dedup: no `SPEC-E2E-REVIVAL-*` exists (only `SPEC-HARNESS-EXECUTE-E2E-001`, unrelated: harness telemetry bugfix)
11. Re-run the count-literal surface grep (D-2 / REQ-E2E-302): `grep -rEn '10 retained agents|9 MoAI-custom|10-agent' <12-file list>` — reconcile against HEAD before M3 (baseline 2026-07-13: 24 sites; the surface can grow between plan and run)

## §D Constraints

- [HARD] Template-First order inside every milestone: template tree → `make build` → local sync. Never local-first.
- [HARD] Template neutrality (§25): the three template artifacts carry ZERO internal SPEC IDs / REQ tokens / dates / SHAs. Self-check each file against the forbidden-class catalogue before commit; the neutrality CI guard is the safety net, not the first line.
- [HARD] The e2e-specialist `tools:` CSV excludes `Agent` and `AskUserQuestion`. Include: Read, Write, Edit, Bash, Grep, Glob, TaskCreate, TaskUpdate, TaskList, TaskGet, Skill (mirror manager-develop's profile minus mcp__context7; final list in design.md §D).
- [HARD] Thin command body <20 non-empty lines with `Skill(` or `subagent` token (TestCommandsThinPattern R3/R4).
- [HARD] CI count-constant edits carry provenance comments in the established comment-ledger style of `catalog_tier_audit_test.go` (append to the existing history block; keep the comment neutral of this SPEC's ID? NO — test files are NOT template content; internal SPEC IDs are the established convention there. Neutrality binds `templates/**` only).
- [HARD] ACs verify REACHABILITY (file-exists + cross-file reference + test-passes chains), never token presence alone.
- Commit discipline: plan-phase commit `feat(SPEC-E2E-REVIVAL-001): plan-phase artifacts (L, 6 artifacts)`; run-phase one commit per milestone `feat(SPEC-E2E-REVIVAL-001): M<n> <slug>`.

## §E Self-Verification (run-phase exit gates)

- E1: AC matrix PASS/FAIL table (all 27), each row citing an executed command + verbatim-output path
- E2: `go test ./internal/template/...` and `go test ./...` exit 0 (full suite, not just touched packages)
- E3: `golangci-lint run` exit 0
- E4: Subagent-boundary grep — e2e-specialist tools line clean of `Agent`/`AskUserQuestion`
- E5: Both-tree parity evidence — `diff` outputs for skill + agent (expected identical), rendered-equivalence note for command pair
- E6: Neutrality self-check — `grep -rn 'SPEC-E2E-REVIVAL\|REQ-E2E-' internal/template/templates/` → 0 matches
- E7: `moai spec lint --strict .moai/specs/SPEC-E2E-REVIVAL-001/spec.md` → 0 findings (the lint CLI takes the spec.md FILE; a directory argument fails with `ParseFailure … is a directory` — executed evidence, audit iter-1 D7)

## §F Milestones

### M1 — e2e-specialist agent (template tree) — Priority: High, most design-sensitive

- Author `internal/template/templates/.claude/agents/moai/e2e-specialist.md` per design.md §D: frontmatter (name, description with PROACTIVELY + "NOT for:" clause, tools CSV, `model: inherit`, `effort: high`, color, `permissionMode`, `memory: project`, `skills:` ≤2), body (scope, platform-toolchain execution knowledge, token-minimization ladder [HARD], blocker-report protocol, artifact-directory conventions, delegation-return contract).
- Sync byte-identical copy to `.claude/agents/moai/e2e-specialist.md`.
- Exit: file exists in both trees, byte-identical; frontmatter COMPLETENESS (name/description/tools CSV/model/effort/color/permissionMode/memory/skills) verified by MANUAL field-by-field checklist — `TestAgentFrontmatterAudit` guards retired-field cleanliness ONLY, not completeness (audit iter-1 D12); the CI runs land in M4 when the count constant moves.

### M2 — e2e workflow skill (template tree) — Priority: High

- Author `internal/template/templates/.claude/skills/moai/workflows/e2e.md` per design.md §E: frontmatter (`user-invocable: false`, triggers, version reset to 3.0.0 lineage), Phase 0 detection (marker matrix §B) → Phase 0.5 selection (orchestrator AskUserQuestion contract — the workflow INSTRUCTS the orchestrator; it never claims the agent prompts) → Phase 1 journey mapping → Phase 2 script creation (per-toolchain file conventions) → Phase 3 execution (CLI-first, bounded output) → Phase 4 recording (native facilities) → Phase 5 report (conversation_language). Embed: platform-toolchain matrix with token-cost column, 3-rung MCP escalation ladder, e2e-specialist delegation directives per phase.
- Sync identical copy to local tree.
- Exit: both trees identical; body contains `e2e-specialist` delegation in ≥3 phase sections; ≤~450 lines; zero neutrality-class violations.

### M3 — Thin command + router/docs re-registration (both trees) — Priority: High

- Author `internal/template/templates/.claude/commands/moai/e2e.md.tmpl` (replicating the sibling render pattern measured in pre-flight #8) + local `.claude/commands/moai/e2e.md`.
- SKILL.md (both trees): add `**e2e**` Priority 1 row (position: after `gate`, before `harness`, keeping alphabetic-cluster precedent of the existing list order); add e2e cue exemplars to Priority 3; add `e2e` to frontmatter description enumeration.
- CLAUDE.md (both trees): §3 `Subcommands:` line + the three §4 literal sites (count text → "exactly 11 retained agents (10 MoAI-custom + 1 Explore)"; "flat-hierarchy 10-agent consolidation rationale" → 11-agent; "one of the 10 retained agents above" → 11) + §4 catalog table row + Selection Decision Tree entry (appended as entry 12; manager-design stays entry 11).
- Count-literal sweep (both trees, full REQ-E2E-302 surface): agent-authoring.md (3 sites), agent-patterns.md (3 sites incl. adding `e2e-specialist` to the MoAI-custom enumeration), model-policy.md (1 site), spec-workflow.md (1 site incl. named list), manager-design.md (1 site). Content-token anchored edits (line numbers drift +1 between trees for agent-authoring.md).
- Exit: grep-verifiable deltas per acceptance.md AC-E2E-017…021 (baseline 0 → 1 per surface, both trees) + invariance CMD-019-INV → 0 stale literals over the 12-file surface.

### M4 — catalog.yaml + CI-guard constants + make build + embed verification — Priority: Medium (mechanical)

- Manually add `e2e-specialist` entry to `catalog.yaml` `core.agents` (name/tier: core/path/version: 1.0.0; hash placeholder).
- Update `expectedAgentCount` 9→10 and `expectedTotal` 37→38 with ledger-style provenance comments.
- `make build` (embeds templates, regenerates ALL hashes including the new entry and the changed moai-skill hash).
- Exit: `go test ./internal/template/...` exit 0 (this single gate transitively proves: thin pattern, frontmatter audit, catalog membership, catalog path resolution, count reconciliation, neutrality).

### M5 — Full verification batch + evidence — Priority: Medium

- Single-turn parallel batch: `go test ./...` | `golangci-lint run` | boundary greps (E4) | parity diffs (E5) | neutrality grep (E6) | `moai spec lint --strict` (E7) | AC matrix sweep with per-AC executed evidence.
- Populate progress.md §E.2/§E.3 (manager-develop ownership).
- Exit: E1–E7 all green; blocker report for anything red.

## §G Anti-Patterns (bind run-phase)

- **AP-1 Token-presence AC**: `grep -c e2e` alone proves nothing — every reachability AC pairs existence + cross-reference + passing CI test (lesson: AC-token-presence-not-reachability).
- **AP-2 Local-first authoring**: editing `.claude/` before `templates/` inverts Template-First and desyncs on next `moai update`.
- **AP-3 Neutrality leak**: copying this SPEC's ID or REQ tokens into template artifact prose (test files are exempt; templates are not).
- **AP-4 Stale-fact inheritance**: porting the retired workflow's tool facts verbatim (e.g., "29 tools") instead of the 2026-verified numbers in research.md.
- **AP-5 Count-constant drift**: editing a count constant without the matching artifact in the SAME commit (breaks bisect).
- **AP-6 Inert agent**: shipping e2e-specialist without the M3 catalog-table/decision-tree/workflow-delegation wiring — grep-green but unreachable (lesson: inert-deliverable reachability).
- **AP-7 Naive parity assumption**: asserting byte-parity for the command `.md` vs `.md.tmpl` pair (they render differently — pre-flight #8).

## §H Cross-References

- research.md — verified tool-stack evidence + sources (this SPEC dir)
- design.md — matrices, agent contract, phase design (this SPEC dir)
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter SSOT
- `internal/template/CLAUDE.md` — Template-First + embed mechanics
- CLAUDE.local.md §2 / §15 / §25 — sync rules, language neutrality, internal-content isolation
- `.claude/rules/moai/core/agent-common-protocol.md` — blocker report + file-redirect contract
- Retired baseline: `git show c6b04d39c~1:.claude/skills/moai/workflows/e2e.md`
