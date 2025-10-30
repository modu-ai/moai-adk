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

# SPEC-NEXTRA-CONTENT-001 Implementation Plan

## HISTORY

### v0.0.1 (2025-10-31)
- **INITIAL**: mkdocs → Nextra 콘텐츠 마이그레이션 실행 계획 작성
- **AUTHOR**: @GOOS
- **SECTIONS**: Milestones, Technical Approach, Migration Strategy, Risks

---

## Milestones (우선순위 기반)

### Primary Goal: 원본 백업 및 마이그레이션 스크립트 준비
**목표**: 원본 `/docs/` 디렉토리를 백업하고 마이그레이션 스크립트를 작성하여 검증

**완료 조건**:
- ✅ `/docs/` 디렉토리를 Git 커밋하여 백업 완료
- ✅ 마이그레이션 스크립트 (`scripts/migrate-content.js`) 작성 완료
- ✅ 40개 페이지 목록 정의 완료
- ✅ 이미지 경로 수정 로직 검증 완료
- ✅ 내부 링크 수정 로직 검증 완료

**주요 작업**:
1. `/docs/` 디렉토리 현황 파악 (총 70개 페이지)
2. 마이그레이션 대상 40개 페이지 선정 (조회수 또는 중요도 기준)
3. Git 백업 커밋: `git commit -m "backup: preserve original docs before migration"`
4. `scripts/migrate-content.js` 스크립트 작성
   - Markdown → MDX 변환 로직
   - 이미지 경로 수정 (`../images/` → `/images/`)
   - 내부 링크 수정 (`../guides/quickstart.md` → `/ko/guides/quickstart`)
5. 테스트 페이지 1-2개로 스크립트 검증

---

### Secondary Goal: 40개 페이지 변환 및 `_meta.json` 생성
**목표**: 40개 페이지를 MDX로 변환하고 네비게이션 구조를 정의

**완료 조건**:
- ✅ `/pages/ko/`, `/pages/en/` 디렉토리에 40개 페이지 변환 완료
- ✅ 모든 하위 디렉토리에 `_meta.json` 파일 생성 완료
- ✅ 이미지 경로 상대 경로 0개 확인
- ✅ 내부 링크 상대 경로 0개 확인
- ✅ Front Matter (title, description) 추가 완료

**주요 작업**:
1. 마이그레이션 스크립트 실행:
   ```bash
   node scripts/migrate-content.js
   ```
2. `_meta.json` 파일 생성:
   - `/pages/ko/_meta.json` (한국어 네비게이션)
   - `/pages/en/_meta.json` (영어 네비게이션)
   - 각 하위 디렉토리에도 `_meta.json` 생성
3. 이미지 경로 검증:
   ```bash
   grep -r "\!\[.*\](\.\./" pages/ko/
   grep -r "\!\[.*\](\.\./" pages/en/
   ```
4. 내부 링크 검증:
   ```bash
   grep -r "\[.*\](\.\./" pages/ko/
   grep -r "\[.*\](\.\./" pages/en/
   ```
5. Front Matter 수동 추가 (자동 생성된 부분 검토)

---

### Final Goal: 빌드 검증 및 Vercel 배포
**목표**: 변환된 페이지가 빌드되고 Vercel에 배포되어 정상 접속되는 것을 확인

**완료 조건**:
- ✅ `npm run build` 실행 시 오류 없이 빌드 완료
- ✅ `out/ko/`, `out/en/` 디렉토리에 모든 페이지 생성 확인
- ✅ Vercel 배포 후 https://adk.mo.ai.kr/ko/, https://adk.mo.ai.kr/en/ 정상 접속
- ✅ 모든 페이지 수동 클릭 테스트 (404 오류 0개)
- ✅ Mermaid 다이어그램 렌더링 확인

**주요 작업**:
1. 로컬 빌드 테스트:
   ```bash
   npm run build
   ls -la out/ko/
   ls -la out/en/
   ```
