# MoAI-ADK 이중 하네스 설계 보고서

> Claude Code + Codex · 공식 문서 및 로컬 실측 기준 · 2026-09-21

## 0. 판정 요약

| 질문 | 판정 | 핵심 근거 |
|---|---|---|
| Codex도 세션 간 통신·지시가 가능한가? | **가능하나 Claude Code와 의미가 완전히 같지는 않다.** | 설치된 `codex-cli 0.155.1`은 공유 App Server의 세션을 찾는 `codex agents`와 기존 세션에 메시지를 넣는 `codex queue --thread … --message …`를 제공한다. 공식 문서는 App Server의 thread/turn 수명주기와 subagent 협업을 설명한다. 반면 Claude Code의 `ListAgents`/`SendMessage`처럼 에이전트가 다른 독립 세션을 직접 탐색하고 메시지를 주고받는 동일 계약은 Codex 공식 문서에서 확인되지 않았다. |
| Codex도 `EnterWorktree` 같은 기능이 있는가? | **worktree는 지원한다. 정확히 같은 in-session 도구는 확인되지 않았다.** | Codex는 `--worktree`, Desktop의 managed/permanent worktree와 Local↔Worktree handoff를 지원한다. Claude Code는 `claude --worktree`와 세션 중 `EnterWorktree` 도구를 공식 제공한다. Codex에서는 새/재개 세션을 해당 경로에 결합하는 방식으로 추상화해야 한다. |
| `.claude/rules`를 없애고 모두 skills로 옮기는 게 나은가? | **아니다. 전면 삭제는 권장하지 않는다.** | 두 제품 모두 항상 적용할 규칙과 필요할 때 불러오는 skill을 보완 관계로 설명한다. 안전·권한·Git 불변식을 skill로만 옮기면 호출되지 않은 턴에서 사라진다. 대신 중립 원본에서 `AGENTS.md`, Claude path rule, 양쪽 skill, hook을 생성하는 편이 맞다. |
| 최종 방향 | **공통 코어 + 얇은 호스트 어댑터 + 기능 협상** | 정책·skill·agent 정의는 하네스 중립 원본 하나로 관리하고, Claude/Codex 산출물은 결정적으로 생성한다. 세션 통신과 worktree는 런타임 어댑터로 분리한다. |

### 핵심 결론

MoAI-ADK는 이미 `--llm claude|gpt|both`, `.agents/skills` 미러, `.codex/hooks.json`, Codex agent emitter, `session_msg_*` MCP 브로커를 갖고 있다. 필요한 것은 새 프레임워크가 아니라 다음 네 가지 정리다.

1. **`.claude`를 원본으로 삼는 구조를 끝낸다.** `internal/template/catalog/`을 유일한 중립 원본으로 둔다.
2. **규칙을 역할별로 분류한다.** 항상 적용, 경로 적용, 작업 절차, 호스트 전용, 기계적 강제로 나눈다.
3. **세션 통신과 worktree에 호스트 어댑터를 둔다.** 네이티브 기능이 있으면 쓰고, 공통 브로커·MoAI worktree로 폴백한다.
4. **동등성을 테스트한다.** 문서가 존재하는지만 보지 말고, 위험 규칙 누락·skill 해시·agent 의미 손실·실제 메시지 전달·worktree 격리를 검사한다.

## 1. 최신 공식 기능 비교

### 1.1 세션 통신

| 기능 | Claude Code | Codex | MoAI 설계 판단 |
|---|---|---|---|
| 독립 세션 목록 | `ListAgents`; 세션 이름·ID·상태·worktree를 조회 | 로컬 실측 `codex agents`; 공유 App Server daemon의 세션 탐색 | `SessionDirectory` 인터페이스로 감싼다. |
| 실행 중 세션에 메시지 | `SendMessage`; 즉시 전달 또는 큐잉 | 로컬 실측 `codex queue --thread … --message …`; 공식 App Server는 thread/turn 제어 제공 | `SessionMessenger` 어댑터를 둔다. CLI 세부 명령은 capability probe 뒤 사용한다. |
| 같은 세션 내부 병렬 작업 | subagent | subagent와 `/agent`; 결과가 주 스레드로 합류 | 기존 agent emitter를 유지한다. |
| 다른 기기/원격 | cross-session route | App Server `--remote` WebSocket/Unix 연결 | 원격 주소·인증은 호스트 설정에만 둔다. |
| 권한 승계 | 수신 메시지는 승인이나 설정 변경이 아님 | Codex `queue`의 권한 승계 계약은 공식 문서에서 확인하지 못함 | **메시지를 사용자 승인으로 취급하지 않는다.** |

