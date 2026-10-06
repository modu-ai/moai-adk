# moai-proxy 다음 세션 핸드오프

작성일: 2026-09-10  
대상 저장소: `/Users/goos/MoAI/moai-adk-go`  
제안 SPEC ID: `SPEC-MOAI-PROXY-001`  
제안 worktree 이름: `moai-proxy-unified`  
제안 worktree 브랜치: `WT-moai-proxy-unified`

## 1. 다음 세션의 목표

다음 세션은 기존 설계 보고서를 근거로 `SPEC-MOAI-PROXY-001`의 plan 산출물을 작성하고 독립 plan 감사를 완료한다. 이 세션에서는 아직 구현을 시작하지 않는다. 구현 착수 전에는 다음 두 조건을 충족해야 한다.

1. 공통 loopback proxy를 통해 하나의 Claude Code 세션에서 Claude → GPT → GLM → Claude 전환을 검증할 수 있는 실행 가능한 시험 설계가 SPEC에 포함되어야 한다.
2. plan 감사 결과와 구현 범위를 사용자에게 보고하고 Implementation Kickoff Approval을 받아야 한다.

기준 문서:

- 읽기 전용 설계 원문: `/Users/goos/MoAI/moai-adk-go/reports/moai-proxy-three-provider-redesign-20260910.md`
- 읽기 전용 HTML 보고서: `/Users/goos/MoAI/moai-adk-go/reports/moai-proxy-three-provider-redesign-20260910.html`

위 두 파일은 현재 primary checkout의 untracked 파일이다. 새 worktree에 자동으로 복사된다고 가정하지 말고, 절대경로에서 읽기만 한다. 새 worktree의 SPEC 산출물 외에는 이 파일들을 수정하거나 이동하지 않는다.

## 2. 확정된 제품 결정

다음 사항은 다시 선택지로 되돌리지 않는다.

- 공개 실행 명령은 `moai cc`, `moai gpt`, `moai glm` 세 개만 둔다.
- `moai gg`와 `moai cg`는 제거한다.
- 세 명령 모두 Claude Code를 실행하고 동일한 Claude/GPT/GLM `/model` 선택 목록을 제공한다.
- 실행 명령의 차이는 최초 선택 모델뿐이다.
- `moai gpt`의 기본 모델은 `gpt-5.6-sol`이다.
- GPT 후보는 `gpt-5.6-sol`, `gpt-5.6-terra`, `gpt-5.6-luna`, `gpt-6-astra`이다.
- `moai gpt login`은 PKCE 로그인과 MoAI 전용 credential 저장을 담당한다.
- `moai gpt logout`은 MoAI가 소유한 credential만 지우며 Codex CLI 인증은 유지한다.
- `moai glm`의 기존 사용자 경험은 보존하되 공통 proxy와 공통 picker 위에서 동작하게 한다.
- provider 간 암묵적 fallback과 예상하지 못한 API 과금 경로를 두지 않는다.
- `/model`에서 세션에만 적용할 때는 `s`를 사용한다. Enter 또는 `/model <name>`이 `~/.claude/settings.json` 기본값을 바꿀 수 있음을 문서와 시험에 반영한다.

## 3. 현재 실측 기준선

핸드오프 작성 세션에서 다음을 관측했다.

```text
git rev-parse --short HEAD
2213871af

git branch --show-current
main

git fetch origin main && git rev-list --count --left-right origin/main...HEAD
0    0

moai session current
01a0896e-4701-7cb0-bd4e-7064dabb0e26
```

추가 관측:

- `SPEC-MOAI-PROXY-001`은 작성 시점에 존재하지 않았다.
- `.claude/worktrees/moai-proxy-unified`도 작성 시점에 존재하지 않았다.
- primary checkout에는 이 과제와 무관한 tracked/untracked 사용자 변경이 다수 있다. 새 세션은 이를 수정·정리·stage·stash·reset하지 않는다.
- 현행 `moai cc --help`에는 `-w, --worktree [name]`이 있고, 짧은 이름을 주면 `.claude/worktrees/<name>/`에서 Claude Code를 시작한다.
- 현행 CLI에는 아직 `moai gpt`가 없다. 따라서 구현 전 새 세션 진입은 `moai cc`를 사용한다.

이 기준선은 다음 세션의 현재 상태를 보증하지 않는다. 다음 세션은 반드시 다시 측정한다.

## 4. 새 터미널에서의 진입 절차

primary checkout에서 다음 명령 하나로 새 세션을 시작한다.

```bash
moai cc -w moai-proxy-unified
```

세션이 열린 직후, 어떤 파일도 쓰기 전에 다음을 수행한다.

