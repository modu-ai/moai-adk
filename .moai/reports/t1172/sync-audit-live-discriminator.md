# t1172 M1-b 실제 판별 LIVE 독립 감사

## Claim

**M1-b 증거 게이트 PASS.** `SPEC-CODEX-PREAPPROVAL-PROBE-001`의 AC-CPP-002·003·004·005·014 판정식이 이번 실제 판별 기록에서 모두 참이다. Codex CLI 0.157.0에서 대조 팔은 승인 정책에 막혔고(`REFUSED`), 처치 팔은 MoAI MCP 도구에 들어가 출력 경로 검증에서 거절됐다(`RAN`). 따라서 도구별 사전 승인 표가 이 실험의 MCP 승인 장벽을 넘긴다는 판별 결과가 있다. `RAN`은 하위 감사 실행 성공이나 카드 완료를 뜻하지 않는다.

**전체 SPEC 판정은 보류.** AC-CPP-006의 운영자 A1/A2 선택, 채택 갈래에 따른 AC-CAR-010, AC-CAR-011 등 나머지 AC는 이 감사 범위 밖이다. AC-CPP-013의 최종 판정도 `adoption.txt`가 아직 없어 실행할 수 없다.

## Evidence

기준 트리: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1172`. 감사 시작 HEAD `cd00bbd3f`, 마지막 HEAD `8ab0fa13b`. 중간 커밋 `8ab0fa13b`는 `internal/template/agentemit/audit_launcher_surface_test.go`와 `internal/template/templates/AGENTS.md.tmpl`만 바꿨고, 이번 감사의 Go 판별 코드와 원자료는 바꾸지 않았다. 마지막 `git status --short` 출력 없음. `codex --version` 출력 `codex-cli 0.157.0`.

원본 증거 디렉터리를 별도 임시 디렉터리에 복사한 뒤 `acceptance.md`에 적힌 **판정 명령 원문**을 실행했다. 실제 LIVE 실행 블록은 실행하지 않았다. 판정 명령이 생성하는 `.sha`·`.recheck` 파일은 임시 복제본에만 썼다.

```text
AC-CPP-002 judge1 exit=0 stdout='true' stderr=''
AC-CPP-002 judge2 exit=0 stdout='true' stderr=''
AC-CPP-003 judge1 exit=0 stdout='true' stderr=''
AC-CPP-004 judge1 exit=0 stdout='true' stderr=''
AC-CPP-004 judge2 exit=0 stdout='true' stderr=''
AC-CPP-005 judge2 exit=0 stdout='true' stderr=''
AC-CPP-014 judge1 exit=0 stdout='true' stderr=''
```

`jq -r 'select(.type=="item.completed" and .item.type=="mcp_tool_call") | [.item.server,.item.tool,.item.status,((.item.error // {})|tostring|.[0:180]),((.item.result // {})|tostring|.[0:180])] | @tsv'`를 두 팔의 원본 JSONL에 각각 적용한 출력:

```text
moai  codex_role_audit  failed  {"message":"MCP tool call requires approval, but approval policy is never"}  {}
moai  codex_role_audit  failed  {}  {"content":[{"type":"text","text":"codex_role_audit: codex audit sync-auditor: destination rejected: AGENTS.md is outside the report tree"}],"structured_content":null}
```

두 팔 모두 `item.completed`의 `moai/codex_role_audit` 호출 1개와 `turn.started` 1개다. 독립 Python 읽기 검사 출력:

```text
control calls 1 turns 1 outcome REFUSED raw_refusal True raw_destination False
control ledger_live 1 raw_sha256 63a59ceb54aa03aa7ea8a26fa9698013b394cd6b19dd8f13bb138d961255f73b
treatment calls 1 turns 1 outcome RAN raw_refusal False raw_destination True
treatment ledger_live 1 raw_sha256 810c2a8f7332619510261367c713a46caaac07ce3b096bcc22d3d44d27789673
auth_hash_equal True
pid_subset True recorded 48 killed 48
evidence_sha256 c719f9179a764c2dce3ebaa73b55da225579dd15335b43dd27ba51eb8667f34f
m1a_first4_sha256 4e9b04acb14ed10b7122f1740fdcfaec147a91698bdebe2867a43123d6cb73ef
```

`discriminator/live.jsonl`에는 `TestCodexPreApprovalDiscriminatorLive`의 `pass` 1개, `fail`·`skip` 0개, `CPP005_EVIDENCE_SHA256 c719f9179a764c2dce3ebaa73b55da225579dd15335b43dd27ba51eb8667f34f`가 있다. `evidence.json`은 `invocations: 2`, `attempt: 1`, `aborted: false`, 네 시작 검사 통과 중 판별 두 팔 `PASS`, 팔마다 자식·기동 기록 0, 로그인 해시 전후 같음을 기록한다. 장부에는 원래 M1-a 시작 검사 4행과 새 검사 4행, LIVE 2행만 있고 `stop`은 없다. 각 새 시작 검사는 약 20초, 두 LIVE는 각각 11.84초·12.27초였다. `internal/cli/codex_audit_launch.go:170-176`의 출력 경로 검증은 자식 실행 준비보다 앞서므로 처치 결과는 MCP 도구 진입까지의 근거다.

비모델 집중 검증:

```text
--- PASS: TestCodexPreApprovalLivePreflight (0.00s)
--- PASS: TestCodexPreApprovalM1aBaseline (0.00s)
--- PASS: TestCodexPreApprovalGateRefusals (0.23s)
--- PASS: TestCodexPreApprovalArgvShape (0.03s)
--- PASS: TestCodexPreApprovalStartupLedgerRetention (0.01s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli 1.037s
```

명령은 `unset MOAI_CODEX_PREAPPROVAL_LIVE MOAI_T1172_EVIDENCE_DIR && go test ./internal/cli -run '^(TestCodexPreApprovalLivePreflight|TestCodexPreApprovalM1aBaseline|TestCodexPreApprovalStartupLedgerRetention|TestCodexPreApprovalGateRefusals|TestCodexPreApprovalArgvShape)$' -count=1 -v`였다. 별도 `go vet ./internal/cli`와 `gofmt -l`(판별·시작 검사·게이트 3개 파일)은 exit 0, 출력 없음. `go mod verify`는 `all modules verified`를 출력했다. 같은 선택 테스트의 `go test -cover`는 `coverage: 5.3% of statements`를 출력했다. 이 수치는 큰 `internal/cli` 패키지에서 선택한 다섯 테스트만 돌린 결과이며, 전체 SPEC의 85% 커버리지 근거가 아니다.

## Baseline-attribution

판정과 원본 대조는 위 트리의 `cd00bbd3f`에 저장된 LIVE 원자료에 귀속된다. 감사 중 `8ab0fa13b`로 HEAD가 전진했으나 변경 경로를 `git diff --name-only cd00bbd3f..HEAD`로 확인해 판별 코드·원자료가 그대로임을 확인했다. 이 감사에서 모델 LIVE를 추가 호출하지 않았다. 복제본 판정은 원본의 내용 해시를 유지한 사본에 대한 재계산이다.

## Dimension Scores

| 차원 | 이 단계 판정 | 근거 |
|---|---|---|
| Functionality | PASS (M1-b 범위) | 위 AC-CPP-002·003·004·005·014 원문 판정 7개 모두 `true` |
| Security | PASS (M1-b 범위) | 로그인 해시 불변, 기록한 PID만 종료, 자식·기동 기록 0, `go mod verify` 통과; 발견한 Critical/High 없음 |
| Craft | UNVERIFIED (전체 SPEC) | 선택 테스트 통과와 `go vet` 통과. 선택 실행의 패키지 커버리지 5.3%는 전체 85% 문턱을 판정하지 못함 |
| Consistency | PASS (M1-b 범위) | `gofmt -l` 출력 없음, `git diff --check` 출력 없음 |

## Findings

이번 M1-b 범위에서 blocking 결함을 발견하지 못했다. 처치 JSONL의 `status: failed`는 출력 경로 검증의 오류 응답과 함께 기록됐다. 이 경우 AC-CPP-005가 정의한 `RAN`은 도구 호출이 승인 장벽을 넘어 검증 함수에 들어간 상태이며, 하위 `codex audit`가 실행됐다는 뜻이 아니다.

## Gaps

- AC-CPP-013의 최종 판정과 A1/A2 채택 결정은 실행하지 않았다. 현재 `adoption.txt`가 없다.
- 실제 Codex 프로세스 내부의 승인 판단을 관찰하는 계측은 없으며, JSONL 결과와 MoAI MCP 응답으로 판별했다.
- `evidence.json`의 자식 0·기동 기록 0은 종료 뒤 삭제된 임시 픽스처에서 독립적으로 다시 읽을 수 없다. 소스의 출력 경로 선행 거부와 실험 기록이 이를 뒷받침한다.
- Windows 런타임, 전체 패키지 테스트·커버리지, car010/car011 LIVE는 이 감사에서 실행하지 않았다.

## Residual-risk

이 결과로 도구별 사전 승인이 Codex CLI 0.157.0의 **이번 구성·이번 도구 호출**에서 통한 것은 뒷받침된다. 다른 CLI 버전, TUI, 실제 하위 감사 실행, 모든 도구에 대한 일반화는 추가 검증이 필요하다. 사용자 A1/A2 결정에 맞춘 구현과 이월 LIVE AC를 완료하기 전에는 t1172 전체를 PASS나 완료로 표시하면 안 된다.
