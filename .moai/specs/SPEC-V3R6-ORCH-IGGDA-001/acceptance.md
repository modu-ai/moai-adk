# Acceptance Criteria — SPEC-V3R6-ORCH-IGGDA-001

> **Testable ACs**. The safe-condition predicate (REQ-IGGDA-004/005) is tested EXHAUSTIVELY on BOTH branches (auto-proceed when safe-condition holds; explicit-gate when it fails). Per the FROZEN-invariant amendment risk, the predicate's AC coverage is the single highest-risk acceptance surface.

---

## §D — AC Matrix

### §D.1 — D2 Path B safe-condition predicate (the FROZEN-invariant amendment) — HIGHEST RISK

#### AC-IGGDA-001 — Implementation Kickoff Approval gate still issued (NOT removed)

**Given** the orchestrator has reached the plan→run boundary (Phase 0.5 plan-auditor verdict obtained)
**When** the orchestrator evaluates the Implementation Kickoff Approval gate
**Then** the orchestrator issues an `AskUserQuestion` round (the gate is NOT silently bypassed, regardless of safe-condition predicate outcome)
**And** the `AskUserQuestion` round's first option is marked `(Recommended)` per askuser-protocol.md
**Evidence**: `grep -n "AskUserQuestion" internal/template/templates/.claude/skills/moai/workflows/run.md` returns the Implementation Kickoff Approval section; the gate is present.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-001.

---

#### AC-IGGDA-002 — Preferences drained at the gate (Phase 0 concentration)

**Given** the Implementation Kickoff Approval `AskUserQuestion` round is presented
**When** the user responds with approval
**Then** ALL user preferences (Tier, mode preference, PR strategy, domain-specific preferences) have been collected at this gate (NOT deferred to mid-run)
**And** the autonomous Phase 1–3 agents (Workflow, `/goal`-turn, recursive-self-diagnosis) can execute without prompting the user
**Evidence**: the `AskUserQuestion` round's options enumerate the preference set; no mid-run `AskUserQuestion` call site exists in the autonomous path.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-002, REQ-IGGDA-024 (inherited C2).

---

#### AC-IGGDA-003 — Non-approval user response blocks regardless of predicate

**Given** the safe-condition predicate holds (all 4 conditions met)
**When** the user responds to the Implementation Kickoff Approval `AskUserQuestion` round with "further review" or "abort" (NOT approval)
**Then** the orchestrator does NOT advance to Phase 1, regardless of the predicate outcome
**And** user intent overrides the safe-condition auto-proceed
**Evidence**: the orchestrator's decision branch on the `AskUserQuestion` response is upstream of the predicate auto-proceed.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-003.

---

#### AC-IGGDA-004 — Safe-condition predicate AUTO-PROCEED branch (all 4 conditions hold)

**Given** (a) Socratic intent clarity is 100% (interview complete per CLAUDE.md §7 Rule 5)
**And** (b) plan-auditor verdict is PASS
**And** (c) SPEC is Tier S or Tier M (NOT Tier L)
**And** (d) the SPEC contains NO security/payment/critical-domain keywords AND no `--pr`-forcing destructive scope
**When** the orchestrator evaluates the safe-condition predicate
**Then** the Implementation Kickoff Approval gate is reduced to a lightweight confirmation that auto-proceeds to Phase 1 after a bounded timeout (the `(Recommended)` option auto-selected)
**And** the `AskUserQuestion` round is STILL ISSUED (the user retains veto authority — the gate is reduced, not removed)
**And** the predicate decision is logged to `progress.md §E IGGDA Kickoff Predicate` with all 4 conditions + the auto-proceed verdict
**Evidence**: a regression test (Go or shell) constructs a Tier M SPEC with no dangerous keywords + 100% intent clarity + plan-auditor PASS; the orchestrator's predicate evaluation returns auto-proceed; `progress.md` logs the 4 conditions.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-004, REQ-IGGDA-006.

---

