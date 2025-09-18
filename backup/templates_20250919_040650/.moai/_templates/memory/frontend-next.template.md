# $PROJECT_NAME 프론트엔드(Next.js) 메모

> Next.js(App Router) 기반 프로젝트 추가 지침입니다. React 공통 메모(`frontend-react.md`)와 함께 참고하세요.

## 1. 구조 & 라우팅
- App Router 기반 폴더 구조(`/app`, layout, loading, error), Route Group 활용
- Server Component 기본, Client Component 최소화(상태/브라우저 API 필요 시)
- Route Handler/API Routes 인증/캐싱 전략 정의, Edge/Node 런타임 분리

## 2. 데이터 & 상태
- fetch 캐싱(`cache`, `revalidate`, `tags`) 정책 수립, ISR/SSG/SSR 구분 명확화
- React Query/TanStack Query 등 클라이언트 캐시 사용 시 boundary 설정
- Env 변수를 `.env.local` 등으로 관리하고, `NEXT_PUBLIC_` prefix 규칙 준수

## 3. 품질/성능
- ESLint(next/core-web-vitals), Prettier, TypeScript strict mode, Jest/Vitest + Testing Library
- 성능: `next build --profile`, Lighthouse CI, Bundle Analyzer(`next-bundle-analyzer`)
- 모니터링: Vercel Analytics/Datadog/Sentry 연동, tracing exporter(OpenTelemetry)

## 4. 배포 전략
- Vercel/Cloudflare/Netlify 등 서버리스 환경 시 Edge Runtime 검토, Node 환경 시 PM2/Express adapter 활용
- 이미지/Font 최적화(Next Image, next/font), CDN 설정과 Cache-Control 헤더 명시
- CI: `pnpm lint && pnpm test && pnpm build`, `next export` 사용 여부 결정

## 5. 참고 문서
- Next.js 공식 문서: https://nextjs.org/docs
- React 메모: `./frontend-react.md`
- 공통 체크리스트/보안/TDD: @.claude/memory/shared_checklists.md / @.claude/memory/security_rules.md / @.claude/memory/tdd_guidelines.md
