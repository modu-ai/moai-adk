---
id: SPEC-USERASSET-DEPLOY-GUARD-001
title: "User-asset install/deploy defect-repair bundle: lock-guard recovery, journal durability, Codex reference normalization, collision safety, frozen-guard user-path protection, deploy-surface and doctor accuracy (card t1591)"
version: "0.1.0"
status: draft
created: 2026-10-09
updated: 2026-10-09
author: manager-spec
priority: P1
phase: "v3.2.0 target"
module: "internal/userassets, internal/cli, internal/template, internal/web, internal/hook"
lifecycle: spec-anchored
tags: "user-assets,install,journal,lock-recovery,codex-normalization,frozen-guard,collision,doctor,agentfm,bundles,defect-repair,t1591"
tier: L
related_specs: [SPEC-USER-ASSET-INSTALL-001, SPEC-PROGRESS-RECORD-IO-001]
---

# SPEC-USERASSET-DEPLOY-GUARD-001 — 사용자 자산·배포 표면 결함 수리 묶음 (카드 t1591)

## HISTORY

| 버전 | 날짜 | 변경 |
|---|---|---|
| 0.1.0 | 2026-10-09 | 최초 plan-phase 아티팩트 발행. 팩토리 카드 t1591 (lane-20 위임, 리더 배차 2026-10-08). 13개 원장 항목을 7개 결함 가족으로 규정 |

## 1. 배경과 목적 (WHY)

본 SPEC은 t1509 "사용자 자산·배포 축" 보존 계약의 후속으로, t1561 게이트 원장과 t1547·t1574·t1498 게이트 릴레이가 누적한 13개 원장 항목(카드 본문이 원전)을 하나의 수리 묶음으로 규정한다. 사용자 공통 자산 설치·갱신의 회복 무결성(잠금·저널), 하니스별 정규화(Codex 변환), 충돌 판정 안전, 보호 구역의 사용자 경로 누락, 배포 표면과 doctor 진단의 정확성이 대상이다.

원장 항목의 성격에 대한 선언: 13개 항목 대부분은 타 레인 턴종료 게이트가 넘긴 **결함 주장(relay)**이다. 텍스트 패턴 릴레이는 가설이지 검증된 결함이 아니다. 본 SPEC의 plan-phase는 각 좌표를 현재 HEAD(`db0c514d3`)에서 재고정(re-anchor)했고, 관측된 것(`observed-at-HEAD`)과 릴레이된 것(`relayed-unverified`)을 전면에서 구별한다(§4 요약표, progress.md 재고정 표 전문). 재고정 결과 일부 항목은 이미 수리되어 있었고(11a, 5의 일부), 일부는 좌표가 이동했으며(8d, 13b), 일부는 판별 불가였다(10a). 이 발견은 묵살하거나 조용히 고치지 않고 §7 열린 질문과 plan.md §B에 기록했다. 모든 항목의 첫 런 페이즈 과업은 본 트리에서 작성·관측하는 RED-first 재현이다.

## 2. 범위 (WHAT)

- 대상: §3의 7개 결함 가족 요구사항 = 13개 원장 항목 전체 (원장 1:1 추적은 acceptance.md §D 매트릭스).
- 상위 계약: 부모 SPEC-USER-ASSET-INSTALL-001 (status: completed)의 C2 경계·REQ-010 사용자 파일 불가침·REQ-023 백업·저널-매니페스트 원자성 계약을 승계하며, 본 SPEC은 그 위의 결함 수리만 규정한다.
- 비대상: §6 Out of Scope.

## 3. 요구사항 (GEARS)

### 3.1 잠금 가족 — R-LOCK (원장 1, 10-P1)

- REQ-LOCK-001: **When** 잠금 획득 프로세스가 `.guard` 마커(`lock_guard_windows.go`의 `path+".guard"`, `lock.go`의 `path+".acquire-guard"`)를 보유한 채 죽으면, the guard shall 마커에 회복 가능한 소유권 증거(PID)를 남기고, 이후 획득자는 소유자 사망을 확인한 뒤 잔존 마커를 회수한다 — **While** 소유자 사망이 입증된(pid 기록 + 사망 확인) 잔존 마커는 회수된다; 소유권 증거가 없는 마커의 자동 회수는 금지되며 획득 거부·대기와 doctor 보고 + 명시적 확인 제거라는 가시 회복 경로로 해소된다. `lock.go`의 `lockOwnerGone` 자세(사망 확인 전 인수 금지)를 재사용한다.
- REQ-LOCK-002: The guard shall windows 빌드와 unix 빌드에서 동일한 회수 의미를 제공한다(플랫폼 패리티 — 원장 10-P1 "init/update/bundle 전면 차단"은 windows 관측). **Where** CI windows 매트릭스가 릴레이 시점 미검증이면, the SPEC shall `GOOS=windows` 빌드 게이트와 플랫폼 중립 표 테스트로 이 패리티를 판정 가능하게 한다.

