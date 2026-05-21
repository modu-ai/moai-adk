---
id: SPEC-V3R5-WORKFLOW-OPT-001
title: "Workflow Optimization — Implementation Plan"
version: "0.1.0"
status: draft
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P1
phase: "v3.0.0 — Round 5"
module: ".claude/rules/moai + .moai/config/sections/workflow.yaml + internal/harness/capture + .claude/agents/moai/plan-auditor.md"
lifecycle: spec-anchored
tags: "workflow, optimization, plan, milestones, brownfield, agent-teams, plan-auditor"
---

# Implementation Plan — SPEC-V3R5-WORKFLOW-OPT-001

## 1. Strategy Overview

This SPEC is a **brownfield meta-workflow extension**: it modifies orchestrator behaviour and tooling without altering any user-facing CLI surface. The implementation strategy is **add-only**:

- Rules and configs are EXTENDED (no deletions).
- The Go code addition (Layer F) places NEW files under `internal/harness/capture/` alongside W3-shipped files; no existing function or type is renamed or removed.
- The agent prompt (Layer G) appends D7 and D8 dimensions without modifying D1–D6.

The 8 layers map to 6 milestones (M1–M6) grouped by artifact-type, risk, and dependency:

| Milestone | Domain      | Layers covered      | Risk    | Depends on |
|-----------|-------------|---------------------|---------|------------|
| M1        | R (rule)    | A (remainder), C, D, E, H | Low     | —          |
| M2        | C (config)  | B                   | Medium  | M1         |
| M3        | F (Go code) | F                   | High    | M1, W3 PR #1024 base |
| M4        | G (agent)   | G                   | Medium  | M1         |
| M5        | Integration | (dogfooding)        | Low     | M1–M4      |
| M6        | Docs        | (cleanup)           | Low     | M5         |

Milestones M2, M3, and M4 are **independent of each other** — they can run in parallel once M1 is complete (this is the very pattern Layer B / REQ-WO-010 enables; if Agent Teams config lands in M2 before M3/M4, then M3/M4 can validate the pattern by being implemented in parallel teammates).

---

## 2. Milestones

### M1 — Rule Domain (Layers A-remainder, C, D, E, H)

**Goal**: Land all rule-only edits in a single milestone; ensure template mirror parity.

**Tasks**:

1. **A.1** Mirror `.claude/rules/moai/development/manager-develop-prompt-template.md` → `internal/template/templates/.claude/rules/moai/development/manager-develop-prompt-template.md` (file copy).
2. **A.2** Add cross-reference in `.claude/skills/moai/workflows/run.md` § Phase 1 pointing to the new prompt template.
3. **C.1** Extend `.claude/rules/moai/workflow/ci-watch-protocol.md` with a "Background watch standardization" section: canonical pattern is `gh pr checks --watch` invoked via `run_in_background: true`; explicit anti-pattern callout for `sleep + check` polling loops.
4. **D.1** Add "Read-only verification batching" subsection to `.claude/rules/moai/core/agent-common-protocol.md` § Parallel Execution, with a 7-item canonical example (test / coverage / grep / sentinel / CLI / benchmark / lint).
5. **D.2** Create new rule `.claude/rules/moai/workflow/verification-batch-pattern.md` formalizing the verification grouping pattern.
6. **E.1** Modify `.claude/rules/moai/workflow/spec-workflow.md` § Phase Transitions to permit run-phase start while plan-PR is in review/CI, and to define the Plan Audit Gate skip policy (PASS ≥ 0.90 threshold).
7. **H.1** Add "Tool optimization patterns" subsection to `.claude/rules/moai/core/agent-common-protocol.md`: canonical patterns for `gh pr checks --json … | jq`, ToolSearch per-turn preload, and `git log --format=…` single-command idioms.
8. **M1.X** Mirror every rule edit to `internal/template/templates/.claude/rules/moai/**`; run `make build`; verify `git diff internal/template/embedded.go` is non-empty (mirror invariant).
9. **M1.Y** Add CI guard test `internal/template/rule_template_mirror_test.go` (or extend existing template audit test) that fails with `RuleTemplateMirrorDrift` when any `.claude/rules/moai/**` file lacks a mirror under `internal/template/templates/**`.

