---
id: SPEC-V3R6-AGENT-FOLDER-SPLIT-001
title: "Implementation Plan — Agent Folder Split"
version: "0.1.1"
status: draft
created: 2026-05-22
updated: 2026-05-22
author: manager-spec
priority: Medium
phase: "v3.0.0"
module: ".claude/agents, internal/template/templates/.claude/agents, internal/template/catalog.yaml"
lifecycle: spec-anchored
tags: "agent, folder-restructure, plan, wave-2"
tier: M
---

# Implementation Plan — SPEC-V3R6-AGENT-FOLDER-SPLIT-001

## HISTORY

| Version | Date | Author | Description |
|---------|------|--------|-------------|
| 0.1.0 | 2026-05-22 | manager-spec | Initial plan — single atomic commit recommendation, M1-M4 milestones |
| 0.1.1 | 2026-05-22 | manager-spec | iter 2: D1 enumeration (10 active files), D2 in-scope Go enumeration (21 files), D6 explicit 3-commit strategy (replaces atomic single-commit), D7 baseline numbers concrete (19 agents, 22 active occurrences, 41 in-scope Go, 42 out-of-scope Go, 19 catalog entries). Scope reduction rationale: 9 Go files out-of-scope → follow-up SPEC `FROZEN-PREFIX-REALIGN-001` (가칭). |

## 1. Approach Overview

Tier M LEAN workflow의 3-artifact (spec/acceptance/plan). orchestrator-direct 또는 manager-develop cycle_type=ddd 위임 둘 다 가능. **권장: manager-develop 위임** (38 git mv + ~82 Edit operations로 작업 volume이 충분하며, Section A-E template 5-section 적용 의무).

### 1.1 Strategy

1. **Branch 전략**: 1인 OSS Hybrid Trunk policy 적용 — `auto_branch: true` → manager-develop 또는 orchestrator가 `feat/SPEC-V3R6-AGENT-FOLDER-SPLIT-001` branch를 origin/main 기반으로 자동 생성 후 작업
2. **순서**: M1 (folder creation + git mv) → M2 (cross-ref update: active rule/skill + in-scope Go + catalog) → M3 (catalog hash regen + Template-First verify) → M4 (cross-platform build + test + 12 ACs verify)
3. **3-commit strategy** (NEW iter 2, D6 fix — replaces single-atomic-commit): M1+M2 → 1차 commit (file move + cross-ref update atomic block) / M3 → 2차 commit (catalog hash regen via `make build`) / M4 verification은 별도 commit 없음 (검증만). 각 commit 후 `go build ./...` + cross-platform build 검증.
4. **선례**: HARNESS-RENAME-001 (5 dir renames + 8 frontmatter + 8 template mirror) 패턴 직접 활용

### 1.2 Milestone Commit Strategy (NEW iter 2)

iter 1 plan-auditor D6 (SHOULD)에 따라 atomic single-commit 권고를 명시적 multi-commit 분할로 변경. 이유: M1 (38 git mv) + M2 (~82 Edit operations)는 부분 실패 시 rollback이 복잡 (특히 19개 file의 history continuity 보존 ↔ cross-ref string replace 동기화). 3-commit으로 분할하면 git bisect + revert 가능성 향상.