```bash
git rev-parse --show-toplevel
git branch --show-current
git rev-parse --short HEAD
git fetch origin main 2>&1
git rev-list --count --left-right origin/main...HEAD
git status --short
```

기대 조건:

- 저장소 루트가 `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/moai-proxy-unified`여야 한다.
- 새 worktree의 출발점은 `origin/main`이어야 한다.
- ahead/behind 결과는 `0 0`이어야 한다.
- 새 worktree 브랜치는 즉시 `git branch -m WT-moai-proxy-unified`로 이름을 바꾼다.

위 조건이 다르면 파일을 수정하지 말고 차이를 보고한다. bare `git worktree add`, primary checkout의 `git switch`, `git checkout`, `git stash`, `git reset --hard`는 사용하지 않는다.

## 5. plan 단계에서 만들어야 할 것

새 worktree 안에서 다음 파일을 작성한다. 기존 MoAI SPEC 템플릿과 소유권 규칙을 먼저 확인하고 그 형식에 맞춘다.

- `.moai/specs/SPEC-MOAI-PROXY-001/spec.md`
- `.moai/specs/SPEC-MOAI-PROXY-001/plan.md`
- `.moai/specs/SPEC-MOAI-PROXY-001/acceptance.md`
- Tier L 판정 시 요구되는 `research.md`, `design.md`
- 감사 증거: `.moai/reports/SPEC-MOAI-PROXY-001/verdict.md`

SPEC에는 최소한 다음 설계 축을 포함한다.

1. CLI 표면: `cc·gpt·glm` 등록, `gg·cg` 제거, 기존 flag 호환성.
2. 공통 supervisor: loopback 서버 시작, Claude child 실행, signal 전달, 종료·정리, 포트 경쟁 처리.
3. Anthropic ingress 정규화: Messages 요청, SSE, tool use/result, reasoning/history 보존.
4. provider adapter: Anthropic passthrough, OpenAI GPT, ZAI GLM의 단일 route와 정확한 model ID 전달.
5. 인증: MoAI 소유 GPT PKCE credential, logout 소유권 경계, Claude/Codex/GLM 인증 불변성.
6. model registry와 picker: 세 launcher에서 동일한 3사 목록, launcher별 초기 선택, `s` 세션 전용 안내.
7. migration: 기존 `team_mode: cg`를 임의로 다른 모드로 바꾸지 않고 doctor에서 `moai cc|gpt|glm` 명시 선택을 요구.
8. 선택적 `cg` 제거: live runtime·현재 문서·배포 템플릿·현재 테스트만 갱신하고 과거 SPEC·release note·감사 증거는 보존.
9. 보안: loopback bind, bearer/header 유출 방지, 로그 마스킹, credential 파일 권한, prompt/request 기록 정책.
10. 실패 의미론: 인증 없음, 허용되지 않은 model ID, upstream 오류, SSE 중단, child 비정상 종료에서 조용한 fallback 금지.

## 6. 필수 수용 시험

설계 보고서의 T01~T20을 acceptance 파일에 검증 가능한 Given/When/Then 또는 동등한 계약으로 옮긴다. 다음 핵심 항목을 축약하거나 삭제하지 않는다.

- T01: 세 launcher의 초기 모델이 각각 기대값과 일치한다.
- T02: 세 launcher에서 동일한 3-provider picker가 보인다.
- T03: 같은 `moai cc` 세션에서 Claude → GPT → GLM → Claude를 `/model`과 `s`로 전환한다.
- T04: 각 provider에서 tool call과 tool result가 왕복한다.
- T05~T06: GPT PKCE login과 MoAI 소유 credential만 지우는 logout.
- T07: 네 GPT model ID 각각의 실제 접근 가능 여부를 구분해 보고한다.
- T08: provider 실패 시 조용한 유료 API fallback이 없다.
- T09~T11: Claude 구독 경로, 외부 인증, 부모 env와 전역 settings 불변성.
- T12~T14: supervisor 생명주기, tool 동시성, SSE 오류.
- T15~T18: history/reasoning, context/input, subagent/background, hooks/MCP/compact 회귀.
- T19: credential·로그·loopback 보안.
- T20: `gg·cg` 제거와 기존 `cc·glm` 회귀.

실측하지 못한 항목은 PASS로 쓰지 않는다. 외부 유료 호출, 실제 계정 로그인, Windows 검증처럼 현재 환경에서 수행하지 못한 것은 Gap으로 남긴다.

## 7. 구현 단계의 우선순위와 중단 조건

Implementation Kickoff Approval을 받은 뒤에만 run 단계로 전환한다. 권장 순서는 다음과 같다.

