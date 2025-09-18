# $PROJECT_NAME 프론트엔드(Vue/Nuxt) 메모

> Vue 3 또는 Nuxt 프로젝트를 위한 기본 지침입니다. Nuxt 사용 시 서버 렌더링 섹션을 함께 참고하세요.

## 1. 환경 & 구조
- Node LTS + pnpm 권장, `.nvmrc` 공유
- Vue 3 + Composition API + `<script setup>` 기본, 컴포넌트는 기능 단위 폴더 구조 유지
- 상태 관리: Pinia 모듈화, Persist/Plugin 정책 정의

## 2. 품질 도구
- ESLint(vue-eslint-parser) + Prettier + Stylelint 구성
- 테스트: Vitest + Vue Testing Library, E2E는 Cypress/Playwright
- 접근성: @vueuse/head, axe-core 체크, i18n 번역 파일 관리

## 3. Nuxt 특화(선택)
- Nitro 서버 + API Route, middleware 정책, server/api 디렉터리 구조 문서화
- Static/Hybrid/SSR 모드 결정, prerender 설정(`nuxt.config.ts`) 명시
- 라우터 전환 hook, layout 시스템 활용, Edge 배포시 `preset` 옵션 확인

## 4. pre-commit 예시
```yaml
repos:
  - repo: https://github.com/pre-commit/mirrors-eslint
    rev: v9.11.1
    hooks:
      - id: eslint
        files: "(src|app)/"
  - repo: https://github.com/pre-commit/mirrors-prettier
    rev: v4.0.0
    hooks:
      - id: prettier
  - repo: https://github.com/stylelint/stylelint
    rev: 16.7.0
    hooks:
      - id: stylelint
```

## 5. 배포 체크
- Nuxt: Vercel/Netlify/Cloudflare 배포 시 서버 자원 제한/캐시 정책 설정
- Vue SPA: Vite build + CDN 배포, SSG 필요 시 `vite-ssg`
- 모듈 federation/Micro Frontend 고려 시 빌드 아웃라인 문서화

## 6. 참고 문서
- Vue 규칙: @.claude/memory/coding_standards/frameworks.md
- 공통 체크리스트/보안/TDD: @.claude/memory/shared_checklists.md / @.claude/memory/security_rules.md / @.claude/memory/tdd_guidelines.md