| Commit | Milestones | Contents | Verification before commit |
|---|---|---|---|
| Commit 1 | M1 + M2 | 38 git mv + 22 active rule/skill Edit + 41 in-scope Go Edit + 19 catalog.yaml `path:` Edit | `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0 |
| Commit 2 | M3 | `make build` execution → catalog.yaml hash regen (19 hash values updated) + Template-First diff verify (`diff -rq` for agents/core, expert, meta, harness) | `TestManifestHashFormat` PASS, `diff -rq` empty |
| (no commit) | M4 | Verification only — 12 ACs run + cross-platform build + baseline FAIL preservation + out-of-scope file PRESERVE check (AC-AFS-012) | All 12 ACs PASS, AC-AFS-012 specifically verifies 9 out-of-scope files byte-identical |

**Critical**: Commit 1과 Commit 2 사이 build state는 functional이어야 함 (commit 1만으로는 catalog hash mismatch 발생 → `TestManifestHashFormat` FAIL). 이는 acceptable intermediate state (binary `make build`가 trivially 해결).

### 1.3 Decision Rationale

- **`core/expert/meta` 3-folder classification** (4-folder 아님 — `harness/`는 별도 PRESERVE): blueprint § Wave 2 명시
- **manager-* → core/** (별도 `manager/` 폴더 아님): blueprint § Wave 2 row 2 `core/` 명시. workflow operational manager는 MoAI의 "core" 역할
- **builder-harness + claude-code-guide + evaluator-active + plan-auditor + researcher → meta/**: 이들은 MoAI 자체를 빌드/평가/감사/조사하는 메타-수준 agent. 단일 분류 폴더로 묶는 것이 navigability ↑
- **agent `name:` field 보존**: 다른 곳 (catalog.yaml, skills, rules)에서 `name`으로 invoke하므로 변경 시 cascade risk 큼. 경로만 변경
- **iter 2 scope reduction (NEW)**: 9 Go files (frozen_guard family + walker tests + pre_tool.go SentinelHarnessFrozenAgent + 2 walker test files) OUT OF SCOPE. 별도 SPEC `SPEC-V3R6-FROZEN-PREFIX-REALIGN-001` (가칭) 후속 처리. 본 SPEC Tier M (5-15 files) 한계 내 유지 + cross-SPEC tension 회피.

## 2. Milestones

본 SPEC은 시간 추정 대신 priority 기반 ordering 사용.

### M1: Folder Creation + File Movement (Priority: Critical)

**Goal**: 38개 git mv 실행하여 19 agent file을 새 폴더로 이동 (local + template).

**Steps**:
1. `mkdir -p .claude/agents/{core,expert,meta}` (local)
2. `mkdir -p internal/template/templates/.claude/agents/{core,expert,meta}` (template)
3. Local: 19 git mv operations per §3 classification matrix (spec.md §3.1-3.3)
4. Template: 19 git mv operations (same classification, paths prefixed with `internal/template/templates/`)
5. Verify: `find .claude/agents -maxdepth 1 -type d | sort` returns 5 entries (.claude/agents, core, expert, harness, meta) — no `moai`
6. Verify: `git status` shows 38 R (renamed) entries

**Critical**: `.claude/agents/moai/` 와 `internal/template/templates/.claude/agents/moai/` 빈 폴더는 git rename 후 자동 untracked (git은 빈 디렉토리 추적 안함). 만약 `.gitkeep` 등 보존 file이 있다면 explicit `rm` + commit.

### M2: Cross-Reference Update (Priority: Critical)

**Goal**: 22 active rule/skill direct path refs + 41 in-scope Go path refs + 19 catalog.yaml path refs를 새 경로로 갱신. 9 out-of-scope Go files PRESERVE.

**Steps**:

#### M2.1 — Active rule + skill files (22 occurrences in 19 unique files) — D1 FIX iter 2

전체 enumeration (precise grep result):

| File (local) | Lines | Occurrence count | Template mirror? |
|---|---|---|---|
| `.claude/rules/moai/development/agent-authoring.md` | 15 | 1 | YES — `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` (1 occ) |
| `.claude/rules/moai/development/model-policy.md` | 45, 46 | 2 | YES — template mirror (2 occ) |
| `.claude/rules/moai/workflow/spec-workflow.md` | 45, 137 | 2 | YES — template mirror (1 occ, drifted -1) |
| `.claude/rules/moai/workflow/team-protocol.md` | 114 | 1 | YES — template mirror (1 occ) |
| `.claude/skills/moai-harness-learner/SKILL.md` | 151 | 1 | YES — template mirror (1 occ) |
| `.claude/skills/moai-meta-harness/SKILL.md` | 352 | 1 | YES — template mirror (1 occ) |
| `.claude/skills/moai/workflows/harness.md` | 175 | 1 | YES — template mirror (1 occ) |
| `.claude/skills/moai/workflows/plan/spec-assembly.md` | 308 | 1 | YES — template mirror (1 occ) |
| `.claude/skills/moai/workflows/project/meta-harness.md` | 227 | 1 | YES — template mirror (1 occ) |
| `.claude/skills/moai/workflows/release.md` | 878 | 1 | NO — local-only per CLAUDE.local.md §21 (dev-only command) |

**Local subtotal**: 12 occurrences in 10 files
**Template subtotal**: 10 occurrences in 9 files (release.md absent)
**Grand total**: **22 occurrences in 19 unique files**

Edit 방식: 각 line의 path string을 §3 classification matrix에 따라 갱신:
- `manager-*` → `core/manager-*`
- `expert-*` → `expert/expert-*`
- `builder-harness` / `claude-code-guide` / `evaluator-active` / `plan-auditor` / `researcher` → `meta/<name>`

#### M2.2 — In-scope Go code references (41 occurrences in 21 files) — D2 FIX iter 2

**Pre-flight scope filter** (out-of-scope 9 files exclude):

```bash
grep -rln '\.claude/agents/moai/' internal/ 2>/dev/null \
  | grep -v -E '(internal/template/templates/|internal/template/catalog\.yaml|internal/harness/frozen_guard\.go|internal/harness/safety/frozen_guard\.go|internal/harness/safety/frozen_guard_test\.go|internal/harness/safety_preservation_test\.go|internal/harness/safety/pipeline_test\.go|internal/harness/meta_invocation_test\.go|internal/hook/pre_tool\.go|internal/template/catalog_tier_audit_test\.go|internal/template/agent_frontmatter_audit_test\.go)' \
  | sort -u
