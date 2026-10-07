# 의존성 그래프

## 현재 최종 통합 트리의 의존성 — e65b3b

기준은 `e65b3b469c0ee71195b0b568b4b66d0b364b6f3d`다. 현재 darwin/arm64의 `go list -deps -json ./...`에서 모듈 패키지170개, 내부 import481쌍, 최상위 집계306쌍을 측정했다. 보존된840826 JSON과 정확한 set 비교에서 `internal/factorylane` → `internal/auditverdict` 한 쌍이 package·folded 집계에 각각 추가됐고 제거는0개였다. 패키지 집합170개는 같고 `go.mod`·`go.sum`도 변하지 않았다. 숫자만 같아서 관계 동일성을 추정한 결과가 아니다.

새 의존성은 SPEC 없는 카드의 감사 admission을 공통 술어로 연결한다. folded fan-in에서 `internal/auditverdict`는4→5로 아래 표에 들어왔고, `internal/factorylane`의 fan-out은3→4다. 게이트 영수증 캐시는 기존 CLI 모듈 안에 추가됐고 Bun 실행은 기존 quality 모듈 안에서 바뀌었다. caller context와 사용자 자산 설치 배선은 유지한다. 아래 fan-in/out 표와 순환 절은 이번 집계이며 이전 판 수치는 각각 당시 기록이다.

## 이전 c572 기준의 의존성

기준은 `c572e7baceaa6fd0cd3c78a9335b4baabd320347`다. `go list -deps -json ./...`에서 내부 import 엣지 480개, 최상위로 접고 self-edge를 뺀 고유 쌍 305개를 측정했다. 이전 ff7722 이후 `internal/cli/worktree` → `internal/factory`, `internal/factory` → `internal/factorylane`, `internal/factory` → `pkg/version`이 추가됐고, `internal/cli/worktree` → `internal/config`는 제거됐다. `go.mod`·`go.sum` 차이는 없다.

현재 최상위 fan-out은 `internal/cli` 76, `internal/hook` 40, `internal/web` 16, `internal/factory` 9다. fan-in은 `internal/config` 30, `internal/atomicfile` 14, `internal/paths` 14, `internal/defs` 13, `internal/template` 7이다. 사용자 자산 패키지의 소스와 기존 설치 배선은 이번 창에서 바뀌지 않았다.

merge-ready는 `internal/cli/factory_merge.go`의 콜백으로 `internal/factorylane/merge.go`가 읽은 후보 tip tree의 기록을 검증한다. complete T16은 `internal/homestate/card_evidence_readers.go`가 실제 병합 tree의 기록을 검증한다. 공통으로 `ReadRemeasureRecord`와 `ValidateRemeasureRecord`를 쓰며 시점과 트리 출처가 다르다.

## 이전 ff7722 기준의 import 대조

`ff7722d2d157dd4e3cffd88ebb644e0f8ead83fa`에서 `go list -deps -json ./...`를 실행했다. 모듈 내부 패키지 170개, 내부 import 엣지 478개, 최상위로 접고 self-edge를 제거한 고유 쌍 303개다. `internal/userassets`는 `internal/template`을 import하며 비테스트 소비자는 `internal/cli` 하나다. 아래 main 기준 fan-in/out과 이전 기록은 각각 명시된 시점의 관측값이다.

> `internal/template/pluginemit`은 PR #1772에서 폐기됐다. 현재 빌드 방출기는 agentemit·commandemit·embedemit이며, 아래 이전 pluginemit 수치는 이력이다.

## 이전 081899 기준의 재측정

`081899adb825935d5263b1699fe730373deaa4fd`에서 `go list -deps -json ./...`를 실행했다. 모듈 내부 Imports 엣지는 478개이고, 최상위 패키지로 접어 self-edge를 뺀 고유 쌍은 302개다. 아래 이전 기록의 그래프 수치는 각 시점의 관측으로 남긴다.

현재 최상위 fan-out은 `internal/cli` 75, `internal/hook` 40, `internal/web` 16이다. fan-in은 `internal/config` 30, `internal/paths` 14, `internal/atomicfile` 14, `internal/defs` 13이다. 신규 빌드 도구 `internal/template/embedemit`은 표준 라이브러리만 사용하고 비테스트 fan-in은 0이다. 런타임에서 호출하지 않는 점은 `agentemit`·`commandemit`과 같다.

**이전 부분 재측정 — t1524, worktree `.claude/worktrees/develop`, 브랜치 `develop`, base `d0378d37c` (2026-10-05).**
문서의 산출 명령으로 내부 import를 다시 쟀다 — 패키지 단위 470→**476**, 최상위 접기 + self-edge 제거 고유 쌍 295→**301**. 신규 패키지 `internal/hygiene`(card t1518)은 내부 import `internal/config` 하나, 소비자 `internal/cli`·`internal/hook` 둘. 움직인 엣지는 전부 이번 창 몫이다 — `cli→hygiene`·`hook→hygiene`·`hygiene→config`(card t1518), `runtime→auditverdict`·`runtime→config`·`cli→auditverdict`(card t1500 — `audit_ceiling.go`와 `spec_ceiling.go`). fan-in 상위에서 움직인 행: `internal/config` 28→30(`hygiene`·`runtime` 합류). fan-out 상위: `internal/cli` 73→75, `internal/hook` 39→40. 작은 fan-in 표에 `internal/auditverdict` 4(`contract`·`homestate`에 `cli`·`runtime` 합류)와 `internal/hygiene` 2가 들어왔다. § 순환은 재확인 결과 변동 없음(신규 엣지는 전부 일방향 — leaf 방향). go.mod는 앵커 이후에도 한 줄도 바뀌지 않았다.

