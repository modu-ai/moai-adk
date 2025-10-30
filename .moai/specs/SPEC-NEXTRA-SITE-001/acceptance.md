---
id: NEXTRA-SITE-001
version: 0.0.1
status: draft
created: 2025-10-31
updated: 2025-10-31
author: @GOOS
priority: critical
---

# SPEC-NEXTRA-SITE-001 Acceptance Criteria

## HISTORY

### v0.0.1 (2025-10-31)
- **INITIAL**: Nextra 4.0 기본 구조 구축 수락 기준 작성
- **AUTHOR**: @GOOS
- **SECTIONS**: Test Scenarios, Quality Gates, Verification Methods

---

## Test Scenarios (Given-When-Then 형식)

### Scenario 1: 프로젝트 초기화 성공
```gherkin
Given: Node.js 20.x LTS가 설치되어 있음
  And: npm 또는 yarn 패키지 매니저가 설치되어 있음
When: 개발자가 다음 명령을 실행함
  ```
  npm init -y
  npm install next@^14.2.0 nextra@^4.0.0 nextra-theme-docs@^4.0.0 react@^18.2.0 react-dom@^18.2.0
  ```
Then: 모든 의존성이 성공적으로 설치됨
  And: `package.json`에 다음 의존성이 포함됨:
    - `next`: ^14.2.0
    - `nextra`: ^4.0.0
    - `nextra-theme-docs`: ^4.0.0
    - `react`: ^18.2.0
    - `react-dom`: ^18.2.0
  And: `node_modules/` 디렉토리가 생성됨
```

**검증 방법**:
```bash
# package.json에서 의존성 확인
cat package.json | grep -E "(next|nextra|react)"

# node_modules 디렉토리 존재 확인
ls -ld node_modules
```

---

### Scenario 2: 개발 서버 시작 성공
```gherkin
Given: 프로젝트 초기화가 완료됨 (Scenario 1)
  And: `next.config.js`, `theme.config.tsx`, `app/layout.tsx`, `app/page.tsx` 파일이 존재함
When: 개발자가 `npm run dev`를 실행함
Then: 개발 서버가 localhost:3000에서 시작됨
  And: 콘솔에 다음 메시지가 표시됨:
    ```
    ▲ Next.js 14.x.x
    - Local:        http://localhost:3000
    - ready in XXXms
    ```
  And: 브라우저에서 http://localhost:3000에 접속 시 Nextra 홈페이지가 표시됨
  And: 페이지에 "MoAI-ADK Documentation" 로고가 표시됨
```

**검증 방법**:
```bash
# 개발 서버 시작
npm run dev

# 브라우저에서 접속 (수동 확인)
# 또는 curl로 HTML 응답 확인
curl http://localhost:3000 | grep "MoAI-ADK"
```

---

### Scenario 3: Hot Module Replacement (HMR) 동작
```gherkin
Given: 개발 서버가 실행 중임 (Scenario 2)
  And: 브라우저에서 http://localhost:3000이 열려 있음
When: 개발자가 `app/page.tsx` 파일을 수정함 (예: 제목 변경)
Then: 3초 이내에 브라우저가 자동으로 새로고침됨
  And: 변경된 내용이 페이지에 반영됨
  And: 페이지 전체가 새로고침되지 않고 변경된 모듈만 교체됨
```

**검증 방법**:
```bash
# 파일 수정 (예시)
echo "export default function Page() { return <h1>Test HMR</h1>; }" > app/page.tsx

# 브라우저에서 자동 새로고침 확인 (수동)
# 또는 브라우저 개발자 도구에서 WebSocket 연결 확인
```

---

### Scenario 4: 프로덕션 빌드 성공
```gherkin
Given: 프로젝트 초기화 및 파일 구성이 완료됨
  And: `next.config.js`에 `output: 'export'` 설정이 추가됨
When: 개발자가 `npm run build`를 실행함
Then: 빌드가 오류 없이 완료됨
  And: `out/` 디렉토리가 생성됨
  And: `out/` 디렉토리에 다음 파일이 존재함:
    - `index.html` (홈페이지)
    - `_next/` (JavaScript, CSS 번들)
  And: 빌드 시간이 3분 이내로 완료됨
  And: 콘솔에 다음 메시지가 표시됨:
    ```
    ✓ Generating static pages (X/X)
    ✓ Finalizing page optimization
    ```
```