2. 빌드 오류 수정 (있을 경우)
3. Vercel 배포:
   ```bash
   git add .
   git commit -m "feat: migrate 40 core pages from mkdocs to Nextra"
   git push origin main
   ```
4. 배포 후 수동 검증:
   - 모든 페이지 클릭 테스트 (404 오류 확인)
   - 이미지 로딩 확인
   - 내부 링크 동작 확인
   - Mermaid 다이어그램 렌더링 확인
5. 변환 보고서 생성 (성공/실패/경고 목록)

---

## Technical Approach (기술 접근 방식)

### 1. 수동 vs. 자동 변환
- **선택**: 반자동 방식 (스크립트 + 수동 검토)
- **이유**:
  - 완전 자동화는 오류 위험 높음 (링크 깨짐, 이미지 누락)
  - 40개 페이지는 수동 검토 가능한 규모
  - 품질 우선: 각 페이지 검토하며 변환

### 2. Markdown → MDX 변환 전략
- **단순 변환**: 대부분의 마크다운 문법은 MDX와 호환됨
- **추가 작업**:
  - Front Matter 추가 (title, description)
  - 이미지 경로 수정 (`../images/` → `/images/`)
  - 내부 링크 수정 (상대 경로 → 절대 경로)
  - Mermaid 다이어그램 코드 블록 유지

### 3. 이미지 관리 전략
- **위치**: 모든 이미지는 `/public/images/` 디렉토리에 저장
- **경로**: 절대 경로 사용 (`/images/logo.png`)
- **장점**:
  - 언어별 디렉토리 구조 변경에 영향 없음
  - 빌드 시 이미지 자동 복사 (Next.js 기본 기능)

### 4. 네비게이션 메뉴 구조 (`_meta.json`)
- **언어별 구분**: `/pages/ko/_meta.json`, `/pages/en/_meta.json`
- **계층 구조**: 각 하위 디렉토리에도 `_meta.json` 생성
- **순서 제어**: `_meta.json`에서 메뉴 순서 정의

---

## Architecture Design (아키텍처 설계)

### Migration Flow
```
┌─────────────────────────────────────────────────────────────┐
│                  Original MkDocs Content                    │
│                  /docs/ (70 pages)                          │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ↓
┌─────────────────────────────────────────────────────────────┐
│              Page Selection (40 core pages)                 │
│  - 홈, 소개, 시작하기, 핵심 개념, 워크플로우, Skills, ...   │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ↓
┌─────────────────────────────────────────────────────────────┐
│           Migration Script (migrate-content.js)             │
│  1. Read Markdown file                                      │
│  2. Convert to MDX (add Front Matter)                       │
│  3. Fix image paths (../images/ → /images/)                 │
│  4. Fix internal links (../guides/... → /ko/guides/...)     │
│  5. Write to /pages/ko/ and /pages/en/                      │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ↓
┌─────────────────────────────────────────────────────────────┐
│              Nextra MDX Content (80 files)                  │
│  - /pages/ko/ (40 pages)                                    │
│  - /pages/en/ (40 pages)                                    │
│  - _meta.json files (navigation structure)                  │
└──────────────────────┬──────────────────────────────────────┘
                       │
                       ↓
┌─────────────────────────────────────────────────────────────┐
│                  Build & Deploy                             │
│  - npm run build → out/ko/, out/en/                         │
│  - Vercel deployment                                        │
└─────────────────────────────────────────────────────────────┘
```

### Content Structure Mapping
```
MkDocs Structure                  Nextra Structure
────────────────────────────────────────────────────────────
/docs/
  index.md                    →   /pages/ko/index.mdx
                                  /pages/en/index.mdx

  introduction/               →   /pages/ko/introduction/
    overview.md                     overview.mdx
    architecture.md                 architecture.mdx
                                  /pages/en/introduction/
                                    overview.mdx
                                    architecture.mdx

  getting-started/            →   /pages/ko/getting-started/
    installation.md                 installation.mdx
    quickstart.md                   quick-start.mdx
                                  /pages/en/getting-started/
                                    installation.mdx
                                    quick-start.mdx

  images/                     →   /public/images/
    logo.png                        logo.png
```

