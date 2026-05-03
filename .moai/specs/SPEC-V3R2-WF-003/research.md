# SPEC-V3R2-WF-003 Deep Research (Phase 0.5)

> Research artifact for `Multi-Mode Router (--mode flag, loop/run/design)`.
> Companion to `spec.md` (v0.2.0). Authored against worktree `feature/SPEC-V3R2-WF-003`.
> Stacked on top of `feature/SPEC-V3R2-WF-004` (PR #765, OPEN).

## HISTORY

| Version | Date       | Author                                 | Description                                                            |
|---------|------------|----------------------------------------|------------------------------------------------------------------------|
| 0.1.0   | 2026-05-03 | MoAI Plan Workflow (run-phase research)| Initial deep research per `.claude/skills/moai/workflows/plan.md` Phase 0.5 |

---

## 1. Goal of Research

Substantiate `spec.md` §1 (Goal) and §4 (Assumptions) with concrete file:line evidence so that the run phase can implement REQ-WF003-001..018 against a known-good baseline. The research answers six questions:

1. Where does the **mode-axis dispatch** of `/moai run`, `/moai loop`, and `/moai design` actually need to live (CLI layer? skill layer?)?
2. What is the current state of `/moai loop` — is it already a thin wrapper around the Ralph engine (per assumption 4.1 in spec.md)?
3. How is the harness level (minimal/standard/thorough) read today, and how can it drive mode auto-selection per REQ-WF003-002/003?
4. What are the exact pre-conditions for `team` mode activation (REQ-WF003-011 / REQ-WF003-012), and where are those checks performed today?
5. How do `/moai design` Path A (Claude Design import), Path B1/B2 (Figma/Pencil), and Path B (code-based copy + brand) map onto the `--mode {autopilot|import|team}` axis without breaking the existing 3-path UX?
6. Which sentinel error keys are owned by WF-003 vs WF-004, and is the cross-spec contract `MODE_PIPELINE_ONLY_UTILITY` consistent between the two SPECs?

---

## 2. Architectural Anatomy of `/moai {run,loop,design}`

### 2.1 Slash command surface (`.claude/commands/moai/`)

All three user-invocable surfaces are **thin wrappers** that route into a single `moai` skill, per the Thin Command Pattern (`.moai/design/v3-redesign/research/r6-commands-hooks-style-rules.md` §1; `internal/template/commands_audit_test.go:11-50`). This is the same convention as the WF-004 utility subcommand surface.

| Subcommand   | File                                          | Body (lines 5-7) |
|--------------|-----------------------------------------------|------------------|
| `/moai run`  | `.claude/commands/moai/run.md:1-7`            | `Use Skill("moai") with arguments: run $ARGUMENTS` |
| `/moai loop` | `.claude/commands/moai/loop.md:1-7`           | `Use Skill("moai") with arguments: loop $ARGUMENTS` |
| `/moai design` | `.claude/commands/moai/design.md:1-7`       | `Use Skill("moai") with arguments: design $ARGUMENTS` |
| `/moai plan` | `.claude/commands/moai/plan.md:1-7`           | `Use Skill("moai") with arguments: plan $ARGUMENTS` |

All four declare `allowed-tools: Skill` only, and pass `$ARGUMENTS` verbatim. **Implication**: the `--mode` flag never crosses a Go process boundary — it travels from the user's slash invocation into the workflow skill body as a positional/named argument. The skill body (Markdown executed by Claude) parses and dispatches on it.

This contrasts with `internal/cli/run.go` etc. — those Go commands exist for the standalone `moai` binary (CLI from terminal), not for the slash command path. The slash command flow is entirely orchestrator + skill-driven; no Go handler intercepts `--mode`.

### 2.2 Workflow skill control flow (`.claude/skills/moai/workflows/`)

| Skill | Lines | Current `--mode` axis support? | What dispatches today |
|-------|-------|--------------------------------|-----------------------|
| `run.md` | 1014 | NO. `--team` flag exists (line 46), but no `--mode` axis | Phase 0.95 Scale-Based Mode Selection (lines 361-382) auto-selects between Fix/Focused/Standard/Full Pipeline/Team modes based on file/domain count; Phase 2A/2B routes by `quality.yaml development_mode` (DDD vs TDD) at lines 537-545 |
| `loop.md` | 249 | NO. Existing flags `--max`, `--auto-fix`, `--sequential`, `--errors`, `--coverage`, `--memory-check`, `--resume` (lines 35-44) — no `--mode` | Direct iteration loop; calls expert agents per-issue-type (lines 106-117). Already exhibits Ralph engine semantics (scan/fix/verify/repeat). |
| `design.md` | 256 | NO. Existing path selection via AskUserQuestion (lines 64-94) splits Path A / B1 / B2 explicitly | Phase 0 pre-flight (brand context check) → Phase 1 route selection → Phase A/B1/B2 → Phase B-Common → Phase C quality gate (GAN loop) |
| `plan.md` | 805 | NO. `--mode` flag NOT present | Auto-routes by SPEC complexity; uses manager-spec subagent. |
| `sync.md` | 1178 | NO. `--mode` flag NOT present | Auto-routes by SPEC artifacts; uses manager-docs subagent. |

#### 2.2.1 `/moai run` — `.claude/skills/moai/workflows/run.md`

The entry point already implements **scale-based mode selection** at Phase 0.95 (lines 361-382):

| Request Pattern | Detection Criteria | Execution Mode | Agents |
|----------------|-------------------|---------------|--------|
| Bug fix / error fix | SPEC scope ≤ 3 files, single domain | **Fix Mode** | expert-debug + expert-testing |
| Single endpoint / function | SPEC scope ≤ 5 files, single domain | **Focused Mode** | relevant expert + expert-testing |
| Feature across 1 domain | SPEC scope 5-10 files, single domain | **Standard Mode** | manager-strategy + relevant expert + manager-quality |
| Multi-domain feature | SPEC scope ≥ 10 files OR ≥ 3 domains | **Full Pipeline** | All agents |
| Large cross-cutting change | complexity score ≥ 7 AND --team flag | **Team Mode** | 3-4 parallel teammates |

This existing infrastructure is the **substrate** for WF-003's `--mode` axis. The four `--mode` values map onto existing infrastructure as follows:

| `--mode` value | Maps to existing path | Reference |
|----------------|----------------------|-----------|
| `autopilot` (default) | Fix/Focused/Standard/Full Pipeline modes per Phase 0.95 — single-lead orchestration | `run.md:361-382` (Phase 0.95) + `run.md:602-686` (Phase 2A/2B DDD/TDD) |
| `loop` | Ralph engine via the existing `moai-workflow-loop` skill | `loop.md:1-249` (entire skill); invocation pattern same as Phase 2A/2B but loop-wrapped |
| `team` | Existing Team Mode (Phase 0.95 row 5) per `.claude/skills/moai/team/run.md` | `run.md:46,927-943` (Team Mode Routing); requires `workflow.team.enabled: true` + `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` |
| `pipeline` | REJECTED — `MODE_PIPELINE_ONLY_UTILITY` per REQ-WF003-016 (cross-ref WF-004) | run.md will need a rejection clause (already partially planned per WF-004 plan.md M3) |

Critical existing infrastructure references:
- **Run-time team activation** (`run.md:927-943`): `1. Verify prerequisites: workflow.team.enabled == true AND CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 env var is set` — this is the exact contract REQ-WF003-011/012 references.
- **Team fallback message** (`run.md:932`): `If prerequisites NOT met: Warn user then fallback to standard sub-agent mode` — REQ-WF003-012 (silent downgrade with info log) effectively codifies the existing fallback into a named sentinel.
- **Mode override hooks** (`run.md:920-922`): `If no execution_mode provided (direct /moai run invocation): Apply existing --team/--solo flag logic` — `--mode` flag will sit at this same junction as a successor to `--team`/`--solo`.

Mapping to WF-003 mode axis:

| WF-003 mode | Phase 0.95 / Team Mode equivalent | What happens at run-time |
|-------------|----------------------------------|--------------------------|
| autopilot   | Fix/Focused/Standard/Full Pipeline (Phase 0.95 rows 1-4) | Single-lead orchestration; manager-strategy + execution agents |
| loop        | (NEW) — invoke `moai-workflow-loop` skill instead of Phase 2A/2B | Ralph engine: scan/fix/verify/repeat per `loop.md:46-138` per-iteration cycle |
| team        | Phase 0.95 row 5 (Large cross-cutting change) + Team Mode Routing | TeamCreate via `team/run.md`; backend-dev + frontend-dev + tester + quality teammates |
| pipeline    | REJECTED — emit `MODE_PIPELINE_ONLY_UTILITY` | Multi-agent subcommand cannot honor pipeline; flag is utility-only per WF-004 |

#### 2.2.2 `/moai loop` — `.claude/skills/moai/workflows/loop.md`

This skill **already implements** the Ralph engine semantics that REQ-WF003-008 references. Specifically:

- Per-iteration cycle (`loop.md:46-138`) — Step 1 Completion Check → Step 2 Memory Pressure Check → Step 3 Parallel Diagnostics → Step 4 Completion Condition Check → Step 5 Task Generation → Step 5.5 Pre-Fix MX Context Scan → Step 6 Fix Execution → Step 7 Verification → Step 7.5 MX Tag Check → Step 8 Snapshot Save → Step 9 Repeat or Exit.
- Fix agent dispatch (`loop.md:106-117`) — agents picked by issue type via static lookup table (same pattern as `fix.md` per WF-004 research §2.2.1, indicating loop is already Agentless-compatible at the per-iteration level).
- Convergence semantics (`loop.md:140-152`) — exit conditions: completion marker, all conditions met, max iterations, memory pressure. This satisfies REQ-WF003-017 (loop terminates on convergence rather than infinitely).

REQ-WF003-004's claim "`/moai loop` shall be an alias resolving to `/moai run --mode loop`" therefore means: the existing `loop.md` skill body must be **invoked from inside `run.md` when `--mode loop` is parsed**, OR the existing `loop.md` skill body must remain authoritative and `run.md --mode loop` must redirect to it.

**Decision rationale (recommended for the run phase)**: keep `loop.md` as the authoritative implementation. Update `loop.md` and `run.md` so that:
1. `loop.md` contains a one-line note: "This skill is invocable directly via `/moai loop` or via `/moai run --mode loop` per SPEC-V3R2-WF-003 REQ-WF003-004."
2. `run.md` at Phase 0.95 routing adds a new branch: when `--mode loop` is present, **delegate execution to the `moai-workflow-loop` skill** (Skill("moai-workflow-loop") with the SPEC-ID and remaining args) rather than entering Phase 2A/2B.
3. The `/moai loop` command file remains unchanged (it already invokes `loop.md` directly).

This preserves the existing user mental model (`/moai loop` works as today) while satisfying REQ-WF003-004 (the alias relationship is documented and behaviorally identical).

#### 2.2.3 `/moai design` — `.claude/skills/moai/workflows/design.md`

This skill currently presents a **3-way path selection** (lines 64-94) via AskUserQuestion:
- Option 1 (Recommended): Path A (Claude Design)
- Option 2: Path B1 (Figma)
- Option 3: Path B2 (Pencil)

Plus an implicit Path B (code-based copywriting + brand-design) reached via Phase B-Common (lines 167-192).

The WF-003 spec.md §2.1 specifies:
- `autopilot` → Path B (code-based, default — currently Phase B-Common)
- `import` → Path A (Claude Design handoff bundle, currently Step A1-A5)
- `team` → "large design brief with parallel copywriting + brand-design" (NEW — parallels TeamCreate spawn for `moai-domain-copywriting` + `moai-domain-brand-design`)

**Critical observation**: the current design.md has FOUR paths (A, B1, B2, B-Common = code-based-brand), while WF-003 specifies THREE modes for design (`autopilot`, `import`, `team`). Mapping:

| WF-003 design mode | Maps to current path | Notes |
|--------------------|----------------------|-------|
| `autopilot` | Phase B-Common (code-based copywriting + brand-design) | Default; matches current B-Common Step BC-2..BC-5 |
| `import` | Phase A (Claude Design) | Single skill: `moai-workflow-design-import` per design.md:111 |
| `team` | (NEW) Phase B-Common with **parallel TeamCreate** for `moai-domain-copywriting` + `moai-domain-brand-design` | Per REQ-WF003-009: spawn both teammates concurrently per the GAN Loop contract |

**Path B1 (Figma) and Path B2 (Pencil) are NOT in the WF-003 mode axis**. This is a gap that the run phase must resolve. Three options:

1. **Subsume Path B1/B2 into `autopilot` mode** — autopilot's behavior depends on input availability (`.pen` files? Figma token? Otherwise default to code-based per Phase B-Common). This preserves all 3 existing paths under `autopilot` umbrella.
2. **Extend the mode axis to include `figma` and `pencil` values** — adds two values, but spec.md §1 strictly specifies 4 mode values (`autopilot|loop|team|pipeline`).
3. **Treat Path B1/B2 as `import`-mode subcategories** — confused with Claude Design import path A.

**Recommendation (for run phase)**: adopt option 1. The `autopilot` mode for `/moai design` performs the same Phase 0 pre-flight checks as today, then continues to Phase 1 path selection (which still asks user to choose between Path A/B1/B2 via AskUserQuestion). When `--mode autopilot` is explicit, the orchestrator skips the path selection entirely and goes straight to Phase B-Common (code-based). When `--mode import` is explicit, orchestrator skips path selection and goes straight to Path A. When `--mode team` is explicit, orchestrator skips path selection and spawns parallel copywriting + brand-design teammates per REQ-WF003-009.

This mapping is **additive**: today's `/moai design` (no `--mode` flag) still presents the AskUserQuestion path selection. With `--mode`, the user opts out of the selection and goes directly to the chosen execution path.

#### 2.2.4 `/moai plan` and `/moai sync` — mode-axis NOT applicable

REQ-WF003-005 mandates that `plan`, `sync`, `project`, `db` ignore `--mode` flag. Verification:

- `plan.md` (805 lines): No mode-axis dispatch present. Workflow is linear (Phase 0 / 0.5 / 1A / 1B / 2 / 3 / 4 per skill body). Adding `--mode` here would be meaningless because plan phase is itself an orchestrator that produces SPEC artifacts; there is no execution style axis to vary.
- `sync.md` (1178 lines): No mode-axis dispatch present. Workflow is linear sync (manager-docs orchestration, multi-language docs sync, PR creation).

Implementation requirement for these two skills: add a one-line "Mode Flag: NOT APPLICABLE" notation per REQ-WF003-005, mirroring the WF-004 plan §2 M3 pattern for utility-vs-implementation skills. In WF-004's M3, plan/run/sync/design all received a `## Mode Flag Compatibility` section. WF-003 will REFINE that section to distinguish between:
- `plan`/`sync` (mode-axis NA per REQ-WF003-005, but `pipeline` mode still rejected with `MODE_PIPELINE_ONLY_UTILITY`)
- `run`/`design` (full mode-axis support per REQ-WF003-001)

### 2.3 Implementation skill comparison: 4 implementation skills vs 5 utility skills (WF-004 cross-ref)

For contrast with WF-004's classification:

| Skill | WF-004 class | WF-003 `--mode` support | Default mode source |
|-------|--------------|-------------------------|--------------------|
| `plan.md` | Multi-Agent | Ignored (REQ-WF003-005) | `autopilot` fixed |
| `run.md` | Multi-Agent | Full axis (autopilot/loop/team) | Harness or config |
| `sync.md` | Multi-Agent | Ignored (REQ-WF003-005) | `autopilot` fixed |
| `design.md` | Multi-Agent | Full axis (autopilot/import/team) | Harness or config |
| `fix.md` | Pipeline (Agentless) | Ignored — `MODE_FLAG_IGNORED_FOR_UTILITY` (REQ-WF004-011) | n/a |
| `coverage.md` | Pipeline (Agentless) | Ignored | n/a |
| `mx.md` | Pipeline (Agentless) | Ignored | n/a |
| `codemaps.md` | Pipeline (Agentless) | Ignored | n/a |
| `clean.md` | Pipeline (Agentless) | Ignored | n/a |

This is the **9-row matrix** that WF-004 plan §2 M4 (`spec-workflow.md:17` insertion point — `## Subcommand Classification` section) is publishing. WF-003 INHERITS this matrix structure and ADDS detail about which `--mode` values each multi-agent subcommand accepts. The two SPECs together define the complete subcommand × mode contract.

---

## 3. Mode Auto-Selection Logic (REQ-WF003-002, REQ-WF003-003)

### 3.1 Harness level reading

Per `.moai/config/sections/harness.yaml:5-103`:

- Harness levels are: `minimal`, `standard`, `thorough` (`harness.yaml:67-103`).
- Auto-detection rules (`harness.yaml:16-31`):
  - `minimal`: `file_count <= 3 AND single_domain` OR `spec_type in [bugfix, docs, config]`
  - `standard`: `file_count > 3 OR multi_domain` OR `spec_type in [feature, refactor]`
  - `thorough`: `security_keywords OR payment_keywords present` OR `spec_priority == critical` OR `domain in [auth, payment, migration, public_api]`
- The harness level is consumed at `run.md:64-83` (Harness Level Routing) — already determines which phases skip and whether evaluator runs.

REQ-WF003-002 (`autopilot` when harness ∈ {minimal, standard}) and REQ-WF003-003 (`team` when harness == thorough AND team.enabled) map to a simple decision function:

```
default_mode_for_run(harness_level, team_enabled, team_env_set):
    if harness_level == "thorough" AND team_enabled AND team_env_set:
        return "team"
    elif harness_level == "thorough" AND (NOT team_enabled OR NOT team_env_set):
        # REQ-WF003-012: silent downgrade with info log
        emit_info("MODE_AUTO_DOWNGRADE: thorough harness wanted team, but prerequisites not met. Falling back to autopilot.")
        return "autopilot"
    else:
        # minimal or standard
        return "autopilot"
```

REQ-WF003-008 / 009 / 013 introduce additional mode-specific behaviors:

- `loop` mode is NEVER auto-selected by harness — it is opt-in via `--mode loop` (or `default_mode: loop` in workflow.yaml per REQ-WF003-014).
- `import` mode (design only) is NEVER auto-selected — it is opt-in via explicit `--mode import` (since it requires a Claude Design bundle path).

### 3.2 default_mode config field (REQ-WF003-014) — schema extension required

Verified via `grep -rn "default_mode" .moai/config/sections/`:
- `.moai/config/sections/research.yaml:24` — `default_mode: terminal` (different namespace; unrelated)
- `.moai/config/sections/workflow.yaml` — **NO `default_mode` field exists today**

Run-phase implementation requirement: extend `workflow.yaml` schema with a new optional field:

```yaml
workflow:
    default_mode: ""  # Optional. Values: autopilot|loop|team. Empty = harness-based auto.
```

Reading order per REQ-WF003-018 precedence (CLI > config > harness auto):
1. If user supplies `--mode` flag → use it.
2. Else if `workflow.default_mode` is set and non-empty → use it.
3. Else fall back to harness-based auto per §3.1 above.

This requires:
- Document the schema extension in `workflow.yaml` (with comment explaining valid values).
- Document the precedence in `run.md` and `design.md` skill bodies.
- (Optional, post-WF-003) Add a Go-side loader for `workflow.default_mode` if a CLI command needs to read it. Currently the slash-command path is skill-only, so the Go loader is not strictly required for WF-003 acceptance.

### 3.3 Mode precedence implementation

REQ-WF003-018 (Unwanted Behavior): `If two conflicting mode selection sources disagree (e.g., --mode autopilot + default_mode: team in workflow.yaml), then the CLI-provided --mode shall win`. This is the standard CLI > config > auto pattern. The skill body MUST document this precedence and the mode resolver pseudocode MUST follow this order strictly.

---

## 4. Sentinel Inventory and Cross-Spec Verification (CRITICAL)

WF-003 introduces two new sentinel error keys, references one shared with WF-004, and is cross-referenced by one WF-004 sentinel. Sentinel ownership table:

| Sentinel | Owner SPEC | Owner REQ | Reference REQ (other SPEC) | Trigger |
|----------|-----------|-----------|---------------------------|---------|
| `MODE_UNKNOWN` | WF-003 | REQ-WF003-010 | (none) | Invalid `--mode` value supplied (e.g., `--mode banana`) |
| `MODE_TEAM_UNAVAILABLE` | WF-003 | REQ-WF003-011 | (none) | `--mode team` requested but team prerequisites missing |
| `MODE_PIPELINE_ONLY_UTILITY` | **SHARED** (WF-003 + WF-004) | REQ-WF003-016 | REQ-WF004-014 | `--mode pipeline` on a multi-agent subcommand |
| `MODE_FLAG_IGNORED_FOR_UTILITY` | WF-004 | REQ-WF004-011 | (referenced by WF-003 §2.3 cross-table, not by REQ) | `--mode <any-value>` on a utility subcommand |

### 4.1 Cross-spec verification: `MODE_PIPELINE_ONLY_UTILITY`

**Verification command**:
```
grep -n "MODE_PIPELINE_ONLY_UTILITY" \
    .moai/specs/SPEC-V3R2-WF-003/spec.md \
    .moai/specs/SPEC-V3R2-WF-004/spec.md
```

**Verification result** (verbatim grep output):
- `.moai/specs/SPEC-V3R2-WF-003/spec.md:161` — `**If** \`--mode pipeline\` is specified on \`/moai run\` (a multi-agent subcommand per SPEC-V3R2-WF-004), **then** the system **shall** reject with \`MODE_PIPELINE_ONLY_UTILITY\` pointing to the utility subcommand set.`
- `.moai/specs/SPEC-V3R2-WF-003/spec.md:183` — `**AC-WF003-11**: Given \`/moai run --mode pipeline\` When executed Then \`MODE_PIPELINE_ONLY_UTILITY\` error is emitted (maps REQ-WF003-016).`
- `.moai/specs/SPEC-V3R2-WF-004/spec.md:155` — `**If** a multi-agent subcommand (e.g., \`plan\`) is forced into pipeline mode via \`--mode pipeline\`, **then** the system **shall** emit \`MODE_PIPELINE_ONLY_UTILITY\` (shared with REQ-WF003-016).`
- `.moai/specs/SPEC-V3R2-WF-004/spec.md:174` — `**AC-WF004-11**: Given \`/moai plan --mode pipeline\` When executed Then \`MODE_PIPELINE_ONLY_UTILITY\` error emerges (maps REQ-WF004-014).`

**Verdict**: PASS. The sentinel string `MODE_PIPELINE_ONLY_UTILITY` is byte-identical in both SPECs. WF-004's REQ-WF004-014 explicitly states "shared with REQ-WF003-016", confirming the bidirectional contract.

### 4.2 Implementation surface for sentinels

Since the slash-command path has no Go handler, sentinel emission is the orchestrator's responsibility (skill body + AskUserQuestion or info-log instruction). Static audit (similar to WF-004's `agentless_audit_test.go`) is the reasonable enforcement mechanism:

