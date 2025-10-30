---
id: NEXTRA-I18N-001
version: 0.0.1
status: draft
created: 2025-10-31
updated: 2025-10-31
author: @GOOS
priority: high
depends_on:
  - NEXTRA-SITE-001
---

# SPEC-NEXTRA-I18N-001 Implementation Plan

## HISTORY

### v0.0.1 (2025-10-31)
- **INITIAL**: next-intl 다국어 지원 실행 계획 작성
- **AUTHOR**: @GOOS
- **SECTIONS**: Milestones, Technical Approach, Architecture, Risks

---

## Milestones (우선순위 기반)

### Primary Goal: next-intl 설치 및 언어 라우팅 구성
**목표**: next-intl을 설치하고 `/ko/`, `/en/` 경로로 언어별 라우팅이 동작하는 것을 확인

**완료 조건**:
- ✅ `next-intl` 의존성 설치 완료
- ✅ `i18n/routing.ts` 파일 생성 및 언어 설정 완료
- ✅ `middleware.ts` 파일 생성 및 언어 감지 로직 구현
- ✅ `/ko/`, `/en/` 경로 접근 시 해당 언어 콘텐츠 표시
- ✅ 루트 경로(`/`) 접속 시 브라우저 언어에 따라 자동 리다이렉트

**주요 작업**:
1. next-intl 설치: `npm install next-intl@^3.25.0`
2. `i18n/routing.ts` 생성 (지원 언어: ko, en / 기본 언어: ko)
3. `middleware.ts` 생성 (Accept-Language 헤더 기반 리다이렉트)
4. `i18n/request.ts` 생성 (번역 파일 로더)
5. `next.config.js` 업데이트 (next-intl 플러그인 추가)
6. `app/[locale]/layout.tsx` 생성 (언어별 동적 라우팅)
7. `app/[locale]/page.tsx` 생성 (언어별 홈페이지)
8. 로컬 개발 서버에서 `/ko/`, `/en/` 경로 동작 확인

---

### Secondary Goal: 번역 파일 구조 생성 및 UI 텍스트 번역
**목표**: 한국어/영어 번역 파일을 생성하고 주요 UI 텍스트를 번역하여 표시

**완료 조건**:
- ✅ `messages/ko.json` 파일 생성 및 기본 번역 완료
- ✅ `messages/en.json` 파일 생성 및 기본 번역 완료
- ✅ Navigation, Common, Home, Footer 네임스페이스 번역 완료
- ✅ `useTranslations` 훅을 사용하여 컴포넌트에서 번역 텍스트 표시
- ✅ 번역 누락 시 기본 언어(한국어)로 폴백 동작 확인

**주요 작업**:
1. `messages/` 디렉토리 생성
2. `messages/ko.json` 생성 (한국어 번역)
   - Navigation: home, docs, api, guide, about
   - Common: learnMore, getStarted, search, language
   - Home: title, subtitle, description
   - Footer: copyright, github, discord
3. `messages/en.json` 생성 (영어 번역)
4. `app/[locale]/page.tsx`에서 `useTranslations('Home')` 훅 사용
5. 번역 폴백 동작 테스트 (번역 키 누락 시 기본 언어 표시)

---

### Final Goal: 언어 전환 UI 구현 및 Vercel 배포
**목표**: 사용자가 드롭다운에서 언어를 선택하여 즉시 전환할 수 있는 UI를 구현하고 Vercel에 배포

**완료 조건**:
- ✅ `components/LanguageSwitcher.tsx` 컴포넌트 생성 완료
- ✅ Nextra 테마 네비게이션에 언어 전환 드롭다운 추가
- ✅ 언어 전환 시 현재 페이지가 해당 언어 경로로 전환됨 (예: `/ko/docs` → `/en/docs`)
- ✅ 빌드 성공: `out/ko/`, `out/en/` 디렉토리 생성 확인
- ✅ Vercel 배포 완료: https://adk.mo.ai.kr/ko/, https://adk.mo.ai.kr/en/ 정상 접속

**주요 작업**:
1. `components/LanguageSwitcher.tsx` 생성
   - `useLocale()` 훅으로 현재 언어 가져오기
   - `useRouter()`와 `usePathname()` 훅으로 경로 전환
   - 드롭다운 UI (한국어/English)
2. `theme.config.tsx` 업데이트
   - `navbar.extraContent`에 LanguageSwitcher 추가
3. 빌드 테스트: `npm run build`
   - `out/ko/index.html` 존재 확인
   - `out/en/index.html` 존재 확인
4. Vercel 배포
5. 배포 후 언어별 URL 동작 확인

---

## Technical Approach (기술 접근 방식)

