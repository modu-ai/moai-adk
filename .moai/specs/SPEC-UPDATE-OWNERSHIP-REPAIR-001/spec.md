---
id: SPEC-UPDATE-OWNERSHIP-REPAIR-001
title: "moai update ownership verdict and legacy-migration repair: evidence-based per-file ownership, profile-reach keep reasons, case-variant paths, emptied-directory pruning, and a remaining-file list (follow-up to SPEC-UPDATE-REMOVAL-FAMILY-001)"
version: "0.1.0"
status: draft
created: 2026-10-18
updated: 2026-10-18
author: "GOOS (via manager-spec, card t1865)"
priority: P1
phase: "v3.2.0 target"
module: "internal/cli"
lifecycle: spec-anchored
tags: "update, ownership, provenance, migration, profile-reach, case-insensitive, remaining-list, rc36"
tier: L
related_specs: [SPEC-UPDATE-REMOVAL-FAMILY-001, SPEC-UPDATE-REPAIR-RC30-001, SPEC-UPDATE-MIGRATION-001]
---

# SPEC-UPDATE-OWNERSHIP-REPAIR-001 — moai update 소유권 판정·이전 호환 수리

## HISTORY

| 버전 | 날짜 | 작성자 | 내용 |
|------|------|--------|------|
| 0.1.0 | 2026-10-18 | GOOS (manager-spec, card t1865) | 초기 plan-phase 초안. 운영자 지시 10-18 (Class C, P1, t1668 후속) 의 현장 진단 일곱 항목과 미등록 파일 보존 강화 조건 (6) 을 요구사항 15개와 인수 기준 19개로 옮겼다. 코드 조사는 research.md, 데이터 모델 결정은 design.md, 결정 지점은 decision-index.md 에 있다. 리더 구속 제약 (board d-20261018T020724Z-48fb) 을 § F 에 그대로 실었다. |

## A. 배경

카드 t1865 는 t1668 (SPEC-UPDATE-REMOVAL-FAMILY-001, lane-11) 의 후속이다. t1668 은 백업 후 제거, 제거 뒤 빈 스킬 디렉터리 정리, 제거 게이트 (`removalEnabled` 기본 false) 를 만들었고, 이 카드는 그 위에서 "무엇을 관리 파일로 볼 것인가" 를 바로잡는다. t1668 이 소유한 세 기계는 다시 구현하지 않는다. 이 트리는 t1668 브랜치 `WT-t1619-spec-b` 의 `e1e163320` 병합을 포함하며, 착지는 t1668 뒤이다.

운영자가 전달한 현장 진단 (사용자 전달값이며 ADK 에서 재측정한 값이 아니다):

- 실행 바이너리 rc.36-t1822c / ee981a932 (dirty), 프로젝트의 템플릿 기록은 rc.35, `removalEnabled=false`.
- 프로필 도달 예외로 에이전트 6개와 스킬 1개가 보존되지만 대응하는 사용자/프로필 파일이 없다.
- commands 17개가 이전 대상에서 빠졌다가 재배포되었다.
- OWASP 스킬에서 `skill.md` / `SKILL.md` 대소문자가 어긋난다.
- 관리 파일만 있던 빈 디렉터리 3개가 정리되지 않았다.
- `.agents/skills` 313개가 매니페스트에 미등록이어서 보존되었다.

2026-10-18 운영자 수정: 미등록 파일의 관리 소유권은 해시 일치만으로 확정하거나 자동 채택하지 않는다. 배포 버전, 원래 배포 경로, 이전 매니페스트 (또는 동등한 출처 증거) 까지 확인된 파일만 관리 대상으로 판정한다. 출처가 불명이면 템플릿과 바이트가 같아도 사용자/미등록으로 보존하고 잔존 목록에 사유를 적는다. 313개 전체를 관리로 재분류한 뒤 제거하는 것, 미등록이라는 사실만으로 삭제하는 것, 폴더 전체 삭제, 제거 스위치를 무조건 켜는 것은 승인되지 않았다.