Claude Code의 cross-session messaging은 2.1.224+에서 `ListAgents`와 `SendMessage`를 제공하며, 다른 세션에서 온 메시지가 권한 승인이나 설정 변경을 대신하지 않는다고 명시한다. Codex 공식 문서는 App Server의 thread 시작·재개·fork·목록과 turn 제어를 설명하고, Codex 클라이언트의 subagent 협업을 설명한다. 이 머신의 Codex 0.155.1에는 세션 목록과 큐 입력 명령이 실제로 존재한다.

따라서 답은 “Codex도 가능”이지만 경계를 붙여야 한다.

- **사람/오케스트레이터 → 기존 Codex 세션 지시:** `codex queue`로 가능하다고 실측했다.
- **한 Codex 세션 → 다른 독립 Codex 세션의 자율 P2P:** Claude의 `ListAgents`/`SendMessage`와 동일한 공식 고수준 계약은 확인하지 못했다.
- **한 Codex 세션 안의 subagent 지시:** 공식 지원한다.
- **Claude↔Codex 상호운용:** 제품 네이티브 기능에 기대지 말고 MoAI broker가 공통 하한선이 되어야 한다.

#### 지시는 메시지가 아니라 권한 있는 작업 레코드여야 한다

현재 MoAI의 `session_msg_send`는 “짧고 자체 완결적인 사실만 전송하고 상태 변경 지시는 보내지 말라”고 명시한다. 이 제한은 유지하는 게 맞다. 자유문자 메시지로 파일 수정 권한까지 전달하면 발신 세션의 사용자 권한을 수신 세션이 오인할 수 있기 때문이다.

권장 계약은 다음과 같다.

- `session_msg_*`: 발견, 알림, 질문, 상태 회신용.
- `dispatch` 레코드: 실제 작업 지시용. `task_id`, `scope`, `authority_ref`, `source_session_id`, `target_session_id`, `expected_readback`, `expires_at`을 가진다.
- 수신 세션은 메시지가 아니라 저장된 dispatch를 읽고 자신의 권한·worktree·카드 소유권을 다시 검사한다.
- 완료는 “보냈다”가 아니라 결과 artifact와 검증 evidence가 dispatch에 연결됐을 때만 성립한다.

### 1.2 worktree

| 항목 | Claude Code | Codex | 통합 규칙 |
|---|---|---|---|
| 새 격리 작업공간 | `claude --worktree [name]` | `codex --worktree` | 둘 다 지원 |
| 세션 도중 진입 | `EnterWorktree` 공식 도구 | 동일 이름·동일 계약의 공식 도구는 미확인 | Codex는 대상 경로에서 새/재개 세션으로 전환 |
| UI handoff | 세션/worktree 중심 흐름 | Desktop Local↔Worktree handoff | 호스트 전용 UX로 취급 |
| 저장 위치 | 프로젝트 `.claude/worktrees/` 등 Claude 관리 | `$CODEX_HOME/worktrees` 아래 Codex 관리 | 호스트 관리 worktree는 임시 용도만 |
| 장기 카드 worktree | MoAI launcher 규칙 사용 | MoAI launcher 규칙 사용 | **MoAI가 경로·branch·정리를 단독 소유** |

Codex 공식 문서는 worktree가 Git 저장소에서 독립적인 대화를 실행하고, managed worktree가 보통 한 chat에 연결되며 detached HEAD일 수 있다고 설명한다. Claude Code는 CLI worktree뿐 아니라 세션 중 `EnterWorktree`를 제공한다.

MoAI에서는 두 모드를 명확히 분리해야 한다.

