---
id: SPEC-UPDATE-REMOVAL-FAMILY-001
title: "moai update removal family: gated project-side removal, backup-before-removal, emptied-directory pruning, and integrity-probe criterion (rc.30 split B)"
version: "0.1.0"
status: completed
created: 2026-10-17
updated: 2026-10-18
author: "GOOS (via manager-spec, card t1668)"
priority: P0
phase: "v3.2.0-rc.30 target"
module: "internal/cli/update"
lifecycle: spec-anchored
tags: "update, removal, migration, backup, profile, integrity-probe, rc30"
tier: M
related_specs: [SPEC-UPDATE-REPAIR-RC30-001, SPEC-UPDATE-MIGRATION-001, SPEC-UPDATE-MIGRATION-FIX-001]
---

# SPEC-UPDATE-REMOVAL-FAMILY-001 — moai update 제거 가족 (rc.30 분할 B)

## HISTORY

| 버전 | 날짜 | 작성자 | 내용 |
|------|------|--------|------|
| 0.1.0 | 2026-10-17 | GOOS (manager-spec, card t1668) | 초기 plan-phase 초안. 범위는 SPEC-UPDATE-REPAIR-RC30-001 (이하 SPEC A) decision-index.md F1 절에서 분할: AC-RC30-01, -02, -03, -11 과 REQ-RC30-003, -004, U-02, U-10 (ruling d-20261009T235135Z-5e35), AC-RC30-26 과 REQ-RC30-021, U-31 (ruling d-20261010T033220Z-3c5f; decision d-20261010T084752Z-d4be), AC-RC30-27 의 제거 반쪽 (TestMigrationRemovesSharedCopyWithoutProfileReach; AC-RC30-01 과 함께 이동), 그리고 번들 G (sync 감사 finding F12; 리더 이관 판정 "bundle G → t1668 이관"). 이동 항목은 ID 를 그대로 유지한다 (F1 규정). |

## A. 배경

이 SPEC 은 카드 t1668 이다. SPEC A (card t1619, completed) 의 계획 단계에서 Tier L 예산을 맞추기 위해 운영자 승인 아래 분할된 두 형제 카드 중 하나로, "제거 가족" (removal family) 을 받았다. 분할 권위 기록은 SPEC A 의 `decision-index.md` F1 절이며, 이동 항목은 ID 를 유지한다.

제거 가족이 다루는 문제: 2026-10-09 의 미추적 제거 사건 (`.claude/skills/moai/**` 46 파일과 여섯 코어 에이전트가 백업 디렉터리 없이 삭제됨 — SPEC A 항목 U-25) 이 보여준 것처럼, `moai update` 의 프로젝트 측 공통 에셋 제거 경로는 데이터 안전 장치가 결여돼 있었다. SPEC A 의 계획 단계에서 제거 스위치 `removalEnabled` (기본 false) 가 inert 상태로 본 트리에 반입되었고, 제거 경로의 판정과 활성화는 이 카드로 넘어왔다.

이 카드가 착지한 트리의 현재 상태 (2026-10-17, 브랜치 `WT-t1619-spec-b`, base `9b4922d8e` 관측):

- 제거 스위치 `removalEnabled` 존재, 기본 false — `internal/cli/migrate_project_assets.go:36-37`. 게이트 분기는 `:210-216` ("removal is not enabled yet" 보존 분기).
- keep 반쪽 (프로필 도달 시 사본 보존 + 표지)은 SPEC A 가 구현·판정 완료 — `internal/cli/migrate_profile_reach_test.go:52` (`TestMigrationKeepsSharedCopyUnderProfileReach`). 제거 반쪽 테스트 (`TestMigrationRemovesSharedCopyWithoutProfileReach`) 는 존재하지 않는다 (이번 실행 grep, 매치 없음).
- 무결성 프로브의 claudeOnly 수선은 card t1578 이 착지 (`ac141b2cb`, 2026-10-10) — `internal/cli/update_integrity_probe.go:55-57` (`appliesTo`), `:71` (claudeOnly 항목), `:154-156` (미적용 항목 스킵). SPEC A 원래 행이 지정한 판정 테스트 이름 `TestIntegrityProbeCodexOnlyNoClaudeSettings` 는 만들어진 적 없다.
- 마이그레이션 제거 지점 (`os.Remove(p)`, `internal/cli/migrate_project_assets.go:217`) 경로에는 백업 쓰기가 없고, 디렉터리 정리 단계도 없다 (이번 실행 grep — 마이그레이션 경로에 prune/RemoveAll 매치 없음).

