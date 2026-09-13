---
id: SPEC-UPDATE-SETTINGS-BASE-SNAPSHOT-001
title: "moai update 가 .claude/settings.json 의 템플릿 값 변경을 전달하도록 — 배포 직후 렌더 스냅숏을 3-way 병합 base 로"
version: "0.4.0"
status: completed
created: 2026-09-11
updated: 2026-09-12
author: manager-spec (card t656)
priority: P1
phase: "v3.2.0 target"
module: "internal/cli/update"
lifecycle: spec-anchored
tags: "update, settings-json, merge, snapshot, provenance, t656"
era: V3R6
tier: M
depends_on: ["SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001"]
related_specs: ["SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001", "SPEC-UPDATE-MERGE-CONFLICT-BLIND-001", "SPEC-UPDATE-HOOK-DELIVERY-001", "SPEC-UPDATE-REINSTALL-LOOP-001"]
---

# SPEC — `.claude/settings.json` 템플릿 값 변경 전달

## HISTORY

- 2026-09-11 — v0.1.0, plan 단계 산출물 작성(Tier M: spec.md + plan.md + acceptance.md + progress.md). 카드 **t656**, init/update 감사 항목 **F3**. 설계 선택지 메모는 `.moai/reports/t656/design-options.md` 에 있고, 운영자가 그중 A1 · B1 · C1 을 확정했다(§B).
- 2026-09-11 — v0.2.0, plan-audit 1회차(FAIL 0.73, `.moai/reports/t656/plan-audit-iter1.md`)를 반영했다.
  - 스냅숏을 대기본과 확정본 두 상태로 정의했다(F-06, §B.2).
  - 배포가 기존 파일을 건너뛴 흐름에서는 기록하지 않는다는 REQ-USB-016 을 새로 두었다(F-05).
  - REQ-USB-013 을 둘로 나눴다(F-15). 013 은 sections 스냅숏, 014 는 공용 엔진이다. 종전 REQ-USB-014(git 무시)는 REQ-USB-002 에 흡수했다.
  - REQ-USB-015 에서 함수 이름을 빼고 동작으로 적었다.
  - 형제 SPEC 과의 관계(D6)를 운영자 결정으로 확정하고 그 SPEC 을 함께 개정했다(§B.3).
- 2026-09-11 — v0.3.0, 운영자 답변(리드 경유)을 반영했다.
  - **D5 확정:** 정교화 규칙을 채택했다. 규칙을 REQ-USB-005 에 덧붙이고 AC-USB-016 을 채택했다(§B.4).
  - **F-05 보완 확정:** A1 범위 안의 보완으로 확정했다(§B.1).
  - **F-17:** 알려진 한계로 수용했다(§B.5, §E).
- 2026-09-11 — v0.4.0, plan-audit 2회차(FAIL 0.75, `.moai/reports/t656/plan-audit-iter2.md`)를 반영했다. 운영자가 한 번에 한해 연장을 승인한 최종 수정이다.
  - REQ-USB-005 를 GEARS 문장 셋으로 다시 썼다(N-01).
  - 남은 대기본의 판정 위치를 "흐름의 어느 단계도 살아 있는 settings.json 을 지우거나 다시 쓰기 전"으로 못박았다. 이전 배치("다음 흐름이 새 대기본을 기록하기 전")는 이것으로 대체된다(N-02, 운영자 지시).
  - 정상 종료 판정 신호를 "병합의 보존 경로"로 정의했다(N-10).
  - 중단 뒤 끼어든 쓰기의 결과를 알려진 한계로 기록했다(N-08).
  - 두 시점 판정 설계를 "운영자 규칙의 구현 방식, 리드 수용"으로 기록했다(§B.4).
- 2026-09-12 — sync-phase close (`status: in-progress → implemented → completed`, 카드 t656). Run-phase 16개 AC 전부 PASS(29/29 뮤턴트 킬), evidence: `progress.md` §E.2/§E.3, `.moai/reports/t656/run/`. `sync_commit_sha` 는 이 커밋 자신을 가리킬 수 없어 `pending-backfill` 로 기록하고, 후속 커밋에서 백필한다(스키마 D3 예외).

## §A 배경

### §A.1 지금의 병합 base 는 어떻게 만들어지는가

`moai update` 는 템플릿을 배포해 사용자의 `.claude/settings.json` 을 덮어쓴 뒤, 배포 전에 메모리에 떠 둔 사용자 파일을 3-way 병합으로 되돌려 넣는다(`internal/cli/update/merge/merge.go:173-258`). 3-way 병합에는 세 입력이 필요하다. 사용자 파일(current)과 방금 배포된 템플릿(updated)은 손에 있지만, 사용자 파일이 원래 배포받았던 템플릿 내용(base)은 어디에도 저장돼 있지 않다.

