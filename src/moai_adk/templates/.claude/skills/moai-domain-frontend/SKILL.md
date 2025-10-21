---

name: moai-domain-frontend
description: React/Vue/Angular development with state management, performance optimization, and accessibility. Use when working on frontend interfaces scenarios.
allowed-tools:
  - Read
  - Bash
---

# Frontend Expert

## Skill Metadata
| Field | Value |
| ----- | ----- |
| Allowed tools | Read (read_file), Bash (terminal) |
| Auto-load | On demand for frontend delivery |
| Trigger cues | Component architecture, design systems, accessibility, performance budgets. |
| Tier | 4 |

## What it does

Provides expertise in modern frontend development using React, Vue, or Angular, including state management patterns, performance optimization techniques, and accessibility (a11y) best practices.

## When to use

- Engages when building or reviewing UI/front-end experiences.
- “Front-end development”, “React components”, “state management”, “performance optimization”
- Automatically invoked when working with frontend projects
- Frontend SPEC implementation (`/alfred:2-run`)

## How it works

**React Development**:
- **Functional components**: Hooks (useState, useEffect, useMemo)
- **State management**: Redux, Zustand, Jotai
- **Performance**: React.memo, useCallback, code splitting
- **Testing**: React Testing Library

**Vue Development**:
- **Composition API**: setup(), reactive(), computed()
- **State management**: Pinia, Vuex
- **Performance**: Virtual scrolling, lazy loading
- **Testing**: Vue Test Utils

**Angular Development**:
- **Components**: TypeScript classes with decorators
- **State management**: NgRx, Akita
- **Performance**: OnPush change detection, lazy loading
- **Testing**: Jasmine, Karma

**Performance Optimization**:
- **Code splitting**: Dynamic imports, route-based splitting
- **Lazy loading**: Images, components
- **Bundle optimization**: Tree shaking, minification
- **Web Vitals**: LCP, FID, CLS optimization

**Accessibility (a11y)**:
- **Semantic HTML**: Proper use of HTML5 elements
- **ARIA attributes**: Roles, labels, descriptions
- **Keyboard navigation**: Focus management
- **Screen reader support**: Alt text, aria-live

## Examples
```bash
$ npm run lint && npm run test
$ npm run build -- --profiling
```

## Inputs
- 도메인 관련 설계 문서 및 사용자 요구사항.
- 프로젝트 기술 스택 및 운영 제약.

## Outputs
- 도메인 특화 아키텍처 또는 구현 가이드라인.
- 연관 서브 에이전트/스킬 권장 목록.

## Failure Modes
- 도메인 근거 문서가 없거나 모호할 때.
- 프로젝트 전략이 미확정이라 구체화할 수 없을 때.

## Dependencies
- `.moai/project/` 문서와 최신 기술 브리핑이 필요합니다.

## References
- Google. "Web.dev Performance Guidelines." https://web.dev/fast/ (accessed 2025-03-29).
- W3C. "Web Content Accessibility Guidelines (WCAG) 2.2." https://www.w3.org/TR/WCAG22/ (accessed 2025-03-29).

## Changelog
- 2025-03-29: 도메인 스킬에 대한 입력/출력 및 실패 대응을 명문화했습니다.

## Works well with

- alfred-trust-validation (frontend testing)
- typescript-expert (type-safe React/Vue)
- alfred-performance-optimizer (performance profiling)

## Best Practices
- 도메인 결정 사항마다 근거 문서(버전/링크)를 기록합니다.
- 성능·보안·운영 요구사항을 초기 단계에서 동시에 검토하세요.
