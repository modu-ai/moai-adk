# SPEC-V3R2-ORC-003 Acceptance Criteria

> Detailed Given-When-Then scenarios and verification evidence for **Effort-Level Calibration Matrix for 17 agents**.
> Companion to `spec.md` v0.1.0, `plan.md` v0.1.0, `research.md` v0.1.0, `tasks.md` v0.1.0.

## HISTORY

| Version | Date       | Author                          | Description                                                                                                                                       |
|---------|------------|---------------------------------|---------------------------------------------------------------------------------------------------------------------------------------------------|
| 0.1.0   | 2026-05-10 | manager-spec (Plan workflow Phase 1B) | Initial acceptance.md. 10 AC scenarios mapped 1:1 to spec.md §6 AC-ORC-003-01..10. Each AC includes Given/When/Then + measurable evidence + REQ traceback + tasks.md task IDs. |

---

## 1. AC Coverage Map

| AC ID | Spec §6 statement | Mapped REQs | Mapped tasks | Evidence type |
|---|---|---|---|---|
| AC-ORC-003-01 | agent-authoring.md contains effort matrix table with 17 agents | REQ-001, REQ-002 | T-ORC003-04, T-ORC003-05, T-ORC003-23 | Markdown grep + table parse |
| AC-ORC-003-02 | 17 agent files have effort: matching canonical matrix | REQ-002, REQ-003 | T-ORC003-09..21 | Frontmatter scan |
| AC-ORC-003-03 | moai agent lint exits 0 with no LR-03 warnings on v3r2 roster | REQ-006 | T-ORC003-32 | CLI exit code |
| AC-ORC-003-04 | Agent without effort: causes CI to fail with LR-03 error | REQ-006, REQ-009 | T-ORC003-24 | CLI exit code + violation |
| AC-ORC-003-05 | Changing effort: on expert-security from xhigh to high fails CI with ORC_EFFORT_MATRIX_DRIFT | REQ-013 | T-ORC003-03, T-ORC003-25 | LR-12 fixture |
| AC-ORC-003-06 | Template and local trees diff equivalent for effort: fields | REQ-004 | T-ORC003-05, T-ORC003-26 | diff -r byte-identical |
| AC-ORC-003-07 | moai-constitution.md §Opus 4.7 Prompt Philosophy contains cross-reference but no duplicate table | REQ-005, REQ-008 | T-ORC003-06, T-ORC003-22 | Markdown grep |
| AC-ORC-003-08 | Attempting effort: ultra (invalid enum) fails with AGT_INVALID_FRONTMATTER | REQ-012 | T-ORC003-03, T-ORC003-25 | LR-13 fixture |
| AC-ORC-003-09 | Running MIG-001 migrator on a v2 agent tree rewrites declared-but-drifted effort values and logs each rewrite | REQ-007 | T-ORC003-07 (advisory) | Documentation cross-link only |
| AC-ORC-003-10 | The 4 explicit-drift corrections are visible in a git diff: expert-security, evaluator-active, plan-auditor, expert-refactoring frontmatter now show effort: xhigh | REQ-002, REQ-003 | T-ORC003-20, T-ORC003-21 | git diff inspection |

---

## 2. Acceptance Criteria

### AC-ORC-003-01: Canonical matrix table in agent-authoring.md

**REQ traceback**: REQ-ORC-003-001 (Ubiquitous — matrix publication), REQ-ORC-003-002 (Ubiquitous — matrix content)

**Mapped tasks**: T-ORC003-04 (table insertion), T-ORC003-05 (template parity), T-ORC003-23 (regression test)

**Given** the run-phase has applied this SPEC's deltas
**And** the file `.claude/rules/moai/development/agent-authoring.md` exists in the repository
**When** an inspector reads the file
**Then** the file MUST contain a section heading `## Effort-Level Calibration Matrix (SPEC-V3R2-ORC-003)`
**And** the section MUST contain a markdown table with at least 18 rows (1 header + 17 agent rows)
**And** the table MUST list these 17 agent names exactly: manager-spec, manager-strategy, manager-cycle, manager-quality, manager-docs, manager-git, manager-project, expert-backend, expert-frontend, expert-security, expert-devops, expert-performance, expert-refactoring, builder-platform, evaluator-active, plan-auditor, researcher
**And** the corresponding effort values MUST be: xhigh, xhigh, high, high, medium, medium, medium, high, high, xhigh, medium, high, xhigh, medium, xhigh, xhigh, xhigh
**And** every effort value MUST be in the 5-value enum {low, medium, high, xhigh, max}

