---
id: SPEC-TOKEN-VERIFY-DIET-001
title: "Verification Output Diet — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
author: manager-spec
priority: P1
phase: "v3.0.0"
module: ".claude/rules/moai/core/agent-common-protocol.md"
lifecycle: spec-anchored
tags: "token-economy, verification, vci, doctrine, plan"
---

# SPEC-TOKEN-VERIFY-DIET-001 — Plan

## §A Context

### §A.1 Problem summary

Three MoAI surfaces currently co-own verification output but none standardizes a file-redirect contract:

1. `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution (canonical 7-item batch — defines parallel-execution obligation, NOT economy obligation).
2. `.claude/rules/moai/workflow/verification-batch-pattern.md` (grouping rationale + Re-sync sentinel at line 30 — explicitly disclaims the verbatim command list).
3. `.claude/output-styles/moai/moai.md` §8 (Verification Matrix + Completion Report banners — re-quote verification results inline as banner row content).

Result: verbatim tool output enters orchestrator context twice (Bash inline + banner re-quote). Originating incident: CHANGELOG full-read autocompact thrashing (see spec.md § Problem GAP-4; current `wc -l CHANGELOG.md` = 7764 lines).

### §A.2 Evidence baselines (measured 2026-07-08 by this agent via Bash, vci §2 attribution)

```
wc -l .claude/rules/moai/core/agent-common-protocol.md        → 456 lines
wc -l .claude/rules/moai/workflow/verification-batch-pattern.md → 70 lines
wc -l .claude/output-styles/moai/moai.md                       → 780 lines
wc -l .claude/rules/moai/core/verification-claim-integrity.md  → 87 lines (preserved invariant)
wc -l CHANGELOG.md                                             → 7764 lines (originating incident artifact)
```

Section anchors (grep-verified):
- `grep -n "Canonical 7-item example" agent-common-protocol.md` → line 284 (M1 anchor).
- `grep -n "Verification Matrix" moai.md` → line 396 (M2 banner anchor).
- `grep -n "Completion Report" moai.md` → line 603 (M2 banner anchor).
- `grep -n "Re-sync sentinel" verification-batch-pattern.md` → line 30 (M2 sentinel anchor).
- `grep -n "Read-only verification batching" agent-common-protocol.md` → line 278 (M1 subsection anchor).

Template-tree mirror existence (Template-First Rule applicability check):
- `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` → exists (29147 bytes).
- `internal/template/templates/.claude/rules/moai/workflow/verification-batch-pattern.md` → exists (3737 bytes).
- `internal/template/templates/.claude/output-styles/moai/moai.md` → exists (56505 bytes).

All 3 surfaces are template-managed → Template-First Rule applies to M1 + M2.

### §A.3 Approach — Doctrine-Primary, Two Milestones

This is a doctrine-primary SPEC. The mechanism is documentation: standardize the file-redirect contract in `agent-common-protocol.md` (M1), then cascade the cross-references in `verification-batch-pattern.md` + `moai.md` §8 (M2). No Go code is touched.

### §A.4 Tier Confirmation — M (standard)

**Tier M confirmed.** Rationale grounded in measured evidence:

- Touches 3 always-loaded / cross-referenced surfaces (`agent-common-protocol.md` is always-loaded per CLAUDE.md core-rule loading; `moai.md` loads at orchestrator launch; `verification-batch-pattern.md` is path-scoped but cross-referenced from the always-loaded file).
- Cascading cross-ref updates — `verification-batch-pattern.md:30` Re-sync sentinel explicitly requires re-sync when the 7-item list changes; this SPEC's representation change triggers that obligation.
- vci preservation obligation (REQ-005) — non-trivial invariant to preserve while changing evidence representation.
- NOT Tier S — 3 files touched (Tier S = isolated single-file edit).
- NOT Tier L — no Go code, no runtime behavior change, no production-code paths touched, doctrine-only.

---

## §B Known Issues (measured)

- **KI-1 — Parallel-execution grep AC**: The 7-item batch's bare-command form is load-bearing for the parallel-execution grep AC at `agent-common-protocol.md` line 339 (`grep -E "go test|coverprofile|grep|sentinel|cmd/moai|bench|lint"`). M1 MUST preserve all 7 keywords when adding file-redirect syntax. Failure = regression on a CI-verified sentinel.
- **KI-2 — moai.md §8 Localization Contract**: moai.md §8 carries a Localization Contract (4 locales en/ko/ja/zh). Any new banner field (file-path column) MUST be considered for localization. File paths themselves are locale-verbatim protocol tokens per moai.md line 734 verbatim-preservation list, but the column HEADER translates per `conversation_language`.
- **KI-3 — Re-sync sentinel scope**: `verification-batch-pattern.md` line 30 Re-sync sentinel currently fires only on 7-item list changes. M2 MUST extend the sentinel to also cover the file-redirect contract representation, or the contract drifts on the next 7-item edit.
- **KI-4 — Template-First Rule**: All 3 edited files live in both LIVE and template trees (mirror existence verified §A.2). Run-phase MUST apply edits to template source first (`internal/template/templates/...`) and rebuild via `make build`, OR identically in both trees. CLAUDE.local.md §2 [HARD] Template-First Rule.
- **KI-5 — moai.md §8 banner HARD rules**: Lines 423-427 carry 5 [HARD] rules governing the Verification Matrix banner (`✓`/`✗` symbols, `(cause: ...)` annotation, max 2 items per line, `📊 N/M PASS` line, criterion-label translation). M2's file-path column MUST NOT violate these rules — it adds a new column, not alters existing HARD-governed rendering.

---

## §C Pre-flight — Devices Verified to Exist

Per `feedback_defect_claim_verification`, all referenced devices were Read/Grep/Bash-verified before this plan was authored:

- ✓ `.claude/rules/moai/core/agent-common-protocol.md` (456 lines; § Parallel Execution at line 270; 7-item batch at line 284; parallel-execution HARD clause at line 272).
- ✓ `.claude/rules/moai/workflow/verification-batch-pattern.md` (70 lines; Re-sync sentinel at line 30).
- ✓ `.claude/output-styles/moai/moai.md` (780 lines; Verification Matrix at line 396; banner template at line 407; Completion Report at line 603; verbatim-preservation list at line 734).
- ✓ `.claude/rules/moai/core/verification-claim-integrity.md` (87 lines; §1.1 binding surfaces 1+2+3).
- ✓ `internal/config/token_budget_guard.go` — sibling device (cross-ref only, NOT this SPEC's target).
- ✓ `internal/runtime/cache_control.go` — sibling device (cross-ref only).
- ✓ Project memory `project_token_economy_epic_handoff.md` — Epic gap definition §C verified to exist (Read in full).
- ✓ Template-tree mirrors for all 3 edit targets — verified to exist (§A.2).

---

## §D Constraints (carried from spec.md)

1. vci preservation (REQ-005) — HARD.
2. Parallel-execution non-regression (REQ-006) — HARD.
3. E1-E7 boundary (REQ-007) — HARD.
4. Doctrine-primary — no Go code.
5. Template-First Rule — edits in template source first OR identically in both trees.
6. GEARS notation — no legacy `IF/THEN`.

---

## §E Self-Verification (plan-phase)

The plan-phase author (manager-spec) self-verifies before handing off to plan-auditor:

- [ ] All 7 GEARS requirements use valid GEARS patterns (Ubiquitous / When / While / Where + generalized subject; no legacy `IF/THEN`).
- [ ] All 4 gap claims (GAP-1..4) cite measured file + measured line range (vci §2 attribution).
- [ ] Out-of-Scope section has ≥1 `### Out of Scope — <topic>` H3 sub-heading with `-` bullets (satisfies `OutOfScopeRule` lint).
- [ ] Frontmatter has all 12 canonical fields, no snake_case aliases (`created`/`updated` NOT `created_at`/`updated_at`; `tags` NOT `labels`).
- [ ] SPEC ID matches `^SPEC(-[A-Z][A-Z0-9]*)+-\d{3}$` (Pre-Write Self-Check printed with `→ PASS` in the authoring turn).
- [ ] Cross-referenced devices verified to exist (§C Pre-flight).
- [ ] progress.md §E skeleton emitted with literal `§E.2`/`§E.3`/`§E.4` headings (era.go parser load-bearing).