그래서 base 를 **유도**한다. 방금 배포된 템플릿을 사용자 파일이 가진 키로만 좁힌 것이 base 가 된다(`internal/cli/update/merge/base.go:112-131`). 좁힌 base 에 남는 키의 값은 언제나 updated 에서 온다(`base.go:128`).

병합 규칙은 공용 엔진이 정한다(`internal/merge/strategies.go:360-467`). 양쪽이 모두 가진 키는 다음 넷 중 하나로 갈린다(`strategies.go:418-462`).

| base 와 비교한 결과 | 쓰이는 값 |
|---|---|
| 아무도 바꾸지 않음 | base 값 |
| 사용자만 바꿈 | 사용자 값 |
| 템플릿만 바꿈 | 템플릿 값 |
| 둘 다 다르게 바꿈 | 충돌을 기록하고 사용자 값 |

배열은 원소 단위가 아니라 값 하나로 통째 비교된다.

### §A.2 왜 값 변경이 도착하지 않는가

유도된 base 에서는 공유 키마다 base 값과 updated 값이 같다. 따라서 "템플릿만 바꿈"과 "둘 다 바꿈" 두 갈래는 공유 leaf 에서 결코 실행되지 않는다. 사용자가 건드리지 않은 키라도 사용자 파일의 값이 새 템플릿 값과 다르기만 하면 "사용자만 바꿈"으로 읽혀 사용자 쪽 값이 남는다.

- 템플릿이 **새로 추가한** 키는 기존 설치에 도착한다. 사용자 파일에 없는 키는 base 에서 빠지므로 템플릿이 추가한 것으로 읽힌다.
- 사용자가 이미 가진 키의 **값을 템플릿이 바꾸면** 그 변경은 도착하지 않는다. 사용자가 그 키를 한 번도 편집한 적이 없어도 마찬가지다.

`base.go:29-34` 의 주석은 이 한계와 함께 해법도 적어 두었다. 배포 시점의 템플릿 내용을 스냅숏으로 남겨 다음 update 에서 진짜 base 로 써야 한다는 것이다. 같은 현상을 형제 SPEC `SPEC-UPDATE-MERGE-CONFLICT-BLIND-001` §A.6 이 다른 트리에서 측정했다. 여기서는 그 측정을 방향을 잡는 근거로만 인용하며, 이 트리에 대한 기준선은 run 단계 M1 에서 새로 잰다(acceptance.md §C).

### §A.3 영향 범위

`.claude/settings.json.tmpl` 은 기계마다 달라지는 값을 렌더한다(`SmartPATH`, `Platform`, `GitMode`, `HookOptIn` — 템플릿 안 자리표시자 줄 18개). 권한 목록, 훅 배선, `statusLine`, `env.PATH` 가 모두 이 파일에 있다. 템플릿이 이런 값을 고쳐도 기존 설치에는 반영되지 않으며, 새로 `moai init` 한 설치에서만 올바른 값을 받는다.

배포 뒤 파일별 병합을 쓰는 곳은 두 군데다. 일반 update(`internal/cli/update_template_sync.go:544`)와 clean-reinstall(`internal/cli/update_clean_install.go:507`)이다.

### §A.4 배포가 settings.json 을 쓰지 않는 경우, 그리고 흐름 안의 첫 재기록

init 은 강제 모드가 아닌 배포기를 쓴다(`internal/cli/init.go:821`). 이 배포기는 이미 있는 파일 가운데 매니페스트 기록이 없거나 사용자 소유로 기록된 것을 건너뛴다(`internal/template/deployer.go:239-255`). 따라서 `.claude/settings.json` 이 이미 있는 디렉터리에서 init 을 돌리면, 배포가 끝난 뒤 디스크에 있는 파일은 렌더가 아니라 사용자 파일이다. 그 파일을 base 로 삼으면 사용자가 고친 값이 모두 "아무도 바꾸지 않음"으로 읽혀 다음 update 에서 템플릿 값으로 덮인다. REQ-USB-016 이 이 경로를 막는다.

update 경로는 사정이 다르다. 관리 경로 정리 단계가 살아 있는 `.claude/settings.json` 을 지우고(`internal/cli/update_template_sync.go:329-337`), 강제 모드 배포가 렌더를 새로 쓴다(`:364`). 정리 직전에는 사용자 파일이 실행 단위 백업 디렉터리의 `in-memory-backups/` 아래로 복사된다(`internal/cli/update_disk_backup.go:47-53`). 그래서 배포 뒤 병합 전에 update 가 중단되면 살아 있는 파일은 순수 렌더다.

