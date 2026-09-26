# Research — SPEC-HANDOFF-NEUTRAL-001 (핸드오프 하네스 중립화, card t1273)

> Plan-phase 조사 산출물. 코드베이스 사실은 worktree HEAD `1b7a88d78` (branch `WT-handoff-neutral`, develop 분기점)에 대해 grep/Read로 실측했다. 외부 문서 사실은 2026-09-26 관측이며, 각 출처의 검증 상태를 §F에 명시한다 — 카드 [HARD] 요건(출처 URL 검증, 미확인 표시) 준수.
>
> 마일스톤 구조는 리드 승인 B안+조정(2026-09-26)을 따른다: **M1 = 조사·설계 + Codex 쪽 저장·주입 경로(새 능력) + 교차 인계 LIVE 검증**. ultrathink·하네스 고유 키워드 제거(SSOT 문서 + 렌더 코드 동시)는 **M2(t1175 병합 뒤)**. AGENTS.md 계약 배치는 **M3(t1243 뒤)**.

---

## §A — 조사 범위와 결론 요약

카드의 문제 선언: Claude Code↔Codex 교차 인계에서 컨텍스트가 손실된다. 4개 축을 조사했다.

| 축 | 조사 전 가설 | 실측 결론 | 영향 |
|------|-------------------------------|---------------------------|------|
| A.1 현재 저장·주입 경로의 하네스 편향 | `moai handoff save`·인젝터가 Claude Code 전용일 것 | **확인** — handoff 표면(internal/cli/handoff.go, internal/hook/handoff/, internal/hook/handoff_inject_render.go)에서 `codex` 참조 **0건** (grep -rni, exit 0·빈 출력). 인젝터는 Claude Code SessionStart `additionalContext` 경로에만 묶임 | M1의 "Codex 쪽 저장·주입 경로"는 신규 구축이지 수리가 아님 |
| A.2 Codex의 세션 시작 주입 수단 | "훅 이벤트 유무"가 카드 조사 과제였음 | **전제 갱신** — Codex에도 라이프사이클 훅이 존재한다: config `features.hooks` + `hooks.<Event>`(SessionStart 포함) + `additionalContextLimit`(기본 2500 토큰). 단 feature flag 필요(기본값 미확인, §F-3) | Claude 측과 구조 대칭인 기계 주입 경로가 이론상 존재 — LIVE 검증이 M1 관문 |
| A.3 크로스 하네스 세션 재개 가능성 | resume 계열이 교차 인계를 담을 수 있나 | **불가 확인** — 양 하네스의 resume은 자기 세션 저장소만 본다(Claude: `~/.claude/projects/.../<id>.jsonl` / Codex: 자체 rollout). Codex가 Claude 세션을 resume하는 수단은 공식 문서에 없음 | 6블록 텍스트가 유일한 교환 매체 — 카드의 중립화 방향이 올바른 축 |
| A.4 ultrathink 등 하네스 고유 지시 | 본문 공통 블록이 Codex에서 무의미한 키워드를 싣고 있을 것 | **확인 + 범위 정정** — 발화 지점은 3층: SSOT 문서(session-handoff.md 등), `moai handoff save --ultrathink` 플래그→`Directives` 필드, 인젝터의 4-로케일 복원 안내 라인. 인젝터 주석은 이미 "NEVER asserts ultrathink is active"(안내-only) — 렌더 코드는 단언이 아니라 안내를 싣고 있으므로 M2 제거는 SSOT 문서와의 동시 정합이 핵심 | 리드 조정의 근거와 일치: 한쪽만 바꾸면 문서↔출력 어긋남 |

---

## §B — 외부 공식 문서 조사 (하네스별 세션 연속성)

### B.1 Claude Code — 세션 관리 (검증됨, §F-1)

