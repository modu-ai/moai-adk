---
id: SPEC-GLM-EFFORT-TUNE-001
title: "GLM effort overlay configuration tune-up (P1/P2/P4) — acceptance criteria"
version: "0.1.0"
status: completed
created: 2026-07-14
updated: 2026-07-14
author: manager-spec
priority: P2
phase: "v3.x config-tune"
module: "internal/template/glm_effort_overlay.go + .moai/config/sections/llm.yaml"
lifecycle: spec-anchored
tags: "glm, effort, overlay, config, template-mirror, reasoning-effort"
related_specs: [SPEC-MODEL-TIER-PLANTYPE-001]
---

# SPEC-GLM-EFFORT-TUNE-001 — acceptance.md

## §D. Acceptance Criteria Matrix

Each AC is Given-When-Then with observable evidence. Severity: **MUST** = must-pass (FAIL blocks implementation PASS); **SHOULD** = strongly expected (FAIL is debt, not blocker); **INFO** = documentation-only.

### P1 — builder-harness removed from the coding-max override set

#### AC-GET-001 [MUST] — override set contains exactly `manager-develop`

- **Given** the file `internal/template/glm_effort_overlay.go` after M1
- **When** the auditor runs `grep -n '"builder-harness": true' internal/template/glm_effort_overlay.go`
- **Then** the command returns no match (absence-grep), AND `grep -n '"manager-develop": true' internal/template/glm_effort_overlay.go` returns exactly one match.
- **Evidence**: verbatim grep output (both commands).
- **Maps to**: REQ-GET-001, REQ-GET-002.

#### AC-GET-002 [MUST] — manager-develop STILL forces reasoning-max

- **Given** the overlay after M1
- **When** a Go test calls `ResolveGLMReasoning("manager-develop", "medium")` (any Claude effort below max)
- **Then** the returned `GLMReasoningState.Name == "reasoning-max"` (the override wins over the collapse).
- **Evidence**: a passing Go test assertion. The M1 test edit MUST include this case.
- **Maps to**: REQ-GET-004.

#### AC-GET-003 [MUST] — builder-harness now reaches reasoning-high (make-or-break for P1)

- **Given** the overlay after M1 (builder-harness removed from the override set)
- **When** a Go test calls `ResolveGLMReasoning("builder-harness", "high")`
- **Then** the returned `GLMReasoningState.Name == "reasoning-high"` (NOT `reasoning-max` — the override no longer applies; the standard collapse of Claude `high` takes effect).
- **Evidence**: a passing Go test assertion (the new test added in M1).
- **Maps to**: REQ-GET-003.

#### AC-GET-004 [MUST] — test rename + cardinality assertion updated

- **Given** `internal/template/glm_effort_overlay_test.go` after M1
- **When** the auditor runs `grep -n 'TestGLMCodingMaxOverrideAgents_ExactlyTwo' internal/template/glm_effort_overlay_test.go`
- **Then** the command returns no match (the old name is gone), AND `grep -n 'len(got) == 1\|len(out) == 1\|want := \[\]string{"manager-develop"}' internal/template/glm_effort_overlay_test.go` returns at least one match (the new cardinality assertion is present).
- **Evidence**: verbatim grep output (both commands).
- **Maps to**: REQ-GET-005.

#### AC-GET-005 [MUST] — doc comments updated (no "two code-producing" framing)

- **Given** `internal/template/glm_effort_overlay.go` after M1
- **When** the auditor runs `grep -n 'two code-producing\|two code-producers\|{manager-develop, builder-harness}' internal/template/glm_effort_overlay.go`
- **Then** the command returns no match.
- **Evidence**: verbatim grep output.
- **Maps to**: REQ-GET-005 (comment half).

#### AC-GET-006 [SHOULD] — full package test suite passes

- **Given** the M1 changes
- **When** the auditor runs `go test ./internal/template/ -count=1`
- **Then** exit 0, no FAIL.
- **Evidence**: verbatim test output tail.
- **Maps to**: cross-cutting (supports REQ-GET-001..005).

### P2 — GLM reasoning-effort mapping exposed in llm.yaml

#### AC-GET-007 [MUST] — exposure block present in BOTH local and template mirror

- **Given** the M2 changes
- **When** the auditor runs `grep -c 'reasoning-effort mapping\|GLM reasoning' .moai/config/sections/llm.yaml internal/template/templates/.moai/config/sections/llm.yaml`
- **Then** BOTH files report ≥1 match (the exposure block exists in both the local config and the template source).
- **Evidence**: verbatim grep -c output (both filenames + counts).
- **Maps to**: REQ-GET-006, REQ-GET-009.