`moai update` 명령 안에서 살아 있는 `.claude/settings.json` 을 가장 먼저 다시 쓸 수 있는 단계는 폐기 v2 deny 규칙 제거(`internal/cli/update.go:384`)다. 이 단계는 `--binary`·`--dry-run` 조기 반환(`update.go:330`, `:335-365`) 뒤, v2 분기(`:405` → clean-reinstall `:420`)와 템플릿 동기화(`:489`)보다 앞에 있다. 버전 일치로 동기화를 건너뛰는 두 지점(`update_template_sync.go:98-105`, `:641`)보다도 앞이다. init 에서 가장 먼저 쓰는 단계는 초기화 실행(`init.go:867`)이다.

### §A.5 선행 SPEC 과 이 SPEC 의 경계

`SPEC-UPDATE-TEMPLATE-BASE-SNAPSHOT-001`(완료)은 `.moai/config/sections/*.yaml` 에 대해 같은 종류의 base 출처 결함을 닫았고, 그 스냅숏을 `.moai/cache/template-snapshot/sections/` 에 둔다(`internal/cli/update/backup/snapshot.go:22`). 그 스냅숏은 **복원(병합)이 끝난 뒤** 디스크 상태를 복사한다. 이 SPEC 은 그 규칙을 settings.json 에 복제하지 않는다. 캐시 루트만 함께 쓰고, 하위 경로와 기록 시점은 따로 정한다(§B C1). sections 쪽 규칙과 동작은 바뀌지 않는다.

## §B 운영자 결정

### §B.1 확정 결정 — A1 · B1 · C1 과 A1 보완

- **A1 — base 의 출처.** 배포가 `.claude/settings.json` 을 렌더해 쓴 **직후, 어떤 병합보다 먼저** 그 바이트를 스냅숏으로 남긴다. 다음 update 는 그것을 settings.json 병합의 base 로 쓴다.
  - 스냅숏이 없거나, 읽히지 않거나, 유효한 JSON 이 아니면 지금의 유도 base 로 돌아가며, 이 폴백은 update 를 막지 않는다.
  - 스냅숏 기록이 실패해도 init 과 update 는 실패하지 않는다.
  - 기록 지점은 `moai init`, 일반 `moai update`, clean-reinstall 세 곳이다.
  - 병합 뒤의 수정(폐기된 v2 deny 규칙 제거와 그 뒤의 모든 settings.json 재기록)은 base 에 섞이지 않는다.
- **A1 보완 — 배포가 실제로 쓴 경우에만 기록 (F-05, 운영자 확정 2026-09-11).** A1 범위 안의 보완이다. 스냅숏은 그 흐름의 배포가 `.claude/settings.json` 을 실제로 썼을 때만 기록한다. 배포가 기존 파일을 보호하려고 건너뛴 흐름에서는 기록하지 않으며, 디스크의 현재 파일을 다시 읽어 기록하지도 않는다(REQ-USB-016).
- **B1 — 둘 다 바꾼 경우.** 공용 엔진의 현재 규칙(사용자 값 유지 + 충돌 보고)을 그대로 두며, `internal/merge` 는 수정하지 않는다. 배열이 통째로 비교된다는 점은 알려진 한계로 기록한다(§E).
- **C1 — sections 스냅숏과의 관계.** 캐시 루트 `.moai/cache/template-snapshot/` 만 공유하고 settings.json 은 별도 하위 경로에 둔다. sections 스냅숏의 규칙과 동작은 바꾸지 않는다.

**하위 경로 선택 — `claude/settings.json`.** 근거는 세 가지다.

1. **대칭.** `sections/` 가 원래 위치 `.moai/config/sections/` 의 마지막 구간을 따르듯, `claude/settings.json` 도 원래 위치 `.claude/settings.json` 을 따른다. `sections/` 와 겹치지 않는다.
2. **점 없는 디렉터리.** Claude Code 는 하위 디렉터리의 `.claude/` 도 탐색한다(이 세션에서 `.claude/worktrees/t656/` 의 스킬이 따로 나열된 것이 그 관찰이다). 캐시 안에 `.claude` 디렉터리가 생기지 않게 한다. 설정 파일까지 탐색하는지는 확인하지 않았다(§H Gaps).
3. **정리 반경 밖, git 추적 밖.** `.moai/cache/` 는 관리 경로 일괄 삭제 대상이 아니다. 저장소 `.gitignore:352` 와 배포 템플릿 `.gitignore:241` 모두 이 경로를 무시한다. 렌더된 settings.json 에는 기계별 `PATH` 가 들어가므로 커밋되면 안 된다.

