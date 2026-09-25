# t1172 M1-b 비모델 사전 감사

## Claim

**FAIL — 94246aa04의 M1-b 하네스는 첫 모델 LIVE를 시작할 준비가 되지 않았다.** 시작 검사가 실패하면 이미 실행한 검사가 공용 장부의 16회 상한에 포함되지 않고, `stop` 행도 남지 않는다. 상한 사전 거부 경로도 `ABORTED`를 장부에 남기지 않는다. 기능 판정은 FAIL이다. Security, Craft, Consistency의 전체 점수는 이 제한된 사전 감사에서 산정하지 않았다.

## Evidence

기준 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1172`, `git rev-parse --short HEAD` 출력 `94246aa04`, 당시 `git status --short` 출력 없음.

정상 집중 테스트:

```text
unset MOAI_CODEX_PREAPPROVAL_LIVE MOAI_T1172_EVIDENCE_DIR && go test ./internal/cli -run '^(TestCodexPreApprovalLivePreflight|TestCodexPreApprovalM1aBaseline|TestCodexPreApprovalGateRefusals|TestCodexPreApprovalArgvShape)$' -count=1 -v
--- PASS: TestCodexPreApprovalLivePreflight (0.00s)
--- PASS: TestCodexPreApprovalM1aBaseline (0.00s)
--- PASS: TestCodexPreApprovalGateRefusals (0.23s)
--- PASS: TestCodexPreApprovalArgvShape (0.02s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli 0.850s
```

실패 경로 재현: M1-a 원본 장부 4행·원자료 12개를 임시 `.moai/reports/t1172`에 복사했다. 가짜 `codex`는 `thread.started` 한 줄만 출력하고 MCP를 띄우지 않으므로 모델 호출은 없었다. `MOAI_CODEX_PREAPPROVAL_LIVE=1 MOAI_T1172_EVIDENCE_DIR=<임시 증거> MOAI_CODEX_BIN=<가짜 codex> go test ./internal/cli -run '^TestCodexPreApprovalStartupPreserve$' -count=1 -v` 출력:

```text
=== RUN   TestCodexPreApprovalStartupPreserve
    codex_preapproval_startup_test.go:164: startup stopped after 2 entries; see ledger.json
--- FAIL: TestCodexPreApprovalStartupPreserve (5.29s)
FAIL
FAIL github.com/modu-ai/moai-adk/internal/cli 6.100s
FAIL
ledger_count_after_failed_startup=4
new_raw_startup_files=3
```

코드 위치: `internal/cli/codex_preapproval_startup_test.go:429-432`에서 실패 행을 로컬 `ledger`에만 넣고, `:481-482`에서 종료한다. 공용 장부 쓰기는 `:510-540`에 있어서 실행되지 않는다. `:213-215`의 16회 사전 상한 거부도 `t.Fatal`만 호출한다. 이 경로의 `stop`/`ABORTED` 기록은 없다.

## Baseline-attribution

이 판정은 `94246aa04` 기준이다. 감사 중 HEAD가 `f496f49e3`로 이동했고 두 테스트 파일의 동시 수정이 시작됐다. 그 뒤 시도한 추가 음성 검사는 jq 합성식 오류와 수정 중 컴파일 오류가 겹쳐 **무효**이며 판정 근거에 포함하지 않는다.

## Findings

- **F1 · High · blocking · 신뢰도 높음** — `internal/cli/codex_preapproval_startup_test.go:429-432,481-540`: 실패한 실제 시작 검사 1회와 `stop`이 공용 장부에 누락된다. REQ-CPP-003 및 AC-CPP-014의 절대 상한·기록 조건을 위반한다. 각 시작 검사가 끝나는 즉시 고유 원자료와 해당 행을 장부에 영속화하고, 실패 시 `stop`을 남긴 후 반환하라. 실패 재현을 회귀 테스트로 추가하라.
- **F2 · Medium · blocking · 신뢰도 높음** — 같은 파일 `:213-215`: 상한 때문에 다음 시작 검사를 거부할 때 장부에 `ABORTED` 또는 `stop`을 남기지 않는다. REQ-CPP-004와 plan §D의 정지 기록 조건을 충족하도록 거부 행을 쓰고, 16행 상태에서 17번째 시작 검사가 기동되지 않는 음성 테스트를 추가하라.

## Gaps

- Codex 모델 LIVE는 실행하지 않았다. 실제 두 팔의 거부·성공 여부와 뒤따르는 car010/car011은 아직 측정되지 않았다.
- 94246aa04 전체 패키지 커버리지, 보안 점검, 형식 검사 및 Windows 실행은 이 제한된 사전 감사에서 측정하지 않았다.
- 동시 수리 중인 `f496f49e3` 이후 상태에는 이 판정을 적용하지 않는다. 수정 커밋에서 F1·F2를 다시 검사해야 한다.

## Residual-risk

현재 녹색 집중 테스트는 시작 검사 실패 시 장부 영속화 경로를 실행하지 않는다. LIVE 전에 음성 재현으로 새 장부 행과 상한을 확인해야 한다.
