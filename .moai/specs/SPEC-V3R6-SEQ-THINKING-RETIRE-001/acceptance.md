---
id: SPEC-V3R6-SEQ-THINKING-RETIRE-001
title: "Acceptance — Sequential-Thinking MCP Retirement"
version: "0.1.0"
status: draft
created: 2026-05-23
updated: 2026-05-23
author: manager-spec
priority: P1
phase: "v3.6.0"
module: ".claude/agents,.claude/skills,.claude/rules,internal/template/templates,.mcp.json,settings.json"
lifecycle: spec-anchored
tags: "v3r6, mcp, retirement, ultrathink, acceptance, given-when-then, binary-ac"
tier: M
---

# Acceptance — SPEC-V3R6-SEQ-THINKING-RETIRE-001

## §0. Acceptance Scope Boundary

### §0.1 Out of Scope

These verification scenarios verify only the retirement of `sequential-thinking` MCP literals and the consolidation onto the `ultrathink` keyword. **Not verified** by this acceptance matrix:

- User-local Sequential Thinking MCP installation flows (out of project namespace).
- New Adaptive Thinking behavior (no new features added).
- Other MCP servers' lifecycle (context7, chrome-devtools, claude-in-chrome, zai-mcp-server are explicitly out of scope and their references must remain intact).
- Pre-existing baseline failures captured in M1 `baseline-test.txt` (only NEW regressions over baseline are reckoned against this SPEC).
- Run-phase commit messages themselves (manager-develop authors them at run-time; this acceptance matrix does not verify message content).

## §1. Traceability Matrix

| AC ID | REQ Coverage | Verification Type | Milestone | Severity |
|-------|--------------|-------------------|-----------|----------|
| AC-STR-001 | REQ-STR-001 | Binary grep | M2+M3+M4+M5+M6 (cumulative) | Blocking |
| AC-STR-002 | REQ-STR-001 | Binary grep + JSON validate | M5 | Blocking |
| AC-STR-003 | REQ-STR-001 | Binary grep + JSON validate | M4+M5 | Blocking |
| AC-STR-004 | REQ-STR-001, REQ-STR-002 | Binary grep | M6 | Blocking |
| AC-STR-005 | REQ-STR-002, REQ-STR-003 | Binary grep (positive + negative) | M4 | Blocking |
| AC-STR-006 | REQ-STR-004, REQ-STR-005 | Test execution | M6 | Blocking |
| AC-STR-007 | REQ-STR-006 | Build + test execution | M7 | Blocking |
| AC-STR-008 | REQ-STR-001, REQ-STR-008 | File diff verification | M7 | Should-fix |
| AC-STR-009 | REQ-STR-001 | Per-file grep | M2 | Blocking |
| AC-STR-010 | REQ-STR-001 | Per-file grep | M3+M4 | Blocking |
| AC-STR-011 | REQ-STR-006 | Cross-platform build | M7 | Blocking |
| AC-STR-012 | REQ-STR-002 | Binary grep | M6 | Should-fix |

## §2. Given/When/Then Scenarios

### AC-STR-001 — Project-namespace cleanliness (cumulative)

**Given** the run-phase milestones M2–M6 are complete and committed to the working tree;

**When** the orchestrator runs the global namespace grep across all source surfaces;

**Then** the command produces zero matches and exits with status 1 (grep convention for no match).

**Verification command**:

```bash
grep -rln 'sequential-thinking\|sequentialthinking' \
  .claude/agents/ \
  .claude/skills/ \
  .claude/rules/ \
  internal/template/templates/.claude/ \
  CLAUDE.md \
  ; echo "exit=$?"
```

**Expected**: `exit=1` (no matches). Matches in `internal/template/seq_thinking_retire_audit_allowlist.txt` (M6 artifact) are tolerated because the file is the audit allow-list, not source code. The grep above does not include that file path.

**Allow-list exception**: the BC marker line in `.claude/skills/moai-foundation-thinking/SKILL.md` ("Sequential Thinking MCP support retired in SPEC-V3R6-SEQ-THINKING-RETIRE-001") and its template mirror are NOT matched by the regex `'sequential-thinking\|sequentialthinking'` because the BC marker uses the title-cased phrase "Sequential Thinking" without a hyphen. Confirm with `grep -n 'Sequential Thinking' .claude/skills/moai-foundation-thinking/SKILL.md` — exactly one occurrence (the BC marker).

