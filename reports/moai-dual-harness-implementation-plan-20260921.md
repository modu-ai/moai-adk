# MoAI-ADK Claude Code + Codex 이중 하네스 상세 구현계획

> Plan / Expert · 2026-09-21 · 코드 변경 없이 구현 순서와 검증 계약을 정의한다.

## 1. 목표와 완료 정의

### 목표

Claude Code 중심으로 구성된 현재 MoAI-ADK를 공통 코어와 두 개의 호스트 어댑터로 재구성한다. 정책·Skill·Agent 정의는 한 번만 작성하고 Claude Code와 Codex가 각자의 네이티브 탐색 규칙으로 소비하도록 생성한다.

### 완료 정의

1. `moai init --harness claude|codex|both`가 선택한 하네스에 필요한 파일만 배포한다.
2. 공통 안전 정책의 유일한 저작 원본이 하네스 중립 경로에 존재한다.
3. Claude/Codex Skill의 공통 본문 digest가 일치한다.
4. 모든 MUST 조항은 persistent instruction, Hook/guard, 또는 검증 테스트를 갖는다. 안전 조항이 Skill에만 존재하는 경우는 0개다.
5. Claude↔Codex 알림 왕복과 readback이 검증된다. 메시지는 권한 승인이나 완료 상태를 직접 바꾸지 않는다.
6. persistent card worktree의 lifecycle owner는 MoAI 하나이며, host-native worktree는 ephemeral 용도로 분리된다.
7. 기존 `--llm claude|gpt|both` 프로젝트는 데이터 손실 없이 업데이트된다.

### 비목표

- Claude와 Codex의 모든 UI·도구 이름을 동일하게 만드는 것
- Codex에서 지원하지 않는 Claude 기능을 지원한다고 가장하는 것
- 기존 83개 rule을 기계적으로 전부 Skill로 변환하는 것
- 이번 계획 단계에서 실제 템플릿이나 Go 코드를 변경하는 것

## 2. 현재 기반과 재사용 범위

| 현재 자산 | 상태 | 계획 |
|---|---|---|
| `moai init --llm claude|gpt|both` | 동작 확인 | `--harness`로 명칭 정리, 기존 플래그 호환 유지 |
| `internal/template/templates/AGENTS.md` | 공통 계약 존재 | 중립 policy source에서 생성하도록 전환 |
| `internal/template/templates/CLAUDE.md` | Claude overlay 존재 | `AGENTS.md` 참조 + Claude 전용 내용만 유지 |
| `internal/template/skill_mirror.go` | `.claude/skills` → `.agents/skills` symlink/copy | 과도기 재사용, 중립 source emitter로 교체 |
| `internal/template/agentemit` | Claude agent → Codex TOML 변환 | neutral agent manifest 방식으로 확장 |
| `.codex/config.toml`, `.codex/hooks.json` | MCP·Hook 연결 | capability manifest와 doctor 검증 추가 |
| `internal/sessionmsg` | register/list/send/poll broker | 공통 fallback transport로 감싼다 |
| `internal/cli/worktree`, launcher | MoAI worktree 기능 | lifecycle-owner 계약을 추가한다 |
| `internal/template/templates/.claude/rules` 83개 | 약 1.2MB | 조항 단위로 분류하고 carrier를 재배치 |

현재 구조를 폐기하지 않는다. `templates/`가 `go:embed`되는 배포 트리라는 계약을 유지하면서, 그 앞에 canonical source와 deterministic emitter를 추가한다.

## 3. 목표 디렉터리 구조

`internal/template`에는 이미 배포용 `catalog.yaml`이 있으므로 새 원본 디렉터리는 혼동을 피하기 위해 `canonical/`로 명명한다.