---

## §F Milestones

> Per `agent-common-protocol.md` § Time Estimation (CLAUDE.md §17 cross-ref), milestones use priority ordering, NOT time estimates.

### M1 — Priority High — Standardize file-redirect contract in agent-common-protocol.md

**Scope**: Edit `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution / Read-only verification batching / Canonical 7-item example. Apply identically to `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` (Template-First), then `make build` if template source is the edit target.

**Changes**:

1. Add a new subsection `### File-redirect contract` immediately after `### Canonical 7-item example` (i.e., after the closing of the 7-command example block at line 327, before `### Anti-pattern: serial verification across turns` at line 317 — confirm exact insertion point at run-phase by re-reading).
2. The new subsection states the contract verbatim: *"The orchestrator redirects verbatim tool output to a file on disk; conversation context carries only exit code + bounded tail summary. The cited file path MUST appear in the Verification Matrix / §E self-verification row so verbatim evidence remains reachable. This preserves `verification-claim-integrity.md` §1.1 surface 1 (orchestrator self-report) and surface 2 (manager §E self-verification) — the contract is 'verbatim evidence lives on disk with a citable path; context carries exit code + bounded tail', NOT 'drop the evidence'."*
3. Update the 7-item example commands to use file redirection. Example transformation (preserve all 7 keywords):
   - `go test ./... > /tmp/moai-verify/1-go-test.log 2>&1; echo "exit=$?"; tail -50 /tmp/moai-verify/1-go-test.log`
   - Same pattern for items 2-7. PRESERVE the keywords `go test`, `coverprofile`, `grep`, `sentinel`, `cmd/moai`, `bench`, `lint` so the parallel-execution grep AC at line 339 still passes.
