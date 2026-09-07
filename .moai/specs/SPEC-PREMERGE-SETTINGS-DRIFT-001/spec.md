---
id: SPEC-PREMERGE-SETTINGS-DRIFT-001
title: "병합 창 진입 전 카드 워크트리의 tracked .claude/settings.json drift 단정 — 검출·보존·거절"
version: "0.6.2"
status: completed
created: 2026-09-06
updated: 2026-09-07
author: manager-spec (card t488)
priority: P2
phase: "v3.2.0 target"
module: internal/cli (moai integration) + internal/kanban
lifecycle: spec-anchored
tags: "integration-window, settings-json, drift-gate, worktree, t488, t487-followup"
tier: M
related_specs: [SPEC-SETTINGS-ORIGIN-001]
---

# SPEC-PREMERGE-SETTINGS-DRIFT-001 — 병합 전 settings.json drift 단정

## HISTORY

| 날짜 | 사건 |
|---|---|
| 2026-09-06 | 생성 (카드 t488, plan 단계). SPEC-SETTINGS-ORIGIN-001(카드 t487)의 권고 C5를 구현 카드로 분리한 것. t487은 권고만 냈고 구현은 별도 카드로 남겼다(`.moai/reports/t487/verdict.md` §Q3). |
| 2026-09-06 | v0.6.2 — 델타 감사 S1 수리: `progress.md`가 인용하던 `dea4be776`은 amend 이전 커밋이라 dangling object이므로 `d02db303b`으로 교체(검증 가능성 주장이 해소되지 않을 앵커에 걸려 있었다). 같은 자리의 문장 접합 오류(S2)를 함께 정리하고, `AC-PSD-013(e)`의 암묵 의존((a)가 lock 레코드 존재를 세운다)을 명시. |
| 2026-09-06 | v0.6.1 — 기록 보완만(요구·수용·설계 변경 0). v0.5.0 행의 어긋남 기술을 양측으로 넓혔다 — 완료 보고와 그 뒤의 판독 **둘 다** 이미 움직인 상태를 서술했다. 아울러 v0.4.0 → v0.5.0 델타 주장이 기계 비교가 아니라 판독임을 명시(v0.4.0 미커밋). |
| 2026-09-06 | v0.6.0 — 확인 감사(PASS 0.930) 잔여 2건 종결 + 스윕 적발 1건. R1 `plan.md:180`을 `status` carrier로, R3 경계 사례를 판정층으로. 상시 조항을 **내용 우선** 어순으로 재배열하고 성공 검사(`err != nil` 포함)를 명시적으로 배제. **스윕 적발**: `REQ-PSD-008`이 무조건 거절을 요구해 REQ-PSD-016과 요구층에서 모순 — 거절만 `Where` 게이트로 조건화하고 보고는 비게이트로 분리. `REQ-PSD-009`도 같은 조건화, `AC-PSD-013`에 (e) 추가로 반증. |
| 2026-09-06 | v0.5.0 — M3 픽스처에 "bare 원격은 만들되 브랜치 push를 잊는" 실수 모양과 그 구성 직후 확인을 명시. 기록: 리드가 이 회차를 marker 1 적용 회차로 지시했으나, marker 1은 v0.4.0에서 이미 종결됐다(리드 측정이 v0.4.0 편집 이전 시점의 것이었고, 디스크 재측정으로 확인 — 미해결 마커 0건, `plan.md:40` 기본값 `false`). 중복 적용하지 않았다. |
| 2026-09-06 | v0.4.0 — 3회차 감사(0.895) 수리 + **marker 1 종결**(운영자 결정: 계열 (가) 기본 OFF + 켜면 거절). REQ-PSD-016 신설(킬 스위치는 거절 층만 게이트), REQ-PSD-014에 `clean`/`drift`/`undetermined` 세 상태 명명, REQ-PSD-012에 판정 상태 필드 상시 포함, AC-PSD-009/010 재작성(거절 층 켠 설정 전제 명시), AC-PSD-013 신설. 감사 수리분: `AC-PSD-007(d-3)`이 어느 저장소를 재는지 명시(N5: 대상 트리 + primary 루트 양쪽), 동일성 단정의 측정 공허를 부재 단정 규율에 흡수하고 (d-3)·AC-008에 내용 기반 대조 적용(N6), 게이트 경로 실행기 우회 뮤턴트를 M2 표와 DoD에 추가해 뮤턴트 6종으로(N7), `AC-PSD-009` 예외를 조항과 marker 1 양쪽에 기록(N8). |
| 2026-09-06 | v0.3.0 — 2회차 감사(0.875, 임계 초과) 수리. 1회차 수리가 심은 N1-N3 종결: 부재 단정 앞 양성 대조 + 기록 독립 직접 관측(N1), 픽스처 `.gitignore`에 의존하던 ignored 분류 단정을 스테이징 부재로 교체(N2), `AC-PSD-012` 청구를 M3 → M4로 이동(N3). 같은 모양으로 `AC-PSD-002`/`-003`/`-006`도 함께 수리했고, REQ-PSD-013에 실행기 단일 원천을 게이트 경로 전체로 확장했다. |
| 2026-09-06 | v0.2.0 — plan 감사 FAIL(0.725, 임계 0.80) 수리. `.moai/reports/t488/plan-audit.md`의 D1-D8 반영: REQ-PSD-014/015 신설(술어 실패·보존 실패가 통과가 되지 않음), REQ-PSD-013에 argv 단일 원천 조항 추가, 새 Go 코드가 계산하는 해시를 sha256으로 이관, AC-PSD-011/012 신설, AC-PSD-007에 (d) 추가. `§1`의 t334 실측 기록(md5)은 실행된 명령의 기록이므로 손대지 않았다. |