```text
internal/template/
├── canonical/                         # 사람이 편집하는 유일한 의미 원본
│   ├── policy/
│   │   ├── standing/                  # 모든 하네스·턴에 적용
│   │   ├── scoped/                    # selector가 있는 정책
│   │   ├── host/                      # claude/codex 전용
│   │   └── coverage.yaml              # clause → carrier/guard/test
│   ├── skills/<skill>/                # 공통 Skill 원본
│   ├── agents/<role>.yaml             # 공통 Agent 의미
│   └── capabilities.yaml              # 버전별 기능·폴백
├── policyemit/                         # policy emitter
├── skillemit/                          # dual Skill emitter
├── agentemit/                          # 기존 emitter 확장
├── templates/                          # 생성된 배포 트리, go:embed 입력
│   ├── AGENTS.md
│   ├── CLAUDE.md
│   ├── .claude/{rules,skills,agents}/
│   ├── .agents/skills/
│   └── .codex/{agents,config.toml,hooks.json}
└── embed.go                            # 기존 all:templates 계약 유지
```

배포 프로젝트:

```text
project/
├── AGENTS.md                           # 공통 standing policy
├── CLAUDE.md                           # @AGENTS.md + Claude overlay
├── .claude/
│   ├── rules/                          # Claude scoped/host policy
│   ├── skills/
│   ├── agents/
│   └── settings.json
├── .agents/skills/                     # Codex Skill 탐색 경로
├── .codex/
│   ├── agents/
│   ├── config.toml
│   └── hooks.json
└── .moai/
    ├── config/harness.yaml
    └── state/{session-msg,dispatch}/
```

## 4. 정책 분류와 Rules 처리

### 분류 규칙

| 클래스 | 의미 | Claude 출력 | Codex 출력 | 강제 장치 |
|---|---|---|---|---|
| `standing` | 매 턴 필요한 안전·권한 불변식 | `CLAUDE.md`가 `AGENTS.md` 참조 | root `AGENTS.md` | Hook/test 병행 |
| `scoped` | 특정 파일군/디렉터리 규칙 | `.claude/rules` paths/glob | 아래 Codex 변환 규칙 | 필요 시 path-aware Hook |
| `workflow` | 요청 시 실행할 반복 절차 | `.claude/skills` | `.agents/skills` | Skill contract test |
| `host` | 제품 고유 기능 | Claude 전용 rule/Skill | Codex 전용 rule/Skill | capability probe |
| `enforced` | 프롬프트만으로 부족한 금지/게이트 | Hook/CLI guard | Hook/CLI guard | negative test 필수 |
| `reference` | 긴 설명·체크리스트 | Skill `references/` | Skill `references/` | 없음 |

### Codex scoped-policy의 의미 차이

Claude의 path-scoped rule과 Codex의 nested `AGENTS.md`는 동등하지 않다. Codex의 instruction chain은 실행 CWD 기준이므로 root에서 작업하며 다른 경로의 파일을 수정할 때 nested AGENTS가 항상 적용된다고 가정하면 안 된다.

Codex 변환은 다음 우선순위를 따른다.

1. 위반 시 데이터 손실·권한 침해가 생기는 규칙은 root `AGENTS.md`와 Hook으로 승격한다.
2. 해당 디렉터리에서 실행하는 것이 작업 계약인 경우 nested `AGENTS.override.md`를 생성하고 launcher가 `codex -C <dir>`를 사용한다.
3. 작업 유형에 따른 절차라면 dual Skill로 이동한다.
4. 파일 경로를 보고 기계적으로 검사할 수 있으면 path-aware Hook으로 강제한다.
5. 위 네 방법으로 의미를 보존할 수 없으면 `coverage.yaml`에 `unsupported`와 잔여 위험을 기록하고 제거하지 않는다.

### coverage schema

```yaml
clauses:
  - id: POL-GIT-001
    source: shared-checkout.md#never-change-branch
    severity: hard
    class: enforced
    carriers:
      common: AGENTS.md
      claude: hook.pre_tool
      codex: hook.pre_tool
    tests:
      - internal/hook/branch_guard_test.go
    skill_only: false
```

빌드는 다음 경우 실패한다.

- `severity: hard`인데 persistent carrier가 없음
- `hard` 조항이 Skill에만 있음
- 원본 clause가 coverage에 등록되지 않음
- Claude/Codex 한쪽에서 의미가 빠졌는데 disposition이 없음
- scoped rule을 Codex nested AGENTS로 내보내면서 요구 CWD가 선언되지 않음

## 5. Skill 이중 배포 계획

### 목표 계약