**Verification command**:
```bash
grep -A 25 "## Effort-Level Calibration Matrix" .claude/rules/moai/development/agent-authoring.md | grep -E "manager-spec.*xhigh|builder-platform.*medium" | wc -l
```
Expected output: ≥ 2 (first agent + last agent rows present)

**Test fixture**: `internal/cli/agent_lint_test.go` `TestAuthoringDocHasEffortMatrix` (T-ORC003-23)

---

### AC-ORC-003-02: 17 agent files declare correct effort

**REQ traceback**: REQ-ORC-003-002 (Ubiquitous — matrix content), REQ-ORC-003-003 (Ubiquitous — frontmatter declaration)

**Mapped tasks**: T-ORC003-09..21 (17 frontmatter edits, each agent × 2 trees)

**Given** the 17 v3r2 agent files exist under `.claude/agents/moai/`
**When** an inspector parses each agent's YAML frontmatter
**Then** each of the 17 agents MUST declare an `effort:` field
**And** the value MUST exactly match the canonical matrix:

| Agent | Required effort |
|---|---|
| manager-spec | xhigh |
| manager-strategy | xhigh |
| manager-cycle | high |
| manager-quality | high |
| manager-docs | medium |
| manager-git | medium |
| manager-project | medium |
| expert-backend | high |
| expert-frontend | high |
| expert-security | xhigh |
| expert-devops | medium |
| expert-performance | high |
| expert-refactoring | xhigh |
| builder-platform | medium |
| evaluator-active | xhigh |
| plan-auditor | xhigh |
| researcher | xhigh |

**Verification command**:
```bash
for agent in manager-spec manager-strategy manager-cycle manager-quality manager-docs manager-git manager-project expert-backend expert-frontend expert-security expert-devops expert-performance expert-refactoring builder-platform evaluator-active plan-auditor researcher; do
  effort=$(awk '/^---$/{flag=!flag; next} flag && /^effort:/{print $2; exit}' ".claude/agents/moai/${agent}.md")
  printf "%-25s %s\n" "$agent" "$effort"
done
```

**Expected output**: 17 lines, each showing the matrix value above.

---

### AC-ORC-003-03: Lint clean on v3r2 roster

**REQ traceback**: REQ-ORC-003-006 (LR-03 promotion verification)

**Mapped tasks**: T-ORC003-32 (final verification)

**Given** the run-phase has populated all 17 agents per AC-02
**When** an operator runs `moai agent lint --path .claude/agents/moai/`
**Then** the command MUST exit with code 0 OR exit with code 1 only if violations are unrelated to LR-03/12/13/14 on the 17 v3r2 roster (out-of-roster agents may still trigger violations)
**And** filtering for LR-03 violations on the 17 v3r2 roster MUST yield zero hits
**And** filtering for LR-12 violations on the 17 v3r2 roster MUST yield zero hits

**Verification command**:
```bash
moai agent lint --path .claude/agents/moai/ --format=json | \
  jq '[.violations[] | select(.rule == "LR-03" or .rule == "LR-12") | select(.file | test("(manager-spec|manager-strategy|manager-cycle|manager-quality|manager-docs|manager-git|manager-project|expert-backend|expert-frontend|expert-security|expert-devops|expert-performance|expert-refactoring|builder-platform|evaluator-active|plan-auditor|researcher)\\.md$"))] | length'
```

**Expected output**: `0`

---

### AC-ORC-003-04: Adding agent without effort: causes CI failure

**REQ traceback**: REQ-ORC-003-006 (LR-03 promotion), REQ-ORC-003-009 (new agent requirement)

**Mapped tasks**: T-ORC003-24 (regression test)

**Given** the canonical matrix and LR-03 are in effect
**When** a contributor adds a new agent file (e.g., `.claude/agents/moai/test-agent.md`) without an `effort:` field
**And** CI runs `moai agent lint --path .claude/agents/moai/`
**Then** the command MUST exit with code 1
**And** the output MUST contain a violation with `Rule: LR-03`, `Severity: Error`, `File: .claude/agents/moai/test-agent.md`

**Test fixture** (`internal/cli/agent_lint_test.go` `TestLintLR03_MissingEffortIsError`, T-ORC003-24):

```go
func TestLintLR03_MissingEffortIsError(t *testing.T) {
    tmpDir := t.TempDir()
    agentFile := filepath.Join(tmpDir, "test-agent.md")
    content := []byte(`---
name: test-agent
description: test
tools: Read, Write
model: sonnet
---