### 3.2 저널 가족 — R-JRN (원장 3, 6a, 10, 10a-릴레이)

- REQ-JRN-001: **When** 런이 보류 저널을 처음 스테이징할 때(첫 `WriteJournal`), the installer shall 회수된 기존 저널 항목을 그 스테이지에 **스테이징 이전에** 병합한다 — 스테이징부터 매니페스트 저장 사이 어느 중단 지점에서도 회수된 항목이 잃지 아니하며(원장 3의 "영구 collision" 결과 차단), 소유권 증거가 비는 창이 존재하지 아니한다 (install.go:239 축).
- REQ-JRN-002: **When** 회수 항목의 대상 파일이 런 안에서 refresh되면, the installer shall 반입된 항목의 `ExpectedSHA256`를 갱신된 값으로 기록한다 (install.go:297 축).
- REQ-JRN-003: **When** 각 자산 쓰기가 완료할 때마다, the installer shall 해당 항목의 `WriteCompleted` 플래그를 즉시 영속화한다 — 루프 종료 후 일괄 영속화는 결함이다 (install.go:255 축 — 중단+사용자 수정 시 미추적 collision 오판).
- REQ-JRN-004: **When** 인식할 수 없는 `schema_version`의 저널이 적재되면, the loader shall 조용히 디코드하지 않고 진단과 함께 거절한다 (journal.go `LoadJournal`의 버전 게이트 부재 관측 — 원장 10a의 v1/v2 오판 주장은 판별 불가, §7).

### 3.3 Codex 정규화 가족 — R-CNV (원장 7a, 13a, 13b)

- REQ-CNV-001: **When** 설치 대상 자산이 Codex 루트(`RootCodexAgents`의 역할 TOML — install.go:392 축, `RootAgentsSkills`의 사용자 스킬 — install.go:373 축)로 향하면, the installer shall 하니스 고유 참조(CLAUDE.md, `.claude/rules/` 등)를 기존 변환면(`template.NormalizeCodexRoleForDeploy`/`normalizedOpen`)을 재사용해 대상 하니스의 배포 면에 맞게 변환·기록하고, 변환된 바이트를 해시 대상으로 삼는다 — 원문 복사·원문 해시는 결함이다.
- REQ-CNV-002: **Where** 참조의 Codex 측 대응물이 존재하지 않으면, the installer shall 파일 단위로 보고한다 — dangling 참조를 조용히 남기지 아니한다.
- 적용 범위 주의: 13a·13b는 미검증 릴레이(§4)이다. 본 요구는 RED-first 재현이 확정한 좌표 집합에만 적용하며, 재현이 반박하면 plan-audit·런 페이즈의 범위 조정 대상이다.

### 3.4 동결 가드 가족 — R-GRD (원장 9a)

- REQ-GRD-001: The protected-zone guard shall 네 사용자 설치 루트(`userassets.ResolveRoots`, paths.go:38 축) 아래 moai 관리 파일을 보호 경로 집합으로 판정한다. 이를 위해 보호 경로 모델(config 목록과 `resolveZoneTarget`)이 사용자 루트 기반 해석을 지원하도록 확장된다 — 현재 config 목록은 절대 경로 항목을 기각하고(internal/config/protected_zone.go:258 관측) 구면 판정은 루트 내부 매치만 형성한다(protected_zone_path.go:235-255 관측). 사용자 루트 불포함 주장 자체는 릴레이 상태이다.
- REQ-GRD-002: **When** 보호 대상 사용자 폴더 자산(예: plan-auditor.md)의 삭제가 시도되면, the guard shall 프로젝트 경로와 동일하게 거절한다 — 사용자 폴더의 삭제가 allow 통과하는 것은 결함이다.

### 3.5 충돌 판정·confined 쓰기 가족 — R-COL (원장 8c, 12, 6b)