## §1 문제 진술

카드 워크트리의 tracked `.claude/settings.json` 워킹 사본이 정체불명의 작성자에 의해 수정된 채 발견된 사례가 두 번 있다. 두 번 모두 병합 창은 이를 알아채지 못했다.

- **1차** — 워크트리 `t452`. 카드 t480 조사 중 발견됐고, 병합이 우연히 거부되면서 드러났다.
- **2차** — 워크트리 `t334`. 카드 t487의 79개 트리 전수 스윕에서 발견됐으며, 1차 발견으로부터 9일이 지나도록 그 사이 아무도 알아채지 못했다.

2차 인스턴스는 본 워크트리에서 카드 착수 전에 독립 재측정됐다(워크트리 가드가 cross-tree `git`을 거부하므로 공유 오브젝트 저장소 판독 + 파일 해싱으로 측정):

```
$ md5 -q <t334>/.claude/settings.json
4f455d9425a396d38c202f2614bc918f            # 워킹 사본
$ git show 4837b05211fc7288747742f380de4048bc43de19:.claude/settings.json | md5 -q
437834679fcc2435761e264cec2ffc15            # 그 트리 HEAD blob
$ md5 -q .moai/reports/t487/preserved-copies/settings.json.t334-dirty
4f455d9425a396d38c202f2614bc918f            # t487 보존 사본과 바이트 일치
```

워킹 사본이 HEAD blob과 다르므로 t334는 dirty이며, t487이 반출한 보존 사본과 정확히 같다.

### t487이 닫아 둔 축 — 재개하지 않는다

t487은 작성자 귀속(누가 썼는가)을 **귀속 불가**로 종결했다(verdict Gaps 2, t485 C4 dead-axis 준수). 본 SPEC은 그 축을 다시 열지 않는다. 여기서 다루는 것은 **검출 지점**이다: 작성자를 특정하지 못해도, 그 산물이 병합에 실리기 전에 잡히면 피해는 막힌다.

### 왜 병합 창인가

t487 verdict §Q3의 premise 세 가지가 이 카드의 근거다.

1. 79개 트리 중 유일하게 발견된 dirty(t334)는 9일 동안 어떤 절차에도 걸리지 않았다. 병합 창 검사만이 **보증된** 검출 지점이다(t452는 병합 거부가 우연히 잡은 것이다).
2. dirty tracked 파일이 병합에 실리면 낡은 내용이 develop으로 흘러들 수 있다(t480 사례는 방향이 반대여서 피해가 없었을 뿐이다).
3. 검사 형태는 `--no-optional-locks`가 필수다. 카드 t485는 `moai statusline`의 평범한 `git status --porcelain`이 인덱스 **쓰기 락**을 수십 ms 잡는 것을 실측했다. 여러 레인이 도는 머신에서 병합 직전에 쓰기 락을 잡는 검사는 자기가 막으려는 경합을 스스로 만들어낸다. t487도 같은 처방(D7)을 79트리 스윕에 적용했다.

