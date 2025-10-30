---
id: NEXTRA-CONTENT-001
version: 0.0.1
status: draft
created: 2025-10-31
updated: 2025-10-31
author: @GOOS
priority: high
depends_on:
  - NEXTRA-SITE-001
  - NEXTRA-I18N-001
---

# SPEC-NEXTRA-CONTENT-001 Acceptance Criteria

## HISTORY

### v0.0.1 (2025-10-31)
- **INITIAL**: mkdocs → Nextra 콘텐츠 마이그레이션 수락 기준 작성
- **AUTHOR**: @GOOS
- **SECTIONS**: Test Scenarios, Quality Gates, Verification Methods

---

## Test Scenarios (Given-When-Then 형식)

### Scenario 1: 40개 페이지 변환 성공
```gherkin
Given: 원본 `/docs/` 디렉토리에 70개 마크다운 파일이 존재함
  And: 마이그레이션 대상 40개 페이지 목록이 정의됨
When: 마이그레이션 스크립트 `node scripts/migrate-content.js`를 실행함
Then: `/pages/ko/` 디렉토리에 40개 MDX 파일이 생성됨
  And: `/pages/en/` 디렉토리에 40개 MDX 파일이 생성됨
  And: 총 80개 MDX 파일이 생성됨
```

**검증 방법**:
```bash
# 페이지 수 확인
find pages/ko -name "*.mdx" | wc -l  # 예상: 40
find pages/en -name "*.mdx" | wc -l  # 예상: 40

# 전체 MDX 파일 수
find pages -name "*.mdx" | wc -l  # 예상: 80
```

---

### Scenario 2: Front Matter 추가 성공
```gherkin
Given: 페이지 변환이 완료됨 (Scenario 1)
When: 변환된 MDX 파일을 검사함
Then: 모든 MDX 파일 상단에 YAML Front Matter가 존재함
  And: Front Matter에 `title` 필드가 포함됨
  And: Front Matter에 `description` 필드가 포함됨 (옵션)
```

**검증 방법**:
```bash
# Front Matter 존재 확인 (첫 줄이 ---로 시작)
head -n 1 pages/ko/index.mdx | grep -q "^---" && echo "✅ Front Matter exists"

# title 필드 확인
grep -q "^title:" pages/ko/index.mdx && echo "✅ title field exists"
```

---

### Scenario 3: 이미지 경로 절대 경로로 수정
```gherkin
Given: 원본 마크다운 파일에 상대 경로 이미지가 포함됨 (예: `../images/logo.png`)
When: 마이그레이션 스크립트가 실행됨
Then: 변환된 MDX 파일의 모든 이미지 경로가 절대 경로로 변경됨 (예: `/images/logo.png`)
  And: 상대 경로 이미지 (`../images/`)가 0개임
```

**검증 방법**:
```bash
# 상대 경로 이미지 확인 (있으면 안 됨)
grep -r "\!\[.*\](\.\./" pages/ko/ && echo "❌ Relative image paths found" || echo "✅ No relative image paths"
grep -r "\!\[.*\](\.\./" pages/en/ && echo "❌ Relative image paths found" || echo "✅ No relative image paths"
```

---

### Scenario 4: 내부 링크 절대 경로로 수정
```gherkin
Given: 원본 마크다운 파일에 상대 경로 내부 링크가 포함됨 (예: `[Quick Start](../getting-started/quickstart.md)`)
When: 마이그레이션 스크립트가 실행됨
Then: 변환된 MDX 파일의 모든 내부 링크가 절대 경로로 변경됨 (예: `[Quick Start](/ko/getting-started/quick-start)`)
  And: 상대 경로 내부 링크 (`../`)가 0개임
  And: `.md` 확장자가 제거됨
```

**검증 방법**:
```bash
# 상대 경로 링크 확인 (있으면 안 됨)
grep -r "\[.*\](\.\./" pages/ko/ && echo "❌ Relative links found" || echo "✅ No relative links"
grep -r "\[.*\](\.\./" pages/en/ && echo "❌ Relative links found" || echo "✅ No relative links"

# .md 확장자 확인 (있으면 안 됨)
grep -r "\.md)" pages/ko/ && echo "❌ .md extension found" || echo "✅ No .md extensions"
```

---

