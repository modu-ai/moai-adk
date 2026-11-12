# 양 호스트 실세션 시험 체크리스트 (I8 #2245, rc.2)

Claude Code와 Codex에서 **AGENTS.md 로드·skill 노출·hook 발화**를 운영자가 한 번에 실측하기 위한 절차다. 레인은 실세션을 실행하지 않는다(카드 '항상 blocked': 미승인 host 실세션·유료 호출 금지). 레인 산출물은 이 문서, 하네스 `tools/host_live_harness.py`, 하네스 시험 `tests/host_live_harness_test.py`까지다.

기존 T2 스크립트(`tests/e2e/host-claude.sh` AC-08C-3/5/6, `tests/e2e/host-codex.sh` AC-08C-4)는 `/moai` 명시 호출과 권한 거부를 다룬다. 이 체크리스트는 그 앞 단계인 "호스트가 지침 파일과 skill을 실제로 읽는가, hook이 실제로 도는가"를 따로 잰다.

## 1. 항목

판정은 모두 구조로 한다. 하네스는 실행마다 무작위 표식(`mp-<16hex>`)을 임시 프로젝트의 지침 파일·skill에 넣고, 호스트 응답에 그 표식이 있는지만 본다. 문장 내용은 판정에 쓰지 않는다.

| AC | host | 방식 | 확인하는 것 | 판정 규칙 |
| --- | --- | --- | --- | --- |
| AC-I8C-1 | Claude Code | 스크립트 | CLAUDE.md가 없을 때 AGENTS.md 로드 | `-p … --output-format json`의 `.result`(문자열, `is_error`≠true)에 AGENTS.md 표식이 있음 |
| AC-I8C-2 | Claude Code | 스크립트(음성 대조) | CLAUDE.md(import 없음)가 있으면 AGENTS.md를 읽지 않음 | `.result`에 AGENTS.md 표식이 **없음** |
| AC-I8C-3 | Claude Code | 스크립트 | project skill 노출·명시 호출 | `/moai-probe` 응답에 probe skill 표식이 있음 |
| AC-I8C-4 | Claude Code | 스크립트 | project `SessionStart` hook 발화(호스트 메커니즘) | hook이 임시 프로젝트에 쓴 `hook-fired.txt`가 표식과 같음 |
| AC-I8C-5 | Claude Code | **결정 대기** | MoAI hook 발화 | 실행하지 않음 — RC1 MoAI hook 0개(P04 = A) |
| AC-I8X-1 | Codex | 스크립트 | AGENTS.md 로드 | `exec --json`에 `turn.completed` 이벤트가 있고, `-o` 마지막 메시지에 표식이 있음 |
| AC-I8X-2 | Codex | 스크립트(대체 대조) | 같은 디렉터리의 AGENTS.override.md가 AGENTS.md를 대체 | override 표식은 있고 AGENTS.md 표식은 없음 |
| AC-I8X-3 | Codex | 스크립트 | project skill 노출·명시 호출 | `$moai-probe` 마지막 메시지에 probe skill 표식이 있음 |
| AC-I8X-4 | Codex | **수동** | project hook 발화(호스트 메커니즘) | §4.3 수동 절차. 스크립트로 돌리려면 hook 신뢰 우회 플래그가 필요한데, 그것은 보안 우회라 쓰지 않는다 |
| AC-I8X-5 | Codex | **결정 대기** | MoAI hook 발화 | 실행하지 않음 — hook 생성은 Codex 계약 C4의 별도 결정 |

AC-I8C-4와 AC-I8X-4는 **호스트의 hook 메커니즘이 이 격리 환경에서 관측 가능한지**를 잴 뿐이다. MoAI hook이 발화한다는 주장(AC-I8C-5, AC-I8X-5)과는 다른 행이고, MoAI hook이 생기기 전에는 passed가 될 수 없다.

각 임시 프로젝트에는 `skills/moai` 전체도 호스트 위치(`.claude/skills/moai` 또는 `.agents/skills/moai`)에 함께 복사된다. receipt의 `skill_digest`는 그 `SKILL.md`의 sha256이다.

## 2. 사전 조건