- 공통 `SKILL.md`, scripts, references는 한 원본에서 생성한다.
- 하네스 전용 차이는 `host/claude.md`, `host/codex.md`처럼 명시한다.
- 같은 공통 파일의 digest는 두 출력에서 같아야 한다.
- 사용자 소유 디렉터리를 덮어쓰지 않는다.

### 단계적 전환

1. `canonical/skills`를 추가하고 기존 `.claude/skills`의 내용을 그대로 import한다.
2. `skillemit`이 `.claude/skills`를 생성하도록 만든다. 이 시점에는 현재 `skill_mirror.go`가 Codex 경로를 계속 만든다.
3. golden/hash 검증으로 기존 배포와 byte-equivalence를 확인한다.
4. 양쪽 경로를 build-time으로 각각 생성할지, neutral payload를 배포하고 두 경로를 symlink할지 플랫폼 probe로 결정한다.
5. Windows symlink 권한이 없을 때 copy fallback과 drift 경고를 유지한다.
6. 기존 runtime mirror를 제거하는 것은 두 방식의 E2E가 통과한 뒤다.

### 변경 파일

- 신규: `internal/template/canonical/skills/**`
- 신규: `internal/template/skillemit/{load,emit,hash,validate}.go`
- 신규: `internal/template/skillemit/*_test.go`
- 수정: `internal/template/skill_mirror.go`
- 수정: `internal/template/skill_mirror_*_test.go`
- 수정: `internal/template/catalog.yaml`

## 6. Agent 이중 배포 계획

현재 `agentemit`의 measured-field와 documented-drop 원칙은 유지한다. 다만 Claude `.md`를 원본으로 삼는 대신 neutral manifest를 원본으로 바꾼다.

```yaml
id: manager-develop
role: implementation
effort: medium
capabilities: [file_read, file_write, shell, subagent]
restrictions: [no_git_push, no_spec_body_edit]
host:
  claude: {}
  codex:
    sandbox_mode: workspace-write
```

출력:

- Claude: `.claude/agents/moai/manager-develop.md`
- Codex: `.codex/agents/moai/manager-develop.toml`

`tools`, `model`, `effort`, sandbox, MCP, Hook, memory와 같이 표현력이 다른 필드는 `emit`, `consequence`, `documented-drop`, `unsupported` 중 하나의 disposition을 반드시 갖는다.

변경 파일:

- 신규: `internal/template/canonical/agents/*.yaml`
- 수정: `internal/template/agentemit/{loader,manifest,emit,writer}.go`
- 대체: `internal/template/agentemit/agents-codex.yaml`
- 수정: `internal/template/agentemit/*_test.go`, `golden_test.go`

## 7. 세션 통신 계획

### Transport 계층

```text
SessionTransport
├── ClaudeNativeTransport     ListAgents / SendMessage
├── CodexAppServerTransport   agents / queue / App Server
└── MoAIBrokerTransport       session_msg_register/list/send/poll
```

```go
type SessionTransport interface {
    Probe(context.Context) CapabilityEvidence
    List(context.Context, SessionFilter) ([]Session, error)
    Notify(context.Context, SessionRef, Message) (Receipt, error)
    Readback(context.Context, Receipt) (DeliveryState, error)
}
```

### 지시와 메시지 분리

현재 broker의 “facts, not mutations” 계약을 유지한다. 실제 작업 지시는 별도의 `DispatchStore`에서 처리한다.

```yaml
dispatch:
  id: dsp_...
  task_id: T-...
  scope: [internal/template/canonical]
  authority_ref: user-turn-or-approved-card
  source_session_id: ...
  target_session_id: ...
  worktree_ref: ...
  expected_readback: [artifact, tests]
  expires_at: ...
```

수신 메시지는 dispatch ID를 알릴 뿐이다. 수신 세션은 dispatch를 claim하기 전에 권한, scope, worktree owner를 검사한다. 전달 receipt와 작업 완료 evidence는 별도 상태다.

변경 파일:

- 신규: `internal/sessiontransport/{types,probe,router}.go`
- 신규: `internal/sessiontransport/{claude,codex,broker}.go`
- 신규: `internal/dispatch/{store,types,validate}.go`
- 수정: `internal/sessionmsg/*`는 broker adapter가 사용할 안정 계약만 노출
- 신규 CLI: `moai session list|send|readback`, `moai dispatch create|claim|complete`

## 8. Worktree 통합 계획

### 두 가지 모드

| 모드 | owner | 용도 | 실행 |
|---|---|---|---|
| `persistent-card` | MoAI | SPEC/카드 구현, 통합 전까지 보존 | `moai worktree enter`, host는 지정 path에서 실행 |
| `ephemeral-session` | Claude 또는 Codex | 탐색·일회성 작업 | `claude --worktree`, `codex --worktree` |

```go
type WorktreeProvider interface {
    Create(context.Context, WorktreeSpec) (WorktreeRef, error)
    Launch(context.Context, WorktreeRef, Harness) (SessionRef, error)
    Inspect(context.Context, WorktreeRef) (WorktreeState, error)
    Close(context.Context, WorktreeRef, IntegrationProof) error
}
```

`WorktreeRef`에는 owner, path, base SHA, branch/detached, card ID, session IDs, cleanup policy를 저장한다.

### 호스트별 실행

- Claude persistent: MoAI가 만든 path에서 Claude를 시작하거나 `EnterWorktree`로 진입한다.
- Codex persistent: `codex -C <path>`로 시작하거나 같은 path에 묶인 thread를 resume한다.
- Claude ephemeral: `claude --worktree [name]`.
- Codex ephemeral: `codex --worktree`; detached HEAD 가능성을 명시한다.

거부 조건:

- 하나의 path에 owner가 둘
- persistent worktree 안에서 host-native worktree를 다시 생성
- 통합 증거 없이 persistent worktree 삭제
- Codex detached HEAD를 카드 branch로 등록
- worktree path와 dispatch scope 불일치

변경 파일:

- 신규: `internal/worktree/provider.go`, `ownership.go`
- 수정: `internal/cli/worktree/**`, `internal/cli/cc.go`
- 신규: Codex launcher 또는 공통 `internal/cli/harness_launch.go`
- 수정: worktree registry·guard 테스트

## 9. CLI와 설정 계획

### CLI

```text
moai init --harness claude|codex|both
moai update --harness auto|claude|codex|both
moai doctor --harness both
moai session list|send|readback
moai worktree enter <name|path> --harness claude|codex
```

호환:

- `--llm claude` → `--harness claude`
- `--llm gpt` → `--harness codex` + deprecated 경고
- `--llm both` → `--harness both`
- 모델 선택 설정은 `model.policy`로 분리한다.

### 설정

```yaml
harness:
  targets: [claude, codex]
  preferred: auto
  capability_probe: on-version-change

session:
  transport_order:
    claude: [native, moai_broker]
    codex: [app_server, moai_broker]
  delivery_readback: required
  message_is_authority: false

worktree:
  default_owner: moai
  persistent_cards: true
  native_ephemeral: true
  forbid_nested: true
```

## 10. Milestone 실행 순서

### M0 — 기준 동결과 coverage inventory · High

작업:

- 83개 rule에서 의무 조항을 추출하고 안정 ID를 부여한다.
- 현재 `AGENTS.md`, `CLAUDE.md`, Skill, Hook, agent에 대한 carrier map을 만든다.
- 설치된 Claude/Codex 버전과 capability evidence를 기록한다.

산출물:

- `canonical/policy/coverage.yaml`
- `canonical/capabilities.yaml`
- coverage validator와 누락 mutant 테스트

종료 게이트:

- 원본 MUST 조항 100%가 coverage에 존재
- 미측정 capability는 `unknown`, 미지원은 `unsupported`로 구분

### M1 — Policy canonicalization · High

작업:

- policy schema와 emitter를 구현한다.
- root `AGENTS.md`, `CLAUDE.md`, Claude rules를 생성한다.
- Codex scoped-policy 변환 판정을 구현한다.

종료 게이트:

- 기존 핵심 계약 semantic snapshot 동등
- hard clause skill-only 0개
- AGENTS 결합 크기 경고와 tail-truncation mutant 통과