```yaml
worktree:
  ownership: moai        # moai | host
  mode: persistent-card # persistent-card | ephemeral-session
  forbid_nested: true
```

- `ownership: moai`: `moai cc -w <name>`/공통 launcher가 생성·등록·정리한다. Claude나 Codex는 주어진 경로에서만 실행한다.
- `ownership: host`: Claude/Codex 네이티브 worktree를 사용한다. 임시 탐색이나 단일 세션 작업에 한정한다.
- 한 물리 worktree의 lifecycle owner는 반드시 하나다. MoAI와 호스트가 동시에 정리하도록 두지 않는다.
- Codex의 detached HEAD 가능성을 branch 기반 카드 흐름과 섞지 않는다. 카드 작업은 MoAI 소유 persistent worktree로 고정한다.

## 2. `.claude/rules`를 전부 skills로 옮기면 안 되는 이유

Claude 공식 기능 개요는 `CLAUDE.md`를 항상 적용되는 지침, `.claude/rules`를 모든 세션 또는 경로별 지침, skills를 작업별 반복 워크플로로 구분한다. OpenAI 공식 문서도 `AGENTS.md`는 항상 알아야 하는 작고 지속적인 지침, skills는 필요할 때 불러오는 전문 절차라고 설명한다.

즉, rules와 skills는 중복 저장소가 아니라 **적용 시점이 다른 전달 수단**이다.

| 내용 종류 | 올바른 carrier | 예시 |
|---|---|---|
| 항상 지켜야 하는 불변식 | root `AGENTS.md` + hook/test | 공유 checkout에서 branch 변경 금지, 증거 없는 완료 주장 금지 |
| 특정 경로에서만 적용 | Claude path-scoped rule + Codex nested `AGENTS.override.md` | 프런트엔드 접근성, DB migration 규칙 |
| 사용자가 특정 작업을 요청할 때 필요한 절차 | dual-published skill | TDD, SPEC 작성, HTML report 제작 |
| 제품 고유 기능 사용법 | host adapter 문서/skill | Claude `EnterWorktree`, Codex App Server |
| 반드시 막아야 하는 위험 동작 | hook/CLI guard + 테스트 | sweep staging, 금지된 branch 조작, 권한 없는 dispatch |
| 긴 참고자료 | skill `references/` | 체크리스트, 예시, 패턴 카탈로그 |

### 권장안: rules 삭제가 아니라 원본 중립화

현재 템플릿의 83개 rule 파일을 다음 다섯 종류로 분류한다.

1. `standing`: 모든 턴에 필요한 핵심 계약. 압축해 `AGENTS.md`에 생성한다.
2. `scoped`: 경로 조건이 있는 계약. Claude rule과 Codex nested AGENTS 파일로 생성한다.
3. `workflow`: 명시적 작업에서만 필요한 절차. `.claude/skills`와 `.agents/skills`로 생성한다.
4. `host`: Claude/Codex 고유 메커니즘. 각 호스트 출력에만 생성한다.
5. `enforced`: 프롬프트만으로 부족한 안전 규칙. hook·CLI guard·테스트로 구현하고 문서는 설명만 한다.

삭제 조건은 파일 수가 아니라 coverage manifest다. 모든 기존 의무 조항이 적어도 하나의 항상 적용 carrier 또는 기계적 guard에 연결되고, 위험 조항이 skill에만 남아 있지 않다는 검사가 통과한 뒤에만 중복 rule을 제거한다.

## 3. 목표 아키텍처

