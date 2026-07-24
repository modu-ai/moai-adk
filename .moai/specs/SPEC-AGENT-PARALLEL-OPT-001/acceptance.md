---
id: SPEC-AGENT-PARALLEL-OPT-001
title: "Agent instruction diet + plan/run/sync parallelization maximization — Acceptance Criteria"
version: "0.2.0"
status: draft
created: 2026-07-25
updated: 2026-07-25
author: manager-spec
priority: P1
phase: "v3.1.0 target"
module: ".claude/agents/moai, .claude/skills/moai/workflows, .claude/rules/moai/core, internal/template/templates"
lifecycle: spec-anchored
tags: "agent-diet, parallelization, fan-out, write-concurrency, workflow-wiring, template-first"
tier: L
---

## §A 판정 원칙

- 모든 AC는 **기계 검증 가능**하거나(grep / diff / `wc -l` / `go test`), 명시적 구조 검사로 판정한다.
- 판정 근거는 명령의 **verbatim 출력**이어야 한다. 요약·추정·이월 수치는 근거로 인정하지 않는다.
- 라인 수 AC는 `wc -l` 출력의 정확한 정수로 판정한다. 근사치("약 340줄") 금지.
- MUST 등급 AC 1건이라도 FAIL이면 SPEC은 close 불가다.

---

## §B Given-When-Then 시나리오

### 시나리오 1 — 겹치지 않는 경로에 대한 동시 write spawn 허용

**Given** `agent-common-protocol.md`가 스코프 한정 형태로 개정되어 있고,
**And** 오케스트레이터가 sync Phase 12에서 CHANGELOG 담당과 project-docs 담당 두 write-capable 에이전트를 준비했고,
**And** 두 spawn 프롬프트가 각각 disjoint path manifest를 선언했으며 두 manifest의 교집합이 공집합일 때,
**When** 오케스트레이터가 두 에이전트를 동시 spawn하면,
**Then** 규칙 위반이 아니며,
**And** 두 에이전트가 쓴 파일 집합은 서로 겹치지 않고,
**And** 오케스트레이터 자신은 그 구간 동안 read-only 작업만 수행한다.

### 시나리오 2 — 겹치는 경로 선언 시 직렬화

**Given** 두 write-capable spawn의 manifest가 `README.md`를 공통으로 포함할 때,
**When** 오케스트레이터가 동시 실행을 검토하면,
**Then** 동시 spawn을 수행하지 않고 직렬로 실행하며,
**And** 직렬화 사유를 기록한다.

### 시나리오 3 — 런타임 미지원 환경에서의 graceful degradation

**Given** 배포 사용자가 3개 스크립트 파일은 수령했으나 런타임이 dynamic workflow를 지원하지 않고(구버전 Claude Code),
**And** `plan.md`가 `plan-research-fanout.js`를 capability-gate 형태로 참조할 때,
**When** 사용자가 `/moai plan "<feature>"`를 실행하면,
**Then** 워크플로우는 기존 단일 Explore 리서치 경로로 진행되고,
**And** 오류·경고·중단이 발생하지 않으며,
**And** 산출되는 SPEC 산출물 집합은 배선 이전과 동일하다.

파일 부재 환경(스크립트를 수동 삭제한 사용자)에서도 동일 결과여야 한다 — gate는 "파일 존재 AND 런타임 지원" 두 조건을 모두 확인한다.

### 시나리오 7 — 배포가 harness Runner 격리를 깨뜨리지 않음

**Given** 3개 generic fan-out 스크립트가 `internal/template/templates/.claude/workflows/`에 배포되어 있을 때,
**When** `go test ./internal/template/...`를 실행하면,
**Then** `TestSplitHarnessNamespaceNoLeak`이 통과하고,
**And** `hns-release-update-run.js` 등 dev-only Runner를 템플릿에 심으면 여전히 FAIL한다(차단 유효성 확인),
**And** `moai update`가 3개 스크립트를 template-managed로, 사용자의 `hns-*` Runner를 user-owned로 분류한다.

### 시나리오 8 — 중립성 판정이 공허하지 않음