#### AC-IGGDA-005a — Safe-condition predicate EXPLICIT-GATE branch (Tier L)

**Given** the SPEC is Tier L (regardless of other conditions)
**When** the orchestrator evaluates the safe-condition predicate
**Then** condition (c) FAILS
**And** the Implementation Kickoff Approval gate REMAINS a mandatory explicit `AskUserQuestion` that blocks until the user responds (no auto-proceed, no timeout)
**And** the predicate decision is logged to `progress.md` with condition (c) = FAIL + explicit-gate verdict
**Evidence**: a regression test constructs a Tier L SPEC; the predicate returns explicit-gate; the `AskUserQuestion` blocks.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-005.

---

#### AC-IGGDA-005b — Safe-condition predicate EXPLICIT-GATE branch (security/payment/critical keywords)

**Given** the SPEC is Tier M BUT the SPEC scope contains security-domain keywords (e.g., "authentication", "authorization", "secret", "credential", "token", "crypto") OR payment-domain keywords (e.g., "payment", "billing", "stripe", "pgp") OR critical-domain keywords (per design.md §F.3 list)
**When** the orchestrator evaluates the safe-condition predicate
**Then** condition (d) FAILS (dangerous keyword detected)
**And** the Implementation Kickoff Approval gate REMAINS a mandatory explicit `AskUserQuestion`
**And** the matched keyword is logged to `progress.md` for audit
**Evidence**: regression tests for each keyword category (security, payment, critical) construct SPECs with the keyword; the predicate returns explicit-gate; the matched keyword is logged.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-005.

---

#### AC-IGGDA-005c — Safe-condition predicate EXPLICIT-GATE branch (--pr-forcing destructive scope)

**Given** the user supplied `--pr` flag OR the SPEC scope is marked destructive (force-push, `rm -rf`, dropping tables, posting externally — per REQ-IGGDA-027)
**When** the orchestrator evaluates the safe-condition predicate
**Then** condition (d) FAILS (destructive scope)
**And** the Implementation Kickoff Approval gate REMAINS a mandatory explicit `AskUserQuestion`
**Evidence**: regression test with `--pr` flag; the predicate returns explicit-gate.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-005.

---

#### AC-IGGDA-005d — Safe-condition predicate RETURN-TO-PHASE-0 branch (intent clarity < 100%)

**Given** the Socratic interview is incomplete (intent clarity < 100% per CLAUDE.md §7 Rule 5)
**When** the orchestrator evaluates the safe-condition predicate
**Then** condition (a) FAILS
**And** the orchestrator returns to Phase 0 Socratic interview (NOT explicit-gate — the interview must complete first)
**Evidence**: regression test with incomplete intent; the predicate returns "return to Phase 0".

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-005, REQ-IGGDA-007.

---

#### AC-IGGDA-005e — Safe-condition predicate EXPLICIT-GATE branch (plan-auditor FAIL/INCONCLUSIVE)

**Given** the plan-auditor verdict is FAIL or INCONCLUSIVE (NOT PASS)
**When** the orchestrator evaluates the safe-condition predicate
**Then** condition (b) FAILS
**And** the orchestrator surfaces the verdict to the user via `AskUserQuestion` (per REQ-IGGDA-023 — independent-audit halt)
**Evidence**: regression test with plan-auditor FAIL; the predicate returns "surface to user".

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-005, REQ-IGGDA-023.

---

#### AC-IGGDA-006 — Safe-condition predicate decision logged to progress.md

**Given** the safe-condition predicate has been evaluated (either branch)
**When** the orchestrator advances (or halts) based on the verdict
**Then** `progress.md §E IGGDA Kickoff Predicate` contains: condition (a) value, condition (b) value, condition (c) value, condition (d) value, final verdict (auto-proceed / explicit-gate / return-to-Phase-0 / surface-to-user), timestamp
**Evidence**: `grep -A 8 "IGGDA Kickoff Predicate" .moai/specs/SPEC-V3R6-ORCH-IGGDA-001/progress.md` returns all 4 conditions + the verdict.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-006.

