# t494 — codex 공식 문서 기반 moai 배선 확장안 (조사 전용)

| 항목 | 값 |
|---|---|
| 카드 | t494 |
| 트리 | `.claude/worktrees/t494` · 브랜치 `WT-codex-doc-survey` · HEAD `ace1c5440` (`origin/develop`과 델타 `0 0`) |
| 조사 시각 | 2026-09-07 |
| 관측 대상 CLI | **codex-cli 0.153.4** (`/Users/goos/.local/bin/codex --version`) |
| 종전 판정 기준 | SPEC-CODEX-SKILL-LOADER-001 = codex-cli **0.152.1** · `agents-codex.yaml` 매니페스트 = **0.147.0** |
| 산출 성격 | 조사 전용. 코드·템플릿 수정 0. 후속 구현 카드(t496·t502)의 입력 |

> **버전 표기의 한계 [HARD]**: OpenAI 공식 문서(`learn.chatgpt.com/docs/*`)에는 **버전 스탬프가 없다**. 따라서 아래 인용은 "codex-cli 0.153.4용 문서"가 아니라 **2026-09-07 조회 시점의 현행 문서**다. 로컬 0.153.4의 실제 표면(`codex --help`, `~/.codex/` 트리)과 대조해 어긋나지 않음을 확인한 항목만 "정합"으로 적었고, 대조하지 못한 항목은 §8 Gaps에 남겼다.

---

## 0. 방법

- 1차 출처는 **리포 밖**에서 가져왔다. `WebSearch`로 후보를 찾고 **`WebFetch`로 실제 본문을 읽은 URL만** 인용한다. 생성한 URL은 없다.
- `developers.openai.com/codex/*`는 전부 **308로 `learn.chatgpt.com/docs/*`로 이전**돼 있다. 두 주소를 함께 적되, 인용 본문은 실제로 읽힌 이전 후 주소의 것이다.
- 리포 쪽 현재 상태는 `file:line`으로 인용한다. 부재 주장에는 **대조군**을 함께 실행해 검사 자체가 살아 있음을 보였다.

### 선행 조사의 오류 — 되풀이하지 않기 위한 기록

앞선 Explore는 "`~/.codex/prompts/*.md`가 codex의 커스텀 커맨드 메커니즘"이라 썼다가 출처 없음으로 철회했다. **이번 조사 결과 그 진술은 절반만 틀렸다**: 그런 메커니즘은 실재했고 공식 문서에 지금도 남아 있으나, **deprecated 상태이며 스킬로 대체**됐다(§2). 즉 철회는 옳았지만 이유가 달랐다 — 근거가 없던 것이지 사실이 없던 것이 아니다. 근거를 리포 안에서 찾으려 한 것이 실패의 원인이고, 이번 카드가 리포 밖으로 나간 이유다.

---

## 1. 리드 실측 재현 (인용 아님 — 이 트리에서 다시 잼)

| # | 주장 | 재현 명령 | 관측 | 판정 |
|---|---|---|---|---|
| M1 | `codex/prompts`·`prompts_dir` 리포 전체 0히트 | `git grep -nE 'codex/prompts\|prompts_dir' -- .` | rc=1, `hits=0`. 대조군 `codex_task` → 115파일 | 일치 |
| M2 | 템플릿 `.codex/` 하위는 `agents/moai/`뿐 | `find internal/template/templates/.codex -maxdepth 2` | `.codex` · `.codex/agents` · `.codex/agents/moai` 3개 디렉터리뿐 | 일치 |
| M3 | `/moai` 커맨드 16본 | `find …/.claude/commands/moai -maxdepth 1 -type f` | 16 (15 `.md.tmpl` + `todo.md`) | 일치 |
| M4 | 11종 중 6종만 adapted | `awk '/^var EventTable/,/^}/' … \| grep -c 'true},'` / `'false},'` | `true`=6, `false`=5, 합 11 | 일치 |

M4 보충: 세 번째 필드는 **위치 인자**이지 `Adapted: true` 명명 필드가 아니다(`grep -c 'Adapted:'` = 0). 리드 서술은 의미상 정확하나, 검증자가 `Adapted:` 문자열로 재현을 시도하면 0을 얻어 없는 결함을 만들 수 있으므로 적어 둔다. 실체는 `internal/codexadapter/events.go:52-64`.

---

## 2. 축 1 — 커스텀 프롬프트 / 커맨드 표면

### (a) 공식 URL
- https://developers.openai.com/codex/custom-prompts → **308** → https://learn.chatgpt.com/docs/custom-prompts (본문 읽음)
- https://developers.openai.com/codex/skills → **308** → https://learn.chatgpt.com/docs/build-skills (본문 읽음)

### (b) 원문 인용
> **Deprecation Status:** Custom prompts are deprecated. The documentation explicitly states: "Use skills for reusable prompts."
>
> **Directory Location:** Custom prompts are stored in `~/.codex/prompts/`.
> **Invocation Syntax:** `/prompts:promptname` in both the Codex CLI and IDE extension.
> **Placeholder Rules:** `$1`–`$9`, named `$FILE` (`KEY=value`), `$ARGUMENTS`, `$$` for a literal `$`.
> Custom prompts "require explicit invocation and live in your local Codex home directory... they're not shared through your repository."