## B. Requirements (GEARS)

요구사항 ID 는 SPEC A 에서 그대로 유지된다 (F1 규정). Tier M 요구사항 상한은 16; 이 SPEC 은 3 개를 가진다.

- **REQ-RC30-003** (Unwanted behavior): The update flow shall not delete a project-side managed skill or agent copy (for example `.claude/skills/moai/**`) until the runnable profile-visibility check AC-RC30-01 has passed on the tree being updated. Items: U-01, U-25 (교차 참조 — SPEC A 소유; U-02·U-10 은 REQ-RC30-004 의 항목이다, § C 대장 — plan-audit-iter1 D5 정정). 판정: AC-RC30-02 (게이트), AC-RC30-27 제거 반쪽, AC-RC30-01 (채택 전제, 운영자 단말).
- **REQ-RC30-004** (Event-driven): **When** the migration removes a file, the run backup shall contain a copy of that file, the migration step shall run after the backup step, and an emptied skill directory shall be removed with its last file. Items: U-02, U-10. 판정: AC-RC30-03, AC-RC30-11.
- **REQ-RC30-021** (State-driven): **While** the Claude harness is hidden for a project (codex-only projects, `hideClaude=true`, `internal/template/harness_fs.go` 의 `isHidden`), the managed-surface integrity probe shall not treat the absence of `.claude/settings.json` as damage, and shall still check every other entry of its set. Item: U-31. 판정: AC-RC30-26.

모든 요구사항은 각각 하나 이상의 인수 기준을 가진다. 요구사항 문장은 SPEC A 본문에서 문자 그대로 조각낸 것으로, SPEC A 가 이 항목들의 원 소유자였던 시점의 의미를 보존한다.

## C. 항목 대장

| 항목 | 출처 | 우선순위 | 처분 | 근거 (본 트리 핀, 2026-10-17 관측) | 판정 |
|------|------|----------|------|-----------------------------------|------|
| U-02 | rc.30 감사 must-fix 2 (D2); REQ-RC30-004 | High | 이 SPEC 이 소유 (ruling d-20261009T235135Z-5e35) | 제거 지점 `os.Remove(p)` `internal/cli/migrate_project_assets.go:217`; 마이그레이션 호출 `internal/cli/update_template_sync.go:1247`; 제거 경로에 백업 쓰기 없음 | AC-RC30-03 |
| U-10 | rc.30 감사 must-fix 7 (D5); REQ-RC30-004 | Medium | 이 SPEC 이 소유 (같은 ruling) | 같은 제거 지점이 호출당 파일 하나만 제거; 디렉터리 정리 단계 없음 | AC-RC30-11 |
| U-31 | 리더 인테이크 10-10 (t1578 card-review M3); REQ-RC30-021 | Medium | 이 SPEC 이 판정을 소유; 코드 수선은 리더 정정에 따라 card t1578 이 소유하며 이미 착지 (`ac141b2cb`) | `internal/cli/update_integrity_probe.go:55-57` (`appliesTo`), `:71` (claudeOnly 항목), `:154-156` (스킵 루프) | AC-RC30-26 |
| G (F12) | 번들 G — 프로필 도달 레인 검증. SPEC A 에서 리더 판정으로 이관 ("bundle G (F12): 이 카드 범위에서 뺀다 → t1668 이관", t1619 카드 기록; sync-audit-30659e8e1a.md F12 행. 카드 기록은 git 비추적이므로 출처 표기이지 권위 앵커가 아니다) | Medium | 이 SPEC 이 소유. AC 가 아니다 (sync-audit-ab0adb383-FAIL.md:42 "이 항목은 AC가 아니다") — 검증 묶음이며 결과는 progress.md §E.2 에 기록 | 원장 픽스처: `profile.RecordLastUsedProfileForProject` (`internal/cli/migrate_profile_reach_test.go:41`), keep 반쪽 판정 근거 d-20261010T124923Z-2b09 1항 | AC 아님 — M6 검증 묶음 |

