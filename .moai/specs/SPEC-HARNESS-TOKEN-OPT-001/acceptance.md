# SPEC-HARNESS-TOKEN-OPT-001 — Acceptance Criteria

> Verification layer. Given-When-Then format. Mechanically testable. Cross-references `spec.md` REQ-HTO-001..010 and `plan.md` M0-M6.

## §D. AC Matrix

### AC-HTO-001 — verification-batch-pattern.md paths: frontmatter present (M0, REQ-HTO-001)

**Given** the file `.claude/rules/moai/workflow/verification-batch-pattern.md` is read after M0 lands,
**When** the first 5 lines are inspected,
**Then** a YAML frontmatter block is present containing a `paths:` key whose value globs include `.moai/specs/**` and at least one run/sync workflow skill path.

**Mechanical verification**: `head -5 .claude/rules/moai/workflow/verification-batch-pattern.md | grep -c "^paths:"` = 1.

### AC-HTO-002 — verification-batch-pattern.md A9 thin pointer (M0, REQ-HTO-001)

**Given** the file after M0 lands,
**When** the A9 attributable diff-check section is inspected,
**Then** the section is ≤ 8 lines and references `agent-common-protocol.md` §Parallel Execution → Attributable diff-check doctrinal switch as the SSOT.

**Mechanical verification**: `awk '/## Attributable diff-check pattern/,/^## /' .claude/rules/moai/workflow/verification-batch-pattern.md | wc -l` ≤ 12 (heading + ≤8 body + blank).

### AC-HTO-003 — nav-tokens.md paths: frontmatter present (M0, REQ-HTO-002)

**Given** the file `.claude/rules/moai/workflow/nav-tokens.md` is read after M0 lands,
**When** the first 5 lines are inspected,
**Then** a YAML frontmatter block is present containing a `paths:` key whose value globs include `.moai/project/*.md` and `**/*.go`.

**Mechanical verification**: `head -5 .claude/rules/moai/workflow/nav-tokens.md | grep -c "^paths:"` = 1.

### AC-HTO-004 — goal-directive.md stub size reduced (M1, REQ-HTO-003)

**Given** the file `.claude/rules/moai/workflow/goal-directive.md` after M1 lands,
**When** byte-size is measured,
**Then** the size is ≤ 12,000 bytes (reduced from baseline 25,755 bytes; target ~8K-10K).

**Mechanical verification**: `wc -c .claude/rules/moai/workflow/goal-directive.md` reports a value ≤ 12000.

### AC-HTO-005 — goal-directive-detail.md lazy companion created with paths: (M1, REQ-HTO-003)

**Given** the file `.claude/rules/moai/workflow/goal-directive-detail.md` after M1 lands,
**When** the file is inspected,
**Then** it exists, has YAML frontmatter with `paths:` scoped to `.moai/state/goal/**` and the goal workflow skill, and contains the relocated sections (Comparing Approaches table, condition templates, Integration Notes, Native /goal Prohibition).

**Mechanical verification**:
- `test -f .claude/rules/moai/workflow/goal-directive-detail.md && echo PASS`.
- `head -5 .claude/rules/moai/workflow/goal-directive-detail.md | grep -c "^paths:"` = 1.
- `grep -c "Native .* Prohibition\|Comparing Autonomous-Continuation" .claude/rules/moai/workflow/goal-directive-detail.md` ≥ 2.

### AC-HTO-006 — goal-directive.md Goal-Presentation Timing arm-only invariant preserved (M1, REQ-HTO-003 + REQ-HTO-010)

**Given** the always-loaded goal-directive.md stub after M1 lands,
**When** the file is grep'd for the arm-only invariant,
**Then** the Goal-Presentation Timing section is preserved (the load-bearing invariant that `/moai goal` never starts work of its own and never substitutes for Implementation Kickoff Approval).

**Mechanical verification**: `grep -c "arm-only\|Goal-Presentation Timing" .claude/rules/moai/workflow/goal-directive.md` ≥ 1.

### AC-HTO-007 — session-handoff.md Diet Constraints relocated (M2, REQ-HTO-004)

**Given** the files `.claude/rules/moai/workflow/session-handoff.md` and `.claude/rules/moai/workflow/session-handoff-examples.md` after M2 lands,
**When** both are inspected,
**Then** the full AP-D-001..005 catalogue and the 9-item pre-emit checklist live in `session-handoff-examples.md` (the lazy sidecar) and the always-loaded `session-handoff.md` retains a 2-example inline summary + 1-line pointer.