**Given** `leakTextExtensions`에 `.js`가 추가되어 있을 때,
**When** 중립화되지 않은 스크립트(`REQ-ATR-018` 등 포함)를 템플릿에 심고 leak 테스트를 실행하면,
**Then** 테스트가 **FAIL**하고(스캐너가 실제로 `.js`를 읽었음의 증거),
**And** 중립화 후 재실행하면 PASS한다.

### 시나리오 4 — 병렬 심사 후 단일 구속 verdict

**Given** run Phase 품질 단계가 병렬 증거 수집 + 단일 verdict 구조로 재구조화되어 있을 때,
**When** 4개 차원 read-only 심사가 병렬 완료되면,
**Then** 최종 PASS/FAIL verdict는 `sync-auditor` 1개 에이전트가 산출하며,
**And** 스크립트가 계산한 집계 점수는 증거로만 인용되고 verdict를 대체하지 않는다.

### 시나리오 5 — 다이어트 후 SSOT 도달성 보존

**Given** `manager-spec.md`에서 12-field frontmatter 스키마 블록이 제거되었을 때,
**When** 독자가 해당 위치를 읽으면,
**Then** `spec-frontmatter-schema.md`를 가리키는 교차참조가 존재하고,
**And** 그 경로의 파일이 실재하며,
**And** 12개 필드 정보가 그 파일에서 조회 가능하다.

### 시나리오 6 — Template-First 위반 탐지

**Given** 미러가 존재하는 파일이 편집되었을 때,
**When** 로컬 파일과 템플릿 미러를 `diff`하면,
**Then** §C Pre-flight에서 확정한 baseline 차이 외의 신규 차이가 0이다.

---

## §C 엣지 케이스

| # | 케이스 | 기대 동작 |
|---|---|---|
| E1 | manifest 교집합 판정이 모호(glob 패턴 중첩) | 직렬화가 기본값 — 판정 불가는 "겹침"으로 취급 |
| E2 | fan-out 중 1개 drafter가 blocker report 반환 | 나머지 결과는 유지, 오케스트레이터가 해당 항목만 재위임 |
| E3 | `moai verify` 스냅샷이 TTL 만료 | 스냅샷 인용 금지, 해당 검사 재실행 |
| E4 | 라인 상한을 충족하나 교차참조 누락 | REQ-APO-068 FAIL — 상한 충족만으로 PASS 불가 |
| E5 | 템플릿 편집 중 SPEC ID 문자열 유입 | 중립성 가드 FAIL, 즉시 제거 |
| E6 | D3 게이트 grep에서 마커 소비자 발견 | REQ-APO-043은 출력 계약 보존 + 주변 산문만 축약으로 전환 |
| E7 | docs-site 주장 정합화 시 4-locale 중 일부 누락 | 4-locale 동시 수정 의무 위반 — FAIL |
| E8 | fan-out 폭이 5를 초과 | Mode 4 상한 위반 — 3-5로 조정 |
| E9 | `.js` 확장자 추가 없이 중립성 green 주장 | 공허 통과 — AC-APO-071 증거 무효, FAIL |
| E10 | 중립화 중 스크립트 실행 로직 변경 | 범위 위반 — 헤더·주석만 대상 |
| E11 | 배포 후 `hns-*` 차단이 함께 완화됨 | dev-only 격리 훼손 — AC-APO-072b FAIL |
| E12 | 3개 스크립트가 user-owned 보존 집합에 편입됨 | `moai update`가 갱신 불가 — AC-APO-073 FAIL |

---

## §D AC 매트릭스

### D.1 Group 1 — write-concurrency (REQ-APO-001..005)

| AC | REQ | 등급 | 판정 |
|---|---|---|---|
| AC-APO-001 | 001 | MUST | `grep -n "overlapping scope" .claude/rules/moai/core/agent-common-protocol.md` ≥ 1건, 그리고 절대 금지 문장("does not run two write-capable agents concurrently" 단독형)이 0건 |
| AC-APO-002 | 002 | MUST | `agent-common-protocol.md`에 disjoint path manifest 계약 서술 존재(`grep -n "disjoint"` ≥ 1) |
| AC-APO-003 | 003 | MUST | 교집합 시 직렬화 규칙 서술 존재, 그리고 "판정 불가 = 직렬화" 기본값 명시 |
| AC-APO-004 | 004 | MUST | `CLAUDE.md`와 `internal/template/templates/CLAUDE.md`의 해당 문장이 동일하고 양쪽 모두 `overlapping scope` 포함 |
| AC-APO-005 | 005 | MUST | `grep -n "orchestrator work concurrent with a write-capable agent"` 문장이 3표면 모두에 보존 |