- **For `MODE_UNKNOWN`** (REQ-WF003-010): the `run.md` and `design.md` skill bodies must contain a section documenting how the orchestrator validates `--mode` values and emits this sentinel. A static audit test asserts the literal string `MODE_UNKNOWN` appears in both skills.
- **For `MODE_TEAM_UNAVAILABLE`** (REQ-WF003-011): the `run.md` skill body must contain a section documenting team prerequisite checks and the sentinel. A static audit asserts the literal string in `run.md`.
- **For `MODE_PIPELINE_ONLY_UTILITY`** (REQ-WF003-016): WF-004's plan.md M3 already adds this sentinel to all 4 implementation skills. WF-003 must NOT remove or alter this — instead, WF-003 adds a parallel rejection clause ensuring the orchestrator emits this error before reaching Phase 2A/2B.

The audit-test pattern is already established by WF-004 plan §2 M1 (`internal/template/agentless_audit_test.go` with `TestImplementationSkillsContainPipelineRejectionSentinel`). WF-003 SHOULD extend this same test file (rather than create a parallel audit test file) to add:
- `TestRunDesignSkillsContainModeUnknownSentinel` — asserts `MODE_UNKNOWN` in `run.md` and `design.md`
- `TestRunSkillContainsModeTeamUnavailableSentinel` — asserts `MODE_TEAM_UNAVAILABLE` in `run.md`