---

#### AC-IGGDA-007 — Phase 0.5 NOT skipped by predicate

**Given** the safe-condition predicate holds (all 4 conditions)
**When** the orchestrator evaluates the auto-proceed verdict
**Then** Phase 0.5 (Plan Audit Gate) was already completed BEFORE the predicate evaluation (Phase 0.5 → predicate → Phase 1, in order)
**And** the predicate does NOT substitute for Phase 0.5
**Evidence**: the run.md ordering documents Phase 0.5 BEFORE the Implementation Kickoff Approval gate; the predicate is evaluated AFTER Phase 0.5 PASS.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-007.

---

#### AC-IGGDA-008 — `[ZONE:Frozen]` marker preserved

**Given** M1 has amended the Implementation Kickoff Approval behavior (safe-condition predicate)
**When** `grep -n "\[ZONE:Frozen\]" internal/template/templates/.claude/skills/moai/workflows/run.md internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md` is run
**Then** the marker is STILL PRESENT on Implementation Kickoff Approval (the BEHAVIOR is amended, the MARKER is preserved-as-audit-trail)
**Evidence**: grep returns ≥ 1 match for the marker on Implementation Kickoff Approval in both files.

**Severity**: MUST-PASS. **Trace**: §G Out of Scope — `[ZONE:Frozen]` marker removal.

---

### §D.2 — D1 + D4 IGGDA 4-phase pipeline + Stop hook driver

#### AC-IGGDA-009 — 4-phase pipeline defined in order

**Given** the IGGDA pipeline definition (D1)
**When** the orchestrator traverses the pipeline
**Then** the 4 phases execute in order: Phase 0 Intent → Phase 1 Plan → Phase 2 Run → Phase 3 Sync + final audit
**And** no phase is skipped (except by graceful degradation per AC-IGGDA-017)
**Evidence**: `grep -n "Phase 0\|Phase 1\|Phase 2\|Phase 3" internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md` returns the 4 phases in order.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-008.

---

#### AC-IGGDA-010 — Phase 1 safe-transition (plan-auditor PASS + Implementation Kickoff Approval cleared)

**Given** Phase 1 (Plan) has completed (manager-spec produced SPEC artifacts + plan-auditor issued PASS verdict)
**When** the Stop hook driver evaluates the Phase 1 safe-transition predicate
**Then** the driver reads `progress.md` for the plan-auditor PASS marker AND the Implementation Kickoff Approval cleared marker
**And** emits a `/goal`-style auto-advance signal to Phase 2
**Evidence**: regression test simulates Phase 1 completion; the driver emits the auto-advance signal.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-009, REQ-IGGDA-010.

---

#### AC-IGGDA-011 — Phase 2 safe-transition (all blocking ACs PASS + go test exit 0 + no out-of-scope test file modified)

**Given** Phase 2 (Run) has completed (all blocking ACs PASS + `go test ./...` exit 0 + `git status` shows no test file outside SPEC scope modified)
**When** the Stop hook driver evaluates the Phase 2 safe-transition predicate
**Then** the driver emits a `/goal`-style auto-advance signal to Phase 3
**Evidence**: regression test simulates Phase 2 completion; the driver emits the auto-advance signal. This AC inherits the SPEC-AUTONOMY-RUN-GOAL-001 `ac_converge` predicate structure.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-009.

---

#### AC-IGGDA-012 — Phase 3 safe-transition (sync-auditor ≥ threshold + moai spec audit 0 MUST-FIX)

**Given** Phase 3 (Sync + final audit) has completed (sync-auditor 4-dimension score ≥ threshold AND `moai spec audit --json --filter-spec=<SPEC-ID>` reports 0 MUST-FIX drift AND `git status` is clean)
**When** the Stop hook driver evaluates the Phase 3 safe-transition predicate
**Then** the driver emits the terminal `/goal clear` signal (IGGDA-complete)
**Evidence**: regression test simulates Phase 3 completion; the driver emits the terminal signal.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-009, REQ-IGGDA-028.

