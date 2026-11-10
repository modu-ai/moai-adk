# #2383 설계(v2, 수리판) — Claude Code 상태줄을 moai-cli로 전환(경로 이름 `moai` 제거 전제)

상태: 설계(design-only) v2 — v1(감사 FAIL, agent-3-t62 코멘트 6767987735, 차단 B1–B4)의 전면 재작성판. 이 문서는 아무것도 바꾸지 않는다. 모든 실측은 2026-11-10 읽기 전용이며, 이 문서가 다루는 "실행"은 전부 후속 구현 단계의 것으로 본 PR은 실행하지 않는다(§5 각 단계의 직전 확인 포함).
요청: 사용자 지시 2026-11-10T11:50Z. 선행: PR #2374(#2350 Rust statusline) 병합·재설치 — #2374 자체는 수정하지 않는다.
v1 대비 정정: 실제 경로를 프로젝트 기준으로 바로잡았고(§0), Go 플러그인 MCP 소비자를 인벤토리에 넣었으며(§1.4), 훅 전환을 "Go 스크립트 복사·sed" 대신 **Rust 책임 4개 기준 대체/폐기 표**(§3)로 다시 설계했고, 되돌리기를 진짜 역연산+해시 검증으로 고쳤으며(§6), 절차를 승인 순서(인벤토리 → 격리 검증 → 미리보기·백업 → 단계별 직전 확인)로 재배열했다(§5).

## 0. 대상 파일 원장 — 소유·mode·해시 (승인 순서 1단계: 맨 먼저)

모든 경로는 이 기계 실측(2026-11-10). `$PRI` = `/Users/goos/MoAI/moai-adk`.

