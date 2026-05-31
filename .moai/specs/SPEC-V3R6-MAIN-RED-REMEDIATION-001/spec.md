---
id: SPEC-V3R6-MAIN-RED-REMEDIATION-001
title: "internal/template main-RED 4-group 일괄 해소 — main green 복구"
version: "0.2.0"
status: implemented
created: 2026-05-30
updated: 2026-05-30
author: manager-spec
priority: P1
phase: "v3.0.0"
module: "internal/template"
lifecycle: spec-anchored
tags: "test-correction, template, mirror-drift, internal-content-leak, hook-count, main-red, ci-green"
tier: M
---

# internal/template main-RED 4-group 일괄 해소 — main green 복구

## HISTORY

| Version | Date | Author | Change |
|---------|------|--------|--------|
| 0.1.0 | 2026-05-30 | manager-spec | 최초 작성 — plan-phase 4개 group (agents-layout / mirror-drift / leak / hook-count) remediation SPEC |

## §A. Context (배경)

origin/main 이 5일째 committed-RED 상태이다. `golangci-lint run` 은 CLEAN (0 issues) 이지만 `go test ./internal/template/...` 패키지가 13개 parent 테스트로 fail 한다. 모든 failure 는 `internal/template` 패키지 한 곳에 국한되며, 4개 그룹으로 군집한다:

- **G1 (agents-layout stale tests, 9 failing)**: 테스트가 `.claude/agents/{core,meta}` 서브폴더 구조를 기대하나, 현재 canonical 레이아웃은 `.claude/agents/moai/` FLAT (7 retained agents) 이다. 근본 원인은 `{core,meta}` split 을 도입했던 선행 SPEC 이 `superseded` 상태로 전이되었고 후속 SPEC 이 FLAT `moai/` 레이아웃을 canonical 로 복원했음에도 테스트가 superseded SPEC 의 기대값을 그대로 carry 하고 있기 때문이다. **테스트가 stale 하므로 테스트를 정정한다** — 디렉토리 구조를 생성하지 않는다.
- **G2 (template-mirror-drift, 2 parent / 8 subtest)**: source `.claude/...` 파일과 mirror `internal/template/templates/.claude/...` 파일이 byte 차이가 난다. 8개 파일이 drift 상태이다.
- **G3 (internal-content-leak, 1 test / 30 violation)**: production template 파일에 30개 internal-content leak (SPEC-ID / REQ-ID / AC-ID / archive date / backup path) 이 잔존한다. 선행 sanitization SPEC 이 `completed` 로 종료되었으나 부분적으로만 적용/regression 되었다.
- **G4 (hook-count, 1 test)**: `hook event count = 21, want 20` — 21번째 event `PreCommit` 가 선행 lifecycle SPEC M3 에서 추가되었으나 테스트의 기대 상수가 20 으로 stale 하다.

본 SPEC 은 4개 그룹을 모두 해소하여 `main green` 을 복구한다.

### A.1 검증된 ground-truth 증거

- 13 parent fail: `TestContractSchemaVerification`, `TestBackwardCompatibility`, `TestContractAssertionsNaturalLanguage`, `TestTemplateNoInternalContentLeak`, `TestAgentFrontmatterAudit`, `TestBuilderSkillPathStructure`, `TestManagerDevelopActiveAgentPresent`, `TestManagerDevelopIsActiveAgent`, `TestEmbeddedTemplates_AgentDefinitions`, `TestTemplateAgentsStructure`, `TestSettingsTemplateHookEventCount`, `TestLateBranchTemplateMirror`, `TestRuleTemplateMirrorDrift`.
- canonical 레이아웃: source `.claude/agents/moai/` + template `internal/template/templates/.claude/agents/moai/` 양쪽 모두 FLAT, 7 파일 (builder-harness, evaluator-active, manager-develop, manager-docs, manager-git, manager-spec, plan-auditor). `core/` / `meta/` / `expert/` 디렉토리는 어디에도 존재하지 않음.
- G2↔G3 충돌 실증: `manager-spec.md` 가 G2 drift 8파일 + G3 leak 30사이트 양쪽에 동시 등장. SOURCE 는 internal content 보유 (working copy), MIRROR 는 일부 sanitized 됨 → byte-parity 와 sanitization 이 동일 파일에서 상호 모순.

