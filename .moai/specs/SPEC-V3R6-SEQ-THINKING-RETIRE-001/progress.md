# Progress — SPEC-V3R6-SEQ-THINKING-RETIRE-001

Run-phase Tier M complete. All 12 ACs PASS. Zero net new test or lint regressions over the M1 baseline.

## SHOULD-FIX 3 (plan-auditor iter 1) — Resolution

| # | Finding | Resolution |
|---|---------|------------|
| D1 | `plan.md:L128` references "AC-STR-003 (skill body literal check)" but body-literal coverage is owned by AC-STR-005 per acceptance.md §2 | Documented here: M4's body verification is AC-STR-005 (positive ultrathink + negative sequential-thinking literal). AC-STR-003 covers configuration files (settings.json + skill body for moai-foundation-thinking) — both pass at the M4/M5 boundary. No spec.md mutation needed. |
| D2 | Inventory count drift (spec.md §6 = 50, plan.md M2 task list = 51) | Reconciled by M1 baseline: **canonical 49 files** = 14 agents × 2 (28) + 5 skills × 2 + 1 thinking × 2 (12) + 1 rule × 2 (2) + 4 config files (.mcp.json local+template + settings.json local+template) + 1 CLAUDE.local.md + 2 new files (audit test + allow-list). `CLAUDE.md` template mirror at `internal/template/templates/CLAUDE.md` has zero sequential-thinking references already (no template edit needed). `.claude/settings.local.json` is runtime-managed per CLAUDE.local.md §2 [HARD] (skipped). |
| D3 | acceptance.md L94 `go build ./cmd/moai` is a template-build proxy only | M7 verification batch adds full `go test ./internal/template/...` which exercises rendering via `TestMCPTemplate*` family. Cross-platform builds via `GOOS=darwin\|linux\|windows go build ./cmd/moai` all exit 0. |

## M1 — Inventory and Baseline Capture

**Commit**: `0209c5e9f`
**Artifacts**:
- `.moai/specs/SPEC-V3R6-SEQ-THINKING-RETIRE-001/baseline.txt` (109 lines)
- `.moai/specs/SPEC-V3R6-SEQ-THINKING-RETIRE-001/baseline-test.txt`
- `.moai/specs/SPEC-V3R6-SEQ-THINKING-RETIRE-001/baseline-lint.txt`

### Pre-existing failures (NOT attributable to this SPEC)

| Test | Notes |
|------|-------|
| `TestRuleTemplateMirrorDrift/manager-develop-prompt-template.md` | source 8496 vs mirror 7180 bytes (pre-existing) |
| `TestRuleTemplateMirrorDrift/spec-workflow.md` | source 29363 vs mirror 26709 bytes (pre-existing) |
| `TestRuleTemplateMirrorDrift/plan-auditor.md` | source 21042 vs mirror 18778 bytes (pre-existing, mirror drift) |
| `TestSkillsContainPlanAuditGateMarkers/solo_run.md` | run.md missing Plan Audit Gate markers (pre-existing) |
| `TestRetirementCompletenessAssertion/manager-{tdd,ddd}_replacement_manager-develop_must_exist` | `.claude/agents/moai/manager-develop.md` not in embedded FS (pre-existing) |
| `TestAgentFrontmatterAudit`, `TestAllAgentsInCatalog`, `TestBackwardCompatibility`, `TestEmbeddedTemplates_AgentDefinitions`, `TestLateBranchTemplateMirror`, `TestLoadCatalog`, `TestLoadEmbeddedCatalog_Success` | Pre-existing — all captured in `baseline-test.txt`. |

**Lint baseline**: 27 issues (8 errcheck, 1 ineffassign, 5 staticcheck, 13 unused). Captured by re-running `golangci-lint run --timeout=2m` on the pre-edit tree via `git stash`.

### Catalog location correction

