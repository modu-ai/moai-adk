# Codex 상태 표시 조사 문서 epoch-3 독립 재검증 기록 (#2357)

- 재검증: agent-10-t6b, `session=agent-10-t6b`, epoch 3, `Moai-Claim: 2357#6759371745`(2026-11-10T04:55:42Z, lease 08:55:42Z).
- 검증 대상: 커밋 `73afc14cd6`(부모 `6bf79ecd85`)에 있는 조사 문서 `codex-statusline-official-scope-2026-11-10.md`와 출처 원장 `…-sources.json`.
- 방법: 인계된 sources.json의 성공 소스 11건을 HTTP GET으로 전부 재확보(읽기 전용, 세션 scratch인 gitignored `target/epoch3-verify/`에 저장 후 저장소 미커밋)하고 HTTP 상태·bytes·SHA-256을 재계산해 인계 원장과 대조했다. 조사 문서 §2 판정 표 12행과 §3·§4·§7의 핵심 주장은 재확보한 원문에서 직접 확인했다. 재검증 시각은 각 재확보 직전 `date -u` 측정값이다.
- 판정 변동: 없다. epoch 3은 supported/unsupported/not_verified 판정을 하나도 바꾸지 않았고, 공식 원문과 어긋나는 사실 서술 2곳을 정정했다(§3).

## 1. 소스 재확보 결과 (sources.json 대조)

| ID | URL | 재확보 시각(UTC) | HTTP·bytes | SHA-256 재계산 | 인계 원장 대조 |
|---|---|---|---|---|---|
| S01 | developer-commands.md | 2026-11-10T04:59:11Z | 200 · 64,320 | `d65c47d1…58181` | 일치 |
| S02 | config-file/config-reference.md | 2026-11-10T04:59:11Z | 200 · 153,298 | `78b5f4af…70c7e96` | 일치 |
| S03 | config-file/config-sample.md | 2026-11-10T04:59:12Z | 200 · 36,506 | `f5200b38…6dab8664` | 일치 |
| S04 | config-file/config-advanced.md | 2026-11-10T04:59:12Z | 200 · 56,830 | `ce85d4b5…5b5bcaf` | 일치 |
| S05 | hooks.md | 2026-11-10T04:59:13Z | 200 · 48,400 | `a50c60b6…f18ab6d782` | 일치 |
| S06 | reference/settings.md | 2026-11-10T04:59:13Z | 200 · 5,007 | `b66dc62c…b4e973e0` | 일치 |
| S07 | app/commands.md | 2026-11-10T04:59:14Z | 200 · 26,950 | `c5e9b64b…1130267233` | 일치 |
| S08 | app-server.md | 2026-11-10T04:59:14Z | 200 · 106,105 | `b9e75507…5ea93efe77f2` | 일치 |
| S09 | changelog | 2026-11-10T04:59:15Z | 200 · 4,771,490 | `443bf502…6ee46c3b3` | **불일치** (인계 4,771,489 bytes 대비 +1) |
| S10 | extend/mcp.md | 2026-11-10T04:59:15Z | 200 · 31,613 | `faf94077…b35c29e3b` | 일치 |
| S11 | mcp-server.md | 2026-11-10T04:59:15Z | 200 · 1,239 | `9fe22399…4f8af32da` | 일치 |

10/11은 HTTP 상태·bytes·SHA-256 전부 일치했다. S09(changelog)만 응답이 1바이트 커져 해시가 달라졌다. 살아있는 HTML 응답의 동적 요소(nonce·타임스탬프류) 차이로 추정하지만 1바이트의 위치는 특정하지 않았다. 문서가 S09를 근거로 든 두 항목은 재확보 원문에서 그대로 확인되므로(§2) 근거 자체는 유지된다. epoch 2 기록의 config-advanced "2바이트 차이"는 web-fetch 기준 측정이었고, 이번 curl 재측정에서는 56,830 bytes로 인계 원장과 정확히 일치했다.

## 2. 주장별 원문 확인 (재확보 원문 인용)

