auditor-model: claude-opus-5-5
verdict: PASS
score: 0.92
audited_sha: 47963571a1cee3b9643ed9baba1a66d534a0e737
receipt: rcpt-c671fc729913939c39b1da42

# t2123 sync 감사 판정 — #1969 Windows race 잔여, 조각 ① (`/bin/sh` 부재 19건)

- 카드: t2123 (t2119 분할), 브랜치 `WT-windows-race-bin`, 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2123`
- 감사 범위: 커밋 `47963571a1` 하나 (`git rev-parse HEAD` → `47963571a1cee3b9643ed9baba1a66d534a0e737`), 부모 `04d8c27a7e`
- 변경 파일 4개: `internal/factorymsg/factory_run_retire_test.go`, `internal/cli/factory_runs_command_test.go`, `internal/agenticfactory/effect/internal/procrun/main_test.go`(신규), `.moai/reports/t2123/progress.md`. 시험 외 Go 파일 변경 0건(`git diff --name-only 04d8c27a7e 47963571a1 -- '*.go' ':!*_test.go'` 출력 없음).

## 차원별 점수

| 차원 | 점수 | 판정 | 근거 |
|---|---|---|---|
| Functionality (40%) | 94/100 | PASS | CI 실패 19건 전부가 이번에 바꾼 두 도우미와 procrun 패키지로 귀속됨(아래 증거 E1). 도우미 호출 시험 10개 함수 전부 통과(E2~E4), 도우미가 만든 pid가 실제로 dead로 분류됨(E5), 도우미를 live로 바꾼 변이체는 시험 4개를 실패시킴(E6) |
| Security (25%) | 100/100 | PASS | 시험 코드만 바뀜. 비밀값 탐침 0건(E8). 자식 프로세스는 자기 시험 바이너리이며 외부 입력이 없음 |
| Craft (20%) | 86/100 | PASS | gofmt 깨끗, Windows 교차 vet 깨끗(E7). 선택 지적 F1(cli 자식 실행 비용), F2(패키지 단위 생략의 가시성), F4(`os.Args[0]` 상대 경로) |
| Consistency (15%) | 92/100 | PASS | `exec.Command(os.Args[0], "-test.run=^$")`는 저장소 선례가 있음(`internal/config/toolpolicy/tier_render_test.go:44`, `:752`, `internal/factorymsg/store_path_context_test.go:95`). 주석 형식과 카드 표기가 주변 파일과 맞음 |

필수 통과 차원(Functionality, Security)은 각각 기준을 넘었다. 차단 지적 0건.

## 지적 목록 (구조화 결함 목록)

- F1 [Low, 확신 높음] [optional] [in-diff] `internal/cli/factory_runs_command_test.go:20` — cli 자식 프로세스는 `-test.run=^$`여도 cli `TestMain` 전체(홈·프로필·git config 샌드박스, gitenv 정리, 명령 트리 예열, config 섹션 해시)를 돈다. 측정값은 호출당 0.60~1.15초(factorymsg는 약 22ms). 호출 지점 6곳이라 패키지 전체에 몇 초가 붙는다. 자식은 부모 샌드박스를 물려받고 자기 임시 디렉터리(`sandboxAuditReceiptFallbackRoot`)는 정상 종료 경로에서 지우므로 잔여물·재귀는 없다(코드 확인, `internal/cli/main_test.go:372-467`, `:535-571`). — 필요한 수정: 없음. Windows race에서 시간이 문제가 되면 그때 가벼운 자식 경로를 검토.
- F2 [Low, 확신 중간] [optional] [in-diff] `internal/agenticfactory/effect/internal/procrun/main_test.go:16-19` — `m.Run()` 전에 `os.Exit(0)`하므로 `go test -json` 스트림에 시험별 skip 이벤트가 없다. CI 집계에는 "시험 3개 생략"이 아니라 "시험 0개 통과"로 보인다. 앞으로 셸이 필요 없는 시험이 이 패키지에 들어와도 Windows에서 함께 빠진다. 지금은 시험 3개가 모두 공용 `sh()`(`procrun_test.go:14-16`)를 쓰므로 판단 자체는 맞다. — 필요한 수정: 없음. 원하면 `sh()`가 `t`를 받아 `t.Skip`하도록 바꾸면 skip이 보인다.
- F3 [Info] [optional] [touched-file] `.moai/reports/t2123/progress.md` Gaps 둘째 항목 — "cli의 다른 `/bin/sh` 참조가 11건에 드는지 개별 확인하지 않았다"는 공백은 이 감사에서 닫혔다. 분류표에서 cli 11행·factorymsg 5행 모두 메시지가 `start short-lived process: …`이고 시험 이름이 바뀐 도우미의 호출자다(E1). — 필요한 수정: 없음(기록용).
- F4 [Low, 확신 중간] [optional] [in-diff] 두 도우미의 `os.Args[0]` — `go test`는 절대 경로로 바이너리를 띄우므로 문제없다(E2~E5 통과로 확인). 다만 `go test -c` 후 `./cli.test`처럼 상대 경로로 직접 띄우면, 도우미 전에 `t.Chdir(root)`하는 시험(`factory_runs_command_test.go:52`)에서 자식 실행이 실패한다. — 필요한 수정: 없음. 원하면 `os.Executable()`로 바꿀 수 있다.
- O1 [Info] [outside] procrun을 감싸는 상위 래퍼 패키지들도 가짜 gh/git 스크립트로 시험한다(`procrun_test.go:11-12` 주석). Windows 실패 분류의 다른 묶음에 들어 있을 수 있으며 이번 조각 범위 밖이다.

## 조각 ①만 착지하는 것의 판단, 그리고 남은 조각 분할

- 조각 ①만 내는 것은 받아들일 수 있다. 원인 하나(실패 19건, 같은 오류 문장)를 닫는 독립 변경이고, ②③④는 progress.md의 Claim·Gaps에 미착수로 적혀 있다. 단, 카드 t2123을 "#1969 Windows race 잔여 전체 완료"로 닫으면 안 되고 조각 ① 완료로만 기록해야 한다.
- 남은 조각은 더 나누는 편이 맞다.
  - ② POSIX 권한 비트(46건): 패키지 묶음으로 나눈다 — mission(21) 1장, cli(11) 1장, 나머지(homestate·factory·graph·yamlpatch·civerdict·cli/update·template·web, 14건) 1장. 셋 다 "Windows에서 0600 기대를 어떻게 다룰지"라는 같은 결정을 먼저 한 번 내리고 들어가야 한다.
  - ③ home guard TEMP(54건): 원인 문장이 하나라 카드 1장이 맞다. Windows 재현이 먼저다.
  - ④ pre-commit e2e(12건) 1장, "기타" 약 200건은 분류 카드 1장을 먼저 두고 분류 결과로 다시 나눈다.
  - SQLite URI(474건)·디렉터리 fsync(326건)는 이 카드 목록에 없으며 분석 문서가 별도 카드로 지목했다.

## 증거 (Claim / Evidence)

E1. 실패 19건의 귀속 — 분석 문서(`/Users/goos/.moai/worktrees/projects/moai-adk-go/t2077/.moai/reports/pr-sweep-20261106/pr1969-ci-analysis.md`)가 가리키는 분류표 `categorized.tsv`에서 `bin/sh` 행을 그대로 뽑음:

```
cli1 BIN_SH_MISSING internal/cli TestFactoryRunsCommandReportsAndRefuses  factory_runs_command_test.go:57: start short-lived process: exec: "/bin/sh": executable file not found in %PATH%
cli1 … TestFactoryRunsRetireRebindTo/{a,b,c,d,g,h/live,h/indeterminate}  factory_runs_rebind_test.go:{43,69,85,104,116,137,137}: start short-lived process: …
cli1 … TestFactoryRunsRetireRefusesInProgressCards  factory_runs_retire_guard_test.go:120: start short-lived process: …
cli1 … TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun  factory_runs_retire_guard_test.go:174: start short-lived process: …
cli3 … TestRelaunchRetireFromRunOpensMoaiPathsRoot  factory_root_alignment_test.go:221: start short-lived process: …
rem  … procrun TestStdoutAndStderrAreSeparateAndTheExitCodeIsAnError  procrun_test.go:21: out="" err=exec: "/bin/sh": executable file not found in %PATH%
rem  … procrun TestTheLimitIsRequiredAndAStreamPastItStopsTheProcess  procrun_test.go:39: … "/bin/sh": executable file not fou…
rem  … procrun TestTheContextEndsTheProcessAndTheArgumentsAndEnvironmentArriveAsGiven  procrun_test.go:54: err=exec: "/bin/sh": …
rem  … factorymsg TestPartialStampWithBrokerStaysIndeterminate  factory_run_boot_proof_legs_test.go:112: start short-lived process: …
rem  … factorymsg TestResolveActiveRunRetiresDeadOwnerAndJoins  factory_run_retire_test.go:103: start short-lived process: …
rem  … factorymsg TestResolveActiveRunBothOwnersDead  factory_run_retire_test.go:128: start short-lived process: …
rem  … factorymsg TestResolveActiveRunMigratesAndReapsLegacyRowsViaPeerFallback  factory_run_retire_test.go:219: start short-lived process: …
rem  … factorymsg TestResolveActiveRunReadOnlyDeadOwnerNextToLiveWritesNothing  resolve_active_run_readonly_test.go:87: start short-lived process: …
```

(cli 11 + procrun 3 + factorymsg 5 = 19. 행 앞부분은 축약했고 메시지 부분은 원문.) `start short-lived process` 문장은 이 트리에서 두 도우미에만 있다:

```
$ grep -rn 'short-lived process' --include='*_test.go' internal | grep -v 'wait for'
internal/factorymsg/factory_run_retire_test.go:41:		t.Fatalf("start short-lived process: %v", err)
internal/cli/factory_runs_command_test.go:22:		t.Fatalf("start short-lived process: %v", err)
```

위 시험 함수 10개는 모두 이 트리에 있고(각 이름 `grep -rln "func <name>("` 확인), 도우미 호출 지점은 cli 6곳·factorymsg 7곳이다. 다른 cli `/bin/sh` 사용 파일(약 50개)은 이 19건에 들어 있지 않다 — 다른 오류 묶음이거나 `//go:build !windows`, `darwin || linux`로 묶여 있다.