스킬 쪽 호출 규약:
> In Codex, invoke skills explicitly using the `$` sigil: `$skill-name` or `/skills` to browse available skills.

### (c) 현재 moai 배선과의 차이
- 로컬 실측: `~/.codex/prompts` **없음**(`ls` → No such file or directory). 반면 `~/.codex/skills`·`~/.codex/plugins`는 **존재**.
- moai는 `.codex/prompts`를 만들지도, `~/.codex/prompts`에 쓰지도 않는다(M1 = 0히트).
- **결론: 차이가 아니라 정합이다.** deprecated 표면을 쓰지 않는 현재 상태가 문서 권고와 같은 방향이다. `/moai` 16본을 codex 쪽 슬래시 표면으로 옮기려면 경로는 **프롬프트가 아니라 스킬**이다.

### (d) 구현 카드 후보
- **C1 (권고: 채택)** — `/moai` 16본을 codex 스킬로 노출하는 경로 확정. moai는 이미 `.agents/skills/<name>` 미러를 만들고 있으므로(§4), 커맨드를 스킬 모양으로 바꾸는 것이 아니라 **커맨드 본문을 스킬로 발행하는 emitter**를 붙이는 형태가 된다.
- **C2 (권고: 기각)** — `~/.codex/prompts` 발행. deprecated 표면이고, 사용자 홈에 쓰는 방식이라 리포 공유가 안 된다. 문서가 명시적으로 스킬을 대안으로 지목한다.

---

## 3. 축 2 — 훅 이벤트 전체 목록

### (a) 공식 URL
- https://developers.openai.com/codex/hooks → **308** → https://learn.chatgpt.com/docs/hooks (본문 읽음)
- https://developers.openai.com/codex/config-reference → **308** → https://learn.chatgpt.com/docs/config-file/config-reference (본문 읽음)

### (b) 원문 인용
> The documentation lists these hook events: `SessionStart`, `SessionEnd`, `SubagentStart`, `SubagentStop`, `PreToolUse`, `PermissionRequest`, `PostToolUse`, `PreCompact`, `PostCompact`, `UserPromptSubmit`, `Stop`, and `Interrupt`.

로딩 경로:
> Codex discovers hooks from these paths: `~/.codex/hooks.json`, `~/.codex/config.toml`, `<repo>/.codex/hooks.json`, `<repo>/.codex/config.toml`.
> "If more than one hook source exists, Codex loads all matching hooks." · "Plugin-bundled hooks load alongside other hook sources."
> "Higher-precedence config layers don't replace lower-precedence hooks." · 프로젝트 훅은 "when the project `.codex/` layer is trusted"에만 로드.

인라인 TOML 형식:
```toml
[[hooks.SessionStart]]
matcher = "startup|resume"

[[hooks.SessionStart.hooks]]
type = "command"
command = "python3 ~/.codex/hooks/session_start.py"
```

동기 / 배경:
> **Synchronous** (기본): "Codex waits for a command hook to finish before continuing the operation that triggered it."
> **Background** (`"async": true`): "When a background hook finishes, Codex delivers supported informational output at the next safe point in the conversation." · "Codex runs up to eight background hooks concurrently per session."

입력 페이로드 공통 필드: `session_id`, `transcript_path`, `cwd`, `hook_event_name`, `model`, `permission_mode` (턴 스코프 훅은 `turn_id` 추가).

출력 / 판정:
```json
{ "continue": true, "stopReason": "optional", "systemMessage": "optional", "suppressOutput": false }
```
> Exit `0`: Success · Exit `2`: Blocking decision (error message written to stderr)
```json
{ "hookSpecificOutput": { "permissionDecision": "deny", "permissionDecisionReason": "reason text" } }
```

관리형 훅:
> `allow_managed_hooks_only = true` — "skip hooks from user, project, session, and plugin sources, but still loads managed hooks from `requirements.toml`."

### (c) 현재 moai 배선과의 차이

**차이 1 — `Interrupt` 이벤트가 moai 테이블에 없다.**
공식 12종 대 moai 11종. 누락은 정확히 `Interrupt` 하나다.
- 측정: `grep -rn 'EventInterrupt' internal/` → **0히트**. 대조군 `EventStop`은 `internal/hook/audit_test.go:493` 등에서 검출 → 검사 자체는 살아 있음.
- moai의 11종: `internal/codexadapter/events.go:52-64`.

**차이 2 — adapted 6종은 문서가 아니라 측정 기준이다.**
소스 주석(`events.go:40-51`)이 스스로 적고 있다: 미adapted 5종은 "counterpart 부재"가 아니라 **측정 커버리지 결정**이고, `SubagentStop`은 발화하지 **않는 것으로 측정**됐다(위임이 `tool_name`이 `collaboration`으로 시작하는 `PostToolUse`로 나타남). 이 서술은 현행 문서에 `SubagentStop`이 정식 이벤트로 남아 있는 사실과 **모순되지 않는다** — 문서는 이벤트의 존재를, 주석은 이 배선에서의 발화 여부를 말한다. 다만 **0.152.1 이전에 잰 관측**이므로 0.153.4에서의 재측정 가치가 있다.

