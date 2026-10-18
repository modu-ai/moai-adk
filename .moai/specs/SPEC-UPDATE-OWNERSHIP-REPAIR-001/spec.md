---
id: SPEC-UPDATE-OWNERSHIP-REPAIR-001
title: "moai update ownership verdict and legacy-migration repair: evidence-based per-file ownership, profile-reach keep reasons, case-variant paths, emptied-directory pruning, and a remaining-file list (follow-up to SPEC-UPDATE-REMOVAL-FAMILY-001)"
version: "0.3.1"
status: completed
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
| 0.2.0 | 2026-10-18 | GOOS (manager-spec, card t1865) | 계획 감사 iteration 1 (FAIL 0.68, MP-8 실패, 차단 결함 5건) 의 결함 D1-D15 를 처리한 수정판. 요구사항 15 → 16 (REQ-OWR-016: 이전 매니페스트 사본이 실패하면 제거 차단), 인수 기준 19 → 20 (AC-OWR-20). REQ-OWR-008/-011/-015 문구를 고치고 `removal-eligible` 정의를 § B 한 곳으로 모았다. 결함별 처분과 증거는 progress.md § E.1 의 "Plan audit iteration 1 repair" 표에 있다. |
| 0.3.0 | 2026-10-18 | GOOS (manager-spec, card t1865) | 계획 감사 iteration 2 (FAIL 0.76, 차단 결함 N1-N5, 필수 백엔드 codex 실패) 의 수정판과 리더 지시 두 건. 요구사항 16 → 17 (REQ-OWR-017: 이전 세대 조회 한도·손상 허용·보존 기간 무변경), 인수 기준 20 → 21 (AC-OWR-21). REQ-OWR-002 에 이전 기록의 출처가 `template_managed` 여야 한다는 조건을 넣었다 (N2). REQ-OWR-005 를 좁은 예외 하나 (`launchCandidateIsUsable` 의 기존 `os.Stat` 한 호출) 와 호출 지점 동결 기준선으로 다시 썼다 (N1, N8, 리더 판정 OQ-5). REQ-OWR-012/-013 이 버전 일치 건너뛰기를 명시한다 (N5). REQ-OWR-015/-016 은 리더 결정 d-20261018T040918Z-def2 (Q2 = 가, 조건부) 의 조건 (1)-(5) 을 싣는다. 결함별 처분은 progress.md § E.1 의 "Plan audit iteration 2 repair" 표. |
| 0.3.1 | 2026-10-18 | GOOS (manager-spec, card t1865) | 천장 보정 커밋 (계획 감사 iteration 3 FAIL 0.84 vs 0.85, 차단 결함 N12·N13; 리더 판정 보드 d-20261018T044408Z-3e9d, 상시 규칙 C). 정확히 두 결함만 고쳤다: N12 (AC-OWR-21 이 이전 매니페스트의 임시 쓰기 + 이름 바꾸기 생성을 관측하지 못함 — design.md D-3 항목 3 에 관측점 두 개 추가, AC-OWR-21 (a) 재설계) 와 N13 (plan.md M8·M9 의 낡은 의무). 요구사항·인수 기준 id 집합, 마일스톤 구성, Q2/Q5/OQ 결정은 바뀌지 않았다. N14-N21 과 N4 잔여는 decision-index.md 에 수용 부채로 기록했다. |

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
- **처분 (disposition)**: 소유권 분류와 별개의 두 번째 단계 — `keep(<이유>)` 또는 `removal-eligible`. 분류가 처분을 정하지 않는다.
- **`removal-eligible`** (정의는 여기 한 곳이다. design.md D-1 과 acceptance.md 는 이 정의를 인용한다): 판정 함수가 이번 실행의 입력으로 내리는 후보 판정이며, 다섯 조건을 모두 갖춘 파일이다 — (i) 분류가 `managed`, (ii) 프로필 도달 유지 대상이 아님, (iii) 사용자 루트 대응물이 확인됨, (iv) 제거 스위치 (`removalEnabled`) 가 켜져 있음, (v) 루트 경계 안 (심링크 탈출 아님). 하나라도 빠지면 `keep(<이유>)` 이고, 스위치가 꺼진 기본 구성에서는 `keep(removal-gate-off)` 이다. 실행 시점에만 알 수 있는 조건 — 백업 사본 쓰기 성공, 읽기·삭제 OS 오류 — 은 판정 밖이다. 따라서 실제로 제거된 집합은 언제나 `removal-eligible` 집합의 부분집합이고, 그런 실패가 없으면 둘은 같다. 판정 함수는 제거하지 않는다.
- **프로필 루트 (profile root)**: launch profile 의 설정 디렉터리 (`<profile base>/<name>`, 프로필 세션에서 `CLAUDE_CONFIG_DIR` 이 가리키는 곳) 아래의 `skills`·`agents` 디렉터리와 그 하위, 그리고 t1702 설계의 `claude-skills@<profile>`·`claude-agents@<profile>` 철자. 코드 수준 사실 (이번 실행 측정): 이 트리의 Go 소스에는 `claude-skills@` 문자열이 없다 (acceptance.md § L-7).
- **기존 접촉 (existing contact)**: 이 카드가 시작되기 전부터 갱신 경로의 호출 닫힘 (`go list -deps ./internal/cli`) 안에 있던 프로필 관련 호출 지점. 둘로 나눈다. (i) **이름 붙은 예외 하나**: `internal/profile.launchCandidateIsUsable` 안의 `os.Stat(filepath.Join(baseDir, name))` 한 호출 (`internal/profile/profile.go:439`, 커밋 `5fb8920ee` 에서 관측) — 프로필 설정 디렉터리 항목 자체의 존재와 디렉터리 여부만 본다. 리더 판정 (OQ-5, 2026-10-18): 이 호출은 카드 이전부터 있었고 REQ-OWR-005 위반이 아니라 "이 카드가 더하지 않는 기존 접촉" 이다. (ii) **동결된 기준선의 나머지**: `profile` 패키지 심볼을 부르는 기존 호출 지점 전부 — 예: `profile.ResolveLaunchProfileForProject` (`migrate_project_assets.go:185`), `profile.GetCurrentNameForProject`·`profile.ReadPreferences`·`profile.SyncToProjectConfig` (`update.go:693-698`), `profile.ReadPreferences(profile.GetCurrentName())` (`update_participation.go:49`). 이 호출들은 프로필 설정 디렉터리의 `preferences` 파일이나 항목 존재를 다루며 `skills`·`agents` 하위를 열지 않는다 (코드 읽기 관측; acceptance.md § L-12). 예외가 아니라 호출 지점별로 개수까지 동결된 기준선 행이며, 새 호출 지점이나 늘어난 개수는 전부 AC-OWR-08 소스 스캔을 붉게 한다.
- **증명된 managed (attested)**: 매니페스트에 항목이 없는 파일이 세 증거 — 배포 버전, 원래 배포 경로, 이전 매니페스트 기록 — 를 모두 갖춰 `managed` 로 분류된 경우. 파일 자신의 바이트에서 나오지 않은 증거여야 한다.
- **잔존 행 (remaining row)**: 마이그레이션이 제자리에 둔 파일 하나의 기록 — 경로, 소유권 분류, 이유, 빠진 증거 항목.
- **템플릿 관리 출처값 (template-managed provenance)**: `internal/manifest` 의 상수 `manifest.TemplateManaged` (`"template_managed"`) 하나. 같은 패키지의 다른 값 — `UserModified`, `UserCreated`, `Deprecated`, `GeneratedManaged` — 과 빈 값·알 수 없는 값은 증거가 되지 못한다. 배포기는 매니페스트에 없는 기존 파일을 만나면 `UserCreated` 로 기록하고 건너뛴다 (`internal/template/deployer.go:265-267`), 그래서 바이트가 템플릿과 같은 파일이 이전 세대에 `UserCreated` 기록을 가질 수 있다.
- **다시 실행 (re-run) 의 두 종류**: 갱신은 프로젝트 템플릿 버전이 바이너리 버전과 같고 `--force` 가 없고 템플릿 트리 해시도 같으면 조정 단계에 들어가기 전에 "이미 최신" 으로 끝난다 (`internal/cli/update_template_sync.go:1239-1248`, `versionMatchSkipsSync`). **조정에 닿는 실행** (첫 실행, `--force`, 버전이나 템플릿 트리가 다른 실행) 과 **건너뛰는 실행** 을 구분한다. 이 SPEC 은 그 관문을 바꾸지 않는다.

