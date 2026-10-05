# t556 — goal 조건 분류: 산문이 mechanical 로 떨어지는 잔여 구멍 폐쇄

카드: t556 (GH #1660) · 브랜치 `WT-goal-cond-classify` · base 로컬 develop `d060e0d13`

## Claim

1. 제보 #1660 의 **원래 증상은 이미 닫혀 있었다.** 선행 카드 t288(`c04188cf8`)이 `conversation` 참조어를 추가했고, t436(`522116b1d`)이 arm 시점 실행가능성 관문과 `model:`/`cmd:` 선언 접두를 넣었다. 카드 문안의 전제("MCP 래퍼가 분류를 거치지 않는다")는 현재 트리에서 거짓이다 — 두 arm 표면 모두 같은 `parseCondition` 을 탄다.
2. **결함 계열 자체는 살아남아 있었다.** 첫 단어가 우연히 실제 명령 이름인 산문(`make …`, `test …`, `go …`)은 t436 관문을 그대로 통과해 mechanical 로 무장되고, 매 턴엔드를 천장까지 막는다.
3. 그 잔여 구멍을 **arm 시점 산문-형태 판별식**으로 닫았다. 두 arm 표면(CLI `goal arm`, MCP `goal_arm`)이 이제 단일 함수 `armTimeConditionGate` 를 공유한다.
4. 대조군은 불변이다 — 실제 명령, `cmd:`/`model:` 선언, 기존 참조어 산문의 분류·무장 결과가 수리 전후로 동일하다.

## Evidence

### E1 — 수리 전 분류 실측 (계기: `armTimeConditionGate` 경유)

| 입력 | 수리 전 | 수리 후 |
|---|---|---|
| `go test ./...` | mechanical / armed | mechanical / armed |
| `go test ./... exits 0` | mechanical / armed | mechanical / armed |
| `true` | mechanical / armed | mechanical / armed |
| `cmd: go test ./...` | mechanical / armed | mechanical / armed |
| `…in the transcript` | model / armed | model / armed |
| `…in the conversation` | model / armed | model / armed |
| `model: every AC row is marked PASS` | model / armed | model / armed |
| `모든 AC가 PASS로 표시된다` | mechanical / REFUSED | mechanical / REFUSED |
| `すべてのACがPASSになる` | mechanical / REFUSED | mechanical / REFUSED |
| `all blocking acceptance criteria are satisfied` | mechanical / REFUSED | mechanical / REFUSED |
| **`make sure every AC row is marked PASS`** | **mechanical / armed** | **REFUSED** |
| **`test coverage reaches 85 percent for every package`** | **mechanical / armed** | **REFUSED** |
| **`go through each SPEC and confirm the fix landed`** | **mechanical / armed** | **REFUSED** |

대조군 7건 전부 불변, 산문 3건이 armed → REFUSED. 측정 하니스는 `internal/cli/zz_t556_measure_test.go`(임시, 커밋하지 않음)로 두 시점 모두 같은 코드에서 돌렸다.

계기 결함 1건을 도중에 잡았다: 최초 측정 하니스가 구 판별식을 인라인으로 복제해 새 관문에 **도달하지 못했다.** 고친 뒤 재측정한 값이 위 표다.

### E2 — 뮤턴트 가드 (공허한 초록 배제)

`if proseShapedCommand(cond.Cmd) {` → `if false && proseShapedCommand(cond.Cmd) {` 로 되돌린 상태에서:

```
--- FAIL: TestGoalArm_ProseShapedConditionRefused
    prose condition was armed (out=armed goal for session T556PROSE (mechanical condition, ceiling 30 turns): make sure every AC row is marked PASS)
--- FAIL: TestMCPGoalArm_ProseShapedConditionRefused
    MCP goal_arm armed a prose condition: [{... text goal_arm: ok}]
    MCP refusal wrote a state file: &{SessionID:T556MCP ... Conditions:[{Type:mechanical Cmd:make sure every AC row is marked PASS ...}] ...}
```

두 표면 테스트가 모두 빨갛다 — 실행되고 있고, 판정이 실제 코드에 걸려 있다.

### E3 — 실제 바이너리 E2E (격리된 임시 프로젝트, 리드 세션 goal 상태 무접촉)

```
$ CLAUDE_PROJECT_DIR=<tmp> ./bin/moai goal arm "make sure every AC row is marked PASS" --session T556E2E
ERROR  ... was classified as a mechanical condition and would be run as a shell command,
       but it reads as a sentence rather than a command — ...
       Declare which one you meant: goal arm "model: ..." ... or goal arm "cmd: ..."
PROSE_EXIT=1

$ ... ./bin/moai goal arm "go test ./internal/goal/... exits 0" --session T556E2ECTRL
armed goal for session T556E2ECTRL (mechanical condition, ceiling 30 turns): go test ./internal/goal/... exits 0
CTRL_EXIT=0

$ ... ./bin/moai goal arm "model: every AC row is marked PASS" --session T556E2EMODEL
armed goal for session T556E2EMODEL (model condition, ceiling 30 turns): model: every AC row is marked PASS
MODEL_EXIT=0

$ ls <tmp>/.moai/state/goal/
T556E2ECTRL.json          # 거절된 arm 은 상태 파일을 쓰지 않았다
```

### E4 — 검증 배치 (건드린 패키지 범위)

| 검사 | 명령 | 결과 |
|---|---|---|
| 단위·통합 | `go test ./internal/cli/... -count=1` | exit 0 |
| goal 엔진 | `go test ./internal/goal/... -count=1` | `ok … 1.334s` |
| vet | `go vet ./internal/cli/` | exit 0 |
| lint | `golangci-lint run ./internal/cli/...` | `0 issues.` |
| gofmt | `gofmt -l internal/cli/` | 출력 없음 |
| 크로스플랫폼 | `GOOS=windows GOARCH=amd64 go build ./...` | exit 0 |
| 템플릿 임베드 | `make build` | 성공, `catalog.yaml` 갱신(12899 bytes) |

## Baseline-attribution

전부 이 트리(`.claude/worktrees/t556`), base 로컬 develop `d060e0d13` 에서 이번 실행에 측정했다. 수리 전 값은 base 커밋 상태에서, 수리 후 값은 작업 트리 상태에서 같은 하니스로 쟀다.

## 설계 판단 — 왜 model 재분류가 아니라 거절인가

카드의 [HARD] 는 "애매하면 mechanical 로 떨어뜨리지 말라, model 이 안전하다" 이다. 두 선택지를 놓고 거절을 골랐다.

- **model 로 조용히 재분류**하면, 오타 난 진짜 명령(`got test ./...`)이 전사 판정 주장으로 바뀐다. 테스트가 한 번도 돌지 않았는데 목표가 수렴했다고 선언되는 **거짓 수렴**이며, 아무 신호도 남기지 않는다.
- **거절**은 아무것도 무장하지 않으므로 세션을 막지 않는다 — 카드가 막으려는 실패 방향(mechanical 오분류 → 매 턴 차단)을 그대로 피한다. 게다가 바로 옆 t436 관문이 이미 채택한 형태라 파일 안의 일관성을 지킨다. 되돌릴 길도 메시지가 직접 알려준다(`model:` / `cmd:`).

즉 [HARD] 를 문자 그대로도(mechanical 로 떨어지지 않는다) 취지대로도(세션이 막히지 않는다) 만족하면서, model 재분류가 새로 여는 거짓 수렴 위험을 열지 않는다.

## Gaps — 관측하지 않은 것

- 전체 스위트(`go test ./...`)를 로컬에서 돌리지 않았다(CLAUDE.local.md §4 금지). 전 패키지 판정은 CI 몫이다.
- `internal/hook/stop-goal` 평가자 경로는 이번에 건드리지 않았고, 실행도 하지 않았다. 이 카드는 arm 시점만 바꾼다.
- 판별식의 임계값 두 개(`proseWordFloor = 5`, `proseBareRatio = 0.8`)는 **유도값이 아니라 조율값**이다. 테스트의 실제 명령 9종·산문 6종으로 고정했을 뿐, 더 넓은 명령 코퍼스로 검증하지 않았다.
- ja/zh 로케일 산문은 t436 관문이 먼저 잡아서 새 판별식까지 도달하지 않는다 — 새 판별식이 그 로케일에서도 동작한다는 것은 단위 테스트(`non-english` 행)로만 봤고, arm 표면에서는 보지 않았다.

## Residual-risk

- **거짓 거절**: 5단어 이상이면서 플래그·경로·파이프가 하나도 없는 진짜 명령(`npm run test coverage report` 류)은 산문으로 판정돼 거절된다. 회복 가능하고(`cmd:` 접두) 메시지가 그 길을 명시하지만, 마찰은 실재한다.
- **거짓 통과**: 경로나 플래그를 여럿 품은 긴 산문은 비율 문턱을 넘지 못해 통과한다. 이때는 평가 시점 exit-127 백스톱이 첫 턴에 잡는 형태로 떨어진다(천장까지 가지 않는다).
- 판별식이 **구문 증거**에만 의존하므로, 셸 구문과 자연어가 겹치는 구간에서는 원리적으로 애매하다. 명시 접두(`model:` / `cmd:`)가 그 구간의 유일한 확정 수단이며, 문서에 그렇게 적었다.

## 후속 제안 (이 카드에서 하지 않음)

- `parseCondition` 의 기본 방향 자체를 뒤집는 안(무접두·무증거 → model)은 **채택하지 않았다.** 문서화된 맨 명령 형태(`/moai goal "go test ./... exits 0"`)를 전부 model 로 만들어버려 회귀가 크다. 별도 판단이 필요하면 카드로 분리할 것.
- 참조어 허용목록(`transcript`/`conversation`)은 여전히 영어 전용이다. 이번 판별식이 그 뒤를 받치지만, 허용목록 자체는 손대지 않았다.

## 병합 트리 재측정 — 통합 창 안 (lane-6)

리드 지명 후 창을 잡고(`moai integration acquire --name lane-6`) 로컬 develop 을 흡수한 트리에서 다시 쟀다. 흡수 전 측정은 근거로 재사용하지 않는다.

**흡수 대상의 최신성.** `git fetch origin develop` 후 `git rev-list --count --left-right origin/develop...develop` → `0	37`. 원격에만 있는 커밋은 없고 로컬이 리드 일괄 push 대기분만큼 앞서 있으므로, 흡수 대상은 로컬 develop `030987513` 이다.

**흡수.** `git merge --no-edit develop` → HEAD `7d4ef0560`, 트리 `ec49cc856`. 흡수 전에 `git merge-tree --write-tree --name-only develop HEAD` 가 예측한 트리 `ec49cc856` 과 같다(충돌 파일 0). `git merge-base --is-ancestor 030987513 HEAD` → rc 0.

**범위 — 영향 범위만.** 기준 이후 develop 이 바꾼 파일과 이 카드가 바꾼 파일은 겹치지 않는다. 겹치는 것은 패키지 둘이다.

| 공유 패키지 | develop 쪽 | 이 카드 쪽 | 잰 것 |
|---|---|---|---|
| `internal/cli` | `update/merge` 테스트 2본, 템플릿 embed 전이(t609) | `goal.go`, `goal_runnable.go`, `mcp_server.go` | 이 카드의 테스트 4개 |
| `internal/template` | `settings.json.tmpl` | `catalog.yaml` 의 `moai` 스킬 해시 | catalog 계열 테스트 |

| 명령 | 결과 | 증거 |
|---|---|---|
| `go test ./internal/cli/ -count=1 -v -run 'TestProseShapedCommand\|TestGoalArm_ProseShape\|TestMCPGoalArm_ProseShape'` | rc 0, `--- PASS` 4건 이름 확인, `ok … 0.971s` | `merge-tree-goal-tests.txt` |
| `go test ./internal/goal/... -count=1` | rc 0, `ok … 0.355s` | `merge-tree-goal-pkg.txt` |
| `go test ./internal/template/ -count=1 -v -run 'TestCatalog\|TestAllSkillsInCatalog\|TestLoadCatalog'` | rc 0, `--- PASS` 10건(`TestCatalogHashCoversSkillSubfiles` 포함), `ok … 0.460s` | `merge-tree-catalog.txt` |

통과 줄은 개수만이 아니라 테스트 이름으로 읽었다 — `-run` 이 0건과 일치해도 `ok` 가 찍히기 때문이다.

**이 절의 Gaps.** `internal/cli` 패키지 전체 판정은 리드가 일괄 push 직전에 한 번 도는 몫이라 여기서 돌리지 않았다. `internal/template` 도 catalog 계열만 쟀고 패키지 전체는 돌리지 않았다. 앞 절에 기록된 선재 실패 4건(`TestHomeState…` 계열)은 이 트리에서 다시 재지 않았다.
