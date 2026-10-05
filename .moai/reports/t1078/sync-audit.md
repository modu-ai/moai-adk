# Sync Audit — t1078

- SPEC: `SPEC-CODEX-LOCALMD-001`
- 감사 대상: `WT-codex-local-md@6af5dc2332c4cb7837608150c6f755072ae9423f` + sync-docs working-tree delta
- 프로필: `default` (flat weighted percentage)
- Overall Verdict: **PASS**
- Score: **100/100**
- 코드 findings: **0**
- evidence gaps: **0**
- non-blocking cache misses: **1**
- 판정 요약: AC-LMD-001~012는 모두 PASS이고 구현 결함은 발견되지 않았다. shared snapshot의 `fresh:false`는 검증 생략을 허용하지 않는 cache miss이며, 이번 감사는 scoped regression을 직접 재실행했으므로 PASS 근거를 충족한다.

## Dimension Scores

| Dimension | Score | Verdict | Evidence |
|---|---:|---|---|
| Functionality (40%) | 100/100 | PASS | scoped regression `go test` exit 0; LIVE output와 rollout에서 두 nonce/source 쌍 및 `turn.completed` 직접 확인 |
| Security (25%) | 100/100 | PASS | 이전 독립 감사의 no-follow/same-descriptor TOCTOU·비정규 파일·`govulncheck` 증거와 committed implementation의 좁은 회귀 PASS |
| Craft (20%) | 100/100 | PASS | 이전 감사에서 변경 함수 96.4%~100%, `golangci-lint` 0 issues; 현재 committed tree scoped regression 재통과 |
| Consistency (15%) | 100/100 | PASS | 구현 HEAD는 고정돼 있고 현재 uncommitted delta는 sync 문서·progress·감사 보고서뿐이다. 단일 funnel·상수·템플릿 계약의 관련 테스트가 재통과했다. |

가중 점수는 `40 + 25 + 20 + 15 = 100`이다.

## Findings

`[]` — 코드 결함 및 blocking evidence finding 없음.

shared snapshot miss는 finding이 아니다. `moai verify --help`가 이를 sibling verifier의 재사용 캐시로 정의하고, stale이면 consumer가 check를 재실행하도록 명시한다. 이번 consumer는 E2의 scoped regression을 실제 재실행했다.

## Per-AC Verdict

| AC | Verdict | 직접 또는 회귀 증거 |
|---|---|---|
| AC-LMD-001 | PASS | bare/cli/app 단일 payload 관련 scoped regression PASS |
| AC-LMD-002 | PASS | spawn/`-w`/factory funnel 관련 scoped regression PASS |
| AC-LMD-003 | PASS | provenance 및 CLAUDE-first/AGENTS-later 합성 테스트 PASS; LIVE rollout 입력에서도 동일 순서 확인 |
| AC-LMD-004 | PASS | absent/empty matrix 테스트 PASS |
| AC-LMD-005 | PASS | 이전 독립 감사에서 양 파일의 Unix no-follow, same-descriptor replacement, open/stat/read fail-closed 테스트 PASS; Windows cross-build PASS. Windows 실제 runtime은 Residual-risk에 남긴다. |
| AC-LMD-006 | PASS | 61,360-byte UTF-8 body slice/hash 및 fresh-read 테스트 PASS |
| AC-LMD-007 | PASS | 다섯 override 표기와 음성 대조 테스트 PASS |
| AC-LMD-008 | PASS | direct/spawn 경계 및 quote-expansion 테스트 PASS |
| AC-LMD-009 | PASS | 모든 funnel의 fixture input identity/hash 및 rename count 0 테스트 PASS; LIVE fixture/source 해시 불변 확인 |
| AC-LMD-010 | PASS | 두 번째 launch fresh-read 테스트 PASS |
| AC-LMD-011 | PASS | help 및 template pin 테스트 PASS |
| AC-LMD-012 | PASS | exact production command exit 0; output·rollout 양쪽의 exact nonce/source 쌍, `turn.completed`, fixture/source 불변 확인 |

## TRUST 5

| Pillar | Verdict | 근거 |
|---|---|---|
| Tested | PASS | AC-LMD-001~012 PASS, 현재 committed tree scoped regression PASS |
| Readable | PASS | 이전 독립 lint 0 issues, 명명된 상수·진단·작은 플랫폼 helper 구조 |
| Unified | PASS | 단일 launch funnel과 기존 error/launch seam 패턴 유지 |
| Secured | PASS | leaf no-follow, opened-descriptor fstat/read, fail-closed, dependency scan 0 callable vulnerabilities |
| Trackable | PASS | SPEC/AC/commit/LIVE thread·rollout·hash traceability가 존재한다. snapshot cache miss는 직접 재실행 증거로 대체됐다. |

## Claim

`6af5dc233` 구현은 AC-LMD-001~012를 충족한다. 특히 실제 Codex 세션이 두 로컬 instruction의 nonce와 source filename을 정확히 반환했고 rollout에도 같은 입력과 최종 응답이 남았다. 코드 finding은 0건이다. current-key snapshot은 없지만 cache 재사용을 생략하고 scoped regression을 직접 실행했으므로 최종 verdict는 PASS다.

