auditor-model: claude-opus-5-5

# t1599 싱크 감사 판정 (sync-auditor)

verdict: FAIL
audited_sha: 270cda84fb997c43ea73dff782c47d9b66d0b53e
required_backend: codex=fail receipt=rcpt-56d38345ae2e668229271b84

## 평가 보고서
카드: t1599 (Class B, SPEC 없음) — 브랜치 `WT-i18n-budget-flake`, 기준 `0bd1da497`
종합 판정: **FAIL** — 필수 백엔드(codex, gate=required)가 `fail`을 냈고, 지시에 따라 필수 백엔드 FAIL은 판정을 FAIL로 강제한다. 수정 코드 자체는 올바르다고 판단하며, 실패 원인은 회귀 테스트의 결정성 부족(F1) 하나다.

### 차원별 점수
| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 90/100 | PASS | 수정 코드: `done` 분기에서 `time.Since(start) > budget`이면 exit 4 + 동일 메시지 반환. 수정 후 `go test -race -count=20 -run 'TestBudget_'` → `ok ... 165.430s` |
| Security (25%) | 95/100 | PASS | 변경 범위는 시간 비교 1줄 + 상수 승격뿐. 새 입력·외부 호출·비밀 없음(grep 결과 기존 `os/exec`/`go/token`만) |
| Craft (20%) | 60/100 | FAIL | 커버리지 `coverage: 86.2% of statements`, vet exit 0, gofmt 무출력. 그러나 RED 테스트가 수정 전 코드에서 확률적으로만 실패(F1) |
| Consistency (15%) | 90/100 | PASS | `gofmt -l` 무출력(exit 0), 기존 주석 양식·메시지 유지 |

### 결함 목록 (structured defect-list)
- **F1** [P2 / 중간] [blocking — 필수 백엔드 codex의 FAIL 근거] `scripts/i18n-validator/main_test.go:530-546` — `TestBudget_OverrunReportedWhenScanFinishesFirst`는 수정 전 코드에서 결정적으로 실패하지 않는다. 타이머 분기가 이기면 수정 전에도 exit 4가 나오므로, 스케줄링에 따라 수정 제거 회귀를 놓친다. 감사자 재현: `GOMAXPROCS=1`에서 수정 전 코드로 5회 중 PASS 3회. 신뢰도 높음. — **필요한 수정**: 타이머 채널을 주입 가능하게 만든다(예: 패키지 변수 `budgetTimer = time.After`). 테스트에서 절대 발화하지 않는 채널로 바꾸고 1ns 예산으로 1회 호출하면, 수정 전에는 결정적으로 exit 0(FAIL), 수정 후에는 결정적으로 exit 4(PASS)가 된다. 20000회 반복 루프는 이 방식으로 대체할 수 있다.
- **F2** [낮음] [optional] `scripts/i18n-validator/main.go:431` — `budgetMsg`가 `--budget` 값과 무관하게 "30s"로 고정. 이번 카드 이전부터 있던 문구(기존 timeout 분기와 동일)이며 이번 변경으로 생긴 결함은 아니다. 신뢰도 높음. — 수정 제안: 필요하면 `fmt.Sprintf`로 실제 예산을 넣는다(별도 카드 감).
- **F3** [정보] [optional] `main.go:451-453` — `done` 분기의 초과 판정 시 위반 목록과 진행 stderr를 버린다. timeout 분기와 같은 동작이라 일관적이다. 신뢰도 중간. — 조치 불필요.
- **F4** [정보] [optional] 감사 초점 (2): 예산 안에 스캔이 끝났지만 수신이 예산 뒤로 밀린 경우 이제 exit 4가 된다. 이 검증기의 예산은 벽시계 예산(`flag.Duration("budget", defaultBudget, "wall-clock budget for full repo scan")`)이므로 측정 시점 기준 초과 보고가 옳은 의미론이라고 판단한다. 수정 전 코드에서도 그 시점에는 타이머 분기가 이미 준비돼 select가 무작위로 고르므로, 결과가 오히려 결정적으로 바뀐 것이다. 30s 기본 예산에서 수신 지연은 마이크로초 단위라 오판 여지는 사실상 없다. — 조치 불필요.

