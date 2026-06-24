---
id: SPEC-V3R6-STATUSLINE-MILLION-BUDGET-001
title: "Statusline memory_test AutoCompactScaling model-env isolation"
version: "0.2.0"
status: completed
created: 2026-06-18
updated: 2026-06-18
author: manager-spec
priority: P2
phase: "v3.0.0"
module: "internal/statusline"
lifecycle: spec-anchored
tags: "statusline, test-fixture, model-env-isolation, debt-cleanup"
---

# Plan — SPEC-V3R6-STATUSLINE-MILLION-BUDGET-001

## §A Context

**One-paragraph summary**: The `TestCollectMemory_AutoCompactScaling` test in `internal/statusline/memory_test.go` fails on GLM-ambient developer machines because `resolveContextWindowOverride()` (`memory.go:139`) reads the ambient `ANTHROPIC_DEFAULT_*_MODEL` env vars and, on a `glm-5.2` match, returns `1_000_000` — overriding the test's `ContextWindowSize: 200000` input and producing `TokensUsed = 850000, want 170000` failures. The fix is **test-only env isolation** (Direction B): neutralize the three model env vars via `t.Setenv(..., "")` inside the test so the production resolver returns 0 on priority-3 lookup and the existing 200K-derived expectations hold. Production `memory.go` is untouched.

**Why Direction B** (not Direction A "update expectations to 1M"): Direction A hardcodes the 1M figure and re-stales on the next GLM model rename or context-table revision. Direction B is durable — it keys the isolation on "any model-env value present" rather than a specific identifier, and it honors the test author's original intent (verify threshold-scaling math at a fixed 200K baseline). Full rationale: spec.md §B.

**Debt lineage**: AC-HNS-011 (PASS-WITH-DEBT) of closed SPEC-V3R6-HARNESS-NAMESPACE-V2-001. This SPEC retires that debt.

**Tier classification**: **Tier S LEAN** — single test file, single concern, no production change, no architectural decision. ACs inline in spec.md §D (no separate acceptance.md).

## §B Pre-flight (verified this session)

- [x] **Race check**: `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD` → `0 0` (clean sync, no concurrent-session divergence).
- [x] **Pre-existing scope confirmed**: failure reproduces on commit `bef24877d` (pre-NAMESPACE-V2). NOT a regression introduced by SPEC-V3R6-HARNESS-NAMESPACE-V2-001 or SPEC-V3R6-HARNESS-MOAI-NAMESPACE-001.
- [x] **Production code references verified**:
  - `memory.go:28` — `"glm-5.2": 1_000_000` in `glmContextWindows` table.
  - `memory.go:139` — `resolveContextWindowOverride()` function entry.
  - `memory.go:146-150` — env slot list (`EnvAnthropicDefaultOpusModel`, `Sonnet`, `Haiku`).
  - `memory.go:199-207` — `contextSize` resolution: input → default 200000 → override application.
- [x] **Test code references verified**:
  - `memory_test.go:110` — `TestCollectMemory_AutoCompactScaling` function entry.
  - `memory_test.go:119-183` — 5 table entries, all with `ContextWindowSize: 200000`.
  - `memory_test.go:186-204` — `t.Run` runner with `t.Setenv("CLAUDE_AUTOCOMPACT_PCT_OVERRIDE", ...)` at `:188`.
  - `memory_test.go:193` / `:196` — the two failing assertions.
- [x] **SPEC ID uniqueness**: no existing SPEC matches `STATUSLINE-MILLION-BUDGET` (closest siblings: `SPEC-V3R6-STATUSLINE-PRESET-RETIRE-001`, `SPEC-V3R5-STATUSLINE-*-001` — distinct scope).
- [x] **SPEC ID regex self-check**: `decomposition: SPEC ✓ | V3R6 ✓ | STATUSLINE ✓ | MILLION ✓ | BUDGET ✓ | 001 ✓ → PASS` (canonical regex `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$`).
- [x] **Frontmatter schema**: 12 canonical fields present on this plan.md and the sibling spec.md; `created`/`updated` (NOT `created_at`/`updated_at`); `tags` comma-separated string (NOT `labels` array); `version` quoted string.

## §C Constraints (carried from spec.md §E)

- Test-only change: `internal/statusline/memory.go` byte-identical pre/post fix (AC-SMB-003).
- Use `t.Setenv(..., "")` for env neutralization (NOT `os.Unsetenv` + deferred restore) — guarantees test-lifetime scoping and parallel safety.
- No new third-party imports; stdlib `testing` + `os` only.
- No `t.Parallel()` introduction; no refactor of the table-driven structure into separate functions.
- Do NOT touch the sibling override-priority tests (`memory_test.go:276+` / `:350+` / `:415+` / `:450+`) — they intentionally populate model env.

## §D Milestone

This Tier S SPEC executes as a **single milestone** (M1). There is no M2/M3 decomposition — the change is one cohesive test-file edit.

### M1 — Apply model-env isolation to TestCollectMemory_AutoCompactScaling

**Scope**: edit `internal/statusline/memory_test.go` only.

