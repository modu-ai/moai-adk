---
id: SPEC-CODEX-WIRING-001
title: "Codex Dual Harness M4 — wiring generator: moai init --agent claude|codex|both, .codex/hooks.json + config.toml, trust guidance, doctor"
version: "0.3.1"
status: completed
created: 2026-08-23
updated: 2026-08-24
author: manager-spec
priority: P1
phase: "v3.1.4 target"
module: internal/cli
lifecycle: spec-anchored
era: V3R6
tier: M
tags: "codex, dual-harness, wiring, hooks-json, mcp_servers, approval-mode, trust, init-flag, statusline"
depends_on: [SPEC-CODEX-HOOK-ADAPTER-001]
related_specs: [SPEC-CODEX-DUAL-AGENTS-001, SPEC-CODEX-SKILLS-CANONICAL-001, SPEC-CODEX-SESSION-MSG-001]
---

# SPEC-CODEX-WIRING-001 — Codex Dual Harness M4 배선 생성기

## HISTORY

| 버전 | 날짜 | 변경 |
|---|---|---|
| 0.1.0 | 2026-08-23 | 최초 작성 (plan-phase, card t88 / M4). 구 M4 블로커("프로젝트 hooks.json 미발화") 전제를 운영자 실측 정정(§A.3)으로 폐기하고 재설계 |
| 0.2.0 | 2026-08-23 | plan-audit review-1(D1 차단 + D2-D6 경미) 적용. D1: §A.4의 "21도구 전부 명시 선언" 주장이 오측이었음을 정정 — 명시 15/21·미선언 6·유효 불일치 4·실효 승인 집합 10으로 교체, REQ-CW-011/AC-CW-010을 유효값 기준선으로 재정의, 4도구 annotation 수정을 plan M2에 명시 + PRESERVE 예외 축소. D2: /hooks 히트수 범위 한정 재기술. D3: §G 격리 주석. D5: 시험명 REQ→AC 이동 |
| 0.3.0 | 2026-08-24 | 운영자 지시 statusline 범위 추가(카드 t88 본문 개정 2026-08-24). §A.6 사실 기준 신설(공식 문서·소스 확정 — 정식 식별자 29종 + 별칭 7종), REQ-CW-013(status_line create-if-absent 배선)·REQ-CW-014(한계 문서화, SHOULD) 신설, REQ-CW-005 MoAI 관리 범위에 tui.status_line 확정 기재, AC 12→14(MUST 13 + SHOULD 1). Tier M REQ/AC 상한 16 이내 |
| 0.3.1 | 2026-08-24 | plan-audit review-3(PASS 0.86) 경미 결함 D1-D3 적용. D1: 미존재 영어 로케일 README 중복 열거 제거(영어 정본=README.md, README 총 4종). D2: rebase 후 낡은 사실 갱신 — §A.1을 작성 기준 사실+해소 이력으로 재서술, §G 격리 주석을 #1602 병합 후 상태(잔여 부재 #1606만)로 갱신, §G 의존 표에 충족 기록, plan §A·§C#4 동일 갱신. D3: 별칭 개수 토큰 수 기준 7종으로 정정(6은 매핑 대상 수) |

## §A. 측정 전제 (Verified baseline)

> 근거 문서: `.moai/reports/t88/codex-support-audit.md` (전수 조사, release/v3.1.3 트리 기준) ·
> `.moai/reports/t88/moai-desktop-compat-20260823.html` (운영자 검증 정정, 2026-08-23) ·
> `.moai/reports/t83/precondition-measurement{,-round3}.md` (codex-cli 0.147.0 실측, release/v3.1.3 수록).

### §A.1 작성 기준 트리(origin/main @ 76b2c4ece) 상태 — 듀얼 하네스 코어 부재 → 해소 이력

본 SPEC의 작성 기준 트리(WT-codex-wiring @ 76b2c4ece = 당시 origin/main tip)에는 Codex 듀얼
하네스 코어가 **하나도 없었다** — `internal/codexadapter` · `internal/template/templates/.codex/` ·
`AGENTS.md` 템플릿 · 스킬 미러 전부가 당시 미병합 release batch PR #1602(release/v3.1.3)에
실려 있었다(`git ls-tree origin/main internal/ | grep codex` → 0건; `origin/release/v3.1.3` → hit).
**해소 이력(2026-08-24)**: #1602가 merge commit 915c310de로 main에 병합됐고 본 브랜치는 그
위로 rebase했다 — 현재 트리에서 `internal/codexadapter` 존재·`go build` green·어댑터 테스트
초록을 실측했다(plan §C Pre-flight). 잔여 미병합은 #1606(t187)뿐(§G).

반면 배선 그 자체는 어느 브랜치에도 없다 — 감사 확정(§2.1/§2.4/§2.6): init 플래그 목록
(`internal/cli/init.go:72-125`)에 `--agent` 부재, `.codex/hooks.json` 생성 코드 경로 0개,
`mcp_servers.moai` 등록 코드 경로 0개, `internal/codexadapter`는 프로덕션 호출자 0개인 라이브러리.