**이전 부분 재측정 — t1485, 브랜치 `WT-codemaps-regen3`, base `83086bec5` (2026-10-04).**
문서의 산출 명령으로 내부 import를 다시 쟀다 — 패키지 단위 461→**470**, 최상위 접기 + self-edge 제거 고유 쌍 289→**295**. 큐 도메인 패키지가 `internal/factory`로 개명되며(card t1399) 그 행들이 새 이름으로 옮겨졌다. 신규 패키지 셋: `internal/decision`(내부 import `internal/homestate` 하나, 소비자 `internal/cli`), `internal/auditverdict`(내부 import 0인 leaf, 소비자 `internal/contract`·`internal/homestate`), 당시 `pluginemit`(`internal/template`·`pkg/version` import, 비테스트 소비자 0 — 빌드타임 방출기). fan-in 상위에서 움직인 행: `internal/config` 27→28, `internal/homestate` 8→9, `pkg/version` 5→6, `internal/lockfile` 5로 상위 진입. fan-out 상위: `internal/cli` 72→73, `internal/web` 15→16, `internal/contract` 9→10, `internal/homestate` 4→6. go.mod는 t1456 앵커 이후에도 한 줄도 바뀌지 않았다.

**이전 부분 재측정 — t1456, worktree `.moai/worktrees/t1456`, 브랜치 `WT-codemaps-regen2`, base `5501c06af` (2026-10-03).**
문서의 산출 명령으로 내부 import를 다시 쟀다 — 패키지 단위 460→**461**, 최상위 접기 + self-edge 제거 고유 쌍 288→**289**. 창의 신규 8파일은 전부 기존 패키지 안에 들어와 신규 패키지는 없다. fan-out·fan-in 상위 표는 전 행 재확인 결과 변동이 없었고(`internal/cli` 72·`internal/hook` 39·`internal/config` 27·`internal/paths`·`internal/atomicfile` 14), 작은 fan-in 표도 변동 없었다. § 순환은 재확인 결과 변동 없음. go.mod는 t1443 앵커 이후에도 한 줄도 바뀌지 않았다 — § 외부 의존성 표는 t1443 판의 32항목이 그대로 유효하다.

**이전 부분 재측정 — t1443, worktree `.moai/worktrees/t1443`, 브랜치 `WT-codemaps-regen`, base `4bf547bca` (2026-10-02).**
문서의 산출 명령으로 내부 import를 다시 쟀다 — 패키지 단위 457→**460**, 최상위 접기 + self-edge 제거 고유 쌍 286→**288**. 신규 패키지 `internal/settings/agentfm`(card t1411)은 순수 leaf다 — 내부 import 0(표준 라이브러리와 `gopkg.in/yaml.v3`만), 내부 소비자는 `internal/web` 하나. fan-out·fan-in 상위 표는 전 행 재확인 결과 변동이 없었다(`internal/cli` 72·`internal/hook` 39·`internal/config` 27·`internal/paths` 14·`internal/atomicfile` 14 포함 — 창의 소폭계 신규 엣지는 기존 최상위 쌍 안쪽의 패키지 세분이다). 작은 fan-in 표도 변동 없었다(`auditreceipt` 3·`jev` 3·`jevcred` 3·`contract` 3·`mission` 2·`chain` 2·`stateanchor` 3·`civerdict` 2 — 같은 명령 재측정). § 순환은 다섯 쌍 재확인 결과 변동 없음(상호쌍 교차 재계산 — `cli`↔`hook` · `cli`↔`kanban` · `hook`↔`migration` · `contract`↔`escalation` · `profile`↔`settings`). go.mod는 앵커 이후 처음 움직였다 — 직접 require 30→32항목: 신규 `github.com/gorilla/websocket` v1.5.3(관리 세션의 Codex App-Server stream 전송 — `internal/cli/managed_codex_factory.go`, card t1375)과 `santhosh-tekuri/jsonschema/v6` v6.0.2의 indirect→직접 승격(`internal/codextools/registry.go`)이며 § 외부 의존성 표에 두 행을 더했다.

**이전 부분 재측정 — t1297, worktree `.moai/worktrees/t1297`, 브랜치 `WT-codemaps-regen`, base `a9f43a6fc` (2026-10-02, 전면 재생성 카드).**
t1333 판 이후 이 문서를 건드린 판이 없어(개별 카드가 `overview.md` 표만 옮겨 왔다) 창 전체를 한 번에 다시 쐈다 — 패키지 단위 449→**457**, 최상위 접기 + self-edge 제거 고유 쌍 277→**286**. go.mod·go.sum은 앵커 이후 한 줄도 바뀌지 않았다. fan-out 상위에서 움직인 행: `internal/cli` 70→72(`internal/factorylane` — card t1338 착지 — 와 세션 앵커 계열 합류). fan-in 상위에서 움직인 행: `internal/config` 25→27, `internal/atomicfile` 13→14, `internal/factory`·`internal/homestate` 7→8, `pkg/models` 8→7. § 순환은 네 쌍→**다섯 쌍** — 다섯째 `internal/profile` ↔ `internal/settings`의 양 엇키는 앵커 시점부터 존재하던 것이(`git grep` 실측 — `internal/profile/sync.go`) 이 판에 처음 표로 들어온 스테일 누락분이다(t1333 판의 auditreceipt·jev 정정과 같은 성격).