**Mechanical verification**:
- `grep -c "AP-D-001\|AP-D-005" .claude/rules/moai/workflow/session-handoff-examples.md` ≥ 2.
- `grep -c "session-handoff-examples.md" .claude/rules/moai/workflow/session-handoff.md` ≥ 1 (the pointer).

### AC-HTO-008 — session-handoff.md Canonical Format + Cut-line Marker preserved (M2, REQ-HTO-004 + REQ-HTO-010)

**Given** the always-loaded `session-handoff.md` after M2 lands,
**When** grep'd for the paste-ready resume contract,
**Then** the 6-block Canonical Format and the Cut-line Marker Spec are preserved verbatim.

**Mechanical verification**:
- `grep -c "✂──── 여기부터 복사" .claude/rules/moai/workflow/session-handoff.md` = 1.
- `grep -c "✂──── 여기까지 복사" .claude/rules/moai/workflow/session-handoff.md` = 1.
- `grep -c "Block 1\|Block 5" .claude/rules/moai/workflow/session-handoff.md` ≥ 2.

### AC-HTO-009 — IK SSOT designated + cross-reference pointer landed (M3, REQ-HTO-005)

**Given** the corpus `.claude/rules/moai/` + `CLAUDE.md` after M3 lands,
**When** the Implementation Kickoff Approval restatements are enumerated,
**Then** (a) `orchestration-mode-selection.md` §E retains the canonical mandate (mandatory, score-independent, ordering-invariant); (b) the count of total restatements is reduced from the 45-match baseline; (c) the 1-line cross-reference `Per the Implementation Kickoff Approval mandatory-restoration invariant (orchestration-mode-selection.md §E).` appears in ≥ 8 cut sites.

**Mechanical verification**:
- `grep -c "mandatory and score-independent\|mandatory, score-independent" .claude/rules/moai/workflow/orchestration-mode-selection.md` ≥ 1 (mandate preserved).
- `grep -rn "Implementation Kickoff Approval mandatory-restoration invariant (orchestration-mode-selection.md §E)" .claude/rules/moai/ CLAUDE.md | wc -l` ≥ 1 (cross-reference pointer landed; ambitious ≥8 target achievable only if run-phase re-classification grows the C-class set beyond the 1 clear-cut identified at plan-phase — see plan.md §F.M3 default-to-preserve policy).
- `grep -rn "Implementation Kickoff Approval" .claude/rules/moai/ CLAUDE.md | wc -l` ≤ 45 (no-regression floor; measured baseline 45 across 12 files, 2026-08-11). The `< 25` ambitious target stays as the reduction ceiling, achievable only if run-phase re-classification grows the C-class set beyond the 1 clear-cut identified at plan-phase.

### AC-HTO-010 — IK mandate itself preserved (M3, REQ-HTO-005 + REQ-HTO-010 do_not_touch)

**Given** the corpus after M3 lands,
**When** the canonical SSOT is inspected,
**Then** the mandate (mandatory, score-independent, ordering-invariant — gate before any autonomy) is preserved verbatim in `orchestration-mode-selection.md` §E, and at least one canonical-quality restatement per file whose logic depends on the gate is preserved (class A + class B).

**Mechanical verification**:
- `grep -B1 -A1 "Implementation Kickoff Approval" .claude/rules/moai/workflow/orchestration-mode-selection.md | grep -c "score-independent"` ≥ 1.
- `grep -c "plan→run HUMAN GATE\|plan-to-implement HUMAN GATE" .claude/rules/moai/workflow/orchestration-mode-selection.md` ≥ 1.

### AC-HTO-011 — A9 default inverted to consume-path (M4, REQ-HTO-006)

**Given** `.claude/rules/moai/core/agent-common-protocol.md` after M4 lands,
**When** the §Parallel Execution → Attributable diff-check doctrinal switch is inspected,
**Then** the default is CONSUME-§E-evidence-on-three-way-match (not re-execute), and the fallback-to-re-execution contract is preserved unchanged on any mismatch.

