---
id: SPEC-V3R6-MAIN-RED-REMEDIATION-001
title: "Plan — internal/template main-RED 4-group 일괄 해소"
version: "0.1.0"
status: draft
created: 2026-05-30
updated: 2026-05-30
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/template"
lifecycle: spec-anchored
tags: "test-correction, template, mirror-drift, internal-content-leak, hook-count, main-red, plan"
tier: M
---

# Plan — internal/template main-RED 4-group 일괄 해소

## §A. Context (위치 + 분기 + 산출물)

- 작업 위치: `/Users/goos/MoAI/moai-adk-go` (project root)
- 현재 branch: `main` (Hybrid Trunk 1-person OSS — Tier M main 직진 push, PR 없음)
- 패키지 scope: `internal/template` 단일 패키지 (모든 13 fail 이 여기 국한)
- SPEC 산출물: `.moai/specs/SPEC-V3R6-MAIN-RED-REMEDIATION-001/{spec,plan,acceptance}.md` (Tier M LEAN 3-artifact)
- PRESERVE 대상: golangci-lint 결과 (0 issues 유지), 모든 다른 패키지, `.claude/agents/` 디렉토리 구조, `core/`/`meta/`/`expert/` 미생성
- EXTEND 대상: `internal/template` 의 4개 group 에 해당하는 테스트 + template 산출물

## §B. Known Issues (자동 주입 — 도메인 관련 카테고리)

본 SPEC 은 markdown/template/test 정정이 주이므로 manager-develop-prompt-template §B 의 카테고리 중 다음만 관련:

- **B1 Cross-platform Build Tags**: syscall 신규 코드 없음 — 그러나 최종 green gate 에서 `GOOS=windows GOARCH=amd64 go build ./...` 통과 의무 (REQ-MRR-011).
- **B2 Cross-SPEC 정책 충돌 스캔**: G1 의 근본 원인이 superseded SPEC 의 stale 기대값 carry 이다. agents-layout 테스트의 주석에 등장하는 선행 SPEC 참조(AGENT-FOLDER-SPLIT-001 / V2-V3-CLEAN-REINSTALL-001 / TEST-REFACTOR-001)는 모두 이미 superseded/completed — reversal 명시 불필요, 테스트 기대값만 canonical FLAT 으로 정정.
- **B4 Frontmatter Canonical Schema**: 본 SPEC 산출물 frontmatter 는 12-field canonical (`created`/`updated`/`tags`) 준수.
- **B5 CI 3-tier 인지**: spec-lint / golangci-lint / Test(per OS) 별도. 본 SPEC 은 Test tier 의 13 fail 해소가 목표; lint tier 는 이미 green (불변).
- **B8 Working Tree Hygiene**: runtime-managed files (`.moai/state/`, `.moai/cache/`, `.moai/harness/`) 손대지 말 것. 무관 untracked (`.moai/research/*`, `scripts/audit-spec-sync-drift.sh`) commit 포함 금지 — specific path staging.
- **B9 Git Commit + Push 자체 수행**: manager-develop 이 본 SPEC scope commit + main 직진 push. Conventional Commits (`fix(SPEC-V3R6-MAIN-RED-REMEDIATION-001): M{N} <subject>`). `--no-verify` 금지.
- **B10 Untouched Paths PRESERVE**: 병렬 세션 race 주의 (메모리상 main-RED remediation 수렴 중 — 중복 SPEC 금지 확인됨). `internal/template` 외 디렉토리 손대지 말 것.
- **B11 AskUserQuestion 금지**: subagent 는 blocker report 만, AskUserQuestion 호출 금지.

## §C. Pre-flight Check (착수 전 의무 검증)

```bash
# 1. baseline 확인
git branch --show-current   # main
git rev-parse HEAD

# 2. cross-platform build baseline
go build ./...
GOOS=windows GOARCH=amd64 go build ./...

# 3. lint baseline (불변 확인 — 0 issues 유지가 목표)
golangci-lint run --timeout=2m 2>&1 | tail -5

# 4. 현재 fail 목록 캡처 (13 parent)
go test ./internal/template/... 2>&1 | grep -E '^--- FAIL' 

# 5. leak site 30개 재캡처 (G3 baseline)
go test ./internal/template/... -run TestTemplateNoInternalContentLeak -v 2>&1 | grep 'class='
```

