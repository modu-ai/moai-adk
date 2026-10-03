# t1448 판정서 — `/moai:todo --auto` 카드 선택 자율화

- 카드: t1448 (Class C, Tier M) · SPEC-TODO-AUTO-PICK-001
- 브랜치: `WT-todo-auto-pick-autonomy` · 트리: `.moai/worktrees/t1448` · 레인: lane-4
- 측정 대상 HEAD: `a501e1b05` (로컬 develop `1e2151a38` 재흡수 병합 커밋). 이 판정서 커밋이 그 뒤에 온다.
- 이 판정서의 최종 PASS/FAIL 판정은 리더의 몫이다. 아래는 레인이 읽고 잰 증거다.

## 운영자 결정 (원문 그대로)

운영자 결정 2026-10-03: sync-audit iter2 네 차원 87.0 PASS-WITH-DEBT, 필수 codex fail(F3 a/c·F14)은 잔여 위험으로 수용, 후속 카드로 종결

`sync-audit-iter2.md`의 `verdict: FAIL` 줄은 고치지 않았다. 그 보고서가 FAIL인 이유(필수 교차 모델 게이트 미충족)가 그대로 사실이고, 위 결정이 그 위에 얹힌 처분이다.

## 1. 주장 (Claim)

| # | 주장 |
|---|---|
| C1 | `moai factory next --card <id>` 지정 임대가 대상 카드만 임대하거나 닫힌 토큰 12종 중 하나로 거절하며, 거절은 큐·기록을 바꾸지 않는다(`raced` 제외). |
| C2 | 레인 세션은 `moai todo --auto` 직렬 사이클이 거절되고 임대 경로로만 카드를 집는다. 보류·표식·blocked·직렬 점유·이미 소유된 카드는 임대로 거절되고, 확인 의존 카드는 세션 판단으로 건너뛰어 보고한다. |
| C3 | 큐 입장(생산)은 운영자 권한 그대로이고, 레인은 `add/drop/edit/hold` 등 큐 변이를 하지 못한다. |
| C4 | 항상 적재되는 `kanban-dispatch.md` 스텁은 병합 기준 대비 커지지 않았다. |
| C5 | 규칙 문서 라이브↔템플릿 거울본이 일치하고 카탈로그 해시·Codex TOML이 재생성됐다. |
| C6 | 동시 두 레인이 서로 다른 카드를 임대하는 회귀 테스트가 반복 실행에서 결정적으로 통과한다. |
| C7 | develop 재흡수 병합(`a501e1b05`) 뒤에도 위 C1~C6이 유지된다. |

## 2. 증거 (Evidence)

모든 명령은 `.moai/worktrees/t1448`, HEAD `a501e1b05`에서 환경 세척(`unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_FACTORY_ROLE MOAI_FACTORY_WORKER && …`) 한 호출 안에서 이번 실행에 직접 쟀다. 무거운 실행은 `moai slot acquire --resource go-test-cli-t1448` 임대 아래에서 돌렸고 `moai slot release`로 반납해 `moai slot status`에서 사라진 것을 확인했다.

| 측정 | 명령 | 관측 |
|---|---|---|
| 빌드 | `go build ./...` | exit 0 |
| Windows 교차 빌드 | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| 템플릿·가드 패키지 | `go test ./internal/template/... ./internal/guardstate -count=1` | `ok internal/template 138.566s`, `ok internal/template/agentemit 0.452s`, `ok internal/template/commandemit 0.608s`, `ok internal/guardstate 0.284s`, exit 0 |
| 지정 임대·교리·레인 거절 테스트 | `go test ./internal/cli -timeout 25m -run '^(TestFactoryNext\|TestTodoLane\|TestTodoNonLane\|TestFactoryFallback\|TestAutoPick\|TestAutoRank\|TestAutoHelp)' -count=2 -v` | `ok internal/cli 852.174s`, `--- FAIL` 0건, 하위 포함 `--- PASS` 806건, `DATA RACE` 0건 |
| 동시성 테스트 반복 | 위 실행 안의 `TestFactoryNextNominateConcurrentLanes` | 2회 모두 PASS (9.54s, 5.77s). `TestFactoryNextParallelizableConcurrentLeases` 2/2 PASS, `TestFactoryNextNominateCompensateRechecksRecord` 2/2 PASS |
| 레이스 검출 | `go test ./internal/cli -race -timeout 20m -run '^(TestFactoryNextNominateConcurrentLanes\|TestFactoryNextParallelizableConcurrentLeases\|TestFactoryNextNominateCompensateRechecksRecord\|TestFactoryNextNominateBeforeRecord)' -count=3` | `ok internal/cli 63.309s`, `DATA RACE` 0건 |
| 스텁 크기 | `wc -c`/`wc -m` 라이브 `kanban-dispatch.md` vs 병합 기준 `1e2151a38` 블롭 | 28,308 → 28,301 바이트, 28,099 → 28,092 글자(증가 없음). 템플릿 거울 27,979 바이트 / 27,771 글자 |
| 거울 일치 | `cmp` 라이브↔템플릿: `auto-semantics.md`, `gtd.md`, `manager-todo.md`, `moai-kanban-foreman/SKILL.md`, `kanban-dispatch-detail.md`, `moai-mcp-tools-catalogue.md` | 6쌍 모두 차이 없음(exit 0) |
| SPEC 린트 | 트리에서 빌드한 `moai spec lint SPEC-TODO-AUTO-PICK-001` | exit 0, `No findings — all SPEC documents are valid` |
| 카탈로그 해시 | `go run ./internal/template/scripts/gen-catalog-hashes.go --all` | `catalog.yaml updated successfully`, 충돌 표식 0건 |