- `--continue`(최근 세션), `--resume <id|name>`, `/resume` 피커, `--fork-session`·`/branch`.
- resume이 복원하는 것: 대화 전체, 모델, 에이전트, 권한 모드, 활성 goal, 예약 작업. **복원 안 됨**: 런치 플래그(`--mcp-config`, `--settings`, `--add-dir` 등) — 세션 경계를 넘는 설정은 사용자가 다시 전달해야 한다.
- `/clear`: 새 빈 컨텍스트 시작. **이전 대화는 보존**돼 `/resume` 복귀 가능 — moai 자동 주입의 `clear` 소스가 의존하는 동작.
- transcripts: `~/.claude/projects/<project>/<session-id>.jsonl` (내부 형식, 버전 간 변경 예고).

### B.2 Claude Code — SessionStart 주입 경로 (검증됨, §F-2)

- SessionStart matcher: `startup | resume | clear | compact`.
- 주입 채널 2개: (a) 평문 stdout이 컨텍스트로 주입, (b) `hookSpecificOutput.additionalContext` — 첫 프롬프트 직전 삽입. moai 인젝터가 쓰는 경로가 (b).
- **10,000자 상한**: additionalContext/systemMessage/plain stdout 공통. 초과 시 파일로 저장되고 "경로 + 처음 2,000자 미리보기"로 강등 — 6블록 길이 예산의 기계적 상한.

### B.3 Codex — 세션·지시·훅 (검증됨, §F-3)

config.toml 키 기준(관측일 2026-09-26):

| 키 | 관측된 설명 | 중립화 관점 |
|---|---|---|
| `developer_instructions` | "Additional developer instructions injected into the session (optional)." string | 세션에 주입되는 공식 키 — 카드가 지목한 후보. 단 config 성격상 정적 |
| `features.hooks` | "Enable lifecycle hooks loaded from hooks.json or inline [hooks] config." + 폐기 별칭 `codex_hooks` | Codex 훅의 총 스위치. **기본값은 이번 판독에서 확인 못 함(미확인)** |
| `hooks.<Event>` | PreToolUse, PermissionRequest, PostToolUse, PreCompact, PostCompact, **SessionStart**, SubagentStart, SubagentStop, UserPromptSubmit, Stop | SessionStart 존재 — Claude 측과 이벤트 대응 |
| `hooks.<Event>[].hooks[].async`·`additionalContextLimit` | additionalContextLimit 기본 **2500 토큰**, 초과 시 디스크 저장 + 짧은 프리뷰 | Claude의 10,000자 상한에 대응하는 Codex 상한. 6블록이 2500 토큰을 넘으면 강등됨 |
| `project_doc_max_bytes` | "Maximum bytes read from AGENTS.md when building project instructions." | AGENTS.md 예산 스위치. **32KiB 기본값은 이 표에 명시 없음 — 수치는 미확인**(§F-4 보조 출처) |
| `model_instructions_file`·`project_doc_fallback_filenames` | AGENTS.md 대체/보조 파일 | M3 배치 논의 입력 |
| `tui.resume_cwd` | "Working directory to use when resuming or forking a session" | `codex resume`/fork 존재의 1차 근거(전용 문서 페이지는 없음 — 부분 검증) |
| `projects.<path>.trust_level` | `"trusted"\|"untrusted"` — untrusted는 project-scoped .codex/ 레이어(로컬 config, hooks, rules) skip | Codex 측 주입이 프로젝트 단위 신뢰에 묶임 |
| `history.persistence`(save-all\|none)·`features.memories` | 세션 지속·메모리(기본 off) | Codex→Codex 연속성의 보완 축 |

또한 "Project-scoped config can't override machine-local provider, auth, …" 제외 목록에 `developer_instructions`·hooks는 **들어 있지 않다** — 프로젝트 단위(.codex/config.toml, trust 필요) 설정이 가능하다는 읽기가 성립한다(공식 명시는 아님 — LIVE 검증 항목).

### B.4 Anthropic 하네스 설계 참고 (검증됨, §F-5)

장기 실행 앱 하네스 설계문: 컨텍스트 리셋 시점마다 **구조화된 핸드오프 아티팩트**(진행 파일)를 남기고, 에이전트 간 통신은 파일 기반으로, 평가자는 실행자와 분리한다. — moai의 pending.json + progress.md 패턴과 같은 방향; 6블록을 "양 하네스가 같은 방식으로 소비하는 파일"로 유지하라는 정당화 근거.

