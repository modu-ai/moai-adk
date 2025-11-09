# Nextra 4 마이그레이션 - 검증 및 위험 대응 체크리스트

## 1. Pre-Migration 검증 체크리스트

### 1.1 호환성 검증

**Next.js 16 호환성**:
- [ ] Next.js 16 공식 문서 검토 완료
  - 확인 URL: https://nextjs.org/docs
  - 주요 변경사항: React 19 필수, TypeScript 5.x 호환성

- [ ] React 19 호환성 확인
  - 문서: https://react.dev/blog/2024/12/05/react-19
  - 주요: Ref 콜백 변경, 새 hooks 추가

- [ ] TypeScript 5.x 호환성
  - 현재: 5.9.3 (호환)
  - 타입 정의 업데이트 필요 여부 확인

**Nextra 4.6.0 호환성**:
- [ ] Nextra 4.6.0 공식 마이그레이션 가이드 검토
  - URL: https://nextra.site/guide/migrate-from-3
  - Breaking changes 확인:
    - [ ] i18n 설정 변경 (page routing → app routing)
    - [ ] theme.config 구조 변경
    - [ ] meta.json 네이밍 (\_meta.json → meta.json)
    - [ ] MDX 컴포넌트 props 변경 여부

- [ ] nextra-theme-docs 4.6.0 호환성
  - 문서: https://nextra.site/docs/guide/theme-configuration
  - 확인 사항:
    - [ ] i18n 설정 여전히 theme.config에서 가능한지
    - [ ] search 설정 (FlexSearch vs Pagefind)
    - [ ] 커스텀 컴포넌트 호환성

**Turbopack 호환성**:
- [ ] Turbopack 실험적 기능 검토
  - URL: https://turbo.build/pack/docs/features/bundling
  - 주의: 아직 실험적 기능, edge case 존재 가능

### 1.2 현재 환경 스냅샷

**기본 정보 기록**:
```bash
# 현재 빌드 성능
npm run build > build-performance-before.log

# 빌드 시간: ___ 초
# 번들 크기: ___ MB
# 에러: (있으면 기록)
```

- [ ] 현재 package-lock.json 또는 uv.lock 백업
- [ ] 현재 main 브랜치 commit hash 기록: `________________`
- [ ] 현재 번들 크기 측정:
  - app.js: ___ KB
  - 기타: ___ KB

**다국어 상태 확인**:
- [ ] /ko 페이지 정상 동작 확인
- [ ] /en 페이지 정상 동작 확인
- [ ] /ja 페이지 정상 동작 확인
- [ ] /zh 페이지 정상 동작 확인
- [ ] 검색 기능 정상 동작 확인
- [ ] 사이드바 모든 섹션 펼쳐짐 확인

### 1.3 의존성 분석

**직접 의존성**:
```json
현재:
- next: 14.2.15
- nextra: 3.3.1
- nextra-theme-docs: 3.3.1
- react: 18.2.0
- react-dom: 18.2.0

업그레이드 후:
- next: ^16.0.0
- nextra: ^4.6.0
- nextra-theme-docs: ^4.6.0
- react: ^19.0.0
- react-dom: ^19.0.0
- pagefind: ^1.1.0 (신규)
```

**Peer Dependency 충돌 예상**:
- [ ] React 18 → 19 업그레이드 시 호환성 확인 필요한 패키지:
  - tailwindcss (호환성 확인)
  - @next/third-parties (확인)
  - 기타 플러그인 (확인)

**대응 방안**:
```bash
# Peer dependency 충돌 발생 시
npm install --legacy-peer-deps
# 또는 uv sync --allow-unsafe-dep
```

---

## 2. Phase별 검증 체크리스트

### Phase 1: 호환성 검증

**완료 조건**:
- [ ] Breaking changes 문서 모두 검토
- [ ] 호환성 이슈 0건 (또는 모두 대응 방안 수립)
- [ ] 팀 검토 및 승인 완료

**위험 신호**:
- [ ] 문서에서 "unstable" 또는 "experimental" 표기
- [ ] GitHub issues에서 "migration" 관련 오픈 이슈 다수
- [ ] 마이그레이션 사례가 매우 적음

