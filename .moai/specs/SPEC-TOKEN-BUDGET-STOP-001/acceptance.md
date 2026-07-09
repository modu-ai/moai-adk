---
id: SPEC-TOKEN-BUDGET-STOP-001
title: "Token Budget Graceful-Abort + /tmp Evidence Persistence — Acceptance Criteria"
version: "0.1.0"
status: in-progress
created: 2026-07-09
updated: 2026-07-09
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/runtime/budget.go"
lifecycle: spec-anchored
era: V3R6
tags: "token-economy, budget, graceful-abort, handoff, file-redirect, acceptance"
---

# SPEC-TOKEN-BUDGET-STOP-001 — Acceptance Criteria

> Acceptance criteria (AC) are observable, testable assertions derived from spec.md §Requirements (GEARS). Each AC traces to one or more REQs. Given-When-Then scenarios provide the verification procedure.

## §D AC Matrix

| AC ID | REQ trace | Severity | Description |
|-------|-----------|----------|-------------|
| AC-TBS-001 | REQ-TBS-001 | MUST-PASS | Graceful-abort signal emitted when `IsAtHardLimit` returns true |
| AC-TBS-002 | REQ-TBS-002 | MUST-PASS | Auto-generated handoff conforms to session-handoff.md 6-block format |
| AC-TBS-003 | REQ-TBS-003 | MUST-PASS | `/clear` NEVER auto-invoked (grep + existing test) |
| AC-TBS-004 | REQ-TBS-004 | MUST-PASS | `RecordCall` continues to return no error (warning-first preserved) |
| AC-TBS-005 | REQ-TBS-005 | MUST-PASS | Evidence persistence obligation stated in file-redirect contract |
| AC-TBS-006 | REQ-TBS-006 | MUST-PASS | vci §1.1 surface 1+2 preserved — cited path resolves to verbatim evidence |
| AC-TBS-007 | REQ-TBS-007 | MUST-PASS | Template-First Rule — doctrine edits mirrored to template tree |
| AC-TBS-008 | (boundary) | MUST-PASS | No mechanical hook script authored (no new `.claude/hooks/moai/*.sh`) |
| AC-TBS-009 | (boundary) | MUST-PASS | No true hard-fail (`RecordCall` signature unchanged — no error return added) |

---

## §D.1 Severity Classification

- **MUST-PASS** (9): AC-TBS-001 through AC-TBS-009. All 9 must pass for the SPEC to close. No SHOULD-PASS or NICE-TO-HAVE ACs in this SPEC — the scope is tight (budget graceful-abort + evidence persistence), and all criteria are load-bearing.

---

## §D.2 Given-When-Then Scenarios

### AC-TBS-001 — Graceful-abort signal emitted when IsAtHardLimit returns true

**Given** a `Tracker` initialized with `HardClearThreshold = 0.90` and `PerAgentBudget["default"] = 1000`
**When** an agent's cumulative usage reaches 900 tokens (90% of 1000) via `RecordCall`
**Then** a graceful-abort method on `Tracker` returns a non-empty handoff message AND a recommendation signal (boolean or non-empty string), WITHOUT returning an error that blocks the next call.

**Verification command**:
```bash
go test -run TestGracefulAbortAtHardLimit ./internal/runtime/ -v
```

**Expected output**: test PASS — the graceful-abort method returns a non-empty handoff string when `IsAtHardLimit` returns true; the method does NOT return an error.

---

### AC-TBS-002 — Auto-generated handoff conforms to session-handoff.md 6-block format

**Given** the graceful-abort method is triggered (agent at hard-limit)
**When** the runtime generates the paste-ready resume message
**Then** the message contains:
- The canonical cut-line markers (`✂──── 여기부터 복사 ────✂` / `✂──── 여기까지 복사 ────✂`)
- The `ultrathink.` opener
- An `applied lessons:` line (or equivalent placeholder)
- A `전제 검증:` / `Preconditions:` header with numbered preconditions
- A `실행:` / `Run:` header with a single primary action
- A `머지 후:` / `After merge:` or `후속:` / `Follow-up:` header (Block 6, when applicable)

**Verification command**:
```bash
go test -run TestGracefulAbortHandoffFormat ./internal/runtime/ -v
```

**Expected output**: test PASS — the generated handoff string contains all 6-block structure tokens. (The exact field values are run-phase discretion; the STRUCTURE is the acceptance criterion.)

