---
id: SPEC-V3R6-AGENT-FOLDER-SPLIT-001
title: "Progress — Agent Folder Split"
version: "0.2.0"
status: implemented
created: 2026-05-22
updated: 2026-05-22
author: manager-develop
priority: Medium
phase: "v3.0.0"
module: ".claude/agents, internal/template/templates/.claude/agents, internal/template/catalog.yaml"
lifecycle: spec-anchored
tags: "agent, folder-restructure, progress, wave-2"
tier: M
---

# Run-Phase Progress — SPEC-V3R6-AGENT-FOLDER-SPLIT-001

## Run-Phase Completion Summary

- **Branch**: `main` (1인 OSS Hybrid Trunk policy, CLAUDE.local.md §23, no L2/L3 worktree)
- **Commits ahead of origin/main**: 3 (this SPEC) + 4 unrelated parallel session commits before run
- **HEAD before run**: `58a235e06` (plan commit)
- **HEAD after run**: `a912e7d5a` (Commit 3)
- **Sprint pattern**: M1+M2 → Commit 1 (1bd083725) → M3 → Commit 2 (fdf325eb6) → M4 discoveries → Commit 3 (a912e7d5a)
- **Wall-time**: single sprint, single session

## Pre-Flight Baseline (Section C results)

| Check | Expected | Actual |
|---|---|---|
| Cross-platform build (darwin) | exit 0 | exit 0 |
| Cross-platform build (windows) | exit 0 | exit 0 |
| Local `.claude/agents/moai/*.md` | 19 | 19 |
| Template `internal/template/templates/.claude/agents/moai/*.md` | 19 | 19 |
| Out-of-scope Go baseline (9 files) | 42 | 42 |
| In-scope Go baseline (21 files) | 41 (occurrences) | 40 (lines, 41 occurrences — line 130 of agent_lint.go has 2 matches) |
| Active rule/skill baseline | 22 | 22 |
| catalog.yaml entries (templates/.claude/agents/moai/) | 19 | 19 |

## Run-Phase Actions Taken

### M1: Folder Creation + File Movement (38 git mv)

- Created 6 new folders: `.claude/agents/{core,expert,meta}` + `internal/template/templates/.claude/agents/{core,expert,meta}`
- 19 local agent files renamed via `git mv` (8 core + 6 expert + 5 meta)
- 19 template mirror files renamed via `git mv`
- Empty `.claude/agents/moai` and `internal/template/templates/.claude/agents/moai` removed via `rmdir`
- 38 R (rename) entries verified via `git status`

### M2: Cross-Reference Update (Commit 1, SHA 1bd083725)

#### M2.1 Active rule + skill files (22 occurrences → 0)

10 local + 9 template (release.md is dev-only per CLAUDE.local.md §21, no template mirror):

| File | Edits |
|---|---|
| `.claude/rules/moai/development/agent-authoring.md` | 1 (line 15: moai → {core,expert,meta,harness}) |
| `.claude/rules/moai/development/model-policy.md` | 2 (lines 45, 46: plan-auditor → meta/, package agents → {core,expert,meta}) |
| `.claude/rules/moai/workflow/spec-workflow.md` | 2 (lines 45, 137: manager-git → core/, plan-auditor → meta/) |
| `.claude/rules/moai/workflow/team-protocol.md` | 1 (line 114: moai → {core,expert,meta,harness}) |
| `.claude/skills/moai-harness-learner/SKILL.md` | 1 (line 151) |
| `.claude/skills/moai-meta-harness/SKILL.md` | 1 (line 352) |
| `.claude/skills/moai/workflows/harness.md` | 1 (line 175) |
| `.claude/skills/moai/workflows/plan/spec-assembly.md` | 1 (line 308: manager-git → core/) |
| `.claude/skills/moai/workflows/project/meta-harness.md` | 1 (line 227) |
| `.claude/skills/moai/workflows/release.md` | 1 (line 878: manager-spec → core/) |
| + 9 template mirrors of all above (excluding release.md) | 10 |
| **Total** | **22 → 0** |