---

#### AC-IGGDA-013 — Stop hook driver reads progress.md + moai spec audit (NOT frontmatter inference)

**Given** the Stop hook driver is evaluating a phase transition
**When** the driver determines phase completion
**Then** the driver reads `progress.md` §E markers AND invokes `moai spec audit --json --filter-spec=<SPEC-ID>`
**And** the driver NEVER infers phase completion from frontmatter text alone (per verification-claim-integrity.md §1.1 surface 3)
**Evidence**: `grep -E "cat .moai/specs|moai spec audit --json --filter-spec" internal/template/templates/.claude/hooks/moai/iggda-phase-driver.sh` returns ≥2 matches; `grep -E "status:implemented|frontmatter" internal/template/templates/.claude/hooks/moai/iggda-phase-driver.sh` returns 0 (no frontmatter-only inference branch).

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-010.

---

#### AC-IGGDA-014 — Recovery-Signal Carve-Out compliance (exit 0 on recovery turn)

**Given** the Stop hook driver fires on a turn whose `stopReason` indicates a recovery signal (PTL, max_output_tokens, media_size, compact-failure) OR a withheld-recoverable error
**When** the driver evaluates the turn
**Then** the driver exits 0 (allows the turn to end / the recovery to proceed)
**And** does NOT exit 2 (block) — per the Recovery-Signal Carve-Out (runtime-recovery-doctrine.md §4)
**Evidence**: regression test simulates a recovery-turn `stopReason`; the driver exits 0.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-011.

---

#### AC-IGGDA-015 — Graceful degradation when /goal unavailable

**Given** the `/goal` primitive is unavailable (runtime version < v2.1.139 OR `disableAllHooks` OR `allowManagedHooksOnly`)
**When** the IGGDA pipeline attempts to set a `/goal`
**Then** the pipeline degrades gracefully to per-turn progression (orchestrator drives each phase manually)
**And** does NOT fail
**Evidence**: regression test simulates `/goal` unavailability; the pipeline emits a graceful-degradation log and continues per-turn.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-012.

---

#### AC-IGGDA-016 — Stop hook driver returns exit codes + JSON (NEVER AskUserQuestion)

**Given** the Stop hook driver is firing
**When** the driver determines a block (exit 2) is required
**Then** the driver returns exit code 2 + JSON with `continue`, `stopReason`, `details`, and `ledger_note` fields
**And** the driver NEVER invokes `AskUserQuestion` (the orchestrator translates the block to `AskUserQuestion`)
**Evidence**: `grep -n "AskUserQuestion" internal/template/templates/.claude/hooks/moai/iggda-phase-driver.sh` returns 0 matches; `grep -n "ledger_note" ...` returns the JSON field.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-013, B11.

---

#### AC-IGGDA-017 — Stop hook driver never invokes AskUserQuestion (subagent boundary)

**Given** the Stop hook driver is registered as a Stop hook
**When** `grep -rn 'AskUserQuestion\|mcp__askuser' internal/template/templates/.claude/hooks/moai/iggda-phase-driver.sh internal/template/templates/.claude/hooks/moai/handle-iggda-phase-driver.sh` is run
**Then** 0 matches (the hook respects the orchestrator-subagent boundary per agent-common-protocol.md § Hook Invocation Surface)
**Evidence**: the grep returns 0 matches.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-013, C-HRA-008 family.

---

#### AC-IGGDA-018 — Stop hook timeout ≤ 5 seconds

**Given** the Stop hook driver is registered in `settings.json`
**When** `grep -A 3 "iggda-phase-driver" internal/template/templates/.claude/settings.json.tmpl` is run
**Then** the `timeout` field is ≤ 5 (per CLAUDE.local.md §7 MoAI policy default)
**Evidence**: grep returns the timeout value.

**Severity**: MUST-PASS. **Trace**: CLAUDE.local.md §7.

