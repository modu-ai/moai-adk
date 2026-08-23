---
id: SPEC-CLI-CLEAN-SYMLINK-001
title: "Acceptance — moai update 청소 경로의 심볼릭 링크 인식"
version: "0.1.1"
status: in-progress
created: 2026-08-22
updated: 2026-08-22
author: manager-spec
priority: P1
phase: "v3.1.3 target"
module: internal/cli/update/deploy
lifecycle: spec-anchored
tier: M
era: V3R6
tags: "update, clean, symlink, deploy, backup, t173"
---

# acceptance.md — SPEC-CLI-CLEAN-SYMLINK-001

> 수치(3면 일치): **픽스처 형태 5종 · AC 11건(AC-CSL-001…011)** — REQ 12건은 각 AC가
> §D.2 추적표에서 커버한다. 모든 링크 단언은 ≥2 관측 축을 결합한다(REQ-CSL-011).
> 단언의 grep 기준 토큰: 링크 경로 + "symlink" 계열 토큰(plan.md §D-7).

## §A. 픽스처 형태 5종 (형태 일치 — REQ-CSL-010 [HARD])

| 형태 | 픽스처 구성 | 대상 AC |
|---|---|---|
| FX-1 라이브 디렉터리 링크 | 관리 뿌리(예: `.claude/rules/moai`, 글로브 배치는 `.claude/skills/moai-livelink`)를 외부 **실재 디렉터리**(센티널 파일 포함)를 가리키는 **디렉터리 링크**로 교체 | AC-CSL-004 |
| FX-2 라이브 파일 링크 | `.claude/settings.json`을 외부 **실재 파일**(내용 센티널)을 가리키는 **파일 링크**로 교체 — 디렉터리 링크로 대체 금지 | AC-CSL-005 |
| FX-3 dangling 링크 | 링크 자체는 존재, 대상 부재 — 3배치: 비-글롭 뿌리 / 글로브 매치 이름 / `.moai/config` 뿌리 | AC-CSL-001/002/003 |
| FX-4 실디렉터리 대조 | 링크 없는 관리 뿌리 + 비관리 파일 1개 이상 (must-not-flag 극) | AC-CSL-006 |
| FX-5 hns 사용자 소유 대조 | `.claude/skills/hns-mine/`(SKILL.md + 내부 dangling 링크) (must-not-flag 극) | AC-CSL-007 |

## §B. 공통 Given 관례

- 모든 테스트는 `t.TempDir()` 하위 임시 프로젝트에서 `moai init` 상태를 재현하거나 청소
  단계가 기대하는 최소 디렉터리 구조를 직접 구성한다. 프로젝트 루트·`~` 미접촉.
- "진행줄" 판정은 청소 단계가 기록하는 출력(io.Writer)에 대해 grep 한다.
- 심볼릭 링크 생성은 헬퍼 경유 — `os.Symlink` 실패 시 `t.Skip`(REQ-CSL-012).

## §C. 심각도 척도

P0 = 출하 결함(사용자 update 불능·데이터 위험) · P1 = 계약·관측성 · P2 = 비회귀 보장.

## §D. AC 매트릭스 (Given-When-Then)

### AC-CSL-001 [P0] — dangling 링크 at 비-글롭 뿌리 (Run D 재현 — M1 RED 기준)

**Given** 임시 프로젝트의 `.claude/agents/moai` 실제 디렉터리를 지우고, 부재 경로를 가리키는
dangling 심볼릭 링크로 교체했다(링크 존재, `os.Stat` 추적 시 대상 부재).
**When** update의 청소 단계가 해당 루트를 처리하고 배포까지 진행한다.
**Then**
1. 링크가 제거되어 있다(해당 경로 `os.Lstat` ENOENT).
2. 진행 출력에 해당 경로와 dangling 심볼릭 링크임을 이름붙인 줄이 존재한다(경로+symlink 토큰).
3. update가 `mkdir ... file exists` (EEXIST)로 중단하지 않는다(청소+배포 완결).
4. `.claude/agents/moai`는 실제 디렉터리로 재배포되어 있다(템플릿 파일 ≥1).
5. 청소·배포를 즉시 재실행해도 동일하게 성공한다(재실행 루프 폐쇄 — 도시에 gap 2의
   실측 전환).