#### M2.2 In-scope Go code (41 occurrences in 21 files → 0)

All 21 files in SPEC enumeration updated. Notable cases:
- `internal/cli/agent_lint.go` (2): docstring + flag description
- `internal/cli/update.go` (1): `isMoaiManaged()` extended with `agents` case recognizing core/expert/meta/harness as MoAI-managed
- `internal/hook/subagent_start.go` (1): `loadAgentFrontmatter()` search order extended to 4 domain subfolders + legacy fallback
- `internal/hook/agent_start_test.go` (2): test fixtures now write to `.claude/agents/core/`
- `internal/research/safety/frozen.go` (1): researcher.md → meta/
- `internal/template/manager_develop_present_test.go` (3): TestPurgedZombieAgentsAbsent now iterates all 4 domains for regression detection

#### M2.3 catalog.yaml (19 entries → 0)

All 19 path entries updated from `agents/moai/` to `agents/{core,expert,meta}/` per §3 classification matrix. File hashes preserved via `git mv` byte-identity.

#### M2.4 Discovered additional walker callers (Commit 3, SHA a912e7d5a — M4 discovery)

6 files NOT in plan-phase enumeration but identified during M4 test verification as having the same walker/ReadDir root cause:

| File | Reason |
|---|---|
| `internal/core/project/initializer.go` | `claudeDirs` initialization list referenced legacy `agents/moai` |
| `internal/core/project/validator.go` | `requiredClaudeDirs` validation list referenced legacy `agents/moai` |
| `internal/template/model_policy.go` | `ApplyEffortPolicy` + `ApplyModelPolicy` walked single `.claude/agents/moai` dir — now iterates 4 domain subfolders with per-domain manifest tracking |
| `internal/template/contract_schema_test.go` | 3 tests now target `core/` for manager-quality + iterate all 4 subfolders for backwards-compat |
| `internal/template/embed_test.go` | `TestEmbeddedTemplates_AgentDefinitions` now walks 4 subfolders |
| `internal/template/model_policy_test.go` | 11 fixture paths updated to match agent name → domain (manager-* → core/, expert-* → expert/, plan-auditor → meta/, unknown → core/) |

### M3: catalog.yaml hash regeneration (Commit 2, SHA fdf325eb6)

- `make build` executed → `gen-catalog-hashes.go --all` regenerated 2 skill content hashes (moai-harness-learner + moai-meta-harness SKILL.md were edited)
- 19 agent file hashes UNCHANGED — `git mv` preserved byte-identical agent content; hash is content-based, not path-based
- `TestManifestHashFormat` PASS (verified via `go test ./internal/template/...`)

### M4: Cross-platform build + Test verification