**Deliverables**: 4 modified rule files (`manager-develop-prompt-template.md`, `ci-watch-protocol.md`, `agent-common-protocol.md`, `spec-workflow.md`), 2 new rule files (`agent-teams-pattern.md` placeholder — full content in M2, `verification-batch-pattern.md`), template mirrors, CI guard test, `make build` artifact.

**Exit criteria**: AC-WO-002, AC-WO-003, AC-WO-006, AC-WO-007, AC-WO-008, AC-WO-013 (partial; full verification in M5).

---

### M2 — Config Domain (Layer B — Agent Teams)

**Goal**: Land `workflow.yaml` extension for Agent Teams role profiles; document experimental flag.

**Tasks**:

1. **B.1** Extend `.moai/config/sections/workflow.yaml` under existing `team:` key to add:
   - `team.role_profile_keys: [implementer, tester, reviewer]` (enumeration of allowed keys; PRESERVE `auto_selection`, `default_model`, `delegate_mode`, `enabled`).
   - `team.role_profiles:` map with 7 entries (5 implementer role_profiles + 1 tester + 1 reviewer). Each entry: `mode`, `model`, `isolation`.
2. **B.2** Populate new `.claude/rules/moai/workflow/agent-teams-pattern.md` (placeholder from M1) with the standardized 5-teammate pattern: when to spawn, role distribution, file ownership, fallback to solo mode.
3. **B.3** Document `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS=1` environment variable requirement; add to `.moai/config/sections/workflow.yaml` comment header.
4. **B.4** Mirror `workflow.yaml` change to `internal/template/templates/.moai/config/sections/workflow.yaml`; `make build`.
5. **B.5** Add `internal/config/workflow_role_profiles_test.go` validating that:
   - `role_profile_keys` contains exactly the 3 allowed values.
   - `role_profiles` map has exactly 7 entries.
   - Every entry has the required 3 sub-keys (`mode`, `model`, `isolation`).

**Deliverables**: Modified `workflow.yaml` (runtime + template), populated `agent-teams-pattern.md`, config test.

**Exit criteria**: AC-WO-009.

**Risk mitigation**: Agent Teams is an experimental Claude Code feature. The added config is gated by `team.enabled` (defaults to existing value) and `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS`. Solo-mode fallback continues to work because no existing key is removed.

---

### M3 — Code Domain (Layer F — Defect Detector Extension)

**Goal**: Extend `internal/harness/capture/` package with defect pattern detection that classifies `manager-develop` failure reports into the 8 known-issue categories (B1–B8).

**Dependency**: M3 starts after W3 PR #1024 merges (the W3 plan creates `internal/harness/capture/` package). If W3 has not merged at M3 start, M3 starts on the W3 feature branch and rebases when W3 merges.

**Tasks**:

1. **F.1** Add `internal/harness/capture/defect_detector.go`:
   - Type `DefectCategory` enum (`B1Crossplatform`, `B2CrossSPECConflict`, `B3SubagentBoundary`, `B4Frontmatter`, `B5CITier`, `B6SpecLintHeading`, `B7PathResolution`, `B8WorkingTreeHygiene`).
   - Function `Classify(report DefectReport) (DefectCategory, float64)` — heuristic-only (string + simple AST-shape matching, no LLM call). Returns category + confidence ∈ [0, 1].
   - Function `AppendObservation(memoryPath string, category DefectCategory, evidence string) error` — appends YAML observation entry to lessons memory log.
2. **F.2** Add `internal/harness/capture/defect_detector_test.go`:
   - Table-driven tests with one synthetic failure-report fixture per B1–B8 category.
   - Confidence threshold assertion: classified category confidence ≥ 0.7.
   - Coverage target: ≥ 90% line coverage on `defect_detector.go`.
