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

# SPEC-NEXTRA-I18N-001 Acceptance Criteria

## HISTORY

### v0.0.1 (2025-10-31)
- **INITIAL**: next-intl 다국어 지원 수락 기준 작성
- **AUTHOR**: @GOOS
- **SECTIONS**: Test Scenarios, Quality Gates, Verification Methods

---

## Test Scenarios (Given-When-Then 형식)

### Scenario 1: 한국어 경로 접근 성공
```gherkin
Given: next-intl이 설치되고 설정이 완료됨
  And: `messages/ko.json` 번역 파일이 존재함
When: 사용자가 https://adk.mo.ai.kr/ko/ 에 접속함
Then: 페이지가 한국어로 표시됨
  And: 페이지 타이틀이 "MoAI-ADK 문서"로 표시됨
  And: 네비게이션 메뉴가 한국어로 표시됨 (홈, 문서, API 레퍼런스 등)
```

**검증 방법**:
```bash
# 한국어 페이지 접속
curl https://adk.mo.ai.kr/ko/ | grep "MoAI-ADK 문서"

# 브라우저에서 접속 (수동 확인)
open https://adk.mo.ai.kr/ko/
```

---

### Scenario 2: 영어 경로 접근 성공
```gherkin
Given: next-intl이 설치되고 설정이 완료됨
  And: `messages/en.json` 번역 파일이 존재함
When: 사용자가 https://adk.mo.ai.kr/en/ 에 접속함
Then: 페이지가 영어로 표시됨
  And: 페이지 타이틀이 "MoAI-ADK Documentation"으로 표시됨
  And: 네비게이션 메뉴가 영어로 표시됨 (Home, Documentation, API Reference 등)
```

**검증 방법**:
```bash
# 영어 페이지 접속
curl https://adk.mo.ai.kr/en/ | grep "MoAI-ADK Documentation"

# 브라우저에서 접속 (수동 확인)
open https://adk.mo.ai.kr/en/
```

---

### Scenario 3: 루트 경로 자동 리다이렉트 (한국어 브라우저)
```gherkin
Given: 사용자의 브라우저 언어 설정이 한국어임 (Accept-Language: ko-KR)
When: 사용자가 https://adk.mo.ai.kr/ 에 접속함
Then: 시스템이 /ko/ 경로로 자동 리다이렉트함
  And: 페이지가 한국어로 표시됨
  And: URL이 https://adk.mo.ai.kr/ko/ 로 변경됨
```

**검증 방법**:
```bash
# Accept-Language 헤더를 ko-KR로 설정하여 요청
curl -H "Accept-Language: ko-KR" -I https://adk.mo.ai.kr/ | grep "Location: /ko/"

# 브라우저에서 확인 (브라우저 언어를 한국어로 설정)
```

---

### Scenario 4: 루트 경로 자동 리다이렉트 (영어 브라우저)
```gherkin
Given: 사용자의 브라우저 언어 설정이 영어임 (Accept-Language: en-US)
When: 사용자가 https://adk.mo.ai.kr/ 에 접속함
Then: 시스템이 /en/ 경로로 자동 리다이렉트함
  And: 페이지가 영어로 표시됨
  And: URL이 https://adk.mo.ai.kr/en/ 로 변경됨
```

**검증 방법**:
```bash
# Accept-Language 헤더를 en-US로 설정하여 요청
curl -H "Accept-Language: en-US" -I https://adk.mo.ai.kr/ | grep "Location: /en/"

# 브라우저에서 확인 (브라우저 언어를 영어로 설정)
```

---

### Scenario 5: 루트 경로 기본 언어 폴백 (지원하지 않는 언어)
```gherkin
Given: 사용자의 브라우저 언어 설정이 지원하지 않는 언어임 (Accept-Language: fr-FR)
When: 사용자가 https://adk.mo.ai.kr/ 에 접속함
Then: 시스템이 기본 언어인 /ko/ 경로로 리다이렉트함
  And: 페이지가 한국어로 표시됨
```

**검증 방법**:
```bash
# Accept-Language 헤더를 fr-FR로 설정하여 요청
curl -H "Accept-Language: fr-FR" -I https://adk.mo.ai.kr/ | grep "Location: /ko/"
```

