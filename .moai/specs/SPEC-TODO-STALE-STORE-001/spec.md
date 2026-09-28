---
id: SPEC-TODO-STALE-STORE-001
title: "트리 안 유령 큐 저장소 — 스테일 프로젝트-로컬 스토어 고지, doctor 발산 점검, 잔존 저장소 처분"
version: "0.1.1"
status: draft
created: 2026-09-29
updated: 2026-09-29
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "internal/kanban + internal/cli"
lifecycle: spec-anchored
tags: "todo-queue, stale-store, doctor, binary-lag, disclosure, t1307"
tier: M
related_specs: [SPEC-TODO-QUEUE-HOME-CANON-001, SPEC-TODO-SQLITE-001, SPEC-BACKLOG-JSON-DISCLOSURE-001, SPEC-WEB-TODO-QUEUE-001]
---

# SPEC-TODO-STALE-STORE-001 — 트리 안 유령 큐 저장소 (카드 t1307)

## HISTORY

- 2026-09-29 v0.1.0 — manager-spec 최초 작성 (카드 t1307, plan-phase). 측정 근거: `.moai/reports/todo-logic-review-20260929.md` P1(2026-09-29 오독 사고 포함), 운영자 제공 실측(홈 DB `~/.moai/db/moai-adk-go-1bd3d038/todo/backlog.db` last_seq 1305 vs 유령 스토어 `.moai/state/todo/backlog.db` seq 661).
- 2026-09-29 v0.1.1 — manager-spec plan-audit iter1 수리 (카드 t1307). D1: M3 삭제 승인 주체 확정(운영자·리드 양쪽 확인 — 카드 t1307 본문 근거, REQ-TSS-021 정합 수정). D3/D4: `-run` 패턴 비공허화(§C.3). D2/D5/D6는 plan.md/acceptance.md 쪽 수리.

## A. 배경과 문제 정의

홈 DB 마이그레이션(SPEC-TODO-QUEUE-HOME-CANON-001, 카드 t470/t472 계열) 이후 정식 큐는
`~/.moai/db/<project-key>/todo/backlog.db` 하나다. 그러나 primary 체크아웃의
`.moai/state/todo/backlog.db`가 **롤백 스냅샷으로 남아** 있고, 이 스토어는

- 어떤 런타임 고지도 없다 — 기존 stderr 고지(`internal/cli/todo_disclosure.go`,
  SPEC-BACKLOG-JSON-DISCLOSURE-001)는 State D(`backlog.json` 이 옆에 있는 형태)만 다룬다.
  스테일 SQLite 스토어는 고지도 고립 마커도 없다.
- 발산 검출이 없다 — 2026-09-29 실제 오독 사고: 리드가 유령 스토어의 `queued=72`를
  살아 있는 큐로 읽었다(홈 DB last_seq 1305 vs 유령 661, 644 seq 격차).
- 잔존 저장소가 처리되지 않았다 — 0바이트 쌍 `.moai/backlog.db`, `.moai/state/backlog.db`
  (카드 t472가 2026-09-03부터 플래그)과 마이그레이션 백업
  `.moai/state/todo-merge-backup-20260913/store-{1,2}/`가 처분 근거 없이 남아 있다.

## B. 요구사항 (GEARS)

### B.1 항목 ① — 스테일 로컬 스토어 고지

- **REQ-TSS-001**: **When** `moai todo` 읽기 동사(bare/list, why, pr, history)가 홈 DB로
  대답하는 동시에 프로젝트-로컬 레거시 스토어(`<root>/.moai/state/todo/backlog.db` 또는
  `<root>/.moai/state/kanban/backlog.db`)가 존재하고 그 `meta.last_seq`가 홈 DB의
  `meta.last_seq`와 다르면, the CLI shall 유령 스토어의 경로와 양쪽 last_seq 값을 밝히는
  한 줄 고지를 stderr에 낸다.
- **REQ-TSS-002**: The CLI shall 고지를 **stderr에만** 낸다 — stdout(`--json` 포함)은
  고지 유무와 무관하게 바이트 동일해야 한다(REQ-BJD-004 선례; foreman 루프가 stdout을
  기계 표면으로 읽는다).
- **REQ-TSS-003**: The 검출 경로 shall 유령 스토어와 홈 DB 어느 쪽도 **쓰지 않는다** —
  마커 파일 부착, 스토어 내부 변경, lock 획득 모두 금지다(REQ-WTQ-001 읽기-경로 순수성
  선례; 유령 스토어는 롤백 스냅샷이라 변경 자체가 롤백 증거 파괴다).
- **REQ-TSS-004**: The 고지 표면과 doctor 점검(§B.2) shall 발산 판별을 **단일 검출기에서**
  얻는다 — 두 번째 프로브는 두 검출기가 갈라질 여지다(SPEC-BACKLOG-JSON-DISCLOSURE-001의
  "no second inspector" 원칙).
- **REQ-TSS-005**: **When** 레거시 스토어가 없거나 양쪽 last_seq가 같으면, the 고지 shall
  아무것도 내지 않는다(REQ-BJD-003 무고지 선례).

### B.2 항목 ② — `moai doctor` 발산 점검 + binary_lag 쌍

