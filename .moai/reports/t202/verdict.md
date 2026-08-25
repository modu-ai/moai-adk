# t202 — 칸반 리드 세션 이름의 런타임 등록 (Fixes #1596)

- 카드: t202 (Class C — 설계 방향 리드 확정: (b) 기본 + (a) 병행, (c) 범위 외)
- 브랜치: `WT-lead-name-register` (base `origin/main` = `28bde4022`)
- 이슈: #1596

## 1. Claim

`moai cc -k` / `moai glm -k` (및 factory `-f`) 로 띄운 **리드 세션이 세션 목록에 자기 이름으로 표시**된다. 표시 이름은 그 세션이 실제로 띄워진 이름 — 운영자가 직접 준 이름이 있으면 그 이름, 없으면 bare 또는 bump된 역할명(`lead`, `lead-1`) — 과 항상 같은 문자열이다. 나중에 `/rename` 하면 그쪽이 이긴다.

## 2. Evidence

### 2.1 근본 원인 — 카드 전제의 정정

카드와 이슈는 "선언만 하고 런타임에 등록 안 함"으로 적었으나, **`--name` 주입은 이미 존재했고 신고된 빌드에도 들어 있었다.**

```
$ git log --oneline -1 -S'func appendLeadName' -- internal/cli/kanban.go
c326eb4e0 feat(kanban): name the lead session by its bare role, not lead-<run-id>

$ git merge-base --is-ancestor c326eb4e0 4b2f203fe && echo "IN 3.1.2 (issue's build)"
IN 3.1.2 (issue's build)
```

즉 이슈 본문의 "`/rename` 이 사실상 유일한 등록 수단"은 **이름(name)** 에는 해당하지 않는다. 이슈 본문이 스스로 적은 "메시징·위임 정상"이 그 증거다 — 메시징 주소는 정상 등록돼 있었다.

실제 결함은 **이름과 제목이 서로 다른 두 개의 등록**이라는 데 있다. `--name` 은 메시징 주소만 등록하고, 세션 목록에 뜨는 **제목**은 별도 기록이다. 제목을 아무도 정하지 않으면 `UserPromptSubmit` 훅의 기존 정책이 그 자리를 채우는데, 그 정책의 2순위가 `detectActiveSpec` — 프로젝트에서 가장 최근 수정된 `spec.md` 의 제목이다.

`internal/hook/user_prompt_submit.go` `buildSessionTitle` (수정 전 순서):

1. first-wins 가드 (ai-title / custom-title 있으면 `""`)
2. **활성 SPEC → `"SPEC-ID: heading"`**  ← 리드 세션이 여기 걸렸다
3. 첫 사용자 프롬프트에서 파생
4. `""`

이슈가 관측한 `"SPEC-MIGRATE-002: — 뷰어 프런트엔드 이식"` 은 git 로그가 아니라 **이 2번 분기의 출력 형태**와 정확히 일치한다.

### 2.2 수정

| 파일 | 변경 |
|---|---|
| `internal/config/envkeys.go` | `EnvMoaiKanbanLeadName = "MOAI_KANBAN_LEAD_NAME"` 신설 — 리드의 **해소된** 세션 이름을 런처 → 훅으로 전달 |
| `internal/cli/kanban.go` | `appendLeadName` 이 `([]string, string)` 반환 (argv + 해소된 이름). 운영자가 이름을 준 경우에도 그 이름을 **보고**한다. `exportLeadSessionName(name) func()` 추가 — 기존 `captureEnvState` + 복원 idiom 준수 |
| `internal/cli/cc.go`, `internal/cli/glm.go` | 리드 분기 4곳(kanban×2, factory×2)에서 `defer exportLeadSessionName(leadName)()` |
| `internal/hook/user_prompt_submit.go` | `buildSessionTitle` 에 리드 분기 추가 — **first-wins 가드 아래, SPEC 분기 위**. `leadSessionTitle()` 헬퍼가 env를 `TrimSpace` 해서 읽음 |
| `internal/hook/session_start_kanban_i18n.go`, `session_start_factory_i18n.go` | (a) 안내문 갱신, 4-locale(en/ko/ja/zh) 전부 — 목록에도 같은 이름이 뜨고, 제목은 **첫 프롬프트에** 등록되며, `/rename` 이 우선한다는 사실 명시 |

분기 위치가 설계의 핵심이다.