## §2 요구사항 (GEARS)

- **REQ-PSD-001** (Ubiquitous): drift 술어는 `git --no-optional-locks status --porcelain -- .claude/settings.json`이며, 판정은 그 출력의 **일치 줄 수(match count)**로 내린다. 0줄이면 통과, 1줄 이상이면 drift다.
- **REQ-PSD-002** (Unwanted): 술어는 어떤 경우에도 인덱스 쓰기 락을 잡는 형태(`--no-optional-locks` 없는 `git status`)로 실행되지 않는다.
- **REQ-PSD-003** (Unwanted): 판정은 프로세스 종료 코드로 내리지 않는다. 종료 코드는 호출자 편의 신호일 뿐이며, 어떤 수용 기준도 종료 코드에 의존하지 않는다.
- **REQ-PSD-004** (Ubiquitous): 술어의 경로 인자는 `.claude/settings.json` 하나로 고정되며, 대상 트리는 호출자의 작업 트리 최상위(`git rev-parse --show-toplevel`)다.
- **REQ-PSD-005** (Event-driven): **When** drift가 검출되면, `moai`는 그 워킹 사본을 primary 체크아웃 아래 보존 디렉터리로 복사하고, 보존 경로·sha256·바이트 크기·카드 id·브랜치·원본 워크트리 경로·측정 시각을 원장(ledger)에 한 줄로 기록한다.
- **REQ-PSD-006** (Unwanted): `moai`는 검출된 `.claude/settings.json`을 **절대** 자동 복원·되돌리기·삭제하지 않는다. 이 파일은 런타임(`moai glm` / `moai cc` / `moai cg`, SessionStart 훅)이 쓰며 머신 고유값(토큰, 절대경로, tmux pane id)을 담을 수 있으므로, 자동 복원은 그 자체로 데이터 파괴다. 처분은 사람이 정한다.
- **REQ-PSD-007** (Unwanted): `moai`는 보존 사본을 자동으로 커밋하거나 push하지 않는다. 보존 사본은 gitignore된 상태 디렉터리에 untracked로 남으며, 저장소 이력에 넣는 결정은 시크릿 스캔을 거친 사람의 판단이다(t487 L2 선례).
- **REQ-PSD-008** (Event-driven + Capability gate): **When** 레인이 병합 창을 잡기 위해 `moai integration acquire`를 실행하고 drift가 검출되면, `moai`는 보존 경로·sha256·리드에게 보고할 문구를 출력한다. **Where** 거절 층이 켜져 있으면(`workflow.settings_drift_gate.enabled: true`), 그 명령은 창을 기록하지 않고 거절한다. 꺼져 있는 기본 설정에서는 창을 기록하되 같은 보고를 낸다 — 보고는 게이트되지 않고 거절만 게이트된다(REQ-PSD-016).
- **REQ-PSD-009** (Capability gate): **Where** 거절 층이 켜져 있고 호출자가 명시적 우회 플래그를 지정하면, `acquire`는 drift가 있어도 창을 기록하되 우회 사실과 보존 경로를 lock 레코드와 출력 양쪽에 남긴다. 거절 층이 꺼져 있으면 창은 어차피 기록되므로 이 플래그는 아무것도 우회하지 않으며, 우회로 기록되지도 않는다 — 우회할 거절이 없는데 우회로 적으면 기록이 거짓말을 한다.
- **REQ-PSD-010** (Unwanted): 우회는 `--force`에 얹지 않는다. `--force`는 "살아 있는 보유자에게서 창을 빼앗는다"는 별개 축이며, 두 결정을 한 플래그에 묶으면 어느 쪽을 의도했는지 기록이 말하지 못한다.
- **REQ-PSD-011** (Capability gate): **Where** 사람이 창과 무관하게 손으로 확인하려 하면, 같은 술어를 단독으로 실행하는 읽기 표면이 제공된다.
- **REQ-PSD-012** (Ubiquitous): 검출 결과는 기계 판독 가능한 형태(`--json`)로도 제공되며, 그 표현은 판정 상태 필드를 **항상** 담고, 판정이 성립한 경우 일치 줄 수 필드를 함께 담는다. 리드는 이 두 필드로 판정을 읽는다.
- **REQ-PSD-013** (Ubiquitous): 술어와 게이트는 뮤턴트 반증 가능해야 한다. 술어 삭제·판정 반전·경로 오지정·`--no-optional-locks` 누락 각각에 대해 실패로 뒤집히는 검증이 존재한다. argv는 정확히 한 곳에서 구성되고, 실행기는 그 구성 결과를 그대로 받으며 자기에게 넘어온 argv를 기록한다. argv를 단정하는 검증은 **실행 경계에서 기록된 argv**를 읽는다 — 빌더를 따로 노출해 그 반환값을 부르는 형태는 이 요구를 만족시키지 못한다(빌더와 실행 경로가 갈리면 어느 쪽에서 플래그를 빼도 검증이 초록으로 남는다). 이 단일 원천은 술어에만 걸리지 않는다 — 게이트가 대상 트리나 그 저장소에 대해 실행하는 **모든** 명령이 같은 실행기를 통과하며, 따라서 기록된 목록은 그 실행의 전부를 담는다. 목록이 실행의 일부만 담으면 그 목록에 대한 어떤 부재 단정도 아무것도 주장하지 못한다.
- **REQ-PSD-014** (Unwanted): 술어 실행이 실패한 상황(대상이 git 저장소가 아니다, `git`이 없다, 권한이 없다)에서 `moai`는 통과 판정을 내리지 않는다. 판정 표면은 **세 상태**를 구별한다 — `clean`(측정했고 drift 없음) / `drift`(측정했고 drift 있음) / `undetermined`(측정하지 못함). 실패는 언제나 `undetermined`이며 `clean`으로 접히지 않는다. `undetermined`에서는 일치 줄 수 필드를 통과처럼 읽히는 값으로 채우지 않고 생략하며, 오류를 함께 보고한다. 신호 부재는 깨끗함의 증거가 아니다.
- **REQ-PSD-015** (Unwanted): 보존 사본 쓰기가 실패한 상황에서 `moai`는 drift 판정을 통과로 뒤집지 않는다. drift 판정은 술어의 일치 줄 수만으로 결정되고, 보존 실패는 그와 별개의 실패로 보고된다.
- **REQ-PSD-016** (Unwanted): 킬 스위치 `workflow.settings_drift_gate.enabled`는 **거절 층만** 게이트한다. 검출·보존·원장 기록은 그 플래그의 값과 무관하게 `acquire`마다 돌며, 플래그가 꺼져 있다는 이유로 생략되지 않는다. 저장소의 네 가드가 지켜 온 규율과 같다(`internal/config/types.go:670-679` — 관찰과 감사는 플래그와 무관하게 돈다).