3. **F.3** Add `internal/harness/capture/lessons_injector.go`:
   - Function `LoadRelevantLessons(promptKeywords []string, memoryPath string, maxEntries int) ([]LessonEntry, error)` — keyword-matching lookup against lessons memory; returns up to `maxEntries` matches.
   - Function `PrependToPrompt(prompt string, lessons []LessonEntry) string` — prepends lessons content to delegation prompt body, idempotent.
4. **F.4** Add `internal/harness/capture/lessons_injector_test.go` — unit tests; coverage target ≥ 90%.
5. **F.5** Wire `Classify` invocation into the existing SubagentStop hook dispatch path created by W3 (modify only the dispatch site, not the W3 hook code itself; if integration touches W3 files, defer to M5 integration phase).
6. **F.6** Subagent-boundary CI guard: `internal/harness/capture/subagent_boundary_test.go` verifying no AskUserQuestion reference appears in any `defect_detector.go` / `lessons_injector.go` non-test source.

**Deliverables**: 4 new Go files (2 production + 2 test) under `internal/harness/capture/`, boundary CI guard, ≥ 90% coverage on new code.

**Exit criteria**: AC-WO-005, AC-WO-010.

**Risk mitigation**:
- C-WO-001 (subagent boundary): `subagent_boundary_test.go` enforces.
- C-WO-002 (no syscall): heuristic-only design means no `os.File`-level operations are needed for classification (filesystem I/O is limited to `AppendObservation` write, which uses `os.OpenFile` only — no `syscall.Flock`).
- C-WO-003 (no W3 file modification): All new code lives in NEW files; W3 capture.go is not touched.

---

### M4 — Agent Domain (Layer G — plan-auditor D7/D8)

**Goal**: Extend `.claude/agents/moai/plan-auditor.md` with two new evaluation dimensions; register weights in evaluator profiles.

**Tasks**:

1. **G.1** Append D7 (Cross-SPEC Reconciliation) dimension to `.claude/agents/moai/plan-auditor.md`:
   - Definition: "Verifies that every SPEC ID referenced in the body has its current status documented in `.moai/specs/<ID>/spec.md` frontmatter; flags BLOCKING if a referenced SPEC is `retired`, `superseded`, or `archived` without explicit reconciliation in the new SPEC body."
   - Verb to use: `git ls-files .moai/specs/<referenced-SPEC-ID>/spec.md | xargs grep '^status:'`.
   - Severity rubric: BLOCKING for unresolved retirement conflict; SHOULD for missing-but-recoverable references.
2. **G.2** Append D8 (Cross-Platform Discipline) dimension to `.claude/agents/moai/plan-auditor.md`:
   - Definition: "Verifies that SPECs introducing `syscall` imports declare a `//go:build` constraint in the SPEC body OR explicitly justify cross-platform exemption."
   - Verb to use: scan SPEC body for `syscall` substring; if present, verify nearby `//go:build` literal OR explicit `EXCL` entry.
   - Severity rubric: BLOCKING if syscall introduced without constraint OR justification.
3. **G.3** Register D7/D8 weights in `.moai/config/evaluator-profiles/default.md` and `.moai/config/evaluator-profiles/frontend.md`.
4. **G.4** Mirror plan-auditor edit to `internal/template/templates/.claude/agents/moai/plan-auditor.md`; mirror evaluator-profiles edits; `make build`.
5. **G.5** Add fixture-based plan-auditor unit test (lightweight markdown audit harness): `internal/cli/plan_audit_d7_d8_test.go`:
   - Fixture 1: SPEC with reference to a retired SPEC ID — expect D7 BLOCKING.
   - Fixture 2: SPEC with `syscall.Flock` reference and no build-tag — expect D8 BLOCKING.
   - Fixture 3: SPEC with `syscall` + documented `//go:build` constraint — expect D8 PASS.

**Deliverables**: Modified `plan-auditor.md` (+ template mirror), modified `default.md`/`frontend.md` evaluator profiles (+ mirrors), 3 fixture tests, `make build` artifact.

**Exit criteria**: AC-WO-004, AC-WO-011, AC-WO-012.

---

### M5 — Integration & Self-Validation (Dogfooding)

