---
id: SPEC-MOAI-SKILL-DOCTRINE-FIX-001
title: "moai Skill Folder Doctrine Drift Remediation (70 Verified Findings)"
version: "0.1.0"
status: completed
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P1
phase: "v3.0.0 target"
module: ".claude/skills/moai"
lifecycle: spec-anchored
tags: "skill-doctrine, drift-remediation, gears, template-neutrality, harness, tier-l, agent-catalog"
---

# SPEC-MOAI-SKILL-DOCTRINE-FIX-001: moai Skill Folder Doctrine Drift Remediation

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-08 | manager-spec | Initial draft — 70 verified findings from 9-way parallel read-only audit of `.claude/skills/moai/` (42 files, ~10,043 lines) |

## §A. Background

A 9-way parallel read-only audit fanned out across the entire `/moai` subcommand skill system (`.claude/skills/moai/`, 42 files, ~10,043 lines) and produced **70 VERIFIED findings** (CRITICAL 7 / MAJOR 32 / MINOR 31). Every finding was independently confirmed against ground-truth by `grep`/`ls`/`diff`/code cross-reference — this SPEC does not re-litigate whether a finding is real; it exists to remediate them.

The findings fall into six cross-cutting themes:

1. **Agent/Route Ownership Drift (T1)** — skill docs unconditionally assign the sync commit / Git operations to `manager-git`, contradicting `spec-workflow.md`'s Frozen HARD Route A (Tier S/M = `manager-develop`/`manager-docs` main-direct push, no PR) vs Route B (Tier L OR `--pr` = `manager-git` PR route).
2. **Harness Config vs Doc Gating Drift (T2)** — skill docs describe harness-level gating (plan_audit skip, sync-auditor invocation) that contradicts the actual `harness.yaml` structure.
3. **Retired/Non-existent Skill & Agent References (T3)** — skill docs reference retired skills (`moai-design-craft`), deprecated skills (`moai-meta-harness`), an absent agent from Quick Reference (`plan-auditor`), and non-canonical team role names (`backend-dev`/`frontend-dev`) that do not exist in `workflow.yaml` `team.role_profiles` / `team-protocol.md` Role Matrix SSOT.
4. **Internal Content Leak / Template Neutrality (T4)** — internal SPEC IDs, multi-segment `REQ`/`AC` tokens, and `C-PH-NNN` citations leak into files that are byte-identical mirrors of `internal/template/templates/.claude/skills/moai/...` (violates `CLAUDE.local.md` §25 Template Internal-Content Isolation), and the CI leak-test regex families structurally cannot detect several of these shapes.
5. **Cross-Reference & Structural Integrity (T5)** — circular dangling references, broken relative links, stale line-number pointers, frontmatter/version mismatches, and config-key drift (flat vs nested booleans) across skill body prose.
6. **CLI/Harness Verb Lifecycle & Trust-Boundary Documentation (T6)** — `harness.md`'s central premise ("CLI verb path retired, no Go binary invoked") is false (SPEC-V3R5-HARNESS-AUTONOMY-001, completed, un-retired lifecycle verbs registered at `root.go:166`), and `moai harness apply --execute` applies a Tier-4 proposal to disk with no `AskUserQuestion` gate — an undocumented parallel trust boundary.

## §B. Constraints (Template-First + Neutrality)

