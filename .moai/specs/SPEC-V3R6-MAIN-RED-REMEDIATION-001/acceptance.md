---
id: SPEC-V3R6-MAIN-RED-REMEDIATION-001
title: "Acceptance — internal/template main-RED 4-group 일괄 해소"
version: "0.1.0"
status: draft
created: 2026-05-30
updated: 2026-05-30
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/template"
lifecycle: spec-anchored
tags: "test-correction, template, mirror-drift, internal-content-leak, hook-count, main-red, acceptance"
tier: M
---

# Acceptance — internal/template main-RED 4-group 일괄 해소

## §A. Definition of Done

4개 그룹 정정 후 `go test ./internal/template/...` 가 0 fail 이고, cross-platform build 가 exit 0 이며, golangci-lint baseline (0 issues) 이 유지된다. 13 parent fail 이 모두 PASS 로 전환된다.

## §B. Given-When-Then 시나리오

### Scenario 1 — G4 hook-count 정정 (최소 변경 검증)

- **Given**: `settings_test.go:512` 의 `const expectedCount = 20` 이 stale 하여 `TestSettingsTemplateHookEventCount` 가 `hook event count = 21, want 20` 으로 fail 한다.
- **When**: 상수를 21 로 정정하고 `PreCommit` 추가 출처 HISTORY 주석을 동반한다.
- **Then**: `go test ./internal/template/... -run TestSettingsTemplateHookEventCount -v` 가 PASS 하고, rendered settings 의 21 event 가 모두 매칭된다.

### Scenario 2 — G1 agents-layout FLAT 정정 (구조 변경 없음)

- **Given**: 8 테스트 파일이 `.claude/agents/{core,meta,expert}` 서브폴더를 기대하나 canonical 레이아웃은 `.claude/agents/moai/` FLAT (7 retained agents) 이다.
- **When**: 디렉토리를 생성/이동하지 않고 테스트의 path 기대값 + domain enum + count threshold 를 FLAT `moai/` (7 retained) 으로 정정한다.
- **Then**: 9 테스트 (`TestTemplateAgentsStructure`, `TestContractSchemaVerification`, `TestContractAssertionsNaturalLanguage`, `TestBackwardCompatibility`, `TestAgentFrontmatterAudit`, `TestEmbeddedTemplates_AgentDefinitions`, `TestManagerDevelopActiveAgentPresent`, `TestManagerDevelopIsActiveAgent`, `TestBuilderSkillPathStructure`) 가 PASS 하고, `.claude/agents/` 디렉토리에는 여전히 `core/`/`meta/`/`expert/` 가 존재하지 않는다.

### Scenario 3 — G3 leak 해소 (30 → 0: 29 prose-substitution + 1 allowlist-path-correction)

- **Given**: production template 에 30개 internal-content leak (§D.2 enumeration) 이 잔존하여 `TestTemplateNoInternalContentLeak` 가 fail 한다. 단 site #1 (`manager-spec.md` 의 `SPEC-V3R6-SPEC-ID-VALIDATION-001`) 은 SPEC ID regex Pre-Write Self-Check pedagogical 예제로서 `pedagogicalAllowlist` 가 합법화하려던 항목이나, allowlist 키가 stale (`core/`) 이어서 false-positive 로 표면화한 것이다.
- **When**: **29 sites** 는 substitution dictionary S1~S6 를 적용하여 generic prose 로 치환한다. **site #1** 은 prose 치환을 적용하지 **않고**, `pedagogicalAllowlist` 의 `File:` 경로를 `core/`→`moai/` 로 정정하여 pedagogical 면제를 회복한다 (실행 코드 차원 라인 233/240, C1). 추가로 leak test 의 코멘트 차원 carrier 경로 예시 (라인 203/254/337–338) 도 `moai/` 로 일관 정정한다.
- **Then**: `go test ./internal/template/... -run TestTemplateNoInternalContentLeak -v` 가 0 occurrence 로 PASS 하며, manager-spec 의 regex teaching 예제 (`SPEC-AUTH-001` / `SPEC-V3R6-SPEC-ID-VALIDATION-001` decomposition 출력) 가 파괴되지 않고 보존된다.

### Scenario 4 — G2 mirror-drift per-file 정책 적용

- **Given**: 8개 파일이 source 와 mirror 간 byte drift 상태이다. ground-truth: drift∩leak 교집합은 `manager-spec.md` **단독 1파일** (나머지 7 파일은 leak 0 = drift-only) 이며, `manager-spec.md` 는 §25 sanitization 대상이면서 동시에 byte-parity allowlist 에 있어 두 contract 가 모순한다.
- **When**: plan.md §G 의 per-file 정책을 적용한다 — **manager-spec.md (단독)** 는 byte-parity allowlist 에서 제외(leak test 가 커버), **나머지 7 drift 파일** (`agent-common-protocol.md` 포함, 모두 leak 0) 은 source→mirror cp resync 후 allowlist 유지. 이후 `make build`.
- **Then**: `go test ./internal/template/... -run 'TestRuleTemplateMirrorDrift|TestLateBranchTemplateMirror' -v` 가 PASS 하고, manager-spec.md mirror 는 0 leak (Scenario 3) 으로 보장되며, 나머지 7 파일은 source 와 mirror 가 byte 동일하다.