## §B. 요구사항 (GEARS Format)

### B.1 G1 — agents-layout 테스트 정정 (REQ-MRR-001 ~ REQ-MRR-003)

- **REQ-MRR-001** (Ubiquitous): The agents-layout 테스트 스위트 shall expect the canonical FLAT layout `.claude/agents/moai/` 를 path 기대값으로 사용하며, retired `{core,meta,expert}` 서브폴더 기대값을 더 이상 carry 하지 않는다.

- **REQ-MRR-002** (Event-driven): **When** 테스트가 retained agent 카탈로그를 enumerate 할 때, the 테스트 shall reflect the 7 retained agents (builder-harness, evaluator-active, manager-develop, manager-docs, manager-git, manager-spec, plan-auditor) 를 FLAT `moai/` 디렉토리 하나에서 읽는 형태로 검증한다.

- **REQ-MRR-003** (Unwanted behavior): The remediation shall not 디렉토리 구조를 생성/이동하여 테스트를 통과시킨다 (no `core/` or `meta/` directory creation). 정정은 테스트 기대값을 canonical 현실에 맞추는 방식으로만 수행한다.

### B.2 G4 — hook-count expected-value 정정 (REQ-MRR-004)

- **REQ-MRR-004** (Event-driven): **When** `TestSettingsTemplateHookEventCount` 가 rendered settings 의 hook event 수를 검증할 때, the 테스트 shall expect 21 events (20 기존 + `PreCommit` 1) 이며, 기대 상수 정정에는 `PreCommit` 추가 출처를 명시하는 HISTORY 주석을 동반한다.

### B.3 G3 — internal-content leak sanitization (REQ-MRR-005 ~ REQ-MRR-007)

