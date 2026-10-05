# MoAI-ADK for Codex: 플랫폼 경계를 반영한 재설계

**조사 기준:** 2026-09-23, 현재 공유 체크아웃 `main` / `2213871af`, 설치 `moai-adk v3.2.0-rc.12` (`g34dea4ff0` dirty 빌드), `codex-cli 0.156.1`  
**산출물 성격:** 코드와 공식 문서의 전수 목록·대표 경로 감사 및 개선 계획. 실제 Codex에서 전체 SPEC 흐름이 완료됐다는 보고가 아니다.  
**범위:** `-f` Factory, Claude/GLM/Windows/Codex 지원 경계, 세션 통신, Codex 초기화·스킬·에이전트·훅·MCP·goal·plan/run/sync, 현재 설치본과 소스의 차이.

템플릿 모드와 Claude 규칙의 중복·토큰 비용 등 전체 하네스 감사는 [동반 종합 보고서](harness-full-audit-20260923-01a0cddd.html)에 정리했다. 이 문서는 플랫폼 지원 판정과 Codex 실행 구조에 초점을 맞춘 후속 설계다.

## 결론

**가능한 목표:** macOS·Linux·Windows의 공식 Codex CLI 실행 환경에서 MoAI의 SPEC, 검증, Todo, goal, 문서화 흐름이 Claude Code판과 같은 **결과 계약**을 만족하도록 만드는 것. Codex가 공식 지원하는 스킬·프로젝트 지침·서브에이전트·훅·MCP·`codex exec`를 사용한다. 각 기능의 존재는 공식 문서와 로컬 CLI로 확인했지만 MoAI의 end-to-end 완주는 아직 확인하지 못했다.

**경계:** 기존 `moai cc/glm -f` Factory는 Claude Code 세션과 그 백엔드를 전제로 한다. Codex와 네이티브 Windows에는 Factory 지원을 선언하지 않는다. Codex에서는 한 세션의 네이티브 서브에이전트와 독립 worktree로 병렬 작업을 구성하고, 카드·검증 증거는 디스크에 남긴다. Claude Code의 `ListAgents`/`SendMessage`와 Codex의 세션 도구를 같은 통신망으로 취급하지 않는다.

## 1. 판정 기준과 현재 기준선

세 갈래로 조사했다. (1) Factory/플랫폼/세션 통신의 실제 코드 경로, (2) 현행 OpenAI Codex 공식 문서와 설치 CLI, (3) MoAI Codex 생성기·어댑터·테스트. `.claude/worktrees/**`와 과거 보고서는 현행 코드 근거에서 제외했다. 과거의 `moai-adk for Codex` 문서는 **목표와 이전 가설**로만 사용했다. 파일 존재, 유닛 테스트, 실제 Codex 호출 성공은 각각 다른 증거 수준이다.

### 이번 조사에서 관측한 명령과 출력

```text
git branch --show-current
main
git rev-parse --short HEAD
2213871af
git status --porcelain=v1 -uno | awk 'END{print "tracked_modified",NR}'
tracked_modified 15

codex --version
codex-cli 0.156.1
moai version
moai-adk v3.2.0-rc.12
[v3.2.0-rc.12] [moai_cp/20260910_130400-2723-g34dea4ff0] [built 2026-09-23T10:48:43Z]

moai init --help | rg -n -- '--llm|--agent'
39:    --llm                    Llm harness to deploy and wire: claude, gpt, or both (default: claude; gpt deploys AGENTS.md + Codex surfaces only — no .claude/ tree; both adds Codex wiring to the claude deployment)

python3 (현재 생성 자산 목록)
skills_dirs 50 symlinks 34 wrappers 16
AGENTS_bytes 14229
codex_agents 11

python3 (현재 .codex/hooks.json 이벤트 목록)
hook_events 8
names PostToolUse,PreToolUse,SessionEnd,SessionStart,Stop,SubagentStart,SubagentStop,UserPromptSubmit
```