**Goal**: Apply the M1–M4 deliverables to the next manager-develop delegation (which is this SPEC's own run-phase) and measure wall-time.

**Tasks**:

1. **M5.1** Verify Layer A: this SPEC's run-phase delegation prompt uses the 5-section template structure (manual reading of the orchestrator's spawn prompt; record screenshot or text snapshot).
2. **M5.2** Verify Layer F: when the run-phase manager-develop completes (or fails), confirm that `internal/harness/capture/defect_detector.go` runs against the dispatch payload (check `.moai/research/observations/` directory for new entry).
3. **M5.3** Verify Layer G: re-run plan-auditor on this SPEC's plan files via `moai plan audit` (or equivalent); confirm D7 = PASS (no cross-SPEC retirement conflict) and D8 = N/A (no syscall introduced).
4. **M5.4** Record wall-time: timestamp at run-phase start (orchestrator delegation prompt sent) and at run-phase end (M1–M4 commits pushed to feature branch). Target ≤ 30 min.
5. **M5.5** Record manager-develop delegation count: count of distinct `Use the manager-develop subagent…` invocations. Target ≤ 1 (1-pass success).

**Deliverables**: Wall-time measurement entry in `progress.md`, screenshot/text snapshot of delegation prompt, observations log entry.

**Exit criteria**: AC-WO-001 (wall-time), AC-WO-002 (template applied), AC-WO-014 (no NEW spec-lint findings).

---

### M6 — Documentation + Lessons Archive

**Goal**: Update lessons memory; index in MEMORY.md; finalize PROGRESS.md.

**Tasks**:

1. **M6.1** Append new lesson entry "lessons #22 (workflow optimization 8-layer pattern)" to lessons memory (`~/.claude/projects/.../memory/lessons.md` or equivalent).
2. **M6.2** Update `MEMORY.md` index with one-line entry pointing to project_v3r5_workflow_opt_001_complete.md.
3. **M6.3** Mark any superseded entries with `[SUPERSEDED by …]` prefix.
4. **M6.4** Finalize `progress.md` with all M1–M5 task completion timestamps and the final wall-time measurement.

**Deliverables**: Lessons archive entry, MEMORY.md index update, final progress.md.

**Exit criteria**: All M1–M5 ACs satisfied; lessons archived for future SPEC cycles.

---

## 3. Brownfield Inventory

### 3.1 PRESERVE List (DO NOT modify)

| File                                                                 | Why preserved                                                                                        |
|----------------------------------------------------------------------|------------------------------------------------------------------------------------------------------|
| `internal/harness/capture/capture.go` (created by W3)                | W3 deliverable; Layer F adds sibling files only.                                                     |
| `internal/harness/capture/capture_test.go` (created by W3)           | W3 deliverable; M3 tests live in new `defect_detector_test.go` and `lessons_injector_test.go`.       |
| `.claude/agents/moai/plan-auditor.md` dimensions D1–D6               | Existing assessment surface; D7 and D8 are appended only.                                            |
| `.moai/config/sections/workflow.yaml` existing keys (`auto_selection`, `default_model`, `delegate_mode`, `enabled`, `mode`-related) | mode dispatch ownership belongs to SPEC-V3R2-WF-003/WF-004. Layer B adds `role_profiles` map only. |
| All existing rule files except those listed in EXTEND list           | Add-only extension; no rewrite.                                                                       |

### 3.2 EXTEND List (modify or add)