```

Precise in-scope Go files (21 files, 41 occurrences):

| File | Occurrence count | Pattern |
|---|---|---|
| `internal/cli/agent_lint.go` | 2 | `.claude/agents/moai/<name>.md` literal |
| `internal/cli/plan_audit_d7_d8_test.go` | 1 | fixture path |
| `internal/cli/update.go` | 1 | string literal in update logic |
| `internal/cli/update_preserve_my_harness_test.go` | 1 | test fixture path |
| `internal/cli/update_safety_test.go` | 1 | test fixture path |
| `internal/core/project/initializer_test.go` | 1 | expert-backend.md fixture |
| `internal/hook/agent_start_test.go` | 2 | hook event path assertion |
| `internal/hook/subagent_start.go` | 1 | subagent path identifier |
| `internal/manifest/types_test.go` | 1 | expert-backend.md path |
| `internal/merge/confirm_coverage_test.go` | 1 | expert-backend.md path |
| `internal/research/safety/frozen.go` | 1 | researcher.md path literal (NOTE: 본 파일은 frozen_guard와 다른 file — `internal/research/safety/` 경로 — in-scope) |
| `internal/research/safety/frozen_test.go` | 3 | researcher.md + expert-backend.md fixtures |
| `internal/template/builder_skill_path_test.go` | 1 | builder-harness.md path |
| `internal/template/deployer_bench_test.go` | 1 | benchmark fixture |
| `internal/template/deployer_mode_test.go` | 8 | expert-backend.md fixtures (multiple test cases) |
| `internal/template/deployer_test.go` | 4 | expert-backend.md fixtures |
| `internal/template/embed.go` | 2 | embed.FS directory hints |
| `internal/template/manager_develop_present_test.go` | 3 | manager-develop.md path |
| `internal/template/rule_template_mirror_test.go` | 3 | plan-auditor.md + manager-git.md mirror paths |
| `internal/template/slim_guard.go` | 1 | builder-harness.md sentinel |
| `internal/template/slim_guard_test.go` | 2 | builder-harness.md + expert-backend.md fixtures |
| **TOTAL** | **41** | **21 files** |

**Order**: source code (`.go` non-test) first (`agent_lint.go`, `update.go`, `subagent_start.go`, `frozen.go`, `embed.go`, `slim_guard.go` = 6 files / 8 occurrences), then test files (15 files / 33 occurrences). Each edit immediate `go build ./...` verify before continuing.

**Critical out-of-scope filter discipline**: 다음 9 files는 절대 수정 금지 (REQ-AFS-010 / AC-AFS-012):
- `internal/harness/frozen_guard.go` (1)
- `internal/harness/safety/frozen_guard.go` (1)
- `internal/harness/safety/frozen_guard_test.go` (8)
- `internal/harness/safety_preservation_test.go` (1)
- `internal/harness/safety/pipeline_test.go` (1)
- `internal/harness/meta_invocation_test.go` (4)
- `internal/hook/pre_tool.go` (2)
- `internal/template/catalog_tier_audit_test.go` (4)
- `internal/template/agent_frontmatter_audit_test.go` (20)

**Total out-of-scope: 42 occurrences in 9 files — DO NOT EDIT.**

#### M2.3 — catalog.yaml (19 entries) — iter 2 correction

`internal/template/catalog.yaml` 다음 19 lines (iter 1 "18"은 오기, builder-harness.md line 341 nested section 포함):

| Line | Old path | New path |
|---|---|---|
| 129 | `templates/.claude/agents/moai/claude-code-guide.md` | `templates/.claude/agents/meta/claude-code-guide.md` |
| 154 | `templates/.claude/agents/moai/evaluator-active.md` | `templates/.claude/agents/meta/evaluator-active.md` |
| 159 | `templates/.claude/agents/moai/expert-refactoring.md` | `templates/.claude/agents/expert/expert-refactoring.md` |
| 164 | `templates/.claude/agents/moai/manager-brain.md` | `templates/.claude/agents/core/manager-brain.md` |
| 169 | `templates/.claude/agents/moai/manager-develop.md` | `templates/.claude/agents/core/manager-develop.md` |
| 174 | `templates/.claude/agents/moai/manager-docs.md` | `templates/.claude/agents/core/manager-docs.md` |
| 179 | `templates/.claude/agents/moai/manager-git.md` | `templates/.claude/agents/core/manager-git.md` |
| 184 | `templates/.claude/agents/moai/manager-project.md` | `templates/.claude/agents/core/manager-project.md` |
| 189 | `templates/.claude/agents/moai/manager-quality.md` | `templates/.claude/agents/core/manager-quality.md` |
| 194 | `templates/.claude/agents/moai/manager-spec.md` | `templates/.claude/agents/core/manager-spec.md` |
| 199 | `templates/.claude/agents/moai/manager-strategy.md` | `templates/.claude/agents/core/manager-strategy.md` |
| 204 | `templates/.claude/agents/moai/plan-auditor.md` | `templates/.claude/agents/meta/plan-auditor.md` |
| 209 | `templates/.claude/agents/moai/researcher.md` | `templates/.claude/agents/meta/researcher.md` |
| 235 | `templates/.claude/agents/moai/expert-backend.md` | `templates/.claude/agents/expert/expert-backend.md` |
| 246 | `templates/.claude/agents/moai/expert-devops.md` | `templates/.claude/agents/expert/expert-devops.md` |
| 303 | `templates/.claude/agents/moai/expert-performance.md` | `templates/.claude/agents/expert/expert-performance.md` |
| 308 | `templates/.claude/agents/moai/expert-security.md` | `templates/.claude/agents/expert/expert-security.md` |
| 328 | `templates/.claude/agents/moai/expert-frontend.md` | `templates/.claude/agents/expert/expert-frontend.md` |
| 341 | `templates/.claude/agents/moai/builder-harness.md` | `templates/.claude/agents/meta/builder-harness.md` |

**4 harness/ entries (lines 134, 139, 144, 149) PRESERVE (no change)**.

**Critical**: `path:` 변경 후 hash 값은 stale 상태가 됨. M3에서 `make build`가 `gen-catalog-hashes.go --all` 자동 실행 (CATALOG-SSOT-001 mechanism) → 19개 hash 자동 regen.

### M3: catalog.yaml hash regeneration + Template-First verify (Priority: High)

**Steps**:
1. `make build` 실행 → `gen-catalog-hashes.go --all` 자동 실행 → catalog.yaml hash field 자동 갱신
2. `diff -rq .claude/agents/core/ internal/template/templates/.claude/agents/core/` → empty (no diff)
3. 동일 검증 `expert/`, `meta/`, `harness/`
4. `TestManifestHashFormat` PASS 확인 (CATALOG-SSOT-001 baseline guard)
5. Commit 2 등록 (3-commit strategy §1.2)

### M4: Cross-platform build + Test verification + 12 ACs PASS (Priority: High)

**Steps**:
1. `go build ./...` → exit 0
2. `GOOS=windows GOARCH=amd64 go build ./...` → exit 0
3. `go test ./...` → in-scope NEW regression 0건 (baseline 3 FAIL + out-of-scope deferred FAIL `TestAllAgentsInCatalog` + `TestAgentFrontmatterAudit` 보존; 그 외 regression 시 blocker)
4. spec-lint NEW regression 0건 확인
5. C-HRA-008 boundary check (manager-develop-prompt-template.md B3): `grep -rn 'AskUserQuestion\|mcp__askuser' internal/harness/ internal/hook/ | grep -v "_test.go" | grep -v "// "` → 0 matches (본 SPEC scope 무관하나 baseline guard)
6. **AC-AFS-001 ~ AC-AFS-012 모두 PASS verify** (12 ACs, AC-AFS-012는 out-of-scope file PRESERVE)

## 3. Technical Approach

### 3.1 Tool Selection

- **`git mv`** (not `Write` + `rm`): 38 file movements 모두 `git mv` 사용 의무. file history 보존.
- **`Edit` tool** (not sed/awk per CLAUDE.local.md §6): catalog.yaml + Go code + skill/rule prose의 path string replacement. Out-of-scope filter 의무 (M2.2 grep pattern).
- **`Bash` tool**: `mkdir`, `find`, `diff -rq`, `go build`, `go test`, `make build` 실행
- **`Grep` tool**: cross-ref enumeration (pre-flight + verification)
- **`Glob` tool**: 19 agent file enumeration

### 3.2 Atomicity (3-commit strategy, NEW iter 2)

D6 fix: §1.2 표 참조. Commit 1 (M1+M2) → Commit 2 (M3) → M4 verification only.

Commit 1 message 권장:
```
feat(SPEC-V3R6-AGENT-FOLDER-SPLIT-001): split .claude/agents/moai/ into core/expert/meta/ (38 git mv + cross-ref)