*RED 기준: 현행 코드에서 단언 1·3·5가 실패한다(단언 1·3은 Run D 실측; 단언 5의 재실행
루프는 Run D 재현 + 코드 추적 — 도시에 gap 2, M1 RED에서 직접 관측된다). 축 결합: 링크
존재 + 메시지 + 종결 + 재배포 + 재실행.*

### AC-CSL-002 [P0] — dangling 링크 at 글로브 매치 이름

**Given** `.claude/skills/moai-dangling-custom`(부재 경로 지목 dangling 링크)와 정상
템플릿 스킬 디렉터리들이 공존한다.
**When** 글로브 루트(`.claude/skills/moai*`)의 청소가 실행된다.
**Then**
1. dangling 링크가 제거되어 있다(종전에는 영구 잔존 — Run D 실측).
2. 해당 링크 경로를 이름붙인 진행줄이 존재한다.
3. 템플릿 스킬들은 정상 재배포되어 있다(스킬 파일 수 ≥1).

### AC-CSL-003 [P1] — dangling 링크 at `.moai/config` 뿌리 (도시에 gap 4 실측 전환)

**Given** `.moai/config`가 부재 경로를 가리키는 dangling 링크다(사전검사 없는 8번째 뿌리).
**When** 청소의 config 제거 단계와 배포가 실행된다.
**Then**
1. 링크가 제거되어 있다.
2. `.moai/config`가 실제 디렉터리로 재배포되어 있다(템플릿 config 파일 ≥1).
3. update가 EEXIST로 중단하지 않는다.
4. 진행 출력에 해당 경로와 dangling임을 이름붙인 줄이 존재한다(경로+symlink 토큰) —
   config 뿌리의 진행줄만 빼먹은 구현이 이 AC를 통과하지 못하게 닫는다(REQ-CSL-002/005).

### AC-CSL-004 [P1] — 라이브 디렉터리 링크 (FX-1, 비-글롭 + 글로브 양배치)

**Given** 관리 뿌리(비-글롭: `.claude/rules/moai`, 글로브: `.claude/skills/moai-livelink`)가
각각 외부 실재 디렉터리(내부에 센티널 파일)를 가리키는 라이브 **디렉터리** 링크다.
**When** update의 청소와 배포가 실행된다.
**Then**
1. 루트 경로는 실제 디렉터리로 재배포되어 있다(링크가 아님).
2. 외부 대상 디렉터리와 센티널 파일이 무결하다(내용 그대로).
3. 각 링크 경로를 이름붙인 진행줄이 존재한다(현행 코드에서 이 단언만 실패 — 무인식).
4. pre-clean 백업 트리에 해당 루트 경로의 파일이 없다(백업 0건 연속 — 축 결합의 보조축으로만).

*축 결합: 링크/실디렉터리 구분 + 대상 무사 + 메시지 (+백업). "백업 0건"을 단독으로 쓰지
않는다(WalkDir-스킵 공허함 — D4).*

### AC-CSL-005 [P1] — 라이브 파일 링크 (FX-2, Run B 준거)

**Given** `.claude/settings.json`이 외부 실재 파일(내용 센티널 `OUTSIDE-SETTINGS-v1` 계열)을
가리키는 라이브 **파일** 링크다.
**When** update의 청소와 배포가 실행된다.
**Then**
1. pre-clean 백업에 링크를 통과해 읽은 대상 바이트가 존재한다(센티널와 일치).
2. 최종 `.claude/settings.json`은 실제 파일이다(링크가 아님).
3. 링크 경로를 이름붙인 진행줄이 존재한다.
4. 최종 내용 흐름이 현행과 동일하다 — 백업 바이트가 사용자 내용으로 복원된다(3-way merge
   보존, Run B 실측 준거).

### AC-CSL-006 [P2] — 실디렉터리 대조 / must-NOT-flag 극 (FX-4)