### 권고
- F1만 고치면 된다. 수정 후 동일 감사 범위(F1 델타)로 재감사하면 충분하다.
- Windows 스킵 제거는 정당하다: 수정 후에는 1ns 예산이 `time.Since(start)`에 의해 어느 분기가 이기든 항상 초과로 판정되므로, 스킵 근거(스케줄링 비결정성)가 사라졌다. 다만 Windows 실행은 CI에서 확인해야 한다(Gaps).

## Claim / Evidence / Baseline-attribution / Gaps / Residual-risk

### Claim
1. 수정은 결정적이고 의미상 올바르다: 초과는 어느 select 분기가 이기든 exit 4로 보고되고, 예산 안의 스캔은 이전처럼 0/1을 반환한다.
2. RED 테스트는 수정 전 코드에서 실패한다 — 단, 확률적으로만(F1).
3. 수정 후 테스트는 안정적으로 통과한다.
4. `go vet`, `go test -race`가 통과한다.
5. 필수 백엔드 codex는 FAIL을 냈다.

### Evidence
- 수정 전 코드 오버레이(`git show 45fb4a20a:scripts/i18n-validator/main.go` → `/private/tmp/sxa/oldmain.go`, `go test -overlay`; 오버레이 파일에 `time.Since(start)` 0건 확인):
  - `go test -overlay ... -race -count=5 -run 'TestBudget_OverrunReportedWhenScanFinishesFirst'` → exit=1, 5회 모두 FAIL
    ```
    main_test.go:544: 7/20000 iterations did not report the budget overrun (want exit 4 + 'budget' in stderr)
    main_test.go:544: 11/20000 iterations did not report the budget overrun (want exit 4 + 'budget' in stderr)
    FAIL	github.com/modu-ai/moai-adk/scripts/i18n-validator	2.285s
    ```
  - `go test -overlay ... -count=5 -run TestBudget_OverrunReportedWhenScanFinishesFirst` (race 없음) → exit=1, `--- FAIL` 5건
  - `GOMAXPROCS=1 go test -overlay ... -count=5 -v -run TestBudget_OverrunReportedWhenScanFinishesFirst` → exit=1, 그러나
    ```
    --- FAIL: TestBudget_OverrunReportedWhenScanFinishesFirst (0.28s)
    --- PASS: TestBudget_OverrunReportedWhenScanFinishesFirst (0.01s)
    --- PASS: TestBudget_OverrunReportedWhenScanFinishesFirst (0.02s)
    --- FAIL: TestBudget_OverrunReportedWhenScanFinishesFirst (0.43s)
    --- PASS: TestBudget_OverrunReportedWhenScanFinishesFirst (0.01s)
    ```
    → codex의 F1 주장(GOMAXPROCS=1에서 수정 전 코드가 통과)을 감사자가 독립 재현함(5회 중 3회 PASS).
- 수정 후(HEAD):
  - `go test -race -count=20 -run 'TestBudget_' ./scripts/i18n-validator/` → `ok  	github.com/modu-ai/moai-adk/scripts/i18n-validator	165.430s`, exit=0
  - `go vet ./scripts/i18n-validator/` → `vet=0`
  - `go test -race -count=1 -cover ./scripts/i18n-validator/` → `ok  	github.com/modu-ai/moai-adk/scripts/i18n-validator	10.489s	coverage: 86.2% of statements`, exit 0
  - `gofmt -l scripts/i18n-validator/` → 출력 없음, exit 0
- 감사 계획: `moai verify audit-plan --project-root <tree>` → `"config_status": "ok"`, `"model": "codex"`, `"cross_model_active": true`, `"enforced_required": ["codex"]`
- `mcp__moai__audit_multi` (target=baseBranch, card_id=t1599) 다이제스트:
  - `overall_verdict`: `fail`
  - `gate_unmet`: 결과에 키 없음(빈 값)
  - `plan_source`: `config`
  - per_backend: `backend=codex gate=required verdict=fail`
  - `audit_receipt`: `rcpt-56d38345ae2e668229271b84`