## §3 수용 기준

`AC-PSD-001`..`AC-PSD-013`은 [`acceptance.md`](./acceptance.md)에 Given-When-Then으로 있다. 요구-검증 대응은 아래와 같다.

- **AC-PSD-001** — 적중 방향 (maps REQ-PSD-001, REQ-PSD-011)
- **AC-PSD-002** — 통과 방향 (maps REQ-PSD-001, REQ-PSD-011)
- **AC-PSD-003** — 경로 오지정 뮤턴트 검출 (maps REQ-PSD-004, REQ-PSD-013)
- **AC-PSD-004** — 경로 인자 누락 뮤턴트 검출 (maps REQ-PSD-004, REQ-PSD-013)
- **AC-PSD-005** — `--no-optional-locks` argv 고정 (maps REQ-PSD-002, REQ-PSD-004, REQ-PSD-013)
- **AC-PSD-006** — 판정이 종료 코드에 걸려 있지 않음 (maps REQ-PSD-003)
- **AC-PSD-007** — 보존과 원장 (maps REQ-PSD-005, REQ-PSD-007, REQ-PSD-012)
- **AC-PSD-008** — 원본 불변 (maps REQ-PSD-006)
- **AC-PSD-009** — 거절이 창을 잡지 않음 (maps REQ-PSD-008, REQ-PSD-012, REQ-PSD-013)
- **AC-PSD-010** — 우회는 기록됨 (maps REQ-PSD-009, REQ-PSD-010)
- **AC-PSD-011** — 술어 실패는 통과가 아님 (maps REQ-PSD-014)
- **AC-PSD-012** — 보존 실패가 판정을 뒤집지 않음 (maps REQ-PSD-015)
- **AC-PSD-013** — 기본 설정은 거절하지 않되 관찰층은 돈다 (maps REQ-PSD-016)

