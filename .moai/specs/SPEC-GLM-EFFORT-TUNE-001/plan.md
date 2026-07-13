---
id: SPEC-GLM-EFFORT-TUNE-001
title: "GLM effort overlay configuration tune-up (P1/P2/P4) — implementation plan"
version: "0.1.0"
status: in-progress
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

# SPEC-GLM-EFFORT-TUNE-001 — plan.md

## §A. Context

This plan implements three surgical configuration-hygiene findings (P1/P2/P4) on the GLM effort overlay subsystem introduced by SPEC-MODEL-TIER-PLANTYPE-001. It does NOT change the overlay's architecture, the `tierProfiles` matrix, or any agent's Claude-side effort. The change set is small but spans Go code + Go tests + template mirror + local config + comments — hence Tier M (standard).

**Approach summary (one-line):** remove one key from a Go map, update the one test that pins the set's cardinality, add a documentation block to two YAML files (template first), and correct 2-tier framing to 3-state in code comments + the new YAML block + any docs-site references found by grep.

**Decision-reversal likelihood ordering** (per CLAUDE.md §7 Rule 1 — highest-change-likelihood decisions first, so review focuses there):

1. **P2 exposure mechanism (design.md §B)** — comments-only vs real `LLMConfig` struct field + loader wiring. This is the decision most likely to be reversed on review, because it trades runtime-extensibility against CI-guard complexity. Decided in M2; documented in design.md.
2. **P1 exclude-vs-include of builder-harness (design.md §A)** — the tradeoff record. The recommendation is EXCLUDE, but the INCLUDE rationale (harness does generate code) is recorded; a reviewer may flip the call.
3. **P2 exposure-block text + honesty-caveat wording** — the exact prose that lands in `llm.yaml` (both surfaces). Review-sensitive because of the verification-claim-integrity constraint.
4. **P4 docs-site grep scope** — which directories to sweep for 2-tier framing.
5. **P1 mechanical edits** — the actual map-key removal + test rename.
6. **P4 comment rewording** — mechanical prose edits in `glm_effort_overlay.go`.

## §B. Known Issues