## §D. Constraints (DO NOT VIOLATE)

- `.claude/agents/` 에 `core/`/`meta/`/`expert/` 디렉토리 생성 금지 (REQ-MRR-003) — G1 은 순수 테스트 정정.
- golangci-lint 관련 변경 절대 금지 (이미 green).
- 3개 미완료 config-schema SPEC entangle 금지.
- §25 doctrine 본문 (CLAUDE.local.md §25) 개정 금지 — substitution dictionary 적용만.
- `--no-verify`, `--amend`, force-push 금지.
- template 수정 시 `make build` 후 `embedded.go` 재생성 의무 (Template-First).
- 무관 untracked 파일 commit 금지 (specific path staging).

### Out of Scope

- golangci-lint 작업 (이미 0 issues — spec.md §D.1 참조).
- `.claude/agents/` 디렉토리 구조 생성/이동 (G1 은 순수 테스트 정정).
- 3개 미완료 V3R5 config-schema SPEC entangle.
- mirror 정책 영구 아키텍처 재설계 (본 SPEC 은 main green 복구 최소 변경만).

## §E. Self-Verification Deliverables

manager-develop 완료 보고 시 각 milestone 별 AC PASS/FAIL matrix + 다음 명령 출력 포함:

- E1: AC binary matrix (acceptance.md §D 참조).
- E2: `go test ./internal/template/...` → 0 fail.
- E3: `GOOS=windows GOARCH=amd64 go build ./...` → exit 0.
- E4: `golangci-lint run --timeout=2m` → 0 issues (baseline 불변).
- E5: `make build` → `embedded.go` 재생성 성공.
- E6: commit SHA 리스트 + main push 결과.

## §F. Milestones (priority-ordered: risk-ascending)

milestone 순서는 **위험 오름차순** — 자명한 expected-value 정정 먼저, design-conflict (G2) 마지막.

### M1 — G4 hook-count expected-value 정정 (가장 작고 안전)

- **대상**: `internal/template/settings_test.go:512` `const expectedCount = 20` → `21`.
- **작업**: 상수를 21 로 변경 + `// = 20 active hook registrations` 주석을 `// = 21 active hook registrations (20 base + PreCommit added by pre-commit spec-status hook)` 로 갱신. 전체 21 event list 검증: SessionStart, PostToolUse, TeammateIdle, TaskCompleted, ConfigChange, CwdChanged, SessionEnd, PostToolUseFailure, SubagentStart, UserPromptSubmit, StopFailure, FileChanged, PreCompact, PermissionDenied, PermissionRequest, PreCommit, PreToolUse, Stop, SubagentStop, PostCompact, InstructionsLoaded.
- **verify**: `go test ./internal/template/... -run TestSettingsTemplateHookEventCount -v`
- **risk**: Trivial — 단일 상수 정정.
- **AC**: AC-MRR-004.

### M2 — G1 agents-layout 테스트 정정 (8 테스트 파일, FLAT 정렬)

- **대상 테스트 파일 + 정정 내용**:
  - `embedded_namespace_test.go` (`TestTemplateAgentsStructure`): `expected = {core, meta}` → FLAT `moai/` 기대. ReadDir `.claude/agents` → `moai` 단일 서브디렉토리 + 7 파일 검증. `HARNESS_NAMESPACE_LEAK` sentinel 유지.
  - `contract_schema_test.go` (`TestContractSchemaVerification`, `TestContractAssertionsNaturalLanguage`, `TestBackwardCompatibility`): `templatesDir = templates/.claude/agents/core` → `.../agents/moai`; `domains := {core, meta}` → `{moai}`; BackwardCompatibility domain enum `{core, expert, meta, harness}` → `{moai}` 정렬.
  - `agent_frontmatter_audit_test.go` (`TestAgentFrontmatterAudit`): 라인 77/198/227 `[]string{"core","meta"}` → `[]string{"moai"}`; 라인 153 주석 path drift 갱신.
  - `embed_test.go` (`TestEmbeddedTemplates_AgentDefinitions`): 라인 49 `domains := {core, meta}` → `{moai}`; 라인 70 `.claude/agents/core/manager-develop.md` → `.../agents/moai/manager-develop.md`; mdCount threshold 를 7 retained 기준으로.
  - `manager_develop_present_test.go` (`TestManagerDevelopActiveAgentPresent`, `TestManagerDevelopIsActiveAgent`): 기대 경로 `.claude/agents/core/manager-develop.md` → `.../agents/moai/manager-develop.md`.
  - `builder_skill_path_test.go` (`TestBuilderSkillPathStructure`): 라인 33 `const agentPath = ".claude/agents/meta/builder-harness.md"` → `".claude/agents/moai/builder-harness.md"`.
