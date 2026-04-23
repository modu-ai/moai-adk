# B2 — Utility Subcommands Audit (/moai fix / loop / codemaps)

**Audit Date**: 2026-04-24
**Auditor**: B2 (Utility Subcommands Team)
**Governing SPEC**: SPEC-V3R2-WF-004 (Agentless Fixed-Pipeline Classification)
**Related SPEC**: SPEC-V3R2-WF-003 (Multi-Mode Router)

---

## Executive Summary

### Subcommands Audited: 3

| Subcommand | Overall Health | Agentless Conformance | Ralph Engine LOC |
|------------|---------------|----------------------|-----------------|
| `/moai fix` | **7/10** | Partial | N/A (skill-only) |
| `/moai loop` | **8/10** | Partial | 337 (engine.go) |
| `/moai codemaps` | **5/10** | No | N/A (skill-only) |

### Ralph Engine Key Files

| File | LOC | Role |
|------|-----|------|
| `internal/ralph/engine.go` | 337 | Decision engine + LSP classifier |
| `internal/loop/controller.go` | 394 | Loop lifecycle + goroutine orchestration |
| `internal/loop/state.go` | 166 | State machine + interfaces |
| `internal/loop/storage.go` | 117 | Atomic JSON persistence |
| `internal/loop/feedback.go` | 86 | Stagnation + quality gate predicates |
| `internal/loop/go_feedback.go` | 260 | Go toolchain feedback collector |
| `internal/loop/feedback_channel.go` | 81 | Bounded event channel |
| **Total (loop + ralph)** | **~1,578** | |

### Top 5 Issues

1. **Codemaps violates Agentless contract**: Phases 1-3 explicitly delegate to `Explore` and `manager-docs` subagents, using LLM-driven control flow in what SPEC-V3R2-WF-004 classifies as a fixed-pipeline command (REQ-WF004-004).

2. **Fix workflow: agent delegation for repair violates strict Agentless interpretation**: The repair phase in `/moai fix` calls specialized subagents (expert-backend, expert-refactoring, expert-debug) for every fix level, contradicting REQ-WF004-004 which requires "only deterministic tool invocations."

3. **No-op path not implemented**: REQ-WF004-007 requires that when localize finds zero targets, the pipeline exits with status "no-op" and exit code 0. Neither `fix.md` nor `codemaps.md` implements this explicitly; the loop workflow partially handles it only through completion condition checks (Step 4 of loop.md).

4. **RalphEngine stagnation uses simple `IsStagnant`, not the more capable `IsStagnantWithDiagnostics`**: `engine.go:64` calls `loop.IsStagnant()` (integer metrics only), ignoring the richer diagnostic-trend-aware function `IsStagnantWithDiagnostics()` defined in `feedback.go:13`. This means LSP diagnostic trend data is not used in the convergence decision, leaving information on the table.

5. **No per-project config schema for `/moai fix` or `/moai codemaps`**: `ralph.yaml` covers loop/LSP settings but lacks dedicated sections for fix-specific (max level, security scan on/off) or codemaps-specific (depth, format, incremental vs full) parameters. Users must pass flags; no persistent project-level overrides exist.

### Top 3 Critical Fixes

1. **Separate codemaps into an LLM-aided but deterministic pipeline**: Replace open-ended Explore/manager-docs agent delegation with deterministic tool calls (Glob, Grep, Read for localize; Write for repair; file existence/consistency check for validate). The analysis intelligence can remain LLM-powered, but the phase structure and control flow must be fixed, not agent-decided.

2. **Implement no-op exit for all three workflows**: Add an explicit check after the localize phase: if zero targets found, emit `status: no-op`, exit code 0, and skip repair + validate. This closes the AC-WF004-07 gap across all three subcommands.

3. **Wire `IsStagnantWithDiagnostics` into RalphEngine**: Replace `loop.IsStagnant(prev, feedback)` at `engine.go:64` with `loop.IsStagnantWithDiagnostics(prev, feedback)` to leverage diagnostic trend data in convergence decisions, preventing false convergence when LSP errors are climbing while integer metrics are flat.

---

## /moai fix Audit

### Thin Command Wrapper (D1 compliance)

**File**: `.claude/commands/moai/fix.md` (7 lines)

The wrapper is fully compliant with the thin-wrapper pattern (SPEC-THIN-CMDS-001). It routes to `Skill("moai")` with `fix $ARGUMENTS` in 7 lines total, well within the 20 LOC limit.

```
allowed-tools: Skill
Use Skill("moai") with arguments: fix $ARGUMENTS
```

### Workflow Skill: fix.md v2.5.0

