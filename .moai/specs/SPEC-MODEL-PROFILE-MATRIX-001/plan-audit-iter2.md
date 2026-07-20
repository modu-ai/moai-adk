# SPEC Review Report: SPEC-MODEL-PROFILE-MATRIX-001
Iteration: 2/3
Verdict: PASS
Overall Score: 0.93 (Tier L threshold 0.85 — PASS; iter-1 was 0.69 → monotonic +0.24, no score regression)

> Reasoning context (the task-brief summary of claimed fixes) was used ONLY as a
> fix-checklist pointer; all verdicts below are grounded in the six SPEC artifacts
> + live working-tree evidence. Matrix A cell decisions are treated as user-settled
> design input and were NOT re-litigated.

## Must-Pass Results

- **[PASS] MP-1 REQ number consistency**: `grep -oE '^\*\*REQ-MPM-[0-9]{3}'` → exactly 40 unique REQs, REQ-MPM-001…040, no gaps, no duplicates, uniform 3-digit zero-padding. (Cosmetic-only: REQ-MPM-040 is physically placed in §B.4 between REQ-022 and REQ-023 rather than at numeric end — this is out-of-numeric-order *placement*, not a gap/dup; MP-1 concerns set completeness + padding, which hold.)
- **[PASS] MP-2 GEARS format compliance**: All 40 REQs use GEARS patterns — Ubiquitous (REQ-001 "The `llm` … **shall** carry"), Event-driven (REQ-003 "**When** the config loader reads … **shall** ignore"; REQ-040 "**When** a web settings save updates … the writer **shall** persist"), Where/capability-gate (REQ-002 "**Where** `llm.profile` is absent … **shall** resolve"), Unwanted/negative (REQ-037/038/039/040 "**shall not**"). acceptance.md ACs are Given-When-Then test scenarios explicitly labeled "Format: Given-When-Then" (spec.md §B is the GEARS requirements layer, acceptance.md is the G/W/T verification layer — the standard MoAI two-layer structure; the ACs are NOT mislabeled as GEARS, so no MP-2 penalty per the §B-REQs-are-GEARS carve-out).
- **[PASS] MP-3 YAML frontmatter validity**: all 12 canonical fields present with correct types — `id`, `title`, `version: "0.1.1"` (quoted semver), `status: draft` (valid enum), `created`/`updated: 2026-07-20` (canonical names, NOT `created_at`/`updated_at` aliases), `author`, `priority: P1`, `phase: "v3.1.0 target"`, `module`, `lifecycle: spec-anchored`, `tags` (comma-separated, NOT `labels`). Optional `tier: L` + `related_specs` present. (Note: `id: SPEC-MODEL-PROFILE-MATRIX-001` is multi-segment, violating the schema-doc single-segment regex `^SPEC-[A-Z][A-Z0-9]+-[0-9]{3}$` — but this is the repo-wide convention and `FrontmatterSchemaRule` checks field presence/emptiness, NOT the regex; NOT an MP-3 failure per established practice, schema-doc-vs-practice divergence at most.)
- **[PASS/N-A] MP-4 language neutrality**: N/A — single-project-scoped internal config (`internal/config`, `internal/template`, `internal/web`, `internal/cli`). Matrix A names the *tool's own* Claude/GLM model aliases (fable/opus/sonnet), NOT user-project 16-language dev-tooling. The one template-tree touch (`internal/template/templates/.moai/config/sections/llm.yaml`) is the model-routing config, which is language-agnostic and hardcodes no language-specific tool (no gopls/pylsp/rust-analyzer etc.). Auto-passes.
- **[PASS] MP-5 D7 cross-SPEC reconciliation**: 3 referenced SPECs resolved live — `SPEC-MODEL-TIER-PLANTYPE-001 → status: completed`, `SPEC-WEBCONF-SIMPLIFY-001 → status: completed`, `SPEC-GLM-EFFORT-TUNE-001 → status: completed`. None retired/superseded/archived → no BLOCKING finding. (SPEC-MODEL-TIER-PLANTYPE-001 is the axis this SPEC replaces; it is `completed`, not `superseded`, and the supersession is explicitly narrated in §A.1/§E — no dangling reconciliation.)
- **[PASS] MP-6 D8 cross-platform discipline**: `grep -c syscall spec.md` → 0. No syscall introduction → D8 auto-PASS, no `//go:build` obligation.
- **[PASS] MP-7 clarification gate**: `grep -rn '\[NEEDS CLARIFICATION' plan.md research.md` → 2 hits, both backtick-quoted PROSE confirming the markers were *resolved* ("Both former `[NEEDS CLARIFICATION]` markers are already resolved as DECISION-001 / DECISION-002"). Neither is a live `[NEEDS CLARIFICATION: <topic>]` marker. The two iter-1 markers (effort-injection capability, default profile) are settled as DECISION-001/002 in plan.md §E with rationale. No unresolved markers at audit time.