설치 바이너리의 `--llm`(claude/gpt/both)과 현재 소스 `internal/cli/init.go:128-154`의 `--agent`(claude/codex/both)는 같은 계약이 아니다. 설치 바이너리의 빌드 커밋 `34dea4ff0`은 현재 HEAD `2213871af`보다 뒤의 계보에 있으므로, 이 차이만으로 설치본의 노후화나 현재 브랜치와의 일치를 판정할 수 없다. `dirty` 빌드의 포함 변경도 복원하지 못했다. 아래 계획은 **출시 바이너리·해당 소스·생성 템플릿을 같은 manifest로 묶는 작업**을 선행 조건으로 둔다.

## 2. 사용자 환경과 기능을 분리한 지원 표

표의 **진입점 확인**은 CLI/코드가 경로를 제공한다는 뜻이고, **운영 지원**은 실제 환경에서 메시지·완료·복구까지 입증됐다는 뜻이다.

| 환경 | 현재 `-f` Factory | Claude 네이티브 세션 통신 | Codex 사용 | MoAI의 공식 지원 판정과 목표 |
|---|---|---|---|---|
| macOS·Linux + Claude Code | `moai cc -f [N]` 진입점 확인 | 조건 충족 시 `ListAgents`/`SendMessage` 경로가 프로젝트 규칙에 명시 | 해당 없음 | Factory 지원 대상. 실제 수신·완료 판정은 설치 버전·공급자·환경 변수 조건에 따라 확인. |
| macOS·Linux + GLM | `moai glm -f [N]` 진입점 확인 | **Claude Code 클라이언트에 GLM 백엔드를 연결**하므로 같은 런타임 제약을 받음 | 해당 없음 | Factory 지원 대상이지만 GLM 모델 사용만으로 네이티브 통신을 보장하지 않음. |
| 네이티브 Windows + Claude/GLM | 코드에 명시적 OS 거부 없음; Windows PID 감시 구현 존재 | 프로젝트 규칙상 네이티브 Windows에는 없음 | 해당 없음 | **Factory 운영 미지원으로 표시.** 현재 CLI가 조용히 진입할 수 있으므로 정책에 맞춘 명시 오류가 필요. |
| Windows의 WSL2 + Claude/GLM | Linux 경로의 잠재 대상 | 프로젝트 규칙은 WSL2를 Linux로 분류 | 해당 없음 | 네이티브 Windows와 분리. Factory 실제 실행 시험 전에는 공식 지원으로 승격하지 않음. |
| macOS·Linux·Windows + Codex CLI | MoAI Codex Factory 진입점·백엔드 미구현 | Claude 채널 사용 불가 | OpenAI는 세 OS의 Codex CLI를 문서화 | **Codex Factory 미지원.** Codex 고유의 단일 세션 + 서브에이전트 경로로 동등한 결과를 목표로 함. |
| `moai cg` 혼합 백엔드 | `FACTORY_MODE_UNSUPPORTED_BACKEND`로 거부 | 해당 없음 | 해당 없음 | 기존 거부 유지. 지원 대상을 늘린다는 이유로 우회하지 않음. |

`internal/cli/cc.go:55-68`, `glm.go:75-88`은 두 Factory 진입점을, `factory.go:398-415`는 `cg` 거부를, `internal/kanban/record.go:21-25`는 Claude/GLM 백엔드만 열거한다. `internal/kanban/factory_alive_windows.go:1-52`는 Windows PID 감시를 구현한다. 따라서 “Windows는 코드가 `-f`를 거부한다”는 명제는 **현재 소스에 맞지 않는다**. 사용자 제품 정책인 “Windows Factory 미지원”을 코드와 도움말에 명시하는 것이 재설계의 일부다.

