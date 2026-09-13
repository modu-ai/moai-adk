# M1 — native Agent(fork)/subtask 가용성 probe (t653)

- 카드: t653, milestone M1 (plan.md §2 M1)
- 기준선: worktree `.claude/worktrees/t653`, branch `WT-gateway-as4-resume`, HEAD `3c34e90ee`
- 수행일: 2026-09-13
- 판정: **전제 실패 NOT-RUN** — 설치본(2.1.270)의 문서화된 호출 표면에서 native Agent(fork)/subtask(부모 이력에서 자식을 분기하는 별도 연결 subtask) 파라미터를 확인할 수 없다.

## 1. Claim

설치된 Claude Code 2.1.270의 CLI·Agent 도구 표면에는 AS-013 세 번째 Given이 요구하는 "native Agent(fork)/subtask"(부모 이력에서 자식을 분기하고 병렬·중첩·resume 가능한 별도 연결 subtask)에 해당하는 파라미터나 문서화된 기능이 존재하지 않는다. 따라서 native fork 양성 실증은 이 환경에서 수행 불가 → NOT-RUN으로 기록한다.

## 2. Evidence (probe 방법 + 관측 출력)

### 2.1 설치본 확인

```
$ which claude && claude --version
/Users/goos/.local/bin/claude
2.1.270 (Claude Code)

$ ls -la /Users/goos/.local/bin/claude
lrwxr-xr-x  1 goos  staff  48 Sep 13 17:00 /Users/goos/.local/bin/claude -> /Users/goos/.local/share/claude/versions/2.1.270
```

### 2.2 CLI 표면 probe — `claude --help`

`fork|subtask|sub-task|resume|continue` 검색 결과, fork 계열은 세션 런처 플래그뿐이다:

```
$ claude --help | grep -in "fork\|subtask\|sub-task"
  97:  --fork-session                        When resuming, create a new session ID
  99:                                        with --resume or --continue)
```

→ `--fork-session`은 `--resume`/`--continue`와 함께 쓰는 **세션 ID 재발급 플래그**(launcher 수준)다. SPEC이 구별하라고 명시한 그 launcher `--fork-session`이며, Agent 도구의 부모-이력 fork가 아니다. `subtask`/`sub-task` 문자열은 0건.

`agent|isolation|worktree` 검색 결과: `--agent`/`--agents`(커스텀 에이전트 정의), `claude agents`(백그라운드 에이전트 관리), `--forward-subagent-text`, `-w --worktree` — subagent fork 파라미터 없음.

### 2.3 `claude agents --help` probe

`claude agents`는 백그라운드 세션 관리("Manage background agents")다. dispatch 옵션(`--agent`, `--model`, `--cwd`, `--mcp-config` 등)만 있고 부모 이력 상속·분기 위치 지정 옵션은 없다.

### 2.4 설치본 바이너리 내부 Agent 도구 스키마 probe

바이너리(`/Users/goos/.local/share/claude/versions/2.1.270`, 207,500,480 bytes)에서 도구 스키마 키를 직접 검색:

```
$ LC_ALL=C grep -ac "subagent_type" <binary>
55

$ LC_ALL=C grep -ao '.\{40\}subagent_type.\{120\}' <binary> | grep -o 'Agent-tool parameters.*'
l used Agent-tool parameters (`prompt`/`subagent_type`).
  → 오류 메시지가 Agent 도구 파라미터를 `prompt`/`subagent_type`으로 열거

$ LC_ALL=C grep -ao 'isolation[^a-zA-Z]\{0,4\}(worktree[^)]\{0,60\}' 등 isolation probe
- `isolation: "worktree"` gives the agent its own git worktree (auto-cleaned if unchanged).
Filesystem isolation: `worktree` runs in a temporary git worktree.
  → isolation enum에서 확인되는 값은 "worktree"(파일시스템 격리)뿐 — 이력 fork 값 없음

$ LC_ALL=C grep -ao 'forkSession[a-zA-Z]*' <binary> | head
forkSessionToBackground / forkSession / forkSessionId
  → 모두 백그라운드 세션 내부 관리 식별자

$ LC_ALL=C grep -ao '.\{60\}"fork".\{80\}' <binary> | head
...e.kind==="fork"?e.root.observers:... (반복)
  → `"fork"` 출현은 전부 백그라운드 세션 kind 내부 분기 로직(e.kind==="fork")이며,
    Agent 도구 입력 스키마의 파라미터가 아님
```

Agent 도구의 실측 파라미터 표면 (바이너리 문자열 기준): `prompt`, `subagent_type`, `run_in_background`, `name`, `isolation`(`"worktree"`), `model`, `effort`. **`fork` 파라미터 없음.** 중첩(2.1.219+ depth-3)·백그라운드 실행은 존재하지만 이것은 "부모 이력에서 자식을 분기"하는 기능이 아니라 새로운 독립 context의 자식 생성이다 (AS-013 첫 번째 Given의 일반 자식 — 별도 실증 대상으로 유지).

## 3. Baseline-attribution

- 위 모든 명령은 2026-09-13 이 워크트리 세션에서, 설치된 바이너리 `versions/2.1.270` (SHA 시점: 2026-09-13 04:55 빌드본)에 대해 직접 실행해 관측한 출력이다.
- 코드 변경 없음 — M1은 probe + 기록 산출물이 전부다 (RED 조건 없음, plan.md §2 M1).

## 4. Gaps (명시적 미관측)

- 대화형 세션 안에서 Agent 도구를 실제 호출해 스키마를 동적으로 요청하는 probe는 수행하지 않았다(환경이 비대화형이고, 바이너리 내장 스키마 문자열이 도구 스키마의 정적 원천이다). 다만 오류 메시지가 파라미터 집합을 `prompt`/`subagent_type`으로 명시적으로 열거하므로 정적 관측의 신뢰도는 높다.
- 설치본이 아닌 상위/하위 버전(예: 미래 릴리스)에서의 가용성은 관측 범위 밖이다. 이 probe의 판정은 2.1.270에 한정된다.

## 5. Residual-risk

- 바이너리는 minify된 번들이라 문자열 검색이 스키마의 일부를 놓쳤을 가능성은 완전히 배제할 수 없다. 그러나 (a) help 텍스트, (b) 오류 메시지의 명시적 파라미터 열거, (c) isolation enum의 "worktree" 단일 실측이 서로 일치하므로, `fork` 파라미터가 숨어 있을 위험은 낮다.
- 미래 설치본에서 기능이 추가되면 이 NOT-RUN은 재판정 대상이다 — probe 방법(§2)을 그대로 재실행해 양성을 확인하면 AS-013 세 번째 Given 실증을 진행할 수 있다.

## 6. 판정에 따르는 계약 효과 (plan.md §2 M1 그대로)

- native fork 양성 의무(AS-013 세 번째 Given)는 **삭제하지 않는다** — 유지된다.
- AS4/전체 지원 완료는 이 증거가 확보될 때까지 **보류**된다.
- 일반 자식(non-fork Agent) 성공이나 launcher `--fork-session` 성공으로 **대체하지 않는다** — 둘은 별개 실증 대상으로 남는다 (M5 + t654).
- resume·모델·압축·명시 분기 등 가용성과 무관한 AS4 경로의 구현·검증은 plan 순서대로 계속 진행한다(본 카드 M2-M6).