---

## §C — 현재 moai 구현 실측 (HEAD 1b7a88d78)

### C.1 저장 — `moai handoff save` (internal/cli/handoff.go)

플래그: `--stdin`/`--body`, `--spec`, `--phase`, `--session`, `--lang`, `--ultrathink`, `--ultracode`, `--goal` (handoff.go:115-123). `--ultrathink`/`--ultracode`/`--goal`는 "restoration guidance only" — 본문이 아니라 복원 안내 메타데이터.

`PendingRecord` (internal/hook/handoff/pending.go:53-67): `SchemaVersion`, `SpecID`, `Phase`, `SavedAt`, `SavedBySession`, `ConversationLanguage`, `Directives{Ultrathink, Ultracode, Goal}`, `EmbeddedGoal`, `Body`. — **저장 파일(.moai/state/handoff/pending.json)과 CLI는 하네스 무관**이다. moai CLI가 쓰는 파일이지 Claude Code가 쓰는 파일이 아니다. 편향은 저장이 아니라 소비에 있다.

### C.2 주입 — Claude Code SessionStart 전용 (internal/hook/handoff_inject_render.go)

`renderHandoffContext`: header + disclaimer + (Directives 있으면) 복원 안내(ultrathink/ultracode/goal, 4-로케일 ko/ja/zh/en: handoff_inject_render.go:86-119) + Body verbatim. 파일 주석이 설계 계약을 명시: "The injected text NEVER asserts that ultrathink / an extended-reasoning mode is active — a hook cannot change effort/model. Mode-change directives are rendered as manual-paste restoration guidance only."

즉 렌더 코드는 이미 (a) 지시를 단언하지 않고 (b) 안내로만 싣는다. M2의 "ultrathink 제거"는 이 안내 라인과 SSOT 문서 본문을 같은 마일스톤에서 맞추는 작업이지, 잘못된 단언을 고치는 작업이 아니다.

### C.3 Codex 배포면 부재 (실측)

- `internal/template/templates/.codex/` — `agents/`(agentemit이 방출하는 .toml)만 존재. **템플릿 배포판에는** handoff 저장·주입 인프라가 없다. "handoff" 문자열 등장은 manager-lead.toml·manager-design.toml 본문 언급뿐(에이전트 지시 텍스트). 단, 로컬 런타임 생성 배선은 이미 존재한다(§C.4).
- AGENTS.md는 템플릿 루트에 `AGENTS.md.tmpl`(19,177 bytes, 렌더링 필요)로 배포된다 — M3 배치 논의의 현 물성.

### C.4 Codex 훅 배선·어댑터는 이미 부분 존재 (실측 — 카드 가정보다 앞선 상태)

- primary 체크아웃의 `.codex/hooks.json` (**untracked** — 워크트리 t1273 체크아웃에는 없음 = develop 트리에 없음 = 런타임 생성)에 `moai hook session-start --harness codex` 배선이 존재한다. 운영자 `~/.codex/config.toml`의 `[hooks.state]`에 `session_start` 실행 기록이 있다.
- `--harness codex` 어댑터가 이미 구현돼 있다: `internal/cli/hook_harness_codex.go` (SPEC-CODEX-WIRING-001 M3) + `internal/codexadapter/` — MoAI 훅 출력을 Codex 형식으로 재작성 (continue:false→decision:block, 이벤트별 systemMessage→additionalContext).
- **정확한 갭**: `internal/codexadapter/output.go:84-88`의 `additionalContextEvents` 집합은 `UserPromptSubmit` **하나뿐** — 주석이 명시한다: "Only UserPromptSubmit was measured delivering it." 반면 handoff 인젝터는 `EventSessionStart`에 등록된다 (`internal/hook/handoff_inject.go:41`). 즉 Codex 경로에서 인젝터가 내놓는 additionalContext는 **버려진다** (어댑터가 매핑하지 않음). P1의 실체는 새 배선 구축이 아니라 **SessionStart 채널의 전달 실측 + 매핑 집합 추가**다.
- **환경 관측 (2026-09-26, codex exec 1회)**: 이 워크트리에서 `codex exec` 최소 호출이 `hook: SessionStart` ×3 발화 + `hook: SessionStart Completed` ×3를 관측했다 — Codex CLI가 SessionStart 훅을 실제로 실행한다는 1차 확인 (관문 a). 사후 판정: 발화한 핸들러는 글로벌 `~/.codex/hooks.json` 소속 운영자 외부 훅 2개(orca·luvus)이며 moai 훅은 아니었다 — 워크트리에 `.codex/hooks.json`이 없어 moai 프로젝트 훅이 로드되지 않았다. moai 관점 상태 변화는 없다(pending.json 부재·consumed 미변동·세션 레지스트리 기록 없음).