### §B.2 스냅숏의 두 상태 — 대기본과 확정본

A1 이 요구하는 두 시점("배포 직후 기록"과 "다음 update 의 base")을 한 파일로 겸하면, 같은 흐름에서 배포 직후 쓴 렌더가 그 흐름의 병합 base 를 덮는다. base 와 updated 가 같아지면 사용자 파일에 없는 템플릿 키가 모두 사용자가 지운 키로 읽혀 새 키 전달이 조용히 깨진다. 그래서 스냅숏을 두 상태로 나눈다.

| 상태 | 경로 (캐시 루트 기준) | 언제 생기는가 | 누가 읽는가 |
|---|---|---|---|
| **대기본** (staging) | `claude/settings.json.pending` | 이 흐름의 배포가 `.claude/settings.json` 에 렌더를 쓴 직후 | 승격 판단만 읽는다. 병합은 읽지 않는다 |
| **확정본** (canonical base) | `claude/settings.json` | 대기본이 **승격**될 때 | 다음 흐름의 settings.json 병합이 base 로 읽는다 |

승격은 대기본을 확정본 자리로 옮기는 행위다. 한 흐름의 병합 단계가 끝나기 전에는 일어나지 않으며, 승격 조건은 §B.4 규칙을 따른다(REQ-USB-005). 확정본의 경로와 형식은 v0.1.0 과 같으므로, 병합이 base 를 찾는 방식은 바뀌지 않는다.

### §B.3 형제 SPEC 과의 관계 — D6 운영자 결정 (2026-09-11)

`SPEC-UPDATE-MERGE-CONFLICT-BLIND-001`(카드 t576, in-progress)의 REQ-UMC-010 은 원래 "the merge subsystem shall continue to resolve a shared key in favour of the user's value" 로 시작해, 이 SPEC 의 REQ-USB-006 과 글자 그대로 부딪쳤다. 운영자 결정은 두 가지다.

- **해석.** REQ-UMC-010 은 그 SPEC 의 REQ-UMC-008/009 개선에만 걸린 한정이다. 그 SPEC 의 REQ-UMC-010 과 제외 항목을 같은 방향으로 고쳐 쓰는 개정을 이 SPEC 과 함께 적용했다(그 SPEC 의 HISTORY 2026-09-11 항목).
- **착지 순서.** 이 SPEC(t656)이 먼저 착지한다. t576 의 M2.1 이 재개될 때는 그 SPEC §A.6 의 전제를 다시 잰다. 이 SPEC 이 반영되면 settings.json 에서 "템플릿만 바꿈" 갈래에 도달할 수 있기 때문이다.

### §B.4 D5 — 승격 조건 (운영자 결정 2026-09-11)

**규칙.** 대기본은 흐름이 끝났을 때 살아 있는 `.claude/settings.json` 이 그 흐름의 렌더를 반영하고 있을 때만 확정본으로 승격한다. 반영하고 있지 않으면 이전 확정본을 그대로 둔다(REQ-USB-005).

| 경우 | 흐름 끝의 살아 있는 파일 | 결과 |
|---|---|---|
| 병합이 병합 결과를 씀 | 렌더를 반영한 병합 결과 | 승격 |
| 병합이 실패해 사용자 파일을 통째로 보존 | 사용자 파일 | 이전 확정본 유지 |
| 배포 뒤 병합 전에 중단 | 순수 렌더(§A.4) | 승격 |
| 중단 뒤 복원이 사용자 파일을 되돌림 | 사용자 파일 | 이전 확정본 유지 |
| init 이 파일을 실제로 씀 | 렌더(이후 자율성 번들이 고쳤을 수 있음) | 승격 |

**판정 방식 — 두 시점.**