4. State the bounded-tail ceiling as a concrete default: **≤50 lines OR ≤2KB, whichever is smaller**. (Run-phase may tune per-domain; the contract holds regardless of the exact number.)

**AC bindings**: AC-VD-001 (contract exists), AC-VD-002 (ceiling + redirect-on-exceed), AC-VD-003 (vci named), AC-VD-005 (7 keywords), AC-VD-006 (HARD clause intact), AC-VD-007 (3-surface consistency partial — SSOT anchor).

**M1 self-verify commands** (run-phase, manager-develop):
```bash
grep -n "File-redirect contract" .claude/rules/moai/core/agent-common-protocol.md
grep -n "verification-claim-integrity" .claude/rules/moai/core/agent-common-protocol.md
grep -E "go test|coverprofile|grep|sentinel|cmd/moai|bench|lint" .claude/rules/moai/core/agent-common-protocol.md
sed -n '270,280p' .claude/rules/moai/core/agent-common-protocol.md   # HARD clause intact
diff .claude/rules/moai/core/agent-common-protocol.md \
     internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md   # Template parity
```

### M2 — Priority Medium — Cross-ref adoption in verification-batch-pattern.md + moai.md §8

**Scope**:

1. `.claude/rules/moai/workflow/verification-batch-pattern.md` line 30 Re-sync sentinel — extend the sentinel clause to cover the file-redirect contract representation (not just the 7-item command list). Wording suggestion: *"If either the 7-item list OR the file-redirect contract representation changes, re-sync this file's grouping rationale and class taxonomy to match."*
2. `.claude/output-styles/moai/moai.md` §8 Verification Matrix banner (line 407) — add a file-path column (or row continuation field) referencing the redirected verbatim evidence. Banner rows cite file path; do NOT re-quote verbatim content as inline row text. Preserve the 5 HARD rules at lines 423-427.
3. `.claude/output-styles/moai/moai.md` §8 Completion Report banner (line 603) — same file-path-column treatment.
4. `moai.md` §8 Localization Contract — confirm the file-path column HEADER translates per `conversation_language` while file-path VALUES remain locale-verbatim (already covered by moai.md line 734 verbatim-preservation list, but state it explicitly next to the new column).
5. Template-First: apply identically to `internal/template/templates/.claude/rules/moai/workflow/verification-batch-pattern.md` and `internal/template/templates/.claude/output-styles/moai/moai.md`; `make build`.

**AC bindings**: AC-VD-004 (banner column), AC-VD-007 (3-surface consistency — all 3 surfaces carry the contract or a cross-ref to SSOT), AC-VD-008 (E1-E7 boundary — confirm no row-structure change in the run-phase diff).

**M2 self-verify commands** (run-phase, manager-develop):
```bash
grep -n "file-redirect\|File-redirect" .claude/rules/moai/workflow/verification-batch-pattern.md
grep -n "file path\|filepath\|File path\|evidence path\|Evidence" .claude/output-styles/moai/moai.md
grep -lE "File-redirect contract|file-redirect contract" \
  .claude/rules/moai/core/agent-common-protocol.md \
  .claude/rules/moai/workflow/verification-batch-pattern.md \
  .claude/output-styles/moai/moai.md   # Expected: 3 file paths
git diff HEAD~N..HEAD -- .claude/agents/moai/manager-develop.md | grep -E "^[+-].*E[1-7]"   # Expected: no matches (AC-VD-008)
```