**위험 대응**:
```
IF 위험 신호 발견 THEN
  1. 해당 이슈 상세 검토
  2. 임시 해결책 연구
  3. 필요시 workaround 구현
  4. 최악: 마이그레이션 일정 연기
```

---

### Phase 2: 기본 구조 전환

**생성 파일 검증**:
- [ ] `app/layout.jsx` 생성 완료
  ```bash
  wc -l app/layout.jsx
  # 예상: 50-100 lines
  ```
- [ ] `app/page.jsx` 생성 완료
- [ ] `app/[locale]/layout.jsx` 생성 완료
- [ ] `app/[locale]/page.jsx` 생성 완료
- [ ] `app/[locale]/[[...slug]]/page.jsx` 생성 완료

**로컬 dev 서버 테스트**:
```bash
npm install
npm run dev

# 확인 사항:
```
- [ ] 서버 정상 시작 (에러 없음)
- [ ] http://localhost:3000/ 접근 가능
- [ ] 기본 언어(ko)로 자동 리다이렉트
- [ ] /ko 접근 가능 (한국어 홈 로드)
- [ ] /en, /ja, /zh 접근 가능

**위험 신호 및 대응**:

| 신호 | 심각도 | 대응 |
|------|--------|------|
| "Cannot find module 'nextra'" | CRITICAL | Nextra 4 설치 재확인, node_modules 정리 |
| "Invalid layout configuration" | HIGH | layout.jsx 구조 재검토, Nextra 문서 확인 |
| "Locale not recognized" | HIGH | i18n 설정 재확인, isValidLocale() 함수 테스트 |
| "/ko 페이지 로드 실패" | CRITICAL | pages/ vs content/ 경로 확인 |

---

### Phase 3: 콘텐츠 마이그레이션

**마이그레이션 스크립트 테스트**:
```bash
# 스크립트 실행 전 backup
cp -r docs/pages docs/pages.backup

# 스크립트 실행
node scripts/migrate-nextra-4.js

# 결과 검증
```

- [ ] content/ko/ 디렉토리 생성 확인
- [ ] content/en/ 디렉토리 생성 확인
- [ ] content/ja/ 디렉토리 생성 확인
- [ ] content/zh/ 디렉토리 생성 확인
- [ ] 모든 MDX 파일 복사 완료
  ```bash
  find content/ -name "*.mdx" | wc -l
  # 예상: 100+ 파일
  ```

**메타 파일 검증**:
```bash
# _meta.json 남아있는 파일 확인
find content/ -name "_meta.json"
# 예상: 0개 (모두 meta.json으로 변환)
```

- [ ] 모든 _meta.json을 meta.json으로 이름 변경 완료
- [ ] meta.json 내용 검증
  ```bash
  cat content/ko/meta.json
  # 예상: JSON 형식 (파싱 가능)
  ```

**링크 검증**:
```bash
node scripts/validate-links.js
```

- [ ] 모든 링크 유효함
- [ ] 깨진 링크 0건
- [ ] 경로 오류 0건

**위험 신호 및 대응**:

| 신호 | 심각도 | 대응 |
|------|--------|------|
| "Failed to copy files" | CRITICAL | 디스크 공간 확인, 권한 확인 |
| "meta.json: Invalid JSON" | HIGH | 파일 수동 검사, JSON 형식 수정 |
| "Broken links found" | HIGH | 스크립트로 자동 수정 또는 수동 수정 |
| "Missing index.mdx" | MEDIUM | 해당 섹션 수동 생성 또는 부모 페이지 생성 |

---

### Phase 4: 의존성 업그레이드

**package.json 업데이트**:
- [ ] `next` 버전 확인: ^16.0.0
- [ ] `nextra` 버전 확인: ^4.6.0
- [ ] `react` 버전 확인: ^19.0.0
- [ ] `pagefind` 추가 확인: ^1.1.0

**의존성 설치**:
```bash
npm install
# 또는
uv lock
```