---

### AC-STR-002 — .mcp.json server entry removed

**Given** M5 is complete;

**When** the orchestrator greps the project MCP configuration files for `sequential-thinking`;

**Then** zero matches are found and both files validate as JSON.

**Verification commands**:

```bash
# Literal grep
grep -n 'sequential-thinking\|sequentialthinking' .mcp.json internal/template/templates/.mcp.json.tmpl ; echo "exit=$?"
# Expected: exit=1 (no matches)

# JSON validity
jq . .mcp.json > /dev/null && echo "local OK"
# .mcp.json.tmpl is a Go template — skip jq, instead test via render
go build ./cmd/moai && ls -la moai && rm -f moai && echo "template builds"
```

**Expected**: First command exits 1. Second command prints `local OK`. Third command builds successfully.

---

### AC-STR-003 — settings.json + skill body sequential-thinking removed

**Given** M4 and M5 are complete;

**When** the orchestrator greps `settings.json` files and the `moai-foundation-thinking` SKILL body for sequential-thinking literals;

**Then** zero matches in each.

**Verification commands**:

```bash
# settings allow-list + enabledMcpjsonServers
grep -n 'mcp__sequential-thinking\|"sequential-thinking"' \
  .claude/settings.json \
  internal/template/templates/.claude/settings.json.tmpl \
  ; echo "exit=$?"
# Expected: exit=1

# moai-foundation-thinking SKILL body
grep -c 'sequential-thinking\|sequentialthinking' \
  .claude/skills/moai-foundation-thinking/SKILL.md \
  internal/template/templates/.claude/skills/moai-foundation-thinking/SKILL.md
# Expected: 0 each
```

**Expected**: First command `exit=1`. Second command prints `0` for each file.

---

### AC-STR-004 — Root documentation retirement note + correction

**Given** M6 is complete;

**When** the orchestrator inspects `CLAUDE.md` §12 and `CLAUDE.local.md` §22.2;

**Then** CLAUDE.md §12 contains the retirement statement (referencing SPEC-V3R6-SEQ-THINKING-RETIRE-001) and CLAUDE.local.md §22.2 no longer mentions `mcp__sequential-thinking` as an alwaysLoad MCP.

**Verification commands**:

```bash
# CLAUDE.md retirement statement
grep -n 'SPEC-V3R6-SEQ-THINKING-RETIRE-001\|Sequential Thinking MCP retired' CLAUDE.md
# Expected: 1+ matches

# CLAUDE.local.md §22.2 correction
grep -n 'mcp__sequential-thinking' CLAUDE.local.md ; echo "exit=$?"
# Expected: exit=1 (no matches — the only previous mention is now removed)
# Tolerated: a "(retired)" or "SPEC-V3R6-SEQ-THINKING-RETIRE-001" mention may exist
grep -n 'SPEC-V3R6-SEQ-THINKING-RETIRE-001' CLAUDE.local.md
# Expected: 1+ matches (the correction reference)
```

**Expected**: First command finds the retirement statement. Second command exits 1. Third command finds the SPEC ID reference in §22.2.

---

### AC-STR-005 — moai-foundation-thinking body: ultrathink+ present, sequential-thinking absent

**Given** M4 is complete;

**When** the orchestrator inspects the redesigned `moai-foundation-thinking/SKILL.md`;

**Then** the body contains `ultrathink` or `Adaptive Thinking` (1 or more occurrences) and zero `sequential-thinking` literal occurrences (the BC marker uses title-cased "Sequential Thinking" without hyphen and is exempt).

**Verification commands**:

```bash
# Negative — no hyphenated sequential-thinking
grep -c 'sequential-thinking\|sequentialthinking' \
  .claude/skills/moai-foundation-thinking/SKILL.md
# Expected: 0

# Positive — ultrathink or Adaptive Thinking
grep -cE 'ultrathink|Adaptive Thinking' \
  .claude/skills/moai-foundation-thinking/SKILL.md
# Expected: 1+ (preserved from original body)

# BC marker uniqueness
grep -c 'Sequential Thinking MCP support retired' \
  .claude/skills/moai-foundation-thinking/SKILL.md
# Expected: 1 (exactly one BC marker line)
```

