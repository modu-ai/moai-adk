# t1074 독립 동기화 감사 — 최종 재판정

- SPEC: `SPEC-FACTORY-MIXED-HOOK-001`
- 대상: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1074`
- branch / baseline HEAD: `WT-factory-mixed-hook` / `8c5d9be993ad45d5c91ae8a95af904b27c935599`
- 판정: **PASS**
- 점수: **85/100**
- 병합 차단 finding: **없음**
- 범위 밖 잔여 gate: repository-wide CI 및 integration-branch 판정

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---:|---|---|
| Functionality (40%) | 100/100 | PASS | AC-FMH-001..015와 AC-FMH-OPS-001..006 모두 현재 evidence contract를 충족한다. 새 AC11 GLM JSONL과 AC12~14 JSONL의 exact predicate가 모두 `true`다. |
| Security (25%) | 100/100 | PASS | 운영자 GLM credential은 기존 test seam으로 child 메모리에만 전달되고 launcher attribution/MCP config와 분리된다. 새 AC11 artifact 및 GLM artifacts에서 key 이름·credential assignment가 검출되지 않았다. |
| Craft (20%) | 25/100 | UNVERIFIED | scoped selector coverage는 package 전체 85%를 증명하지 않는다. 이는 병합 차단 코드 결함이 아니라 명시적 coverage GAP이다. |
| Consistency (15%) | 100/100 | PASS | focused race tests, `gofmt -l`, `git diff --check`, launcher process-primitive guard가 통과했다. |

가중 점수는 `100×0.40 + 100×0.25 + 25×0.20 + 100×0.15 = 85`다. must-pass인 Functionality와 Security가 모두 통과했고 활성 blocking finding이 없어 전체 판정은 PASS다.

## Findings

활성 finding 없음.

## Per-AC verdict

| Criterion | Verdict | 근거 |
|---|---|---|
| AC-FMH-001..009 | PASS | M5 10-selector artifact/predicate와 scoped regression evidence. |
| AC-FMH-010 | PASS | Codex↔Codex nonce `75bb84b10c8ee1397752a7c03a3681ed`, `acknowledged=2`. |
| AC-FMH-011 | PASS | `m5-reg-ac11-glm.jsonl`; nonce `057d60face93a6325425c3330f1914de`, `acknowledged=2`, exact predicate `true`. |
| AC-FMH-012 | PASS | `m5-reg-ac12-glm-rerun.jsonl`; nonce `8e74ffdc2458360dfdf1a9a3fc9c552f`, `acknowledged=2`, exact predicate `true`. |
| AC-FMH-013 | PASS | `m5-reg-ac13-glm.jsonl`; nonce `8103c0a80c989e669bc79b9c6af17ea4`, `acknowledged=2`, exact predicate `true`. |
| AC-FMH-014 | PASS | `m5-reg-ac14-glm.jsonl`; idle pending truth 및 next-turn receipt, exact predicate `true`. |
| AC-FMH-015 | PASS | M5 benchmark selector와 기록된 성능 예산 evidence. |
| AC-FMH-OPS-001..006 | PASS | exact 8-selector race, LIVE16 production chain, parser/trust/cleanup 검증. |

## Claim

현재 구현은 factory launcher pending 등록, POSIX exec owner/실패 rollback, SessionStart 원자적 bind, 첫 non-empty UserPromptSubmit의 inbox 전 authoritative rotation, empty/whitespace no-bind, bound endpoint 무쓰기, read-only roster, rollout result 귀속과 cross-turn 누적, production trust/cleanup/MCP owner 검증을 충족한다. Claude 접근 remediation은 freshly built fixture `moai glm -- ...`을 사용하며, manual fixture attribution과 production launcher attribution을 분리하고 fixture root의 명시적 `.moai` marker를 사용한다. 모든 scoped AC/OPS criterion의 최종 evidence가 PASS다.

## Evidence

### 1. AC11 외부 인증 GAP 해소 artifact

감사 중 처음에는 AC11 PASS가 `go test -v` 전사만 있고 canonical JSONL이 없어 evidence blocker로 판정했다. 재실행 후 다음 artifact를 직접 검사했다.

```text
.moai/reports/t1074/m5-reg-ac11-glm.jsonl
mtime: 2026-09-23T08:04:04+0900
size: 16887 bytes
sha256: 83d00e1c1516f16ab34cf137ea99b3c9928ad6bc4cc60c3ca2b1307a5d10f52c
```

Exact predicate:

```text
jq -se '([.[]|select(.Action=="pass" and .Test=="TestFactoryLiveCodexClaude")]|length)==1 and ([.[]|select(.Action=="fail")]|length)==0 and ([.[]|select(.Action=="skip")]|length)==0 and ([.[]|select((.Output//"")|contains("NOT_RUN"))]|length)==0' .moai/reports/t1074/m5-reg-ac11-glm.jsonl
true
```

Verbatim outcome:

```text
live separate-context nonce/receipt verified: case=codex-claude nonce=057d60face93a6325425c3330f1914de acknowledged=2
--- PASS: TestFactoryLiveCodexClaude (112.72s)
PASS
```

Parsed security/evidence shape:

```json
{
  "pass": ["TestFactoryLiveCodexClaude"],
  "fail": [],
  "skip": [],
  "not_run": false,
  "key_name": false,
  "credential_shape": false
}
```

### 2. AC11~14 final artifact matrix

각 파일에 named PASS 1개, fail 0개, skip 0개, `NOT_RUN` 0개를 요구한 동일 predicate의 출력:

```text
m5-reg-ac11-glm.jsonl               true
m5-reg-ac12-glm-rerun.jsonl         true
m5-reg-ac13-glm.jsonl               true
m5-reg-ac14-glm.jsonl               true
```

AC12~14 verbatim sentinels:

```text
live separate-context nonce/receipt verified: case=claude-codex nonce=8e74ffdc2458360dfdf1a9a3fc9c552f acknowledged=2
live separate-context nonce/receipt verified: case=claude-claude nonce=8103c0a80c989e669bc79b9c6af17ea4 acknowledged=2
idle pending truth and next-turn receipt verified: message=c26d88099996c5d86ab3c4d06808abc1
```

과거 direct-Claude OAuth FAIL artifacts와 최초 AC12 diagnostic FAIL은 이력으로 보존됐고 최종 PASS artifact와 파일명이 분리돼 있다.

### 3. GLM command·credential·attribution 경계

```text
unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY MOAI_TEST_GLM_KEY && MOAI_HOME=/tmp/t1074-sync-glm-home GOCACHE=/tmp/t1074-sync-glm-cache go test -race ./internal/cli -run '^(TestFactoryLiveClaudeCommandRoutesThroughBuiltMoaiGLM|TestFactoryLiveOperatorGLMKeyRestoresIsolatedHome|TestFactoryLiveGLMLauncherEnvDefersAttributionToMCPConfig)$' -count=1 -timeout=60s -v
=== RUN   TestFactoryLiveClaudeCommandRoutesThroughBuiltMoaiGLM
--- PASS: TestFactoryLiveClaudeCommandRoutesThroughBuiltMoaiGLM (0.00s)
=== RUN   TestFactoryLiveOperatorGLMKeyRestoresIsolatedHome
--- PASS: TestFactoryLiveOperatorGLMKeyRestoresIsolatedHome (0.00s)
=== RUN   TestFactoryLiveGLMLauncherEnvDefersAttributionToMCPConfig
--- PASS: TestFactoryLiveGLMLauncherEnvDefersAttributionToMCPConfig (0.00s)
PASS
ok  github.com/modu-ai/moai-adk/internal/cli  1.782s
```

`internal/cli/factory_live_test.go`의 최종 경로는 다음을 기계적으로 고정한다.

- backend `claude`는 fixture에서 빌드한 `f.moai glm -- ...`을 실행한다.
- operator `.env.glm`은 key 값만 메모리로 읽고 기존 isolated `MOAI_HOME`을 정확히 복원한다.
- launcher child env에서는 stale/manual factory attribution을 제거한다.
- run/session/role attribution은 turn별 strict `.mcp.json`의 MCP server env에만 기록한다.
- `MOAI_TEST_GLM_KEY`는 `.mcp.json`에 기록하지 않는다.
- fixture root에는 명시적 `.moai` directory를 생성한다.

### 4. 이전 코드 finding delta

POSIX Claude exec 실패 orphan은 앞선 iteration에서 수정됐다.

```text
TestAuditClaudePOSIXExecFailureRollsBackPending             PASS
TestCodexDirectPOSIXExecFailureRollsBackExactPending        PASS
TestClaudePOSIXExecFailureRollsBackPending                  PASS
TestCodexSpecFiles_ExecPrimitivesCodexOnly                  PASS
```

OPS exact 8-selector race predicate와 production LIVE16 predicate도 각각 `true`로 관측됐다.

### 5. Consistency/security probe

```text
gofmt -l internal/cli/factory_live_test.go internal/cli/factory_live_command_test.go
<no output>