Plan.md M7 references `.claude/skills/catalog.yaml`, but the actual file is at `internal/template/catalog.yaml` (293 → 14317 bytes post-regen). Hash regeneration tool is `internal/template/scripts/gen-catalog-hashes.go --all`. M7 used the correct path.

## M2 — Agent Frontmatter Cleanup

**Commit**: `514717d33`
**Files**: 14 agents × 2 mirrors = **28 files**.

**Pattern**: removed `mcp__sequential-thinking__sequentialthinking[, ]` token from `tools:` field. Three variants handled:
- Most: `, mcp__sequential-thinking__sequentialthinking, mcp__context7__...` → `, mcp__context7__...` (12 agents)
- Last-in-list (`manager-quality`, `evaluator-active`): `, mcp__sequential-thinking__sequentialthinking` → `` (2 agents)
- Mid-list (`plan-auditor`): `Bash, mcp__sequential-thinking__sequentialthinking, Write, Edit` → `Bash, Write, Edit`

### AC-STR-009 verification

```bash
$ for f in .claude/agents/core/manager-{develop,docs,project,quality,spec,strategy}.md \
            .claude/agents/expert/expert-{backend,devops,frontend,refactoring,security}.md \
            .claude/agents/meta/{builder-harness,evaluator-active,plan-auditor}.md \
            internal/template/templates/.claude/agents/core/manager-{develop,docs,project,quality,spec,strategy}.md \
            internal/template/templates/.claude/agents/expert/expert-{backend,devops,frontend,refactoring,security}.md \
            internal/template/templates/.claude/agents/meta/{builder-harness,evaluator-active,plan-auditor}.md; do
    count=$(grep -c 'mcp__sequential-thinking__sequentialthinking' "$f" 2>/dev/null)
    [ "$count" != "0" ] && echo "FAIL: $f -> $count"
  done; echo DONE
DONE
```
**Result**: AC-STR-009 PASS — zero FAIL lines.

## M3 — Skill Frontmatter + Body Cleanup (5 skills)

**Commit**: `856a85ee9`
**Files**: 5 skills × 2 mirrors = **10 files**.

**Per-skill changes**:
- `moai-foundation-cc/reference/claude-code-settings-official.md`: removed `sequential-thinking` MCP server example block from `mcpServers` example + permission example.
- `moai-foundation-cc/reference/skill-formatting-guide.md`: removed `mcp__sequential-thinking__*` from multi-MCP `allowed-tools` example.
- `moai-foundation-core/modules/agents-reference.md`: removed `mcp-sequential-thinking` from agent category table example + agent catalog table row.
- `moai/workflows/plan/clarity-interview.md` and `moai/workflows/run/context-loading.md`: removed "Note: Sequential Thinking MCP remains available" line — deep-reasoning path now consolidated on `ultrathink` keyword per REQ-STR-002.

### AC-STR-010 partial verification (5 of 6 skills, M3 scope only)

```bash
$ grep -c 'sequential-thinking\|sequentialthinking' \
    .claude/skills/moai-foundation-cc/reference/claude-code-settings-official.md \
    .claude/skills/moai-foundation-cc/reference/skill-formatting-guide.md \
    .claude/skills/moai-foundation-core/modules/agents-reference.md \
    .claude/skills/moai/workflows/plan/clarity-interview.md \
    .claude/skills/moai/workflows/run/context-loading.md
0
0
0
0
0
```
**Result**: AC-STR-010 partial PASS (final 6th skill in M4).

## M4 — moai-foundation-thinking Redesign + BC Marker

**Commit**: `7dc601c38`
**Files**: `SKILL.md` local + template mirror = **2 files**.

