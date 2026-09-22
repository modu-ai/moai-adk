# t1078 독립 sync 감사

## 판정

- Overall: **FAIL / un-PASS**
- 코드 findings: **0건**
- Evidence gap: **1건 — AC-LMD-012 실제 Codex LIVE 미완료**
- 분류: 외부 실행 승인 경계에서 차단된 수용 증거 GAP이며, 확인된 코드 결함이 아니다.

## Claim

독립 감사자는 `07f3b78d448` 위 uncommitted 구현을 SPEC-CODEX-LOCALMD-001과 대조했다. AC-LMD-001~011의 로컬 계약에서 merge-blocking 코드 결함을 찾지 못했다. 그러나 `acceptance.md`가 AC-LMD-012 `NOT_RUN`/미완료 시 전체 PASS를 금지하므로 카드 완료 판정은 거부한다.

## Evidence

독립 감사 관측:

- scoped Codex selector PASS
- 안전 open / same-descriptor / 교체 경합 / non-regular / prelaunch 차단의 37개 하위 셀 PASS
- 다섯 operator override 표기, direct/spawn 크기 경계와 quote 팽창, 모든 launcher funnel, 입력 불변 PASS
- race 3회, lint, govulncheck, Windows cross-build PASS
- `moai verify check --key-current`: `fresh=false`, snapshot 미기록

부모 세션의 현재 트리 재측정:

```text
$ unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1078-root-verify-home GOCACHE=/tmp/t1078-root-verify-cache go test ./internal/cli -run '^(TestCodexLocal|TestCodexInstructionContract)' -count=1 -timeout=90s
ok  github.com/modu-ai/moai-adk/internal/cli  2.261s

$ unset ANTHROPIC_BASE_URL ANTHROPIC_API_KEY ANTHROPIC_AUTH_TOKEN Z_AI_API_KEY && MOAI_HOME=/tmp/t1078-root-race-home GOCACHE=/tmp/t1078-root-race-cache go test -race ./internal/cli -run '^(TestCodexLocal|TestCodexInstructionContract)' -count=1 -timeout=90s
ok  github.com/modu-ai/moai-adk/internal/cli  3.705s

$ git diff --check
<empty>; exit 0
```

AC-LMD-012 시도:

- built binary: `/tmp/t1078-moai-live`
- SHA-256: `e0116af703c574e8023a639a2c5e4785aa69434b0a5d3eb8cb10c02a0ead596c`
- 1차: exit 130, 독립 fixture의 `.codex/hooks.json`/`.codex/config.toml` 부재로 init 요구; child Codex 응답 전 종료
- 같은 binary로 fixture init 후 `CLAUDE.local.md`와 `AGENTS.local.md` 해시 불변 확인
- 2차 sandbox: exit 1, `failed to initialize in-process app-server client: Operation not permitted`
- sandbox 밖 재실행: 두 fixture 지침의 외부 Codex 서비스 전송에 대한 명시 승인이 없어 승인 단계에서 거부됨; 우회하지 않음

## AC 판정

| AC | 판정 | 근거 |
|---|---|---|
| AC-LMD-001~005 | PASS | 두 파일 합성·순서·matrix·충돌·safe-open scoped 테스트 |
| AC-LMD-006~009 | PASS | large UTF-8, direct/spawn prelaunch bound, 입력 불변, fresh read scoped 테스트 |
| AC-LMD-010~011 | PASS | help/template와 all-funnel scoped 테스트 |
| AC-LMD-012 | GAP / UNVERIFIED | 실제 인증 Codex 응답과 session log 쌍을 얻지 못함 |

## Baseline-attribution

- 코드 기준선: `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1078`, branch `WT-codex-local-md`, HEAD `07f3b78d448` + 본 카드 uncommitted 구현
- 로컬 재측정은 위 절대 WT에서 실행했다.
- LIVE fixture는 `/tmp/t1078-live-fixture-01a0c58d`; 리포의 실제 local instruction 파일을 전송하지 않았다.

## Gaps

- 실제 Codex 응답 JSON과 Codex session log에서 두 nonce/source 쌍을 확인하지 못했다.
- Windows 및 unsupported 플랫폼 runtime은 실행하지 않았다. cross-build는 runtime 증거가 아니다.
- `verify` fresh snapshot은 없다.

## Residual-risk

코드 레벨 합성 계약이 통과해도 외부 Codex가 `developer_instructions`를 실제 세션에 동일하게 적용한다고 단정할 수 없다. AC-LMD-012의 승인된 LIVE가 exit 0이고 응답·session log·입력 해시 불변을 모두 만족하기 전에는 카드 완료나 develop 병합을 승인하지 않는다.