## C. Requirements (GEARS)

요구사항 문장은 GEARS 문법 토큰을 원형으로 둔다. 각 요구사항은 § E 의 추적표에서 카드 범위 항목 (1)-(7) 과 인수 기준에 연결된다.

- **REQ-OWR-001** (Ubiquitous): The migration, the dry-run preview, and the remaining-file report shall obtain each file's ownership class from one verdict function whose classes are `managed`, `user-modified`, `user-created`, and `unregistered`, decided from the project manifest record and the file's current bytes.
- **REQ-OWR-002** (Unwanted behavior): The update shall not classify an unregistered file as `managed` on the strength of byte equality with a template — the embedded template tree included — unless the deployed version, the original deployed path, and a prior manifest record of that path whose recorded provenance is the template-managed value are all confirmed from evidence that does not derive from the file's own bytes; a prior record of any other provenance value, or of none, shall not count as evidence.
- **REQ-OWR-003** (Event-driven): **When** an unregistered file lacks any one of the three evidence items, the update shall leave the file unchanged — no removal, rename, overwrite, or manifest adoption — and shall record it as `unregistered` together with the names of the missing evidence items.
- **REQ-OWR-004** (State-driven): **While** the launch ledger maps the project to a profile, the migration shall keep each template-managed copy under the common-asset roots, shall not classify it `removal-eligible`, and shall record whether a user-root counterpart with a matching hash was confirmed.
- **REQ-OWR-005** (Unwanted behavior): The update shall not open, read, enumerate, stat, walk, supply, or write any profile root (the `skills` and `agents` directories under a profile configuration directory, and the `claude-skills@<profile>` and `claude-agents@<profile>` directories), shall not add a call site of any symbol that produces or opens a path in a profile configuration directory beyond the frozen existing contacts, and shall not change `internal/profile`, `internal/config/userroots.go` or `internal/config/protected_zone.go`; the single existing contact this requirement names as tolerated is the `os.Stat` of the profile configuration directory entry itself inside `launchCandidateIsUsable` (`internal/profile/profile.go:439`), which this SPEC neither calls anew nor alters, and no other function, call, file or package is excepted.
- **REQ-OWR-006** (Ubiquitous): The update shall give each file under `.claude/commands/` and `.claude/agents/` exactly one disposition per run, and after every run — an interrupted and then recovered run included — every command and agent file the template carries shall be present.
- **REQ-OWR-007** (Event-driven): **When** a path found on disk differs from a manifest key only in letter case and, on a case-insensitive filesystem, names the same file as that key, the verdict function shall treat the two spellings as one file, apply the registered entry's verdict, report the spelling difference, and not rename the file; on a case-sensitive filesystem the two spellings shall remain distinct files.
- **REQ-OWR-008** (Event-driven): **When** a skill directory (a directory below `.claude/skills/` or `.agents/skills/`) is empty and the prior manifest record shows that every file once recorded under it was template-managed and none remains on disk, the migration shall remove the directory — whether this run or an earlier one emptied it — and otherwise shall keep the empty directory and record that no provenance record covers it.
- **REQ-OWR-009** (Ubiquitous): The update shall produce a deterministic remaining-file list with one row per file the migration leaves in place under the four common-asset roots, each row carrying path, ownership class, keep reason, and missing evidence items, and shall render it grouped by reason with per-file rows under `--verbose`.
- **REQ-OWR-010** (Event-driven): **When** the migration removes a file, the run backup shall hold a byte-identical copy of that file and a copy of the prior manifest, such that returning both to their original locations reproduces the pre-run bytes of the file and of the manifest.
- **REQ-OWR-011** (Unwanted behavior): The `moai update --dry-run` command — every step it runs, the ownership preview included — shall not create, modify, or delete any file or directory, and the removal candidates it announces shall be exactly the files the verdict function marks `removal-eligible`.
- **REQ-OWR-012** (Ubiquitous): A repeated update over an unchanged tree that reaches reconciliation shall leave the project tree byte-identical outside the timestamped run-backup directories and shall produce an identical remaining-file list; a repeated update that the version-match gate skips shall run no migration and shall leave the tree and every report unchanged.
- **REQ-OWR-013** (Event-driven): **When** an update run is interrupted or fails partway through the migration, the update shall have preserved every file it removed in the run backup, and a re-run that reaches reconciliation after recovery shall reach the same final state as an uninterrupted run, without resurrecting a removed manifest entry and without leaving an emptied managed directory; this SPEC does not change the version-match gate, so a plain same-version re-run that the gate skips is not a repair path (plan.md OQ-6).
- **REQ-OWR-014** (Unwanted behavior): The update shall not enable the removal switch: `removalEnabled` shall keep its default `false`, no production code path shall set it to true, and every removal-related test shall enable it only inside its own fixture.
- **REQ-OWR-015** (Ubiquitous): The update shall record in each manifest entry it deploys in a run the moai binary version that wrote it and shall not set, back-fill or guess that record on any entry that run did not deploy; each deploy shall refresh the manifest-level version record to the writing binary's version at the deploy-stage save and shall change the manifest-level deployment timestamp only when that version changes, so that a repeat run at the same version leaves the manifest bytes unchanged; and each update run shall write the manifest as loaded at the start of the run, byte for byte, into its run backup, creating that copy atomically (a temporary file renamed into place) so that a partial or failed write never reads as a valid generation.
- **REQ-OWR-016** (Event-driven): **When** the run backup cannot hold the byte copy of any removal candidate or the prior-manifest copy, the migration shall remove no file in that run and shall list each candidate as kept with the reason that the backup failed, while the rest of the update continues.
- **REQ-OWR-017** (Ubiquitous): The prior-generation lookup shall read at most the newest `priorManifestGenerations` (8) run backups as a lookup bound only, shall treat a generation that is absent, unreadable, partial, or not a valid manifest as no evidence (the file stays preserved), shall treat an entry without the deployed-version field as lacking that evidence, and shall not delete, prune or shorten the retention of any run backup; manifest files written before this SPEC shall still load unchanged.