### 1. next-intl 선택 이유
- **Next.js 14 App Router 호환**: 공식적으로 Next.js 14 지원
- **파일 기반 번역**: JSON 파일로 번역 관리 (간단한 구조)
- **타입 안전성**: TypeScript와 잘 통합됨
- **Middleware 지원**: Accept-Language 헤더 자동 감지
- **활발한 커뮤니티**: 주간 다운로드 100만+, GitHub 스타 5k+

### 2. URL 기반 라우팅 선택 이유
- **SEO 친화적**: 검색 엔진이 언어별 페이지를 개별적으로 인덱싱
- **사용자 명확성**: URL만 보고 현재 언어 파악 가능
- **공유 가능**: 특정 언어 링크를 다른 사람에게 공유 가능
- **북마크 지원**: 사용자가 선호하는 언어로 북마크 가능

### 3. Accept-Language 헤더 기반 자동 리다이렉트
- **이유**: 사용자가 처음 방문 시 브라우저 언어로 자동 전환
- **장점**:
  - 사용자 경험 향상 (수동 언어 선택 불필요)
  - 대부분의 브라우저가 Accept-Language 헤더 전송
- **단점**:
  - 브라우저 설정이 잘못된 경우 의도하지 않은 언어로 표시
- **대응**:
  - 언어 전환 UI 제공 (사용자가 수동으로 변경 가능)
  - 기본 언어(한국어)로 폴백

### 4. 파일 기반 번역 vs. 데이터베이스 기반 번역
- **선택**: 파일 기반 번역 (JSON)
- **이유**:
  - 정적 사이트이므로 빌드 시 번역 파일 포함 가능
  - 데이터베이스 불필요 (비용 절감)
  - Git으로 번역 버전 관리 가능
  - 번역자가 JSON 파일만 수정하면 됨 (단순함)

---

## Architecture Design (아키텍처 설계)

### High-Level Architecture
```
┌─────────────────────────────────────────────────────────────┐
│                     User (Browser)                          │
│              Accept-Language: ko-KR, en-US                  │
└──────────────────────┬──────────────────────────────────────┘
                       │ GET /
                       ↓
┌─────────────────────────────────────────────────────────────┐
│                  Middleware (middleware.ts)                 │
│  - Accept-Language 헤더 파싱                                │
│  - 언어 감지: ko-KR → ko, en-US → en                        │
│  - 리다이렉트: / → /ko/ 또는 /en/                           │
└──────────────────────┬──────────────────────────────────────┘
                       │ 302 Redirect to /ko/
                       ↓
┌─────────────────────────────────────────────────────────────┐
│          App Router ([locale]/layout.tsx)                   │
│  - generateStaticParams: ['ko', 'en']                       │
│  - getMessages(): messages/ko.json 로드                     │
│  - NextIntlClientProvider로 하위 컴포넌트 래핑              │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ↓
┌─────────────────────────────────────────────────────────────┐
│          Component (useTranslations('Home'))                │
│  - t('title') → messages/ko.json의 Home.title 반환         │
│  - 번역 누락 시 기본 언어(ko)로 폴백                         │
└─────────────────────────────────────────────────────────────┘
```

### Middleware Flow
```
사용자 요청: /
    ↓
middleware.ts 실행
    ↓
Accept-Language 헤더 확인
    ├─ ko-KR → /ko/로 리다이렉트
    ├─ en-US → /en/으로 리다이렉트
    └─ 기타 → /ko/로 리다이렉트 (기본 언어)
    ↓
정적 페이지 제공 (out/ko/index.html 또는 out/en/index.html)
```

### Translation Loading Flow
```
빌드 시:
  - messages/ko.json, messages/en.json 읽기
  - i18n/request.ts에서 번역 파일 로드
  - generateStaticParams()로 언어별 페이지 생성
  - out/ko/index.html, out/en/index.html 생성

런타임 시:
  - 사용자가 /ko/ 접속
  - NextIntlClientProvider가 messages/ko.json 주입
  - useTranslations('Home') 훅으로 번역 가져오기
  - t('title') → "MoAI-ADK 문서" 반환
```

---

## Dependency Management (의존성 관리)

### New Dependencies
| Package | Version | Purpose |
|---------|---------|---------|
| `next-intl` | `^3.25.0` | Next.js 14 App Router i18n |

### Configuration Updates
| File | Change |
|------|--------|
| `next.config.js` | next-intl 플러그인 추가 |
| `middleware.ts` | 언어 감지 미들웨어 생성 |
| `i18n/routing.ts` | 언어 라우팅 설정 생성 |
| `i18n/request.ts` | 번역 파일 로더 생성 |