조사 결과의 핵심 (research.md C절): 같은 질문에 답하는 소유권 판정이 네 곳에 흩어져 있고, 각각 다른 입력과 어휘를 쓴다. 파일별 배포 버전은 어디에도 기록되지 않으며, 이전 매니페스트 세대를 보존하는 곳도 없다. 이 SPEC 은 판정을 한 곳으로 모으고, 증거가 없을 때의 기본값을 "보존 + 사유 기록"으로 고정한다.

## B. 용어

- **소유권 분류 (ownership class)**: 네 값의 닫힌 집합 — `managed`, `user-modified`, `user-created`, `unregistered`. 파일 하나가 정확히 하나를 받는다.
- **처분 (disposition)**: 소유권 분류와 별개의 두 번째 단계 — `keep(<이유>)` 또는 `removal-eligible`. 분류가 처분을 정하지 않는다. `removal-eligible` 이어도 실제 제거는 t1668 의 게이트 (`removalEnabled`, 백업 선행, 사용자 대응물 확인) 뒤에서만 일어난다.
- **증명된 managed (attested)**: 매니페스트에 항목이 없는 파일이 세 증거 — 배포 버전, 원래 배포 경로, 이전 매니페스트 기록 — 를 모두 갖춰 `managed` 로 분류된 경우. 파일 자신의 바이트에서 나오지 않은 증거여야 한다.
- **잔존 행 (remaining row)**: 마이그레이션이 제자리에 둔 파일 하나의 기록 — 경로, 소유권 분류, 이유, 빠진 증거 항목.

## C. Requirements (GEARS)

요구사항 문장은 GEARS 문법 토큰을 원형으로 둔다. 각 요구사항은 § E 의 추적표에서 카드 범위 항목 (1)-(7) 과 인수 기준에 연결된다.

- **REQ-OWR-001** (Ubiquitous): The migration, the dry-run preview, and the remaining-file report shall obtain each file's ownership class from one verdict function whose classes are `managed`, `user-modified`, `user-created`, and `unregistered`, decided from the project manifest record and the file's current bytes.
- **REQ-OWR-002** (Unwanted behavior): The update shall not classify an unregistered file as `managed` on the strength of byte equality with a template — the embedded template tree included — unless the deployed version, the original deployed path, and a prior manifest record of that path are all confirmed from evidence that does not derive from the file's own bytes.
- **REQ-OWR-003** (Event-driven): **When** an unregistered file lacks any one of the three evidence items, the update shall leave the file unchanged — no removal, rename, overwrite, or manifest adoption — and shall record it as `unregistered` together with the names of the missing evidence items.
- **REQ-OWR-004** (State-driven): **While** the launch ledger maps the project to a profile, the migration shall keep each template-managed copy under the common-asset roots, shall not classify it `removal-eligible`, and shall record whether a user-root counterpart with a matching hash was confirmed.
- **REQ-OWR-005** (Unwanted behavior): The update shall not read, enumerate, supply, or write any profile root (the `claude-skills@<profile>` and `claude-agents@<profile>` directories), and shall not change `internal/config/userroots.go` or `internal/config/protected_zone.go`.
- **REQ-OWR-006** (Ubiquitous): The update shall give each file under `.claude/commands/` and `.claude/agents/` exactly one disposition per run, and after every run — an interrupted and then recovered run included — every command and agent file the template carries shall be present.
- **REQ-OWR-007** (Event-driven): **When** a path found on disk differs from a manifest key only in letter case and, on a case-insensitive filesystem, names the same file as that key, the verdict function shall treat the two spellings as one file, apply the registered entry's verdict, report the spelling difference, and not rename the file; on a case-sensitive filesystem the two spellings shall remain distinct files.
- **REQ-OWR-008** (Event-driven): **When** a skill directory (a directory below `.claude/skills/` or `.agents/skills/`) is empty and the prior manifest record shows that every file once recorded under it was template-managed and none remains on disk, the migration shall remove the directory — whether this run or an earlier one emptied it — and otherwise shall keep the empty directory and record `empty-dir-no-provenance`.
- **REQ-OWR-009** (Ubiquitous): The update shall produce a deterministic remaining-file list with one row per file the migration leaves in place under the four common-asset roots, each row carrying path, ownership class, keep reason, and missing evidence items, and shall render it grouped by reason with per-file rows under `--verbose`.
- **REQ-OWR-010** (Event-driven): **When** the migration removes a file, the run backup shall hold a byte-identical copy of that file and a copy of the prior manifest, such that returning both to their original locations reproduces the pre-run bytes of the file and of the manifest.
- **REQ-OWR-011** (Unwanted behavior): The dry-run shall not create, modify, or delete any file or directory, and the candidates it announces shall be exactly those the verdict function marks `removal-eligible`.
- **REQ-OWR-012** (Ubiquitous): A repeated update over an unchanged tree shall leave the project tree byte-identical outside the timestamped run-backup directories and shall produce an identical remaining-file list.
- **REQ-OWR-013** (Event-driven): **When** an update run is interrupted or fails partway through the migration, the update shall have preserved every file it removed in the run backup, and a re-run after recovery shall reach the same final state as an uninterrupted run, without resurrecting a removed manifest entry and without leaving an emptied managed directory.
- **REQ-OWR-014** (Unwanted behavior): The update shall not enable the removal switch: `removalEnabled` shall keep its default `false`, no production code path shall set it to true, and every removal-related test shall enable it only inside its own fixture.
- **REQ-OWR-015** (Ubiquitous): Each manifest entry a deploy writes shall record the version of the moai binary that wrote it, and each update run shall retain the manifest as loaded at the start of the run in the run backup, so that a later run can confirm a deployed version and a prior record.