---

## Migration Strategy (마이그레이션 전략)

### Phase 1: Preparation (준비)
1. **원본 백업**: Git 커밋하여 `/docs/` 디렉토리 보존
2. **페이지 선정**: 40개 핵심 페이지 목록 작성
3. **스크립트 작성**: `scripts/migrate-content.js` 작성
4. **테스트 실행**: 1-2개 페이지로 스크립트 검증

### Phase 2: Conversion (변환)
1. **자동 변환**: 마이그레이션 스크립트 실행
2. **수동 검토**: 각 페이지 Front Matter 검토 및 수정
3. **이미지 복사**: `/docs/images/` → `/public/images/`
4. **_meta.json 생성**: 네비게이션 구조 정의

### Phase 3: Validation (검증)
1. **이미지 경로 검증**: 상대 경로 0개 확인
2. **내부 링크 검증**: 상대 경로 0개 확인
3. **빌드 테스트**: `npm run build` 성공 확인
4. **수동 클릭 테스트**: 모든 페이지 404 오류 확인

### Phase 4: Deployment (배포)
1. **Git 커밋**: 변환된 페이지 커밋
2. **Vercel 배포**: Git 푸시하여 자동 배포
3. **프로덕션 검증**: 배포 후 모든 페이지 수동 확인
4. **변환 보고서 작성**: 성공/실패/경고 목록

---

## Dependency Management (의존성 관리)

### 선행 조건
| SPEC | 상태 | 필요성 |
|------|------|--------|
| NEXTRA-SITE-001 | 완료 필요 | Next.js + Nextra 기본 구조 |
| NEXTRA-I18N-001 | 완료 필요 | `/pages/ko/`, `/pages/en/` 디렉토리 구조 |

### 후속 작업
| Task | 의존성 |
|------|--------|
| 나머지 30개 페이지 마이그레이션 | NEXTRA-CONTENT-001 완료 후 |
| 콘텐츠 업데이트 및 개선 | Phase 2 (v1.1) |
| 스크린샷 재생성 | Phase 2 (v1.1) |

---

## Risks & Mitigation (위험 요소 및 대응 방안)

### Risk 1: 마이그레이션 스크립트 오류로 일부 페이지 누락
- **확률**: Medium
- **영향**: High (특정 페이지 접근 불가)
- **대응**:
  - 마이그레이션 전 원본 백업 (Git 커밋)
  - 변환 결과 보고서 생성 (성공/실패 목록)
  - 수동 검증: `find pages/ko -name "*.mdx" | wc -l` (예상: 40)

### Risk 2: 이미지 경로 오류로 이미지 깨짐
- **확률**: High
- **영향**: Medium (일부 이미지 표시 안 됨)
- **대응**:
  - 이미지 경로 검증 스크립트 실행:
    ```bash
    grep -r "\!\[.*\](\.\./" pages/ko/
    ```
  - 브라우저 개발자 도구에서 404 오류 확인
  - 누락된 이미지는 `/public/images/`로 복사

### Risk 3: 내부 링크 깨짐으로 404 오류
- **확률**: High
- **영향**: High (사용자 경험 저하)
- **대응**:
  - 링크 검증 스크립트 실행:
    ```bash
    grep -r "\[.*\](\.\./" pages/ko/
    ```
  - 빌드 후 모든 페이지 수동 클릭 테스트
  - 깨진 링크는 즉시 수정

### Risk 4: Mermaid 다이어그램 렌더링 실패
- **확률**: Low
- **영향**: Medium (일부 다이어그램 표시 안 됨)
- **대응**:
  - Nextra mermaid 플러그인 설치 확인
  - 브라우저에서 다이어그램 렌더링 확인
  - 렌더링 실패 시 이미지로 대체

### Risk 5: 빌드 시간 증가
- **확률**: Medium
- **영향**: Low (빌드 시간 3분 초과 가능)
- **대응**:
  - 빌드 시간 측정: `time npm run build`
  - Incremental Static Regeneration (ISR) 고려 (Phase 2)