### C.5 세션-메시징 브로커와의 관계 (계층 구분)

Codex↔Claude 실시간 메시징은 이미 `mcp__moai__session_msg_*` 브로커(cross-session-messaging.md § Codex broker path)가 담당한다. 그러나 그것은 **살아 있는 세션끼리의 넛지** 채널이고, 핸드오프는 **세션 경계를 넘는 상태 이전**이다 — 세션이 죽은 뒤에도 소비돼야 하므로 pending.json 같은 디스크 매체가 필요하다. 두 계층은 보완 관계이며 이 카드는 후자만 다룬다.

---

## §D — 크로스 핸드오프 갭 분석

교차 인계(Claude→Codex, Codex→Claude)에서 실제로 끊기는 지점:

1. **주입 부재 (심각)**: Claude 세션이 `moai handoff save`로 pending.json을 남겨도 Codex 세션 시작 시 그것을 읽는 경로가 없다. 사용자가 6블록을 보기 위해서는 Claude 쪽 응답 본문에서 복사해 Codex에 붙여넣는 수작업이 유일하다 — 수동 paste 경로는 하네스 중립적(텍스트이므로)이지만 저장→재발견 사슬이 끊긴다.
2. **본문의 하네스 고유 토큰 (중간)**: 6블록 Block 1의 `ultrathink.` 오프너, Block 5의 `/moai run`류 슬래시 명령은 Claude Code 세션에서만 의미가 있다. Codex에 붙여넣으면 `ultrathink.`는 무해한 잔문(해석 불가 토큰)이고 `/moai run`은 moai CLI 호출 지시로는 유효하나 슬래시-커맨드 형태가 아니다. **M2가 다루는 축** — 단 SSOT(session-handoff.md)와 렌더(handoff.go·pending.go·handoff_inject_render.go)가 같은 문구를 다루므로 동시 변경(리드 조정)이 선행 조건.
3. **Codex→Claude 방향 (중간)**: Codex 세션이 반대 방향 인계를 남길 표준 형식이 없다. moai handoff save는 CLI라 Codex 세션도 실행할 수 있으므로(shell 호출), 형식이 중립적이면 양방향이 같은 매체를 쓴다 — 방향 비대칭은 없다.
4. **에이전트 정의 사본 (낮음, 범위 밖)**: Codex가 받는 .toml 에이전트 정의(agentemit)는 이미 존재. 이 카드는 에이전트가 아니라 세션-수준 인계를 다룬다.

---

## §E — M1 설계 시사점: Codex 저장·주입 경로 후보

카드가 지목한 `moai codex developer_instructions` 후보를 포함해 4개 경로를 비교한다. 최종 선택은 design.md; 여기는 근거.