**Mechanical verification**:
- `grep -c "CONSUMES the attributable §E evidence\|consume.*§E evidence.*default" .claude/rules/moai/core/agent-common-protocol.md` ≥ 1.
- `grep -c "fallback to re-execution\|fallback-to-re-execution" .claude/rules/moai/core/agent-common-protocol.md` ≥ 3.
- `grep -c "snapshot_key_drift\|command_drift\|missing_section_e\|output_drift" .claude/rules/moai/core/agent-common-protocol.md` ≥ 4 (mismatch-reason enum preserved).
- `grep -c "VCI §1.1 invariant holds on every path" .claude/rules/moai/core/agent-common-protocol.md` ≥ 1.

### AC-HTO-012 — CLAUDE.local.md consolidated (M5, REQ-HTO-007)

**Given** `CLAUDE.local.md` after M5 lands,
**When** byte-size and structure are inspected,
**Then** (a) byte-size ≤ 32,000 bytes (from baseline 42,164); (b) a single `## References` section exists; (c) §5 Version Management and §7 Hook Development content moved to `.moai/docs/` files.

**Mechanical verification**:
- `wc -c CLAUDE.local.md` reports value ≤ 32000.
- `grep -c "^## References$" CLAUDE.local.md` = 1.
- `test -f .moai/docs/version-management.md && test -f .moai/docs/hook-development.md && echo PASS`.
- `grep -c "version-management.md\|hook-development.md" CLAUDE.local.md` ≥ 2 (the pointers).

### AC-HTO-013 — Template mirror parity (M6, REQ-HTO-008)

**Given** the template source `internal/template/templates/.claude/rules/moai/` after M6 lands,
**When** each mirrored target file is diff'd against its local counterpart,
**Then** the diff is empty (byte-identical parity) for: verification-batch-pattern.md, nav-tokens.md, goal-directive.md, goal-directive-detail.md, session-handoff.md, session-handoff-examples.md, agent-common-protocol.md, orchestration-mode-selection.md, and every other file touched in M3.

**Mechanical verification** (one diff per file; exit 0 expected):
- `diff .claude/rules/moai/workflow/verification-batch-pattern.md internal/template/templates/.claude/rules/moai/workflow/verification-batch-pattern.md`.
- `diff .claude/rules/moai/workflow/nav-tokens.md internal/template/templates/.claude/rules/moai/workflow/nav-tokens.md`.
- `diff .claude/rules/moai/workflow/goal-directive.md internal/template/templates/.claude/rules/moai/workflow/goal-directive.md`.
- `diff .claude/rules/moai/workflow/goal-directive-detail.md internal/template/templates/.claude/rules/moai/workflow/goal-directive-detail.md`.
- `diff .claude/rules/moai/workflow/session-handoff.md internal/template/templates/.claude/rules/moai/workflow/session-handoff.md`.
- `diff .claude/rules/moai/workflow/session-handoff-examples.md internal/template/templates/.claude/rules/moai/workflow/session-handoff-examples.md`.
- `diff .claude/rules/moai/core/agent-common-protocol.md internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md`.
- `diff .claude/rules/moai/workflow/orchestration-mode-selection.md internal/template/templates/.claude/rules/moai/workflow/orchestration-mode-selection.md`.
- (plus every additional file touched in M3).

### AC-HTO-014 — make build succeeds + catalog.yaml regenerated (M6, REQ-HTO-008)

**Given** the template source after mirror,
**When** `make build` is run,
**Then** the build exits 0 and `internal/template/catalog.yaml` is regenerated (and committed).

**Mechanical verification**:
- `make build` exit 0.
- `git status --porcelain internal/template/catalog.yaml` shows the file is staged or committed (not stale).

### AC-HTO-015 — Template neutrality CI guard passes (M6, REQ-HTO-009)

**Given** the template source after mirror + build,
**When** the §25 template-neutrality CI guard runs,
**Then** zero forbidden content classes are detected: no SPEC-ID (`SPEC-HARNESS-TOKEN-OPT-001`), no REQ-token (`REQ-HTO-`), no audit citation, no internal work date beyond routine doc dates, no commit SHA, no macOS-bias absolute path.

**Mechanical verification**:
- `go test ./internal/template/...` exit 0 (covers `internal_content_leak_test.go` + `split_namespace_test.go`).
- `grep -rn "SPEC-HARNESS-TOKEN-OPT-001\|REQ-HTO-" internal/template/templates/` returns zero matches.
- `grep -rn "08eef9a0f" internal/template/templates/` returns zero matches (no commit SHA leakage).

### AC-HTO-016 — verification-claim-integrity.md §1-§3 preserved (REQ-HTO-010 do_not_touch)