- [ ] 설치 완료 (에러 없음)
- [ ] Peer dependency 경고 0건 (또는 --legacy-peer-deps 사용)
- [ ] package-lock.json 또는 uv.lock 업데이트

**TypeScript 검증**:
```bash
npm run type-check
```

- [ ] 타입 에러 0건
- [ ] 경고 0건 (또는 무시 가능한 수준)

**빌드 시간 측정**:
```bash
time npm run build
```

- [ ] 빌드 성공 (에러 없음)
- [ ] 빌드 시간 기록: ___ 초
- [ ] 이전 대비 성능: ___ % (목표: 50% 개선)

**위험 신호 및 대응**:

| 신호 | 심각도 | 대응 |
|------|--------|------|
| "Peer dependency conflict" | MEDIUM | --legacy-peer-deps 사용 또는 버전 조정 |
| "Type error after upgrade" | HIGH | 타입 정의 업데이트, 호환성 패치 검색 |
| "Build fails with webpack error" | CRITICAL | Turbopack 비활성화 후 재시도 |
| "Build time increased" | MEDIUM | 번들 분석, 코드 스플릿 최적화 |

---

### Phase 5: 검색 엔진 마이그레이션

**.pagefindrc.json 설정**:
- [ ] 파일 생성 완료
- [ ] 모든 언어 설정 추가 (ko, en, ja, zh)
- [ ] CJK 토큰화 설정 추가
- [ ] root_selector 설정 (article)
- [ ] exclude_selectors 설정

**Pagefind 설치 및 테스트**:
```bash
npm install pagefind

# 빌드
npm run build
# 기대: npm run build && pagefind --site out
```

- [ ] 빌드 성공
- [ ] public/pagefind/ 디렉토리 생성 확인
  ```bash
  ls -la public/pagefind/
  # 예상: index.json, pagefind.js, ui.css 등
  ```

**검색 기능 테스트**:
```bash
npm run dev
```

로컬에서 검색 UI 테스트:
- [ ] /ko 페이지에서 검색 바 표시
- [ ] "API" 검색 (한국어 결과 나옴)
- [ ] "installation" 검색 (영어 결과 나옴)
- [ ] 검색 결과 클릭 → 해당 페이지 이동

**다국어 검색 검증**:
- [ ] 한국어 검색: "명령어", "설정" 등 (최소 3개 단어)
- [ ] 영어 검색: "installation", "guide" 등 (최소 3개 단어)
- [ ] 일본어 검색: 테스트 (가능한 경우)
- [ ] 중국어 검색: 테스트 (가능한 경우)

**위험 신호 및 대응**:

| 신호 | 심각도 | 대응 |
|------|--------|------|
| "pagefind command not found" | CRITICAL | npx pagefind 사용 또는 전역 설치 |
| "No index files generated" | HIGH | .pagefindrc.json 설정 재확인, 빌드 디렉토리 확인 |
| "Search returns no results" | HIGH | 인덱싱 로그 확인, 선택자(selector) 재검토 |
| "CJK 검색 작동하지 않음" | MEDIUM | splitting_strategy: "cjk" 설정 확인 |

---

### Phase 6: 빌드 및 배포 설정

**next.config.mjs 검증**:
- [ ] 파일 형식: ESM (import/export)
- [ ] Turbopack 설정 포함:
  ```javascript
  experimental: {
    turbo: {
      enabled: true,
    },
  }
  ```
- [ ] 문법 오류 0건

**Vercel 설정 검토**:
- [ ] 빌드 명령어: `next build && pagefind --site out`
- [ ] 출력 디렉토리: `.next` (기본값)
- [ ] Node.js 버전: 18.x 이상

**로컬 프로덕션 빌드 검증**:
```bash
npm run build
npm start
```

- [ ] 빌드 성공 (에러 0건)
- [ ] http://localhost:3000/ 접근 가능
- [ ] 모든 페이지 로드 가능
- [ ] 검색 기능 동작
- [ ] 다크 모드 전환 동작

**위험 신호 및 대응**:

| 신호 | 심각도 | 대응 |
|------|--------|------|
| "SyntaxError in next.config.mjs" | CRITICAL | 파일 재작성, ESM 문법 확인 |
| "Turbopack compilation failed" | HIGH | 실험적 기능 비활성화, SWC 사용 |
| "next start fails" | CRITICAL | .next 디렉토리 재생성 |

---

### Phase 7: 정적 생성 및 라우팅

**generateStaticParams 검증**:
```bash
npm run build 2>&1 | grep "generated"
```

- [ ] 모든 페이지 정적 생성됨
- [ ] 정적 생성 시간 기록: ___ 초
- [ ] 생성된 페이지 수: 100+ (기대값)

**라우팅 검증**:
```bash
npm run dev
```

테스트 URL 목록:
- [ ] `/` → 리다이렉트 to `/ko` (또는 언어 선택)
- [ ] `/ko` → 한국어 홈 로드
- [ ] `/en` → 영어 홈 로드
- [ ] `/ja` → 일본어 홈 로드
- [ ] `/zh` → 중국어 홈 로드
- [ ] `/ko/guides/alfred/1-plan` → 정상 로드
- [ ] `/en/reference/agents/index` → 정상 로드
- [ ] `/ko/non-existent` → 404 페이지
- [ ] `/invalid-locale/page` → 404 또는 리다이렉트

**위험 신호 및 대응**:

| 신호 | 심각도 | 대응 |
|------|--------|------|
| "Some pages failed to generate" | HIGH | 빌드 로그 상세 검토, generateStaticParams 재확인 |
| "404 페이지가 제대로 표시 안 됨" | MEDIUM | not-found.jsx 파일 생성, 라우트 우선순위 확인 |
| "라우팅 버그: 다른 페이지가 로드됨" | CRITICAL | 라우트 정의 재검토, slug 파싱 로직 확인 |

---

### Phase 8: 메타데이터 및 SEO

**generateMetadata 검증**:
```bash
curl -s http://localhost:3000/ko | grep -i "<meta"
```

- [ ] 모든 페이지에서 메타데이터 표시됨
- [ ] og:title 설정됨
- [ ] og:description 설정됨
- [ ] hreflang alternate 링크 설정됨

**SEO 검증**:
- [ ] /ko 페이지: 한국어 메타데이터
- [ ] /en 페이지: 영어 메타데이터
- [ ] 다국어 버전: hreflang 교차 링크 확인

**위험 신호 및 대응**:

| 신호 | 심각도 | 대응 |
|------|--------|------|
| "메타데이터 없음" | MEDIUM | generateMetadata 함수 미구현 또는 오류 |
| "hreflang 링크 깨짐" | MEDIUM | URL 구성 재확인, locale 변수 검증 |

---

### Phase 9: 성능 최적화

**이미지 최적화 검증**:
- [ ] Next.js Image 컴포넌트 사용
- [ ] 모든 이미지에 width/height 설정
- [ ] WebP 포맷 적용 확인

**코드 분할 검증**:
```bash
npm run build
# .next/static/chunks/ 디렉토리 분석
```

- [ ] 청크 파일 생성됨
- [ ] 청크 크기 적절 (< 100KB)

**Lighthouse 점수**:
```bash
npm install -g lighthouse
npm run build
npm start
# 새 터미널에서:
lighthouse http://localhost:3000/ko --view
```

**기대 점수**:
- [ ] Performance: > 90
- [ ] Accessibility: > 90
- [ ] Best Practices: > 90
- [ ] SEO: > 90

**실제 측정값**:
- Performance: ___ (이전: ___)
- Accessibility: ___ (이전: ___)
- Best Practices: ___ (이전: ___)
- SEO: ___ (이전: ___)

**개선 항목** (필요시):
- [ ] 항목 1: ___
- [ ] 항목 2: ___
- [ ] 항목 3: ___

**위험 신호 및 대응**:

| 신호 | 심각도 | 대응 |
|------|--------|------|
| "LCP > 3.0s" | MEDIUM | 이미지 최적화, 주요 콘텐츠 우선순위 |
| "CLS > 0.1" | MEDIUM | 레이아웃 안정성 개선, 스켈레톤 로더 추가 |
| "Performance < 80" | HIGH | 상세 분석, 코드 스플릿 재검토 |

