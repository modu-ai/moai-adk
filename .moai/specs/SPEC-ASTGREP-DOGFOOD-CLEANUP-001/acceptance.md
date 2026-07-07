---
id: SPEC-ASTGREP-DOGFOOD-CLEANUP-001
title: "Local Dogfood ast-grep Ruleset Curated-Baseline Cleanup — Acceptance"
version: "0.1.0"
status: draft
created: 2026-07-08
updated: 2026-07-08
author: GOOS
priority: P3
phase: "maintainer-tooling"
module: ".moai/config/astgrep-rules"
lifecycle: spec-anchored
tags: "astgrep, dogfood, cleanup, tooling, local"
---

# 인수 기준 — SPEC-ASTGREP-DOGFOOD-CLEANUP-001

## §D AC 매트릭스 (각 항목 기계 검증 가능)

모든 명령은 저장소 루트(`/Users/goos/MoAI/moai-adk-go`)에서 실행. `run-phase` 완료 후 검증.

| AC | REQ | 검증 명령 | 기대 결과 |
|----|-----|-----------|-----------|
| AC-ADC-001 | REQ-ADC-001 | `ls .moai/config/astgrep-rules/go-hardcoding.yml 2>&1` | `No such file or directory` (orphan 부재) |
| AC-ADC-002 | REQ-ADC-002 | `grep -cE '^\s*-\s+(go\|security)\s*$' .moai/config/astgrep-rules/sgconfig.yml` | `2` (go + security만) |
| AC-ADC-003a | REQ-ADC-003 | `grep -c 'utils' .moai/config/astgrep-rules/sgconfig.yml` | `0` (utils ruleDir 부재) |
| AC-ADC-003b | REQ-ADC-002 | `grep -cE '^\s*-\s+' .moai/config/astgrep-rules/sgconfig.yml` | `2` (ruleDir 항목 총 2개) |
| AC-ADC-004 | REQ-ADC-004 | `for d in cpp flutter java javascript python r rust scala swift; do ls -d .moai/config/astgrep-rules/$d 2>/dev/null; done \| wc -l \| tr -d ' '` | `0` (빈 stub 9개 전부 부재) |
| AC-ADC-005 | REQ-ADC-005 | `for d in csharp elixir kotlin php ruby; do ls -d .moai/config/astgrep-rules/$d 2>/dev/null; done \| wc -l \| tr -d ' '` | `0` (데모 stub 5개 전부 부재) |
| AC-ADC-006 | REQ-ADC-006 | `grep -c 'SPEC-' .moai/config/astgrep-rules/sgconfig.yml` | `0` (내부 SPEC-ID 부재) |
| AC-ADC-007 | REQ-ADC-007 | `sg scan --config .moai/config/astgrep-rules/sgconfig.yml . >/dev/null 2>sg.err; grep -ic 'ruleDir\|no such file\|not found\|utils' sg.err` | `0` (누락 ruleDir 오류 부재; sg exit 성공) |
| AC-ADC-008 | REQ-ADC-008 | `git status --porcelain internal/template/templates/.moai/config/astgrep-rules/ \| wc -l \| tr -d ' '` | `0` (템플릿 트리 무변경) |
| AC-ADC-009a | REQ-ADC-009 | `find .moai/config/astgrep-rules/go -type f \| wc -l \| tr -d ' '` | `5` (go 실제 룰 5개 보존) |
| AC-ADC-009b | REQ-ADC-009 | `find .moai/config/astgrep-rules/security -type f \| wc -l \| tr -d ' '` | `5` (security 실제 룰 5개 보존, untracked credentials.yml 포함) |
| AC-ADC-010 | REQ-ADC-010 | `git diff --name-only internal/template/templates/ \| wc -l \| tr -d ' '` | `0` (template-neutrality CI 미트리거 조건) |