## §4 제약 [HARD — run 단계 구속]

- **t334 워크트리는 읽기 전용이다.** 운영자가 보존 전용·복원 금지로 정했다. 세 번째 인스턴스가 나타났을 때 대조할 대조 표본이므로, 본 카드는 그 트리의 어떤 파일도 수정·삭제하지 않는다.
- **검증은 카드 범위로 한정한다.** 전체 스위트를 로컬에서 돌리지 않는다(`CLAUDE.local.md` §4/§6). 건드린 패키지만 돌리고 전 패키지 판정은 CI에 맡긴다.
- **테스트 픽스처는 `t.TempDir()` 안에서 구성한다.** 프로젝트 루트나 실 워크트리를 픽스처로 쓰지 않는다.
- **배경 부하를 만들지 않는다.** 이 머신에서는 여러 레인이 동시에 돈다.
- 작성자 귀속을 재개하지 않는다(t487 Gaps 2 / t485 C4).

## §5 범위

들어오는 것: `.claude/settings.json` 하나에 대한 drift 술어, 그 술어를 부르는 CLI 표면(창 진입 전제 + 단독 확인 표면), 보존·원장 기록, 거절과 명시적 우회, 그리고 이 전부에 대한 양방향 뮤턴트 반증 테스트. 관련 doctrine 문서의 절차 문구 갱신.

닫는 모양: 병합 트리에서 건드린 패키지 테스트가 통과하고, 두 방향(hit / pass) 검증이 실측 출력으로 인용된 상태.

### Out of Scope — 감시 대상 파일 확대

- `.claude/settings.local.json`, `.moai/config/**`, 그 밖의 어떤 파일도 이 게이트의 대상이 아니다. 운영자는 확대를 선택하지 않았다. 확대가 타당하다는 근거가 나오면 plan.md의 `[NEEDS CLARIFICATION]` 항목으로 남기고, 결정 없이 넓히지 않는다.

### Out of Scope — 자동 복원과 자동 커밋

- 검출된 파일의 복원·되돌리기·삭제(REQ-PSD-006), 보존 사본의 자동 커밋·push(REQ-PSD-007). 처분은 사람의 결정이다.

### Out of Scope — 작성자 귀속

- dirty 사본을 누가 썼는지 특정하는 어떤 조사도 하지 않는다. t487이 귀속 불가로 종결했고, 본 카드는 검출만 담당한다.

### Out of Scope — described-source-diff 술어 개선

- 카드 t478이 남긴 잔여 항목(공백만 바뀐 파일을 drift로 오계수하는 문제)은 별개 카드 후보로 이미 큐에 있다. 본 카드는 그 코드에 손대지 않는다.

### Out of Scope — 기존 통합 락 의미 변경

- `moai integration`의 창 보유·해제 의미, `--force`의 의미, PreToolUse 통합 락 가드의 판정 로직은 그대로 둔다. 본 카드는 `acquire` 앞에 전제를 추가할 뿐이다.

## §6 의존과 참조

- `.moai/reports/t487/verdict.md` — 본 카드의 근거 판정서. §Q3가 이 카드의 원문 권고이고, E1/E2가 t334 실측이다.
- `.moai/specs/SPEC-SETTINGS-ORIGIN-001/` — t487의 SPEC 아티팩트(문제 정의와 지문).
- `internal/cli/integration.go` — `moai integration` 세 동사, `integrationLockRoot()`(:46), `integrationSessionID()`(:80), `resolveIntegrationTarget()`(:117), `acquire`의 `--card`/`--force` 플래그 정의(:275-276).
- `internal/kanban/integration_lock.go` — lock 레코드 구조(`IntegrationLock`, :109)와 경로(`integrationLockPath`, :151 — `<root>/.moai/state/integration-lock.json`).
- `internal/hook/integration_lock_guard.go` — PreToolUse 가드의 fail-open 규율과 deny sentinel 형태(모방 대상).
- `internal/config/types.go:650` `IntegrationLockConfig` / `internal/config/defaults.go:884` — 저장소의 opt-in 게이트 선례.
- `.claude/rules/moai/workflow/kanban-dispatch.md` § Integration into the release branch is self-served — 레인이 창을 잡는 절차.
- `CLAUDE.local.md` §4.1 — 레인의 통합 창 운영 절차.