---

### Phase 10: 통합 테스트 및 QA

**기능 테스트 체크리스트**:

| 기능 | 테스트 | 상태 |
|------|--------|------|
| 다국어 라우팅 | /ko, /en, /ja, /zh 접근 | ☐ PASS |
| 페이지 렌더링 | 100+ MDX 파일 로드 | ☐ PASS |
| 검색 기능 | 각 언어별 검색 실행 | ☐ PASS |
| 링크 검증 | 내부 링크 클릭 | ☐ PASS |
| 외부 링크 | 샘플 5개 링크 클릭 | ☐ PASS |
| 다크 모드 | 테마 전환 | ☐ PASS |
| 사이드바 | 섹션 토글 | ☐ PASS |
| 목차(TOC) | 헤딩 링크 클릭 | ☐ PASS |
| 반응형 | 모바일(375px), 태블릿(768px), 데스크톱 | ☐ PASS |
| 접근성 | keyboard navigation (Tab, Enter) | ☐ PASS |

**Staging 배포 테스트**:
```bash
# staging 브랜치로 푸시
git push origin feature/nextra-4-migration:staging
```

- [ ] Vercel preview URL 생성됨
- [ ] Preview 배포 성공 (URL: ___________________)
- [ ] 모든 기능 테스트 통과 (위 체크리스트 반복)
- [ ] 성능 측정: Lighthouse 점수 (Performance: ___)

**위험 신호 및 대응**:

| 신호 | 심각도 | 대응 |
|------|--------|------|
| "예상치 못한 페이지 렌더링 오류" | CRITICAL | 에러 로그 상세 검토, pages vs content 경로 확인 |
| "특정 언어 페이지 로드 실패" | HIGH | 해당 언어 content 디렉토리 검증 |
| "Staging에서만 발생하는 버그" | MEDIUM | 환경변수, API 엔드포인트 확인 |

---

### Phase 11: 프로덕션 배포

**배포 전 최종 체크**:
- [ ] main 브랜치 백업 생성: `git branch backup/nextjs-14 main`
- [ ] 롤백 절차 문서화 완료
- [ ] Vercel 배포 권한 확인
- [ ] 팀 공지 완료
- [ ] 모니터링 도구 준비 (Sentry, Analytics 등)

**배포 실행**:
```bash
git checkout feature/nextra-4-migration
git checkout -b deploy/nextra-4-migration
git push origin deploy/nextra-4-migration:main
```

- [ ] main 브랜치 push 성공
- [ ] Vercel 배포 시작 감지
- [ ] 배포 로그 모니터링 (Vercel Dashboard)

**배포 진행률**:
- [ ] Build phase: 0-10분
- [ ] Deploy phase: 10-15분
- [ ] Propagation: 15-20분 (CDN 캐시)

**배포 완료 확인**:
```bash
curl -I https://moai-adk.gooslab.ai
# 기대: HTTP/2 200
```

- [ ] 상태 코드: 200 OK
- [ ] Content-Type: text/html
- [ ] 응답 시간: < 500ms

**배포 후 검증** (5분 후):
- [ ] 홈페이지 로드 성공
- [ ] 기본 언어 리다이렉트 동작
- [ ] 검색 기능 작동
- [ ] 콘솔 에러: 0건

**위험 신호 및 대응**:

| 신호 | 심각도 | 대응 |
|------|--------|------|
| "배포 실패: Build Error" | CRITICAL | 빌드 로그 검토, rollback 준비 |
| "배포 성공 후 500 에러" | CRITICAL | Sentry 확인, 롤백 실행 |
| "페이지 깨짐 (CSS/JS 로드 실패)" | HIGH | CDN 캐시 확인, Hard refresh (Cmd+Shift+R) |
| "일부 언어만 오류" | MEDIUM | 해당 언어 content 빌드 재확인 |

---

## 3. 롤백 절차

### 롤백 결정 기준

