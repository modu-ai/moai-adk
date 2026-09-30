---
id: SPEC-CODEX-GATE-SCOPE-001
title: "수락 기준 — codex 리뷰 게이트 스코핑"
version: "0.1.0"
created: 2026-10-01
---

# SPEC-CODEX-GATE-SCOPE-001 — 수락 기준

## §A 관측 규율 [HARD]

모든 스코프 AC 는 **게이트가 조립한 검사 요청의 대상 필드**(target·cwd)와 **receipt 의 바인딩 상태**(스코프 키·ConfigDigest)를 관측 대상으로 삼는다. verdict 값은 요청과 무관하게 스크립트될 수 있으므로 단독 근거가 못 된다.

라이브 codex 를 요구하는 AC 는 없다 — 판별기와 측정식은 주입형 seam 으로 검증한다(선례: `reviewGateChangeDetector`, `codex_review_gate_live_test.go` 의 skip 규율). skip 은 통과가 아니라 **미관측**이다.

## §B 픽스처 설계 (판정 기준의 원천)

**카드 워크트리 픽스처**: 실제 git 저장소 + 연결 워크트리. 카드 워크트리 — 브랜치 `WT-fixture-card`, develop 분기 뒤 커밋 1건(파일 A) + 미커밋 수정 1건(파일 B). primary 역할 트리 — 미커밋 외부 WIP 1건(파일 F, 레인 소유 아님).

**env 매트릭스** (REQ-CGS-004):

| 사례 | 브랜치 | env | 기대 스코프 |
|---|---|---|---|
| 카드 세션 | `WT-*` | 무엇이든 | 카드 |
| 리더/일반 세션 | `develop` 등 | 있든 없든 | 트리 |
| stale env | `develop` | `MOAI_FACTORY_WORKER=<임의>` | 트리 (env 무시) |
| 미식별 | detached / 비 WT- | 없음 | 트리 (명시된 페일오픈) |

env 값은 어떤 문자열이든 같게 거동해야 한다 — 값 의존이 곧 라벨 형식 결합(t1378)이므로, 픽스처는 서로 다른 임의 값 2종 이상을 쓴다.

## §C RED 확립 규율 [HARD]

AC-CGS-001·003~010·012·013 은 **변경 전 구현에서 실패해야 한다** — 프로덕션 변경 전에 검사만 추가해 `--- FAIL` 을 관측하고 `.moai/reports/t1383/red/` 에 출력을 남긴다. AC-CGS-002·011 은 회귀선(현행 초록)이다.

RED 실행은 `-v` 로 함께 돌려 `=== RUN` 행에서 대상 케이스가 실제 실행됐음을 확인한다 — 셀렉터 0매칭이 초록으로 보이는 사고 방지. RED 기록에는 캡처 시점의 구현 전 트리 SHA(`git rev-parse HEAD` 출력)를 함께 남긴다(D3) — 판정 가치는 구현 전 상태 귀속에서 나온다.

## §D AC 매트릭스

| AC | 요구 | RED 로 시작 | 관측 대상 |
|---|---|---|---|
| AC-CGS-001 | REQ-CGS-002 | 예 | 검사 요청 target·cwd |
| AC-CGS-002 | REQ-CGS-003 | 아니오 (회귀, 초록) | 검사 요청 target·cwd |
| AC-CGS-003 | REQ-CGS-004 | 예 | 스코프 판별 결과 |
| AC-CGS-004 | REQ-CGS-004 | 예 | 스코프 판별 결과 |
| AC-CGS-005 | REQ-CGS-004 | 예 | 스코프 판별 결과 |
| AC-CGS-006 | REQ-CGS-005 | 예 | 검사 요청 cwd·diff 내용 |
| AC-CGS-007 | REQ-CGS-006 | 예 | reviewer 미호출 + ALLOW |
| AC-CGS-008 | REQ-CGS-001+003 | 예 | 스코프 판별 결과 |
| AC-CGS-009 | REQ-CGS-007 | 예 | receipt 매치 행동 |
| AC-CGS-010 | REQ-CGS-009 | 예 | 양 경로의 스코프 동일성 |
| AC-CGS-011 | REQ-CGS-008 | 아니오 (회귀, 초록) | ALLOW + 부재 사유 |
| AC-CGS-012 | REQ-CGS-002 | 예 | 스코프 키 기저 |
| AC-CGS-013 | REQ-CGS-010 | 예 | 구조화 로그 행 |

REQ 매핑 누락 없음: REQ-CGS-001→AC-CGS-008, 002→001·012, 003→002·008, 004→003·004·005, 005→006, 006→007, 007→009, 008→011, 009→010, 010→013.

---

### AC-CGS-001 — 카드 세션의 검사는 카드 diff 로 간다

**Given** 카드 워크트리 픽스처(§B — 커밋 1건 + 미커밋 1건)와 primary 역할 트리의 외부 WIP(파일 F)가 주어지고,
**When** 카드 세션의 턴이 게이트에 닿으면,
**Then** 게이트가 조립한 검사 요청의 대상은 카드 diff(merge-base..HEAD + 카드 트리 미커밋분)이고, 파일 F 는 대상에 없으며, cwd 는 카드 워크트리다.

### AC-CGS-002 — 비카드 세션의 검사는 변하지 않는다

**Given** develop(또는 비 WT- 브랜치) 트리와 미커밋 변경이 주어지고,
**When** 그 세션의 턴이 게이트에 닿으면,
**Then** 검사 요청은 현행과 shape-identical 이다 — 대상 = 해상 트리의 uncommittedChanges, REQ-CRT-006 직렬화 형태 유지.