| 후보 | 형상 | 강점 | 약점 | 권고 |
|---|---|---|---|---|
| **P1 — Codex hooks (features.hooks + hooks.SessionStart + additionalContext)** | `.codex/` 프로젝트 레이어에 hook 정의; moai가 pending.json을 소비해 additionalContext로 주입 | Claude 측 인젝터와 구조 대칭 — 같은 pending.json, 같은 복원 안내. 기계적 자동 주입 | feature flag 필요(기본값 미확인); `additionalContextLimit` 2500 토큰 — 6블록이 넘으면 디스크+프리뷰 강등; trust_level 의존; **LIVE 검증 전에는 존재·동작이 가설** | M1 후보 — LIVE 검증 관문 통과 시 |
| **P2 — developer_instructions (config.toml)** | `moai handoff save`가 resume 본문을 Codex config 레이어에 기록 | 공식 "injected into the session" 키 — 문서상 주입 보장 | config는 **영속적 정적 지시** 매체. 세션마다 다른 휘발성 resume를 쓰면 이전 세션 잔문이 누적·오염되고, 소비 시점(어느 세션이 읽는가)을 save가 제어할 수 없음. 사용자 config 오염 위험 | **부적합 판정 권고** — 카드가 지목했으나 정적/휘발성 특성이 어긋남. 근거는 design.md에 상세화 |
| **P3 — AGENTS.md 소비 계약 + `moai handoff show`** | AGENTS.md(.tmpl)에 "세션 시작 시 pending.json 확인" 절차 2-3줄 추가; `moai handoff show`가 저장본을 붙여넣기 가능한 형태로 재출력 | 하네스 중립(파일+텍스트); AGENTS.md는 Codex가 **항상** 읽는 문서(32KiB 예산 내); 즉시 구현 가능; fail-open 정합 | 자동 주입이 아님 — 모델이 절차를 따르는 간접 주입. AGENTS.md 예산 소비(M3 t1243 조정 필요) | M1 병행 권고 — P1과 배타가 아님 |
| **P4 — codex resume (같은 하네스 내 연속성)** | Codex→Codex 세션 재개 | Codex 쪽 연속성 보완 | **크로스가 아님** — Claude 세션을 재개할 수 없으므로 카드의 교차 인계 축이 아님 | 범위 밖 표시(카드 범위와 구분 기록) |

**LIVE 검증 설계 입력 (카드 [HARD] — 상한 선언 후 실행)**: P1 타당성은 세 관문으로 좁힌다 — (a) 이 환경의 Codex가 `features.hooks`+SessionStart를 실제로 발화하는가, (b) additionalContext가 세션에 도달하는가, (c) 2500 토큰 한계에서 6블록 실측 길이(≈600-900 토큰 예상)가 강등 없이 통과하는가. 상한은 관문당 실행 3회·벽시계 30분을 선언한다(교훈: LIVE 측정은 상한 선언 후 — feedback_live_measurement_needs_declared_caps).

**관문 상태 (2026-09-26 갱신 — 리드 조건①에 따른 쿼터 선확인 결과)**:
- **쿼터 차단 확인**: `codex exec` 최소 호출 1회가 `You've hit your usage limit... try again at Sep 28th, 2026 2:37 PM`를 반환 (t1203 레인 실측과 동일). 관문 (b)(c)는 **9/28 14:37 이후**로 연기하고, 그동안 design.md는 리드 지시대로 **"P1 조건부(관문 통과 시) + P3 기본"** 구조로 진행한다.
- **관문 (a) 1차 통과**: 같은 호출에서 `hook: SessionStart` ×3 발화·`Completed` 관측 — SessionStart 훅 실행은 이 환경 사실이다. (모델 응답은 쿼터로 차단돼 (b) 관측은 불가했다.)
- 관문 (b)가 P1의 핵심 잔여 위험이다: UserPromptSubmit만 어댑터 측정돼 있고(§C.4) SessionStart 채널은 미측정 — Codex가 SessionStart의 additionalContext를 소비하는지가 additionalContextEvents 집합 확장의 전제.

---

## §F — 출처 목록 (검증 상태)