---

## Testing Strategy (테스트 전략)

### Unit Tests (Phase 1에서는 제외)
- **이유**: 정적 콘텐츠 변환이므로 복잡한 로직 없음
- **Phase 2 고려 사항**: 링크 검증 테스트 자동화

### Integration Tests
- **빌드 테스트**:
  - `npm run build` 실행 → 오류 없이 완료
  - `out/ko/`, `out/en/` 디렉토리 존재 확인
- **이미지 경로 테스트**:
  - 상대 경로 이미지 0개 확인
- **내부 링크 테스트**:
  - 상대 경로 링크 0개 확인

### E2E Tests (Phase 1에서는 수동)
- **로컬 개발 서버 테스트**:
  - `npm run dev` 실행 → 모든 페이지 접속
  - 이미지 로딩 확인
  - 내부 링크 클릭 확인

- **배포 후 프로덕션 테스트**:
  - https://adk.mo.ai.kr/ko/ 접속 → 모든 페이지 수동 클릭
  - https://adk.mo.ai.kr/en/ 접속 → 모든 페이지 수동 클릭
  - Mermaid 다이어그램 렌더링 확인

---

## Validation Scripts (검증 스크립트)

### 1. 이미지 경로 검증 스크립트
```bash
#!/bin/bash
# scripts/validate-image-paths.sh

echo "=== Validating image paths ==="

# 한국어 페이지 확인
KO_RELATIVE=$(grep -r "\!\[.*\](\.\./" pages/ko/ | wc -l)
if [ "$KO_RELATIVE" -eq 0 ]; then
  echo "✅ Korean pages: No relative image paths"
else
  echo "❌ Korean pages: $KO_RELATIVE relative image paths found"
  grep -r "\!\[.*\](\.\./" pages/ko/
  exit 1
fi

# 영어 페이지 확인
EN_RELATIVE=$(grep -r "\!\[.*\](\.\./" pages/en/ | wc -l)
if [ "$EN_RELATIVE" -eq 0 ]; then
  echo "✅ English pages: No relative image paths"
else
  echo "❌ English pages: $EN_RELATIVE relative image paths found"
  grep -r "\!\[.*\](\.\./" pages/en/
  exit 1
fi

echo "=== Image path validation passed ==="
```

### 2. 내부 링크 검증 스크립트
```bash
#!/bin/bash
# scripts/validate-internal-links.sh

echo "=== Validating internal links ==="

# 한국어 페이지 확인
KO_RELATIVE=$(grep -r "\[.*\](\.\./" pages/ko/ | wc -l)
if [ "$KO_RELATIVE" -eq 0 ]; then
  echo "✅ Korean pages: No relative internal links"
else
  echo "❌ Korean pages: $KO_RELATIVE relative links found"
  grep -r "\[.*\](\.\./" pages/ko/
  exit 1
fi

# 영어 페이지 확인
EN_RELATIVE=$(grep -r "\[.*\](\.\./" pages/en/ | wc -l)
if [ "$EN_RELATIVE" -eq 0 ]; then
  echo "✅ English pages: No relative internal links"
else
  echo "❌ English pages: $EN_RELATIVE relative links found"
  grep -r "\[.*\](\.\./" pages/en/
  exit 1
fi

echo "=== Internal link validation passed ==="
```

### 3. 페이지 수 검증 스크립트
```bash
#!/bin/bash
# scripts/validate-page-count.sh

echo "=== Validating page count ==="

KO_COUNT=$(find pages/ko -name "*.mdx" | wc -l)
EN_COUNT=$(find pages/en -name "*.mdx" | wc -l)

echo "Korean pages: $KO_COUNT (expected: 40)"
echo "English pages: $EN_COUNT (expected: 40)"

if [ "$KO_COUNT" -eq 40 ] && [ "$EN_COUNT" -eq 40 ]; then
  echo "✅ Page count validation passed"
else
  echo "❌ Page count mismatch"
  exit 1
fi
```

---

## Performance Optimization (성능 최적화)