1. High — T03의 same-session `/model` 전환을 loopback mock으로 먼저 재현한다.
2. High — CLI 계약과 model registry를 테스트 우선으로 구현한다.
3. High — supervisor와 provider-neutral ingress를 구현한다.
4. High — Anthropic/OpenAI/ZAI adapter와 인증 소유권 경계를 구현한다.
5. Medium — live `cg` runtime과 현재 배포 문서·템플릿을 선택적으로 제거한다.
6. Medium — T01~T20 범위의 표적 시험과 실패 경로를 수행한다.
7. Medium — sync 감사 후에만 문서 동기화와 배포 판단을 한다.

다음 중 하나가 발생하면 추측으로 계속하지 말고 중단해 보고한다.

- Claude Code가 unknown model ID를 한 번 전달하는 것과 달리, 같은 세션의 `/model` 변경을 proxy 요청에 반영하지 않는다.
- GPT PKCE token이 현재 사용하려는 upstream에서 해당 model ID에 접근할 수 없다.
- Claude Code 내부 schema 또는 tool semantics를 손실 없이 provider-neutral 형태로 매핑할 수 없다.
- `cg` 제거 대상과 역사 기록 보존 대상의 경계가 파일 단위로 모호하다.
- primary checkout 또는 worktree 기준선이 사용자 변경과 충돌한다.

## 8. 금지 사항

- `gpt-5.3`을 후보에 넣지 않는다.
- `gpt-6-astra`를 기본값으로 두지 않는다.
- `moai gg`, `moai cg`, `--provider mixed`, `claude_glm`을 새 API로 되살리지 않는다.
- Codex CLI credential을 읽기·수정·삭제하는 logout을 구현하지 않는다.
- 성공하지 않은 인증이나 model access를 지원된다고 문서화하지 않는다.
- 텍스트 검색 결과만으로 결함·완료를 단정하지 않는다.
- 과거 SPEC, release note, audit evidence를 일괄 치환하지 않는다.
- `git add -A`, `git add .`, `git commit -a`를 사용하지 않는다.
- 명시 지시 전에는 push, PR 생성, merge, worktree 제거를 하지 않는다.

## 9. 다음 세션에 붙여 넣을 본문

아래 블록은 새 worktree 세션에 그대로 붙여 넣는다. 첫 줄의 실행 명령은 새 터미널 진입용이며, 이미 해당 세션 안이라면 다시 실행하지 않는다.

✂──── 여기부터 복사 ────✂

[새 터미널 — WORKTREE에서 시작]
$ moai cc -w moai-proxy-unified
   └─ 현행에는 moai gpt가 아직 없으므로 최초 진입은 moai cc를 사용한다. 세션이 열리면 파일 쓰기 전에 `git branch -m WT-moai-proxy-unified`를 수행한다.

ultrathink. SPEC-MOAI-PROXY-001 plan 진입.
applied lessons: /Users/goos/MoAI/moai-adk-go/reports/moai-proxy-three-provider-redesign-20260910.md, /Users/goos/MoAI/moai-adk-go/reports/moai-proxy-next-session-handoff-20260910.md
source_session_id: 01a0896e-4701-7cb0-bd4e-7064dabb0e26

전제 검증:
0) `git rev-parse --show-toplevel` → `.claude/worktrees/moai-proxy-unified`로 끝나는 경로
1) `git branch --show-current` → `WT-moai-proxy-unified`
2) `git fetch origin main && git rev-list --count --left-right origin/main...HEAD` → `0 0`
3) 두 기준 보고서를 절대경로에서 읽고, primary checkout의 dirty 파일은 변경하지 않았음을 확인

실행: `/moai plan SPEC-MOAI-PROXY-001` — cc·gpt·glm 공통 proxy와 세션 내 `/model` 전환을 설계하고 plan 감사 뒤 구현 승인문 앞에서 멈춘다.

후속: 승인 후 같은 worktree에서 `/moai run SPEC-MOAI-PROXY-001`; push·PR·merge·worktree 제거는 별도 지시가 있을 때만 수행한다.

✂──── 여기까지 복사 ────✂

## 10. 다음 세션의 보고 형식

각 단계 종료 보고에는 다음 다섯 구획을 둔다.

1. Claim — 실제로 주장할 수 있는 상태.
2. Evidence — 이번 세션에서 실행한 명령과 관측한 원문 출력.
3. Baseline-attribution — 어떤 tree, branch, HEAD, 환경을 기준으로 측정했는지.
4. Gaps — 수행하지 않았거나 관측하지 못한 항목.
5. Residual-risk — 관측을 마쳐도 남는 위험.

빈 Gaps 구획은 “아무것도 남지 않았다”는 강한 주장이다. 실제로 모든 항목을 관측한 경우가 아니면 비워 두지 않는다.