**Given** 링크가 전혀 없는 관리 루트(예: `.claude/hooks/moai`)에 비관리 파일 1개 이상이 있다.
**When** update의 청소와 배포가 실행된다.
**Then**
1. 비관리 파일이 pre-clean 백업에 존재한다(파일 수 ≥1).
2. 진행 출력에 "symlink" 계열 토큰을 이름붙인 줄이 **없다**(링크 없는 실행에서 링크
   진행줄이 나오면 오탐).
3. 해당 루트는 정상 재배포되어 있다.

### AC-CSL-007 [P2] — hns 사용자 소유 미접촉 / must-NOT-flag 극 (FX-5, Run C 준거)

**Given** `.claude/skills/hns-mine/SKILL.md`(센티널 내용)과 그 내부의 dangling 링크
`badlink`가 있다.
**When** update 전 과정이 실행된다.
**Then**
1. `hns-mine/` 디렉터리와 SKILL.md 내용이 무결하다.
2. 내부 dangling 링크 `badlink`가 그대로 잔존한다(사용자 소유 링크는 청소 대상 아님 —
   FX-3b의 제거가 관리 네임스페이스에만 적용됨을 입증).
3. 백업 트리 전체에서 `hns-mine` 경로 검색 결과가 0건이다.

### AC-CSL-008 [P1] — 순서 독립 (미러-대상 연결, REQ-CSL-008)

**Given** 청소 대상 두 항목이 링크로 연결되어 있다 — 글로브 루트의 실재 디렉터리 A(비관리
파일 포함)와 A를 가리키는 라이브 디렉터리 링크 B(글로브 매치 이름).
**When** 두 항목을 (a) B→A 순서와 (b) A→B 순서로 각각 청소하고 배포까지 실행한다.
**Then** 두 순서의 최종 트리 상태가 동일하다 — A 재배포 상태, B 부재(템플릿 비보유 이름),
A의 비관리 파일 백업 존재가 순서와 무관하게 일치한다.

### AC-CSL-009 [P1] — 청소 집합 ↔ 배포 집합 교차 계약 (REQ-CSL-009)

**Given** 임베디드 템플릿 파일시스템과 `ManagedCleanTargets`(+ config 뿌리).
**When** 계약 검사를 실행한다.
**Then**
1. 모든 비-글롭 청소 루트는 **배포자가 렌더링 후 기록하는 경로**다 — "보유"의 판정 집합은
   렌더링 목적지(`.tmpl` 접미사 제거 포함)이며 원시 임베디드-FS 경로가 아니다(루트 1
   `.claude/settings.json`은 템플릿이 `settings.json.tmpl`로만 보유 — 원시 판독은 도시에
   §1.1(b) 실측으로 반증됨). 위반은 결함 — 청소가 재배포 없이 사용자 경로를 순삭하는 형태.
2. 모든 글로브 청소 패턴은 템플릿 경로 ≥1과 매치된다.
위반 시 테스트는 실패한다. t81(가) `.agents/` 배포 추가 후 release/v3.1.3 통합 시점에
같은 검사가 통과해야 한다(§D.7).

### AC-CSL-010 [P1] — 픽스처 형태 일치·비공허 자점검 (리뷰 AC — REQ-CSL-010/011)

**Given** M1~M3에서 완성된 테스트 파일.
**When** 리뷰(또는 plan-audit)가 픽스처와 단언을 검사한다.
**Then**
1. 모든 링크 픽스처의 형태가 시험 대상 AC의 제품 형태와 일치한다(FX-2는 파일 링크,
   FX-1은 디렉터리 링크 — §A 표와 대조).
2. 모든 링크 단언이 ≥2 관측 축을 결합한다 — bare "백업 파일 수 == 0" 단언은 0건.

### AC-CSL-011 [P2] — 플랫폼 스킵 (REQ-CSL-012)

**Given** `os.Symlink`가 실패하는 호스트(예: 권한 없는 Windows).
**When** 심볼릭 링크 테스트가 실행된다.
**Then** 테스트는 `t.Skip`으로 건너뛰고 실패하지 않는다(링크 생성 성공을 가정한 단언 없음).

## §D.1 심각도 요약

P0: AC-CSL-001, AC-CSL-002 (출하 결함 직접 수정) · P1: AC-CSL-003/004/005/008/009/010 ·
P2: AC-CSL-006/007/011.

