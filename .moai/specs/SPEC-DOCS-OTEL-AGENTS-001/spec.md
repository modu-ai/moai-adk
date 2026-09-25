---
id: SPEC-DOCS-OTEL-AGENTS-001
title: "docs-site: drop the project-settings telemetry row and restate /agents as a reminder-only command (CC 2.1.281 / 2.1.282)"
version: "0.2.0"
status: in-progress
created: 2026-09-25
updated: 2026-09-25
author: manager-spec
priority: P2
phase: "v3.2.0 target"
module: "docs-site/content"
lifecycle: spec-anchored
tags: "docs-site, i18n, claude-code-upstream, telemetry, opentelemetry, slash-commands, subagents, documentation-only"
tier: M
card: t1181
---

## HISTORY

| Date | Author | Change |
|------|--------|--------|
| 2026-09-25 | manager-spec | 최초 작성 — 카드 t1181 (Class C, Tier S, docs 전용) plan-phase 산출물. 근거 조사: `.moai/research/cc-update-2.1.274-to-2.1.282.md` T1-1·항목 t·카드 후보 C1. 드리프트는 이 트리(base `a520187f1`)에서 grep 으로 실측했다(§A.4). |
| 2026-09-25 | manager-spec | 0.2.0 — plan-audit iter-1 FAIL(0.74) 반영. Tier S→M 승격(아래 Tier 근거), §A.2 전제를 "범위 모호"로 정정, REQ-DOA-004·007 문언 조정, ko 절 제목 AFTER 교체. acceptance.md 는 AC-002 를 AC-001 에 흡수하고 AFTER 행 단위 정확 일치 AC 를 더해 9개로 재번호, 카드 커밋 단위 판정으로 바꿨다. |

**Tier 근거:** Tier M — `.moai/config/sections/harness.yaml:76` 이 Tier S 를 5개 파일 미만으로 정의하는데, 이 SPEC 은 docs-site 파일 9개를 고친다.

## §A 배경과 문제

Claude Code 2.1.281 과 2.1.282 이후 docs-site 네 로케일에 실린 서술 두 가지가 문제가 되었다. 텔레메트리 변수 행은 범위가 모호해져 가장 먼저 떠올릴 해석에서 틀리게 되었고(결함 A), `/agents` 서술은 현재 동작과 어긋난다(결함 B·C). 두 사실 모두 이 SPEC 작성 시점에 오케스트레이터가 공식 원문을 WebFetch 로 확인했고, 아래 인용은 원문 그대로다.

### §A.1 업스트림 사실 (공식 출처)

1. **프로젝트·로컬 설정의 텔레메트리 변수 무시 (2.1.282).** <https://raw.githubusercontent.com/anthropics/claude-code/main/CHANGELOG.md> § `## 2.1.282`:
   > "Changed project and local settings to ignore OpenTelemetry variables that turn on export, set its endpoint, or capture content, like `CLAUDE_CODE_ENABLE_TELEMETRY` and `OTEL_LOG_*`"
2. **허용 범위.** <https://code.claude.com/docs/en/monitoring-usage>:
   > "Claude Code ignores the OpenTelemetry exporter variables in a repository's `.claude/settings.json` and `.claude/settings.local.json`, so a repository can't use them to turn telemetry on, choose where it goes, or capture content. Set them in managed settings, or have each developer set them in their shell or `~/.claude/settings.json`."
3. **`/agents` 메뉴 항목 제거 (2.1.281).** 같은 CHANGELOG § `## 2.1.281`:
   > "Removed the leftover \"(removed)\" `/agents` entry from the command menu and `/help`; typing `/agents` still explains where the wizard went"
4. **`/agents` 현재 동작.** <https://code.claude.com/docs/en/commands> 의 `/agents` 행:
   > "As of v2.1.198, running `/agents` prints a reminder to ask Claude to create or manage subagents, or to edit `.claude/agents/` or `~/.claude/agents/` directly. On v2.1.197 and earlier, opens an interactive interface for creating and managing subagent configurations"

### §A.2 결함 A — 텔레메트리 변수 행 (en/ja/zh)