REQ-OWR-015 ~ -017 은 결정 지점 Q2 (decision-index.md) 에 걸린다. Q2 는 리더가 옵션 (가) 로 조건부 결정했다 (보드 d-20261018T040918Z-def2, 2026-10-18; 설계 결정일 뿐이며 사용자 보안 승인도, 실제 홈 접근 승인도, Kickoff 사전 승인도 아니다). 따라서 (가) 가 이 SPEC 의 기록된 계획이다. 옵션 (나) 의 뺄셈 지도 (plan.md § G.1) 는 기록된 대안으로만 남는다. 조건 일곱 가지는 REQ-OWR-015 ~ -017 과 AC-OWR-06, -20, -21 이 싣는다. 확인 전과 무관하게 REQ-OWR-001 과 -003 은 증거가 없으면 모든 미등록 파일을 보존하고, REQ-OWR-002 의 "증명된 managed" 양성 경로는 함수 수준에서 증거를 직접 공급하는 시험으로 닫힌다 (plan.md M1a).

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
| (7) 안전 속성 | — | fix (검증 중심) | 실행 백업의 바이트 사본, 이전 매니페스트 원본 바이트 사본, 복원 영역 동일성, 잔존 목록 | 백업 실패 (파일 바이트 또는 매니페스트 사본) 시 그 실행은 파일을 하나도 제거하지 않고 모두 보존 |