REQ-OWR-015 은 결정 지점 Q2 (decision-index.md) 가 운영자 확인을 기다린다. 확인 전에도 REQ-OWR-001 과 -003 은 이행 가능하며 (증거가 없으면 모든 미등록 파일이 보존된다), REQ-OWR-002 의 "증명된 managed" 양성 경로와 REQ-OWR-010 의 매니페스트 사본이 REQ-OWR-015 에 의존한다 (plan.md M1b).

## D. 항목별 판정표

판정은 `preserve` (제자리 보존), `migrate` (조건이 갖춰질 때만 제거 대상), `fix` (동작 수리) 중 하나로 적는다. "근거"는 판정을 받치는 증거이고, "근거 불완전 시 기본값"은 증거가 빠졌을 때 취하는 행동이다. 어느 행도 근거가 불확실한 채로 `migrate` 로 기울지 않는다.

| 항목 | 현장 증상 (전달값) | 판정 | 근거 | 근거 불완전 시 기본값 |
|------|-------------------|------|------|------------------------|
| (1) 파일별 소유권 판정 | 실행 rc.36 dirty, 템플릿 기록 rc.35 | fix — 판정 함수 하나로 통합 | 매니페스트 항목 (`Provenance`, `TemplateHash`) + 현재 바이트의 해시. 내장 템플릿과의 바이트 일치는 판정 근거가 아니다 | `unregistered` 또는 `user-modified` 로 보존하고 이유를 기록 |
| (2) 프로필 도달 | 에이전트 6·스킬 1 보존, 대응 파일 없음 | preserve | 런치 원장의 프로젝트→프로필 매핑 + 사용자 루트 대응물의 존재와 해시. 프로필 루트는 읽지 않는다 | 보존. 이유 `profile-reach`, 대응물 상태 `confirmed`/`absent`/`unconfirmed` 를 함께 기록. 프로필 쪽 대응물은 `not-checked` |
| (3) commands | 17개가 이전 제외 후 재배포 | fix — 한 실행에 한 처분 + 존재 보증 | 프로젝트 매니페스트 + reconcile 분류 (template-owned/user-modified/user-owned) + 템플릿 운반 | 제자리 보존. 삭제 후 재배포를 별개 처분으로 세지 않는다 |
| (4) 대소문자 | `skill.md` / `SKILL.md` 불일치 | fix — 한 파일로 취급 | `os.SameFile` 이 참 + 키가 대소문자만 다름 + 파일시스템 대소문자 프로브 | 두 철자를 별개 파일로 두고 보존·보고. 이름 바꾸기는 하지 않는다 (skillresidue 소유) |
| (5) 빈 디렉터리 | 3개 정리 누락 | migrate (조건부 정리) | 이전 매니페스트 기록: 디렉터리 안에 있던 파일이 전부 `template_managed` 이고 지금은 전부 없음 + 디렉터리가 비어 있음 | 빈 디렉터리 유지 + `empty-dir-no-provenance` 기록 |
| (6) 미등록 313 | `.agents/skills` 313개 보존 | preserve | 배포 버전 + 원래 배포 경로 + 이전 매니페스트 기록 셋 모두. 바이트 일치만으로는 불충분 | 템플릿과 바이트가 같아도 `unregistered` 로 보존, 빠진 증거 항목을 잔존 행에 기록 |
| (7) 안전 속성 | — | fix (검증 중심) | 실행 백업의 바이트 사본, 이전 매니페스트 사본, 트리 해시, 잔존 목록 | 실패 시 해당 파일을 제거하지 않고 보존 |