`docs-site/content/{en,ja,zh}/advanced/settings-json.md:763` 의 "Key Environment Variable Reference" 표는 `settings.json` `env` 예시 아래에 `CLAUDE_CODE_ENABLE_TELEMETRY` | `"1"` | OpenTelemetry 수집 활성화 행을 싣고 있다. 이 행은 거짓이 아니라 **범위가 모호한** 행이다. 문서는 `settings.json` 을 특정 범위로 한정하지 않고(en L87 은 "global settings file" 로 소개하고, 앞 절에서 네 범위를 모두 다룬다), 이 변수는 `~/.claude/settings.json`·managed settings·셸에서는 지금도 유효하다(사실 2). 그러나 MoAI 사용자가 가장 먼저 따를 해석 — 저장소의 프로젝트 `.claude/settings.json`(또는 `.claude/settings.local.json`)에 넣는 것 — 은 2.1.282 부터 무시된다(사실 1·2). 이 행을 지우면 사용자 범위에서는 참인 정보도 함께 빠진다는 절충은 plan.md 결정 D1 이 다룬다. ko 판은 이미 재작성되어 이 표가 없다 — docs-site 전체에서 `CLAUDE_CODE_ENABLE_TELEMETRY`·`OTEL_` 는 이 세 행에만 나온다.

### §A.3 결함 B·C — `/agents` 서술

- **B (commands.md, 4 로케일).** 내장 명령 표의 `/agents` 행(ko/en :45, ja/zh :46)이 `/agents` 를 서브에이전트 "관리(UI)" 명령으로 설명한다. en/ja/zh 는 여기에 "공식 문서에는 2026-07 시점 탭 UI 가 아직 남아 있다"는 단서를 덧붙이는데, 공식 명령 문서는 이제 안내 문구만 출력하는 동작을 기술하므로(사실 4) 이 단서는 거짓이다. ko/en 의 `### /agents` 절(:123-125)은 첫 문장에서 `/agents` 를 서브에이전트를 "살펴보는/inspect" 명령이라 부르는데, v2.1.198 이후 이 명령은 안내만 띄운다. 게다가 v2.1.281 부터는 명령 메뉴와 `/help` 에서도 빠졌다(사실 3).
- **C (sub-agents.md, ja/zh 형제).** `ja/.../sub-agents.md:102`, `zh/.../sub-agents.md:100` 은 서브에이전트를 "`/agents` 명령으로 대화형 생성"할 수 있다고 쓰고, "공식 문서에는 2026-07 시점 `/agents` 인터페이스가 남아 있으니 실제 2.1.198 세션에서 확인하라"는 단서를 단다. 사실 3·4 와 모순이다. 같은 문장의 ko :120·en :120 은 이미 옳다.

### §A.4 실측 기준선 (base `a520187f1`, 이 실행)

- `grep -rn 'CLAUDE_CODE_ENABLE_TELEMETRY\|OTEL_' docs-site/content` → 3행(en/ja/zh `settings-json.md:763`).
- `grep -c '2026-07'`: commands.md ko 0 / en 1 / ja 1 / zh 1, sub-agents.md ko 0 / en 0 / ja 1 / zh 1 — 모두 `/agents` 행·문장에 있다.
- ja/zh `sub-agents.md` 에서 `/agents` 가 있는 행의 `v2.1.198` 계수 → ja 0 / zh 0(해당 문장은 `v` 없는 `CC 2.1.198` 을 쓴다). ja 파일 전체의 `v2.1.198` 계수는 4 — 다른 행에 있다.
- 옛 제목(`^### /agents — 서브에이전트 관리$` ko 1, `^### /agents — Managing Subagents$` en 1)과 옛 표 행 서술자(ko `서브에이전트 관리 (v2.1.198`, en `Manage subagent configuration`, ja `サブエージェント管理 UI`, zh `子智能体管理 UI` 각 1).
- ja/zh commands.md 에는 `### /agents` 절이 없다(표 행만 있음). 두 파일의 description(5행)에도 `/agents` 가 없다.
- `/agents` 를 다루는 다른 docs-site 페이지는 없다(`commands.md`·`sub-agents.md` 밖 grep 0건). `commands#agents` 앵커를 거는 링크도 0건이다.
- docs-site Hugo 빌드(`hugo --source docs-site --minify --gc`, 출력은 스크래치 디렉터리): exit 0, `WARN` 0건.

## §B 요구사항 (GEARS)

### REQ-DOA-001 — 텔레메트리 변수 행 제거 (Unwanted)

The en/ja/zh `advanced/settings-json.md` guides shall not list `CLAUDE_CODE_ENABLE_TELEMETRY` in the environment-variable table placed beneath the `settings.json` `env` example, and no replacement row or prose shall be added in its place.

### REQ-DOA-002 — `/agents` 명령표 행 정정 (Ubiquitous)