### §A.2 소비할 어댑터 표면 (M3가 남긴 것, release/v3.1.3 수록)

| 표면 | 제공 | 본 SPEC의 소비 |
|---|---|---|
| `EventTable` / `Resolve` | 11 이벤트 표(6개 adapted: PreToolUse·PostToolUse·SessionStart·SessionEnd·Stop·UserPromptSubmit)와 이벤트명→dispatcher 인자 매핑 | 생성기가 어느 이벤트에 어떤 command를 emit할지의 **단일 데이터 원천** |
| `ValidateConfig` | 측정된 키 화이트리스트 검증(최상위 `{description, hooks}`, 진입 `{matcher, hooks}`, 훅 `{type, command, timeout}`) | 생성기의 **매 쓰기 전 선검증** — 잘못된 키는 Codex에서 파일 전체를 무음 사망시킴(t83 Finding D) |
| `MapOutput` | `continue:false`→`decision:block`(이유 기본문구 치환), `systemMessage`→UserPromptSubmit `additionalContext` 라우팅, 그 외 패스스루 | 런타임 `--harness codex` 모드의 출력 래핑 |
| `RecordDiscards` | 전달 불가 메시지의 관측 기록(내용 길이만, `.moai/logs/codex-adapter.jsonl`) | 런타임 모드가 호출 — M3 REQ-3 의무("무음 금지")의 활성화 |
| `ClassifyStderr` | 이벤트별 stderr 의미(차단 사유 vs 연속 프롬프트) | 런타임 모드의 stderr 패스스루 근거 |

M3 SPEC(§B Out of Scope)은 명시적으로 남겨둔 것이 이 배선이다: "This SPEC ships the constraint
validator (REQ-5) but does not build the thing that writes a Codex hooks file" — 본 SPEC이 그
첫 프로덕션 호출자가 된다.

### §A.3 구 블로커 폐기 — 운영자 실측 정정 (2026-08-23)

t83 실측(codex-cli 0.147.0 셸 CLI)은 "프로젝트 `<proj>/.codex/hooks.json` 미발화, 무음"을
M4의 선결 블로커로 남겼다(§F Blocker Candidate). **이 전제는 폐기됐다** — 데스크톱 호환
보고서 "핵심 갱신" 절(운영자 검증, 원문 인용):

> "구 M4 블로커 서술 폐기. 'Codex가 프로젝트 .codex/hooks.json을 발화하지 않는다'는 종전 전제는
> 낡았다: 현재 공식 문서는 프로젝트 훅을 지원하고 0.149.0-alpha에서 격리 프로젝트 훅이 실제 발화했다.
> 남은 차단은 Codex가 아니라 MoAI의 훅 생성·배선 부재(M4/t88)다."

측정 조합: Codex 앱 26.818.32112(내장 codex 0.149.0-alpha.4.1) + 셸 CLI 0.147.0 + moai
v3.1.3-rc.0. 운영 권장 사역(같은 절): "MCP 등록 시 `default_tools_approval_mode = "writes"`
(쓰기 6종 … 만 승인받기). 프로젝트 config·훅은 신뢰된 프로젝트에서만 로드되며 해시 변동 시 재승인 필요."

### §A.4 t187 정합성 — `writes` 승인 모드는 **capability 기반** (리드 지정 정합성 검증)

질문: `default_tools_approval_mode = "writes"`는 capability 분류(신규 쓰기 도구 자동 포함)인가,
아니면 도구명 열거인가? **답: capability 기반.** 공식 문서 원문 인용(2026-08-23 fetch, §D URL 목록):

> "The `writes` mode prompts for tools that aren't marked read-only."
> — Codex MCP 문서 (`learn.chatgpt.com/codex/extend/mcp`)

즉 `writes`는 도구별 **read-only 표식(MCP ReadOnlyHint annotation)** 에 기반하며, 도구명 열거
메커니즘은 별도의 `enabled_tools`/`disabled_tools` 목록이다. 서버 실태(2026-08-23 실측 — plan-audit
review-1의 기계 집계를 본 SPEC 작성자가 등록 블록별 awk 매핑으로 재확인):

- moai MCP 서버의 **명시 선언은 15/21** — true 11 + false 4. false 4종은
  codex_task·codex_job_cancel·glm_task·glm_job_cancel.
- **6종은 선언 부재**(goal_arm·verify_snapshot·audit_cache·codex_audit·glm_audit·audit_multi)
  — `mcp.NewTool`의 기본 시딩(false)에 맡겨져 있다. goal_arm·verify_snapshot은 catalog가 WRITE라
  유효값이 우연히 일치하지만, **audit_cache·codex_audit·glm_audit·audit_multi의 4종은 catalog
  분류가 READ인데 유효 hint가 false인 유효 불일치**다.
- 따라서 capability 기반 `writes` 하부의 base 트리 **실효 승인 집합은 6이 아니라 10**이다
  (catalog 쓰기 6 + 미선언 READ 4). 이 4종의 annotation 추가(`true`)는 plan M2가 수행하고
  유효값 기준선 가드(REQ-CW-011)가 닫는다.