### Scenario 5: `_meta.json` 파일 생성 성공
```gherkin
Given: 페이지 변환이 완료됨 (Scenario 1)
When: `_meta.json` 파일을 생성함
Then: `/pages/ko/_meta.json` 파일이 존재함
  And: `/pages/en/_meta.json` 파일이 존재함
  And: 각 하위 디렉토리에도 `_meta.json` 파일이 존재함
  And: `_meta.json`에 네비게이션 메뉴 구조가 정의됨
```

**검증 방법**:
```bash
# _meta.json 파일 존재 확인
test -f pages/ko/_meta.json && echo "✅ pages/ko/_meta.json exists"
test -f pages/en/_meta.json && echo "✅ pages/en/_meta.json exists"

# 하위 디렉토리 _meta.json 확인
find pages/ko -name "_meta.json" | wc -l  # 예상: 10+ (디렉토리 수에 따라)
```

---

### Scenario 6: Mermaid 다이어그램 코드 블록 유지
```gherkin
Given: 원본 마크다운 파일에 Mermaid 다이어그램이 포함됨
When: 마이그레이션 스크립트가 실행됨
Then: 변환된 MDX 파일에 Mermaid 코드 블록이 유지됨
  And: 코드 블록 형식이 동일함 (```mermaid ... ```)
```

**검증 방법**:
```bash
# Mermaid 코드 블록 확인
grep -r "```mermaid" pages/ko/ && echo "✅ Mermaid diagrams preserved"
grep -r "```mermaid" pages/en/ && echo "✅ Mermaid diagrams preserved"
```

---

### Scenario 7: 프로덕션 빌드 성공
```gherkin
Given: 40개 페이지가 변환됨 (Scenario 1)
  And: 이미지 경로와 내부 링크가 수정됨 (Scenario 3, 4)
  And: `_meta.json` 파일이 생성됨 (Scenario 5)
When: 개발자가 `npm run build`를 실행함
Then: 빌드가 오류 없이 완료됨
  And: `out/ko/` 디렉토리에 40개 HTML 페이지가 생성됨
  And: `out/en/` 디렉토리에 40개 HTML 페이지가 생성됨
```

**검증 방법**:
```bash
# 빌드 실행
npm run build

# 빌드 오류 확인
if [ $? -eq 0 ]; then
  echo "✅ Build succeeded"
else
  echo "❌ Build failed"
  exit 1
fi

# HTML 파일 수 확인
find out/ko -name "*.html" | wc -l  # 예상: 40+
find out/en -name "*.html" | wc -l  # 예상: 40+
```

---

### Scenario 8: Vercel 배포 성공
```gherkin
Given: 프로덕션 빌드가 성공함 (Scenario 7)
  And: Git 커밋 및 푸시가 완료됨
When: Vercel이 자동 배포를 시작함
Then: 배포가 성공적으로 완료됨 (상태: "Ready")
  And: https://adk.mo.ai.kr/ko/ 에서 모든 페이지가 정상 표시됨
  And: https://adk.mo.ai.kr/en/ 에서 모든 페이지가 정상 표시됨
```

**검증 방법**:
```bash
# Vercel 배포 상태 확인
vercel ls --scope moai-adk | grep "Ready"

# 사이트 접속 확인
curl -I https://adk.mo.ai.kr/ko/ | grep "HTTP/2 200"
curl -I https://adk.mo.ai.kr/en/ | grep "HTTP/2 200"

# 브라우저에서 수동 확인
open https://adk.mo.ai.kr/ko/
open https://adk.mo.ai.kr/en/
```

---

### Scenario 9: 모든 페이지 접근 가능 (404 오류 0개)
```gherkin
Given: Vercel 배포가 완료됨 (Scenario 8)
When: 개발자가 모든 40개 페이지에 수동으로 접속함
Then: 모든 페이지가 정상 표시됨 (HTTP 200)
  And: 404 오류가 발생한 페이지가 0개임
```

**검증 방법**:
```bash
# 모든 페이지 URL 목록 생성
cat > check-urls.txt << 'EOF'
https://adk.mo.ai.kr/ko/
https://adk.mo.ai.kr/ko/introduction/overview
https://adk.mo.ai.kr/ko/getting-started/installation
... (나머지 37개 URL)
EOF

# 모든 URL 상태 확인
while read url; do
  status=$(curl -o /dev/null -s -w "%{http_code}" "$url")
  if [ "$status" -eq 200 ]; then
    echo "✅ $url"
  else
    echo "❌ $url (HTTP $status)"
  fi