## E. 추적표

| 카드 항목 | 요구사항 | 인수 기준 |
|-----------|----------|-----------|
| (1) | REQ-OWR-001 | AC-OWR-01, AC-OWR-16 |
| (6) | REQ-OWR-002 | AC-OWR-02, -03, -05 |
| (6) | REQ-OWR-003 | AC-OWR-02, -04, -14 |
| (2) | REQ-OWR-004 | AC-OWR-07 |
| (2) | REQ-OWR-005 | AC-OWR-08 |
| (3) | REQ-OWR-006 | AC-OWR-09, -10 |
| (4) | REQ-OWR-007 | AC-OWR-11, -12 |
| (5) | REQ-OWR-008 | AC-OWR-13 |
| (6), (7) | REQ-OWR-009 | AC-OWR-14 |
| (7) | REQ-OWR-010 | AC-OWR-15 |
| (7) | REQ-OWR-011 | AC-OWR-16 |
| (7) | REQ-OWR-012 | AC-OWR-17 |
| (7) | REQ-OWR-013 | AC-OWR-18 |
| (1), (7) 게이트 | REQ-OWR-014 | AC-OWR-19 |
| (6) 증거 생산 | REQ-OWR-015 | AC-OWR-05, -06 |

요구사항 15개 (Tier L 상한 25), 인수 기준 19개 (상한 25). 인수 기준이 없는 요구사항은 없다.

## F. 제약