---

### AC-TBS-003 — /clear NEVER auto-invoked

**Given** the `internal/runtime/` package
**When** a grep is performed for `/clear` invocation mechanisms
**Then** the package contains:
- NO `os/exec` import used for `/clear` invocation
- NO shell invocation of `clear` or `/clear`
- NO `syscall.Exec` used to replace the process with a clear command

**Verification commands**:
```bash
# 1. Grep for os/exec import in internal/runtime/
grep -rn '"os/exec"' internal/runtime/ | grep -v "_test.go" | grep -v "// "
# Expected: no matches (os/exec not imported for /clear invocation)

# 2. Grep for /clear or clear invocation
grep -rn '/clear\|clear()' internal/runtime/ | grep -v "_test.go" | grep -v "// " | grep -v "TestNoAuto"
# Expected: no matches (no /clear invocation in production code)

# 3. Existing test continues to pass
go test -run TestNoAutoClearInvocation ./internal/runtime/ -v
# Expected: PASS
```

**Expected output**: grep returns no matches; existing `TestNoAutoClearInvocation` (budget_test.go:328-341) continues to PASS.

---

### AC-TBS-004 — RecordCall continues to return no error (warning-first preserved)

**Given** a `Tracker` with `PerAgentBudget["default"] = 1000`
**When** `RecordCall("default", 900, 0)` is called (90% — at hard-limit) followed by `RecordCall("default", 500, 0)` (140% — over budget)
**Then** `RecordCall` returns without error (the function signature is `func (t *Tracker) RecordCall(agentName string, tokensIn, tokensOut int)` — no error return) AND `IsAtHardLimit("default")` returns true.

**Verification command**:
```bash
go test -run TestRecordCallNoErrorOnBudgetExhaustion ./internal/runtime/ -v
```

**Expected output**: test PASS — `RecordCall` compiles without an error return (signature unchanged); calling it on an over-budget agent completes without panic or error.

**Additional grep verification**:
```bash
# RecordCall signature unchanged (no error return added)
grep -n 'func (t \*Tracker) RecordCall' internal/runtime/budget.go
# Expected: "func (t *Tracker) RecordCall(agentName string, tokensIn, tokensOut int)" — NO error return
```

---

### AC-TBS-005 — Evidence persistence obligation stated in file-redirect contract

**Given** the file-redirect contract section in `.claude/rules/moai/core/agent-common-protocol.md`
**When** a reader inspects the contract
**Then** the contract:
- Names `.moai/state/verify/<session>/` as the persistent evidence location
- States that evidence MUST remain reachable at audit time, including after `/tmp` clearance
- States that the exact persist mechanism (direct write vs. /tmp write + copy step) is a run-phase implementation detail

**Verification command**:
```bash
# 1. Persistence location named
grep -n '\.moai/state/verify' .claude/rules/moai/core/agent-common-protocol.md
# Expected: ≥1 match (the persistent evidence location is named)

# 2. Reachability obligation stated
grep -n 'reachable at audit time\|survives.*tmp\|persistent' .claude/rules/moai/core/agent-common-protocol.md
# Expected: ≥1 match (the persistence obligation is stated)
```

**Expected output**: grep returns matches confirming the persistence obligation is stated in the doctrine.

---

### AC-TBS-006 — vci §1.1 surface 1+2 preserved — cited path resolves to verbatim evidence

**Given** the file-redirect contract section in `.claude/rules/moai/core/agent-common-protocol.md` and the Verification Matrix / Completion Report banners in `.claude/output-styles/moai/moai.md`
**When** a reader inspects both surfaces
**Then**:
- The contract explicitly names `verification-claim-integrity` and the §1.1 surface 1+2 preservation obligation
- The contract explicitly rejects the "drop the evidence" interpretation
- The cited path in a Verification Matrix row (moai.md §8) resolves to a file containing the command + full verbatim output

**Verification command**:
```bash
# 1. vci preservation stated
grep -n 'verification-claim-integrity\|§1.1\|surface 1\|surface 2' .claude/rules/moai/core/agent-common-protocol.md
# Expected: ≥1 match (the vci preservation obligation is named)

# 2. "drop the evidence" rejected
grep -n 'drop the evidence\|NOT.*drop' .claude/rules/moai/core/agent-common-protocol.md
# Expected: ≥1 match (the "drop the evidence" interpretation is explicitly rejected)

# 3. Evidence path in moai.md §8 banner
grep -n 'evidence:' .claude/output-styles/moai/moai.md
# Expected: matches at lines ~413 + ~613 referencing the evidence path
```