Claude 네이티브 통신의 범위는 `.claude/rules/moai/workflow/cross-session-messaging.md:17-27`에 macOS/Linux(+WSL2), 공급자·버전·feature-flag 제약으로 적혀 있다. 이 규칙은 프로젝트의 현재 계약이며, 이번 조사에서 모든 조합의 실제 소켓 전달을 실행하지 않았다. GLM 런처는 `internal/cli/glm.go:173-175,282-284`처럼 Claude Code 클라이언트를 사용한다. 

### “크로스 통신 불가”를 정확히 나누기

- **Claude Code 네이티브 채널:** `ListAgents`/`SendMessage`는 Codex 런타임의 공통 API가 아니다. 네이티브 Windows에서도 프로젝트 규칙상 사용할 수 없다.
- **MoAI의 별도 파일 브로커:** `session_msg_register/list/send/poll` 네 MCP 도구가 `internal/cli/mcp_server.go:439-481`, `mcp_session_msg.go:30-55`에 있다. 같은 프로젝트의 `.moai/state/session-msg/`를 공유하는 Claude·Codex 세션용으로 설계됐다. 이는 **수동 폴링**이며 Claude 네이티브 채널과 동일한 지연·도착 보장을 뜻하지 않는다. 이번 조사에서 실제 양쪽 호스트 세션 간 전달은 시험하지 않았다.
- **Codex 자체 세션 표면:** 설치 CLI의 `codex queue --help`는 기존 스레드에 메시지를 대기시키는 명령을 보였다. 실제 전달·Claude 상호 운용은 미검증이고, 공식 App Server 문서의 스레드 API도 Claude 채널과의 연동을 보장하지 않는다. [Codex App Server](https://learn.chatgpt.com/docs/app-server)

Factory/재설계의 완료 판정은 항상 **디스크 카드 상태와 검증 증거**로 한다. 메시지는 알림이다. 이는 `.claude/rules/moai/workflow/kanban-dispatch.md:67-69`의 기존 계약과 일치한다.

## 3. 공식 Codex 문서가 허용하는 설계 범위

| 공식 기능 | 확인한 계약 | MoAI 설계 사용처 | 남은 검증 |
|---|---|---|---|
| 운영체제 | Codex CLI는 macOS·Linux·Windows를 제공하고 Windows 네이티브/WSL2 경로를 문서화한다. WSL1은 현 문서상 대상이 아니다. [CLI](https://learn.chatgpt.com/docs/codex/cli), [Windows](https://learn.chatgpt.com/docs/windows/windows-sandbox), [WSL](https://learn.chatgpt.com/docs/windows/wsl) | Codex 기본 워크플로를 세 OS에서 설계; MoAI Factory 지원 여부와 분리 | 세 OS의 실제 MoAI 설치·검증 |
| 지침·스킬 | `AGENTS.md` 계층 병합, `.agents/skills`, `$skill-name` 명시 호출과 설명 기반 선택. [AGENTS.md](https://learn.chatgpt.com/docs/agent-configuration/agents-md), [스킬](https://learn.chatgpt.com/docs/build-skills) | 공통 안전 계약은 짧은 AGENTS, 실행 절차는 기능별 Codex 스킬 | 스킬 발견 이후 본문 실행과 도구 사용 |
| 서브에이전트 | `/agent`와 `.codex/agents/*.toml` 사용자 지정 역할. [서브에이전트](https://learn.chatgpt.com/docs/agent-configuration/subagents) | 한 Codex 세션 안의 독립 연구·구현·감사 작업 | 실제 위임·병합·승인 경계 |
| 훅 | `command`·`mcp_tool` 핸들러, 비관리 훅의 해시별 신뢰. `prompt`·`agent` 핸들러는 실행하지 않는다. [훅](https://learn.chatgpt.com/docs/hooks) | 정책 게이트와 goal 평가를 Codex 이벤트에 맞게 연결 | 프로젝트 trust, 이벤트 발화, 차단 출력 |
| MCP | STDIO·HTTP 서버 연결과 도구 승인 설정. [MCP](https://learn.chatgpt.com/docs/extend/mcp) | MoAI 상태/검증/브로커의 구조화 도구 | 서버 기동, 읽기·쓰기 승인, 오류 전달 |
| 비대화식 실행 | `codex exec --json` JSONL, 출력 스키마, 재개. [비대화식 모드](https://learn.chatgpt.com/docs/non-interactive-mode) | 단계별 자동 검증과 CI fixture | 같은 SPEC의 plan→run→sync 완료 |
| App Server | 스레드·턴·steer API가 문서화되고 WebSocket remote는 실험적이다. [App Server](https://learn.chatgpt.com/docs/app-server) | 장래 UX 연구 후보 | 운영 Factory나 Claude 교차 통신의 기초로 삼지 않음 |
| 모델 제공자 | 사용자 지정 `wire_api`는 Responses 형식만 허용한다. [설정 참조](https://learn.chatgpt.com/docs/config-file/config-reference) | Codex 기본 모델 경로 우선 | Claude/GLM 프로토콜을 그대로 연결한다는 주장은 금지 |

OpenAI의 [Claude 플러그인 전환 가이드](https://developers.openai.com/plugins/guides/submit-claude-plugin)는 Claude 명령·재사용 절차를 스킬로 옮기고 전용 호출·훅을 Codex에 맞게 바꾸도록 안내한다. **추론:** MoAI는 공유 비즈니스 로직을 유지하되, 스킬 본문과 호스트 호출 계약은 Codex용으로 생성하는 편이 맞다.

## 4. 현재 MoAI Codex 연동의 확인된 빈틈

| ID · 우선순위 | 확인 근거 | 영향과 판정 경계 | 개선 방향 |
|---|---|---|---|
| C01 · 높음 · 초기화 계약 분기 | 설치본은 `--llm`의 claude/gpt/both, 현재 소스 `internal/cli/init.go:128-154`는 `--agent`의 claude/codex/both | 같은 안내를 두 구현에 적용할 수 없다. 설치본이 Codex 기능이 전혀 없다는 주장은 아니다. | 릴리스별 manifest와 호환 별칭 정책을 정하고 같은 빌드의 도움말·init 산출물 검증 |
| C02 · 높음 · 스킬의 Claude 호출 | `internal/template/skill_mirror.go:3-20,174-185`는 본문 변환 없이 연결한다. `.agents/skills/moai/SKILL.md:8,371-387`에 `Agent`, `Skill`, `AskUserQuestion`, `TaskCreate`가 남음 | 파일 로딩과 실제 실행은 다르다. | Codex용 작은 dispatcher와 기능별 스킬 본문을 생성하고 도구 capability map으로 금지 호출 제거 |
| C03 · 높음 · 11개 에이전트 본문 | `internal/template/agentemit/writer.go:99-109`는 Claude 본문을 거의 그대로 TOML에 싣는다. `agents-codex.yaml:111-142`는 의미 손실을 기록한다. | TOML 게시 테스트는 Codex 위임 성공을 증명하지 않는다. | 워크플로는 스킬에, 전문 역할은 Codex agent TOML에 두고 역할별 실제 위임 테스트 |
| C04 · 높음 · goal 계속 실행 | Claude는 `.claude/settings.json`에 `stop-goal`을 등록한다. Codex 생성기 `internal/codexwiring/hooks.go:95-113`은 일반 Stop만 생성한다. | Codex에서 MoAI goal의 조건 재평가·다음 턴 계속 실행은 미입증. | Codex Stop에 goal 판정 합성, 중복 평가·루프 상한 처리, 실제 턴 연속성 테스트 |
| C05 · 높음 · 훅 적용 범위·신뢰 | 소스 `internal/codexadapter/events.go:38-64`는 11개 중 6개를 adapted로 둔다. 현재 `.codex/hooks.json`에는 8개 이벤트가 있다. 기준 측정은 Codex 0.147.0, 설치는 0.156.1. | 현재 파일 8개와 현재 소스의 6개를 같은 릴리스 증거로 합칠 수 없다. 파일 존재는 trust·발화 증거가 아니다. | 현 버전의 발화/입출력/차단을 재측정하고 생성물 갱신; 미지원 이벤트는 명시 GAP |
| C06 · 중간 · 진단이 정적 | `internal/cli/doctor_codex.go:34-119`는 파일·whitelist·sidecar·PATH/MCP 표면을 검사한다. | doctor 결과만으로 스킬 실행·훅 신뢰·MCP 호출을 PASS로 볼 수 없다. | 정적 점검과 opt-in 동적 스모크 결과를 다른 상태로 보고 |
| C07 · 중간 · 통신의 두 계층 | `.claude/rules/moai/workflow/cross-session-messaging.md:125-139`는 Codex에 MCP 파일 브로커를 지정한다. | 네이티브 Claude 통신, MoAI 파일 브로커, Codex 로컬 queue는 서로 다른 보장. | 각 채널의 등록·송신·수신·ack·재시도·권한을 따로 검사; 완료는 디스크 증거로만 판정 |
| C08 · 중간 · Windows Factory 허위 기대 | `-f` 진입점에는 Windows 거부가 없고 Windows PID 감시 코드가 있다. 운영 정책은 미지원이다. | Windows 사용자가 명령 진입을 성공으로 오해할 수 있다. | 운영 미지원 정책이 확정이면 `FACTORY_MODE_UNSUPPORTED_PLATFORM`으로 조기 거부하고 WSL2 문구 분리 |
| C09 · 낮음 · Factory 소켓 설명 | `docs-site/content/en/advanced/factory-mode.md:51`은 소켓이 실제로 열린다고 쓰지만 `internal/kanban/bootstrap.go:298-324`는 표시용 주소라고 적는다. | 운영자가 주소 표시를 연결 가능성으로 오인할 수 있다. | 실제 전송 계층과 표시용 경로를 문서에 구분 |

`session_msg_*` 네 MCP 도구는 구현돼 있고 관련 단위 테스트가 있다. 독립 Codex·Claude 세션에서의 종단 전달, 네이티브 Windows, 다른 머신 간 전달은 이번 조사에서 실행하지 않았다. `codex queue`는 설치 CLI 도움말에만 확인했고 공식 문서·실제 전달 검증이 없어 제품 아키텍처의 필수 경로로 채택하지 않았다.

## 5. Codex 전용 목표 구조

```text
사용자 → Codex CLI 스킬($moai-plan / $moai-run / $moai-sync / $moai-goal)
               ↓
       Codex 전용 dispatcher + 역할별 agent TOML
               ↓
       공통 Go 코어(SPEC, Todo, 검증, Git 안전 규칙, 증거)
               ↓
       Codex 훅/MCP 어댑터 → 디스크 상태·검증 로그
               ↓
       독립 감사가 결과 판정 → 다음 단계
```

### A. 명령과 컨텍스트

`AGENTS.md`에는 모든 세션에 적용할 안전 계약과 출처만 둔다. 현재 파일은 14,229바이트이며 이 조사에서 Codex 로더의 실제 주입은 재실행하지 않았다. 기능별 스킬은 짧은 진입점에 필요한 워크플로 파일을 지연 로드한다. Claude용 `Skill()`·`Agent()`·`AskUserQuestion` 문구는 Codex 진입점에서 제거한다. 호출 도구는 추상 이름이 아니라 실제 Codex API 또는 MoAI CLI/MCP의 확정된 명령으로 적는다.

### B. 병렬성과 Factory 대체

Codex의 기본 경로는 **한 운영 세션이 카드 한 개를 끝까지 소유**하는 plan→run→sync다. 서로 독립된 조사·구현·감사는 Codex의 네이티브 서브에이전트에 맡긴다. 동시에 파일을 쓰는 일은 카드별 worktree와 명시 소유 파일로 분리한다. 여러 카드의 배치 병렬화는 Factory라고 부르지 않고, 별도 opt-in 실험으로 두며 카드 선택·승인·완료 판단을 디스크 큐와 증거로 고정한다. Codex 세션 간 메시지나 App Server remote가 없어도 기본 흐름은 완료돼야 한다.

### C. 훅, goal, 통신

Codex 프로젝트 trust 및 훅 해시 신뢰 후에만 게이트가 살아난다는 상태를 명확히 표시한다. 훅은 등록·신뢰·실제 발화·정책 차단을 각각 검사한다. `stop-goal`은 Codex Stop 흐름에서 하나의 평가자로 합성하고, 반복 상한·중복 발화·중단 후 재개를 실증한다. `session_msg_*` 브로커는 같은 프로젝트를 공유하는 세션의 보조 알림으로 유지한다. 전달 실패 또는 폴링 부재가 카드 완료를 가로막지 않도록 한다.

### D. 플랫폼 경계

Codex 기본 기능은 macOS·Linux·Windows에서 같은 acceptance contract를 사용하되 셸·경로·프로세스 호출 구현은 OS별로 검사한다. 네이티브 Windows에서 Unix 셸·tmux·Claude 네이티브 메시지를 가정하지 않는다. WSL2는 별도 행으로 시험하고 네이티브 Windows 성공을 추론하지 않는다. Claude/GLM Factory는 현행 경로로 유지하되 사용자에게 공급자/버전/메시지 가용성 진단 결과를 보여준다.

## 6. 성능 목표와 측정 방법

“Claude Code처럼 성능”은 모델 성능이나 속도가 자동으로 같다는 뜻으로 쓰지 않는다. 같은 SPEC의 **결과 정확성, 작업 완주, 컨텍스트 소비, 불필요한 호출, 복구성**으로 비교한다.

| 지표 | 현재 관측·제약 | 목표 실험 |
|---|---|---|
| 작업 완주 | Codex plan→run→sync 종단 실행 없음 | 같은 작은 SPEC을 두 호스트에서 각 단계와 AC/증거까지 완주 |
| 컨텍스트 | 현재 스킬 디렉터리 50개(34 링크+16 래퍼), AGENTS 14,229바이트. 실제 토큰 수는 이번 조사에서 측정하지 않음 | `codex debug prompt-input`과 사용량 기록으로 시작 입력·캐시·스킬 본문 로드 전후 비교 |
| 병렬 효율 | Codex native subagent 공식 지원, MoAI 역할 실행은 미검증 | 독립 작업을 서브에이전트로 분리한 경우와 직렬 경우의 완료율·도구 호출량 비교 |
| 재작업 | 현재 스킬/agent의 Claude 호출이 Codex에서 의미 있게 실행되는지 미검증 | 도구 미존재·잘못된 경로·반복 읽기 횟수를 CI fixture에서 0으로 만들기 |
| 안전·복구 | 훅 신뢰와 goal 연속 실행 미측정 | 중단/재개, 권한 거부, 부분 실패, 디스크 증거 불일치에서 잘못된 PASS 0건 |

측정값 없이 “토큰 절약”이나 “Claude Code와 동급 속도”를 확정하지 않는다. 긴 스킬을 무조건 줄이지 않고 역할 선택 정확도와 성공률이 유지되는지 함께 본다.

## 7. 실행 순서와 완료 게이트

| 단계 | 우선순위 | 작업 묶음 | 기계적 완료 조건 |
|---|---|---|---|
| P0 | 높음 | 출시 기준 고정: 소스·바이너리·템플릿·CLI 도움말 manifest, `--llm`/`--agent` 호환 계약 | 동일 빌드로 세 OS fixture init, 생성 파일 목록·help 대조 PASS |
| P1 | 높음 | 지원 표 고정: Windows native Factory 조기 거부, Codex `-f` 거부, WSL2 조건 문서, 소켓 문구 수정 | 모든 OS/백엔드 조합의 허용·거부가 테스트 표와 일치 |
| P2 | 높음 | Codex 스킬·11 agent를 실제 도구 계약으로 변환, 역할별 권한 확인 | Claude 전용 필수 호출 0건, 스킬 발견·본문 실행·대표 agent 위임 PASS |
| P3 | 높음 | 훅·goal·MCP의 현재 Codex 버전 실사와 어댑터 정렬 | trust 후 발화/차단, goal 계속 실행, MCP 읽기·쓰기 승인 PASS |
| P4 | 중간 | Codex 단일 세션 plan→run→sync와 native subagent 분업 | 동일 fixture SPEC에서 세 단계·AC·독립 감사 완료, 잘못된 PASS 0건 |
| P5 | 중간 | macOS/Linux/Windows/WSL2 호환·성능 회귀 매트릭스 | 각 지원 OS의 CLI 실증과 컨텍스트·호출량 비교 보고; 미지원은 명시 오류 |
| P6 | 조건부 | 여러 카드 병렬 배치 실험 | 카드별 worktree·큐·증거·충돌 복구가 검증될 때만 별도 기능으로 승격 |

Factory는 macOS/Linux Claude/GLM 지원 경로로 유지한다. Codex 경로의 성공 판정에 Factory나 Claude 네이티브 통신을 필수 조건으로 두지 않는다. 이 구분을 코드, 도움말, 문서, 테스트, 릴리스 노트에 동일하게 반영한다.

## 8. 검증 근거, 공백, 잔여 위험

**관측된 좁은 테스트:** 아래 출력은 위 기준 트리에서 이번 조사 중 직접 실행한 값이다. 이는 소스의 일부 분기 검증이다.

```text
go test ./internal/cli -run 'Test(ParseFactory|Factory|RejectFactory|RunCGFactory|CGReject)' -count=1 -timeout 90s
ok github.com/modu-ai/moai-adk/internal/cli 0.536s

go test ./internal/codexadapter ./internal/codexwiring ./internal/template/agentemit ./internal/sessionmsg -count=1 -run 'Test(EventTable|Wire|Render|Emit|Register|Send|Poll|Golden|MapOutput|ValidateConfig)' -timeout 90s
ok github.com/modu-ai/moai-adk/internal/codexadapter 0.333s
ok github.com/modu-ai/moai-adk/internal/codexwiring 0.354s
ok github.com/modu-ai/moai-adk/internal/template/agentemit 0.217s
ok github.com/modu-ai/moai-adk/internal/sessionmsg 0.331s
```

**명시적 공백:** Factory macOS/Linux 실세션 실행, GLM 공급자별 네이티브 메시지, Windows 네이티브/WSL2 Factory 운영, Codex 실제 스킬·agent·hook·MCP·goal·SPEC 완주, Claude↔Codex 브로커 종단 전달, 설치본과 현재 HEAD의 동일성은 관측하지 않았다. Windows 교차 빌드 시도는 로컬 Darwin Go 도구체인/캐시 문제로 실패했으므로 Windows 비호환의 증거로 쓰지 않았다.

**잔여 위험:** Codex 버전 변화, 사용자 project trust와 훅 재승인, 공급자별 모델 프로토콜, 전역 AGENTS 병합 크기, 세션 간 메시지 손실, 여러 worktree의 동시 쓰기가 남는다. 따라서 파일 생성·정적 doctor·유닛 테스트만으로 제품 지원을 선언하지 않는다.

**최종 판정:** `moai-adk for Codex`는 공식 Codex 기능 안에서 Claude Code판에 가까운 **결과 품질과 작업 완주**를 목표로 재설계할 수 있다. 현재는 파일 게시와 일부 단위 경로까지 확인된 **부분 지원**이다. Factory `-f`와 Claude 네이티브 세션 통신은 Codex 지원 약속에서 제외하고, Codex 고유 실행 경로를 위 게이트로 검증해야 한다.