**Expected**: First command prints `0`. Second command prints `1` or more. Third command prints `1`.

---

### AC-STR-006 — New CI guard test PASS

**Given** M6 is complete and the new test file `internal/template/seq_thinking_retire_audit_test.go` exists;

**When** the orchestrator runs `go test -run TestSeqThinkingRetired ./internal/template/...`;

**Then** the test exits with status 0 and the sentinel `SEQ_THINKING_REINTRODUCED` is defined as a const or fail-message string in the source.

**Verification commands**:

```bash
# Test execution
go test -run TestSeqThinkingRetired ./internal/template/... ; echo "exit=$?"
# Expected: exit=0

# Sentinel defined
grep -n 'SEQ_THINKING_REINTRODUCED' internal/template/seq_thinking_retire_audit_test.go
# Expected: 1+ matches (sentinel constant or error message)
```

**Expected**: First command `exit=0`. Second command finds the sentinel.

**Failure injection sanity check** (optional run by orchestrator during M6 development): introduce a fake `sequential-thinking` literal to a temp test file, rerun the test, confirm it fails with the sentinel message, then revert.

---

### AC-STR-007 — Build + test net regressions = 0

**Given** M7 is complete and `baseline-test.txt` from M1 is on disk;

**When** the orchestrator reruns the full test suite and compares against the baseline;

**Then** zero NEW failures appear, where "NEW" means present in the post-M7 result but not in the M1 baseline.

**Verification commands**:

```bash
# Full test suite post-M7
go test ./... 2>&1 | tee post-m7-test.txt
echo "exit=$?"
# Compare baseline vs post-M7
diff <(grep -E '^(--- FAIL|FAIL\s+)' .moai/specs/SPEC-V3R6-SEQ-THINKING-RETIRE-001/baseline-test.txt | sort -u) \
     <(grep -E '^(--- FAIL|FAIL\s+)' post-m7-test.txt | sort -u)
# Expected: no diff lines starting with '>' (no NEW failures)

# Build
make build ; echo "exit=$?"
# Expected: exit=0
```

**Expected**: diff produces no `>`-prefixed lines (no new failures). `make build` exits 0.

---

### AC-STR-008 — catalog.yaml hashes regenerated

**Given** M7 is complete;

**When** the orchestrator inspects `catalog.yaml` modification time and content;

**Then** the catalog file has been touched in M7 and its content reflects post-edit hashes for the 14 modified agents + 6 modified skills.

**Verification commands**:

```bash
# File modification time post-M1
ls -la .claude/skills/catalog.yaml internal/template/templates/.claude/skills/catalog.yaml
# Expected: mtime within the M7 commit window

# Hash entries exist for retired skill
grep -A2 'moai-foundation-thinking' .claude/skills/catalog.yaml | head -5
# Expected: skill entry exists with a hash that differs from M1 baseline

# Sanity: catalog parses as YAML
python3 -c "import yaml; yaml.safe_load(open('.claude/skills/catalog.yaml'))" ; echo "exit=$?"
# Expected: exit=0
```

**Expected**: catalog.yaml has been regenerated. YAML parses cleanly.

---

### AC-STR-009 — 14 agent frontmatter tools tokens zero

**Given** M2 is complete;

**When** the orchestrator runs per-file grep on each of the 14 agent files (local + template = 28 files);

**Then** each file produces zero matches.

**Verification command**:

```bash
for f in \
  .claude/agents/core/manager-{develop,docs,project,quality,spec,strategy}.md \
  .claude/agents/expert/expert-{backend,devops,frontend,refactoring,security}.md \
  .claude/agents/meta/{builder-harness,evaluator-active,plan-auditor}.md \
  internal/template/templates/.claude/agents/core/manager-{develop,docs,project,quality,spec,strategy}.md \
  internal/template/templates/.claude/agents/expert/expert-{backend,devops,frontend,refactoring,security}.md \
  internal/template/templates/.claude/agents/meta/{builder-harness,evaluator-active,plan-auditor}.md \
; do
  count=$(grep -c 'mcp__sequential-thinking__sequentialthinking' "$f" 2>/dev/null || echo "MISSING")
  if [ "$count" != "0" ]; then
    printf "FAIL: %s -> %s\n" "$f" "$count"
  fi
done
echo "DONE"
```