**Changes**:
1. **Frontmatter**: removed `mcp__sequential-thinking__sequentialthinking` from `allowed-tools:`. Replaced `sequential-thinking` token in `tags:` with `adaptive-thinking`.
2. **Description**: changed "Sequential Thinking MCP (absorbed from moai-workflow-thinking)" → "Adaptive Thinking via the `ultrathink` keyword".
3. **Body — Adaptive Thinking section**: removed three stranded numbered list items (lines 303-305 in pre-edit file) that referenced `branchFromThought` / `nextThoughtNeeded` / "Adaptive Thinking handles reasoning automatically" — these were leftover sequential-thinking guidance. Replaced with a clean prose paragraph that documents the activation pattern (orchestrator prepends `ultrathink` keyword + Opus 4.7+ Adaptive Thinking handles depth).
4. **BC marker**: added a new `## Backward Compatibility Marker` section at the file end with verbatim text: "Sequential Thinking MCP support retired in SPEC-V3R6-SEQ-THINKING-RETIRE-001. For deep reasoning use the `ultrathink` keyword (triggers Opus 4.7+ Adaptive Thinking). The earlier 'Sequential Thinking MCP (absorbed from moai-workflow-thinking)' content is superseded; this skill's deep-reasoning path is consolidated onto ultrathink + the creative frameworks above."

**PRESERVE**: Critical Evaluation, Diverge-Converge, Deep Questioning sections unchanged. First Principles absorbed content (`Five-Phase Process`) unchanged.

### AC-STR-005 verification

```bash
$ grep -c 'sequential-thinking\|sequentialthinking' .claude/skills/moai-foundation-thinking/SKILL.md
0

$ grep -cE 'ultrathink|Adaptive Thinking' .claude/skills/moai-foundation-thinking/SKILL.md
8

$ grep -c 'Sequential Thinking MCP support retired' .claude/skills/moai-foundation-thinking/SKILL.md
1
```
**Result**: AC-STR-005 PASS (0 hyphenated, 8 ultrathink/Adaptive Thinking, 1 BC marker line — exactly one).

## M5 — Settings + .mcp.json + Rule Cleanup

**Commit**: `2ba0d7cf3`
**Files**: 6 files.

**Per-file changes**:
1. `.mcp.json` (local): removed `"sequential-thinking"` MCP server entry block (5 lines). chrome-devtools + context7 + zai-mcp-server retained.
2. `internal/template/templates/.mcp.json.tmpl`: removed `"sequential-thinking"` block including `{{- if eq .Platform "windows"}}` cross-platform command branches (11 lines). context7 + moai-lsp retained.
3. `.claude/settings.json` (local): removed `"mcp__sequential-thinking__*"` from `permissions.allow` array.
4. `internal/template/templates/.claude/settings.json.tmpl`: removed same `allow` entry + removed `"sequential-thinking"` from `enabledMcpjsonServers` array. Now contains `["chrome-devtools", "context7"]`.
5. `.claude/rules/moai/core/settings-management.md`: removed `- sequential-thinking: Complex problem analysis` from Standard MCP servers list. Added retirement note `> Sequential Thinking MCP was retired in SPEC-V3R6-SEQ-THINKING-RETIRE-001. Use the ultrathink keyword (Adaptive Thinking on Opus 4.7+) for deep reasoning.` Renamed "Sequential Thinking Usage" subsection → "Adaptive Thinking Usage" with refined prose.
6. `internal/template/templates/.claude/rules/moai/core/settings-management.md`: byte-identical mirror.

**Note on `settings.local.json`**: per CLAUDE.local.md §2 [HARD] runtime-managed; intentionally NOT modified. Future `moai cg`/`moai glm` invocations will not re-add sequential-thinking entries because the template no longer carries them.

### AC-STR-002 / AC-STR-003 verification

```bash
$ grep -n 'sequential-thinking\|sequentialthinking' .mcp.json internal/template/templates/.mcp.json.tmpl
(no output — exit 1)

$ grep -n 'mcp__sequential-thinking\|"sequential-thinking"' .claude/settings.json internal/template/templates/.claude/settings.json.tmpl
(no output — exit 1)

$ jq . .mcp.json > /dev/null && echo OK
OK

$ jq . .claude/settings.json > /dev/null && echo OK
OK
```
**Result**: AC-STR-002 PASS. AC-STR-003 PASS.

