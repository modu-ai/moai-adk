# Implementation Plan — SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001

> Tier M. 3-artifact set (spec.md + plan.md + acceptance.md) + progress.md skeleton.
> Mirrors the RULE-SCOPING-001 milestone shape (template-first → make build → live parity → verification).

---

## A. Context

- **Goal**: Apply Anthropic's "Write an effective CLAUDE.md" per-line test to CLAUDE.md's always-loaded body, reducing always-on token cost via the three REAL reduction mechanisms (M-DELETE / M-POINTER / M-SCOPE) while preserving every behavioral [HARD] directive and keeping the two trees in byte-parity.
- **Baseline**: both `CLAUDE.md` and `internal/template/templates/CLAUDE.md` are byte-identical (650 lines / 35778 bytes / 14 `[HARD]` / 14 `[ZONE:*]` / 2 `@import`). diff exit 0.
- **Higher risk than RULE-SCOPING-001**: that SPEC was frontmatter-only; THIS is body editing. The §C classification table below is precise enough that run-phase is mechanical.

### A.5 PRESERVE list (do NOT touch — B8/B10 working-tree hygiene)

The working tree carries UNRELATED residue. Run-phase MUST touch ONLY the 2 CLAUDE.md trees + this SPEC's 4 artifacts. Do NOT touch:
- modified `.moai/project/codemaps/*.md` (5 files)
- deleted `internal/evaluator/evaluator_test.go`
- untracked `.moai/specs/SPEC-CLEANUP-EVALUATOR-001/`, `.moai/specs/SPEC-SIMPLICITY-LADDER-001/`, `.moai/specs/SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001/`
- untracked `.moai/reports/*`, `.moai/design/web-console-handoff/`
- runtime-managed files (`.moai/state/*`, `.moai/cache/*`, `.moai/harness/*`)

---

## B. Known Issues (filtered to relevant categories per Tier M)