---

### Scenario 6: 언어 전환 드롭다운 동작 (한국어 → 영어)
```gherkin
Given: 사용자가 https://adk.mo.ai.kr/ko/docs 페이지에 있음
When: 사용자가 언어 전환 드롭다운에서 "English"를 선택함
Then: 시스템이 https://adk.mo.ai.kr/en/docs 로 전환함
  And: 페이지가 영어로 표시됨
  And: URL이 /en/docs 로 변경됨
```

**검증 방법**:
```bash
# 브라우저에서 수동 확인:
# 1. https://adk.mo.ai.kr/ko/docs 접속
# 2. 언어 전환 드롭다운 클릭
# 3. "English" 선택
# 4. URL이 /en/docs 로 변경되고 페이지가 영어로 표시되는지 확인
```

---

### Scenario 7: 언어 전환 드롭다운 동작 (영어 → 한국어)
```gherkin
Given: 사용자가 https://adk.mo.ai.kr/en/docs 페이지에 있음
When: 사용자가 언어 전환 드롭다운에서 "한국어"를 선택함
Then: 시스템이 https://adk.mo.ai.kr/ko/docs 로 전환함
  And: 페이지가 한국어로 표시됨
  And: URL이 /ko/docs 로 변경됨
```

**검증 방법**:
```bash
# 브라우저에서 수동 확인:
# 1. https://adk.mo.ai.kr/en/docs 접속
# 2. 언어 전환 드롭다운 클릭
# 3. "한국어" 선택
# 4. URL이 /ko/docs 로 변경되고 페이지가 한국어로 표시되는지 확인
```

---

### Scenario 8: 번역 파일 정상 로드
```gherkin
Given: `messages/ko.json`과 `messages/en.json` 파일이 존재함
When: 개발자가 `npm run build`를 실행함
Then: 빌드가 성공적으로 완료됨
  And: `out/ko/index.html`에 한국어 텍스트가 포함됨
  And: `out/en/index.html`에 영어 텍스트가 포함됨
```

**검증 방법**:
```bash
# 빌드 실행
npm run build

# 한국어 페이지 확인
cat out/ko/index.html | grep "MoAI-ADK 문서"

# 영어 페이지 확인
cat out/en/index.html | grep "MoAI-ADK Documentation"
```

---

### Scenario 9: 번역 누락 시 기본 언어로 폴백
```gherkin
Given: `messages/en.json`에 "Home.newKey" 번역이 누락됨
  And: `messages/ko.json`에 "Home.newKey": "새로운 키" 번역이 존재함
When: 사용자가 `/en/` 페이지에서 `useTranslations('Home')` 훅을 사용하여 "newKey"를 가져옴
Then: 시스템이 기본 언어(한국어)의 번역을 사용하여 "새로운 키"를 표시함
  Or: next-intl 설정에 따라 번역 키가 그대로 표시됨 (예: "Home.newKey")
```

**검증 방법**:
```bash
# 개발 모드에서 확인 (콘솔에 경고 표시)
npm run dev
# 브라우저 콘솔에서 "[next-intl] Missing message: Home.newKey" 경고 확인
```

---

### Scenario 10: 정적 빌드 시 언어별 디렉토리 생성
```gherkin
Given: `generateStaticParams()` 함수가 ['ko', 'en']을 반환함
When: 개발자가 `npm run build`를 실행함
Then: `out/` 디렉토리에 다음 하위 디렉토리가 생성됨:
  - `out/ko/` (한국어 페이지)
  - `out/en/` (영어 페이지)
  And: 각 디렉토리에 `index.html` 파일이 존재함
```

**검증 방법**:
```bash
# 빌드 실행
npm run build

# 디렉토리 구조 확인
ls -la out/ko/
ls -la out/en/

# 인덱스 파일 확인
test -f out/ko/index.html && echo "✅ ko/index.html exists"
test -f out/en/index.html && echo "✅ en/index.html exists"
```

---

## Quality Gates (품질 게이트)