#### AC-GET-008 [MUST] — 3-state vocabulary present in the exposure block

- **Given** the M2 exposure block
- **When** the auditor runs `grep -c 'thinking-off' .moai/config/sections/llm.yaml internal/template/templates/.moai/config/sections/llm.yaml`
- **Then** BOTH files report ≥1 match (the 3rd state, thinking-off, is named — not just high/max).
- **Evidence**: verbatim grep -c output.
- **Maps to**: REQ-GET-006, REQ-GET-011.

#### AC-GET-009 [MUST] — overlay is documented as SSOT (no parallel runtime path)

- **Given** the M2 exposure block
- **When** the auditor runs `grep -i 'runtime SSOT\|documentation-only\|Go overlay' .moai/config/sections/llm.yaml internal/template/templates/.moai/config/sections/llm.yaml`
- **Then** at least one match in each file naming the Go overlay as the runtime SSOT / describing the YAML block as documentation-only.
- **Evidence**: verbatim grep output.
- **Maps to**: REQ-GET-007.

#### AC-GET-010 [MUST] — honesty caveat present (wire NOT claimed as validated)

- **Given** the M2 exposure block
- **When** the auditor runs `grep -i 'live validation pending\|live-validation pending\|implemented.*wired' .moai/config/sections/llm.yaml internal/template/templates/.moai/config/sections/llm.yaml`
- **Then** at least one match in each file carrying the honesty caveat, AND `grep -iE 'validated|guaranteed|works' .moai/config/sections/llm.yaml internal/template/templates/.moai/config/sections/llm.yaml` returns no match inside the exposure block (no overclaim wording).
- **Evidence**: verbatim grep output (both commands).
- **Maps to**: REQ-GET-008, constraint C-3.

#### AC-GET-011 [SHOULD] — mirror parity test passes (if it covers llm.yaml)

- **Given** the M2 changes
- **When** the auditor runs `go test ./internal/template/ -run TestRuleTemplateMirror -count=1`
- **Then** PASS (or, if `rule_template_mirror_test.go` does not cover `llm.yaml`, the absence is documented in research.md §D and this AC is INFO).
- **Evidence**: verbatim test output tail OR research.md §D absence note.
- **Maps to**: REQ-GET-009, constraint C-1.

#### AC-GET-012 [MUST] — no config-loader CI-guard regression

- **Given** the M2 changes (comments-only recommendation adopted — no `LLMConfig` struct field added)
- **When** the auditor runs `go test ./internal/config/... -count=1`
- **Then** PASS — `YAML_SECTION_NO_LOADER` and `CONFIG_STRUCT_YAML_MISMATCH` both clean (no new YAML key without a loader; no new struct field without a YAML key).
- **Evidence**: verbatim test output tail.
- **Maps to**: REQ-GET-007, constraint C-2.

### P4 — 3-state framing corrected in comments and docs

#### AC-GET-013 [MUST] — overlay doc comments frame 3 states, not 2-tier

- **Given** `internal/template/glm_effort_overlay.go` after M3
- **When** the auditor runs `grep -n '2-tier\|two-tier\|2 tier' internal/template/glm_effort_overlay.go`
- **Then** no match.
- **Evidence**: verbatim grep output (empty).
- **Maps to**: REQ-GET-010.

#### AC-GET-014 [MUST] — 3-state framing in the llm.yaml exposure block

- **Given** the M2 exposure block (also bound by P4)
- **When** the auditor runs `grep -c 'thinking-off' .moai/config/sections/llm.yaml internal/template/templates/.moai/config/sections/llm.yaml`
- **Then** BOTH ≥1 (already required by AC-GET-008; restated here as the P4 framing half).
- **Evidence**: verbatim grep -c output.
- **Maps to**: REQ-GET-011.

#### AC-GET-015 [MUST] — docs-site grep evidence recorded (presence OR absence)

- **Given** the M4 docs-site sweep
- **When** the auditor reviews research.md §D
- **Then** research.md §D records the verbatim grep command AND either (a) the list of corrected docs-site paths with before/after framing, OR (b) the empty grep result proving absence.
- **Evidence**: research.md §D content.
- **Maps to**: REQ-GET-012.

### Cross-cutting

#### AC-GET-016 [MUST] — full repo test suite passes

- **Given** all M1–M4 changes committed
- **When** the auditor runs `go test ./... -count=1`
- **Then** exit 0, no FAIL.
- **Evidence**: verbatim test output tail.
- **Maps to**: cross-cutting.

#### AC-GET-017 [MUST] — lint + vet clean (or pre-existing baseline)

