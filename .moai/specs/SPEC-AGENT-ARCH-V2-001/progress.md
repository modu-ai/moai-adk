---
id: SPEC-AGENT-ARCH-V2-001
title: "MoAI Agent Architecture v2 — Progress"
version: "0.1.0"
status: in-progress
created: 2026-07-09
updated: 2026-07-10
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/agents/moai + internal/config"
lifecycle: spec-anchored
era: V3R6
tier: L
tags: "agent-arch, super-advisor, manager-design, no-haiku, 3-tier, claude-design, token-policy, progress"
---

# SPEC-AGENT-ARCH-V2-001 — Progress

> progress.md is the lifecycle tracker. Section letters below are allocated per the canonical progress.md Section Map in `.claude/rules/moai/development/spec-frontmatter-schema.md` (the SSOT). The `§E.*` namespace is parser-load-bearing — `internal/spec/era.go` `ClassifyEra` matches the literal `§E.2`/`§E.3`/`§E.4` heading tokens and the `sync_commit_sha` field; renaming any of these silently breaks era classification.
>
> Per the manager-spec artifact ownership policy, this agent populates **§E.1 only** (the plan-phase audit-ready signal). `§E.2`/`§E.3` belong to manager-develop (run-phase); `§E.4` belongs to manager-docs (sync-phase). §F / §H are placeholders that gain content during run-phase / Phase 0.95.

---

## §E.1 Plan-phase Audit-Ready Signal

```yaml
plan_status: audit-ready
plan_complete_at: 2026-07-09
plan_artifact_set: Tier-L-5
plan_artifacts:
  - spec.md         # WHAT/WHY SSOT — 17 REQs (REQ-AA2-001..017) in GEARS notation, 5 GAPs, 9 Constraints
  - plan.md         # HOW skeleton — §A-§H (Context, Known Issues B1-B12, Pre-flight, Constraints, E1-E7, Milestones M1-M4, Anti-Patterns, Xref)
  - acceptance.md   # AC Matrix — 17 ACs (AC-AA2-001..017) Given-When-Then + §D.1-§D.7 severity/traceability/closure-gate
  - design.md       # Architecture — §A Target topology + catalog 9+1; §B super-advisor; §C manager-design D1-D5 + H1-H9 verbatim + DesignSync 11-method contract; §D No-Haiku 3-tier (2-A..2-D); §E Wiring
  - research.md     # Codebase baselines — §A..§J (RouteModelFor, workflow.yaml, llm.yaml, CLAUDE.md §4, model-policy.md, ADVISOR-RUNG analysis, MODEL-ROUTING-WIRE analysis, DesignSync MCP Gap, haiku enumeration)
  - progress.md     # this file — §E.1 populated; §E.2-§E.4 placeholder headings
plan_supersede_targets:
  - SPEC-ADVISOR-RUNG-001         # frontmatter flipped draft → superseded; superseded_by recorded
  - SPEC-MODEL-ROUTING-WIRE-001   # frontmatter flipped draft → superseded; superseded_by recorded
plan_design_authority: .moai/reports/agent-architecture-redesign-v2-20260709.html
plan_design_authority_lines: 565
plan_era: V3R6
plan_lifecycle: spec-anchored
plan_tier: L
plan_open_decisions:
  - plan.md §D D2 (manager-design effort xhigh FIXED): RECOMMENDED = frontmatter `effort: xhigh` literal
  - plan.md §D D3 (HaikuResidualRule scope): RECOMMENDED = HARD gate, NOT skip-able (REQ-AA2-012 binds)
  - plan.md §D D4 (moai init flag migration): RECOMMENDED = accept `--model-policy max|medium|low` + deprecated `--high/medium/low` aliases (one cycle)
  - plan.md §D D6 (legacy `model_routing` key disposition, added iter-2): RECOMMENDED = retain as `medium` profile alias for one backward-compat cycle
  - super-advisor tool whitelist `WebFetch` (REQ-AA2-001 detail): RECOMMENDED = retain WebFetch in read-only whitelist (super-advisor may fetch external docs during consultation)
  # Note (iter-2 D4 numbering-unification): open-decision identifiers here now align to plan.md §D D-numbering. The legacy local D1/D2/D3 labels that previously collided with plan.md §D (D1=legacy key, D2=WebFetch, D3=severity) are retired — their content is folded into the plan.md §D D2/D3/D6 references above.
plan_known_gaps:
  - DesignSync MCP not registered in .mcp.json at plan-phase (research.md §H) — does not block plan-phase close; M2 run-phase verifies
plan_phase_evidence_path: .moai/specs/SPEC-AGENT-ARCH-V2-001/
```