| File                                                                                                  | Action  | Layer |
|-------------------------------------------------------------------------------------------------------|---------|-------|
| `.claude/rules/moai/development/manager-develop-prompt-template.md`                                   | MIRROR  | A     |
| `.claude/rules/moai/workflow/ci-watch-protocol.md`                                                    | EXTEND  | C     |
| `.claude/rules/moai/core/agent-common-protocol.md`                                                    | EXTEND  | D, H  |
| `.claude/rules/moai/workflow/spec-workflow.md`                                                        | EXTEND  | E     |
| `.claude/rules/moai/workflow/agent-teams-pattern.md`                                                  | NEW     | B     |
| `.claude/rules/moai/workflow/verification-batch-pattern.md`                                           | NEW     | D     |
| `.moai/config/sections/workflow.yaml`                                                                 | EXTEND  | B     |
| `internal/harness/capture/defect_detector.go`                                                         | NEW     | F     |
| `internal/harness/capture/defect_detector_test.go`                                                    | NEW     | F     |
| `internal/harness/capture/lessons_injector.go`                                                        | NEW     | F     |
| `internal/harness/capture/lessons_injector_test.go`                                                   | NEW     | F     |
| `internal/harness/capture/subagent_boundary_test.go`                                                  | NEW     | F     |
| `.claude/agents/moai/plan-auditor.md`                                                                 | EXTEND  | G     |
| `.moai/config/evaluator-profiles/default.md`                                                          | EXTEND  | G     |
| `.moai/config/evaluator-profiles/frontend.md`                                                         | EXTEND  | G     |
| `internal/template/templates/.claude/rules/moai/**` (all mirrored paths above)                        | MIRROR  | M1.X  |
| `internal/template/templates/.moai/config/sections/workflow.yaml`                                     | MIRROR  | M2.B.4 |
| `internal/template/templates/.claude/agents/moai/plan-auditor.md`                                     | MIRROR  | G     |
| `internal/template/templates/.moai/config/evaluator-profiles/{default,frontend}.md`                   | MIRROR  | G     |
| `internal/template/rule_template_mirror_test.go` (or extend existing audit test)                      | NEW     | M1.Y  |
| `internal/config/workflow_role_profiles_test.go`                                                      | NEW     | M2.B.5 |
| `internal/cli/plan_audit_d7_d8_test.go` (or fixture-based audit harness)                              | NEW     | G     |

### 3.3 Cross-SPEC Reconciliation (per Vision §8)

| Touched area                                       | Related SPEC(s)                                  | Impact                                                                                              |
|----------------------------------------------------|--------------------------------------------------|------------------------------------------------------------------------------------------------------|
| `internal/harness/capture/`                        | SPEC-V3R5-HARNESS-AUTONOMY-001 (W3, in-flight)   | Layer F adds NEW files only; W3 capture.go/capture_test.go untouched. PRESERVED.                    |
| `.claude/agents/moai/plan-auditor.md` dimensions   | (existing plan-auditor agent; no parent SPEC located) | Layer G appends D7+D8 only; D1–D6 preserved verbatim.                                          |
| `.moai/config/sections/workflow.yaml`              | SPEC-V3R2-WF-003, SPEC-V3R2-WF-004 (mode dispatch) | Layer B adds `role_profiles` keys only; mode dispatch keys untouched. PRESERVED.                 |
| `.claude/rules/moai/workflow/spec-workflow.md`    | SPEC-V3R2-SPC-001 (EARS hierarchical AC schema)  | Layer E modifies Phase Transitions section only; AC schema untouched.                              |
| `.claude/rules/moai/workflow/ci-watch-protocol.md` | (no specific owner SPEC; rule extension)         | Layer C extends with new section; existing content preserved.                                       |

This SPEC operates **add-only** — no existing SPEC's retirement, supersession, or reversal is required.

---

## 4. Technical Approach

### 4.1 Defect Detector Heuristics (Layer F detail)

Heuristic table (initial design — refined during M3):

| Category | Detection signal                                                                                    | Confidence base |
|----------|------------------------------------------------------------------------------------------------------|-----------------|
| B1 Cross-platform | Report mentions `syscall.` + (`windows` OR `darwin` OR `linux`) without preceding `build tag`    | 0.85            |
| B2 Cross-SPEC     | Report mentions `Retired` OR `Superseded` OR `deprecation-marker` + SPEC-ID pattern               | 0.80            |
| B3 Subagent       | Report mentions `AskUserQuestion` OR `mcp__askuser` in non-orchestrator file path                  | 0.95            |
| B4 Frontmatter    | Report mentions `FrontmatterInvalid` OR snake_case field name (`created_at:`, `updated_at:`, `labels:`) | 0.90        |
| B5 CI Tier        | Report distinguishes spec-lint/golangci-lint/test failure OR mentions "chicken-and-egg"            | 0.75            |
| B6 Heading        | Report mentions `MissingExclusions` OR "h3 sub-section"                                            | 0.90            |
| B7 Path           | Report mentions `observer.go` OR `os.Getwd` OR "path resolution"                                   | 0.70            |
| B8 Working tree   | Report mentions untracked files OR "stray files" OR `usage-log.jsonl`                              | 0.70            |