- **정상적으로 끝난 흐름은 그 흐름 끝에서 판정한다.** 신호는 바이트 비교가 아니라 병합이 **보존 경로**를 탔는지다. 보존 경로란 흐름 이전 사용자 파일을 통째로 되돌려 쓰는 분기로, `merge.go:197-204`(배포 파일을 읽지 못함), `:217-225`(base 를 만들 수 없음), `:229-237`(병합 실패)이다. 사용자 파일이 렌더와 같아 병합을 건너뛴 경우(`:207`)는 보존 경로가 아니다. init 처럼 병합이 없는 흐름도 보존 경로를 타지 않은 것으로 본다.
- **승격 판정 전에 중단된 흐름은 다음 흐름에서 판정한다.** 중단된 흐름은 대기본을 남긴다. 다음 흐름은 **그 흐름의 어느 단계도 살아 있는 `.claude/settings.json` 을 지우거나 다시 쓰기 전에** 살아 있는 파일을 남은 대기본과 비교한다. 바이트 동일하면 승격하고, 다르면 버린다.
  - `moai update` 에서는 `--binary`·`--dry-run` 조기 반환 뒤, 폐기 v2 deny 규칙 제거(`update.go:384`) 앞이다.
  - `moai init` 에서는 초기화 실행(`init.go:867`) 앞이다.
  - 이 위치는 clean-reinstall 분기와 두 버전 일치 건너뛰기(§A.4)보다 앞이므로 한 곳이 모든 update 흐름을 덮는다. 배포 뒤나 백업 단계에서 판정하면 새 렌더나 이미 고쳐진 파일과 비교하게 된다.

**결정 기록.**

- 처음 고른 1회차 권고안 (a)의 목적은 새 템플릿 키가 사용자의 삭제로 잘못 읽히는 일을 절대 만들지 않는 것이다. 이 목적은 그대로 지켜진다. 병합이 사용자 파일을 통째로 보존한 흐름에서 렌더를 승격하지 않으므로, 그 렌더가 새로 들인 키는 다음 update 에서 사용자 삭제로 읽히지 않는다.
- "배포 뒤 병합 전에 중단되면 이전 스냅숏을 유지한다"는 이전 설명은 틀린 전제에 기대고 있었다. 중단 시 살아 있는 파일을 사용자 파일로 보았지만, 실제로는 순수 렌더다. 그 설명은 이 규칙으로 대체되었다(운영자 결정 2026-09-11).
- **운영자 규칙의 구현 방식, 리드 수용 (2026-09-11).** 위의 두 시점 판정 설계는 운영자 규칙을 구현하는 방식이며, 리드가 수용했다. 중단된 흐름을 중단 시점에 판정하면 뒤이은 복원을 볼 수 없어, 운영자의 넷째와 다섯째 경우를 함께 성립시킬 수 없기 때문이다. 남은 대기본의 판정 위치는 plan-audit 2회차 N-02 에 따른 배치(흐름의 어느 단계도 살아 있는 파일을 지우거나 다시 쓰기 전)이며, 이전 배치("다음 흐름이 새 대기본을 기록하기 전")를 대체한다(운영자 지시 2026-09-11).
- **복원 명령에 대한 판독.** `moai update --restore` 는 `RestoreFromBackupDir` 를 거쳐 `RestoreMoaiConfig` 만 두 번 부른다(`internal/cli/update/backup/restore_entry.go:47-79`). 감사 판독에 따르면 이 경로는 `.moai/config` 만 쓴다. 따라서 "중단 뒤 복원이 사용자 파일을 되돌림"은 사용자가 파일을 손으로 되돌린 경우에만 생길 수 있다. 실행으로 확인하지는 않았으며, run 단계 M1 의 검증 항목이다(plan.md Decision D5).

### §B.5 사용자가 지운 템플릿 키

사용자가 지운 템플릿 키는 이후 템플릿 수정을 받지 않는다 — 운영자 확인 2026-09-11

## §C 요구사항 (GEARS)

**REQ-USB-001** — **When** `moai init`, 일반 `moai update`, clean-reinstall 가운데 어느 흐름의 배포가 `.claude/settings.json` 에 렌더 바이트를 실제로 쓰면, the update subsystem shall 그 흐름이 이 파일을 다른 방식으로 고치기 전에 그 렌더 바이트를 settings.json 대기본으로 기록한다.

**REQ-USB-002** — The settings.json 스냅숏 shall 사용자 프로젝트의 git 무시 경로인 캐시 루트 `.moai/cache/template-snapshot/` 아래에 놓이며, 확정본은 `claude/settings.json`, 대기본은 `claude/settings.json.pending` 을 쓰고, sections 스냅숏의 `sections/` 하위 경로와 겹치지 않는다.

**REQ-USB-003** — The settings.json 대기본 shall not 렌더 이후에 이루어진 수정을 담는다. 여기에는 settings.json 3-way 병합 결과, 폐기된 v2 deny 규칙 제거, init 의 자율성 단계(autonomy tier) 권한 번들 적용, 그 밖의 모든 후속 재기록이 포함된다.

**REQ-USB-004** — **Where** settings.json 확정본이 존재하고 읽히며 JSON 객체로 해석되면, the update subsystem shall 그 확정본을 `.claude/settings.json` 3-way 병합의 base 로 쓴다.

