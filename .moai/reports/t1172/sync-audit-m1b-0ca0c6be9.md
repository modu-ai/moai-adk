# t1172 M1-b 비모델 사전 증분 재감사 — 0ca0c6be9

## Claim

**PASS — 첫 모델 LIVE 직전의 M1-b 코드 게이트 범위.** F1(시작 검사 실패 기록), F2(16회 상한 정지), F3(정지 시각과 후속 LIVE 거부)가 측정한 경로에서 충족됐다. 이 판정은 판별 LIVE 자체나 SPEC 전체 완료 판정이 아니다.

## Evidence

기준 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1172`; `git rev-parse --short HEAD` → `0ca0c6be9`; 최초 `git status --short` → 출력 없음. 정식 장부 확인: `canonical_rows=4 startup=4 live=0`; 처음 네 행의 고정 sha256 `4e9b04acb14ed10b7122f1740fdcfaec147a91698bdebe2867a43123d6cb73ef`.

F3 독립 검사: 저장소 파일을 수정하지 않는 Go `-overlay` 임시 테스트를 만들었다. 테스트는 `preApprovalWriteLedger`로 장부를 작성하고, 시각 없는 stop 및 stop 뒤 LIVE 거부 전후의 파일 바이트를 직접 비교했다. 명령: `unset MOAI_CODEX_PREAPPROVAL_LIVE MOAI_T1172_EVIDENCE_DIR && go test -overlay <임시 overlay.json> ./internal/cli -run '^TestAuditF3Independent$' -count=1 -v`.

```text
=== RUN   TestAuditF3Independent
    audit_f3_overlay_test.go:19: nil timestamp rejected; ledger bytes unchanged
    audit_f3_overlay_test.go:23: stop started_ns positive=true
    audit_f3_overlay_test.go:28: late LIVE rejected; ledger bytes unchanged
    audit_f3_overlay_test.go:32: equal-time LIVE and evidence-directory reuse rejected
--- PASS: TestAuditF3Independent (0.00s)
PASS
ok   github.com/modu-ai/moai-adk/internal/cli 0.759s
```

AC-CPP-014의 실제 정지 절(`stop.started_ns` 최소값 뒤 모든 LIVE가 그보다 앞이어야 함)을 `jq -n`의 합성 장부에 적용했다:

```text
LIVE(started_ns=2), stop(started_ns=3), LIVE(started_ns=4): stop_value=3, ac_stop_clause=false
LIVE(started_ns=2), stop(started_ns=3): stop_value=3, ac_stop_clause=true
```

집중 테스트: `unset MOAI_CODEX_PREAPPROVAL_LIVE MOAI_T1172_EVIDENCE_DIR && go test ./internal/cli -run '^(TestCodexPreApprovalStartupLedgerRetention|TestCodexPreApprovalLivePreflight|TestCodexPreApprovalM1aBaseline|TestCodexPreApprovalGateRefusals|TestCodexPreApprovalArgvShape)$' -count=1 -v`:

```text
--- PASS: TestCodexPreApprovalLivePreflight (0.00s)
--- PASS: TestCodexPreApprovalM1aBaseline (0.00s)
--- PASS: TestCodexPreApprovalGateRefusals (0.24s)
--- PASS: TestCodexPreApprovalArgvShape (0.02s)
--- PASS: TestCodexPreApprovalStartupLedgerRetention (0.01s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli 0.941s
```

`gofmt -l`(세 관련 Go 파일) 및 `git diff --check` → 둘 다 출력 없음. `0ca0c6be9` diff에서 판별 LIVE의 stop closure는 `preApprovalAppendStop`을 호출하고, 그 함수는 정수 `started_ns`를 쓴다. `preApprovalWriteLedger`는 시각 없는 stop, stop 이후 또는 같은 시각의 LIVE를 직렬화 전에 거부한다.

## Baseline-attribution

F3는 위 명령으로 `0ca0c6be9`에서 새로 검증했다. F1·F2는 직전 커밋 `d7cdda463`에서 임시 복제 장부로 각각 `startup=5 stop=1 raw=3`, `startup=16 stop=1 new_raw=0`을 실측했고, 이번 변경은 그 함수의 stop 작성 경로를 공통 `preApprovalAppendStop`으로 연결했다. 이번 집중 테스트가 해당 경로를 다시 통과했다. 실제 Codex 0.157.0의 모델 비접속 시작 검사 4개는 직전 감사에서 독립 임시 장부로 PASS(`moai=1` 네 번, car010 `decoy=1`, `items=0`, `live=0`)였고, 이번에는 16회 예산을 보존하기 위해 되풀이하지 않았다.

## Findings

- **F1 · 해소:** 실패한 시작 검사 1회와 stop이 공용 장부에 기록되는 경로가 유지됐다.
- **F2 · 해소:** 16회 상한 뒤 추가 검사 없이 stop을 기록하는 경로가 유지됐다.
- **F3 · 해소:** 시각 없는 stop 및 정지 뒤 LIVE가 거부되고, 거부 전후 장부 바이트가 같다. AC-CPP-014 정지 절도 올바른 참/거짓을 냈다.

## Gaps

실제 판별 모델 LIVE 및 car010/car011 LIVE는 실행하지 않았다. 도구 사전 승인이 `RAN`인지 `REFUSED`인지 아직 측정되지 않았다. 이 증분 감사는 전체 패키지 커버리지, Windows 실행, 최종 SPEC의 모든 AC를 판정하지 않는다.

## Residual-risk

모델 LIVE 중 실제 Codex 출력 형식이나 인증 상태가 예상과 다르면 분류가 실패할 수 있다. 판별 실행 뒤 원자료·공용 장부·AC 판정식을 다시 독립 검증해야 한다.