## D. 범위 출처 (분할 추적)

SPEC A `decision-index.md` F1 절 (커밋된 기록) 에서 이 카드에 배정된 행, 원문 인용:

- "Card t1668 (ruling d-20261009T235135Z-5e35): AC-RC30-01, AC-RC30-02, AC-RC30-03, AC-RC30-11 with REQ-RC30-003, REQ-RC30-004 and U-02, U-10. The removal half of AC-RC30-27 (TestMigrationRemovesSharedCopyWithoutProfileReach) moves with AC-RC30-01."
- "Card t1668 (ruling d-20261010T033220Z-3c5f, decision d-20261010T084752Z-d4be): AC-RC30-26 with REQ-RC30-021 and U-31."
- 번들 G (F12) 의 이관은 F1 절에 없다 — sync 감사 N3 가 정확히 그 공백을 기록했다 ("SPEC progress.md has no record of the bundle G transfer to t1668"). 이관의 권위는 (1) 운영자 위임 아래 리더가 이름을 정해 재pick 한 카드 t1668 의 배차 범위 선언 ("Bundle G (item F12, cross-card relation t1642) migrated INTO this card"), (2) SPEC A `acceptance.md:240` 의 커밋된 문장 ("Interim: the lane's own reach test (M2 bundle G) confirms reach and does not replace this criterion") 이다. 이 SPEC 이 그 공백을 커밋된 기록으로 처음 담는다.

분할에서 제외된 것 (SPEC A 잔여): U-01 본체 (keep 반쪽, AC-RC30-27 keep 반쪽과 AC-RC30-28 reach 판정으로 SPEC A 에서 판정됨), U-25 (운영자 필드 런 조사로 deferred), AC-RC30-25 묶음 (card t1669).

## E. 제약

1. **제거는 게이트 뒤에 머무른다.** 이 SPEC 은 제거 기계 장치를 기본 꺼진 상태(`removalEnabled = false`)로 전달한다. 프로덕션에서 스위치를 켜는 일은 AC-RC30-01 이 대상 트리에서 통과된 뒤에만 허용된다 (REQ-RC30-003). 스위치 활성화 자체는 이 SPEC 의 범위가 아니며 채택 게이트(운영자 단말)에 속한다.
2. **프로필 세션 항목은 운영자 단말이다.** AC-RC30-01 은 실제 Claude Code 프로필 세션을 필요로 한다. 워크트리 레인 세션은 실행할 수 없다 (SPEC A Q8, decision d-20261010T223148Z-1e76; 워크트리 가드가 인라인 HOME 교체를 거부한 전례 — ruling d-20261010T124923Z-2b09 2항).
3. **요약 보고 불변식.** 제거 경로의 변화는 SPEC-UPDATE-MIGRATION-001 의 REQ-UPM-031 ("the summary shall never report a deletion-free run while the removal machinery deleted files") 과 REQ-UPM-030/032 를 훼손해서는 안 된다. 현재 `removedPaths` 는 리포터로 흐른다 (`internal/cli/update_template_sync.go:1256`).
4. **dry-run 미리보기 게이트 유지.** 제거 후보 미리보기는 `removalEnabled` 로 게이트돼 있다 (`internal/cli/update_dryrun_preview.go:84`). 이 SPEC 의 제거 경로 변화는 이 게이트를 유지해야 한다.
5. **언어.** 이 SPEC 산출물의 산문은 한국어 (깨끗한 문어체). GEARS 요구사항 문법 토큰(shall/When/While/Where)과 식별자·경로·명령은 원형 그대로 둔다.