### M2 — Skill·Agent dual emitter · High

작업:

- Skills와 Agents를 canonical source로 이동한다.
- 양쪽 출력과 documented-drop manifest를 생성한다.
- 현재 runtime mirror와 backward compatibility를 유지한다.

종료 게이트:

- Skill 공통 digest 100% 일치
- 11개 Agent 출력 등록 테스트 통과
- 사용자 소유 경로 overwrite negative test 통과

### M3 — Session transport와 Dispatch · High

작업:

- 세 transport adapter와 capability router를 구현한다.
- 권한 있는 dispatch store와 receipt/readback을 구현한다.

종료 게이트:

- Claude→Codex, Codex→Claude 왕복 E2E
- offline/expired/duplicate/at-least-once delivery 테스트
- 메시지만으로 승인·완료가 바뀌지 않는 테스트

### M4 — Worktree provider · High

작업:

- owner metadata와 공통 provider를 구현한다.
- Claude/Codex launch 전략을 연결한다.
- persistent와 ephemeral을 분리한다.

종료 게이트:

- 이중 owner와 nested worktree 거부
- detached HEAD/branch 구분
- 통합 증거 없는 삭제 거부
- 재진입 후 동일 card/session metadata 확인

### M5 — CLI, update, doctor · Medium

작업:

- `--harness` 도입과 `--llm` 호환 layer를 추가한다.
- init/update 조합별 golden을 만든다.
- doctor에 policy/skill/agent/hook/session/worktree 검사를 추가한다.

종료 게이트:

- `claude`, `codex`, `both` golden 통과
- 기존 프로젝트 fixture update 통과
- capability 측정 버전 drift 경고 확인

### M6 — 전환과 중복 제거 · Medium

작업:

- 한 호환 주기 동안 구·신 emitter 결과를 비교한다.
- coverage와 E2E가 통과한 항목만 오래된 원본에서 제거한다.
- manifest 기반 rollback을 검증한다.

종료 게이트:

- clean regeneration 후 git diff 0
- 구 버전 프로젝트 rollback/readback 성공
- `.claude/rules`에는 Claude scoped/host 규칙만 남음

## 11. 테스트 매트릭스

| 계층 | 테스트 | 핵심 실패 mutant |
|---|---|---|
| Policy | schema, coverage, snapshot | hard clause를 Skill-only로 이동 |
| Skill | dual digest, symlink/copy, user-owned path | Codex 경로에서 SKILL 누락 |
| Agent | Claude/Codex golden, registration | unsupported field 조용히 삭제 |
| Hook | 이벤트 parity, path-aware enforcement | Codex Hook 한 이벤트 누락 |
| Session | 왕복, ack, duplicate, offline, expiry | 메시지를 사용자 승인으로 처리 |
| Dispatch | authority/scope/worktree/readback | 다른 worktree가 claim |
| Worktree | owner, detached, nested, cleanup | 이중 owner 허용 |
| CLI | init/update/doctor 조합 | `--llm gpt` 프로젝트 파손 |
| E2E | Claude-only, Codex-only, both | 한쪽 산출물만 존재 |

권장 검증 명령:

```bash
go test ./internal/template/... ./internal/codexwiring/... -count=1
go test ./internal/sessionmsg/... ./internal/sessiontransport/... ./internal/dispatch/... -count=1
go test ./internal/worktree/... ./internal/cli/... -count=1
go test ./internal/hook/... -count=1
moai doctor --harness both --json
```

전체 suite는 push 후 CI가 실행하고, 로컬은 변경 영향 범위로 제한한다.

## 12. 위험과 완화

| 위험 | 영향 | 완화 |
|---|---|---|
| rule 분류 중 의무 손실 | 안전 계약 누락 | clause ID, coverage 100%, mutant test |
| Codex scoped rule을 Claude와 동일하다고 오판 | 특정 경로에서 규칙 미적용 | CWD 계약, root 승격, Hook, unsupported disposition |
| dual Skill drift | 하네스별 동작 차이 | digest manifest, clean regeneration CI |
| App Server/CLI 버전 변화 | session transport 파손 | versioned capability probe와 broker fallback |
| worktree 이중 소유 | branch 유실·오정리 | owner 필수, nested 거부, integration proof |
| symlink 제한 | Windows에서 Skill 누락 | copy fallback, drift warning, 플랫폼 E2E |
| AGENTS 결합 크기 초과 | tail 지침 유실 | size budget, 중요도 순서, truncation mutant |
| 기존 프로젝트 업데이트 | 사용자 설정 덮어쓰기 | managed-entry ownership, backup, explicit diff, rollback |