**REQ-USB-005** — The update subsystem shall 확정본을 대기본의 승격으로만 바꾸며, 한 흐름의 대기본을 그 흐름의 settings.json 병합 단계가 끝나기 전에 승격하지 않는다. **When** 한 흐름이 정상적으로 끝나면, the update subsystem shall 그 흐름의 settings.json 병합이 흐름 이전의 사용자 파일을 통째로 되돌려 쓰는 보존 경로를 타지 않았을 때 그 흐름의 대기본을 승격하고, 보존 경로를 탔으면 그 대기본을 버린다. **When** 한 흐름이 이전 흐름이 남긴 대기본을 발견하면, the update subsystem shall 그 흐름의 어느 단계도 살아 있는 `.claude/settings.json` 을 지우거나 다시 쓰기 전에, 살아 있는 파일이 남은 대기본과 바이트 동일하면 남은 대기본을 승격하고 다르면 버린다.

**REQ-USB-006** — **Where** 확정본 base 가 쓰이고 **When** 사용자가 바꾸지 않은 공유 leaf 의 값이 확정본과 새 렌더 사이에서 달라졌으면, the merge shall 새 렌더의 값을 기록한다.

**REQ-USB-007** — **Where** 확정본 base 가 쓰이고 **When** 사용자가 공유 leaf 의 값을 바꿨고 템플릿은 그 값을 바꾸지 않았으면, the merge shall 사용자 값을 기록한다.

**REQ-USB-008** — **Where** 확정본 base 가 쓰이고 **When** 사용자와 템플릿이 같은 공유 leaf 를 서로 다른 값으로 바꿨으면, the merge shall 사용자 값을 기록하고 그 파일의 병합 결과를 충돌로 보고한다.

**REQ-USB-009** — **When** settings.json 확정본이 없거나, 읽을 수 없거나, JSON 으로 해석되지 않거나, JSON 객체가 아니면, the update subsystem shall 지금의 유도 base 로 settings.json 을 병합하고, 그 결과는 이 SPEC 이전과 같으며, update 를 실패시키지 않는다.

**REQ-USB-010** — **When** 대기본 기록이나 승격이 실패하면, the update subsystem shall sections 스냅숏의 경고와 구별되는 경고를 출력하고, 그 흐름을 부른 `moai init` 또는 `moai update` 의 결과는 바꾸지 않는다.

**REQ-USB-011** — The update subsystem shall 사용자 파일에도 확정본에도 없고 새 렌더에만 있는 키를 병합 결과에 추가한다. 이 동작은 확정본이 있든 없든 유지된다.

**REQ-USB-012** — **Where** 확정본 base 가 쓰이고 **When** 확정본과 새 렌더에는 있는 키가 사용자 파일에서 빠져 있으면, the merge shall 그 키를 사용자의 삭제로 취급해 결과에 넣지 않는다.

**REQ-USB-013** — The sections 스냅숏 shall 위치, 기록 시점, 내용, 읽는 경로를 이 SPEC 이전 그대로 유지한다.

**REQ-USB-014** — The 공용 병합 엔진 `internal/merge` shall 이 SPEC 이전의 규칙과 코드를 그대로 유지한다.

**REQ-USB-015** — The 확정본 base 선택 shall `.claude/settings.json` 에만 적용되며, 배포 뒤 파일별 병합이 다루는 다른 파일(`.mcp.json`, `.moai/status_line.sh`)의 base 는 바뀌지 않는다.

**REQ-USB-016** — **When** 배포가 이미 있는 `.claude/settings.json` 을 보호하려고 그 파일 쓰기를 건너뛰면, the update subsystem shall not 그 흐름에서 대기본을 기록하며, 대기본 바이트를 디스크에 있는 현재 파일을 다시 읽어 만들지 않는다.

## §D 비기능 제약

- **NFR-USB-001 (교차 플랫폼)** — 경로는 모두 `filepath.Join` 으로 조립한다. 승격은 한 파일 시스템 안의 교체로 하며, `GOOS=windows GOARCH=amd64 go build ./...` 로 확인한다.
- **NFR-USB-002 (테스트 격리)** — 모든 테스트는 `t.TempDir()` 안에서만 쓴다. 홈 디렉터리가 필요하면 `userHomeDirFn` 시접(`internal/cli/homedir.go`)으로 주입하며, `t.Setenv("HOME", ...)` 은 쓰지 않는다. 실제 홈이나 실제 `moai update` 실행에 닿지 않는다. 시접을 바꾸는 테스트는 `t.Parallel()` 을 부르지 않는다.
- **NFR-USB-003 (새 의존성 없음)** — 표준 라이브러리와 기존 모듈만 쓴다.
- **NFR-USB-004 (템플릿 중립성)** — `internal/template/templates/` 아래 파일은 추가도 수정도 하지 않는다.
- **NFR-USB-005 (서브에이전트 경계)** — 이 변경은 사용자 질문 채널을 부르지 않는다.
- **NFR-USB-006 (조용한 퇴행 금지)** — 확정본이 없는 설치의 settings.json 병합 결과는 이 SPEC 이전과 바이트 수준으로 같다.