**Given** `.claude/rules/moai/core/verification-claim-integrity.md` (NOT edited by this SPEC),
**When** grep'd for the no-unobserved-claim invariant,
**Then** §1 (all four binding surfaces), §2 (baseline-attribution), §3 (5-section evidence report format) are preserved verbatim.

**Mechanical verification**:
- `grep -c "no unobserved-claim\|no-unobserved-claim" .claude/rules/moai/core/verification-claim-integrity.md` ≥ 1.
- `grep -c "Baseline-Integrity Attribution\|baseline-attribution" .claude/rules/moai/core/verification-claim-integrity.md` ≥ 1.
- `grep -c "5-Section Evidence-Bearing Report\|Claim.*Evidence.*Baseline-attribution.*Gaps.*Residual-risk" .claude/rules/moai/core/verification-claim-integrity.md` ≥ 1.
- `diff` against `origin/main -- .claude/rules/moai/core/verification-claim-integrity.md` shows ZERO changes (this file is not edited by this SPEC).

### AC-HTO-017 — AskUserQuestion channel monopoly + preload mandate preserved (REQ-HTO-010 do_not_touch)

**Given** `.claude/rules/moai/core/askuser-protocol.md` and `.claude/rules/moai/core/moai-constitution.md` after M3 lands (these files may have IK restatements cut),
**When** grep'd for the channel monopoly + preload mandate,
**Then** both are preserved verbatim.

**Mechanical verification**:
- `grep -c "AskUserQuestion is the.*exclusive.*channel\|only user-facing question channel" .claude/rules/moai/core/askuser-protocol.md` ≥ 1.
- `grep -c 'ToolSearch(query: "select:AskUserQuestion")' .claude/rules/moai/core/askuser-protocol.md` ≥ 1.

### AC-HTO-018 — manager-develop §E E1-E8 triple + sync-auditor weights preserved (REQ-HTO-010 do_not_touch)

**Given** the manager-develop prompt template and the sync-auditor scoring rules (NOT edited by this SPEC, except the A9 default inversion in `agent-common-protocol.md` which is orthogonal),
**When** grep'd for the §E triple and the 4-dimension weights,
**Then** both are preserved verbatim.

**Mechanical verification**:
- `grep -c "verbatim.*observed\|attributable" .claude/rules/moai/development/manager-develop-prompt-template.md` ≥ 2 (returns 2 today, 2026-08-11; the canonical text uses "exact invocation" / "observed output" / "baseline-attribution" on separate bold-labeled lines, NOT "verbatim command ... observed output" on one line — the broader disjunction resolves against the canonical phrasing).
- `grep -c "Attribution discipline (SPEC-SYNC-PARALLEL-DOCS-001 A9)" .claude/rules/moai/development/manager-develop-prompt-template.md` ≥ 1 (the §E attribution-discipline section header).
- `grep -c "Functionality.*40\|Functionality 40" .claude/agents/moai/sync-auditor.md` ≥ 1 (the sync-auditor weights SSOT — verified 2026-08-11: returns 2; the weights do NOT live under `.claude/rules/moai/`, they live in the agent definition + evaluator-profiles).
- `grep -c "Functionality.*40\|Functionality 40" .moai/config/evaluator-profiles/default.md` ≥ 1 (the derived profile; verified 2026-08-11: returns 2).

## §D.1 Severity Classification

| AC | Severity | Rationale |
|---|---|---|
| AC-HTO-013, AC-HTO-015, AC-HTO-016, AC-HTO-017, AC-HTO-018 | MUST-PASS | Parity, neutrality, and do_not_touch — safety-critical |
| AC-HTO-006, AC-HTO-008, AC-HTO-010, AC-HTO-011 | MUST-PASS | Load-bearing invariants preserved |
| AC-HTO-001..005, AC-HTO-007, AC-HTO-009, AC-HTO-012, AC-HTO-014 | MUST-PASS | Functional delivery of each recommendation |

All 18 ACs are MUST-PASS (this SPEC carries no nice-to-have criteria; every recommendation is user-approved P0+P1).

## §D.2 Traceability Matrix