1. 실행 위치는 이 저장소 루트다(하네스가 `skills/moai`를 읽는다). receipt 디렉터리(`--out`)는 체크아웃 **밖**에 둔다. 예: `"$(mktemp -d)/i8"`.
2. 자격 증명은 운영자가 이미 가진 것을 환경 변수로만 넘긴다: Claude Code는 `ANTHROPIC_API_KEY`, Codex는 `CODEX_API_KEY`. 하네스는 변수 이름만 기록하고, 호스트 원시 출력에 값이 나타나면 `[REDACTED:<이름>]`으로 바꾼 뒤 보관한다.
3. `--live` 실행은 API 유료 호출이다. 스크립트 항목 7개는 각각 호스트를 한 번 띄우고, 호스트마다 `--version`도 한 번 실행한다. 실행 여부는 운영자가 정한다.
4. 호스트 버전: Claude Code 2.1.277 이상(AGENTS.md 직접 읽기 최소 버전, `src/claude.rs` `MIN_AGENTS_VERSION`).

## 3. 격리 방식과 그 근거

항목마다 새 임시 프로젝트와 새 임시 HOME을 만든다. 호스트 프로세스의 환경은 `PATH`·`LANG`·`LC_ALL`·`TMPDIR`와 해당 자격 증명 변수만 남기고, 다음 값으로 덮는다.

| 변수 | 값 | 근거 (2026-11-09 공식 문서 확인) |
| --- | --- | --- |
| `HOME` | 임시 HOME | 사용자 `~/.claude`·`~/.codex`·`~/.agents`·`~/.moai-cli`를 설정으로 읽지도 쓰지도 않게 한다 |
| `CLAUDE_CONFIG_DIR` | `<임시 HOME>/.claude` | settings: "set `CLAUDE_CONFIG_DIR`; Claude Code then stores your settings, session history, and plugins there instead" |
| `CODEX_HOME` | `<임시 HOME>/.codex` | environment variables: "Sets the root for Codex state, including config, auth, logs, sessions, skills, and standalone package metadata" |

임시 트리는 항목이 끝나면 지운다. 원시 출력은 `--out` 디렉터리에만 남는다.

**`--bare`를 쓰지 않는 이유.** cli-reference는 `--bare`가 "auto-discovery of hooks, skills, custom commands, subagents, installed plugins, MCP servers, auto memory, and CLAUDE.md"를 건너뛴다고 적는다. 이 체크리스트가 재려는 skill·hook·지침 파일 탐색을 끄는 옵션이므로 쓰지 않고, 위 격리로 대신한다.

**문서로 확인하지 못한 점(실측으로 확인할 것):**

- Claude Code: 새 `CLAUDE_CONFIG_DIR`에서 `-p` 첫 실행이 온보딩 없이 진행되는지, `--bare`가 아닐 때 macOS 키체인의 OAuth를 함께 읽는지는 문서에 없다. 운영자 계정 상태와 다르게 동작하면 그 사실을 receipt 비고에 적는다.
- Claude Code: AGENTS.md 탐색이 `--bare`에서도 꺼지는지는 문서에 적혀 있지 않다(CLAUDE.md만 명시).
- Codex: `codex exec`가 AGENTS.md와 skill을 읽는지, hook을 실행하는지는 non-interactive mode 문서에 적혀 있지 않다. 하네스는 임시 프로젝트에서 `git init`을 해 프로젝트 루트를 고정한다(AGENTS.md 문서: "root (usually Git root)").
- headless `-p`는 project `.claude/settings.json` hook을 신뢰 대화 없이 실행한다(headless: "Without `--bare`, a `-p` session runs the hooks in a project's `.claude/settings.json` … even in a folder you've never trusted"). AC-I8C-4는 이 동작을 임시 프로젝트 안에서만 쓴다.

## 4. 실행 절차

### 4.1 dry-run (자격 증명·호스트 불필요, 아무것도 띄우지 않음)

```sh
OUT="$(mktemp -d)/i8"
python3 -I tools/host_live_harness.py --out "$OUT"; echo "exit=$?"
```

기대 결과: exit 0, 10개 항목 모두 `not_run`, 스크립트 항목의 `detail`에 "dry-run: not executed: <실행할 명령>"이 나온다. `"$OUT/rows.md"`에는 not_run 행 표만 있다.

### 4.2 live (운영자 1회)

```sh
# 자격 증명은 이미 환경에 있는 값만 쓴다. 값을 명령줄에 적지 않는다.
python3 -I tools/host_live_harness.py --out "$OUT" --live; echo "exit=$?"
```