## §E 알려진 한계

- **첫 사이클.** 확정본이 아직 없는 설치는 이 SPEC 이 반영된 뒤 첫 `moai update` 에서도 값 변경을 받지 못한다. 그 update 는 폴백 경로(REQ-USB-009)로 병합하고, 그 흐름의 대기본이 승격된 뒤부터 base 가 생긴다.
- **배열은 통째로 비교된다 (B1).** 사용자가 `permissions.allow` 에 한 줄만 더해도, 그 배열에 대한 템플릿의 이후 변경은 전부 "둘 다 바꿈"이 되어 도착하지 않는다. 사용자 배열이 남고 충돌만 보고된다.
- **템플릿이 지운 키는 지워지지 않는다 (B1).** 확정본과 사용자 파일에는 있고 새 렌더에는 없는 키는 공용 엔진 규칙에 따라 사용자 쪽 값으로 남는다.
- **사용자가 지운 템플릿 키는 이후 템플릿 수정을 받지 않는다 (A1+B1, 운영자 확인 2026-09-11).** 지금은 사용자가 지운 템플릿 키가 update 마다 다시 추가된다. 확정본 base 에서는 같은 상황이 사용자의 삭제로 읽혀 결과에 넣지 않으며, 그 키에 대한 이후 템플릿 수정도 도착하지 않는다(REQ-USB-012, §B.5).
- **중단 뒤 끼어든 쓰기는 승격을 폐기로 바꾼다 (N-08, 안전한 쪽으로 벗어남).** 배포 뒤 병합 전에 중단된 흐름이 있고, 다음 흐름이 시작하기 전에 살아 있는 `.claude/settings.json` 에 어떤 쓰기든 일어나면 남은 대기본은 승격되지 않고 버려진다. 손 편집, Claude Code 의 프로젝트 설정 기록, `moai tool-policy build`, 자율성 번들이 그런 쓰기다. 운영자 규칙 문구(중단 시 살아 있는 파일은 순수 렌더이므로 승격)와 다른 결과다. 그러나 사용자 데이터는 잃지 않는다. 중단된 렌더의 템플릿 변경이 사용자 변경으로 읽혀, 같은 leaf 에 대한 이후 템플릿 변경이 충돌로 남을 뿐이다.
- **기계가 바뀌면 렌더 값 변화가 템플릿 변경으로 읽힌다.** `env.PATH` 처럼 기계마다 달라지는 값은 사용자가 손대지 않았다면 새 렌더 값으로 바뀐다. 이것은 의도한 동작이다. 사용자가 직접 고친 값이면 충돌이 보고되고 사용자 값이 남는다.
- **세 기록 지점 밖의 재기록.** `ApplyAutonomyTierBundle` 이나 권한 정책 생성처럼 흐름 밖에서 프로젝트 settings.json 을 다시 쓰는 작성자는 사용자 변경으로 올바르게 읽힌다. 다만 그런 작성자가 **지운** 키는 사용자의 삭제로 읽혀 복구되지 않는다.

## §F Exclusions

### Out of Scope — sections 스냅숏 규칙

- `.moai/config/sections/*.yaml` 스냅숏의 위치, 기록 시점, 내용, 읽는 경로는 이 SPEC 에서 바꾸지 않는다.
- sections 스냅숏이 복원 뒤 상태를 담는 탓에 두 번째 update 에서 사용자 편집이 덮일 수 있다는 결함 의심은 별도 카드에서 다룬다.

### Out of Scope — 공용 병합 엔진

- `internal/merge` 의 병합 규칙, 충돌 판정, 출력 형식은 수정하지 않는다(B1).

### Out of Scope — 배열 원소 단위 병합

- 배열을 원소 단위로 3-way 병합하는 방식(추가·삭제만 반영)은 도입하지 않는다.

### Out of Scope — 템플릿 키 강제 적용

- 지정한 키에 템플릿 값을 강제로 덮어쓰는 관리 키 목록은 만들지 않는다.