---

#### AC-IGGDA-019 — Stop hook cross-platform (Windows stub)

**Given** the Stop hook driver is a shell script
**When** `GOOS=windows GOARCH=amd64 go build ./...` is run (or the CI Windows job runs)
**Then** a Windows stub exists (`iggda-phase-driver_windows.cmd` or equivalent) that emits a graceful-degradation log
**And** does NOT break the Windows build
**Evidence**: the Windows stub file exists; the Windows CI job passes.

**Severity**: SHOULD-PASS. **Trace**: CLAUDE.local.md §7.

---

#### AC-IGGDA-020 — Driver does not write to runtime-managed paths

**Given** the Stop hook driver is firing
**When** the driver reads `progress.md` and emits its JSON verdict
**Then** the driver does NOT write to `.moai/harness/*`, `.moai/state/*`, `.moai/cache/*` (runtime-managed per B8)
**And** writes only to its own log location (if any)
**Evidence**: the driver's shell script source shows no write to runtime-managed paths.

**Severity**: MUST-PASS. **Trace**: B8.

---

### §D.3 — D3 Bounded recursive self-diagnosis loop

#### AC-IGGDA-021 — Max-3-iteration bound enforced

**Given** the recursive self-diagnosis loop is active on a mechanical failure
**When** the loop completes iteration 3 without VERIFY exit 0
**Then** the loop halts (iteration 4 is prohibited)
**And** the orchestrator is signaled to run an `AskUserQuestion` escalation (continue manual / revert + re-plan / abort)
**Evidence**: regression test simulates 3 failing iterations; the loop halts and signals escalation.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-014.

---

#### AC-IGGDA-022 — Mechanical failure triggers DIAGNOSE-PATCH-VERIFY

**Given** a mechanical failure (lint rule violation, type error, build error, missing import, formatting drift) is surfaced during Phase 2
**When** the recursive self-diagnosis sub-agent activates
**Then** the sub-agent executes DIAGNOSE (read failing output) → PATCH (minimal fix) → VERIFY (re-run check)
**And** on VERIFY exit 0, advances; on VERIFY still failing, increments iteration counter and repeats from DIAGNOSE
**Evidence**: the run.md `## Recursive Self-Diagnosis Loop (bounded)` section documents the pattern; a regression test simulates a lint error and verifies the loop patches it.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-015.

---

#### AC-IGGDA-023 — Semantic failure escalates IMMEDIATELY (NEVER auto-fixed)

**Given** a semantic failure (data race, deadlock, panic, test assertion failure, concurrency hazard) is surfaced during Phase 2
**When** the recursive self-diagnosis loop evaluates the failure classification
**Then** the loop halts IMMEDIATELY (no DIAGNOSE-PATCH-VERIFY attempt)
**And** the orchestrator is signaled to run an `AskUserQuestion` human escalation
**And** the semantic failure is NEVER auto-patched
**Evidence**: regression test simulates a data race; the loop halts without patching and signals escalation. This AC inherits SPEC-AUTONOMY-RUN-GOAL-001 REQ-ARG-009 semantic-failure escalation.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-016, run.md:152.

---

#### AC-IGGDA-024 — All 5 circuit-breaker invariants honored

**Given** the recursive self-diagnosis loop is active
**When** the loop evaluates its compliance with runtime-recovery-doctrine.md §3
**Then** all 5 invariants are honored: (1) max-3 same-rung failures → escalate rung; (2) no self-loop within one turn (`hasAttemptedReactiveCompact`); (3) compact-can-PTL last-resort escape to rung-4 abort + preserve; (4) abort-closes-ledger (persist to `progress.md`); (5) narrative-consistency (5-Section Evidence-Bearing Report)
**Evidence**: the run.md recursive-loop section cross-references all 5 invariants; a regression test simulates each invariant's trigger condition and verifies compliance.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-017.

---

#### AC-IGGDA-025 — Iteration log written to progress.md

