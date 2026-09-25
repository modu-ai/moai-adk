---
id: SPEC-DOCS-OTEL-AGENTS-001
plan_version: "0.2.0"
created: 2026-09-25
updated: 2026-09-25
author: manager-spec
---

# Implementation Plan — SPEC-DOCS-OTEL-AGENTS-001

## §A Context

카드 t1181. Claude Code 2.1.281·2.1.282 이후 낡거나 범위가 모호해진 docs-site 서술 세 곳(결함 A·B·C, spec.md §A)을 네 로케일에서 한 커밋으로 고친다. 코드·템플릿 변경은 없다. 브랜치 `WT-docs-otel-env`, base `a520187f1`(= 로컬 develop). Tier M(대상 9개 파일 — spec.md 프론트매터 아래 근거 참조).

## §B Known Issues

- ko `settings-json.md` 는 이미 재작성되어 en/ja/zh 와 본문 구조가 다르다(272행 대 1,100행대). 이 카드는 그 차이를 줄이지 않는다 — 문제의 행 하나만 지운다(spec.md Out of Scope).
- ja/zh `commands.md` 에는 `### /agents` 절이 없고, description(5행)에도 `/agents` 가 없다. 결함 B 의 절·description 판단은 ko/en 에만 해당한다.
- zh 는 파일마다 용어가 다르다: `commands.md` 는 「子智能体」, `sub-agents.md` 는 「子代理」. 각 파일의 기존 용어를 따른다.
- 표 행의 버전 열 `v2.1.139+` 는 업스트림 CHANGELOG(`/agents` 첫 등장 `## 1.0.60`)와 맞지 않을 수 있으나 이 카드에서 검증·수정하지 않는다(spec.md Out of Scope — 알려진 한계).

## §C Pre-flight (run-phase 진입 점검)

1. `git branch --show-current` → `WT-docs-otel-env`, `git status --short` → 빈 출력.
2. 아래 M2 가 지목하는 BEFORE 행이 현재 트리에 그대로 있는지 `grep -F` 로 확인한다(develop 흡수로 행 번호가 움직였을 수 있으므로 행 번호가 아니라 내용으로 찾는다).
3. acceptance.md §A 의 `after` 헬퍼로 M1 의 AFTER 블록 10개가 각각 비지 않은 한 줄로 추출되는지 확인한다.
4. base 빌드 기준선 재측정: `hugo --source docs-site --minify --gc --destination <scratch>` → exit 0, `grep -c WARN` → 0.

## §D Constraints

- 편집은 내용 기준 정확 문자열 치환(Edit)으로 한다. 행 번호 기반 sed 금지.
- 아홉 파일 모두 한 커밋에 담는다. 커밋 메시지에 카드 id `t1181` 을 넣는다.
- `docs-site/` 아래에서는 아홉 파일 밖을 건드리지 않는다. SPEC 진행 기록(`progress.md`)과 증거(`.moai/reports/t1181/verdict.md`)는 `docs-site/` 밖이므로 이 제약과 무관하다.
- Hugo 빌드 출력은 트리 밖(스크래치 디렉터리)으로 보낸다 — `docs-site/public/` 에 쓰지 않는다.
- 배포·push 금지.

## §E Self-Verification

acceptance.md 의 AC-DOA-001~009 명령을 한 턴에 병렬로 실행하고 출력 원문을 `progress.md` §E.2 에 싣는다.

## §F Milestones (우선순위 순, 시간 추정 없음)

### M1 — 문구 결정 (Priority: Critical — 가장 바뀌기 쉬운 결정)

아래 AFTER 문구가 이 SPEC 의 결정이다. 감사·운영자 검토는 이 절에 집중한다.

**결정 D1 — 텔레메트리 행은 삭제만 한다(절충을 알고 택함).** 이 행은 거짓이 아니라 범위가 모호하다. `CLAUDE_CODE_ENABLE_TELEMETRY` 는 `~/.claude/settings.json`·managed settings·셸에서는 여전히 유효하고, 프로젝트 `.claude/settings.json`·`.claude/settings.local.json` 에서만 2.1.282 부터 무시된다(spec.md §A.1 사실 1·2). 따라서 삭제는 사용자 범위에서는 참인 정보를 지우는 선택이다. 그래도 삭제를 택하는 이유는 로케일 간 일관성이다 — ko 판에는 이 표 자체가 없으므로, 범위 한정 주석을 붙인 행을 en/ja/zh 에만 남기면 네 로케일의 내용 차이가 더 벌어진다. 대체 행·산문도 같은 이유로 넣지 않는다.

