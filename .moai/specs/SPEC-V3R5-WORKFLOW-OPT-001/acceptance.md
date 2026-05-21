---
id: SPEC-V3R5-WORKFLOW-OPT-001
title: "Workflow Optimization — Acceptance Criteria"
version: "0.1.0"
status: draft
created: 2026-05-20
updated: 2026-05-20
author: GOOS Kim
priority: P1
phase: "v3.0.0 — Round 5"
module: ".claude/rules/moai + .moai/config/sections/workflow.yaml + internal/harness/capture + .claude/agents/moai/plan-auditor.md"
lifecycle: spec-anchored
tags: "workflow, acceptance, given-when-then, traceability, binary-ac"
---

# Acceptance Criteria — SPEC-V3R5-WORKFLOW-OPT-001

## 1. Conventions

- Every AC is **binary** (PASS or FAIL).
- Every AC includes a **verification command** that produces an unambiguous PASS/FAIL signal.
- Given-When-Then phrasing follows EARS-compatible structure (per SPEC-V3R2-SPC-001).
- "AC PASS" means the verification command exits 0 and matches the expected output pattern.

---

## 2. Functional Acceptance Criteria

### AC-WO-001 — Self-Dogfooding Wall-Time ≤ 30 min

**Given** this SPEC's run-phase begins (orchestrator delegates to manager-develop with milestone scope M1+M2+M3+M4),
**When** the run-phase completes (all M1–M4 commits pushed to feature branch),
**Then** the wall-clock elapsed time between delegation send and final push **shall** be ≤ 30 minutes.

**Verification command**:
```bash
# Recorded in progress.md as `wall_time_seconds: <value>`
# Verification:
ELAPSED=$(grep '^wall_time_seconds:' .moai/specs/SPEC-V3R5-WORKFLOW-OPT-001/progress.md | awk '{print $2}')
test "$ELAPSED" -le 1800     # 30 min = 1800 s
```

**Maps to**: REQ-WO-002, REQ-WO-003 (downstream beneficiaries).

---

### AC-WO-002 — Known-Issues B1–B8 Auto-Injection

**Given** the orchestrator generates a `manager-develop` delegation prompt for any SPEC,
**When** the prompt body is serialized,
**Then** the prompt **shall** contain all 8 string markers `B1`, `B2`, `B3`, `B4`, `B5`, `B6`, `B7`, `B8` under a section heading containing `Known Issues`.