## E. 추적표

| 카드 항목 | 요구사항 | 인수 기준 |
|-----------|----------|-----------|
| (1) | REQ-OWR-001 | AC-OWR-01, AC-OWR-16 |
| (6) | REQ-OWR-002 | AC-OWR-01, -02, -03, -05 |
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
| (6) 증거 생산 | REQ-OWR-015 | AC-OWR-05, -06, -17, -21 |
| (7) | REQ-OWR-016 | AC-OWR-20, -18 |
| (6), (7) 증거 읽기 | REQ-OWR-017 | AC-OWR-06, -21 |

요구사항 17개 (Tier L 상한 25), 인수 기준 21개 (상한 25). 인수 기준이 없는 요구사항은 없다.

## F. 제약

1. **구속 제약 (리더, board d-20261018T020724Z-48fb).** (a) 이 SPEC 과 구현은 프로필 루트 (`claude-skills@<profile>/`, `claude-agents@<profile>/`) 에 쓰지 않고, 프로필 루트를 열거하거나 공급하지 않으며, `internal/config/userroots.go` 와 `internal/config/protected_zone.go` 를 건드리지 않는다. 프로필 처리는 읽기 전용 확인이며 불확실하면 보존한다. `internal/profile` 도 바꾸지 않는다 (카드 지시). t1702 는 3.3 보류를 유지하고, 이 SPEC 은 그 판정을 되돌리거나 우회하거나 의존하지 않는다. 요구사항이 프로필 루트 쓰기나 F2 보호를 필요로 하게 되면 멈추고 blocker 보고를 낸다.
2. **t1668 안전 게이트 유지.** `removalEnabled` 기본 false (`internal/cli/migrate_project_assets.go:38-39`), 프로덕션 활성화 없음, 제거 관련 시험은 픽스처 안에서만 스위치를 주입한다. 삭제 로직의 추가, 보호 완화, 별도 보안 권한, `removalEnabled` 활성화, 카드 t1702 범위가 어떤 항목에 섞여 들어오면 그 항목은 이 SPEC 에 싣지 않고 멈춰 보고한다 (리더 결정 d-20261018T040918Z-def2 조건 7).
3. **시험 실행 범위.** 카드 워크트리와 픽스처/임시 HOME 만 쓴다. `TMPDIR` 는 `/private/tmp` 아래 (HOME 이나 저장소 아래 금지, 카드 토큰이 없는 이름이며 디렉터리는 사전 점검 P0 (plan.md § C) 이 만든다). 실제 홈, mo.ai.kr, `moai update` 재실행, 바이너리 교체를 쓰지 않는다. 시험은 `HOME` 과 `CLAUDE_CONFIG_DIR` 을 직접 지정한다.
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
| AC-OWR-02 | 보존 픽스처 1 — 내장 템플릿의 실재 경로와 바이트가 같은 미등록 파일 (템플릿 조회 성공을 먼저 단언, 승격 변이 대조군 포함) | 002, 003 |
| AC-OWR-03 | 보존 픽스처 2 — 사용자 수정 파일 | 002 |
| AC-OWR-04 | 보존 픽스처 3 — 출처 불명 미등록 파일, 이유와 빠진 증거 기록 | 003 |
| AC-OWR-05 | 양성 픽스처 4 — 증거 완비 파일만 managed, 증거 하나씩 빼거나 이전 기록 출처가 `template_managed` 가 아니면 보존 (변이 시험, 함수 수준 + 운영 원천 한 건) | 002, 015 |
| AC-OWR-06 | 배포 버전 기록 (이번 실행이 배포한 항목에만), 이전 매니페스트 원본 바이트 보존, 이전 세대 원천 읽기 | 015, 017 |
| AC-OWR-07 | 프로필 지도 프로젝트: 대응 파일 없는 사본 보존과 이유 | 004 |
| AC-OWR-08 | 프로필 루트 무접근 — `go list -deps` 로 도출한 호출 닫힘 전체의 심볼 단위 소스 스캔 (호출 지점 동결 기준선, 이름 붙은 예외 하나) + 변이 대조군 + 정규화한 차분 카나리 + 파일 불변 | 005 |
| AC-OWR-09 | commands 한 처분, 존재 (같은 버전 / 낡은 기록 두 변형) | 006 |
| AC-OWR-10 | 중단 뒤에도 commands/agents 탐색 가능 (commands 와 agents 의 형식 분리) | 006 |
| AC-OWR-11 | 대소문자 변형 — 양쪽 파일시스템 모드 | 007 |
| AC-OWR-12 | 대소문자 변형 — 사용자 매니페스트 키, `.agents/skills` 철자 | 007 |
| AC-OWR-13 | 관리 파일만 있던 빈 디렉터리 정리 | 008 |
| AC-OWR-14 | 잔존 목록과 313개 보존 (재실행은 `--force` 로 조정에 닿게 하고 닿았음을 단언) | 003, 009 |
| AC-OWR-15 | 백업·복원 바이트 동일 — 복원 영역 (제거·재작성된 파일과 매니페스트) 의 동일성 + 나머지 차이가 계산된 기대 집합과 같음 | 010 |
| AC-OWR-16 | `moai update --dry-run` 전체 진입의 무변경 (디렉터리 mtime 포함)과 판정 일치 | 001, 011 |
| AC-OWR-17 | 반복 갱신 멱등 (`--force` 로 조정에 닿은 실행 + 건너뛰는 실행 사례 분리) | 012, 015 |
| AC-OWR-18 | 중단 → 복구 → 재실행 수렴 (재실행은 `--force`) | 013, 016 |
| AC-OWR-19 | 제거 스위치 기본값과 프로덕션 비활성 | 014 |
| AC-OWR-20 | 백업 실패 시 제거 0건 (fail-closed) — 운영 호출 경로 전체 (Backup 단계 → 마이그레이션) 로 검증 | 016 |
| AC-OWR-21 | 이전 세대 원자적 생성, 손상·부분·읽기 불가 = 증거 없음, 옛 JSON 왕복, 다운그레이드, 같은 버전 반복, 보존 기간 무변경 | 015, 017 |