- REQ-COL-001: **When** 충돌 사전 판정이 대상을 읽을 때, the installer shall 먼저 항목 유형을 분류(Lstat 선행)하고 정규 파일이 아닌 대상(FIFO 등)에서 무한정 막히지 아니한다 (install.go:211 축).
- REQ-COL-002: **When** confined 쓰기가 검증된 부모에 기록할 때, the installer shall 부모 디렉터리를 핸들로 고정하고 핸들 경유로 기록한다 — 검증 이후의 경로 재해석이 외부 기록을 만들지 아니한다 (install.go:750 축 — symlink-swap 창의 구조적 폐쇄; 경로 재검증 방식은 TOCTOU를 닫지 못함이 관측됐다).
- REQ-COL-003: **When** the installer가 설치 대상 중 실행 가능한 스크립트 자산(예: .sh)을 기록할 때, the installer shall 실행 권한(0755)을 보존해 설치한다 — 0644로의 하락은 결함이다 (install.go:743 축 — navigator-audit.sh 재현).

### 3.6 배포·번들 표면 가족 — R-SRF (원장 4, 5-잔여, 8a, 8b, 8d, 11a, 11b)

- REQ-SRF-001: **When** 프로젝트 공통 자산 이동이 미등록 mirrorSkills 사본(매니페스트 항목 없음)을 만나면, the cli shall 분류·보고하여 가시화하고, 제거는 생성 출처 입증(매니페스트·저널 증거) 또는 명시적 삭제 승인 하에서만 수행한다 — 해시 일치는 소유권 증명이 아니다(사용자 복사본도 해시가 일치한다). 조용한 무보존 방치는 결함이다 (migrate_project_assets.go:89·:119 축).
- REQ-SRF-002 (회귀 가드): **When** doctor가 구 `.codex/agents` 배출물 검사에서 비교되지 않은 항목을 만나면, the doctor shall 이를 누락(CheckFail)으로 보고한다 — compared 미달의 OK 집계로의 퇴행은 결함이다. HEAD에서 이미 CheckFail이 관측된다(doctor_agentemit_embed.go:150-165) — 본 요구는 그 불변의 고정이다.
- REQ-SRF-003: **When** `skills disable --codex`가 사용자 설치 스킬을 대상으로 하고 프로젝트 미러가 없으면, the cli shall 사용자 계층 면을 해석해 동작한다 — absent(0) 무시("Nothing to disable")는 결함이다 (internal/cli/skills.go 축 — 원장 좌표 deployer_mode.go:37은 M7 퇴역으로 생 면이 이동, §4).
- REQ-SRF-004: **When** 사용자 에이전트와 프로젝트 에이전트가 같은 이름이면, the web form shall 동명 행을 스코프 출처가 병기된 단일 설정으로 통합 렌더하고, POST는 그 단일 설정을 저장하며 재독록까지 일관한다 — 중복 행과 첫 행 우선(`PostFormValue`)에 의한 소실은 결함이다 (agentfm.go:128·:282 축; 저장 계약은 design.md §7 — 이름당 단일 값 유지).
- REQ-SRF-005: **When** 번들 선택이 설치되면, the installer shall 카탈로그 `depends_on` 클로저를 함께 설치하고, 보존 집합은 보존된 번들의 의존 번들을 포함한다 (install.go:459 축 — 수집 경로의 클로저 확장 부재 관측; prune 측 R-f-② 유예 팔 존재는 관측, `depends_on` 인지 여부는 재현으로 변별).
- REQ-SRF-006 (회귀 가드): **When** update가 확인창에서 취소되면, the cli shall 프로젝트 관리 자산을 바이트 그대로 유지한다 — 원장 11a의 "migrate가 확인창 선행" 결함은 HEAD에서 이미 수리 관측(update.go:566-574 이전 코멘트 + update_template_sync.go:1185 확인창 뒤 호출)이므로, 본 요구는 회귀 가드로 고정한다.
- REQ-SRF-007: **When** init이 사용자 자산 ensure 실패로 끝나면, the cli shall 재실행 시 "already initialized" 거절 대신 부족분 완성 재개 경로를 제공한다 (init.go:947·:905 축).

### 3.7 doctor 진단 가족 — R-DOC (원장 2, 7b, 9b)