- 19 agents split: 8 core/ + 6 expert/ + 5 meta/
- Template-First mirror: identical structure in internal/template/templates/
- 22 active rule/skill ref updates in 19 unique files (12 local + 10 template)
- 41 in-scope Go ref updates in 21 files
- 19 catalog.yaml path entries updated (hashes pending M3 make build)
- .claude/agents/harness/ PRESERVE (HARNESS-RENAME-001 result)
- 9 out-of-scope Go files PRESERVE (FROZEN guard + walker tests, follow-up SPEC FROZEN-PREFIX-REALIGN-001)

🗿 MoAI <email@mo.ai.kr>
```

Commit 2 message 권장:
```
chore(SPEC-V3R6-AGENT-FOLDER-SPLIT-001): regenerate catalog.yaml hashes (19 entries)

- make build → gen-catalog-hashes.go --all (CATALOG-SSOT-001 mechanism)
- 19 hash values updated to reflect new path-based hash inputs
- TestManifestHashFormat PASS

🗿 MoAI <email@mo.ai.kr>
```

### 3.3 Verification batching (manager-develop-prompt-template.md §E)

run-phase 완료 보고 시 다음 8 verification을 single-turn parallel Bash 호출로 실행 (verification-batch-pattern.md, AC-AFS-012 추가):

```bash
# 1. AC-AFS-001 folder structure
find .claude/agents -maxdepth 1 -type d | sort