## F. 교차 참조와 의존성

- **SPEC-UPDATE-REPAIR-RC30-001** (parent, completed): 분할 출처. 이 SPEC 은 parent 의 REQ-RC30-001, REQ-RC30-002 를 교차 참조한다 (소유 이동 없음). parent 의 AC-RC30-24 는 REQ-RC30-003 을 교차 참조하는 관계로 남는다 (SPEC A acceptance.md:7).
- **SPEC-UPDATE-MIGRATION-001** (completed): REQ-UPM-030/031/032 — 제거 경로가 지켜야 할 요약 보고 불변식.
- **SPEC-UPDATE-MIGRATION-FIX-001** (card t1578): U-31 의 코드 수선 소유자. 이 SPEC 은 그 수선의 판정 기준(AC-RC30-26)을 소유한다.
- **card t1642 (3.2-2-5)**: AC-RC30-28 (`moai doctor` reach 판정) 의 소유자. 번들 G 는 AC-RC30-28 이 실행될 때까지의 잠정 회귀 가드다 (SPEC A acceptance.md:240). t1642 가 reach 판정을 doctor 에 넣으면 번들 G 의 잠정 역할은 끝난다.
- **선행 카드 t1619**: merged-local 확인 (카드 진행 기록, 2026-10-17).

## G. 제외 (Exclusions)

### Out of Scope — 제거 스위치의 프로덕션 활성화

- `removalEnabled` 를 프로덕션에서 true 로 바꾸는 것, 그리고 그 채택 증거 수집은 이 SPEC 의 범위가 아니다. 활성화는 AC-RC30-01 채택 (운영자 단말) 이후의 별도 결정이다 (§ E 제약 1).
- 스위치 기본값 변경: 없음. 기본 false 는 유지된다.

### Out of Scope — keep 반쪽과 reach 판정

- 프로필 도달 시 사본 보존(keep) 동작과 그 판정 (AC-RC30-27 keep 반쪽, AC-RC30-28) 은 SPEC A 와 card t1642 의 소유다. 이 SPEC 은 제거 반쪽만 판정한다.
- 프로필 세션이 사본을 실제로 적재하는지의 확인 — Claude Code 런타임 동작이라 Go 시험으로 볼 수 없고 (ruling d-20261010T124923Z-2b09 1항 원문), 운영자 확인으로 남는다 (AC-RC30-28, t1642).

### Out of Scope — U-25 미추적 제거 사건의 원인 규명

- 2026-10-09 미추적 제거의 로그 추적과 원인 규명은 SPEC A 가 deferred 로 기록한 운영자 몫이다 (3.2-6-2 rc.30 필드 런). 이 SPEC 은 재발 방지 장치(백업·게이트)만 만든다. "Do not fix blindly" 원칙 유지.

### Out of Scope — 요약 출력과 dry-run 미리보기의 표현 변화

- 업데이트 요약의 형식, 배너, 진행 렌더링은 SPEC-UPDATE-MIGRATION-001 개정과 card t1527 계승 묶음의 소유다. 이 SPEC 은 제거 사실이 요약에 빠지지 않는다는 불변식만 지킨다 (§ E 제약 3).

## H. 의사결정 지점

이 SPEC 의 의사결정은 `decision-index.md` 에 기록한다 (interview.decision_gate: on). 네 개의 FOUNDER 행: AC-RC30-27 제거 반쪽의 픽스처 의미 (Q1), 번들 G 의 형태 (Q2), AC-RC30-26 판정 명령의 재지정 (Q3), AC-RC30-01 운영자 단말 항목의 종결 처분 (Q4). 네 질문 모두 기본값이 적용되었고(2026-10-17T11:54:29Z) plan-audit-iter1 이 확정했다; 운영자는 언제든 뒤집을 수 있고, 그 경우 plan.md 의 해결 항목이 결정 기록을 따라 다시 쓰인다.
