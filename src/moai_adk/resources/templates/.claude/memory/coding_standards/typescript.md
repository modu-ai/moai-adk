# TypeScript 규칙(요약)

- tsconfig: strict true, noImplicitAny, exactOptionalPropertyTypes
- 도구: ESLint, Prettier, Vitest/Jest, ts-node(개발)
- 타입: 좁히기/never 처리, unknown 선호, any 금지
- 런타임 검증: Zod/valibot로 입력 스키마 검사
- 구조: feature-first, barrel 최소, public API 명확화
- 테스트: 단위/통합 분리, mocking 최소, 커버리지 ≥80%