**Given** each recursive self-diagnosis iteration completes
**When** the iteration's VERIFY result is obtained
**Then** `progress.md §E Recursive Self-Diagnosis Log` is appended with: iteration number, failure classification (mechanical/semantic), root-cause summary, patch summary, VERIFY result, (on halt) escalation reason
**Evidence**: `grep -A 10 "Recursive Self-Diagnosis Log" .moai/specs/<SPEC-ID>/progress.md` returns the iteration entries.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-018.

---

#### AC-IGGDA-026 — Loop does not modify forbidden paths

**Given** the recursive self-diagnosis loop is patching
**When** the PATCH step selects files to modify
**Then** the loop does NOT modify `.env`, `.env.*`, credentials files, secrets, `scripts/ci-watch/run.sh`, or any Wave 2 infrastructure scripts
**And** the patch surface is the SPEC scope only
**Evidence**: the run.md recursive-loop section enumerates the forbidden-paths list (inherited from CONST-V3R5-011/013 + manager-develop-prompt-template.md autofix escalation).

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-019.

---

### §D.4 — D5 Independent-audit preservation

#### AC-IGGDA-027 — plan-auditor spawned in fresh context (NOT continuation of manager-spec)

**Given** the IGGDA pipeline reaches Phase 1 plan-audit
**When** the orchestrator spawns plan-auditor
**Then** plan-auditor is a separate `Agent(subagent_type: "plan-auditor")` call in a fresh isolated context
**And** NOT a continuation of manager-spec's turn
**And** plan-auditor's verdict is the FINAL guarantee that Phase 1 is safe to advance
**Evidence**: the orchestration-mode-selection.md § IGGDA Independent-Audit Preservation documents the fresh-context spawn; a regression guard verifies the spawn is a separate `Agent()` call.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-020, REQ-IGGDA-021.

---

#### AC-IGGDA-028 — sync-auditor spawned in fresh context (NOT continuation of manager-docs)

**Given** the IGGDA pipeline reaches Phase 3 sync-audit
**When** the orchestrator spawns sync-auditor
**Then** sync-auditor is a separate `Agent(subagent_type: "sync-auditor")` call in a fresh isolated context
**And** NOT a continuation of manager-docs's turn
**Evidence**: regression guard verifies the spawn is a separate `Agent()` call.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-020, REQ-IGGDA-021.

---

#### AC-IGGDA-029 — FAIL/INCONCLUSIVE auditor verdict halts auto-advance

**Given** plan-auditor OR sync-auditor returns FAIL or INCONCLUSIVE
**When** the Stop hook driver receives the verdict
**Then** the driver halts auto-advance (no transition to the next phase)
**And** the orchestrator surfaces the verdict to the user via `AskUserQuestion`
**And** the FAIL/INCONCLUSIVE is a hard stop regardless of prior phase PASS
**Evidence**: regression test simulates plan-auditor FAIL; the driver halts; the orchestrator surfaces via `AskUserQuestion`.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-023.

---

#### AC-IGGDA-030 — "Self-audit" vs "Independent audit" disambiguation documented

**Given** the IGGDA documentation uses both "self-audit" and "independent audit"
**When** a reader searches the orchestration-mode-selection.md
**Then** the disambiguation is explicit: "self-audit" = D3 recursive loop (Phase 2 first-party); "independent audit" = plan-auditor (Phase 1) + sync-auditor (Phase 3)
**And** the two are documented as complementary, NOT interchangeable
**Evidence**: `grep -n "self-audit\|independent audit" internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md` returns the disambiguation section.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-022.

---

### §D.5 — IGGDA-completeness terminal gate (REQ-IGGDA-028)

#### AC-IGGDA-031 — IGGDA-completeness requires moai spec audit 0 MUST-FIX + sync-auditor ≥ threshold + git clean