---

## Risks & Mitigation (위험 요소 및 대응 방안)

### Risk 1: next-intl 3.x와 Nextra 4.0 호환성 이슈
- **확률**: Medium
- **영향**: High (프로젝트 전체 지연)
- **대응**:
  - next-intl GitHub Issues에서 Nextra 관련 이슈 검색
  - 필요시 next-intl 2.x로 다운그레이드
  - Nextra 커뮤니티(Discord)에 문의
  - 대안: react-intl 또는 i18next 고려

### Risk 2: 번역 파일 동기화 문제 (키 누락)
- **확률**: High
- **영향**: Medium (일부 텍스트가 번역되지 않음)
- **대응**:
  - 번역 키 검증 스크립트 작성 (npm script):
    ```bash
    # 번역 키 비교 스크립트
    node scripts/check-translations.js
    ```
  - CI/CD에서 번역 파일 검증 자동화
  - 개발 모드에서 누락된 번역 키 경고 표시 (next-intl 기본 기능)

### Risk 3: 정적 빌드 시 언어별 페이지 생성 실패
- **확률**: Medium
- **영향**: High (특정 언어 페이지 접근 불가)
- **대응**:
  - `generateStaticParams()` 함수 검증:
    ```typescript
    export function generateStaticParams() {
      return [{ locale: 'ko' }, { locale: 'en' }];
    }
    ```
  - 로컬 빌드에서 `out/ko/`, `out/en/` 디렉토리 확인
  - 빌드 오류 발생 시 next-intl 문서 재검토

### Risk 4: 브라우저 언어 감지 오류 (Accept-Language 헤더 없음)
- **확률**: Low
- **영향**: Low (기본 언어로 폴백)
- **대응**:
  - 기본 언어(한국어)로 폴백 설정
  - 언어 전환 UI 제공 (사용자가 수동으로 변경 가능)

---

## Testing Strategy (테스트 전략)

### Unit Tests (Phase 1에서는 제외)
- **이유**: 정적 콘텐츠 위주이므로 복잡한 로직 없음
- **Phase 2 고려 사항**: 번역 파일 검증 테스트 추가

### Integration Tests
- **언어 라우팅 테스트**:
  - `/ko/` 경로 접근 → 한국어 콘텐츠 표시 확인
  - `/en/` 경로 접근 → 영어 콘텐츠 표시 확인
  - `/` 경로 접근 → Accept-Language 헤더에 따라 리다이렉트 확인

### E2E Tests (Phase 1에서는 수동)
- **로컬 개발 서버 테스트**:
  - `npm run dev` 실행 → `/ko/`, `/en/` 접속
  - 언어 전환 드롭다운 클릭 → 경로 변경 확인
  - 번역 텍스트 표시 확인

- **배포 후 프로덕션 테스트**:
  - https://adk.mo.ai.kr/ko/ 접속 → 한국어 표시
  - https://adk.mo.ai.kr/en/ 접속 → 영어 표시
  - 언어 전환 동작 확인

---

## Translation File Validation Script (번역 파일 검증 스크립트)

### scripts/check-translations.js
```javascript
const fs = require('fs');
const path = require('path');

const messagesDir = path.join(__dirname, '../messages');
const locales = ['ko', 'en'];

function flattenObject(obj, prefix = '') {
  return Object.keys(obj).reduce((acc, k) => {
    const pre = prefix.length ? prefix + '.' : '';
    if (typeof obj[k] === 'object' && obj[k] !== null) {
      Object.assign(acc, flattenObject(obj[k], pre + k));
    } else {
      acc[pre + k] = obj[k];
    }
    return acc;
  }, {});
}

function validateTranslations() {
  const translations = {};

  // 모든 언어의 번역 파일 로드
  locales.forEach(locale => {
    const filePath = path.join(messagesDir, `${locale}.json`);
    const content = JSON.parse(fs.readFileSync(filePath, 'utf-8'));
    translations[locale] = flattenObject(content);
  });

  // 기본 언어(한국어) 키를 기준으로 검증
  const baseKeys = Object.keys(translations['ko']);
  let hasErrors = false;

  locales.forEach(locale => {
    if (locale === 'ko') return; // 기본 언어는 스킵

    const localeKeys = Object.keys(translations[locale]);

    // 누락된 키 확인
    const missingKeys = baseKeys.filter(key => !localeKeys.includes(key));
    if (missingKeys.length > 0) {
      console.error(`❌ Missing keys in ${locale}.json:`);
      missingKeys.forEach(key => console.error(`   - ${key}`));
      hasErrors = true;
    }

    // 추가된 키 확인
    const extraKeys = localeKeys.filter(key => !baseKeys.includes(key));
    if (extraKeys.length > 0) {
      console.warn(`⚠️  Extra keys in ${locale}.json (not in ko.json):`);
      extraKeys.forEach(key => console.warn(`   - ${key}`));
    }
  });

  if (!hasErrors) {
    console.log('✅ All translation files are in sync!');
  } else {
    process.exit(1);
  }
}

validateTranslations();
```

