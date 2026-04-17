---
id: SPEC-OPUS47-COMPAT-001
version: 0.1.1
status: draft
created: 2026-04-17
updated: 2026-04-17
author: GOOS행님
priority: critical
issue_number: 671
---

# SPEC-OPUS47-COMPAT-001 — Compact Summary

## EARS Requirements

- **REQ-OC-001** [Ubiquitous, MODIFY]: 시스템은 `claude-opus-4-7` 모델 및 5단계 effort(`low`/`medium`/`high`/`xhigh`/`max`)를 프로파일/런처/템플릿에서 일급 시민으로 지원해야 한다.
- **REQ-OC-002** [Event-Driven, MODIFY]: `moai init`/`moai update` 시점에 6개 reasoning agent(manager-spec, manager-strategy, plan-auditor, evaluator-active, expert-security, expert-refactoring)에 `effort` 필드 및 Opus 4.7 프롬프트 철학을 적용해야 한다.
- **REQ-OC-003** [State-Driven, MODIFY]: Claude Code v2.1.110+ 환경에서 (a) moai doctor MCP scope 중복 경고, (b) PermissionRequest updatedInput deny 재검증, (c) Bash tool timeout 상한(600,000ms)을 agent-authoring.md / agent-common-protocol.md에 문서화(enforcement는 Claude Code 런타임이 담당), (d) disableBypassPermissionsMode 정책 필드를 제공해야 한다.
- **REQ-OC-004** [Optional, MODIFY]: Windows 환경에서 SessionStart 훅이 `CLAUDE_ENV_FILE` 환경변수를 주입해야 한다. macOS/Linux는 기존 동작 유지.
- **REQ-OC-005** [Unwanted, MODIFY]: Opus 4.7 대상 agent 프롬프트 본문에서 Opus 4.6 시대 스캐폴딩("double-check X", "verify N times" 류)과 고정 Thinking 예산 지시를 제거해야 한다.

## Given / When / Then 시나리오 요약

- **GWT-1** (REQ-OC-001): `moai profile`에서 Opus 4.7 + xhigh 선택 시 preferences.yaml에 정확히 기록
- **GWT-2** (REQ-OC-002): 신규 프로젝트 배포 후 manager-spec.md에 `effort: xhigh` frontmatter 존재, 22개 비대상 agent에는 `effort` 키 부재
- **GWT-3** (REQ-OC-002): `rg "max is Opus 4.6 only"` 결과 0건
- **GWT-4** (REQ-OC-003): `moai doctor` 실행 시 MCP scope 중복 warning 출력, exit 0
- **GWT-5** (REQ-OC-003): 렌더링된 settings.json에 `disableBypassPermissionsMode: false` 키 존재 + agent-authoring.md/agent-common-protocol.md에 Bash timeout 600,000ms 상한 문서화 매치
- **GWT-6** (REQ-OC-004): Windows에서 `CLAUDE_ENV_FILE` 주입, macOS/Linux 회귀 없음
- **GWT-7** (REQ-OC-005): manager-spec/manager-strategy/plan-auditor/evaluator-active/expert-security/expert-refactoring 6개 파일에서 "double-check|verify N times|explicitly confirm before" 패턴 검색 시 각각 0건(하드 기준), 워크플로우 스텝 수 유지 또는 증가
- **GWT-8** (REQ-OC-005): `rg "thinking\.budget_tokens|thinking budget \d+"` 결과 0건, Sequential Thinking MCP 지시는 유지

## 파일 목록 (28 files)

### P0 — 15 files (코어 호환성 + 프롬프트 철학)

1. `internal/profile/preferences.go` [MODIFY]
2. `internal/template/model_policy.go` [MODIFY]
3. `internal/cli/profile_setup_translations.go` [MODIFY]
4. `internal/cli/launcher.go:489` [MODIFY]
5. `internal/template/templates/.moai/config/sections/quality.yaml` [MODIFY]
6. `internal/template/templates/.moai/config/sections/llm.yaml` [MODIFY]
7. `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` [MODIFY]
8. `internal/template/templates/.claude/rules/moai/development/skill-authoring.md` [MODIFY]
9. `internal/template/templates/.claude/agents/moai/manager-spec.md` [MODIFY]
10. `internal/template/templates/.claude/agents/moai/manager-strategy.md` [MODIFY]
11. `internal/template/templates/.claude/agents/moai/plan-auditor.md` [MODIFY]
12. `internal/template/templates/.claude/agents/moai/evaluator-active.md` [MODIFY]
13. `internal/template/templates/.claude/agents/moai/expert-security.md` [MODIFY]
14. `internal/template/templates/.claude/agents/moai/expert-refactoring.md` [MODIFY]
15. `internal/template/templates/.claude/skills/moai-workflow-thinking/SKILL.md` [MODIFY]

### P1 — 9 files (v2.1.110 Runtime 대응)

16. `internal/cli/doctor.go` [MODIFY]
17. `internal/hook/permission_request.go` [MODIFY]
18. `internal/hook/session_start.go` [MODIFY]
19. `internal/template/templates/.claude/settings.json` [MODIFY]
20. `internal/template/templates/.claude/rules/moai/core/agent-common-protocol.md` [MODIFY]
21. `internal/template/templates/.claude/rules/moai/development/agent-authoring.md` (P0 파일 7과 동일, Bash timeout 문서화 내용 추가) [MODIFY]
22. `internal/template/templates/.claude/rules/moai/core/moai-constitution.md` [MODIFY]
23. `internal/template/templates/.claude/rules/moai/workflow/worktree-integration.md` [MODIFY]
24. `internal/template/templates/.moai/config/sections/harness.yaml` [MODIFY]

### P2 — 4 files (문서 레이어)

25. `internal/template/templates/.claude/rules/moai/development/coding-standards.md` [MODIFY]
26. `CLAUDE.local.md` [MODIFY]
27. `internal/template/templates/CLAUDE.md §12` [MODIFY]
28. `CHANGELOG.md` [MODIFY]

## Exclusions (9)

1. Vertex AI / AWS Bedrock 지원 추가
2. `/tui`, `/focus`, Push notification, Remote Control 등 v2.1.110 UX 기능
3. `plugin_errors` stream-json 구조 구현
4. TRACEPARENT/TRACESTATE SDK 분산 트레이싱
5. `CLAUDE_CODE_ENABLE_AWAY_SUMMARY` opt-out 구현
6. `/ultrareview` vs `/moai review` 관계 정리
7. 기존 Opus 4.6/Sonnet 4.6/Haiku 4.5 모델 제거
8. 실제 Go 코드·YAML·Markdown 파일의 구현 변경 (본 SPEC은 Plan)
9. Sequential Thinking MCP(`--deepthink`) 비활성화