This consolidates all mode-related static audits into one test file, matching the WF-004 architecture decision.

---

## 5. `/moai loop` Wrapper Pattern (REQ-WF003-004) and Thin-Wrapper Discipline

### 5.1 Current `/moai loop` thin-wrapper status

`.claude/commands/moai/loop.md:1-7` (verified verbatim):
```
---
description: Iteratively fix issues until all resolved or max iterations reached
argument-hint: "[--max N] [--auto-fix] [--seq]"
allowed-tools: Skill
---

Use Skill("moai") with arguments: loop $ARGUMENTS
```

This is **already a thin wrapper** (7 lines total, body = single skill invocation). Per REQ-WF003-004 ("`/moai loop` shall be an alias resolving to `/moai run --mode loop` with identical arguments"), the wrapper itself does NOT need rewriting — the alias semantics are achieved at the **skill orchestration layer**, not the command file layer.

Implementation interpretation:
- `/moai loop $ARGUMENTS` → currently routes to `Skill("moai") with arguments: loop $ARGUMENTS` → orchestrator loads `.claude/skills/moai/workflows/loop.md`.
- After WF-003: `/moai loop $ARGUMENTS` → still routes to `Skill("moai") with arguments: loop $ARGUMENTS` → orchestrator loads `.claude/skills/moai/workflows/loop.md` → `loop.md` body now contains a header note: "This skill is also reachable as `/moai run --mode loop` per SPEC-V3R2-WF-003 REQ-WF003-004. Both invocation paths produce identical behavior."