**Expected**: no `FAIL:` lines printed (only `DONE` at the end).

---

### AC-STR-010 — 6 skill frontmatter allowed-tools tokens zero

**Given** M3 and M4 are complete;

**When** the orchestrator runs per-file grep on each of the 6 skill files (local + template = 12 files);

**Then** each file produces zero matches in its `allowed-tools` field.

**Verification command**:

```bash
for f in \
  .claude/skills/moai-foundation-cc/reference/claude-code-settings-official.md \
  .claude/skills/moai-foundation-cc/reference/skill-formatting-guide.md \
  .claude/skills/moai-foundation-core/modules/agents-reference.md \
  .claude/skills/moai-foundation-thinking/SKILL.md \
  .claude/skills/moai/workflows/plan/clarity-interview.md \
  .claude/skills/moai/workflows/run/context-loading.md \
  internal/template/templates/.claude/skills/moai-foundation-cc/reference/claude-code-settings-official.md \
  internal/template/templates/.claude/skills/moai-foundation-cc/reference/skill-formatting-guide.md \
  internal/template/templates/.claude/skills/moai-foundation-core/modules/agents-reference.md \
  internal/template/templates/.claude/skills/moai-foundation-thinking/SKILL.md \
  internal/template/templates/.claude/skills/moai/workflows/plan/clarity-interview.md \
  internal/template/templates/.claude/skills/moai/workflows/run/context-loading.md \
; do
  count=$(grep -c 'mcp__sequential-thinking__sequentialthinking' "$f" 2>/dev/null || echo "MISSING")
  if [ "$count" != "0" ]; then
    printf "FAIL: %s -> %s\n" "$f" "$count"
  fi
done
echo "DONE"
```

**Expected**: no `FAIL:` lines printed.

---

### AC-STR-011 — Cross-platform builds pass

**Given** M7 is complete;

**When** the orchestrator builds for darwin/linux/windows;

**Then** each build exits 0.

**Verification commands** (issued in single multi-Bash turn for parallel execution per `.claude/rules/moai/workflow/verification-batch-pattern.md`):

```bash
GOOS=darwin GOARCH=amd64 go build -o /tmp/moai-darwin ./cmd/moai && rm /tmp/moai-darwin && echo "darwin OK"
GOOS=linux GOARCH=amd64 go build -o /tmp/moai-linux ./cmd/moai && rm /tmp/moai-linux && echo "linux OK"
GOOS=windows GOARCH=amd64 go build -o /tmp/moai.exe ./cmd/moai && rm /tmp/moai.exe && echo "windows OK"
```

**Expected**: three `OK` lines printed.

---

### AC-STR-012 — CLAUDE.md §12 retirement statement

**Given** M6 is complete;

**When** the orchestrator greps CLAUDE.md for the retirement statement;

**Then** at least one line contains both the phrase "Sequential Thinking MCP retired" (or equivalent retirement wording) and the SPEC ID reference.

**Verification command**:

```bash
grep -E 'Sequential Thinking MCP retired|retired in SPEC-V3R6-SEQ-THINKING-RETIRE-001' CLAUDE.md
# Expected: 1+ matches

# Section anchor verification
grep -n '^## 12\.' CLAUDE.md
# Expected: matches the §12 MCP Servers section heading
```

**Expected**: First command finds the retirement statement. Second command confirms the §12 anchor exists.

---

## §3. Definition of Done

All 12 acceptance criteria above produce their expected output. The orchestrator records evidence (command + output) in `progress.md` under each AC heading. spec.md status transitions from `draft` to `implemented` and version bumps from `0.1.0` to `0.2.0`.

Run-phase merge is gated on:
- All 12 ACs PASS.
- Net new test regressions = 0 (per AC-STR-007).
- Net new lint regressions = 0 (per `golangci-lint run` in M7 verification batch).
- C-HRA-008 sentinel grep returns zero matches in `internal/harness/` and `internal/hook/` (per M7 verification batch item 3).
- `SEQ_THINKING_REINTRODUCED` sentinel is grep-discoverable in `internal/template/seq_thinking_retire_audit_test.go`.

