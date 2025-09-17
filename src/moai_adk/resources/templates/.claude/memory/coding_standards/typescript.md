# TypeScript 규칙

## ✅ 필수
- `tsconfig.json`: `strict: true`, `noImplicitAny`, `exactOptionalPropertyTypes` 활성화
- ESLint + Prettier + Vitest/Jest 조합, lint/test를 PR 게이트에 포함
- 타입은 `unknown` → 좁히기 패턴을 사용, `any` 금지, `never` 처리 필수
- 런타임 입력 검증은 Zod/valibot 등 스키마 도구 사용
- 테스트는 단위/통합 분리, 커버리지 80% 이상, shared_checklists 확인

## 👍 권장
- 폴더 구조는 feature-first, barrel 파일 최소화, 퍼블릭 API만 export
- React/Next 사용 시 Server/Client 컴포넌트 구분, Suspense/CSR 전략 명시
- Zustand/Redux/Recoil 등 상태 관리 시 파편화를 피하고 디버깅 도구 사용
- API 호출은 fetch wrapper + 타입 안전한 응답 파서 구성

## 🚀 확장/고급
- tsup/ts-node/tsx 등 실행 도구를 목적별로 분리, bundle 분석으로 성능 추적
- Storybook/Playwright와 연계하여 UI 테스트 자동화
- ESLint custom rule/TypeScript transformer로 팀 전용 규칙 강화
- SWC/ESBuild 기반 빌드 파이프라인 최적화 및 monorepo(workspace) 구성
