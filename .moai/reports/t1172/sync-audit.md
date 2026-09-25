# t1172 M1-a 독립 감사 — iterations 1–2

SPEC: `SPEC-CODEX-PREAPPROVAL-PROBE-001` v0.5.1
첫 대상: `11e350eda1978188158bb391436a08eb8f48b7b3`; 재감사 대상: `655a9286b`
**최신 범위 판정: F1·F2 수리 PASS (100/100). M1-b LIVE는 아직 준비 중.**

아래 iteration 1의 `FAIL (59/100)`은 최초 커밋의 판정으로 보존했다. 최신 PASS는 F1·F2 결함 수리와 M1-a 비모델 선행 검사에 한정한다. SPEC 전체 AC 완료 판정이 아니다.

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---:|---|---|
| Functionality (40%) | 50/100 | FAIL | 네 시작 검사 원자료와 32개 반출 파일 해시는 일치했지만, LIVE 전 게이트가 추가 diff와 빈 시작 검사 결과를 허용했다. 아래 F1·F2. |
| Security (25%) | 75/100 | PASS | `go mod verify` → `all modules verified`; 임시 루트·표지·git top-level 확인과 timeout이 있다. F1·F2가 안전 게이트에 닿으므로 LIVE 보류. |
| Craft (20%) | 25/100 | FAIL | 이 좁은 `-run` 범위의 `go test -cover`가 `internal/cli` 5.3%, `internal/codexwiring` 1.7%였다. 패키지 전체 커버리지라 추가된 테스트 전용 함수의 개별 커버리지로 해석하지 않는다. 기본 프로필의 85% 기준은 이 측정으로 입증되지 않았다. |
| Consistency (15%) | 100/100 | PASS | `gofmt -l` 출력 없음, `go vet ./internal/cli ./internal/codexwiring` 출력 없음, `git diff --check 7ff8b9f52..11e350eda` 출력 없음. |

## Findings

- **F1 [High, blocking, confidence high]** `internal/cli/codex_preapproval_gate_test.go:59-76` — `preApprovalOnlyProjectConfigDiff`는 첫 승인 표 diff 뒤의 `Binary files … differ` 행을 무시한다. 실제 `diff -r`에 다른 파일의 바이너리 차이를 더한 뒤 Go overlay 테스트에서 `helper_accepts_extra_binary_diff=true`였다. REQ-CPP-002와 AC-CPP-003의 “두 팔의 차이는 승인 표 하나” 및 LIVE 전 차단 조건에 어긋난다. **Required fix:** 허용된 단일 diff 헤더·hunk·선택적 빈 줄·정확한 표 두 줄을 전체 문자열 기준으로 검사하고, 그 밖의 줄은 모두 거부한다. 추가 바이너리 차이의 음성 테스트를 넣는다.
- **F2 [High, blocking, confidence high]** `internal/cli/codex_preapproval_gate_test.go:36-55` — `startupOutput`이 빈 문자열이어도 LIVE callback이 실행된다. Go overlay 테스트가 `empty_startup_live_callbacks=1 row_kind=live`를 출력했다. `thread.started`, moai 기동 수, 설정 적재 결과를 게이트가 양성으로 확인하지 않는다. REQ-CPP-003의 “시작 검사 통과 뒤 LIVE”를 강제하지 못한다. **Required fix:** 문자열 패턴의 부재 대신 구조화된 시작 검사 결과(`thread.started` 1개, 비허용 item 0개, 설정 오류 없음, moai 기동 1개 이상, car010 decoy 1개 이상)를 입력으로 받아 모두 확인한 후에만 callback을 실행하고, 빈/임의 결과는 `stop`으로 기록하는 테스트를 추가한다.
- **F3 [Medium, optional, confidence high]** `.moai/reports/t1172/verdict.md:491-496` — 시작 검사 원시 JSONL·장부·매니페스트는 현재 이 워크트리의 ignored 로컬 산출물이다. 이 카드의 원자료 재현성은 워크트리 보존에 의존한다. 저장소의 카드별 보고서 정책에 따라 리드가 파일 보존·수집을 명시적으로 관리한다.

## Claim

Codex CLI 0.157.0 로더 점검, 네 픽스처의 비모델 시작 검사 원자료, 반출 해시, 운영자 확정 기록은 현재 트리에서 확인했다. 그러나 F1·F2 때문에 이 커밋의 LIVE 전 게이트는 불충분하다. `OPERATOR-CONFIRM`은 첫 LIVE 전 한 줄 존재하며 LIVE 장부 행은 0개다. `model_provider="dead"`는 AC-CPP-004가 시작 검사에 요구한 제공자 선택이며, 금지된 `model` 지정으로 세지 않는다.