## H. 교차 참조와 의존성

- **리더 결정 d-20261018T040918Z-def2** (2026-10-18): Q2 = 옵션 (가), 조건부. 설계 결정일 뿐이며 사용자 보안 승인도, 실제 홈 접근 승인도, Kickoff 사전 승인도 아니다 (Kickoff 는 독립 계획 감사 뒤에 정한다). REQ-OWR-015 ~ -017 이 조건을 싣는다.
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
- 호출 순서: Backup 단계 안에서 마이그레이션 (`internal/cli/update_template_sync.go:815-817`), 배포 단계 저장 (`:672-673`), 지연된 `skill.md` 이름 바꾸기 (`:369`). 버전 일치 건너뛰기: `:1239-1248` (`versionMatchSkipsSync`, `internal/cli/update_template_sync_guard.go:55`). 이 트리에서 측정한 동작: 첫 실행은 건너뛰지 않고, 같은 버전의 일반 재실행은 건너뛰며, `--force` 실행은 건너뛰지 않는다 (acceptance.md § L-13).
- 기존 백업 보존 정책: `backup.PruneBackups(projectRoot, 5, pastKeep)` 가 최신 5개 실행을 통째로 두고 그보다 오래된 실행은 선택 규칙이 요구하는 것만 남긴다 (`internal/cli/update_template_sync.go:944`, `internal/cli/update/backup/retention_prune.go:82`). 이 SPEC 은 그 정책을 바꾸지 않는다 (design.md D-3).
- 미리보기 판정: `internal/cli/update_dryrun_preview.go:84`, `:120`. 읽기 전용 로더 `internal/cli/update/reconcile.go:563`.
- commands: 관리 청소 대상 `internal/cli/update/deploy/deploy.go:65-75`, 기본 경로는 reconcile (`update_template_sync.go:609`).
- 이미 있는 이름 바꾸기 장치: `internal/cli/update/skillresidue/residue.go` (`.claude/skills` 의 배포되는 스킬만).
- `internal/config/userroots.go` 는 이 트리에 없다. `internal/config/protected_zone.go` 는 있고 건드리지 않는다.
- 프로필 접근의 기존 사실: 마이그레이션이 부르는 런치 원장 판정 `profile.ResolveLaunchProfileForProject` (`migrate_project_assets.go:185`) 는 `launchCandidateIsUsable` 안에서 `os.Stat(<profile base>/<name>)` 로 프로필 설정 디렉터리 항목의 존재와 디렉터리 여부만 본다 (`internal/profile/profile.go:439`). 안의 `skills`·`agents` 는 열지 않는다. 이 접근은 Q1 (DECIDED) 의 매핑 판정에 속하는 기존 코드이며 이 SPEC 이 더하거나 바꾸지 않는다. 갱신 경로 호출 닫힘 (`go list -deps ./internal/cli`) 은 모듈 안 패키지 142개이며 `internal/profile`·`internal/config`·`internal/manifest`·`internal/template`·`internal/core/project`·`internal/userassets` 를 모두 포함한다 (acceptance.md § L-11). `profile` 패키지의 경로 생산 또는 열기 심볼은 acceptance.md § L-12 에 나열했다.
- `probeCaseSensitiveFS` (`update_cleanup.go:387`) 는 비시험 호출자가 없다 (acceptance.md § L-9). 판정·미리보기 경로는 이 함수를 부르지 않는다 (design.md D-4).
- `moai update --dry-run` 의 전체 진입은 `update.go:361` 의 분기이며 (`previewReconciliation` 은 그 안의 한 단계), 기존 시험 헬퍼 `runDryRunUpdateInProject` (`update_dry_run_reach_test.go:84`) 가 이 분기를 `runUpdate` 로 구동한다.