- **Given** all M1–M4 changes committed
- **When** the auditor runs `go vet ./...` and `golangci-lint run --timeout=2m`
- **Then** vet clean; golangci-lint clean OR only pre-existing findings (no NEW finding introduced by this SPEC).
- **Evidence**: verbatim command output tail.
- **Maps to**: cross-cutting.

#### AC-GET-018 [MUST] — `make build` succeeds (template recompiled)

- **Given** the M2 template edit
- **When** the auditor runs `make build`
- **Then** exit 0.
- **Evidence**: verbatim make output tail.
- **Maps to**: constraint C-1.

## §D.1 Severity summary

| Severity | Count | IDs |
|----------|-------|-----|
| MUST | 14 | AC-GET-001, 002, 003, 004, 005, 007, 008, 009, 010, 012, 013, 014, 015, 016, 017, 018 |
| SHOULD | 2 | AC-GET-006, AC-GET-011 |
| INFO | 0 | — |

(Note: 16 MUST + 2 SHOULD = 18 ACs total; the table enumerates MUST IDs explicitly.)

## §D.2 Traceability

| REQ | AC |
|-----|----|
| REQ-GET-001 (set contains exactly manager-develop) | AC-GET-001 |
| REQ-GET-002 (builder-harness NOT in set) | AC-GET-001 |
| REQ-GET-003 (builder-harness → reasoning-high) | AC-GET-003 |
| REQ-GET-004 (manager-develop → reasoning-max preserved) | AC-GET-002 |
| REQ-GET-005 (test rename + comment update) | AC-GET-004, AC-GET-005 |
| REQ-GET-006 (exposure block in both files) | AC-GET-007, AC-GET-008 |
| REQ-GET-007 (overlay is SSOT, no parallel path) | AC-GET-009, AC-GET-012 |
| REQ-GET-008 (honesty caveat) | AC-GET-010 |
| REQ-GET-009 (mirror parity) | AC-GET-007, AC-GET-011 |
| REQ-GET-010 (3-state in code comments) | AC-GET-013 |
| REQ-GET-011 (3-state in llm.yaml exposure) | AC-GET-008, AC-GET-014 |
| REQ-GET-012 (docs-site grep + correction-or-absence) | AC-GET-015 |

## §D.3 Indirect verification

The `IsGLMCodingMaxOverrideAgent("builder-harness")` predicate is indirectly verified by AC-GET-001 (set membership) + AC-GET-003 (ResolveGLMReasoning behavioral outcome). No direct unit test on the predicate is required beyond the existing test (updated in M1) that enumerates set membership.

## §D.4 Closure gates (Definition of Done)

A SPEC PASS requires:

1. All 14 MUST ACs PASS with verbatim evidence (no summarize-only).
2. The 2 SHOULD ACs either PASS or carry an explicit debt note in progress.md §E.2.
3. `moai spec lint` returns 0 errors on this SPEC's spec.md (warnings acceptable; `StatusGitConsistency` is structural and expected at plan-phase).
4. The P3 Out-of-Scope bound holds: no live-validation code is added by this SPEC.
5. The honesty caveat (AC-GET-010) is present in BOTH llm.yaml surfaces — this is non-negotiable per C-3.

## §D.5 Forward-looking check (post-implementation)

After this SPEC lands, the MODEL-TIER-PLANTYPE-001 follow-up #1 (P3 — live wire-effectiveness validation) becomes the next actionable item on the GLM overlay subsystem. progress.md §E.4 (sync-phase) should record a pointer to that follow-up so it is not lost.

## §D.6 Anti-fabrication input validation

Every AC's "Evidence" field MUST be the verbatim output of a command actually run in the run-phase audit — NOT a paraphrase. An AC row marked PASS without verbatim grep/test output in the cited evidence path violates `verification-claim-integrity.md` §1.1 surface 2 (manager-agent self-verification).

## §D.7 Edge cases to cover in M1 tests

- `ResolveGLMReasoning("manager-develop", "low")` → `reasoning-max` (override wins even at lowest Claude effort)
- `ResolveGLMReasoning("manager-develop", "max")` → `reasoning-max` (override agrees with collapse)
- `ResolveGLMReasoning("builder-harness", "low")` → `thinking-off` (now follows standard collapse)
- `ResolveGLMReasoning("builder-harness", "high")` → `reasoning-high` (the make-or-break AC-GET-003 case)
- `ResolveGLMReasoning("builder-harness", "xhigh")` → `reasoning-max` (collapse of xhigh, NOT override)
- `ResolveGLMReasoning("manager-spec", "high")` → `reasoning-high` (unchanged, non-override agent)