Confidence below 0.7 is treated as "unclassified" and logged separately for human triage (no automatic memory append).

### 4.2 Lessons Injector Algorithm (Layer F detail)

```
function LoadRelevantLessons(promptKeywords, memoryPath, maxEntries):
    entries := parseYAMLEntries(memoryPath)
    scored := []
    for entry in entries:
        score := jaccard_similarity(entry.tags ∪ entry.keywords, promptKeywords)
        if score > 0.3:
            scored.append((entry, score))
    sort scored by score desc
    return first maxEntries of scored
```

Maximum 3 entries per delegation prompt (avoid context bloat). Each entry truncated to ≤ 500 characters.

### 4.3 plan-auditor D7/D8 Verb Design (Layer G detail)

D7 verb (executed inside plan-auditor agent during audit):
```bash
# Extract SPEC-ID references from new SPEC body
grep -Eo 'SPEC-[A-Z][A-Z0-9]+-[0-9]+' <new-spec.md> | sort -u | while read SID; do
  if [ -f ".moai/specs/$SID/spec.md" ]; then
    STATUS=$(grep '^status:' ".moai/specs/$SID/spec.md" | head -1 | cut -d: -f2 | tr -d ' ')
    case "$STATUS" in
      retired|superseded|archived)
        echo "BLOCKING: $SID has status=$STATUS but is referenced without reconciliation"
        ;;
    esac
  fi
done
```

D8 verb:
```bash
# Detect syscall introduction
if grep -q 'syscall' <new-spec.md>; then
  if ! grep -qE '//go:build|cross-platform exemption|EXCL.*syscall' <new-spec.md>; then
    echo "BLOCKING: SPEC references syscall but no //go:build constraint or EXCL justification"
  fi
fi
```

Both verbs are executed by the plan-auditor agent at audit time, not at SPEC creation.

---

## 5. Testing Strategy

### 5.1 Unit Testing

- **M3 (Layer F)**: Table-driven Go tests; 8 B-category fixtures + edge cases. Coverage target ≥ 90% via `go test -coverprofile=coverage.out ./internal/harness/capture/`.
- **M2 (Layer B)**: Config validation test (`internal/config/workflow_role_profiles_test.go`); verifies role_profiles map shape.
- **M4 (Layer G)**: Fixture-based plan-auditor harness test; 3 fixtures (D7 BLOCKING / D8 BLOCKING / D8 PASS).
- **M1.Y (CI guard)**: Mirror drift detection test; runs as part of `go test ./internal/template/...`.

### 5.2 Integration Testing

- **M5 dogfooding**: This SPEC's own run-phase serves as the integration test for Layers A, D, E, F, G. Wall-time measurement is the integration acceptance criterion.

### 5.3 spec-lint Regression

- Run `go run ./cmd/moai spec lint --strict` before and after each milestone commit. NEW findings = 0 (delta-only); existing baseline (1 pre-existing warning unrelated to this SPEC) is acceptable.

### 5.4 CI 3-Tier Verification (lessons #19)

- Tier 1 (spec-lint): zero NEW findings introduced.
- Tier 2 (golangci-lint): zero NEW findings introduced.
- Tier 3 (Test per OS): green on linux, macos, windows for all M3 Go additions. `GOOS=windows GOARCH=amd64 go build ./...` smoke check (since Layer F is pure-Go heuristic with no syscall use, this should pass trivially).

---

## 6. Risk Register