### Gate 1: 언어 라우팅 품질
- ✅ `/ko/` 경로 접근 시 한국어 콘텐츠 표시
- ✅ `/en/` 경로 접근 시 영어 콘텐츠 표시
- ✅ 루트 경로(`/`) 접속 시 브라우저 언어에 따라 자동 리다이렉트
- ✅ 지원하지 않는 언어는 기본 언어(한국어)로 폴백

**자동화 방법**:
```bash
# 테스트 스크립트
#!/bin/bash

# Test 1: 한국어 경로
curl -s https://adk.mo.ai.kr/ko/ | grep -q "MoAI-ADK 문서" && echo "✅ /ko/ OK" || echo "❌ /ko/ FAIL"

# Test 2: 영어 경로
curl -s https://adk.mo.ai.kr/en/ | grep -q "MoAI-ADK Documentation" && echo "✅ /en/ OK" || echo "❌ /en/ FAIL"

# Test 3: 루트 리다이렉트 (한국어)
curl -H "Accept-Language: ko-KR" -I -s https://adk.mo.ai.kr/ | grep -q "Location: /ko/" && echo "✅ Redirect /ko/ OK" || echo "❌ Redirect /ko/ FAIL"

# Test 4: 루트 리다이렉트 (영어)
curl -H "Accept-Language: en-US" -I -s https://adk.mo.ai.kr/ | grep -q "Location: /en/" && echo "✅ Redirect /en/ OK" || echo "❌ Redirect /en/ FAIL"
```

---

### Gate 2: 번역 파일 품질
- ✅ `messages/ko.json`과 `messages/en.json` 파일이 존재함
- ✅ 두 번역 파일의 키 구조가 동일함 (누락된 키 없음)
- ✅ JSON 형식이 올바름 (파싱 오류 없음)
- ✅ 모든 번역 키가 실제로 사용됨 (사용하지 않는 키 없음)

**자동화 방법**:
```bash
# 번역 파일 검증 스크립트 실행
npm run check-translations

# 예상 출력:
# ✅ All translation files are in sync!
```

---

### Gate 3: 언어 전환 UI 품질
- ✅ 언어 전환 드롭다운이 모든 페이지에 표시됨
- ✅ 드롭다운에서 언어 선택 시 즉시 경로가 변경됨 (새로고침 없음)
- ✅ 현재 언어가 드롭다운에서 선택된 상태로 표시됨
- ✅ 언어 전환 시 현재 페이지 경로가 유지됨 (예: `/ko/docs` → `/en/docs`)

**자동화 방법**:
```bash
# E2E 테스트 (Playwright 또는 Cypress 사용)
# 예시: Playwright
test('Language switcher works', async ({ page }) => {
  await page.goto('https://adk.mo.ai.kr/ko/docs');
  await page.selectOption('select[aria-label="언어"]', 'en');
  await expect(page).toHaveURL('https://adk.mo.ai.kr/en/docs');
});
```

---

### Gate 4: 빌드 품질
- ✅ `npm run build` 실행 시 오류 없이 빌드 완료
- ✅ `out/ko/`, `out/en/` 디렉토리가 생성됨
- ✅ 각 디렉토리에 `index.html` 파일이 존재함
- ✅ 빌드 시간 증가가 10% 이내 (NEXTRA-SITE-001 대비)

**자동화 방법**:
```bash
# 빌드 테스트
npm run build 2>&1 | tee build.log

# 오류 확인
grep -c "ERROR" build.log

# 디렉토리 확인
test -d out/ko && echo "✅ out/ko/ exists" || echo "❌ out/ko/ missing"
test -d out/en && echo "✅ out/en/ exists" || echo "❌ out/en/ missing"
```

---

### Gate 5: 배포 품질
- ✅ Vercel 배포 성공 (상태: "Ready")
- ✅ https://adk.mo.ai.kr/ko/ 정상 접속
- ✅ https://adk.mo.ai.kr/en/ 정상 접속
- ✅ 페이지 로딩 시간 < 2초 (초기 방문 기준)