git diff --check
<no output; exit 0>
```

Final GLM JSONL에서 `MOAI_TEST_GLM_KEY`, `GLM_API_KEY=`, `ANTHROPIC_AUTH_TOKEN=`, `Z_AI_API_KEY=` 문자열은 검출되지 않았다.

## Verified non-findings

- direct `claude -p` 경로가 최종 Claude-backed LIVE 명령에 남아 있지 않다.
- credential이 fixture `MOAI_HOME`, `.mcp.json`, JSONL output에 복사 또는 출력되지 않는다.
- GLM launcher가 inherited stale run/session/worker attribution을 사용하지 않는다.
- MCP server에는 현재 peer의 non-secret run/session/PID/role attribution만 전달된다.
- `.moai` root 부재로 GLM lead가 잘못된 project root를 선택하던 최초 AC12 diagnostic은 재현 가능한 fixture setup으로 해소됐다.
- parser는 output-only spoof, 다른 turn output, 중복 ID, 잘못된 wrapper를 거부하고 distinct turn의 valid call을 누적한다.
- roster 조회는 message claim이나 peer state write를 하지 않는다.

## Baseline-attribution

- 새 감사 명령은 HEAD `8c5d9be993ad45d5c91ae8a95af904b27c935599`와 현재 사용자 승인 t1074 미커밋 변경에서 실행했다.
- AC11은 새 raw JSON event stream을 직접 읽었고 전사된 요약만으로 PASS하지 않았다.
- AC12~14도 각 final raw JSONL에 named PASS가 정확히 하나 있고 fail/skip/`NOT_RUN`이 없음을 다시 판정했다.
- 기존 OPS/rollback/LIVE16 근거는 같은 worktree baseline의 직전 독립 감사에서 관측한 결과를 iteration history로 유지했다.
- 생산·테스트 구현, SPEC status, commit, branch, push, merge는 변경하지 않았다. 이 보고서만 갱신했다.

## Gaps

- 전체 package/repository coverage 85%는 미관측이다. scoped function coverage와 selector PASS를 전체 coverage로 일반화하지 않는다.
- Windows 실제 process-tree/LIVE는 실행하지 않았다.
- repository-wide CI 및 integration-branch 결과는 아직 이 감사 범위 밖이다.

## Residual-risk

- GLM credential bridge는 test-only LIVE harness에서 process-global `MOAI_HOME`을 매우 짧게 교체한다. 현재 테스트는 병렬 실행하지 않고 원값 복원을 검증하지만, 향후 이 helper를 parallel test에서 호출하면 환경 경쟁 위험이 생긴다.
- Codex/GLM CLI output 또는 rollout shape가 바뀌면 fail-closed parser/harness가 새 shape를 거부할 수 있다.
- trust TUI와 provider availability는 외부 상태에 민감하므로 향후 재실행에서 일시 실패할 수 있다.

## Iteration history

1. **초기 감사 — FAIL/code blocker:** POSIX Claude exec 실패 시 provisional row 잔존을 overlay RED로 재현했다.
2. **코드 delta — blocker 해소:** exact rollback, Codex/Claude regression, tmux sole-owner guard, OPS race가 모두 통과했다.
3. **외부 인증 감사 — FAIL/GAP:** direct Claude OAuth 만료로 AC11~14가 미충족이었다.
4. **GLM remediation 1차 — AC11 evidence GAP:** AC12~14 raw JSONL은 PASS했으나 AC11 PASS는 `-v` 전사만 있어 canonical artifact contract를 충족하지 못했다.
5. **최종 재감사 — PASS:** `m5-reg-ac11-glm.jsonl`이 named PASS/no-fail/no-skip/no-`NOT_RUN`, nonce/receipt 및 credential 비노출을 충족했다. 활성 finding은 없다.

## Recommendation

scoped implementation은 병합 가능하다. 다음 단계는 코드 추가 수정이 아니라 integration-branch CI와 일반 merge workflow다. coverage/Windows LIVE는 별도 품질 개선 또는 플랫폼 검증 항목으로 추적하되 이 SPEC의 이미 통과한 scoped AC를 다시 열 필요는 없다.