## Category Scores (0.0-1.0, rubric-anchored)

| Dimension | Score | Rubric Band | Evidence |
|-----------|-------|-------------|----------|
| Clarity | 0.90 | 1.0 band (single unambiguous interpretation; one minor nuance) | REQ-MPM-025/026/027 precisely partition model-vs-effort consumption; DECISION-001/002 remove prior ambiguity. Minor: REQ-027's divergence example (spec.md:140) cites `manager-spec` (plain xhigh), not the two `xhigh (FIXED)` agents (super-advisor/manager-design, agent-authoring.md:344/349) whose GLM-overlay reporting differs — a small under-specification (see D1). |
| Completeness | 0.95 | 1.0 band | All sections present: HISTORY, §A Context/Goal (WHY), §B 40 REQs (WHAT), §C Success Criteria, §D 6× `### Out of Scope — <topic>` H3 subheadings (spec.md:186/189/192/195/198/201) each with a specific `-` bullet, §E References. Tier L 5-file set complete: design.md (§A-§G) + research.md (§A-§F, command+output evidence blocks) authored (D3 resolved). |
| Testability | 0.90 | 1.0 band | ACs carry concrete verifiable commands: AC-018/AC-025 grep predicates, AC-020 `make build` + `go test`, AC-005 exact Matrix A cell values, AC-019 exact GLM collapse (`low→thinking-off`, coding-max→`reasoning-max`) verified against live `CollapseClaudeEffortToGLM`. Minor: AC-016/017 "documented intent" is a semantic annotation, but the mechanical assertions (resolver `--json` output equals Matrix A; lint-clean; frontmatter unmutated) are binary-testable. No weasel words (appropriate/adequate/reasonable) found. |
| Traceability | 0.95 | 1.0 band | Every REQ-MPM-001…040 covered by ≥1 AC (full map verified). Every AC traces to valid REQ(s); AC-MPM-025→REQ-040 (new). Sole exception: AC-MPM-024 cites "HaikuResidualRule + CI guards" not a REQ — a deliberate invariant-preservation/CI-guard AC (relates to §C + REQ-023/037), an accepted pattern, not a broken orphan. |

Aggregate (harmonic mean of the four dimensions) ≈ 0.935; arithmetic mean 0.9375. Reported as **0.93**, well above the Tier L 0.85 threshold.

## Defects Found (structured defect-list)

D1. **fixed-effort-GLM-reporting-underspecified** — spec.md:140 (REQ-MPM-027) / acceptance.md:97 (AC-MPM-017) — Severity: **minor** — REQ-MPM-027's "profile effort = documented intent, frontmatter effort remains effective for the named-subagent spawn" is correct and implementable, but its worked example uses `manager-spec` (plain `xhigh`). The two agents declared `xhigh (FIXED)` in agent-authoring.md §Effort-Level Calibration Matrix (super-advisor L349, manager-design L344) get a *lower* GLM-overlay-reported effort under Matrix A (e.g. super-advisor profile `max` → `medium` → `CollapseClaudeEffortToGLM` = `reasoning-high`, vs frontmatter `xhigh` → `reasoning-max`). This is a *resolver-reporting* divergence, honesty-bounded by REQ-MPM-039 ("live-validation pending") and non-contradictory at the named-subagent Claude spawn (frontmatter `xhigh` stays effective per REQ-027). NON-BLOCKING. Required fix (run-phase, optional): add one sentence to REQ-MPM-027 or design.md §C.2 noting that for the FIXED-xhigh agents the resolver's GLM-collapsed profile effort is a *reported/documented-intent* value that does NOT lower the named-subagent spawn effort below the frontmatter FIXED `xhigh`. Does not require plan-artifact rework to proceed.

D2. **REQ-040-out-of-numeric-order** (cosmetic) — spec.md:130 — Severity: **minor** — REQ-MPM-040 is authored inside §B.4 physically between REQ-022 and REQ-023 (numeric end would be after REQ-039). Does not affect MP-1 (set complete, no gaps/dups). Required fix (optional): none required; if desired, a HISTORY note already explains 040 was the iter-2 web-save-path addition. Leave as-is is acceptable.

(No critical or major defects. All four iter-1 FAIL drivers are resolved — see Regression Check.)

## Chain-of-Verification Pass