E2. `go test ./internal/factorymsg -run 'Retire|Dead' -count=1` → exit 0
```
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	16.460s
```
E2b. 위 패턴에 걸리지 않는 두 호출자: `go test ./internal/factorymsg -run 'TestPartialStampWithBrokerStaysIndeterminate|TestResolveActiveRunMigratesAndReapsLegacyRowsViaPeerFallback' -count=1 -v` → exit 0
```
--- PASS: TestPartialStampWithBrokerStaysIndeterminate (0.82s)
--- PASS: TestResolveActiveRunMigratesAndReapsLegacyRowsViaPeerFallback (2.35s)
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	3.577s
```
E3. `go test ./internal/cli -run 'TestFactoryRuns|TestRelaunchRetireFromRunOpensMoaiPathsRoot' -count=1` → exit 0
```
ok  	github.com/modu-ai/moai-adk/internal/cli	22.714s
```
E4. `go test ./internal/agenticfactory/effect/internal/procrun -count=1 -v` → (darwin, 생략 분기 미진입)
```
--- PASS: TestStdoutAndStderrAreSeparateAndTheExitCodeIsAnError (0.02s)
--- PASS: TestTheLimitIsRequiredAndAStreamPastItStopsTheProcess (0.04s)
--- PASS: TestTheContextEndsTheProcessAndTheArgumentsAndEnvironmentArriveAsGiven (0.11s)
ok  	github.com/modu-ai/moai-adk/internal/agenticfactory/effect/internal/procrun	0.251s
```
E5. 도우미 pid의 분류 — overlay로 임시 시험 파일을 두 패키지에 넣어(트리 미변경) 도우미를 5회씩 부르고 `homestate.DefaultOwnerClassifier(pid, fp)`로 분류. `go test -overlay <scratch>/overlay.json ./internal/factorymsg ./internal/cli -run TestScratchT2123DeadHelperIsDead -count=1 -v` → exit 0
```
factorymsg: iter=0 pid=29382 fp="1793971421.595533" class=dead elapsed=22.0865ms … iter=4 pid=29404 class=dead elapsed=25.127ms
ok  	github.com/modu-ai/moai-adk/internal/factorymsg	0.516s
cli: iter=0 pid=73798 fp="1793971619.926198" class=dead elapsed=595.870916ms
cli: iter=1 pid=73939 class=dead elapsed=1.148632s   iter=2 elapsed=966.9115ms   iter=3 elapsed=749.985917ms   iter=4 pid=74714 class=dead elapsed=691.841875ms
ok  	github.com/modu-ai/moai-adk/internal/cli	6.790s
```
10회 모두 dead. 자식이 Go 런타임 시작 시간만큼 살아 있어 지문은 자리표시 문자열이 아닌 실제 시작 시각으로 잡혔다(이전 `/bin/sh -c "exit 0"`은 탐침 전에 끝나 자리표시가 들어갈 수 있었음). 두 경우 모두 dead로 분류되므로(`ClassifyOwnerWith`: Wait 뒤 pid dead → OwnerDead, 재사용 pid → 지문 불일치 → OwnerDead) 동작은 보존된다.