| 파일 | 소유 | mode | sha256(앞 16) | 역할 |
| --- | --- | --- | --- | --- |
| `$PRI/.moai/status_line.sh` | goos | `-rwxr-xr-x@` | `2f5eb54cfb1094dd` | 상태줄 wrapper — settings.json `statusLine.command`(같은 파일 10행)이 이것을 실행 |
| `~/go/bin/moai` | goos | `-rwxr-xr-x` | `bbedd136ab623dd3` (1510B) | 라우팅 shim — **PATH에 있음**(`~/go/bin`). launcher 단어→Rust, Go 전용 64단어(hook·statusline·mcp-server·session 포함)→`~/go/bin/moai.old` |
| `~/.moai-cli/path/moai` | goos | `-rwxr-xr-x` | `bbedd136ab623dd3` (1510B) | 위 shim과 바이트 동일 사본 — 디렉터리가 PATH에 없음 |
| `$PRI/.moai-cli/install/moai-shim.sh` | goos | `-rw-r--r--@` | `bbedd136ab623dd3` (1510B) | shim **설치 원본** — live shim 2곳과 `cmp` 동일 실측. 되돌리기의 정본(§6) |
| `~/go/bin/moai.old` | goos | `-rwxr-xr-x` | — (78,709,760바이트 Go 바이너리) | Go 빌드 — 훅·statusline·mcp-server·session의 현재 실행체 |
| `~/.moai-cli/path/moai-cli` | goos | 심링크 | → `versions/2026.11.0-rc.1+f259f15a1/bin/moai-cli` | Rust 빌드(PR #2374의 statusline은 미포함 — 미병합) |
| `$PRI/.claude/hooks/moai/` (디렉터리) | goos | `drwxr-xr-x` | — | Go 훅 포워더 `.sh` **46개**(전부 .sh — 디렉터리에 비-.sh 항목 없음) |

참고: v1의 "디렉터리 49항목 중 비-.sh 3"은 셸 별칭(`ls`→`ls -l`)이 만든 착시였다. 정정: `ls -A` 항목 46 = `.sh` 46, 비-.sh 없음(이전 감사 t62 관측과 일치).

## 1. 활성 consumer 인벤토리 (승인 순서 1단계 — "무엇이 지금 moai를 쓰는가")

### 1.1 Go 플러그인 MCP 서버 — **v1에서 누락된 살아 있는 소비자 (B1 정정)**

- `~/.claude/settings.json` `enabledPlugins`: `"moai@moai-adk": true`.
- 설치본 `~/.claude/plugins/cache/moai-adk/moai/3.1.3/.mcp.json`: `{"mcpServers": {"moai": {"command": "moai", "args": ["mcp-server"]}}}` — **PATH 이름 `moai`로 기동하는 절대 경로 아닌 등록**.
- 실행 실측(2026-11-10 재측정): `moai mcp-server`(PATH 이름) 프로세스 **22개**, `/Users/goos/go/bin/moai.old mcp-server` **11개** — 합계 33개가 살아 있다(이전 감사 t62의 12:59Z 측정 22+11과 동일 수준).

### 1.2 Go 훅 포워더 — settings 배선

`.claude/settings.json`의 `hooks/moai/` 참조 36곳, 20 이벤트에 **30개 스크립트** 배선. 차단 가능 이벤트에 걸린 13개: handle-pre-tool(PreToolUse)·handle-user-prompt-submit·handle-stop·handle-stop-goal·handle-security-turn·handle-security-commit·sync-phase-quality-gate·handle-multi-review-gate·handle-codex-review-gate·handle-codex-review-entry·chain-event·handle-subagent-stop·handle-stop-failure. 호출 형태는 전부 `moai hook <sub>`(exec 또는 `printf | moai hook …`)이며, 스크립트 내부에 PATH 폴백 사슬이 있다(handle-pre-tool.sh 70–86행: `command -v moai` → `$HOME/go/bin/moai` → `~/.local/bin/moai` → exit 0).

### 1.3 statusline

`.claude/settings.json` `statusLine.command` = `"$CLAUDE_PROJECT_DIR/.moai/status_line.sh"`(refreshInterval 60). 스크립트의 PATH 의존 행: 22–23(`command -v moai` + `exec moai statusline`), 29–30(절대 `~/go/bin/moai` = **shim**), 34–35(같은 파일 $HOME 확장), 39–40(`~/.local/bin/moai`, 부재). 9–11행 `~/.moai/.env.glm` source(stdin 닫음 — 값은 다루지 않는다).

### 1.4 그 외

- 사용자 전역 `~/.claude/settings.json` statusLine: Orca wrapper — `moai` 문자열 없음(영향 없음).
- MCP 등록: 프로젝트·사용자 settings에 `mcpServers` 키 없음, 저장소 루트 `.mcp.json` 없음. 유일한 moai MCP 배선이 §1.1의 플러그인이다(v1의 "MCP 0건"은 이 배선을 놓쳤다 — 정정).
- 세션 런처: shim의 launcher 단어(cc·glm·codex·omp·-l·-a)는 Rust로 간다. `moai-cli` 자체는 `~/.moai-cli/path`가 PATH에 없어 **이름으로는 해석되지 않는다**(절대 경로 또는 shim 경유).

## 2. 영향 분석 — "moai 제거"의 두 층

| 항목 | (i) PATH 이름만 제거(shim 파일 유지) | (ii) shim 파일 제거 |
| --- | --- | --- |
| 상태줄(현행) | `command -v moai` 실패 → 29행 절대 폴백=shim exec → **Go 렌더 유지** | 폴백 전부 실패 → exit 0 → **상태줄 소멸** |
| 훅 44파일(가드) | PATH 호출 실패 → `$HOME/go/bin/moai`(shim) 절대 폴백 → moai.old 훅 실행 → **가드 유지** | 폴백 실패 → exit 0(침묵) 또는 exec 실패 → **13개 차단 가드 포함 전부 소멸** |
| 플러그인 MCP(§1.1) | 기존 33프로세스는 생존, **신규 기동은 실패**(PATH 이름 의존) — 세션 재시작 시마다 증가 | 신규 기동 실패는 동일 |
| launcher(cc·glm·-l·-a) | shim 경유 Rust — 유지 | `~/.moai-cli/path/moai-cli` 절대 호출만 가능 |

## 3. 훅 전환 — Rust 책임 4개 기준 대체/폐기 표 (B2 정정: Go 스크립트 복사·sed 없음)

전제 실측: Rust 빌드에 `hook` 하위 명령이 없다(`src/main.rs` grep 0건 — 2026-11-10). 따라서 **지금 시점의 전환은 없으며**, 아래 표는 Rust 훅 진입점이 구현된 이후의 대체/폐기 계획표다. 그 이전 단계에서 훅은 전부 현행(PATH shim 경유 Go)으로 **무변경 유지**가 기본이고, 이것이 v1 (가)의 "복사·sed"를 대체한다.

| Rust 책임 | 커버하는 Go 훅(스크립트) | 전환 시 대체 | 버려지는 것(의도된 폐기) |
| --- | --- | --- | --- |
| **session-check** | handle-session-start(-navigator)·handle-session-end·session-start-compact | 세션 시작·종료 기록과 게이트를 Rust가 판정 | Go의 세션 기록 포맷 의존(이관 후 폐기) |
| **tool-preflight** | handle-pre-tool(PreToolUse 차단)·handle-user-prompt-submit·chain-event·handle-subagent-stop | **exit-2/JSON 거부 계약을 Rust가 이관** — 이관 완료 전까지 Go 훅 제거 금지 | Go 폴백 사슬(command -v/절대 경로 3단 — Rust는 자체 바이너리 경로를 쓰므로 불필요) |
| **tool-observe** | handle-post-tool·handle-post-tool-failure·handle-harness-observe·handle-codex-review-entry | 관찰 전용(비차단) 기록 이관 | 기록 포맷의 Go 호환 필드 |
| **turn-validation** | handle-stop·handle-stop-goal·handle-security-turn·handle-security-commit·sync-phase-quality-gate·handle-multi-review-gate·handle-codex-review-gate·handle-task-completed·handle-teammate-idle | 턴 완료 검증 게이트 이관 | 동일 |
| **해당 없음(책임 밖)** | handle-compact·handle-post-compact·handle-config-change·handle-cwd-changed·handle-file-changed·handle-instructions-loaded·handle-permission-denied·handle-permission-request·status-transition-ownership·handle-elicitation* ·trace-ledger·verify-sync-backup(moai 언급 없음 2개 포함) | 대체 없음 — Go 유지 또는 별도 결정 | 없음 |

표의 운용 규칙: 각 행의 "이관 완료"는 (a) Rust 진입점 존재, (b) 동일 stdin 계약 시험, (c) 기존 fixture 시험 녹색 — 세 조건으로 판정하고, 미충족 행은 Go 유지. **13개 차단 가드가 조용히 꺼지는 상태**(v1 (가)의 결함)는 이 규칙으로 원천 차단된다.

## 4. 영향받는 사용자 설정과 활성 세션 보존 (B4 정정)

- 살아 있는 moai MCP 프로세스 33개(§1.1)는 PATH 이름 제거 시 **신규 기동만 실패**(기존 프로세스는 계속 렌더·응답). 세션을 재시작하지 않는 한 즉시 깨지지 않는다 — 그래도 제거는 §5 최종 단계로 미루고, 제거 전 플러그인 MCP의 처리(비활성화 / 절대 경로化 / 유지)는 **사용자 결정**으로 올린다(§5 단계 6의 직전 확인 항목).
- 현재 세션들이 사용 중인 훅 30스크립트는 §3 규칙으로 무변경 유지된다. 본 PR·본 설계는 그 어떤 파일도 쓰지 않았다(§0 원장의 해시가 적용 전후 동일함을 확인하는 용도).

## 5. 승인 순서 절차 (각 실제 변경·영구 삭제의 직전 별도 확인)

격리 검증이 실제 변경보다 **먼저** 온다. `$PRI` 기준 실제 경로를 쓴다.

1. **원장·인벤토리 기록**(완료 — §0·§1): 소유·mode·sha256·활성 consumer. 변경 후 이 표와 재대조한다.
2. **격리 HOME 선검증(실제 변경 전)**: 임시 HOME·임시 프로젝트에서 (a) §4 대체 statusline 스크립트가 같은 stdin에 같은 출력을 내는지(Issue AC-2 — PATH에서 moai 제거된 셸), (b) moai-cli 부재 시 Go 폴백 렌더, (c) 전환 후 훅 경로가 도구 호출을 막지 않음(AC-3). 실패 시 여기서 중단 — 실제 파일은 아직 무변경.
3. **미리보기 제시**: §6의 before/after diff를 사용자에게 보인다.
4. **복구 백업(사본 생성 — 원본 불변)**: `cp $PRI/.moai/status_line.sh $PRI/.moai/status_line.sh.before-2383` + sha256 기록. (훅은 무변경이므로 백업 불필요 — §3 무변경 유지 원칙.)
5. **실행 직전 확인 ①**: 상태줄 교체(`status_line.sh` 갱신) — 운영자 확인을 받은 직후에만 실행. readback: sha256 재대조 + `bash -n` 통과.
6. **실행 직전 확인 ②**: PATH에서 `moai` 제거(`rm ~/go/bin/moai` — **영구 삭제**) — (a) §4 대체 스크립트 적용 완료, (b) 훅 무변경 유지 확인, (c) 플러그인 MCP 처리 사용자 결정 완료, (d) 격리 검증 GREEN — 네 조건 확인 직후에만. 제거 후 readback: `command -v moai` 실패 + 상태줄 렌더 유지 + 훅 통과(AC-2·AC-3 재관측).
7. **기록**: decisions log에 각 단계의 승인·readback·되돌리기 명령을 함께 기록.

되돌리기(§6)는 모든 단계에서 유효하며 단계별로 즉시 실행 가능해야 한다.

## 6. 되돌리기 — 진짜 역연산 + 해시 검증 (B3 정정: 실제 경로·정본 원본)

```sh
# R1. 상태줄 스크립트 복원 (5단계 되돌리기) — 백업 사본에서 복원 후 해시 검증
cp /Users/goos/MoAI/moai-adk/.moai/status_line.sh.before-2383 \
   /Users/goos/MoAI/moai-adk/.moai/status_line.sh
shasum -a 256 /Users/goos/MoAI/moai-adk/.moai/status_line.sh   # → 2f5eb54cfb10… 일치 확인

# R2. PATH 이름 복원 ((6)단계 되돌리기) — 설치 원본에서 복원 + 해시 검증
cp /Users/goos/MoAI/moai-adk/.moai-cli/install/moai-shim.sh "$HOME/go/bin/moai"
shasum -a 256 "$HOME/go/bin/moai"                              # → bbedd136ab62… 일치 확인
command -v moai                                                # 재등장 관찰
```

- v1의 `mv moai.old moai`는 **삭제한다**: moai.old를 없애는 것이 역연산이 아니며, §3 스크립트·(가) 치환 뒤의 절대 호출을 깨뜨린다(t62 지적 정당). moai.old는 항상 그 자리에 남는다.
- 설치 원본의 정본: `$PRI/.moai-cli/install/moai-shim.sh`(sha `bbedd136ab623dd3…`, live shim 2곳과 `cmp` 동일 실측). v1이 상대했던 `~/.moai-cli/install/`은 존재하지 않는다 — 경로 정정 기록.
- 훅은 §3 원칙상 무변경이므로 훅 되돌리기는 존재하지 않는다(무엇도 바꾸지 않았다).

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

`~/.claude/settings.json`·`$PRI/.claude/settings.json`·`$PRI/.moai/status_line.sh`·`$PRI/.claude/hooks/moai/*`·`~/go/bin/*`·`~/.moai-cli/*`·PATH — 본 PR은 `docs/`의 이 문서 하나만 추가한다(§0 원장 해시가 적용 전후 동일한지 확인하는 용도로 쓰인다).

## 9. v1 대비 정정 일람 (감사 6767987735 대응)

B1 → §1.1(플러그인 MCP·33프로세스)·§2 표·§4·§5-6. B2 → §3(복사·sed 폐기, Rust 책임 4개 표, 13가드 사일런스 원천 차단). B3 → §0 원장·§6(실제 경로, 설치 원본+해시 검증, moai.old 보존). B4 → §5 재배열(인벤토리 맨 먼저·격리 선검증·단계별 직전 확인·활성 세션 보존). 비차단 → §0 정산 정정(46 전부 .sh)·§7 측정 표준 교체·clone 표기는 §실측에 "워크트리 커밋 트리"로 명시. t62가 옳다고 확인한 항목(.sh 46·44 언급·182행·배선 30/13·shim 해시·status_line 행)은 전부 유지한다.

🗿 MoAI