- **SPEC 분기 위**: 리드가 SPEC 있는 프로젝트에 앉아 있는 상황이 곧 이 결함의 발생 조건이다.
- **first-wins 가드 아래**: `/rename` 은 다른 모든 출처를 이기듯 이 분기도 이긴다. 매 턴 제목을 다시 쓰는 회귀(#1198)를 되살리지 않는다.

역할명을 훅에서 다시 유추하지 않고 런처가 해소한 값을 그대로 실어 보내는 이유: bump된 리드(`lead-1`)에서 제목과 실제 주소가 갈리기 때문이다.

### 2.3 검증

```
$ go build ./...                                     (무출력)
$ gofmt -l <touched 9 files>                         (무출력)
$ golangci-lint run ./internal/hook/... ./internal/cli/... ./internal/config/...
0 issues.
$ go test ./internal/hook/ ./internal/config/ -count=1
ok  github.com/modu-ai/moai-adk/internal/hook    23.510s
ok  github.com/modu-ai/moai-adk/internal/config   2.302s
$ go test ./internal/cli/ -count=1 -timeout 600s
ok  github.com/modu-ai/moai-adk/internal/cli    344.867s
```

신규 회귀 테스트 `TestBuildSessionTitle_LeadNameWinsOverSPEC` 5개 서브테스트 전부 PASS:

```
--- PASS: .../not_a_lead_->_SPEC_title_(unchanged_default)
--- PASS: .../lead_->_the_session's_own_name,_not_the_SPEC
--- PASS: .../bumped_lead_->_the_bumped_name
--- PASS: .../existing_title_wins_over_the_lead_name
--- PASS: .../whitespace-only_value_->_falls_through_to_the_SPEC_title
```

첫 서브테스트는 **음성 대조군**이다 — 변수가 없을 때 SPEC 제목이 그대로 나오는 것을 먼저 확인하지 않으면, 양성 케이스가 "모든 세션에 이름을 붙이는" 잘못된 구현으로도 통과한다.

기존 `TestAppendLeadName_OperatorNameWins` 은 반환값 변경에 맞춰 갱신하고, 운영자 이름이 **보고되는지**를 추가로 단언한다.

## 3. Baseline-attribution

- 기준 트리: `WT-lead-name-register`, base `origin/main` = `28bde4022`
- 모든 수치는 이 트리에서 이 라운드에 실행한 명령의 출력이다. 캐시 비활성(`-count=1`).

## 4. Gaps (미검증)

- **실제 세션에서 제목이 바뀌는지 육안 확인 안 함.** 검증은 `buildSessionTitle` 의 반환값까지다 — 훅이 반환한 `SessionTitle` 을 Claude Code가 실제로 세션 목록에 반영하는 구간은 런타임 소관이고, 이 트리에서 관측하지 않았다. 기존 SPEC 제목 기능이 같은 경로로 동작해 왔다는 것이 유일한 간접 근거다.
- **동반 세션(plan/run/sync)과 factory lane의 제목은 고치지 않았다.** 동일한 결함이 그대로 남아 있다 — 이름은 `--name` 으로 등록되지만 목록 제목은 여전히 SPEC에서 온다. 카드 범위가 리드였고, `MOAI_KANBAN_LABEL` 이 이미 해소된 라벨을 들고 있어 분기 하나면 되지만 범위 밖이라 손대지 않았다. **후속 카드 후보.**
- **(c) `moai doctor` 진단 항목은 범위 외** — 리드 지시.
- 전체 스위트(`go test ./...`)는 로컬에서 돌리지 않았다(CLAUDE.local.md §4). 전 패키지 판정은 CI 몫.
- Windows/Linux 매트릭스 미검증 — CI 몫. env 전달과 `TrimSpace` 뿐이라 플랫폼 의존은 없어 보이나 관측한 바 없다.

## 5. Residual-risk

- **제목은 첫 프롬프트에 등록된다, 런치 시점이 아니다.** `UserPromptSubmit` 훅이 그 지점이라 구조적이다. 운영자가 첫 프롬프트를 보내기 전에 세션 목록을 보면 여전히 이름이 없다. 안내문(a)에 이 사실을 명시한 이유다.
- **훅이 꺼진 환경(`disableAllHooks`)에서는 제목이 등록되지 않는다.** 메시징 이름은 `--name` 이라 영향 없다.
- `os.Setenv` 는 프로세스 전역이다. 신규 테스트는 그래서 병렬이 아니며, 런처 쪽은 기존 `captureEnvState` 복원 규율(`@MX:ANCHOR` on `enterKanbanMode`)을 그대로 따른다. 복원 누락은 같은 바이너리의 뒤따르는 테스트를 오염시킨다.
- 운영자가 리드에 자기 이름을 준 경우 그 이름이 제목이 된다. 의도된 동작이지만, 종전에는 SPEC 제목이 나왔으므로 그 운영자에게는 **행동 변화**다.