Body.
`)
    require.NoError(t, os.WriteFile(agentFile, content, 0644))

    violations, err := lintAgentFile(agentFile, false)
    require.NoError(t, err)

    var lr03Violations []LintViolation
    for _, v := range violations {
        if v.Rule == "LR-03" {
            lr03Violations = append(lr03Violations, v)
        }
    }
    require.Len(t, lr03Violations, 1)
    require.Equal(t, SeverityError, lr03Violations[0].Severity)
}
```

**Expected**: PASS.

---

### AC-ORC-003-05: Drift introduction fails CI with ORC_EFFORT_MATRIX_DRIFT

**REQ traceback**: REQ-ORC-003-013 (Unwanted Behavior — drift re-introduction)

**Mapped tasks**: T-ORC003-03 (LR-12 implementation), T-ORC003-25 (LR-12 test fixture)

**Given** the canonical matrix declares `expert-security: xhigh`
**And** LR-12 is wired into `lintAgentFile` dispatcher
**When** a contributor changes `.claude/agents/moai/expert-security.md` frontmatter to `effort: high`
**And** runs `moai agent lint --path .claude/agents/moai/`
**Then** the command MUST exit with code 1
**And** the output MUST contain a violation with:
- `Rule: LR-12`
- `Severity: Error`
- `File: .claude/agents/moai/expert-security.md`
- `Message` containing the literal string `ORC_EFFORT_MATRIX_DRIFT`
- `Message` containing both the declared value (`high`) and the expected value (`xhigh`)

**Test fixture** (`internal/cli/agent_lint_test.go` `TestLintLR12_MatrixDrift_DriftedAgent`, T-ORC003-25):

```go
func TestLintLR12_MatrixDrift_DriftedAgent(t *testing.T) {
    tmpDir := t.TempDir()
    agentFile := filepath.Join(tmpDir, "expert-security.md")
    content := []byte(`---
name: expert-security
description: ...
tools: Read, Grep
model: sonnet
effort: high
---
`)
    require.NoError(t, os.WriteFile(agentFile, content, 0644))

    violations, err := lintAgentFile(agentFile, false)
    require.NoError(t, err)

    var lr12 []LintViolation
    for _, v := range violations {
        if v.Rule == "LR-12" {
            lr12 = append(lr12, v)
        }
    }
    require.Len(t, lr12, 1)
    require.Equal(t, SeverityError, lr12[0].Severity)
    require.Contains(t, lr12[0].Message, "ORC_EFFORT_MATRIX_DRIFT")
    require.Contains(t, lr12[0].Message, "high")
    require.Contains(t, lr12[0].Message, "xhigh")
}
```

**Expected**: PASS.

---

### AC-ORC-003-06: Template-local byte-identical parity

**REQ traceback**: REQ-ORC-003-004 (template sync)

**Mapped tasks**: T-ORC003-05 (`make build`), T-ORC003-26 (`diff -r` final gate)

**Given** the run-phase has edited both `.claude/agents/moai/*.md` and `internal/template/templates/.claude/agents/moai/*.md`
**And** `make build` has regenerated `internal/template/embedded.go`
**When** an operator runs `diff -r .claude/agents/moai/ internal/template/templates/.claude/agents/moai/`
**Then** the command MUST exit with code 0 (byte-identical trees)

**And** the same MUST hold for the rule files modified:
```bash
diff -r .claude/rules/moai/core/moai-constitution.md \
        internal/template/templates/.claude/rules/moai/core/moai-constitution.md
diff -r .claude/rules/moai/development/agent-authoring.md \
        internal/template/templates/.claude/rules/moai/development/agent-authoring.md
```
Both MUST exit code 0.

**Verification command**:
```bash
diff -r .claude/agents/moai/ internal/template/templates/.claude/agents/moai/ && \
diff .claude/rules/moai/core/moai-constitution.md internal/template/templates/.claude/rules/moai/core/moai-constitution.md && \
diff .claude/rules/moai/development/agent-authoring.md internal/template/templates/.claude/rules/moai/development/agent-authoring.md
echo $?
```

**Expected output**: `0`

---

### AC-ORC-003-07: Constitution cross-reference, no duplicate table

**REQ traceback**: REQ-ORC-003-005 (constitution cross-reference), REQ-ORC-003-008 (matrix is unique source)

**Mapped tasks**: T-ORC003-06 (constitution edit), T-ORC003-22 (regression test)