The `/moai run --mode loop` path is the NEW invocation route. When `run.md` parses `--mode loop`, it MUST delegate to `Skill("moai-workflow-loop")` (which is the same target as the `/moai loop` route). Mechanism to be implemented in `run.md` Phase 0.95 routing.

### 5.2 Constraint check: 20 LOC thin-wrapper rule

spec.md §7 Constraints: "`/moai loop` wrapper는 20 LOC thin-wrapper 규격 유지." Current `/moai loop` command file is 7 lines (well under 20). After WF-003, no changes to the command file are planned — the file remains ≤ 20 LOC. Constraint satisfied.

### 5.3 alias behavior identity (REQ-WF003-004)

REQ-WF003-004 mandates "identical arguments". This means:
- `/moai loop SPEC-001 --max 50 --auto-fix` MUST behave identically to `/moai run SPEC-001 --mode loop --max 50 --auto-fix`.
- Both routes invoke `loop.md` with the same argument set (positional + flags).
- The CI test (audit-content style) should verify both routes are documented as equivalent in their respective skill bodies.

A behavioral-equivalence test (running both invocations on the same fixture) is the strongest verification, but adds runtime test complexity. As an interim, document equivalence in skill bodies and rely on static cross-referencing.

---

## 6. Team Mode Prerequisites and Silent Downgrade (REQ-WF003-011, REQ-WF003-012)