**즉시 롤백 필요 상황**:
1. 홈페이지 표시 안 됨 (500 에러)
2. 모든 페이지 404 오류
3. 데이터베이스 연결 실패 (있는 경우)
4. 검색 기능 완전 마비
5. 30분 이상 배포 지연

### 롤백 실행

**Option 1: Vercel Dashboard (권장)**
```
1. Vercel Dashboard 접속
2. 프로젝트 선택: moai-adk-docs
3. Deployments 탭 선택
4. 이전 배포(backup/nextjs-14)로 이동
5. "Promote to Production" 클릭
→ 즉시 이전 버전으로 복구 (< 1분)
```

**Option 2: Git 롤백**
```bash
# 이전 main 브랜치 강제 푸시
git checkout backup/nextjs-14
git push origin backup/nextjs-14:main --force

# 또는 특정 commit으로 롤백
git reset --hard <commit-hash>
git push origin main --force
```

**Option 3: 수동 복구** (마지막 수단)
```bash
1. Vercel CLI 설치: npm install -g vercel
2. 이전 deployment ID 확인
3. vercel promote <deployment-id>
4. 확인: https://moai-adk.gooslab.ai
```

### 롤백 후 조치

- [ ] 롤백 완료 확인 (홈페이지 정상 로드)
- [ ] 팀 공지: 롤백 이유 및 진행 상황
- [ ] Root cause analysis:
  - 오류 로그 분석
  - 배포 설정 재검토
  - 코드 문제 파악
- [ ] 문제 해결 후 재배포 계획 수립

**예상 롤백 시간**: 5-10분

---

## 4. 모니터링 계획 (배포 후 24시간)

### 실시간 모니터링 (배포 후 1시간)

**측정 항목**:
- [ ] 페이지 응답 시간: < 500ms
- [ ] 에러율: < 0.1%
- [ ] 검색 응답 시간: < 200ms
- [ ] 방문자 수: 정상 수준

**모니터링 도구**:
- Vercel Analytics (기본 제공)
- Sentry (설정된 경우)
- Google Analytics (설정된 경우)
- 브라우저 DevTools (수동)

**매시간 체크** (배포 후 1-6시간):
```bash
# 상태 확인 스크립트
curl -w "\n%{http_code}\n" https://moai-adk.gooslab.ai/ko
curl -w "\n%{http_code}\n" https://moai-adk.gooslab.ai/en
curl -w "\n%{http_code}\n" https://moai-adk.gooslab.ai/search-test

# 기대: 모두 200
```

- [ ] 0시간: 모든 페이지 정상
- [ ] 1시간: 모든 페이지 정상, 트래픽 정상 수준
- [ ] 2시간: 에러 로그 확인, 문제 없음
- [ ] 3시간: 성능 지표 안정화 확인
- [ ] 6시간: 최종 안정성 확인

### 일일 모니터링 (배포 후 1-7일)

**매일 확인**:
- [ ] 에러율 < 0.1%
- [ ] 검색 성공률 > 99%
- [ ] Core Web Vitals:
  - LCP: < 2.5s
  - FID: < 100ms
  - CLS: < 0.1

**주간 리포트**:
- Lighthouse 점수 측정
- 트래픽 분석
- 성능 비교 (이전 vs 현재)

---

## 5. 성능 비교 기준

### 기대 개선 사항

| 지표 | 현재 | 목표 | 개선도 |
|------|------|------|--------|
| 빌드 시간 | 120초 | 60초 | 50% ↓ |
| LCP | 2.8초 | 2.1초 | 25% ↓ |
| FID | 85ms | 40ms | 53% ↓ |
| CLS | 0.12 | 0.05 | 58% ↓ |
| 번들 크기 | - | -20% | 20% ↓ |
| 검색 응답시간 | 300ms | 150ms | 50% ↓ |

### 측정 방법

**빌드 시간**:
```bash
npm run build > build.log 2>&1
tail -20 build.log | grep -i "build"
```

**LCP, FID, CLS**:
```bash
# Lighthouse CLI
lighthouse https://moai-adk.gooslab.ai/ko --output-path=./lighthouse.html
```