**결정 D2 — ko/en description(5행)은 `/agents` 를 유지한다.** `/agents` 는 공식 명령 문서에 여전히 실린 명령이고, 입력하면 안내를 출력한다. 본문 절이 그 동작을 정확히 서술하도록 고치므로 요약 줄에서 지울 이유가 없다.

**결정 D3 — ko/en 절 제목을 바꾼다.** `commands#agents...` 앵커를 거는 링크가 docs-site 에 0건이므로(spec.md §A.4) 제목 변경은 링크를 깨지 않는다. ko 제목은 `/agents` 가 이제 안내만 띄운다는 사실을 제목만 읽어도 알 수 있게 쓴다.

**AFTER 블록 형식.** 각 AFTER 는 `<!-- after:<key> -->` 표지 바로 뒤의 ```` ```text ```` 펜스 안에 **정확히 한 줄**로 둔다. 그 한 줄이 대상 파일의 한 행 전체를 대체한다. 펜스 바깥의 백틱은 AFTER 의 일부가 아니며, 펜스 안의 백틱은 모두 AFTER 의 일부다. acceptance.md AC-DOA-007 이 이 블록을 기계적으로 추출해 대상 파일과 행 단위로 정확히 대조한다.

**B-row (commands.md `/agents` 표 행 전체):**

<!-- after:ko-row -->
```text
| `/agents` | 서브에이전트 안내 — v2.1.198부터 마법사 대신 Claude에게 요청하거나 `.claude/agents/`를 직접 편집하라는 안내만 표시, v2.1.281부터 명령 메뉴와 `/help`에서 빠짐 | v2.1.139+ |
```

<!-- after:en-row -->
```text
| `/agents` | Subagent pointer — since v2.1.198 prints a reminder to ask Claude or edit `.claude/agents/` directly instead of opening a wizard; hidden from the command menu and `/help` since v2.1.281 | v2.1.139+ |
```

<!-- after:ja-row -->
```text
| `/agents` | サブエージェントの案内 — v2.1.198 以降はウィザードを開かず、Claude に依頼するか `.claude/agents/` を直接編集するよう案内のみ表示。v2.1.281 以降はコマンドメニューと `/help` に表示されない | v2.1.139+ |
```

zh 는 같은 파일의 기존 행 표기(` — `)를 따른다.

<!-- after:zh-row -->
```text
| `/agents` | 子智能体指引 — 自 v2.1.198 起不再打开向导，只提示让 Claude 代劳或直接编辑 `.claude/agents/`；自 v2.1.281 起不再出现在命令菜单和 `/help` 中 | v2.1.139+ |
```

**B-section (ko/en `commands.md` 의 `### /agents` 제목 행, 그리고 제목 다음 문단 한 행 전체):**

<!-- after:ko-head -->
```text
### /agents — 이제는 안내만 띄우는 명령
```

<!-- after:ko-para -->
```text
`/agents`는 이제 서브에이전트를 만들거나 관리하는 화면을 열지 않습니다. v2.1.198부터는 입력하면 Claude에게 요청하거나 `.claude/agents/`(개인용은 `~/.claude/agents/`)를 직접 편집하라는 안내만 띄우고, v2.1.281부터는 명령 메뉴와 `/help` 목록에서도 빠졌습니다. 새 서브에이전트를 만드는 길은 두 가지입니다.
```

<!-- after:en-head -->
```text
### /agents — Where Subagent Creation Went
```

<!-- after:en-para -->
```text
`/agents` no longer opens a screen for creating or managing subagents. Since v2.1.198, typing it only prints a reminder to ask Claude or to edit `.claude/agents/` (or `~/.claude/agents/` for personal ones) directly, and since v2.1.281 it is hidden from the command menu and `/help`. There are two ways to create a new subagent.
```