**File**: `.claude/skills/moai/workflows/fix.md` (293 lines)

#### D1 — Agentless Pipeline Conformance

**Localize phase (Phase 1)**: Well-defined. Three parallel scanners (LSP, AST-grep, Linter) run via Bash with `run_in_background`. Output is normalized into a unified structured issue record (`file`, `line`, `column`, `severity`, `code`, `message`, `source`, `language`). This phase is deterministic and agentless. **Conformant.**

**Repair phase (Phase 3)**: **Partially non-conformant.** The fix workflow uses specialized subagents for repair:
- Level 1 (formatting): `expert-backend` or `expert-frontend` subagent
- Level 2 (rename, type): `expert-refactoring` subagent
- Level 3 (logic, API): `expert-debug` or `expert-backend` subagent

This directly invokes `Agent()` for repair, contradicting REQ-WF004-004 ("only deterministic tool invocations permitted"). However, the SPEC's own definition at §4 acknowledges that "Agentless pipeline is LLM-enabled but control flow is fixed." The question is whether agent delegation for repair constitutes LLM-driven control flow or LLM-driven execution. The spirit of REQ-WF004-004 is to prevent the agent from deciding what to do (localize/repair/validate ordering); execution via sub-agents is arguably separate. **Contested partial conformance.**

**Validate phase (Phase 4)**: Re-runs affected diagnostics on modified files, confirms fixes resolved issues, and detects regressions. This is deterministic tool execution. **Conformant.**

**LLM control flow**: The phase ordering itself (localize → repair → validate) is fixed and not LLM-decided. The use of agents within repair is execution delegation, not flow control. The `--dry` flag, severity levels, and task creation are all deterministic. **Phase ordering is agentless; repair execution uses agents.**

#### D3 — Fix Engine Details

**16-language LSP detection** (`fix.md:50-71`): Complete matrix. All 16 MoAI-supported languages are covered with indicator files and corresponding LSP commands. Language detection is based on indicator file presence.

**Error classification** (`fix.md:127-132`):
- Level 1 (Immediate): No approval, auto-apply. Examples: import sorting, whitespace, formatting.
- Level 2 (Safe): Log only, auto-apply. Examples: rename variable, add type annotation.
- Level 3 (Review): User approval via AskUserQuestion required. Examples: logic changes, API modifications.
- Level 4 (Manual): Auto-fix not allowed. Examples: security vulnerabilities, architecture changes.

This classification maps well to the Go-layer `ErrorLevel` constants in `internal/ralph/engine.go:96-121`:
- `ErrorLevelAutoFix (1)` = Level 1 in skill
- `ErrorLevelLogOnly (2)` = Level 2 in skill
- `ErrorLevelApproval (3)` = Level 3 in skill
- `ErrorLevelManual (4)` = Level 4 in skill
- `ErrorLevelBlocker (5)` = above Level 4, no auto-fix

**Gap**: The skill-level fix classification (Levels 1-4) and the Go-layer `ErrorLevel` (1-5) are parallel but not explicitly cross-referenced. `ErrorLevelBlocker` (compiler errors) has no direct mapping in the skill's Level system. A compiler-level blocker in the Go code would fall to Level 4 (Manual) in the skill — this is correct behavior but undocumented.

**Safety rails**:
- Reproduction-first: Phase 3 references `CLAUDE.md Section 7 Safe Development Protocol`. The skill states "Write a failing test that reproduces the bug before fixing" for Level 3+ fixes.
- MX Context Scan (Phase 2.5): Before repair, scans target files for `@MX:ANCHOR`, `@MX:WARN`, `@MX:NOTE`, `@MX:TODO` tags. Anchor functions are passed as "do not break" constraints to fix agents.
- Snapshot-based resume: Snapshots saved to `$CLAUDE_PROJECT_DIR/.moai/cache/fix-snapshots/`.

**Missing safety rail**: No backup is taken before destructive modifications. The SPEC defines no `.moai/backups/` usage requirement, but D6 asks about this. Rollback capability relies on git history only.

**Idempotency gap (D8)**: If no issues are found during Phase 1, the workflow does not explicitly state "exit with no-op status." The flow would continue through Classification (empty list), Pre-Fix MX Context Scan (no targets), and reach the reporting phase with no changes. This implicitly handles it but does not implement the explicit "no-op" exit specified in REQ-WF004-007.

#### D5 — Performance

- Phase 1 parallelizes 3 scanners via background Bash: expected ~8s vs ~30s sequential (claimed 3-4x speedup in skill body).
- Sequential fallback via `--sequential` flag.
- No latency target documented.

#### D6 — Safety + Reversibility