한 호스트만 잴 때는 `--only AC-I8C-1 --only AC-I8C-3`처럼 항목을 고른다. 다른 경로의 호스트 바이너리는 `--bin claude=<path>` / `--bin codex=<path>`로 준다. 바이너리가 없거나 자격 증명 변수가 비어 있으면 그 항목은 실행하지 않고 `not_run`으로 남는다.

exit 코드: 0 = receipt 작성, 1 = 입력을 읽을 수 없음(`skills/moai/SKILL.md` 없음 등) 또는 I/O 실패, 2 = 사용법 오류. 항목의 failed는 exit 코드가 아니라 receipt의 `result`로 나온다.

### 4.3 수동 항목 (대화형)

모두 임시 프로젝트에서 한다. 사용자 HOME 설정은 바꾸지 않는다.

```sh
WORK="$(mktemp -d)"
mkdir -p "$WORK/home" "$WORK/project/.claude/skills" "$WORK/project/.agents/skills"
cp -R skills/moai "$WORK/project/.claude/skills/moai"
cp -R skills/moai "$WORK/project/.agents/skills/moai"
printf '# Probe\n\nMOAI_PROBE_MARKER=manual-check\n' > "$WORK/project/AGENTS.md"
```

| 기록 id | host | 절차 | passed 조건 |
| --- | --- | --- | --- |
| AC-I8C-6 | Claude Code | `cd "$WORK/project" && HOME="$WORK/home" CLAUDE_CONFIG_DIR="$WORK/home/.claude" claude` → `/memory` | 목록에 프로젝트 `AGENTS.md`가 보임 |
| AC-I8C-7 | Claude Code | 같은 세션에서 `/` 메뉴 | `/moai`가 나옴 |
| AC-I8X-6 | Codex | `cd "$WORK/project" && git init -q && HOME="$WORK/home" CODEX_HOME="$WORK/home/.codex" codex` → `/skills` | `moai`가 나옴 |
| AC-I8X-4 | Codex | 임시 프로젝트에 `.codex/hooks.json`으로 `SessionStart` hook(임시 프로젝트 안의 파일을 쓰는 명령)을 두고 Codex를 시작 → `/hooks`에서 검토·신뢰 → 재시작 | hook이 쓴 파일이 생김. 신뢰 기록은 임시 `CODEX_HOME`에만 남는다 |

AC-I8X-4는 문서상 "Non-managed hooks must be reviewed and trusted before they run", "load only when the project `.codex/` layer is trusted"라서 대화형 검토가 필요하다. 프로젝트 신뢰 키는 `projects.<path>.trust_level`이고, hook 신뢰를 기록하는 정확한 키는 문서에 없다.

수동 결과는 하네스와 같은 receipt 형식(`schemas/e2e-receipt.schema.json`)으로 한 줄씩 `"$OUT/receipts.jsonl"`에 덧붙인다. `detail`에는 관찰 내용(예: "/memory 목록에 AGENTS.md 보임")을 적고 자격 증명 값은 적지 않는다.

```json
{"ac": "AC-I8C-6", "os": "Darwin", "arch": "arm64", "host": "claude-code", "host_exact_version": "<claude --version>", "mode": "local", "skill_digest": "<shasum -a 256 skills/moai/SKILL.md>", "moai_cli_sha": "absent", "result": "passed", "ran_at": "<date -u +%Y-%m-%dT%H:%M:%SZ>", "detail": "manual: /memory lists project AGENTS.md", "probe": "I8"}
```

수동 항목은 하네스의 `rows.md`에 나오지 않는다. 지원표에는 §5의 형식대로 직접 행을 적는다.

## 5. 결과 양식과 지원표 갱신

하네스 출력(REQ-I8-01):

- `"$OUT/receipts.jsonl"`: 항목마다 1줄. 필드는 receipt 스키마 필수 필드(`ac`, `os`, `arch`, `host`, `host_exact_version`, `mode`, `skill_digest`, `moai_cli_sha`, `result`, `ran_at`)에 `detail`·`probe: "I8"`·`kind`가 더해진다. `result`는 `passed` / `failed` / `not_run`이다.
- `"$OUT/<AC>.out"`, `.err`, (Codex) `.last.md`: 판정에 쓴 원시 출력. `detail`은 `--out` 기준 파일 이름과 판정 이유다(예: `AC-I8C-1.out (marker check held)`). dry-run `detail`의 명령은 호스트 이름과 `<out>` 자리표시자로 적어 기계별 절대 경로를 남기지 않는다.
- `"$OUT/rows.md"`: `docs/support-matrix.md` 행. passed는 "지원 행" 16열 형식(결과 `passed`, receipt `receipts.jsonl`), 나머지는 "not_run 행" 6열 형식(결과 `not_run`/`failed`, 사유 = `detail`).