## M6 — CI Guard + Documentation Updates

**Commit**: `57e00d6b3`
**Files**: 4 files (2 new + 2 modified).

**New files**:
- `internal/template/seq_thinking_retire_audit_test.go` (260 lines): `TestSeqThinkingRetired` walks `.claude/{agents,skills,rules}` + template mirrors + `.mcp.json` + `.claude/settings.json` + `internal/template/templates/.mcp.json.tmpl` + `internal/template/templates/.claude/settings.json.tmpl` + `CLAUDE.md`. Pattern: `sequential-thinking|sequentialthinking` regex (case-sensitive, hyphenated form). On violation emits sentinel `SEQ_THINKING_REINTRODUCED` with file:line list. Allow-list loaded from `seq_thinking_retire_audit_allowlist.txt` (substring match).
- `internal/template/seq_thinking_retire_audit_allowlist.txt`: registers two permitted substrings (`retired in SPEC-V3R6-SEQ-THINKING-RETIRE-001` and `SPEC-V3R6-SEQ-THINKING-RETIRE-001`) — these cover the retirement notes in CLAUDE.md §12, CLAUDE.local.md §22.2, settings-management.md (local + template), moai-foundation-thinking BC marker, and progress.md cross-references.

**Documentation updates**:
- `CLAUDE.md` §12: appended retirement statement (`> Sequential Thinking MCP retired in SPEC-V3R6-SEQ-THINKING-RETIRE-001. The ultrathink keyword (Adaptive Thinking on Opus 4.7+) is the canonical deep-reasoning path. Users who require Sequential Thinking for personal workflow may install it independently via ~/.claude/settings.json (user-local scope); the project namespace remains clean.`).
- `CLAUDE.local.md` §22.2: corrected the alwaysLoad reference (was "mcp__context7 / mcp__sequential-thinking", now "mcp__context7" only) and appended Korean retirement note.

### AC-STR-004 / AC-STR-006 / AC-STR-012 verification

```bash
$ go test -run TestSeqThinkingRetired ./internal/template/...
--- PASS: TestSeqThinkingRetired (0.08s)
PASS
ok      github.com/modu-ai/moai-adk/internal/template   0.534s

$ grep -n 'SEQ_THINKING_REINTRODUCED' internal/template/seq_thinking_retire_audit_test.go
16:// Sentinel: SEQ_THINKING_REINTRODUCED
26:const SeqThinkingSentinel = "SEQ_THINKING_REINTRODUCED"
182:// On failure, the test emits the SEQ_THINKING_REINTRODUCED sentinel

$ grep -n 'SPEC-V3R6-SEQ-THINKING-RETIRE-001' CLAUDE.md
403:> Sequential Thinking MCP retired in SPEC-V3R6-SEQ-THINKING-RETIRE-001. ...

$ grep -n 'mcp__sequential-thinking' CLAUDE.local.md
(no output — exit 1)
```
**Result**: AC-STR-004 PASS, AC-STR-006 PASS (test PASS + sentinel grep-discoverable), AC-STR-012 PASS.

## M7 — Validation + Catalog Regeneration + Status

**Commit**: (this commit)
**Files**: 4 files.

### M7 tasks completed

1. **Catalog hashes regenerated**: `go run ./internal/template/scripts/gen-catalog-hashes.go -all` updated `internal/template/catalog.yaml` (293 → final size 14317 bytes). All 14 modified agents + 6 modified skills (including retired moai-foundation-thinking redesign) have refreshed hashes.