두 갈래 목록(1·2번)과 그 뒤 두 문단(정의 파일 안내, 백그라운드·중첩 스폰 문단)은 그대로 둔다.

**C (ja/zh `sub-agents.md` 의 서브에이전트 정의 문단 한 행 전체) — ko/en :120 과 같은 뜻:**

<!-- after:ja-sub -->
```text
サブエージェントは YAML フロントマターを持つマークダウンファイルで定義します。Claude に作成を依頼することも、ファイルを直接書くこともできます。v2.1.198 以降、`/agents` コマンドは対話式の作成ウィザードを開かず、Claude に依頼するか `.claude/agents/` を直接編集するよう案内するだけです（ファイル形式と保存場所は変わりません）。
```

<!-- after:zh-sub -->
```text
子代理通过带有 YAML 前置元数据的 Markdown 文件来定义。既可以请 Claude 创建，也可以直接手写文件。自 v2.1.198 起，`/agents` 命令不再打开交互式创建向导，只提示让 Claude 代劳或直接编辑 `.claude/agents/`（文件格式和存放位置不变）。
```

**AFTER 가 함의하는 커밋 단위 변경량(`git show --numstat`, 추가·삭제 순).** 위 AFTER 를 base `a520187f1` 사본에 적용해 잰 값(spec 작성 실행, 스크래치 사본):

| 파일 | numstat |
|---|---|
| en/ja/zh `advanced/settings-json.md` | `0 1` |
| ko/en `claude-code/foundations/commands.md` | `3 3` (표 행·제목·문단) |
| ja/zh `claude-code/foundations/commands.md` | `1 1` |
| ja/zh `claude-code/agentic/sub-agents.md` | `1 1` |

### M2 — 편집 적용 (Priority: High)

1. 결함 A: en/ja/zh `settings-json.md` 에서 `CLAUDE_CODE_ENABLE_TELEMETRY` 행 하나를 지운다. 앞뒤 행(`CLAUDE_AUTOCOMPACT_PCT_OVERRIDE`, `CLAUDE_CODE_DISABLE_BACKGROUND_TASKS`)은 건드리지 않는다.
2. 결함 B: 네 로케일 표 행(``| `/agents` |`` 로 시작하는 행)을 `<l>-row` 로 바꾸고, ko/en 의 `### /agents` 제목 행을 `<l>-head` 로, 그다음 문단 행(백틱으로 감싼 `/agents` 뒤에 ko 는 「는 세션 안에서」, en 은 「 is a command to inspect」가 이어지는 행)을 `<l>-para` 로 바꾼다.
3. 결함 C: ja/zh 서브에이전트 정의 문단 행(ja `サブエージェントは YAML フロントマター`, zh `子代理通过带有 YAML` 로 시작)을 `<l>-sub` 로 바꾼다.

### M3 — 검증·커밋 (Priority: High)

아홉 파일을 명시적 pathspec 으로 스테이징 → 한 커밋(`docs(t1181): ...`) → acceptance.md AC 배치를 한 턴에 실행(AC-DOA-001·008 은 커밋이 있어야 판정된다) → 증거를 `.moai/reports/t1181/verdict.md` 와 `progress.md` §E.2 에 기록. push 하지 않는다.

## §G Anti-Patterns

- ko 판을 기준으로 en/ja/zh 본문을 대량 재작성하는 것.
- 낡은 단서를 지우는 대신 "2026-09 기준" 같은 새 날짜 단서로 갈아 끼우는 것 — 같은 부패가 반복된다.
- AFTER 펜스의 바깥 백틱까지 옮기거나 안쪽 백틱을 빠뜨리는 것 — AC-DOA-007 의 행 전체 대조가 실패한다.
- `git add -A` / `git add .` 로 스테이징하는 것.
- Hugo 출력을 `docs-site/public/` 에 남기는 것.

## §H Cross-references

- spec.md §A.1 (공식 출처 4건), §C (대상 파일 9개)
- acceptance.md §D (AC-DOA-001~009)
- `.moai/docs/docs-site-i18n-rules.md`
