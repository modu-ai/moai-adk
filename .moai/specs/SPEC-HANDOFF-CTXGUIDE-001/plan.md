---
id: SPEC-HANDOFF-CTXGUIDE-001
title: "256K 윈도우 핸드오프 안내 임계 결함 수정 — 구현 계획"
version: "0.1.0"
status: in-progress
created: 2026-07-03
updated: 2026-07-03
author: MoAI
priority: P1
phase: "v3.0.0"
module: "internal/statusline"
lifecycle: spec-anchored
tags: "statusline, handoff-guide, context-window, threshold, tier-s, epic-handoff-v2"
era: V3R6
---

# 구현 계획 — SPEC-HANDOFF-CTXGUIDE-001 (Tier S)

## §A.1 Tier 판정

Tier **S**. 근거: Go 단일 함수 로직 교체 + 문서 표 1행 추가 + 단위 테스트 케이스 추가. 신규 의존성 0, 신규 config 0, 신규 상태 파일 0. 영향 파일 3개(<5), 변경 LOC <100(<300). plan-auditor PASS 임계 0.75.

## §A.2 변경 파일 (정확한 타깃)

1. `internal/statusline/renderer.go` — `shouldShowHandoffGuide`(약 571–589행) switch를 크기 무관 밴드 로직으로 교체. 함수 상단 `@MX:NOTE`의 "1M=50%/200K=90%" 문구를 밴드 표현(`≥500K→50% / <500K→90%`)으로 갱신.
2. `internal/statusline/stdinfields_test.go` — 기존 `TestShouldShowHandoffGuide_*` 개별 함수 6개(L31–116)가 위치한 파일. AC-256K-001..006 케이스를 **같은 파일**에 추가(256K show/hide, 500K/499999 경계, cwSize=0). 기존 개별 함수 옆에 추가하거나 테이블 주도로 전환 — `renderer_test.go`에 분산 금지(plan-auditor D1).
3. `.claude/rules/moai/workflow/context-window-management.md` — § Context Window Targets 표에 256K 행 추가(윈도우 256,000 / 임계 90% / 절대 여유 ≈ 230K). session-handoff.md Trigger #1은 이 표를 SSOT로 참조하므로 수치 편집 불필요.

## §A.3 구현 순서 (RED → GREEN, cycle_type=tdd)

1. **RED**: stdinfields_test.go에 AC-256K-001/002/005/006 케이스 추가 → 256K/500K-band 실패 확인(현재 default:false).
2. **GREEN**: renderer.go switch → 밴드 로직 교체. 테스트 통과 확인.
3. **REFACTOR**: `@MX:NOTE` 문구 갱신. 기존 1M/200K 케이스(AC-003/004) 무손상 확인.
4. 문서 표 256K 행 추가(AC-007).
5. `go test ./internal/statusline/` 전체 + 크로스플랫폼 build(AC-008/009).

## §A.4 PRESERVE (변경 금지)

- `renderBarsInline`의 `(⚠️/clear)` 렌더 로직(단일 단계 유지 — 2단계는 M4).
- `internal/statusline/memory.go`의 `ContextWindowSize`/`TokenBudget` 유도 로직(입력 계약 불변).
- 무관 untracked/modified 파일(README.ko.md, docs-site/*, 병렬 세션 산출물).
- runtime-managed: `.moai/state/*`, `.moai/cache/*`, `.moai/harness/*`.
- **미커밋 ♻️ 변경**(renderer.go:311 renderCacheHit, cache_hit_test.go) — 별개 작업이므로 이 SPEC 커밋에 섞지 말 것. renderer.go 편집 시 해당 라인 무접촉.

## §A.5 위험 · 완화

- **위험 R1**: renderer.go에 이미 미커밋 ♻️ 변경(311행)이 있어 같은 파일을 편집. → 완화: `shouldShowHandoffGuide`(약 571–589행)만 편집, 311행 무접촉. 커밋 시 함수 단위 확인으로 ♻️ 변경과 분리(또는 사용자 결정에 따라 별도 커밋).
- **위험 R2**: 500K 컷오프가 미래 300K/400K 클래스 등장 시 재검토 필요. → 완화: 90% 밴드가 보수적 기본값이라 안전; 정밀 튜닝은 M4 config로.
- **위험 R3**: 문서 표 행 추가 시 세션-핸드오프 SSOT 참조 정합. → 완화: session-handoff.md는 표를 참조만 하므로 수치 중복 없음(편집 불필요) 확인.

## §A.6 자가 검증 (완료 시)

- AC-256K-001..009 PASS/FAIL 행렬 + 실제 명령 출력.
- `go test ./internal/statusline/ -run TestShouldShowHandoffGuide -count=1` 출력.
- `go build ./...` + `GOOS=windows GOARCH=amd64 go build ./...` exit 0.
- `grep -n '256' .claude/rules/moai/workflow/context-window-management.md` 표 행 확인.

### §A.7 Out of Scope

이 절은 spec-lint Exclusions 요구를 만족한다. 상세는 spec.md §1.3 참조.

- 2단계 안내(🛑/clear!, 95%) — M4 소관.
- config 오버라이드 / `handoff.yaml` / `HandoffConfig` — M3 landing · M4 소비.
- `.moai/state/context-usage.json` 영속화 — M4 소관.
- Detection Heuristics "state file first" 재작성 — M4 소관.