2. **Pre-existing test suite update**: 3 tests in `internal/template/settings_test.go` (`TestMCPTemplateRequiredServers`, `TestMCPTemplateExistingFieldsPreserved`, `TestMCPTemplateAlwaysLoadOnSequentialThinking`) hard-coded `sequential-thinking` as a required MCP server. Updated:
   - `TestMCPTemplateRequiredServers`: removed `sequential-thinking` from `requiredServers` list and added an inverted assertion that `sequential-thinking` MUST NOT be present (retirement enforcement).
   - `TestMCPTemplateExistingFieldsPreserved`: removed the block that asserted sequential-thinking command/args presence; replaced with a comment referencing the retirement SPEC.
   - `TestMCPTemplateAlwaysLoadOnSequentialThinking`: renamed to `TestMCPTemplateSequentialThinkingRetired` and inverted the assertion to fail when sequential-thinking server entry is present.

3. **Spec.md status**: `status: draft` → `status: implemented`, `version: 0.1.0` → `version: 0.2.0`.

### M7 verification batch (parallel multi-Bash)

```text
1) go test ./...                                   → 10 pre-existing failures (zero NEW)
2) go test -coverprofile=cover.out ./internal/template/...  → coverage 85.1%
3) C-HRA-008 sentinel grep                         → 0 violations
4) Sentinel-key audit                              → SEQ_THINKING_REINTRODUCED + NAMESPACE_LEAK_* discoverable
5) go run ./cmd/moai --version                     → moai-adk v3.0.0-rc1
6) Lint baseline                                   → 27 issues (zero NEW over baseline)
7) Cross-platform build                            → darwin OK, linux OK, windows OK
```

### AC-STR-001 / AC-STR-007 / AC-STR-008 / AC-STR-011 verification

```bash
$ grep -rln 'sequential-thinking\|sequentialthinking' \
    .claude/agents/ .claude/skills/ .claude/rules/ \
    internal/template/templates/.claude/ CLAUDE.md
(no output)

$ go test ./internal/template/... 2>&1 | grep -E '^--- FAIL: Test' | sort -u | diff - \
    <(grep -E '^--- FAIL: Test' .moai/specs/SPEC-V3R6-SEQ-THINKING-RETIRE-001/baseline-test.txt | sort -u)
(only microsecond timing differences — zero NEW failures)

$ wc -c internal/template/catalog.yaml
14317 internal/template/catalog.yaml

$ GOOS=darwin GOARCH=amd64 go build -o /tmp/moai-darwin ./cmd/moai && rm /tmp/moai-darwin && echo darwin OK
darwin OK
$ GOOS=linux GOARCH=amd64 go build -o /tmp/moai-linux ./cmd/moai && rm /tmp/moai-linux && echo linux OK
linux OK
$ GOOS=windows GOARCH=amd64 go build -o /tmp/moai.exe ./cmd/moai && rm /tmp/moai.exe && echo windows OK
windows OK
```
**Result**: AC-STR-001 PASS, AC-STR-007 PASS, AC-STR-008 PASS, AC-STR-011 PASS.

## Final AC Matrix

| AC | Status | Owner Milestone | Evidence |
|----|--------|-----------------|----------|
| AC-STR-001 | **PASS** | M2+M3+M4+M5+M6 cumulative | Global grep returns zero matches in canonical scope |
| AC-STR-002 | **PASS** | M5 | .mcp.json + .mcp.json.tmpl zero literals; jq validates |
| AC-STR-003 | **PASS** | M4+M5 | settings.json + settings.json.tmpl + moai-foundation-thinking SKILL.md zero literals |
| AC-STR-004 | **PASS** | M6 | CLAUDE.md retirement statement + CLAUDE.local.md §22.2 correction |
| AC-STR-005 | **PASS** | M4 | 0 hyphenated, 8 ultrathink/Adaptive Thinking, 1 BC marker |
| AC-STR-006 | **PASS** | M6 | TestSeqThinkingRetired PASS + sentinel grep-discoverable |
| AC-STR-007 | **PASS** | M7 | zero NEW test regressions, zero NEW lint over M1 baseline, make build OK |
| AC-STR-008 | **PASS** | M7 | catalog.yaml regenerated (14317 bytes), parses cleanly |
| AC-STR-009 | **PASS** | M2 | All 28 agent files (14 local + 14 template) zero matches |
| AC-STR-010 | **PASS** | M3+M4 | All 12 skill files (6 local + 6 template) zero matches |
| AC-STR-011 | **PASS** | M7 | darwin + linux + windows builds OK |
| AC-STR-012 | **PASS** | M6 | CLAUDE.md §12 contains "Sequential Thinking MCP retired in SPEC-V3R6-SEQ-THINKING-RETIRE-001" |