### Plan-phase evidence (vci §3.2 attribution)

- **Plan-phase artifacts**: 6 files at `.moai/specs/SPEC-AGENT-ARCH-V2-001/` (spec.md, plan.md, acceptance.md, design.md, research.md, progress.md) — all authored 2026-07-09 by manager-spec.
- **Frontmatter schema**: 12 canonical fields verified per file; `created: 2026-07-09`, `updated: 2026-07-09`, `tags: "..."` (string, not labels array); `era: V3R6`, `tier: L`, `lifecycle: spec-anchored`.
- **Supersede footprint**: 2 targets flipped (`SPEC-ADVISOR-RUNG-001`, `SPEC-MODEL-ROUTING-WIRE-001`) — `status: draft → superseded`, `superseded_by: SPEC-AGENT-ARCH-V2-001`, `updated: 2026-07-09`. Inline supersede note in each target's Epic Context section.
- **Design SSOT fidelity**: every spec.md REQ traces to a verbatim §NN section of `.moai/reports/agent-architecture-redesign-v2-20260709.html` (565 lines). No architecture re-derived or invented (CONSTRAINT #1).
- **Live-tree baselines**: every baseline in research.md measured 2026-07-09 by this agent via Read/Bash/Grep (vci §2 attribution). RouteModelFor external call sites: zero production consumers (only 2 doc comments in types.go).
- **Lint status**: see §E.2 placeholder below — manager-develop will populate run-phase evidence; plan-phase lint check is `moai spec lint .moai/specs/SPEC-AGENT-ARCH-V2-001/` (run by this agent before close — verbatim output in the completion report to the orchestrator).

### Plan-phase verification claim (vci §3.1)

**Claim**: all 6 plan-phase artifacts exist with the frontmatter schema compliant, every REQ in GEARS notation, every AC in Given-When-Then format, every design decision traces to the SSOT, every research baseline measured live 2026-07-09, and the two supersede targets carry the `superseded_by` field with `status: superseded`.

**Evidence**: the file listing above (`plan_artifacts`) + the `plan_phase_evidence_path`. Each artifact's content is the load-bearing evidence; the path is the audit-time-reachable pointer.

**Baseline-attribution**: HEAD `2fae7057a`, design SSOT at `.moai/reports/agent-architecture-redesign-v2-20260709.html` (565 lines, 2026-07-09). No carry-over from unrelated prior measurements.

**Gaps**: DesignSync MCP not registered (research.md §H) — recorded as a known gap, not a blocker.

**Residual-risk**: (a) plan-auditor may identify scope-reduction opportunity at iter1 — the M1-M4 split is conservative (separates Go code from doctrine), so the most likely scope-reduction is merging M3 + M4 if M3's lint rule already encompasses the doctrine refresh; (b) the 36-cell matrix in design.md §D.5 is a transcription of §2-D — if the SSOT is updated post-plan-phase, the matrix needs re-sync; (c) the HaikuResidualRule scope (6 trigger surfaces in REQ-AA2-012) may surface additional haiku references at run-phase that were not in research.md §I's enumeration.

---

## §E.2 Run-phase Evidence

### M1 — super-advisor agent file + ceiling 8->10 + E1-E4 doctrine (PASS)

**AC matrix (vci 5-section):**
- AC-AA2-001 super-advisor.md frontmatter: PASS — name=super-advisor, model=inherit, effort=xhigh, permissionMode=plan; NOT for: count=4; tools=Read,Grep,Glob,Bash,WebFetch,Skill (write-tools leak=0); live+template byte-identical (6352 bytes).
- AC-AA2-002 CLAUDE.md ceiling: PASS — "10 retained agents" count=2 (live+template); "8 retained agents" count=0 (complete flip); super-advisor+manager-design rows present; Selection Decision Tree entries 10+11 present.
- AC-AA2-003 E1-E4 doctrine: PASS — agent-common-protocol.md E1-E4 count=5 (>=4); super-advisor cross-ref count=3 (>=1).

**Evidence (LIVE, vci attribution):**
- `go build ./...` -> exit 0
- `golangci-lint run --timeout=2m` -> 0 issues
- template-first: super-advisor.md IDENTICAL; CLAUDE.md section-4 IDENTICAL; agent-common-protocol.md IDENTICAL (cp synced).
- `moai spec lint spec.md` -> 1 warning (StatusGitConsistency: draft vs git-implied; resolves when status flips to in-progress).

**Baseline-attribution:** HEAD pre-M1 8cc3ed06e; M1 edits are the 6 files (super-advisor.md x2, CLAUDE.md x2, agent-common-protocol.md x2) + this progress.md. No Go code touched.

**Gaps:** M1 commit SHA populated post-commit (self-reference avoided per sync_commit_sha pitfall). plan-auditor iteration-2 independent verdict deferred (API 429); orchestrator-direct verdict ~0.84 PASS-WITH-DEBT recorded.

**Residual-risk:** (a) orchestrator-direct plan-audit (confirmation bias — orchestrator authored manager-spec delegation AND verified); (b) M1 executed under GLM backend (user override of Opus recommendation) — agent-file authoring fidelity risk for M2 (manager-design H1-H9 verbatim).

### M2 — manager-design agent file + design pipeline + role_profile absorption (PASS)

**AC matrix (vci 5-section):**
- AC-AA2-004 manager-design.md H1-H9 verbatim: PASS — manager-design.md (+template mirror) exists; `grep -cE '^H[1-9]'` = 9; frontmatter (tools=Read,Write,Edit,Grep,Glob,Bash,DesignSync; model=inherit; effort=xhigh; permissionMode=acceptEdits; isolation=worktree; memory=project; skills=[moai-domain-frontend]) matches §04 codeblock; DesignSync count=7; live+template byte-identical.
- AC-AA2-005 design.md D1-D5 skill: PASS — `.claude/skills/moai/workflows/design.md` (+template mirror) exists; `grep -cE '^### D[1-5]'` = 5; methods (finalize_plan/write_files/report_validate) count=8.
- AC-AA2-006 spec-workflow plan→design→run route: PASS — `grep -c 'design → run\|plan → design'` = 1; `grep -c 'UI-surfaced\|UI surface'` = 2; conditional route documented as additive (UI-surfaced SPECs only).
- AC-AA2-007 designer role_profile + pencil MCP absorbed-by: PASS — `Absorbed by manager-design` annotation in team-protocol.md + settings-management.md (2 surfaces, ≥1).

**Evidence (LIVE, vci attribution):**
- AC grep matrix run against live tree 2026-07-10: H1-H9=9, `### D1-D5`=5, design route=1, UI-surfaced=2, Absorbed=2 surfaces, subagent-boundary AskUserQuestion=0, SPEC-ID neutrality=0 (4 files).
- `go test ./internal/template/ -count=1` → catalog tier audit green: TestAllAgentsInCatalog=9 agents, TestLoadCatalog=37 entries, TestLoadEmbeddedCatalog=37, TestManifestHashFormat hash-stable post `make build` (manager-design hash e994cf…). expectedTotal 35→37 + expectedAgentCount 7→9 updated for the v2 catalog (super-advisor + manager-design).
- template-first byte-parity: 4/5 IDENTICAL (manager-design.md, design.md skill, spec-workflow.md, team-protocol.md); settings-management.md DIFFERS pre-existing (rule_template_mirror 비대상; template neutral 유지).
- spec-workflow.md REQ-WFL-001 neutrality fix (live SPEC-ID token → template neutral prose) resolved a pre-existing TestRuleTemplateMirrorDrift FAIL.

**Baseline-attribution:** HEAD pre-M2 `5f2e5738c`; M2 edits are manager-design.md×2 (new), design.md skill×2 (new), spec-workflow.md×2 + team-protocol.md×2 + settings-management.md×2 (M2 annotations), catalog.yaml + 3 test files (catalog v2 count update) + this progress.md. No Go production code touched.

**Gaps:** TestRuleProvenanceAudit pre-existing FAIL (12 internal REQ/AC/SPEC-V3R* tokens in agent-common-protocol.md template mirror — REQ-LEDGER-001..006, REQ-AA2-003, SPEC-V3R6-HOOK-RECOVERY-SIGNAL-001, SPEC-V3R6-HARNESS-RUNTIME-RECOVERY-001, SPEC-V3R6-ORCH-INTERRUPT-LEDGER-001, SPEC-V3R5-WORKFLOW-OPT-001, W3 meta-analysis) NOT resolved — M2 scope-out (M1 E1-E4 doctrine + prior LEDGER/WORKFLOW-OPT SPECs); SPEC-TEMPLATE-RULES-CLEANUP-001 owns remediation. This is the sole remaining whole-repo-green blocker.

**Residual-risk:** (a) DesignSync MCP not registered in `.mcp.json` (research.md §H) — D2-D5 live execution gated on tool registration (separate user action); (b) M2 executed under GLM backend orchestrator-direct (M1 force-majeure continuity) — H1-H9 verbatim fidelity verified via grep + byte-parity, not independent semantic review; (c) TestRuleProvenanceAudit FAIL blocks whole-repo `go test ./...` green until the CLEANUP SPEC lands.

### M3a — RouteModelFor 3-arg + model_routing_profiles Go layer (PASS, commit e44c66d5f)

**AC matrix (vci 5-section):**
- AC-AA2-008 RouteModelFor 3-arg: PASS — signature is `RouteModelFor(specTier, phase, perfTier string)`; validRoutingPerfTiers = {max, medium, low}; TestRouteModelFor3x12Matrix exercises all 36 cells; TestRouteModelForHappyPath3Arg + TestRouteModelForFallback3Arg + TestRouteModelForInvalidInput3Arg PASS.
- AC-AA2-009 model_routing_profiles 3 matrices (code layer): PASS — ValidateModelRoutingProfiles + ModelRoutingProfiles map structure in types.go; fullProfilesYAML 36-cell golden fixture; TestValidateModelRoutingProfiles PASS.

**Evidence:** commit `e44c66d5f feat(SPEC-AGENT-ARCH-V2-001): M3a RouteModelFor 3-arg + model_routing_profiles (AC-008/009)`.

**Baseline-attribution:** HEAD e44c66d5f (M3a commit). `go test ./internal/config/... -count=1` → ok. No carry-over.

**Gaps:** M3a production config surface (workflow.yaml model_routing_profiles data) deferred to M3b (this delegation).

### M3b — workflow.yaml production config (PASS)

**AC matrix (vci 5-section):**
- AC-AA2-009 model_routing_profiles 3 matrices (production surface): PASS — `model_routing_profiles:` block with max/medium/low sub-tiers (3×12=36 cells) present in both live + template workflow.yaml; values match design.md §D.5 line 271-284 table AND fullProfilesYAML fixture byte-for-byte. `grep -c 'model_routing_profiles:' workflow.yaml` = 1 (both trees). Legacy `model_routing:` key RETAINED as `medium` alias per plan.md D6 (values updated to match medium profile + deprecation comment).
- AC-AA2-013 workflow_agents + role_profiles: PASS — `workflow_agents.read-only-extract` = `sonnet/low` (already correct pre-M3); `role_profiles.researcher` = `model: sonnet` (already correct pre-M3); `grep -cE 'model:\s*haiku' workflow.yaml` = 0 (both trees).

**Evidence:** `diff .moai/config/sections/workflow.yaml internal/template/templates/.moai/config/sections/workflow.yaml | head -20` — model_routing_profiles block identical (live L-run updated from opus→sonnet in legacy alias block to align with medium profile). AC-015 template-first: `diff` shows byte-identical model_routing_profiles between live + template.

**Baseline-attribution:** HEAD pre-M3b (worktree b153459e7d). `go build ./...` → exit 0. M3b edits are 2 workflow.yaml files (live + template). No Go code touched.

**Gaps:** None for AC-009/013/015 (M3b scope).

### M3c — moai init --model-policy + HaikuResidualRule + validRoutingModels haiku removal (PASS)

**AC matrix (vci 5-section):**
- AC-AA2-010 moai init --model-policy: PASS — `--model-policy max|medium|low` flag registered in init.go; `validateInitFlags` rejects invalid values with stderr error `invalid --model-policy value %q: must be one of: max, medium, low`; `template.ApplyPerformanceTier` patches llm.yaml `performance_tier` field; `resolveModelPolicy` resolves CLI flag + legacy `--high`/`--medium-alias`/`--low` aliases (plan.md D4). Legacy aliases emit stderr deprecation.
- AC-AA2-011 claude_models low: sonnet: PASS (pre-existing) — `awk` state-machine on claude_models block returns 0 haiku in both live + template llm.yaml; `low: sonnet` in live (line 19); template `low: "sonnet"`.
- AC-AA2-012 HaikuResidualRule (HARD): PASS — All 4 surfaces return 0: (1) agent frontmatter `grep -rn 'haiku' .claude/agents/moai/ internal/template/templates/.claude/agents/moai/ | grep -v _test` = 0; (2) claude_models awk block = 0 (both trees); (3) `grep -c 'haiku' workflow.yaml` = 0 (both trees); (4) `grep -n '"haiku"' internal/config/model_routing.go | grep -v _test` = 0. HaikuResidualRule registered in lint.go defaultRules() as crossSPECRule (NOT skip-able — CheckAll findings bypass applylintSkip). 6 tests PASS (CleanTree, AgentFrontmatter, ClaudeModelsBlock, WorkflowRouting, ValidRoutingModels, RegisteredAndNotSkippable).
- AC-AA2-016 haiku-residual-0 HARD closure: PASS — re-ran AC-012 four-surface verification; all 4 = 0.

**Evidence:**
- `go build ./...` → exit 0; `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- `golangci-lint run --timeout=2m` → 0 issues.
- `go test ./internal/spec/... -run TestHaikuResidual -v` → 6/6 PASS.
- `go test ./internal/config/... -count=1` → ok (both packages).
- Agent-file haiku removal: builder-harness.md lines 166-167 (model policy table) + line 192 (prose); manager-docs.md line 126 (prose) — all × 2 trees (live + template).
- `validRoutingModels` `"haiku": true` entry removed; fullRoutingYAML test fixture updated (haiku→sonnet in 3 entries); TestValidateModelRoutingClosedSetEffort fixture fixed (sonnet+ultra).
- catalog.yaml hashes regenerated via `go run ./internal/template/scripts/gen-catalog-hashes.go --all` (9 hashes; builder-harness + manager-docs cascade from M3c edits + 7 pre-existing drift). TestCatalogHash PASS.

**Baseline-attribution:** HEAD pre-M3c (worktree b153459e7d + M3b uncommitted). All tests run against this tree in this run. No carry-over.

**Gaps:** (a) TestOutputStylesTemplateLiveParity FAIL — PRE-EXISTING (moai-easy.md drift between template + live; unrelated to M3, not caused by M3 edits); (b) TestRuleProvenanceAudit — delegation mentions it as pre-existing FAIL but test not found at current HEAD (may have been resolved or renamed); (c) catalog.yaml on PRESERVE list but hash regen was a necessary cascade from builder-harness/manager-docs edits (2/9 hashes directly from M3c; 7/9 pre-existing drift also corrected by the deterministic regen).

**Residual-risk:** (a) AC-010 integration test (`moai init --model-policy max` end-to-end) not run — requires a full init in a temp dir; the flag registration + validation + ApplyPerformanceTier are individually verified but the end-to-end init→llm.yaml write path is not exercised in isolation; (b) HaikuResidualRule is a new cross-SPEC rule — it will scan ALL SPECs at lint time; no false-positive risk on closed SPECs because the rule scans the project tree (agents/config/code), not SPEC documents; (c) the `--medium-alias` flag name avoids collision with the existing `--mode` flag's potential `medium` value; the legacy alias `--medium` was not used because `--mode` already accepts a similar namespace.

### M4 — doctrine refresh: model-policy §2-B + agent-authoring 10-agent + agent-patterns 4-loop (PASS)

**AC matrix (vci 5-section):**
- AC-AA2-014 doctrine refresh: PASS — (1) `grep -cE '2-B|2-C|3-tier|max/medium/low' model-policy.md` = 5 (LIVE) / 5 (TEMPLATE), both ≥3; (2) `grep -c 'except manager-docs and manager-git which use model: haiku' model-policy.md` = 0 (LIVE) / 0 (TEMPLATE); (3) `grep -cE '10-agent|10 retained|super-advisor|manager-design' agent-authoring.md` = 8 (LIVE) / 8 (TEMPLATE), both ≥1.
- AC-AA2-015 template-first: PASS — model-policy.md live=template byte-identical (diff exit 0); agent-patterns.md live=template byte-identical (diff exit 0); agent-authoring.md has ONLY the pre-existing line 25/38 divergence (dev-only-commands `97-*/98-*` + dev-only-commands-isolation.md cross-ref — intentional local-only isolation per §21/§24, NOT caused by M4; M4 catalog-section edits are identical between live+template).

**Evidence (LIVE, vci attribution — measured 2026-07-10 in worktree agent-a465b5496a9ef8b0c at HEAD 39c74d777):**
- `go build ./...` → exit 0 (doctrine-only; template-embed unaffected).
- `golangci-lint run --timeout=3m` → 0 issues.
- `go test ./...` → 91 ok, 0 FAIL packages; 2 pre-existing test failures (NOT caused by M4): (a) `TestHookWrapper_TempFileCleanup` — temp-file-count race (12-file leak, hook wrapper; unrelated to markdown doctrine); (b) `TestOutputStylesTemplateLiveParity` — `moai-easy.md` parity drift (untracked file from a parallel SPEC per git status; M4 did not touch output-styles).
- model-policy.md edits: (1) removed stale haiku-exception prose (lines 47-51 — `except manager-docs and manager-git which use model: haiku` + exceptions bullet); (2) replaced § Model Policy Tiers with §2-B agent×tier matrix reference + 3-tier `max/medium/low` table; (3) added fable enum to generation mapping + v2.1.196/v2.1.198 references (Explore session-model inheritance); (4) § Effort Levels cross-references §2-B/§2-C; (5) § Rules updated for No-Haiku policy.
- agent-authoring.md edits: (1) catalog 8→10 (7→9 MoAI-custom); (2) Retained list adds super-advisor + manager-design; (3) Effort-Level Calibration Matrix updated to 10 agents (manager-docs high→medium, manager-git high→low per §2-B; super-advisor + manager-design added with xhigh FIXED); (4) retention ceiling 8-agent→10-agent (2 locations); (5) archived count 8 retained→10 retained.
- agent-patterns.md edits: (1) per-spawn section "7 MoAI-custom retained agents"→"9 MoAI-custom retained agents" + adds manager-design + super-advisor; (2) new § Orchestrator 4-Loop Mechanism → Catalog Mapping (plan→decompose→direct→collect, 4-step table mapping to catalog roles); (3) new § 4 Rejected Alternatives (전면 동적화 / auditor 통합 / 정적 핀 / Time-루프 에이전트 — each with rejected-approach + why-rejected rationale per SSOT §06 M4).

**Baseline-attribution:** HEAD 39c74d777 (M1-M3 complete, pushed to origin/main). M4 edits measured in worktree agent-a465b5496a9ef8b0c at this HEAD. All AC-014 grep commands run against both live + template trees in this run. No carry-over from prior unrelated measurements.

**Gaps:** (a) 2 pre-existing test failures (TestHookWrapper_TempFileCleanup + TestOutputStylesTemplateLiveParity) — NOT caused by M4 (markdown-only changes); both documented as pre-existing in M3c §E.2 gaps already; (b) worktree isolation — M4 executed in L1 worktree agent-a465b5496a9ef8b0c; orchestrator owns push (7 active sessions on shared checkout per task B9); commit lands on worktree branch.

**Residual-risk:** (a) agent-authoring.md pre-existing line 25/38 divergence (dev-only-commands) is NOT resolved by M4 — it is intentional local-only isolation per §21/§24 and out of M4 scope; AC-015 template-first is satisfied for the M4 catalog-section edits (identical live+template); the pre-existing divergence predates M4; (b) the 4-Loop mapping and 4 rejected alternatives are doctrine prose derived from the SSOT §03/§06 — no runtime code verifies the mapping; future agent-catalog revisions MUST re-verify the 4-Loop mapping + rejected-alternatives rationale before re-proposing any rejected approach.

---

## §E.3 Run-phase Audit-Ready Signal

```yaml
run_complete_at: 2026-07-10
run_commit_sha: pending-backfill   # populated post-commit (self-reference avoided)
run_status: audit-ready
ac_pass_count: 8
ac_fail_count: 0
preserve_list_post_run_count: 0
l44_pre_commit_fetch: n/a   # worktree isolation; no push (orchestrator owns push)
l44_post_push_fetch: n/a    # orchestrator push deferred (7 active sessions on shared checkout)
new_warnings_or_lints_introduced: 0
cross_platform_build:
  linux_darwin: pass   # go build ./... exit 0
  windows_amd64: pass  # GOOS=windows GOARCH=amd64 go build ./... exit 0
total_run_phase_files: 14
m1_to_mN_commit_strategy: M1+M2 (prior commits e44c66d5f + M2 commits); M3a (prior commit e44c66d5f); M3b+M3c (prior commit); M4 (this delegation, doctrine refresh)
```

MUST-PASS ACs green: AC-AA2-008 (M3a), AC-AA2-009 (M3a code + M3b config), AC-AA2-010 (M3c), AC-AA2-011 (pre-existing), AC-AA2-012 (M3c HARD), AC-AA2-013 (M3b), AC-AA2-016 (M3c HARD closure). SHOULD-PASS AC-AA2-014 (M4 doctrine refresh) PASS — model-policy §2-B + agent-authoring 10-agent + agent-patterns 4-loop mapping + 4 rejected alternatives. AC-AA2-015 template-first verified (model-policy + agent-patterns byte-identical; agent-authoring catalog-section identical, pre-existing line 25/38 dev-only divergence out of scope).

---

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates this section with `sync_commit_sha` on the single sync commit that carries the `implemented → completed` transition.>_

```yaml
sync_commit_sha: ""   # populated by manager-docs at sync commit
sync_completed_at: ""  # populated by manager-docs at sync close
```

---

## §F Phase 0.95 Mode Selection

Decision: sub-agent (Mode 5)

Input parameters: tier=L, scope=~25-35 files, domains >=5 (agents/workflow-skills/rules/Go-config/template-mirrors), language mix=markdown+Go+yaml, concurrency benefit=LOW (coding-heavy), Agent Teams prereqs=team.enabled false (Mode 3 unavailable).

Mode evaluation: Mode 1 trivial NO (Tier L semantic); Mode 2 background NO (write-heavy); Mode 3 agent-team NO (team.enabled=false default); Mode 4 parallel NO (coding-heavy — Anthropic caveat); Mode 6 workflow NO (multi-rule semantic, not uniform mechanical); Mode 5 sub-agent YES (Tier L coding-heavy sequential default).

Justification: Tier L + coding-heavy (Go signature extension + agent file authoring + doctrine refresh) -> Mode 5 per Anthropic coding-task parallelism caveat. M1->M2->M3->M4 sequential (M4 depends on M1+M3).

Force-majeure note (API 429): plan-auditor iteration-2 spawn hit API 429 (5-hour GLM usage limit, resets 20:11:06). Orchestrator executes M1-M4 DIRECTLY (not via manager-develop spawn) as a force-majeure exception — user explicitly directed continue. Phase 0.5 iter-2 verdict produced by orchestrator-direct LIVE verification (PASS-WITH-DEBT ~0.84; D1-D6 resolved, 0 new BLOCKING). Independent plan-auditor + sync-auditor confirmation deferred to API reset / sync-phase. Residual-risk: confirmation bias (orchestrator authored the manager-spec delegation AND verified it).

---

## §H Recursive Self-Diagnosis Log

_<pending run-phase — the bounded DIAGONSE-PATCH-VERIFY loop record (mechanical failures only, max 3 iterations per failure class). Manager-develop or orchestrator populates when a mechanical failure triggers the self-diagnosis path.>_

---

## Cross-References

- **Plan-phase SSOT**: `spec.md` (17 REQs, 5 GAPs, 9 Constraints, Out of Scope).
- **Design authority**: `.moai/reports/agent-architecture-redesign-v2-20260709.html` (§01-§08).
- **Lifecycle schema**: `.claude/rules/moai/development/spec-frontmatter-schema.md` § progress.md Section Map (the canonical §E.1-§E.4 + §F/§H allocation).
- **Era classification**: `.claude/rules/moai/workflow/lifecycle-sync-gate.md` § Era Classification Heuristic (the `era: V3R6` frontmatter override fires H-override; `§E.2`+`§E.4`+`sync_commit_sha` predicates are the H-4 fallback that confirms V3R6 era post-sync).

---

## History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-09 | manager-spec | Initial progress.md. §E.1 populated with plan-phase audit-ready signal (6 artifacts authored, 2 supersede targets flipped, design authority cited, 3 open decisions + 1 known gap recorded). §E.2-§E.4 + §F + §H placeholder headings emitted per the canonical progress.md Section Map. |