**검증 방법**:
```bash
# 빌드 실행 및 시간 측정
time npm run build

# out 디렉토리 확인
ls -la out/
ls -la out/_next/

# index.html 존재 확인
test -f out/index.html && echo "✅ index.html exists" || echo "❌ index.html missing"
```

---

### Scenario 5: Vercel 배포 성공
```gherkin
Given: 프로덕션 빌드가 성공함 (Scenario 4)
  And: Git 저장소가 초기화되고 코드가 커밋됨
  And: Vercel 프로젝트가 생성되고 Git 저장소가 연동됨
  And: `vercel.json` 파일이 생성됨
When: 개발자가 `git push origin main`을 실행함
Then: Vercel이 자동으로 배포를 시작함
  And: 배포가 성공적으로 완료됨 (Vercel 대시보드에서 "Ready" 상태 표시)
  And: https://adk.mo.ai.kr에서 사이트가 정상 표시됨
  And: 페이지 로딩 시간이 2초 이내
```

**검증 방법**:
```bash
# Git 푸시
git push origin main

# Vercel 배포 상태 확인 (CLI)
vercel ls

# 사이트 접속 확인
curl -I https://adk.mo.ai.kr | grep "HTTP/2 200"

# 브라우저에서 https://adk.mo.ai.kr 접속 (수동 확인)
```

---

### Scenario 6: 도메인(adk.mo.ai.kr) 설정 성공
```gherkin
Given: Vercel 배포가 성공함 (Scenario 5)
  And: DNS에 CNAME 레코드가 설정됨 (adk.mo.ai.kr → cname.vercel-dns.com)
When: 개발자가 https://adk.mo.ai.kr에 접속함
Then: Vercel이 요청을 처리하여 사이트를 표시함
  And: SSL 인증서가 자동으로 적용됨 (HTTPS)
  And: 브라우저 주소창에 자물쇠 아이콘이 표시됨
  And: 페이지가 2초 이내에 로딩됨
```

**검증 방법**:
```bash
# DNS 설정 확인
nslookup adk.mo.ai.kr
dig adk.mo.ai.kr

# SSL 인증서 확인
curl -vI https://adk.mo.ai.kr 2>&1 | grep "SSL certificate verify ok"

# HTTPS 응답 확인
curl -I https://adk.mo.ai.kr | grep "HTTP/2 200"
```

---

### Scenario 7: 보안 헤더 응답 확인
```gherkin
Given: Vercel 배포가 완료됨 (Scenario 5)
  And: `vercel.json`에 보안 헤더가 설정됨
When: 개발자가 https://adk.mo.ai.kr에 HTTP 요청을 보냄
Then: 응답 헤더에 다음 보안 헤더가 포함됨:
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY`
  - `X-XSS-Protection: 1; mode=block`
```

**검증 방법**:
```bash
# 보안 헤더 확인
curl -I https://adk.mo.ai.kr | grep -E "(X-Content-Type-Options|X-Frame-Options|X-XSS-Protection)"

# 예상 출력:
# X-Content-Type-Options: nosniff
# X-Frame-Options: DENY
# X-XSS-Protection: 1; mode=block
```

---

### Scenario 8: Lighthouse 성능 점수 90+ 확인
```gherkin
Given: Vercel 배포가 완료됨 (Scenario 5)
When: 개발자가 Chrome DevTools Lighthouse를 실행함
  And: 모바일 및 데스크톱 모드에서 테스트함
Then: Lighthouse 성능 점수가 90 이상임
  And: First Contentful Paint (FCP)가 1.8초 이내
  And: Largest Contentful Paint (LCP)가 2.5초 이내
  And: Cumulative Layout Shift (CLS)가 0.1 이하
```

**검증 방법**:
```bash
# Lighthouse CLI 설치 (선택사항)
npm install -g lighthouse

