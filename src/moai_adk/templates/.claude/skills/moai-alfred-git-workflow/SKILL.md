---

name: moai-alfred-git-workflow
description: Automates Git operations with MoAI-ADK conventions (feature branch, locale-based TDD commits, Draft PR, PR Ready transition). Use when orchestrating GitFlow checkpoints, commits, or PR transitions.
allowed-tools:
  - Read
  - Write
  - Edit
  - Bash
  - TodoWrite
---

# Alfred Git Workflow

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file), Bash (terminal), TodoWrite (todo_write) |
| Auto-load | /alfred:2-run Git automation |
| Trigger cues | Branch provisioning, commit batching, draft PR preparation within Alfred flows. |

## What it does

Automates Git operations following MoAI-ADK conventions: branch creation, locale-based TDD commits, Draft PR creation, and PR Ready transition.

## When to use

- Activates when Alfred must manage branches, commits, or PR transitions.
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

## Best Practices
- 사용자에게 보여주는 문구는 TUI/보고서용 표현으로 작성합니다.
- 도구 실행 시 명령과 결과 요약을 함께 기록합니다.

## Examples
```markdown
- /alfred 커맨드 내부에서 이 스킬을 호출해 보고서를 생성합니다.
- Completion Report에 요약을 추가합니다.
```

## Inputs
- MoAI-ADK 프로젝트 맥락 (`.moai/project/`, `.claude/` 템플릿 등).
- 사용자 명령 또는 상위 커맨드에서 전달한 파라미터.

## Outputs
- Alfred 워크플로우에 필요한 보고서, 체크리스트 또는 추천 항목.
- 후속 서브 에이전트 호출을 위한 구조화된 데이터.

## Failure Modes
- 필수 입력 문서가 없거나 권한이 제한된 경우.
- 사용자 승인 없이 파괴적인 변경이 요구될 때.

## Dependencies
- cc-manager, project-manager 등 상위 에이전트와 협력이 필요합니다.

## References
- Vincent Driessen. "A successful Git branching model." https://nvie.com/posts/a-successful-git-branching-model/ (accessed 2025-03-29).
- GitHub Docs. "GitHub Flow." https://docs.github.com/en/get-started/using-github/github-flow (accessed 2025-03-29).

## Changelog
- 2025-03-29: Alfred 전용 스킬에 입력/출력/실패 대응을 추가했습니다.

## Works well with

- alfred-ears-authoring (SPEC ID-based branch naming)
- alfred-trust-validation (PR Ready quality check)