## J. 제외 (Exclusions)

### Out of Scope — 제거 스위치의 프로덕션 활성화

- `removalEnabled` 를 프로덕션에서 true 로 바꾸는 것, 기본값 변경, 켜는 운영 경로의 추가는 이 SPEC 의 범위가 아니다 (REQ-OWR-014).
- AC-RC30-01 (프로필 세션 가시성 확인) 의 실행과 채택 증거 수집은 t1668 과 운영자 단말 몫이다.

### Out of Scope — 프로필 루트의 열거·쓰기·보호

- 다중 프로필 루트 도달은 카드 t1702 (3.3 보류) 의 범위이다. 열거가 필요해지면 구현하지 않고 멈춰 보고한다.
- `internal/config/userroots.go`, `internal/config/protected_zone.go` 의 변경, 프로필 쪽 대응물의 존재·해시 읽기는 하지 않는다 (decision-index.md Q5, 리더 판정 2026-10-18: 프로필 루트는 읽지 않고 프로필 쪽 상태는 `not-checked` 로 기록하며 프로필 지도 사본은 무조건 보존한다).
- 기존 런치 원장 판정이 하는 프로필 설정 디렉터리 항목의 `os.Stat` (`launchCandidateIsUsable`, `internal/profile/profile.go:439`; § B 의 이름 붙은 예외 하나) 은 이 SPEC 이 바꾸지도 새로 부르지도 않는다. 그 접근을 없애는 것은 Q1 의 매핑 판정을 바꾸는 일이며 범위 밖이다. `internal/profile` 은 수정하지 않는다.
- 버전 일치 관문 (`versionMatchSkipsSync`) 의 변경은 범위 밖이다 (plan.md OQ-6).

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

