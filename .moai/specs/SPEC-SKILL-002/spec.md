---
id: SPEC-SKILL-002
version: 1.0.0
status: completed
created: 2026-04-09
updated: 2026-04-09
author: GOOS
priority: high
issue_number: 0
---

# SPEC-SKILL-002: MoAI-ADK 스킬 전면 최적화 — 공식 가이드 준수

## HISTORY

- 2026-04-09: 57개 스킬 전수 감사 결과 + 공식 PDF 가이드 대조 분석 기반 초안

## Overview

Claude Code 공식 Skills 가이드(PDF 28p)의 권고사항을 MoAI-ADK 57개 스킬에 전면 적용한다. 핵심 변경: 16개 언어 스킬을 paths-based rules로 전환, references/ 링크 복원, 이름 규약 준수, 불완전 파일 복원.

## Environment

- Claude Code v2.1.97+
- moai-adk-go template system
- 57 skill definitions (.claude/skills/*/SKILL.md)
- 16 language rules (.claude/rules/moai/languages/*.md)
- Go embed system (make build required)

## Requirements

### REQ-001: 16개 언어 스킬을 Rules로 전환 (Ubiquitous)

시스템은 언어별 개발 가이드를 skill이 아닌 paths-based rule로 제공해야 한다.

대상: moai-lang-go, moai-lang-python, moai-lang-typescript, moai-lang-javascript, moai-lang-rust, moai-lang-java, moai-lang-kotlin, moai-lang-swift, moai-lang-csharp, moai-lang-php, moai-lang-ruby, moai-lang-elixir, moai-lang-cpp, moai-lang-r, moai-lang-scala, moai-lang-flutter

전환 절차:
1. 기존 rule 파일(~47줄)에 스킬의 풍부한 콘텐츠(~173줄)를 병합
2. `paths:` frontmatter 유지 (이미 설정됨)
3. 스킬 디렉토리(SKILL.md + references/) 삭제
4. 에이전트 frontmatter에서 해당 스킬 참조 제거 (이미 SPEC-AGENT-002에서 제거됨)

장점:
- .go 파일 작업 시 Go 규칙 자동 로드 (확정적, 파일 패턴 기반)
- 스킬 수 57→41개 감소 (컨텍스트 부하 28% 절감)
- Progressive disclosure Level 1 메타데이터 로드 불필요

### REQ-002: moai-foundation-claude 이름 변경 (Ubiquitous)

시스템의 모든 스킬 이름에 "claude" 또는 "anthropic"이 포함되지 않아야 한다.

- `moai-foundation-claude` → `moai-foundation-cc` 로 변경
- 디렉토리명, SKILL.md name 필드, 모든 참조(에이전트 frontmatter, CLAUDE.md, 다른 스킬) 일괄 수정

### REQ-003: 41개 스킬 references/ 링크 복원 (Event-Driven)

WHEN 스킬에 references/ 디렉토리가 존재할 때 THEN SKILL.md에서 `${CLAUDE_SKILL_DIR}/references/` 경로로 명시적 링크가 있어야 한다.

- 41개 스킬에 references/ 링크 추가
- 패턴: `For detailed patterns, consult ${CLAUDE_SKILL_DIR}/references/{filename}.md`

### REQ-004: 4개 대형 스킬 references/ 분리 (State-Driven)

WHILE 스킬 본문이 300줄을 초과하는 상태에서 references/ 디렉토리가 없으면 THEN 상세 콘텐츠를 references/로 분리해야 한다.

대상:
- moai-platform-deployment (409줄) — 코드 예제를 references/로 이동
- moai-workflow-ddd (394줄) — 상세 패턴을 references/로 이동
- moai-design-tools (353줄) — 설정 예제를 references/로 이동
- moai-platform-chrome-extension (318줄) — API 참조를 references/로 이동

### REQ-005: 15개 스킬 progressive_disclosure 추가 (Ubiquitous)

시스템의 모든 스킬에 progressive_disclosure 설정이 있어야 한다.

대상: moai-docs-generation, moai-domain-backend, moai-domain-database, moai-domain-frontend, moai-domain-uiux, moai-formats-data, moai-framework-electron, moai-library-mermaid, moai-library-nextra, moai-library-shadcn, moai-tool-ast-grep, moai-tool-svg, moai-workflow-ddd, moai-workflow-loop, moai

### REQ-006: 코드 예제 references/ 이동 (Ubiquitous)

시스템의 스킬 SKILL.md에 개념 설명용 코드 예제가 직접 포함되지 않아야 한다.

대상:
- moai-tool-svg: SVG 패턴 코드 블록 → references/examples.md
- moai-platform-deployment: JSON 설정 예제 → references/config-examples.md
- moai-library-shadcn: TSX 코드 예제 → references/component-examples.md

### REQ-007: 3개 불완전 언어 스킬 복원 (Unwanted Behavior)

IF 스킬 파일이 불완전하게 잘려 있다면 THEN 다른 언어 스킬 수준으로 복원해야 한다.

대상: moai-lang-swift (141줄), moai-lang-csharp (133줄), moai-lang-flutter (124줄)

참고: REQ-001에 의해 rules로 전환되므로, 복원된 내용이 rule 파일에 병합됨.

## Exclusions (What NOT to Build)

- Shall NOT modify Agency skills (agency-*) — 이미 전체 준수
- Shall NOT modify moai-ref-* reference skills — 이미 준수
- Shall NOT change skill naming conventions beyond claude→cc 변경
- Shall NOT restructure workflow skills (moai-workflow-*) — 콘텐츠 유지
- Shall NOT change moai unified orchestrator skill beyond references/ 링크
- Will NOT implement scripts/ directory usage — 향후 별도 검토