**차이 3 — 커맨드 형태는 정합.**
`internal/codexwiring/hooks.go:104` + `codexwiring.go:46`이 `moai hook <arg> --harness codex`를 `.codex/hooks.json`에 쓴다. 공식 `type="command"` 규약과 같은 모양이다. 파일 타깃도 `HooksRelPath = ".codex/hooks.json"`(`codexwiring.go:29`)로 문서의 `<repo>/.codex/hooks.json`과 일치.

**차이 4 — `async` / `matcher` 미사용.** moai가 발행하는 핸들러에 `async`나 `matcher`가 실린다는 근거를 찾지 못했다. 없다고 단정하지 않는다(§8 Gaps).

**차이 5 — `allow_managed_hooks_only`.** `requirements.toml` 경유 관리형 훅 경로는 moai에 배선이 없다. 기업 배포 축이라 우선순위는 낮다.

### (d) 구현 카드 후보
- **H1 (권고: 채택, 소)** — `Interrupt`를 `EventTable`에 미adapted 행으로 추가. 12종 열거를 문서와 일치시키고, adapted 여부는 측정 뒤에 결정. 리스크 낮음.
- **H2 (권고: 채택, 중)** — 0.153.4에서 미adapted 5종 재측정(특히 `SubagentStop` 발화 여부). 판정이 뒤집히면 adapted 6→최대 11로 확장 가능.
- **H3 (조건부)** — 배경 훅(`async: true`) 도입. 동시 8개 상한이 문서에 명시돼 있으므로 상한을 넘기지 않는 설계가 전제.
- **H4 (권고: 보류)** — `allow_managed_hooks_only` / `requirements.toml`. 수요 근거 없음.

---

## 4. 축 3 — `[[skills.config]]` 규약

### (a) 공식 URL
- https://learn.chatgpt.com/docs/build-skills (본문 읽음)
- https://learn.chatgpt.com/docs/config-file/config-reference (본문 읽음)

### (b) 원문 인용
탐색 경로 — **`.codex/skills`가 아니라 `.agents/skills`다**:

| Scope | Path |
|---|---|
| REPO | `$CWD/.agents/skills` |
| REPO | `$CWD/../.agents/skills` |
| REPO | `$REPO_ROOT/.agents/skills` |
| USER | `$HOME/.agents/skills` |
| ADMIN | `/etc/codex/skills` |
| SYSTEM | Bundled with Codex |

디렉터리 레이아웃과 프론트매터:
```
my-skill/
├── SKILL.md (Required)
├── scripts/ (Optional)
├── references/ (Optional)
├── assets/ (Optional)
└── agents/openai.yaml (Optional)
```
```yaml
---
name: skill-name
description: Explain exactly when this skill should and should not trigger.
---
```

설정 키 — 비활성화 전용:
> To disable skills without deletion, use this in `~/.codex/config.toml`:
```toml
[[skills.config]]
path = "/path/to/skill/SKILL.md"
enabled = false
```

config-reference의 필드 정의:
> **`[[skills.config]]`** — Array of skill configurations: `path` (string, path to skill folder), `enabled` (boolean)

> **주의 — 문서 내부 불일치**: build-skills 예시는 `path`를 **`SKILL.md` 파일**로 적고, config-reference는 **"path to skill folder"**라 적는다. 두 문서가 같은 키의 값 모양을 다르게 기술한다. 어느 쪽이 참인지는 **측정하지 않았다**(§8 Gaps). 이 축을 구현할 카드는 이 모호성을 먼저 실측으로 닫아야 한다.

### (c) 현재 moai 배선과의 차이

**정합 1 — 미러는 이미 옳은 경로에 있다.**
`internal/template/skill_mirror.go:52` — `mirrorSkillsRelDir = filepath.Join(".agents", "skills")`. 파일 머리말(`:4-5`)이 근거를 직접 적는다: codex는 `.claude/skills/`를 보지 않고 `<repo>/.agents/skills/`를 스캔하므로, 배포한 스킬마다 `.agents/skills/<name>` → `../../.claude/skills/<name>` 상대 심볼릭 링크를 만든다(링크 불가 플랫폼에서는 실 디렉터리 복사로 폴백). 진입점은 `WithSkillMirror`, `Deploy`당 1회(`:171`).
→ **"moai에 codex 스킬 배선이 없다"는 진술은 거짓이다.** 있다. 다만 `[[skills.config]]`를 통한 것이 아니라 **경로 관례**를 통한 것이다.

**차이 1 — `[[skills.config]]`는 읽기만 하고 쓰지 않는다.**
`internal/codexwiring/skills.go`는 자기 주석(`:9`)에 "READ-ONLY by construction — nothing here writes"라 적는다. 소비처는 `moai doctor`의 권고 표면뿐(`internal/cli/doctor_codex.go:303,363,475`). 발행자는 없다.
- 측정: `git grep -n 'skills.config' -- internal/ pkg/ cmd/` → 히트 전부 `doctor_codex.go` 및 그 테스트. 발행 경로 0.