### 6.1 Team mode prerequisite check

Per `.claude/skills/moai/workflows/run.md:927-943` (Team Mode Routing section):

```
1. Verify prerequisites: workflow.team.enabled == true AND CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1 env var is set
2. If prerequisites met: Read ${CLAUDE_SKILL_DIR}/team/run.md and execute the team workflow
3. If prerequisites NOT met: Warn user then fallback to standard sub-agent mode
```

This existing check IS the foundation for REQ-WF003-011 and REQ-WF003-012:

- REQ-WF003-011 (Event-Driven, error case): When user EXPLICITLY supplies `--mode team` AND prerequisites missing → emit `MODE_TEAM_UNAVAILABLE` error and suggest `--mode autopilot` fallback. This is an **explicit failure** (user opted into team mode and we cannot honor it).
- REQ-WF003-012 (State-Driven, downgrade case): When mode auto-selection (harness-based) DECIDES `team` BUT prerequisites missing → silently downgrade to `autopilot` and emit info log. This is an **implicit fallback** (no user opt-in).

**Difference is intent**:
- Explicit user request (`--mode team`): treat as user-facing error (REQ-WF003-011).
- Auto-selection: treat as informational (REQ-WF003-012).

Implementation requires the orchestrator to track WHERE the `team` selection came from (CLI flag vs auto-resolved from harness) and route to the appropriate sentinel.

### 6.2 workflow.yaml team prerequisites verification

Verified `.moai/config/sections/workflow.yaml:17-26`:
```yaml
team:
    auto_selection:
        min_complexity_score: 7
        min_domains_for_team: 3
        min_files_for_team: 10
    default_model: opus[1m]
    delegate_mode: true
    enabled: true
    max_teammates: 10
```

`team.enabled: true` is the current default. `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` is an env-var check at runtime. Both must be satisfied for `team` mode to activate.

### 6.3 Info log specification for downgrade (REQ-WF003-012)

The exact info-log message format is not specified in spec.md. Recommended format (for plan.md M-target):

```
[mode-auto-downgrade] Harness level=thorough requested team mode, but prerequisites are not satisfied (workflow.team.enabled=<bool>, CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=<set/unset>). Falling back to autopilot mode. To use team mode, set workflow.team.enabled=true and export CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1.
```

Sentinel string for static audit: `MODE_AUTO_DOWNGRADE` (proposed; not in spec.md — orchestration internal). The skill body MUST document this downgrade behavior so the user understands why team mode did not activate.

---

## 7. Path A (Claude Design Import) and Path B (Code-Based Design) Mode Mapping

### 7.1 Existing path A skill: moai-workflow-design-import

Verified WF-001 24-skill catalog status (`.moai/specs/SPEC-V3R2-WF-001/spec.md:5,254,258`):
- WF-001 status: **completed**
- `moai-workflow-loop` (KEEP — Ralph) — row 16
- `moai-workflow-design-import` (KEEP — Path A handoff) — row 20
- `moai-domain-copywriting` (KEEP — FROZEN agency contract) — row 27
- `moai-domain-brand-design` (KEEP — FROZEN agency contract) — row 28

