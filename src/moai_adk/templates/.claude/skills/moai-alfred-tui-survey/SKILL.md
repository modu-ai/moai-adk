---

name: moai-alfred-tui-survey
description: Standardizes Claude Code Tools AskUserQuestion TUI menus for surveys, branching approvals, and option picking across Alfred workflows. Use when gathering approvals or decisions via Alfred’s TUI menus.
allowed-tools:
  - Read
  - Write
  - Edit
  - TodoWrite
---

# Alfred TUI Survey Skill

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), TodoWrite (todo_write) |
| Auto-load | On demand when AskUserQuestion menus are built |
| Trigger cues | Branch approvals, survey menus, decision gating via AskUserQuestion. |

## What it does

Provides ready-to-use patterns for Claude Code's AskUserQuestion TUI selector so Alfred agents can gather user choices, approvals, or survey answers with structured menus instead of ad-hoc text prompts.

## When to use

- Activates when Alfred needs to gather structured choices through the TUI selector.
- Need confirmation before advancing to a risky/destructive step.
- Choosing between alternative implementation paths or automation levels.
- Collecting survey-like answers (persona, tech stack, priority, risk level).
- Any time a branched workflow depends on user-selected options rather than free-form text.

## How it works

1. **Detect gate** – Pause at steps that require explicit user choice.
2. **Shape options** – Offer 2–5 focused choices with concise labels.
3. **Render menu** – Emit the `AskUserQuestion({...})` block for the selector.
4. **Map follow-ups** – Note how each option alters the next action/agent.
5. **Fallback** – Provide plain-text instructions when the UI cannot render.

### Templates & examples

- [Single-select deployment decision](examples.md#single-select-template)
- [Multi-select diagnostics checklist](examples.md#multi-select-variation)
- [Follow-up drill-down prompt](examples.md#follow-up-prompt-for-deeper-detail)

## Best Practices
- 질문·선택지·후속 조치를 한 화면에 담아 맥락 전환을 줄입니다.
- 옵션은 위험도·우선순위 등 비교 기준에 맞춰 정렬합니다.
- 승인/보류/취소처럼 안전 장치가 필요한 경우 경고 문구를 붙입니다.
- 제출 결과는 Sync 단계나 보고서에 재사용할 수 있도록 기록합니다.

## Inputs
- Alfred 워크플로우가 수집한 의사결정 시나리오와 후보 옵션.
- 옵션별 후속 행동(다음 커맨드, 서브 에이전트, TODO 등)에 대한 정의.

## Outputs
- AskUserQuestion 블록 및 선택지 → 후속 액션 매핑.
- 선택 결과 요약과 추가 확인이 필요한 TODO 항목.

## Failure Modes
- 모호하거나 중복된 옵션은 선택 지연을 야기합니다.
- 모든 옵션이 동일하게 보이면 우선순위 판단이 어렵습니다.
- TUI 비활성 환경에서는 AskUserQuestion이 동작하지 않아 플랜 B가 필요합니다.

## Dependencies
- 주요 `/alfred:*` 커맨드와 연동되며 필요 시 TodoWrite로 후속 조치를 예약합니다.
- moai-foundation-ears · moai-foundation-tags와 결합하면 요구사항 → 선택지 → TAG 기록이 연결됩니다.

## References
- Jakubovic, J. "Designing Effective CLI Dialogs." ACM Queue, 2021.
- NCurses. "Programming Guide." https://invisible-island.net/ncurses/man/ncurses.3x.html (accessed 2025-03-29).

## Changelog
- 2025-03-29: Alfred 전용 스킬에 입력/출력/실패 대응을 추가했습니다.

## Works well with

- `moai-foundation-ears` – Combine structured requirement patterns with menu-driven confirmations.
- `moai-alfred-git-workflow` – Use menus to choose branch/worktree strategies.
- `moai-alfred-code-reviewer` – Capture reviewer focus areas through guided selection.

## Examples
```markdown
- `/alfred:1-plan` 단계에서 사용자 우선순위를 수집한 뒤 결과를 PLAN 보드에 기록합니다.
- `/alfred:2-run` 실행 중 위험도가 높은 작업 전 사용자 승인을 확인합니다.
```