## §4. Edge Cases and Risk Verification

### EC1 — BC marker false positive

The M4 BC marker uses title-cased "Sequential Thinking" (no hyphen). Confirm via:

```bash
grep -n 'Sequential Thinking' .claude/skills/moai-foundation-thinking/SKILL.md
# Expected: exactly 1 line (the BC marker)
grep -n 'sequential-thinking' .claude/skills/moai-foundation-thinking/SKILL.md
# Expected: 0 lines (hyphenated form fully removed)
```

The CI guard's allow-list pattern (M6 artifact) MUST be a substring match on the full BC marker line text, not a regex generalization, to prevent allow-list creep.

### EC2 — settings.local.json runtime overwrite

`settings.local.json` is runtime-managed per CLAUDE.local.md §2. If a future `moai cg`/`moai glm` invocation rewrites `settings.local.json`, it should not re-add sequential-thinking entries. Verify by inspecting `internal/cli/glm_tools.go` and `internal/cli/cg.go` for any literal `sequential-thinking` references:

```bash
grep -rn 'sequential-thinking\|sequentialthinking' internal/cli/ ; echo "exit=$?"
# Expected: exit=1
```

This grep is included in the AC-STR-001 global scope. If runtime tooling references the literal anywhere in `internal/cli/`, it must also be cleaned (silent expansion of scope — manager-develop has authority to clean within Tier M boundaries).

### EC3 — Empty enabledMcpjsonServers after removal

After removing `"sequential-thinking"` from `enabledMcpjsonServers`, the array should still contain `"chrome-devtools"` and `"context7"`. Confirm:

```bash
python3 -c "import json; d=json.loads(open('.claude/settings.json').read()); print(d.get('enabledMcpjsonServers', []))"
# Expected: ['chrome-devtools', 'context7'] (or similar non-empty list)
```

Empty array is valid JSON but suggests over-removal — investigate if the result is `[]`.

### EC4 — Catalog hash regeneration tool absence

If `gen-catalog-hashes.go --all` does not exist in the repo at run-time, M7 falls back to manual hash regeneration via `sha256sum` per file. Verify the tool exists before M7:

```bash
ls -la internal/skill/gen-catalog-hashes.go 2>/dev/null || find . -name '*catalog*hash*' -type f
```

If absent, manager-develop logs the absence in M7 progress.md and uses the fallback procedure.

## §5. Test Inventory

The CI guard test added in M6 is the only new test introduced by this SPEC:

| Test name | File | Type | Purpose |
|-----------|------|------|---------|
| `TestSeqThinkingRetired` | `internal/template/seq_thinking_retire_audit_test.go` | Spec / namespace audit | Verify zero sequential-thinking literals outside allow-list |

Existing tests that may interact (regression watch):
- `TestNamespaceLeakMyHarnessSkills`, `TestNamespaceLeakHarnessAgentsDir` (precedent `d0782a365`) — sibling pattern.
- `TestTemplate*` family — confirms template still compiles after `.mcp.json.tmpl` and `settings.json.tmpl` edits.

## §6. Quality Gate Criteria

Mapped to TRUST 5:

| Pillar | Criterion | Verified by |
|--------|-----------|-------------|
| Tested | New CI guard test PASS + existing tests no new regressions | AC-STR-006, AC-STR-007 |
| Readable | BC marker prose is clear and self-documenting | M4 redesign review (manager-develop self-check) |
| Unified | Local + template byte-equivalent for paired changes | per-milestone diff inspection |
| Secured | No new external attack surface (MCP server removal reduces it) | trivially satisfied; M5 verifies no new entries |
| Trackable | Each milestone is one commit with SPEC ID in message | manager-develop commit log + M7 progress.md |

## §7. Post-Run-Phase Acceptance Sign-off

Manager-develop authors `progress.md` under this SPEC directory documenting:

1. M1–M7 commit SHAs in chronological order.
2. Per-AC PASS/FAIL with command output excerpts.
3. Baseline-vs-post-M7 test diff summary.
4. Cross-platform build evidence.
5. Catalog hash regeneration evidence.
6. Any deviations from this plan (with rationale).

When progress.md confirms all 12 ACs PASS and net regressions = 0, the SPEC transitions to `status: implemented` (M7 task 5).