**이전 재측정 — t1333, worktree `.moai/worktrees/t1333`, 브랜치 `WT-codemaps-refresh9`, base `145c3d98c` (2026-09-29).**
문서의 산출 명령으로 내부 import를 다시 졌다 — 패키지 단위 449개, 최상위 접기 + self-edge 제거 고유 쌍은 277개으로 앞 판과 같았다. fan-in·fan-out 상위 표 전 행(`internal/cli` 70·`internal/hook` 39 포함)과 순환 4쌍은 재확인 결과 변동이 없었고, go.mod·go.sum도 앵커 이후 한 줄도 바뀌지 않았다. 이 판의 유일한 갱신지는 작은 fan-in 표의 스테일 수치 정정이다 — `internal/auditreceipt` 2→3(`internal/closure` 합류 — git grep으로 앵커 시점에 이미 존재하는 엇키였음이 확인된 스테일 값), `internal/jev` 2→3·`internal/jevcred` 2→3(`internal/contract` 합류 — 같은 성격), `internal/contract` 1→3(`internal/closure`·`internal/escalation` — 역시 같은 성격). civerdict 2·mission 2·chain 2·stateanchor 3은 변동 없었다.

**이전 재측정 — t1305, worktree `.claude/worktrees/t1305`, 브랜치 `WT-codemaps-refresh8`, base `afecf81e9e96` (2026-09-29).**
문서의 산출 명령(`go list -f` + 모듈 경로 필터)으로 내부 import를 다시 쟀다 — 패키지 단위 449개,
최상위 접기 + self-edge 제거 고유 쌍은 277개다. 현재 최상위 fan-out은 `internal/cli` 70,
`internal/hook` 39이며 fan-in은 `internal/config` 25, `internal/paths` 14,
`internal/defs`와 `internal/atomicfile` 각 13이다. 이번 배치(card t1246 에이전트 모델 은퇴·card t1289
커밋 신원 가드·card t1240 codex factory 복원·card t1294 lane slot 상한)는 패키지 경계를 둘 없앴다 —
cellguard(하네스의 fan-in 0 가드 — 엣지 변동 없음)와 agentfm(settings 하위 — 소멸로
settings의 fan-out이 줄었다). `internal/gitenv` 소비자가 둘에서 여섯으로 늘었고(§ 작은 fan-in → 상위 표),
§ 순환은 세 쌍에서 네 쌍으로 늘었다(`internal/contract` ↔ `internal/escalation`). § 외부 의존성은
`go.mod`가 앵커 이후 한 줄도 바뀌지 않은 것으로 확인했다.

**이전 재측정 — t1295, `develop` `cee197917` (2026-09-28).**
`go list -deps -json ./...`가 성공한 트리의 내부 import는 패키지 단위 452개,
`internal/<X>`·`cmd/<X>`·`pkg/<X>`로 접고 자기 엣지를 뺀 고유 쌍은 278개다.
이번 갱신의 구조 변화인 MoAI L1 워크트리·Factory 역할 전환은 기존 패키지 안에서
일어났으며 새 패키지 경계를 만들지 않았다.

**이전 재측정 — t1187, `origin/develop` `a8a9b9376` (2026-09-25).**
아래 import 엣지는 같은 `go list -f` 명령으로 다시 셌다. 비테스트 소스
변경 41개 중 Codex 감사 런처와 Factory 런 은퇴가 기존 `internal/cli`·
`internal/homestate`·`internal/factorymsg` 경계를 사용하며, 새 Go 패키지는
늘지 않았다.

> `/moai codemaps`로 생성됐습니다. 내부 엣지만 대상이며 stdlib·서드파티는 제거했습니다.