**Given** the run-phase has edited `.claude/rules/moai/core/moai-constitution.md`
**When** an inspector reads §Opus 4.7 Prompt Philosophy in that file
**Then** the file MUST contain a cross-reference to `.claude/rules/moai/development/agent-authoring.md` § Effort-Level Calibration Matrix (specifically, the literal substring `agent-authoring.md` MUST appear in §Opus 4.7 Prompt Philosophy)
**And** the file MUST NOT contain a duplicate 17-agent matrix table (`grep "manager-cycle.*high"` over the constitution should NOT match if the inline table was replaced; if it does match, the inline list must include cross-reference text instead)
**And** the §Opus 4.7 Prompt Philosophy "Effort level selection" bullet MUST still convey the 3-tier role categorization (reasoning-intensive / implementation / template+speed-critical) so that no FROZEN doctrinal content is lost.

**Verification command**:
```bash
grep -c "agent-authoring.md" .claude/rules/moai/core/moai-constitution.md
```
**Expected output**: ≥ 1 (cross-reference present).

**Test fixture** (`internal/cli/agent_lint_test.go` `TestConstitutionCrossReference`, T-ORC003-22):

```go
func TestConstitutionCrossReference(t *testing.T) {
    content, err := os.ReadFile(".claude/rules/moai/core/moai-constitution.md")
    require.NoError(t, err)

    s := string(content)
    // Find the Opus 4.7 Prompt Philosophy section
    sectionIdx := strings.Index(s, "## Opus 4.7 Prompt Philosophy")
    require.GreaterOrEqual(t, sectionIdx, 0, "Opus 4.7 Prompt Philosophy section missing")

    // Find the next section heading to bound the search
    nextSectionIdx := strings.Index(s[sectionIdx+5:], "\n## ")
    if nextSectionIdx == -1 {
        nextSectionIdx = len(s) - sectionIdx - 5
    }
    section := s[sectionIdx : sectionIdx+5+nextSectionIdx]

    require.Contains(t, section, "agent-authoring.md",
        "Opus 4.7 Prompt Philosophy section must cross-reference agent-authoring.md (SPEC-V3R2-ORC-003 REQ-005)")
}
```

**Expected**: PASS.

---

### AC-ORC-003-08: Invalid effort enum rejected

**REQ traceback**: REQ-ORC-003-012 (Unwanted Behavior — invalid enum rejection)

**Mapped tasks**: T-ORC003-03 (LR-13 implementation), T-ORC003-25 (LR-13 fixture)

**Given** LR-13 is wired into `lintAgentFile` dispatcher
**When** a contributor declares `effort: ultra` (not in {low, medium, high, xhigh, max}) on any agent
**And** runs `moai agent lint`
**Then** the command MUST exit with code 1
**And** the output MUST contain a violation with:
- `Rule: LR-13`
- `Severity: Error`
- `Message` containing `AGT_INVALID_FRONTMATTER` or equivalent enum-rejection signal
- `Message` referencing the illegal value `ultra`
- `Message` enumerating the valid 5-value enum

**Test fixture** (`internal/cli/agent_lint_test.go` `TestLintLR13_InvalidEffortEnum`, T-ORC003-25):

```go
func TestLintLR13_InvalidEffortEnum(t *testing.T) {
    tmpDir := t.TempDir()
    agentFile := filepath.Join(tmpDir, "bad-agent.md")
    content := []byte(`---
name: bad-agent
description: ...
tools: Read
model: sonnet
effort: ultra
---
`)
    require.NoError(t, os.WriteFile(agentFile, content, 0644))

    violations, err := lintAgentFile(agentFile, false)
    require.NoError(t, err)

    var lr13 []LintViolation
    for _, v := range violations {
        if v.Rule == "LR-13" {
            lr13 = append(lr13, v)
        }
    }
    require.Len(t, lr13, 1)
    require.Equal(t, SeverityError, lr13[0].Severity)
    require.Contains(t, lr13[0].Message, "ultra")
    require.Contains(t, lr13[0].Message, "low")  // valid enum present in message
}
```

**Expected**: PASS.

---

### AC-ORC-003-09: MIG-001 migrator drift rewrite (advisory)

**REQ traceback**: REQ-ORC-003-007 (Event-Driven — migrator drift rewrite)

**Mapped tasks**: T-ORC003-07 (research.md cross-link to MIG-001)

**Given** SPEC-V3R2-MIG-001 migrator is implemented (separate SPEC, downstream consumer)
**When** an operator runs the migrator against a v2 agent tree containing drifted effort values
**Then** the migrator SHOULD rewrite each drifted value to the canonical matrix entry
**And** SHOULD emit a migration log line per rewrite, naming the agent and the rewrite (`<old> → <new>`)