**12 of 12 ACs PASS.** Net new regressions = 0.

## Commit SHA list

| Milestone | SHA | Subject |
|-----------|-----|---------|
| M1 | `0209c5e9f` | chore(SPEC-V3R6-SEQ-THINKING-RETIRE-001): M1 inventory + progress baseline |
| M2 | `514717d33` | refactor(agents): M2 sequential-thinking tool removal — 14 agents (28 files) |
| M3 | `856a85ee9` | refactor(skills): M3 sequential-thinking removal — 5 skills (10 files) |
| M4 | `7dc601c38` | refactor(moai-foundation-thinking): M4 ultrathink-centric redesign + BC marker |
| M5 | `2ba0d7cf3` | chore(config): M5 .mcp.json + settings.json + settings-management.md 정리 |
| M6 | `57e00d6b3` | feat(template): M6 SEQ_THINKING_REINTRODUCED CI guard + allow-list + docs sync |
| M7 | (this commit) | chore(SPEC-V3R6-SEQ-THINKING-RETIRE-001): M7 mark implemented v0.2.0 + catalog hash refresh |

## Deviations from plan.md

- **Plan.md M2 task list** referenced `.claude/agents/{core,expert,meta}` paths plus three additional patterns. Inventory confirms exactly 14 agents, as expected.
- **Plan.md M5 task 5** mentioned editing `.claude/settings.local.json`. Skipped per CLAUDE.local.md §2 [HARD] runtime-managed contract. Future runtime invocations of `moai cg`/`moai glm` will not re-add sequential-thinking because the source template no longer carries them.
- **Plan.md M7 task 1** referenced `gen-catalog-hashes.go --all` from a path that did not exist. Resolved to `internal/template/scripts/gen-catalog-hashes.go -all` (no leading `.claude/skills/` and single-dash flag).
- **Plan.md M7 task 2** seven-item verification batch added an 8th step: updating three pre-existing settings_test.go MCP template unit tests that hard-coded sequential-thinking as a required server. These were not in scope at plan-time but became failing tests once the .mcp.json.tmpl block was removed. Per Section D EXTEND scope, the tests were updated to reflect the post-retirement state (one renamed `TestMCPTemplateAlwaysLoadOnSequentialThinking` → `TestMCPTemplateSequentialThinkingRetired` with inverted assertion).

## CLAUDE.md template mirror

`internal/template/templates/CLAUDE.md` already had zero sequential-thinking references before this SPEC. The retirement note was added only to the local `CLAUDE.md` (M6) since the template version is not the source of the §12 section content for user projects (Template-First Rule does not apply here — the template `CLAUDE.md` is a separate user-facing artifact with simpler content).

## Milestone status (final)

- [x] M1 — Inventory + baseline ✓ `0209c5e9f`
- [x] M2 — Agent frontmatter cleanup (28 files) ✓ `514717d33`
- [x] M3 — Skill frontmatter + body cleanup (10 files, 5 skills × 2) ✓ `856a85ee9`
- [x] M4 — moai-foundation-thinking redesign + BC marker (2 files) ✓ `7dc601c38`
- [x] M5 — Settings + .mcp.json + rules cleanup (6 files) ✓ `2ba0d7cf3`
- [x] M6 — CI guard + docs (4 files) ✓ `57e00d6b3`
- [x] M7 — Validation + catalog regeneration + status: implemented ✓ (this commit)