| REQ | Milestone | ACs |
|---|---|---|
| REQ-HTO-001 (verification-batch paths + thin pointer) | M0 | AC-HTO-001, AC-HTO-002 |
| REQ-HTO-002 (nav-tokens paths) | M0 | AC-HTO-003 |
| REQ-HTO-003 (goal-directive split) | M1 | AC-HTO-004, AC-HTO-005, AC-HTO-006 |
| REQ-HTO-004 (session-handoff Diet lazy) | M2 | AC-HTO-007, AC-HTO-008 |
| REQ-HTO-005 (IK SSOT consolidation) | M3 | AC-HTO-009, AC-HTO-010 |
| REQ-HTO-006 (A9 default inversion) | M4 | AC-HTO-011 |
| REQ-HTO-007 (CLAUDE.local.md consolidation) | M5 | AC-HTO-012 |
| REQ-HTO-008 (Template mirror + make build) | M6 | AC-HTO-013, AC-HTO-014 |
| REQ-HTO-009 (Template neutrality) | M6 | AC-HTO-015 |
| REQ-HTO-010 (do_not_touch preservation) | cross-cutting | AC-HTO-006, AC-HTO-008, AC-HTO-010, AC-HTO-016, AC-HTO-017, AC-HTO-018 |

## §D.3 Indirect Verification (goals the ACs imply but do not directly test)

- **Token recovery ≥ 18,000 tokens/turn**: implied by AC-HTO-004 (goal-directive stub size ≤ 12K from 25.7K baseline = ~13.5K bytes saved) + AC-HTO-001/003 (verification-batch + nav-tokens become lazy, full file size saved when not loaded) + AC-HTO-007 (session-handoff Diet relocated) + AC-HTO-009 (IK restatements reduced). Direct measurement requires a token-counter pass over the always-loaded set; deferred to sync-phase.
- **Wall-clock recovery 30-120s per run-phase completion**: implied by AC-HTO-011 (A9 default = consume-path). Direct measurement requires orchestrator self-timing instrumentation; deferred to sync-phase or a follow-up SPEC.

## §D.4 Closure Gates (Definition of Done)

- [ ] All 18 ACs PASS (mechanically verified, not asserted).
- [ ] `make build` succeeds (AC-HTO-014).
- [ ] `go test ./internal/template/...` succeeds (AC-HTO-015).
- [ ] Template-source diff parity for every mirrored file (AC-HTO-013).
- [ ] do_not_touch grep sentinels all PASS (AC-HTO-006, AC-HTO-008, AC-HTO-010, AC-HTO-016, AC-HTO-017, AC-HTO-018).
- [x] `[NEEDS CLARIFICATION: IK restatement classification]` in plan.md — RESOLVED 2026-08-11 (user confirmed "보존 우선 / default-to-preserve" via AskUserQuestion; full A/B/C table in plan.md §F.M3: A=3 canonical, B=41 preserve, C=1 cut). Zero unresolved markers remain in plan.md.

## §D.5 Forward-Looking Checks (post-merge, not blocking)

- Run-phase wall-clock delta on a real SPEC completion (compare before/after A9 default inversion).
- Always-loaded token-count delta (instrument the orchestrator's session-start prefix).
- sync-auditor independence: confirm the consolidation did not introduce a self-audit path (no editor of an IK restatement is also the auditor of that restatement).

## §D.6 Edge Cases

- **Edge — `goal-directive-detail.md` `paths:` matches a session that armed a goal via `moai goal arm` but did not touch `.moai/state/goal/**` directly** (the goal state file is written by the CLI, not by the orchestrator's Edit tool). The `paths:` glob should include `.claude/skills/moai/workflows/goal.md` (the goal workflow skill), which IS touched when the orchestrator loads the goal workflow. Mitigation: AC-HTO-005 verifies both globs are present.
- **Edge — M3 cuts a restatement that is later needed for a file's internal logic**. R2 mitigation: classify-before-cut; default-to-preserve; plan-auditor reviews classification.
- **Edge — Template mirror introduces a SPEC-ID leak via a `Related SPECs` cross-reference in a rule body**. Mitigation: AC-HTO-015 grep sentinel `grep -rn "SPEC-HARNESS-TOKEN-OPT-001" internal/template/templates/` returns zero.
- **Edge — `verification-batch-pattern.md` `paths:` restriction prevents it from loading when the orchestrator needs the WHY rationale at run-phase**. Mitigation: the HARD batching obligation lives in always-loaded `agent-common-protocol.md` §Parallel Execution; the WHY in `verification-batch-pattern.md` is complementary, not load-bearing. The `paths:` glob includes `.moai/specs/**` and run/sync workflow skills, both of which are touched at run-phase completion.