**최초 측정 트리**: worktree `.claude/worktrees/t592`, 브랜치 `WT-home-state-rollout`, HEAD `e7bd89ee3`, 2026-09-10
**재측정 트리**: worktree `.claude/worktrees/t869`, 브랜치 `WT-codemaps-refresh`, HEAD `a851b205c`, 2026-09-18 — 엣지 수, fan-in·fan-out 표 전체, 상호 참조 쌍, 새 leaf 표, `go.mod` 직접 require 항목 수와 버전, § 이례적인 것 7. § 이례적인 것 1~6의 서술은 이번에 버전·사용처 줄을 다시 대조했고 판단은 앞 판을 이어받았습니다.
**정기 재측정**: worktree `.claude/worktrees/t999`, 브랜치 `WT-codemaps-remediation`, HEAD `56c64891a`, 2026-09-20 — 엣지 수(365→371 · 222→227), fan-in 표에서 움직인 한 행(`internal/atomicfile` 10→11), fan-out 표에서 움직인 두 행(`internal/cli` 62→63 · `internal/hook` 32→35), 그리고 작은 fan-in 표의 신규 세 항목. 나머지 행은 같은 명령으로 재확인해 변동이 없었고, § 외부 의존성과 § 순환은 이번에 다시 재지 않았습니다(앞 판 인계).
**정기 재측정**: worktree `.claude/worktrees/t1069`, 브랜치 `WT-graph-restamp`, HEAD `0314801c2`, 2026-09-22 — 엣지 수(371→378 · 227→234), fan-in 표에서 움직인 두 행(`internal/defs` 11→12 · `internal/paths` 11→12), fan-out 표에서 움직인 두 행(`internal/cli` 63→65 · `internal/web` 15→16)과 하나의 정정(`internal/spec` — 앞 판 행이 3으로 적혔으나 스탬프 트리에서도 4였다), 작은 fan-in 표의 `internal/stateanchor` 2→3(소비자에 `internal/session` 합류)과 신규 세 행(`internal/jev` · `internal/jevcred` · `internal/jevmeasure`). § 순환은 같은 방법으로 재확인해 세 쌍 그대로였고, § 외부 의존성은 `go.mod`가 스탬프 이후 한 줄도 바뀌지 않은 것으로 확인했습니다.
**부분 재측정**: worktree `.claude/worktrees/t1092`, 브랜치 `WT-codemaps-restamp`, base `08113ff0f`, 2026-09-23 — 카드 t1092. 엣지 수(378→381 · 234→237). 신규 패키지 `internal/factorymsg`는 fan-in 2(`internal/cli` · `internal/hook`이 import)로 작은 fan-in 표(상위 14 밖)에 속하며, fan-out 상위 표에는 두 소비자 쪽 수치 변화가 반영됐지만 순위표 자체는 움직이지 않았습니다(`internal/cli`·`internal/hook` 모두 기존에도 상위권). § 순환·§ 외부 의존성·상호 참조 쌍 목록은 이번 변경과 무관해 손대지 않았습니다.
**정기 재측정**: worktree `.claude/worktrees/t1274`, 브랜치 `WT-codemaps-refresh4`, base `cf4b82755`, 2026-09-26 — 카드 t1274. 엣지 수(386→409 · 241→256; 최상위 집계는 계보 방식대로 고유 쌍 집합 기준). 증가분의 대부분은 신규 패키지 `internal/escalation`(비테스트 소비자 `internal/hook` 1개, 스스로는 `internal/config`·`internal/contract`·`internal/spec`·`internal/constitution`·`internal/homestate`·`internal/navigator/astx` 등을 import)와 t1235 계열 cli·hook·config 변경이 가져왔습니다. § 순환·§ 외부 의존성은 이번 변경과 무관해 손대지 않았습니다.
**정기 재측정**: worktree `.claude/worktrees/t1278`, 브랜치 `WT-codemaps-refresh5`, base `6d514f9b7`, 2026-09-27 — 카드 t1278. 엣지 수(409→413 · 256→260). 신규 엣지는 정확히 넷: 신규 패키지 `internal/civerdict`로 향하는 둘(`internal/cli`·`internal/escalation`)과 `internal/escalation`→`internal/verify`(ciLimb 의 HasLocalPass 소비), `internal/contract`→`internal/mission`(projection_mission 투영). t1242 가 지운 `internal/cli`→`internal/homestate` 접힌 엣지는 타 cli 파일이 유지해 상위 집계에 변동이 없습니다. fan-out 상위 표는 `internal/cli` 66→67(civerdict 합류) 한 행, 작은 fan-in 표는 `internal/mission` 1→2(소비자에 `internal/contract` 합류)와 신규 `internal/civerdict` 2 한 행. § 순환·§ 외부 의존성·상호 참조 쌍은 이번 변경과 무관해 손대지 않았습니다.

두 가지 해상도로 봅니다 — 패키지 단위 **476 엣지**, 이를 `internal/<X>` · `pkg/<X>` · `cmd/<X>`
최상위로 접고 self-edge를 제거한 **301 엣지**. 아래 표는 후자 기준입니다.

산출:

```
$ go list -deps -json ./... 의 프로젝트 패키지 Imports 중 모듈 내부 경로
476
```

> 앵커 `25a3212a9` 판은 이 자리에 1638을 적었습니다. 위 명령으로 재현되지 않고 그 판의
> 명령 인용이 생략형이라 무엇을 셌는지 복원할 수 없으므로, 이후 판은 위 명령의 출력을 싣습니다.
> 최상위 집계는 205 → 214 → 222 → 227 → 234 → 237로 움직였습니다.

---

## fan-in 상위 — 다른 최상위 패키지에게 import 당한 수

| # | 패키지 | 피import |
|---|---|---|
| 1 | `internal/config` | 30 |
| 2 | `internal/atomicfile` | 14 |
| 2 | `internal/paths` | 14 |
| 4 | `internal/defs` | 13 |
| 5 | `internal/homestate` | 9 |
| 6 | `internal/core` | 8 |
| 6 | `internal/factory` | 8 |
| 6 | `internal/spec` | 8 |
| 9 | `internal/execerr` | 7 |
| 9 | `internal/template` | 7 |
| 9 | `pkg/models` | 7 |
| 12 | `internal/gitenv` | 6 |
| 12 | `internal/hook` | 6 |
| 12 | `pkg/version` | 6 |
| 15 | `internal/auditverdict` | 5 |
| 15 | `internal/lockfile` | 5 |
| 15 | `internal/lsp` | 5 |
| 15 | `internal/statusline` | 5 |

산출은 최상위 집계 엣지 목록의 목적지 열을 `sort | uniq -c | sort -rn` 한 것입니다.