### Out of Scope — 백업 보존 정책 변경과 삭제 로직

- 기존 백업 보존 정책 (`PruneBackups`) 의 변경, 새 자동 삭제·정리, 보존 기간 단축은 하지 않는다 (REQ-OWR-017). `priorManifestGenerations` 는 조회 한도일 뿐이다.
- 보호 완화, 별도 보안 권한, 실제 홈 접근, 카드 t1702 범위는 이 SPEC 이 싣지 않는다.

## K. 의사결정 지점

이 SPEC 의 결정은 decision-index.md 에 있다 (`interview.decision_gate: on`). 다섯 행: Q1 (프로필 지도 사본 유지, DECIDED), Q2 (배포 버전 기록과 이전 매니페스트 보존 — 리더가 옵션 (가) 로 조건부 결정, 보드 d-20261018T040918Z-def2. 제품 수준 행이며 이 결정은 설계 결정일 뿐 사용자 보안 승인·실제 홈 접근 승인·Kickoff 사전 승인이 아니다), Q3 (잔존 목록의 형태, 기본값 적용), Q4 (대소문자 변형의 처리 범위, 기본값 적용), Q5 (프로필 쪽 대응물 읽기 — 리더 판정 2026-10-18 로 보수적 읽기 확정, 라벨은 증거 대기를 유지). Q2 의 운영자 확인 여부가 자율 Kickoff 에 미치는 효과는 이 문서가 정하지 않는다 (독립 계획 감사 뒤 리더와 Kickoff 관문이 정한다).