**Verification command** (applied to this SPEC's own delegation prompt, captured at run-phase start as `.moai/specs/SPEC-V3R5-WORKFLOW-OPT-001/delegation-prompt.txt`):
```bash
for tag in B1 B2 B3 B4 B5 B6 B7 B8; do
  grep -q "\\b$tag\\b" .moai/specs/SPEC-V3R5-WORKFLOW-OPT-001/delegation-prompt.txt \
    || { echo "FAIL: $tag missing from delegation prompt"; exit 1; }
done
grep -q '^## .*Known Issues' .moai/specs/SPEC-V3R5-WORKFLOW-OPT-001/delegation-prompt.txt
```

Exit code 0 = PASS.

**Maps to**: REQ-WO-001, REQ-WO-040.

---

### AC-WO-003 — Template Mirror Parity + Successful `make build`

**Given** any rule edit under `.claude/rules/moai/**` introduced by M1, M2, or M4,
**When** the change is staged for commit,
**Then** the corresponding file under `internal/template/templates/.claude/rules/moai/**` (or the templates path for `.moai/config/` and `.claude/agents/moai/`) **shall** exist with byte-identical content, AND `make build` **shall** produce a non-empty diff in `internal/template/embedded.go`.

**Verification command**:
```bash
# For each modified rule file, verify mirror parity
git diff --name-only main...HEAD -- '.claude/rules/moai/**' '.moai/config/**' '.claude/agents/moai/**' | while read src; do
  if [[ "$src" =~ ^\.claude/rules/moai/ ]]; then
    mirror="internal/template/templates/$src"
  elif [[ "$src" =~ ^\.moai/config/ ]]; then
    mirror="internal/template/templates/$src"
  elif [[ "$src" =~ ^\.claude/agents/moai/ ]]; then
    mirror="internal/template/templates/$src"
  else
    continue
  fi
  test -f "$mirror" || { echo "FAIL: mirror missing for $src"; exit 1; }
  diff -q "$src" "$mirror" || { echo "FAIL: mirror drift for $src"; exit 1; }
done

# Verify make build regenerated embedded.go
make build
git status --porcelain internal/template/embedded.go | grep -q '^.M' \
  || { echo "FAIL: make build did not regenerate embedded.go"; exit 1; }
```

Exit code 0 = PASS.

**Maps to**: REQ-WO-004, REQ-WO-041.

---

### AC-WO-004 — D7 Detects Simulated V3R4 Retirement Conflict

**Given** a synthetic test fixture SPEC body referencing a `retired` SPEC ID (e.g., `SPEC-V3R4-FOO-001` whose `status:` field is `retired`),
**When** plan-auditor D7 evaluation runs against the fixture,
**Then** D7 **shall** emit a BLOCKING finding identifying the conflicting SPEC ID.

**Verification command**:
```bash
go test -run TestPlanAuditD7_RetiredSPECConflict -v ./internal/cli/...
# Expected: PASS line "BLOCKING: SPEC-V3R4-FOO-001 has status=retired but is referenced without reconciliation"
```

Exit code 0 = PASS.

**Maps to**: REQ-WO-012, REQ-WO-021.

---

### AC-WO-005 — Defect Detector Coverage ≥ 90%

**Given** Layer F is implemented (M3 complete),
**When** Go coverage tool runs against `internal/harness/capture/defect_detector.go`,
**Then** the line-coverage ratio **shall** be ≥ 0.90.

**Verification command**:
```bash
go test -coverprofile=cover.out ./internal/harness/capture/...
COVERAGE=$(go tool cover -func=cover.out | grep 'defect_detector.go' | awk '{print $3}' | sed 's/%//')
awk "BEGIN { exit !($COVERAGE >= 90) }"
```

Exit code 0 = PASS.

**Maps to**: REQ-WO-020.

---

### AC-WO-006 — Background CI Watch Pattern Documented

**Given** Layer C edit to `ci-watch-protocol.md`,
**When** the rule body is inspected,
**Then** it **shall** contain both string literals `gh pr checks --watch` AND `run_in_background: true` within the same "Background watch" section.

**Verification command**:
```bash
awk '/^##.*Background.*watch/,/^##[^#]/' .claude/rules/moai/workflow/ci-watch-protocol.md \
  | grep -q 'gh pr checks --watch' \
  && awk '/^##.*Background.*watch/,/^##[^#]/' .claude/rules/moai/workflow/ci-watch-protocol.md \
  | grep -q 'run_in_background: true'
```

Exit code 0 = PASS.

**Maps to**: REQ-WO-003.

---

### AC-WO-007 — Verification Batch Example in Parallel Execution Section

**Given** Layer D edit to `agent-common-protocol.md`,
**When** the rule body is inspected,
**Then** the `## Parallel Execution` (or sub-section) **shall** include an example block that demonstrates exactly 7 distinct Bash verification commands invoked within a single agent turn (test / coverage / grep / sentinel / CLI / benchmark / lint).

**Verification command**:
```bash
# Count distinct verification keywords in the canonical example
awk '/^##.*Parallel/,/^##[^#]/' .claude/rules/moai/core/agent-common-protocol.md > /tmp/parallel.txt
for kw in 'go test' 'coverprofile' 'grep ' 'sentinel' 'cmd/moai' 'bench' 'lint'; do
  grep -q "$kw" /tmp/parallel.txt || { echo "FAIL: $kw missing from parallel example"; exit 1; }
done
```

Exit code 0 = PASS.

**Maps to**: REQ-WO-002.

---

### AC-WO-008 — Plan Audit Gate Skip Policy Documented

**Given** Layer E edit to `spec-workflow.md`,
**When** the rule body is inspected,
**Then** the "Phase Transitions" section **shall** contain the literal threshold `0.90` (or `PASS ≥ 0.90`) within a sentence describing run-phase audit gate skip semantics.

**Verification command**:
```bash
awk '/Phase Transitions/,/^##[^#]/' .claude/rules/moai/workflow/spec-workflow.md \
  | grep -qE '(0\.90|PASS ≥ 0\.90|Plan Audit Gate.*skip)'
```

Exit code 0 = PASS.

**Maps to**: REQ-WO-011.

---

### AC-WO-009 — `workflow.yaml` role_profiles Map Has Exactly 7 Keys

**Given** Layer B edit to `.moai/config/sections/workflow.yaml`,
**When** the file is parsed as YAML,
**Then** the `workflow.team.role_profiles` map **shall** contain exactly 7 entries, AND each entry **shall** declare exactly 3 sub-keys (`mode`, `model`, `isolation`).

**Verification command**:
```bash
go test -run TestWorkflowRoleProfilesShape -v ./internal/config/...
# Test asserts:
#   len(cfg.Workflow.Team.RoleProfiles) == 7
#   for each entry: all 3 sub-keys present
```

Exit code 0 = PASS.

**Maps to**: REQ-WO-010.

---

### AC-WO-010 — Defect Classification Confidence ≥ 0.7

**Given** a synthetic `manager-develop` failure report for any B-category (B1–B8),
**When** `defect_detector.Classify(report)` runs,
**Then** the returned category **shall** match the expected B-category AND the confidence **shall** be ≥ 0.7.

**Verification command**:
```bash
go test -run TestDefectClassify_AllCategories -v ./internal/harness/capture/...
# Test iterates B1..B8 fixtures and asserts:
#   got_category == want_category
#   got_confidence >= 0.7
```

Exit code 0 = PASS.

**Maps to**: REQ-WO-020, REQ-WO-023.

---

### AC-WO-011 — D7 Dimension Defined in plan-auditor.md

**Given** Layer G edit to `.claude/agents/moai/plan-auditor.md`,
**When** the file is inspected,
**Then** it **shall** contain a section heading or labelled list item matching `D7` AND the body **shall** mention `cross-SPEC` (or equivalent) AND reference `.moai/specs/` status check.

**Verification command**:
```bash
grep -E '^(##|###|\*\s+\*\*?)D7\b' .claude/agents/moai/plan-auditor.md \
  && grep -E 'D7|cross-SPEC' .claude/agents/moai/plan-auditor.md | grep -q '\.moai/specs/'
```

Exit code 0 = PASS.

**Maps to**: REQ-WO-012, REQ-WO-021.

---

### AC-WO-012 — D8 Dimension Defined in plan-auditor.md

**Given** Layer G edit to `.claude/agents/moai/plan-auditor.md`,
**When** the file is inspected,
**Then** it **shall** contain a section heading or labelled list item matching `D8` AND the body **shall** mention both `syscall` AND `//go:build`.

**Verification command**:
```bash
grep -E '^(##|###|\*\s+\*\*?)D8\b' .claude/agents/moai/plan-auditor.md \
  && awk '/D8/,/^(##|###)[^#]/' .claude/agents/moai/plan-auditor.md | grep -q 'syscall' \
  && awk '/D8/,/^(##|###)[^#]/' .claude/agents/moai/plan-auditor.md | grep -q '//go:build'
```

Exit code 0 = PASS.

**Maps to**: REQ-WO-013, REQ-WO-022.

---

### AC-WO-013 — gh pr checks --json | jq Pattern Referenced

**Given** Layer H edit to `agent-common-protocol.md`,
**When** the rule body is inspected,
**Then** it **shall** contain the canonical pattern string `gh pr checks --json` AND `jq` in proximity (within the same section).

**Verification command**:
```bash
# Search for both within the same Tool Optimization section
awk '/Tool [Oo]ptimization|Tool [Uu]sage/,/^##[^#]/' .claude/rules/moai/core/agent-common-protocol.md \
  | grep -q 'gh pr checks --json' \
  && awk '/Tool [Oo]ptimization|Tool [Uu]sage/,/^##[^#]/' .claude/rules/moai/core/agent-common-protocol.md \
  | grep -q 'jq'
```

Exit code 0 = PASS.

**Maps to**: REQ-WO-002 (downstream beneficiary).

---

### AC-WO-014 — Zero NEW spec-lint / golangci-lint / Test Findings

**Given** all milestone commits land on the feature branch,
**When** spec-lint, golangci-lint, and `go test ./...` run against the feature branch HEAD,
**Then** the count of findings introduced by this SPEC's files **shall** equal 0 (delta-only semantics; existing baseline is preserved).

**Verification command**:
```bash
# Compare baseline (main) vs feature branch
git stash || true
git checkout main
BL_SPEC=$(go run ./cmd/moai spec lint --strict 2>&1 | grep -cE '^(WARNING|ERROR)')
BL_LINT=$(golangci-lint run ./... 2>&1 | grep -c '^.*: warning\|^.*: error' || true)
BL_TEST=$(go test ./... 2>&1 | grep -c '^FAIL\|^--- FAIL' || true)
git checkout -
git stash pop 2>/dev/null || true

FB_SPEC=$(go run ./cmd/moai spec lint --strict 2>&1 | grep -cE '^(WARNING|ERROR)')
FB_LINT=$(golangci-lint run ./... 2>&1 | grep -c '^.*: warning\|^.*: error' || true)
FB_TEST=$(go test ./... 2>&1 | grep -c '^FAIL\|^--- FAIL' || true)

test "$FB_SPEC" -le "$BL_SPEC" && test "$FB_LINT" -le "$BL_LINT" && test "$FB_TEST" -le "$BL_TEST"
```

Exit code 0 = PASS.

**Maps to**: REQ-WO-004, REQ-WO-041.

---

## 3. Edge Cases

### EC-WO-001 — Defect Detector Below-Threshold Match

**Scenario**: A failure report has classification confidence between 0.0 and 0.7.

**Expected behaviour**: `Classify` returns `DefectUnclassified` category; no observation is appended to memory; the report is written to `.moai/research/observations/unclassified-<timestamp>.yaml` for human triage.

**Verification command**:
```bash
go test -run TestDefectClassify_UnclassifiedSubThreshold -v ./internal/harness/capture/...
```

---

### EC-WO-002 — Lessons Injector Empty Memory File

**Scenario**: `LoadRelevantLessons` is called when memory file is empty or missing.

**Expected behaviour**: Function returns `([]LessonEntry{}, nil)` (no error). Delegation prompt is dispatched without lessons prepend.

**Verification command**:
```bash
go test -run TestLessonsInjector_EmptyMemory -v ./internal/harness/capture/...
```

---

### EC-WO-003 — plan-auditor D7 Missing Referenced SPEC

**Scenario**: SPEC body references a SPEC ID that does NOT exist in `.moai/specs/` (typo or future SPEC).

**Expected behaviour**: D7 emits SHOULD severity (not BLOCKING), with message indicating "referenced SPEC not found".

**Verification command**:
```bash
go test -run TestPlanAuditD7_MissingReferencedSPEC -v ./internal/cli/...
```

---

### EC-WO-004 — Agent Teams Disabled Fallback

**Scenario**: `CLAUDE_CODE_EXPERIMENTAL_AGENT_TEAMS` env var is unset.

**Expected behaviour**: `workflow.yaml team.enabled` is read; if true, role_profiles map is loaded but ignored at runtime; orchestrator falls back to solo-mode delegation (existing behaviour, REQ-WO-010 does not fire).

**Verification command**: Manual integration check; documented in M2 task notes.

---

### EC-WO-005 — Mirror Drift Test Detects Missing Mirror

**Scenario**: Developer edits a rule file but forgets to mirror.

**Expected behaviour**: `internal/template/rule_template_mirror_test.go` fails with `RuleTemplateMirrorDrift` and identifies the missing path.

**Verification command**:
```bash
go test -run TestRuleTemplateMirrorDrift -v ./internal/template/...
```

---

### EC-WO-006 — Template Build Idempotency

**Scenario**: `make build` runs twice in succession with no source change.

**Expected behaviour**: Second run produces no diff in `internal/template/embedded.go` (idempotent).

**Verification command**:
```bash
make build
git status --porcelain internal/template/embedded.go > /tmp/diff1
make build
git status --porcelain internal/template/embedded.go > /tmp/diff2
diff /tmp/diff1 /tmp/diff2     # must be identical (both empty or both same content)
```

---

## 4. Quality Gates

| Gate           | Threshold                                                                                    | Verification                                                                                       |
|----------------|----------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------|
| Coverage       | `internal/harness/capture/defect_detector.go` ≥ 90% line coverage                            | AC-WO-005                                                                                          |
| Coverage       | `internal/harness/capture/lessons_injector.go` ≥ 90% line coverage                           | `go test -coverprofile=cover.out ./internal/harness/capture/...; go tool cover -func=cover.out`    |
| spec-lint      | 0 NEW findings (delta-only)                                                                  | AC-WO-014 tier 1                                                                                   |
| golangci-lint  | 0 NEW findings (delta-only)                                                                  | AC-WO-014 tier 2                                                                                   |
| Test (linux)   | All packages PASS                                                                            | AC-WO-014 tier 3                                                                                   |
| Test (macos)   | All packages PASS                                                                            | AC-WO-014 tier 3                                                                                   |
| Test (windows) | All packages PASS (Layer F has no syscall use, so trivial)                                   | AC-WO-014 tier 3                                                                                   |
| Subagent boundary | 0 AskUserQuestion references in `internal/harness/capture/` non-test source                | `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/capture/ \| grep -v "_test.go" \| grep -v "// "` |

---

## 5. Manual Verification Items

These items lack automated verification commands; they are confirmed via human inspection during M5/M6:

- **REQ-WO-030** (Optional evaluator-active auto-spawn): observed only if a P0/thorough SPEC is run during M5; not a M5 gating condition.
- **AC-WO-001** wall-time measurement requires human stopwatch at run-phase start/end (or git log timestamp diff between delegation commit and final M4 commit).

---

## 6. Definition of Done

- [ ] AC-WO-001 through AC-WO-014: all PASS.
- [ ] Edge cases EC-WO-001 through EC-WO-006: all PASS.
- [ ] All Quality Gates green.
- [ ] Run-PR merged with `status: implemented` frontmatter update.
- [ ] sync-PR merged with `status: completed` frontmatter update.
- [ ] Lessons #22 archived to memory.

---

## 7. REQ ↔ AC Traceability Matrix

| REQ ID        | Description (abridged)                              | AC Coverage                          |
|---------------|------------------------------------------------------|--------------------------------------|
| REQ-WO-001    | 5-section template in every manager-develop prompt   | AC-WO-002, AC-WO-003                 |
| REQ-WO-002    | Single-turn multi-Bash for verifications             | AC-WO-007 (canonical example)        |
| REQ-WO-003    | No idle during CI wait                               | AC-WO-006                            |
| REQ-WO-004    | Template mirror invariant                            | AC-WO-003, AC-WO-014                 |
| REQ-WO-010    | Agent Teams 5+ teammate spawn permission             | AC-WO-009                            |
| REQ-WO-011    | Plan Audit Gate skip on PASS ≥ 0.90                  | AC-WO-008                            |
| REQ-WO-012    | D7 cross-SPEC retirement BLOCKING                    | AC-WO-004, AC-WO-011                 |
| REQ-WO-013    | D8 syscall + build-tag BLOCKING                      | AC-WO-012                            |
| REQ-WO-020    | SubagentStop defect classification + append          | AC-WO-005, AC-WO-010                 |
| REQ-WO-021    | plan-PR open → D7 auto-scan                          | AC-WO-004, AC-WO-011                 |
| REQ-WO-022    | syscall in SPEC → D8 build-tag check                 | AC-WO-012                            |
| REQ-WO-023    | New defect → auto-prepend in next delegation         | AC-WO-010 (lessons_injector tests)   |
| REQ-WO-030    | (Optional) evaluator-active auto-spawn               | Manual (§5)                          |
| REQ-WO-040    | Delegation without Section B → halt                  | AC-WO-002                            |
| REQ-WO-041    | Rule without mirror → CI fail                        | AC-WO-014 + EC-WO-005                |

100% REQ coverage by automated ACs (REQ-WO-030 is the sole `may` requirement, covered by manual §5).