### Build Time Optimization
- **목표**: 빌드 시간 증가 < 20% (NEXTRA-I18N-001 대비)
- **전략**:
  - 이미지 최적화: WebP 형식 사용 (Phase 2)
  - 불필요한 페이지 제외 (나머지 30개는 Phase 2)

### Content Loading Optimization
- **목표**: 페이지 로딩 시간 < 2초
- **전략**:
  - 정적 사이트 생성 (SSG): 모든 페이지 빌드 시 생성
  - CDN 캐싱: Vercel Edge Network 활용

---

## Rollback Strategy (롤백 전략)

### Scenario 1: 마이그레이션 스크립트 실행 중 오류 발생
- **대응**:
  1. 스크립트 중단: `Ctrl+C`
  2. Git으로 복원: `git checkout -- pages/`
  3. 스크립트 수정 후 재실행

### Scenario 2: 빌드 실패 (변환된 페이지 문제)
- **대응**:
  1. 빌드 로그 확인: `npm run build 2>&1 | tee build.log`
  2. 오류 페이지 식별 및 수정
  3. 재빌드: `npm run build`

### Scenario 3: 배포 후 404 오류 다수 발생
- **대응**:
  1. Vercel 대시보드에서 이전 배포 버전으로 롤백
  2. 로컬에서 링크 검증 재실행
  3. 문제 수정 후 재배포

---

## Documentation Plan (문서 작성 계획)

### 1. 콘텐츠 마이그레이션 가이드 (`@DOC:CONTENT-MIGRATION-001`)
**내용**:
- 마이그레이션 스크립트 사용법
- 이미지 경로 수정 규칙
- 내부 링크 수정 규칙
- `_meta.json` 작성 방법

### 2. MDX 작성 스타일 가이드 (`@DOC:CONTENT-STYLE-001`)
**내용**:
- Front Matter 작성 규칙
- Mermaid 다이어그램 사용법
- 코드 블록 작성법
- 이미지 삽입 방법

---

## Success Criteria (성공 기준)

### Milestone 1 완료 기준
- ✅ 원본 `/docs/` 디렉토리 백업 완료
- ✅ 마이그레이션 스크립트 작성 완료

### Milestone 2 완료 기준
- ✅ 40개 페이지 변환 완료
- ✅ `_meta.json` 파일 생성 완료

### Milestone 3 완료 기준
- ✅ 빌드 성공
- ✅ Vercel 배포 성공

### 전체 SPEC 완료 기준
- ✅ 모든 Milestone 완료
- ✅ 콘텐츠 마이그레이션 가이드 작성 완료
- ✅ Git 커밋 및 PR 생성 완료

---

## Next Steps (다음 단계)

### 의존 관계
- **선행 작업**: @SPEC:NEXTRA-SITE-001, @SPEC:NEXTRA-I18N-001
- **후속 작업**: 나머지 30개 페이지 마이그레이션 (Phase 2)

### Handoff Points
- **구현 담당**: TDD-implementer (via `/alfred:2-run SPEC-NEXTRA-CONTENT-001`)
- **문서 작성**: doc-syncer (via `/alfred:3-sync`)
- **Git 관리**: git-manager (자동)

---

## Appendix

### Useful Commands
```bash
# 원본 백업
git add docs/
git commit -m "backup: preserve original docs before migration"

# 마이그레이션 스크립트 실행
node scripts/migrate-content.js

# 이미지 경로 검증
bash scripts/validate-image-paths.sh

# 내부 링크 검증
bash scripts/validate-internal-links.sh

# 페이지 수 검증
bash scripts/validate-page-count.sh

# 빌드 및 배포
npm run build
git add pages/ public/
git commit -m "feat: migrate 40 core pages from mkdocs to Nextra"
git push origin main
```

### Reference Links
- [Nextra MDX Guide](https://nextra.site/docs/guide/markdown)
- [MkDocs to Nextra Migration](https://nextra.site/docs/guide/migration)
- [MDX Syntax](https://mdxjs.com/docs/)