**자동화 방법**:
```bash
# 배포 상태 확인 (Vercel CLI)
vercel ls --scope moai-adk | grep "Ready"

# 사이트 접속 확인
curl -I https://adk.mo.ai.kr/ko/ | grep "HTTP/2 200"
curl -I https://adk.mo.ai.kr/en/ | grep "HTTP/2 200"
```

---

## Verification Methods (검증 방법)

### 1. 로컬 개발 환경 검증
```bash
# Step 1: next-intl 설치 확인
npm list next-intl

# Step 2: 개발 서버 시작
npm run dev

# Step 3: 브라우저 접속
open http://localhost:3000/ko/
open http://localhost:3000/en/

# Step 4: 언어 전환 동작 확인
# - 언어 전환 드롭다운 클릭
# - 다른 언어 선택
# - URL 변경 및 콘텐츠 번역 확인
```

### 2. 번역 파일 검증
```bash
# Step 1: 번역 파일 존재 확인
ls -la messages/ko.json
ls -la messages/en.json

# Step 2: 번역 키 동기화 확인
npm run check-translations

# Step 3: JSON 형식 검증
node -e "JSON.parse(require('fs').readFileSync('messages/ko.json', 'utf-8'))" && echo "✅ ko.json valid"
node -e "JSON.parse(require('fs').readFileSync('messages/en.json', 'utf-8'))" && echo "✅ en.json valid"
```

### 3. 프로덕션 빌드 검증
```bash
# Step 1: 빌드 실행
npm run build

# Step 2: 언어별 디렉토리 확인
ls -la out/ko/
ls -la out/en/

# Step 3: 인덱스 파일 확인
cat out/ko/index.html | grep "MoAI-ADK 문서"
cat out/en/index.html | grep "MoAI-ADK Documentation"
```

### 4. Vercel 배포 검증
```bash
# Step 1: Git 커밋 및 푸시
git add .
git commit -m "feat: implement i18n with next-intl"
git push origin main

# Step 2: Vercel 배포 상태 확인
vercel ls

# Step 3: 배포된 사이트 접속
curl https://adk.mo.ai.kr/ko/
curl https://adk.mo.ai.kr/en/

# Step 4: 브라우저에서 수동 확인
open https://adk.mo.ai.kr/ko/
open https://adk.mo.ai.kr/en/
```

### 5. Accept-Language 헤더 테스트
```bash
# Test 1: 한국어 브라우저
curl -H "Accept-Language: ko-KR" -I https://adk.mo.ai.kr/ | grep "Location: /ko/"

# Test 2: 영어 브라우저
curl -H "Accept-Language: en-US" -I https://adk.mo.ai.kr/ | grep "Location: /en/"

# Test 3: 지원하지 않는 언어 (프랑스어)
curl -H "Accept-Language: fr-FR" -I https://adk.mo.ai.kr/ | grep "Location: /ko/"

# Test 4: Accept-Language 헤더 없음
curl -I https://adk.mo.ai.kr/ | grep "Location: /ko/"
```

---

## Definition of Done (완료 정의)

### SPEC 완료 조건
- ✅ **모든 Test Scenarios 통과**: 10개 시나리오 모두 성공
- ✅ **모든 Quality Gates 통과**: 5개 게이트 모두 충족
- ✅ **문서 작성 완료**:
  - `@DOC:I18N-GUIDE-001`: 다국어 개발 가이드
  - `@DOC:I18N-TROUBLESHOOTING-001`: i18n 트러블슈팅 가이드
- ✅ **Git 커밋 완료**: "feat: implement i18n with next-intl (SPEC-NEXTRA-I18N-001)"
- ✅ **Vercel 배포 성공**: https://adk.mo.ai.kr/ko/, https://adk.mo.ai.kr/en/ 정상 접속
- ✅ **후속 SPEC 준비**: NEXTRA-CONTENT-001 준비 완료