- `--dry` flag provides preview mode without applying changes.
- No `.moai/backups/` usage documented.
- Rollback relies on git state; no explicit rollback on failure path.

#### D7 — Observability

- TaskCreate/TaskUpdate tracking for all discovered issues.
- Report generated with file:line evidence of changes.
- MX Tag Report section generated after Phase 4.5.
- No structured log artifact written to `.moai/logs/`.

#### D9 — Configuration Gaps

`ralph.yaml` covers LSP and loop settings. Fix-specific configuration missing:
- No YAML knob for default `--level` (max fix level).
- No per-language fix tolerance configuration.
- `ast_grep.auto_fix: false` in ralph.yaml is present but the fix workflow doesn't explicitly reference it.

**Fix audit verdict: 7/10** — Solid localize and validate phases. Repair uses agent delegation which is contested for Agentless conformance. Missing no-op exit, backup mechanism, and fix-specific config schema.

---

## /moai loop Audit (Ralph Engine Deep-Dive)

### Thin Command Wrapper

**File**: `.claude/commands/moai/loop.md` (7 lines)

Fully compliant thin wrapper. Routes to `Skill("moai")` with `loop $ARGUMENTS`.

**Note per SPEC-V3R2-WF-003 REQ-WF003-004**: `/moai loop` is intended to become an alias for `/moai run --mode loop`. The current thin wrapper delegates to the moai skill's `loop` subcommand handler, not explicitly to `/moai run --mode loop`. This alias relationship is not yet implemented.

### Workflow Skill: loop.md v2.5.0

**File**: `.claude/skills/moai/workflows/loop.md` (250 lines)

#### D1 — Agentless Pipeline Conformance

The loop workflow is structurally the most complex of the three. SPEC-V3R2-WF-004 §4 explicitly maps the loop analog as:
- localize → error discovery (Step 3 parallel diagnostics)
- repair → fix execution (Step 6)
- validate → verification (Step 7)

**Localize (Step 3)**: Four parallel diagnostic tools via background Bash (LSP, AST-grep, test runner, coverage). Aggregated into unified report. **Deterministic, conformant.**

**Repair (Step 6)**: Same pattern as fix.md — uses specialized subagents (`expert-debug`, `expert-testing`, `expert-security`, `expert-performance`, etc.). The `[HARD]` mandate states: "Agent delegation mandate: ALL fix tasks MUST be delegated to specialized agents. NEVER execute fixes directly." This is a direct contradiction of REQ-WF004-004 at face value. **Contested partial conformance.**

**Validate (Step 7)**: TaskUpdate to completed. Re-running full diagnostics happens at the next iteration's Step 3 rather than a dedicated per-fix validate step. **Implicit, not explicit.**

**LLM control flow**: The iteration structure (Steps 1-9) is fixed. The decision to continue vs. exit is made by convergence conditions (zero errors + tests passing + coverage), not LLM reasoning. The loop exit is deterministic. **Phase ordering is agentless.**

#### D2 — Loop Engine Architecture (Ralph Engine Deep-Dive)

##### State Machine (`internal/loop/state.go:27-47`)

Four-phase cycle: `PhaseAnalyze → PhaseImplement → PhaseTest → PhaseReview`. Phase transitions are validated via `validTransitions` map. `NextPhase()` returns the valid successor or defaults to `PhaseAnalyze` on unknown input.

```
PhaseAnalyze -> PhaseImplement -> PhaseTest -> PhaseReview -> PhaseAnalyze
```

##### RalphEngine Decision Priority (`internal/ralph/engine.go:29-89`)

```
1. Max iterations reached → abort
2. Quality gate satisfied (zero failures + zero lint + build success + 85% coverage) → converge
3. Stagnation detected (auto_converge enabled) → converge
4. Human review requested + PhaseReview → request_review
5. Default → continue
```

This is a clean priority cascade. Each check is independent. No LLM reasoning involved in the decision logic.

##### Convergence Detection (`internal/loop/feedback.go:54-61`)

`IsStagnant(prev, curr)` checks: `TestsFailed == prev.TestsFailed && LintErrors == prev.LintErrors && Coverage == prev.Coverage`.

**Critical gap**: Engine uses `loop.IsStagnant` (`engine.go:64`), which checks only integer metrics. The more complete `IsStagnantWithDiagnostics` (`feedback.go:13`) additionally checks `len(LSPDiagnostics)` trend (decreasing = not stagnant). This function is tested in `stagnation_diag_test.go` but **not used by the engine**. The engine therefore misses cases where integer metrics are flat but LSP errors are increasing — a real stagnation signal.