- 결과 검사: `moai verify audit-plan --project-root <tree> --result-file .moai/state/audit-plan-result.json` → `"convergence_check": { "ok": true, "unmet": [] }`, exit 0. (게이트 충족 = 필수 백엔드가 응답함을 뜻하며, 그 응답 자체는 `fail`이다.)

### Baseline-attribution
- 트리: `/Users/goos/MoAI/moai-adk-go/.moai/worktrees/t1599`, HEAD `270cda84fb997c43ea73dff782c47d9b66d0b53e` (`git rev-parse` 실측), 작업 트리 clean.
- 수정 전 비교 기준: `45fb4a20a`의 `main.go`(RED 커밋, 수정 미포함)를 오버레이로 HEAD 테스트와 결합.
- 모든 `go test`는 `TMPDIR=GOTMPDIR=/private/tmp/sxa|sxb|sxc`, darwin, 이번 실행에서 측정.
- 도구 출처(§2.2): `go` 툴체인은 이 트리를 직접 컴파일하므로 지연 없음. `moai` CLI/MCP 서버 빌드는 `ee981a932`로, 이 트리 HEAD의 조상(build_lag 보고됨). 감사 계획·결과 검사 출력은 그 빌드가 판정한 것이다.

### Gaps
- Windows에서 실행하지 않음. 스킵 제거의 정당성은 코드 분석에 근거하며, 교차 플랫폼 확인은 CI 몫.
- 원래의 부하 걸린 ubuntu 러너 플레이크는 재현하지 않음(빈 디렉터리 유사 재현만).
- 저장소 전체 테스트는 실행하지 않음(패키지 범위만).
- 감사 중 Bash 호출 4건이 worktree 격리 가드에 거부됨(복합 명령/정규식 인자). 각각 단순 명령으로 쪼개 재실행했고, 그 결과가 위 Evidence이다. 대체로 인한 측정 손실은 없음. 다만 `-run '^...$'` 앵커 정규식 형태는 가드 때문에 비앵커 형태로 실행함(해당 이름으로 시작하는 테스트는 하나뿐이라 동일 대상).
- `moai` 바이너리가 트리 HEAD보다 오래됨(`ee981a932`) — 감사 계획·수렴 검사가 최신 코드 기준이 아닐 수 있음.
- `/private/tmp/sx*` 아래 원시 출력은 머신 로컬 스크래치이며 인용 대상이 아니다. 판정을 가른 줄은 위에 옮겨 적었다.

### Residual-risk
- F1 미수정 상태에서는, 누군가 `done` 분기 검사를 지워도 단일 코어·저부하 CI에서 회귀 테스트가 초록으로 통과할 수 있다.
- 수정 코드 자체에 대해서는 잔여 위험이 낮다: 1ns 예산은 `time.Since(start)`가 항상 넘으므로 수정 후 두 예산 테스트는 결정적이다. 정상 30s 예산에서는 스캔 후 수신까지의 지연만 경계 오판 요인이며, 이는 벽시계 예산 의미론상 올바른 쪽으로 기운다(F4).

## 반복 이력
- 1회차(이 보고서): FAIL — 필수 백엔드 codex FAIL, 사유 F1.
- 보충(같은 회차): Stop 훅의 영수증 확인 요구로 `mcp__moai__codex_audit`(native, baseBranch)를 추가 호출함 → `verdict: pass`, findings 0건, receipt `rcpt-16bce58f88a1c681e7689d6b`. native 리뷰는 F1을 짚지 않았으나, card 바인딩된 audit_multi(codex required)의 FAIL과 감사자의 독립 재현(GOMAXPROCS=1에서 수정 전 코드 3/5 PASS)이 유효하므로 판정은 FAIL을 유지한다.