# Lighthouse 테스트 실행
lighthouse https://adk.mo.ai.kr --output html --output-path ./lighthouse-report.html

# 또는 Chrome DevTools 사용 (수동):
# 1. Chrome에서 https://adk.mo.ai.kr 접속
# 2. F12 → Lighthouse 탭 → "Generate report" 클릭
# 3. Performance 점수 확인
```

---

### Scenario 9: 반응형 디자인 확인 (모바일/태블릿/데스크톱)
```gherkin
Given: Vercel 배포가 완료됨 (Scenario 5)
When: 개발자가 다음 화면 크기에서 사이트를 확인함:
  - 모바일: 375px x 667px (iPhone SE)
  - 태블릿: 768px x 1024px (iPad)
  - 데스크톱: 1920px x 1080px (FHD)
Then: 모든 화면 크기에서 레이아웃이 올바르게 표시됨
  And: 텍스트가 읽기 쉬운 크기로 표시됨
  And: 네비게이션 메뉴가 화면 크기에 맞게 조정됨 (모바일: 햄버거 메뉴)
  And: 이미지가 화면 크기에 맞게 조정됨
```

**검증 방법**:
```bash
# Chrome DevTools 사용 (수동):
# 1. Chrome에서 https://adk.mo.ai.kr 접속
# 2. F12 → Toggle device toolbar (Ctrl+Shift+M)
# 3. 다양한 디바이스 프리셋 선택 (iPhone SE, iPad, Responsive)
# 4. 레이아웃 확인

# 또는 BrowserStack 같은 크로스 브라우저 테스트 도구 사용
```

---

### Scenario 10: 빌드 시간 3분 이내 확인
```gherkin
Given: 프로젝트 구성이 완료됨
When: 개발자가 `npm run build`를 실행함
Then: 빌드가 3분(180초) 이내에 완료됨
  And: 콘솔에 빌드 완료 메시지가 표시됨
```

**검증 방법**:
```bash
# 빌드 시간 측정
time npm run build

# 예상 출력:
# real    1m30.456s  (< 3분이면 성공)
```

---

## Quality Gates (품질 게이트)

### Gate 1: 빌드 프로세스 품질
- ✅ `npm run build` 실행 시 오류 0개
- ✅ 빌드 경고 < 5개 (허용 가능한 경고만)
- ✅ 빌드 시간 < 3분
- ✅ `out/` 디렉토리에 최소 10개 이상의 파일 생성 (HTML, CSS, JS)

**자동화 방법**:
```bash
# CI/CD에서 실행
npm run build 2>&1 | tee build.log

# 오류 카운트
grep -c "ERROR" build.log

# 빌드 시간 체크
if [ $(date +%s -d "$(grep 'real' build.log)") -lt 180 ]; then
  echo "✅ Build time OK"
else
  echo "❌ Build time exceeded 3 minutes"
fi
```

---

### Gate 2: 성능 품질
- ✅ Lighthouse 성능 점수 ≥ 90 (모바일/데스크톱)
- ✅ First Contentful Paint (FCP) ≤ 1.8초
- ✅ Largest Contentful Paint (LCP) ≤ 2.5초
- ✅ Time to Interactive (TTI) ≤ 3.8초
- ✅ Cumulative Layout Shift (CLS) ≤ 0.1

**자동화 방법**:
```bash
# Lighthouse CI 설정
npm install -g @lhci/cli

# Lighthouse 테스트 실행
lhci autorun --collect.url=https://adk.mo.ai.kr