### AC-CGS-003 — stale env 는 스코프를 바꾸지 못한다

**Given** develop 세션에 `MOAI_FACTORY_WORKER=<임의 라벨>`이 설정돼 있고,
**When** 턴이 게이트에 닿으면,
**Then** 스코프는 트리로 판별되고, env 값은 판정에 쓰이지 않는다(관측 맥락으로만 기록).

### AC-CGS-004 — 브랜치만으로 카드 스코프가 정해진다

**Given** WT- 브랜치 세션에 env 가 전혀 없고,
**When** 턴이 게이트에 닿으면,
**Then** 스코프는 카드로 판별된다 — env 부재는 카드 판별을 막지 않는다.

### AC-CGS-005 — 라벨 형식에 결합하지 않는다

**Given** 같은 WT- 세션에 서로 다른 임의 env 라벨 2종을 번갈아 설정하고,
**When** 각각 턴이 게이트에 닿으면,
**Then** 스코프 판별 결과는 두 값에서 동일하다 — 판별기가 라벨 텍스트를 소비하지 않는다(t1378 정합). 이 AC 는 부재 주장이 아니라 **양성 관측**(두 값 동일 결과)으로 판정한다.

### AC-CGS-006 — freeze 된 projectDir 를 우회한다

**Given** 세션 cwd 는 카드 워크트리이고 `CLAUDE_PROJECT_DIR` 는 primary 를 가리키며(spec.md §A.3 (b)), primary 에 외부 WIP 가 있고,
**When** 턴이 게이트에 닿으면,
**Then** 스코프는 세션 cwd 트리에서 정해지고, 검사 대상에 primary 전용 변경(파일 F)은 없다.

### AC-CGS-007 — 빈 카드 diff 는 reviewer 를 부르지 않는다

**Given** 카드 스코프 세션의 카드 diff 가 비어 있고(커밋 0, 미커밋 0),
**When** 턴이 게이트에 닿으면,
**Then** 게이트는 reviewer 를 호출하지 않고 ALLOW 한다 — 셀프게이트가 검사할 스코프와 같은 것을 잰다.

### AC-CGS-008 — 미식별 세션은 트리 스코프로 떨어진다

**Given** detached HEAD 또는 비 WT- 브랜치 + env 부재 세션이 주어지고,
**When** 턴이 게이트에 닿으면,
**Then** 스코프는 트리로 판별된다 — 명시된 페일오픈(decision-index Q5).

### AC-CGS-009 — 판정은 스코프 경계를 넘지 않는다

**Given** 카드 스코프에서 fail verdict receipt 가 기록돼 있고, 별개의 리더 세션이 다른 스코프 상태로 턴을 끝내며,
**When** 리더 턴이 게이트(멤버 6 포함)에 닿으면,
**Then** 그 receipt 로 차단되지 않는다. 대신 **Given** 같은 카드의 같은 diff 상태에서,
**When** 다시 턴이 닿으면,
**Then** 기록된 fail 로 재차단된다 — 스코프 상태가 변할 때까지.

### AC-CGS-010 — 두 경로가 같은 스코프를 본다

**Given** 같은 세션 상태(트리·브랜치·env)가 주어지고,
**When** 턴 종료 경로와 `moai verify codex-review` 각각이 스코프를 해상하면,
**Then** 두 결과의 스코프 클래스와 검사 대상 표현이 동일하다 — 서로 다른 판별 구현이 존재하지 않는다.

### AC-CGS-011 — fail-open 은 카드 스코프에서도 유지된다

**Given** 카드 스코프 세션에서 codex 바이너리가 없거나 리뷰 호출이 오류·inconclusive 이고,
**When** 턴이 게이트에 닿으면,
**Then** ALLOW 한다 — 어느 스코프에서든 차단 경로는 명시적 fail verdict 뿐이다(REQ-MCP-012).

### AC-CGS-012 — 스코프 기저는 핀되지 않는다

**Given** 카드 세션이 develop 의 신규 커밋을 흡수한 직후이고,
**When** 게이트가 카드 diff 를 계산하면,
**Then** 기저는 그 시점의 `git merge-base develop HEAD` 이다 — 흡수 전에 핀된 값이 남지 않는다(gitflow-lane-protocol §8 의 게이트 내 적용).

### AC-CGS-013 — 스코프 판정은 관측 가능하다

**Given** 어느 세션이든 턴이 게이트에 닿으면,
**When** 게이트가 스코프를 판별하면,
**Then** 구조화 로그에 클래스(card/tree)와 근거(branch match/none)·env 맥락이 클래스별로 구별 가능하게 기록된다.

---

## §E 품질 게이트와 DoD

- **DoD**: §D 전 AC 최종 상태 PASS · RED 관측 기록이 `.moai/reports/t1383/red/` 에 존재 · `go vet`(darwin+windows) 신규 0 · 대상 패키지 커버리지 85% 이상(`quality.yaml` `test_coverage_target`) · §D 추적표 상 미매핑 REQ 없음.
- **간접 검증**: AC-CGS-005 는 양성 귀결을 포함한다 — "라벨을 소비하지 않는다"의 검증은 다른 값 2종에서 동일 결과라는 양성 관측이어야 하고, 코드에 라벨 패턴이 없다는 부재 주장 단독이 아니다.
- **선행 폐쇄 게이트**: M1 회귀선(AC-CGS-002·011)이 변경 전 트리에서 초록으로 관측되지 않으면 이후 마일스톤을 시작하지 않는다(spec.md §D).
