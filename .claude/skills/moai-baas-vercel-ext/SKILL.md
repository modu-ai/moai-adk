# Skill: moai-baas-vercel-ext

## 메타데이터

```yaml
skill_id: moai-baas-vercel-ext
skill_name: Vercel 배포 및 Edge Functions
version: 1.0.0
created_date: 2025-11-09
language: korean
triggers:
  - keywords: ["Vercel", "Edge Functions", "Next.js", "배포", "ISR", "Serverless"]
  - contexts: ["vercel-detected", "pattern-a", "pattern-b", "pattern-d"]
agents:
  - frontend-expert
  - devops-expert
freedom_level: high
word_count: 600
context7_references:
  - url: "https://vercel.com/docs/deployments/overview"
    topic: "배포 방식 비교"
  - url: "https://vercel.com/docs/functions/edge-functions"
    topic: "Edge Functions 상세"
  - url: "https://vercel.com/docs/concepts/image-optimization"
    topic: "이미지 최적화"
spec_reference: "@SPEC:BAAS-ECOSYSTEM-001"
```

---

## 📚 내용

### 1. Vercel 배포 원리 (150 words)

**Vercel**은 Next.js 최적화된 클라우드 배포 플랫폼입니다.

**배포 프로세스**:
```
Git Push
   ↓
GitHub/GitLab 연동
   ↓
Vercel: 자동 빌드
   ├─ npm install
   ├─ npm run build (Next.js)
   └─ 최적화
   ↓
엣지 네트워크에 배포 (200+개 위치)
   ↓
CDN 캐싱
   ↓
Live!
```

**Next.js 렌더링 방식**:

| 방식 | 빌드 시점 | 캐싱 | 사용 시기 |
|-----|---------|------|---------|
| **SSG** | 빌드 타임 | 영구 | 블로그, 문서 |
| **ISR** | 백그라운드 | 시간 기반 | 준 정적 콘텐츠 |
| **SSR** | 요청마다 | 없음 | 실시간 데이터 |
| **CSR** | 클라이언트 | 없음 | 대시보드 |

**예제: ISR (Incremental Static Regeneration)**
```typescript
// pages/blog/[slug].tsx
export async function getStaticProps({ params }) {
  const post = await getPost(params.slug);

  return {
    props: { post },
    revalidate: 60 // 60초마다 재생성
  };
}
```

---

### 2. Edge Functions (200 words)

**Edge Functions**: 사용자에 가장 가까운 엣지에서 실행되는 Serverless 함수.

**Serverless vs Edge**:

```
Client Request
   ↓
┌─────────────────────────────────┐
│ Edge Functions (가깝고 빠름)      │
├─────────────────────────────────┤
│ - 실행 위치: 지역별 엣지 (200+곳) │
│ - 응답 시간: < 100ms            │
│ - 유효 기간: 15분                │
│ - 용도: 인증, 리다이렉트, 변환    │
└─────────────────────────────────┘
   ↓ (필요 시에만)
┌─────────────────────────────────┐
│ Serverless Functions (중앙집중)   │
├─────────────────────────────────┤
│ - 실행 위치: 중앙 데이터센터      │
│ - 응답 시간: 100-1000ms         │
│ - 유효 기간: 5분 (cold start)    │
│ - 용도: DB 쿼리, 계산, API 호출  │
└─────────────────────────────────┘
```

**Edge Functions 예제**:

```typescript
// api/middleware.ts - Supabase와 함께 사용
import { NextRequest, NextResponse } from 'next/server';

export async function middleware(req: NextRequest) {
  // 1. 인증 확인 (엣지에서 고속 실행)
  const token = req.cookies.get('auth_token');

  if (!token) {
    return NextResponse.redirect(new URL('/login', req.url));
  }

  // 2. 사용자 정보 조회 (선택: Supabase)
  const res = await fetch('https://xxx.supabase.co/rest/v1/users', {
    headers: {
      'Authorization': `Bearer ${token.value}`,
      'apikey': process.env.NEXT_PUBLIC_SUPABASE_KEY
    }
  });

  if (!res.ok) {
    return NextResponse.redirect(new URL('/unauthorized', req.url));
  }

  return NextResponse.next();
}

export const config = {
  matcher: ['/dashboard/:path*', '/api/:path*']
};
```