All four skills WF-003 needs are present and active in the post-WF-001 catalog. **No new skills are required by WF-003.** Constraint per spec.md §7 ("SPEC-V3R2-WF-001 24-skill 카탈로그 준수: 신규 skill 추가 금지") is satisfied.

### 7.2 Path A invocation pattern (REQ-WF003-013)

REQ-WF003-013 (State-Driven): `While /moai design is executing with --mode import, the system shall skip path B (copywriting + brand-design) and only run moai-workflow-design-import for handoff bundle parsing.`

Current `design.md:111` (Step A3): `Invoke moai-workflow-design-import skill: Pass: bundle file path, project brief, .moai/config/sections/design.yaml. Expected output: .moai/design/tokens.json, .moai/design/components.json, .moai/design/assets/`.

Implementation: when `--mode import` is parsed, the orchestrator skips Phase 1 path selection (which today asks user to pick A/B1/B2) and goes straight to Step A2 (collect bundle path) → Step A3 (invoke design-import). It also SKIPS Phase B-Common (Step BC-3 which loads moai-domain-copywriting + moai-domain-brand-design + moai-workflow-gan-loop), satisfying REQ-WF003-013's "skip path B" clause.

### 7.3 Path B + team mode (REQ-WF003-009)

REQ-WF003-009 (Event-Driven): `When a user runs /moai design --mode team, the system shall spawn moai-domain-copywriting and moai-domain-brand-design teammates in parallel per the GAN Loop contract.`

Current `design.md:183-186` (Step BC-3): `Load code-based design skills: Load moai-domain-copywriting; Load moai-domain-brand-design; Load moai-workflow-gan-loop`.

In autopilot mode, these skills are loaded sequentially (single-lead). In team mode, they are spawned as parallel teammates via `TeamCreate`. Implementation:
- Team prerequisite check (per §6.1) MUST run first.
- If prerequisites satisfied: spawn `moai-domain-copywriting` teammate + `moai-domain-brand-design` teammate in parallel via `TeamCreate` with `role_profiles.designer` (per workflow.yaml). Both teammates feed into `moai-workflow-gan-loop` for evaluation.
- If prerequisites not satisfied: route per REQ-WF003-011 (explicit `--mode team` request → `MODE_TEAM_UNAVAILABLE` error).

Reference for team spawn pattern: `.claude/rules/moai/workflow/worktree-integration.md:133` — "Implementation teammates in team mode (role_profiles: implementer, tester, designer) MUST use isolation: worktree when spawned via Agent()". The `designer` role profile from workflow.yaml:52-56 (`mode: acceptEdits, model: sonnet, isolation: worktree`) is the canonical fit for both copywriting and brand-design teammates.

---

## 8. Subcommand × Mode Matrix (cross-ref WF-004 publication target)

The matrix to be published in `.claude/rules/moai/workflow/spec-workflow.md` (per WF-004 plan §2 M4) MUST be EXTENDED by WF-003 to specify which `--mode` values each multi-agent subcommand accepts. The combined post-WF-003+WF-004 matrix:

| Subcommand   | Class          | Default mode | `--mode` honored? | Valid `--mode` values | Sentinel on invalid mode |
|--------------|----------------|--------------|-------------------|-----------------------|-------------------------|
| `/moai plan`     | Multi-Agent    | autopilot (fixed) | NO (REQ-WF003-005) | (none — flag ignored, but `pipeline` rejected) | `MODE_PIPELINE_ONLY_UTILITY` (if `pipeline`) |
| `/moai run`      | Multi-Agent    | per harness/config | YES | `autopilot`, `loop`, `team` | `MODE_UNKNOWN` (other), `MODE_TEAM_UNAVAILABLE` (team without prereqs), `MODE_PIPELINE_ONLY_UTILITY` (`pipeline`) |
| `/moai sync`     | Multi-Agent    | autopilot (fixed) | NO (REQ-WF003-005) | (none — flag ignored, but `pipeline` rejected) | `MODE_PIPELINE_ONLY_UTILITY` (if `pipeline`) |
| `/moai design`   | Multi-Agent    | per harness/config | YES | `autopilot`, `import`, `team` | `MODE_UNKNOWN` (other), `MODE_TEAM_UNAVAILABLE` (team without prereqs), `MODE_PIPELINE_ONLY_UTILITY` (`pipeline`) |
| `/moai loop`     | Multi-Agent (alias) | n/a (alias for `run --mode loop`) | (alias) | (alias) | (delegated to `/moai run` validation) |
| `/moai fix`      | Pipeline (Agentless) | n/a (no mode axis) | NO (REQ-WF004-011) | (none — all flags ignored with info log) | `MODE_FLAG_IGNORED_FOR_UTILITY` (info log only) |
| `/moai coverage` | Pipeline | n/a | NO | (none) | `MODE_FLAG_IGNORED_FOR_UTILITY` |
| `/moai mx`       | Pipeline | n/a | NO | (none) | `MODE_FLAG_IGNORED_FOR_UTILITY` |
| `/moai codemaps` | Pipeline | n/a | NO | (none) | `MODE_FLAG_IGNORED_FOR_UTILITY` |
| `/moai clean`    | Pipeline | n/a | NO | (none) | `MODE_FLAG_IGNORED_FOR_UTILITY` |

Implementation requirement: the WF-004 plan M4 inserts the basic 9-row matrix; the WF-003 run phase EXTENDS that matrix (adds the columns "Default mode", "Valid `--mode` values", "Sentinel on invalid mode") and adds the `/moai loop` alias row.

Coordination with WF-004 run phase: WF-004 lands its matrix first (per spec.md §9 dependencies — WF-003 is blocked-by WF-004). WF-003 run phase EDITS the existing matrix to add the new columns and the loop row.

---

## 9. Reference Patterns to Preserve (Read-Only)

The run phase MUST NOT modify the following files — they encode the rhythm WF-003 is integrating with:

- **Reference**: `.claude/skills/moai/workflows/loop.md:46-138` — Per-iteration cycle (Steps 1-9). This IS the Ralph engine. WF-003's `--mode loop` route delegates here unchanged.
- **Reference**: `.claude/skills/moai/workflows/run.md:361-382` — Phase 0.95 Scale-Based Execution Mode Selection. This is the EXISTING mode infrastructure that WF-003 augments with the explicit `--mode` flag.
- **Reference**: `.claude/skills/moai/workflows/run.md:927-943` — Team Mode Routing. This is the EXISTING team-prereq check that REQ-WF003-011/012 codifies into named sentinels.
- **Reference**: `.claude/skills/moai/workflows/design.md:64-94` — Phase 1 Path Selection AskUserQuestion. WF-003 makes this conditional (skip when explicit `--mode` provided).
- **Reference**: `.claude/skills/moai/workflows/design.md:111` — Step A3 invokes `moai-workflow-design-import`. WF-003 reuses this exact invocation under `--mode import`.
- **Reference**: `.claude/skills/moai/workflows/design.md:183-186` — Step BC-3 loads copywriting + brand-design + gan-loop. WF-003 reuses this under `--mode autopilot` (sequential) and `--mode team` (parallel via TeamCreate).
- **Reference**: `.moai/config/sections/workflow.yaml:17-26` — team configuration block. WF-003 extends this section with optional `default_mode` field but does NOT modify the `team:` subkeys.
- **Reference**: `.moai/config/sections/harness.yaml:5-103` — Harness depth configuration. WF-003 reads this for auto-selection but does NOT modify it.
- **Reference**: `.claude/rules/moai/workflow/spec-workflow.md` — Phase Overview table at lines 9-15. WF-004 plan M4 extends this with `## Subcommand Classification`. WF-003 EDITS the new section (does not create a duplicate).
- **Reference**: `internal/template/commands_audit_test.go:1-60` — Thin Command Pattern audit scaffold. WF-003 audit tests follow the same pattern, extending `internal/template/agentless_audit_test.go` (created by WF-004 M1).

---

## 10. Risks (extends `spec.md` §8 with run-phase mitigations)

| Risk | Probability | Impact | File:Line Mitigation Anchor |
|------|-------------|--------|-----------------------------|
| Mode auto-selection picks team but env not set, user not informed | M | M | REQ-WF003-012 mandates info log; `run.md` Phase 0.95 routing must emit explicit downgrade message per §6.3 above |
| Path B1 (Figma) and Path B2 (Pencil) excluded from `--mode` axis confuses users | M | M | Recommendation §2.2.3: `autopilot` mode preserves AskUserQuestion path selection; explicit `--mode autopilot` skips selection and goes straight to Path B-Common. Document in `design.md` skill body that `--mode autopilot` is "Path B (code-based) only"; users wanting Figma/Pencil should NOT supply `--mode` flag. |
| `default_mode` config field added without Go-side loader | L | L | Skill-only consumption is sufficient for MVP; Go loader can be added in SPEC-V3R2-MIG-003 (per spec.md §9.2 blocks list) |
| `/moai loop` alias diverges in behavior from `/moai run --mode loop` over time | M | M | Static audit test cross-references both invocation paths' skill bodies; behavioral equivalence noted in CI documentation |
| Mode precedence ambiguity in conflicting sources | L | M | REQ-WF003-018 hard-codes precedence (CLI > config > harness auto); skill body MUST document this with example |
| Sentinel string drift between WF-003 and WF-004 implementations | L | H | Static audit `TestImplementationSkillsContainPipelineRejectionSentinel` (added in WF-004 M1) already enforces literal string; WF-003 extends same test file — no separate audit needed |
| Run phase touches `feedback`/`review`/`e2e`/`fix`/`coverage`/`mx`/`codemaps`/`clean` skills accidentally | L | M | spec.md §1.2 explicitly excludes these; tasks.md M-targets enumerate exactly the 5 affected skills (run, loop, design, plan, sync) — utility skills are WF-004's domain |
| `--mode team` for `/moai design` requires teammate role_profile not yet defined | L | M | workflow.yaml:52-56 already defines `designer` role_profile; copywriting+brand-design teammates use this profile (per spec.md §2.2 in-scope clause "기존 workflow.yaml role_profiles 재사용") |
| Stacked PR base transition timing (WF-004 merge before WF-003 base-switch) | M | M | plan.md will document pre-merge base-switch hook (CLAUDE.local.md §18.11 case study) — switch base from `feature/SPEC-V3R2-WF-004` → `main` BEFORE PR #765 lands |

---

## 11. Recommendations for the Run Phase

1. **Methodology**: Per `.moai/config/sections/quality.yaml development_mode: tdd`, this project uses TDD. Following WF-004's pattern, audit tests for sentinel presence MUST be written first (RED). Specifically, extend `internal/template/agentless_audit_test.go` (created in WF-004 M1) with two new test functions per §4.2 above. Confirm RED, then add the sentinel content + dispatch logic to skill bodies (GREEN).

2. **Worktree discipline**: All run-phase work continues in `/Users/goos/.moai/worktrees/moai-adk/SPEC-V3R2-WF-003`. Branch `feature/SPEC-V3R2-WF-003` is already created with base `feature/SPEC-V3R2-WF-004` (HEAD `5ab409292`). Worktree absolute path is the **only** legal write surface.

3. **Stacked PR protocol**: Before PR #765 (WF-004) merges to main, the WF-003 PR's base MUST be transitioned from `feature/SPEC-V3R2-WF-004` to `main`. Per CLAUDE.local.md §18.11 case study (v2.14.0 stacked PR auto-close incident), this transition prevents the WF-003 PR from being auto-closed when its parent merges.

4. **MX targets**: Pre-mark the new `## Mode Dispatch` sections in `run.md` and `design.md` with `@MX:NOTE` documenting why mode axis exists. Pre-mark the mode resolver function (when implemented) with `@MX:ANCHOR` (high fan_in: every `/moai run` invocation passes through this resolver).

5. **Quality gate**: Per `.claude/rules/moai/workflow/spec-workflow.md:172-204` Phase 0.5 Plan Audit Gate, plan-auditor will audit this research + plan + acceptance + tasks before the implementation phase begins. Treat each artifact as audit-ready output.

