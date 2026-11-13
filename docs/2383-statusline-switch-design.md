# #2383 설계(v3, 2차 수리판) — Claude Code 상태줄을 moai-cli로 전환(경로 이름 `moai` 제거 전제)

상태: 설계(design-only) v3 — v2(재감사 FAIL 6768843277: F1–F3, 보충 6768932473)의 수리판(agent-3-hB, claim 2383#6769016485, 리더 handoff 6769011126). **F1(Go moai 제거 시점과 Go 훅 가드)은 사용자 결정(Issue #2383 댓글 6769057244)대로** — #2391 준비 전까지 Go 훅만 Go 바이너리 절대 경로로 유지 — §2.1·§2.2·§5에 반영했다. v2는 v1(감사 FAIL, agent-3-t62 코멘트 6767987735, 차단 B1–B4)의 전면 재작성판이었다. 이 문서는 아무것도 바꾸지 않는다. 모든 실측은 2026-11-10 읽기 전용이며, 이 문서가 다루는 "실행"은 전부 후속 구현 단계의 것으로 본 PR은 실행하지 않는다(§5 각 단계의 직전 확인 포함).
요청: 사용자 지시 2026-11-10T11:50Z. 선행: PR #2374(#2350 Rust statusline) 병합·재설치 — #2374 자체는 수정하지 않는다.
v2 대비 정정은 §10 표에 있다. v1 대비 정정: 실제 경로를 프로젝트 기준으로 바로잡았고(§0), Go 플러그인 MCP 소비자를 인벤토리에 넣었으며(§1.1), 훅 전환을 "Go 스크립트 복사·sed" 대신 **Rust 책임 4개 기준 대체/폐기 표**(§3)로 다시 설계했고, 되돌리기를 진짜 역연산+해시 검증으로 고쳤으며(§6), 절차를 승인 순서(인벤토리 → 격리 검증 → 미리보기·백업 → 단계별 직전 확인)로 재배열했다(§5).

## 0. 대상 파일 원장 — 소유·mode·해시 (승인 순서 1단계: 맨 먼저)

모든 경로는 이 기계 실측(2026-11-10). `$PRI` = `/Users/goos/MoAI/moai-adk`.

| 파일 | 소유 | mode | sha256(앞 16) | 역할 |
| --- | --- | --- | --- | --- |
| `$PRI/.moai/status_line.sh` | goos | `-rwxr-xr-x@` | `2f5eb54cfb1094dd` | 상태줄 wrapper — settings.json `statusLine.command`(같은 파일 10행)이 이것을 실행 |
| `~/go/bin/moai` | goos | `-rwxr-xr-x` | `bbedd136ab623dd3` (1510B) | 라우팅 shim — **PATH에 있음**(`~/go/bin`). launcher 단어→Rust, Go 전용 64단어(hook·statusline·mcp-server·session 포함)→`~/go/bin/moai.old` |
| `~/.moai-cli/path/moai` | goos | `-rwxr-xr-x` | `bbedd136ab623dd3` (1510B) | 위 shim과 바이트 동일 사본 — 디렉터리가 PATH에 없음 |
| `$PRI/.moai-cli/install/moai-shim.sh` | goos | `-rw-r--r--@` | `bbedd136ab623dd3` (1510B) | shim **설치 원본** — live shim 2곳과 `cmp` 동일 실측. **git 미추적**(`git ls-files --error-unmatch` 실패)이고 mode가 실행 불가 `644`라 그대로는 되돌리기 원본이 될 수 없다 — 되돌리기는 §5-4에서 만든 mode 보존 백업을 쓴다(§6 R2) |
| `~/go/bin/moai.old` | goos | `-rwxr-xr-x` | — (78,709,760바이트 Go 바이너리) | Go 빌드 — 훅·statusline·mcp-server·session의 현재 실행체 |
| `~/.moai-cli/path/moai-cli` | goos | 심링크 | → `versions/2026.11.0-rc.1+f259f15a1/bin/moai-cli` | Rust 빌드(PR #2374의 statusline은 미포함 — 미병합) |
| `$PRI/.claude/hooks/moai/` (디렉터리) | goos | `drwxr-xr-x` | — | Go 훅 포워더 `.sh` **46개**(전부 .sh — 디렉터리에 비-.sh 항목 없음) |

참고: v1의 "디렉터리 49항목 중 비-.sh 3"은 셸 별칭(`ls`→`ls -l`)이 만든 착시였다. 정정: `ls -A` 항목 46 = `.sh` 46, 비-.sh 없음(이전 감사 t62 관측과 일치).

## 1. 활성 consumer 인벤토리 (승인 순서 1단계 — "무엇이 지금 moai를 쓰는가")

### 1.1 Go 플러그인 MCP 서버 — **v1에서 누락된 살아 있는 소비자 (B1 정정)**

- **v3 갱신 — 플러그인은 사용자가 이미 제거했다(2026-11-10T13:15Z 실행, 리더 기록).** 13:39Z 실측: `~/.claude/settings.json` `enabledPlugins`에 `moai@moai-adk` 항목 없음. 캐시 디렉터리 `~/.claude/plugins/cache/moai-adk`는 남아 있다(삭제는 이 설계 범위 밖, 사용자 확인 사항).
- 제거 전 설치본 `.mcp.json`은 `{"mcpServers": {"moai": {"command": "moai", "args": ["mcp-server"]}}}`였다 — PATH 이름 `moai`로 기동하는 등록. 제거됐으므로 **새 세션이 PATH 이름으로 MCP를 띄우는 소비자는 더 없다.**
- 남은 것은 제거 이전에 뜬 프로세스다. 13:39Z 실측: `moai mcp-server`(PATH 이름) **14개**, `moai.old mcp-server` **4개**(12:59Z에는 22+11). 해당 세션이 끝나면 함께 사라진다. 강제 종료는 이 설계가 하지 않는다(활성 세션 보존, §4).

### 1.2 Go 훅 포워더 — settings 배선

`.claude/settings.json`의 `hooks/moai/` 참조 36곳, 20 이벤트에 **30개 스크립트** 배선. 차단 가능 이벤트에 걸린 13개: handle-pre-tool(PreToolUse)·handle-user-prompt-submit·handle-stop·handle-stop-goal·handle-security-turn·handle-security-commit·sync-phase-quality-gate·handle-multi-review-gate·handle-codex-review-gate·handle-codex-review-entry·chain-event·handle-subagent-stop·handle-stop-failure. 호출 형태는 전부 `moai hook <sub>`(exec 또는 `printf | moai hook …`)이며, 스크립트 내부에 PATH 폴백 사슬이 있다(handle-pre-tool.sh 70–86행: `command -v moai` → `$HOME/go/bin/moai` → `~/.local/bin/moai` → exit 0).

### 1.3 statusline

`.claude/settings.json` `statusLine.command` = `"$CLAUDE_PROJECT_DIR/.moai/status_line.sh"`(refreshInterval 60). 스크립트의 PATH 의존 행: 22–23(`command -v moai` + `exec moai statusline`), 29–30(절대 `~/go/bin/moai` = **shim**), 34–35(같은 파일 $HOME 확장), 39–40(`~/.local/bin/moai`, 부재). 9–11행 `~/.moai/.env.glm` source(stdin 닫음 — 값은 다루지 않는다).

### 1.4 그 외

- 사용자 전역 `~/.claude/settings.json` statusLine: Orca wrapper — `moai` 문자열 없음(영향 없음).
- MCP 등록: 프로젝트·사용자 settings에 `mcpServers` 키 없음, 저장소 루트 `.mcp.json` 없음. 유일한 moai MCP 배선이었던 §1.1의 플러그인은 13:15Z에 제거됐다(v1의 "MCP 0건"은 제거 전 이 배선을 놓쳤다 — 정정).
- 세션 런처: shim의 launcher 단어(cc·glm·codex·omp·-l·-a)는 Rust로 간다. `moai-cli` 자체는 `~/.moai-cli/path`가 PATH에 없어 **이름으로는 해석되지 않는다**(절대 경로 또는 shim 경유).

## 2. 영향 분석 — "moai 제거"의 두 층

| 항목 | (i) PATH 이름만 제거(shim 파일 유지) | (ii) shim 파일 제거 |
| --- | --- | --- |
| 상태줄(현행) | `command -v moai` 실패 → 29행 절대 폴백=shim exec → **Go 렌더 유지** | 폴백 전부 실패 → exit 0 → **상태줄 소멸** |
| 훅 44파일(가드) | PATH 호출 실패 → `$HOME/go/bin/moai`(shim) 절대 폴백 → moai.old 훅 실행 → **가드 유지** | 폴백 실패 → exit 0(침묵) 또는 exec 실패 → **13개 차단 가드 포함 전부 소멸** |
| 플러그인 MCP(§1.1) | 플러그인 제거됨(13:15Z) — 새 기동 소비자 없음. 남은 18프로세스는 해당 세션 종료 때 사라짐 | 같음 |
| launcher(cc·glm·-l·-a) | shim 경유 Rust — 유지 | `~/.moai-cli/path/moai-cli` 절대 호출만 가능 |

**(i)층은 이 기계에서 shim 파일을 둔 채로는 만들 수 없다.** `~/go/bin`이 PATH에 있으므로(`~/.zshenv` 33행·`~/.zshrc` 153행) `~/go/bin/moai` 파일이 있는 한 PATH 이름 `moai`도 해석된다. shim을 `~/.local/bin/moai`로 옮겨도 `~/.local/bin`이 PATH 맨 앞에 있어(`~/.zshenv` 20행) 마찬가지다. PATH 밖의 프로젝트 로컬 폴백(`$CLAUDE_PROJECT_DIR/.moai/bin/moai`·`$CLAUDE_PROJECT_DIR/bin/moai`)은 차단 가드 13개 중 **4개**에만 있고 9개에는 없다(13:5xZ 실측). 따라서 "PATH 이름 `moai` 제거"는 현행 Go 훅에 대해 실질적으로 (ii)층이다.

### 2.1 F1 — 사용자 결정: Go 훅만 Go 바이너리 절대 경로로 유지

**결정(사용자, Issue #2383 댓글 6769057244, 리더 기록 2026-11-10T13:49Z):** "Rust hook 공통 엔진(#2391)이 준비될 때까지 Go hook만 절대 경로로 유지한다." PATH의 Go moai 제거(PR #2374 병합·재설치 뒤)를 하더라도 Go hook 스크립트는 Go 바이너리의 절대 경로를 부르도록 전환 절차에 넣어, 폴백 `exit 0`으로 가드 13개가 사라지는 일을 막는다. **Go hook 제거는 #2391 완료와 격리 HOME 검증 뒤 별도 확인을 거친다.**

결정 이전의 충돌(기록): "PATH의 Go moai는 PR #2374 병합·재설치 뒤 제거"와 "Go 훅 스크립트 무변경"을 함께 지키면, 무변경 Go 훅의 폴백 사슬이 `exit 0`에 이르러 차단 가드 13개가 **조용히** 꺼진다(handle-pre-tool.sh 70–89행: `command -v moai` → `[ -f $HOME/go/bin/moai ]` → `[ -f ~/.local/bin/moai ]` → `exit 0`). Rust에는 아직 이 13개를 대신할 훅 진입점이 없다(§3).

### 2.2 Go 훅 절대 경로 전환 방식 — 스크립트 맨 위 한 줄 + PATH 밖 심링크

**방식.** PATH 밖의 새 디렉터리에 Go 바이너리를 가리키는 심링크 하나를 만들고, `moai`를 언급하는 Go 훅 44개 각각의 shebang 바로 다음 줄에 그 디렉터리를 스크립트 안에서만 PATH 앞에 붙이는 한 줄을 넣는다.

```sh
# symlink to create (absent today; ~/go itself is not on PATH)
$HOME/go/hook-bin/moai -> $HOME/go/bin/moai.old
# the one line inserted as line 2 of each of the 44 hooks
PATH="$HOME/go/hook-bin:$PATH"  # #2383: Go hooks call the Go binary by absolute path until #2391
```

**이 방식을 고른 이유(실측).**
- 44개 훅의 `moai` 호출은 182행이고 형태가 여러 가지다. `command -v moai` 분기, `$HOME/go/bin/moai` 절대 폴백, `MOAI_PATH_BIN="$(command -v moai)"` 뒤 `[ -f … ] && [ -x … ]` 검사(handle-codex-review-gate.sh 69–71행), 프로젝트 로컬 폴백이 섞여 있다.
- v1처럼 줄마다 치환하면 일부 형태를 놓친다(재감사 B2: 182행 중 69행만 치환). 한 줄 PATH 접두는 형태와 무관하게 이름 `moai`를 모두 같은 절대 경로로 풀어 준다.
- 셸 함수(`moai() { … }`)는 `command -v`가 실제 파일 경로를 돌려주지 않아 `-f`/`-x` 검사에서 떨어지므로 기각했다.
- 심링크 대상 `moai.old`는 지금 shim이 Go 전용 단어(hook 포함)를 넘기는 바로 그 실행체다(`~/go/bin/moai` 12행 `GO="${MOAI_GO_BIN:-$HOME/go/bin/moai.old}"`, 21행 `exec "$GO" "$@"`). 따라서 호출 결과는 지금과 같다.
- 이 한 줄은 훅 프로세스 안의 PATH만 바꾼다. 사용자 셸의 PATH에는 `moai`가 다시 나타나지 않는다.
- 주의: `sync-phase-quality-gate.sh`는 시험에서 `source`로 읽히기도 하므로, 그때는 읽는 셸의 PATH에도 접두가 남는다. 영향은 그 셸 안의 이름 `moai`가 Go를 가리키는 것뿐이다.

**메커니즘 시험(2026-11-10T13:5xZ, 세션 scratchpad, 실제 파일 무변경).** 시험 조건은 임시 HOME, `PATH=/usr/bin:/bin`(PATH에 `moai` 없음)이다. 실제 Go 대신 호출을 기록하고 exit 2로 끝나는 가짜 `moai.old`를 썼다.

| 훅 | 원본(한 줄 없음) | 한 줄 넣은 사본 |
| --- | --- | --- |
| handle-pre-tool | exit 0, Go 호출 없음 — **가드 꺼짐** | exit 2, `…/go/hook-bin/moai hook pre-tool` 호출 |
| handle-stop | exit 0, Go 호출 없음 | exit 2, `…/go/hook-bin/moai hook stop` 호출 |
| chain-event | exit 0, Go 호출 없음 | exit 0(설계상 fail-open), `…/go/hook-bin/moai hook chain-event` 호출 |

이 시험은 경로 해석과 종료 코드 전달만 보인다. 실제 Go 판정(어떤 입력을 거부하는지)은 §5-2 격리 선검증에서 진짜 `moai.old`로 확인한다.

**수명.** 이 한 줄과 심링크는 #2391(hook 공통 엔진)이 준비될 때까지의 다리다. Go 훅을 지우는 일은 이 설계의 단계가 아니다. #2391 완료 → 격리 HOME 검증 → **별도 확인**을 거쳐 그때의 Issue에서 한다(§5-9).

모든 단계의 readback은 "훅 통과"가 아니라 **거부 fixture가 실제로 거부되는지**(exit 2 또는 deny JSON)를 관찰한다. `exit 0` 통과는 가드가 꺼져도 똑같이 나오기 때문이다.

## 3. 훅 전환 — Rust 책임 4개 기준 대체/폐기 표 (B2 정정: Go 스크립트 복사·sed 없음)

전제 실측(v3 정정, 보충 6768932473): Rust 빌드에는 Go처럼 `hook <sub>` 형태의 하위 명령이 없다. 다만 훅 진입점은 이미 몇 개 있다 — `claim-verify --hook`(host PreToolUse 입력, 거부 시 exit 2; `src/main.rs` 233–237행), `hook-receipt --host`(관찰 전용, 훅을 실패시키지 않음; 405–409행), `hooks-plan`·`hooks-apply`(훅 집합 미리보기·적용; 388–396행). 이것들이 아래 표의 "이관 완료" 판정 (a)의 기준선이며, 4책임 공통 엔진은 #2391에서 만든다. 따라서 **지금 시점의 전환은 없으며**, 아래 표는 Rust 훅 진입점이 구현된 이후의 대체/폐기 계획표다. 그 이전 단계에서 훅은 전부 현행(PATH shim 경유 Go)으로 **무변경 유지**가 기본이고, 이것이 v1 (가)의 "복사·sed"를 대체한다.

| Rust 책임 | 커버하는 Go 훅(스크립트) | 전환 시 대체 | 버려지는 것(의도된 폐기) |
| --- | --- | --- | --- |
| **session-check** | handle-session-start(-navigator)·handle-session-end·session-start-compact | 세션 시작·종료 기록과 게이트를 Rust가 판정 | Go의 세션 기록 포맷 의존(이관 후 폐기) |
| **tool-preflight** | handle-pre-tool(PreToolUse 차단)·handle-user-prompt-submit·handle-codex-review-entry(UserPromptSubmit 배선 — v3에서 tool-observe에서 옮김)·chain-event·handle-subagent-stop | **exit-2/JSON 거부 계약을 Rust가 이관** — 이관 완료 전까지 Go 훅 제거 금지 | Go 폴백 사슬(command -v/절대 경로 3단 — Rust는 자체 바이너리 경로를 쓰므로 불필요) |
| **tool-observe** | handle-post-tool·handle-post-tool-failure·handle-harness-observe | 관찰 전용(비차단) 기록 이관 | 기록 포맷의 Go 호환 필드 |
| **turn-validation** | handle-stop·handle-stop-goal·handle-security-turn·handle-security-commit·sync-phase-quality-gate·handle-multi-review-gate·handle-codex-review-gate·handle-task-completed·handle-teammate-idle | 턴 완료 검증 게이트 이관 | 동일 |
| **해당 없음(책임 밖)** | handle-compact·handle-post-compact·handle-config-change·handle-cwd-changed·handle-file-changed·handle-instructions-loaded·handle-permission-denied·handle-permission-request·status-transition-ownership·handle-elicitation* ·trace-ledger·verify-sync-backup(moai 언급 없음 2개 포함) | 대체 없음 — Go 유지 또는 별도 결정 | 없음 |

표의 운용 규칙: 각 행의 "이관 완료"는 (a) Rust 진입점 존재, (b) 동일 stdin 계약 시험, (c) 기존 fixture 시험 녹색 — 세 조건으로 판정하고, 미충족 행은 Go 유지. 이 규칙만으로는 가드가 지켜지지 않는다. Go 훅이 기대는 shim을 지우면 가드는 조용히 꺼진다 — v2의 "원천 차단" 문구는 틀렸다(재감사 F1). 가드를 지키는 것은 사용자 결정(§2.1)에 따른 절대 경로 전환(§2.2)이며, shim 제거(§5-7)는 그 전환의 readback 뒤에만 한다.

## 4. 영향받는 사용자 설정과 활성 세션 보존 (B4 정정)

- 플러그인 MCP는 사용자가 13:15Z에 제거했다(§1.1). 남은 moai MCP 프로세스(13:39Z 14+4)는 이미 떠 있으므로 PATH 이름이 없어져도 계속 응답하고, 해당 세션이 끝날 때 사라진다. 이 설계는 그 프로세스를 종료하지 않는다.
- 상태줄은 매 렌더마다 새로 실행되므로, §5-5 교체 직후 다음 렌더(refreshInterval 60초 이내)부터 새 스크립트가 쓰인다. 세션 재시작은 필요 없다.
- 훅은 매 이벤트마다 새 프로세스로 실행되므로, §5-6의 한 줄 전환은 다음 이벤트부터 적용되고 세션 재시작이 필요 없다. 전환 전후 모두 같은 Go 실행체(`moai.old`)가 판정하므로 실행 중인 세션의 동작은 바뀌지 않는다.
- 본 PR·본 설계는 그 어떤 파일도 쓰지 않았다(§0 원장의 해시가 적용 전후 동일함을 확인하는 용도).

### 4.1 대체 상태줄 스크립트 (F3 정정 — §5-2·§5-5가 쓰는 실제 내용)

대상은 `$PRI/.moai/status_line.sh`의 내용만이다. `.claude/settings.json`의 `statusLine.command`(같은 경로)는 바꾸지 않는다. 사용자 결정 "상태줄은 Go 화면 구성 계승"은 PR #2374(`moai-cli statusline`, Go 동등 렌더 — golden `--check`)가 맡고, 이 스크립트는 그것을 PATH 이름 없이 부르기만 한다.

```bash
#!/bin/bash
# MoAI statusline wrapper (#2383) - renders with the Rust moai-cli by absolute path.
# No PATH lookup of `moai`: the Go `moai` name leaves PATH after PR #2374 (#2383).
# stdin is handed straight to the exec'd process; this script creates no temporary
# file (an exec'd process never runs an EXIT trap, so a temp buffer would leak).

# A regular file that is executable: `-x` alone is true for a directory, which exec
# cannot run (exit 126, no fallback) - #2503.
MOAI_CLI="$HOME/.moai-cli/path/moai-cli"
if [ -f "$MOAI_CLI" ] && [ -x "$MOAI_CLI" ]; then
	exec "$MOAI_CLI" statusline
fi

# Fallback while the Go build is still installed: the same Go render as today.
MOAI_GO="$HOME/go/bin/moai.old"
if [ -f "$MOAI_GO" ] && [ -x "$MOAI_GO" ]; then
	exec "$MOAI_GO" statusline
fi

# Nothing installed - exit silently (Claude Code handles a missing statusline).
exit 0
```

저장소 원본: `tools/claude-statusline.sh`(mode 755). 위 블록과 바이트 단위로 같아야 하며 `tests/claude_statusline_wrapper_test.py`가 그 일치와 동작(임시 HOME·가짜 실행체로 stdin·argv·cwd·exit 보존, Rust 우선, 폴백, PATH의 `moai`/`moai-cli` 미호출, `.env.glm` 미source)을 확인한다. §5-5의 교체는 이 파일을 쓴다. 실제 설치본·실제 Go로 한 실행은 이 시험의 범위가 아니다.

현행(§1.3)과 다른 점과 그 근거:
- `command -v moai`와 `~/go/bin/moai`·`~/.local/bin/moai` 폴백을 없앴다. PATH 이름에 기대지 않는 것이 이 카드의 목적이다.
- 존재 검사를 `-f`에서 `-x`로 바꿨다. 실행 권한 없는 파일을 exec해 126으로 끝나는 경우를 막는다(재감사 F2와 같은 원인).
- #2503에서 `-f`를 다시 함께 둔다(`[ -f X ] && [ -x X ]`, Rust·Go 두 경로 모두). `-x`만으로는 디렉터리(또는 디렉터리를 가리키는 링크)도 참이라 exec가 126으로 끝나고 폴백이 돌지 않았다(PR #2482 감사 N1). 이제 디렉터리는 없는 것과 같게 다음 단계로 넘어간다.
- Rust 경로는 `~/.moai-cli/path/moai-cli`(§0, 심링크)다. **#2374가 병합·재설치되기 전의 설치본에는 `statusline` op가 없다.** 그래서 §5-5는 #2374 재설치 뒤에만 실행한다. 그 전에 실행하면 Rust가 사용 오류로 끝나고 폴백이 돌지 않는다(exec 뒤라서).
- Go 폴백은 shim을 거치지 않고 `moai.old`를 직접 부른다. shim이 `statusline` 단어를 `moai.old`로 그대로 넘기므로(§2.2) 출력은 같다.
- **현행 9–11행의 `~/.moai/.env.glm` source를 넣지 않았다.** Rust는 `~/.moai`를 읽지 않는다는 결정(#2393)을 따르고, 상태줄 렌더에 자격 증명이 필요하지 않기 때문이다. 이 env가 Go 렌더 결과를 바꾸는지는 아직 확인하지 않았다(not_run). §5-2(b)에서 같은 stdin으로 env 유무를 비교하고, 출력이 다르면 중단해 #2393에 올린다.

### 4.2 #2503 후속 확인 — 보호 장치 오탐과 Go 재작성 위험

**보호 장치 오탐(AC-2).** PR #2482 작업 중 golden 재검 명령 1건이 사례 이름 `b01-git-clean` 속 단어 때문에, 다른 1건이 `$(…)` 때문에 거부됐다(보고 #2383 6810975788, 실행 안 됨). 그 기록에는 거부한 장치의 이름과 거부 문구가 없다. 저장소의 tool-preflight 가드(`src/hook_guards.rs`, #2391)는 명령을 단어가 아니라 명령으로 읽는다: `tests/statusline_guard_false_positive.rs`가 `--case b01-git-clean`이 든 golden `--check` 명령, `$(…)`가 앞에 붙은 같은 명령, `--case` 없는 전체 `--check`를 `bash_guard`·`bash_confirm` 모두 통과(None)로 고정하고, 대조군으로 실제 `git clean -fd`는 여전히 확인 대상임을 고정한다. 그래서 이 가드에서는 재현되지 않는다. 다른 장치(세션 쪽 보호 장치)에서 다시 거부되면:
- 다른 경로로 같은 명령을 다시 보내지 않는다(AGENTS.md "거부된 동작을 다른 도구·경로로 다시 하지 않는다"). 거부 문구·장치 이름·명령을 기록해 리더에게 보고한다.
- 사례를 고를 필요가 없으면 `--case` 없이 전체 `--check`를 쓴다. 명령 줄에 사례 이름이 없으므로 이 오탐이 생기지 않는다.
- `$(…)`는 필요 없다. 경로는 `--bin`에 바로 쓴다.

**전환 뒤 Go `moai update`/`init`이 `status_line.sh`를 다시 쓰는가(AC-3, 보고 #2383 6769648298).** Go 실행체 실행은 규칙상 금지라 실행 시험은 not_run이다. 대신 Go 저장소(`modu-ai/moai-adk` main `4fa2aa9db2`, 읽은 파일은 그 커밋과 같음 — 작업 트리 변경 없음 확인)를 읽었다. 설치된 `moai.old`가 이 커밋으로 빌드됐는지는 확인하지 않았다(미확인).
- `moai update`(일반·clean reinstall): `.moai/status_line.sh`는 병합 대상 집합에 있다(`internal/cli/update_template_sync.go` `mergeableFilePaths`, `internal/cli/update_clean_install.go` 416행). 배포 전에 사용자 파일을 읽어 두고(`update_template_sync.go` 815–821행), 배포 뒤 3-way 병합에서 `.sh`는 키 구조가 없어 base를 만들지 못하므로 사용자 내용을 그대로 다시 쓴다(`internal/cli/update/merge/merge.go` 287–296행, `merge/base.go` 68–74행). 기존 파일에 쓰므로 mode(755)는 유지된다. 즉 소스상 §4.1 내용은 update 뒤에도 남는다. 배포와 병합 사이에는 잠시 Go 원본이 놓이며, 병합 단계가 오류로 끝나면 Go 원본이 남을 수 있다(실행 확인 not_run).
- `moai init`: 이미 초기화된 프로젝트에서는 `--force` 없이 거부한다(`internal/core/project/phase.go` 181–194행). `--force`면 `.moai/` 전체를 백업 디렉터리로 옮긴 뒤 다시 배포한다(`validator.go` 266–287행). `CarryManifestForward`는 `.moai/` 밖 파일만 보호하므로 `.moai/status_line.sh`는 Go 원본으로 바뀌고 §4.1 내용은 백업 쪽에 남는다.
- 인접 위험(이 Issue 범위 밖): `.claude/hooks/moai/*.sh`는 병합 대상 집합에 없다. §2.2의 훅 2행 한 줄이 update 배포로 사라질 수 있는지는 별도 확인이 필요하다.
- 따라서 전환 뒤 Go `moai update`나 `moai init --force`를 실행했다면 §5-5 readback(상태줄 sha256·mode·렌더)과 §5-6 readback(훅 해시·거부 fixture)을 다시 실행한다. Go 명령 실행을 막을지는 사용자 결정이며 이 문서가 정하지 않는다.

## 5. 승인 순서 절차 (각 실제 변경·영구 삭제의 직전 별도 확인)

격리 검증이 실제 변경보다 **먼저** 온다. `$PRI` 기준 실제 경로를 쓴다.

선행 조건: PR #2374 병합과 `moai-cli` 재설치(설치본에 `statusline` op 존재 확인). 그 전에는 1–4단계(읽기·격리·미리보기·백업)까지만 할 수 있다.

1. **원장·인벤토리 기록**(완료 — §0·§1): 소유·mode·sha256·활성 consumer. 실행 당일 같은 명령으로 다시 측정하고, 이 문서의 값과 다르면 차이를 기록한 뒤 진행 여부를 묻는다.
2. **격리 HOME 선검증(실제 변경 전)** — 임시 HOME·임시 프로젝트, `PATH=/usr/bin:/bin`(PATH에 `moai` 없음). 하나라도 실패하면 여기서 중단한다. 실제 파일은 아직 무변경이다.
   - (a) 상태줄: §4.1 스크립트를 임시 프로젝트에 두고 #2374의 `tests/fixtures/statusline/stdin/*.json` 각각에 대해 스크립트 출력과 `moai-cli statusline` 직접 출력이 바이트 동일한지 본다. 그리고 `python3 -I tools/statusline_golden.py --bin "$HOME/.moai-cli/path/moai-cli" --check`가 전부 일치해야 한다(`--argv` 를 주지 않아 하네스 기본값의 고정 시계·렌더 예산을 쓴다, #2756)(Go 화면 구성 계승 — Issue AC-2).
   - (b) Go 폴백: 임시 HOME에 `moai-cli`가 없을 때 스크립트 출력이 `moai.old statusline` 직접 출력과 같은지 본다. 같은 stdin으로 `.env.glm` 유무 출력도 비교한다(§4.1).
   - (c) Go 훅: 44개 훅을 임시 프로젝트에 시험 fixture로 두고 §2.2의 한 줄을 넣는다. 이것은 교체용 사본이 아니라 시험 입력이며, 시험 뒤 임시 디렉터리와 함께 버린다. 임시 HOME의 `go/hook-bin/moai`는 진짜 `moai.old`를 가리킨다.
     - 차단 13개 각각이 원본에서는 Go를 부르지 않고(대조군 — 가드 꺼짐 재현), 한 줄 넣은 쪽에서는 Go를 부르는지 확인한다.
     - **거부 fixture**: 먼저 현행 배선(shim 있음)에서 Go `hook pre-tool`이 거부하는 입력 하나를 찾아 기록한다. 그 입력이 한 줄 넣은 훅 + PATH에 `moai` 없음 조건에서도 똑같이 거부되는지(exit 2 또는 deny JSON) 확인한다. fixture는 아직 고르지 않았다(not_run).
   - (d) 한 줄 넣은 44개 전부 `bash -n` 통과.
3. **미리보기 제시**: 사용자에게 아래를 보인다.
   - `status_line.sh` 전후 unified diff
   - 훅 44개의 diff(각 +1행이고 다른 줄은 무변경임을 `diff` 출력으로)
   - 새로 만들 심링크 하나
   - 7단계에서 지울 파일 하나(`~/go/bin/moai`)
4. **복구 백업(사본 생성 — 원본 불변)**: `B="$PRI/.moai-cli/backups/2383-<UTC>"`에 `cp -p`로 mode를 보존해 복사한다.
   - 대상: `$PRI/.moai/status_line.sh`, `$PRI/.claude/hooks/moai/*.sh`(46개 전부), `~/go/bin/moai`(shim — 설치 원본이 git 미추적이라 이 백업이 되돌리기 정본이다).
   - 배치: `$B/status_line.sh`, `$B/hooks/<이름>.sh`, `$B/moai`.
   - `$B`에서 `shasum -a 256 status_line.sh hooks/*.sh moai > MANIFEST.sha256`, `stat -f '%Lp %N' status_line.sh hooks/*.sh moai > MANIFEST.mode`를 만든다.
   - readback: 백업 사본에서 다시 계산한 해시·mode가 원본과 모두 같아야 한다.
5. **실행 직전 확인 ①** — 상태줄 교체: `$PRI/.moai/status_line.sh` 내용을 §4.1로 바꾼다. 운영자 확인을 받은 직후에만 실행한다.
   - readback: sha256이 3단계 미리보기의 "후" 해시와 같을 것, `stat -f %Lp`가 `755`일 것, `bash -n` 통과, fixture 하나의 렌더가 2(a)와 같을 것.
6. **실행 직전 확인 ②** — Go 훅 절대 경로 전환(§2.2). shim은 아직 그대로 둔다.
   - 실행: `mkdir "$HOME/go/hook-bin" && ln -s "$HOME/go/bin/moai.old" "$HOME/go/hook-bin/moai"`. 이어서 `moai`를 언급하는 44개 훅의 2행에 한 줄을 넣는다(2개는 언급이 없어 제외: `trace-ledger.sh`·`verify-sync-backup.sh`).
   - readback: 44개 파일 해시가 미리보기 "후" 해시와 같을 것, mode가 MANIFEST와 같을 것, `bash -n` 통과, `readlink`가 `moai.old`를 가리킬 것, 2(c) 거부 fixture가 실제 배선에서 거부될 것.
7. **실행 직전 확인 ③** — PATH의 Go moai 제거: `rm "$HOME/go/bin/moai"`(**영구 삭제** — 백업은 4단계).
   - 조건: PR #2374 병합·재설치, 5·6단계 readback GREEN.
   - 영향: launcher 단어(cc·glm·codex·-l·-a)를 이름 `moai`로 부르는 길도 함께 사라진다(§2 표). 이후에는 `moai-cli` 절대 경로로 부른다.
   - readback(새 로그인 셸): `command -v moai` 실패, 상태줄 fixture 렌더가 2(a)와 같을 것, 거부 fixture가 거부될 것(가드 13개 유지 — F1 결정의 확인).
8. **기록**: decisions log에 각 단계의 승인 시각·readback 결과·되돌리기 명령을 함께 적는다.
9. **이 설계 밖 — Go 훅 제거는 별도**: §2.2의 한 줄과 심링크는 #2391(hook 공통 엔진) 준비 전까지 남는다. Go 훅 제거는 #2391 완료 → 격리 HOME 검증(Rust 엔진이 같은 거부 fixture를 거부하는지) → **별도 확인**의 순서로, 그때의 Issue에서 한다. 이 PR과 위 1–8단계는 Go 훅을 지우지 않는다.

되돌리기(§6)는 모든 단계에서 유효하며 단계별로 즉시 실행 가능해야 한다.

## 6. 되돌리기 — 진짜 역연산 + 해시 검증 (B3 정정: 실제 경로·정본 원본)

모든 복원은 4단계 백업 `$B`에서 `cp -p`로 하고, 해시와 **mode**를 함께 확인한다. v2의 R2는 mode `644`인 설치 원본을 새 파일로 복사해 실행 권한 없는 shim을 만들었다. 재감사 F2에서 재현했다: 결과 `-rw-r--r--`, `command -v moai` exit 1, 훅의 `[ -f ]` 폴백은 참이라 실행 불가 파일을 exec. 아래는 그 정정이다. 단계 역순(R3 → R2 → R1)으로 필요한 것만 실행한다.

```sh
PRI=/Users/goos/MoAI/moai-adk
B="$PRI/.moai-cli/backups/2383-<UTC>"     # step 4 backup directory

# R1. restore the statusline script (undo step 5)
cp -p "$B/status_line.sh" "$PRI/.moai/status_line.sh"
shasum -a 256 "$PRI/.moai/status_line.sh"        # expect 2f5eb54cfb1094dd...
stat -f '%Lp' "$PRI/.moai/status_line.sh"        # expect 755

# R2. undo the Go hook absolute-path switch (undo step 6): hooks first, then the symlink
cp -p "$B"/hooks/*.sh "$PRI/.claude/hooks/moai/"
(cd "$B" && shasum -a 256 -c MANIFEST.sha256)    # the backup itself is intact
for f in "$B"/hooks/*.sh; do cmp "$f" "$PRI/.claude/hooks/moai/${f##*/}" || echo "MISMATCH $f"; done   # expect no output
(cd "$PRI/.claude/hooks/moai" && stat -f '%Lp hooks/%N' *.sh) | diff - <(grep ' hooks/' "$B/MANIFEST.mode")   # expect no output
[ "$(readlink "$HOME/go/hook-bin/moai")" = "$HOME/go/bin/moai.old" ] \
  && rm "$HOME/go/hook-bin/moai" && rmdir "$HOME/go/hook-bin"   # removes only what step 6 created

# R3. restore the Go moai shim on PATH (undo step 7): from the backup, with the exec bit
cp -p "$B/moai" "$HOME/go/bin/moai"
chmod 0755 "$HOME/go/bin/moai"                   # executable whatever mode the backup carried
stat -f '%Lp' "$HOME/go/bin/moai"                # expect 755
shasum -a 256 "$HOME/go/bin/moai"                # expect bbedd136ab623dd3...
zsh -lc 'command -v moai'                        # expect $HOME/go/bin/moai again
```

- `MANIFEST.sha256`·`MANIFEST.mode`는 4단계에서 `$B` 기준 상대 경로로 만든다(훅은 `hooks/` 접두).
- 백업을 잃었을 때 R3의 보조 원본은 `$PRI/.moai-cli/install/moai-shim.sh`다. 이 경우 `install -m 0755 "$PRI/.moai-cli/install/moai-shim.sh" "$HOME/go/bin/moai"`로 mode를 지정해 설치하고 위와 같이 해시·mode를 확인한다. 이 파일은 git 미추적이므로(§0) 1차 정본이 아니다.
- v1의 `mv moai.old moai`는 계속 쓰지 않는다. moai.old를 없애는 것은 역연산이 아니며, §2.2의 심링크와 §4.1의 Go 폴백이 `moai.old`를 절대 경로로 부른다. moai.old는 항상 그 자리에 남는다.
- R2의 `rm`·`rmdir`은 6단계가 새로 만든 심링크와 빈 디렉터리만 지운다. `readlink` 확인이 실패하면 지우지 않고 멈춘다.

## 7. 실측 방법 (재현 가능 — BSD grep 괄호 문제 회피)

v1의 `grep -rhE '([;&|…])'` 명령은 macOS BSD grep에서 `brackets not balanced`로 실패했다(정정). 측정 표준은 Python `re`로 통일한다:

```sh
python3 - <<'EOF'
import re, pathlib, collections
d = pathlib.Path('/Users/goos/MoAI/moai-adk/.claude/hooks/moai')
pat = re.compile(r'(^|[;&|"`\s(])(command -v\s+)?moai(?=[\s;&|)"\']|$)')
total = 0
for f in sorted(d.glob('*.sh')):
    n = sum(len(pat.findall(l)) for l in f.read_text().splitlines())
    total += n
print('moai call lines:', total)
EOF
```

이 방식으로 182행(2026-11-10 실측 — t62의 Python 측정과 일치). 배선·차단 목록은 settings.json 파싱(python3 + json)으로, 해시·mode는 §0 표의 명령으로 재현한다.

## 8. 이 문서가 바꾸지 않는 것

`~/.claude/settings.json`·`$PRI/.claude/settings.json`·`$PRI/.moai/status_line.sh`·`$PRI/.claude/hooks/moai/*`·`~/go/bin/*`·`~/go/hook-bin`·`~/.moai-cli/*`·PATH — 본 PR은 `docs/`의 이 문서 하나만 추가한다(§0 원장 해시가 적용 전후 동일한지 확인하는 용도로 쓰인다). v3의 메커니즘 시험과 되돌리기 리허설은 세션 scratchpad의 가짜 HOME·가짜 프로젝트에서만 했다.

## 9. v1 대비 정정 일람 (감사 6767987735 대응)

B1 → §1.1(플러그인 MCP)·§2 표·§4. B2 → §3(복사·줄 단위 sed 폐기, Rust 책임 4개 표). B3 → §0 원장·§6(실제 경로, moai.old 보존). B4 → §5 재배열(인벤토리 맨 먼저·격리 선검증·단계별 직전 확인·활성 세션 보존). 비차단 → §0 정산 정정(46 전부 .sh)·§7 측정 표준 교체. t62가 옳다고 확인한 항목(.sh 46·44 언급·182행·배선 30/13·shim 해시·status_line 행)은 전부 유지한다.

## 10. v2 대비 정정 일람 (재감사 6768843277·보충 6768932473 대응, v3)

| 지적 | 정정 | 위치 |
| --- | --- | --- |
| F1 shim 제거 시 Go 훅 폴백 `exit 0` → 가드 13개 소멸 | 사용자 결정(6769057244) 반영: Go 훅만 Go 바이너리 절대 경로로 유지. 스크립트 2행 한 줄 + PATH 밖 심링크. 메커니즘 시험(가짜 Go): 원본 exit 0·호출 없음 → 한 줄 넣은 쪽 Go 호출·exit 2 전달. shim 제거(7단계)는 전환 readback 뒤에만. Go 훅 제거는 #2391 완료 → 격리 검증 → 별도 확인 | §2.1·§2.2·§3·§5-6·§5-7·§5-9 |
| F2 R2 `cp`가 mode 644 shim 생성 | 모든 복원을 4단계 `cp -p` 백업에서, R3에 `chmod 0755`와 `stat` mode 확인. 설치 원본은 git 미추적이라 보조 원본으로만(`install -m 0755`). 4단계 MANIFEST(해시·mode). scratchpad 리허설: 백업 48항목 → 변경 → R3(mode 755·해시·`command -v` 성공)·R2(46 일치·mode 일치·심링크 제거)·R1(해시·mode 755) 통과 | §0·§5-4·§6 |
| F3 "§4 대체 상태줄 스크립트" 부재 | 실제 스크립트 본문과 현행 대비 차이·근거를 §4.1에 넣음. 검증은 #2374 golden `--check`와 fixture별 바이트 동일(5-2(a)) | §4.1·§5-2·§5-5 |
| 비차단: 플러그인 "enabled" 서술 | 13:15Z 사용자 제거 반영, 13:39Z 잔존 프로세스 14+4 | §1.1·§1.4·§2·§4 |
| 비차단: codex-review-entry 분류 | UserPromptSubmit 배선 실측 → tool-preflight 행으로 이동 | §3 |
| 비차단: "§실측" 가리키는 절 없음 | 해당 문구 삭제 | §9 |
| 보충: "Rust hook 하위 명령 grep 0건" | `claim-verify --hook`·`hook-receipt`·`hooks-plan`/`hooks-apply` 진입점을 기준선으로 명시, 공통 엔진은 #2391 | §3 |

v3 추가 실측(2026-11-10T13:5xZ, 읽기 전용): `~/go/hook-bin` 부재, `~/go`는 PATH에 없음, `~/.local/bin`·`~/go/bin`은 PATH에 있음, 차단 훅 13개 중 프로젝트 로컬 폴백이 있는 것은 4개, shim은 Go 단어를 `$HOME/go/bin/moai.old`로 exec.

🗿 MoAI