- **REQ-MRR-005** (Ubiquitous): The production template 파일 shall contain zero internal-content leak — `TestTemplateNoInternalContentLeak` 가 0 occurrence 를 보고해야 한다. 30개 site 는 **29 sites prose-substituted + 1 site (site #1 manager-spec pedagogical 예제) allowlist-path-corrected** 의 두 경로로 해소되며, site #1 은 pedagogical teaching 예제이므로 prose 치환을 적용하지 않는다 (§C.2 참조).

- **REQ-MRR-006** (Event-driven): **When** 29개 prose-substitution-대상 leak 토큰을 sanitize 할 때, the remediation shall apply the canonical generic-prose substitution dictionary (선행 isolation SPEC design §B 의 6-rule S1~S6) 를 적용하여 SPEC-ID / REQ-ID / AC-ID / archive-date / backup-path 를 generic prose 로 치환한다. **When** site #1 (manager-spec pedagogical 예제) 을 처리할 때, the remediation shall NOT 치환을 적용하고 대신 `pedagogicalAllowlist` 의 `File:` 경로를 `core/`→`moai/` 로 정정하여 pedagogical 면제를 회복한다.

- **REQ-MRR-007** (Where, capability gate): **Where** leak 잔존이 사용자에게 deploy 되는 `internal/template/templates/**` 경로에 한정될 때, the remediation shall sanitize the mirror 산출물을 우선 대상으로 하며, source `.claude/**` working copy 의 internal content 보존 여부는 §B.4 의 mirror 정책 결정에 종속한다.

### B.4 G2 — mirror-drift 해소 + design 결정 (REQ-MRR-008 ~ REQ-MRR-010)

- **REQ-MRR-008** (Ubiquitous): The mirror-drift 테스트 (`TestRuleTemplateMirrorDrift` + `TestLateBranchTemplateMirror`) shall pass — 각 allowlist 항목의 source 와 mirror 가 정책에 부합해야 한다.

- **REQ-MRR-009** (State-driven): **While** 동일 파일이 §25 sanitization 대상이면서 동시에 mirror byte-parity allowlist 에 등록된 상태일 때, the remediation shall resolve the contradiction via plan.md §G 에서 채택한 단일 mirror 정책 (byte-parity-wins / sanitization-wins / per-file) 에 따라 source 와 mirror, 그리고 해당 테스트를 일관되게 정렬한다.

- **REQ-MRR-010** (Event-driven): **When** template 산출물이 수정될 때, the remediation shall run `make build` 로 `embedded.go` 를 재생성하여 source 변경이 embed 산출물에 반영되도록 한다 (Template-First rule).

### B.5 통합 green gate (REQ-MRR-011)

- **REQ-MRR-011** (Event-driven): **When** 4개 그룹 정정이 모두 완료될 때, the remediation shall verify that `go test ./internal/template/...` 가 0 fail 로 통과하고 `GOOS=windows GOARCH=amd64 go build ./...` 가 exit 0 이며, golangci-lint baseline (0 issues) 이 유지됨을 확인한다.

## §C. Cross-Group Coupling (그룹 간 결합 위험)

[HARD] 본 SPEC 의 가장 중요한 위험은 **G1 ↔ G3 경로 결합** 과 **G2 ↔ G3 byte-parity vs sanitization 모순** 두 가지이다.

### C.1 G1 ↔ G3 경로 결합 (leak test 의 `core/` 참조 — 2개 차원)

`internal_content_leak_test.go` 의 `core/` 참조는 **균일하지 않다**. 2개 차원으로 분리된다:

- **코멘트 차원 (라인 203 / 254 / 337–338)**: docstring/설명 코멘트로 `.claude/agents/core/manager-spec.md` 를 예시 경로로 인용. 갱신해도 실행에 무해(harmless)하나 일관성을 위해 `moai/` 로 정정.
- **실행 코드 차원 (라인 233 / 240)**: `pedagogicalAllowlist` 구조체 (라인 210 정의) 의 **`File:` 필드** 로 `.claude/agents/core/manager-spec.md` 를 키로 사용. 이 필드는 `isPedagogicallyAllowed(relPath, matched)` (라인 257) 가 소비한다. canonical 레이아웃은 `moai/` FLAT 이므로, M2 가 경로를 `moai/` 로 바꾸면 이 allowlist `File:` 키 (`core/`) 가 실제 mirror 경로 (`moai/`) 와 매칭 실패 → pedagogical 면제가 작동하지 않아 leak site #1 (manager-spec 의 self-check 예제 SPEC-ID) 이 false-positive 로 계속 mis-fire 한다.

[HARD] 따라서 G3 milestone(M3)은 다음을 **명시적으로 분리**하여 처리한다: (a) `pedagogicalAllowlist` 구조체 `File:` 필드 `core/`→`moai/` 정정 (실행 코드 차원) + (b) 코멘트 라인 203/254/337–338 `core/`→`moai/` 정정 (무해, 일관성). 두 그룹은 동일 테스트 파일(`internal_content_leak_test.go`)을 공유하므로 M3 에서 atomic 하게 처리한다.

### C.2 leak site #1 self-referential hazard (manager-spec 의 pedagogical 예제)

[HARD] leak site #1 = `manager-spec.md` 의 `SPEC-V3R6-SPEC-ID-VALIDATION-001` 토큰이다. ground-truth: 이 토큰은 source 와 mirror **양쪽 모두** (각 2회, 라인 ~146/161) 에 등장하며, **SPEC ID regex Pre-Write Self-Check Protocol 의 pedagogical 예제** (decomposition 출력 예시) 다. 이것은 leak test 의 `pedagogicalAllowlist` 가 합법화(legalize)하려던 바로 그 항목이지만, allowlist 키가 stale (`.claude/agents/core/manager-spec.md`) 이어서 실제 mirror 경로 (`.claude/agents/moai/manager-spec.md`) 와 매칭 실패 → site #1 로 false-positive 표면화한다.

[HARD] 따라서 site #1 은 **substitution dictionary 대상이 아니다**. prose 치환을 적용하면 manager-spec 의 regex teaching 예제 (`SPEC-AUTH-001` / `SPEC-V3R6-SPEC-ID-VALIDATION-001`) 를 파괴한다. site #1 은 "allowlist `File:` 경로 정정 (core/→moai/) 으로 pedagogical 면제 회복" 으로 재분류한다 (C.1 실행 코드 차원과 동일 작업). 순효과: 30개 leak site 는 **29 sites prose-substituted + 1 site allowlist-path-corrected** 로 분해된다.

### C.3 G2 ↔ G3 byte-parity vs sanitization 모순

`manager-spec.md` 가 G2 drift 8파일 + G3 leak 30사이트 양쪽에 동시 등장한다. SOURCE 는 internal content (SPEC-ID, REQ-ID, PR #) 를 보유한 working copy 이고, MIRROR 는 §25 정책상 sanitized 되어야 한다. `TestRuleTemplateMirrorDrift`/`TestLateBranchTemplateMirror` 는 byte-for-byte 동일성을 요구한다. 이 두 요구사항은 **internal content 를 가진 source 와 sanitized mirror 가 byte 동일할 수 없으므로 상호 모순**이다. plan.md §G 의 design 결정으로 단일 정책을 채택하여 해소한다 (권장: per-file 정책 — §25 sanitization 대상 파일은 byte-parity allowlist 에서 제외하고 leak test 로 커버; 그 외 파일은 byte-parity 유지). drift∩leak 교집합은 `manager-spec.md` **단독 1파일** 이다 (나머지 7 drift 파일은 leak 0 = cp-resync 대상; plan.md §G.3 참조).

## §D. Exclusions (What NOT to Build)

### D.1 Out of Scope

- **golangci-lint 작업**: lint 는 이미 CLEAN (0 issues) 이다. lint 관련 어떤 변경도 본 SPEC 범위 밖이다.
- **agents 디렉토리 구조 변경**: `core/` / `meta/` / `expert/` 서브폴더를 생성하거나 agent 파일을 이동하는 어떤 작업도 금지 (REQ-MRR-003). G1 은 순수 테스트 정정이다.
- **3개 미완료 V3R5 config-schema SPEC (GIT-STRATEGY-SCHEMA / WORKFLOW-SCHEMA-EXTEND / INIT-WIZARD-EXPANSION)**: 본 SPEC 과 독립이며 entangle 하지 않는다.
- **CLAUDE.local.md §25 doctrine 본문 개정**: §25 정책 자체의 재설계는 범위 밖이다. 본 SPEC 은 §25 가 정의한 substitution dictionary 를 적용만 한다.
- **leak 검출 엔진(`internal_content_leak_test.go` 의 class 정규식) 확장**: 새 leak class 추가는 범위 밖이다. 기존 30 site 를 0 으로 만드는 것만 수행한다.
- **manager-git PR phase**: Hybrid Trunk 1-person OSS 정책상 Tier M 은 main 직진 push 이다. 별도 PR phase 불필요.

### D.2 비목표 (Non-Goals)

- 다른 패키지의 테스트 failure 해소 (본 SPEC 은 `internal/template` 패키지에만 국한).
- mirror 정책의 영구 아키텍처 재설계 (본 SPEC 은 main green 복구를 위한 최소 일관성 결정만 수행; 영구 정책 표준화는 후속 SPEC 후보).
- agent body 내용 자체의 개선/리라이팅 (G2 resync 는 byte 정렬 또는 sanitization 정렬만 수행, 내용 개선 아님).

## §E. Acceptance Criteria (요약)

상세 binary AC 는 `acceptance.md` 참조. 핵심 closure gate:

- AC-MRR-G1: G1 9개 테스트 PASS (FLAT `moai/` 기대값).
- AC-MRR-G2: G2 mirror-drift 2 parent 테스트 PASS (정책 적용 후).
- AC-MRR-G3: `TestTemplateNoInternalContentLeak` 0 occurrence.
- AC-MRR-G4: `TestSettingsTemplateHookEventCount` PASS (21 expected).
- AC-MRR-GREEN: `go test ./internal/template/...` 0 fail + cross-platform build exit 0 + lint baseline 유지.

## §F. Cross-References

- plan.md §G — G2 mirror 정책 design 결정 (3 option 평가 + 권장).
- acceptance.md §D — binary AC matrix + 30 leak site enumeration.
- `.claude/rules/moai/development/spec-frontmatter-schema.md` — frontmatter 12-field SSOT.
- CLAUDE.local.md §25 — Template Internal-Content Isolation doctrine (substitution dictionary 출처).
- CLAUDE.local.md §15 / internal/template/CLAUDE.md — 16-language neutrality + mirror parity 규약.
