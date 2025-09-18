# $PROJECT_NAME 프론트엔드(React) 메모

> React 기반 UI 개발을 위한 공통 지침입니다. Next.js 등 상위 프레임워크를 사용하는 경우 `frontend-next.md`도 함께 참고하세요.

## 1. 환경 & 도구
- Node LTS(≥18) + pnpm/npm 설정, `.nvmrc` 또는 `.node-version` 공유
- Lint/Format: ESLint(typescript-eslint) + Prettier, 상태 관리 기본 전술 정의(Zustand/Redux 등)
- Styling: Tailwind/SCSS 사용 시 설계 문서와 토큰 체계 기록

## 2. pre-commit Hook
```yaml
repos:
  - repo: https://github.com/pre-commit/mirrors-eslint
    rev: v9.11.1
    hooks:
      - id: eslint
        files: "(src|apps)/"
        types: [ts, tsx, js, jsx]
        args: ["--max-warnings=0"]
  - repo: https://github.com/pre-commit/mirrors-prettier
    rev: v4.0.0
    hooks:
      - id: prettier
        additional_dependencies:
          - prettier
          - prettier-plugin-tailwindcss
```
- CI에서는 `pnpm lint`, `pnpm test`, `pnpm build` 순으로 실행

## 3. 테스트 & 품질
- Vitest/Jest + Testing Library, Storybook(필요 시)으로 UI 스냅샷/시나리오 검증
- 커버리지 80% 이상, 접근성 검사(Lighthouse, axe)와 E2E(Cypress/Playwright) 계획
- 성능 기준(TTFB, LCP 등) 측정 및 모니터링 도구 연결 (Sentry/Datadog/LogRocket)

## 4. 구조 & 배포
- feature-first 폴더 구조, barrel 최소화, public API 명시
- env 구분(`.env.local`, `.env.production`)과 시크릿 관리(.env.example, Vault 등)
- CI/CD: build output 분석(Bundle Analyzer) + 캐시 전략 설정, CDN/Edge 배포 고려

## 5. 참고 문서
- TypeScript/React 규칙: @.claude/memory/coding_standards/typescript.md, @.claude/memory/coding_standards/frameworks.md
- 공통 체크리스트/보안/TDD: @.claude/memory/shared_checklists.md / @.claude/memory/security_rules.md / @.claude/memory/tdd_guidelines.md