### D.2 Group 2 — fan-out 배선 (REQ-APO-010..016)

| AC | REQ | 등급 | 판정 |
|---|---|---|---|
| AC-APO-010 | 010 | MUST | `grep -rn "plan-research-fanout" .claude/skills/moai/workflows/` ≥ 1건 |
| AC-APO-011 | 011 | MUST | 3개 참조 각각에 capability-gate 문구 동반 — gate 조건이 **파일 존재 AND 런타임 지원** 두 가지를 모두 명시하고, 참조 건수와 gate 문구 건수 일치 |
| AC-APO-012 | 012 | MUST | `grep -rn "sync-audit-4dim" .claude/skills/moai/workflows/` ≥ 1건 |
| AC-APO-013 | 013 | MUST | 각 스크립트 참조 지점 인근에 verdict 소유권 보존 문장 존재(auditor가 verdict 소유) |
| AC-APO-014 | 014 | MUST | `grep -rn "codemaps-extract" .claude/skills/moai/workflows/` ≥ 1건, 그리고 high-count 스코핑 문구 동반 |
| AC-APO-015 | 015 | MUST | zero-orphan: 3개 스크립트명 각각이 `.claude/skills/moai/workflows/` 하위에서 최소 1건 매치 (3/3) |
| AC-APO-016 | 016 | MUST | docs-site 4-locale `workflows.md`의 파이프라인 투입 주장이 **참임이 검증**됨 — AC-APO-015(zero-orphan) AND AC-APO-069(배포 존재) 동시 PASS가 근거. 배선/배포 미완 시에는 4개 로케일 동시 정정으로 대체 판정 |

### D.3 Group 3 — 재구조화 (REQ-APO-020..030)

| AC | REQ | 등급 | 판정 |
|---|---|---|---|
| AC-APO-020 | 020 | MUST | plan Phase 11 서술에 병렬 read-only 심사 + plan-auditor 단일 verdict 구조 명시 |
| AC-APO-021 | 021 | MUST | plan Phase 10 단일 writer 유지 명시 + 단일 턴 병렬 Write 지시 존재 |
| AC-APO-022 | 022 | SHOULD | RED 단계 drafter pool + 단일 applier 구조 서술 존재 |
| AC-APO-023 | 023 | MUST | run Phase 13/16/17 축약 구조 서술 존재, 그리고 최대 3회 반복 상한 문구 보존(`grep -n "max 3 iterations"` ≥ 1) |
| AC-APO-024 | 024 | MUST | sync Phase 12에 5개 read-only drafter 구조 서술 + 단일 `manager-docs` 순차 적용 명시; disjoint-writer 변형 서술 0건 |
| AC-APO-024b | 024b | MUST | Phase 12 서술이 Group 1 결과와 독립임이 명시 — M1 미완/철회 시에도 M3 진행 가능함이 문서상 확인 |
| AC-APO-025 | 025 | SHOULD | sync Phase 10 패키지별 fan-out 서술 존재 |
| AC-APO-026 | 026 | MUST | sync Phase 1/7에 `moai verify` 스냅샷 소비 서술 + 신선도(키 일치/TTL) 조건 동반 |
| AC-APO-027 | 027 | SHOULD | MX 스캔 샤딩 서술 존재 |
| AC-APO-028 | 028 | MUST | 모든 신규 fan-out 서술이 오케스트레이터 launch로 기술 — subagent nesting 의존 문구 0건 |
| AC-APO-029 | 029 | MUST | 게이트 토큰 4종(`Decision Point 1` / `Implementation Kickoff Approval` / `gate-sync-1` / `gate-sync-2`) 모두 편집 후에도 존재 |
| AC-APO-030 | 030 | MUST | drafter/judge 서브에이전트 서술에 blocker report 반환 규범 존재, `AskUserQuestion` 호출 지시 0건 |

### D.4 Group 4 — 본문 다이어트 (REQ-APO-040..055)