지원표 반영은 지원표 소유 카드(08-D #2015 / 02-C #1993 계열) 또는 리더가 PR로 한다. 이 카드는 지원표를 직접 고치지 않는다.

1. receipt를 저장소로 옮긴다: `"$OUT/receipts.jsonl"` 내용을 `docs/e2e-receipts/receipts.jsonl` 끝에 덧붙이고, passed 행의 원시 출력은 `docs/e2e-receipts/I8-<AC>-detail.txt`로 복사한다. 이때 `detail`의 파일 이름을 저장소 경로(`docs/e2e-receipts/I8-<AC>-detail.txt`)로 고친다. 원시 출력에 자격 증명 값·개인 절대 경로가 없는지 본다.
2. `rows.md`의 passed 행은 "지원 행" 표에, 나머지는 "not_run 행" 표에 붙인다. 기존 not_run 행 중 같은 항목이 있으면 바꾼다.
3. "집계" 절의 지원 행 수와 passed receipt 수를 고친다.
4. 검사기를 돌린다. exit 0이어야 한다:

   ```sh
   python3 -I tools/check_support_matrix.py docs/support-matrix.md docs/e2e-receipts/; echo "exit=$?"
   ```

5. `docs/e2e-matrix.md` T2 표의 해당 행 `capability_status`를 실측 결과에 맞게 고친다(passed → `probed`).

## 6. 이 카드에서 남는 것

- 스크립트 7항목(AC-I8C-1~4, AC-I8X-1~3) + 수동 4항목(AC-I8C-6·7, AC-I8X-4·6): 전부 not_run(운영자 실측 대기).
- AC-I8C-5, AC-I8X-5: 결정 대기(MoAI hook 생성 결정 — Codex 계약 C4, P04).
- 프로젝트 skill 위치(P04 HOLD)와 EP-H Q-A/B/C는 이 카드가 정하지 않는다. 하네스가 skill을 프로젝트 위치에 두는 것은 측정용 격리 fixture일 뿐, 설치 위치 결정이 아니다.
- 후속 후보(가설, 미검증): `tests/e2e/host-claude.sh`의 AC-08C-3/6은 `claude --bare`로 `/moai`를 부른다. 위 cli-reference 문구대로라면 `--bare`는 skill 자동 탐색을 끄므로, 그 시험이 skill을 찾지 못할 수 있다. 실행해 보지 않았으므로 결함으로 단정하지 않는다. 또 그 스크립트의 `detail`은 EXIT trap이 지우는 임시 경로를 가리킨다. 이 하네스는 원시 출력을 `--out`에 남긴다.

## 출처 (2026-11-09 확인)

- https://code.claude.com/docs/en/cli-reference — `--bare` 범위, `--include-hook-events`, `--setting-sources`
- https://code.claude.com/docs/en/headless — `--bare` 인증("never reads OAuth credentials or the system keychain"), `-p`의 project hook 실행, `system/init` 필드
- https://code.claude.com/docs/en/settings — `CLAUDE_CONFIG_DIR`
- https://code.claude.com/docs/en/hooks — `SessionStart`("When a session begins or resumes")
- https://code.claude.com/docs/en/memory — AGENTS.md 로드 조건(`docs/host-support-matrix.md` Claude `agents_loading` 행)
- https://learn.chatgpt.com/docs/config-file/environment-variables — `CODEX_HOME`, `CODEX_API_KEY`("Provides an API key to a non-interactive Codex process")
- https://learn.chatgpt.com/docs/non-interactive-mode — `exec --json` 이벤트(`thread.started`, `turn.started`, `turn.completed`, `turn.failed`, `item.*`, `error`; 페이지 뒷부분은 요약으로만 읽음)
- https://learn.chatgpt.com/docs/hooks — `SessionStart`, hook 신뢰 요구
- https://learn.chatgpt.com/docs/agent-configuration/agents-md — override 우선순위(`docs/host-support-matrix.md` Codex `agents_loading` 행)
- https://learn.chatgpt.com/docs/build-skills — `.agents/skills` 탐색, `$`·`/skills` 호출