- **KI-1 (test rename)**: the existing test is named `TestGLMCodingMaxOverrideAgents_ExactlyTwo` and explicitly asserts "exactly 2 members" (`glm_effort_overlay_test.go:116,119,121`). The rename is part of P1, not a side-effect — the test name and its assertion count both change. Forgetting the rename leaves a structurally-passing test with a misleading name.
- **KI-2 (comment-only "coding-max default" references)**: `internal/cli/glm.go:242` and `internal/cli/launcher.go:820` contain comment-level references to "hardcoded coding-max default". These describe `SessionGLMReasoningState()` (which unconditionally returns `glmReasoningMax` and is NOT affected by P1's set change), but a grep for "coding-max" will surface them. M1 must verify these are comment-only and explain in research.md why they do NOT change.
- **KI-3 (P4 docs-site absence vs presence)**: P4's REQ-GET-012 is an event-detected requirement — IF a docs-site reference frames GLM as 2-tier, correct it; ELSE record the absence. The M4 grep may return zero matches, which is a valid P4 outcome (the framing drift may be comments-only). M4 must NOT manufacture a docs-site edit to "fix" a non-existent reference.
- **KI-4 (template-mirror parity)**: the local `.moai/config/sections/llm.yaml` and the template `internal/template/templates/.moai/config/sections/llm.yaml` already diverge in indentation (local = 2-space; template = 2-space) and in the context_windows comment block (template has it, local has a shorter form). M2 must verify whether `rule_template_mirror_test.go` actually enforces parity for `llm.yaml` — if it does, both files must be edited to keep parity; if it does not, the edit is still made to both for consistency but no CI guard binds.

## §C. Pre-flight (verifiable preconditions)

1. `git -C /Users/goos/MoAI/moai-adk-go rev-parse --show-toplevel` → `/Users/goos/MoAI/moai-adk-go`
2. `grep -n 'builder-harness' /Users/goos/MoAI/moai-adk-go/internal/template/glm_effort_overlay.go` → exactly one match at line 109 (the override-set key)
3. `grep -n 'TestGLMCodingMaxOverrideAgents_ExactlyTwo' /Users/goos/MoAI/moai-adk-go/internal/template/glm_effort_overlay_test.go` → exactly one match at line 116
4. `grep -c 'reasoning_effort' /Users/goos/MoAI/moai-adk-go/.moai/config/sections/llm.yaml /Users/goos/MoAI/moai-adk-go/internal/template/templates/.moai/config/sections/llm.yaml` → both 0 (no exposure block today)

## §D. Constraints (carried from spec.md §D)

C-1 Template-First · C-2 no runtime behavior change for P2 · C-3 honesty caveat · C-4 P3 Out-of-Scope · C-5 case-sensitive agent names · C-6 no frontmatter/tierProfiles change.

Additional plan-level constraints:

- **C-7 no new dependency**: the SPEC introduces zero new Go imports, zero new YAML keys beyond the documentation block.
- **C-8 commit subject prefix `feat()`** for the plan-phase artifact commit (per `feedback_plan_commit_subject_feat_prefix`); subsequent run-phase commits follow the same `feat(SPEC-GLM-EFFORT-TUNE-001):` prefix.

## §E. Self-Verification (plan-phase, before plan-audit)

The manager-spec self-checks for plan-phase audit-readiness:

- [ ] SPEC ID regex PASS printed (decomposition trace committed)
- [ ] All 12 canonical frontmatter fields present in spec.md / plan.md / acceptance.md / progress.md
- [ ] Out of Scope section carries ≥1 `### Out of Scope — <topic>` H3 heading with `-` bullets
- [ ] GEARS notation used (Ubiquitous / When / While / Where / event-detected); no `IF/THEN` in NEW requirements
- [ ] Every code-line claim in spec.md §B cites a real line number read in this session (not memory)
- [ ] `acceptance.md` Given-When-Then scenarios reference observable evidence (grep output, Go test result, file existence)
- [ ] design.md records BOTH options for P1 (exclude vs include) and BOTH options for P2 (struct field vs comments-only)

## §F. Milestones

### M1 — P1 override set change (Go code + Go test) [Priority High]

The single most decision-relevant change: remove `builder-harness` from `glmCodingMaxOverrideAgents`, update the one test that pins cardinality, refresh the doc comments.

Files:
- `internal/template/glm_effort_overlay.go` — remove `"builder-harness": true,` line from the map; update the doc comments at lines 102-106 ("the two code-producing retained agents" → "the code-producing run-phase agent") and 118-120 (drop "exactly {manager-develop, builder-harness}"); leave `IsGLMCodingMaxOverrideAgent`, `GLMCodingMaxOverrideAgents`, `ResolveGLMReasoning` function signatures unchanged.
- `internal/template/glm_effort_overlay_test.go` — rename `TestGLMCodingMaxOverrideAgents_ExactlyTwo` → `TestGLMCodingMaxOverrideAgents_ExactlyOne` (or equivalent); change `want := []string{"builder-harness", "manager-develop"}` → `want := []string{"manager-develop"}`; change the `len(got) == 2` assertion → `len(got) == 1`; remove the `IsGLMCodingMaxOverrideAgent("builder-harness")` true-assertion; ADD a positive `ResolveGLMReasoning("builder-harness", "high").Name == "reasoning-high"` assertion (this is the make-or-break AC for P1's behavioral claim).

Verification:
- `go test ./internal/template/ -run TestGLMCodingMaxOverrideAgents -count=1` → PASS
- `go test ./internal/template/ -run 'TestResolveGLMReasoning|TestCollapse' -count=1` → PASS (no regression on the collapse table)
- `go test ./internal/template/ -count=1` → full package PASS
- `grep -n '"builder-harness": true' internal/template/glm_effort_overlay.go` → no match (absence-grep)

### M2 — P2 llm.yaml exposure block (template-first) [Priority High]

Add the documentation block under the existing `glm:` key in BOTH the template source and the local rendered file. design.md §B records the struct-field-vs-comments-only decision; **this plan recommends comments-only** for v0.1.0 (the exposure is documentation; introducing a real `LLMConfig.GLMReasoning*` struct field + loader wiring is out of proportion to a documentation goal, and would bind the `YAML_SECTION_NO_LOADER` + `CONFIG_STRUCT_YAML_MISMATCH` CI guards for no runtime benefit). A real struct field can be added later as a user-override extension point without re-designing the exposure.

Files:
- `internal/template/templates/.moai/config/sections/llm.yaml` (Template-First) — add a `# GLM reasoning-effort mapping (documentation-only — the Go overlay is the runtime SSOT)` block under `glm:`. Block contents: the 3 reachable states, the 5→3 collapse table, the coding-max-override note (now naming only `manager-develop`), the manager-git→thinking-off note, and the honesty caveat ("implemented + wired; live validation pending").
- `.moai/config/sections/llm.yaml` (local rendered) — mirror the block.
- `make build` — recompile to embed the updated template.

Verification:
- `grep -c 'reasoning-effort mapping' .moai/config/sections/llm.yaml internal/template/templates/.moai/config/sections/llm.yaml` → both ≥1
- `grep -c 'thinking-off' .moai/config/sections/llm.yaml internal/template/templates/.moai/config/sections/llm.yaml` → both ≥1
- `grep -c 'live validation pending\|live-validation pending' .moai/config/sections/llm.yaml internal/template/templates/.moai/config/sections/llm.yaml` → both ≥1 (honesty caveat present)
- `go test ./internal/template/ -run TestRuleTemplateMirror -count=1` → PASS (if the parity test covers llm.yaml; if not, document the absence in research.md §D)
- `go test ./internal/config/... -count=1` → PASS (no struct/loader change → no CI-guard violation)

### M3 — P4 3-state framing in code comments [Priority Medium]

Correct the framing in `glm_effort_overlay.go` package-level + function-level doc comments. The named constants at lines 26-33 are already correct (3 states); the drift is in the prose framing the system as 2-tier in places.

Files:
- `internal/template/glm_effort_overlay.go` — package doc comment (lines 1-20) already mentions "3-state reasoning control" — verify and tighten; `glmCodingMaxOverrideAgents` doc (lines 102-106) — already targeted by M1; `CollapseClaudeEffortToGLM` doc (lines 74-87) — verify the 5→3 table is framed as 3 states; `SessionGLMReasoningState` doc (lines 162-171) — verify.

Verification:
- `grep -n '2-tier\|two-tier\|2 tier' internal/template/glm_effort_overlay.go` → no match (absence-grep for the drift framing)
- `grep -c 'thinking-off\|reasoning-high\|reasoning-max' internal/template/glm_effort_overlay.go` ≥ existing baseline (no regression in 3-state vocabulary)

### M4 — P4 docs-site grep + correction (if found) [Priority Medium]

Sweep the repo for 2-tier framing of GLM in Markdown files. Per KI-3, the absence of a match is a valid outcome.

Files (conditional):
- `grep -rln 'GLM.*2-tier\|GLM.*two-tier\|glm.*2 tier\|reasoning_effort.*2.level\|high/max.*GLM\|GLM.*high/max' README.md README.ko.md README.ja.md README.zh.md docs/ .claude/ .moai/docs/` → record the matches (if any) in research.md §D
- IF matches found: correct each to 3-state framing in the same commit; document each correction site.
- IF zero matches: record the absence-grep evidence (the exact command + empty result); no docs-site edit is made.

Verification:
- IF corrections made: `grep -rln '2-tier\|two-tier' <corrected-paths>` → no match (absence-grep after correction)
- IF no corrections: research.md §D records the verbatim empty grep output.

### M5 — full-suite verification + commit [Priority High]

- `make build` → exit 0
- `go test ./internal/template/ -count=1` → PASS
- `go test ./internal/config/... -count=1` → PASS
- `go test ./... -count=1` → PASS (catch cascading failures per CLAUDE.local.md §6)
- `go vet ./...` → clean
- `golangci-lint run --timeout=2m` → clean (or pre-existing-baseline only)
- Commit (plan-phase + run-phase commit boundary per `feedback_plan_commit_subject_feat_prefix`): `feat(SPEC-GLM-EFFORT-TUNE-001): M1-M4 GLM effort overlay tune-up (P1/P2/P4)`

## §G. Anti-Patterns

- **AP-1 manufacturing a docs-site edit when the grep is empty** (KI-3). P4's REQ-GET-012 is event-detected; zero matches is a valid PASS.
- **AP-2 editing `tierProfiles` or agent frontmatter while doing P1** (C-6). The P1 change is to the GLM-side override set ONLY. Any temptation to "also lower builder-harness's Claude effort" is scope creep.
- **AP-3 introducing a parallel runtime path for the llm.yaml exposure** (C-2). The YAML block is documentation; making it a real config struct field without going through design.md §B's struct-field path is a CI-guard violation waiting to happen.
- **AP-4 overclaiming the wire** (C-3). Words like "validated", "guaranteed", "works" in the exposure block violate the honesty caveat.
- **AP-5 ignoring the test rename** (KI-1). A test named `_ExactlyTwo` that asserts `_ExactlyOne` is a future-confusion trap.
- **AP-6 carry-over from MODEL-TIER-PLANTYPE-001 memory** (`feedback_defect_claim_verification`). Every line-number claim in research.md must be re-verified in this session against the current file state.

## §H. Cross-References

- spec.md §B (background) · §C (REQ-GET-001..012) · §D (constraints) · §E (Out of Scope, including P3)
- acceptance.md §D (AC matrix mapping REQ → observable evidence)
- design.md §A (P1 exclude-vs-include tradeoff) · §B (P2 struct-field-vs-comments-only decision)
- research.md §A/§B/§C/§D (line-number-cited source evidence)
- `feedback_plan_commit_subject_feat_prefix` (commit subject prefix)
- `feedback_defect_claim_verification` (line-number citation discipline)
- `verification-claim-integrity.md` §1.1 surface 3 (honesty caveat)
