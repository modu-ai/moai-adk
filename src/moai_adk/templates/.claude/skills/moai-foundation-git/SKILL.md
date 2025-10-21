---

name: moai-foundation-git
description: Git workflow automation (branching, TDD commits, PR management). Use when standardizing Git practices across the project.
allowed-tools:
  - Read
  - Bash
  - Write
  - Edit
  - TodoWrite
---

# Alfred Git Workflow

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | SessionStart (foundation bootstrap) |
| Trigger cues | Branch creation, commit convention, PR readiness, and release gating requests. |

## What it does

Automates Git operations following MoAI-ADK conventions: branch creation, locale-based TDD commits, Draft PR creation, and PR Ready transition.

## When to use

- Activates when Git workflow automation is needed for branching, commits, or PR promotion.
- “Create branch”, “Create PR”, “Create commit”
- Automatically invoked by `/alfred:1-plan`, `/alfred:2-run`, `/alfred:3-sync`
- Git workflow automation needed

## How it works

**1. Branch Creation**:
```bash
git checkout develop
git checkout -b feature/SPEC-AUTH-001
```

**2. Locale-based TDD Commits**:
- **Korean (ko)**: 🔴 RED: [Test Description]
- **English (en)**: 🔴 RED: [Test description]
- **Japanese (ja)**: 🔴 RED: [テスト説明]
- **Chinese (zh)**: 🔴 RED: [测试说明]

Configured via `.moai/config.json`:
```json
{"project": {"locale": "ko"}}
```

**3. Draft PR Creation**:
Creates Draft PR with SPEC reference and test checklist.

**4. PR Ready Transition** (via `/alfred:3-sync`):
- Updates PR from Draft → Ready
- Adds quality gate checklist
- Verifies TRUST 5-principles

## Examples
```markdown
- 표준 문서를 스캔하여 누락 섹션을 보고합니다.
- 변경된 규약을 CLAUDE.md에 반영합니다.
```

## Best Practices
- 표준 변경 시 변경 사유와 근거 문서를 함께 기록합니다.
- 단일 소스 원칙을 지켜 동일 항목을 여러 곳에서 수정하지 않도록 합니다.

## Inputs
- 프로젝트 표준 문서(예: `CLAUDE.md`, `.moai/config.json`).
- 관련 서브 에이전트의 최신 출력물.

## Outputs
- MoAI-ADK 표준에 맞는 템플릿 또는 정책 요약.
- 재사용 가능한 규칙/체크리스트.

## Failure Modes
- 필수 표준 파일이 없거나 접근 권한이 제한된 경우.
- 상충하는 정책이 감지되어 조정이 필요할 때.

## Dependencies
- cc-manager와 함께 호출될 때 시너지가 큽니다.

## References
- Vincent Driessen. "A successful Git branching model." https://nvie.com/posts/a-successful-git-branching-model/ (accessed 2025-03-29).
- GitHub Docs. "GitHub Flow." https://docs.github.com/en/get-started/using-github/github-flow (accessed 2025-03-29).

## Changelog
- 2025-03-29: Foundation 스킬 템플릿을 베스트 프랙티스 구조에 맞게 보강했습니다.