1. **구속 제약 (리더, board d-20261018T020724Z-48fb).** (a) 이 SPEC 과 구현은 프로필 루트 (`claude-skills@<profile>/`, `claude-agents@<profile>/`) 에 쓰지 않고, 프로필 루트를 열거하거나 공급하지 않으며, `internal/config/userroots.go` 와 `internal/config/protected_zone.go` 를 건드리지 않는다. 프로필 처리는 읽기 전용 확인이며 불확실하면 보존한다. t1702 는 3.3 보류를 유지하고, 이 SPEC 은 그 판정을 되돌리거나 우회하거나 의존하지 않는다. 요구사항이 프로필 루트 쓰기나 F2 보호를 필요로 하게 되면 멈추고 blocker 보고를 낸다.
2. **t1668 안전 게이트 유지.** `removalEnabled` 기본 false (`internal/cli/migrate_project_assets.go:38-39`), 프로덕션 활성화 없음, 제거 관련 시험은 픽스처 안에서만 스위치를 주입한다.
3. **시험 실행 범위.** 카드 워크트리와 픽스처/임시 HOME 만 쓴다. `TMPDIR` 는 `/private/tmp` 아래 (HOME 이나 저장소 아래 금지, 카드 토큰이 없는 이름). 실제 홈, mo.ai.kr, `moai update` 재실행, 바이너리 교체를 쓰지 않는다. 시험은 `HOME` 과 `CLAUDE_CONFIG_DIR` 을 직접 지정한다.
4. **필수 시험군.** 명령/에이전트 탐색 회귀, 프로필 환경 (`CLAUDE_CONFIG_DIR`), macOS 대소문자 무구분 경로 (`skill.md` vs `SKILL.md`), 중도 중단 → 복구 → 재실행, 백업/복원 바이트 동일, dry-run 무변경, 반복 갱신 뒤 잔존 목록, 보존 픽스처 네 가지 (§ G 의 AC-OWR-02 ~ -05).
5. **금지.** 시험 제외, 제한시간 증가, 가드/감사 우회, 거절 기록 삭제, 승인창 선택. 가드나 감사가 추가 승인을 요구하면 막힘으로 보고하고 우회하지 않는다.
6. **RED 선행.** 모든 행동 변화는 실패하는 시험을 먼저 기록한 뒤 구현한다 (plan.md).
7. **요약 보고 불변식.** 제거 경로의 변화는 SPEC-UPDATE-MIGRATION-001 의 REQ-UPM-030/031/032 (삭제가 있었는데 삭제 없음 보고 금지) 를 훼손하지 않는다. 제거된 경로는 기존대로 `removedPaths` 로 요약에 올라간다 (`internal/cli/migrate_project_assets.go:323`).
8. **언어.** 산출물의 산문은 한국어 (깨끗한 문어체), GEARS 토큰과 식별자·경로·명령은 원형을 둔다.

## G. 인수 기준 개요

세부 Given-When-Then, 판정 명령, 기대 출력, RED 이유는 acceptance.md 에 있다. 이 절은 요구사항과의 대응만 적는다.

| AC | 요지 | 요구사항 |
|----|------|----------|
| AC-OWR-01 | 판정 함수의 네 분류 | 001 |
| AC-OWR-02 | 보존 픽스처 1 — 템플릿과 바이트가 같은 사용자 작성 미등록 파일 | 002, 003 |
| AC-OWR-03 | 보존 픽스처 2 — 사용자 수정 파일 | 002 |
| AC-OWR-04 | 보존 픽스처 3 — 출처 불명 미등록 파일, 이유와 빠진 증거 기록 | 003 |
| AC-OWR-05 | 양성 픽스처 4 — 증거 완비 파일만 managed, 증거 하나씩 빼면 보존 (변이 시험) | 002, 015 |
| AC-OWR-06 | 배포 버전 기록과 이전 매니페스트 보존 | 015 |
| AC-OWR-07 | 프로필 지도 프로젝트: 대응 파일 없는 사본 보존과 이유 | 004 |
| AC-OWR-08 | `CLAUDE_CONFIG_DIR` 환경, 프로필 루트 무접근 | 005 |
| AC-OWR-09 | commands 한 처분, 존재 | 006 |
| AC-OWR-10 | 중단 뒤에도 commands/agents 탐색 가능 | 006 |
| AC-OWR-11 | 대소문자 변형 — 양쪽 파일시스템 모드 | 007 |
| AC-OWR-12 | 대소문자 변형 — 사용자 매니페스트 키, `.agents/skills` 철자 | 007 |
| AC-OWR-13 | 관리 파일만 있던 빈 디렉터리 정리 | 008 |
| AC-OWR-14 | 잔존 목록과 313개 보존 | 003, 009 |
| AC-OWR-15 | 백업·복원 바이트 동일 | 010 |
| AC-OWR-16 | dry-run 무변경, 판정 일치 | 001, 011 |
| AC-OWR-17 | 반복 갱신 멱등 | 012 |
| AC-OWR-18 | 중단 → 복구 → 재실행 수렴 | 013 |
| AC-OWR-19 | 제거 스위치 기본값과 프로덕션 비활성 | 014 |