- REQ-DOC-001: **When** doctor가 매니페스트 적재에 실패하면, the doctor shall 실패 등급(파손 포함)을 별도 행으로 보고하고 진단 중 복구 사본을 덮어쓰지 아니한다 — 읽기 전용 적재 자세. 도달 가능한 RED 앵커는 프로젝트 측 checkProjectVsLock의 임의 적재 오류 OK 위장(doctor_user_install.go:134-137 관측)이고, 사용자 측 checkUserInstallIntegrity는 이미 정직하다(:24-38 관측 — 본 팔은 회귀 가드).
- REQ-DOC-002: **When** 프로젝트 트리가 잠금 비교 가능한 면을 전혀 갖지 않으면, the doctor shall 그 공허 조건을 보고한다 — "project tree matches the lock file" 위장은 결함이다 (doctor_user_install.go:165·:184 축).
- REQ-DOC-003: **When** 워크플로 조회 루트를 선택할 때, the doctor shall 필수 파일의 존재로 선택한다 — 빈 workflows 디렉터리의 존재만으로 프로젝트 경로를 택해 정상 사용자 워크플로 전체를 읽기 실패로 만드는 것은 결함이다 (doctor_harness.go:57 축).
- REQ-DOC-004: **When** 워크플로 폴백이 발동하면, the doctor shall L1 검사를 프로젝트 디렉터리에 유지하고 L6 참조 해석이 `skillsDir` 통째 교체로 오 판정되지 아니하게 한다 (doctor_harness.go:58 축 — 원장 7b, 원장 9b와 실행 시 병합).

### 3.8 비기능 요구사항

- REQ-NFR-001: **While** 본 SPEC의 수리가 착지하는 동안, the codebase shall 영향 패키지(internal/userassets, internal/cli 해당 파일, internal/template, internal/web, internal/hook)의 커버리지를 85% 이상으로, 저널·잠금·변환의 critical 경로를 90% 이상으로 유지한다.

## 4. 좌표 상태 — 관측 vs 릴레이 (요약)

전문(결정적 코드 라인 인용 포함)은 progress.md 재고정 표. 본 표는 판정 요약이다.

| 원장 | 릴레이 좌표 | HEAD(db0c514d3) 판정 |
|---|---|---|
| 1 / 10-P1 | lock_guard_windows.go:23 | observed-at-HEAD — 마커 무소유권·무회수 |
| 2 | doctor_user_install.go:134/:165 | observed-at-HEAD — 적재 오류 OK 위장·공허 일치 |
| 3 | install.go:239 | observed-at-HEAD — 첫 스테이징 전 병합 부재 |
| 4 | migrate_project_assets.go:89/:119 | observed-at-HEAD — 미등록 사본 영구 보존 |
| 5 | deployer.go:192/:347, bundle.go:145 | observed-at-HEAD — 단, ListTemplates 제외는 **이미 수리**(t1547 수리 라운드, 758→414 측정 코멘트), emission compared 축도 **repaired-at-HEAD/회귀 가드**(CheckFail 이미 관측 — AC-013·REQ-SRF-002 가드화). 생존 분항: skills disable. `internal/template/bundle.go`는 **HEAD 부재** |
| 6a / 6b | install.go:297/:743 | observed-at-HEAD — 반입 해시 미갱신·0644 하드코딩 |
| 7a | install.go:392 | observed-at-HEAD — 무변환 복사 경로. 참조 내용 주장은 relayed-unverified |
| 7b / 9b | doctor_harness.go:58/:57 | observed-at-HEAD — 통째 교체·빈 디렉터리 존재 선택 |
| 8a | agentfm.go:128/:282 | observed-at-HEAD — 홈 우선 중복 행·PostFormValue 첫 행 |
| 8b | install.go:459 | observed-at-HEAD — 수집에 클로저 확장 부재. prune R-f-② 유예 팔 존재(remove.go) |
| 8c | install.go:211 | observed-at-HEAD — `os.ReadFile` 직접 판정 |
| 8d | deployer_mode.go:37 | **좌표 퇴역**(M7 리타이어, 45행) — 생 면 internal/cli/skills.go로 재표현 |
| 9a | paths.go:38 | observed-at-HEAD — 네 루트 정의. 보호 집합 config 적재(pre_tool.go:355), 사용자 루트 불포함 주장은 relayed |
| 10 | install.go:255 | observed-at-HEAD — 플래그 일괄 영속화 |
| 10a | install.go:260 | **판별 불가** — `SchemaVersion=1` 뿐, v1/v2 오판 좌표 부재. REQ-JRN-004로 재정식화 |
| 11a | update.go:565 | **이미 수리 관측** — 확인창 뒤 호출로 이전됨. 회귀 가드로 고정 |
| 11b | init.go:947/:905 | observed-at-HEAD — 재개 경로 부재 |
| 12 | install.go:750 | observed-at-HEAD — rename 직전 재검증 없음 |
| 13a / 13b | install.go:373/:584 | 복사 경로 관측. 결함 주장 자체는 **relayed-unverified**(런 재현 확정 대상 — 카드 본문 명시) |

