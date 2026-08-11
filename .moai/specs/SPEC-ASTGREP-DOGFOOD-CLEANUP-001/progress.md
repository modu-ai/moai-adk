---
id: SPEC-ASTGREP-DOGFOOD-CLEANUP-001
title: "Local Dogfood ast-grep Ruleset Curated-Baseline Cleanup — Progress"
version: "0.2.0"
status: in-progress
created: 2026-07-08
updated: 2026-08-12
author: GOOS
priority: P3
phase: "maintainer-tooling"
module: ".moai/config/astgrep-rules"
lifecycle: spec-anchored
tags: "astgrep, dogfood, cleanup, tooling, local"
---

# 진행 기록 — SPEC-ASTGREP-DOGFOOD-CLEANUP-001

## §E.1 Plan-phase Audit-Ready Signal

- Plan-phase 산출물 4종 (spec.md + plan.md + acceptance.md + progress.md), 전부 `status: draft`.
- **v0.2.0 refresh (2026-08-12, MoAI-Easy 직접 편집, worktree `adc-plan-refresh`)**: REQ-ADC-002/003은
  `SPEC-CONFIG-AUDIT-REPAIR-001` M2 (#1142, commit 6e70a29fd)가 해소 → 산출물에서 제거 (번호 공간 보존).
  잔여 활성 REQ 8개(001/004/005/006/007/008/009/010). 정리 대상 4개(001/004/005/006)는 순수 위생.
  REQ-004 빈 stub 개수 9→10 정정(typescript 누락). REQ-006 타깃 `SPEC-ASTG-UPGRADE-001`→현재 잔존
  `SPEC-GATE-ASTGREP-REPAIR-001`(sgconfig.yml:10)로 갱신. §A.3 gate 프레이밍 #1142 이후 실제(warn-only ON)로 정정.
- 실측 앵커 (2026-08-12 재검증): ruleDirs 이미 `[go, security]` 정렬(#1142); utils 부재; orphan
  `go-hardcoding.yml` 잔존(Go/config 참조 0 = 진짜 orphan); 빈 stub 10(cpp/flutter/java/javascript/
  python/r/rust/scala/swift/typescript); 데모 stub 5(csharp/elixir/kotlin/php/ruby); sgconfig.yml:10에
  `SPEC-GATE-ASTGREP-REPAIR-001` 잔존 → AC-006 현재 FAIL(이것이 run-phase REQ-006으로 해소 대상).
- Tier: **M (standard)** (유지). 스코프 결정: **curated-baseline alignment** (16-lang 전면 저작 아님).
- GEARS 요구사항: 활성 8개 + 해소 2개(#1142). 활성 AC 9개(AC-001/004/005/006/007/008/009a/009b/010),
  해소 AC 3개(AC-002/003a/003b). Given-When-Then 3.
- Out of Scope 4개 H3 sub-heading + `-` bullet 확보 (OutOfScopeRule 충족).
- 커밋 미수행 (오케스트레이터 소관).

## §E.2 Run-phase Evidence

**실행 (2026-08-12, MoAI-Easy 직접 구동, worktree `worktree-adc-plan-refresh` @ ed70e4354):**

- **M1 pre-flight (read-only)**: divergence `origin/main...HEAD` = `0 0` (병렬 세션 레이스 없음);
  `git status --porcelain` on go/·security/ = empty (dirty 없음); `credentials.yml` = **tracked**
  (2026-07-08 untracked 상태에서 이후 커밋됨 — in-flight 우려 해소). 삭제 대상 31 tracked 파일 확정.
- **M2 경로-한정 `git rm -r`**: orphan `go-hardcoding.yml` + 빈 stub 10 디렉터리 + 데모 stub 5 디렉터리
  = 31 tracked 파일 스테이징 삭제 (`git add -A` 미사용 — 경로 한정 원칙 준수).
- **M3 sgconfig.yml 편집**: `SPEC-GATE-ASTGREP-REPAIR-001` SPEC-ID 접두 strip + 삭제된
  `go-hardcoding.yml` 참조 주석(구 line 7-8) 제거. ast-grep config-mode globs 기술 노트 본문은 보존.
- **M4/M5 검증**: 아래 AC 매트릭스.

**AC 매트릭스 (실측 command + 관측 출력, baseline = 본 run @ ed70e4354 worktree):**

| AC | command | observed output | PASS? |
|----|---------|-----------------|-------|
| AC-001 | `ls .moai/config/astgrep-rules/go-hardcoding.yml` | `No such file or directory` | ✅ |
| AC-004 | `ls -F .moai/config/astgrep-rules/` | `go/ security/ sgconfig.yml` (10 빈 stub dir 전부 부재) | ✅ |
| AC-005 | (상동 tree listing) | csharp/elixir/kotlin/php/ruby 전부 부재 | ✅ |
| AC-006 | `grep -c 'SPEC-' sgconfig.yml` | `0` | ✅ |
| AC-007 | `sg scan --config .../sgconfig.yml <self>` | exit 0, ruleDir 오류 없음 | ✅ |
| AC-008 | `git status --porcelain internal/template/templates/.moai/config/astgrep-rules/` | empty | ✅ |
| AC-009a | `ls go/` | 5 파일 (concurrency/error-handling/hardcoding/idioms/resource-safety) | ✅ |
| AC-009b | `ls security/` | 5 파일 (credentials/crypto/injection/secrets/web) | ✅ |
| AC-010 | `git status --porcelain internal/template/templates/` | empty | ✅ |

**Gaps**: 없음 — 모든 활성 AC 관측됨. (AC-002/003a/003b는 #1142 해소로 본 SPEC에서 제거됨.)
**Residual-risk**: sg scan 대상을 self(sgconfig.yml)로 한정해 파싱 검증. 전체 리포(`.`) 스캔은
findings가 많아 stdout을 suppress했으나, config PARSE 오류(ruleDir)는 target 무관 config-load 시점에
발생하므로 self-target으로 동등 검증됨.

## §E.3 Run-phase Audit-Ready Signal

- **Run-phase 완료**: M1~M5 전부 완료, 활성 AC 9개(AC-001/004/005/006/007/008/009a/009b/010) 전부 PASS.
- **TRUST 5**: Go 코드 무변경 → go test/golangci-lint 회귀 없음. 유일 기능 검증 = sg config-mode 파싱(AC-007 PASS).
- **MX**: 본 변경은 파일/설정 위생으로 @MX 태깅 대상 아님 (Go 코드 0).
- **미커밋 상태**: working tree에 (plan-refresh 4 SPEC md) + (run-phase 31 삭제 + sgconfig edit + progress 증거) 가 unstaged/uncommitted로 대기. 커밋은 오케스트레이터/사용자 소관 (Route B PR — repo-local-pr-policy §23).

## §E.4 Sync-phase Audit-Ready Signal

_<pending sync-phase — manager-docs 소관>_
