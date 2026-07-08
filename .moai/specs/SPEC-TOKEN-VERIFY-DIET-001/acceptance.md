---
id: SPEC-TOKEN-VERIFY-DIET-001
title: "Verification Output Diet — Acceptance Criteria"
version: "0.1.0"
status: in-progress
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/core/agent-common-protocol.md"
lifecycle: spec-anchored
tags: "token-economy, verification, vci, doctrine, acceptance"
---

# SPEC-TOKEN-VERIFY-DIET-001 — Acceptance

## §A Summary

- **AC count**: 8 (`AC-VD-001`..`AC-VD-008`).
- **Tier**: M (standard) — see plan.md §A.4.
- **Era**: V3R6 (3-phase close: plan→run→sync).
- **Traceability**: Each AC binds to one or more of `REQ-001`..`REQ-007` (see §C AC Matrix).
- **Verification posture**: all ACs are mechanically verifiable via `grep` / `sed -n` / `git diff` against the edited files — no subjective judgment calls.

---

## §B Severity Classification

| Severity | ACs | Definition |
|----------|-----|------------|
| **MUST-PASS (blocker)** | AC-VD-001, AC-VD-003, AC-VD-005, AC-VD-006, AC-VD-007, AC-VD-008 | vci preservation, contract existence, keyword non-regression, parallel-execution non-regression, 3-surface consistency, E1-E7 boundary — any failure blocks close |
| **SHOULD-PASS** | AC-VD-002, AC-VD-004 | Bounded-tail ceiling stated, banner file-path column — important but not close-blocking if a debt entry is recorded in progress.md §E.2 with rationale |

---

## §C AC Matrix

| AC ID | Requirement | Milestone | Severity | Verifies |
|-------|-------------|-----------|----------|----------|
| AC-VD-001 | REQ-001, REQ-005 | M1 | MUST-PASS | File-redirect contract subsection exists in `agent-common-protocol.md`; contract terms stated (verbatim → file on disk; context → exit code + bounded tail) |
| AC-VD-002 | REQ-002, REQ-003 | M1 | SHOULD-PASS | Bounded-tail ceiling stated as concrete value; redirect-on-exceed rule present |
| AC-VD-003 | REQ-005 | M1 | MUST-PASS | Contract section explicitly names `verification-claim-integrity` and §1.1 surface 1+2 preservation; explicitly rejects the "drop the evidence" interpretation |
| AC-VD-004 | REQ-001, REQ-004 | M2 | SHOULD-PASS | `moai.md` §8 Verification Matrix + Completion Report banner templates carry a file-path column or row-continuation field; verbatim content NOT embedded as inline row text |
| AC-VD-005 | REQ-006 | M1, M2 | MUST-PASS | 7 verification keywords (`go test`, `coverprofile`, `grep`, `sentinel`, `cmd/moai`, `bench`, `lint`) still grep-able in `agent-common-protocol.md` § Parallel Execution (parallel-execution grep AC at line 339 still passes) |
| AC-VD-006 | REQ-006 | M1, M2 | MUST-PASS | Parallel-execution HARD clause at `agent-common-protocol.md` line 272 preserved (single-turn multi-Bash obligation intact); file-redirect contract alters output representation only |
| AC-VD-007 | REQ-005 | M1, M2 | MUST-PASS | 3-surface consistency: file-redirect contract (or cross-reference to SSOT) appears in all 3 surfaces (`agent-common-protocol.md`, `verification-batch-pattern.md`, `moai.md`) |
| AC-VD-008 | REQ-007 | M2 | MUST-PASS | E1-E7 row structure of `manager-develop.md`'s §E self-verification matrix unchanged — only evidence-surfacing representation is touched |

---

## §D Given-When-Then Scenarios

### AC-VD-001 — File-redirect contract subsection exists

**Given** the file `.claude/rules/moai/core/agent-common-protocol.md` after M1 lands,
**When** an auditor greps the file for the literal `File-redirect contract`,
**Then** the grep returns ≥1 match AND the matching section defines the contract terms: (a) verbatim tool output redirected to a file on disk, (b) conversation context carries exit-code + bounded-tail summary, (c) the cited file path appears in the Verification Matrix / §E row.

**Verification command**:
```bash
grep -n "File-redirect contract" .claude/rules/moai/core/agent-common-protocol.md
```

### AC-VD-002 — Bounded-tail ceiling stated + redirect-on-exceed