# 2. AC-AFS-002 file count
ls .claude/agents/core/*.md | wc -l  # + expert/, meta/, harness/

# 3. AC-AFS-006 in-scope Go refs (filtered)
grep -rn '\.claude/agents/moai/' internal/ 2>/dev/null \
  | grep -v -E '(internal/template/templates/|internal/template/catalog\.yaml|internal/harness/frozen_guard\.go|internal/harness/safety/frozen_guard\.go|internal/harness/safety/frozen_guard_test\.go|internal/harness/safety_preservation_test\.go|internal/harness/safety/pipeline_test\.go|internal/harness/meta_invocation_test\.go|internal/hook/pre_tool\.go|internal/template/catalog_tier_audit_test\.go|internal/template/agent_frontmatter_audit_test\.go)' \
  | wc -l

# 4. AC-AFS-007 active rule/skill refs
grep -rn '\.claude/agents/moai/' .claude/rules/ .claude/skills/ \
  internal/template/templates/.claude/rules/ internal/template/templates/.claude/skills/ \
  CLAUDE.md internal/template/templates/CLAUDE.md 2>/dev/null | wc -l

# 5. AC-AFS-008 Template-First Rule
diff -rq .claude/agents/core/ internal/template/templates/.claude/agents/core/

# 6. AC-AFS-009 catalog.yaml
grep -c 'agents/moai/' internal/template/catalog.yaml

# 7. AC-AFS-012 (NEW) out-of-scope file PRESERVE
git diff origin/main -- internal/harness/frozen_guard.go internal/harness/safety/ \
  internal/harness/safety_preservation_test.go internal/harness/meta_invocation_test.go \
  internal/hook/pre_tool.go internal/template/catalog_tier_audit_test.go \
  internal/template/agent_frontmatter_audit_test.go | wc -l

# 8. Cross-platform build
go build ./... && GOOS=windows GOARCH=amd64 go build ./...
```

## 4. Risks and Mitigations

### Risk R-AFS-001 (M/H): In-scope Go test fixture path 갱신 누락 (21 files)

**Mitigation**:
- Pre-flight grep으로 ALL in-scope occurrences enumerate (M2.2 21-file 표)
- Out-of-scope 9 files filter 의무 (M2.2 grep pattern + `grep -v -E ...`)
- Edit 후 즉시 `go build ./...` (build error 즉시 발견)
- 마지막 verification에서 AC-AFS-006 → 0 in-scope, AC-AFS-012 → 0 out-of-scope diff

### Risk R-AFS-002 (L/M): catalog.yaml hash mismatch

**Mitigation**:
- CATALOG-SSOT-001의 `make build` self-healing manifest gate에 의존
- M3에서 `make build` 강제 실행 + `TestManifestHashFormat` PASS verify

### Risk R-AFS-003 (L/H): 38 git mv 부분 실패

**Mitigation**:
- 3-commit strategy로 atomic 단위 분할 (M1+M2 / M3 / M4)
- M1 완료 후 `git status` 확인 (38 R entries)
- 부분 실패 시 `git restore --staged` + 재시도

### Risk R-AFS-004 (L/H): frontmatter 손상

**Mitigation**:
- `git mv`는 file content 변경 없음 → byte-identical 자동 보장
- AC-AFS-005에서 5 sample agents frontmatter inspection

### Risk R-AFS-005 (M/M): post-implementation 다른 PR가 옛 경로로 add file

**Mitigation**: run-phase 권장 결정 — CI guard test 추가 여부 (`TestNoAgentMoaiPathRemnant` grep-based, deferred to run-phase decision)

### Risk R-AFS-006 (NEW iter 2, M/H): 9 out-of-scope Go files 실수로 수정

**Mitigation**:
- M2.2 grep filter pattern 의무 사용
- AC-AFS-012 verification (`git diff origin/main -- <9 files>` empty) 의무
- 실수 발견 시 즉시 `git restore --source=origin/main -- <file>`로 revert
- progress.md near-miss 기록 (Scenario 4)

### Risk R-AFS-007 (NEW iter 2, L/M): manager-develop Agent worktree spawn 실패

**Background**: `project_v3r6_agent_isolation_worktree_removal` (2026-05-22) 기록에 따르면 manager-develop 등 5 agents의 `isolation: worktree` frontmatter가 제거되었으나, worktree-integration.md HARD 규칙은 여전히 write-heavy agent에 isolation 의무를 명시. 본 SPEC도 manager-develop 위임 시 WorktreeCreate hook regression 가능성.

**Mitigation**:
- Plan A: manager-develop 위임 시도 → 실패 시 orchestrator-direct fallback (CLAUDE.md §16 "위임 대상 부재" 조항)
- Plan B: orchestrator-direct로 처음부터 진행. 위임 prompt 5-section template을 self-application
- 선례: V3R6-CATALOG-SSOT-001, V3R6-HARNESS-RENAME-001 모두 orchestrator-direct로 완료

### Risk R-AFS-008 (NEW iter 2, L/M): post-implementation FAIL state confusion

**Background**: 본 SPEC implemented 후 `TestAllAgentsInCatalog` + `TestAgentFrontmatterAudit`는 walker가 `.claude/agents/moai` 빈 디렉토리를 읽어 FAIL 상태로 전환. 이는 baseline 3 FAIL (`TestRuleTemplateMirrorDrift`, `TestLateBranchTemplateMirror`, `TestSkillsContainPlanAuditGateMarkers`)과 다른 카테고리.

**Mitigation**:
- progress.md에 명시적 카테고리 분리:
  - **Pre-existing baseline FAIL (3건)**: TestRuleTemplateMirrorDrift / TestLateBranchTemplateMirror / TestSkillsContainPlanAuditGateMarkers
  - **Out-of-scope deferred FAIL (2건, NEW post-implementation)**: TestAllAgentsInCatalog / TestAgentFrontmatterAudit → follow-up SPEC `FROZEN-PREFIX-REALIGN-001` scope
- Follow-up SPEC plan-phase에서 명시적 supersession

## 5. Sprint Breakdown

본 SPEC은 Tier M이지만 sequential file operation 위주이므로 sprint 분할 불필요. M1→M2→M3→M4 단일 sprint로 진행. 3-commit strategy 적용 (§1.2).

### Sprint 1 (Single Sprint, 3 commits)

- M1: 38 git mv (folder creation + file movement)
- M2: cross-ref update (22 active rule/skill + 41 in-scope Go + 19 catalog.yaml)
- Commit 1 (M1+M2)
- M3: make build + Template-First diff verify
- Commit 2 (M3)
- M4: cross-platform build + test + 12 ACs verify (no commit)

Total estimated operations: 38 git mv + ~82 Edit operations + 1 `make build` invocation + 2 commits.

## 6. Cross-Reference Enumeration Cheat Sheet

run-phase pre-flight 단계에서 다음 grep batch 실행하여 모든 affected refs를 enumerate:

```bash
# Active rules + skills (22 expected per AC-AFS-007)
grep -rln '\.claude/agents/moai/' .claude/rules/ .claude/skills/ \
  internal/template/templates/.claude/rules/ internal/template/templates/.claude/skills/ \
  CLAUDE.md internal/template/templates/CLAUDE.md 2>/dev/null

# In-scope Go code (41 expected per AC-AFS-006)
grep -rn '\.claude/agents/moai/' internal/ 2>/dev/null \
  | grep -v -E '(internal/template/templates/|internal/template/catalog\.yaml|internal/harness/frozen_guard\.go|internal/harness/safety/frozen_guard\.go|internal/harness/safety/frozen_guard_test\.go|internal/harness/safety_preservation_test\.go|internal/harness/safety/pipeline_test\.go|internal/harness/meta_invocation_test\.go|internal/hook/pre_tool\.go|internal/template/catalog_tier_audit_test\.go|internal/template/agent_frontmatter_audit_test\.go)'

# Out-of-scope Go code (42 expected, MUST NOT BE EDITED, AC-AFS-012)
grep -c '\.claude/agents/moai/' \
  internal/harness/frozen_guard.go \
  internal/harness/safety/frozen_guard.go \
  internal/harness/safety/frozen_guard_test.go \
  internal/harness/safety_preservation_test.go \
  internal/harness/safety/pipeline_test.go \
  internal/harness/meta_invocation_test.go \
  internal/hook/pre_tool.go \
  internal/template/catalog_tier_audit_test.go \
  internal/template/agent_frontmatter_audit_test.go 2>/dev/null

# catalog.yaml (19 expected per AC-AFS-009)
grep -n 'templates/\.claude/agents/moai/' internal/template/catalog.yaml
```

## 7. Open Questions (run-phase 결정 대상)

본 plan-phase는 다음 3개 항목에 대해 default proceed 입장. 사용자 또는 manager-develop이 다른 결정을 원할 시 blocker report로 escalation 가능:

### OQ-AFS-001: CI guard test 추가 여부

`TestNoAgentMoaiPathRemnant` 같은 grep-based guard test 추가는 본 SPEC scope 외이지만 후속 PR 안전성 ↑. **Default decision**: 본 SPEC run-phase에서 추가하지 않음 (separate SPEC 또는 follow-up). manager-develop이 추가가 cheap하다고 판단 시 in-scope로 흡수 가능.

### OQ-AFS-002: agent-memory 폴더 경로

`.claude/agent-memory/<agent-name>/` 폴더는 `<agent-name>`을 그대로 사용 (folder 변경 영향 없음). 따라서 본 SPEC scope 무관. **Default decision**: 변경 없음.

### OQ-AFS-003: docs-site advanced/agents 문서

`docs-site/content/{ko,en,ja,zh}/advanced/agents.md` 등 user-facing docs에 옛 경로 문자열이 있을 가능성. **Default decision**: 본 SPEC scope 외 (docs-only SPEC 별도). manager-develop이 발견 시 progress.md에 기록.

### OQ-AFS-004 (NEW iter 2): Follow-up SPEC trigger timing

`SPEC-V3R6-FROZEN-PREFIX-REALIGN-001` (가칭) plan-phase 진입 timing — 본 SPEC implemented 직후 vs Wave 6 (RULES-COMPLIANCE-001) 완료 후. **Default decision**: 사용자 결정 보류 (본 SPEC run-phase 완료 후 progress.md에 follow-up SPEC 후보 명시 + 사용자 prompt).

## 8. Cross-References

- spec.md (this SPEC's specification, v0.1.1)
- acceptance.md (this SPEC's binary AC matrix, v0.1.1, 12 ACs)
- v3.0.0 redesign blueprint: `.moai/research/v3-redesign-blueprint-2026-05-22.md` § Wave 2
- HARNESS-RENAME-001 SPEC pattern: `.moai/specs/SPEC-V3R6-HARNESS-RENAME-001/`
- CATALOG-SSOT-001 self-healing manifest: `.moai/specs/SPEC-V3R6-CATALOG-SSOT-001/`
- Frontmatter schema: `.claude/rules/moai/development/spec-frontmatter-schema.md`
- Agent authoring: `.claude/rules/moai/development/agent-authoring.md`
- manager-develop prompt template: `.claude/rules/moai/development/manager-develop-prompt-template.md` (Section A-E 5-section + B4/B5/B6/B8 known issues)
- verification batch pattern: `.claude/rules/moai/workflow/verification-batch-pattern.md`
- CLAUDE.local.md §2 (Template-First Rule), §21 (Dev-Only Commands Isolation — workflows/release.md), §23 (1인 OSS Hybrid Trunk)

## 9. Project Memory Anchors

본 SPEC 작업 시 다음 memory file 참조 강력 권장 (Section B autoload):

- `project_v3r6_wave2_harness_rename_001_plan_complete` — Tier S → Tier M transition pattern, plan-auditor self-est option, manager-spec 1-pass success
- `project_v3r6_harness_rename_001_run_complete` — 5 dir renames + 8 frontmatter + 8 Template-First mirror + 8 catalog entries + 3 cascade test count adjustments (HARNESS-RENAME-001 선례)
- `project_v3r6_template_mirror_drift_audit_2026_05_22` — pre-existing baseline 10 drift files (out of scope) + 3 baseline FAIL preservation rule
- `project_v3r6_catalog_ssot_001_run_complete` — `make build` self-healing manifest gate (gen-catalog-hashes.go --all 자동 실행)
- `project_v3r6_agent_isolation_worktree_removal` — WorktreeCreate hook regression 우회 (manager-develop Agent 호출 실패 시 fallback)
- `feedback_worktree_autonomous` — user policy 2026-05-17 L1 worktree Claude Code runtime autonomous
- `feedback_dead_config_audit_spec_catalog_cross_check` — pre-flight SPEC catalog cross-check 의무 (본 SPEC depends_on satisfied)
