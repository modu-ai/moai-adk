# #2383 설계 — Claude Code 상태줄을 moai-cli로 전환(경로 이름 `moai` 제거 전제)

상태: 설계(design-only) — 이 문서는 아무것도 바꾸지 않는다. 읽기 전용 실측만 수행했다.
요청: 사용자 지시 2026-11-10T11:50Z ("기존 `moai`를 PATH 제거하고 moai-cli로 claude code statusline은 moai 제거해도 동일하게 statusline이 나오도록").
선행: PR #2374(#2350 Rust statusline) 병합·재설치 — 본 문서는 그 전제 위의 전환 설계다. #2374 자체는 수정하지 않는다.

## 1. 현재 상태 실측 인벤토리 (2026-11-10, 읽기 전용 — 방법은 §9)

### 1.1 `moai` 이름이 존재하는 파일 3종

| 경로 | 실체 | PATH 도달성 |
| --- | --- | --- |
| `~/go/bin/moai` | 라우팅 shim 스크립트(2026-11-10 리더 설치). launcher 단어(cc·glm·codex·omp·-l·-a)→ Rust(`~/.moai-cli/path/moai-cli`), Go 전용 하위명령(hook·statusline·mcp-server·session 등 80단어)→ `~/go/bin/moai.old` | **PATH에 있음**(`~/go/bin`) |
| `~/.moai-cli/path/moai` | 위 shim과 **바이트 동일**한 사본(2026-11-10 15:28 설치) | 디렉터리가 PATH에 없음 |
| `~/.moai-cli/path/moai-cli` | Rust 빌드 심링크(versions/2026.11.0-rc.1+f259f15a1) | 디렉터리가 PATH에 없음 |

shim 헤더에 Undo가 문서화되어 있다: `mv "$HOME/go/bin/moai.old" "$HOME/go/bin/moai"` (및 `~/.moai-cli/path/moai` 제거).

### 1.2 PATH 이름 `moai`에 의존하는 파일·행 (전수)

1. **`.moai/status_line.sh`** — 상태줄 wrapper. PATH 의존 행:
   - 22행 `if command -v moai &> /dev/null; then` / 23행 `exec moai statusline`
   - 29–30행 절대 폴백 `exec "/Users/goos/go/bin/moai" statusline` — **가리키는 것은 moai.old가 아니라 shim 자체**
   - 34–35행 `$HOME/go/bin/moai`(위와 같은 파일), 39–40행 `~/.local/bin/moai`(부재)
   - 9–11행 `~/.moai/.env.glm` source(stdin 닫음 — 값은 출력·복사하지 않음, 기존 행동 보존)
2. **`.claude/hooks/moai/*.sh`** — `.sh` 46개 중 44개가 `moai` 문자열 포함, 명령형 호출 행 182행. 호출 형태: `moai hook <sub>`(exec 또는 `printf | moai hook …`) 전부 — Go 훅 포워더. 폴백 사슬(handle-pre-tool.sh 70–86행이 대표): PATH `moai` → `$HOME/go/bin/moai`(shim) → `~/.local/bin/moai` → exit 0(침묵).
3. **`.claude/settings.json`** — `hooks/moai/` 참조 36곳, 20 이벤트에 **30개 스크립트 배선**. 차단 가능 이벤트(PreToolUse·UserPromptSubmit·Stop·SubagentStop·StopFailure)에 걸린 스크립트 **13개**(handle-pre-tool·handle-stop·handle-stop-goal·handle-security-turn·handle-security-commit·sync-phase-quality-gate·handle-multi-review-gate·handle-codex-review-gate·handle-codex-review-entry·handle-user-prompt-submit·chain-event·handle-subagent-stop·handle-stop-failure). statusLine 자체는 10행 `.moai/status_line.sh` 참조.
4. **MCP 서버** — 등록 무: 프로젝트 `.claude/settings.json`·`~/.claude/settings.json` 모두 `mcpServers` 키 없음, `.mcp.json` 파일 없음, `~/.claude.json`의 mcpServers는 무관 항목(higgsfield)만. `moai mcp-server`는 Go 능력으로 존재하지만 어디에도 배선되어 있지 않다.
5. **사용자 전역 `~/.claude/settings.json` statusLine** — Orca wrapper 스크립트, `moai` 문자열 없음(영향 없음).

정산: 리더 실측 "42/48" 대비 본 문서 "44/46" — 측정 방법 차이(리더: 훅 스크립트 48개 기준 moai 호출 여부 / 본 문서: `.sh` 확장자 46파일 중 moai 문자열 포함 44, 디렉터리 49항목 중 비-스크립트 3). 어느 쪽도 규칙 판정에는 영향 없으며 이후 카드는 본 문서의 방법(§9 명령)을 표준으로 쓴다.