**차이 2 — 종전 판정의 유효 범위.**
SPEC-CODEX-SKILL-LOADER-001은 `completed`이며 `progress.md:90`에 `codex_version_observed: "codex-cli 0.152.1"`을 기록한다. 판정 요지: `<repo>/.agents/skills/<name>/SKILL.md`는 **실제로 로드된다**(분기 A, 두 독립 탐침으로 확인). 반면 `skills` 키의 세 성질(존재·값 모양·잘못된 값에서의 동작)은 **"확인되지 않음"**으로 남았고(AC-CSL-004), `skill-loader` 매니페스트 행은 `documented-drop`으로 처리됐다.
매니페스트 최상단 `codex_measured_version`은 `"0.147.0"`으로 **의도적으로 유지**됐다(`agents-codex.yaml:14`) — 이번에 잰 축이 하나뿐이라 최상단을 올리면 갖지 않은 커버리지를 주장하게 되기 때문(`progress.md:118`). 조용한 미갱신이 아니라 기록된 판단이다.
→ **t502의 성패는 이 지점에 걸려 있다.** 현행 문서는 `[[skills.config]]`를 `path` + `enabled` 두 필드로 **명시적으로 문서화**한다. 즉 0.152.1 시점에 "확인되지 않음"이던 축이 지금은 **문서상 확정**됐다. 종전 기각은 그때는 옳았고 지금은 재검토 대상이다.

**차이 3 — 용도의 방향.**
문서상 `[[skills.config]]`는 **비활성화 수단**("To disable skills without deletion")이지 등록 수단이 아니다. 등록은 경로 관례가 한다. 따라서 moai가 이 테이블에 쓸 이유는 "스킬을 codex에 보이게 하려고"가 아니라 "특정 스킬을 끄려고"일 때뿐이다.

**참고 — 유령 등록.** SPEC은 저자 머신 `~/.codex/config.toml`에 존재하지 않는 경로를 가리키는 `[[skills.config]]` 49건이 남아 있었고, 이는 Go 이전 moai가 남긴 죽은 장부라고 기록한다. `moai doctor`가 이를 stale로 보고한다(`doctor_codex.go:363`).

### (d) 구현 카드 후보
- **S1 (권고: 채택)** — `path` 값 모양(파일 vs 폴더) 실측으로 문서 불일치를 닫기. 0.153.4에서 두 형태를 각각 넣고 `codex debug prompt-input`으로 로드 여부 판정. **다른 모든 skills.config 작업의 선행 조건.**
- **S2 (조건부, S1 종속)** — 선택적 비활성화 발행. `moai update`가 사용자 스킬을 끄는 것은 침습적이므로 기본 비활성 + 명시적 opt-in.
- **S3 (권고: 채택, 소)** — 유령 `[[skills.config]]` 정리 verb. doctor가 이미 탐지만 하고 치우지는 않는다.
- **S4 (권고: 채택)** — 매니페스트 `codex_measured_version`을 0.153.4로 올리는 **전수 재측정** 카드. 지금 방식(축별 개별 기록)은 정직하지만 축이 늘수록 읽기 어려워진다.

---

## 5. 축 4 — agents TOML 소비 규약

### (a) 공식 URL
- https://developers.openai.com/codex/subagents.md → **308** → https://learn.chatgpt.com/docs/agent-configuration/subagents.md (본문 읽음)
- https://learn.chatgpt.com/docs/config-file/config-reference (본문 읽음)

### (b) 원문 인용
> **File Locations** — Personal agents: `~/.codex/agents/` · Project-scoped agents: `.codex/agents/`
> Each file is a standalone TOML document defining one custom agent.
>
> **Required Fields** — Every custom agent file must include: `name`, `description`, `developer_instructions`.

```toml
name = "pr_explorer"
description = "Read-only codebase explorer for gathering evidence before changes."
model = "gpt-5.3-codex-spark"
model_reasoning_effort = "medium"
sandbox_mode = "read-only"
developer_instructions = """
Stay in exploration mode. Trace execution paths, cite files and symbols.
"""
```

> **Built-in Agents** — `default`, `worker`, `explorer`
> **Optional Config Keys in Agent Files** — `model`, `model_reasoning_effort`, `sandbox_mode`, `mcp_servers`, `skills.config`
> **Global `[agents]` Table** — `enabled`, `interrupt_message`, `max_concurrent_threads_per_session`, `max_threads`, `default_subagent_model`, `default_subagent_reasoning_effort`; 커스텀 롤은 `[agents.<name>]`에 `config_file`(경로) + `description`
> **Precedence** — "Precedence resolves from explicit spawn values, then `[agents]` defaults, then parent values."

### (c) 현재 moai 배선과의 차이

**정합 1 — 경로와 필수 3필드는 일치.**
`internal/template/templates/.codex/agents/moai/*.toml` **11본**. 전수 확인 결과 모두 `name` / `description` / `developer_instructions`를 갖는다. 대표 헤더(`manager-lead.toml:1-6`):
```toml
# Generated by the MoAI agent dual-publication emitter; regenerate, do not edit.
# Neutral source: .claude/agents/moai/manager-lead.md (body carried verbatim).
name = "manager-lead"
description = '''
```
문서가 말하는 프로젝트 스코프 경로 `.codex/agents/`의 하위(`moai/`)에 있다.