- t187(PR #1606, 미병합)은 4도구를 추가하며 catalog 21→25, WriteCapable 6→9 — 신규 3쓰기 도구
  (session_msg_register·send·poll)에 annotation false, session_msg_list에 true를 **명시 선언**한다
  (`origin/WT-codex-session-msg:internal/cli/mcp_server.go` diff에서 확인).

**정합성 입장(본 SPEC이 채택하는 결론)**: 카드의 6도구 열거는 **문서(작성 시점 쓰기 집합의 나열)이지
구성이 아니다**. config.toml에는 **도구명 열거를 일절 넣지 않는다**. 다만 "capability 기반이므로
자동으로 정확하다"는 결론은 공짜가 아니다 — 그 정확성은 **유효 annotation 값이 catalog 분류와
일치한다**는 불변식 위에 서고, base 트리는 지금 그 불변식을 4도구에서 위반 중이다. M2의 4종
수정(READ 도구에 `WithReadOnlyHintAnnotation(true)` 1줄씩 — Claude 측에서는 advisory라 행동 변화
없음)이 불변식을 복원하고, 그 후 승인 집합은 catalog 진실과 일치한다: base 21도구 트리 6개,
t187 병합 후 25도구 트리 9개(병합 순서 독립). 불변식의 기계화가 REQ-CW-011의 **유효값 기준선**
가드이고, `moai-mcp-tools.md`의 25도구 기술과도 모순 없다 — 능력 분류는 annotation에 살고
열거형 config에는 살지 않는다.

### §A.5 신뢰(trust) 모델 — 공식 문서 확정 사실

공식 hooks 문서(`learn.chatgpt.com/docs/hooks`, 2026-08-23 fetch) 원문 인용:

- 레이어 병합: "If more than one hook source exists, Codex loads all matching hooks" —
  전역+프로젝트 레이어는 **병합되어 둘 다 실행**되므로, 생성기는 사용자 훅을 절대 덮어쓰지 않아야 한다.
- 신뢰: "Before a non-managed command hook can run, Codex requires you to review and trust the
  exact hook definition." 신뢰는 **훅 정의의 해시**에 기록되고, 변경된 훅은 "marked for review and
  skipped until trusted" — 즉 `moai update`가 hooks.json을 다시 쓰면 내용 해시가 바뀌어
  **조용히 실행 정지**한다. `/hooks` 명령으로 "review new or changed hooks, trust hooks" 가능.
- 프로젝트 레이어: "load only when the project `.codex/` layer is trusted".
- hooks.json 4개 위치: `~/.codex/hooks.json` · `~/.codex/config.toml`(inline) ·
  `<repo>/.codex/hooks.json` · `<repo>/.codex/config.toml`. 본 SPEC의 배포 대상은 **프로젝트 레이어**.
- handler 스키마: `type`(command만 실행)·`command`·`timeout`(초, 기본 600; **SessionEnd는 기본 1,
  상한 3**) + 공식 문서가 언급하는 부가 키들(statusMessage 등). t83 측정 화이트리스트
  `{type, command, timeout}`의 부분집합만 emit하면 양쪽 모두 만족한다.

### §A.6 statusline 사실 기준 (운영자 지시 2026-08-24, 공식 문서·소스 확정)

카드 t88 본문에 운영자 지시로 추가된 statusline 범위의 조사 확정물(2026-08-24 fetch):

- **명령형 statusline 미지원**: openai/codex#17827(OPEN) — Claude의 `status_line.sh` 같은
  command-backed statusline은 Codex에 없다. 내장 식별자 배열 구성만 가능하므로 MoAI 고유
  정보(goal·todo·SPEC 상태) 노출은 이 이슈 해소 전 불가하다(REQ-CW-014가 문서화).
- **공식 문서** (`learn.chatgpt.com/docs/developer-commands` `/statusline` 절): "The footer
  status line updates immediately and persists to `tui.status_line` in `config.toml`.
  Available status-line items include model, model+reasoning, context stats, rate limits,
  git branch, token counters, session id, current directory/project root, and Codex version."
- **Config 레퍼런스** (`learn.chatgpt.com/docs/config-file/config-reference.md`):
  `tui.status_line` 타입 `array<string> | null` — "Ordered list of TUI footer status-line
  item identifiers. `null` disables the status line."
- **샘플 설정** (`learn.chatgpt.com/docs/config-file/config-sample.md`): 미설정 시 Codex 기본값
  `["model-with-reasoning", "context-remaining", "current-dir"]`.
- **식별자 원천** (소스 `openai/codex` `codex-rs/tui/src/bottom_pane/status_line_setup.rs`,
  main @ 2026-08-24): `StatusLineItem` enum(`serialize_all = "kebab_case"`) — **정식 토큰 29종**:
  model, model-with-reasoning, reasoning, current-dir, project-name, hostname, git-branch,
  pull-request-number, branch-changes, run-state, permissions, approval-mode,
  context-remaining, context-used, five-hour-limit, weekly-limit, codex-version,
  context-window-size, used-tokens, total-input-tokens, total-output-tokens,
  thread-credits, estimated-thread-cost, thread-id, fast-mode, raw-output, thread-title,
  workspace-headline, task-progress. **파싱 전용 별칭 7종**(구값 호환 — 발행에 쓰지 않는다;
  별칭 토큰 수 기준이며 매핑 대상 정식 토큰은 6종):
  model-name→model, project→project-name, project-root→project-name, status→run-state,
  approval→approval-mode, context-usage→context-used, session-id→thread-id.
- **카드 제안 5종의 정식 토큰 확정**: `model-with-reasoning` · `context-remaining`(카드의
  "context" — Claude statusline CW% 게이지에 대응하는 컨텍스트 게이지, used% = 100 − remaining%) ·
  `git-branch` · `current-dir`(카드의 "cwd") · `thread-id`(카드의 "session-id"의 정식 토큰).
  Codex 자체 기본값 3종의 상위집합에 git-branch·thread-id를 더한 구성.
- **업스트림 증가 대응**: 허용목록은 본 저장소의 고정 상수로 유지한다(블라인드 업스트림 추적
  없음 — 검증 가능성 우선). 식별자 추가는 문서 갱신 시점의 명시적 판단이다.

## §B. User Story

> Codex(데스크톱 앱·CLI)에서 moai-adk를 쓰는 사용자로서, `moai init --agent codex` 한 번으로
> 훅 계층(브랜치 가드·위험 패턴·stop 체인)과 MCP 25도구가 내 프로젝트에 배선되기를 원한다.
> 수동 `codex mcp add`나 손수 쓴 hooks.json 없이. 그리고 moai가 훅을 다시 썼을 때
> "왜 갑자기 훅이 안 돌지?"가 아니라 "다시 승인하라는 안내"를 받고 싶다.

## §C. Scope Summary

**In Scope (요약)**

1. `moai init --agent claude|codex|both` 플래그 — 폐쇄집합 검증, 기본 `claude`(플래그 부재 =
   오늘과 동일한 동작, backward compat).
2. 신규 패키지 `internal/codexwiring` — `.codex/hooks.json` 생성기(EventTable 유도, 병합 보존,
   ValidateConfig 선검증, 멱등) + `.codex/config.toml` `[mcp_servers.moai]` 테이블 등록.
3. 런타임 seam: `moai hook <arg> --harness codex` — dispatcher 출력을 codexadapter로 래핑
   (MapOutput·RecordDiscards·이벤트 일관성 검증). codexadapter의 **첫 프로덕션 호출자**.
4. 신뢰 안내 — 생성 시점 stdout 안내 + 내용 변경 재생성 시 `/hooks` 재신뢰 안내 + 신뢰 사이드카.
5. `moai doctor` "Codex Wiring" 진단 — 존재·ValidateConfig·해시 divergence·PATH 해석·테이블 존재.
6. annotation 정합성 가드 테스트(WriteCapable ↔ ReadOnlyHint) — t187 정합성의 기계화.
7. `.codex/config.toml` `[tui]` `status_line` 기본 구성(운영자 지시 2026-08-24) — 정식 식별자
   5종 create-if-absent 배선 + 고정 허용목록 회귀 테스트. command-backed 미지원 한계(#17827)의
   문서화(SHOULD, REQ-CW-014).

### Out of Scope — 플러그인 패키징 (M6/t90)

- `.codex-plugin/plugin.json` 패키징, 스킬+MCP 번들 — M6/t90 소관, 수요 게이트 상태.
- M0 반증(`plugin_hooks` removed)으로 재범위 예정인 영역을 본 SPEC이 선행하지 않는다.

### Out of Scope — 에이전트 TOML 재생성 (M5/t89)

- `.codex/agents/moai/*.toml` 11종의 생성·수정·재발행 — agentemit(make agents-emit)과 템플릿
  순회 배포가 이미 소유. 본 SPEC은 이 파일들을 **건드리지 않는다**(REQ-CW-012). 정합성(존재 확인,
  산출물과의 무충돌)만 담당.

### Out of Scope — Claude 사이드 변경

- `.claude/` 템플릿·hook 래퍼·settings.json 구조 변경 — 전부 무변경. Claude 사이드에서 유일한
  변경은 `--agent` 값이 `.mcp.json` provisioning 호출을 게이팅하는 플래그 배관뿐이다(REQ-CW-001).
- 템플릿 신규 추가 없음 — hooks.json/config.toml은 **코드가 생성하는 사용자 프로젝트 파일**이다
  (.mcp.json provisioning과 동일한 분류; Template-First 대상 아님). 구현 중 template 변경이
  생기는 경우에만 run-phase에서 Template-First + `make build` 적용.

### Out of Scope — Codex 내부 신뢰 저장소 판독

- Codex가 신뢰 해시를 어디에 어떤 형태로 저장하는지는 비문서화면이다. doctor는 그 저장소를
  읽지 않고, **사이드카 대비 divergence**(마지막 생성 내용의 sha256 vs 현재 파일)로 간접 검증한다.

### Out of Scope — 미측정 이벤트·미측정 config 키

- EventTable의 미적응 5이벤트(compact·post-compact·permission-request·subagent-start·
  subagent-stop)는 emit하지 않는다(M3의 측정 범위 결정 존중).
- `features.hooks` 등 측정되지 않은 config.toml 키를 능동 주입하지 않는다(§H Risks 참조).

## §D. Requirements (GEARS)

> 표기는 GEARS. `주어진 선택자`는 capability gate(`Where`), `~할 때`는 이벤트(`When`),
> `~인 동안`은 상태(`While`)이며 수식 어순으로 자유 결합한다.

### REQ-CW-001 — init 플래그와 폐쇄집합

`moai init` 명령은 `--agent` 문자열 플래그를 제공하고, 그 값은 폐쇄집합
`{claude, codex, both}`로 제한하며, 기본값은 `claude`이다. `moai init`은 집합 외 값을
플래그 검증 단계에서 0이 아닌 exit code와 유효값을 나열하는 진단으로 fail-loud 거부한다.

**플래그 부재 하에서** `moai init`은 오늘의 동작과 동일하게 수행한다 — 동일한 템플릿 배포,
동일한 `.mcp.json` 기본 provisioning(`provisionMCPEntryUnlessDeclined` 경로), `.codex/hooks.json`
및 `.codex/config.toml` 미생성. `--agent claude`는 플래그 부재와 동일하게 동작한다.

**`--agent codex`가 주어진 경우** `moai init`은 Claude 측 `.mcp.json` provisioning을 수행하지
않는다(사용자가 자신의 하네스가 Codex라고 선언했음을 존중; 플래그가 wizard 답변에 우선).
**`--agent both`가 주어진 경우** 양쪽 provisioning을 모두 수행한다.

### REQ-CW-002 — hooks.json 생성 (어댑터 유도)

**`--agent codex` 또는 `--agent both`가 주어진 경우** 생성기는 `<project>/.codex/hooks.json`에
`internal/codexadapter`의 `EventTable` 중 `Adapted`인 행 전부(6 이벤트)를 배선한다 — 이벤트 키는
Codex 설정 키 문법(PascalCase)으로, 각 handler의 command는 `moai hook <dispatcher-arg> --harness codex`
형태로, handler의 `type`은 `"command"`로 emit한다. 생성기는 이벤트 집합을 자체 열거하지 않고
`EventTable`을 읽어 유도한다(신규 이벤트 적응 시 배선이 자동 따라온다).

생성기는 `timeout`을 이벤트별로 emit하되 **SessionEnd의 timeout은 3 이하**로 제한한다(공식 문서의
SessionEnd 상한). 그 외 이벤트의 기본 timeout은 테이블 상수로 정의한다.

### REQ-CW-003 — 화이트리스트 선검증 (무음 사망 방지)

**생성기가 hooks.json을 쓰기 직전마다** 렌더된 바이트를 `codexadapter.ValidateConfig`로 검증하고,
위반이 1건이라도 있으면 **그 파일을 쓰지 않고** 진단과 함께 중단한다. Codex는 잘못된 최상위 키
하나로 파일 전체를 조용히 무력화하므로(t83 Finding D), 검증 통과 바이트만 디스크에 도달한다.

### REQ-CW-004 — config.toml MCP 등록 (열거 없는 writes)

**`--agent codex` 또는 `--agent both`가 주어진 경우** 생성기는 `<project>/.codex/config.toml`에
`[mcp_servers.moai]` 테이블을 확보한다 — `command = "moai"`, `args = ["mcp-server"]`(Claude 측
`.mcp.json` provisioning이 쓰는 것과 동일한 상수·PATH 해석, 절대경로 없음),
`default_tools_approval_mode = "writes"`. 생성기는 이 목적을 위해 도구명 열거
(`enabled_tools`/`disabled_tools`)를 config에 남기지 않는다 — 승인 정책은 서버의 read-only
annotation이 담당한다(§A.4).

### REQ-CW-005 — 사용자 산출물 보존 (병합, 무분쇄)

**`.codex/hooks.json` 또는 `.codex/config.toml`이 이미 존재하는 경우** 생성기는 MoAI 관리
엔트리만 갱신하고 그 외 모든 내용을 보존한다 — hooks.json에서 MoAI 관리 엔트리는 command가
`moai hook `으로 시작하는 handler들이며(갱신 시 낡은 MoAI handler들을 제거한 뒤 현재 표를
append), 그 외 모든 엔트리·이벤트 키·description은 바이트 보존한다. config.toml에서 MoAI 관리
범위는 `[mcp_servers.moai]` 테이블뿐이다. **이미 존재하는 `[mcp_servers.moai]` 테이블은 덮어쓰지
않는다** — 사용자 소유 판정이며, 정본과의 불일치는 doctor가 보고한다(REQ-CW-010).
**v0.3.0 확장**: `[tui]` 테이블의 `status_line` 키가 MoAI 관리 create-if-absent 대상에
추가된다(REQ-CW-013) — 이미 존재하는 키는 사용자 소유로 바이트 불변이며, `[tui]`의 다른
키·다른 테이블 전부는 기존대로 무결정·보존이다.

### REQ-CW-006 — 멱등성

**입력이 변하지 않은 채 생성이 재실행되는 경우** 산출 파일들은 바이트 동일이다 — 고정 키 순서,
타임스탬프 없음, 절대경로 없음, 실행 시점 환경값 없음(agentemit의 결정론 규칙과 같은 기준).

### REQ-CW-007 — 런타임 `--harness codex` 모드 (어댑터 활성화)

`moai hook` dispatcher는 `<dispatcher-arg>` 부속명령에 `--harness codex` 모드를 제공하고,
**그 모드로 실행되는 경우** dispatcher 출력을 `codexadapter.MapOutput`으로 재작성하고
(`continue:false`→`decision:block`, UserPromptSubmit `systemMessage`→`additionalContext`),
전달 불가 메시지를 `RecordDiscards`로 `.moai/logs/codex-adapter.jsonl`에 기록하며, exit code와
stderr는 변경 없이 통과시킨다(M3 REQ-2/3/4 의무의 런타임 활성화).

**`--harness codex` 모드에서 payload를 읽는 경우** dispatcher는 payload의 `hook_event_name`이
`codexadapter.Resolve`로 해당 부속명령의 dispatcher 인자에 정확히 매핑되는지 검증하고,
불일치 시 0이 아닌 exit와 진단으로 거부한다(생성 시점 표와 런타임 표의 어긋남 방지).

`internal/hook/` 하위 기존 파일은 수정하지 않는다(M3 REQ-7 정신 — seam은 dispatcher 앞뒤에
살고 결정 로직 안에 들어가지 않는다).

### REQ-CW-008 — 신뢰 안내 (생성 시점 + 변경 시점)

**생성기가 hooks.json을 처음 만드는 경우** stdout에 Codex 신뢰 흐름 안내를 출력한다 — 프로젝트
`.codex/` 레이어 신뢰와 `/hooks` 검토·승인 절차를 이름한다. **재생성이 hooks.json 내용을 변경하는
경우** 안내에 변경된 훅이 재신뢰 전까지 실행 정지함을 명시하고 `/hooks to re-trust` 지시를
포함한다(§A.5 신뢰 모델 — 해시 변경 = 조용한 정지).

생성기는 마지막 생성 내용의 sha256을 신뢰 사이드카(`.moai/state/codex-wiring.json`)에 기록하여,
doctor의 divergence 판정과 재생성 전후 비교의 기준을 제공한다.

### REQ-CW-009 — update 갱신 규칙 (존재 = opt-in)

**`moai update`가 실행되는 경우** updater는 이미 존재하는 배선 파일만 갱신한다 — hooks.json이나
config.toml의 moai 테이블이 없는 프로젝트(`--agent claude`/플래그 부재 init)에는 배선을
**만들지 않는다**. 파일 존재가 사용자의 opt-in 지속 표식이다. 갱신 결과 hooks.json 내용이
변하는 경우 REQ-CW-008의 재신뢰 안내를 출력하고, 무변경인 경우 안내를 출력하지 않는다.

### REQ-CW-010 — doctor "Codex Wiring" 진단

`moai doctor`는 "Codex Wiring" 진단 항목을 제공하고, 배선 활성 프로젝트(파일 존재)에서 다음을
검증한다: hooks.json 존재·`ValidateConfig` 통과, 사이드카 해시 대비 현재 파일 divergence
(divergence 시 `/hooks to re-trust` 안내), `moai` 바이너리 PATH 해석, config.toml의
`[mcp_servers.moai]` 테이블 존재·정본 일치. 배선 비활성 프로젝트에서는 정보성 스킵 상태를
보고한다. 진단은 advisory(비게이팅)이며 실패 시에도 doctor 나머지를 중단하지 않는다
(`checkBinaryFreshness`의 t184 선례와 같은 형태).

### REQ-CW-011 — annotation 유효값 정합성 가드 (t187 정합성의 기계화)

**moai MCP 서버가 도구를 등록하는 경우** 각 도구의 **유효** read-only 값 — 선언된 annotation
값, 선언이 없으면 기본 false — 이 catalog(`internal/mcp`의 `WriteCapable`) 분류와 일치해야 한다
(유효 read-only ⟺ `WriteCapable=false` 도구별 동치). 이 동치를 주장하는 가드 테스트가 변경
세트에 포함된다 — `writes` 승인 모드의 승인 집합은 유효 annotation에 의존하므로(§A.4),
catalog-read 도구의 annotation 누락(유효 false)과 catalog-write 도구의 read-only 선언이 모두
가드에서 실패한다. 기준선은 **유효값**이지 선언의 존재가 아니어서, base 트리가 이미 지닌
선언-부재 상태(goal_arm·verify_snapshot 등)를 선행 수정 없이도 통과할 수 있고 catalog 진실과의
어긋남만을 잡는다.

### REQ-CW-012 — M5 산출물 무변경

배선 생성기와 doctor의 Codex 점검은 `.codex/agents/**` 을 생성·수정·삭제하지 않으며, 생성
산출물(hooks.json·config.toml)에 에이전트 정의를 중복 수록하지 않는다 — 에이전트 TOML은
M5(agentemit + 템플릿 순회 배포)의 단일 소유면이다.

### REQ-CW-013 — config.toml status_line 기본 구성

**`--agent codex` 또는 `--agent both`가 주어진 경우** 생성기는 `<project>/.codex/config.toml`의
`[tui]` 테이블에 `status_line` 키를 기본 구성 배열
`["model-with-reasoning", "context-remaining", "git-branch", "current-dir", "thread-id"]`로
create-if-absent 발행한다. **`status_line` 키가 이미 존재하는 경우** 생성기는 그 키를 바이트
불변으로 남긴다(사용자 소유 — REQ-CW-005 병합 모델의 v0.3.0 확장). `[tui]` 테이블이 없으면
파일 말미에 신규 테이블 섹션으로, 테이블은 있으나 키가 없으면 그 섹션 안에 키를 삽입한다.
발행 토큰은 정식(canonical) 토큰만 사용하며(§A.6 — 파싱 전용 별칭은 발행하지 않는다), 생성기는
발행·검증에 고정 허용목록 상수(`statusLineAllowlist`)를 사용하고 **기본 구성 ⊆ 허용목록**
동치를 회귀 테스트로 고정한다. Codex는 command-backed statusline을 지원하지 않는다
(openai/codex#17827 OPEN) — 내장 식별자 배열 구성이 이 REQ의 전부이며, 그 한계는 REQ-CW-014가
문서화한다.

### REQ-CW-014 — statusline 한계 문서화 (SHOULD)

배선 문서(README Codex 관련 절·docs-site)는 MoAI 고유 정보(goal·todo·SPEC 상태)의 statusline
노출이 openai/codex#17827 해소 전 불가함을 명시한다. 문서화는 sync-phase 산출물이며 본 REQ의
검증은 AC-CW-014(SHOULD)로 한다 — 미달 시 사유와 함께 부채 기록.

## §E. Acceptance Criteria

Given-When-Then 시나리오 14건(AC-CW-001..014 — MUST 13 + SHOULD 1)은 `acceptance.md` §A
매트릭스에 실행 가능한 명령과 함께 명세된다. grep 계열 AC의 토큰은 사전구현 트리에서 0반환을
실측 기록했다(2026-08-23, 본 트리 — `default_tools_approval_mode`·`checkCodexWiring`·
`"Codex Wiring"`·`/hooks to re-trust`·`codex /hooks` 전부 0hit. 2026-08-24 v0.3.0 추가 —
`statusLineAllowlist` 0hit, `17827`(README 4종·docs-site/content) 0hit). 채택 제외 토큰 2종:
`/hooks` 단독(기존 문구 충돌 — 히트수는 범위 의존적이라 명령으로 귀속한다:
`grep -rn 're-approve\|/hooks' internal/cli/*.go | grep -v _test | wc -l` → **10**, 글롭형·
비재귀, internal/cli 최상위 .go 한정; 재귀 범위를 넓히면 수치가 달라진다)와 `status_line`
단독(internal/ Go 46hit — Claude statusline 코드와 충돌; 스크래치 산출물 파일 대상 grep으로만
사용하고 저장소 코드 토큰은 `statusLineAllowlist`로 대체).

## §F. Constraints (non-functional)

- **C-HRA-008**: CLI 경로에서 AskUserQuestion 금지 — 모든 상호작용은 플래그·positional·구조화
  stderr 오류로만.
- **§14 하드코딩 금지**: 환경변수명은 `internal/config/envkeys.go` 상수, 임계값·문구 상수는
  패키지 상수로. 테스트의 Codex 격리는 `CODEX_HOME`(scratch) 위생을 따른다(t83 방법론).
- **크로스플랫폼**: 산출 파일은 JSON/TOML 텍스트(플랫폼 중립). 생성기·런타임은
  darwin/linux/windows에서 컴파일·동작(`GOOS=windows` vet 게이트). PATH 해석 `moai` 명령의
  Windows 실행 세부는 run-phase 검증 항목(§H).
- **16언어 템플릿 중립성**: template 반입 물질에는 SPEC-ID·내부 날짜·SHA 불포함(본 SPEC은
  template 추가를 예상하지 않는다 — §C Out of Scope).
- **무음 실패 금지**: Codex의 실패 양상은 조용하다(파일 무력화·해시 정지). 본 SPEC의 모든
  경로(생성·재생성·doctor)는 상태를 기계적으로 관측 가능하게 남긴다(사이드카·ValidateConfig·
  divergence 보고).
- **best-effort 배선**: 배선 실패는 init을 실패시키지 않는다 — `.mcp.json` provisioning과
  동일한 경고-후-계행 원칙(단, REQ-CW-003의 "쓰기 거부"는 보존: 검증 실패 파일은 디스크에
  남지 않는다).

## §G. Dependencies

| 대상 | 유형 | 내용 |
|---|---|---|
| SPEC-CODEX-HOOK-ADAPTER-001 (M3, release PR #1602) | **강함 (depends_on)** | `internal/codexadapter` 전체 표면을 소비. run-phase 베이스에 PR #1602 병합 필수 — **2026-08-24 충족(915c310de merge + rebase, pre-flight 실측 초록)** |
| SPEC-CODEX-DUAL-AGENTS-001 (M5, 동일 PR) | 참조 (related) | `.codex/agents/moai/*.toml` 배포면 — 본 SPEC은 무변경(REQ-CW-012) |
| SPEC-CODEX-SKILLS-CANONICAL-001 (M1, 동일 PR) | 참조 (related) | `.agents/skills` 미러 — 본 SPEC과 무관하나 같은 배포 체인 |
| SPEC-CODEX-SESSION-MSG-001 (t187, PR #1606 미병합) | 정합성 (related) | §A.4 — `writes` capability 기반 판정으로 병합 순서 독립. annotation 가드(REQ-CW-011)가 양 트리에서 성립 |

> **브랜치 격리 주석 (plan-audit review-1 D3; review-3 D2 갱신)**: 작성 시점(76b2c4ece)에는
> 위 4개 SPEC 디렉터리가 이 브랜치의 `.moai/specs/`에 없었다(#1602·#1606 미병합 산출물).
> 2026-08-24 #1602 병합(915c310de) 후 rebase로 4종 중 3종(HOOK-ADAPTER·DUAL-AGENTS·
> SKILLS-CANONICAL, 전부 `completed`)이 이 트리에 존재한다. 잔여 부재는 #1606(t187,
> SESSION-MSG)뿐 — 병합 순서 독립성은 §A.4가 이미 담보하므로 run 차단 아니다.

## §H. Risks

| 위험 | 완화 |
|---|---|
| `features.hooks` config 키의 기본값 미확정 — 특정 빌드에서 기본 off라면 hooks.json이 로드 자체가 안 될 수 있음 | 측정되지 않은 키 주입 금지 원칙 유지(§C). run-phase 검증 항목 #1: 격리 CODEX_HOME에서 프로젝트 훅 로드 재확인(운영자 실측 0.149.0-alpha는 default로 발화 관측). 필요 시 후속 SPEC에서 managed 키 추가 |
| Codex 신뢰 저장소 비문서화 — 사이드카 divergence가 실제 신뢰 상태와 어긋날 수 있음 | doctor 안내문을 단정("정지됨")이 아니라 조치 지시("/hooks 확인")로 서술. 재생성 시점 안내(REQ-CW-008)가 1차 방어 |
| Windows에서 PATH 해석 `moai`(및 `.exe` 해석) 미검증 | 산출물은 텍스트라 위험은 런타임 실행 한정. run-phase: `GOOS=windows` vet + 문서화된 수동 확인. 필요 시 wrapper 파일 폴백(skill_mirror 선례)을 후속 카드로 |
| t83 측정(0.147.0)과 운영자 실측(0.149.0-alpha)의 버전 혼재 | hooks.json 스키마는 두 버전 모두에서 동작 관측(0hit 키 없이 최소 집합 emit). doctor가 ValidateConfig으로 상시 검증 |
| config.toml 텍스트 레벨 테이블 편집의 파싱 경계(인라인 테이블·코멘트 변형) | 생성기는 create-if-absent + 테이블 미수정 원칙(REQ-CW-005)으로 편집 경로를 최소화. 갱신은 doctor의 drift 보고로 대체 |

## §I. Cross-References

- 감사 원본: `.moai/reports/t88/codex-support-audit.md` (§2.1 init 플래그, §2.4 어댑터, §2.5 M5,
  §2.6 MCP 부재, §3 갭 1-5)
- 운영자 정정: `.moai/reports/t88/moai-desktop-compat-20260823.html` "핵심 갱신" 절
- M3 실측: `.moai/reports/t83/precondition-measurement{,-round3}.md` + `probe/` (release/v3.1.3 수록)
- CLI·템플릿 아키텍처: `.claude/skills/hns-moaiadk-patterns` (Template-First, add-a-hook)
- Codex 공식 문서(§D의 URL 목록은 §A.4/§A.5에 인용 위치별 기재)
- statusline(v0.3.0): `learn.chatgpt.com/docs/developer-commands` `/statusline` 절 ·
  `learn.chatgpt.com/docs/config-file/config-reference.md`(tui.status_line) ·
  `learn.chatgpt.com/docs/config-file/config-sample.md`(기본값) ·
  소스 `openai/codex` `codex-rs/tui/src/bottom_pane/status_line_setup.rs`(식별자 원천) ·
  openai/codex#17827(command-backed 미지원) — §A.6 원문 인용