### Scenario 5 — 통합 green gate

- **Given**: M1~M4 정정이 완료된 상태.
- **When**: 전체 패키지 테스트 + cross-platform build + lint baseline 을 검증한다.
- **Then**: `go test ./internal/template/...` 0 fail + `GOOS=windows GOARCH=amd64 go build ./...` exit 0 + `golangci-lint run` 0 issues + 다른 패키지 cascading failure 없음.

## §C. Edge Cases

- **EC-1 (make build 후 embedded.go drift)**: template 수정 후 `make build` 를 실행하지 않으면 `embedded.go` 가 stale 하여 `TestTemplateAgentsStructure` (embedded FS 검사) 가 여전히 fail 할 수 있다. 각 milestone 후 `make build` 의무.
- **EC-2 (leak test가 새 leak 탐지)**: sanitization 중 generic prose 가 우연히 다른 leak class 정규식에 매칭되면 새 leak 보고. 치환 후 즉시 leak test 재실행으로 확인.
- **EC-3 (allowlist 오분류 미검출)**: M4 에서 leak 보유 파일을 byte-parity 에 남기면 mirror-drift test 가 영구 fail. leak test 보고서 대조로 evidence-based 분류 의무.
- **EC-4 (병렬 세션 중복 commit)**: 메모리상 main-RED remediation 수렴 중. pre-spawn fetch + mystery commit 확인.

## §D. Binary AC Matrix

### D.1 AC 목록