**Given** the IGGDA pipeline has reached Phase 3 convergence
**When** the Stop hook driver evaluates the terminal IGGDA-completeness predicate
**Then** ALL THREE hold: (1) `moai spec audit --json --filter-spec=<SPEC-ID>` reports `drift_findings: []` (0 MUST-FIX); (2) sync-auditor 4-dimension score ≥ threshold; (3) `git status` is clean
**And** the `/goal` clears ONLY when all three hold
**Evidence**: regression test simulates each of the 3 conditions failing; the `/goal` does not clear until all 3 hold.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-028.

---

### §D.6 — Cross-cutting (inherited C-series + IGGDA additions)

#### AC-IGGDA-032 — Preferences collected before autonomy (inherited C2)

**Given** the IGGDA pipeline is about to launch Phase 1–3 autonomous agents
**When** the orchestrator confirms the pre-launch state
**Then** ALL user preferences have been collected in Phase 0 (Intent)
**And** no autonomous agent needs to prompt the user mid-run
**Evidence**: the Phase 0 Socratic interview log in `progress.md` enumerates the collected preferences.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-024.

---

#### AC-IGGDA-033 — Flat hierarchy (inherited C3 — no agent spawns agent)

**Given** the IGGDA pipeline is executing autonomously
**When** any autonomous agent (recursive-self-diagnosis, Workflow, `/goal`-turn) needs to delegate
**Then** the agent returns a blocker report; the ORCHESTRATOR spawns the next agent
**And** no autonomous agent spawns another autonomous agent (flat hierarchy per Anthropic Finding A1)
**Evidence**: grep the autonomous-agent prompt bodies for `Agent(` spawn calls; 0 matches (the orchestrator is the sole spawner).

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-025.

---

#### AC-IGGDA-034 — Background agents read-only (inherited C4)

**Given** the IGGDA pipeline spawns a background agent (`run_in_background: true`)
**When** the background agent attempts Write/Edit
**Then** the attempt is denied (background write prohibition)
**And** the recursive-self-diagnosis sub-agent runs in FOREGROUND (it patches code)
**Evidence**: the recursive-self-diagnosis spawn in the orchestrator's prompt body specifies `run_in_background: false`.

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-026.

---

#### AC-IGGDA-035 — Autonomy does not authorize destructive operations (REQ-IGGDA-027)

**Given** the IGGDA pipeline is executing autonomously
**When** the pipeline reaches a point where a destructive operation (force-push, `rm -rf`, dropping tables, posting externally) or PR creation is contemplated
**Then** the operation is surfaced as an explicit gate AFTER Phase 3 convergence (NOT auto-executed mid-run)
**And** autonomy does NOT pre-approve hard-to-reverse or shared-system actions
**Evidence**: the orchestration-mode-selection.md documents the non-substitution clause (inherited from run.md:154-156).

**Severity**: MUST-PASS. **Trace**: REQ-IGGDA-027.

---

### §D.7 — Indirect verification (coverage / lint / spec-lint)

#### AC-IGGDA-036 — spec-lint clean on this SPEC

**Given** this SPEC's artifacts (spec.md, plan.md, acceptance.md, design.md, research.md) are authored
**When** `moai spec lint .moai/specs/SPEC-V3R6-ORCH-IGGDA-001/` is run
**Then** 0 findings (FrontmatterSchemaRule, OutOfScopeRule, GEARS notation, OwnershipTransitionRule all pass)
**Evidence**: lint output shows 0 findings.

**Severity**: MUST-PASS. **Trace**: §E Self-Verification.

---

#### AC-IGGDA-037 — Go test suite passes (M5 only)

**Given** M5 has added/modified `internal/spec/audit.go` (if the `--filter-spec` flag was absent)
**When** `go test ./internal/spec/...` is run
**Then** all tests pass; coverage on `internal/spec/audit.go` ≥ 85%
**Evidence**: `go test -cover ./internal/spec/...` output.

**Severity**: MUST-PASS (if M5 touches Go); N/A (if M5 is a no-op). **Trace**: plan.md M5.

---

#### AC-IGGDA-038 — Template neutrality CI guard passes