done < check-urls.txt
```

---

### Scenario 10: 이미지 정상 로딩
```gherkin
Given: Vercel 배포가 완료됨 (Scenario 8)
When: 개발자가 이미지가 포함된 페이지에 접속함
Then: 모든 이미지가 정상 로딩됨
  And: 이미지 404 오류가 0개임
```

**검증 방법**:
```bash
# 브라우저 개발자 도구에서 확인:
# 1. F12 → Network 탭
# 2. 페이지 새로고침
# 3. 이미지 리소스 상태 확인 (모두 200이어야 함)

# 또는 curl로 이미지 경로 확인
curl -I https://adk.mo.ai.kr/images/logo.png | grep "HTTP/2 200"
```

---

## Quality Gates (품질 게이트)

### Gate 1: 페이지 변환 품질
- ✅ 한국어 페이지 40개 변환 완료
- ✅ 영어 페이지 40개 변환 완료
- ✅ 총 80개 MDX 파일 생성
- ✅ 모든 MDX 파일에 Front Matter 포함

**자동화 방법**:
```bash
#!/bin/bash

KO_COUNT=$(find pages/ko -name "*.mdx" | wc -l)
EN_COUNT=$(find pages/en -name "*.mdx" | wc -l)

if [ "$KO_COUNT" -eq 40 ] && [ "$EN_COUNT" -eq 40 ]; then
  echo "✅ Gate 1: Page conversion quality PASS"
else
  echo "❌ Gate 1: Page conversion quality FAIL (KO: $KO_COUNT, EN: $EN_COUNT)"
  exit 1
fi
```

---

### Gate 2: 경로 수정 품질
- ✅ 이미지 상대 경로 0개 (절대 경로로 수정 완료)
- ✅ 내부 링크 상대 경로 0개 (절대 경로로 수정 완료)
- ✅ `.md` 확장자 0개 (제거 완료)

**자동화 방법**:
```bash
#!/bin/bash

# 이미지 상대 경로 확인
IMG_KO=$(grep -r "\!\[.*\](\.\./" pages/ko/ | wc -l)
IMG_EN=$(grep -r "\!\[.*\](\.\./" pages/en/ | wc -l)

# 내부 링크 상대 경로 확인
LINK_KO=$(grep -r "\[.*\](\.\./" pages/ko/ | wc -l)
LINK_EN=$(grep -r "\[.*\](\.\./" pages/en/ | wc -l)

if [ "$IMG_KO" -eq 0 ] && [ "$IMG_EN" -eq 0 ] && [ "$LINK_KO" -eq 0 ] && [ "$LINK_EN" -eq 0 ]; then
  echo "✅ Gate 2: Path correction quality PASS"
else
  echo "❌ Gate 2: Path correction quality FAIL"
  exit 1
fi
```

---

### Gate 3: 빌드 품질
- ✅ `npm run build` 실행 시 오류 0개
- ✅ 빌드 경고 < 5개
- ✅ `out/ko/`, `out/en/` 디렉토리 생성 확인
- ✅ 빌드 시간 < 5분

**자동화 방법**:
```bash
#!/bin/bash

# 빌드 실행
npm run build 2>&1 | tee build.log

# 오류 확인
ERROR_COUNT=$(grep -c "ERROR" build.log)

if [ "$ERROR_COUNT" -eq 0 ] && [ -d "out/ko" ] && [ -d "out/en" ]; then
  echo "✅ Gate 3: Build quality PASS"
else
  echo "❌ Gate 3: Build quality FAIL (Errors: $ERROR_COUNT)"
  exit 1
fi
```

---

### Gate 4: 배포 품질
- ✅ Vercel 배포 성공 (상태: "Ready")
- ✅ https://adk.mo.ai.kr/ko/ 정상 접속 (HTTP 200)
- ✅ https://adk.mo.ai.kr/en/ 정상 접속 (HTTP 200)
- ✅ 페이지 로딩 시간 < 2초

**자동화 방법**:
```bash
#!/bin/bash