### Out of Scope — 다른 병합 대상 파일

- `.mcp.json` 과 `.moai/status_line.sh` 의 base 는 바꾸지 않는다. `.mcp.json` 에도 같은 종류의 한계가 있지만 이 SPEC 의 대상이 아니다.

### Out of Scope — 과거 스냅숏 재구성

- 확정본이 없는 기존 설치를 위해 git 이력이나 매니페스트 해시로 과거 렌더를 되살리지 않는다. 첫 사이클 한계(§E)를 그대로 받아들인다.

### Out of Scope — 충돌 신호 계측

- 충돌을 알리는 사용자 출력의 모양, 보안 관련 키 신호, `HasConflict` 도달성 계측은 `SPEC-UPDATE-MERGE-CONFLICT-BLIND-001` 의 소관이다. 이 SPEC 은 기존 충돌 보고 경로를 그대로 쓴다.

## §G 수용 기준 교차 참조

요구사항과 수용 기준의 대응표는 `acceptance.md` §B 에 있다. 수용 기준은 모두 Go 테스트이거나 읽기 전용 명령이다. 이 SPEC 이 새로 요구하는 동작에는 이 SPEC 이전 트리에서 실패하는 RED 단계가 있다.

## §H 교차 참조

- `internal/cli/update/merge/base.go:29-34` — 한계를 스스로 적은 주석. `:53-60` 은 `templateManaged`, `:112-131` 은 `pruneToShared`
- `internal/cli/update/merge/merge.go:173-258` — 배포 뒤 파일별 병합(보존 경로 `:197-204`, `:217-225`, `:229-237`, 동일 건너뛰기 `:207`, 충돌 출력 `:245-247`)
- `internal/merge/strategies.go:360-467` — 공용 엔진의 맵 병합(공유 키 네 갈래 `:418-462`)
- `internal/template/deployer.go:239-255` — 기존 파일 보호 건너뛰기
- `internal/cli/update.go:330`, `:335-365` — `--binary`·`--dry-run` 조기 반환. `:384` 폐기 deny 규칙 제거, `:405`·`:420` v2 분기와 clean-reinstall, `:489` 템플릿 동기화
- `internal/cli/update_template_sync.go:98-105`, `:641` — 버전 일치 건너뛰기. `:329-337` 정리, `:364` 배포, `:491` 백업 단계 읽기, `:495-547` Restore Settings
- `internal/cli/update_clean_install.go:400` — 백업 읽기, `:459` 배포, `:507` 병합, `:531` 폐기 deny 규칙 제거
- `internal/cli/init.go:821` — init 의 비강제 배포기, `:867` 초기화 실행, `:889-900` 자율성 단계 권한 번들
- `internal/cli/update_disk_backup.go:47-53` — 정리 직전 사용자 파일의 디스크 백업 대상
- `internal/cli/update/backup/restore_entry.go:47-79` — `RestoreFromBackupDir` 는 `RestoreMoaiConfig` 만 부름
- `internal/cli/update/backup/snapshot.go:22` — sections 스냅숏 경로 상수, `:114` `HasSnapshot`
- `.moai/specs/SPEC-UPDATE-MERGE-CONFLICT-BLIND-001/spec.md` REQ-UMC-010 — 2026-09-11 개정(§B.3)
- `.moai/reports/t656/design-options.md`, `plan-audit-iter1.md`, `plan-audit-iter2.md`
- **Gaps**
  - 줄 인용은 트리 `04a8ab731` 에서 읽었다. `81c1d58f9` 이후 `internal/` 변경이 없음을 `git diff --stat` 로 확인했다. 코드 동작을 실행해 확인한 것은 없다.
  - Claude Code 가 하위 디렉터리의 `.claude/settings.json` 을 설정으로 읽는지는 확인하지 않았다.
  - `RestoreMoaiConfig` 가 `.claude/settings.json` 을 쓰지 않는다는 판단은 감사 판독에 기댄 것이다. 직접 읽은 것은 `restore_entry.go:47-79` 까지다(run 단계 M1 검증 항목).
  - `update.go:139-379` 사이에 프로젝트 settings.json 을 쓰는 다른 단계가 없다는 판단은 `update.go` 의 `settings.json` 토큰 grep(주석 두 줄만 적중)에 기댄 것이며, 호출되는 함수의 내부까지 따라가지는 않았다.
  - init 호출 지점까지 `DeployResult.ProtectedSkips` 가 전달되는지는 확인하지 않았다. `initializer.go:423-430` 은 거울 알림만 경고로 넘긴다.