**성능 최적화**:
```typescript
// ✅ Edge Functions 사용
- 인증 토큰 검증
- 지역별 리다이렉트
- A/B 테스트
- 요청 변환

// ❌ Edge Functions 사용 금지
- 데이터베이스 쿼리 (느림)
- 파일 업로드 처리
- 복잡한 계산
- Realtime 구독
```

**Supabase와 함께 사용**:
```typescript
// 예: Edge에서 인증 후 Supabase 쿼리
const { data, error } = await supabase
  .from('posts')
  .select('*')
  .eq('user_id', userId)
  .limit(10);
```

---

### 3. Environment Variables (100 words)

**환경 변수 관리**:

```bash
# .env.local (로컬 개발)
NEXT_PUBLIC_SUPABASE_URL=xxx
NEXT_PUBLIC_SUPABASE_ANON_KEY=xxx
SUPABASE_SERVICE_KEY=xxx  # 서버만

# vercel.yml (프로덕션)
env:
  NEXT_PUBLIC_SUPABASE_URL: @supabase_url
  NEXT_PUBLIC_SUPABASE_ANON_KEY: @supabase_key
  SUPABASE_SERVICE_KEY: @supabase_service_key
```

**Secrets 관리**:
```bash
# Vercel CLI로 secrets 추가
vercel env add SUPABASE_SERVICE_KEY

# 또는 대시보드
Settings → Environment Variables → 추가
```

**주의사항**:
- ✅ `NEXT_PUBLIC_` = 클라이언트에 노출 (공개 정보만)
- ❌ 키 노출 = 즉시 재생성 필요
- ✅ Service role key는 절대 클라이언트에 노출 금지

---

### 4. Monitoring & Analytics (150 words)

**Web Vitals 추적**:

```typescript
// app/layout.tsx
import { Analytics } from '@vercel/analytics/react';

export default function RootLayout({ children }) {
  return (
    <html>
      <body>
        {children}
        <Analytics /> {/* 자동 추적 */}
      </body>
    </html>
  );
}
```

**추적 항목**:
- **LCP** (Largest Contentful Paint): 콘텐츠 로드 시간 (< 2.5s)
- **FID** (First Input Delay): 상호작용 지연 (< 100ms)
- **CLS** (Cumulative Layout Shift): 레이아웃 이동 (< 0.1)

**성능 최적화**:

```typescript
// 1. 동적 import (코드 분할)
const HeavyComponent = dynamic(() => import('./Heavy'), {
  loading: () => <Skeleton />
});

// 2. 이미지 최적화 (자동)
import Image from 'next/image';

export default function Page() {
  return (
    <Image
      src="/photo.jpg"
      width={400}
      height={300}
      // Vercel이 자동으로 최적화:
      // - WebP 변환
      // - 반응형 이미지
      // - Lazy loading
    />
  );
}

// 3. 폰트 최적화
import { Inter } from 'next/font/google';

const inter = Inter({ subsets: ['latin'] });
```

**Error Tracking**:
- Vercel 대시보드 → Logs → Errors 확인
- 자동 감지: 500 에러, 사용자 보고

**비용 모니터링**:
- Edge Requests: 무료 (일부)
- Serverless Functions: 요청당 과금
- 데이터 전송: 월별 제한
- 빌드: 월 100회 무료

---

## 🎯 사용 방법

### Agent에서 호출

```python
# frontend-expert, devops-expert에서
Skill("moai-baas-vercel-ext")

# Vercel 패턴 감지 시 자동 로드
```

### Context7 자동 로딩

Vercel 감지 시:
- 배포 방식 비교 (SSG vs ISR vs SSR)
- Edge Functions 상세 가이드
- 성능 최적화 체크리스트

---

## 📚 참고 자료

- [Vercel 배포 가이드](https://vercel.com/docs/deployments/overview)
- [Edge Functions](https://vercel.com/docs/functions/edge-functions)
- [이미지 최적화](https://vercel.com/docs/concepts/image-optimization)

---

## ✅ 검증

- [x] 배포 원리
- [x] Edge Functions 심화
- [x] Environment Variables
- [x] Monitoring & Analytics
- [x] 600 단어 목표