| ID    | Risk                                                                                                       | Probability | Impact | Mitigation                                                                                                                |
|-------|------------------------------------------------------------------------------------------------------------|-------------|--------|---------------------------------------------------------------------------------------------------------------------------|
| R-WO-01 | Agent Teams (Layer B) is experimental; instability could break solo-mode fallback                        | Medium      | High   | Layer B adds keys only; existing solo-mode keys untouched. Solo-mode regression test added in M2.                         |
| R-WO-02 | Lessons auto-capture (Layer F) produces false positives that pollute memory                                | Medium      | Medium | Heuristic-only (no LLM); confidence ≥ 0.7 threshold; manual triage for sub-threshold matches.                             |
| R-WO-03 | plan-auditor D7/D8 over-strict — blocks legitimate SPECs                                                   | Low         | Medium | iteration-1 plan-auditor verdict already follows the user-review pattern (existing). D7/D8 inherit that gate.            |
| R-WO-04 | M3 (capture extension) blocked by W3 PR #1024 not yet merged                                               | Medium      | Medium | M3 starts on W3 feature branch; rebase on merge. M3 deliverables are NEW files so rebase has no conflict surface.        |
| R-WO-05 | Template mirror drift accidentally introduced                                                              | Low         | High   | M1.Y CI guard (`rule_template_mirror_test.go`) fails build; mandatory `make build` before commit.                         |
| R-WO-06 | Background CI watch (`run_in_background: true`) blocks main session unexpectedly                          | Low         | Medium | Layer C rule includes explicit notification pattern; if `gh pr checks --watch` hangs, orchestrator can foreground-poll.   |
| R-WO-07 | Dogfooding self-validation (M5) does not achieve ≤ 30 min — AC-WO-001 fails                                | Medium      | Medium | Even partial improvement (e.g., 60 min vs 91 min) is a measurable gain; M5 records actual and updates target for follow-up. |

---

## 7. Implementation Hints (for manager-develop)

### 7.1 Order of Operations

1. **M1 first**: lands all rule changes in one go. Rules are independent of code; landing them early enables M2/M3/M4 to start in parallel.
2. **M2, M3, M4 in parallel** (if Agent Teams is enabled and configured by M1 completion of B.2). Otherwise sequential.
3. **M5 last**: requires M1–M4 deliverables on the feature branch.
4. **M6**: post-PR-merge cleanup.

### 7.2 Make Build Discipline

After EVERY rule or config edit:
```bash
make build
git status -- internal/template/embedded.go    # must be modified
git add internal/template/embedded.go
```

### 7.3 Subagent Boundary Self-Check (C-WO-001)

Before committing any Go file under `internal/harness/capture/`:
```bash
grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/capture/ | grep -v "_test.go" | grep -v "// "
# Expected output: (empty)
```

### 7.4 Cross-Platform Smoke (C-WO-002)

Before committing Layer F:
```bash
GOOS=windows GOARCH=amd64 go build ./internal/harness/capture/...
GOOS=darwin GOARCH=arm64 go build ./internal/harness/capture/...
GOOS=linux GOARCH=amd64 go build ./internal/harness/capture/...
```

### 7.5 spec-lint Delta-Only Check (AC-WO-014)

Before opening run-PR:
```bash
git stash    # set aside changes
BEFORE=$(go run ./cmd/moai spec lint --strict 2>&1 | grep -cE '^(WARNING|ERROR)')
git stash pop
AFTER=$(go run ./cmd/moai spec lint --strict 2>&1 | grep -cE '^(WARNING|ERROR)')
test "$AFTER" -le "$BEFORE"    # delta must be ≤ 0 NEW findings
```

---

## 8. Definition of Done

- [ ] All 6 milestones (M1–M6) complete.
- [ ] All 14 ACs (AC-WO-001 through AC-WO-014) verified — see acceptance.md §7.
- [ ] No NEW spec-lint, golangci-lint, or Test failures introduced (delta-only semantics).
- [ ] Run-PR opened, plan-auditor verdict re-run, status `draft → implemented` after merge.
- [ ] Lessons #22 entry archived in memory; MEMORY.md index updated.
- [ ] Wall-time measurement recorded in progress.md (target ≤ 30 min, actual TBD).