```text
internal/template/catalog/                 ← 유일한 사람이 편집하는 원본
├── policy/
│   ├── standing/                          ← 항상 적용
│   ├── scoped/                            ← path selector 포함
│   └── coverage.yaml                      ← 조항 → carrier/guard/test
├── skills/<skill>/                        ← 하네스 중립 skill 패키지
├── agents/<role>.yaml                     ← 중립 역할·capability
└── capabilities.yaml                      ← 호스트별 지원/폴백/측정 버전
                 │
                 ▼ deterministic emit
internal/template/templates/               ← 배포 산출물, 직접 편집 금지
├── AGENTS.md                              ← 공통 standing policy
├── CLAUDE.md                              ← AGENTS import + Claude 전용
├── .claude/rules/                         ← Claude scoped/host rules
├── .claude/skills/                        ← Claude skill publication
├── .agents/skills/                        ← Codex skill publication
├── .claude/agents/                        ← Claude agents
├── .codex/agents/                         ← Codex TOML agents
├── .claude/settings.json                  ← Claude hooks/MCP
├── .codex/config.toml
└── .codex/hooks.json

runtime
├── SessionTransport
│   ├── ClaudeNativeTransport              ← ListAgents / SendMessage
│   ├── CodexAppServerTransport            ← agents / queue / App Server
│   └── MoAIBrokerTransport                ← session_msg_* 공통 폴백
└── WorktreeProvider
    ├── MoAIPersistentProvider             ← 카드용, 기본
    ├── ClaudeEphemeralProvider
    └── CodexEphemeralProvider
```

### 3.1 기능 협상

하네스 이름으로 동작을 추정하지 말고 session start 때 기능을 측정한다.

```yaml
capabilities:
  session_directory:
    claude: { native: ListAgents, fallback: moai_broker }
    codex:  { probe: "codex agents --help", fallback: moai_broker }
  session_message:
    claude: { native: SendMessage, fallback: moai_broker }
    codex:  { probe: "codex queue --help", fallback: moai_broker }
  enter_worktree:
    claude: { native: EnterWorktree }
    codex:  { strategy: restart_or_resume_at_path }
```

probe 결과에는 `harness_version`, `checked_at`, `capability`, `status`, `evidence_digest`를 저장한다. 미지원과 미측정을 구분하고, 버전이 바뀌면 다시 측정한다. 현재 `agents-codex.yaml`의 측정 버전은 0.147.0인데 설치 버전은 0.155.1이므로 doctor가 재측정을 요구해야 한다.

### 3.2 세션 API

```go
type SessionTransport interface {
    Capabilities(ctx context.Context) CapabilitySet
    List(ctx context.Context, filter SessionFilter) ([]Session, error)
    Notify(ctx context.Context, target SessionRef, msg Message) (Receipt, error)
    Readback(ctx context.Context, receipt Receipt) (DeliveryState, error)
}

type DispatchStore interface {
    Create(ctx context.Context, d AuthorizedDispatch) (DispatchRef, error)
    Claim(ctx context.Context, ref DispatchRef, actor SessionRef) error
    Complete(ctx context.Context, ref DispatchRef, evidence []ArtifactRef) error
}
```

`Notify`와 `Dispatch`를 분리하는 것이 핵심이다. Claude `SendMessage`, Codex `queue`, MoAI broker 중 무엇을 쓰더라도 전달 성공은 작업 완료가 아니다. 상태 전이는 persisted dispatch와 evidence가 담당한다.

### 3.3 worktree API

```go
type WorktreeProvider interface {
    Create(ctx context.Context, spec WorktreeSpec) (WorktreeRef, error)
    Launch(ctx context.Context, ref WorktreeRef, host Harness) (SessionRef, error)
    Inspect(ctx context.Context, ref WorktreeRef) (WorktreeState, error)
    Close(ctx context.Context, ref WorktreeRef, proof IntegrationProof) error
}
```

`WorktreeRef`에는 `owner`, `path`, `branch_or_detached`, `base_sha`, `session_ids`, `card_id`, `cleanup_policy`를 담는다. `Close`는 통합·원격 반영 증거가 없으면 거부한다.

## 4. 설정 개선안

### 4.1 CLI 명칭

현재 `moai init --llm claude|gpt|both`는 실제 동작하지만 제품 이름이 `gpt`라서 Codex 하네스와 모델 계열을 혼동한다.

- 신규 표준: `--harness claude|codex|both`
- 호환: `--llm gpt`는 `--harness codex`의 deprecated alias로 유지
- 저장 설정: 모델 선택과 하네스 선택을 분리