### 1.3 moai 제거 시 항목별 영향 ("제거"의 두 층을 구분)

`moai` 제거에는 두 층이 있다 — **(i) PATH 이름 제거**(shim 파일은 유지)와 **(ii) shim 파일 제거**(~/go/bin/moai 삭제·이동). 결과가 다르다.

| 항목 | (i) PATH 이름만 제거 | (ii) shim 파일까지 제거 |
| --- | --- | --- |
| 상태줄(현행 스크립트) | `command -v moai` 실패 → 29행 절대 폴백이 shim 파일을 exec → **Go statusline 계속 렌더**(변화 없음) | 29·34행 폴백 모두 실패(`~/.local/bin/moai` 부재) → 42행 exit 0 → **상태줄 소멸** |
| 훅 44파일(가드) | PATH 호출 실패 → `$HOME/go/bin/moai`(shim) 절대 폴백 → moai.old 훅 계속 실행 → **가드 유지**(차단 계약 포함) | 폴백 전부 실패 → handle-pre-tool 등은 exit 0(침묵, **차단하지 않는 안전 방향**) · `exec moai hook …`형도 exit 0 → **가드 소멸**(보호 없음) |
| MCP 서버 | 등록 자체가 없음 → **영향 없음** | 동일 |
| session 런처 | shim의 `session` 단어는 moai.old session(Go 세션 런처)로 라우팅 — PATH 이름 제거 후에도 shim 파일이 있으면 동작. `moai-cli session`(Rust)은 별개 이름으로 상시 가능 | `moai session` 소멸(Rust `moai-cli session`은 계속) |
| launcher(cc·glm·-l·-a) | shim 경유 Rust — shim 파일 유지 시 무관. `moai-cli` 직접 호출은 `~/.moai-cli/path`가 PATH에 없어 **절대 경로 필요** | `moai-cli` 절대 경로로만 가능 |

결론: **(i)만으로는 아무것도 깨지지 않는다**(shim 파일이 절대 폴백으로 모든 것을 계속 서빙). 실질 전환의 본체는 상태줄 스크립트 교체(§4)이고, 훅은 §3 선택지에 따른다.

## 2. 훅 3 선택지 판정