## H. 교차 참조와 의존성

- **SPEC-UPDATE-REMOVAL-FAMILY-001** (card t1668, 선행): 백업 후 제거, 빈 스킬 디렉터리 정리, 제거 게이트를 소유한다. 이 SPEC 은 그 코드 (`backupRemovedAsset`, `pruneEmptiedSkillDir`, `migrationRunContext`, `verifyMigrationFinalState`, `removalEnabled`) 위에서 일하고 다시 만들지 않는다. frontmatter 의 `depends_on` 에는 넣지 않았다 — 이 트리에서 그 SPEC 의 status 가 `in-progress` 라서 Phase 1 의 depends_on 사전 점검이 미충족으로 읽어 실행을 막기 때문이다. 착지 순서 (t1668 뒤) 는 이 문서와 plan.md 가 선언한다.
- **SPEC-UPDATE-REPAIR-RC30-001** (completed): 프로필 도달 옵션 (c) 의 결정 근거 (decision-index.md Q2, d-20261010T060848Z-dc3b). REQ-OWR-004 는 그 결정을 그대로 따른다.
- **SPEC-UPDATE-MIGRATION-001** (completed): 요약 보고 불변식 REQ-UPM-030/031/032.
- **카드 t1702** (보류): 다중 프로필 루트 도달. 이 SPEC 의 범위 밖이며 의존하지 않는다.
- **카드 t1642 / t1643**: `KeptBy` 표지를 읽어 되찾는 doctor 와 프로필 도달 판정. 이 트리에는 `KeptBy` 를 읽는 비시험 코드가 없다 (research.md D(2)).

## I. 저장소 사실 (이 SPEC 이 의존한 것)

좌표는 HEAD `cdd72a814` 의 것이다. 다른 트리에서는 다시 측정해야 한다. 근거 명령과 출력은 research.md 에 있다.

- 제거 스위치와 게이트: `internal/cli/migrate_project_assets.go:38-39` (기본 false), 게이트 분기 `:296-302`.
- 소유권 판정 순서: 같은 파일 `:198-302`. 미등록 파일은 `:216-226` 에서 보존·보고된다.
- 프로필 도달 유지: `:181-185` (매핑 판정), `:255-272` (유지 + `KeptBy`). 사용자 대응물은 이 유지를 게이트하지 않는다 (`:259-260`).
- 사용자 대응물 확인: `:377-452` — `$HOME` 기준 네 루트만 본다 (`internal/userassets/paths.go:36-56`).
- 백업: `backupRemovedAsset` `:84-104`. 빈 디렉터리 정리: `pruneEmptiedSkillDir` `:526-559`, 호출 `:332-336`. 최종 상태 검증: `:573-601`.
- 매니페스트 데이터 모델: `internal/manifest/types.go:93-113` — 파일별 배포 버전 필드가 없다. 매니페스트 수준 `Version`/`DeployedAt` 은 `internal/core/project/initializer.go:607-608` 에서만 쓰인다. 갱신 배포가 기록하는 것은 `TemplateTreeHash` 이다 (`internal/cli/update_template_sync_guard.go:29-40`).
- 거울: 심링크 거울과 복사 폴백 거울은 의도적으로 매니페스트에 등록되지 않는다 (`internal/template/skill_mirror.go:1-21`, `internal/template/deployer.go:310-318`). 매니페스트에 없는 기존 파일을 만난 배포기는 `UserCreated` 로 기록하고 건너뛴다 (`deployer.go:265-267`).
- 호출 순서: Backup 단계 안에서 마이그레이션 (`internal/cli/update_template_sync.go:815-817`), 배포 단계 저장 (`:672-673`), 지연된 `skill.md` 이름 바꾸기 (`:369`).
- 미리보기 판정: `internal/cli/update_dryrun_preview.go:84`, `:120`. 읽기 전용 로더 `internal/cli/update/reconcile.go:563`.
- commands: 관리 청소 대상 `internal/cli/update/deploy/deploy.go:65-75`, 기본 경로는 reconcile (`update_template_sync.go:609`).
- 이미 있는 이름 바꾸기 장치: `internal/cli/update/skillresidue/residue.go` (`.claude/skills` 의 배포되는 스킬만).
- `internal/config/userroots.go` 는 이 트리에 없다. `internal/config/protected_zone.go` 는 있고 건드리지 않는다.

