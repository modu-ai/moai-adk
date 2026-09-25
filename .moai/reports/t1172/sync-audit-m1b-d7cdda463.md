# t1172 M1-b 비모델 사전 재감사 — d7cdda463

## Claim

**FAIL.** 앞선 F1·F2(실패한 시작 검사와 16회 상한의 장부 기록)는 수정된 것으로 실측했다. 그러나 판별 LIVE의 `stop` 행에 `started_ns`가 없어 AC-CPP-014의 정지 뒤 호출 판정이 실행되지 않는다(F3). 첫 모델 LIVE는 보류한다.

## Evidence

기준: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1172`; `git rev-parse --short HEAD` → `d7cdda463`; `git status --short` → 출력 없음. 설치 CLI `codex --version` → `codex-cli 0.157.0`. 아래 음성 검사는 M1-a 장부 4행·원자료 12개를 각각 독립 임시 디렉터리에 복사했다. 가짜 Codex는 `thread.started`만 출력하여 모델 호출은 없었다.

실패 시작 검사 음성: `MOAI_CODEX_PREAPPROVAL_LIVE=1 MOAI_T1172_EVIDENCE_DIR=<임시 증거> MOAI_CODEX_BIN=<가짜 codex> go test ./internal/cli -run '^TestCodexPreApprovalStartupPreserve$' -count=1 -v`:

```text
=== RUN   TestCodexPreApprovalStartupPreserve
    codex_preapproval_startup_test.go:165: startup MCP launch or non-model gate failed
--- FAIL: TestCodexPreApprovalStartupPreserve (5.15s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 5.796s
FAIL
rows=6 startup=5 stop=1 last=stop
new_raw_startup_files=3
```

상한 음성: 같은 원본 첫 4행을 네 번 연결해 시작 검사 16행의 합성 장부를 만든 뒤 위 명령을 실행했다:

```text
=== RUN   TestCodexPreApprovalStartupPreserve
    codex_preapproval_startup_test.go:165: startup invocation budget would be exceeded
--- FAIL: TestCodexPreApprovalStartupPreserve (0.00s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 0.650s
FAIL
rows=17 startup=16 stop=1 last=stop
new_raw_startup_files=0
```

집중 테스트: `unset MOAI_CODEX_PREAPPROVAL_LIVE MOAI_T1172_EVIDENCE_DIR && go test ./internal/cli -run '^(TestCodexPreApprovalStartupLedgerRetention|TestCodexPreApprovalLivePreflight|TestCodexPreApprovalM1aBaseline|TestCodexPreApprovalGateRefusals|TestCodexPreApprovalArgvShape)$' -count=1 -v`:

```text
--- PASS: TestCodexPreApprovalLivePreflight (0.00s)
--- PASS: TestCodexPreApprovalM1aBaseline (0.00s)
--- PASS: TestCodexPreApprovalGateRefusals (0.24s)
--- PASS: TestCodexPreApprovalArgvShape (0.02s)
--- PASS: TestCodexPreApprovalStartupLedgerRetention (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli 1.046s
```

실제 Codex의 모델 비접속 시작 검사: 독립 임시 복제 장부에서 `MOAI_CODEX_PREAPPROVAL_LIVE=1 MOAI_T1172_EVIDENCE_DIR=<임시 증거> go test ./internal/cli -run '^TestCodexPreApprovalStartupPreserve$' -count=1 -v`:

```text
disc-control startup exit=-1 moai=1 decoy=0 items=0
disc-treatment startup exit=-1 moai=1 decoy=0 items=0
car010 startup exit=-1 moai=1 decoy=1 items=0
car011 startup exit=-1 moai=1 decoy=0 items=0
--- PASS: TestCodexPreApprovalStartupPreserve (87.07s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli 87.662s
run_status=0
rows=8 startup=8 stop=0 live=0
```

새 4행 각각의 `passed`, 임시 루트, 반출 뒤 시작, 반출·표준 출력·표준 오류·기동 원자료 해시는 독립 Python 읽기 검사에서 모두 `True`; `all_checks= True`. M1-a 첫 4행의 고정 해시는 `4e9b04acb14ed10b7122f1740fdcfaec147a91698bdebe2867a43123d6cb73ef`로 일치했다. `gofmt -l`(세 관련 Go 파일)과 `git diff --check`는 출력이 없었다.

잔여 정지 판정 음성: `jq -n '[{"kind":"live","started_ns":2},{"kind":"stop"},{"kind":"live","started_ns":3}] as $all | ($all|map(select(.kind=="stop"))|map(.started_ns)|min) as $stop | {stop_value:$stop,ac_stop_clause:($stop==null or ($all|map(select(.kind=="live"))|all(.started_ns < $stop)))}'`:

```json
{"stop_value":null,"ac_stop_clause":true}
```

## Baseline-attribution

모든 결과는 이 감사에서 관측한 `d7cdda463`와 별도 임시 증거 디렉터리에 귀속된다. 실패 테스트의 FAIL은 의도한 음성 경로의 신호이며 제품 정상 경로 PASS로 세지 않았다. 모델 LIVE 호출은 0회다.

## Findings

- **F1 · 재검사 PASS:** 실패한 실제 시작 검사 1회와 `stop`이 장부에 추가되고 원자료 3개가 남았다.
- **F2 · 재검사 PASS:** 16행 상한에서 추가 시작 검사 0회, `stop` 1행이 남았다.
- **F3 · Medium · blocking · 신뢰도 높음:** `internal/cli/codex_preapproval_discriminator_live_test.go:615-620`의 LIVE `stop` closure가 `started_ns` 없이 행을 쓴다. AC-CPP-014의 `$stop`은 `null`이 되어 정지 뒤 LIVE 허용 여부 검사 절이 항상 참으로 빠진다. `time.Now().UnixNano()`를 정지 행에 기록하고, stop 뒤 LIVE를 넣은 합성 장부가 거짓이 되는 음성 검사를 추가하라.

## Gaps

판별 LIVE 및 car010/car011 LIVE는 실행하지 않았다. 실제 도구 승인 결과는 확인되지 않았다. 전체 패키지 커버리지·Windows 런타임은 이 범위에서 측정하지 않았다.

## Residual-risk

F3를 고친 뒤 정지 판정 음성과 집중 테스트를 다시 실행해야 한다. 정상 시작 검사를 정식 장부에 다시 실행하면 16회 예산을 소비하므로 이 감사의 정상 검사는 임시 복제본에서만 수행했다.