```go
// engine.go:64 — uses IsStagnant
if prev != nil && loop.IsStagnant(prev, feedback) {

// feedback.go:13 — IsStagnantWithDiagnostics also checks diagnostic trends
// NOT called by engine
```

##### Quality Gate (`internal/loop/feedback.go:66-74`)

```go
func MeetsQualityGate(fb *Feedback) bool {
    return fb.TestsFailed == 0 &&
        fb.LintErrors == 0 &&
        fb.BuildSuccess &&
        fb.Coverage >= float64(config.DefaultTestCoverageTarget) // 85%
}
```

The quality gate is hard-wired to `config.DefaultTestCoverageTarget` (85%). The `ralph.yaml` config allows `completion.coverage_threshold: 85` but the Go code does not read this value — it uses the compile-time default. **Configuration drift: runtime config not wired into quality gate.**

##### FeedbackChannel (`internal/loop/feedback_channel.go:14-81`)

Bounded channel (capacity 64 by default). Overflow drops oldest event with slog.Warn. Uses double-select pattern for safe drain-and-insert. The `@MX:WARN` tag at `feedback_channel.go:14` correctly identifies the bounded drop semantics.

**Risk**: If the loop produces feedback faster than the controller consumes it (unlikely in typical use), older feedback events are silently dropped. This could cause missed stagnation signals in high-frequency scenarios.

##### LoopController (`internal/loop/controller.go:29-394`)

The `@MX:ANCHOR` tag on `NewLoopController` (fan_in=18+) is appropriate. The goroutine-based `runLoop` is tagged with `@MX:WARN` for orphan goroutine risk, also appropriate.

**Failure handling when iteration worsens**: The engine has no explicit "regression detection" path. If `TestsFailed` increases from one iteration to the next, the engine will:
1. Not meet the quality gate → continue
2. Not be stagnant (values changed) → continue

This means a loop that is actively regressing (more errors each iteration) will run until `maxIter`, then abort. There is no early termination for monotonically worsening state. The skill's loop.md Step 4 mentions re-routing to coverage workflow if coverage is the only gap, but does not address active regression.

##### Storage (`internal/loop/storage.go:14-117`)

Atomic writes via temp-file + rename pattern. JSON validation before unmarshal. Field validation (SpecID non-empty, phase valid). All correct.

**Idempotency**: `DeleteState` is idempotent (returns nil on `os.ErrNotExist`). `SaveState` overwrites. Loading non-existent state returns error — callers must handle.

#### Skill-to-Go-Engine Mapping

The `loop.md` workflow skill operates at the LLM/Claude Code layer, not directly calling the `internal/loop` Go package. The Go engine is invoked via the `stop__loop_controller` hook (referenced in `moai-workflow-loop/SKILL.md`). This creates a two-layer architecture:

- **Claude Code layer**: `loop.md` skill drives Claude's behavior (agent delegation, task tracking, MX tags, snapshots in `.moai/cache/loop-snapshots/`)
- **Hook layer**: `stop__loop_controller` hook reads `.moai/cache/.moai_loop_state.json` and signals loop continuation (exit code 1) or completion (exit code 0)
- **Go layer**: `internal/loop` + `internal/ralph` implement the decision engine and state persistence

The Claude Code layer uses its own snapshot format (`loop-snapshots/iteration-N.json`) while the Go layer uses `{specID}.json` in the loop state directory. These are separate artifacts.

#### D5 — Performance

- Parallel diagnostics: 4 tools via background Bash.
- Memory pressure detection: session duration > 25 minutes or iteration time doubling triggers checkpoint.
- Physical memory check via `free -m` / `vm_stat` (optional, requires `memory_guard.enabled` in quality.yaml).
- Max iterations default: 10 (ralph.yaml), skill states 100 as default — **inconsistency between config and skill documentation**.

```yaml
# ralph.yaml:26
max_iterations: 10

# loop.md:37
--max N (alias --max-iterations): Maximum iteration count (default 100)
```

#### D6 — Safety + Reversibility

- Snapshot-based resume at `$CLAUDE_PROJECT_DIR/.moai/cache/loop-snapshots/`.
- `--resume` flag to restore from latest or named snapshot.
- Pre-exit clean sweep runs `clean` workflow on modified files.
- No git backup taken before destructive edits.

#### D7 — Observability

- Per-iteration snapshots preserve error count, coverage, and phase.
- MX_TAG_REPORT section generated after each iteration.
- `slog.Default().Warn()` used in controller for save failures.
- No structured log artifact written to `.moai/logs/`.
- Progress tracking via TaskCreate/TaskUpdate.

#### D8 — Idempotency