**Verification**: This AC is **advisory** in this SPEC. Code path is owned by SPEC-V3R2-MIG-001. This SPEC contributes:
1. The canonical matrix that MIG-001 references at migration time.
2. Documentation cross-link in `research.md` § 5.7 (T-ORC003-07).

**This AC will not be tested in this SPEC's CI**; the MIG-001 SPEC owns the runtime test.

**Status**: SOFT-PASS (documentation cross-link only; runtime verification deferred to MIG-001 acceptance gate).

---

### AC-ORC-003-10: 4 explicit drift corrections visible in git diff

**REQ traceback**: REQ-ORC-003-002, REQ-ORC-003-003 (drift correction)

**Mapped tasks**: T-ORC003-20 (DRIFT-3-A expert-security), T-ORC003-21 (DRIFT-3-B/C/D evaluator-active, plan-auditor, expert-refactoring)

**Given** the run-phase has applied the 4 drift corrections
**When** an operator runs `git diff plan/SPEC-V3R2-ORC-003..feat/SPEC-V3R2-ORC-003 -- '.claude/agents/moai/expert-security.md' '.claude/agents/moai/evaluator-active.md' '.claude/agents/moai/plan-auditor.md' '.claude/agents/moai/expert-refactoring.md'`
**Then** the diff output MUST contain four `-effort: high` removals and four `+effort: xhigh` additions, one per file

**Verification command** (post-merge):
```bash
git log --oneline main -- .claude/agents/moai/expert-security.md | head -1
# locate the merge commit
git show <merge-sha> -- .claude/agents/moai/expert-security.md | grep -E "^\+effort: xhigh|^-effort: high"
```
Expected: 1 removal of `-effort: high` + 1 addition of `+effort: xhigh` for each of the 4 agents (4 × 2 = 8 diff lines).

**Verification command** (during run-phase, before merge):
```bash
for agent in expert-security evaluator-active plan-auditor expert-refactoring; do
  git diff origin/main -- ".claude/agents/moai/${agent}.md" | grep -E "^\+effort: xhigh|^-effort: high"
done
```
Expected: 8 lines total (4 removals + 4 additions).

---

## 3. Definition of Done

The SPEC is considered DONE when:

- [ ] All 10 ACs (AC-ORC-003-01 through AC-ORC-003-10) verified
- [ ] All 14 REQs (REQ-ORC-003-001 through REQ-ORC-003-014) traced to ≥1 task per plan §1.5
- [ ] All 27 tasks (T-ORC003-01 through T-ORC003-27 per tasks.md) marked complete
- [ ] `go test -race -count=1 ./...` PASS
- [ ] `golangci-lint run` clean
- [ ] `make build` regenerates `internal/template/embedded.go` correctly
- [ ] `moai agent lint --path .claude/agents/moai/` reports 0 LR-03/12/13/14 errors for the 17 v3r2 agents
- [ ] Template-local parity verified via `diff -r`
- [ ] CHANGELOG entry written in Unreleased section
- [ ] @MX tags applied per plan §6
- [ ] Run PR squash-merged into main
- [ ] Sync PR squash-merged into main
- [ ] Worktree disposed via `moai worktree done SPEC-V3R2-ORC-003`

---

## 4. Quality Gate Criteria

Per `.moai/config/sections/quality.yaml`:

| Criterion | Target | Verification |
|---|---|---|
| Tested | All new functions covered | `go test -cover ./internal/cli/` ≥ 85% |
| Readable | English comments + clear naming | `golangci-lint run` clean |
| Unified | Style + import order | `gofmt -l ./internal/cli/` empty |
| Secured | No new attack surface | LR-12/13/14 are read-only; no shell injection |
| Trackable | All commits + CHANGELOG | git log inspection + CHANGELOG diff |

---

## 5. Risk-based AC Prioritization

| AC | Priority | Why |
|---|---|---|
| AC-01, AC-02, AC-10 | P0 | Core deliverable (matrix + frontmatter) |
| AC-04, AC-05, AC-08 | P0 | Lint enforcement (regression prevention) |
| AC-03, AC-06, AC-07 | P0 | Verification gates (CI required) |
| AC-09 | P2 | Advisory; deferred to MIG-001 |

---

End of acceptance.

Version: 0.1.0
Status: Acceptance artifact for SPEC-V3R2-ORC-003