## Evidence

실행 명령과 관측 출력(이 감사 턴):

```text
codex --version
codex-cli 0.157.0

go test ./internal/codexwiring -run '^TestGeneratedConfigKeysLoadInCodex$' -count=1 -v -timeout=30s
--- PASS: TestGeneratedConfigKeysLoadInCodex (0.09s)
    --- PASS: TestGeneratedConfigKeysLoadInCodex/generated_config (0.05s)
    --- PASS: TestGeneratedConfigKeysLoadInCodex/positive_control_misspelled_key (0.02s)
    --- PASS: TestGeneratedConfigKeysLoadInCodex/positive_control_bad_enum (0.02s)
    --- PASS: TestGeneratedConfigKeysLoadInCodex/diagnostic_matcher_controls (0.00s)
PASS

unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test ./internal/cli -run '^TestCodexPreApprovalGateRefusals$' -count=1 -v -timeout=40s
--- PASS: TestCodexPreApprovalGateRefusals (0.24s)
    --- PASS: TestCodexPreApprovalGateRefusals/missing_input (0.00s)
    --- PASS: TestCodexPreApprovalGateRefusals/arm_diff_exceeds (0.00s)
    --- PASS: TestCodexPreApprovalGateRefusals/startup_check_fails (0.23s)
    --- PASS: TestCodexPreApprovalGateRefusals/clean_refuses_non_fixture_root (0.00s)
    --- PASS: TestCodexPreApprovalGateRefusals/kills_only_recorded_pids (0.00s)
PASS

go mod verify
all modules verified

go test ./internal/cli ./internal/codexwiring -run '^(TestCodexPreApprovalGateRefusals|TestGeneratedConfigKeysLoadInCodex)$' -count=1 -cover -timeout=40s
ok  github.com/modu-ai/moai-adk/internal/cli 1.068s coverage: 5.3% of statements
ok  github.com/modu-ai/moai-adk/internal/codexwiring 0.602s coverage: 1.7% of statements
```

임시 Go overlay는 저장소 파일을 바꾸지 않고 대상 테스트 파일에 한 개의 감사 테스트만 덧붙였다. F1의 입력은 실제 `diff -r inputs/control inputs/treatment` 출력이다:

```text
diff_exit 1
diff -r inputs/control/project-config.toml inputs/treatment/project-config.toml
1a2,4
> [빈 줄: 실제 diff에는 `>` 뒤 공백 하나]
> [mcp_servers.moai.tools.codex_role_audit]
> approval_mode = "approve"
Binary files inputs/control/z.bin and inputs/treatment/z.bin differ
go_test_exit 0
=== RUN   TestAuditBinaryDiff
    codex_preapproval_gate_test.go:235: helper_accepts_extra_binary_diff=true
--- PASS: TestAuditBinaryDiff (0.00s)
```

F2의 임시 Go overlay 결과:

```text
go_test_exit 0
=== RUN   TestAuditEmptyStartup
    codex_preapproval_gate_test.go:235: empty_startup_live_callbacks=1 row_kind=live
--- PASS: TestAuditEmptyStartup (0.00s)
```

원자료를 읽어 SHA-256과 순서를 다시 계산한 출력:

```text
discriminator declared 14 actual 14 mismatch 0 missing 0
car010 declared 9 actual 9 mismatch 0 missing 0
car011 declared 9 actual 9 mismatch 0 missing 0
disc-control export_before_start True duration_s 20.0 config_hash_match True home_hash_match True auth False False
disc-treatment export_before_start True duration_s 20.0 config_hash_match True home_hash_match True auth False False
car010 export_before_start True duration_s 20.0 config_hash_match True home_hash_match True auth False False
car011 export_before_start True duration_s 20.0 config_hash_match True home_hash_match True auth False False
```

`jq -r '.[]|[.fixture,.moai_launches,.decoy_launches]|@tsv' .moai/reports/t1172/ledger.json`에서 moai는 네 픽스처 모두 1회, car010 decoy는 1회였다. 네 `startup/*.jsonl`에는 각각 `thread.started` 1개·`turn.started` 1개·`error` 3개와 `item.*` 0개가 있었다. `rg -n '^OPERATOR-CONFIRM ' .moai/reports/t1172/verdict.md`는 450행 한 줄을 찾았고, `jq -r '[.[]|select(.kind=="live")]|length' .moai/reports/t1172/ledger.json` 출력은 `0`이었다.

## Baseline-attribution

