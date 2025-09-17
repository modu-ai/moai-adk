# 프레임워크 가이드

## React / Next.js
- App Router 우선, Server Action/Suspense/SSR 전략을 명시하고 클라이언트 상태는 query/cache/store 로 균형 분배
- CSR 전환 시 SEO 고려, Edge/Serverless 배포 시 환경 변수/시크릿 주의
- Storybook/Playwright/Testing Library로 UI/통합 테스트 구성

## Vue / Nuxt
- Composition API + `<script setup>`, Pinia 상태 모듈화, auto import 최소화
- Route middleware 정책 문서화, Nitro 서버와 API 라우팅 분리
- Vite 플러그인/ESLint 구성 공유, 테스트는 Vitest + Cypress

## Angular
- Standalone 컴포넌트, OnPush ChangeDetection, RxJS pipeable 연산자 사용
- DI 스코프/Providers 구조 문서화, Nx workspace로 모노레포 구성 고려
- Jest/Karma 대신 Vitest/Jest + Testing Library 조합 권장

## FastAPI / Django
- Pydantic(BaseModel) 검증, 의존성 주입/미들웨어 경계 명확화, DB 세션 범위 한정
- Django는 settings 분리/환경 변수 관리, DRF Serializer 검증 강화
- TestClient/pytest + FactoryBoy, contract tests(OpenAPI) 작성

## Spring Boot
- configuration properties 검증(@Validated), Profile/Actuator 관리, Transaction boundary 명시
- Spring Security/Method Security로 최소 권한 원칙 적용, Observability(Micrometer)
- MockMvc/WebTestClient + Testcontainers, Spring REST Docs/Swagger 문서화