6. **No new skills**: WF-001 24-skill catalog stability requires that WF-003 reuse existing skills. Verified §7.1 above — all four targets (`moai-workflow-loop`, `moai-workflow-design-import`, `moai-domain-copywriting`, `moai-domain-brand-design`) are present and active. The run phase MUST NOT create any new skill directories.

7. **BC-WF003 scope check**: spec.md frontmatter `breaking: false` and `bc_id: []`. The `--mode` flag is purely additive — users not supplying `--mode` get exactly today's behavior (Phase 0.95 auto-selection on `/moai run`, AskUserQuestion path selection on `/moai design`, direct Ralph engine on `/moai loop`). REQ-WF003-005 (plan/sync ignore) is also additive (plan/sync today don't accept `--mode`; they continue not to). No existing user-facing behavior changes.

8. **Cross-SPEC sync**: WF-003 BLOCKS-BY WF-004. Once WF-004 PR #765 merges, the `## Subcommand Classification` matrix in `spec-workflow.md` exists; WF-003 run phase EDITS that matrix (adds columns + loop alias row) rather than creating a parallel matrix.

---

## 12. v3 Master design cross-anchors

Per `spec.md` §10 traceability, the master design document references WF-003 in the same Phase 6 Multi-Mode Workflow section as WF-004:

- **`docs/design/major-v3-master.md:L993`** — §9 Phase 6 Multi-Mode Workflow: positions WF-003 and WF-004 as sibling deliverables sharing the `--mode` flag matrix.
- **Pattern source**: `.moai/design/v3-redesign/synthesis/pattern-library.md` §O-4 (Multi-Mode Router, R2 §3 OMC reduction).
- **Research source**: `.moai/design/v3-redesign/research/r2-opensource-tools.md` §3 (OMC 6-mode design — WF-003 adopts 4 of 6).

Pattern-library §O-4 and R2 §3 are reproduced inline in spec.md §1.1 and need no further extraction.

---

## 13. Citation Summary (file:line anchors used)

1. `.moai/specs/SPEC-V3R2-WF-003/spec.md:1-249` (entire SPEC, contract source)
2. `.moai/specs/SPEC-V3R2-WF-004/spec.md:144,155,173,174` (sentinel cross-reference)
3. `.moai/specs/SPEC-V3R2-WF-001/spec.md:5,209,214,216,254,258,265,266` (24-skill catalog status: completed)
4. `.moai/specs/SPEC-V3R2-WF-002/spec.md:1-30` (related thin-wrapper SPEC, MERGED #761)
5. `.claude/commands/moai/run.md:1-7` (thin wrapper)
6. `.claude/commands/moai/loop.md:1-7` (thin wrapper, REQ-WF003-004 base)
7. `.claude/commands/moai/design.md:1-7` (thin wrapper)
8. `.claude/commands/moai/plan.md:1-7` (thin wrapper, mode-NA)
9. `.claude/skills/moai/workflows/run.md:46` (--team flag — existing infrastructure)
10. `.claude/skills/moai/workflows/run.md:64-83` (Harness Level Routing — mode auto-select substrate)
11. `.claude/skills/moai/workflows/run.md:361-382` (Phase 0.95 Scale-Based Mode Selection — substrate)
12. `.claude/skills/moai/workflows/run.md:537-545` (Development Mode Routing — DDD/TDD selection)
13. `.claude/skills/moai/workflows/run.md:602-686` (Phase 2A/2B implementation routing)
14. `.claude/skills/moai/workflows/run.md:920-922` (--team/--solo flag junction)
15. `.claude/skills/moai/workflows/run.md:927-943` (Team Mode Routing prerequisite check)
16. `.claude/skills/moai/workflows/loop.md:1-249` (entire skill — Ralph engine)
17. `.claude/skills/moai/workflows/loop.md:46-138` (per-iteration cycle Steps 1-9)
18. `.claude/skills/moai/workflows/loop.md:106-117` (per-issue-type fix agent dispatch)
19. `.claude/skills/moai/workflows/loop.md:140-152` (convergence and exit conditions)
20. `.claude/skills/moai/workflows/design.md:64-94` (Phase 1 path selection AskUserQuestion)
21. `.claude/skills/moai/workflows/design.md:97-123` (Phase A: Claude Design Import)
22. `.claude/skills/moai/workflows/design.md:111` (Step A3: invoke moai-workflow-design-import)
23. `.claude/skills/moai/workflows/design.md:167-192` (Phase B-Common: code-based path)
24. `.claude/skills/moai/workflows/design.md:183-186` (Step BC-3: load copywriting + brand-design + gan-loop)
25. `.claude/skills/moai/workflows/plan.md:1-805` (mode-NA verification)
26. `.claude/skills/moai/workflows/sync.md:1-1178` (mode-NA verification)
27. `.claude/rules/moai/workflow/spec-workflow.md:9-17` (Phase Overview table — WF-004 matrix insertion anchor)
28. `.claude/rules/moai/workflow/worktree-integration.md:133` (designer role isolation: worktree)
29. `.moai/config/sections/harness.yaml:5-103` (harness levels + auto-detection rules)
30. `.moai/config/sections/harness.yaml:67-103` (level definitions: minimal/standard/thorough)
31. `.moai/config/sections/workflow.yaml:17-26` (team configuration)
32. `.moai/config/sections/workflow.yaml:42-61` (role_profiles: designer, implementer, tester, etc.)
33. `.moai/config/sections/quality.yaml:1-3` (development_mode: tdd for run phase)
34. `internal/template/commands_audit_test.go:1-60` (audit-test scaffold pattern, WF-004 ref)
35. `internal/template/agentless_audit_test.go` (NEW per WF-004 M1; WF-003 EXTENDS this file)
36. `.moai/specs/SPEC-V3R2-WF-004/plan.md:39-43` (M3 mode-flag rejection clauses — WF-003 builds on this)
37. `.moai/specs/SPEC-V3R2-WF-004/plan.md:140-189` (M4 Subcommand Classification matrix — WF-003 extends columns)
38. `docs/design/major-v3-master.md:L993` (§9 Phase 6 Multi-Mode Workflow anchor)

Total distinct file:line citations: **38** (exceeds the §Hard-Constraints minimum of 10 for research.md).

---

End of research.md.