---

## §G Anti-Patterns (regression hazards)

- **AP-VD-001 — Drop the evidence**: Interpreting "diet" as "skip the verbatim output entirely". REQ-005 explicitly forbids — the cited file path MUST resolve to a file containing the command + full verbatim output.
- **AP-VD-002 — Inline verbatim preservation (triple-burn)**: Keeping the inline verbatim output AND adding a file path alongside (Bash inline + banner inline + file on disk). The contract REPLACES inline verbatim with file-path citation, not ADDS a file path alongside.
- **AP-VD-003 — Parallel-execution regression**: Splitting the 7 commands across turns "because the output is now on disk anyway". REQ-006 forbids — single-turn multi-Bash obligation is preserved (HARD clause at line 272 remains).
- **AP-VD-004 — E1-E7 scope creep**: Rewriting `manager-develop`'s §E matrix row structure while editing the evidence-surfacing column. REQ-007 forbids — AC-VD-008 verifies the diff is clean of E1-E7 row changes.
- **AP-VD-005 — Template drift**: Editing only the LIVE `agent-common-protocol.md` without mirroring to `internal/template/templates/`. CLAUDE.local.md §2 [HARD] Template-First Rule violated; CI guard may flag.
- **AP-VD-006 — Grep-keyword regression**: Rewriting the 7-item batch in a way that drops one of the 7 keywords. The parallel-execution grep AC at `agent-common-protocol.md` line 339 fails. M1 must preserve all 7.
- **AP-VD-007 — moai.md §8 HARD-rule violation**: Adding the file-path column in a way that breaks any of the 5 HARD rules at moai.md lines 423-427 (e.g., introducing a non-`✓`/`✗` symbol, dropping the `📊 N/M PASS` line). M2 must preserve all 5.

---

## §H Cross-References

- **Spec**: `.moai/specs/SPEC-TOKEN-VERIFY-DIET-001/spec.md` (this SPEC's requirements + gap definitions).
- **Acceptance**: `.moai/specs/SPEC-TOKEN-VERIFY-DIET-001/acceptance.md` (AC matrix + Given-When-Then scenarios).
- **Progress**: `.moai/specs/SPEC-TOKEN-VERIFY-DIET-001/progress.md` (§E skeleton + plan-phase audit-ready signal).
- **Epic context**: project memory `project_token_economy_epic_handoff.md` §C.
- **Preserved invariant**: `.claude/rules/moai/core/verification-claim-integrity.md` §1.1.
- **Sibling Epic SPECs**: SPEC-TOKEN-ACCOUNTING-001 (A, closed); SPEC-TOKEN-ROUTING-001 (B, closed); SPEC-TOKEN-BUDGET-STOP-001 (D, deferred).

---

## §I Open Design Decisions (run-phase discretion, NOT blockers)

These are design decisions the run-phase implementer (manager-develop) may tune within the contract's spirit. **None are blockers requiring orchestrator+user clarification** — they are within-contract tuning.

1. **Bounded-tail ceiling exact value**: Plan proposes `≤50 lines OR ≤2KB` default. Run-phase may tune per-domain (test output may need more lines than lint output). The contract holds regardless of the exact number; REQ-002 / REQ-003 bind only on the *exceeds-ceiling* predicate.
2. **File path convention**: Plan proposes `/tmp/moai-verify/<session-id>/<N>.<cmd-slug>.log` (per-session, cleared on session end). Run-phase may choose `.moai/state/verify/<session>/` instead (persisted, gitignored). The contract requires only that the path be (a) cited in the banner and (b) reachable at render time — it does not prescribe the directory.
3. **Localization of the file-path column header**: The column header (e.g., "Evidence" / "증거" / "証拠" / "证据") translates per `conversation_language`; file path values are locale-verbatim (moai.md line 734 already covers this category). Run-phase renders the header per the existing Localization Contract machinery.
4. **Whether to add a moai.md §8 column or a row-continuation field**: The Verification Matrix banner currently uses 2-column layout (lines 408-411). Adding a 3rd column may crowd the layout; a row-continuation field (`   └─ evidence: /tmp/...`) may be cleaner. Run-phase decides based on visual fit; either satisfies REQ-004.
