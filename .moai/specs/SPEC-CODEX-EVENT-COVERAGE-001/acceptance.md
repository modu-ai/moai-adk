---
id: SPEC-CODEX-EVENT-COVERAGE-001
title: "acceptance — Codex hook 이벤트 커버리지"
version: "0.1.0"
created: 2026-09-07
updated: 2026-09-07
author: manager-spec
---

# acceptance — SPEC-CODEX-EVENT-COVERAGE-001

## D. AC 매트릭스

### M1 (Axis A)

- **AC-CEV-001** (maps REQ-CEV-001) — **Given** M1 구현 후 트리의 `internal/codexadapter/events.go` **When** `awk '/^var EventTable/,/^}/' internal/codexadapter/events.go | grep -c 'true},'` 와 `'false},'` 를 실행 **Then** 합계 12이고 `false}=6` (기존 5 + Interrupt 1), `true}=6`.
- **AC-CEV-002** (maps REQ-CEV-002, REQ-CEV-005) — **Given** M1 구현 후 트리 **When** `grep -rn 'EventInterrupt' internal/ | grep -v _test | grep -v codexadapter` 를 실행 **Then** 매치 0행 (internal/hook 무변경 회귀 판별식, REQ-CEV-002/005).
- **AC-CEV-003** (maps REQ-CEV-003) — **Given** `Resolve("Interrupt")` **When** `go test ./internal/codexadapter/ -run 'TestResolve' -v` 실행 **Then** Interrupt 케이스가 `ErrUnadapted` 래핑과 "메시지가 `dispatcher arg` 존재를 단언하지 않음"을 모두 단언하는 테스트가 GREEN (REQ-CEV-003).
- **AC-CEV-004** (maps REQ-CEV-001, REQ-CEV-003) — **Given** M1 구현 후 **When** `go test ./internal/codexadapter/ ./internal/codexwiring/` 실행 **Then** `TestDispatcherArgsExist`가 갱신돼 GREEN이고, `TestEventTableRowCount`는 12를 단언 (REQ-CEV-001, §D1).
- **AC-CEV-005** (maps REQ-CEV-004) — **Given** M1 구현 전후의 RenderHooks **When** 동일 입력에 대한 `RenderHooks` 출력의 Interrupt 키 존재를 단언하는 테스트(신설 1개: Interrupt 행이 설치 바이트에 나타나지 않음)를 실행 **Then** GREEN — `.codex/hooks.json` 설치 표면 불변 (REQ-CEV-004).
- **AC-CEV-006** (maps REQ-CEV-006) — **Given** M1 이후 `internal/codexadapter/events.go` **When** 아래 3개 명령을 실행 **Then** (1) `grep -c 'All eleven' internal/codexadapter/events.go` → 0, (2) `grep -c 'never an absence of' internal/codexadapter/events.go` → 0 (M1 후 Interrupt는 정확히 부재로 제외되므로 events.go:41-44의 그 문장은 거짓이 된다), (3) 갱신된 doc comment 블록이 `Interrupt`를 언급하며 그 `no MoAI dispatcher counterpart` 성격을 기재한다. 단, "Six rows are adapted" 문구는 M1 후에도 참(6 adapted 불변)이므로 그 삭제를 요구하지 않는다 — 대신 adapted/held-back census가 Interrupt의 사유를 포함해 12행 전체를 셈한다 (REQ-CEV-006).

### M2 (Axis B — 캠페인 기록 기반, 런 페이즈)

- **AC-CEV-010** (maps REQ-CEV-007, REQ-CEV-009) — **Given** `.moai/reports/t496/codex-event-campaign.md` **When** 6종 이벤트(PreCompact, PostCompact, PermissionRequest, SubagentStart, SubagentStop, Interrupt) 각각에 대해 판정 행 존재를 점검 **Then** 6행 모두 fired/not-fired 판정 + 실행 명령 + 관측 출력을 운반한다 (REQ-CEV-007/009). 판정 근거가 문서 인용만인 행은 FAIL.
- **AC-CEV-011** (maps REQ-CEV-008) — **Given** 캠페인 기록 **When** P0 절차 기록을 점검 **Then** `CODEX_HOME` 지원 검증 결과(지원/미지원과 그 근거 명령)와, 미지원 시의 폴백 경로 기록이 존재하며, 연산자 `~/.codex` 변경을 나타내는 기록이 0건이다 (REQ-CEV-008).
- **AC-CEV-012** (maps REQ-CEV-011) — **Given** 캠페인 기록 **When** SubagentStop 행을 점검 **Then** 0.147.0 종전 관측을 인용만 한 게 아니라 0.153.4 재측정 명령+출력을 운반한다 (REQ-CEV-011).
- **AC-CEV-013** (maps REQ-CEV-010) — **Given** 캠페인 기록 **When** 6종 각각의 처분 필드를 점검 **Then** fires 판정 행은 어댑터 확장 결정(페이로드 형상+사상 결정)을, not-fired 행은 미발화 문서화를 운반하며, 미처분 행이 0건이다 (REQ-CEV-010).

### M3 (조건부)

- **AC-CEV-020** (maps REQ-CEV-012) — **Given** M2 처분 **When** adapt-now 표시 행을 셈 **Then** (a) 0건이면 M3 산출물 변경 0이고 M2 기록이 종결 산출물로 기재돼 있거나, (b) 1건 이상이면 해당 이벤트만 테이블 `Adapted` true + 사상 갱신 + `go test ./internal/codexadapter/` GREEN이 성립한다.

## D.1 엣지 케이스

- 사용자가 `.codex/hooks.json`에 손으로 Interrupt 훅을 넣은 경우 → 런타임 Resolve가 ErrUnadapted로 거부(기존 5종과 동일 클래스). harness 거부 경로 테스트로 커버.
- 빈 `DispatcherArg` 행이 `Resolve` adapted 경로에 도달하는 경우 → 구조적으로 불가(Adapted=false 고정). 테스트로 고정 단언.
- 캠페인 중 codex-cli 버전이 0.153.4가 아닌 경우 → 캠페인 기록에 실측 버전을 명시하고, 버전 불일치 시 재측정 전제로 기재.

## D.2 품질 게이트

- 영향 패키지 테스트: `go test ./internal/codexadapter/ ./internal/codexwiring/` GREEN (+ 영향 시 `./internal/cli/ -run 'Codex|Hooks'`)
- `golangci-lint run ./internal/codexadapter/...` 클린
- `go vet` 클린 (영향 패키지)

## D.3 Definition of Done

1. AC-CEV-001~006 전부 GREEN (M1) — 명령+출력이 progress.md §E.2에 귀속
2. AC-CEV-010~013 전부 충족 (M2) — 캠페인 기록 경로가 progress.md에 기재
3. AC-CEV-020 충족 (M3 조건 판정 기록 포함)
4. `internal/hook` diff 0행 (REQ-CEV-005 — `git diff --stat $(git merge-base HEAD origin/develop)..HEAD -- internal/hook/` 공허. HEAD 대상 diff로는 이 브랜치에 이미 커밋된 변경을 못 잡으므로 merge-base 비교가 판별식이다)
5. TRUST 5 — Tested(영향 패키지), Readable/Unified(주석 en·gofmt), Secured(연산자 설정 무변경), Trackable(Conventional Commit)