E6. 변이 검사 — cli 도우미가 live 신원(`os.Getpid()`, `CurrentProcessFingerprint()`)을 돌려주도록 바꾼 사본을 overlay로 넣고 `go test -overlay <scratch>/overlay_mutant.json ./internal/cli -run 'TestFactoryRuns' -count=1` → exit 1
```
--- FAIL: TestFactoryRunsCommandReportsAndRefuses (2.60s)
    factory_runs_command_test.go:85: --retire run-dead: factory run owner is not dead: run run-dead owner classified live
--- FAIL: TestFactoryRunsRetireRebindTo (14.53s)
--- FAIL: TestFactoryRunsRetireRefusesInProgressCards (1.10s)
--- FAIL: TestFactoryRunsRetireReportsStrandedCardsOfRetiredRun (1.94s)
FAIL	github.com/modu-ai/moai-adk/internal/cli	24.373s
```
시험들이 도우미가 dead pid를 내는 데 실제로 의존함을 확인(변이체 사살).

E7. `GOOS=windows go vet ./internal/factorymsg ./internal/cli ./internal/agenticfactory/effect/internal/procrun` → 출력 없음, `vet_exit=0`. `gofmt -l` 변경 Go 파일 3개 → 출력 없음, exit 0.

E8. `git show 47963571a1 --unified=0 | grep -inE 'token|secret|password|api[_-]?key|http://|https://'` → 일치 0건(exit 1).