- If zero issues found: completion conditions met at Step 4, loop exits cleanly with success.
- File checksums not used; idempotency is behavioral (re-running diagnostics).
- Re-running `/moai loop` on a clean project should exit immediately at Step 4 (completion check passes).

#### D9 — Configuration

`ralph.yaml` provides:
- `loop.max_iterations: 10`
- `loop.completion.coverage_threshold: 85`
- `loop.completion.zero_errors: true`
- `loop.completion.tests_pass: true`
- `loop.auto_fix: false`
- `loop.require_confirmation: true`

**Gaps**:
- `loop.completion.coverage_threshold` in ralph.yaml not wired into `MeetsQualityGate()` in Go code.
- No per-language iteration tolerance.
- `loop.md` states `--max N` default is 100, ralph.yaml says 10.

**Loop audit verdict: 8/10** — Go engine is architecturally sound with proper state machine, atomic persistence, and stagnation detection. Key gaps: `IsStagnant` vs `IsStagnantWithDiagnostics` mismatch, coverage threshold config drift, max_iterations documentation inconsistency, no regression worsening detection.

---

## /moai codemaps Audit

### Thin Command Wrapper

**File**: `.claude/commands/moai/codemaps.md` (7 lines)

Fully compliant thin wrapper. Routes to `Skill("moai")` with `codemaps $ARGUMENTS`.

### Workflow Skill: codemaps.md v1.0.0

**File**: `.claude/skills/moai/workflows/codemaps.md` (150 lines, version 1.0.0)

**Note**: codemaps.md is at version 1.0.0, while fix.md and loop.md are at v2.5.0. This version gap suggests codemaps received significantly less iteration and is the youngest of the three workflows.

#### D1 — Agentless Pipeline Conformance

This is the most significant conformance failure in this audit.

**Localize (Phase 1)**: Explicitly delegates to `Explore` subagent:
```
[HARD] Delegate codebase exploration to the Explore subagent.
```
This is LLM-driven control flow — the Explore agent makes decisions about what constitutes module boundaries, entry points, and architecture patterns.

**Repair (Phase 3)**: Explicitly delegates to `manager-docs` subagent for map generation:
```
[HARD] Delegate map generation to the manager-docs subagent.
```
Again, LLM-driven content creation, not deterministic tool invocation.

**Validate (Phase 4)**: Performed by "MoAI orchestrator (verification checks)" — this remains in the workflow but is the only phase that could be considered deterministic.

**Verdict**: All three phases use LLM-driven agent delegation for control flow. This directly violates REQ-WF004-004: "Agentless subcommand execution shall not invoke Agent() for control flow decisions; only deterministic tool invocations are permitted." The SPEC's own mapping for codemaps is:
```
codemaps: localize(stale map) → repair(regenerate) → validate(diff)
```
The actual implementation inverts this: it discovers stale maps via an agent (not a diff of existing vs. codebase), regenerates via an agent, and validates via the orchestrator. **Full non-conformance.**

#### D4 — Codemap Generation Details

**Scan strategy**: Delegates to the `Explore` subagent which uses AST/structural analysis (the Explore subagent has access to Read, Grep, Glob). No direct Go AST or Tree-sitter integration. The scan strategy is determined by the Explore agent's judgment, not a fixed algorithm.

**16-language support**: Indirectly, via the Explore subagent which is language-agnostic. No explicit language-specific scan logic.

**Output format** (`codemaps.md:88-99`): Five fixed output files:
- `overview.md` — High-level architecture summary
- `modules.md` — Module catalog
- `dependencies.md` — Dependency graph (text + optional mermaid)
- `entry-points.md` — Entry point catalog
- `data-flow.md` — Data flow paths

**Format options**: `--format markdown` (default), `--format mermaid`, `--format json`. The mermaid option adds diagrams to markdown; the json option generates machine-readable alongside markdown. Flags exist but implementation depends entirely on manager-docs generating appropriate content.

**Incremental updates**: If `--force` not set, existing `.moai/project/codemaps/` content is passed to manager-docs for incremental updates. This is the intended mechanism but relies on the agent correctly computing a diff.

**Actual output files** (observed on disk as of Mar 31 2022):
```
.moai/project/codemaps/
  data-flow.md    (11.6KB)
  dependencies.md  (8.3KB)
  entry-points.md  (9.6KB)
  modules.md      (16.8KB)
  overview.md      (6.4KB)
```
The output format matches the spec. Content quality is not assessed in this audit.

**Integration with project docs**: codemaps output lives in `.moai/project/codemaps/` while project docs (product.md, structure.md, tech.md) live in `.moai/project/`. No explicit wiring — `/moai project` and `/moai sync` reference these directories independently.

#### D5 — Performance