| 선택지 | 내용 | 즉시 가능 | 되돌리기 | 부작용 |
| --- | --- | --- | --- | --- |
| (가) 훅은 절대 경로 Go(moai.old) 호출로 변환, PATH 이름만 제거 | 44파일의 `moai hook` 호출을 `"$HOME/go/bin/moai.old" hook`으로 일괄 치환(sed 1패스) 후 `~/go/bin/moai` 파일 제거 | **가능**(기계적 치환) | 치환 역방향 sed + shim 복원(§7) | 훅은 계속 Go 동작 — Rust 전환은 나중으로. shim 삭제로 launcher 단어는 `moai-cli` 직접 호출 필요(§4 스크립트가 그 경로) |
| (나) 훅을 Rust hook 진입점으로 이전 후 제거 | Rust 쪽 훅 핸들러 구현이 선행 | **불가** — Rust 훅 진입점이 아직 없다(#2374는 statusline만) | — | #2374 이후 별도 카드 |
| (다) 이 프로젝트 Go 훅 종료 | settings.json 훅 블록 제거 | 가능 | 훅 블록 백업 복원 | **가드 소멸**(차단 계약·품질 게이트 전부 off) — 실행 중 세션 보호 상실 |

**권고: (가)를 이 전환의 훅 방침으로 채택, (나)를 #2374 병합 이후 후속 카드로 등록.** 근거: (가)는 무중단·기계적·완전 되돌리기 가능이고 가드를 유지한다. (나)는 전제(구현)가 없어 지금 설계할 수 없고, (다)는 사용자가 유지 중인 가드를 없애는 방향이라 본 카드의 목적(statusline만 동일하게)과 무관하게 범위를 넓힌다.

(가)의 정확한 절차(후속 구현 카드용, 본 PR은 실행하지 않음):
1. 사본: `cp -R ~/.claude/hooks/moai ~/.claude/hooks/moai.bak-<date>` (사용자 확인 후).
2. 치환: `*.sh`에서 `moai hook ` → `"$HOME/go/bin/moai.old" hook ` (182행 중 `moai hook`형 전부; 다른 형태는 §1.2 분포 표 기준 없음을 사전 grep으로 재확인).
3. 검증: 치환 후 각 파일의 남은 `moai ` 호출 0건(grep), dry 실행 1건(한 스크립트에 빈 stdin → 기존과 동일 분류).
4. 제거: `rm ~/go/bin/moai`(shim) — `~/.moai-cli/path/moai`는 비-PATH라 유지해도 무방하나 혼동 방지를 위해 함께 제거 권고.
5. 이 시점부터 launcher는 `moai-cli` 이름으로 호출(§4 스크립트가 절대 경로 폴백을 포함하므로 PATH 등록과 무관하게 동작).

## 3. 대체 statusline 스크립트 (`.moai/status_line.sh` 교체안 전문)

설계 원칙: (1) `moai-cli` 우선 — 절대 경로가 PATH 변화에 무관하게 먼저, (2) moai-cli 부재 시 **Go(moai.old) 절대 경로** 폴백 — 상태줄이 사라지지 않는다, (3) 마지막 폴백으로 현행 shim(전환 실패 시에도 렌더 유지), (4) `.env.glm` source 행동은 현행 그대로 보존(stdin 닫음·값 출력 없음), (5) 임시 파일 없음(현행 exec 규약 유지).

```bash
#!/bin/bash
# MoAI Statusline Wrapper — Issue #2383 design (replaces the moai-PATH lookup).
# 1) moai-cli (Rust, absolute install path first), 2) legacy Go via absolute
# moai.old, 3) the routing shim. stdin JSON is passed straight through; no
# temp files are created (exec replaces the shell, see the old comment below).

if [ -f "$HOME/.moai/.env.glm" ]; then
	# shellcheck disable=SC1091
	source "$HOME/.moai/.env.glm" </dev/null
fi

# 1) The Rust statusline (PR #2374). Absolute install path first: it does not
#    depend on PATH at all.
if [ -x "$HOME/.moai-cli/path/moai-cli" ]; then
	exec "$HOME/.moai-cli/path/moai-cli" statusline
fi
if command -v moai-cli &> /dev/null; then
	exec moai-cli statusline
fi

# 2) Legacy Go statusline by absolute path — keeps the same rendering when
#    moai-cli is missing. Deliberately NOT the `moai` routing shim: this script
#    must render even after the `moai` name leaves PATH.
if [ -x "$HOME/go/bin/moai.old" ]; then
	exec "$HOME/go/bin/moai.old" statusline
fi

# 3) Last resort: the old shim (renders via Go today). Removed in option (가)
#    step 4, so this arm exists only before that step runs.
if [ -x "/Users/goos/go/bin/moai" ]; then
	exec "/Users/goos/go/bin/moai" statusline
fi

# Not found - exit silently (Claude Code handles missing statusline gracefully)
exit 0
```

폴백 사슬 요약: `~/.moai-cli/path/moai-cli` → PATH `moai-cli` → `~/go/bin/moai.old` → 기존 shim → exit 0. "moai-cli가 없을 때"는 2단계(Go)로 렌더가 유지되고, "moai가 PATH에 없는 셸"에서도 1단계 절대 경로는 영향받지 않는다(Issue AC-2).

## 4. 사용자에게 보일 변경 전/후 diff (`.moai/status_line.sh`)

전환 적용 시 사용자에게 승인·공개할 diff의 요지(전문은 §3의 after와 현재 파일이 before):

```diff
--- a/.moai/status_line.sh   (before — PATH 이름 의존)
+++ b/.moai/status_line.sh   (after — moai-cli 우선 + Go 절대 폴백)
@@ -9,7 +9,7 @@
 if [ -f "$HOME/.moai/.env.glm" ]; then
     source "$HOME/.moai/.env.glm" </dev/null
 fi
-
-if command -v moai &> /dev/null; then
-	exec moai statusline
-fi
-
-# Try the installed binary's own resolved path (captured at init/update time).
-if [ -f "/Users/goos/go/bin/moai" ]; then
-	exec "/Users/goos/go/bin/moai" statusline
-fi
-
-# Try the Go bin directory ($HOME-relative for portability).
-if [ -f "$HOME/go/bin/moai" ]; then
-	exec "$HOME/go/bin/moai" statusline
-fi
-
-# Try user local bin directory.
-if [ -f "$HOME/.local/bin/moai" ]; then
-	exec "$HOME/.local/bin/moai" statusline
-fi
+if [ -x "$HOME/.moai-cli/path/moai-cli" ]; then
+	exec "$HOME/.moai-cli/path/moai-cli" statusline
+fi
+if command -v moai-cli &> /dev/null; then
+	exec moai-cli statusline
+fi
+if [ -x "$HOME/go/bin/moai.old" ]; then
+	exec "$HOME/go/bin/moai.old" statusline
+fi
+if [ -x "/Users/goos/go/bin/moai" ]; then
+	exec "/Users/goos/go/bin/moai" statusline
+fi

 # Not found - exit silently (Claude Code handles missing statusline gracefully)
 exit 0
```

사용자 관점 변화: `settings.json`은 **무변경**(statusLine.command는 같은 스크립트를 가리킨다). 화면에 보이는 상태줄은 #2374의 Rust statusline이 Go와 바이트 동일임이 구현 단계에서 검증된 뒤에야 바뀐다(Issue AC-1 — 이 카드의 실행 단계에서 수행).

## 5. 되돌리기 명령 (각 변경 1:1)

```sh
# 0) 적용 직전 사본 (변경의 일부로 먼저 수행)
cp ~/.moai/status_line.sh ~/.moai/status_line.sh.before-2383

# 1) 상태줄 스크립트 되돌리기 (사본 복원 — 즉시 효과, 다음 렌더부터 Go 렌더)
cp ~/.moai/status_line.sh.before-2383 ~/.moai/status_line.sh

# 2) 훅 치환 되돌리기 ((가)를 실행한 경우 — 사본 디렉터리 역복사)
rm -r ~/.claude/hooks/moai && cp -R ~/.claude/hooks/moai.bak-<date> ~/.claude/hooks/moai

# 3) PATH/shim 되돌리기 (shim 헤더의 Undo 그대로)
mv "$HOME/go/bin/moai.old" "$HOME/go/bin/moai" 2>/dev/null || true
#   (~/.moai-cli/path/moai를 제거했다면: cp "$HOME/go/bin/moai" "$HOME/.moai-cli/path/moai")

# 4) 확인: command -v moai → shim 재등장, 상태줄 즉시 Go 렌더 복귀
```

되돌리기 전제: 1·2는 사본이 있어야 하고(적용 절차 0단계), 3은 shim 헤더에 문서화된 공식 Undo다.

## 6. 적용 절차 (후속 구현 카드용 — 본 PR은 실행하지 않음)

0. **운영자 승인 1회**(저장소 설정·사용자 설정 변경 — release-checklist 규칙) + 사본 생성(§5-0, 훅은 §2-(가)-1).
1. **#2374 병합·재설치 확인**: `~/.moai-cli/path/moai-cli statusline`이 실행되는지(없으면 이 카드의 실행 부분은 대기).
2. **바이트 비교(Issue AC-1)**: 기록된 실제 stdin 사본 N개에 대해 `$HOME/go/bin/moai.old statusline` 출력 vs `moai-cli statusline` 출력 — 다르면 차이 목록·원인을 문서화하고 전환 보류.
3. **교체**: §3 스크립트로 `.moai/status_line.sh` 갱신(§5-0 사본 후) + 사용자에게 §4 diff 제시.
4. **검증(Issue AC-2)**: `env PATH=/usr/bin:/bin` (moai 이름 없는 셸)에서 같은 stdin → 같은 출력(1단계 절대 경로가 잡음). **AC-3**: 임시 HOME·임시 프로젝트에서 훅이 도구 호출을 막지 않음을 선택지 (가) 기준으로 확인(훅은 moai.old 절대 호출 유지).
5. **훅 (가) 전환**(§2 절차) — 별도 승인 시.
6. **`moai` 이름 제거**: `rm ~/go/bin/moai` (+ `~/.moai-cli/path/moai`) — 2·4·5가 전부 관찰된 뒤에만.
7. **기록**: decisions log에 적용·검증·되돌리기 명령과 함께 기록.

## 7. 이 문서의 실측 방법 (재현 가능한 명령)

- shim 3종: `which moai` + `diff ~/.moai-cli/path/moai ~/go/bin/moai`(IDENTICAL) + `ls ~/.moai-cli/path/`
- 훅 수: `ls .claude/hooks/moai/*.sh | wc -l`(46) · `grep -lic moai .claude/hooks/moai/*.sh | wc -l`(44) · `grep -rhE '(^|[;&|"'\'' (])(command -v )?moai([ ;&|)"'\''|$)' .claude/hooks/moai/ | wc -l`(182)
- 배선: `.claude/settings.json` 파싱 — 이벤트 20종, `hooks/moai/` 참조 36곳, 스크립트 30종, 차단 이벤트 스크립트 13종(§1.2 스크립트별 매핑은 본 감사 세션 로그의 이벤트→스크립트 표)
- statusLine: `.claude/settings.json` `statusLine.command`(10행 앵커) · `~/.claude/settings.json` statusLine(Orca wrapper, moai 문자열 없음)
- MCP: settings 2종 + `.mcp.json`(부재) + `~/.claude.json` mcpServers(moai 무관) — moai MCP 배선 0건
- status_line.sh: 행번호 목록(§1.2-1) · 전문 인용

## 8. 이 문서가 바꾸지 않는 것

`~/.claude/settings.json`·프로젝트 `.claude/settings.json`·`.moai/status_line.sh`·`.claude/hooks/moai/*`·PATH·`~/go/bin/*` — 본 PR은 `docs/`의 이 문서 하나만 추가한다. 모든 실행은 위 절차의 운영자 승인 후 구현 카드에서 수행한다.