감사 이력(카드 트리의 `.moai/reports/t1448/`, 폐기 전 hoist 필요):

- plan-audit: 1회차 FAIL 0.73 → 2회차 FAIL 0.77 → 3회차 FAIL 0.80 → 운영자 승인 델타 감사 PASS 0.85 (Tier M 임계 0.80, 상한 2회를 리더가 한 번 더 승인).
- sync-audit 1회차(`sync-audit.md`, 감사 대상 `8de769d81`): FAIL, 네 차원 조화평균 77.2/100(기능성 65 필수통과 미달). 막는 지적은 AC-TAU-002 동시성 테스트가 22회 중 5회만 통과한 것.
- 수리: `e55aaeb1b`(코드·테스트: 테스트를 결정적으로 만들고 보상 재검증을 큐 잠금 안에서 기록 재독으로 바꿈), `9a5cb0dcb`·`01ccc8b33`(교리·문서 범위 한정).
- sync-audit 2회차(`sync-audit-iter2.md`, 감사 대상 `01ccc8b33`): 네 차원 88·90·85·85, 조화평균 87.0/100 PASS-WITH-DEBT, 막는 지적 없음. 보고서 자체 판정 줄은 FAIL로 남음 — 필수 교차 모델 게이트 미충족(codex 백엔드 fail, `plan_source` 부재로 `moai verify audit-plan --result-file`의 `convergence_check.ok=false`; MCP 서버 빌드 `d194083fb`가 그 표면보다 오래됨).

### 2회 흡수 부기

판정서 커밋 뒤 로컬 develop에 카드 t1430의 문서 커밋 2건(`22194a0ad`, 병합 `43f5f85a5`; ultracode 문구, 규칙 `dynamic-workflows.md`와 그 거울, docs-site 9쪽, 10파일)이 더 들어왔다. 충돌 없이 흡수해 HEAD `29a4d56d6`이 됐고, 그 트리에서 다시 쟀다: `go build ./...` ok, `go test ./internal/template/... ./internal/guardstate -count=1 -timeout 25m` → `ok internal/template 378.777s`, `ok agentemit 0.699s`, `ok commandemit 0.628s`, `ok guardstate 0.319s`, exit 0. 위 표의 `internal/cli` 측정은 그 흡수 전 HEAD `a501e1b05`에서 한 것이고, 흡수가 건드린 파일은 `internal/cli`와 무관한 문서·규칙 거울뿐이라 `internal/cli` 테스트는 다시 돌리지 않았다(미검증으로 남긴다).

## 3. 기준선 귀속 (Baseline-attribution)

- 측정한 트리: 카드 브랜치 HEAD `a501e1b05`(병합 후). 병합 기준(공통 조상)은 로컬 `develop` 팁 `1e2151a38` 자체이며, 스텁 크기 비교의 기준 블롭은 그 커밋의 `.claude/rules/moai/workflow/kanban-dispatch.md`다(`git show`로 읽어 28,308 바이트/28,099 글자 확인).
- 이전 흡수 시점(`7109e0900`)의 기준 값(28,308/28,099, 거울 27,986/27,778)은 이번 재흡수 뒤에도 같은 값으로 재현됐다.
- 모든 Go 측정은 이번 실행에서 이 트리에 대해 새로 측정했다. 이전 실행의 수치(`-count=30` 30/30, `-race -count=5`, 와이드 셀렉터 57 pass 등 progress.md §E.2 기록)는 이 판정서에서 인용하지 않는다.

## 4. 미검증 (Gaps)