**Given** the new `### File-redirect contract` subsection,
**When** an auditor reads the subsection prose,
**Then** the subsection states (a) a concrete bounded-tail ceiling (e.g., `≤50 lines OR ≤2KB`) — NOT vague phrasing like "some lines" or "appropriate amount" — AND (b) the redirect-on-exceed rule: when verbatim output exceeds the ceiling, redirect to disk and surface exit-code + tail in context.

**Verification**: manual read of the subsection; the ceiling value is a concrete number or formula. `grep -nE "50 lines|2KB|ceiling|bounded" .claude/rules/moai/core/agent-common-protocol.md` returns a match near the contract section.

### AC-VD-003 — vci preservation named explicitly

**Given** the new `### File-redirect contract` subsection,
**When** an auditor greps the subsection for the literal `verification-claim-integrity`,
**Then** the grep returns ≥1 match AND the surrounding prose (a) names the §1.1 surface 1 (orchestrator self-report) and surface 2 (manager §E) preservation obligation, AND (b) explicitly states *"NOT 'drop the evidence'"* or equivalent rejection of the drop interpretation.

**Verification commands**:
```bash
grep -n "verification-claim-integrity" .claude/rules/moai/core/agent-common-protocol.md
grep -nE "NOT.*drop|drop the evidence|do not drop|must not drop" .claude/rules/moai/core/agent-common-protocol.md
```

### AC-VD-004 — moai.md §8 banner file-path column

**Given** `.claude/output-styles/moai/moai.md` §8 after M2 lands,
**When** an auditor reads the Verification Matrix banner template (line 407+) and the Completion Report banner template (line 603+),
**Then** both banner templates carry either (a) a file-path column in the row layout OR (b) a row-continuation field (`   └─ evidence: /tmp/...`) for citing redirected verbatim evidence, AND the banner description states that verbatim content is NOT embedded as inline row text. The 5 HARD rules at lines 423-427 remain intact.

**Verification command**:
```bash
grep -nE "file path|filepath|File path|evidence path|Evidence|└─ evidence" .claude/output-styles/moai/moai.md
sed -n '423,427p' .claude/output-styles/moai/moai.md   # 5 HARD rules intact
```

### AC-VD-005 — 7-keyword non-regression

**Given** `agent-common-protocol.md` § Parallel Execution after M1 lands,
**When** an auditor greps the section for the 7 verification keywords,
**Then** all 7 keywords still appear in the section (the parallel-execution grep AC at `agent-common-protocol.md` line 339 — `go test`, `coverprofile`, `grep`, `sentinel`, `cmd/moai`, `bench`, `lint` — still passes).

**Verification command**:
```bash
grep -E "go test|coverprofile|grep|sentinel|cmd/moai|bench|lint" .claude/rules/moai/core/agent-common-protocol.md
# Expected: ≥7 matches (one per keyword; some keywords may match multiple lines)
```

### AC-VD-006 — Parallel-execution HARD clause intact

**Given** `agent-common-protocol.md` § Parallel Execution after M1 lands,
**When** an auditor reads the opening HARD clause (line 272),
**Then** the HARD single-turn multi-Bash obligation remains — the clause still states the orchestrator MUST execute every read-only verification batch as a single-turn multi-Bash call. The file-redirect contract ALTERS output representation (redirect + tail vs inline verbatim), NOT the parallel-execution obligation.

**Verification command**:
```bash
sed -n '270,285p' .claude/rules/moai/core/agent-common-protocol.md
# Expected: HARD clause "MUST execute every read-only verification batch as a single-turn multi-Bash call" intact
```

### AC-VD-007 — 3-surface consistency

**Given** all 3 surfaces after M2 lands,
**When** an auditor greps for the file-redirect contract across the 3 surfaces,
**Then** the grep returns matches in all 3 files (`agent-common-protocol.md`, `verification-batch-pattern.md`, `moai.md`) — either the literal `File-redirect contract` section OR a cross-reference pointing back to the SSOT in `agent-common-protocol.md`.

**Verification command**:
```bash
grep -lE "File-redirect contract|file-redirect contract" \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/workflow/verification-batch-pattern.md \
  .claude/output-styles/moai/moai.md
# Expected: 3 file paths printed (one per surface)
```

### AC-VD-008 — E1-E7 boundary (out-of-scope verification)