### package.json에 스크립트 추가
```json
{
  "scripts": {
    "check-translations": "node scripts/check-translations.js"
  }
}
```

---

## Performance Optimization (성능 최적화)

### Build Time Optimization
- **목표**: 빌드 시간 증가 최소화 (< 10% 증가)
- **전략**:
  - 번역 파일은 빌드 시 한 번만 로드
  - 언어별 페이지는 정적 생성 (SSG)

### Runtime Performance
- **번들 크기**: next-intl은 경량 라이브러리 (gzip 후 ~10KB)
- **전략**:
  - 번역 파일은 클라이언트에 포함 (정적 사이트이므로 불가피)
  - 사용하지 않는 번역 키 제거 (tree-shaking)

---

## Rollback Strategy (롤백 전략)

### Scenario 1: next-intl 호환성 이슈
- **대응**:
  1. next-intl 제거: `npm uninstall next-intl`
  2. 이전 커밋으로 복원: `git revert HEAD`
  3. 대안 라이브러리 검토 (react-intl, i18next)

### Scenario 2: 빌드 실패 (언어별 페이지 생성 안 됨)
- **대응**:
  1. 로컬에서 빌드 재현: `npm run build`
  2. 빌드 로그 확인 및 오류 수정
  3. 문제 해결 전까지 NEXTRA-SITE-001 버전 유지

---

## Documentation Plan (문서 작성 계획)

### 1. 다국어 개발 가이드 (`@DOC:I18N-GUIDE-001`)
**내용**:
- 번역 파일 작성 방법 (`messages/ko.json`, `messages/en.json`)
- `useTranslations` 훅 사용법
- 번역 키 네이밍 규칙
- 번역 파일 검증 스크립트 실행 방법

### 2. i18n 트러블슈팅 가이드 (`@DOC:I18N-TROUBLESHOOTING-001`)
**내용**:
- 번역이 표시되지 않는 경우 해결 방법
- 언어 전환이 동작하지 않는 경우 해결 방법
- 빌드 실패 시 디버깅 방법

---

## Success Criteria (성공 기준)

### Milestone 1 완료 기준
- ✅ `/ko/`, `/en/` 경로 접근 시 해당 언어 콘텐츠 표시
- ✅ 루트 경로(`/`) 접속 시 브라우저 언어에 따라 자동 리다이렉트

### Milestone 2 완료 기준
- ✅ 번역 파일(`messages/ko.json`, `messages/en.json`) 정상 로드
- ✅ UI 텍스트가 번역 파일에서 가져와짐

### Milestone 3 완료 기준
- ✅ 언어 전환 드롭다운 동작
- ✅ Vercel 배포 후 https://adk.mo.ai.kr/ko/, https://adk.mo.ai.kr/en/ 정상 접속

### 전체 SPEC 완료 기준
- ✅ 모든 Milestone 완료
- ✅ 개발 가이드 작성 완료
- ✅ Git 커밋 및 PR 생성 완료

---

## Next Steps (다음 단계)

### 의존 관계
- **선행 작업**: @SPEC:NEXTRA-SITE-001 (Next.js 14 + Nextra 4.0 기본 구조)
- **후속 작업**: @SPEC:NEXTRA-CONTENT-001 (콘텐츠 마이그레이션)

### Handoff Points
- **구현 담당**: TDD-implementer (via `/alfred:2-run SPEC-NEXTRA-I18N-001`)
- **문서 작성**: doc-syncer (via `/alfred:3-sync`)
- **Git 관리**: git-manager (자동)

---

## Appendix

### Useful Commands
```bash
# next-intl 설치
npm install next-intl@^3.25.0

# 번역 파일 검증
npm run check-translations

# 개발 서버 시작
npm run dev

# 언어별 경로 접속
open http://localhost:3000/ko/
open http://localhost:3000/en/

# 빌드 및 언어별 디렉토리 확인
npm run build
ls -la out/ko/
ls -la out/en/
```

### Reference Links
- [next-intl Documentation](https://next-intl.dev)
- [Next.js 14 Internationalization](https://nextjs.org/docs/app/building-your-application/routing/internationalization)
- [Nextra i18n Guide](https://nextra.site/docs/guide/i18n)