```yaml
harness:
  targets: [claude, codex]
  preferred: auto
  capability_probe: on-version-change

model:
  policy: high

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

### 4.2 생성기 구조

현재 `skill_mirror.go`는 `.claude/skills`를 canonical로 삼고 `.agents/skills`를 상대 symlink 또는 copy로 만든다. 양쪽에서 읽히는 목적은 달성하지만 Claude 경로가 원본이라는 편향이 남는다.

권장 변경:

- `internal/template/catalog/skills`를 중립 원본으로 이동한다.
- build-time emitter가 `.claude/skills`와 `.agents/skills`를 생성한다.
- 배포 시 같은 파일시스템이면 두 호스트 경로를 중립 배포 디렉터리에 대한 상대 symlink로 만들 수 있다.
- symlink가 불가능하면 copy하되 manifest에 source digest와 mirror digest를 기록한다.
- update/doctor가 drift를 검출하고, 사용자 소유 경로는 덮어쓰지 않는다.

`agentemit`이 이미 하고 있는 “지원 필드만 내보내고 의미 손실을 documented drop으로 기록”하는 방식을 policy와 skill에도 확장하면 된다.

### 4.3 doctor와 품질 게이트

`moai doctor --harness both`에 다음 검사를 추가한다.

- `AGENTS.md` 결합 크기와 32KiB 한계 여유
- `coverage.yaml`의 모든 MUST 조항 carrier 존재
- 안전 조항이 skill-only가 아닌지
- Claude/Codex skill 목록과 content digest 동등성
- Claude agent와 Codex TOML의 역할·effort·MCP disposition 동등성
- hook 이벤트 매핑과 trust 상태
- `session_msg_*` MCP 등록·왕복·ack
- 설치 버전 기준 `agents`, `queue`, `--worktree`, Claude `EnterWorktree` capability probe
- worktree owner 단일성, branch/detached 상태, registry 일치
- 오래된 측정 버전 경고

## 5. 단계별 마이그레이션

### Phase A — 분류와 계약 고정 · High

- 템플릿 83개 rule의 각 의무 조항에 ID를 부여한다.
- `standing/scoped/workflow/host/enforced`로 분류한다.
- `coverage.yaml`을 만들고 현재 carrier와 목표 carrier를 기록한다.
- 위험 조항이 skill에만 배치되면 빌드를 실패시킨다.

### Phase B — 중립 카탈로그와 결정적 emitter · High

- `internal/template/catalog`을 도입한다.
- policy emitter, skill emitter를 추가한다.
- 생성된 `templates/`에 수동 수정 금지 헤더와 재생성 검사를 넣는다.
- 기존 `.claude/skills → .agents/skills` 미러는 한 릴리스 호환 경로로 유지한다.

### Phase C — 세션 transport · High

- 기존 `session_msg_*`를 `MoAIBrokerTransport`로 감싼다.
- Claude native와 Codex App Server/CLI transport를 추가한다.
- capability probe와 버전별 결과 캐시를 넣는다.
- 실제 작업 지시는 typed dispatch로 분리하고 권한 참조를 강제한다.

### Phase D — worktree provider · High

- MoAI persistent와 host ephemeral provider를 분리한다.
- lifecycle owner가 둘이면 생성/정리를 거부한다.
- Codex detached HEAD를 카드 branch로 오인하지 않는 테스트를 추가한다.
- re-entry는 공통 `moai worktree enter` UX로 제공하되 내부 동작은 호스트별로 다르게 한다.

### Phase E — CLI·템플릿 전환 · Medium

- `--harness claude|codex|both`를 표준화한다.
- `--llm gpt` 호환 경고와 migration을 제공한다.
- init/update/doctor의 golden test를 두 하네스 조합별로 만든다.

### Phase F — 중복 제거 · Medium

- coverage, golden, E2E가 모두 통과한 뒤에만 오래된 `.claude/rules` 중복 원본을 제거한다.
- 배포 결과의 `.claude/rules` 자체는 Claude path-scoped/host 규칙이 남는 한 유지한다.
- rollback은 이전 템플릿 manifest로 재생성 가능해야 한다.

## 6. 수용 기준

1. `moai init --harness both`가 한 번의 실행으로 양쪽 설정·agent·skill·hook을 생성한다.
2. 같은 skill ID의 Claude/Codex `SKILL.md` digest가 같고, 호스트 전용 오버레이만 명시적으로 다르다.
3. 모든 안전 MUST 조항이 `AGENTS.md`, scoped rule/AGENTS, hook/test 중 하나 이상에 있고 skill-only 항목은 0개다.
4. Claude→Codex, Codex→Claude 알림이 전달되고 receipt/readback이 기록된다.
5. 메시지 수신만으로 권한 승인 또는 완료 상태가 바뀌는 negative test가 통과한다.
6. 한 카드 worktree에 lifecycle owner가 둘이면 작업이 거부된다.
7. Codex `--worktree`의 detached HEAD와 MoAI 카드 branch가 구분된다.
8. 설치 하네스 버전이 capability manifest의 측정 버전보다 새로우면 doctor가 재측정 경고를 낸다.
9. root instruction 결합 크기가 Codex 32KiB 한도 아래이며 tail truncation mutant가 검출된다.
10. 기존 `--llm gpt` 프로젝트가 명시적 경고와 함께 동일 산출물로 업데이트된다.

## 7. 현재 저장소 기준 Evidence

### Claim

현재 저장소에는 이중 하네스의 핵심 기반이 이미 있다: dual init 플래그, Codex MCP/hook, skill mirror, agent emitter, cross-harness broker. 다만 원본이 `.claude` 중심이고 Codex capability 측정 버전이 설치 버전보다 오래됐다.

### Evidence

기준 동기화:

```console
$ git fetch origin main 2>&1
From https://github.com/modu-ai/moai-adk
 * branch                main       -> FETCH_HEAD