No performance targets documented. Scan time on large repos depends entirely on the Explore subagent's thoroughness and the manager-docs generation time. No timeout configured.

**`--depth N` flag** (default 4): Limits directory exploration depth, providing a performance knob. This is the only performance control.

#### D6 — Safety + Reversibility

- `--force` flag regenerates all maps, overwriting existing content.
- No backup of existing `.moai/project/codemaps/` before regeneration.
- If `--force` is used, all previous maps are overwritten without recovery option.
- No dry-run mode.

#### D7 — Observability

- TaskCreate/TaskUpdate tracking for each map file.
- Phase 5 completion report with generated file list.
- No structured log to `.moai/logs/`.
- No iteration history (single-pass generation).

#### D8 — Idempotency

- Without `--force`: If maps already exist and codebase hasn't changed, the Explore agent should produce identical output and manager-docs should detect no change. But this relies on agent judgment, not a deterministic hash comparison.
- With `--force`: Always regenerates, not idempotent.
- **No explicit no-op exit**: REQ-WF004-007 requires that if the localize phase finds no stale maps (codemaps up-to-date), the pipeline exits with "no-op." The current implementation has no such check — it always runs the full pipeline.

#### D9 — Configuration Gaps

No dedicated config section for codemaps in any YAML file. All configuration is via flags:
- `--force`: Full regeneration
- `--area AREA`: Focus area
- `--format FORMAT`: Output format
- `--depth N`: Exploration depth

No project-level defaults for format or depth. No per-project codemaps configuration possible without modifying the skill invocation.

**Codemaps audit verdict: 5/10** — The output format and file structure are well-designed. The existing generated codemaps demonstrate the feature works. However, the workflow fundamentally violates SPEC-V3R2-WF-004's Agentless contract by using LLM-driven agents for all three phases. Missing no-op exit, no backup, no dry-run, and no config schema.

---

## Agentless Pipeline Conformance Matrix