이 감사의 기준은 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1172`, 브랜치 `WT-codex-preapproval-probe`, HEAD `11e350eda`와 설치된 `codex-cli 0.157.0`이다. 원자료는 이 워크트리의 `.moai/reports/t1172/`에서 읽었다. `git merge-base --is-ancestor 51d3e5be6 HEAD` 종료 코드는 0이었다. 이전 트리·다른 패키지의 통과 수치를 이번 측정으로 쓰지 않았다.

## Gaps

- 판별 LIVE 호출과 실제 도구별 승인 효과는 실행하지 않았다. 장부 LIVE 행은 0개다.
- 현재 이월 AC용 `argv.txt`·프롬프트는 M1-a 준비용 값이다. M1-b의 실제 인자가 정해지면 반출·매니페스트·마지막 시작 검사를 다시 수행해야 한다.
- `AC-CPP-002`·`003`·`004`의 전체 판정 명령은 LIVE 산출물을 요구해 이 단계에서 PASS 처리하지 않았다. 원자료의 일부 조건만 독립 확인했다.
- `-cover` 수치는 좁힌 테스트 선택으로 계산된 패키지 전체 수치다. 추가된 테스트 전용 함수만의 커버리지를 뜻하지 않는다.

## Residual-risk

F1·F2를 고친 뒤에도 실제 Codex 승인 결과는 별도 LIVE 판별이 필요하다. 수정 후 음성 대조를 먼저 통과시키고, 반출본과 실제 LIVE 인자·프롬프트의 해시 일치, 네 픽스처의 마지막 시작 검사, 전체 상한과 장부 순서를 확인한 다음 M1-b 호출을 시작해야 한다.

---

## Iteration 2 — F1·F2 수정 재감사

**Scoped Verdict: PASS (100/100, F1·F2 수리와 M1-a 비모델 선행 검사).** 앞의 iteration 1 FAIL은 `11e350eda`에 대한 기록으로 남긴다. 이 범위 판정은 SPEC 전체의 PASS를 뜻하지 않는다.

| Finding | 최신 상태 | 독립 확인 |
|---|---|---|
| F1 High blocking | RESOLVED | 실제 `diff -r`의 `z.bin` 추가 차이 → helper `false`, 게이트 `stop`, LIVE callback 0회. 정상 표만 있는 diff는 허용. |
| F2 High blocking | RESOLVED | 빈·임의 시작 검사 proof → 게이트 `stop`, callback 0회. `thread.started` 1개와 moai 기동 1회의 정상 proof → `live`, callback 1회. |
| F3 Medium optional | OPEN | 원시 JSONL·장부·매니페스트는 여전히 카드 워크트리의 ignored 로컬 산출물. 보고서는 리드가 명시 수집한다. |

### Claim

수정 커밋의 허용 diff 검사는 전체 출력의 줄 수와 모든 허용 줄을 검사하므로 추가 바이너리 차이를 거부한다. 시작 검사 proof는 `thread.started` 1개, moai 기동, car010 decoy 기동, 비허용 item 부재, 적재 오류 부재를 양성으로 검사한다. 기존 5개 거부 하위 테스트가 통과했고, 이 감사가 저장소 파일을 바꾸지 않는 Go overlay로 F1·F2의 실패 입력 및 정상 입력을 다시 실행했다. 실제 Codex CLI 0.157.0의 네 비모델 시작 검사도 독립 재실행으로 통과했다.

### Evidence

```text
python3 <temporary Go-overlay audit runner>  # 기존 테스트 파일을 임시 파일로만 대체; worktree 코드는 수정하지 않음
exit 0
=== RUN   TestAuditF1F2Delta
    codex_preapproval_gate_test.go:349: F1 actual_extra_binary_diff_rejected=true
    codex_preapproval_gate_test.go:357: F2 empty callbacks=0 row=stop
    codex_preapproval_gate_test.go:357: F2 arbitrary callbacks=0 row=stop
    codex_preapproval_gate_test.go:357: F2 valid callbacks=1 row=live