$ git rev-list --count --left-right origin/main...HEAD
0	0
```

설치 버전:

```console
$ codex --version
codex-cli 0.155.1
$ claude --version
2.1.278 (Claude Code)
$ /Users/goos/go/bin/moai version
[moai_cp/20260910_130400] [moai_cp/20260910_130400-1452-gf67d2193f] [built 2026-09-17T15:23:26Z]
```

Codex 세션·worktree 기능 실측:

```console
$ codex --help
agents  Browse all agent sessions on the shared local app-server daemon
queue   Queue a message for an existing session
--worktree  Run the session in a new managed Git worktree

$ codex queue --help
Usage: codex queue [OPTIONS] --thread <THREAD> --message <TEXT>
--thread <THREAD>  Session UUID or exact session name
```

현재 템플릿 규모:

```console
template_rules=83
template_rules_bytes=1265664
template_agents_bytes=14229
template_claude_bytes=19766
claude_skills_bytes=3960832
agents_skills_bytes=65536
```

관련 구현 테스트:

```console
$ GOCACHE=/tmp/moai-dual-sessionmsg-cache go test ./internal/sessionmsg -count=1
ok  	github.com/modu-ai/moai-adk/internal/sessionmsg	0.361s

$ GOCACHE=/tmp/moai-dual-sessionmsg-cache go test ./internal/sessionmsg ./internal/cli -run 'SessionMsg' -count=1
ok  	github.com/modu-ai/moai-adk/internal/sessionmsg	0.103s [no tests to run]
ok  	github.com/modu-ai/moai-adk/internal/cli	0.509s

$ GOCACHE=/tmp/moai-dual-template-cache go test ./internal/template -run 'TestSkillMirror|TestCodexAgents' -count=1
ok  	github.com/modu-ai/moai-adk/internal/template	2.694s