1. 첫 `internal/cli` 재측정(`-count=3`, 기본 `-timeout`)은 `panic: test timed out after 10m0s`로 끝났다. `--- FAIL`은 없었고 기본 10분 제한을 넘긴 것이다. 같은 선택자를 `-timeout 25m -count=2`로 다시 돌려 통과했으므로 `-count=3` 전체는 관측하지 못했다(동시성 테스트 자체는 `-race -count=3`을 따로 3/3 관측).
2. 이 실행의 슬롯 임대 자원명은 카드명(`go-test-cli-t1448`)이었다. 같은 시각 다른 레인이 `go-test-cli-shared`를 쥐고 있었고 두 자원은 서로를 배제하지 않았다. 부하 경합이 측정 시간(852초)에 섞였을 수 있다. 통과/실패 판정에는 영향이 없다고 보지만 시간은 기준값이 아니다.
3. `moai spec lint`는 이 트리에서 별도로 빌드한 바이너리(재흡수 전 카드 트리 커밋에서 빌드; `a501e1b05`의 조상)로 돌렸다. 재흡수 병합이 SPEC 문서를 건드리지 않았음은 `git diff`로 별도 확인하지 않았다. 판정에 도구 빌드 커밋을 병기해야 하는 규칙(§2.2) 기준으로 이 항목은 부분 귀속이다.
4. `MOAI_GR_BASE` 기준 가드 4종(`TestContractModeInheritedDivergence`, `TestContractModeConstitutionDriftNotIncreased`, `TestContractModeAlwaysLoadedBudget`, `TestContractModeEmitterSites`)은 기준 변수 없이 skip이다. 통과도 적색도 주장하지 않는다.
5. `internal/cli` 패키지 전체 커버리지는 재지 않았다. 바뀐 함수의 테스트만 실행했다.
6. 지정 임대 아래서 레인이 쓰는 결정 기록 한 줄은 이 새 교리 아래서 한 번도 실제로 쓰인 적이 없다(관측 0건).
7. 라이브 `kanban-dispatch.md`와 템플릿 거울에는 기존부터 있던 라이브 전용 `moai worktree sweep …` 문장 차이가 남아 있어 그 한 쌍은 `cmp`로 비교하지 않았다(크기만 쟀다).
8. 교차 모델 감사 게이트는 충족되지 않았다(위 운영자 결정으로 수용). 재감사 PASS는 없다.
9. 감사 후 교정 커밋 `86281f836`과 그 뒤 문서 커밋은 재감사를 받지 않았다(리더 허용 범위의 저비용 교정).

## 5. 잔여 위험 (Residual-risk)

- 큐와 팩토리 기록 사이의 세 창은 운영자 결정으로 수용된 잔여 위험이다: (a) 직렬 카드 둘을 같은 순간 지정하면 둘 다 임대될 수 있는 창, (c) 임대 보상이 운영자의 새 pick과 자기 승격을 구별하지 못하는 창, F14 `factoryEnsureCardWorktree`의 `git branch -m` 동시 충돌(임대는 성공했는데 동사가 오류로 끝날 수 있음). 후속 카드(큐와 팩토리 기록에 걸친 원자적 임대, t1458)가 닫는다. 이 카드는 닫지 않는다.
- 보류 상태는 구조 상태 `hold`(`moai gtd hold`만 기록)와 본문이 `[보류`로 시작하는 표식 둘 다로 식별한다. 레인은 둘 중 어느 것도 쓰지 못한다. t810(`picked`), t1294·t1383(`queued`)은 현재 둘 다 없으므로 운영자나 리더가 보류하거나 표식을 달기 전까지 자율 레인이 고를 수 있다. 이 레인은 큐를 건드리지 않았다.
- `kanban-dispatch-detail.md`는 이미 40,000자 예산 초과다(develop 42,675자 → 이 카드 이후 43,138자, +463자). 분할 후속 카드가 리더 몫으로 남아 있다.
- `moai todo --help`의 `--auto` 문구는 여전히 "큐와 그 밖에는 아무것도"라고 배치 승인을 서술한다(후속 한계 e).
- t1407(시리얼 슬롯 도우미 변경)은 아직 develop에 착지하지 않았다. 추출한 `factorySerialInFlightExcluding(cards, classOf, cardID)`는 t1407 변경이 먼저 들어오면 그쪽 줄을 옮겨 담아야 한다. lane-11은 지정 임대 경로(`factoryNextValidate`)를 t1407 범위에서 빼고 t1458로 넘긴다고 알려 왔고, 이 읽기와 일치한다(지정 임대 경로는 `ignoreAssigned=false`).
- t1399(이름 변경)가 먼저 착지하면 이 카드 트리에서 이름 변경 스크립트·검색·재빌드가 필요하다.

## 보류 식별 한 줄 (리더 보고용)

이 SPEC은 보류 집합을 본문 표식(`[보류`로 시작하는 본문)과 구조 상태(`hold`, `moai gtd hold`만 기록) 둘 다로 식별한다.

## 후속 · 정리 목록

1. 큐·팩토리 기록 원자적 임대 카드(F3 a/c, F14 `git branch -m` 충돌) — 리더 발행, t1458.
2. `kanban-dispatch-detail.md` 40,000자 예산 분할 카드.
3. t810·t1294·t1383 보류 또는 표식 처분 — 운영자·리더.
4. 카드 트리에만 있는 보고서(`plan-audit-iter1.md`~`plan-audit-iter4-delta.md`, `sync-audit.md`, `sync-audit-iter2.md`, `wait-record.txt`)는 트리 폐기 전에 primary로 반출해야 한다.
5. 형식 격리로 만들어져 비어 있는 `.claude/worktrees/agent-a1738e79aac1cb93a`(HEAD `7109e0900`)는 레인이 처분하지 않는다.