**Expected output**: grep confirms vci preservation is stated, "drop the evidence" is rejected, and the evidence path is cited in the banner.

---

### AC-TBS-007 — Template-First Rule — doctrine edits mirrored to template tree

**Given** the doctrine edits to `agent-common-protocol.md` and `moai.md` in the LIVE tree
**When** the template-tree mirrors are inspected
**Then** both template files carry the same persistence-obligation edits as their LIVE-tree counterparts AND no Go code is mirrored to the template tree.

**Verification commands**:
```bash
# 1. Template mirrors exist and carry the persistence obligation
grep -n '\.moai/state/verify' internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md
# Expected: ≥1 match (same edit mirrored)

grep -n 'evidence:.*verify' internal/template/templates/.claude/output-styles/moai/moai.md
# Expected: ≥1 match (same evidence-path edit mirrored)

# 2. No Go code mirrored to template tree
ls internal/template/templates/internal/runtime/ 2>&1
# Expected: "No such file or directory" (Go code NOT templated)
```

**Expected output**: template mirrors carry the same edits; `internal/template/templates/internal/runtime/` does not exist.

---

### AC-TBS-008 — No mechanical hook script authored

**Given** the D SPEC's run-phase deliverables
**When** the `.claude/hooks/moai/` directory is inspected for new files
**Then** NO new hook script is authored by D. D's persistence layer is runtime/Go + doctrine, not a hook.

**Verification command**:
```bash
# Check for new hook scripts authored by D (git diff against pre-D HEAD)
git diff --name-only <pre-D-HEAD>..HEAD -- .claude/hooks/moai/
# Expected: no new files (D does not author hook scripts)

# Alternatively, if pre-D-HEAD is not available, check that no hook script
# references the BUDGET-STOP SPEC ID:
grep -rn 'SPEC-TOKEN-BUDGET-STOP-001\|BUDGET-STOP' .claude/hooks/moai/
# Expected: no matches (no hook script references D)
```

**Expected output**: no new hook scripts authored by D; no hook script references the BUDGET-STOP SPEC ID.

---

### AC-TBS-009 — No true hard-fail (RecordCall signature unchanged)

**Given** the `RecordCall` method in `internal/runtime/budget.go`
**When** the method signature is inspected after D's run-phase
**Then** the signature remains `func (t *Tracker) RecordCall(agentName string, tokensIn, tokensOut int)` — NO error return was added. The graceful-abort is a separate method, NOT a `RecordCall` error.

**Verification command**:
```bash
# RecordCall signature unchanged
grep -n 'func (t \*Tracker) RecordCall' internal/runtime/budget.go
# Expected: "func (t *Tracker) RecordCall(agentName string, tokensIn, tokensOut int)"
#   — NO "error" return type at the end

# Existing warning-first test continues to pass
go test -run TestHardLimitWarning ./internal/runtime/ -v
# Expected: PASS (TestHardLimitWarning at budget_test.go:126-144 verifies no error is returned)
```

**Expected output**: `RecordCall` signature has no error return; existing `TestHardLimitWarning` continues to PASS.

---

## §D.3 Edge Cases

- **EC-1 — Agent not yet at hard-limit**: When `IsAtHardLimit` returns false, the graceful-abort method returns an empty handoff string (or `shouldAbort = false`). No handoff is generated. The agent continues normally. (Test: `TestGracefulAbortBelowHardLimit` — verify empty return when usage < 90%.)
- **EC-2 — Unknown agent (not in PerAgentBudget)**: The graceful-abort uses the "default" budget (via `budgetFor`). If the "default" budget is exhausted, the graceful-abort triggers. (Test: reuse `TestUnknownAgentUsesDefaultBudget` pattern.)
- **EC-3 — Concurrent RecordCall calls**: Multiple goroutines call `RecordCall` concurrently; one pushes usage over `HardClearThreshold`. The graceful-abort is triggered exactly once (not once per goroutine). The existing `sync.Mutex` in `Tracker` (budget.go:24) protects the graceful-abort trigger. (Test: `TestConcurrentGracefulAbort` — verify single trigger under concurrent calls.)
- **EC-4 — /tmp already cleared at audit time**: Evidence was written to `/tmp/moai-verify/` and `/tmp` was cleared before audit. If D's persist mechanism is Option (a) (/tmp + copy), the copy step MUST have completed before the crash. If Option (b) (direct write to `.moai/state/verify/`), the evidence survives. The doctrine obligation (AC-TBS-005) states the reachability requirement; the Go mechanism enforces it. (Test: `TestEvidencePersistsAfterTmpClear` — write evidence, simulate /tmp clear, verify `.moai/state/verify/` path resolves.)
- **EC-5 — Graceful-abort when no SPEC ID is available**: The graceful-abort method needs a `specID` for the handoff message. If no SPEC is active (e.g., a utility command), the handoff uses a generic fallback. (Test: `TestGracefulAbortNoSpecID` — verify graceful degradation.)