> **t1305 판에서 움직인 행.** `internal/config` 22→25, `internal/paths` 12→14, `internal/atomicfile`
> 11→13, `internal/spec` 6→8, `internal/factory`·`internal/homestate` 6→7, 그리고 상위 신규 진입
> `internal/gitenv` 2→6 — 소비자가 `cli`·`hook`에 `contract`(하위 `revoke`·`kickoff`)·`core`·
> `escalation`·`homestate` 넷이 합류했다(커밋 신원 가드 계열의 자식 git 프로세스 격리 확산,
> card t1289). 나머지 행은 같은 명령으로 재확인해 변동이 없었습니다.

상위 7행 중 **3개**가 cross-cutting leaf입니다(`atomicfile` · `paths` · `defs`).
나머지 넷은 `internal/config`·`internal/homestate`(data)와 `internal/core`·`internal/factory`(domain)이고,
이 배치는 안정 의존성 원칙에 부합하는 **건강한 신호**입니다.

> **정정 이력.** 이 자리의 cross-cutting 비율은 판마다 레이어 칸을 그대로 세어 갱신한다 —
> 어느 판이 「상위 7개 중 6개」로 적은 것을 5로 바로잡았고, t1305 판에서는 `internal/spec`(domain)이
> 상위 7행에 합류하며 4가 됐다. 집합이 아니라 수를 옮겨 적은 자리였음을 이 이력이 보여준다.

다만 `internal/hook`(11위, 6)과 `internal/statusline`(14위, 5)은 **presentation인데 피의존
대상**입니다. 방향이 뒤집혀 있고, 이것이 `overview.md`가 "레이어링이 국소적으로 무너진다"고
적은 근거입니다.

`internal/core`(8)의 소비자는 `cli` · `github` · `homestate` · `hook` · `kanban` · `stateanchor` ·
`statusline` · `workflow`입니다. `internal/factory`(8)의 소비자 중 `cmd/t657-merge`는 배포되지 않는
일회성 도구이고, `internal/graph`는 이 판의 GTD 비공개 투영(`internal/graph/gtd_private.go`)이
만든 엣지입니다.

### 작은 fan-in의 seam과 leaf

| 패키지 | fan-in | 비고 |
|---|---|---|
| `internal/hygiene` | 2 | **t1524 판에서 새로 들어왔다.**(card t1518, SPEC-MOAI-HYGIENE-001) 소비자는 `internal/cli`(`clean.go` — 수동 표면)와 `internal/hook`(`session_start_hygiene.go` — SessionStart 자동 경로)둘이며, 패키지 스스로는 `internal/config` 하나만 import 한다(workflow.hygiene 6키 — 두 경로가 같은 임계값을 읽는다) |
| `internal/auditverdict` | 5 | **t1485 판에서 새로 들어오고 t1524 판에서 넷이 됐다.** 소비자는 `internal/contract`·`internal/contract/kickoff`·`internal/homestate`(`card_audit_kickoff.go`)에 이번 판의 `internal/cli`(`spec_ceiling.go`)·`internal/runtime`(`audit_ceiling.go`) 합류(card t1500) — 현재 e65b3b에서는 `internal/factorylane`의 SPEC 없는 카드 admission이 추가돼 folded 소비자는5개다 |
| `internal/stateanchor` | 3 | 상태 앵커 seam. 소비자는 `internal/statusline`, `internal/cli`, 그리고 이 판에 합류한 `internal/session` — 레지스트리 경로 해석이 같은 seam을 쓰기 시작했다(워크트리마다 갈라지던 레지스트리 하나로 모으기) |
| `internal/chain` | 2 | 워크트리 세션 origin-trail 원장. 소비자는 `internal/cli`와 `internal/hook` |
| `internal/auditreceipt` | 3 | **t999 판에서 새로 들어왔다.** 소비자는 `internal/cli`와 `internal/hook` — 생산 쪽(MCP 도구 호출)과 소비 쪽(훅 가드)이 각각 하나씩이며, 그 비대칭이 아니라 대칭이 이 패키지의 설계다 **t1333 판 정정: 소비자는 cli·hook·closure 셋이다 — closure 엇키는 앵컰 이전부터 존재했고 이 판이 스테일 값을 바로잛었다** |
| `internal/jev` | 3 | **이 판에서 새로 들어왔다.** 소비자는 `internal/cli`(doctor·todo admission·숨은 suggest 앵커 세 파일)와 `internal/jevmeasure`(살아 있는 `Answerer` 구현) **t1333 판 정정: 소비자는 cli·contract·jevmeasure 셋이다 — contract 엇키는 앵컰 이전부터 존재했다** |
| `internal/jevcred` | 3 | **이 판에서 새로 들어왔다.** 소비자는 `internal/cli`와 `internal/web` — 위자드·doctor 쪽과 콘솔 Jev 패널이 각각 하나씩이며, 두 표면이 하나의 reader를 공유하는 것이 이 패키지의 요건이다 **t1333 판 정정: 소비자는 cli·web·contract 셋이다 — contract 엇키는 앵컰 이전부터 존재했다** |
| `internal/mission` | 2 | 소비자는 `internal/cli`(실제로는 `internal/cli/goal.go` 한 파일)와 — t1278 판 합류 — `internal/contract`(`projection_mission.go` 의 단방향 투영이 import 한다)다 |
| `internal/contract` | 3 | **t1238 판에서 새로 들어왔다.** 비테스트 소비자는 `internal/cli/contract.go` 한 파일과 같은 계열의 `internal/contract/sign`뿐이다. 코어는 표준 라이브러리와 `gopkg.in/yaml.v3` — **t1278 판부터 `internal/mission` 이 그 옆에 더해졌다**(`projection_mission.go` 의 단방향 투영; 더 이상 순수 leaf가 아니다) — 만 import 하며, 부수효과를 지는 `internal/contract/sign`은 코어와 `internal/atomicfile`을 import 한다 — 방향은 sign→core 한쪽뿐이다. `internal/hook`은 Frozen 지시 파일 목록을 테스트에서만 고정하므로 이 칸에 들어오지 않는다 **t1333 판 정정: 비테스트 소비자는 cli·closure·escalation 셋이다 — closure·escalation 엇키는 앵컰 이전부터 존재했다** |
| **`internal/civerdict`** | **2** | **t1278 판에서 새로 들어왔다.** 비테스트 소비자는 `internal/cli`(`ci_verdict.go` — `moai ci-verdict` 저장 경로)와 `internal/escalation`(`ciLimb` 판정)둘이며, 패키지 스스로는 내부 import 0인 순수 leaf다 |
| `internal/codextools` | 0 | 비테스트 소비자 없음(`modules.md` §네거티브 스페이스) |
| `internal/jevmeasure` | 0 | **이 판에서 새로 들어왔고, 0은 설계다 — 그러나 종류가 다른 0이다.** 테스트 시점 가드도 빌드타임 도구도 아니고, 측정 게이트가 실행되지 않은 **게이트 미실행 상태**라 소비자가 원리상 아직 없다. 게이트가 통과하면 소비자가 붙는 것이 이 0의 의미다(`modules.md` §네거티브 스페이스) |
| `internal/harness/rosterguard` | 0 | **t999 판에서 새로 들어왔고, 0이 정상이다.** 테스트 시점 가드라 비테스트 소비자가 원리상 없다 — `internal/template/agentemit` · `commandemit`과 같은 이유이고 `codextools`와는 다른 이유다(`modules.md` §네거티브 스페이스). 형제였던 `cellguard`는 t1246 배치에서 profile-matrix 표면 은퇴와 함께 이 표에서 내려갔다(행동 변화 없음 — fan-in 0이었다) |