| 조사 문서 주장 | 원문 근거 (재확본 위치) | 결과 |
|---|---|---|
| `/statusline`, `/title` 존재 (행1·2) | S01 497–498행: "Configure TUI status-line fields interactively", "Configure terminal window or tab title fields interactively" | 확인 |
| footer 내장 항목 목록 (행1) | S01 849행: "model, model+reasoning, context stats, rate limits, git branch, token counters, session id, current directory/project root, and Codex version" | 문서 표와 일치 |
| title 내장 항목 목록 (행2) | S01 869행: "app name, project, spinner, status, thread, git branch, model, and task progress" | 문서 표와 일치 |
| picker 즉시 반영·config.toml 저장·세션 한정 인자 (행1) | S01 843–844행 "updates immediately and persists to `tui.status_line` in `config.toml`", 850행 "To set the footer for this session only" | 확인 |
| `tui.status_line` (행1), `null` 비활성 (행35) | S02 1668–1672행: type "array<string> \| null", "`null` disables the status line." | 확인 |
| `tui.terminal_title` (행2) | S02 1674행 | 확인 |
| `tui.notifications`·`tui.notification_method`(auto/osc9/bel)·`tui.notification_condition` (행4) | S02 1602–1608행, 1603행 "auto \| osc9 \| bel" | 확인 |
| `notify`·`agent-turn-complete`·단일 JSON 인자 (행5) | S04 719행 "Use `notify` to trigger an external program whenever Codex emits supported events (currently only `agent-turn-complete`)", 766행 "The script receives a single JSON argument." | 확인 |
| hook stdin JSON `session_id`·`transcript_path`·`hook_event_name`·`model`·`turn_id` (행6) | S05 482–488행 (482행 "Subagent hooks use the parent session id", 488행 turn-scoped `turn_id`) | 확인 |
| `statusMessage`는 hook 실행 중 표시, `systemMessage`는 경고 UI/event stream (행6) | S05 271행 "Optional message shown while the hook runs", 521행 "Surfaced as a warning in the UI or event stream" | 확인 |
| transcript 파일 형식은 안정 인터페이스 아님 (행6) | S05 496–497행 "the transcript format isn't a stable interface for hooks and may change over time" | 확인 |
| Desktop 완료 알림·외관·활동 기본 기능 (행7) | S06 Notifications 절 "Choose when turn completion notifications appear", Appearance 절, Profile 절 "activity insights, lifetime tokens, peak tokens" | 확인 |
| Desktop 명령·단축키의 terminal·side chat (행7) | S07 "Toggle terminal (Codex only)", "Open side chat (Codex only)" | 확인 (명칭 정정 §3-1) |
| Desktop 커스텀 footer/Rust command slot 부재 (행8 not_verified) | S06·S07 원문 전체 스캔에서 그 계약 문구 미발견 | not_verified 유지 |
| app-server thread/turn·plan·token usage·rate limit 이벤트 (행9·§3) | S08 129행 `turn/started`, 171·234행 `turn/completed`, 338·844행 `thread/status/changed`, 1420행 `turn/plan/updated`, 1425행 `thread/tokenUsage/updated`, 2043–2044행 `account/rateLimits/read`·`updated` | 확인 |
| WebSocket experimental/unsupported (행9) | S08 61행 "`websocket` (`--listen ws://IP:PORT`, experimental and unsupported)", 76행 "WebSocket transport is experimental and unsupported"; `experimentalApi` 게이팅 291–296행 | 확인 |
| `codex mcp-server`·standalone 바이너리 제거 (행11) | S11 전문: "The `codex mcp-server` command and the standalone `codex-mcp-server` binary have been removed", 18–19행 "The app-server command is experimental and isn't supported for production workloads." | 확인 |
| 외부 MCP 서버 STDIO/Streamable HTTP (행10) | S10 190–192행 "**STDIO servers**… **Streamable HTTP servers**", 217행 "choose **STDIO** or **Streamable HTTP**" | 확인 |
| changelog 실험 UI renderer (행12) | S09 830행 "Add an experimental UI tree renderer for the status line" | 확인 |
| changelog credits/cost 자격 범위 (행13) | S09: "View estimated thread credits or cost in `/status`, status lines, and terminal titles for eligible workspaces." | 확인 |

## 3. epoch 3 정정 2건 (판정 변동 없음)

1. **행7의 "Codex view"** — S07 재확보 원문에 "Codex view"라는 명칭은 없다. 실측되는 명칭은 "Toggle terminal (Codex only)", "Open side chat (Codex only)", "Open review tab (Codex only)"다. 조사 문서 §2 행7을 이 실측 명칭으로 고쳤다. 이는 Desktop UI 지원 판정을 바꾸는 정정이 아니라 기능명 표기의 정확화다.
2. **행35의 "sample은 빈 배열 `[]`을 설명한다"** — S03 재확보 원문에서 status_line 관련 표기는 주석 예시 `# status_line = ["model", "context-remaining", "git-branch"]`(686행)뿐이고 비활성 빈 배열 표기는 없다(`[]`는 `project_doc_fallback_filenames`·`writable_roots` 등 다른 키의 예시에 있다). 조사 문서 §35 문단을 실측 서술로 고쳤다.

t3j 원문은 커밋 `73afc14cd6`의 git history에 보존되어 있다.

## 4. Desktop 지원 경계 재확인 (지시 조건 이행)

재검증 지시 조건은 "공식 소스가 보여주지 않으면 Desktop UI 지원을 진술하지 않는 것"이다. 이행 기록:

- 행7 supported는 Desktop의 **기본 UI 기능**(Settings의 완료 알림·외관·활동, Codex 전용 terminal·side chat·review tab 단축키)에 한정되며 S06·S07 재확보 원문이 그 내용을 실제로 담는다.
- Desktop의 사용자 정의 고정 footer/Rust command slot은 **not_verified 유지** — S06·S07 원문 전체에서 그 계약 문구를 찾지 못했고, `tui.*` 키가 Desktop 확장 슬롯이라는 근거도 없다. 이번 재검증에서 Desktop 관련 판정을 상향하지 않았다.
- S09의 changelog 항목은 CLI 상태줄 renderer·credits 표시 언급이며 Desktop 표시 확정 근거가 아니다(조사 문서 행12·13이 이미 같은 경계를 명시).

## 5. 한계와 not_run

- S09 해시 불일치의 1바이트 차이 원인은 특정하지 않았다. 동적 응답 요소 차이 추정이며, 인계 응답 원문이 세션에 없어 두 응답의 직접 diff는 불가했다. 근거 텍스트 2건의 존재로 주장 검증에는 영향이 없다.
- 부재 주장(행3 외부 상시 상태줄, 행8 Desktop 커스텀 footer)은 키워드 스캔으로 계약 문구를 찾지 못한 것이며, 문서 부재가 기능 부재의 증거가 아니라는 조사 문서 §1 원칙을 그대로 따른다.
- 라이브 동작 검증(Codex CLI/Desktop·hook·notify·MCP·app-server 실행), 설치본 버전 확인, Windows/Linux 확인은 이번에도 `not_run`이다 — 조사 문서 §6과 같은 경계다.
- 검증에 쓴 재확보 원문은 gitignored `target/epoch3-verify/fetched/`에만 있고 저장소에 커밋하지 않았다. 근거 인용은 본 문서 §2의 원문 인용과 sources.json의 해시 원장으로 남는다.

## 6. epoch 4 감사 수리 기록 (2026-11-10, 감사 `6762170355`의 F1·F2에 따름)

- **F1 — `/status` 누락 수리**: 독립 감사(agent-4-t63)가 Desktop·CLI 모두의 공식 표시 수단 `/status`가 §2·§4·§5 표에서 빠져 있다고 지적했다. 수리로 새 출처 **S12**(`reference/slash-commands.md`)를 재확보했다 — **2026-11-10T08:08:36Z, HTTP 200, 5,012 bytes, SHA-256 `31116b5621fa26a4fe944a92517cab42941ca18f68495f04b1f19370571b7c07`** — 감사자의 재수신 기록(07:53:00Z, 같은 bytes·해시)과 완전히 일치한다. 원문 48행 "`/status` — Show the chat ID, context usage, and rate limits."를 직접 확인했다. S01도 재독했다 — **2026-11-10T08:08:36Z, HTTP 200, 64,320 bytes, SHA-256 `d65c47d1…`(인계 원장과 일치)** — 494행 "Display session configuration and token usage … remaining context capacity", 806–813행 "Inspect the session with `/status`", 14–26행의 Desktop 명령 안내(Slash commands 페이지에 `/status` 명시)를 직접 확인했다. 조사 문서 §2에 CLI·Desktop `/status` 행을 신설(요청형·일시 표시 한계 명시)하고 §4 context·사용 한도 행과 §5 선택지(B′)에 반영했다.
- **F2 — §6 시점 정정**: 조사 문서 §6의 not_run·미실행 문장들이 epoch 1 시점 기록인데 현재 사실처럼 읽힌다고 지적됐다. §6 머리에 범위 주석(epoch 1 이력임을 명시, epoch 2–4 진행 안내)을 두고 각 문장을 epoch 1 시점 표기로 고쳤다.
- **S09 주석 권고 이행**: §7 원장과 sources.json의 changelog-full 엔트리에 재현 불일치 주석을 달았다(epoch-1 해시 `5ca7955b…`·4,771,489 bytes는 이후 재확보에서 `443bf502…`·4,771,490 bytes로 재현되지 않음 +1바이트 동적 요소, 근거 문장 2건은 재확보 원문에서 그대로 확인 — 본 문서 §1과 감사자 재수신(10/11 일치)이 같은 값을 보고).
- **비차단 표현 정정**: PR 본문의 "`73afc14cd6`의 trailer" 표현은 interpret-trailers가 파싱하지 않는 형태라 감사자가 부정확하다고 봤다 — 본문 표현을 line-match 규칙 기준 표현으로 고친다.
- 라이브 동작·설치본·Windows/Linux는 이번 수리에서도 `not_run`이며, 감사자와 동일하게 본 카드의 경계를 유지한다.
