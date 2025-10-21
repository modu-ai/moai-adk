---

name: moai-claude-code
description: Scaffolds and audits Claude Code agents, commands, skills, plugins, and settings with production templates. Use when configuring or reviewing Claude Code automation inside MoAI workflows.
allowed-tools:
  - Read
  - Write
  - Edit
---

# MoAI Claude Code Manager

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Write (write_file), Edit (edit_file) |
| Auto-load | SessionStart (Claude Code bootstrap) |
| Trigger cues | Agent/command/skill/plugin/settings authoring, Claude Code environment setup. |

Create and manage Claude Code's five core components according to official standards.

## Components covered

- **Agents** `.claude/agents/` — Persona, tools, and workflow definition
- **Commands** `.claude/commands/` — Slash command entry points
- **Skills** `.claude/skills/` — Reusable instruction capsules
- **Plugins** `settings.json › mcpServers` — MCP integrations
- **Settings** `.claude/settings.json` — Tool permissions, hooks, session defaults

## Reference files

- `reference.md` — 작성 가이드와 체크리스트
- `examples.md` — 완성된 아티팩트 샘플
- `templates/` — 다섯 구성요소용 마크다운/JSON 골격
- `scripts/` — settings 검증 및 템플릿 무결성 검사 스크립트

## Workflow

1. 사용자 요청을 분석해 필요한 구성요소(Agent/Command/Skill/Plugin/Settings)를 결정합니다.
2. `templates/`에서 해당 스텁을 복사하고 프로젝트 맥락에 맞춰 placeholder를 교체합니다.
3. 필요 시 `scripts/` 검증기를 실행해 필수 필드·권한·후속 링크를 확인합니다.

## Guardrails

- Anthropic 공식 가이드라인에 맞춰 최소 권한과 progressive disclosure를 유지합니다.
- 템플릿을 직접 수정하기보다 reference.md에서 안내한 hook/field만 업데이트합니다.
- 생성된 파일과 settings.json은 Git 버전 관리에 포함시키고 변경 이력을 남깁니다.

**Official documentation**: https://docs.claude.com/en/docs/claude-code/skills  
**Version**: 1.0.0

## Examples
```markdown
- 새 프로젝트에서 spec-builder 에이전트와 /alfred 명령 세트를 생성합니다.
- 기존 settings.json을 검토해 허용 도구와 훅 구성을 업데이트합니다.
```

## Best Practices
- 출력 템플릿은 재적용 시에도 안전하도록 idempotent하게 설계합니다.
- 자세한 절차는 reference.md(작성 가이드)와 examples.md(샘플 아티팩트)로 분리해 필요할 때만 로드합니다.

## When to use
- Activates when someone asks to scaffold or audit Claude Code components.
- 새 프로젝트에 Claude Code 구성을 부트스트랩할 때.
- 기존 에이전트/커맨드/스킬/플러그인을 표준에 맞게 재검토할 때.
- `/alfred:0-project` 등의 초기화 워크플로우에서 설정 검증이 필요할 때.

## What it does
- 다섯 핵심 구성요소를 공식 템플릿으로 생성·갱신합니다.
- 허용 도구, 모델 선택, progressive disclosure 링크를 검증합니다.
- templates/·scripts/ 리소스를 통해 재사용 가능한 스텁과 검증 절차를 제공합니다.

## Inputs
- 사용자의 구성 요청(예: “새 커맨드 추가”, “settings.json 검토”)과 현재 `.claude/` 디렉터리 상태.
- 프로젝트별 템플릿 요구사항 또는 보안/권한 정책.

## Outputs
- `.claude/agents|commands|skills/` 하위에 정식 마크다운 정의 파일.
- 최신 설정이 반영된 `.claude/settings.json`과 허용 도구·후속 TODO가 정리된 요약본.

## Failure Modes
- 템플릿 경로나 placeholder가 최신 버전과 어긋나 결과물이 손상될 수 있습니다.
- settings.json 권한 정책이 프로젝트 규칙과 충돌하거나 검증 스크립트 실행이 차단될 수 있습니다.

## Dependencies
- cc-manager, doc-syncer, moai-foundation-git과 함께 사용하면 생성→검증→배포 흐름이 완성됩니다.
- templates/ 및 scripts/ 디렉터리에 버전 관리된 리소스가 있어야 자동화가 정상 작동합니다.

## References
- Anthropic. "Claude Code Style Guide." https://docs.claude.com/ (accessed 2025-03-29).
- Prettier. "Opinionated Code Formatter." https://prettier.io/docs/en/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: Claude 코드 포맷 스킬에 베스트 프랙티스 구조를 추가했습니다.