E9. 자식 실행의 부작용 검토(코드 읽기) — factorymsg `TestMain`은 `brokerPathGitHelper()`(바이너리 이름이 `git`일 때만 동작, `store_path_context_test.go:20`)와 `gitenv.ScrubProcess()`뿐이라 자식은 바로 끝난다. cli `TestMain`은 재실행 자식을 위해 설계돼 있어 홈·프로필·git config 샌드박스를 부모 환경변수로 물려받고 새로 만들지 않으며(`main_test.go:169-179`, `:279-286`, `:490-496`, `:538-540`), 새로 만드는 영수증 임시 디렉터리는 `os.Exit` 직전 복원에서 지운다(`:461`, `:560-571`). `-test.run=^$`이므로 시험이 돌지 않아 재귀 실행은 없다. Windows `ProbeProcessIdentity`는 `OpenProcess`의 `ERROR_INVALID_PARAMETER`를 dead로, 지문을 생성 시각으로 판정한다(`pid_state_windows.go`, `process_fingerprint_windows.go`) — Wait 후 자식은 dead로 판정될 것으로 읽히지만 실행으로 확인하지는 못했다(Gaps).

E10. Codex 감사 1회 — `mcp__moai__codex_audit(project_root=<이 트리>, target=baseBranch, delta_base=04d8c27a7e, mode=native)`:
```
codex verdict: pass
summary: 지정 기준 커밋 대비 변경에서 보고할 결함은 찾지 못했어. 정적 검토만 했고, 테스트·빌드·재현 명령은 실행하지 않았어.
findings: []
resolved_model: gpt-6.1-sol, requested_effort: high
review_base: delta_base (merge base 04d8c27a7ed77afc09217f8623519806d48df3c4)
audit_receipt: rcpt-c671fc729913939c39b1da42
build_commit: 04d8c27a7e (build_lag: 설치 바이너리가 HEAD 47963571a의 조상)
```