**Implementation approach** (manager-develop will execute):

1. **Choose isolation shape** (manager-develop picks one, both satisfy REQ-SMB-001):
   - **Option M1-a (inline, 3 lines per subtest)**: add three `t.Setenv` calls at the top of each `t.Run` body, immediately after the existing `t.Setenv("CLAUDE_AUTOCOMPACT_PCT_OVERRIDE", tt.threshold)` at line 188:
     ```go
     t.Run(tt.name, func(t *testing.T) {
         t.Setenv("CLAUDE_AUTOCOMPACT_PCT_OVERRIDE", tt.threshold)
         t.Setenv(config.EnvAnthropicDefaultOpusModel, "")
         t.Setenv(config.EnvAnthropicDefaultSonnetModel, "")
         t.Setenv(config.EnvAnthropicDefaultHaikuModel, "")
         // ... rest of subtest
     })
     ```
     Pro: maximally local, no abstraction. Con: 15 lines of repetition across 5 subtests; if a 6th subtest is added without the pattern, isolation is lost (REQ-SMB-003 risk).
   - **Option M1-b (shared helper)**: extract a helper invoked once per subtest:
     ```go
     // isolateModelEnv neutralizes ambient ANTHROPIC_DEFAULT_*_MODEL env vars
     // so resolveContextWindowOverride() returns 0 and the subtest's
     // StdinData.ContextWindowSize is the sole context-size source.
     // Required because GLM-ambient dev shells populate these vars and would
     // otherwise override the 200K test input with a 1M GLM context window.
     // See SPEC-V3R6-STATUSLINE-MILLION-BUDGET-001.
     func isolateModelEnv(t *testing.T) {
         t.Helper()
         t.Setenv(config.EnvAnthropicDefaultOpusModel, "")
         t.Setenv(config.EnvAnthropicDefaultSonnetModel, "")
         t.Setenv(config.EnvAnthropicDefaultHaikuModel, "")
     }
     ```
     Then call `isolateModelEnv(t)` as the first line of each `t.Run` body. Pro: DRY, makes the intent explicit, a future contributor adding a 6th subtest sees the helper and is nudged toward reuse (mitigates REQ-SMB-003 risk). Con: adds one function to the test file.

   **Recommendation**: **Option M1-b (shared helper)**. The helper's doc-comment also serves as the in-code anchor for the SPEC reference, making the debt-retirement provenance discoverable by future grep (`grep -rn "STATUSLINE-MILLION-BUDGET" internal/`).

2. **Import check**: `internal/statusline/memory_test.go` must import the `config` package (for `config.EnvAnthropicDefaultOpusModel` etc.) OR use the raw string literals `"ANTHROPIC_DEFAULT_OPUS_MODEL"` etc. directly. Prefer the `config` package constants for single-source-of-truth alignment with `memory.go:147-149`. Verify the import path — likely `github.com/modu-ai/moai-adk-go/internal/config`. If the test file does not already import `config`, add the import.

3. **Verify no expectation changes**: after the edit, `git diff internal/statusline/memory_test.go` MUST show only the `t.Setenv`/helper additions — the `wantUsed` / `wantBudget` / `wantPctApprox` table values MUST be byte-identical (AC-SMB-002). If the diff touches any table value, STOP — that is Direction A and is rejected per spec.md §B.

4. **Run the target test under GLM-ambient conditions** (the proof):
   ```bash
   # Simulate the GLM-ambient shell
   ANTHROPIC_DEFAULT_OPUS_MODEL=glm-5.2 \
   ANTHROPIC_DEFAULT_SONNET_MODEL=glm-5.2 \
   ANTHROPIC_DEFAULT_HAIKU_MODEL=glm-5.2 \
   go test ./internal/statusline/ -run TestCollectMemory_AutoCompactScaling -v -count=1
   ```
   Expected: all 5 subtests PASS. Pre-fix this command fails with `TokensUsed = 850000, want 170000`.

5. **Run the full statusline test suite** (regression check):
   ```bash
   go test ./internal/statusline/... -v -count=1
   ```
   Expected: all tests PASS, including the sibling override-priority tests (AC-SMB-004). The `t.Setenv` test-lifetime scoping guarantees the isolation does NOT leak into sibling tests.

6. **AC-SMB-005 durability sanity check** (then revert):
   - Temporarily edit `memory.go:28` to add a hypothetical entry like `"glm-99.9": 2_000_000`.
   - `ANTHROPIC_DEFAULT_OPUS_MODEL=glm-99.9 go test ./internal/statusline/ -run TestCollectMemory_AutoCompactScaling -v -count=1`
   - Expected: all 5 subtests still PASS (isolation is keyed on "env var non-empty", not on a specific model name).
   - **REVERT** the `memory.go:28` edit before commit. `git diff internal/statusline/memory.go` must be empty at commit time (AC-SMB-003).

