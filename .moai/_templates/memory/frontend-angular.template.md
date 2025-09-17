# $PROJECT_NAME 프론트엔드(Angular) 메모

> Angular 기반 애플리케이션 개발 시 참고할 운영 지침입니다.

## 1. 환경 & 구조
- Node LTS + pnpm/npm, Angular CLI(`ng`) 버전 고정
- Standalone 컴포넌트 + OnPush ChangeDetection + Signal/RxJS 조합 사용
- 폴더 구조: feature-first, shared 모듈 최소화, 라우트 lazy loading 정책 정의

## 2. 품질 도구
- ESLint + Prettier + Stylelint(JIT) 구성, `angular-eslint` 룰셋 활용
- 테스트: Jest/Vitest + Testing Library, Karma는 필요 시 유지
- 접근성: `@angular/cdk/a11y`, axe-core 스캔, i18n pipeline

## 3. pre-commit 예시
```yaml
repos:
  - repo: https://github.com/pre-commit/mirrors-eslint
    rev: v9.11.1
    hooks:
      - id: eslint
        files: "src/"
  - repo: https://github.com/pre-commit/mirrors-prettier
    rev: v4.0.0
    hooks:
      - id: prettier
  - repo: https://github.com/pre-commit/mirrors-stylelint
    rev: 16.7.0
    hooks:
      - id: stylelint
```

## 4. 빌드 & 배포
- 환경 설정: `environment.{env}.ts`, build-time secret 주입 방식 정의
- 빌드 최적화: `ng build --configuration production --budgets`, source map 정책 수립
- 관측성: Angular DevTools, Web Vitals 측정, Sentry/Datadog 통합

## 5. 참고 문서
- Framework 가이드: @.claude/memory/coding_standards/frameworks.md
- 공통 체크리스트/보안/TDD: @.claude/memory/shared_checklists.md / @.claude/memory/security_rules.md / @.claude/memory/tdd_guidelines.md