**Given** the SPEC body (REQ-007) and the full run-phase changeset,
**When** an auditor diffs the run-phase commits against `.claude/agents/moai/manager-develop.md`,
**Then** no commit modifies the E1-E7 row structure of the §E self-verification matrix section — no E1-E7 row additions or deletions in the diff. Only evidence-surfacing representation language (file-redirect + bounded tail phrasing) is touched, if at all.

**Verification command**:
```bash
git diff <run-phase-base>..HEAD -- .claude/agents/moai/manager-develop.md | grep -E "^[+-].*\bE[1-7]\b"
# Expected: no matches (no E1-E7 row additions/deletions in manager-develop.md)
```

---

## §E Edge Cases

- **EC-1 — Verbatim output below the ceiling**: If a verification command's verbatim output is BELOW the bounded-tail ceiling (e.g., `go run ./cmd/moai --version` produces 1 line), the contract PERMITS inline quotation. The file-redirect obligation triggers only on exceedance. REQ-003 `While` modifier captures this — no redirect required when the ceiling is not exceeded.
- **EC-2 — File path unreachable at audit time**: If the cited file path is in `/tmp/` and the auditor's session has been restarted (OS cleared tmp), the verbatim evidence is temporarily unreachable at audit time. This is a residual risk — the contract requires only that the file path be reachable AT THE TIME the banner is rendered, not indefinitely. Long-term persistence of verification evidence is D's domain (budget-stop handoff infrastructure).
- **EC-3 — Multi-row Verification Matrix with 7+ file paths**: A Verification Matrix with 7 rows, each citing a different file path, adds 7 file-path references. moai.md §8 banner template MUST handle this without overflow (existing columnar layout at lines 408-411 should accommodate; if not, run-phase may switch to row-continuation form per plan.md §I.4).
- **EC-4 — Legacy SPECs grandfathered**: Pre-V3R6 SPECs that authored inline-verbatim Verification Matrices are NOT regenerated. The contract applies to NEW Verification Matrices rendered after M2 lands. Sync-auditor's retrospective reach does not extend to rewriting historical banners.
- **EC-5 — Concurrent verification batches**: If two parallel orchestrator sessions emit Verification Matrices simultaneously, both may write to `/tmp/moai-verify/<session-id>/` — the `<session-id>` discriminator prevents collision. Run-phase chooses the path scheme (plan.md §I.2).

---

## §F Quality Gate Criteria

- [ ] All 6 MUST-PASS ACs green (AC-VD-001, AC-VD-003, AC-VD-005, AC-VD-006, AC-VD-007, AC-VD-008).
- [ ] SHOULD-PASS ACs (AC-VD-002, AC-VD-004) green OR recorded as debt with rationale in progress.md §E.2.
- [ ] 7 verification keywords still grep-able in `agent-common-protocol.md` (AC-VD-005).
- [ ] Parallel-execution HARD clause at line 272 intact (AC-VD-006).
- [ ] vci §1.1 surface 1+2 preservation named in the contract section (AC-VD-003).
- [ ] 3-surface consistency (AC-VD-007) — contract reachable from all 3 surfaces.
- [ ] No E1-E7 structural change in manager-develop.md (AC-VD-008).
- [ ] Template-First Rule honored — edits applied to template source first OR identically in both trees (KI-4).
- [ ] moai.md §8 5 HARD rules at lines 423-427 preserved (KI-5).
- [ ] No snake_case frontmatter aliases; SPEC ID matches canonical regex.

---

## §G Definition of Done

- [ ] M1 + M2 milestones complete (see plan.md §F).
- [ ] All ACs in §F Quality Gate Criteria green (or debt-recorded for SHOULD-PASS with rationale).
- [ ] `verification-claim-integrity.md` §1.1 surface 1+2 explicitly preserved (AC-VD-003 + AC-VD-007).
- [ ] progress.md §E.2 Run-phase Evidence populated with verbatim command outputs (file-redirected, with cited paths) — **manager-develop owns this at run-phase**.
- [ ] progress.md §E.3 Run-phase Audit-Ready Signal green — **manager-develop owns**.
- [ ] sync-phase: manager-docs closes 3-phase (plan→run→sync) per V3R6 lifecycle; progress.md §E.4 `sync_commit_sha` populated — **manager-docs owns**.
- [ ] sync-auditor 4-dimension score ≥ Tier-M threshold (standard 0.80 or project-configured).