**차이 1 — `model`을 발행하지 않는다.**
측정: `grep -l '^model = ' …/*.toml` → **0본**. 발행되는 것은 `model_reasoning_effort`와 `sandbox_mode`뿐. 문서상 이 경우 서브에이전트는 부모의 모델을 상속한다("If you don't configure a subagent model... the subagent inherits the parent agent's model"). 의도된 설계일 수 있으나 **명시적으로 기록된 근거를 찾지 못했다**(§8 Gaps). moai가 Claude 쪽에서는 스폰마다 모델을 주입하는 규율(`agent-common-protocol.md` § Per-Spawn Model Injection)을 갖고 있어 축이 비대칭이다.

**차이 2 — `mcp_servers`는 일부만.**
11본 중 `manager-develop` / `manager-docs` / `manager-lead` 3본이 에이전트 파일 안에 `[mcp_servers.moai]`를 품는다(`manager-develop.toml:228-230`). 문서가 허용하는 키이므로 형식은 적법하다. 나머지 8본이 왜 안 갖는지는 이 조사 범위 밖.

**차이 3 — `skills.config`를 에이전트 파일에 넣지 않는다.**
측정: 11본 전부 0히트. 문서는 에이전트 파일 안 `skills.config`를 허용한다 → **에이전트별 스킬 스코핑**이 미사용 표면으로 남아 있다.

**차이 4 — `[agents]` 전역 테이블 미배선.**
`enabled`, `max_threads`, `default_subagent_model`, `interrupt_message`, `[agents.<name>].config_file` 어느 것도 moai가 `.codex/config.toml`에 쓰지 않는다(§6 — `EnsureMCPTable`·`EnsureStatusLine` 두 표면만 쓴다).

**차이 5 — 11본 대 12본.** Claude 쪽 유지 카탈로그는 12(커스텀 11 + 빌트인 `Explore`). codex 쪽 11본은 커스텀 11과 정확히 대응하고, `Explore`는 codex 빌트인 `explorer`로 갈음 가능해 **보인다** — 다만 이는 추론이고 측정하지 않았다(§8 Gaps).

### (d) 구현 카드 후보
- **A1 (권고: 채택)** — `[agents]` 전역 테이블 발행. `default_subagent_model` / `default_subagent_reasoning_effort` / `max_concurrent_threads_per_session`은 moai가 이미 Claude 쪽에서 갖고 있는 정책(모델 프로파일, fanout 상한)의 codex 대응물이다. 배선 공백이 가장 큰 축.
- **A2 (권고: 채택)** — 에이전트별 `skills.config` 스코핑. moai는 에이전트마다 `skills:` 프론트매터를 이미 갖고 있어 소스가 존재한다.
- **A3 (권고: 조사 선행)** — `model` 미발행이 의도인지 누락인지 판정. 의도라면 emitter 매니페스트에 근거를 남기고, 누락이라면 모델 정책을 codex 쪽에도 실어야 한다.
- **A4 (권고: 보류)** — `[agents.<name>].config_file` 커스텀 롤 등록. 파일 자동 발견이 이미 되므로 추가 이득이 불분명하다.

---

## 6. 축 5 — MCP / statusline 최신 스키마

### (a) 공식 URL
- https://learn.chatgpt.com/docs/config-file/config-reference (본문 읽음)
- https://learn.chatgpt.com/docs/config-file/config-advanced (본문 읽음)
- https://developers.openai.com/plugins/build/plugins (본문 읽음)

### (b) 원문 인용

**`[mcp_servers.<id>]` 전체 필드**:
> `command`, `args`, `cwd`, `env` (map), `env_vars` (array), `url` (for HTTP servers), `bearer_token_env_var`, `http_headers`, `http_headers_helper`, `enabled`, `required`, `startup_timeout_sec`, `tool_timeout_sec`, `enabled_tools`, `disabled_tools`, `default_tools_approval_mode`, `tools.<tool>.approval_mode`, `tools.<tool>.output_token_limit`

**`[tui]`**:
> Terminal UI settings: `keymap.<context>.<action>`, `notifications`, `status_line` (or `tui.status_line` array), `theme`, `vim_mode_default`

**`notify`**:
> "Use `notify` to trigger an external program whenever Codex emits supported events (currently only `agent-turn-complete`)." · 형태: `notify = ["python3", "/path/to/notify.py"]`

**그 밖의 최상위 테이블(config-reference 열거)**: `[agents]`, `[mcp_servers.<id>]`, `[permissions.<name>]`, `[[skills.config]]`, `[hooks]`, `[tui]`, `[otel]`. 스칼라 키: `model`, `web_search`, `approval_policy`, `sandbox_mode`, `default_permissions`, `notify`, `log_dir`, `file_opener`, `personality`.