- **B4 Frontmatter canonical schema**: `created:`/`updated:`/`tags:` (no snake_case). 12 required + `tier: M` + `era: V3R6`. (Applied to this SPEC's spec.md.)
- **B6 spec-lint Heading**: Out-of-Scope uses `### Out of Scope — <topic>` h3 sub-headings with `-` bullets (applied in spec.md §D). A bare `## Out of Scope` h2 triggers `MissingExclusions` ERROR.
- **B8/B10 Working-Tree Hygiene + Scope Discipline**: §A.5 PRESERVE list. `git add` specific paths only.
- **B9 Git commit + push self-perform** (Hybrid Trunk): Conventional Commits, `feat(SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001): M{N} ...`, no `--no-verify`.
- **B11 AskUserQuestion prohibition**: subagent returns blocker report; never prompts user.
- NOT relevant (omitted): B1 cross-platform build tags (no Go change), B2 retired-SPEC scan (markdown only), B3/C-HRA-008 subagent-boundary grep (no `internal/harness` code), B5 CI 3-tier (no Go test impact), B7 observer.go path resolution (N/A).

---

## C. Pre-flight + the KEEP/CUT/POINTER classification table (CORE DELIVERABLE)

### C.1 Pre-flight baselines (run BEFORE any edit; record observed)

```bash
# 1. line + byte counts both trees
wc -l CLAUDE.md internal/template/templates/CLAUDE.md
wc -c CLAUDE.md internal/template/templates/CLAUDE.md

# 2. byte-parity (must be exit 0 at baseline AND after diet)
diff CLAUDE.md internal/template/templates/CLAUDE.md; echo "DIFF_EXIT=$?"

# 3. directive counts (the protected set)
grep -c '\[HARD\]' CLAUDE.md
grep -c '\[ZONE:' CLAUDE.md

# 4. @import inventory (structure-only; never counted as reduction)
grep -nE '^@[.]' CLAUDE.md

# 5. neutrality baseline (no internal artifacts pre-existing in body)
go test ./internal/template/ -run TestTemplateNeutralityAudit 2>&1 | tail -3 || true
```

Baseline observed (2026-06-22): 650/650 lines, 35778/35778 bytes, diff exit 0, 14 `[HARD]`, 14 `[ZONE:*]`, 2 `@import` (L331/L332).

### C.2 SSOT verification for POINTER candidates — the STRONGER prose-duplication bar (verification-claim-integrity.md C-7)

[HARD] **(D1 strengthening)** A POINTER classification is a DEFECT-CLAIM ("this CLAUDE.md prose is duplicated in an SSOT") and is therefore subject to verification-claim-integrity.md §1.1 — it MUST be tool-verified, not assumed. **Heading-presence is NOT sufficient.** The iter-1 §C.2 verified only that each SSOT file contained the named HEADING; a heading match does NOT prove the specific CLAUDE.md prose is actually carried there. The corrected bar:

> **A POINTER classification is valid ONLY when a grep confirms the SSOT rule file carries the ACTUAL DUPLICATED PROSE** — the distinctive load-bearing content of the CLAUDE.md section (its specific bullets, thresholds, names, or directives), NOT merely a section heading. When the distinctive content is UNIQUE to CLAUDE.md (0 SSOT prose hits), the section is GENUINELY UNIQUE — pointer-izing it would DELETE behavioral content, which is the core diet risk. Such a section MUST be classified KEEP, not POINTER.

This bar is encoded as a run-phase precondition (acceptance.md AC-CMD-009): before any POINTER edit, run-phase MUST re-run the prose-duplication grep for that section and confirm ≥1 hit of the DISTINCTIVE content; a 0-hit result blocks the POINTER edit and forces KEEP.

Re-audit against the stronger bar (command → observed; this tree, 2026-06-22 — DISTINCTIVE content, not heading):

| CLAUDE.md § | Distinctive content grepped in SSOT | Observed | Verdict |
|-------------|--------------------------------------|----------|---------|
| §7 Rule 5 Discovery | 4 trigger phrases (`Ambiguous pronouns`/`Multi-interpretable`/`Unclear boundaries`/`conflict with existing`) in askuser-protocol.md | `3`+ hits ✓ (all 4 triggers present) | **POINTER** confirmed (genuinely duplicated) |
| §8 User Interaction | `select:AskUserQuestion`/`max 4 questions`/`First option MUST` in askuser-protocol.md | `9` hits ✓ | **POINTER** confirmed |
| §10 Web Search | `WebSearch`/`WebFetch`/`verify each URL`/`Sources:` in moai-constitution.md | `3` hits ✓ | **POINTER** confirmed |
| §11 Error Recovery bullets (Token-limit/Permission/MoAI-ADK) | `Token limit`/`Permission error` in agent-common-protocol.md | **`0` hits** | **DEMOTE → KEEP** (unique recovery bullets; §11 only the EXISTING pointer-line `> Canonical rule:` stays) |
| §11 Resumable Agents (`agentId`) | `agentId`/`Resumable` in agent-common-protocol.md | **`0` hits** (D1) | **DEMOTE → KEEP** (unique) |
| §13 Progressive Disclosure | `Level 1`/`Level 2`/`67%`/`skillListingBudget` in skill-authoring.md | `5` hits ✓ | **POINTER** confirmed |
| §14 File-Write-Conflict / Loop-Prevention bullets | `File Write Conflict`/`Loop Prevention`/`Background.*Agent` in moai-constitution.md | **`0` hits** | **DEMOTE → KEEP** (unique operational bullets; Background-Agent is also `[HARD]`) |
| §14 Worktree Isolation Rules subsection | `Worktree Isolation`/`isolation.*worktree` in worktree-integration.md (the cited SSOT) | `32` hits ✓ | **POINTER** confirmed (only THIS subsection) |
| §15 Activation/Mode/Team-API (non-CG) | `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`/`SendMessage`/`TeammateIdle` in spec-workflow.md | `5` hits ✓ | **POINTER** confirmed (non-CG portion only) |
| §15 CG-Mode subsection (`moai cg`/tmux diagram) | `moai cg` in spec-workflow.md | **`0` hits** (D1) | **DEMOTE → KEEP** (unique cost-optimization content) |
| §16 When-to-Search/Process + `150,000` threshold | `150,000`/`Search Process` in context-window-management.md | **`0` hits** (`150,000`) + 1 weak (D1) | **DEMOTE → KEEP** (unique process + unique threshold; §16 only the EXISTING pointer-line stays) |

**Net (D1 reclassification)**: 5 sections/sub-sections DEMOTED from POINTER → KEEP because their distinctive content is UNIQUE (0 SSOT prose hits): §11 Error-Recovery bullets, §11 Resumable Agents, §14 File-Write-Conflict/Loop-Prevention/Background-Agent bullets, §15 CG-Mode, §16 process+threshold. The surviving POINTER set (§7, §8, §10, §13, §14-Worktree-subsection, §15-non-CG) is each confirmed to carry its DISTINCTIVE content in the SSOT, not merely a heading.

### C.3 The per-section KEEP / CUT / POINTER classification table

Disposition codes: **KEEP** (behavioral; removal causes mistakes), **CUT** (M-DELETE; removal causes no mistakes), **POINTER** (M-POINTER; collapse explanation to a 1-line cross-ref, PRESERVE unique [HARD] directives), **TRIM** (partial M-DELETE within an otherwise-KEEP section). Already-pointer-ized sections (those that already open with a `> Canonical rule:` line) only need their residual restated prose trimmed.

| § | Title | Lines | Disposition | Rationale (per-line test) |
|---|-------|-------|-------------|----------------------------|
| (title) | `# MoAI Execution Directive` | L1 | KEEP | identity |
| 1 | Core Identity + HARD Rules | L3-29 | **KEEP** | The 11 `[HARD]` bullets are the canonical always-loaded directive list — removal causes mistakes. KEEP verbatim. (Recommendations L23-27 = TRIM candidate: low behavioral value, but tiny; default KEEP.) |
| 2 | Request Processing Pipeline | L31-72 | **TRIM** | Phase 1-4 prose is largely explanatory; the Skill() load list (L43-45) + routing categories (L51-55) are useful but verbose. TRIM the Phase 3 examples (L61-63 duplicate §5 chain). Keep the 4-phase skeleton. |
| 3 | Command Reference | L74-93 | **TRIM** | Subcommand list (L80) is useful; the "Allowed Tools" line (L83) duplicates skill frontmatter. TRIM redundant prose, keep the command inventory. |
| 4 | Agent Catalog | L96-145 | **KEEP + REQ-CMD-008 fix** | The 8-agent table + archived-agent list ARE behavioral (spawn-rejection depends on them). KEEP. The "Watch (2.1.172)" nesting note (L100) gets the CONTENT-CORRECTNESS fix (REQ-CMD-008), NOT a token cut. |
| 5 | SPEC-Based Workflow | L147-185 | **TRIM/POINTER** | Command Flow (L153-155) + Agent Chain (L161-166) are behavioral KEEP. MX Tag Integration (L168-184) duplicates `mx-tag-protocol.md` (already cited L182) → POINTER (collapse the MX type list to the existing cross-ref). |
| 6 | Quality Gates | L188-213 | **POINTER** | Already cites `moai-constitution.md` (L190) + config paths. Harness/LSP prose duplicates `harness.yaml`/`quality.yaml` config + spec-workflow. POINTER: keep harness level names + config paths, collapse explanation. |
| 7 | Safe Development Protocol | L217-303 | **POINTER (heavy)** | Rules 1-5 (L223-293) — Rule 5 Discovery (L256-293) duplicates `askuser-protocol.md` § Ambiguity Triggers (verified). Rules 1-4 are behavioral but their detail duplicates the §1 HARD bullets. POINTER: keep the 5 rule NAMES + the §1 HARD-bullet binding, collapse the expanded prose to a cross-ref to askuser-protocol.md + this section's own summary. Language-Specific Guidelines (L295-303) = KEEP (small, behavioral toolchain list). **Largest single diet target.** |
| 8 | User Interaction Architecture | L307-323 | **POINTER** | Already cites `askuser-protocol.md` (L313) + `agent-common-protocol.md` (L320). The 2 `[ZONE:Frozen][HARD]` lines (L309, L311) are UNIQUE behavioral directives → KEEP. Collapse the duplicated "Key rules"/"Agent interaction boundary" bullet explanations to the existing cross-refs. |
| 9 | Configuration Reference | L327-361 | **TRIM** | The 2 `@import` lines (L331-332) are STRUCTURE-ONLY (REQ-CMD-004) — KEEP but NOT counted as reduction. Design System Config list (L344-353) + Language Rules (L355-361) duplicate config files → TRIM to pointers. |
| 10 | Web Search Protocol | L364-389 | **POINTER** | Already cites `moai-constitution.md` (L366) + `glm-web-tooling.md` (L380). Execution Steps + Prohibited Practices duplicate the constitution anti-hallucination section → POINTER. Deep Research (L382-388) duplicates `dynamic-workflows.md` → POINTER. |
| 11 | Error Handling | L392-408 | **KEEP (D1 demoted)** | ALREADY self-declares "high-level overview" + cites `agent-common-protocol.md` § Error Recovery Pattern (L394). BUT the Error Recovery bullets (Token-limit / Permission / MoAI-ADK recovery, L398-402) + Resumable Agents (`agentId`, L404-408) are UNIQUE — 0 prose hits in the SSOT (§C.2 re-audit). Pointer-izing would DELETE behavioral content. KEEP the section verbatim; the EXISTING `> Canonical rule:` pointer-line (L394) already stays. Only safe action: the archived-agent recovery bullet (L398) may TRIM redundancy with §4, but the recovery mechanism stays. |
| 12 | MCP Servers & Deep Analysis | L412-422 | **TRIM/POINTER** | ultrathink/Adaptive Thinking prose duplicates `moai-workflow-thinking` skill (cited L417) + `settings-management.md` (cited L422). POINTER. Keep the keyword names. |
| 13 | Progressive Disclosure | L426-442 | **POINTER** | ALREADY cites `skill-authoring.md` (L428, verified). Level 1-3 prose + Benefits duplicate the SSOT → collapse to the existing pointer. |
| 14 | Parallel Execution Safeguards | L446-466 | **SPLIT (D1 refined)** | Already cites `moai-constitution.md` (L448). The constitution § Parallel Execution carries only the GENERAL principle — the operational bullets File-Write-Conflict / Loop-Prevention (L450-453) + the `[ZONE:Frozen][HARD]` Background-Agent line (L455) are UNIQUE (0 SSOT prose hits, §C.2 re-audit) → **KEEP**. Only the Worktree Isolation Rules subsection (L457-466) carries distinctive prose duplicated in `worktree-integration.md` (32 hits, cited L466) → **POINTER** (collapse to the existing cross-ref). Iter-1's wholesale "§14 POINTER" was over-aggressive. |
| 15 | Agent Teams (Experimental) | L470-540 | **SPLIT (D1 refined)** | Already cites `spec-workflow.md` + `workflow.yaml` (L500). The Activation / Mode-Selection / Team-API / Team-Hook / Dynamic-Team prose (L470-500) carries distinctive content duplicated in `spec-workflow.md` § Agent Teams Variant (5 hits, §C.2) → **POINTER**. BUT the CG-Mode subsection (L502-534, `moai cg`/tmux ASCII diagram + when-to/when-not-to) is UNIQUE — 0 `moai cg` hits in spec-workflow.md (§C.2 re-audit, D1) → **KEEP**. The Dynamic-Workflows subsection (L536-540) already cites `dynamic-workflows.md` → POINTER. Pointer-ize the non-CG team prose; KEEP the CG-Mode cost-optimization content. |
| 16 | Context Search Protocol | L544-589 | **KEEP (D1 demoted)** | ALREADY cites `context-window-management.md` + `session-handoff.md` (L546). BUT the distinctive content — the When-to-Search / When-NOT / 6-step Search Process / Token-Budget (`150,000` threshold, `5,000` injection cap) / Manual-Trigger / Integration-Notes (L550-589) — is UNIQUE: 0 hits for `150,000` and only 1 weak hit for the process in context-window-management.md (§C.2 re-audit, D1). The cited SSOT covers `/clear` THRESHOLDS, NOT the cross-session SEARCH protocol. Pointer-izing would DELETE the search protocol. KEEP the section; the EXISTING `> Canonical rule:` pointer-line (L546) already stays. |
| 17 | Troubleshooting | L593-630 | **TRIM** | Debug commands (L599-608) + Common Issues table (L614-619) are mildly useful but rarely needed always-on. TRIM: keep the Common Issues table (behavioral lookup), CUT the verbose debug-flag prose + Reading-Large-PDFs (L621-630, duplicates Read tool docs). |
| footer | Version + Changes | L634-650 | **CUT (changelog) / KEEP (identity)** | `Version:`/`Language:`/`Core Rule:` (L634-636) = KEEP (system identifier, neutral). "Changes in v14.2.0…"/"Changes in v14.1.0…" (L638-648) = **CUT** (M-DELETE — changelog narrative fails per-line test, git is SSOT). Final Skill() pointer (L650) = KEEP. |

### C.4 Derived target (REQ-CMD-009) — recomputed after D1 demotions

The D1 re-audit DEMOTED 5 sections/sub-sections from POINTER → KEEP (their distinctive content is UNIQUE — pointer-izing would delete behavioral content). The reduction sources therefore SHRINK relative to iter-1:

- **M-DELETE** (unchanged): changelog footer (~13 lines) + Phase 3 dup examples + Reading-Large-PDFs + scattered TRIM prose.
- **M-POINTER** (REDUCED set after D1): §5 MX, §6, §7 (heavy), §8, §10, §12, §13, §14-Worktree-subsection-only, §15-non-CG-only. The demoted §11 (full KEEP), §16 (full KEEP), §14-operational-bullets (KEEP), §15-CG-Mode (KEEP) NO LONGER contribute reduction.

**Recomputed derived target range: 400-470 lines** (from 650, revised UP from iter-1's 350-430 because 5 sections were correctly demoted to KEEP). The heaviest surviving POINTER target is §7 Safe Development (~87 lines); §15-non-CG and §13/§8/§10 contribute the rest. This range is ~2.0-2.35× the official 200-line target — an HONEST diet that PRESERVES the unique §11/§16/§14-bullets/§15-CG content rather than over-cutting it. Landing slightly higher than iter-1's estimate is the CORRECT outcome of demoting genuinely-unique content: monotonic improvement means NOT regressing into the over-cut that D1 caught. The exact landing point is a run-phase outcome; the range + this justification satisfy REQ-CMD-009. **Behavioral KEEP content (§1 HARD bullets, §4 agent catalog, §11 recovery, §16 search protocol, §14 operational bullets, §15 CG-Mode, the 14 [HARD]/14 [ZONE] lines) is NOT cut to chase a number.**

---

## D. Key Decisions

### D.1 @import honesty decision (the load-bearing one)

The 2 existing `@import` lines (§9 user.yaml + language.yaml) are RETAINED as structure-only. We do NOT introduce new `@import`-based "indexing" and we do NOT count `@import` toward the token reduction. The reduction reported in AC-CMD-006 is attributable SOLELY to M-DELETE + M-POINTER + M-SCOPE. Rationale: `@import` inline-expands at load (coding-standards.md § Paths Frontmatter, verified §F.1) — `@import`-ing always-loaded files yields identical cost. This is REQ-CMD-004 + the load-bearing exclusion.

### D.2 M-SCOPE usage decision

This SPEC does NOT relocate CLAUDE.md body blocks into NEW paths-scoped rules (that is closer to the FUTURE GUARDRAIL-HOOK / RULE-SCOPING territory and would edit `.claude/rules/`). The diet is achieved via M-DELETE + M-POINTER only. M-SCOPE is documented as the third mechanism for completeness but is OUT OF SCOPE for the actions in this SPEC (no new rule files created). If run-phase finds a block that genuinely belongs in a new scoped rule, it returns a blocker rather than silently creating one.

### D.3 Load-bearing [HARD] lines that MUST survive (REQ-CMD-006 / AC-CMD-003)

The 14 `[HARD]` lines are the protected set. Enumerated by section:
- §1: 11 `[HARD]` directive bullets (L9-19) — KEEP verbatim.
- §8: 2 `[ZONE:Frozen][HARD]` (L309 AskUserQuestion-only, L311 deferred-tool preload) — KEEP.
- §14: 1 `[ZONE:Frozen][HARD]` (L455 Background-Agent Write Restriction) — KEEP.

Default rule: **every `[HARD]` line is preserved.** A `[HARD]` line may drop ONLY if run-phase proves it is a verbatim duplicate of an SSOT `[HARD]` line AND the SSOT is always-loaded for the same trigger — and each such drop is enumerated + justified in the run-phase report. The plan-phase default count to preserve is **14** (AC-CMD-003 guards this). The 14 `[ZONE:*]` count tracks alongside (most [HARD] lines are also [ZONE]-tagged).

### D.4 Pointer-ization style (consistency)

Each pointer-ized section keeps the pattern already used by §11/§13/§16 (which open with `> Canonical rule: see <path> for <subject>`). New pointer-izations adopt the SAME `> Canonical rule:` / `For <subject>, see <path>.` style for consistency. The section's unique [HARD] directive (if any) stays ABOVE or BELOW the pointer line, never inside the collapsed prose.

### D.5 §4 nesting-note correctness (REQ-CMD-008, NOT a token action)

The §4 note (L100) is corrected to reflect the verified official mechanism: a subagent spawns nested subagents when `Agent` is in its `tools` list; to prevent, omit `Agent`; depth is fixed at 5 and not configurable; "a subagent at depth five does not receive the Agent tool." This is a content fix; it may be net-neutral in length. It is committed in the SAME run-phase but is a distinct REQ so it is not conflated with token-diet measurements.

---

## E. Milestones (priority-based, no time estimates — mirrors RULE-SCOPING M1→M5 shape)

- **M1 — Template-first body diet (TEMPLATE tree)**: Apply the §C.3 classification to `internal/template/templates/CLAUDE.md`: M-DELETE the changelog footer + TRIM blocks, M-POINTER the duplicated sections (preserving all 14 [HARD] lines), and apply the §4 nesting-note correction (REQ-CMD-008). Edit the TEMPLATE tree FIRST per Template-First (C-1).
- **M2 — Re-embed**: `make build` to regenerate `internal/template/embedded.go` from the edited template tree.
- **M3 — Live-tree parity diet**: Apply the IDENTICAL diet to live `CLAUDE.md` so the two trees stay byte-identical. Verify `diff CLAUDE.md internal/template/templates/CLAUDE.md` → exit 0 (AC-CMD-002).
- **M4 — Verification**: Run acceptance.md AC commands — line-count delta per tree into the derived range (AC-CMD-001), byte-parity (AC-CMD-002), [HARD]-count ≥ 14 preserved (AC-CMD-003), @import not double-counted (AC-CMD-004), changelog footer gone (AC-CMD-005), byte-sum reduced per tree (AC-CMD-006), neutrality grep + `internal_content_leak_test.go` clean (AC-CMD-007), §4 note reconciled (AC-CMD-008). Also: `go build ./...` exit 0 (embedded.go compiles), `go test ./internal/template/...` green (mirror-parity + neutrality guards).
- **M5 — Commit + push**: Conventional Commits per Milestone or bundled; `feat(SPEC-STEERING-ALIGN-CLAUDEMD-DIET-001): M{N} ...`; main-direct push per Hybrid Trunk (B9). No `--no-verify`.

### Milestone risk notes
- M1/M3 are the judgment-free mechanical application of §C.3 — if any line's disposition is ambiguous at run-time, that is a plan defect; return a blocker rather than improvising (C-2).
- M2 `make build` is the parity linchpin: editing live before template, or skipping `make build`, breaks byte-parity. Strict M1→M2→M3 order.

---

## F. Anti-Patterns

- Counting `@import` restructuring as token reduction (violates REQ-CMD-004 / D.1).
- Cutting a `[HARD]`/`[ZONE:*]` behavioral directive to hit "~300 lines" (violates REQ-CMD-006 / REQ-CMD-009 / C-3).
- Editing live `CLAUDE.md` before the template tree, or skipping `make build` (breaks byte-parity / Template-First).
- Pointer-izing a section whose claimed SSOT was NOT grep-verified (violates C-7 / verification-claim-integrity).
- Injecting internal SPEC-ID / date / SHA / audit-citation / memory-path into the CLAUDE.md body (violates REQ-CMD-007 / neutrality).
- Editing a `.claude/rules/...` SSOT file under this SPEC (out of scope D; M-POINTER points AT the SSOT, never modifies it).
- Touching any §A.5 PRESERVE-list path.

---

## G. Cross-References

- spec.md §A.3 (3-mechanism token-reduction taxonomy), §B (REQ-CMD-001..009), §F (evidence).
- acceptance.md (re-runnable AC commands).
- `.claude/rules/moai/development/coding-standards.md` § File Size Limits + § Paths Frontmatter.
- `.claude/rules/moai/development/manager-develop-prompt-template.md` § Applicability (Tier M = full Section A-E delegation; Section B filtered to relevant categories per §B above).
- CLAUDE.local.md §2 / §15 / §25 (Template-First / neutrality / isolation).
- SPEC-STEERING-ALIGN-RULE-SCOPING-001 plan.md (milestone-shape precedent).
