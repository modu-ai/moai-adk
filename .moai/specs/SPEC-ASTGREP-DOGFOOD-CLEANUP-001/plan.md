---
id: SPEC-ASTGREP-DOGFOOD-CLEANUP-001
title: "Local Dogfood ast-grep Ruleset Curated-Baseline Cleanup — Plan"
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

# 구현 계획 — SPEC-ASTGREP-DOGFOOD-CLEANUP-001

## §A 맥락 (Context)

Tier **M (standard)**. 로컬 dogfood ast-grep 트리를 완료된 템플릿 `[go, security]` curated
baseline으로 정렬하는 단일 관심사 위생 작업. 런타임 동작 변경 없음(gate OFF). 로컬 전용,
git-tracked, 완전 되돌림 가능.

### Tier 판정 근거 (S/M/L)

- **Tier L 아님**: 새로운 16개 언어 룰 저작이 아니라 curated 정렬 + 삭제만. 아키텍처 변경 없음.
- **Tier S 아님**: 단순 1파일 typo 수정을 넘어선다 — ~30개 tracked 파일 삭제 + sgconfig 편집 +
  `sg` config-mode 검증 게이트 + 병렬 세션 레이스 완화 + 스코프 결정을 포함한다.
- **Tier M 확정**: 다중 파일 조율 정리 + 실제 검증 게이트(`sg scan`) + 2개 원천 교차참조. 표준 harness.

## §B 알려진 이슈 / 제약 (Known Issues & Constraints)

- **[CRITICAL] 병렬 세션 레이스**: `go/`·`security/`가 오늘(2026-07-08 02:29) 수정됨 +
  `security/credentials.yml` untracked(in-flight). 다른 세션이 astgrep 보안 룰을 동시 작업 중일
  가능성. run-phase는 spawn 전 반드시 Pre-Spawn Sync Check + 로컬 dirty 상태 확인 필수. go/security
  **내용은 무접촉** (REQ-ADC-009).
- **`sg` 실행 금지 (plan-phase)**: 본 plan-phase에서는 sg를 실행하지 않는다. REQ-ADC-007의
  "sg가 누락 ruleDir 오류를 낸다"는 아직 **가설**이며 run-phase에서 기계적으로 검증한다
  (verification-claim-integrity: 도구로 관측하기 전엔 defect 단언 금지).
- **`sg` 설치 확인됨**: ast-grep 0.40.5 (`/opt/homebrew/bin/sg`). 따라서 REQ-ADC-007은 run-phase에서
  실제 실행 가능한 기계 검증이다 (수동 fallback 불필요).
- **로컬 전용**: `internal/template/templates/` 밖이므로 template-neutrality CI 미트리거 (REQ-ADC-010).

## §C 사전 점검 (Pre-flight — run-phase 착수 전)

1. `git fetch origin main` + `git rev-list --count --left-right origin/main...HEAD` (Pre-Spawn Sync Check).
2. `git status --porcelain .moai/config/astgrep-rules/` — go/security 및 untracked 파일 in-flight 여부 확인.
3. go/security가 dirty이거나 origin이 앞서면 STOP → 사용자에게 surface (병렬 레이스).
4. 삭제 대상 목록을 삭제 전 실측 재확인 (`git ls-files`로 tracked 여부 확정).

## §D 기술 접근 (Technical Approach)

### D.1 삭제 (git rm — tracked 파일)

- orphan root: `.moai/config/astgrep-rules/go-hardcoding.yml` (REQ-ADC-001)
- 빈 stub 디렉터리 9개 (`.gitkeep`): cpp, flutter, java, javascript, python, r, rust, scala, swift (REQ-ADC-004)
- 데모 stub 디렉터리 5개: csharp, elixir, kotlin, php, ruby (각 `.gitkeep` + 3 boilerplate .yml) (REQ-ADC-005)

> 삭제는 `git rm`(tracked) 사용. `git add -A` 금지 — untracked `credentials.yml` in-flight
> 파일이 실수로 staged/삭제되지 않도록 **경로 한정 삭제**만 수행. (병렬 레이스 방어 lesson 적용.)

### D.2 편집 (Edit — sgconfig.yml)

- `ruleDirs`를 `[go, security]`로 축소 (템플릿 미러) → REQ-ADC-002, REQ-ADC-003 (utils 제거)
- 주석의 `SPEC-ASTG-UPGRADE-001` 내부 ID 제거 → REQ-ADC-006
- (선택) 주석 문안을 템플릿 sgconfig.yml의 curated 설명 톤으로 정렬 — 로컬이므로 필수 아님

### D.3 무접촉 (보존)

- `go/` 5파일, `security/` 5파일 내용 → REQ-ADC-009 (무접촉)
- untracked `security/credentials.yml` → REQ-ADC-009 (무접촉)
- 템플릿 트리 전체 → REQ-ADC-008 (무접촉)

### D.4 검증 (sg config-mode — run-phase)

- `sg scan --config .moai/config/astgrep-rules/sgconfig.yml <target>` 실행 → 누락 ruleDir 오류 부재 확인 (REQ-ADC-007)
- (참고) sg는 read-only scan; 파일 수정 없음.

## §E 자기 검증 (Self-Verification — plan-phase audit-ready signal)

plan-phase 산출물 완결성 체크:
- [ ] 12-필드 canonical frontmatter (4개 산출물 전부 status: draft)
- [ ] GEARS 요구사항 10개 (REQ-ADC-001..010), 각 기계 검증 가능
- [ ] Out of Scope 섹션 4개 `### Out of Scope —` H3 + `-` bullet
- [ ] acceptance.md에 mechanically-verifiable AC + 최소 2 Given-When-Then
- [ ] progress.md §E.1~§E.4 skeleton (§E.2~§E.4 placeholder)
- [ ] 실측 앵커 재검증 완료 (2026-07-08, manager-spec)

## §F 마일스톤 (Milestones — 우선순위 기반, 시간 추정 없음)

- **M1 (Priority High)**: Pre-flight 레이스 점검 + 삭제 대상 실측 재확인. go/security dirty면 STOP.
- **M2 (Priority High)**: orphan + 14개 stub/demo 디렉터리 경로-한정 `git rm`. (REQ-ADC-001/004/005)
- **M3 (Priority High)**: sgconfig.yml 편집 — ruleDirs `[go, security]` + SPEC-ID strip. (REQ-ADC-002/003/006)
- **M4 (Priority Medium)**: `sg scan` config-mode 검증 — 누락 ruleDir 오류 부재. (REQ-ADC-007)
- **M5 (Priority Medium)**: 무접촉 검증 — go/security 내용 + untracked + 템플릿 diff 0. (REQ-ADC-008/009/010)

## §G 안티패턴 (Anti-Patterns — 하지 말 것)

- `git add -A` / `git add .` — untracked in-flight `credentials.yml`을 삼킴. 경로-한정만.
- `rm -rf`로 go/ 또는 security/ 전체 삭제 — 실제 vetted 룰 유실. 삭제는 명시 경로 14개 + orphan 1개만.
- 템플릿 트리 수정 — REQ-ADC-008 위반.
- go/security 메시지 영어화 착수 — 본 SPEC out-of-scope + 레이스 위험.
- plan-phase에서 sg 실행 — 사용자 제약 위반.

## §H 교차 참조 (Cross-References)

- spec.md §A.1 (템플릿 baseline SSOT), §C (Out of Scope).
- `CLAUDE.local.md §2.2` (원천), §15/§25 (격리 doctrine).
- `.claude/rules/moai/core/agent-common-protocol.md` § Pre-Spawn Sync Check (병렬 레이스 완화).