> **[정정 — t507이 실측으로 추가, 2026-09-07. 위 서술은 지우지 않는다]**
> 이 절은 매니페스트를 `.codex-plugin/plugin.json` 하나로만 적어 **두 층을 구분하지 않았다.** 실측(codex-cli 0.153.4, `.moai/reports/t507/verdict.md`)은 층이 갈린다는 것을 보였다:
>
> - **플러그인 매니페스트** = `.codex-plugin/plugin.json` — **수용된다.** 아래 인용은 이 층에서 정확하다.
> - **마켓플레이스 루트** = `.agents/plugins/marketplace.json` 또는 `.claude-plugin/marketplace.json`. **`.codex-plugin/`을 마켓플레이스 루트로 두면 거부된다** (`marketplace root does not contain a supported manifest`, EXIT=1).
>
> 즉 이 절만 읽고 `.codex-plugin/`에 마켓플레이스를 만들려 하면 **첫 명령에서 막힌다.** 아래 § 마켓플레이스 인용이 이미 `.agents/plugins/`를 적고 있으나, 두 경로가 서로 다른 층이라는 사실은 t507 전까지 이 문서에 없었다.

**플러그인 매니페스트** (`.codex-plugin/plugin.json`):
> "Only `plugin.json` belongs in `.codex-plugin/`. Keep `skills/`, `hooks/`, `assets/`, `.mcp.json`, and `.app.json` at the plugin root."
```json
{
  "name": "my-plugin", "version": "0.1.0", "description": "...",
  "skills": "./skills/", "mcpServers": "./.mcp.json",
  "apps": "./.app.json", "hooks": "./hooks/hooks.json",
  "interface": { "displayName": "...", "category": "...", "capabilities": ["Read","Write"] }
}
```
> **Marketplace file locations** — Repo: `$REPO_ROOT/.agents/plugins/marketplace.json` · Personal: `~/.agents/plugins/marketplace.json`
> ```bash
> codex plugin marketplace add owner/repo
> codex plugin marketplace add ./local-marketplace-root
> codex plugin marketplace list / upgrade / remove <name>
> ```
> 모든 매니페스트 경로는 `./`로 시작하고 플러그인 루트를 벗어나지 않아야 한다.

### (c) 현재 moai 배선과의 차이

**정합 1 — MCP 등록은 적법하고 보수적이다.**
`internal/codexwiring/configtoml.go:11-22` + `wire.go:110-117`이 다음을 **없을 때만** 쓴다:
```toml
[mcp_servers.moai]
command = "moai"
args = ["mcp-server"]
default_tools_approval_mode = "writes"
```
`writes`를 고른 근거가 소스 주석에 있다: 승인 집합이 서버의 `ReadOnlyHint` 어노테이션을 타지, 툴 이름 열거를 타지 않게 하려는 것. 문서 필드명과 일치.

**차이 1 — MCP 필드 18개 중 3개만 사용.**
미사용: `startup_timeout_sec`, `tool_timeout_sec`, `enabled`, `required`, `enabled_tools`, `disabled_tools`, `tools.<tool>.approval_mode`, `tools.<tool>.output_token_limit`, `env`, `cwd` 등. 특히 `tool_timeout_sec` 부재는 `moai mcp-server`의 장시간 툴(codex/glm 위임)에서 의미가 있을 수 있다.

**차이 2 — statusline 허용목록은 문서 근거가 없다.**
`configtoml.go:25-40`의 29토큰 허용목록은 **소스 열거형(openai/codex `StatusLineItem`)에서 뜬 것**이며 SPEC-CODEX-WIRING-001 §A.6에 기록됐다고 주석이 밝힌다. 이번 조사에서 **공식 문서는 이 토큰들을 열거하지 않았다** — config-advanced를 읽었으나 "does not enumerate valid status-line item identifiers or tokens"였다.
→ 즉 이 허용목록은 문서로 검증 불가하고, 주석 자체가 "no blind upstream tracking; additions are an explicit judgment at documentation-refresh time"라고 정책을 밝힌다. 결함이 아니라 **기록된 선택**이다. 코드상 허용목록 실측 개수 = **29**(주석의 29와 일치). 기본값 5토큰: `model-with-reasoning`, `context-remaining`, `git-branch`, `current-dir`, `thread-id`(`configtoml.go:44-47`).

**차이 3 — 플러그인 표면 전체가 미배선.**
- 측정: `git grep -lE 'codex-plugin' -- internal/ pkg/ cmd/` → **0히트**. 대조군 `codexwiring` → 22파일.
- 로컬 0.153.4에는 `codex plugin {add,list,marketplace,remove}`가 실재하고, `~/.codex/plugins/cache/` 아래 9개 마켓플레이스가 이미 깔려 있다(`moai-cowork` 포함). **표면은 살아 있고 moai만 안 쓰고 있다.**
- 플러그인은 스킬 + MCP + 훅을 **한 묶음으로 배포**하는 경로다. 지금 moai는 이 셋을 각각 다른 방법(경로 미러 / config.toml 편집 / hooks.json 편집)으로 사용자 트리에 밀어 넣는다. 플러그인은 그 셋을 하나의 설치 단위로 대체할 수 있는 유일한 공식 경로다.