## 5. 성공 기준

- §3의 모든 REQ에 1:1 대응하는 기계 판정 가능한 AC가 acceptance.md에 존재하고, 각 AC는 M0에서 세 분류 중 하나로 판정된다 — **RED-first**(재현 관측 후 수리 마일스톤에서 GREEN 전환), **born-green 회귀 가드**(현행 불변의 고정, RED 관측 시 퇴행 결함 보고), **빌드·측정 게이트**(절차 판정). 분류는 각 AC에 표기된다.
- REQ-NFR-001의 커버리지 기준 충족.
- 열린 질문(§7)이 plan-audit에서 회신·기록된다.

## 6. Out of Scope

### Out of Scope — t1547 도메인 정합 축 (형제 카드 분리)
- `update/reconcile.go:169/:590/:204/:184`, `update_template_sync.go:541` — 리더가 별도 후속 카드로 보유하는 축이다.
- 형제 카드 t1594 (t1547 r5–r12 잔여 축) — "발행 단위 분리"로 인접 분할된 카드다. 본 SPEC이 흡수하지 않는다.

### Out of Scope — 타 원장 이관분
- t1560 게이트 이관분 — SPEC-PROGRESS-RECORD-IO-001 원장에 상호참조로 존속하며 본 SPEC이 중복 소관을 만들지 않는다. 참고: 그 SPEC은 카드 t1598에서 발급 진행 중(in-flight)이며 아직 main에 미착지라 본 트리 카탈로그에는 없다 — 착지 전까지 상호참조는 명목상 참조다.
- t1547 r10 회람 중 `agentfm.go:126` — :128 동일 결함의 좌표 이동으로 원장에서 중복 제외되었다(각주 승계).

### Out of Scope — 기능 확장 및 부활
- 카탈로그 스키마 확장 — depends_on 클로저는 기존 `DependsOn` 필드(catalog_loader.go:70)의 소비만 추가한다.
- plugin carrier·deployer_mode 부활 — SPEC-USER-ASSET-INSTALL-001 M7의 퇴역을 유지한다.
- 프로젝트 경로 보호 동작의 변경 — 사용자 설치 경로의 추가만 한다.
- 이전 세대 /tmp 오버레이 스위트(`/tmp/t1547-review-overlay.json` 등)에의 의존 — 스테일 판정, 본 트리 재현으로 대체한다.

## 7. 열린 질문 (plan-audit 회신 대상)

1. 원장 11a: HEAD에서 이미 수리 관측. RED-first가 재현 불가로 끝나면 REQ-SRF-006을 회귀 가드(현행)로 확정할지, 원장 항목을 소관 종료로 표기할지 — plan-audit 판정 대상.
2. 원장 5: ListTemplates 제외 분항은 이미 수리(deployer.go:332-340 코멘트, 758→414 측정). 생존 분항(emission compared-0/12, skills disable)만 REQ-SRF-002/003으로 규정했다 — 수리 분항의 중복 수리 금지 확인 대상.
3. 원장 8d: 좌표 퇴역(deployer_mode.go:37 메커니즘 소멸). REQ-SRF-003은 생 면(skills.go)에 재표현했으나, 원장 좌표와의 대응 관계를 plan-audit이 확인한다.
4. 원장 10a: v1/v2 오판 주장은 HEAD에서 판별 불가(SchemaVersion=1 뿐, LoadJournal 버전 게이트 부재만 관측). REQ-JRN-004로 재정식화한 것이 타당한지.
5. 원장 13a/13b: 미검증 릴레이(카드 본문 "미검증 전달"). 재현이 반박할 경우 REQ-CNV 적용 범위 축소 — 런 페이즈 M0의 관측이 우선한다.
6. `internal/template/bundle.go:145` 좌표: 파일이 HEAD에 없다. 최근생 면은 internal/cli/bundle.go(RemoveBundle 영역)이나 ListTemplates 행은 없다 — 원장 좌표의 소관 종료 또는 재지정 판정 대상.
7. 원장 8b prune 측: remove.go에 R-f-② 의존 유예 팔이 이미 존재한다. 재현이 prune 결함을 반박하면 REQ-SRF-005는 설치 클로저 절반만 남는다.
8. windows 커버리지: 원장 1은 windows 마커이며 CI windows 매트릭스는 릴레이 시점 미검증이다. REQ-LOCK-002의 빌드 게이트 + 플랫폼 중립 표 테스트 구성이 충분한지.