- **작업 원칙**: 디렉토리 구조 생성/이동 절대 금지. 테스트의 path 기대값 + domain enum + count threshold 만 canonical FLAT `moai/` (7 retained) 으로 정정.
- **verify**: `go test ./internal/template/... -run 'TestTemplateAgentsStructure|TestContractSchemaVerification|TestContractAssertionsNaturalLanguage|TestBackwardCompatibility|TestAgentFrontmatterAudit|TestEmbeddedTemplates_AgentDefinitions|TestManagerDevelopActiveAgentPresent|TestManagerDevelopIsActiveAgent|TestBuilderSkillPathStructure' -v`
- **risk**: Low-Medium — 다수 파일이나 mechanical path 정정. **C1 coupling 주의**: 본 milestone 이 끝나면 §H 의 G1↔G3 경로 결합을 M3 에서 처리해야 함.
- **AC**: AC-MRR-001, AC-MRR-002, AC-MRR-003.

### M3 — G3 internal-content leak 해소 (29 prose-substitution + 1 allowlist-path-correction) + G1↔G3 경로 결합

[HARD] 30개 leak site 는 두 경로로 분해된다: **29 sites prose-substituted + 1 site (site #1) allowlist-path-corrected** (§spec.md C.2 참조). site #1 은 manager-spec 의 SPEC ID regex Pre-Write Self-Check pedagogical 예제이므로 **prose 치환 금지** — `pedagogicalAllowlist` 의 `File:` 경로 정정으로 해소.

- **대상**: 30 leak site (acceptance.md §D.2 enumeration). 파일별:
  - `templates/.claude/agents/moai/manager-spec.md` (1 site: SPEC-ID — **site #1, prose 치환 NOT, allowlist-path-correction 으로 해소**)
  - `templates/.claude/output-styles/moai/moai.md` (5 site: SPEC-ID ×2 + REQ-TII-001 + AC-TII-007 + REQ-COORD-010)
  - `templates/.claude/rules/moai/development/agent-authoring.md` (4 site: SPEC-ID + Finding A3 + archive-date + backup-path)
  - `templates/.claude/rules/moai/development/spec-frontmatter-schema.md` (3 site: SPEC-ID ×3)
  - `templates/.claude/rules/moai/workflow/archived-agent-rejection.md` (10 site: SPEC-ID + REQ ×3 + AC ×2 + Finding ×3 + archive-date + backup-path)
  - `templates/.claude/rules/moai/workflow/spec-workflow.md` (3 site: SPEC-ID + archive-date + backup-path)
  - `templates/.claude/skills/moai/workflows/harness.md` (1 site: SPEC-ID)
  - `templates/.claude/skills/moai/workflows/plan/spec-assembly.md` (1 site: SPEC-ID)
  - `templates/.moai/config/sections/cache.yaml` (1 site: SPEC-ID)
- **작업 (29 prose-substitution sites)**: 선행 isolation SPEC design §B 의 substitution dictionary S1~S6 적용. SPEC-ID → "선행 SPEC"/"the canonical ... policy", REQ/AC token → prose 설명 또는 삭제, Finding citation → generic Anthropic best-practice prose, archive-date/backup-path → generic placeholder 또는 삭제. **site #1 (manager-spec SPEC-ID) 은 제외** — pedagogical 예제이므로 치환 금지.
- **작업 (site #1 allowlist-path-correction)**: `pedagogicalAllowlist` 구조체 (라인 210) 의 **`File:` 필드 `core/`→`moai/` 정정** (라인 233/240 — **실행 코드 차원**). 이것이 pedagogical 면제를 회복하여 site #1 false-positive 를 해소한다. prose 치환을 적용하지 않으므로 manager-spec 의 regex teaching 예제 (`SPEC-AUTH-001` / `SPEC-V3R6-SPEC-ID-VALIDATION-001` decomposition 출력) 가 보존된다.
- **G1↔G3 경로 결합 (C1) — 2개 차원 분리**:
  - **실행 코드 차원 (라인 233/240)**: 위 site #1 allowlist-path-correction 작업과 동일 (`pedagogicalAllowlist` `File:` 필드 정정). `isPedagogicallyAllowed` 가 소비하는 키이므로 반드시 mirror 경로 `moai/` 와 일치해야 함.
  - **코멘트 차원 (라인 203/254/337–338)**: docstring/설명 코멘트의 `.claude/agents/core/manager-spec.md` 예시 경로를 `moai/` 로 정정 (무해, 일관성 목적). 실행에 영향 없으나 stale 참조 제거.
  - **이 작업은 M3 에 포함** (M2 가 FLAT 경로로 바꿨으므로 leak test 의 실행 코드 + 코멘트 경로를 같은 정정 흐름에서 atomic 처리).
- **verify**: `go test ./internal/template/... -run TestTemplateNoInternalContentLeak -v` → 0 occurrence.
- **risk**: Medium — 29 site prose 치환은 의미 보존 판단 필요 (design §B dictionary 가 가이드); site #1 은 allowlist `File:` 경로 정정으로 단순 해소.
- **AC**: AC-MRR-005, AC-MRR-006, AC-MRR-007.

### M4 — G2 mirror-drift design 결정 적용 + resync

- **전제**: §G 의 design 결정 (권장: per-file 정책) 을 적용.
- **대상 8 drift 파일**: agents/moai/{manager-spec, manager-git, plan-auditor}.md, rules/core/agent-common-protocol.md, rules/workflow/verification-batch-pattern.md, rules/development/manager-develop-prompt-template.md, rules/workflow/ci-watch-protocol.md, rules/workflow/agent-teams-pattern.md.
- **작업 (per-file 정책 채택 시)** — ground-truth: drift∩leak 교집합은 `manager-spec.md` **단독 1파일**:
  1. **leak-test-coverage 로 이관 (manager-spec.md 단독)**: source 가 internal content 보유 working copy 이고 mirror 는 sanitized 되어야 하므로, **mirror 를 sanitized 상태로 유지** + `manager-spec.md` 를 mirror-drift allowlist 에서 제외하고 leak test 로 커버. allowlist 제외는 `rule_template_mirror_test.go` 의 `lateBranchMirroredPaths` slice 에서 `.claude/agents/moai/manager-spec.md` entry 제거 + `// per-file: §25 sanitization 대상 — leak test 로 커버` 주석.
  2. **cp-resync (나머지 7 drift 파일 — byte-parity 유지)**: `agent-common-protocol.md`, `manager-git.md`, `plan-auditor.md`, `verification-batch-pattern.md`, `manager-develop-prompt-template.md`, `ci-watch-protocol.md`, `agent-teams-pattern.md` 는 모두 leak test 보고 **0건** (drift-only) 이다. 따라서 source → mirror 재복사 (`cp source mirror`) 로 byte-parity 복원하고 allowlist 에 유지한다. (`agent-common-protocol.md` 는 leak 0 이므로 exclusion 대상이 **아니다** — cp-resync 대상.)
- **G2↔G3 충돌 검증 (C2)**: M3 에서 manager-spec.md mirror 가 0 leak 으로 sanitize 된 후, byte-parity allowlist 에서 manager-spec.md 만 제외되었고 나머지 7 파일은 cp-resync 후 byte 동일임을 확인하여 모순 없음을 입증.
- **make build**: template 변경 후 `make build` 로 `embedded.go` 재생성.
- **verify**: `go test ./internal/template/... -run 'TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror' -v`
- **risk**: Medium-High — design 결정 적용 + allowlist 수정 + cp resync 혼합. 가장 신중한 milestone.
- **AC**: AC-MRR-008, AC-MRR-009, AC-MRR-010.

### M5 — 통합 green gate + cross-platform build

- **작업**: 전체 패키지 테스트 통과 + cross-platform build + lint baseline 확인.
- **verify** (병렬 batch):
  ```bash
  go test ./internal/template/...           # 0 fail 의무
  go build ./...                            # exit 0
  GOOS=windows GOARCH=amd64 go build ./...  # exit 0
  golangci-lint run --timeout=2m            # 0 issues (불변)
  go test ./... 2>&1 | grep -E '^(FAIL|ok)' # 다른 패키지 cascading 확인
  ```
- **risk**: Low — 검증 전용. 단 다른 패키지 cascading failure 발견 시 blocker report.
- **AC**: AC-MRR-011, AC-MRR-GREEN.

## §G. G2 Mirror 정책 Design 결정 (가장 어려운 부분 — 충분한 분석)

### G.1 문제 정의

`TestRuleTemplateMirrorDrift`/`TestLateBranchTemplateMirror` 는 source `.claude/...` 와 mirror `internal/template/templates/.claude/...` 의 **byte-for-byte 동일성**을 요구한다. 반면 CLAUDE.local.md §25 (+ §15) 는 mirror 가 internal content 를 제거한 **sanitized** 버전이어야 한다고 요구한다. 동일 파일이 양쪽 contract 에 걸리면 (실증: `manager-spec.md` — source 는 `SPEC-V3R6-SPEC-ID-VALIDATION-001`/`PR #1046` 보유, mirror 는 "the canonical agent responsibility realignment policy" 로 일부 sanitized) **internal content 를 가진 source 와 sanitized mirror 가 byte 동일할 수 없으므로 두 요구는 상호 모순**이다.

### G.2 3개 옵션 평가

| 옵션 | 메커니즘 | 장점 | 단점 | 평가 |
|------|---------|------|------|------|
| **(a) byte-parity wins** | mirror = source 의 정확한 복사. G3 leak 을 SOURCE 에서 sanitize 하여 source·mirror 양쪽 clean | mirror test 단순 유지, 단일 SSOT | **SOURCE 는 maintainer working copy 이며 internal content (SPEC-ID 추적 등)가 개발에 필요** — source sanitization 은 maintainer 의 추적성을 파괴. §25 의 전제(source ≠ mirror 가 의도된 격리)와 정면 충돌 | **미선택** — source 의 개발 추적성 손실이 치명적 |
| **(b) sanitization wins** | mirror-drift 테스트를 sanitization-aware 비교 (§25 substitution dictionary modulo diff) 로 교체/완화 | 모든 파일에 일관 적용, source 추적성 보존 | 테스트 로직 대폭 변경 (substitution-aware comparator 신규 구현) — 본 SPEC scope(main green 복구) 대비 over-engineering; comparator 자체가 새 결함 표면 | **미선택** — scope 과대, 신규 comparator 위험 |
| **(c) per-file 정책** | §25 sanitization 대상 파일은 byte-parity allowlist 에서 제외(→ leak test 로 커버); internal content 미보유 파일만 byte-parity 유지 | source 추적성 보존 + mirror sanitized 보존 + 테스트 로직 변경 최소(allowlist entry 제거만) + 두 contract 가 파일별로 disjoint 하게 적용되어 모순 제거 | allowlist 에 어느 파일을 넣고 뺄지 판단 필요 (leak 보유 여부로 결정) | **선택 (권장)** |

### G.3 권장: 옵션 (c) per-file 정책

**근거**:
1. **source 추적성 보존**: maintainer 의 `.claude/**` working copy 는 internal content (SPEC-ID, REQ-ID, PR#) 를 유지 — §25 의 의도된 source≠mirror 격리와 정합.
2. **테스트 로직 변경 최소**: `rule_template_mirror_test.go` 의 `workflowOptMirroredPaths`/`lateBranchMirroredPaths` slice 에서 sanitization 대상 entry 만 제거 (신규 comparator 불필요). 제거된 파일의 mirror 무결성은 `TestTemplateNoInternalContentLeak` (0 leak) 이 대신 보장.
3. **모순 제거**: byte-parity contract 와 sanitization contract 가 파일별로 disjoint 하게 적용 — 한 파일이 두 contract 에 동시에 걸리지 않음.
4. **scope 부합**: main green 복구라는 본 SPEC 목표에 최소 변경으로 도달.

**파일 분류 (ground-truth 확정 — leak test 보고서 대조 완료)**:

drift∩leak 교집합은 `manager-spec.md` **단독 1파일** 이다. leak test 가 보고한 30 site 중 mirror-drift allowlist 와 겹치는 유일한 파일이 `manager-spec.md` 다 (다른 7 drift 파일은 leak 0).

- byte-parity allowlist 에서 **제외** (leak 보유 → leak test 가 커버): **`manager-spec.md` 단독**. `lateBranchMirroredPaths` slice 에서 entry 제거.
- byte-parity **유지** (leak 0 = drift-only → cp resync): **7 파일** — `agent-common-protocol.md`, `manager-git.md`, `plan-auditor.md`, `verification-batch-pattern.md`, `manager-develop-prompt-template.md`, `ci-watch-protocol.md`, `agent-teams-pattern.md`. 이들은 모두 leak test 보고 0건이므로 source → mirror 복사로 byte 동일성 복원하고 allowlist 에 유지한다.

[HARD] `agent-common-protocol.md` 는 **exclusion 대상이 아니다** — leak 0 (drift-only) 이므로 byte-parity (cp-resync) 카테고리에 속한다. (직전 판본의 "manager-spec.md, agent-common-protocol.md 등" 예시는 오분류였으며 ground-truth 대조로 정정함.)

**검증 근거**: `go test -run TestTemplateNoInternalContentLeak` 보고서 대조 결과 — `agent-common-protocol.md` leak 0건 확인 (`grep -c 'agent-common-protocol' <leak-report>` → 0). M4 에서 8 drift 파일 각각의 leak 보유 여부를 동일 보고서로 재확인하여 allowlist in/out 을 evidence-based 로 유지 (사전 추측 금지).

## §H. Cross-References

- spec.md §C — Cross-Group Coupling (C1 G1↔G3 경로 2개 차원, C2 leak site #1 self-referential hazard, C3 G2↔G3 모순).
- acceptance.md §D — binary AC + 30 leak site enumeration (29 prose + 1 allowlist-path) + 8 drift file list (1 exclude + 7 cp-resync).
- `.claude/rules/moai/development/manager-develop-prompt-template.md` — Section A-E delegation 템플릿.
- CLAUDE.local.md §25 — substitution dictionary 출처.
- `internal/template/CLAUDE.md` — mirror parity + Template-First 규약.

## §I. Risks

- **R1 (G1↔G3 경로 불일치 — site #1 false-positive)**: M2 가 path 를 `moai/` 로 정정하나 M3 의 `pedagogicalAllowlist` `File:` 필드 (라인 233/240 실행 코드) 갱신을 누락하면 site #1 (manager-spec pedagogical 예제) 이 false-positive 로 계속 leak 보고. 완화: M3 에 실행 코드 차원 (allowlist `File:`) + 코멘트 차원 (203/254/337–338) 분리 명시 (C1).
- **R2 (G2 allowlist 오분류)**: leak 보유 파일을 byte-parity 에 남기거나 그 반대로 분류하면 둘 중 한 테스트 fail. 완화: M4 에서 leak test 보고서 대조로 evidence-based 결정.
- **R3 (make build 누락)**: template 수정 후 `embedded.go` 미재생성 시 embed 산출물 stale → embedded_namespace_test fail 가능. 완화: M2~M4 각각 후 `make build`, M5 에서 최종 재확인.
- **R4 (병렬 세션 race)**: 메모리상 main-RED remediation 수렴 중. 완화: pre-spawn fetch + 중복 SPEC 미존재 확인 (본 SPEC 이 유일 remediation).