### 검증 체크리스트
```markdown
- [ ] Scenario 1: 한국어 경로 접근 성공
- [ ] Scenario 2: 영어 경로 접근 성공
- [ ] Scenario 3: 루트 경로 자동 리다이렉트 (한국어)
- [ ] Scenario 4: 루트 경로 자동 리다이렉트 (영어)
- [ ] Scenario 5: 기본 언어 폴백 (지원하지 않는 언어)
- [ ] Scenario 6: 언어 전환 (한국어 → 영어)
- [ ] Scenario 7: 언어 전환 (영어 → 한국어)
- [ ] Scenario 8: 번역 파일 정상 로드
- [ ] Scenario 9: 번역 누락 시 폴백
- [ ] Scenario 10: 정적 빌드 시 언어별 디렉토리 생성

- [ ] Gate 1: 언어 라우팅 품질
- [ ] Gate 2: 번역 파일 품질
- [ ] Gate 3: 언어 전환 UI 품질
- [ ] Gate 4: 빌드 품질
- [ ] Gate 5: 배포 품질

- [ ] 다국어 개발 가이드 작성 완료
- [ ] i18n 트러블슈팅 가이드 작성 완료
- [ ] Git 커밋 완료
- [ ] Vercel 배포 완료
```

---

## Rollback Criteria (롤백 기준)

### 롤백이 필요한 경우
1. **언어 라우팅 실패**: `/ko/` 또는 `/en/` 경로 접근 시 404 또는 오류 발생
2. **빌드 실패**: `npm run build` 실행 시 5회 이상 연속 실패
3. **번역 파일 오류**: JSON 파싱 오류 또는 번역 키 누락으로 인한 화이트 스크린
4. **배포 실패**: Vercel 배포가 3회 이상 실패
5. **성능 저하**: 페이지 로딩 시간이 5초 이상

### 롤백 절차
```bash
# Step 1: Vercel 대시보드에서 이전 배포 버전으로 롤백
# https://vercel.com/dashboard → Deployments → "Rollback to this deployment"

# Step 2: 로컬 Git에서 이전 커밋으로 복원
git revert HEAD

# Step 3: next-intl 제거 (필요시)
npm uninstall next-intl

# Step 4: 문제 원인 분석 및 수정 후 재배포
```

---

## Appendix

### Automated Testing Script Example
```bash
#!/bin/bash
# test-nextra-i18n.sh

echo "=== NEXTRA-I18N-001 Automated Testing ==="

# Test 1: 한국어 경로
echo "Test 1: Korean route"
curl -s https://adk.mo.ai.kr/ko/ | grep -q "MoAI-ADK 문서"
if [ $? -eq 0 ]; then
  echo "✅ /ko/ route OK"
else
  echo "❌ /ko/ route FAIL"
  exit 1
fi

# Test 2: 영어 경로
echo "Test 2: English route"
curl -s https://adk.mo.ai.kr/en/ | grep -q "MoAI-ADK Documentation"
if [ $? -eq 0 ]; then
  echo "✅ /en/ route OK"
else
  echo "❌ /en/ route FAIL"
  exit 1
fi

# Test 3: 루트 리다이렉트 (한국어)
echo "Test 3: Root redirect (Korean)"
curl -H "Accept-Language: ko-KR" -I -s https://adk.mo.ai.kr/ | grep -q "Location: /ko/"
if [ $? -eq 0 ]; then
  echo "✅ Root redirect /ko/ OK"
else
  echo "❌ Root redirect /ko/ FAIL"
  exit 1
fi

# Test 4: 루트 리다이렉트 (영어)
echo "Test 4: Root redirect (English)"
curl -H "Accept-Language: en-US" -I -s https://adk.mo.ai.kr/ | grep -q "Location: /en/"
if [ $? -eq 0 ]; then
  echo "✅ Root redirect /en/ OK"
else
  echo "❌ Root redirect /en/ FAIL"
  exit 1
fi

# Test 5: 번역 파일 검증
echo "Test 5: Translation files validation"
npm run check-translations
if [ $? -eq 0 ]; then
  echo "✅ Translation files OK"
else
  echo "❌ Translation files FAIL"
  exit 1
fi

echo "=== All tests passed ==="
```

### Reference Links
- [next-intl Testing Guide](https://next-intl.dev/docs/testing)
- [Next.js 14 i18n Best Practices](https://nextjs.org/docs/app/building-your-application/routing/internationalization)
- [Nextra i18n Documentation](https://nextra.site/docs/guide/i18n)