Second-look findings: none new that change the verdict. Verified by re-reading:
- **D1 fix completeness (spec.md §A.4 + REQ-025/026/027 + AC-016/017)** — `grep -niE "effort.*(override|inject).*(spawn|frontmatter)|per-spawn effort"` returned only NEGATION statements ("does **NOT** accept a per-spawn `effort`", "**NOT** a per-spawn override", "the injection channel therefore carries model only"). The iter-1 unimplementable "effort overrides frontmatter at spawn" claim is fully removed.
- **All 4 ApplyTierProfile call sites (live grep)** — initializer.go:195, update.go:486, update.go:1447, web/agentfm.go:108 all confirmed present; agentfm.go doc comment (lines 74-83) confirms it "re-applies the {model, effort} tier profile to the shipped agent .md frontmatter" — the iter-2 "web save-path mutates frontmatter" premise is factually correct (D2 correction verified, not a fabrication).
- **research.md evidence blocks** — spot-checked §A (4 call sites), §C.2 (template `model: inherit` vs local `model: opus`), §C.3 (DefaultModelPolicy:27 / DefaultPerformanceTier:77), §C.4 (template llm.yaml lines 9/15/22), §D (glm_effort_overlay symbols) — every cited command reproduces the claimed output against the live tree. Line numbers drift slightly (tierProfiles at 294 vs cited 297-315) but research.md self-discloses "line numbers are content-token anchors; symbol names are authoritative" — symbols all verified.
- **design.md ↔ spec.md consistency** — design.md §A schema matches REQ-001/006/009; §B group-membership matches REQ-011; §C injection split (model=spawn / effort=documented-intent) matches REQ-025/026/027; §D web change matches REQ-018..022/040; §F removal matches REQ-024/028. No design-vs-spec contradiction found.
- **Traceability end-to-end** — mapped all 40 REQs to ACs individually (not sampled); all covered.

## Regression Check (iter-1 → iter-2)

Defects from iteration 1 (FAIL 0.69):
- **D1 (effort-injection over-claim: REQ-025/027 asserted unimplementable per-spawn effort override)** — **RESOLVED**: REQ-MPM-025/026/027 + AC-MPM-016/017 rewritten to model-only spawn injection; effort reframed as documented intent (frontmatter default + GLM overlay + Workflow/`Agent(general-purpose)` prompt). Verified via CoV grep (all effort-at-spawn mentions are negations). DECISION-001 grounds it against the confirmed Agent-tool capability.
- **D2 (ApplyTierProfile call-site under-enumeration; web-save-path "display-only" premise wrong)** — **RESOLVED**: all 4 production call sites now enumerated in spec.md §A.4 + research.md §A + plan.md §B (verified live: initializer.go:195, update.go:486, update.go:1447, web/agentfm.go:108). REQ-MPM-040 + AC-MPM-025 added for the web save-path frontmatter-mutation retirement; agentfm.go doc comment confirms the mutation is real (not display-only).
- **D3 (Tier L design.md + research.md missing)** — **RESOLVED**: both authored. research.md carries per-finding command+observed-output evidence blocks (§A-§F); design.md is architecture §A-§G consistent with the REQs. Tier L 5-file contract satisfied (6 files incl. progress.md).
- **D4 / MP-7 (2 unresolved [NEEDS CLARIFICATION] markers)** — **RESOLVED**: both settled as DECISION-001 (Agent-tool per-spawn = model-only) and DECISION-002 (default profile = medium) in plan.md §E, dated, with rationale. Live grep confirms no unresolved markers.

Stagnation: none. All four defects moved from open→resolved in a single iteration; aggregate score rose 0.69→0.93 (monotonic increase → no STOP escalation, no scope-reduction needed).

## Recommendation

**PASS** — proceed to Implementation Kickoff Approval (the plan→run HUMAN GATE, mandatory and score-independent per CLAUDE.local.md §19.1). Rationale, per must-pass:
- MP-1 (40 REQs 001-040, no gaps/dups), MP-2 (all GEARS), MP-3 (12 canonical fields) — all cited above with live grep evidence.
- MP-5 (D7): 3 referenced SPECs all `completed`, no BLOCKING.
- MP-6 (D8): 0 syscall, auto-PASS.
- MP-7: no live clarification markers; iter-1 markers resolved as DECISION-001/002.
- All four iter-1 FAIL drivers verified resolved against the live working tree, not merely asserted.

Two minor non-blocking items to hand to the run phase (do NOT gate kickoff):
1. (D1) In run-phase, add one clarifying sentence that the FIXED-`xhigh` agents (super-advisor, manager-design) keep `xhigh` as the effective named-subagent spawn effort even though the resolver's GLM-overlay report shows the Matrix A profile effort collapsed — so an implementer does not accidentally lower those two agents below `xhigh`.
2. (D2) REQ-MPM-040's out-of-numeric-order placement is cosmetic; no action required.

This is a clean PASS (not PASS-WITH-DEBT): the only carried debt — GLM wire-effectiveness live-validation — is an explicit, honesty-bounded scope exclusion inherited from the completed SPEC-MODEL-TIER-PLANTYPE-001 (REQ-MPM-039, §D Out of Scope), not a plan defect requiring rework.