| AC | REQ | 등급 | 판정 |
|---|---|---|---|
| AC-APO-040 | 040 | MUST | `plan-auditor.md` 내 12-field 열거 블록 1개 이하 (현재 2개: MP-3 + FC-1..12) |
| AC-APO-041 | 041 | MUST | `manager-spec.md`에 12-field 열거 블록 0개, `spec-frontmatter-schema.md` 교차참조 ≥ 1건 |
| AC-APO-042 | 042 | MUST | `grep -c "Chain-of-Verification" .claude/agents/moai/plan-auditor.md` == 0 |
| AC-APO-043 | 043 | MUST | **선행 게이트**: `grep -rn "decomposition\|segment match trace" internal/ .github/ .claude/` 출력이 `progress.md`에 verbatim 기록됨. **소비자 0건**이면 마커 강제 제거 + 블록 축약; **소비자 ≥1건**이면 마커의 출력 계약을 보존한 채 주변 산문만 축약. 어느 경우든 실행 Bash 검사는 존치 |
| AC-APO-044 | 044 | MUST | Step 5 체크리스트 항목 중 Step 4 서술의 축자 재진술 0건 |
| AC-APO-045 | 045 | MUST | `manager-spec.md`에 GEARS/EARS 패턴 표 0개, `moai-workflow-spec` 교차참조 ≥ 1건 |
| AC-APO-046 | 046 | MUST | `manager-spec.md` Step 4의 산출물 개수 서술이 실제 열거 개수와 일치 |
| AC-APO-047 | 047 | MUST | `manager-develop.md`에 DDD/TDD 전문 2회 기술 부재 — 공통 골격 + 모드 차이 구조 |
| AC-APO-048 | 048 | MUST | "one atomic change" 제약에 패키지 내부 한정 수식어 존재 |
| AC-APO-049 | 049 | MUST | `grep -c "scoring model" sync-auditor.md` 기준 복수 모델 서술 0건, report template 섹션 1개 |
| AC-APO-050 | 050 | MUST | `grep -ciE "nextra\|wcag\|page.?speed\|lighthouse" manager-docs.md` == 0 |
| AC-APO-051 | 051 | MUST | `e2e-tester.md`에 비-호스트 OS 레시피 본문 부재 + 스킬 레퍼런스 경로 참조 존재(참조 대상 파일 실재) |
| AC-APO-052 | 052 | MUST | `grep -c "merge_method" manager-git.md` == 1 |
| AC-APO-053 | 053 | MUST | `builder-harness.md`의 `Model/effort escalation` 중복 문장 1개 이하, model-policy 표 재진술 0건 |
| AC-APO-054 | 054 | MUST | `grep -l "verification-batch-pattern\|Parallel Execution" .claude/agents/moai/*.md` 결과 파일 수 ≥ 8 (현재 1) |
| AC-APO-055 | 055 | MUST | `wc -l .claude/agents/moai/*.md` 합계 ≤ 1907, 그리고 파일별 상한(`spec.md` §D.2) 전량 충족 |

### D.5 Group 5 — 불변식 (REQ-APO-060..068)

| AC | REQ | 등급 | 판정 |
|---|---|---|---|
| AC-APO-060 | 060 | MUST | 편집된 미러 쌍 전량 `diff` 결과가 Pre-flight baseline 차이 외 0 |
| AC-APO-061 | 061 | MUST | `grep -rnE "SPEC-[A-Z0-9-]+-[0-9]{3}\|REQ-APO-\|20[0-9]{2}-[0-9]{2}-[0-9]{2}" internal/template/templates/<edited files>` == 0 |
| AC-APO-062 | 062 | MUST | `ls internal/template/templates/.claude/workflows/` 결과에 `hns-*` / `harness-*` 접두 파일 0개 (generic fan-out 3개만 존재) |
| AC-APO-063 | 063 | MUST | 시나리오 3 통과 — 스크립트 부재 시 fallback 경로가 문서상 완결 |
| AC-APO-064 | 064 | MUST | `go test ./...` exit 0; template-neutrality CI guard green |
| AC-APO-065 | 065 | MUST | `git diff` 상 `.claude/agents/moai/*.md` frontmatter 블록 변경 0줄 |
| AC-APO-066 | 066 | MUST | archived 12개 에이전트명이 편집 산출물에서 신규 매치 0건 |
| AC-APO-067 | 067 | MUST | `go test ./internal/template/...` exit 0 (`split_namespace_test.go`, `internal_content_leak_test.go` 포함) |
| AC-APO-068 | 068 | MUST | 제거된 각 중복 블록 위치에 SSOT 교차참조 존재, 참조 경로 파일 전량 실재 |