- **REQ-TSS-010**: The 새 doctor 점검 shall 레거시 로컬 스토어의 `meta.last_seq`와 홈 DB의
  `meta.last_seq`를 비교해 세 상태를 판정한다 — 발산(양쪽 값 함께 보고), 홈 DB 부재,
  레거시 스토어 부재(이때는 문제 없음으로 PASS).
- **REQ-TSS-011**: The 새 doctor 점검 shall **상수 등록 + `namesAddedAfterBaseline`
  allowlist 항목 + `TestBinaryLag_AllowlistKeysAreLiveNames` 통과**를 한 묶음으로 착지한다.
  이 저장소 규율상 binary_lag 쌍 없는 doctor 점검은 결함이다(카드 t1251 선례).
- **REQ-TSS-012**: **When** 새 점검이 doctor 출력 골든 스냅샷에 나타나면, the 착지 shall
  골든 파일을 `UPDATE_GOLDEN=1`로 재생성해 같은 커밋에 포함한다.
- **REQ-TSS-013**: The doctor 점검 shall 읽기 전용이다 — 어떤 분기에서도 마이그레이션,
  DDL, lock을 일으키지 않는다.

### B.3 항목 ③ — 잔존 저장소 처분 (측정 먼저, 확인 게이트)

대상: 0바이트 쌍 `.moai/backlog.db`, `.moai/state/backlog.db`;
마이그레이션 백업 `.moai/state/todo-merge-backup-20260913/store-{1,2}/`.

- **REQ-TSS-020**: The 처분 절차 shall **삭제 전에** 각 잔존 저장소의 보존 필요성을
  측정하고 그 결과를 기록한다 — 크기·참조 유무·홈 DB 백업/검증 상태.
- **REQ-TSS-021**: The 시스템 shall **운영자와 리드 양쪽의 명시적 확인(카드 t1307 본문
  「파기 전 운영자·리드 확인」) 없이는** 잔존 저장소를 삭제하지 않는다. 자동 삭제
  경로는 존재하지 않는다. **Where** 확인이 기록되지 않았으면, the 처분 shall 증거
  기록까지만 진행하고 멈춘다.
- **REQ-TSS-022**: The 처분 shall 측정 내용, 확인 주체와 시점, 삭제된 경로를 SPEC
  progress 기록(§E.2)에 남긴다.

## C. 성공 기준

1. 유령 스토어가 읽히는 상황에서 읽기 표면이 stderr로 경고한다(오독 사고 재발 차단 —
   2026-09-29 사고의 재발 방지가 이 SPEC의 존재 이유다).
2. `moai doctor`가 발산을 재현 가능하게 판정한다(명령 한 줄로 증명).
3. binary_lag 쌍이 기계적으로 검증된다 — `go test ./internal/cli/ -run '^(TestBinaryLag_OneSeamServesBothSurfaces|TestBinaryLag_NonGitDirectoryKeepsDoctorExitZero|TestBinaryLag_AllowlistKeysAreLiveNames|TestBinaryLag_DoctorCheckNameSetIsUnchanged)$'`(각 분기 anchored 열거형. 실행 시점에 최소 1개 테스트 적중이 필수다 — 빈 적중도 exit 0이라 녹색으로 위장하며, 단일 이름 형태 `'^TestBinaryLag$'`는 일치하는 실제 테스트가 없어 0개 적중이라 금지).
4. 잔존 저장소는 측정·확인·증거 3단계를 통과한 뒤에만 사라진다.

## D. 제약

- 개발 언어/주석/godoc: 영어(`code_comments: en`). 사용자 대상 stderr 문자열: 영어
  (기존 고지 문안 선례와 동일).
- 테스트 임시 디렉터리는 `t.TempDir()` — 실제 홈 DB와 primary 체크아웃 `.moai/`를
  테스트가 건드리지 않는다. 홈 DB 경로는 기존 시임(`HomeDirFn`, `paths.EnvHome`)으로 주입.
- 읽기 동사 stdout 바이트 동일성 유지 — foreman 루프 기계 표면 보호(REQ-BJD-004).

## E. Out of Scope

### Out of Scope — 유령 스토어의 자동 정리/재동기화

- 유령 스토어를 최신 큐로 재동기화하거나 자동 삭제하는 기능. 롤백 스냅샷의 처분은 §B.3의
  수동 확인 게이트 안에서만 일어난다.

### Out of Scope — 홈 DB 마이그레이션 기능 자체의 변경

- 마이그레이션 상태 머신(openEngine State A/B/D), 홈 DB 경로 해석, adoption 정책의 변경.
  본 SPEC은 검출·고지·처분만 다룬다.

### Out of Scope — todo-logic-review-20260929 보고서의 P2..P6 항목

- 측정 보고서가 함께 지적한 P1 외 항목은 본 카드 범위 밖이며 별도 카드로 간다.

## F. 교차 참조

- SPEC-TODO-QUEUE-HOME-CANON-001 — 홈 DB 정식화(근원 SPEC)
- SPEC-BACKLOG-JSON-DISCLOSURE-001 — stderr 고지 표면과 "no second inspector" 원칙
- SPEC-WEB-TODO-QUEUE-001 — 읽기-경로 순수성(REQ-WTQ-001)
- 카드 t472(0바이트 쌓 플래그), t1251(binary_lag 쌍 규율 선례)