> AC-ADC-007 주: sg는 read-only scan이므로 대상(`.`)에 findings가 있어도 무방하다. 검증 대상은
> "누락 ruleDir로 인한 **설정 파싱 오류** 부재"이며 findings 개수가 아니다. run-phase에서 실제
> sg 출력을 관측해 오류 메시지 패턴이 없음을 확인한다 (verification-claim-integrity: 관측된 증거).

## §D.1 Given-When-Then 시나리오 (최소 2)

### 시나리오 1 — sgconfig 정렬 후 config-mode 파싱 (정상 경로)

- **Given** 로컬 `sgconfig.yml`의 `ruleDirs`가 `[go, security]`로 정렬되고 존재하지 않는 `utils`
  항목이 제거된 상태에서,
- **When** maintainer가 `sg scan --config .moai/config/astgrep-rules/sgconfig.yml .`을 실행하면,
- **Then** ast-grep 엔진은 누락 ruleDir 오류(`utils` 디렉터리 없음) 없이 설정을 파싱하고 정상
  종료한다 (AC-ADC-007).

### 시나리오 2 — 병렬 세션 레이스 방어 (go/security 보존)

- **Given** `go/`·`security/`가 오늘 수정되었고 `security/credentials.yml`이 untracked(in-flight)인 상태에서,
- **When** 정리 run-phase가 orphan + 14개 stub/demo 디렉터리를 경로-한정 `git rm`으로 삭제하면,
- **Then** `go/` 5파일과 `security/` 5파일(untracked credentials.yml 포함)은 변경 없이 보존되고,
  삭제는 명시된 15개 경로에만 적용된다 (AC-ADC-009a/009b).

### 시나리오 3 — 템플릿 트리 무접촉 (경계 가드)

- **Given** 배포 대상 템플릿 트리 `internal/template/templates/.moai/config/astgrep-rules/`가 완료된
  `SPEC-ASTGREP-MULTILANG-001`에 의해 확정된 상태에서,
- **When** 로컬 dogfood 정리가 완료되면,
- **Then** `git status --porcelain internal/template/templates/.moai/config/astgrep-rules/`는 빈 출력이며,
  template-neutrality CI는 트리거되지 않는다 (AC-ADC-008, AC-ADC-010).

## §D.2 엣지 케이스

- **EC-1**: Pre-flight에서 `go/`·`security/`가 dirty이거나 origin이 앞서면 (병렬 세션 활동 감지) →
  run-phase는 STOP하고 blocker report를 오케스트레이터에 반환 (삭제 미수행).
- **EC-2**: `sg`가 정리 시점에 미설치(경로 변동)면 → AC-ADC-007을 수동 검증 단계로 강등하고
  progress.md에 관측 불가 사유 기록 (verification-claim-integrity: 미관측을 Gap으로 명시).
- **EC-3**: 데모 stub 디렉터리에 `.gitkeep` 외 untracked 신규 파일이 발견되면 → 해당 파일은
  in-flight 가능성이 있으므로 삭제 보류하고 사용자에게 surface.

## §D.3 완료 정의 (Definition of Done)

- [ ] AC-ADC-001 ~ AC-ADC-010 전부 PASS (실제 명령 출력 관측 증거 첨부)
- [ ] 시나리오 1~3 통과
- [ ] `git status`에서 삭제 파일이 정확히 15개 경로(orphan 1 + stub/demo 14 디렉터리)로 한정됨
- [ ] 템플릿 트리 diff 0 확인
- [ ] progress.md §E.2/§E.3 run-phase 증거 기록 (manager-develop 소관)
- [ ] 커밋은 오케스트레이터가 처리 (경로-한정, `git add -A` 금지)

## §D.4 품질 게이트

- Go 코드 무변경이므로 `go test`/`golangci-lint` 회귀 없음 (파일/설정만 변경).
- 런타임 동작 변경 없음 (gate OFF) — 회귀 위험 최소.
- 유일한 기능 검증은 `sg` config-mode 파싱 (AC-ADC-007).