| AC | Group | Binary 기준 | Verification Command |
|----|-------|-------------|----------------------|
| AC-MRR-001 | G1 | `TestTemplateAgentsStructure` + `TestContractSchemaVerification` PASS (FLAT moai 기대) | `go test ./internal/template/... -run 'TestTemplateAgentsStructure\|TestContractSchemaVerification' -v` |
| AC-MRR-002 | G1 | `TestContractAssertionsNaturalLanguage` + `TestBackwardCompatibility` PASS (domain enum = moai) | `go test ./internal/template/... -run 'TestContractAssertionsNaturalLanguage\|TestBackwardCompatibility' -v` |
| AC-MRR-003a | G1 | `TestAgentFrontmatterAudit` + `TestEmbeddedTemplates_AgentDefinitions` PASS | `go test ./internal/template/... -run 'TestAgentFrontmatterAudit\|TestEmbeddedTemplates_AgentDefinitions' -v` |
| AC-MRR-003b | G1 | `TestManagerDevelopActiveAgentPresent` + `TestManagerDevelopIsActiveAgent` + `TestBuilderSkillPathStructure` PASS | `go test ./internal/template/... -run 'TestManagerDevelopActiveAgentPresent\|TestManagerDevelopIsActiveAgent\|TestBuilderSkillPathStructure' -v` |
| AC-MRR-003c | G1 | `.claude/agents/` 에 core/meta/expert 디렉토리 미생성 (구조 불변) | `ls .claude/agents/ \| grep -E 'core\|meta\|expert' ; echo "exit=$?"` (no match expected) |
| AC-MRR-004 | G4 | `TestSettingsTemplateHookEventCount` PASS (expectedCount=21) | `go test ./internal/template/... -run TestSettingsTemplateHookEventCount -v` |
| AC-MRR-005 | G3 | `TestTemplateNoInternalContentLeak` 0 occurrence | `go test ./internal/template/... -run TestTemplateNoInternalContentLeak -v` |
| AC-MRR-006a | G3 | 29개 prose-substitution-대상 site 모두 substitution dictionary 적용 (site #1 제외, 0 잔존) | `go test ./internal/template/... -run TestTemplateNoInternalContentLeak -v 2>&1 \| grep -c 'class='` → 0 |
| AC-MRR-006b | G3 | site #1 (manager-spec pedagogical 예제) 은 prose 치환 NOT — manager-spec.md 의 SPEC-ID self-check 예제 토큰 보존 | `grep -c 'SPEC-V3R6-SPEC-ID-VALIDATION-001\|SPEC-AUTH-001' .claude/agents/moai/manager-spec.md` ≥ 2 (teaching 예제 보존) |
| AC-MRR-007 | G3↔G1 | leak test 의 `pedagogicalAllowlist` `File:` 필드 + 코멘트가 `agents/moai/manager-spec.md` 로 정정 (core/ 미참조) | `grep -c 'agents/core/manager-spec.md' internal/template/internal_content_leak_test.go` → 0 |
| AC-MRR-008 | G2 | `TestRuleTemplateMirrorDrift` PASS | `go test ./internal/template/... -run TestRuleTemplateMirrorDrift -v` |
| AC-MRR-009 | G2 | `TestLateBranchTemplateMirror` PASS | `go test ./internal/template/... -run TestLateBranchTemplateMirror -v` |
| AC-MRR-010 | G2 | per-file 정책 적용 — manager-spec.md 단독 allowlist 제외 주석 존재 | `grep -c 'per-file\|§25 sanitization' internal/template/rule_template_mirror_test.go` ≥ 1 |
| AC-MRR-011a | green | `make build` 후 build 성공 | `make build && go build ./...` → exit 0 |
| AC-MRR-011b | green | `embedded.go` 가 실제로 drift-free 재생성됨 (make build exit 0 만으로는 재생성 미입증) | `make build && git diff --exit-code internal/template/embedded.go` → exit 0 (재생성 후 staged 변경분이 commit 에 포함됨을 별도 확인; embedded.go 가 source 변경과 일치) |
| AC-MRR-GREEN | green | `go test ./internal/template/...` 0 fail + cross-platform build + lint baseline | §D.3 batch 참조 |

### D.2 G3 — 30 Leak Site Enumeration (baseline 캡처)

manager-develop 은 **site #2~#30 (29개)** 를 substitution dictionary 로 치환하고, **site #1 은 `pedagogicalAllowlist` `File:` 경로 정정 (core/→moai/)** 으로 해소하여 전체 0 잔존을 검증한다. [HARD] site #1 은 prose 치환 대상이 **아니다** (pedagogical 예제 보존 — §spec.md C.2):

| # | File | class | match token | 해소 방식 |
|---|------|-------|-------------|----------|
| 1 | templates/.claude/agents/moai/manager-spec.md | C1-spec-id-prefix | SPEC-V3R6-SPEC-ID-VALIDATION-001 | **allowlist-path-correction** (prose 치환 NOT — pedagogical 예제) |
| 2 | templates/.claude/output-styles/moai/moai.md | C1-spec-id-prefix | SPEC-V3R6-XXX-001 |
| 3 | templates/.claude/output-styles/moai/moai.md | C1-spec-id-prefix | SPEC-V3R6-MULTI-SESSION-COORD-001 |
| 4 | templates/.claude/output-styles/moai/moai.md | C2-req-ac-internal-prefix | REQ-TII-001 |
| 5 | templates/.claude/output-styles/moai/moai.md | C2-req-ac-internal-prefix | AC-TII-007 |
| 6 | templates/.claude/output-styles/moai/moai.md | C2-req-ac-internal-prefix | REQ-COORD-010 |
| 7 | templates/.claude/rules/moai/development/agent-authoring.md | C1-spec-id-prefix | SPEC-V3R6-AGENT-TEAM-REBUILD-001 |
| 8 | templates/.claude/rules/moai/development/agent-authoring.md | C4-finding-or-internal-archive-date | Finding A3 |
| 9 | templates/.claude/rules/moai/development/agent-authoring.md | C4-finding-or-internal-archive-date | archive-2026-05-25 |
| 10 | templates/.claude/rules/moai/development/agent-authoring.md | C5-memory-archive-path | .moai/backups/agent-archive- |
| 11 | templates/.claude/rules/moai/development/spec-frontmatter-schema.md | C1-spec-id-prefix | SPEC-V3R6-AGENT-RESPONSIBILITY-REALIGN-001 |
| 12 | templates/.claude/rules/moai/development/spec-frontmatter-schema.md | C1-spec-id-prefix | SPEC-V3R6-AGENT-TEAM-REBUILD-001 |
| 13 | templates/.claude/rules/moai/development/spec-frontmatter-schema.md | C1-spec-id-prefix | SPEC-V3R6-ANTHROPIC-AUDIT-TIER3-001 |
| 14 | templates/.claude/rules/moai/workflow/archived-agent-rejection.md | C1-spec-id-prefix | SPEC-V3R6-AGENT-TEAM-REBUILD-001 |
| 15 | templates/.claude/rules/moai/workflow/archived-agent-rejection.md | C2-req-ac-internal-prefix | REQ-ATR-009 |
| 16 | templates/.claude/rules/moai/workflow/archived-agent-rejection.md | C2-req-ac-internal-prefix | REQ-ATR-014 |
| 17 | templates/.claude/rules/moai/workflow/archived-agent-rejection.md | C2-req-ac-internal-prefix | AC-ATR-005 |
| 18 | templates/.claude/rules/moai/workflow/archived-agent-rejection.md | C2-req-ac-internal-prefix | REQ-ATR-015 |
| 19 | templates/.claude/rules/moai/workflow/archived-agent-rejection.md | C2-req-ac-internal-prefix | AC-ATR-016 |
| 20 | templates/.claude/rules/moai/workflow/archived-agent-rejection.md | C4-finding-or-internal-archive-date | archive-2026-05-25 |
| 21 | templates/.claude/rules/moai/workflow/archived-agent-rejection.md | C4-finding-or-internal-archive-date | Finding A5 |
| 22 | templates/.claude/rules/moai/workflow/archived-agent-rejection.md | C4-finding-or-internal-archive-date | Finding A6 |
| 23 | templates/.claude/rules/moai/workflow/archived-agent-rejection.md | C4-finding-or-internal-archive-date | Finding A1 |
| 24 | templates/.claude/rules/moai/workflow/archived-agent-rejection.md | C5-memory-archive-path | .moai/backups/agent-archive- |
| 25 | templates/.claude/rules/moai/workflow/spec-workflow.md | C1-spec-id-prefix | SPEC-V3R6-AGENT-TEAM-REBUILD-001 |
| 26 | templates/.claude/rules/moai/workflow/spec-workflow.md | C4-finding-or-internal-archive-date | archive-2026-05-25 |
| 27 | templates/.claude/rules/moai/workflow/spec-workflow.md | C5-memory-archive-path | .moai/backups/agent-archive- |
| 28 | templates/.claude/skills/moai/workflows/harness.md | C1-spec-id-prefix | SPEC-V3R6-HARNESS-CLASSIFIER-WIRING-001 |
| 29 | templates/.claude/skills/moai/workflows/plan/spec-assembly.md | C1-spec-id-prefix | SPEC-V3R6-GEARS-MIGRATION-001 |
| 30 | templates/.moai/config/sections/cache.yaml | C1-spec-id-prefix | SPEC-V3R6-PROMPT-CACHE-001 |

> **해소 방식 (site #2~#30)**: 위 29개 site 는 모두 **prose-substitution** (substitution dictionary S1~S6 적용) 으로 해소한다. **site #1 만** allowlist-path-correction (위 표 #1 행 참조). 30 → 0 분해: **29 prose-substituted + 1 allowlist-path-corrected**.

### D.3 통합 green gate batch (AC-MRR-GREEN)

```bash
# 모두 PASS 의무
go test ./internal/template/...               # 0 fail (13 parent → 0)
go build ./...                                # exit 0
GOOS=windows GOARCH=amd64 go build ./...      # exit 0
golangci-lint run --timeout=2m                # 0 issues (baseline 불변)
go test ./... 2>&1 | grep -c '^FAIL'          # 0 (cascading 없음)
```

## §E. Quality Gate Criteria

- [ ] 13 parent fail → 0 fail (`go test ./internal/template/...`)
- [ ] cross-platform build exit 0 (linux + windows/amd64)
- [ ] golangci-lint 0 issues (baseline 불변 — 본 SPEC 이 새 lint issue 도입하지 않음)
- [ ] `.claude/agents/` 구조 불변 (core/meta/expert 미생성)
- [ ] 30 leak site → 0 (29 prose-substituted + 1 allowlist-path-corrected)
- [ ] site #1 (manager-spec pedagogical 예제) prose 치환 NOT — teaching 예제 보존 (AC-MRR-006b)
- [ ] `embedded.go` drift-free 재생성 검증 (`make build && git diff --exit-code internal/template/embedded.go`)
- [ ] G1↔G3 경로 결합 일관 정정 — `pedagogicalAllowlist` `File:` (실행 코드) + 코멘트 양쪽 core/ 미참조
- [ ] G2 per-file 정책 적용 (manager-spec.md 단독 exclude + 7 cp-resync; mirror 모순 해소)
- [ ] 무관 untracked 파일 미포함 commit (specific path staging)

## §E.1 AC 비대상

### Out of Scope

- golangci-lint 관련 AC (이미 0 issues — baseline 불변 확인만 AC-MRR-GREEN 에 포함, 신규 lint AC 없음).
- `.claude/agents/` 디렉토리 구조 생성 검증 (G1 은 테스트 정정 — AC-MRR-003c 는 미생성 불변만 검증).
- leak class 정규식 확장 AC (기존 30 site → 0 만 검증, 새 class 추가 없음).
- mirror 정책 영구 표준화 AC (per-file 적용 결과만 검증; 영구 아키텍처는 후속 SPEC).

## §F. Cross-References

- spec.md §B (REQ-MRR-001~011) — 본 acceptance 가 검증하는 요구사항.
- spec.md §C — Cross-Group Coupling (C1 경로 2차원 / C2 site #1 hazard / C3 G2↔G3 모순).
- plan.md §F (M1~M5) — milestone 별 verify 명령.
- plan.md §G — G2 mirror 정책 design 결정 (per-file 권장).