| # | 출처 | 상태 | 관측일 |
|---|---|---|---|
| F-1 | https://code.claude.com/docs/en/sessions | **검증됨** (webReader 직접 판독) | 2026-09-26 |
| F-2 | https://code.claude.com/docs/en/hooks | **검증됨** (동일) | 2026-09-26 |
| F-3 | https://developers.openai.com/codex/config-reference | **검증됨** (동일; 단 `features.hooks` 기본값은 표에서 확인 못 함 → 해당 수치만 미확인) | 2026-09-26 |
| F-4 | Codex `project_doc_max_bytes` 기본값 32KiB | **미확인(수치)** — GitHub 이슈 보조 출처만; F-3 표에는 키 존재만 명시. LIVE 검증(`codex exec` 등으로 실측) 전까지 근거로 인용 금지 | 2026-09-26 |
| F-5 | https://www.anthropic.com/engineering/harness-design-long-running-apps | **검증됨** (webReader 직접 판독) | 2026-09-26 |
| F-6 | `codex resume` CLI 문법(예: `--last`) | **부분 검증** — developers.openai.com에 전용 페이지 없음(검색 1차); F-3의 `tui.resume_cwd` 키로 resume/fork 존재만 1차 확인. 문법 상세는 GitHub 저장소 문서 보조 출처 | 2026-09-26 |
| F-7 | 환경 실험 관측 — `codex exec` 최소 호출 1회 (session 01a0dd57, 워크트리 t1273 cwd) | **실험 관측** — 쿼터 차단 메시지(9/28 14:37) + SessionStart 훅 발화·Completed 관측. **사후 판정**: 발화한 SessionStart 핸들러는 글로벌 `~/.codex/hooks.json` 소속 운영자 외부 훅 2개(orca codex-hook.sh·luvus-agent-hook.sh)이지 moai 훅이 아니다 — 워크트리 t1273에는 `.codex/hooks.json`이 없어 moai 프로젝트 훅은 로드되지 않았다. 상태 변화 없음: 워크트리·primary 모두 `pending.json` 부재(소매 없음), `consumed/` 미변동, moai 세션 레지스트리 기록 없음(moai 훅 미실행) | 2026-09-26 |

`developers.openai.com/codex/guides/agents`는 2회 시도 모두 500 오류 — AGENTS.md 가이드 공식 페이지는 미확인으로 둔다(이 카드는 config-reference 관측으로 충분).

---

## 기준 귀속

- 코드베이스 실측: worktree `/Users/goos/MoAI/moai-adk-go/.claude/worktrees/t1273`, HEAD `1b7a88d78`, branch `WT-handoff-neutral`, 2026-09-26.
- grep 명세: `grep -n "ultrathink" internal/cli/handoff.go internal/hook/handoff/pending.go internal/hook/handoff/persist.go internal/hook/handoff_inject_render.go` (10행); `grep -rni "codex"` 동일 대상 (빈 출력, exit 0).
- 외부 문서: z.ai webReader(GLM 백엔드 WebFetch 등가) 직접 판독; 검색(webSearchPrime)은 발견 수단으로만 사용.

## 미검증 (Gaps)

- Codex `features.hooks`의 기본 on/off — F-3 표에서 기본값 열을 이번 판독으로 확정하지 못했다. LIVE 검증 전까지 P1은 가설.
- Codex hooks의 SessionStart 이벤트가 실제 발화하고 additionalContext가 모델에 도달하는지 — 이 환경 실측 없음. M1 LIVE 검증 항목.
- 6블록의 Codex 측 토큰 실측 길이 — 예상치만 있음.
- `internal/cli/handoff.go` 전체 판독(플래그 표면·save 파이프라인 외 부분) — M1 구현 시 design.md 작성 단계에서 수행.

## 잔여 위험

- Codex 훅 표면은 실험적 feature flag로 보이므로(폐기 별칭 존재가 재설계 이력을 시사), P1이 검증을 통과해도 Codex 버전 변동에 취약하다 — P3(파일 기반)를 병행하는 이중화가 실제 권고의 골자다.
- Claude Code 문서는 버전 간 변경 여지가 있는 내부 동작(transcripts 형식 등)을 포함한다 — 본 문서 관측일(2026-09-26) 이후 드리프트 가능.