7. **Commit** (orchestrator commits; manager-develop stages):
   - Subject: `test(statusline): isolate AutoCompactScaling from ambient GLM model env (SPEC-V3R6-STATUSLINE-MILLION-BUDGET-001)`
   - Body: names the 5 subtests, references AC-HNS-011 debt retirement, states Direction B rationale in one sentence.
   - `Authored-By-Agent: manager-develop` trailer.

## §E Self-Verification (manager-develop §E reporting, plan-phase skeleton only)

This section is the placeholder for manager-develop's run-phase §E self-verification matrix. manager-spec populates only the plan-phase expectations; manager-develop fills the evidence at run-phase.

### §E.1 Plan-phase Audit-Ready Signal

- SPEC artifacts: spec.md + plan.md (this file) authored. Tier S LEAN → no separate acceptance.md (ACs inline in spec.md §D).
- Frontmatter: 12 canonical fields on both artifacts; `created`/`updated` (NOT snake_case); `tags` CSV string; `version` quoted.
- SPEC ID regex: `SPEC-V3R6-STATUSLINE-MILLION-BUDGET-001` → PASS (decomposition trace in §B above).
- GEARS compliance: REQ-SMB-001 (Ubiquitous), REQ-SMB-002 (State-driven `While`), REQ-SMB-003 (Capability gate `Where`), REQ-SMB-004 (event-detected `When` for unwanted behavior), REQ-SMB-005 (Ubiquitous) — five distinct patterns, no residual `IF/THEN` legacy modality.
- Exclusions: spec.md §F lists 6 explicit NOT-in-scope items (production code, Direction A, sibling tests, t.Parallel, glmContextWindows table, llm.yaml).
- Direction B rationale: spec.md §B + this file §A document the rejection of Direction A with 3 durability/intent/minimal-blast-radius reasons.

### §E.2 Run-phase Evidence

_<pending run-phase — manager-develop populates with verbatim test command outputs>_

### §E.3 Run-phase Audit-Ready Signal

_<pending run-phase>_

### §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs populates>_

### §E.5 Mx-phase Audit-Ready Signal

_<pending Mx-phase — manager-docs populates>_

## §F Anti-Patterns to Avoid (run-phase)

- **AP-SMB-001 — Direction A creep**: manager-develop "helpfully" updates `wantBudget: 170000` → `850000` because the test was failing and that looks like the obvious fix. **Block**: Direction A is explicitly rejected per spec.md §B. The fix MUST be env isolation only; expectation values are immutable.
- **AP-SMB-002 — `os.Unsetenv` instead of `t.Setenv`**: using `os.Unsetenv` with a deferred restore loses the test-runtime's automatic cleanup guarantee and breaks under parallel-test execution. **Block**: use `t.Setenv(..., "")` exclusively.
- **AP-SMB-003 — isolating the sibling override-priority tests**: a broad "neutralize all model env in the whole test file" sweep would defeat the intentional env-population in `TestCollectMemory_*` at `:276+` / `:350+` / `:415+` / `:450+`. **Block**: isolation is scoped to `TestCollectMemory_AutoCompactScaling` ONLY.
- **AP-SMB-004 — leaving the AC-SMB-005 sanity-check edit in `memory.go`**: forgetting to revert the temporary `"glm-99.9": 2_000_000` table entry. **Block**: `git diff internal/statusline/memory.go` MUST be empty at commit time; AC-SMB-003 enforces this.
- **AP-SMB-005 — claiming PASS without the GLM-ambient simulated run**: manager-develop runs `go test ./internal/statusline/` in a Claude-ambient (non-GLM) shell, sees green, and reports PASS. **Block**: the proof command (§D step 4) MUST be run with the `glm-5.2` env vars explicitly exported, and the verbatim output MUST appear in §E.2. This is the verification-claim-integrity invariant — unobserved PASS is not PASS.

## §G Cross-References

- **spec.md** (sibling): `internal/statusline` SPEC body with GEARS REQs + inline ACs.
- **Debt origin**: AC-HNS-011 of `.moai/specs/SPEC-V3R6-HARNESS-NAMESPACE-V2-001/` (closed, PASS-WITH-DEBT).
- **Verification doctrine**: `.claude/rules/moai/core/verification-claim-integrity.md` — "no unobserved-claim" invariant; §E.2 evidence must be verbatim command output, not summary.
- **Frontmatter SSOT**: `.claude/rules/moai/development/spec-frontmatter-schema.md` — 12 canonical fields, status enum, snake_case rejection.
- **GEARS authoring**: `.claude/skills/moai-workflow-spec/SKILL.md` § GEARS Format — five patterns, compound clause, generalized `<subject>`.
- **Tier S LEAN policy**: `.claude/rules/moai/workflow/spec-workflow.md` § SPEC Complexity Tier — Tier S permits inline ACs (no separate acceptance.md).

## HISTORY

- **2026-06-18** (v0.1.0, manager-spec): plan-phase draft authored. Single-milestone (M1) Tier S LEAN plan. Direction B adopted. Two implementation options (M1-a inline / M1-b helper) surfaced for manager-develop; M1-b recommended. Run-phase + plan-auditor + Implementation Kickoff deferred to follow-up session per orchestrator directive.