[HARD] Per `CLAUDE.local.md` §2 Template-First Rule: **31 of the 42 audited files are byte-identical to their `internal/template/templates/.claude/skills/moai/...` mirror** (confirmed via `diff -q` during plan-phase authoring; 3 files — `workflows/run/task-decomposition.md`, `workflows/plan/clarity-interview.md`, `workflows/review.md` — currently DIFFER from their template mirror due to unrelated in-flight local edits, and MUST be reconciled, not blindly overwritten, before this SPEC's fixes land). For every file with a template mirror, run-phase implementation MUST:

1. Edit the **template source** (`internal/template/templates/.claude/skills/moai/<relative-path>`) FIRST.
2. Run `make build` to regenerate the embedded binary.
3. Sync/copy the rendered result to the local project path (`.claude/skills/moai/<relative-path>`).
4. Re-verify byte-identity with `diff -q` (or reconcile intentional local-only drift with an explicit note in the run-phase report).

[HARD] Per `CLAUDE.local.md` §25 Template Internal-Content Isolation: template-mirrored files MUST NOT contain internal SPEC IDs, REQ/AC tokens, `Audit N Finding` citations, internal dates, commit SHAs, or archive/memory paths. Every fix that touches prose in a template-mirrored file MUST use generic/neutral phrasing (mechanism description, not internal-provenance citation) — this applies most sharply to REQ-SKF-025 (sync-auditor gating), REQ-SKF-027 (SPEC ID leak), and REQ-SKF-035 (`REQ-HRN-FND-NNN` leak in `harness.md`).

**One narrow exception to the `.claude/skills/moai` module scope**: REQ-SKF-007 sub-clause (b) targets a single stale source-code comment at `internal/cli/root.go:157-160` (Go source, not a skill doc). This is a comment-only edit with no behavior change — justified because that stale comment is the root cause of the false "CLI verb path retired" claim REQ-SKF-007's primary clause corrects in `harness.md`/`SKILL.md`. No other Go source file is in scope for this SPEC.

## §C. GEARS Requirements

Severity tags: **[P0]** = CRITICAL (7 findings), **[P1]** = MAJOR (32 findings, consolidated into 28 REQs where sub-findings share one root cause), **[P2]** = MINOR (31 findings, consolidated into 17 REQs). Theme tags: T1 Agent/Route Ownership · T2 Harness Config Gating · T3 Retired References · T4 Content Leak · T5 Structural Integrity · T6 CLI/Trust-Boundary. Write-group tags (WG-A..WG-I) map to the file-disjoint parallel fanout groups defined in `plan.md` §F.

### C.1 Theme T1 — Agent/Route Ownership Drift

- **REQ-SKF-001** [P0][T1][WG-D] — **When** `workflows/run/task-decomposition.md` Phase 3 "Git Operations" is read by an implementing agent, the skill body **shall** branch on Tier/`--pr` (Route A: `manager-develop` commits + pushes directly to `main`, no PR, for Tier S/M default; Route B: `manager-git` creates a feature branch + opens a PR, for Tier L or explicit `--pr`) instead of unconditionally spawning `manager-git` for every SPEC. Source: `workflows/run/task-decomposition.md:267-296` vs `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Phase Discipline Route A/B (Frozen HARD).
- **REQ-SKF-002** [P0][T1][WG-F] — **When** `workflows/sync/delivery.md` line 33 assigns sync-phase commit ownership, the skill body **shall** branch on the same Tier/route gate (Route A: `manager-docs` emits the single sync commit directly on `main`; Route B: `manager-git` opens the sync PR) instead of unconditionally assigning `manager-git`.
- **REQ-SKF-023** [P1][T1,T3][WG-F] — **When** `workflows/sync/doc-execution.md:186` and `workflows/sync/quality-gates-quality.md:58` cite `archived-agent-rejection.md` §C row 2 (`manager-quality`) to justify spawning `sync-auditor`, the skill body **shall** drop that citation — `sync-auditor` is a RETAINED agent (per the 8-agent catalog) and needs no archived-agent migration justification.

### C.2 Theme T2 — Harness Config vs Doc Gating Drift

- **REQ-SKF-003** [P0][T2][WG-C] — **Where** the harness level is `minimal`, `workflows/plan/spec-assembly.md:160,164,215` **shall** state that plan-audit runs as a **lightweight 1-iteration audit** (`harness.yaml` `minimal.plan_audit.enabled: true`, `require_must_pass: false`, plus `plan_audit_global.always_enabled: true`) rather than claiming plan-audit is skipped entirely at the minimal level.
- **REQ-SKF-025** [P1][T2][WG-F] — **Where** the harness level is `minimal` (`harness.yaml` `minimal.evaluator: false`), `workflows/sync/quality-gates-quality.md:56-66` **shall** gate the `sync-auditor` invocation on harness level instead of invoking it unconditionally on every sync.

### C.3 Theme T3 — Retired/Non-existent Skill & Agent References

- **REQ-SKF-004** [P0][T3][WG-C,WG-B] — **When** `workflows/plan/clarity-interview.md:135` and `workflows/review.md:296,309,338` reference the skill `moai-design-craft` (retired 2026-04-25; no shipping successor `moai-design-system`), the skill body **shall** either remove the skill-name citation (retain only the per-spawn `Agent(general-purpose)` frontend specialist + whitelist reference) or reference a currently-shipping skill — never a retired name.
- **REQ-SKF-009** [P1][T3][WG-A] — The `SKILL.md` Quick Reference agent lists (plan `L104`, default `L176`) **shall** include `plan-auditor` (plan-phase independent audit quality gate) alongside the other agents already listed there.
- **REQ-SKF-011** [P1][T3][WG-A] — `workflows/moai.md:211` **shall** replace the undefined team roles `backend-dev + frontend-dev` with the canonical `team-protocol.md` Role Matrix SSOT roles (`implementer`, `tester`, `reviewer`) for the Run phase implementation team description.
- **REQ-SKF-013** [P1][T3][WG-E] — `team/run.md:49-61` Role Profile table **shall** correct the two wrong model cells (`architect`: `sonnet` → `opus`; `reviewer`: `haiku` → `sonnet`) to match `workflow.yaml` `role_profiles`, OR delete the duplicated table and point to `team-protocol.md` § Role Matrix as the single source of truth.
- **REQ-SKF-016** [P1][T3][WG-B] — `workflows/mx.md:76-83` tag taxonomy **shall** add the 5th tag type `@MX:DEBT` (with its `P5`/DEBT scan-logic row), which is implemented and tested but currently absent from the documented taxonomy.
- **REQ-SKF-022** [P1][T3][WG-D] — `workflows/run/mode-orchestration.md:40,43` **shall** replace the non-canonical team roles `backend-dev`/`frontend-dev`/`quality` with the canonical 7-role set from `team-protocol.md` Role Matrix SSOT.
- **REQ-SKF-028** [P1][T3][WG-G] — `workflows/project/meta-harness.md` Phase 5/6 **shall** stop invoking the deprecated 16-question interview that calls the deprecated `moai-meta-harness` skill (`status: deprecated`, superseded by the `/moai:harness` v4 Builder); the phase **shall** be rewritten to be architecturally compatible with the current `harness-build-entry.md` + `harness-builder.md` v4 flow, or removed if fully superseded.
- **REQ-SKF-037** [P2][T3][WG-B] — `workflows/mx.md:24` frontmatter `agents` list **shall** match `SKILL.md`'s agent roster, and the body **shall** add the missing "Agent Chain Summary" + "Team Mode" sections present in sibling workflow files.

### C.4 Theme T4 — Internal Content Leak / Template Neutrality

- **REQ-SKF-027** [P1][T4][WG-F] — The internal SPEC ID `SPEC-DB-SYNC-RELOC-001` leaked into `workflows/sync/quality-gates-context.md:153` (a byte-identical template mirror) **shall** be removed or replaced with generic/neutral phrasing, per `CLAUDE.local.md` §25.
- **REQ-SKF-032** [P1][T4][WG-G] — Internal SPEC/`REQ`/`AC`/`C-PH-NNN` tokens leaked into `workflows/project/doc-generation.md`, `workflows/project/meta-harness.md`, and `workflows/project.md` (byte-identical template mirrors) **shall** be removed or replaced with generic/neutral phrasing, per `CLAUDE.local.md` §25.
- **REQ-SKF-035** [P1][T4][WG-H] — The ~25 `REQ-HRN-FND-NNN` (4-segment) citations in `workflows/harness.md` (a byte-identical template mirror) **shall** be removed or replaced with generic/neutral mechanism descriptions, per `CLAUDE.local.md` §25.
- **REQ-SKF-053** [P1][T4][WG-I] — The `internal/template/internal_content_leak_test.go` leak-class regex families **shall** be extended to detect: (a) 2-segment `REQ-NNN`/`AC-N` short-code tokens (no domain segment, e.g. `REQ-006`/`AC-6`), (b) the `C-PH-NNN` citation shape, (c) non-`V3R[2-6]`/`AGENCY`/`WORKTREE` SPEC-ID prefixes (e.g. `SPEC-DB-SYNC-RELOC-001`-style single-domain-segment IDs that escape the current `\bSPEC-(V3R[2-6]|AGENCY|WORKTREE)-[A-Z0-9-]+\b` pattern), and (d) the 4-segment `REQ-HRN-FND-NNN` shape (which the current single-segment `\b(REQ|AC)-[A-Z][A-Z0-9]*-[0-9]+\b` regex structurally cannot match). `.github/workflows/template-neutrality-check.yaml` trigger paths **shall** be reviewed for coverage of the touched skill directories.

### C.5 Theme T5 — Cross-Reference & Structural Integrity

- **REQ-SKF-005** [P0][T5][WG-G] — `workflows/project/meta-harness.md:225-227`'s FROZEN path list **shall** be corrected to exactly match `internal/harness/frozen_guard.go` `frozenPrefixes` (verified: **4** entries — `.claude/agents/moai/`, `.claude/skills/moai-`, `.claude/skills/moai/`, `.claude/rules/moai/`). The doc's genuine error is bundling `harness/` into the agents line (`.claude/agents/harness/` is in `allowedPrefixes`, NOT frozen — the agents line **shall** read `.claude/agents/moai/` only); the doc's separate omission is the exact `.claude/skills/moai/` prefix (distinct from the already-correctly-listed `.claude/skills/moai-*` wildcard form). The fix is narrow and additive-and-subtractive, NOT a wholesale rewrite: (a) drop `harness/` from the agents line, (b) add the missing exact `.claude/skills/moai/` entry — the doc's other two already-correct entries (`.claude/skills/moai-*/`, `.claude/rules/moai/`) **shall be preserved**, not deleted.
- **REQ-SKF-006** [P0][T5][WG-G] — The circular dangling "Detection Keywords Reference" between `workflows/project.md:89-93` and `workflows/project/doc-generation.md:196-197` **shall** be resolved: the 16-language manifest/ORM keyword table that Phase 4.1a depends on **shall** be authored in exactly one of the two files, and the other file's pointer **shall** be corrected to reference it (no circular "see the other file" with neither containing the table).
- **REQ-SKF-008** [P1][T5][WG-A] — `SKILL.md:102` and `workflows/moai.md:124` **shall** replace the stale "EARS format" reference with "GEARS format" (GEARS is canonical per the manager-spec agent body; EARS legacy syntax is supported only during the documented 6-month backward-compatibility window).
- **REQ-SKF-010** [P1][T5][WG-A] — `references/reference.md:195-206` **shall** remove the dead Coverage/E2E flag sections that reference retired subcommands (cross-ref `SPEC-SUBCOMMAND-RETIRE-001`).
- **REQ-SKF-012** [P1][T5][WG-A] — `SKILL.md:244` (declares `--worktree` valid for the default pipeline) vs `workflows/moai.md` (never parses `--worktree`) **shall** be reconciled: either correct the `SKILL.md` claim, or add `--worktree` parsing to `workflows/moai.md`.
- **REQ-SKF-014** [P1][T5][WG-E] — `team/glm.md`'s documentation of the legacy `CLAUDE_CODE_TEAMMATE_DISPLAY=tmux` env var (superseded — `internal/cli/glm.go` now uses the `settings.local.json` `teammateMode: "tmux"` field and actively deletes the legacy env var) **shall** be updated to document the `teammateMode` field mechanism.
- **REQ-SKF-015** [P1][T5][WG-E] — The contradictory `team_mode: ""` semantics between `team/glm.md` and `team/run.md` **shall** be reconciled with a single cross-referenced explanation of the evaluation point (and, where relevant, disambiguated from the distinct `llm.yaml` `team_mode` field per `CLAUDE.local.md` §22.3).
- **REQ-SKF-017** [P1][T5][WG-B] — `workflows/clean.md:62-65`'s language-neutrality table (covering 4/16 languages) **shall** be expanded to all 16 supported languages, or explicitly deferred to `workflows/fix.md` Scanner 3's table with a cross-reference (no partial/misleading subset presented as complete).
- **REQ-SKF-018** [P1][T5][WG-B] — `workflows/feedback.md:54-60,163-167` `AskUserQuestion` option sets **shall** carry the `(Recommended)`/`(권장)` first-option label per the HARD Option Description Standards (`askuser-protocol.md`).
- **REQ-SKF-019** [P1][T5][WG-D] — `workflows/run/phase-execution.md:214-237` Phase 0.95 description **shall** instruct writing the `progress.md` `§F Phase 0.95 Mode Selection` log entry per `orchestration-mode-selection.md` §D HARD logging contract (currently omitted).
- **REQ-SKF-020** [P1][T5][WG-D] — `workflows/run/context-loading.md:90`'s broken relative markdown link **shall** be corrected (needs 4× `../`, not the current path depth) and its anchor slug **shall** match the target heading.
- **REQ-SKF-021** [P1][T5][WG-D] — `workflows/run/task-decomposition.md:156-169` Phase 2.75 (lint/format/type-check) **shall** be rewritten to reference the single-turn multi-Bash parallel batch + file-redirect (`/tmp/moai-verify/`) contract (`agent-common-protocol.md` § Parallel Execution) instead of presenting the checks as serial steps.
- **REQ-SKF-024** [P1][T5][WG-F] — `workflows/sync/doc-execution.md:78-81,139-151,200-202`'s "SPEC Lifecycle Level 1/2/3 (spec-first/spec-anchored/spec-as-source)" taxonomy **shall** be corrected to the canonical `spec-frontmatter-schema.md` `lifecycle` enum (`spec-anchored`\|`spec-lite`\|`exploratory`, default `spec-anchored`).
- **REQ-SKF-026** [P1][T5][WG-F] — `workflows/sync/quality-gates-quality.md:62-66`'s review "Perspectives" (Security/Performance/Quality/UX) **shall** be corrected to match the canonical `sync-auditor` rubric (Functionality 40 / Security 25 / Craft 20 / Consistency 15 weighted, Security hard-fail).
- **REQ-SKF-029** [P1][T5][WG-G] — `workflows/project.md:41-85` routing table/flow **shall** include the mandatory Phase 7 (5-Layer Activation), currently omitted.
- **REQ-SKF-030** [P1][T5][WG-G] — `workflows/project/doc-generation.md:269-300` Phase 5/6/7 **shall** be made reachable from the Phase 4.2 branch logic (currently dead — no opt-in wiring connects them).
- **REQ-SKF-031** [P1][T5][WG-G] — `workflows/project/doc-generation.md:188,206,277,282,291` (5 occurrences) **shall** correct `db.auto_sync: true` (flat boolean) to the actual nested `db.yaml` structure (`db.auto_sync.enabled: true`).
- **REQ-SKF-033** [P1][T5,T6][WG-A] — `SKILL.md` Branch A.1 (lines 210, 213, 214) **shall** include `doctor` in its own section title, Verbs list, and CLI examples — currently omitted despite the line-198 dispatch table and `harness-build-entry.md`/`harness-builder.md` both including it. This also resolves the self-contradiction between `SKILL.md:70` ("CLI verb path retired... no Go binary invoked") and `SKILL.md:198,213,214` (which document `list`/`edit`/`remove`/`doctor` as dispatching to the `moai harness <verb>` Go binary subcommand).
- **REQ-SKF-036** [P2][T5][WG-B] — `references/mx-tag.md:24` frontmatter `phases` list **shall** add the missing `"sync"` entry; `:44` grammar `sub_value` enum **shall** add the missing `CEILING`/`UPGRADE` values.
- **REQ-SKF-038** [P2][T5][WG-B] — `workflows/fix.md:134`'s stale pointer to "sync.md Phase 0.6.1" **shall** be corrected to "`sync/quality-gates-quality.md` Step 0.6.1".
- **REQ-SKF-039** [P2][T5][WG-B] — `workflows/review.md:64-68` vs `:327` `--security` agent-routing ambiguity in the Agent Chain Summary **shall** be resolved to a single unambiguous routing statement.
- **REQ-SKF-040** [P2][T5][WG-C] — `workflows/plan/clarity-interview.md:34-39`'s manual "Type your own answer" option (which duplicates the auto-appended "Other" option and wastes an option slot) **shall** be removed.
- **REQ-SKF-041** [P2][T5][WG-C] — The `plan/` subfolder files **shall** add a `ToolSearch(query: "select:AskUserQuestion")` preload citation at each `AskUserQuestion` call site currently missing one (~9 of 10 sites).
- **REQ-SKF-042** [P2][T5][WG-C] — `workflows/plan/spec-assembly.md`'s DP2/DP3 decision points **shall** apply the `(Recommended)`/`(권장)` first-option label consistently (currently inconsistent between the two).
- **REQ-SKF-043** [P2][T5][WG-A] — `workflows/moai.md:81`'s time-prediction "2-3x speedup (15-30s vs 45-90s)" **shall** be replaced with priority/phase-ordering language per the Time Estimation HARD rule (no duration predictions).
- **REQ-SKF-044** [P2][T5][WG-A] — `workflows/moai.md:35-46` Supported Flags table **shall** add the missing `--sequential` flag; frontmatter/footer version-date mismatches in `moai.md` and `plan.md` **shall** be corrected to the current values.
- **REQ-SKF-045** [P2][T5][WG-A] — `workflows/plan.md:37`'s broken internal "Phase 6" reference **shall** be corrected (Git Env Setup is actually Phase 3).
- **REQ-SKF-046** [P2][T5][WG-D] — `workflows/run/task-decomposition.md:292`'s hardcoded stale "Claude Opus 4.6" commit co-author string **shall** be genericized (no hardcoded model-version string in a commit template).
- **REQ-SKF-047** [P2][T5][WG-D] — The `run/` file-naming mismatch (`phase-execution.md` ↔ `task-decomposition.md` content/filename swap risk) **shall** be verified and corrected if present; the retired "Sprint" term **shall** be replaced with "Epic"/"Milestone" per `sprint-round-naming.md` at all 3 occurrences (`context-loading.md:116`, `phase-execution.md:402`, `task-decomposition.md:179`); the teammate-count table **shall** be corrected from "3-4" to "3-5" (Anthropic guidance: "Start with 3-5 teammates").
- **REQ-SKF-048** [P2][T5][WG-H] — `workflows/harness.md:269`'s retired "Wave C" term **shall** be replaced with the current Epic/Milestone taxonomy term.
- **REQ-SKF-049** [P2][T5][WG-G] — `workflows/project/mode-detection.md:64` and `workflows/project/doc-generation.md:160`'s detection glob (covering only 7/16 languages) **shall** be expanded to all 16 supported languages.
- **REQ-SKF-050** [P2][T5][WG-E,WG-A,WG-B] — `team/debug.md`'s informal prose spawn instructions (inconsistent with sibling `Agent()` code-block conventions) **shall** be normalized (owned by WG-E); `SKILL.md`'s fix entry **shall** add a team-mode pointer (owned by WG-A); the `workflows/fix.md:43` ("team-debug.md") vs `:279` ("team/debug.md") naming inconsistency **shall** be corrected to the single actual path `team/debug.md` (owned by WG-B). This REQ spans 3 write-groups by design — each owning write-group applies its own sub-clause as a real assigned edit (not a cross-reference), per the same multi-file-REQ pattern used by REQ-SKF-007 `[WG-H,WG-A]` and REQ-SKF-004 `[WG-C,WG-B]`.
- **REQ-SKF-051** [P2][T5][WG-E] — `team/glm.md`'s role→model table **shall** add the missing `architect` row; its tmux env-var table **shall** add the missing `ANTHROPIC_DEFAULT_FABLE_MODEL` entry.
- **REQ-SKF-052** [P2][T5][WG-F] — `workflows/sync/doc-execution.md`'s 3rd divergent `sync-auditor` description **shall** be unified with the canonical description; `workflows/sync/delivery.md:239`'s language undercutting the Route A default **shall** be corrected; `workflows/sync/quality-gates-quality.md:105-118`'s Stop-hook/agent conflation **shall** be clarified; the phase-number↔`harness.yaml` `skip_phases` namespace **shall** be made unambiguous.

### C.6 Theme T6 — CLI/Harness Verb Lifecycle & Trust-Boundary Documentation

- **REQ-SKF-007** [P0][T6][WG-H,WG-A] — `workflows/harness.md:3-7,29-30,34,263` and `SKILL.md:70,203`'s central premise ("CLI verb path retired, no Go binary invoked") **shall** be corrected. Ground truth (verified against `internal/cli/harness_route.go:59-148` `newHarnessRouterCmd()` and the CI-enforced `internal/cli/harness_retirement_test.go` `TestHarnessV3R5VerbSurface`): **ALL** harness verbs — including the historically-named "learning-lifecycle" verbs `status`/`apply`/`rollback`/`disable` — are registered **Go-binary Cobra subcommands** under `moai harness` (`SPEC-V3R5-HARNESS-AUTONOMY-001` un-retired them; the CI guard asserts they MUST remain registered). There is **NO** learning-vs-v4 Go-binary-dispatch split — every verb (`route`/`validate`/`status`/`apply`/`rollback`/`disable`/`mute`/`mute-list`/`unmute`/`verify`/`propose`/`install`/`list`/`edit`/`remove`/`doctor`) dispatches through the single unified `newHarnessRouterCmd()` tree (registered at `root.go:166`). The corrected framing **shall NOT** introduce a workflow-body-only-vs-Go-binary split — that split does not exist in the code and would replace one false claim with a different false claim.
  - **Sub-clause (b) — stale source comment.** The comment at `internal/cli/root.go:157-160` (which describes the superseded, unregistered `newHarnessCmd()` factory and asserts the lifecycle verbs have "no Go binary invocation") **shall** be corrected to state that `newHarnessRouterCmd()` (registered at `root.go:166`) is the live, unified registration site for all harness verbs. This is a narrow, comment-only edit (no behavior change) to `internal/cli/root.go` — the one exception to this SPEC's `.claude/skills/moai` module scope, justified because the stale comment is the root cause of the false "CLI verb path retired" doc claim it exists to correct.
- **REQ-SKF-034** [P1][T6][WG-H] — `workflows/harness.md`'s `moai harness apply --execute` flag (which applies a Tier-4 proposal to disk with **no** `AskUserQuestion` gate) **shall** be documented as an explicit, separate trust-boundary path — distinct from the default `apply` flow (which IS gated by `AskUserQuestion` per `REQ-HRN-FND-004`-equivalent language, phrased generically per §B neutrality).

## §D. Traceability Summary

| REQ range | Severity | Count | Themes touched | Write-groups touched |
|-----------|----------|-------|-----------------|-----------------------|
| REQ-SKF-001..007 | P0 (CRITICAL) | 7 | T1, T2, T3, T5, T6 | WG-A, WG-B, WG-C, WG-D, WG-F, WG-G, WG-H |
| REQ-SKF-008..035 | P1 (MAJOR) | 28 (consolidates 32 findings) | T1, T2, T3, T4, T5, T6 | WG-A..WG-I |
| REQ-SKF-036..052 | P2 (MINOR) | 17 (consolidates 31 findings) | T3, T5 | WG-A..WG-H |
| REQ-SKF-053 | P1 (CI hardening) | 1 | T4 | WG-I |

Full per-file mapping, milestone ordering, and write-group file lists: see `plan.md` §F.

## §E. Exclusions

### Out of Scope — Code/Config Behavior Changes

- Changing the actual `harness.yaml`, `workflow.yaml`, `db.yaml`, or `spec-frontmatter-schema.md` values referenced by findings — this SPEC corrects **documentation to match** existing config/code ground truth; it does NOT change ground-truth config or Go source behavior (e.g., REQ-SKF-003 corrects the doc claim about `plan_audit.enabled`, it does not change `harness.yaml`).
- Implementing the actual `moai harness apply --execute` `AskUserQuestion` gate if REQ-SKF-034's audit reveals the gate should exist but doesn't — that is a code-behavior change belonging to a follow-up SPEC against `internal/harness/`. This SPEC only requires accurate **documentation** of the current (gated or ungated) behavior.
- Fixing the retired `moai-design-craft` / deprecated `moai-meta-harness` skills themselves (reviving or rewriting their bodies) — REQ-SKF-004 and REQ-SKF-028 only require the **referencing** skill docs to stop citing them incorrectly.

### Out of Scope — Non-Audited Surfaces

- The **13 no-finding files** (`references/anti-patterns.md`, `references/file-reading-optimization.md`, `workflows/codemaps.md`, `workflows/loop.md`, `workflows/gate.md`, `team/plan.md`, `team/review.md`, `team/sync.md`, `workflows/sync.md`, `workflows/project/codebase-analysis.md`, `workflows/harness-build-entry.md`, `workflows/harness-builder.md`, `workflows/run.md`) — included in the 42-file enumeration for completeness but received zero findings (0 REQ-SKF-NNN cites any of these 13); no changes required beyond a pass-through regression check (see `plan.md` §G for the full per-write-group enumeration). `workflows/run.md`'s no-finding status was confirmed during plan-audit review (no REQ in this SPEC cites it; annotated consistent with its 12 counted siblings). 42 total files − 13 no-finding = **29 files carry ≥1 REQ**.
- Any `.claude/skills/` folder outside `moai/` (e.g. `moai-foundation-cc`, `moai-workflow-spec`) — out of the audited scope.
- Agent body files (`.claude/agents/**/*.md`) and rule files (`.claude/rules/**/*.md`) — these are the ground-truth this SPEC corrects skill docs *against*; modifying them is out of scope unless a specific REQ explicitly says otherwise (none do).

### Out of Scope — CI/Test Infrastructure Beyond REQ-SKF-053

- Building a new, separate CI workflow for skill-doctrine-drift detection — REQ-SKF-053 only extends the existing `internal_content_leak_test.go` regex families and reviews (not necessarily rewrites) `.github/workflows/template-neutrality-check.yaml` trigger paths.
- Retroactively re-auditing the rest of `.claude/skills/` (non-`moai/` skills) or `.claude/rules/` for the same drift classes — a distinct future SPEC's scope.