---

## §D.4 Indirect Verification (grep-based sentinels)

- **IV-1 — GEARS notation compliance**: `grep -rn 'IF.*THEN' .moai/specs/SPEC-TOKEN-BUDGET-STOP-001/spec.md` → no matches (no legacy IF/THEN modality).
- **IV-2 — Out of Scope H3 sub-headings**: `grep -c '### Out of Scope —' .moai/specs/SPEC-TOKEN-BUDGET-STOP-001/spec.md` → ≥1 (satisfies `OutOfScopeRule` lint).
- **IV-3 — 12 canonical frontmatter fields**: `grep -c '^id:\|^title:\|^version:\|^status:\|^created:\|^updated:\|^author:\|^priority:\|^phase:\|^module:\|^lifecycle:\|^tags:' .moai/specs/SPEC-TOKEN-BUDGET-STOP-001/spec.md` → 12 (all canonical fields present).
- **IV-4 — No snake_case aliases**: `grep -c 'created_at:\|updated_at:\|labels:\|spec_id:' .moai/specs/SPEC-TOKEN-BUDGET-STOP-001/spec.md` → 0 (no snake_case aliases).
- **IV-5 — era: V3R6**: `grep -c '^era: V3R6' .moai/specs/SPEC-TOKEN-BUDGET-STOP-001/spec.md` → 1 (era field present).

---

## §D.5 Closure Gates

- **Gate-1 (run-phase)**: All 9 MUST-PASS ACs pass (AC-TBS-001 through AC-TBS-009). Verified by manager-develop §E self-verification.
- **Gate-2 (sync-phase)**: `sync_commit_sha` populated in progress.md §E.4. The single sync commit carries the `implemented → completed` transition.
- **Gate-3 (lint)**: `moai spec lint SPEC-TOKEN-BUDGET-STOP-001` returns 0 findings (or only INFO-level findings). No `FrontmatterInvalid`, `OutOfScopeRule`, `LegacyEARSKeyword`, or `OwnershipTransitionInvalid` findings.

---

## §D.6 Forward-Looking Checks

- **FL-1 — Epic closure**: D is the final SPEC (4/4) in the Token-Economy Epic. Closing D completes the Epic. The Epic memory (`project_token_economy_epic_handoff.md`) should be updated to reflect 4/4 completion.
- **FL-2 — True hard-fail (future SPEC)**: D fulfills the graceful path. If true hard-fail (error return blocking next call) is ever needed, a future SPEC owns it. D's §Out of Scope — True hard-fail records this boundary.
- **FL-3 — Mechanical enforcement hook (future SPEC)**: D's persistence layer is runtime/Go + doctrine. If a mechanical hook that validates cited paths at tool-call time is ever needed, a future SPEC owns it. D's §Out of Scope — Mechanical enforcement hook script records this boundary.

---

## §D.7 Definition of Done

D is "done" when:
1. All 9 MUST-PASS ACs pass (Gate-1).
2. `sync_commit_sha` populated in progress.md §E.4 (Gate-2).
3. `moai spec lint` clean (Gate-3).
4. The Token-Economy Epic reaches 4/4 completion (FL-1).
5. No regression on existing tests (`go test ./internal/runtime/...` PASS, including `TestNoAutoClearInvocation`, `TestHardLimitWarning`, `TestHardLimitAt90Pct`, `TestPersistProgressAt75Pct`).

---

## History

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-07-09 | manager-spec | Initial draft — 9 MUST-PASS ACs + 5 edge cases + 5 indirect verifications + 3 forward-looking checks. Tier M. |