--- PASS: TestAuditF1F2Delta (0.01s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli 0.760s

unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && go test ./internal/cli -run '^TestCodexPreApprovalGateRefusals$' -count=1 -v -timeout=40s
--- PASS: TestCodexPreApprovalGateRefusals (0.23s)
    --- PASS: TestCodexPreApprovalGateRefusals/missing_input (0.00s)
    --- PASS: TestCodexPreApprovalGateRefusals/arm_diff_exceeds (0.00s)
    --- PASS: TestCodexPreApprovalGateRefusals/startup_check_fails (0.22s)
    --- PASS: TestCodexPreApprovalGateRefusals/clean_refuses_non_fixture_root (0.00s)
    --- PASS: TestCodexPreApprovalGateRefusals/kills_only_recorded_pids (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli 0.805s

unset MOAI_KANBAN MOAI_KANBAN_ID MOAI_KANBAN_LABEL MOAI_KANBAN_LEAD_ADDR MOAI_KANBAN_SETTINGS_INJECTED MOAI_KANBAN_BACKEND MOAI_FACTORY_WORKER MOAI_FACTORY_WORKERS && MOAI_CODEX_PREAPPROVAL_LIVE=1 MOAI_T1172_EVIDENCE_DIR=.moai/reports/t1172 go test ./internal/cli -run '^TestCodexPreApprovalStartup$' -count=1 -v -timeout=150s
    codex_preapproval_startup_test.go:342: disc-control startup exit=-1 moai=1 decoy=0 items=0
    codex_preapproval_startup_test.go:342: disc-treatment startup exit=-1 moai=1 decoy=0 items=0
    codex_preapproval_startup_test.go:342: car010 startup exit=-1 moai=1 decoy=1 items=0
    codex_preapproval_startup_test.go:342: car011 startup exit=-1 moai=1 decoy=0 items=0
--- PASS: TestCodexPreApprovalStartup (87.19s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli 87.778s

go vet ./internal/cli ./internal/codexwiring && echo 'go vet: clean'
go vet: clean
gofmt -l <세 변경 테스트 파일>
(출력 없음)
git diff --check 11e350eda..655a9286b -- internal/cli internal/codexwiring && echo 'code diff-check: clean'
code diff-check: clean
go mod verify
all modules verified
```

원자료 독립 재계산:

```text
discriminator manifest_files 14 hash_mismatches 0
car010 manifest_files 9 hash_mismatches 0
car011 manifest_files 9 hash_mismatches 0
disc-control export_before_start True thread_started 1 non_error_items 0 moai 1 decoy 0 config_hash_match True auth_present False
disc-treatment export_before_start True thread_started 1 non_error_items 0 moai 1 decoy 0 config_hash_match True auth_present False
car010 export_before_start True thread_started 1 non_error_items 0 moai 1 decoy 1 config_hash_match True auth_present False
car011 export_before_start True thread_started 1 non_error_items 0 moai 1 decoy 0 config_hash_match True auth_present False
live_rows 0
```

### Baseline-attribution

이번 재감사는 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1172`의 HEAD `655a9286b`와 설치된 `codex-cli 0.157.0`을 기준으로 했다. 새 `moai-build.txt`는 `commit=655a9286b`, 바이너리 SHA-256 `2af8b956ac4df37b01b2226e57b0592a21d9a603a8bd96ec0f5c8f4c2f25a4fd`를 기록한다. 처음 실패한 `11e350eda`와 측정 기준을 섞지 않았다.

### Gaps — M1-b LIVE 착수 조건

- `TestCodexPreApprovalDiscriminatorLive`가 아직 코드에 없고 `preApprovalGate`는 현재 가짜 테스트에서만 호출된다. 따라서 이 재감사 직후 바로 모델 LIVE를 시작할 실행 경로는 없다.
- 현재 반출된 이월 AC의 `argv.txt`·프롬프트는 M1-a 준비용이다. 실제 판별·이월 호출의 인자와 프롬프트를 확정한 뒤 호출 라벨별로 다시 반출하고 매니페스트를 갱신해야 한다. 이어서 그 **마지막 반출 뒤** 네 픽스처의 시작 검사를 다시 실행하고, `AC-CPP-002`·`003`·`004`의 전체 판정으로 해시·순서·팔 차이를 닫아야 한다.
- 첫 LIVE 전에 `plan.md` §D의 시도별·절대 상한을 `verdict.md`에 기록하고, 인증 파일 전후 해시·pid 기록/정리·장부 `attempt`·`temp_root`·`fixture_root`·사용자 턴 1개 상한을 실제 LIVE 하네스에 연결해야 한다. 현재 `OPERATOR-CONFIRM`은 정확히 한 줄이고 장부의 LIVE 행은 0개다.
- `approval_mode="approve"`가 실제 무인 Codex 호출에서 승인 요구를 없애는지는 아직 측정하지 않았다. 이 질문 자체가 M1-b LIVE의 목적이다.

### Residual-risk

F1·F2가 수리됐어도 LIVE 호출을 시작하기 전에 위 선행 조건을 기계적으로 확인해야 한다. 첫 실제 호출은 대조 팔부터 실행하고, 판정식이 요구하는 MCP 도구 호출·거부 문구·호출 수·로그인 해시를 기록해야 한다. `NOT MEASURED` 또는 `SKIP`을 PASS로 세지 않는다.