## 13. Rollout·Rollback

Rollout 순서:

1. `both`를 opt-in으로 배포하고 기존 기본값을 유지한다.
2. doctor에서 경고만 내는 관찰 단계로 시작한다.
3. golden·E2E가 안정되면 신규 프로젝트의 표준 옵션을 `--harness`로 전환한다.
4. 기존 `--llm`은 deprecated 경고와 함께 유지한다.
5. 중복 원본 제거는 가장 마지막에 한다.

Rollback:

- 이전 template manifest와 canonical source digest를 보존한다.
- update 전 `.moai` managed artifact manifest를 백업한다.
- 사용자 소유 파일은 rollback 대상에 포함하지 않는다.
- session broker와 기존 `--llm` 경로는 한 호환 주기 동안 유지한다.

## 14. Evidence와 경계

### Claim

현재 저장소에는 이중 하네스 전환에 필요한 주요 기반이 이미 있으며, 가장 큰 작업은 새 기능 개발보다 policy/Skill/Agent 원본의 중립화와 의미 동등성 검증이다.

### Evidence

```console
$ git fetch origin main 2>&1
From https://github.com/modu-ai/moai-adk
 * branch                main       -> FETCH_HEAD
$ git rev-list --count --left-right origin/main...HEAD
0	0
$ git rev-parse --short HEAD
2213871af
$ git branch --show-current
main
```

확인된 현재 파일:

- `internal/template/skill_mirror.go`
- `internal/template/agentemit/**`
- `internal/template/embed.go`
- `internal/template/templates/AGENTS.md`
- `internal/template/templates/CLAUDE.md`
- `internal/sessionmsg/**`
- `internal/codexwiring/**`
- `internal/cli/worktree/**`

### Baseline-attribution

- 저장소: `/Users/goos/MoAI/moai-adk-go`
- branch/HEAD: `main` / `2213871af`
- 원격 대비: `origin/main...HEAD = 0 0`
- 계획 기준일: 2026-09-21
- 공식 기능 근거는 동반 아키텍처 보고서의 OpenAI·Anthropic 공식 출처를 사용했다.

### Gaps

- 이번 산출물은 계획이며 코드 변경은 하지 않았다.
- 83개 rule의 clause-level inventory는 M0의 실제 구현 작업이다.
- Claude↔Codex live E2E, Codex remote App Server, Windows symlink fallback은 실행하지 않았다.
- Codex scoped instructions의 실제 적용 범위는 구현 전에 버전별 probe fixture로 다시 고정해야 한다.

### Residual-risk

- 하네스 버전이 바뀌면 capability 의미가 달라질 수 있다.
- 사용자 범위 AGENTS가 합쳐지므로 프로젝트 instruction 크기만으로 truncation 안전을 완전히 증명할 수 없다.
- 기존 dirty 프로젝트에서 update 충돌이 발생할 수 있으므로 ownership manifest와 non-overwrite 계약이 필수다.

## 15. 공식 출처

- OpenAI: Git worktrees, Subagents, App Server, Customization, AGENTS.md, Build skills
- Anthropic: Cross-session messaging, Agent teams, Git worktrees, Features overview, Memory
- 상세 URL과 기능 판정: `reports/moai-dual-harness-design-20260921.{html,md}`

## 최종 권고

첫 구현 단위는 세션 통신이나 worktree가 아니라 **M0 policy coverage inventory**여야 한다. 규칙의 의미를 먼저 고정하지 않고 디렉터리만 옮기면 가장 중요한 안전 조항이 조용히 사라질 수 있다. 그다음 canonical source와 emitter를 만들고, session/worktree adapter를 붙인 뒤, 마지막에 중복 `.claude` 원본을 제거한다.