The `/agents` row of the built-in command table in `claude-code/foundations/commands.md` shall, in each of the four locales (ko/en/ja/zh), state that since v2.1.198 `/agents` prints a reminder (ask Claude, or edit `.claude/agents/` directly) instead of opening a wizard, and that since v2.1.281 it is hidden from the command menu and `/help`.

### REQ-DOA-003 — 낡은 공식 문서 단서 제거 (Unwanted)

The touched `/agents` lines in `commands.md` (en/ja/zh) and `sub-agents.md` (ja/zh) shall not carry the "official docs still document the `/agents` interface as of 2026-07" clause or its locale equivalents.

### REQ-DOA-004 — ko/en `/agents` 절 정정 (Ubiquitous)

The ko/en `### /agents` section of `commands.md` shall carry a heading that no longer presents `/agents` as a subagent-management command and a lead paragraph that describes `/agents` as reminder-only, and shall preserve the existing two-ways-to-create list and the background-execution / nested-spawn paragraph unchanged.

### REQ-DOA-005 — ja/zh 서브에이전트 문장 정렬 (Ubiquitous)

The ja/zh `claude-code/agentic/sub-agents.md` sentence on creating subagents shall carry the same meaning as the ko/en sentence: a subagent is created by asking Claude or by writing the file directly, and since v2.1.198 `/agents` only shows a reminder.

### REQ-DOA-006 — 단일 커밋·경고 없는 빌드 (Event-driven)

When the run-phase edit lands, the four-locale edits shall land in a single commit, and the docs-site Hugo build shall complete with exit 0 and zero `WARN` lines.

### REQ-DOA-007 — 배포·원격 반영 금지 (Unwanted)

The run phase shall not deploy docs-site, push any branch, or edit any file under `docs-site/` other than the nine target files listed in §C.

## §C 검증 가능한 범위 요약

수정 대상은 정확히 아홉 파일이다.

| 결함 | 파일 | 로케일 |
|---|---|---|
| A | `docs-site/content/<l>/advanced/settings-json.md` | en, ja, zh |
| B | `docs-site/content/<l>/claude-code/foundations/commands.md` | ko, en, ja, zh |
| C | `docs-site/content/<l>/claude-code/agentic/sub-agents.md` | ja, zh |

카드 증거 경로: `.moai/reports/t1181/verdict.md`.

## Out of Scope

### Out of Scope — 로케일 간 본문 재정렬

- ko `settings-json.md` 재작성본을 정본으로 삼아 en/ja/zh 본문 전체를 다시 맞추는 일. 이 카드는 범위가 모호해진 한 행만 지운다.
- en/ja/zh 에만 대체 산문(텔레메트리는 셸·`~/.claude/settings.json`·managed settings 에 두라는 안내)을 추가하는 일 — ko 에 없는 내용을 세 로케일에만 넣으면 로케일 간 차이가 더 벌어진다.
- ja/zh `commands.md` 에 ko/en 과 같은 `### /agents` 절을 새로 만드는 일.

### Out of Scope — docs-site 밖의 표면

- 템플릿(`internal/template/templates/**`)·로컬 `.claude/**`·README 수정. 조사 문서 N-1 에서 템플릿과 로컬 설정에 텔레메트리 변수가 0건임이 이미 확인됐다.
- docs-site 배포(Vercel), 브랜치 push, PR 생성.
- Mermaid·아이콘 shortcode·디자인 컴포넌트 변경.

### Out of Scope — 알려진 한계 (이 카드에서 다루지 않음)

- `/agents` 표 행의 버전 열 `v2.1.139+` 는 검증하지 않고 그대로 이어받는다. 업스트림 CHANGELOG 는 `/agents` 첫 등장을 `## 1.0.60` 에 두므로 틀렸을 수 있으며, 고치려면 별도 카드가 필요하다.
- `scripts/docs-i18n-check.sh`(develop push 때 warn-only 로 도는 `docs-i18n-check.yml`)는 인수 기준에 넣지 않았다. 이 편집은 파일 수·title·H1·용어집을 건드리지 않는다.

## Cross-references

- 조사 근거: `.moai/research/cc-update-2.1.274-to-2.1.282.md` § T1-1, 항목 t, 카드 후보 C1 (primary 체크아웃에만 있는 미추적 파일 — 읽기 전용 인용)
- docs-site 4 로케일 규칙: `.moai/docs/docs-site-i18n-rules.md`
- 인수 기준: `acceptance.md` / 계획: `plan.md` / 진행 기록: `progress.md`