| Subcommand | Localize Phase | Repair Phase | Validate Phase | LLM Control Flow? | Overall Conformance |
|------------|---------------|--------------|---------------|-------------------|---------------------|
| `/moai fix` | Deterministic (3 parallel scanners) | Agent delegation (expert-*) | Deterministic (re-run diagnostics) | Phase ordering fixed, repair uses agents | **Partial** |
| `/moai loop` | Deterministic (4 parallel diagnostics) | Agent delegation (expert-*) | Implicit (next iteration's Step 3) | Phase ordering fixed, repair uses agents | **Partial** |
| `/moai codemaps` | LLM agent (Explore subagent) | LLM agent (manager-docs) | Orchestrator verification | All phases LLM-controlled | **No** |

### Conformance Notes

The Agentless paper (Xia et al. 2024) cited in SPEC-V3R2-WF-004 §1.1 defines three-phase non-agentic pipelines as having no LLM-driven control flow. The SPEC itself acknowledges (§7 Constraints): "Agentless pipeline is LLM-enabled but control flow is fixed." This creates ambiguity: does using a specialized subagent for repair violate "control flow"?

Interpretation A (strict): Any `Agent()` call in the pipeline violates Agentless. → fix and loop are non-conformant.

Interpretation B (lenient): Control flow means the agent cannot decide what phase to run next. Using agents for execution within a fixed phase order is acceptable. → fix and loop are conformant; codemaps is not (agents decide exploration strategy).

This audit adopts **Interpretation B** as the practical reading, consistent with the SPEC's own observation that the utility subcommands are "already quasi-Agentless" (§4). Under this interpretation, fix and loop are Partial (they use agents for repair, but the pipeline itself is fixed), and codemaps is No (agents determine exploration strategy and content, not just execution).

---

## Cross-Cutting Findings

### Finding 1: Workflow-to-Go-Engine Loose Coupling

The Claude Code workflow skills (fix.md, loop.md) and the Go engine (`internal/loop/`, `internal/ralph/`) are loosely coupled through the hook system. The skills describe a 9-step per-iteration cycle; the Go engine implements a 4-phase state machine (Analyze→Implement→Test→Review). These are different abstractions of the same process, which works but creates a semantic gap:

- The skill's "Step 3 Parallel Diagnostics" maps to Go's Test phase feedback collection.
- The skill's "Step 6 Fix Execution" maps to no Go phase — the Go engine only evaluates, it does not execute fixes.

The Go engine is a decision oracle (continue/converge/abort), while the actual repair is Claude-driven through the skill. This separation is architecturally clean but should be documented.

### Finding 2: Configuration Drift (ralph.yaml vs skill defaults)

| Parameter | ralph.yaml | Skill Documentation | Go Code Default |
|-----------|-----------|---------------------|----------------|
| `max_iterations` | 10 | 100 (`loop.md:37`) | 5 (`controller.go:31`) |
| `coverage_threshold` | 85 | 85 (`loop.md:221`) | `config.DefaultTestCoverageTarget` (85) |
| `auto_fix` | false | Described as Level 1 only | N/A (skill layer) |

The three-way mismatch on `max_iterations` is the most critical. A user reading `loop.md` expects 100 iterations; the Go engine defaults to 5 (when constructed without config); ralph.yaml sets 10. The hook system determines which value wins.

### Finding 3: Stagnation Convergence Uses Incomplete Predicate

`RalphEngine.Decide()` at `engine.go:64` calls `loop.IsStagnant(prev, feedback)` which checks only three integer metrics (TestsFailed, LintErrors, Coverage). The function `IsStagnantWithDiagnostics()` at `feedback.go:13` additionally considers LSP diagnostic count trends. The engine does not use the richer predicate. Given that REQ-LL-010 was implemented to add diagnostic-trend stagnation detection, not wiring it into the engine is an oversight.

### Finding 4: No Regression Worsening Gate

None of the three workflows (fix, loop, codemaps) has an explicit gate that stops the pipeline when an iteration makes things measurably worse. The engine's continue action fires whenever quality gate is not met and stagnation is not detected. A scenario where each iteration increases `TestsFailed` will run to max_iterations silently.

### Finding 5: Missing `MODE_FLAG_IGNORED_FOR_UTILITY` Logging

REQ-WF004-011 requires that when a `--mode` flag is passed to a utility subcommand, it must be ignored and a `MODE_FLAG_IGNORED_FOR_UTILITY` info log emitted. None of the three workflow skills implement this log message. The fix and loop skills do not mention `--mode` at all. This is expected for now (SPEC is draft), but noted as a gap for implementation.

### Finding 6: Codemaps Version Lag

`codemaps.md` is at v1.0.0; fix.md and loop.md are at v2.5.0. The 1.5 version gap indicates codemaps has not received the MX tag integration, 16-language support documentation, or structured error output normalization that fix and loop received in v2.2.0+. Codemaps should be brought up to version parity.

---

## Integration with v3R2 SPECs

### SPEC-V3R2-WF-004 (Agentless Fixed-Pipeline) — This Audit's Primary SPEC

**Status**: DRAFT (0.1.0). Not yet implemented in any workflow skill.

**Required changes to implement WF-004**:

1. Add 3-phase pipeline header to `fix.md`, `loop.md`, `codemaps.md` per §10 implementation path.
2. Update `workflow-modes.md` with pipeline classification matrix (AC-WF004-06).
3. Implement `MODE_FLAG_IGNORED_FOR_UTILITY` logging for `--mode` on utility subcommands (REQ-WF004-011).
4. Add no-op exit path for zero-target localize phase (REQ-WF004-007).
5. Refactor `codemaps.md` to use deterministic tool calls for localize phase, removing Explore subagent.

**Codemaps non-conformance**: The SPEC's own §4 assumption states "3-phase contract maps naturally to each subcommand." For codemaps, the assumption `localize(stale map) → repair(regenerate) → validate(diff)` is reasonable but the current implementation inverts localize and repair to be agent-driven. This requires architectural change, not just documentation.

### SPEC-V3R2-WF-003 (Multi-Mode Router)

**Impact on `/moai loop`**: Per REQ-WF003-004, `/moai loop` shall become an alias for `/moai run --mode loop`. The current thin wrapper (`loop.md` command) delegates to `Skill("moai")` with `loop $ARGUMENTS`, not to the `run` skill with `--mode loop`. This alias relationship needs implementing when WF-003 is implemented.

**Impact on `/moai fix` and `/moai codemaps`**: REQ-WF003-006 states utility subcommands operate in pipeline mode per WF-004. REQ-WF003-011 covers `--mode` rejection for utility subcommands.

### Workflow-modes.md Pipeline Classification Table

The `workflow-modes.md` rule file does not currently contain the subcommand × mode matrix required by REQ-WF004-005 and AC-WF004-06. This table needs to be added, listing fix/coverage/mx/codemaps/clean as "pipeline (agentless)" and plan/run/sync/design as "multi-agent."

---

## Recommendations

### Priority High

**P1-1: Wire `IsStagnantWithDiagnostics` into RalphEngine**
- File: `internal/ralph/engine.go:64`
- Change: `loop.IsStagnant(prev, feedback)` → `loop.IsStagnantWithDiagnostics(prev, feedback)`
- Impact: Convergence uses richer diagnostic data; prevents false stagnation-converge when LSP errors increase while integer metrics hold flat.
- Effort: 2-line code change + test verification.

**P1-2: Implement no-op exit in all three workflows**
- Files: `fix.md`, `loop.md`, `codemaps.md`
- Change: After localize phase, if zero targets found → emit `status: no-op, exit code 0`, skip repair and validate.
- Required by: REQ-WF004-007, AC-WF004-07.

**P1-3: Wire ralph.yaml coverage threshold into `MeetsQualityGate()`**
- File: `internal/loop/feedback.go:66`
- Change: Accept a threshold parameter or read from config; replace hardcoded `config.DefaultTestCoverageTarget`.
- Impact: Makes quality gate responsive to project-level configuration.

**P1-4: Refactor `codemaps.md` localize phase to deterministic tools**
- File: `.claude/skills/moai/workflows/codemaps.md`
- Change: Replace Explore subagent with deterministic Glob/Grep/Read to discover existing codemaps and diff against current file structure. Keep manager-docs for repair (content generation), but make localize phase tool-only.
- Impact: Partial WF-004 conformance for codemaps (localize becomes deterministic).

### Priority Medium

**P2-1: Resolve max_iterations documentation inconsistency**
- Files: `loop.md:37` (states 100), `ralph.yaml:26` (states 10), `controller.go:31` (states 5)
- Change: Align all three to the same default; document the precedence chain.

**P2-2: Add regression worsening gate to loop engine**
- File: `internal/ralph/engine.go`, `internal/loop/feedback.go`
- Change: Add `IsWorsening(prev, curr *Feedback) bool` predicate; add `ActionEscalate` for monotonic regression. After 3 iterations of worsening, escalate to human review instead of continuing.

**P2-3: Add per-project config schemas for fix and codemaps**
- New sections in `ralph.yaml`: `fix.max_level`, `fix.security_scan_default`, `codemaps.default_format`, `codemaps.default_depth`.

**P2-4: Implement `MODE_FLAG_IGNORED_FOR_UTILITY` logging**
- Files: `fix.md`, `codemaps.md` (loop becomes alias per WF-003)
- Change: Add flag parsing section; log MODE_FLAG_IGNORED_FOR_UTILITY if `--mode` is detected.

### Priority Low

**P3-1: Version-align codemaps.md to v2.x**
- File: `codemaps.md` (currently v1.0.0)
- Change: Port MX tag integration, structured error output normalization, and 16-language documentation from fix.md/loop.md v2.2.0+ changelog.

**P3-2: Add `.moai/logs/` artifact for each workflow**
- Append machine-readable JSONL entry to `.moai/logs/fix.log`, `.moai/logs/loop.log`, `.moai/logs/codemaps.log` on completion.
- Enables external tooling to analyze workflow history without parsing snapshots.

**P3-3: Implement `/moai loop` → `/moai run --mode loop` alias**
- Per SPEC-V3R2-WF-003 REQ-WF003-004.
- Currently loop.md is a parallel implementation, not an alias.

**P3-4: Add explicit validate phase to loop (per-fix, not per-iteration)**
- Currently validate is implicit: the next iteration's Step 3 re-runs diagnostics globally.
- A per-fix validate would run targeted diagnostics on the modified file immediately after repair, before moving to the next issue.
- Provides faster feedback on regressions introduced by a specific fix.

---

**Report version**: 1.0.0
**Generated**: 2026-04-24
**Scope coverage**: D1-D9 per audit specification
**Files cited**:
- `.claude/skills/moai/workflows/fix.md` (fix.md:37, 50-71, 127-132, 157-163)
- `.claude/skills/moai/workflows/loop.md` (loop.md:37, 49-140, 196-225)
- `.claude/skills/moai/workflows/codemaps.md` (codemaps.md:44, 68, 86)
- `.claude/commands/moai/fix.md`, `loop.md`, `codemaps.md`
- `.claude/skills/moai-workflow-loop/SKILL.md`
- `internal/ralph/engine.go` (engine.go:29-89, 64, 96-121, 337 LOC)
- `internal/loop/controller.go` (controller.go:29-394)
- `internal/loop/state.go` (state.go:27-47)
- `internal/loop/storage.go` (storage.go:14-117)
- `internal/loop/feedback.go` (feedback.go:13-74)
- `internal/loop/go_feedback.go` (go_feedback.go:37-64, 260 LOC)
- `internal/loop/feedback_channel.go` (feedback_channel.go:14-81)
- `.moai/config/sections/ralph.yaml`
- `.moai/specs/SPEC-V3R2-WF-004/spec.md`
- `.moai/specs/SPEC-V3R2-WF-003/spec.md`