`internal/homestate`는 leaf가 아니라 최상위 fan-in **8**의
data/persistence seam입니다. 패키지 단위로 풀면 직접 소비자는 `internal/cli`,
`internal/cli/ptycaptest`, `internal/hook`, `internal/hook/handoff`, `internal/factory`,
`internal/web`, `internal/factorymsg` 등의 표면이며, 이 표면들이 프로젝트 키 경로·Factory 인계·프로필 lease·migration
admission 계약을 공유합니다.

두 방출기(`internal/template/agentemit`, `internal/template/commandemit`)는 이 표에 **나타나지
않습니다** — 비테스트 fan-in이 0이기 때문입니다. 고아가 아니라 빌드타임 도구이며, 소비자가
`make agents-emit` / `make commands-emit` 타깃과 골든 테스트입니다(`modules.md` §네거티브 스페이스).

---

## fan-out 상위 — 다른 최상위 패키지를 import 한 수

| # | 패키지 | import |
|---|---|---|
| 1 | `internal/cli` | 76 |
| 2 | `internal/hook` | 40 |
| 3 | `internal/web` | 16 |
| 4 | `internal/core` | 13 |
| 5 | `internal/escalation` | 12 |
| 6 | `internal/contract` | 10 |
| 7 | `internal/factory` | 9 |
| 8 | `internal/statusline` | 8 |
| 9 | `internal/settings` | 7 |
| 10 | `internal/feedback` | 6 |
| 10 | `internal/homestate` | 6 |
| 12 | `internal/closure` | 5 |
| 12 | `internal/codexwiring` | 5 |
| 12 | `internal/harness` | 5 |
| 15 | `internal/discovery` | 4 |
| 15 | `internal/factorylane` | 4 |
| 15 | `internal/spec` | 4 |
| 15 | `internal/update` | 4 |
| 19 | `internal/config` | 3 |
| 19 | `internal/graph` | 3 |
| 19 | `internal/loop` | 3 |
| 19 | `internal/lsp` | 3 |
| 19 | `internal/profile` | 3 |
| 19 | `internal/ralph` | 3 |
| 19 | `internal/runtime` | 3 |
| 19 | `internal/session` | 3 |
| 19 | `internal/template` | 3 |

`internal/cli`가 다른 최상위 패키지 **76개**를 import 한다(e65b3b의 최상위 집계). 아래 판별 설명과 변화 이력은 기존 구조 분석의 기록이다.
합성 루트(`internal/cli/deps.go`)가 여기 있으므로 일부는 의도된 것이지만, 상당수는
`deps.go`가 아니라 **개별 verb 파일에서 직접** 들어옵니다. 이것이 "명령 하나 = 파일 하나 = 그 명령이
필요한 것 전부 import"라는 수직 슬라이스 성격을 만듭니다.
**t1305 판에서 움직인 행**: `internal/cli` 67→70, `internal/hook` 36→39(커밋 신원 가드·served-model
계열 확장의 누적), `internal/web` 16→15(agentfm 탭 삭제), `internal/core` 12→13, 상위 신규 진입
`internal/escalation` 12·`internal/contract` 9(계약 의사결정이 감지기 루트를 읽는 방향 — § 순환의
새 넷째 쌍), `internal/codexwiring`·`internal/closure` 5. `spec` 행의 4는 새 엣지가 아니라 앞 판의
**정정**이다 — 스탬프 트리에서도 `constitution`을 포함해 4였다.
**t1297 판에서 움직인 행**: `internal/cli` 70→72 — `internal/factorylane`(card t1338 착지)과 세션 앵커 계열(card t1339)이 합류했다. 나머지 상위 행(`internal/hook` 39 포함)은 같은 명령으로 재확인해 변동이 없었습니다.