**번들 크기**:
```bash
ls -lh .next/static/chunks/
du -sh .next/
```

---

## 6. 최종 승인 기준

### 배포 승인 조건 (All must be green)

- [ ] Phase 1-10 모두 완료
- [ ] 모든 체크리스트 항목 PASS
- [ ] Staging 배포 성공 및 검증 완료
- [ ] 롤백 계획 수립 및 테스트 완료
- [ ] 팀 승인 완료
- [ ] 배포 시간 최적 (트래픽 최소)

### 배포 후 승인 해제 조건

- [ ] 배포 후 24시간 모니터링 완료
- [ ] 에러율 < 0.1%
- [ ] 모든 기능 정상 동작
- [ ] Core Web Vitals 목표 달성
- [ ] 사용자 피드백 수집 및 검토

---

## 7. 문제 해결 가이드 (Troubleshooting)

### 자주 발생하는 문제

#### 문제 1: "Cannot find module 'content/ko/index.mdx'"

**원인**: MDX 파일이 pages/에 남아있거나 content/에 복사되지 않음

**해결책**:
```bash
# 1. content 디렉토리 확인
ls -la content/ko/

# 2. 파일 수 확인 (100+ 예상)
find content/ -name "*.mdx" | wc -l

# 3. 필요시 수동 복사
cp -r pages/ko/* content/ko/

# 4. pages/ 디렉토리 삭제 여부 확인
# (삭제하지 않았다면 삭제)
```

#### 문제 2: "i18n is not defined"

**원인**: lib/i18n.js를 import하지 않았거나 함수명 오류

**해결책**:
```javascript
// app/[locale]/layout.jsx 확인
import { isValidLocale } from '@/lib/i18n'

// 또는
import * as i18n from '@/lib/i18n'
i18n.isValidLocale(locale)
```

#### 문제 3: "/ko 페이지가 로드되지 않음"

**원인**: 여러 가지 가능 (라우트 정의, MDX 로더, 레이아웃)

**진단 절차**:
```bash
# 1. 콘텐츠 파일 확인
ls -la content/ko/index.mdx

# 2. 레이아웃 파일 확인
ls -la app/[locale]/page.jsx

# 3. 빌드 로그 확인
npm run build 2>&1 | grep -i "error\|ko"

# 4. 라우트 정의 확인
# app/[locale]/page.jsx의 generateStaticParams() 확인
```

#### 문제 4: "Pagefind 인덱스 생성 안 됨"

**원인**: .pagefindrc.json 설정 오류 또는 빌드 디렉토리 잘못됨

**해결책**:
```bash
# 1. 설정 파일 확인
cat .pagefindrc.json | jq .

# 2. 빌드 디렉토리 확인
ls -la out/ (또는 .next/)

# 3. Pagefind 수동 실행
npx pagefind --site out --debug

# 4. 생성된 파일 확인
ls -la public/pagefind/
```

---

## 8. 의사결정 트리

```
배포 진행할 준비가 되었는가?
  ├─ 아니오 (문제 발생)
  │   └─ 문제 해결 가이드 참고
  │       └─ 해결됨? → 다시 배포 준비 확인
  │       └─ 미해결 → 롤백 고려
  │
  ├─ 예 (모두 준비됨)
  │   ├─ Staging 배포 검증 완료?
  │   │   ├─ 예 → main 브랜치로 배포 진행
  │   │   └─ 아니오 → Staging 재테스트
  │   │
  │   └─ main으로 배포
  │       ├─ 배포 성공?
  │       │   ├─ 예 → 24시간 모니터링 시작
  │       │   └─ 아니오 → 즉시 롤백
  │       │
  │       └─ 24시간 모니터링 완료
  │           ├─ 모든 지표 정상?
  │           │   ├─ 예 → 배포 완료, 마이그레이션 종료
  │           │   └─ 아니오 → 문제 분석 및 hotfix
  │           │
  │           └─ 마이그레이션 완료!
```

---

**문서 완성**: 2025-11-10
**다음 단계**: Phase별 검증 실행 및 기록