## §D.2 추적표 (REQ ↔ AC)

| REQ | AC | | REQ | AC |
|---|---|---|---|---|
| REQ-CSL-001 (분류) | AC-CSL-001/002/003/004/005 (전 형태가 분기 진입을 입증) | | REQ-CSL-007 (사용자 소유) | AC-CSL-007 |
| REQ-CSL-002 (dangling) | AC-CSL-001/002/003 | | REQ-CSL-008 (순서 독립) | AC-CSL-008 |
| REQ-CSL-003 (라이브 디렉터리 링크) | AC-CSL-004 | | REQ-CSL-009 (교차 계약) | AC-CSL-009 |
| REQ-CSL-004 (라이브 파일 링크) | AC-CSL-005 | | REQ-CSL-010 (형태 일치) | AC-CSL-010 |
| REQ-CSL-005 (관측성) | AC-CSL-001/002/004/005 (진행줄 단언) + AC-CSL-006 (부재 단언) | | REQ-CSL-011 (비공허) | AC-CSL-010 |
| REQ-CSL-006 (실재 비회귀) | AC-CSL-006 | | REQ-CSL-012 (플랫폼) | AC-CSL-011 |

12 REQ 전부 최소 1개 AC에 연결된다.

## §D.3 경계 사례 (edge cases)

- **파일 뿌리 dangling**(settings.json이 dangling 링크): 배포는 `atomicWriteFile`의
  rename(deployer.go:28)이 목적지 링크를 대체하므로 성공 추정(도시에 gap 3 — 추적만 있음).
  run-phase에서 직접 확인할 것. REQ-CSL-002의 "어디든" 조건에 포함된다.
- **링크 대상이 청소 집합 안에 있는 경우**(링크 A→청소 루트 B): B가 먼저 지워져도 A는
  dangling 처분으로 제거+재배포되어 수렴 — AC-CSL-008이 입증.
- **루프 링크 / 자기 지목**: 제거-only 처분이라 무한 순회 없음(백업이 링크를 따라가지
  않음 — WalkDir-스킵 연속). run-phase에서 1회 스팟 확인 권장.
- **`--restore` 상호작용**: Run D 사후 복구 경로(`moai update --restore`)는 본 변경으로
  불필요해지는 경로지만 기능 자체는 무관(범위 밖 확인만).

## §D.4 간접 검증 (indirect verification)

- 기존 `deploy` 패키지 테스트 전량 녹색 = 실 디렉터리/실 파일 비회귀(REQ-CSL-006)의
  간접 증거.
- 기존 init/update 통합 테스트(링크 없는 경로) 녹색 = 전체 흐름 비회귀.
- 비-darwin CI 매트릭스에서 skip-not-fail 관측 = REQ-CSL-012.

## §D.5 닫힘 게이트 (closure gates)

- AC 11건 전부 통과(P0 우선). TDD: 각 형태 처분에 대응하는 RED 관측 기록(M1의 Run D
  전환 테스트가 기준선).
- plan.md §E 명령 전량 통과 + deploy 패키지 커버리지 ≥85%(커밋당 ≥80%).
- AC-CSL-010 자점검 표 완료(형태 일치·축 표기).

## §D.6 Definition of Done

- 12 REQ 구현, 11 AC 통과, 5형태 픽스처가 테스트에 상주.
- 진행줄 토큰(경로+symlink)이 안정적으로 grep 가능.
- 문서: 사용자 문서에 update 출력 예시가 있다면 sync-phase 갱신 후보로 보고(강제 아님).

## §D.7 전방 확인 (forward-looking checks)

1. **t81(가) 착지 후 교차 계약 재실행** — `.agents/` 미러 배포가 추가된 release/v3.1.3
   통합 트리에서 AC-CSL-009 재실행. 미러가 청소 집합에 들어가는지 여부는 t81(가)의
   소관이나, 발산이 발견되면 이 SPEC의 REQ-CSL-009 위반이다.
2. **파일 뿌리 dangling 직접 실측**(§D.3 첫 항 — 도시에 gap 3).
3. **대기 순서 제약**: 보존 처분으로의 전환은 본 SPEC 수정을 요구한다(spec.md §B.2).