---

## 순환

e65b3b의 실제 package adjacency에 SCC 분석을 적용한 nontrivial component는 0개다. `go list -deps -json ./...`도 exit0이었다. 상위 디렉터리로 접은 그래프에는 15개 단위 그룹 1개와 contract/escalation의 2개 단위 그룹 1개가 있다. 서로 다른 하위 패키지를 같은 부모로 합친 집계 결과를 Go import cycle로 해석하지 않는다.

직접 양방향으로 연결된 최상위 쌍은 5개다. 아래는 현재 package adjacency에서 확인한 실제 경로다. SCC는 더 긴 경로도 포함하므로 직접 양방향 쌍의 개수와 별개다.

| 최상위 쌍 | 실제 package import의 예 |
|---|---|
| `internal/cli` ↔ `internal/factory` | cli·cli/worktree → factory, factory → cli/specid |
| `internal/cli` ↔ `internal/hook` | cli → hook 및 hook 하위 패키지, hook → cli/preference |
| `internal/contract` ↔ `internal/escalation` | contract/kickoff·contract/revoke → escalation, escalation → contract |
| `internal/hook` ↔ `internal/migration` | hook → migration, migration/migrations → hook |
| `internal/profile` ↔ `internal/settings` | profile → settings/yamlpatch, settings → profile |

15개 단위 그룹은 cli·codexadapter·codexwiring·discovery·factory·factorymsg·feedback·graph·hook·migration·permission·profile·settings·statusline·web이다. 정확한 package 관계는 481쌍의 집합으로 따로 유지하며 최상위 집계와 혼용하지 않는다.

---

## 외부 의존성

현재 `go.mod`의 직접 require는 33개다. 아래 버전은 현재 커밋에서 다시 대조했다. 이전 판의 의존성 불변 설명은 현재 기준으로 적용하지 않는다. 용도와 사용처는 기존 모듈 설명을 유지한다.

| 모듈 | 용도 | 사용처 |
|---|---|---|
| `github.com/gorilla/websocket` v1.5.3 | **t1443 판 신규 직접 의존** — Codex App-Server stream 전송 | `internal/cli/managed_codex_factory.go` (card t1375) |
| `github.com/santhosh-tekuri/jsonschema/v6` v6.0.3 | **t1443 판 indirect→직접 승격** — JSON Schema 검증 | `internal/codextools/registry.go` |
| `github.com/spf13/cobra` v1.10.2 | CLI 명령 트리 | `internal/cli` 전역 |
| `github.com/spf13/pflag` v1.0.10 | cobra 플래그 | 동상 |
| `charm.land/fang/v2` v2.0.1 | cobra 위 help/error/version/completion 렌더러 | `internal/cli/fang.go` |
| `charm.land/bubbletea/v2` v2.0.10 | TUI 이벤트 루프 | `internal/cli/wizard` |
| `charm.land/bubbles/v2` v2.2.1 | TUI 컴포넌트 | 동상 |
| `charm.land/huh/v2` v2.0.3 | 폼/프롬프트 | `internal/cli/wizard` 4개 파일, `internal/cli/ptycaptest/formdriver.go` |
| `charm.land/lipgloss/v2` v2.0.6 | 스타일링 | `internal/tui` |
| `github.com/charmbracelet/lipgloss` v1.1.1-0.20250404203927-76690c660834 | **v2와 병존하는 v1 스타일링** | `internal/statusline`, `internal/cli` |
| `github.com/charmbracelet/glamour` v1.0.0 | 마크다운 터미널 렌더 | `internal/cli/spec_view.go` |
| `github.com/charmbracelet/colorprofile` v0.4.3 | 컬러 프로파일 감지 | tui |
| `github.com/charmbracelet/x/powernap` v0.1.6 | LSP JSON-RPC 전송 | `internal/lsp/transport`, `lsp/core` |
| `github.com/muesli/termenv` v0.16.0 | 터미널 능력 감지 | tui / statusline |
| `github.com/mattn/go-isatty` v0.0.24 | TTY 판별 | 출력 분기 |
| `github.com/mattn/go-runewidth` v0.0.30 | 동아시아 문자폭 계산 | 테이블 / statusline 정렬 |
| `github.com/mark3labs/mcp-go` v1.1.1 | MCP 서버 SDK (stdio 전송) | `internal/cli/mcp_server.go` |
| `github.com/a-h/templ` v0.3.1020 | 타입 세이프 HTML 템플릿 컴파일러 | `internal/web/*.templ` |
| `golang.org/x/net` v0.59.0 | HTML 파싱 | **비테스트 사용처 0 — 테스트 전용** |
| `github.com/smacker/go-tree-sitter` | 16개 언어 AST 심볼 추출 | `internal/navigator/astx`, `internal/hook/mx/complexity` |
| `mvdan.cc/sh/v3` v3.14.1 | 셸 명령 파싱 | `internal/permission/stack.go` — 유일 사용처 |
| `github.com/go-playground/validator/v10` v10.30.5 | 구조체 태그 기반 설정 검증 | `internal/config/validation.go` — 유일 사용처 |
| `github.com/fsnotify/fsnotify` v1.10.1 | 파일 변경 감시 | `internal/web/events.go`, `internal/hook/config_change.go` |
| `golang.org/x/tools` v0.50.0 | Go 패키지/AST 로딩 | `internal/lsp/config` |
| `gopkg.in/yaml.v3` v3.0.1 | 설정·카탈로그·프론트매터 파싱 + **노드 트리 수술**(`internal/settings/yamlpatch`) | 트리 전역 |
| `golang.org/x/sync` v0.23.0 | errgroup 등 동시성 유틸 | 병렬 스캔 경로 |
| `golang.org/x/sys` v0.48.0 | syscall 래퍼 (파일 락, PID 조회) | `*_unix.go` / `*_windows.go` |
| `golang.org/x/term` v0.46.0 | 터미널 제어 | CLI 터미널 경로 |
| `golang.org/x/text` v0.42.0 | 유니코드 / 인코딩 | 정규화 경로 |
| `github.com/stretchr/testify` v1.12.1 | 테스트 단언 | 테스트 전용 |
| `go.uber.org/goleak` v1.3.0 | 고루틴 누수 검출 | `internal/hook` 등 |
| `github.com/google/uuid` v1.6.0 | UUID 생성 | 세션·에이전트 식별자 발급 경로 |
| `modernc.org/sqlite` v1.60.1 | CGO 없는 SQLite 드라이버 | `internal/factory`, `internal/homestate` |