## Baseline-attribution

모든 명령은 이 감사 실행에서 트리 `/Users/goos/.moai/worktrees/projects/moai-adk-go/t2123`, HEAD `47963571a1cee3b9643ed9baba1a66d534a0e737`, darwin에서 실행했다. 시험은 `go test`(이 트리 소스로 빌드)로 돌았다. 감사 계획(`moai verify audit-plan --project-root <이 트리>`)은 `config_status: ok`, codex `required`, claude·glm `off`, `cross_model_active: true`였다. 이 명령과 Codex 감사는 설치 바이너리/MCP 서버(빌드 `04d8c27a7e`)로 돌았다.

## Gaps

- Windows에서 실행하지 않았다. 이 커밋이 Windows race의 19건을 실제로 없애는지는 CI Windows 행렬만이 판정한다. 교차 vet과 Windows 프로브 코드 읽기(E9)까지가 이번 근거다. 특히 Windows에서 자식이 Wait 뒤 `ERROR_INVALID_PARAMETER`로 dead 판정되는지는 읽기 추론이다.
- RED 재현 없음(Windows 전용 실패라 이 기계에서 재현 불가).
- 감사 계획은 `cross_model_active: true`로 `audit_multi` 경로를 가리키지만, 리더 지시로 `audit_multi`는 금지되고 `codex_audit` 1회로 대체했다. 필수 게이트인 codex는 영수증으로 충족했고, claude·glm 게이트는 `off`다.
- 도구 출처(§2.2): `moai verify audit-plan`과 `codex_audit`를 낸 빌드 `04d8c27a7e`는 HEAD의 엄격한 조상이다. 이 커밋은 시험 파일과 progress.md만 바꾸고 감사 도구 코드는 건드리지 않으므로 판정에 영향은 없다고 보지만, 트리에서 빌드한 바이너리로 다시 돌리지는 않았다.
- 분류표 `categorized.tsv`는 다른 세션의 scratch(`/private/tmp/claude-501/-Users-goos-MoAI-moai-adk-go/43472e68-…/scratchpad/`)에 있는 기계 로컬 파일이다. 결정에 쓴 행은 E1에 옮겨 적었다.
- cli·factorymsg 전체 스위트, golangci-lint, 커버리지는 돌리지 않았다(시험 코드만 바뀐 변경이라 커버리지 기준은 해당 없음).

## Residual-risk

- Windows race에서 cli 자식 실행(F1)이 호출당 수 초까지 늘 수 있다. 시험 시간 초과로 번질 가능성은 낮지만 0은 아니다.
- procrun 패키지 전체가 Windows에서 빠지므로(F2) 그쪽 커버리지가 사라지고, CI 집계에는 생략으로 보이지 않는다.
- 조각 ②③④와 SQLite·fsync 묶음이 남아 있어 #1969의 Windows race 행렬은 이 커밋만으로는 녹색이 되지 않는다.

## 반복 이력

- 1회차(이번): PASS 0.92 @ `47963571a1`, 차단 지적 0건.