**차이 4 — `notify` 미배선.** 지원 이벤트가 `agent-turn-complete` 하나뿐이라 이득이 작다.

**차이 5 — `[permissions.<name>]` / `[otel]` 미배선.** 조사 범위 밖이나 미사용 표면으로 기록해 둔다.

### (d) 구현 카드 후보
- **M1 (권고: 채택, 소)** — `[mcp_servers.moai]`에 `tool_timeout_sec` 추가. 위임 툴의 장시간 실행이 실제 문제였다면 근거가 있다(선행 확인 필요).
- **M2 (권고: 채택, 대) — 가장 큰 기회.** moai를 **codex 플러그인**으로 패키징. `.codex-plugin/plugin.json` + `skills/` + `hooks/hooks.json` + `.mcp.json` 한 묶음. 지금의 3중 침습(사용자 config.toml 편집·hooks.json 편집·심볼릭 링크 생성)을 **설치 단위 하나**로 대체할 수 있고, `$REPO_ROOT/.agents/plugins/marketplace.json`으로 자체 배포도 가능하다. 규모가 크므로 별도 SPEC 권장.
- **M3 (권고: 채택, 소)** — statusline 허용목록의 갱신 시점을 명시적 카드로 만들기. 지금은 "documentation-refresh time"이라는 정책만 있고 트리거가 없다.
- **M4 (권고: 보류)** — `notify`, `[permissions]`, `[otel]`.

---

## 7. 구현 카드 후보 종합

| 후보 | 축 | 권고 | 규모 | 선행 조건 |
|---|---|---|---|---|
| **M2** moai를 codex 플러그인으로 패키징 | 5 | 채택 | 대 | 없음 (별도 SPEC) |
| **A1** `[agents]` 전역 테이블 발행 | 4 | 채택 | 중 | 없음 |
| **S1** `skills.config` `path` 값 모양 실측 | 3 | 채택 | 소 | 없음 — S2의 선행 |
| **H2** 미adapted 5종 0.153.4 재측정 | 2 | 채택 | 중 | 없음 |
| **C1** `/moai` 16본의 codex 스킬 발행 | 1 | 채택 | 중 | S1 |
| **A2** 에이전트별 `skills.config` 스코핑 | 4 | 채택 | 중 | S1 |
| **H1** `Interrupt` 이벤트 행 추가 | 2 | 채택 | 소 | 없음 |
| **S4** 매니페스트 전수 재측정 → 0.153.4 | 3 | 채택 | 중 | H2·S1 |
| **A3** `model` 미발행 판정 | 4 | 조사 | 소 | 없음 |
| **S3** 유령 `skills.config` 정리 verb | 3 | 채택 | 소 | 없음 |
| **M1** `tool_timeout_sec` | 5 | 채택 | 소 | 문제 실재 확인 |
| **M3** statusline 허용목록 갱신 트리거 | 5 | 채택 | 소 | 없음 |
| **S2** 선택적 스킬 비활성화 발행 | 3 | 조건부 | 중 | S1 |
| **H3** 배경 훅(`async`) | 2 | 조건부 | 중 | H2 |
| **C2** `~/.codex/prompts` 발행 | 1 | **기각** | — | deprecated |
| **A4** `[agents.<name>].config_file` | 4 | 보류 | — | 이득 불분명 |
| **H4** `allow_managed_hooks_only` | 2 | 보류 | — | 수요 없음 |
| **M4** `notify` / `[permissions]` / `[otel]` | 5 | 보류 | — | 이득 작음 |

가장 짧은 경로: **H1 → S1 → A1**(전부 선행 조건 없음 또는 소규모). 가장 큰 이득: **M2**.

---

## 8. 판정서

### Claim
1. 리드 실측 4건은 이 트리에서 전부 재현된다.
2. codex의 커스텀 프롬프트(`~/.codex/prompts`)는 실재하나 **deprecated**이고 공식 대체는 **스킬**이다.
3. 공식 훅 이벤트는 **12종**이고 moai는 **11종**을 열거한다 — 누락은 `Interrupt` 하나.
4. `[[skills.config]]`는 **공식 문서화된 실재 키**(`path` + `enabled`)이며 **비활성화 수단**으로 기술된다. moai는 읽기만 하고 쓰지 않는다.
5. codex 스킬 탐색 경로는 `.codex/skills`가 아니라 **`.agents/skills`**이고, moai는 **이미 그 경로에 미러를 만든다**.
6. agents TOML 필수 3필드(`name`/`description`/`developer_instructions`)를 moai의 11본이 모두 충족한다. `model`·`skills.config`는 발행하지 않고 `[agents]` 전역 테이블은 미배선이다.
7. codex 플러그인 표면(`.codex-plugin/plugin.json` + 마켓플레이스)은 0.153.4에 실재하며 moai는 **전혀 쓰지 않는다**.