## Evidence

### E1 — 현재 baseline

명령:

```sh
git rev-parse --show-toplevel
git rev-parse HEAD
git branch --show-current
git status --short
```

관측 출력:

```text
/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1078
6af5dc2332c4cb7837608150c6f755072ae9423f
WT-codex-local-md
```

첫 AC/LIVE 재감사 시점의 `git status --short` 출력은 비어 있었다. 이후 sync-docs worker가 plan §F 소유 문서와 progress를 수정했다. G1 재검토 시 현재 상태는 다음과 같았고 구현 코드는 바뀌지 않았다.

```text
 M .moai/reports/t1078/sync-audit.md
 M .moai/specs/SPEC-CODEX-LOCALMD-001/progress.md
 M AGENTS.md
 M CHANGELOG.md
 M docs-site/content/en/advanced/codex-dual-harness.md
 M docs-site/content/ja/advanced/codex-dual-harness.md
 M docs-site/content/ko/advanced/codex-dual-harness.md
 M docs-site/content/zh/advanced/codex-dual-harness.md
```

### E2 — AC-LMD-001~011 좁은 회귀

명령:

```sh
unset CLAUDECODE CODEX_THREAD_ID ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY
export MOAI_HOME=/tmp/t1078-final-audit-home
export GOCACHE=/tmp/t1078-final-audit-cache
go test ./internal/cli -run '^TestCodex(LocalInstructions|InstructionContract)' -count=1
```

관측 출력, exit 0:

```text
ok  	github.com/modu-ai/moai-adk/internal/cli	2.394s
```

### E3 — LIVE binary와 입력/source 해시

명령:

```sh
shasum -a 256 /tmp/t1078-moai-live /tmp/t1078-live-fixture-01a0c58d/CLAUDE.local.md /tmp/t1078-live-fixture-01a0c58d/AGENTS.local.md AGENTS.md CLAUDE.md CLAUDE.local.md
```

관측 출력:

```text
e0116af703c574e8023a639a2c5e4785aa69434b0a5d3eb8cb10c02a0ead596c  /tmp/t1078-moai-live
5c7658929a91562abc8ffae06c3117435eb02808455ba8783311789c7216fb0a  /tmp/t1078-live-fixture-01a0c58d/CLAUDE.local.md
24fdc86b2b5107b471d41a5e29c78b139ba08c8a43cb62b17372a90c2756ce85  /tmp/t1078-live-fixture-01a0c58d/AGENTS.local.md
3d6aa759e52870f2bc2689a5920845c646c2682c7e546bbb6fa7e6154bea2dbe  AGENTS.md
bc9b57c7a0ea55313889dbfda21d243b6d0166e3e176dec93c3c500a58496de4  CLAUDE.md
089b962a1b760b52da18d02576acd54a6b0cf467f2bd41dc410ce22d1d98cd04  CLAUDE.local.md
AGENTS.local.md ABSENT
```

운영자가 제공한 LIVE pre/post 값과 직접 판독한 당시 post 값이 일치했다. fixture 두 파일 및 source WT 네 입력의 상태는 LIVE 전후 불변이다. 그 뒤 `AGENTS.md`는 plan §F에 따른 sync 문서 작업으로 의도적으로 수정됐으며, 이는 launcher의 입력 변조와 구분된다.

### E4 — AC-LMD-012 production LIVE

실행 명령:

```sh
cwd=/tmp/t1078-live-fixture-01a0c58d
/opt/homebrew/bin/gtimeout 120 /tmp/t1078-moai-live codex -- exec --json --skip-git-repo-check '<two-nonce JSON prompt>'
```

환경 scrub 후 실행됐고 exit 0이었다. `/tmp/t1078-live-output.jsonl`의 관측 출력:

```jsonl
{"type":"thread.started","thread_id":"01a0c861-5d93-7540-9bbe-85d7218e4c2a"}
{"type":"turn.started"}
{"type":"item.completed","item":{"id":"item_1","type":"agent_message","text":"{\"claude\":{\"nonce\":\"t1078-claude-01a0c58d-20260922-a\",\"source\":\"CLAUDE.local.md\"},\"agents\":{\"nonce\":\"t1078-agents-01a0c58d-20260922-b\",\"source\":\"AGENTS.local.md\"}}"}}
{"type":"turn.completed","usage":{"input_tokens":33108,"cached_input_tokens":6656,"cache_write_input_tokens":0,"output_tokens":67,"reasoning_output_tokens":0}}
```

rollout 직접 판독 경로:

```text
/Users/goos/.codex/sessions/2026/09/22/rollout-2026-09-22T18-10-19-01a0c861-5d93-7540-9bbe-85d7218e4c2a.jsonl
```

직접 검색에서 developer input에 다음 순서가 관측됐다.