## J. 제외 (Exclusions)

### Out of Scope — 제거 스위치의 프로덕션 활성화

- `removalEnabled` 를 프로덕션에서 true 로 바꾸는 것, 기본값 변경, 켜는 운영 경로의 추가는 이 SPEC 의 범위가 아니다 (REQ-OWR-014).
- AC-RC30-01 (프로필 세션 가시성 확인) 의 실행과 채택 증거 수집은 t1668 과 운영자 단말 몫이다.

### Out of Scope — 프로필 루트의 열거·쓰기·보호

- 다중 프로필 루트 도달은 카드 t1702 (3.3 보류) 의 범위이다. 열거가 필요해지면 구현하지 않고 멈춰 보고한다.
- `internal/config/userroots.go`, `internal/config/protected_zone.go` 의 변경, 프로필 쪽 대응물의 존재·해시 읽기는 하지 않는다 (decision-index.md Q5).

### Out of Scope — 미등록 파일의 일괄 재분류·자동 채택·삭제

- `.agents/skills` 313개를 관리로 일괄 재분류하는 것, 매니페스트에 자동 채택하는 것, 미등록이라는 사실만으로 삭제하거나 폴더를 통째로 지우는 것은 승인되지 않았다.
- 거울 복사본을 배포 시점에 매니페스트나 별도 원장에 등록하는 배포기 변경은 이 SPEC 이 하지 않는다 (design.md D-5 에 후속 후보로만 적는다).

### Out of Scope — 출력 TUX, U-25 원인 규명, 이름 바꾸기 소유

- 업데이트 출력의 표현은 카드 t1669 의 소유이다. 이 SPEC 은 잔존 목록을 원장 행으로 내는 데까지만 정한다.
- 2026-10-09 미추적 제거 사건 (U-25) 의 원인 규명은 하지 않는다.
- `skill.md` → `SKILL.md` 이름 바꾸기는 `skillresidue` 의 소유로 남는다. 이 SPEC 의 마이그레이션은 이름을 바꾸지 않고 같은 파일로 판정만 한다.

### Out of Scope — 복원 명령과 파일 모드 보존

- `removed-assets/` 를 되돌리는 `moai` 하위 명령은 만들지 않는다. 복원의 바이트 동일은 사본을 원위치에 되돌리는 절차에 대해서만 검증한다.
- 백업 사본의 파일 모드 보존 (현재 `defs.FilePerm` 으로 정규화) 은 범위 밖이다. 잔여 위험으로 기록한다.

## K. 의사결정 지점

이 SPEC 의 결정은 decision-index.md 에 있다 (`interview.decision_gate: on`). 다섯 행: Q1 (프로필 지도 사본 유지, DECIDED), Q2 (배포 버전 기록과 이전 매니페스트 보존 — 운영자 확인 대기, 제품 수준), Q3 (잔존 목록의 형태, 기본값 적용), Q4 (대소문자 변형의 처리 범위, 기본값 적용), Q5 (프로필 쪽 대응물 읽기, 증거 대기). Q2 가 비어 있는 동안 자율 Kickoff 는 막히고 운영자에게 간다.