$ GOCACHE=/tmp/moai-dual-emit-cache go test ./internal/template/agentemit ./internal/codexwiring -count=1
ok  	github.com/modu-ai/moai-adk/internal/template/agentemit	0.447s
ok  	github.com/modu-ai/moai-adk/internal/codexwiring	0.445s
```

실제 파일에서 확인한 구조:

- `internal/template/skill_mirror.go`: `.claude/skills`를 canonical로 두고 `.agents/skills`에 상대 symlink, 실패 시 copy.
- `.codex/config.toml`: `moai mcp-server` 등록.
- `.codex/hooks.json`: SessionStart, UserPromptSubmit, Pre/PostToolUse, Stop, SubagentStart/Stop, SessionEnd를 Codex harness로 연결.
- `internal/template/agentemit/agents-codex.yaml`: Claude agent 의미를 Codex TOML로 변환하고 미지원 의미를 documented drop으로 기록. 측정 버전은 0.147.0.
- `.claude/rules/moai/workflow/cross-session-messaging.md`: Claude native 경로와 `session_msg_register/list/send/poll` Codex broker 경로를 구분.

### Baseline-attribution

- 저장소: `/Users/goos/MoAI/moai-adk-go`
- branch/HEAD: `main` / `2213871af`
- 원격 대비: `origin/main...HEAD = 0 0`
- 문서 확인일: 2026-09-21
- 실행 환경: macOS, Codex 0.155.1, Claude Code 2.1.278

### Gaps

- 실제 다른 세션에 `codex queue`를 보내지는 않았다. 외부 세션 상태를 바꾸는 동작이므로 이번 보고서 범위에서는 help 계약만 측정했다.
- Claude↔Codex 실제 왕복 E2E와 원격 App Server 경로는 실행하지 않았다.
- Codex 공식 developer command 색인에는 이 머신에서 실측된 `agents`/`queue`가 아직 명시적으로 보이지 않았다. 그러므로 해당 CLI는 버전 probe가 필요한 기능으로 분류했다.
- Codex에 Claude `EnterWorktree`와 동일한 in-session 도구가 없다는 것은 확인한 공식 문서·CLI 표면 기준이다. 숨은/실험 API의 부재까지 증명한 것은 아니다.
- 이 보고서는 설계 산출물이며 코드·템플릿 migration은 구현하지 않았다.

### Residual-risk

- Codex App Server와 CLI의 실험/안정 경계가 후속 버전에서 바뀔 수 있다.
- 83개 rule을 자동 분류하면 의미가 축약될 수 있으므로 조항 단위 coverage와 adversarial mutant가 필요하다.
- symlink/copy 폴백은 Windows 권한과 파일시스템에 따라 달라질 수 있다.
- 메시지 transport가 성공해도 수신 세션이 종료·차단·다른 worktree에 있으면 작업 결과는 보장되지 않는다.
- root instruction은 사용자 범위 `AGENTS.md`와 합쳐지므로 프로젝트 파일 크기만으로 truncation 안전을 증명할 수 없다.

## 8. 공식 출처

확인일은 모두 2026-09-21이다.

- OpenAI, [Git worktrees](https://learn.chatgpt.com/docs/environments/git-worktrees)
- OpenAI, [Subagents](https://learn.chatgpt.com/docs/agent-configuration/subagents)
- OpenAI, [Codex App Server](https://learn.chatgpt.com/docs/app-server)
- OpenAI, [Customization overview](https://learn.chatgpt.com/docs/customization/overview)
- OpenAI, [AGENTS.md](https://learn.chatgpt.com/docs/agent-configuration/agents-md)
- OpenAI, [Build skills](https://learn.chatgpt.com/docs/build-skills)
- OpenAI, [Developer commands](https://learn.chatgpt.com/docs/developer-commands)
- Anthropic, [Cross-session messaging](https://code.claude.com/docs/en/cross-session-messaging)
- Anthropic, [Agent teams](https://code.claude.com/docs/en/agent-teams)
- Anthropic, [Git worktrees](https://code.claude.com/docs/en/worktrees)
- Anthropic, [Features overview](https://code.claude.com/docs/en/features-overview)
- Anthropic, [Manage memory](https://code.claude.com/docs/en/memory)

---

**최종 권고:** `.claude/rules`를 skills로 통째로 옮기지 않는다. 대신 하네스 중립 카탈로그를 만들고, 항상 적용되는 정책은 `AGENTS.md`와 guard로, 경로 규칙은 각 하네스의 scoped carrier로, 반복 절차는 dual-published skill로, 제품 고유 기능은 얇은 adapter로 생성한다. 세션 통신은 native-first/broker-fallback, 카드 worktree는 MoAI 단독 소유가 기본이다.
