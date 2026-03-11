---
id: SPEC-UPDATE-002
version: "1.0.0"
status: draft
created: "2026-03-11"
updated: "2026-03-11"
author: GOOS
priority: high
---

## HISTORY

| Date | Version | Author | Description |
|------|---------|--------|-------------|
| 2026-03-11 | 1.0.0 | GOOS | Initial SPEC creation |

---

# SPEC-UPDATE-002: moai-domain-uiux 및 moai-design-tools 스킬 전체 업데이트

## 1. 개요

moai-domain-uiux와 moai-design-tools 스킬을 최신 외부 생태계 변화에 맞게 업데이트한다.
핵심 변경: Figma MCP 공식 API 전환, Pencil MCP 팩트 수정, Anti-AI Slop 디자인 방향성 프레임워크 추가.

## 2. 배경

- 2026-03-10 세션에서 Top 3 외부 UI/UX 스킬(Anthropic Frontend Design, Vercel Web Design Guidelines, AccessLint)과 비교 분석 완료
- moai-design-tools의 Figma 섹션이 가상 API를 사용하고 있어 팩트 오류 심각
- moai-domain-uiux에 디자인 방향성/AI Slop 방지 가이드 부재
- Pencil MCP의 .pen 파일 형식에 대한 잘못된 정보 수정 필요

## 3. 요구사항 (EARS Format)

### R1: Figma MCP 공식 전환 (CRITICAL)

**Where** moai-design-tools reference/figma.md가 사용될 때,
**the system shall** Figma 공식 Remote MCP 서버(mcp.figma.com)의 검증된 도구 목록을 제공한다.

- 검증된 도구: generate_figma_design, get_design_context, get_screenshot, get_variable_defs, get_metadata, get_code_connect_map, add_code_connect_map, get_figjam, generate_diagram, create_design_system_rules, whoami
- 설치: `claude plugin install figma@claude-plugins-official`
- Code-to-Canvas (generate_figma_design) v1 제한사항 문서화

### R2: Pencil MCP 팩트 수정 (CRITICAL)

**Where** moai-design-tools reference/pencil-renderer.md가 사용될 때,
**the system shall** .pen 파일이 순수 JSON 형식(Git diff/merge 가능)임을 정확히 안내한다.

- ".pen files are encrypted" 문구 삭제
- MCP 자동 설정(수동 mcpServers 불필요) 반영
- UI Kit 4종(Shadcn UI, Halo, Lunaris, Nitro) 선택 가이드 추가

### R3: Pencil-to-Code 가상 API 제거 (CRITICAL)

**Where** moai-design-tools reference/pencil-code.md가 사용될 때,
**the system shall** 가상 export API(pencil.export_to_react, pencil.config.js)를 제거하고 실제 prompt 기반 워크플로우로 교체한다.

### R4: Anti-AI Slop 디자인 방향성 (HIGH)

**Where** moai-domain-uiux가 UI 디자인 가이드로 사용될 때,
**the system shall** AI Slop 방지를 위한 디자인 방향성 프레임워크를 제공한다.

- Purpose → Tone → Constraints → Differentiation 프로세스
- 금지 패턴: Inter/Roboto/Arial 폰트, 보라색 그라데이션, 예측 가능한 레이아웃
- 대안 제시: 스타일 극단(brutally minimal ~ maximalist chaos 등)
- 모션/공간 구성 가이드

### R5: 기술 스택 버전 업데이트 (MEDIUM)

**Where** moai-domain-uiux SKILL.md가 참조될 때,
**the system shall** TypeScript 5.9+, Tailwind CSS 4.x 등 최신 버전을 반영한다.

- Hugeicons 아이콘 라이브러리 추가 (Nova 프리셋 기본값)
- Nova 프리셋 상호참조 추가

### R6: 중복 콘텐츠 제거 (MEDIUM)

**Where** moai-domain-uiux design-system-tokens.md의 Pencil MCP 섹션이 로드될 때,
**the system shall** moai-design-tools로의 상호참조로 대체하여 중복을 제거한다.

## 4. 영향 범위

### 수정 대상 파일 (10개)

**moai-design-tools:**
1. SKILL.md — allowed-tools, 버전 업데이트
2. reference/figma.md — 전체 재작성
3. reference/pencil-renderer.md — .pen 형식 수정, MCP 설정
4. reference/pencil-code.md — 가상 API 제거, 재작성
5. reference/comparison.md — figma/pencil 변경 반영

**moai-domain-uiux:**
6. SKILL.md — 버전, 기술 스택, 트리거 업데이트
7. modules/web-interface-guidelines.md — Anti-AI Slop, 모션, 모바일 추가
8. modules/design-system-tokens.md — Pencil 중복 제거
9. modules/theming-system.md — Nova 상호참조, Tailwind v4
10. modules/icon-libraries.md — Hugeicons 추가

### 변경 없음 (2개)
- modules/component-architecture.md
- modules/accessibility-wcag.md

### 추가 작업
- `make build` — 임베디드 템플릿 재생성
- 로컬 `.claude/skills/` 디렉토리 동기화

## 5. 제외 범위

- Figma MCP allowed-tools를 moai-design-tools frontmatter에 추가하지 않음 (별도 MCP 서버)
- 미검증 Pencil 도구(search_all_unique_properties, replace_all_matching_properties) 포함하지 않음
- moai-domain-uiux 모듈 YAML frontmatter 추가 (별도 SPEC으로 분리)