### Evidence
- 재현 명령과 출력은 §1 표 및 각 축의 (c)절에 명령·`file:line`과 함께 인용.
- 부재 주장 3건은 대조군을 동반: `M1 hits=0` vs `codex_task 115파일` · `EventInterrupt 0` vs `EventStop 검출` · `codex-plugin 0` vs `codexwiring 22파일`.
- 공식 인용 7페이지는 전부 `WebFetch`로 본문을 읽었다. 검색 요약만 있고 본문을 읽지 못한 페이지는 인용하지 않았다.
- 서브에이전트 인벤토리의 핵심 4건(MCP/statusline 상수, agents TOML 필드, `.agents/skills` 미러, 기록된 codex 버전)은 **내가 직접 재측정해 일치를 확인**했다. 재측정하지 않은 인벤토리 항목은 Gaps에 든다.

### Baseline-attribution
- 트리: `.claude/worktrees/t494`, HEAD `ace1c5440`, `git rev-list --count --left-right origin/develop...HEAD` = `0 0`.
- CLI: `codex --version` → `codex-cli 0.153.4`.
- 문서: 2026-09-07 조회본. 문서에 버전 스탬프 없음.

### Gaps — 관측하지 않은 것
1. **`skills.config`의 `path` 값 모양**(파일 vs 폴더). 두 공식 페이지가 서로 다르게 기술한다. 실측하지 않았다. → S1.
2. **미adapted 5종의 0.153.4 발화 여부.** 종전 관측은 0.152.1 이전 것이다. 재측정하지 않았다. → H2.
3. **`Interrupt`의 실제 발화 조건과 페이로드.** 문서 목록에 있다는 것만 확인했다.
4. **moai가 발행하는 훅 핸들러의 `matcher`/`async` 유무.** 근거를 찾지 못했다 — 없다고 단정하지 않는다.
5. **agents TOML에 `model`을 안 싣는 것이 의도인지.** 판정하지 않았다.
6. **codex 빌트인 `explorer`가 Claude `Explore`를 갈음하는지.** 추론일 뿐 측정하지 않았다.
7. **statusline 29토큰이 0.153.4 소스 열거형과 아직 일치하는지.** 공식 문서가 열거하지 않아 문서로는 검증 불가하고, 소스도 확인하지 않았다.
8. **서브에이전트 인벤토리 중 내가 재측정하지 않은 항목**: `internal/cli/hook.go:51-78`의 25 서브커맨드, `mcp_codex.go`의 codex 서브커맨드 호출 표, `codexadapter/output.go`·`stderr.go`·`diagnostics.go`의 동작 서술. 이들은 이 보고서의 결론을 지지하는 데 쓰이지 않았거나 보조적으로만 쓰였다.
9. **플러그인 설치가 moai의 기존 3중 침습을 실제로 대체 가능한지.** M2는 문서상 가능성이지 실증이 아니다.
10. **로컬 `~/.codex/config.toml`의 `notify = [..., "turn-ended"]`** — 문서는 지원 이벤트를 `agent-turn-complete` 하나로 적는데 실제 파일에는 `turn-ended`가 있다. 이 불일치는 조사하지 않았다(OpenAI 자체 앱이 쓴 값으로 보이나 확인하지 않음).

### Residual-risk
- 문서에 버전 스탬프가 없어 **0.153.4와 문서 사이 시차**를 배제할 수 없다. 문서가 앞설 수도, 뒤질 수도 있다.
- `learn.chatgpt.com`으로의 308 이전이 진행 중이라, 문서 구조가 다시 바뀌면 인용 URL이 깨질 수 있다. 원 주소(`developers.openai.com/codex/*`)를 함께 남긴 이유다.
- 로컬 `~/.codex/` 상태는 **이 머신 한 대의 관측**이다. `~/.codex/skills`가 존재하고 `~/.codex/prompts`가 없다는 사실은 이 머신의 상태이지 codex의 규약이 아니다 — 규약 판정은 문서 인용에 기댔다.
- 후보 표의 "규모"는 판단이지 측정이 아니다.

---

## Sources

- [Custom Prompts — learn.chatgpt.com/docs/custom-prompts](https://learn.chatgpt.com/docs/custom-prompts) (원 주소: https://developers.openai.com/codex/custom-prompts)
- [Build skills — learn.chatgpt.com/docs/build-skills](https://learn.chatgpt.com/docs/build-skills) (원 주소: https://developers.openai.com/codex/skills)
- [Hooks — learn.chatgpt.com/docs/hooks](https://learn.chatgpt.com/docs/hooks) (원 주소: https://developers.openai.com/codex/hooks)
- [Configuration Reference — learn.chatgpt.com/docs/config-file/config-reference](https://learn.chatgpt.com/docs/config-file/config-reference) (원 주소: https://developers.openai.com/codex/config-reference)
- [Advanced Configuration — learn.chatgpt.com/docs/config-file/config-advanced](https://learn.chatgpt.com/docs/config-file/config-advanced)
- [Subagents — learn.chatgpt.com/docs/agent-configuration/subagents.md](https://learn.chatgpt.com/docs/agent-configuration/subagents.md) (원 주소: https://developers.openai.com/codex/subagents.md)
- [Package your plugin — developers.openai.com/plugins/build/plugins](https://developers.openai.com/plugins/build/plugins)
- [Plugins — learn.chatgpt.com/docs/plugins](https://learn.chatgpt.com/docs/plugins) (원 주소: https://developers.openai.com/codex/plugins)