- `go build ./...` → exit 0
- `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
- `go test ./...` → see Test Classification below

## 12 AC Binary Matrix

| AC | Status | Verification Command | Actual Output |
|---|---|---|---|
| AC-AFS-001 | PASS | `find .claude/agents -maxdepth 1 -type d \| sort` | 5 lines exact match: `.claude/agents`, `.claude/agents/core`, `.claude/agents/expert`, `.claude/agents/harness`, `.claude/agents/meta` |
| AC-AFS-002 | PASS | per-folder counts (local + template) | 8/6/5/4 both sides |
| AC-AFS-003 | PASS | `git log --follow --format="%h %s" -2 .claude/agents/core/manager-spec.md` | 2 commits returned (1bd083725 split commit + 81d42a1ae previous edit visible via --follow through rename) |
| AC-AFS-004 | PASS | `test ! -d .claude/agents/moai && test ! -d internal/template/templates/.claude/agents/moai` | PASS |
| AC-AFS-005 | PASS | sample 5 frontmatter inspections | 5/5 agents `name:` field matches basename (manager-spec, expert-backend, builder-harness, plan-auditor, manager-develop) |
| AC-AFS-006 | PASS | in-scope Go grep (21 files filter) | 0 (was 41) |
| AC-AFS-007 | PASS | active rule/skill grep | 0 (was 22) |
| AC-AFS-008 | PARTIAL | `diff -rq` 4 folders | 3/4 folders zero-drift; meta/plan-auditor.md inherits pre-existing template-mirror drift (21042B local vs 18778B template, baseline pre-existing per `project_v3r6_template_mirror_drift_audit_2026_05_22`) |
| AC-AFS-009 | PASS | catalog.yaml path counts | 0 old, 19 new (was 19/0) |
| AC-AFS-010 | PASS | harness/ name list | 4 names match expected: cli-template-specialist, hook-ci-specialist, quality-specialist, workflow-specialist |
| AC-AFS-011 | PASS | dirty tree PRESERVE count | 10 files (≥10 expected) |
| AC-AFS-012 | PASS | `git diff origin/main -- <9 out-of-scope files> \| wc -l` | 0 (byte-identical) + 42 preserved occurrences |

**AC-AFS-008 partial classification rationale**: SPEC acceptance.md §AC-AFS-008 Note explicitly excludes pre-existing template-mirror drift from this SPEC's scope: "pre-existing baseline drift (plan-auditor.md / hooks-system.md 등 10 files per `project_v3r6_template_mirror_drift_audit_2026_05_22`) is out of this SPEC's scope. AC-AFS-008 verifies that **agents/** specifically has zero drift after run-phase." The drift inherited from the `git mv` of `plan-auditor.md` (moai/ → meta/) is the SAME pre-existing drift (file content was already drifted before rename). 3 of 4 folders (core/, expert/, harness/) have ZERO drift. The plan-auditor.md drift is also independently captured by `TestRuleTemplateMirrorDrift` baseline FAIL.

## Test Classification (acceptance.md §Quality Gate)

### Category A: Pre-existing baseline 3 FAIL (preserved per spec.md §10.4)

| Test | Reason |
|---|---|
| `TestRuleTemplateMirrorDrift` (manager-develop-prompt-template.md / plan-auditor.md / spec-workflow.md) | Documented baseline drift in 10 files (template mirror parity unrelated to this SPEC scope) |
| `TestLateBranchTemplateMirror/spec-assembly.md` | Same baseline drift category |
| `TestSkillsContainPlanAuditGateMarkers` | Pre-existing missing markers in run.md (baseline) |

### Category B: Inherited unrelated baseline 1 FAIL

| Test | Reason |
|---|---|
| `TestRenderPRSegment_Absence/pr_present_+_unset_(legacy_backward_compat)` | Pre-existing per `project_v3r5_statusline_fmc_001_run_complete` — statusline domain, not agent-folder scope |

### Category C: Out-of-scope deferred residual 3 FAIL (NEW post-implementation, REQ-AFS-010 scope reduction)

These are NEW failures caused by walker logic in OUT-OF-SCOPE files (per REQ-AFS-010 / spec.md §6.2.2). Acceptable per acceptance.md §Quality Gate "NEW out-of-scope deferred FAIL":

| Test | Source File (out-of-scope) | Failure Reason |
|---|---|---|
| `TestAllAgentsInCatalog` | `internal/template/catalog_tier_audit_test.go:204` | `fs.WalkDir(fsys, ".claude/agents/moai", ...)` returns 0 files post-rename → `len(diskAgents) != 19` |
| `TestAgentFrontmatterAudit` | `internal/template/agent_frontmatter_audit_test.go:88` | 3 WalkDir calls + per-agent fs.Stat checks on `.claude/agents/moai/` |
| `TestRetirementCompletenessAssertion` | `internal/template/agent_frontmatter_audit_test.go:159, 172` | Hardcoded `.claude/agents/moai/manager-develop.md` retirement-replacement path expectations (the literal is in the OUT-OF-SCOPE per-agent fixture array at line 257-365) |

### Category D: NEW regression introduced by this SPEC

**0 NEW regressions.** All test failures are categorized into A/B/C above.

### Follow-up SPEC for Category C resolution

**Provisional**: `SPEC-V3R6-FROZEN-PREFIX-REALIGN-001` (가칭, 본 SPEC implemented 후 별도 plan-phase). Scope:
- 9 out-of-scope Go files (REQ-AFS-010 PRESERVE list)
- `pre_tool.go:720` SentinelHarnessFrozenAgent table semantic supersession decision (semantic legacy vs multi-prefix support)
- `catalog_tier_audit_test.go` + `agent_frontmatter_audit_test.go` walker logic restructure to walk 3 new folders OR consolidate via fs.Glob pattern
- `frozen_guard_test.go` expected-array semantic alignment
- Estimated change: ~17 LOC + walker logic restructuring

## Cross-Platform Build Evidence

```
$ go build ./...                            → exit 0
$ GOOS=windows GOARCH=amd64 go build ./...  → exit 0
```

Both verified twice (before Commit 1 and after Commit 3).

## C-HRA-008 Subagent Boundary Baseline

```
$ grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ internal/hook/ | grep -v "_test.go" | grep -v "^[^:]*:[0-9]*:[ \t]*//" | wc -l
0
```

(Not modified by this SPEC; verified as baseline guard.)

## PRESERVE List (REQ-AFS-011) — Verified Unchanged

- 4 modified files: `.moai/harness/usage-log.jsonl`, `docs-site/hugo.toml`, `docs-site/layouts/_default/baseof.html`, `docs-site/layouts/partials/menu.html` — UNCHANGED
- 2 untracked research files (`moai-adk-current-state-2026-05-22.md`, `v3.0-design-2026-05-22.md`) — UNCHANGED
- `docs-site/content/{en,ja,ko,zh}/book/`, `docs-site/data/menu/extra.yaml`, `docs-site/layouts/_default/redirect.html`, `docs-site/scripts/gen_menu.py`, `docs-site/static/book/`, `internal/hook/.moai/` — UNCHANGED
- 9 out-of-scope Go files (REQ-AFS-010) — `git diff origin/main -- <9 files> | wc -l` returned 0, 42 occurrences preserved (AC-AFS-012)

Working tree noise observed during run-phase (parallel session edits, all preserved untouched):
- `internal/template/templates/.github/actions/*.yml.tmpl` (2 files) + `internal/template/templates/.github/workflows/*.yml.tmpl` (6 files) — parallel session escaping of GitHub Actions template variables
- `internal/template/renderer.go` (+1 line `$GITHUB_WORKSPACE` passthrough token) — parallel session
- `internal/template/github_tmpl_parse_test.go` (untracked NEW test file) — parallel session

All parallel-session changes were carefully excluded from this SPEC's commits via specific `git add <path>` (never `git add -A`).

## Commit Boundary Summary

| Commit | SHA | Files | Purpose |
|---|---|---|---|
| Commit 1 (M1+M2) | `1bd083725` | 79 (38 R + 41 M) | 38 git mv + cross-ref update + isMoaiManaged extension + loadAgentFrontmatter extension |
| Commit 2 (M3) | `fdf325eb6` | 1 | catalog.yaml hash regen for 2 skill content edits |
| Commit 3 (M4 discovery) | `a912e7d5a` | 6 | walker caller fixes (initializer/validator/model_policy + 3 test files) |

## Spec Status Update

- `status`: draft → implemented
- `version`: 0.1.1 → 0.2.0
- HISTORY v0.2.0 row appended to spec.md

## Next Steps (Post-Run)

1. **Sync-phase**: Batch sync via `/moai sync` (covers this SPEC + any other queued SPECs)
2. **Follow-up SPEC**: `SPEC-V3R6-FROZEN-PREFIX-REALIGN-001` (가칭) plan-phase — handles 9 out-of-scope Go files + walker logic supersession
3. **Wave 2 next SPEC**: `SPEC-V3R6-META-HARNESS-PATH-001` (per spec.md §2.2 Wave 2 sequence)