### D.6 Group 6 — 배포 (REQ-APO-069..073)

| AC | REQ | 등급 | 판정 |
|---|---|---|---|
| AC-APO-069 | 069 | MUST | `ls internal/template/templates/.claude/workflows/` 에 3개 파일 존재; `make build` 후 embedded 트리에서도 조회 가능 |
| AC-APO-070 | 070 | MUST | 본 SPEC frontmatter에 `partially_supersedes: [SPEC-DWF-CODEMAPS-PILOT-001]` 존재 **AND** `spec.md`가 supersede 대상 AC를 명시 **AND** `SPEC-DWF-CODEMAPS-PILOT-001` 아티팩트에 supersession 주석이 추가되어 상호 참조 성립 |
| AC-APO-071 | 071 | MUST | 배포된 3개 파일에 대해 `grep -nE "SPEC-[A-Z0-9-]+-[0-9]{3}\|REQ-[A-Z]+-[0-9]\|AC-[A-Z]+-[0-9]\|20[0-9]{2}-[0-9]{2}-[0-9]{2}\|/Users/\|[0-9a-f]{9,40}"` == 0. **선행 조건**: AC-APO-072 PASS (미충족 시 본 AC는 공허하여 무효) |
| AC-APO-072 | 072 | MUST | `leakTextExtensions`에 `".js": true` 존재 (`grep -n '".js"' internal/template/internal_content_leak_test.go` ≥ 1) **AND** 시나리오 8의 RED/GREEN 왕복이 관측됨 — 미중립 스크립트 심었을 때 FAIL, 중립화 후 PASS |
| AC-APO-072b | 062/069 | MUST | `TestSplitHarnessNamespaceNoLeak` PASS **AND** 차단 유효성 확인: `hns-release-update-run.js`를 템플릿에 심고 실행 시 `SPLIT_HARNESS_NAMESPACE_LEAK`으로 FAIL (심은 파일은 제거) |
| AC-APO-073 | 073 | MUST | `internal/cli/update/plan/plan.go`의 user-owned 판정이 3개 generic 스크립트에 대해 false 반환 — 접두사 `hns-`/`harness-` 미매치로 확인. 보존 목록 소스 무변경(`git diff` 0줄) |

---

## §E 품질 게이트

| 게이트 | 기준 |
|---|---|
| 테스트 | `go test ./...` green |
| 커버리지 | 회귀 없음(본 SPEC은 Go 코드 무변경이므로 baseline 유지) |
| 린트 | `golangci-lint run` green |
| SPEC 린트 | `moai spec lint` 본 SPEC 디렉터리 대상 0 errors |
| 템플릿 중립성 | CI guard green **AND** 스캐너가 `.js`를 실제로 읽음이 RED/GREEN 왕복으로 입증(공허 green 금지) |
| 미러 파리티 | 편집 쌍 전량 0-diff |
| 격리 불변식 | `TestSplitHarnessNamespaceNoLeak` green + 차단 유효성 확인 |

---

## §F Definition of Done

- [ ] MUST 등급 AC 전량 PASS, verbatim 명령 출력으로 근거 제시
- [ ] SHOULD 등급 AC 미충족 시 사유가 `progress.md`에 기록
- [ ] §B 시나리오 8개 전량 통과
- [ ] 사용자 결정 D1 / D2 / D3 반영 확인 — D3는 게이트 grep 출력이 `progress.md`에 verbatim 기록
- [ ] `spec.md` §D.2 라인 상한 전량 충족(실측 `wc -l` 인용)
- [ ] Template-First 순서 준수 감사 통과
- [ ] HUMAN GATE 4종 보존 확인
- [ ] verdict 소유권 불변 확인
- [ ] `SPEC-DWF-CODEMAPS-PILOT-001` supersession 상호 참조 성립
- [ ] `progress.md` §E.2 / §E.3 run-phase 증거 기록