# 한국어 사이트 접속 확인
KO_STATUS=$(curl -o /dev/null -s -w "%{http_code}" https://adk.mo.ai.kr/ko/)

# 영어 사이트 접속 확인
EN_STATUS=$(curl -o /dev/null -s -w "%{http_code}" https://adk.mo.ai.kr/en/)

if [ "$KO_STATUS" -eq 200 ] && [ "$EN_STATUS" -eq 200 ]; then
  echo "✅ Gate 4: Deployment quality PASS"
else
  echo "❌ Gate 4: Deployment quality FAIL (KO: $KO_STATUS, EN: $EN_STATUS)"
  exit 1
fi
```

---

### Gate 5: 링크 유효성 품질
- ✅ 404 오류 페이지 0개
- ✅ 이미지 404 오류 0개
- ✅ 내부 링크 404 오류 0개

**자동화 방법**:
```bash
#!/bin/bash

# 모든 페이지 URL 체크 (예시: 샘플 5개)
URLS=(
  "https://adk.mo.ai.kr/ko/"
  "https://adk.mo.ai.kr/ko/introduction/overview"
  "https://adk.mo.ai.kr/ko/getting-started/installation"
  "https://adk.mo.ai.kr/en/"
  "https://adk.mo.ai.kr/en/introduction/overview"
)

ERROR_COUNT=0
for url in "${URLS[@]}"; do
  status=$(curl -o /dev/null -s -w "%{http_code}" "$url")
  if [ "$status" -ne 200 ]; then
    echo "❌ $url (HTTP $status)"
    ERROR_COUNT=$((ERROR_COUNT + 1))
  fi
done

if [ "$ERROR_COUNT" -eq 0 ]; then
  echo "✅ Gate 5: Link validity quality PASS"
else
  echo "❌ Gate 5: Link validity quality FAIL (Errors: $ERROR_COUNT)"
  exit 1
fi
```

---

## Verification Methods (검증 방법)

### 1. 로컬 개발 환경 검증
```bash
# Step 1: 마이그레이션 스크립트 실행
node scripts/migrate-content.js

# Step 2: 페이지 수 확인
find pages/ko -name "*.mdx" | wc -l  # 예상: 40
find pages/en -name "*.mdx" | wc -l  # 예상: 40

# Step 3: 이미지 경로 검증
bash scripts/validate-image-paths.sh

# Step 4: 내부 링크 검증
bash scripts/validate-internal-links.sh

# Step 5: 개발 서버 시작
npm run dev

# Step 6: 브라우저 접속
open http://localhost:3000/ko/
open http://localhost:3000/en/
```

### 2. 프로덕션 빌드 검증
```bash
# Step 1: 빌드 실행
npm run build

# Step 2: 빌드 결과 확인
ls -la out/ko/
ls -la out/en/

# Step 3: HTML 파일 수 확인
find out/ko -name "*.html" | wc -l  # 예상: 40+
find out/en -name "*.html" | wc -l  # 예상: 40+
```

### 3. Vercel 배포 검증
```bash
# Step 1: Git 커밋 및 푸시
git add pages/ public/
git commit -m "feat: migrate 40 core pages from mkdocs to Nextra"
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

### 4. 링크 유효성 검증
```bash
# 모든 페이지 URL 체크 스크립트 실행
bash scripts/check-all-urls.sh

# 이미지 로딩 확인 (브라우저 개발자 도구)
# F12 → Network 탭 → 페이지 새로고침 → 이미지 상태 확인
```

---

## Definition of Done (완료 정의)

### SPEC 완료 조건
- ✅ **모든 Test Scenarios 통과**: 10개 시나리오 모두 성공
- ✅ **모든 Quality Gates 통과**: 5개 게이트 모두 충족
- ✅ **문서 작성 완료**:
  - `@DOC:CONTENT-MIGRATION-001`: 콘텐츠 마이그레이션 가이드
  - `@DOC:CONTENT-STYLE-001`: MDX 작성 스타일 가이드
- ✅ **Git 커밋 완료**: "feat: migrate 40 core pages from mkdocs to Nextra (SPEC-NEXTRA-CONTENT-001)"
- ✅ **Vercel 배포 성공**: 모든 페이지 정상 접속
- ✅ **변환 보고서 생성**: 성공/실패/경고 목록

### 검증 체크리스트
```markdown
- [ ] Scenario 1: 40개 페이지 변환 성공
- [ ] Scenario 2: Front Matter 추가 성공
- [ ] Scenario 3: 이미지 경로 절대 경로로 수정
- [ ] Scenario 4: 내부 링크 절대 경로로 수정
- [ ] Scenario 5: _meta.json 파일 생성 성공
- [ ] Scenario 6: Mermaid 다이어그램 유지
- [ ] Scenario 7: 프로덕션 빌드 성공
- [ ] Scenario 8: Vercel 배포 성공
- [ ] Scenario 9: 모든 페이지 접근 가능 (404 오류 0개)
- [ ] Scenario 10: 이미지 정상 로딩

- [ ] Gate 1: 페이지 변환 품질
- [ ] Gate 2: 경로 수정 품질
- [ ] Gate 3: 빌드 품질
- [ ] Gate 4: 배포 품질
- [ ] Gate 5: 링크 유효성 품질

- [ ] 콘텐츠 마이그레이션 가이드 작성 완료
- [ ] MDX 작성 스타일 가이드 작성 완료
- [ ] Git 커밋 완료
- [ ] Vercel 배포 완료
- [ ] 변환 보고서 생성 완료
```

---

## Rollback Criteria (롤백 기준)

### 롤백이 필요한 경우
1. **페이지 변환 실패**: 40개 페이지 중 10개 이상 변환 실패
2. **빌드 실패**: `npm run build` 실행 시 5회 이상 연속 실패
3. **404 오류 다수**: 배포 후 10개 이상 페이지에서 404 오류 발생
4. **이미지 깨짐**: 이미지 404 오류가 20% 이상 발생
5. **사이트 접속 불가**: 5분 이상 adk.mo.ai.kr 접속 불가

### 롤백 절차
```bash
# Step 1: Vercel 대시보드에서 이전 배포 버전으로 롤백
# https://vercel.com/dashboard → Deployments → "Rollback to this deployment"

# Step 2: 로컬 Git에서 이전 커밋으로 복원
git revert HEAD

# Step 3: 원본 백업에서 복원 (필요시)
git checkout backup-branch -- docs/

# Step 4: 문제 원인 분석 및 수정 후 재마이그레이션
```

---

## Appendix

### Automated Testing Script Example
```bash
#!/bin/bash
# test-nextra-content.sh

echo "=== NEXTRA-CONTENT-001 Automated Testing ==="

# Test 1: 페이지 수 확인
echo "Test 1: Page count"
KO_COUNT=$(find pages/ko -name "*.mdx" | wc -l)
EN_COUNT=$(find pages/en -name "*.mdx" | wc -l)
if [ "$KO_COUNT" -eq 40 ] && [ "$EN_COUNT" -eq 40 ]; then
  echo "✅ Page count OK (KO: $KO_COUNT, EN: $EN_COUNT)"
else
  echo "❌ Page count FAIL (KO: $KO_COUNT, EN: $EN_COUNT)"
  exit 1
fi

# Test 2: 이미지 경로 검증
echo "Test 2: Image paths"
IMG_KO=$(grep -r "\!\[.*\](\.\./" pages/ko/ | wc -l)
IMG_EN=$(grep -r "\!\[.*\](\.\./" pages/en/ | wc -l)
if [ "$IMG_KO" -eq 0 ] && [ "$IMG_EN" -eq 0 ]; then
  echo "✅ Image paths OK"
else
  echo "❌ Image paths FAIL (KO: $IMG_KO, EN: $IMG_EN relative paths)"
  exit 1
fi

# Test 3: 내부 링크 검증
echo "Test 3: Internal links"
LINK_KO=$(grep -r "\[.*\](\.\./" pages/ko/ | wc -l)
LINK_EN=$(grep -r "\[.*\](\.\./" pages/en/ | wc -l)
if [ "$LINK_KO" -eq 0 ] && [ "$LINK_EN" -eq 0 ]; then
  echo "✅ Internal links OK"
else
  echo "❌ Internal links FAIL (KO: $LINK_KO, EN: $LINK_EN relative links)"
  exit 1
fi

# Test 4: 빌드 테스트
echo "Test 4: Build"
npm run build > /dev/null 2>&1
if [ $? -eq 0 ]; then
  echo "✅ Build OK"
else
  echo "❌ Build FAIL"
  exit 1
fi

echo "=== All tests passed ==="
```

### Reference Links
- [Nextra MDX Documentation](https://nextra.site/docs/guide/markdown)
- [MkDocs to Nextra Migration Guide](https://nextra.site/docs/guide/migration)
- [MDX Syntax Reference](https://mdxjs.com/docs/)