**Given** the template edits (M1–M4) have been made in `internal/template/templates/`
**When** `go test ./internal/template/... -run TestTemplateNeutralityAudit` is run
**Then** 0 internal-content-leak findings (no SPEC IDs, commit SHAs, macOS-bias paths, `feedback_` refs in template content)
**Evidence**: test output shows 0 findings.

**Severity**: MUST-PASS. **Trace**: CLAUDE.local.md §2.1, §25.

---

## §E — Quality Gate Criteria

- **MUST-PASS**: AC-IGGDA-001 through AC-IGGDA-038 inclusive, treating AC-IGGDA-005 as 5 sub-branches (005a/005b/005c/005d/005e). Total: 41 MUST-PASS ACs. AC-IGGDA-037 is MUST-PASS when M5 touches Go, N/A when M5 is a no-op (per Q4 pre-flight resolution).
- **SHOULD-PASS**: AC-IGGDA-019 (Windows stub).
- **Definition of Done**: all MUST-PASS ACs pass + spec-lint clean + Tier L PR merged + `moai spec audit` 0 MUST-FIX + sync-auditor ≥ threshold
- **plan-auditor scrutiny focus** (per the FROZEN-invariant amendment risk): AC-IGGDA-004 (auto-proceed branch) + AC-IGGDA-005a–005e (all 5 explicit-gate branches) are the highest-risk acceptance surface. The plan-auditor MUST verify each branch has a regression test and the predicate logic is deterministic.

---

## §F — Edge Cases

1. **EC-1 — Safe-condition predicate with contradictory signals**: Tier M SPEC with security keywords AND 100% intent clarity AND plan-auditor PASS. Condition (d) FAILS (security keyword) → explicit-gate. The predicate evaluates conditions in order (a)→(b)→(c)→(d); the FIRST failing condition determines the verdict rationale, but the gate decision is the same (explicit-gate) for any failure.
2. **EC-2 — Recursive loop hits iteration 3 on a mechanical failure that is actually semantic**: the loop's failure classifier mis-classifies a data race as a "test assertion failure" (which is semantic, but a mis-classifier might treat it as mechanical). REQ-IGGDA-016 + AC-IGGDA-023 enumerate test assertion failure as SEMANTIC — the classifier must align. A regression test for the classifier is in M2.
3. **EC-3 — Stop hook driver fires during a compact**: the Recovery-Signal Carve-Out (AC-IGGDA-014) mandates exit 0. The driver must NOT block a compact-recovery turn.
4. **EC-4 — `moai spec audit` returns non-zero exit but 0 MUST-FIX drift**: the driver treats exit code separately from drift findings. A non-zero exit due to INFO-level findings is NOT a block; only MUST-FIX drift blocks.
5. **EC-5 — User overrides auto-proceed via the lightweight confirmation**: even when the predicate auto-proceeds, the `AskUserQuestion` round is still issued (AC-IGGDA-004). The user CAN select "abort" within the timeout window; the predicate does NOT override user intent (AC-IGGDA-003).

---

## §G — Forward-Looking Checks (post-implementation)

1. **FL-1**: After IGGDA v0.1.0 ships, measure the auto-proceed vs explicit-gate ratio across the next 10 SPECs. If > 80% auto-proceed, the predicate may be too permissive — consider tightening condition (c) or (d). If < 20% auto-proceed, the predicate may be too restrictive — consider relaxing.
2. **FL-2**: After IGGDA v0.1.0 ships, measure the recursive-loop escalation rate. If > 30% of SPECs hit the max-3-iteration bound, the loop may be under-powered — consider raising the bound to 5 OR improving the DIAGNOSE root-cause step.
3. **FL-3**: The version-preflight follow-up (`SPEC-V3R6-IGGDA-PREFLIGHT-001` candidate, EX-1) should be authored if the graceful-degradation path (AC-IGGDA-015) fires more than rarely.

---

Version: 0.1.0 (plan-phase)
Status: Active — awaits plan-auditor independent audit