### 이례적인 것

1. **charm 계열 v1 / v2 straddle이 한쪽만 남았습니다.** `huh`는 v2 하나로 정리돼
   `github.com/charmbracelet/huh`(v1)가 `go.mod`에서 사라졌고, v1과 v2를 한 파일에서
   함께 import 하던 테마 파일도 트리에 없습니다.

   > 이전 판이 그 straddle의 증거로 들던 `internal/cli/huh_theme.go`는 삭제된 경로이며
   > 지금 트리에 존재하지 않습니다. 같은 자리를 대신하는 `internal/cli/theme.go`는
   > `internal/tui`에 위임하는 13줄짜리 래퍼로, huh를 전혀 import 하지 않습니다.

   남은 straddle은 `lipgloss` 한 쌍뿐입니다: v2는 `internal/tui` 쪽이고, v1은
   `internal/cli`·`internal/cli/uikit`·`internal/cli/wizard`·`internal/cli/worktree`·
   `internal/cli/agentlint`·`internal/statusline` 여섯 패키지에 걸친 12개 비테스트 파일이
   아직 씁니다. 마이그레이션은 절반이 아니라 한 축이 끝난 상태입니다.
2. **`modernc.org/sqlite`는 별도 require 블록에 있지만 `// indirect`가 아닌 직접 의존성입니다.**
   `internal/factory`과 `internal/homestate`가 드라이버를 직접 씁니다. 순수 Go 구현이라
   `CGO_ENABLED=0`에서 동작하는 것은 CLI 배포에 맞는 선택이지만 `modernc.org/libc` 등이
   딸려 와 바이너리가 커집니다.
3. **`golang.org/x/net`이 direct require인데 비테스트 사용처가 없습니다.** 사용처 6곳이 전부
   `*_test.go`입니다. direct 블록에 있을 이유가 없습니다.
4. **CLI 치고 의외인 조합** — tree-sitter(cgo), sqlite, LSP 클라이언트, HTML 템플릿 컴파일러,
   로컬 HTTP 서버가 한 바이너리에 다 들어 있습니다. 이 도구는 CLI라기보다 개발 환경
   런타임에 가깝습니다.
5. **`github.com/a-h/templ`이 `tool` 지시어로 등록**돼 있습니다(`go.mod:100` — templ CLI를
   가리키는 `tool` 한 줄). `.templ` → `_templ.go` 생성이 빌드 전제이며
   생성물이 트리에 커밋돼 있습니다 — 트리 최대 비테스트 파일 두 개
   (`fieldsets_templ.go` 168KiB, `screens_templ.go` 124KiB)가 그 산물입니다.
7. **`github.com/santhosh-tekuri/jsonschema/v6`는 `go.mod:83`에 `// indirect`로 적혀 있지만
   직접 import 됩니다.** 비테스트 사용처는 `internal/codextools/registry.go` 하나입니다.
   `go mod tidy`라면 직접 require로 옮겼을 모양이고, 그 유일한 사용처인 `internal/codextools`는
   비테스트 fan-in이 0입니다(`modules.md` §네거티브 스페이스). 두 관찰이 같은 패키지에 모입니다.
6. **`gopkg.in/yaml.v3`의 쓰임이 두 축입니다.** 대부분은 마샬/언마샬이지만
   `internal/settings/yamlpatch`는 같은 라이브러리의 **노드 트리**를 직접 수술해 주석과
   미모델링 키를 보존합니다. 이 두 번째 쓰임이 typed struct 재직렬화가 파괴하는 것을
   보존하는 유일한 경로입니다.