```text
<!-- source: CLAUDE.local.md -->
CLAUDE_NONCE=t1078-claude-01a0c58d-20260922-a
<!-- source: AGENTS.local.md -->
AGENTS_NONCE=t1078-agents-01a0c58d-20260922-b
```

같은 rollout의 final answer:

```json
{"claude":{"nonce":"t1078-claude-01a0c58d-20260922-a","source":"CLAUDE.local.md"},"agents":{"nonce":"t1078-agents-01a0c58d-20260922-b","source":"AGENTS.local.md"}}
```

### E5 — shared snapshot

명령:

```sh
moai verify check --key-current
```

관측 출력, exit 1:

```text
{
  "fresh": false,
  "key": "6af5dc2332c4cb7837608150c6f755072ae9423f:6e340b9cffb37a98",
  "reason": "no snapshot recorded for the current working-tree key"
}

ERROR

Stale: no snapshot recorded for the current working-tree key.
```

### E6 — snapshot 계약 재판독

명령:

```sh
/Users/goos/go/bin/moai verify --help
```

관측 출력:

```text
Shared diagnostic snapshot contract.

Quality-check results (tests, lint, coverage, ...) recorded under the current
working-tree key can be reused by sibling verification layers instead of
re-executing, under a strict freshness rule: the stored key must equal the key
recomputed from the current tree AND the recorded timestamp must be within the
TTL (default 10 minutes). A stale snapshot is never citable evidence — on any
mismatch the consumer re-executes the check.

Verbs:
  verify record   record one executed check result under the current-tree key
  verify check    freshness query — exit 0 fresh / exit 1 stale
```

해석: snapshot은 검사를 대체해 재사용하는 캐시다. stale 상태 자체가 검증 실패를 뜻하지 않으며, consumer가 직접 검사를 재실행하면 된다. E2에서 바로 그 재실행을 관측했다. SPEC의 AC-LMD-001~012와 Definition of Done에는 current-key snapshot 존재 조건이 없다.

## Baseline-attribution

- 코드/AC 회귀 기준선: `WT-codex-local-md@6af5dc2332c4cb7837608150c6f755072ae9423f`. scoped regression 실행 뒤 추가된 delta는 sync 문서·progress·감사 보고서뿐이다.
- LIVE 기준선: binary sha256 `e0116af703c574e8023a639a2c5e4785aa69434b0a5d3eb8cb10c02a0ead596c`, fixture `/tmp/t1078-live-fixture-01a0c58d`, thread `01a0c861-5d93-7540-9bbe-85d7218e4c2a`.
- 이전 독립 감사의 security/craft 세부 측정은 같은 구현 bytes가 commit된 직전 worktree에서 수행됐다. 이번 재감사는 새 HEAD의 관련 회귀와 LIVE delta를 다시 측정했다.

## Gaps

- 필수 evidence gap 없음.
- current working-tree key의 shared verification snapshot은 없다. 이는 sibling verifier 재사용 최적화의 non-blocking cache miss이며 E2의 직접 재실행으로 검증 공백을 남기지 않았다.
- Windows cross-build는 통과했지만 Windows 실제 runtime/symlink 권한 환경은 실행하지 않았다. acceptance의 platform gap으로 보존하며 현재 host 기능 PASS를 뒤집지 않는다.
- cross-model backend audit 및 receipt는 이 감사에서 실행하지 않았다. 프로젝트가 required codex audit gate를 명시한 증거는 관측하지 않았다.

## Residual-risk

- LIVE 한 세션의 PASS는 향후 Codex 버전·config 변경까지 보장하지 않는다.
- Windows reparse-point 동작은 compile evidence만 있고 실제 Windows runtime 증거는 없다.
- shared snapshot이 생성되기 전에는 다른 verifier가 동일한 current-key 결과를 캐시로 재사용할 수 없어 검사를 다시 실행해야 한다.

## Recommendations

- 코드 수정은 필요 없다.
- 선택 사항: sibling verifier의 중복 실행을 줄이려면 workflow가 정한 snapshot producer로 current-key 결과를 기록한다. PASS의 선행조건은 아니다.

## Iteration History

| Iteration | 결과 | 변경점 |
|---|---|---|
| 1 | FAIL / un-PASS | AC-LMD-012가 외부 실행 승인 경계에서 UNVERIFIED였고, current-key snapshot도 없었다. 코드 finding은 0건이었다. |
| 2 | FAIL, 100/100 | production LIVE exit 0과 output/rollout/hash 증거로 AC-LMD-012가 PASS가 됐다. 남은 유일한 gap은 새 clean HEAD의 shared snapshot 부재다. |
| 3 | PASS, 100/100 | `moai verify --help`를 직접 판독해 snapshot이 선택적 재사용 캐시임을 확인했다. scoped regression은 이미 직접 재실행됐고 SPEC AC/DoD에 snapshot hard gate가 없으므로 Iteration 2의 G1 blocking 판정을 철회했다. |

AUDIT-VERDICT: PASS spec=SPEC-CODEX-LOCALMD-001 receipts=none