# 점수 확인 (JSON 출력)
```

---

### Gate 3: 보안 품질
- ✅ 모든 HTTP 응답에 보안 헤더 포함:
  - `X-Content-Type-Options: nosniff`
  - `X-Frame-Options: DENY`
  - `X-XSS-Protection: 1; mode=block`
- ✅ HTTPS 강제 (HTTP → HTTPS 리다이렉트)
- ✅ SSL 인증서 유효 (Vercel 자동 발급)

**자동화 방법**:
```bash
# 보안 헤더 확인 스크립트
#!/bin/bash
headers=$(curl -I https://adk.mo.ai.kr 2>/dev/null)

if echo "$headers" | grep -q "X-Content-Type-Options: nosniff"; then
  echo "✅ X-Content-Type-Options OK"
else
  echo "❌ X-Content-Type-Options missing"
fi

if echo "$headers" | grep -q "X-Frame-Options: DENY"; then
  echo "✅ X-Frame-Options OK"
else
  echo "❌ X-Frame-Options missing"
fi
```

---

### Gate 4: 접근성 품질
- ✅ ARIA 레이블이 모든 인터랙티브 요소에 적용됨
- ✅ 키보드 네비게이션 지원 (Tab/Enter 키로 모든 기능 접근 가능)
- ✅ 색상 대비 WCAG AA 기준 충족 (최소 4.5:1)
- ✅ 스크린 리더 호환성 확인 (NVDA 또는 VoiceOver)

**자동화 방법**:
```bash
# axe-core CLI 사용
npm install -g @axe-core/cli

# 접근성 테스트 실행
axe https://adk.mo.ai.kr

# 또는 Lighthouse 접근성 점수 확인
lighthouse https://adk.mo.ai.kr --only-categories=accessibility
```

---

### Gate 5: 배포 품질
- ✅ Vercel 배포 성공 (상태: "Ready")
- ✅ 도메인(adk.mo.ai.kr) 정상 접속
- ✅ SSL 인증서 유효
- ✅ 페이지 로딩 시간 < 2초 (초기 방문 기준)

**자동화 방법**:
```bash
# 배포 상태 확인 (Vercel CLI)
vercel ls --scope moai-adk | grep "Ready"

# 사이트 접속 확인
curl -I https://adk.mo.ai.kr | grep "HTTP/2 200"

# 페이지 로딩 시간 측정
curl -w "@curl-format.txt" -o /dev/null -s https://adk.mo.ai.kr

# curl-format.txt 내용:
# time_total: %{time_total}s\n
```

---

## Verification Methods (검증 방법)

### 1. 로컬 개발 환경 검증
```bash
# Step 1: 의존성 설치 확인
npm list --depth=0 | grep -E "(next|nextra|react)"

# Step 2: 개발 서버 시작
npm run dev

# Step 3: 브라우저 접속
# http://localhost:3000

# Step 4: HMR 동작 확인
# app/page.tsx 파일 수정 → 브라우저 자동 새로고침 확인
```

### 2. 프로덕션 빌드 검증
```bash
# Step 1: 빌드 실행
npm run build

# Step 2: 빌드 결과 확인
ls -la out/

# Step 3: 로컬 프리뷰 (선택사항)
npx serve out

# Step 4: 브라우저에서 http://localhost:5000 접속
```

### 3. Vercel 배포 검증
```bash
# Step 1: Git 커밋 및 푸시
git add .
git commit -m "feat: initial Nextra setup"
git push origin main

# Step 2: Vercel 대시보드에서 배포 상태 확인
# https://vercel.com/dashboard

# Step 3: 배포된 사이트 접속
# https://adk.mo.ai.kr

# Step 4: Lighthouse 테스트 실행
lighthouse https://adk.mo.ai.kr --view
```

### 4. 성능 검증
```bash
# Lighthouse CLI 실행
lighthouse https://adk.mo.ai.kr \
  --output html \
  --output json \
  --output-path ./lighthouse-report \
  --view

# 성능 점수 확인 (JSON 파싱)
cat lighthouse-report.json | jq '.categories.performance.score * 100'
```

### 5. 보안 검증
```bash
# 보안 헤더 확인
curl -I https://adk.mo.ai.kr | grep -E "(X-Content-Type-Options|X-Frame-Options|X-XSS-Protection)"

# SSL 인증서 확인
openssl s_client -connect adk.mo.ai.kr:443 -servername adk.mo.ai.kr < /dev/null 2>&1 | grep "Verify return code"
```

### 6. 접근성 검증
```bash
# axe-core CLI 실행
axe https://adk.mo.ai.kr --save ./axe-report.json

# 위반 사항 확인
cat axe-report.json | jq '.violations'
```

---

## Definition of Done (완료 정의)

### SPEC 완료 조건
- ✅ **모든 Test Scenarios 통과**: 10개 시나리오 모두 성공
- ✅ **모든 Quality Gates 통과**: 5개 게이트 모두 충족
- ✅ **문서 작성 완료**:
  - `DOC:SITE-GUIDE-001`: 개발 가이드
  - `DOC:SITE-DEPLOY-001`: 배포 가이드
- ✅ **Git 커밋 완료**: "feat: implement Nextra 4.0 basic structure (SPEC-NEXTRA-SITE-001)"
- ✅ **Vercel 배포 성공**: https://adk.mo.ai.kr 정상 접속
- ✅ **후속 SPEC 준비**: NEXTRA-I18N-001, NEXTRA-CONTENT-001 준비 완료

### 검증 체크리스트
```markdown
- [ ] Scenario 1: 프로젝트 초기화 성공
- [ ] Scenario 2: 개발 서버 시작 성공
- [ ] Scenario 3: HMR 동작 확인
- [ ] Scenario 4: 프로덕션 빌드 성공
- [ ] Scenario 5: Vercel 배포 성공
- [ ] Scenario 6: 도메인 설정 성공
- [ ] Scenario 7: 보안 헤더 응답 확인
- [ ] Scenario 8: Lighthouse 성능 점수 90+
- [ ] Scenario 9: 반응형 디자인 확인
- [ ] Scenario 10: 빌드 시간 3분 이내

- [ ] Gate 1: 빌드 프로세스 품질
- [ ] Gate 2: 성능 품질
- [ ] Gate 3: 보안 품질
- [ ] Gate 4: 접근성 품질
- [ ] Gate 5: 배포 품질

- [ ] 개발 가이드 작성 완료
- [ ] 배포 가이드 작성 완료
- [ ] Git 커밋 완료
- [ ] Vercel 배포 완료
```

---

## Rollback Criteria (롤백 기준)

### 롤백이 필요한 경우
1. **빌드 실패**: 5회 이상 연속 빌드 실패
2. **배포 실패**: Vercel 배포가 3회 이상 실패
3. **성능 저하**: Lighthouse 점수가 70 이하
4. **보안 문제**: 보안 헤더 누락 또는 SSL 인증서 오류
5. **사이트 접속 불가**: 5분 이상 adk.mo.ai.kr 접속 불가

### 롤백 절차
```bash
# Step 1: Vercel 대시보드에서 이전 배포 버전으로 롤백
# https://vercel.com/dashboard → Deployments → "Rollback to this deployment"

# Step 2: 로컬 Git에서 이전 커밋으로 복원
git revert HEAD

# Step 3: 문제 원인 분석 및 수정 후 재배포
```

---

## Appendix

### Automated Testing Script Example
```bash
#!/bin/bash
# test-nextra-site.sh

echo "=== NEXTRA-SITE-001 Automated Testing ==="

# Test 1: Build process
echo "Test 1: Build process"
npm run build
if [ $? -eq 0 ]; then
  echo "✅ Build succeeded"
else
  echo "❌ Build failed"
  exit 1
fi

# Test 2: Output directory
echo "Test 2: Output directory"
if [ -d "out" ]; then
  echo "✅ out/ directory exists"
else
  echo "❌ out/ directory missing"
  exit 1
fi

# Test 3: Index file
echo "Test 3: Index file"
if [ -f "out/index.html" ]; then
  echo "✅ index.html exists"
else
  echo "❌ index.html missing"
  exit 1
fi

# Test 4: Security headers (requires deployed URL)
echo "Test 4: Security headers"
curl -I https://adk.mo.ai.kr | grep "X-Content-Type-Options: nosniff"
if [ $? -eq 0 ]; then
  echo "✅ Security headers OK"
else
  echo "❌ Security headers missing"
  exit 1
fi

echo "=== All tests passed ==="
```

### Reference Links
- [Nextra Testing Guide](https://nextra.site/docs/guide/testing)
- [Next.js Testing Documentation](https://nextjs.org/docs/testing)
- [Vercel Deployment Checklist](https://vercel.com/docs/deployments/overview)
- [Web Vitals Reference](https://web.dev/vitals/)
